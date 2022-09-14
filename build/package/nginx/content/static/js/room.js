/* Constants */

const UNKNOWN_USER_NAME = "~Unknown";
const DELAY_CONTEXT_MENU_AFTER_EXIT_BIG_MODE_MS = 250;

const TEXT_SEARCH_AREA_MSG_ID = 'message-id';
const TEXT_SEARCH_AREA_MSG_AUTHOR = 'author-name';
const TEXT_SEARCH_AREA_MSG_TIME = 'message-time';
const TEXT_SEARCH_AREA_R_MESSAGE_ID = 'reply-to-message-id';
const TEXT_SEARCH_AREA_R_USER = 'reply-to-user-name';
const TEXT_SEARCH_AREA_R_MESSAGE = 'reply-to-message-text';
const TEXT_SEARCH_AREA_TEXT = 'text';

const CLASS_MESSAGE_TO_CLEAN_UP = 'message-to-cleanup';

/* Variables */

let isRoomReconnectInProgress = false;

let clientBotsCluster;

let lastUserInfoByIdCache;
let lastUserListTimestamp;
//if some messages came before their user is known - remember them and wait for further 'members changed' messages to provide user name
let userIdToMessageIdsWithUnknownAuthor;
let userIdToMessageIdsWithUnknownReplyToUserName;

let roomMessageIdToDOMElem;
let folkPicksMessageIdToDOMElem;
let messageIdToTextSearchInfo;

let roomUUID;
let roomHasPassword;
let userInRoomUUID;
let roomCreatorUserInRoomUUID;
let currentUserInfo;

let currentRoomName;
let isLoggedIn;
let isRoomDescriptionLoaded;

let isRedirectedFromHomePageRoomForm;

let isDrawingSendInProgress;

let messageEditingInProgressId;
let messageTextSelectionInProgressId;
//possible for desktop only
let messageTextSelectionInProgressIsFolkPick;

let userReplyInProgressToId;
let messageReplyInProgressToId;

let nextUserActionIndicationId;
let currentlyExpectedToCompleteUserActionIndicationId;

let lastCurrentUserSentMessageId;

let textareaBigModeExitedAt;
let textSearchDataChangedAt;

let deletedMessageIds;

let wsReconnectTimeout;
let onSelectionChangeTimeout;
let onScrollBodyTimeout;
let onMessageTextTypingTimeout;
let longTouchTimeout;

let isInsideLongTouch = false;

let reconnectAttempt = 1;

//text search context
let currentTextSearchContext;

//queue of unprocessed commands, (ones that depend on some message and thus waiting for it)
let commandsToProcessChangeRoomDescription;
let commandsToProcessEditMessage;
let commandsToProcessDeleteMessage;
let commandsToProcessVoteMessage;

//last processed command timestamps
let lastChangeRoomDescription;


function initializeVariables () {
    clientBotsCluster = {};

    lastUserInfoByIdCache = {};
    lastUserListTimestamp = null;

    userIdToMessageIdsWithUnknownAuthor = {};
    userIdToMessageIdsWithUnknownReplyToUserName = {};

    roomMessageIdToDOMElem = {};
    folkPicksMessageIdToDOMElem = {};
    messageIdToTextSearchInfo = {};

    roomUUID = null;
    roomHasPassword = false;
    userInRoomUUID = null;
    currentUserInfo = null;
    roomCreatorUserInRoomUUID = null;

    currentRoomName = null;
    isLoggedIn = false;
    isRoomDescriptionLoaded = false;

    isRedirectedFromHomePageRoomForm = false;

    isDrawingSendInProgress = false;

    messageEditingInProgressId = null;
    messageTextSelectionInProgressId = null;
    messageTextSelectionInProgressIsFolkPick = false;

    userReplyInProgressToId = null;
    messageReplyInProgressToId = null;

    nextUserActionIndicationId = 0;
    currentlyExpectedToCompleteUserActionIndicationId = null;

    lastCurrentUserSentMessageId = 0;

    textareaBigModeExitedAt = 0;
    textSearchDataChangedAt = 0;

    deletedMessageIds = {};

    wsReconnectTimeout = -1;
    onSelectionChangeTimeout = -1;
    onScrollBodyTimeout = -1;
    onMessageTextTypingTimeout = -1;

    currentTextSearchContext = {};

    commandsToProcessChangeRoomDescription = [];
    commandsToProcessEditMessage = {};
    commandsToProcessDeleteMessage = {};
    commandsToProcessVoteMessage = {};

    lastChangeRoomDescription = null;
}

function init () {
    initializeVariables();

    keepAlive();

    //get redirect variable (if any). If this page loaded after redirect from home page (using room form) - must draw notification
    const redirectVariableStr = LOCAL_STORAGE.getItem(REDIRECT_VARIABLE_LOCAL_STORAGE_KEY);

    if (redirectVariableStr) {
        const redirectVariable = JSON.parse(redirectVariableStr);

        //check if redirect is from home page and if variable is not expired
        if (redirectVariable.redirectFrom === REDIRECT_FROM_HOME_PAGE_ROOM_FORM &&
            new Date().getTime() - redirectVariable.redirectedAt < REDIRECT_VARIABLE_TTL_MS) {

            isRedirectedFromHomePageRoomForm = true;

            if (redirectVariable.isCreatedRoom) {
                showNotification(REQUEST_PROCESSING_DETAILS_ROOM_CREATED_MESSAGE, NOTIFICATION_SHOW_MS);
            } else {
                showNotification(REQUEST_PROCESSING_DETAILS_ROOM_JOINED_MESSAGE, NOTIFICATION_SHOW_MS);
            }
        }

        LOCAL_STORAGE.removeItem(REDIRECT_VARIABLE_LOCAL_STORAGE_KEY);
    }

    /* Create or join room, passed in page query path */
    createOrJoinRoom();

    //load visited rooms
    loadVisitedRooms(visitedRoomOnClickCallback);

    /* Setup UI */

    $roomInfoLeaveRoomBtn.on('click', function () {
        shutdownSocket();
        redirectToURL("/");
    });

    $userJoinedAsChangeLink.on('click', onChangeUserName);

    $sendMessageButton.on('mousedown', function (e) {
        stopPropagationAndDefault(e);

        const sentOk = sendUserMessage();

        if (sentOk) {
            LOCAL_STORAGE.setItem(LAST_UNSENT_MESSAGE_TEXT_LOCAL_STORAGE_KEY, '');
        }
    });

    $messageEditCancelButton.on('click', function (e) {
        stopPropagationAndDefault(e);

        const isChatAtBottom = isChatScrolledBottom();

        cancelMessageEdit();
        resizeWrappersHeight(isChatAtBottom);
    });

    $userJoinedAsNameWr.on('click', showUserNameChangingBlock);
    $userJoinedAsTitle.on('click', showUserNameChangingBlock);

    $userMessageTextarea.on('focusin', textareaSwitchToBigMode);
    $userMessageTextarea.on('focusout', textareaSwitchToSmallMode);

    $answerToUserCloseButton.on('mousedown', cancelReplyToUser);

    $answerToMessageCloseButton.on('mousedown', cancelReplyToMessage);

    $searchBarPrevButton.on('click', textSearchPrev);
    $searchBarNextButton.on('click', textSearchNext);

    //when text selection happens
    $document.on('selectionchange', onSelectionChange);

    $body.on('scroll', moveMobileTextSelectionControlsToMessageDelayed);

    initUserInputDrawing();


    /* User bots */

    $botsShowExampleBtn.on('click', toggleBotExample);

    $botsAddNewBtn.on('click', function () {
        hideBotExample();

        const newBotId = new Date().getTime();

        const $newBotBlock = $botsEmptyBotForCloningBlock.clone();
        $newBotBlock.removeClass('empty-bot-dom-for-cloning');
        $newBotBlock.removeClass('d-none');

        $newBotBlock.attr('data-bot-id', newBotId);

        $newBotBlock.find('.bot-delete-btn').on('click', onBotDeleteClick);

        $newBotBlock.find('.bots-list-item-enabled, .bots-item-keyphrase, .bots-item-keyusername, .bots-item-template')
            .on('change keyup paste', onBotConfigChangedDelayedFunc);

        $botsListWr.append($newBotBlock);
    });

    let soredBotsClusterStr = LOCAL_STORAGE.getItem(BOTS_LIST_LOCAL_STORAGE_KEY);
    if (!soredBotsClusterStr) {
        clientBotsCluster = {};
    } else {
        clientBotsCluster = JSON.parse(soredBotsClusterStr);

        for (const botId in clientBotsCluster) {
            const botInfo = clientBotsCluster[botId];

            const $botBlock = $botsEmptyBotForCloningBlock.clone();
            $botBlock.removeClass('empty-bot-dom-for-cloning');
            $botBlock.removeClass('d-none');

            $botBlock.attr('data-bot-id', botId);

            $botBlock.find('.bots-list-item-enabled').prop('checked', botInfo.isEnabled);

            if (botInfo.matchStr) {
                $botBlock.find('.bots-item-keyphrase').val(botInfo.matchStr);
            }
            if (botInfo.matchUser) {
                $botBlock.find('.bots-item-keyusername').val(botInfo.matchUser);
            }
            if (botInfo.matchStrPayload) {
                $botBlock.find('.bots-item-template').val(botInfo.matchStrPayload);
            }

            $botBlock.find('.bot-delete-btn').on('click', onBotDeleteClick);

            $botBlock.find('.bots-list-item-enabled, .bots-item-keyphrase, .bots-item-keyusername, .bots-item-template')
                .on('change keyup paste', onBotConfigChangedDelayedFunc);

            $botsListWr.append($botBlock);

            //re-init bot
            clientBotsCluster[botId].onMessage = onRoomMessageBotAction;
            clientBotsCluster[botId].vars = {
                counter: 0,
            };
        }
    }

    //if only example and placeholder bot blocks are present - show example
    if ($('.bots-list-item-wr').length === 2) {
        toggleBotExample();
    }

    /* Set text fields keyboard behaviour */

    $userMessageTextarea.on('keydown', function (e) {
        if (e.which === KEY_CODE_ARROW_UP) {
            if ($userMessageTextarea.val().trim() === '' && lastCurrentUserSentMessageId) {
                const $lastCurrentUserSentMessageBlock = $roomMessagesWr.find('div[data-msg-id=' + lastCurrentUserSentMessageId + ']');
                editUserMessage(lastCurrentUserSentMessageId, $lastCurrentUserSentMessageBlock)

                scrollToTargetMsg($lastCurrentUserSentMessageBlock);

                return;
            }
        }

        if (e.which === KEY_CODE_ENTER && !e.shiftKey && !isMobileClientDevice) {
            stopPropagationAndDefault(e);

            const sentOk = sendUserMessage();

            if (sentOk) {
                clearTimeout(onMessageTextTypingTimeout);
                LOCAL_STORAGE.setItem(LAST_UNSENT_MESSAGE_TEXT_LOCAL_STORAGE_KEY, '');

                return;
            }
        }

        clearTimeout(onMessageTextTypingTimeout);

        if (!messageEditingInProgressId) {
            onMessageTextTypingTimeout = setTimeout(function () {
                LOCAL_STORAGE.setItem(LAST_UNSENT_MESSAGE_TEXT_LOCAL_STORAGE_KEY, $userMessageTextarea.val());
            }, 100);
        }
    });

    $userJoinedAsChangeInput.on('keydown', function (e) {
        if (e.which === KEY_CODE_ENTER) {
            onChangeUserName(e);
        }
    });

    $roomInfoDescriptionCreatorChangeInput.on('keydown', function (e) {
        if (e.which === KEY_CODE_ENTER) {
            onChangeRoomDescription(e);
        }
    });

    $searchBarInput.on('keydown', function (e) {
        if (e.which === KEY_CODE_ENTER) {
            textSearchPrev();
        }
    });

    $document.on('keydown', function(e) {
        if (e.which === KEY_CODE_ESCAPE) {
            $messageContextMenu.addClass('d-none');
            hideFolkPicksMobile();
            hideMenuMobile();
            hideMyRecentRoomsPopup();
            hideShareRoomPopup();
            hideUserInputDrawingBlock();
            hideBotsUI();

            $globalTransparentOverlay.addClass('d-none');
            cancelActionsUnderGlobalTransparentOverlay();

            cancelMessageTextSelectionMode();

            resizeWrappersHeight();
        }
    });
}

function moveMobileTextSelectionControlsToMessageDelayed () {
    clearTimeout(onScrollBodyTimeout);

    onScrollBodyTimeout = setTimeout(moveMobileTextSelectionControlsToMessage, 5);
}

function moveMobileTextSelectionControlsToMessage () {
    if (messageTextSelectionInProgressId) {
        const $messageBlock = findMessageBlockById(messageTextSelectionInProgressId);

        //horizontal offset
        if (currentViewportWidth < MOBILE_CONTAINER_MAIN_LARGE_PADDING_STARTS_FROM_WIDTH_PX) {
            $roomMessagesWr.css('padding-right', '2.1rem');
        } else {
            $roomMessagesWr.css('padding-right', '1.6rem');
        }

        if ($containerMainCenterWr.hasClass('col-xl-9')) {
            $containerMainCenterWr
                .addClass('col-xl-12')
                .removeClass('col-xl-9');
        }

        //take dimensions after block classes changed
        const rect = $messageBlock[0].getBoundingClientRect();

        //vertical offset
        const messageTopEdgeToViewportTopPx = rect.top;
        const messageBotEdgeToViewportTopPx = rect.bottom;


        /* Move text selection control buttons */

        const defaultTopOffset = currentViewportHeight * 0.75;

        const approxControlsBlockHeightPx = 72;

        let controlsTopOffset;

        //if default controls offset is NOT between message top/bot edges stick buttons to top or bot
        if (messageTopEdgeToViewportTopPx < defaultTopOffset && messageBotEdgeToViewportTopPx - approxControlsBlockHeightPx < defaultTopOffset) {
            controlsTopOffset = messageBotEdgeToViewportTopPx - approxControlsBlockHeightPx;

        } else if (messageTopEdgeToViewportTopPx > defaultTopOffset && messageBotEdgeToViewportTopPx - approxControlsBlockHeightPx > defaultTopOffset) {
            controlsTopOffset = messageTopEdgeToViewportTopPx;

        } else {
            //default offset is between top/bot edges (i.e. inside message). Use default one
            controlsTopOffset = defaultTopOffset;
        }

        $mobileTextSelectionButtonsBlock.css('top', controlsTopOffset + 'px');


        /* Move text selection to-message overlay */

        //some margin in px, to allow showing to-message overlay while control buttons are still visible
        const showOverlayEarlierMargin = 15;

        if (messageTopEdgeToViewportTopPx > (currentViewportHeight - showOverlayEarlierMargin)) {
            $mobileTextSelectionToMessageOverlayTopBlock.removeClass('d-none');
            $mobileTextSelectionToMessageOverlayBotBlock.addClass('d-none');

        } else if (messageBotEdgeToViewportTopPx < showOverlayEarlierMargin) {
            $mobileTextSelectionToMessageOverlayTopBlock.addClass('d-none');
            $mobileTextSelectionToMessageOverlayBotBlock.removeClass('d-none');

        } else {
            $mobileTextSelectionToMessageOverlayTopBlock.addClass('d-none');
            $mobileTextSelectionToMessageOverlayBotBlock.addClass('d-none');
        }
    }
}

function onSelectionChange (e) {
    stopPropagationAndDefault(e);

    /* Start message text selection mode, using timer to singleton it*/

    clearTimeout(onSelectionChangeTimeout);

    onSelectionChangeTimeout = setTimeout(function () {
        const someTextIsSelected = !!getSelectionText();

        //if already in text selection mode -
        if (messageTextSelectionInProgressId) {
            const newTextSelectionStartNode = getTextSelectionStartNode();

            if (isMobileClientDevice) {
                transformMobileTextSelectionCopyButton(someTextIsSelected);
            } else {
                transformDesktopTextSelectionCopyButton(someTextIsSelected);
            }

            //if selection is removed but message is still in 'select text' mode - do nothing
            if (!newTextSelectionStartNode) {
                return;

            } else {
                const $currentlySelectedMessageBlock = messageTextSelectionInProgressIsFolkPick
                    ? findFolkPicksMessageBlockById(messageTextSelectionInProgressId)
                    : findMessageBlockById(messageTextSelectionInProgressId);

                const $newTextSelectionStartNodeParentMessage = $(newTextSelectionStartNode).closest('.room-msg-main-wr');

                //if selection starts in any child of currently selected message - do nothing
                if ($newTextSelectionStartNodeParentMessage.length &&
                    $newTextSelectionStartNodeParentMessage[0].isEqualNode($currentlySelectedMessageBlock[0])) {

                    return;
                } else {
                    //user attempts to select text in another message - restart selection
                    cancelMessageTextSelectionMode();
                }
            }
        }

        //start/cancel selection mode for message
        if (someTextIsSelected) {
            const $messageBlock = $(window.getSelection().anchorNode.parentElement)
                .closest('.room-msg-main-wr');

            if ($messageBlock.length) {
                startMessageTextSelectionMode($messageBlock);
            } else {
                cancelMessageTextSelectionMode();
            }

        } else {
            cancelMessageTextSelectionMode();
        }
    }, 20);
}

function createOrJoinRoom () {
    let roomName = formatRoomNameInput(ROOM_NAME);

    //if room name passed in query path is invalid
    if (!roomName || !validateRoomName(roomName)) {
        //save validation result and roomname into local storage before redirect
        redirectToHomePageWithError(ROOM_TO_HOME_PG_REDIRECT_ERROR_BAD_ROOM_NAME_LENGTH);
    } else {
        authorizeInRoom(roomName);
    }
}

function authorizeInRoom(roomName) {
    showSpinnerOverlay();

    //retrieve proper backend address for provided roomName, then open websocket to that backend and send 'room-create/join auth-only request'
    pickBackend(roomName, function (roomName, backendInstanceAddr, alternativeRoomNamePostfixes) {
        initRoomPageWebSocket(roomName, backendInstanceAddr, alternativeRoomNamePostfixes, function (roomName) {
            currentRoomName = roomName;

            sendData(ws,
                JSON.stringify({
                    p: 'unknown'
                })
            );

            sendData(ws,
                JSON.stringify({
                    c: COMMANDS.RoomCreateJoin,
                    uN: null,
                    rq: "room_c_j_done",
                    r: {
                        n: ROOM_NAME
                    }
                })
            );

            reconnectAttempt = 1;
        });
    });
}

function initRoomPageWebSocket (roomName, backendInstanceAddr, alternativeRoomNamePostfixes, onOpen) {
    closeWsIfExists();

    try {
        ws = new WebSocket(WS_ENDPOINT + '?' + URL_PARAM_BACKEND_HOST + '=' + backendInstanceAddr);
    } catch (ex) {
        redirectToHomePageWithError(ROOM_TO_HOME_PG_REDIRECT_ERROR_CONNECTION);
        return;
    }

    ws.onopen = function () {
        onOpen(roomName);
    };

    ws.onerror = function () {
        if (isLoggedIn || isRoomReconnectInProgress) {
            ws.onerror = function () {};
            ws.onclose = function () {};
            reconnectToRoom();
        } else {
            redirectToHomePageWithError(ROOM_TO_HOME_PG_REDIRECT_ERROR_CONNECTION);
        }
    }

    ws.onclose = function () {
        if (isLoggedIn || isRoomReconnectInProgress) {
            ws.onerror = function () {};
            ws.onclose = function () {};
            reconnectToRoom();
        } else {
            redirectToHomePageWithError(ROOM_TO_HOME_PG_REDIRECT_ERROR_CONNECTION);
        }
    }

    ws.onmessage = function (e) {
        let message = JSON.parse(e.data);

        switch (message.c) {
            //business error
            case COMMANDS.Error:
                processErrorCommand(message, alternativeRoomNamePostfixes);
                break;

            //request processed
            case COMMANDS.RequestProcessed:
                processRequestProcessedCommand(message);
                break;

            case COMMANDS.RoomMembersChanged:
                processRoomMembersChangedCommand(message);
                break;

            case COMMANDS.AllTextMessages:
                processAllTextMessagesCommand(message);
                break;

            case COMMANDS.TextMessage:
                processTextMessageCommand(message);
                break;

            case COMMANDS.TextMessageEdit:
                processTextMessageEditCommand(message);
                break;

            case COMMANDS.TextMessageDelete:
                processTextMessageDeleteCommand(message);
                break;

            case COMMANDS.UserDrawingMessage:
                processUserDrawingMessageCommand(message);
                break;

            case COMMANDS.TextMessageSupportOrReject:
                processSupportOrRejectCommand(message);
                break;

            case COMMANDS.RoomChangeDescription:
                processRoomDescriptionChangedCommand(message);
                break;

            case COMMANDS.NotifyMessagesLimitApproaching:
            case COMMANDS.NotifyMessagesLimitReached:
                processRoomNotificationCommand(message);
                break;
        }
    }
}

function pickBackend (roomName, callback) {
    $.ajax(PICK_BACKEND_ENDPOINT + roomName)
        .done(function (data) {
            if (data && data.e) {
                if (isRoomReconnectInProgress) {
                    reconnectToRoom();
                } else {
                    redirectToHomePageWithError(ROOM_TO_HOME_PG_REDIRECT_ERROR_CONNECTION_PICK_BACKEND, data.e);
                }

                return;
            }

            if (data && data.bA) {
                callback(roomName, data.bA, data.aN);
            }
        })
        .fail(function () {
            if (isRoomReconnectInProgress) {
                reconnectToRoom();
            } else {
                redirectToHomePageWithError(ROOM_TO_HOME_PG_REDIRECT_ERROR_CONNECTION_PICK_BACKEND);
            }
        });
}

function redirectToHomePageWithError (error, errorDetails, errorBusiness, alternativeRoomNamePostfixes) {
    shutdownSocket();

    LOCAL_STORAGE.setItem(REDIRECT_VARIABLE_LOCAL_STORAGE_KEY, JSON.stringify({
        redirectFrom: REDIRECT_FROM_ROOM_PAGE_ON_ERROR,
        redirectedAt: new Date().getTime(),
        error: error,
        errorDetails: errorDetails,
        errorBusiness: errorBusiness,
        roomName: ROOM_NAME,
        alternativeRoomNamePostfixes: alternativeRoomNamePostfixes,
    }));

    redirectToURL("/");
}


/* Actions */

function sendUserMessage () {
    if (!isLoggedIn) {
        return false;
    }

    const messageTxt = $userMessageTextarea.val();
    const replyToUserId = userReplyInProgressToId;
    const replyToMessageId = messageReplyInProgressToId;

    const editMessageId = messageEditingInProgressId;

    if (unicodeStringLength(messageTxt) > MAX_TEXT_MESSAGE_LENGTH) {
        showError(BUSINESS_ERRORS[ERROR_CODE_MESSAGE_IS_TOO_LONG].text);

        return false;
    }

    const messageTextareaInLargeMode = $userMessageTextarea.isBig;

    //editing existing message
    if (editMessageId) {
        let nextUserActionIndicationIdVal = nextUserActionIndicationId++;

        const sent = sendData(ws,
            JSON.stringify({
                c: "TM_E",
                rq: "u_action_" + nextUserActionIndicationIdVal,
                r: {
                    n: currentRoomName
                },
                m: {
                    id: editMessageId,
                    t: encodeURI(messageTxt.trim()),
                    rU: replyToUserId || undefined,
                    rM: parseInt(replyToMessageId) || undefined,
                }
            })
        );

        if (sent) {
            currentlyExpectedToCompleteUserActionIndicationId = nextUserActionIndicationIdVal;
            $userMessageCollapseWrSendingIndication.removeClass('d-none');

            cancelMessageEdit();
            resizeWrappersHeight(false);

            const $roomMessage = findMessageBlockById(editMessageId);
            if ($roomMessage) {
                scrollToElement($roomMessage[0]);

                animateBlockForAttention($roomMessage, true, 700);
            }

            cancelReplyToUser();
            cancelReplyToMessage();
        }

        return sent;
    }

    //sending new message
    if (!messageTxt || messageTxt.trim().length === 0) {
        return false;
    }

    let nextUserActionIndicationIdVal = nextUserActionIndicationId++;

    const sent = sendData(ws,
        JSON.stringify({
            c: "TM",
            rq: "u_action_" + nextUserActionIndicationIdVal,
            r: {
                n: currentRoomName
            },
            m: {
                t: encodeURI(messageTxt.trim()),
                rU: replyToUserId || undefined,
                rM: parseInt(replyToMessageId) || undefined,
            }
        })
    );

    if (sent) {
        currentlyExpectedToCompleteUserActionIndicationId = nextUserActionIndicationIdVal;
        $userMessageCollapseWrSendingIndication.removeClass('d-none');

        $userMessageTextarea.val('');

        cancelReplyToUser();
        cancelReplyToMessage();
    }

    if (messageTextareaInLargeMode) {
        $userMessageTextarea.focus();
    }

    scrollChatBottom();

    return sent;
}

function editUserMessage (editMessageId, $messageBlock) {
    const isChatAtBottom = isChatScrolledBottom();

    LOCAL_STORAGE.setItem(LAST_UNSENT_MESSAGE_TEXT_LOCAL_STORAGE_KEY, '');

    cancelReplyToUser();
    cancelReplyToMessage();
    cancelMessageEdit();
    hideFolkPicksMobile();
    cancelTextSearch();

    messageEditingInProgressId = editMessageId;

    const $cancelMessageEditOverlay = $(
        '<div class="message-cancel-editing-overlay">' +
        '<span class="message-cancel-editing-overlay-text">cancel editing</span>' +
        '</div>'
    );

    $cancelMessageEditOverlay.on('mousedown', function (e) {
        stopPropagationAndDefault(e);

        const isChatAtBottom = isChatScrolledBottom();

        cancelMessageEdit();
        resizeWrappersHeight(isChatAtBottom);
    });

    const $roomMessage = findMessageBlockById(editMessageId);
    if ($roomMessage) {
        $roomMessage.append($cancelMessageEditOverlay);
    }

    const $folkPicksMessage = findFolkPicksMessageBlockById(editMessageId);
    if ($folkPicksMessage) {
        $folkPicksMessage.append($cancelMessageEditOverlay.clone(true, true));
    }

    //if message has reply to message or user - maintain
    const replyToMessageId = parseInt($messageBlock.attr('data-reply-to-msg-id'));
    const replyToUserId = $messageBlock.attr('data-reply-to-user-id');

    if (replyToMessageId) {
        let originalMessageShortText = MESSAGE_UNAVAILABLE_PLACEHOLDER_TEXT;

        const $originalMessageBlock = findMessageBlockById(replyToMessageId);
        if ($originalMessageBlock) {
            const $innerText = $originalMessageBlock.find('.room-msg-text-inner');

            if ($innerText.length) {
                originalMessageShortText = cutMessageTextForResponse($innerText.text());
            }
        }

        const $messageUserAnonBlock = $messageBlock.find('.room-msg-name-anon');
        const messageAuthorName = $messageBlock.find('.room-msg-name-val').text();

        startReplyToMessage(replyToMessageId, originalMessageShortText, replyToUserId, messageAuthorName, !!$messageUserAnonBlock.length);

    } else if (replyToUserId) {
        const $userAnonBlock = $messageBlock.find('.room-msg-name-anon');
        const userNameFromMessage = $messageBlock.find('.room-msg-name-val').text();

        startReplyToUser(replyToUserId, userNameFromMessage, !!$userAnonBlock.length);
    }

    $messageEditCancelButton.removeClass('d-none');
    //hide user name info to cleanly occupy same space
    $userJoinedAsNameWr.addClass('d-none');
    $userJoinedAsTitle.addClass('d-none');

    let messageCleanText;

    //if any link has preview loaded - then swap outer link with inner to get its text clearly
    if ($messageBlock.find('.message-link-preview').length) {
        const $messageTextBlockClone = $messageBlock.find('.room-msg-text-inner').clone();
        const $childLinkWithPreview = $messageTextBlockClone.find('.message-link-preview');

        const $innerLink = $childLinkWithPreview.find('.message-highlight-link');
        $childLinkWithPreview.replaceWith($innerLink);

        messageCleanText = $messageTextBlockClone.text();

    } else {
        messageCleanText = $messageBlock.find('.room-msg-text-inner').text();
    }

    $userMessageTextarea.val(messageCleanText);

    $sendMessageButton.text('edit');

    //show user message block if it is minimized
    if ($userMessageContentWr.css('display') === 'none') {
        toggleUserMessageBlock();
    }

    resizeWrappersHeight(isChatAtBottom);

    scrollToElement($messageBlock[0]);

    $userMessageTextarea.focus();
}

function cancelMessageEdit () {
    if (messageEditingInProgressId) {
        const $roomMessage = findMessageBlockById(messageEditingInProgressId);
        if ($roomMessage) {
            $roomMessage.find('.message-cancel-editing-overlay').remove();
        }
        const $folkPicksMessage = findFolkPicksMessageBlockById(messageEditingInProgressId);
        if ($folkPicksMessage) {
            $folkPicksMessage.find('.message-cancel-editing-overlay').remove();
        }

        messageEditingInProgressId = null;

        $messageEditCancelButton.addClass('d-none');
        //bring back user name info blocks
        $userJoinedAsNameWr.removeClass('d-none');
        $userJoinedAsTitle.removeClass('d-none');

        $userMessageTextarea.val('');
        $sendMessageButton.text('send');

        cancelReplyToUser();
        cancelReplyToMessage();
    }
}

function deleteUserMessage (deleteMessageId) {
    let nextUserActionIndicationIdVal = nextUserActionIndicationId++;

    const sent = sendData(ws,
        JSON.stringify({
            c: "TM_D",
            rq: "u_action_" + nextUserActionIndicationIdVal,
            r: {
                n: currentRoomName
            },
            m: {
                id: deleteMessageId
            }
        })
    );

    if (sent) {
        currentlyExpectedToCompleteUserActionIndicationId = nextUserActionIndicationIdVal;
        $userMessageCollapseWrSendingIndication.removeClass('d-none');
    }
}

function supportRejectMsg (messageId, isSupport) {
    let nextUserActionIndicationIdVal = nextUserActionIndicationId++;

    const sent = sendData(ws,
        JSON.stringify({
            c: "TM_S_R",
            rq: "u_action_" + nextUserActionIndicationIdVal,
            r: {
                n: currentRoomName
            },
            srM: isSupport,
            m: {
                id: messageId
            }
        })
    );

    if (sent) {
        currentlyExpectedToCompleteUserActionIndicationId = nextUserActionIndicationIdVal;
        $userMessageCollapseWrSendingIndication.removeClass('d-none');
    }
}

function changeRoomDescription () {
    const newRoomDescr = $roomInfoDescriptionCreatorChangeInput.val().trim();

    if (newRoomDescr.toLowerCase() === $roomInfoDescriptionText.text().toLowerCase()) {
        return;
    }

    if (unicodeStringLength(newRoomDescr) > MAX_ROOM_DESCRIPTION_LENGTH) {
        showError(BUSINESS_ERRORS[ERROR_CODE_ROOM_INVALID_DESCRIPTION_LENGTH].text);
        return;
    }

    let nextUserActionIndicationIdVal = nextUserActionIndicationId++;

    const sent = sendData(ws,
        JSON.stringify({
            c: "R_CH_D",
            rq: "u_action_" + nextUserActionIndicationIdVal,
            r: {
                n: currentRoomName
            },
            m: {
                t: encodeURI(newRoomDescr)
            }
        })
    );

    if (sent) {
        currentlyExpectedToCompleteUserActionIndicationId = nextUserActionIndicationIdVal;
        $userMessageCollapseWrSendingIndication.removeClass('d-none');

        $roomInfoDescriptionCreatorChangeInput.val("");
    }
}

function changeUserName () {
    const newUserName = $userJoinedAsChangeInput.val().trim();

    if (
        !newUserName ||
        newUserName.toLowerCase() === UNKNOWN_USER_NAME.toLowerCase() ||
        (currentUserInfo && newUserName.toLowerCase() === decodeURIComponent(currentUserInfo.name).toLowerCase())
    ) {
        return;
    }

    if (unicodeStringLength(newUserName) < USER_NAME_MIN_LENGTH || unicodeStringLength(newUserName) > USER_NAME_MAX_LENGTH) {
        showError(BUSINESS_ERRORS[ERROR_CODE_INVALID_USER_NAME_LENGTH].text
            + ", must be between " + USER_NAME_MIN_LENGTH + " and " + USER_NAME_MAX_LENGTH + " characters");
        return;
    }

    let nextUserActionIndicationIdVal = nextUserActionIndicationId++;

    const sent = sendData(ws,
        JSON.stringify({
            c: "R_CH_UN",
            rq: "u_action_" + nextUserActionIndicationIdVal,
            uN: encodeURI(newUserName),
            r: {
                n: currentRoomName
            },
        })
    );

    if (sent) {
        currentlyExpectedToCompleteUserActionIndicationId = nextUserActionIndicationIdVal;
        $userMessageCollapseWrSendingIndication.removeClass('d-none');
    }
}

function sendUserDrawingMessage (fileName, fileGroupPrefix) {
    let nextUserActionIndicationIdVal = nextUserActionIndicationId++;

    const sent = sendData(ws,
        JSON.stringify({
            c: "DM",
            rq: "u_action_" + nextUserActionIndicationIdVal,
            r: {
                n: currentRoomName
            },
            m: {
                t: encodeURI(MESSAGE_META_MARKER_TYPE_DRAWING + fileName + "@" + fileGroupPrefix),
            }
        })
    );

    if (sent) {
        currentlyExpectedToCompleteUserActionIndicationId = nextUserActionIndicationIdVal;
        $userMessageCollapseWrSendingIndication.removeClass('d-none');
    }

    scrollChatBottom();
}


/* Incoming commands processing */

function processRoomDescriptionChangedCommand (message) {
    if (!isLoggedIn) {
        //if room join process is not complete yet - wait for it to complete and after that - process current message
        commandsToProcessChangeRoomDescription.push(message);

        return;
    }

    const newChangeRoomDescriptionCreatedAt = message.cAt;

    //skip if command is already outdated
    if (lastChangeRoomDescription && newChangeRoomDescriptionCreatedAt <= lastChangeRoomDescription) {
        return;
    }

    lastChangeRoomDescription = message.cAt;

    roomCreatorUserInRoomUUID = message.rCuId;

    //for room's creator user and self user - add visual indication to user's messages and nameplates
    for (let msgId in roomMessageIdToDOMElem) {
        const $messageBlock = roomMessageIdToDOMElem[msgId];
        const messageAuthorUserId = $messageBlock.attr('data-user-id');

        if (roomCreatorUserInRoomUUID === messageAuthorUserId) {
            $messageBlock.addClass('room-msg-room-creator');

            $messageBlock.find('.message-marks-wr')
                .append($('<span>adm</span>'));

            const $messageInPicksBlock = findFolkPicksMessageBlockById(msgId);

            if ($messageInPicksBlock) {
                $messageInPicksBlock.addClass('room-msg-room-creator');

                $messageInPicksBlock.find('.message-marks-wr')
                    .append($('<span>adm</span>'));
            }
        }

        if (userInRoomUUID === messageAuthorUserId) {
            $messageBlock.addClass('room-msg-author-self');

            if (roomCreatorUserInRoomUUID !== userInRoomUUID) {
                $messageBlock.find('.room-msg-buttons').addClass('d-none');
            }

            const $messageInPicksBlock = findFolkPicksMessageBlockById(msgId);

            if ($messageInPicksBlock) {
                $messageInPicksBlock.addClass('room-msg-author-self');
            }
        }
    }

    $roomInfoOnlineUsers.find(".room-info-online-user[data-user-id='" + roomCreatorUserInRoomUUID + "']")
        .addClass('room-info-user-is-room-creator');

    $roomInfoOnlineUsers.find(".room-info-online-user[data-user-id='" + userInRoomUUID + "']")
        .addClass('room-info-user-self');

    const newRoomDescription = message.m[0].t;

    $roomInfoDescriptionText.text(decodeURIComponent(newRoomDescription));
    $roomInfoDescriptionCreatorChangeInput.val(decodeURIComponent(newRoomDescription));

    //if new description is not empty - show it, else - show placeholder
    if (newRoomDescription) {
        $roomInfoDescriptionText.removeClass('d-none');
        $roomInfoDescriptionEmptyWr.addClass('d-none');

    } else {
        $roomInfoDescriptionText.addClass('d-none');
        $roomInfoDescriptionEmptyWr.removeClass('d-none');
    }

    //if current user is a creator of room - show room description changing interface
    if (userInRoomUUID && userInRoomUUID === roomCreatorUserInRoomUUID) {
        if (!$roomInfoDescriptionEmptyText.hasClass('d-none')) {
            $roomInfoDescriptionEmptyText.addClass('d-none');
            $roomInfoDescriptionEmptyTextCreator.removeClass('d-none');
        }

        //1st time we got room-descr-change after user is fully logged in into room and user id is known - check if user is the room's creator and if so - setup descr changing interface
        if (!$roomInfoDescription.hasClass('room-descr-is-creator')) {
            $roomInfoDescription.addClass('room-descr-is-creator');

            $roomInfoDescription.on('click', showRoomDescriptionEditingBlock);

            $roomInfoDescriptionCreatorChangeLink.on('click', onChangeRoomDescription);
        }

    } else {
        if ($roomInfoDescriptionEmptyText.hasClass('d-none')) {
            $roomInfoDescriptionEmptyText.removeClass('d-none');
            $roomInfoDescriptionEmptyTextCreator.addClass('d-none');
        }
    }

    //notify if this isn't the 1st description load
    if (isRoomDescriptionLoaded) {
        showTopNotification("room description changed", TOP_NOTIFICATION_SHOW_MS * 2, true);
    }

    isRoomDescriptionLoaded = true;
}

function processRoomMembersChangedCommand(message) {
    const newUserListTimestamp = message.cAt;
    const newUsersList = message.rU;
    const newUserInfoByIdCache = {};

    if (lastUserListTimestamp && newUserListTimestamp <= lastUserListTimestamp) {
        return;
    }

    lastUserListTimestamp = newUserListTimestamp;

    const isChatAtBottom = isChatScrolledBottom();

    //prepare new version of user info-by-id cache
    for (let i = 0; i < newUsersList.length; i++) {
        const user = newUsersList[i];

        newUserInfoByIdCache[user.uId] = {id: user.uId, name: user.n, isAnon: user.an, isOnlineInRoom: user.o};
    }

    const newUserOnlineNameplateBlocksArr = [];

    //if new users list is different from old one - re-draw online users list and do other stuff
    if (isUserInfoByIdCacheUpdated(newUserInfoByIdCache, lastUserInfoByIdCache)) {
        $roomInfoOnlineUsers.empty();

        for (let userId in newUserInfoByIdCache) {
            const userCachedInfo = newUserInfoByIdCache[userId];
            const userName = decodeURIComponent(userCachedInfo.name);

            //drawing UI block for user if currently online in room
            if (userCachedInfo.isOnlineInRoom) {
                const $roomOnlineUserBlock = $('<div class="room-info-online-user">');
                $roomOnlineUserBlock.attr("data-user-id", userId);

                if (roomCreatorUserInRoomUUID && roomCreatorUserInRoomUUID === userId) {
                    $roomOnlineUserBlock.addClass('room-info-user-is-room-creator');
                }

                if (userInRoomUUID && userInRoomUUID === userId) {
                    $roomOnlineUserBlock.addClass('room-info-user-self');
                }

                const $anonBlock = $('<span class="room-info-online-user-anon-pref">anon&nbsp;</span>');
                if (!userCachedInfo.isAnon) {
                    $anonBlock.addClass('d-none');
                }

                $roomOnlineUserBlock.append($anonBlock);

                const $roomOnlineUserNameBlock = $('<span class="room-info-online-user-name">');
                $roomOnlineUserNameBlock.text(userName);

                $roomOnlineUserBlock.append($roomOnlineUserNameBlock);

                $roomOnlineUserBlock.on('click', function (e) {
                    const $userBlock = $(e.currentTarget).closest('.room-info-online-user');

                    const messageAuthorUserId = $userBlock.attr("data-user-id");
                    const $anonBlock = $userBlock.find('.room-info-online-user-anon-pref');
                    const messageAuthorName = $userBlock.find('.room-info-online-user-name').text();

                    startReplyToUser(messageAuthorUserId, messageAuthorName, !!$anonBlock.length);
                });

                if (userId === userInRoomUUID) {
                    $roomOnlineUserBlock.addClass('room-info-online-user-self');
                }

                newUserOnlineNameplateBlocksArr.push($roomOnlineUserBlock);
            }


            //check if existing user name changed. If so - update any messages related to this user with new name
            const existingUserInfo = lastUserInfoByIdCache[userId];
            const oldUserName = existingUserInfo ? decodeURIComponent(existingUserInfo.name) : null;
            const newUserName = decodeURIComponent(userCachedInfo.name);

            if (oldUserName !== newUserName) {

                if (existingUserInfo) {
                    showTopNotification(
                        "'" +
                        (existingUserInfo.isAnon ? "anon " : "") +
                        oldUserName +
                        "' changed name to '" +
                        (userCachedInfo.isAnon ? "anon " : "") +
                        newUserName +
                        "'",
                        TOP_NOTIFICATION_SHOW_MS * 2,
                        true
                    );
                }

                //update fast message search info storage
                for (let messageSearchInfoId in messageIdToTextSearchInfo) {
                    const messageSearchInfo = messageIdToTextSearchInfo[messageSearchInfoId];

                    if (userId === messageSearchInfo.authorUserId) {
                        messageSearchInfo.authorName = userName;
                    }

                    if (userId === messageSearchInfo.replyToUserId) {
                        messageSearchInfo.replyToUserName = userName;
                    }
                }

                //if operating user's name changed - change 'logged as' nameplate
                if (userCachedInfo.id === userInRoomUUID) {
                    if (userCachedInfo.isAnon) {
                        $userJoinedAsAnonPref.removeClass('d-none');
                    } else {
                        $userJoinedAsAnonPref.addClass('d-none');
                    }

                    $userJoinedAsName.text(userName);
                    $userJoinedAsChangeInput.val(userName);

                    $roomWelcomeMessageUsername.find('.user-message-joined-as-block').remove();
                    $roomWelcomeMessageUsername.append(
                        $userJoinedAsNameWr.clone()
                    );

                    currentUserInfo = userCachedInfo;
                }

                //change user name for all user's messages
                $roomMessagesWr.find(".room-msg-main-wr[data-user-id='" + userId + "']").each(function () {
                    const $roomMessage = $(this);
                    $roomMessage.find(".room-msg-name-val")
                        .text(userName);

                    if (userCachedInfo.isAnon) {
                        $roomMessage.find('.room-msg-name-anon').removeClass('d-none');
                    } else {
                        $roomMessage.find('.room-msg-name-anon').addClass('d-none');
                    }
                });

                $folkPicksMessagesWr.find(".room-msg-main-wr[data-user-id='" + userId + "']").each(function () {
                    const $messageInFolkPicks = $(this);
                    $messageInFolkPicks.find(".room-msg-name-val")
                        .text(userName);

                    if (userCachedInfo.isAnon) {
                        $messageInFolkPicks.find('.room-msg-name-anon').removeClass('d-none');
                    } else {
                        $messageInFolkPicks.find('.room-msg-name-anon').addClass('d-none');
                    }
                });


                //change user name for messages with reply-to-user that points to this user
                $roomMessagesWr.find(".message-reply-to-user[data-user-id='" + userId + "']").each(function () {
                    const $replyToUserIdBlockInMessage = $(this);
                    $replyToUserIdBlockInMessage.find(".message-reply-to-user-name")
                        .text(userName + ',');

                    if (userCachedInfo.isAnon) {
                        $replyToUserIdBlockInMessage.find('.message-reply-to-user-anon-pref').removeClass('d-none');
                    } else {
                        $replyToUserIdBlockInMessage.find('.message-reply-to-user-anon-pref').addClass('d-none');
                    }
                });

                $folkPicksMessagesWr.find(".message-reply-to-user[data-user-id='" + userId + "']").each(function () {
                    const $replyToUserIdBlockInFolkPicksMessage = $(this);
                    $replyToUserIdBlockInFolkPicksMessage.find(".message-reply-to-user-name")
                        .text(userName + ',');

                    if (userCachedInfo.isAnon) {
                        $replyToUserIdBlockInFolkPicksMessage.find('.message-reply-to-user-anon-pref').removeClass('d-none');
                    } else {
                        $replyToUserIdBlockInFolkPicksMessage.find('.message-reply-to-user-anon-pref').addClass('d-none');
                    }
                });

                if ((oldUserName && oldUserName.toLowerCase().includes(currentTextSearchContext.text)) ||
                    (newUserName && newUserName.toLowerCase().includes(currentTextSearchContext.text))) {
                    textSearchDataChangedAt = new Date().getTime();
                }
            }


            //check if there are any messages with unknown author user name, that points to this user id
            const thisUserMessagesWithUnknownAuthor = userIdToMessageIdsWithUnknownAuthor[userId];

            if (thisUserMessagesWithUnknownAuthor) {
                for (let i = 0; i < thisUserMessagesWithUnknownAuthor.length; i++) {
                    const msgId = thisUserMessagesWithUnknownAuthor[i];

                    //find / set author name for messages
                    const $roomMessage = findMessageBlockById(msgId);
                    if ($roomMessage) {
                        const $messageNameValBlock = $roomMessage.find(".room-msg-name-val");
                        $messageNameValBlock.text(userName);

                        if (userCachedInfo.isAnon) {
                            $roomMessage.find('.room-msg-name-anon').removeClass('d-none');
                        } else {
                            $roomMessage.find('.room-msg-name-anon').addClass('d-none');
                        }
                    }

                    //find / set author name for folk picks
                    const $messageInFolkPicks = findFolkPicksMessageBlockById(msgId);
                    if ($messageInFolkPicks) {
                        const $picksMessageNameValBlock = $messageInFolkPicks.find(".room-msg-name-val");
                        $picksMessageNameValBlock.text(userName);

                        if (userCachedInfo.isAnon) {
                            $messageInFolkPicks.find('.room-msg-name-anon').removeClass('d-none');
                        } else {
                            $messageInFolkPicks.find('.room-msg-name-anon').addClass('d-none');
                        }
                    }

                    //update fast message search info storage
                    messageIdToTextSearchInfo[msgId].authorName = userName;
                }

                delete userIdToMessageIdsWithUnknownAuthor[userId];
            }

            //same - check if there are any messages with unknown reply-to-user, that points to this user id
            const messagesWithUnknownUserNameReplyTo = userIdToMessageIdsWithUnknownReplyToUserName[userId];

            if (messagesWithUnknownUserNameReplyTo) {
                for (let i = 0; i < messagesWithUnknownUserNameReplyTo.length; i++) {
                    const msgWithReplyToId = messagesWithUnknownUserNameReplyTo[i];

                    //find / set reply-to-user-name for messages
                    const $roomMessageWithReplyTo = findMessageBlockById(msgWithReplyToId);
                    if ($roomMessageWithReplyTo) {
                        $roomMessageWithReplyTo.find(".message-reply-to-user-name")
                            .text(userName + ',');

                        if (userCachedInfo.isAnon) {
                            $roomMessageWithReplyTo.find('.message-reply-to-user-anon-pref').removeClass('d-none');
                        } else {
                            $roomMessageWithReplyTo.find('.message-reply-to-user-anon-pref').addClass('d-none');
                        }
                    }

                    //find / set reply-to-user-name for folk picks
                    const $messageInFolkPicksWithReplyTo = findFolkPicksMessageBlockById(msgWithReplyToId);
                    if ($messageInFolkPicksWithReplyTo) {
                        $messageInFolkPicksWithReplyTo.find(".message-reply-to-user-name")
                            .text(userName + ',');

                        if (userCachedInfo.isAnon) {
                            $messageInFolkPicksWithReplyTo.find('.message-reply-to-user-anon-pref').removeClass('d-none');
                        } else {
                            $messageInFolkPicksWithReplyTo.find('.message-reply-to-user-anon-pref').addClass('d-none');
                        }
                    }

                    //update fast message search info storage
                    messageIdToTextSearchInfo[msgWithReplyToId].replyToUserName = userName;
                }

                delete userIdToMessageIdsWithUnknownReplyToUserName[userId];
            }
        }

        lastUserInfoByIdCache = newUserInfoByIdCache;

        //sort all newly-created user nameplates by name and append to DOM
        newUserOnlineNameplateBlocksArr.sort(function (a, b) {
            const userNameA = a.find('.room-info-online-user-name').text();
            const userNameB = b.find('.room-info-online-user-name').text();

            return userNameA.localeCompare(userNameB);
        });

        $roomInfoOnlineUsers.append(...newUserOnlineNameplateBlocksArr);

        //set users count text
        $roomInfoUsersCount.text(getOnlineUsersCountString());

        resizeWrappersHeight(isChatAtBottom);
    }
}

function processSupportOrRejectCommand(message) {
    const newVotedTimestamp = message.cAt;
    const voteMessage = message.m[0];
    const messageId = voteMessage.id;

    const $roomMessage = findMessageBlockById(messageId);

    if ($roomMessage) {
        const lastVotedTimestamp = $roomMessage.attr('data-voted-at');
        if (lastVotedTimestamp && parseInt(newVotedTimestamp) <= parseInt(lastVotedTimestamp)) {
            return;
        }

        $roomMessage.attr('data-voted-at', parseInt(newVotedTimestamp));

        $roomMessage.find('.room-msg-buttons-support-val').text(voteMessage.sC);
        $roomMessage.find('.room-msg-buttons-reject-val').text(voteMessage.rC);

        let $messageInFolkPicksScreen = findFolkPicksMessageBlockById(messageId);

        //if this message is already displayed on folk-picks screen - update values
        if ($messageInFolkPicksScreen) {
            if (voteMessage.sC === 0 && voteMessage.rC === 0) {
                $messageInFolkPicksScreen.remove();
                delete folkPicksMessageIdToDOMElem[messageId];

                if (!anyMessageIsFolkPicked()) {
                    $folkPicksMessagesWr.append($folkPicksEmptyMessagesWr);
                }

                return;
            }

            $messageInFolkPicksScreen.find('.room-msg-buttons-support-val').text(voteMessage.sC);
            $messageInFolkPicksScreen.find('.room-msg-buttons-reject-val').text(voteMessage.rC);

        } else if (voteMessage.sC !== 0 || voteMessage.rC !== 0) {
            //if not yet - copy message to folk-picks and update values

            //check if need to sort folk-picks after inserting new message
            const $lastFolkPicksMessage = $folkPicksMessagesWr.children().last();

            let needToSortFolkPicks = false;

            if ($lastFolkPicksMessage.length) {
                const lastFolkPicksMessageId = parseInt($lastFolkPicksMessage.attr('data-msg-id'));

                //initially 1st message in block is a welcome block without any id
                if (!isNaN(lastFolkPicksMessageId) && lastFolkPicksMessageId > messageId) {
                    needToSortFolkPicks = true;
                }
            }

            //copy message to folk picks
            $messageInFolkPicksScreen = $roomMessage.clone(true, true);

            //if this is 1st voted message
            if ($folkPicksEmptyMessagesWr.css('display') === 'block') {
                $folkPicksEmptyMessagesWr.remove();
            }

            //cancel text selection mode for message block copy
            if (messageTextSelectionInProgressId) {
                attachLongTouchContextMenuToElement($messageInFolkPicksScreen);

                $messageInFolkPicksScreen.removeClass('text-selection-active');

                $messageInFolkPicksScreen.addClass('noselect');

                $messageInFolkPicksScreen.css('opacity', '');
            }
            //remove search highlights from message block copy
            if (isTextSearchInProgress()) {
                cancelTextSearchHighlightsForBlock($messageInFolkPicksScreen);
            }

            if (messageTextSelectionInProgressId) {
                cancelMessageTextSelectionMode();
            }

            $folkPicksMessagesWr.append($messageInFolkPicksScreen);

            folkPicksMessageIdToDOMElem[messageId] = $messageInFolkPicksScreen;

            if (needToSortFolkPicks) {
                sortFolkPics();
            }
        }

        if (voteMessage.sC >= voteMessage.rC) {
            $messageInFolkPicksScreen.removeClass('folk-rejected');
            $messageInFolkPicksScreen.addClass('folk-supported');
        } else {
            $messageInFolkPicksScreen.removeClass('folk-supported');
            $messageInFolkPicksScreen.addClass('folk-rejected');
        }
    } else {
        //if target text message is not received yet - add this vote message to queue and process once target message arrives
        if (!deletedMessageIds[messageId]) {
            if (!commandsToProcessVoteMessage[messageId]) {
                commandsToProcessVoteMessage[messageId] = [];
            }
            commandsToProcessVoteMessage[messageId].push(message);
        }
    }
}

function processUserDrawingMessageCommand (message) {
    const chatScrolledBottom = isChatScrolledBottom();

    const drawingMessage = message.m[0];
    const messageId = drawingMessage.id;
    const messageText = decodeURIComponent(drawingMessage.t);
    const needToSortInfo = processTextMessage(drawingMessage, null, null, true, chatScrolledBottom);

    const textMeta = messageText.split(MESSAGE_META_MARKER_TYPE_DRAWING)[1];
    const fileName = textMeta.split("@")[0];
    const fileGroupName = textMeta.split("@")[1];

    setTimeout(function () {
        getUserDrawingFromFileServer(
            fileName,
            fileGroupName,
            function (data) {
                displayUserDrawingMessage(data, messageId, chatScrolledBottom);
            },
            function () {
                setTimeout(function () {
                        getUserDrawingFromFileServer(
                            fileName,
                            fileGroupName,
                            function (data) {
                                displayUserDrawingMessage(data, messageId, chatScrolledBottom);
                            },
                            function () {
                                console.log('failed to load user drawing after 2nd attempt');
                            });
                    },
                    3000
                );
            }
        );
    }, randomIntBetween(1, 200));

    if (needToSortInfo.needToSortMessages) {
        sortMainMessages();
    }
    if (needToSortInfo.needToSortFolkPicks) {
        sortFolkPics();
    }
}

function displayUserDrawingMessage (data, messageId, needToScrollChatToBottom) {
    const $messageBlock = findMessageBlockById(messageId);
    const $messageTextBlock = $messageBlock.find('.room-msg-text-inner');

    $messageBlock.attr('data-meta-marker', MESSAGE_META_MARKER_TYPE_DRAWING);

    const $drawingMessagePreviewBlock = $('<img src="' + data + '" alt="user drawing" class="user-drawing-message-preview" />');

    $drawingMessagePreviewBlock.on('click', function (e) {
        stopPropagationAndDefault(e);

        showGlobalTransparentOverlay();

        showDrawingFullSizeView($(this).attr('src'));
    });

    if (needToScrollChatToBottom) {
        $drawingMessagePreviewBlock.on('load', scrollChatBottom);
    }

    //clear message text
    $messageTextBlock.text('');
    $messageTextBlock.append($drawingMessagePreviewBlock);

    const $textMessageInFolkPicks = findFolkPicksMessageBlockById(messageId);

    if ($textMessageInFolkPicks) {
        $textMessageInFolkPicks.attr('data-meta-marker', MESSAGE_META_MARKER_TYPE_DRAWING);

        const $folPicksMessageTextBlock = $textMessageInFolkPicks.find('.room-msg-text-inner');
        $folPicksMessageTextBlock.text('');
        $folPicksMessageTextBlock.append($drawingMessagePreviewBlock.clone(true, true));
    }
}

function processTextMessageCommand (message) {
    const chatScrolledBottom = isChatScrolledBottom();
    const textMessage = message.m[0];
    const messageAuthorUserId = textMessage.uId;
    const text = decodeURIComponent(textMessage.t);

    const needToSortInfo = processTextMessage(textMessage, null, null, true, chatScrolledBottom);

    if (needToSortInfo.needToSortMessages) {
        sortMainMessages();
    }
    if (needToSortInfo.needToSortFolkPicks) {
        sortFolkPics();
    }

    if (chatScrolledBottom) {
        scrollChatBottom();
    }

    if (text.toLowerCase().includes(currentTextSearchContext.text)) {
        textSearchDataChangedAt = new Date().getTime();
    }

    //execute bot actions, if any are set up by user
    const messageAuthorUserInfo = lastUserInfoByIdCache[messageAuthorUserId];
    let userName;

    if (messageAuthorUserInfo) {
        userName = decodeURIComponent(messageAuthorUserInfo.name);
    }

    if (!text.startsWith("[BOT]: ")) {
        for (let botId in clientBotsCluster) {
            if (clientBotsCluster[botId].isEnabled) {
                clientBotsCluster[botId].onMessage(botId, {text: text, authorName: userName});
            }
        }
    }
}

function processAllTextMessagesCommand (message) {
    //if this page was re-loaded and there are some old messages to delete - do it
    $roomMessagesWr.find(".room-msg-main-wr." + CLASS_MESSAGE_TO_CLEAN_UP).each(function () {
        $(this).remove();
    });
    $folkPicksMessagesWr.find(".room-msg-main-wr." + CLASS_MESSAGE_TO_CLEAN_UP).each(function () {
        $(this).remove();
    });

    const allTextMessages = message.m;
    const roomMessagesCopedAt = message.cAt;

    let needToSortMessages = false;
    let needToSortFolkPicks = false;

    for (let i = 0; i < allTextMessages.length; i++) {
        const message = allTextMessages[i];
        const messageId = message.id;
        const messageText = decodeURIComponent(message.t);

        //for 1st message - check existing messages order in case some messages got here before all-messages command
        if (i === 0) {
            const needToSortInfoAtInit = processTextMessage(message, roomMessagesCopedAt, allTextMessages, true, true);

            needToSortMessages = needToSortInfoAtInit.needToSortMessages;
            needToSortFolkPicks = needToSortInfoAtInit.needToSortFolkPicks;
        } else {
            const needToSortInfo = processTextMessage(message, roomMessagesCopedAt, allTextMessages, false, true);

            needToSortFolkPicks = needToSortInfo.needToSortFolkPicks;
        }

        //if this is a user drawing message - apply additional processing
        if (messageText.startsWith(MESSAGE_META_MARKER_TYPE_DRAWING)) {
            const textMeta = messageText.split(MESSAGE_META_MARKER_TYPE_DRAWING)[1];
            const fileName = textMeta.split("@")[0];
            const fileGroupName = textMeta.split("@")[1];

            setTimeout(function () {
                getUserDrawingFromFileServer(
                    fileName,
                    fileGroupName,
                    function (data) {
                        displayUserDrawingMessage(data, messageId, true);
                    },
                    function () {
                        setTimeout(function () {
                                getUserDrawingFromFileServer(
                                    fileName,
                                    fileGroupName,
                                    function (data) {
                                        displayUserDrawingMessage(data, messageId, true);
                                    },
                                    function () {
                                        console.log('failed to load user drawing after 2nd attempt');
                                    });
                            },
                            3000
                        );
                    }
                );
            }, randomIntBetween(1, 100));
        }
    }

    if (needToSortMessages) {
        sortMainMessages();
    }
    if (needToSortFolkPicks) {
        sortFolkPics();
    }

    scrollChatBottom();

    if (!anyMessageIsFolkPicked()) {
        $folkPicksMessagesWr.append($folkPicksEmptyMessagesWr);
    }

    textSearchDataChangedAt = new Date().getTime();
}

function processTextMessage(textMessage, roomMessagesCopedAt, allTextMessages, needToCheckMessagesOrder, needToScrollChatToBottom) {
    const messageId = textMessage.id;
    const messageText = decodeURIComponent(textMessage.t);
    const messageAuthorUserId = textMessage.uId;
    const messageAuthorUserInfo = lastUserInfoByIdCache[messageAuthorUserId];
    let userName;

    //if user info of message author is not yet received - use 'unknown' as username and wait for info to come
    if (messageAuthorUserInfo) {
        userName = decodeURIComponent(messageAuthorUserInfo.name);
    } else {
        if (!userIdToMessageIdsWithUnknownAuthor[messageAuthorUserId]) {
            userIdToMessageIdsWithUnknownAuthor[messageAuthorUserId] = [];
        }

        userIdToMessageIdsWithUnknownAuthor[messageAuthorUserId]
            .push(messageId);

        userName = UNKNOWN_USER_NAME;
    }

    //check if last message has lower id than new message. If no - will need to sort messages after appending new one
    let needToSortMessages = false;

    if (needToCheckMessagesOrder) {
        const lastMessage = $roomMessagesWr.children().last();

        if (lastMessage.length) {
            const lastMessageId = parseInt(lastMessage.attr('data-msg-id'));

            //initially 1st message in block is a welcome block without any id
            if (!isNaN(lastMessageId) && lastMessageId > messageId) {
                needToSortMessages = true;
            }
        }
    }

    const originalMessageShortText = textMessage.rM
        ? getOriginalMessageShortText(textMessage.rM, allTextMessages)
        : null;

    //create DOM block for new message and append to main messages wrapper
    const $roomMessage = createTextMessageDom(
        textMessage,
        userName,
        messageAuthorUserInfo ? messageAuthorUserInfo.isAnon : true,
        originalMessageShortText
    );

    const $textMessageBlock = $roomMessage.find('.room-msg-text-inner');

    //for single character message with emoji symbol - enlarge text
    if (unicodeStringLength(messageText.trim()) < 2 && stringContainsEmoji(messageText)) {
        $textMessageBlock.addClass('room-msg-text-inner-enlarge');
    }

    //if this call comes from 'all messages' command - set roomMessagesCopedAt as message's lastEdited/lastVoted timestamp
    if (roomMessagesCopedAt) {
        $roomMessage.attr('data-edited-at', roomMessagesCopedAt);
        $roomMessage.attr('data-voted-at', roomMessagesCopedAt);
    }

    //set voting behaviour

    $roomMessage.find('.room-msg-buttons-support-text, .room-msg-buttons-support-text-short, .room-msg-buttons-support-val')
        .on('mousedown', function (e) {
            if (userInRoomUUID && $roomMessage.attr('data-user-id') === userInRoomUUID && userInRoomUUID !== roomCreatorUserInRoomUUID) {
                return;
            }

            const $messageBlock = $(e.currentTarget).closest('.room-msg-main-wr');

            supportRejectMsg(parseInt($messageBlock.attr('data-msg-id')), true);
        });

    $roomMessage.find('.room-msg-buttons-reject-text, .room-msg-buttons-reject-text-short, .room-msg-buttons-reject-val')
        .on('mousedown', function (e) {
            if (userInRoomUUID && $roomMessage.attr('data-user-id') === userInRoomUUID && userInRoomUUID !== roomCreatorUserInRoomUUID) {
                return;
            }

            const $messageBlock = $(e.currentTarget).closest('.room-msg-main-wr');

            supportRejectMsg(parseInt($messageBlock.attr('data-msg-id')), false);
        });

    /* Set replying behaviour */

    $roomMessage.find('.room-msg-name').on('click', onRoomMessageAuthorClick);
    $roomMessage.find('.room-msg-name-anon').on('click', onRoomMessageAuthorClick);

    $roomMessage.find('.room-msg-id').on('click', onRoomMessageIdClick);

    if (roomCreatorUserInRoomUUID && roomCreatorUserInRoomUUID === messageAuthorUserId) {
        $roomMessage.addClass('room-msg-room-creator');

        //if this message is from room admin - add message mark
        $roomMessage.find('.message-marks-wr')
            .append($('<span>adm</span>'));
    }

    if (userInRoomUUID && userInRoomUUID === messageAuthorUserId) {
        $roomMessage.addClass('room-msg-author-self');
        $roomMessage.addClass('room-msg-main-wr-self');

        if (roomCreatorUserInRoomUUID && roomCreatorUserInRoomUUID !== userInRoomUUID) {
            $roomMessage.find('.room-msg-buttons').addClass('d-none');
        }

        //track last user's message processed by server
        if (messageId > lastCurrentUserSentMessageId) {
            lastCurrentUserSentMessageId = messageId;
        }
    }

    //set scrolling to target message
    const $replyToMessageIdBlock = $roomMessage.find('.reply-msg-to-id');
    if ($replyToMessageIdBlock.length) {
        $replyToMessageIdBlock.on('mousedown', scrollToRepliedMessage);
    }

    //set right-click and long-touch behaviour
    attachLongTouchContextMenuToElement($roomMessage);

    $roomMessage.on('click', function (e) {
        const $clickedBlock = $(e.target);

        //normally if context menu is opened - there will be global overlay to cancel it. But for mobile folk picks
        //it is impossible to show global overlay as mobile folk picks block is already position absolute
        if (!$messageContextMenu.hasClass('d-none')) {
            $messageContextMenu.addClass('d-none');
        }

        const $messageBlock = $(e.currentTarget).closest('.room-msg-main-wr');
        if (parseInt($messageBlock.attr('data-msg-id')) !== messageTextSelectionInProgressId) {
            cancelMessageTextSelectionMode();
        }

        //if clicked on anything except link - prevent default
        if (!$clickedBlock.hasClass('message-highlight-link')) {
            stopPropagationAndDefault(e);
        }
    });

    //turn all links text occurrences from message into tags
    const allLinksInfos = findAllLinksInText(messageText);
    findLinksTextAndTurnIntoLinks(allLinksInfos, $textMessageBlock, messageId, needToScrollChatToBottom);

    const messageTimeText = $roomMessage.find('.room-msg-name-time-part').text();

    //save message text/author name for faster search
    messageIdToTextSearchInfo[messageId] = {
        authorUserId: messageAuthorUserId,
        messageTimeText: messageTimeText,
        authorName: userName,
        replyToUserId: textMessage.rU,
        replyToUserName: null,
        replyToMessageId: textMessage.rM,
        replyToMessageText: originalMessageShortText,
        text: messageText,
    };

    if (textMessage.rU) {
        let replyToUserName = UNKNOWN_USER_NAME;
        const replyToUserInfo = lastUserInfoByIdCache[textMessage.rU];

        if (replyToUserInfo) {
            replyToUserName = decodeURIComponent(replyToUserInfo.name);
        }

        messageIdToTextSearchInfo[messageId].replyToUserName = replyToUserName;
    }

    //append message block to DOM
    $roomMessagesWr.append($roomMessage);
    roomMessageIdToDOMElem[messageId] = $roomMessage;


    /* Folk pics */

    let needToSortFolkPicks = false;

    //if this message has non zero support/reject count - it needs to be displayed on folk-picks screen
    if (textMessage.sC > 0 || textMessage.rC > 0) {
        //check if need to sort folk-picks messages after adding new one
        const $lastFolkPicksMessage = $folkPicksMessagesWr.children().last();

        if ($lastFolkPicksMessage.length) {
            const lastFolkPicksMessageId = parseInt($lastFolkPicksMessage.attr('data-msg-id'));

            if (lastFolkPicksMessageId > messageId) {
                needToSortFolkPicks = true;
            }
        }

        //copy message to folk picks
        const $messagePicksWrBlock = $roomMessage.clone(true, true);

        if (textMessage.sC >= textMessage.rC) {
            $messagePicksWrBlock.addClass('folk-supported');
        } else {
            $messagePicksWrBlock.addClass('folk-rejected');
        }

        if ($folkPicksEmptyMessagesWr.css('display') === 'block') {
            $folkPicksEmptyMessagesWr.remove();
        }

        $folkPicksMessagesWr.append($messagePicksWrBlock);

        folkPicksMessageIdToDOMElem[messageId] = $messagePicksWrBlock;
    }

    /* Check if some command messages are waiting for this text message to arrive */

    const voteCommands = commandsToProcessVoteMessage[messageId];
    const editCommands = commandsToProcessEditMessage[messageId];
    const deleteCommand = commandsToProcessDeleteMessage[messageId];

    if (voteCommands) {
        for (let i = 0; i < voteCommands.length; i++) {
            processSupportOrRejectCommand(voteCommands[i]);
        }

        delete commandsToProcessVoteMessage[messageId];
    }

    if (editCommands) {
        for (let i = 0; i < editCommands.length; i++) {
            processTextMessageEditCommand(editCommands[i]);
        }

        delete commandsToProcessEditMessage[messageId];
    }

    if (deleteCommand) {
        processTextMessageDeleteCommand(deleteCommand);

        delete commandsToProcessDeleteMessage[messageId];
    }

    return {needToSortMessages: needToSortMessages, needToSortFolkPicks: needToSortFolkPicks};
}

function processTextMessageEditCommand (message) {
    const textMessage = message.m[0];
    const messageId = textMessage.id;
    const newEditedTimestamp = message.cAt;

    const isChatAtBottom = isChatScrolledBottom();

    const $roomMessage = findMessageBlockById(messageId);
    if ($roomMessage) {
        const newMessageText = decodeURIComponent(textMessage.t);

        //save message text/author name for faster search
        const messageSearchInfo = messageIdToTextSearchInfo[messageId];
        const oldMessageText = messageSearchInfo.text;

        messageSearchInfo.text = newMessageText;

        if (textMessage.rU !== messageSearchInfo.replyToUserId) {
            messageSearchInfo.replyToUserId = textMessage.rU;

            if (textMessage.rU) {
                let replyToUserName = UNKNOWN_USER_NAME;
                const replyToUserInfo = lastUserInfoByIdCache[textMessage.rU];

                if (replyToUserInfo) {
                    replyToUserName = decodeURIComponent(replyToUserInfo.name);
                }

                messageSearchInfo.replyToUserName = replyToUserName;
            } else {
                messageSearchInfo.replyToUserName = null;
            }
        }

        if (textMessage.rM !== messageSearchInfo.replyToMessageId) {
            messageSearchInfo.replyToMessageId = textMessage.rM;

            if (textMessage.rM) {
                messageSearchInfo.replyToMessageText = getOriginalMessageShortText(textMessage.rM, null);
            } else {
                messageSearchInfo.replyToMessageText = null;
            }
        }

        if (oldMessageText.toLowerCase().includes(currentTextSearchContext.text) ||
            newMessageText.toLowerCase().includes(currentTextSearchContext.text)) {
            textSearchDataChangedAt = new Date().getTime();
        }

        const lastEditedTimestamp = $roomMessage.attr('data-edited-at');
        if (lastEditedTimestamp && parseInt(newEditedTimestamp) <= parseInt(lastEditedTimestamp)) {
            return;
        }

        $roomMessage.attr('data-edited-at', parseInt(newEditedTimestamp));

        const $messageTextBlock = $roomMessage.find('.room-msg-text');
        $messageTextBlock.children().remove();

        if (!$roomMessage.find('.message-edited-title').length) {
            $roomMessage.find('.room-msg-name-time').append($('<span class="message-edited-title">(edited)</span>'));
        }


        /* If message has reply to message or user */

        if (textMessage.rM) {
            appendReplyToMessageBlock($messageTextBlock, textMessage.rM, textMessage.rU, textMessage.id,
                getOriginalMessageShortText(textMessage.rM, null));

            $roomMessage.attr('data-reply-to-user-id', textMessage.rU);
            $roomMessage.attr('data-reply-to-msg-id', textMessage.rM);

            //set scrolling to target message
            const $replyToMessageIdBlock = $roomMessage.find('.reply-msg-to-id');
            if ($replyToMessageIdBlock.length) {
                $replyToMessageIdBlock.on('click', scrollToRepliedMessage);
            }

        } else if (textMessage.rU) {
            $roomMessage.removeAttr('data-reply-to-msg-id');

            appendReplyToUserBlock($messageTextBlock, textMessage.rU, textMessage.id);
            $roomMessage.attr('data-reply-to-user-id', textMessage.rU);
        } else {
            $roomMessage.removeAttr('data-reply-to-msg-id');
            $roomMessage.removeAttr('data-reply-to-user-id');
        }

        const $messageTextInnerBlock = $('<span class="room-msg-text-inner">');
        $messageTextInnerBlock.text(newMessageText);

        $messageTextBlock.append($messageTextInnerBlock);

        //for single character message with emoji symbol - enlarge text
        if (unicodeStringLength(newMessageText.trim()) < 2 && stringContainsEmoji(newMessageText)) {
            $messageTextInnerBlock.addClass('room-msg-text-inner-enlarge');
        }

        //turn all links text occurrences from message into tags
        const allLinksInfos = findAllLinksInText(newMessageText);
        findLinksTextAndTurnIntoLinks(allLinksInfos, $messageTextInnerBlock, messageId, isChatAtBottom);

        const $folkPicksMessage = findFolkPicksMessageBlockById(messageId);

        if ($folkPicksMessage) {
            const $folkPicksMessageTextBlock = $folkPicksMessage.find('.room-msg-text');
            $folkPicksMessageTextBlock.children().remove();


            /* If message has reply to message or user */

            if (textMessage.rM) {
                appendReplyToMessageBlock($folkPicksMessageTextBlock, textMessage.rM, textMessage.rU, textMessage.id,
                    getOriginalMessageShortText(textMessage.rM, null));

                $folkPicksMessage.attr('data-reply-to-user-id', textMessage.rU);
                $folkPicksMessage.attr('data-reply-to-msg-id', textMessage.rM);

                //set scrolling to target message
                const $picksReplyToMessageIdBlock = $folkPicksMessage.find('.reply-msg-to-id');
                if ($picksReplyToMessageIdBlock.length) {
                    $picksReplyToMessageIdBlock.on('click', scrollToRepliedMessage);
                }

            } else if (textMessage.rU) {
                $folkPicksMessage.removeAttr('data-reply-to-msg-id');

                appendReplyToUserBlock($folkPicksMessageTextBlock, textMessage.rU, textMessage.id);
                $folkPicksMessage.attr('data-reply-to-user-id', textMessage.rU);
            } else {
                $folkPicksMessage.removeAttr('data-reply-to-msg-id');
                $folkPicksMessage.removeAttr('data-reply-to-user-id');
            }

            $folkPicksMessageTextBlock.append($messageTextInnerBlock.clone());
        }
    } else {
        //if target text message is not received yet - add this edit message to queue and process once target message arrives
        if (!deletedMessageIds[messageId]) {
            if (!commandsToProcessEditMessage[messageId]) {
                commandsToProcessEditMessage[messageId] = [];
            }
            commandsToProcessEditMessage[messageId].push(message);
        }
    }
}

function processTextMessageDeleteCommand (message) {
    const textMessage = message.m[0];
    const messageId = textMessage.id;

    const $roomMessage = findMessageBlockById(messageId);
    if ($roomMessage) {
        deletedMessageIds[messageId] = true;

        delete messageIdToTextSearchInfo[messageId];

        if ($roomMessage.find('.room-msg-text-inner').text().toLowerCase()
            .includes(currentTextSearchContext.text)) {
            textSearchDataChangedAt = new Date().getTime();
        }

        $roomMessage.off();
        $roomMessage.children().remove();
        $roomMessage.append('<p class="message-deleted-text">this message was deleted</p>');

        const $folkPicksMessage = findFolkPicksMessageBlockById(messageId);
        if ($folkPicksMessage) {
            $folkPicksMessage.off();
            $folkPicksMessage.children().remove();
            $folkPicksMessage.append('<p class="message-deleted-text">this message was deleted</p>');
        }

        //look for message quotes and clear them
        for (let roomMsgId in roomMessageIdToDOMElem) {
            const $roomMsgBlock = roomMessageIdToDOMElem[roomMsgId];

            if (messageId === parseInt($roomMsgBlock.attr('data-reply-to-msg-id'))) {
                $roomMsgBlock.find('.original-message-short-text').text(MESSAGE_UNAVAILABLE_PLACEHOLDER_TEXT);
            }
        }

        for (let picksMsgId in folkPicksMessageIdToDOMElem) {
            const $picksMsgBlock = folkPicksMessageIdToDOMElem[picksMsgId];

            if (messageId === parseInt($picksMsgBlock.attr('data-reply-to-msg-id'))) {
                $picksMsgBlock.find('.original-message-short-text').text(MESSAGE_UNAVAILABLE_PLACEHOLDER_TEXT);
            }
        }
    } else {
        //if target text message is not received yet - add this delete message to queue and process once target message arrives
        commandsToProcessDeleteMessage[messageId] = message;

        delete commandsToProcessEditMessage[messageId];
        delete commandsToProcessVoteMessage[messageId];
    }
}

function processRoomNotificationCommand (message) {
    switch (message.c) {
        case COMMANDS.NotifyMessagesLimitApproaching:
            showNotification(NOTIFICATION_TEXT_MESSAGE_LIMIT_APPROACHING);

            break;

        case COMMANDS.NotifyMessagesLimitReached:
            showNotification(NOTIFICATION_TEXT_MESSAGE_LIMIT_REACHED);

            break;
    }
}

function processRequestProcessedCommand (message) {
    //happens when initial room page info fetch completes
    if (message.rq === "room_c_j_done") {

        //if we just reconnected and build version changed - ask user to reload page
        if (message.bN !== currentBuildNumber) {
            $spinnerOverlayMidContentWr.addClass('d-none');

            showVersionChangedSpinner();

            //Not setting isRoomReconnectInProgress=false here

            shutdownSocket();

            return;
        }

        const isChatAtBottom = isChatScrolledBottom();

        const messageProcessingDetails = message.pd;
        const roomCreatedOrUpdatedStr = messageProcessingDetails.split(";")[0];
        const roomHasPasswordStr = messageProcessingDetails.split(";")[1];

        //if this page loaded after redirect from home page room form - then already notified user. Else - notify
        if (!isRedirectedFromHomePageRoomForm) {
            const isCreatedRoom = roomCreatedOrUpdatedStr === REQUEST_PROCESSING_DETAILS_ROOM_CREATED;

            if (isCreatedRoom) {
                showNotification(REQUEST_PROCESSING_DETAILS_ROOM_CREATED_MESSAGE, NOTIFICATION_SHOW_MS);
            } else {
                showNotification(REQUEST_PROCESSING_DETAILS_ROOM_JOINED_MESSAGE, NOTIFICATION_SHOW_MS);
            }
        }

        roomHasPassword = roomHasPasswordStr === REQUEST_PROCESSING_DETAILS_ROOM_HAS_PASSWORD;

        if (roomHasPassword) {
            $shareRoomHasPassword.removeClass('d-none');
        }

        roomUUID = message.rId;
        userInRoomUUID = message.uId;
        isLoggedIn = true;
        currentUserInfo = lastUserInfoByIdCache[message.uId];

        const lastVisitedRoomId =  LOCAL_STORAGE.getItem(LAST_VISITED_ROOM_ID_LOCAL_STORAGE_KEY);

        if (lastVisitedRoomId && lastVisitedRoomId === roomUUID) {
            const lastUnsentMessageText = LOCAL_STORAGE.getItem(LAST_UNSENT_MESSAGE_TEXT_LOCAL_STORAGE_KEY);

            if (lastUnsentMessageText) {
                $userMessageTextarea.val(lastUnsentMessageText);
            }
        } else {
            LOCAL_STORAGE.setItem(LAST_UNSENT_MESSAGE_TEXT_LOCAL_STORAGE_KEY, '');
        }

        LOCAL_STORAGE.setItem(LAST_VISITED_ROOM_ID_LOCAL_STORAGE_KEY, roomUUID);

        //will be absent only if room users list didn't manage to load in time, before current message with login acknowledgment
        if (currentUserInfo) {
            if (currentUserInfo.isAnon) {
                $userJoinedAsAnonPref.removeClass('d-none');
            } else {
                $userJoinedAsAnonPref.addClass('d-none');
            }

            $userJoinedAsName.text(decodeURIComponent(currentUserInfo.name));
        } else {
            $userJoinedAsName.text(UNKNOWN_USER_NAME);
        }

        setUserNameChangingInputToCurrentValue();

        setRoomInfo(getMillsFromNano(message.cAt), getOnlineUsersCountString());

        $roomWelcomeMessageUsername.append(
            $userJoinedAsNameWr.clone()
        );

        //set visuals for own user nameplate
        $roomInfoOnlineUsers.find(".room-info-online-user[data-user-id='" + userInRoomUUID + "']")
            .addClass('room-info-online-user-self');

        //cycle all messages and set visuals for own messages
        $roomMessagesWr.find(".room-msg-main-wr").each(function () {
            const $roomMessage = $(this);

            if ($roomMessage.attr('data-user-id') === userInRoomUUID) {
                $roomMessage.addClass('room-msg-main-wr-self');

                const messageId = parseInt($roomMessage.attr('data-msg-id'));

                if (messageId > lastCurrentUserSentMessageId) {
                    lastCurrentUserSentMessageId = messageId;
                }
            }
        });

        /* Add current room to visited rooms */
        addNewVisitedRoom();
        loadVisitedRooms(visitedRoomOnClickCallback);

        /* If some change-room-descr messages came before room login process completed (which is happening now) - process them */
        if (commandsToProcessChangeRoomDescription.length) {
            for (let i = 0; i < commandsToProcessChangeRoomDescription.length; i++) {
                const nextUnprocessedCommand = commandsToProcessChangeRoomDescription[i];

                processRoomDescriptionChangedCommand(nextUnprocessedCommand);
            }

            commandsToProcessChangeRoomDescription = [];
        }

        resizeWrappersHeight(isChatAtBottom);

        hideSpinnerOverlay();

        if (userInRoomUUID === roomCreatorUserInRoomUUID) {
            setTimeout(function () {
                showTopNotification(NOTIFICATION_TEXT_ROOM_CREATOR_SELF_VOTE, TOP_NOTIFICATION_SHOW_MS * 2, false);
            }, 2000)
        }

        scrollChatBottom();

        isRoomReconnectInProgress = false;

    } else {
        if (
            currentlyExpectedToCompleteUserActionIndicationId != null &&
            message.rq === "u_action_" + currentlyExpectedToCompleteUserActionIndicationId
        ) {
            currentlyExpectedToCompleteUserActionIndicationId = null;

            $userMessageCollapseWrSendingIndication.addClass('d-none');
        }
    }
}

function processErrorCommand (message, alternativeRoomNamePostfixes) {
    const businessError = BUSINESS_ERRORS[message.m[0].t];

    //if user just connected to same room again under same session token
    if (businessError.code === ERROR_CODE_ROOM_USER_DUPLICATION) {
        showSpinnerOverlay();
        $spinnerOverlayMidContentWr.addClass('d-none');
        $spinnerOverlayRoomUserDuplicationWr.removeClass('d-none');

        //Not setting isRoomReconnectInProgress=false here

        shutdownSocket();

        return;
    }

    if (isLoggedIn) {
        showError(businessError.text);
    } else {
        redirectToHomePageWithError(ROOM_TO_HOME_PG_REDIRECT_ERROR_BUSINESS, null, businessError, alternativeRoomNamePostfixes);
    }
}


/* Misc */

function attachLongTouchContextMenuToElement($elem) {
    $elem.off('contextmenu touchstart touchmove touchend');

    if (!isMobileClientDevice) {
        $elem.on('contextmenu', roomMessageOnContextMenu);
    } else {
        //logic to detect long touch, needed for ios safari

        $elem.on("touchstart", function(e) {
            // timer to detect long-touch
            longTouchTimeout = setTimeout(function() {
                isInsideLongTouch = true;
            }, 800);
        });
        $elem.on("touchmove", function(e) {
            clearTimeout(longTouchTimeout);

            if (isInsideLongTouch) {
                isInsideLongTouch = false;
            }
        });
        $elem.on("touchend", function(e) {
            clearTimeout(longTouchTimeout);

            if (isInsideLongTouch) {
                isInsideLongTouch = false;

                roomMessageOnContextMenu(e);
            }
        });
    }
}

function roomMessageOnContextMenu (e) {
    const $clickedBlock = $(e.target);

    cancelMessageTextSelectionMode();

    hideMobileKeyboard();

    //if clicked on link or started desktop text selection
    if ($clickedBlock.hasClass('message-highlight-link') ||
        (!isMobileClientDevice && getSelectionText())) {

        return;
    }

    stopPropagationAndDefault(e);

    //prevent calling context menu on wrong message, while exiting textarea 'big' mode
    if (new Date().getTime() - textareaBigModeExitedAt < DELAY_CONTEXT_MENU_AFTER_EXIT_BIG_MODE_MS) {
        return;
    }

    const $messageBlock = $(e.currentTarget);
    const messageId = parseInt($messageBlock.attr('data-msg-id'));
    const isMessageInFolkPicks = !!$messageBlock.closest('.picks-messages-wr').length;
    const isUserDrawingMessage = $messageBlock.attr('data-meta-marker') === MESSAGE_META_MARKER_TYPE_DRAWING;

    //for desktop - position menu block relative to clicked message block
    if (!isMobileClientDevice) {
        $messageContextMenu.removeClass('message-context-menu-mid-screen');

        //click coordinates inside element
        const clickCoordsInsideElementPair = getClickCoordinatesInsideElement($messageBlock[0], e);
        const x = clickCoordsInsideElementPair.x;
        const y = clickCoordsInsideElementPair.y;
        const messageBlockRect = clickCoordsInsideElementPair.rect;

        //how far click coords are from element's (!) left border
        const horizontalOffsetPercent = x * 100 / messageBlockRect.width;

        if (horizontalOffsetPercent < 50) {
            $messageContextMenu.css('right', 'unset');
            $messageContextMenu.css('left', x);
        } else {
            $messageContextMenu.css('left', 'unset');
            $messageContextMenu.css('right', messageBlockRect.width - x);
        }

        //how far click coords are from screen's (!) top border
        const verticalOffsetPercent = e.pageY * 100 / $window.height();

        if (verticalOffsetPercent < 50) {
            $messageContextMenu.css('bottom', 'unset');
            $messageContextMenu.css('top', y);
        } else {
            $messageContextMenu.css('top', 'unset');
            $messageContextMenu.css('bottom', messageBlockRect.height - y);
        }
    } else {
        //for mobile - draw conext menu at the middle of screen
        $messageContextMenu.addClass('message-context-menu-mid-screen');
    }

    /* Menu buttons */

    $messageContextMenuGotoButton.off();
    if (isMessageInFolkPicks) {
        $messageContextMenuGotoButton.removeClass('d-none');
        $messageContextMenuGotoButton.on('click', function (e) {
            stopPropagationAndDefault(e);

            hideFolkPicksMobile();
            $messageContextMenu.addClass('d-none');

            //find because current $messageBlock points to message copy in picks wrapper
            scrollToTargetMsg(findMessageBlockById(messageId));
        });

    } else {
        $messageContextMenuGotoButton.addClass('d-none');
    }

    //text selection button is only shown on mobile
    $messageContextMenuSelectButton.off();
    if (isMobileClientDevice && !isMessageInFolkPicks && !isUserDrawingMessage) {
        $messageContextMenuSelectButton.removeClass('d-none');

        const selectedText = getSelectionText();

        $messageContextMenuSelectButton.on('click', function (e) {
            stopPropagationAndDefault(e);

            hideGlobalTransparentOverlay();
            $messageContextMenu.addClass('d-none');

            startMessageTextSelectionMode($messageBlock);
        });
    } else {
        $messageContextMenuSelectButton.addClass('d-none');
    }

    $messageContextMenuCopyButton.off();
    if (!isUserDrawingMessage) {
        $messageContextMenuCopyButton.on('click', function (e) {
            stopPropagationAndDefault(e);

            hideGlobalTransparentOverlay();
            $messageContextMenu.addClass('d-none');

            const $anonBlock = $messageBlock.find('.room-msg-name-anon');

            const $messageTextBlockClone = $messageBlock.find('.room-msg-text-inner').clone();

            //if any link has preview loaded - then swap outer link with inner to get its text clearly
            const $messageLinkPreviewBlock = $messageTextBlockClone.find('.message-link-preview');
            if ($messageLinkPreviewBlock.length) {
                const $innerLink = $messageLinkPreviewBlock.find('.message-highlight-link');
                $messageLinkPreviewBlock.replaceWith($innerLink);
            }

            $messageTextBlockClone
                .find('.message-reply-to-message, .message-reply-to-user')
                .remove();

            $messageTextBlockClone.text($messageTextBlockClone.text());

            const messageToCopy =
                ($anonBlock.hasClass('d-none') ? '' : $anonBlock.text())
                + ($messageBlock.find('.room-msg-name').text()  + ':\n')
                + $messageTextBlockClone.text();

            copyStringToClipboard(messageToCopy,
                null,
                function () {
                    showError(ERROR_NOTIFICATION_TEXT_FAILED_COPY_MESSAGE_TO_CLIPBOARD);
                }
            );
        });
    } else {
        $messageContextMenuCopyButton.addClass('d-none');
    }

    $messageContextMenuRespondMessageButton.off();
    $messageContextMenuRespondMessageButton.on('click', function (e) {
        stopPropagationAndDefault(e);

        hideGlobalTransparentOverlay();
        $messageContextMenu.addClass('d-none');

        if (isUserDrawingMessage) {
            const $drawingBlock = $messageBlock.find('.user-drawing-message-preview');
            loadImageToCanvas($drawingBlock.attr('src'));

            hideFolkPicksMobile();
            showUserInputDrawingBlock();

        } else {
            const messageShortText = cutMessageTextForResponse(
                $messageBlock.find('.room-msg-text-inner').text()
            );

            const $anonBlock = $messageBlock.find('.room-msg-name-anon');
            const messageAuthorName = $messageBlock.find('.room-msg-name-val').text();
            const messageAuthorUserId = $messageBlock.attr('data-user-id');

            startReplyToMessage(messageId, messageShortText, messageAuthorUserId, messageAuthorName, !!$anonBlock.length);
        }
    });

    $messageContextMenuDownloadDrawingButton.off();
    if (isUserDrawingMessage) {
        $messageContextMenuDownloadDrawingButton.removeClass('d-none');

        $messageContextMenuDownloadDrawingButton.on('click', function (e) {
            stopPropagationAndDefault(e);

            saveBase64AsFile(
                $messageBlock.find('.user-drawing-message-preview').attr('src'),
                currentRoomName + '-' + messageId + '.png'
            );
        });
    } else {
        $messageContextMenuDownloadDrawingButton.addClass('d-none');
    }

    $messageContextMenuEditButton.off();
    $messageContextMenuDeleteButton.off();
    $messageContextMenuRespondUserButton.off();

    if ($messageBlock.attr('data-user-id') === userInRoomUUID) {
        $messageContextMenuRespondUserButton.addClass('d-none');

        if (!isUserDrawingMessage) {
            $messageContextMenuEditButton.removeClass('d-none');
            $messageContextMenuEditButton.on('click', function (e) {
                stopPropagationAndDefault(e);

                hideGlobalTransparentOverlay();
                $messageContextMenu.addClass('d-none');

                editUserMessage(messageId, $messageBlock);
            });
        } else {
            $messageContextMenuEditButton.addClass('d-none');
        }

        $messageContextMenuDeleteButton.removeClass('d-none');
        $messageContextMenuDeleteButton.on('click', function () {
            stopPropagationAndDefault(e);

            hideGlobalTransparentOverlay();
            $messageContextMenu.addClass('d-none');

            deleteUserMessage(messageId);
        });
    } else {
        $messageContextMenuRespondUserButton.removeClass('d-none');

        $messageContextMenuEditButton.addClass('d-none');
        $messageContextMenuDeleteButton.addClass('d-none');

        $messageContextMenuRespondUserButton.on('click', function (e) {
            stopPropagationAndDefault(e);

            hideGlobalTransparentOverlay();
            $messageContextMenu.addClass('d-none');

            const $anonBlock = $messageBlock.find('.room-msg-name-anon');
            const messageAuthorName = $messageBlock.find('.room-msg-name-val').text();
            const messageAuthorUserId = $messageBlock.attr('data-user-id');

            startReplyToUser(messageAuthorUserId, messageAuthorName, !!$anonBlock.length);
        });
    }

    //for desktop - position relative to message block, for mobile - middle of screen
    if (!isMobileClientDevice) {
        $messageBlock.append($messageContextMenu);
    } else {
        $body.append($messageContextMenu);
    }

    $messageContextMenu.removeClass('d-none');

    //if this message was clicked from mobile folk picks - dont show global overlay (folk picks on mobile are already absolutely positioned so wont work)
    if (!isMessageInFolkPicks || $folkPicksWr.css('position') !== 'fixed') {
        showGlobalTransparentOverlay();
    }
}

function setUserNameChangingInputToCurrentValue() {
    if (currentUserInfo) {
        $userJoinedAsChangeInput.val(decodeURIComponent(currentUserInfo.name));
    }
}

function onRoomMessageAuthorClick (e) {
    if (!isMobileClientDevice && getSelectionText()) {
        stopPropagationAndDefault(e);
        return;
    }

    const $messageBlock = $(e.currentTarget).closest('.room-msg-main-wr');

    const $anonBlock = $messageBlock.find('.room-msg-name-anon');
    const messageAuthorName = $messageBlock.find('.room-msg-name-val').text();
    const messageAuthorUserId = $messageBlock.attr('data-user-id');

    startReplyToUser(messageAuthorUserId, messageAuthorName, !!$anonBlock.length);
}

function startReplyToUser(messageAuthorUserId, messageAuthorName, isAnon) {
    if (messageAuthorUserId === userInRoomUUID) {
        return;
    }

    const isChatAtBottom = isChatScrolledBottom();

    cancelReplyToMessage();
    hideFolkPicksMobile();

    userReplyInProgressToId = messageAuthorUserId;

    //show user message block if it is minimized
    if ($userMessageContentWr.css('display') === 'none') {
        toggleUserMessageBlock();
    }

    if (isAnon) {
        $answerToUserWr.find('.answer-to-user-anon-pref').removeClass('d-none');
    } else {
        $answerToUserWr.find('.answer-to-user-anon-pref').addClass('d-none');
    }

    $answerToUserWr.find('.answer-to-user-name').text(messageAuthorName);

    $answerToUserWr.removeClass('d-none');

    resizeWrappersHeight(isChatAtBottom);

    $userMessageTextarea.focus();

    setTimeout(function () {
        $body[0].scrollTop = getOffsetTop($answerToUserWr[0]);
    }, 200);
}

function cancelReplyToUser() {
    if (userReplyInProgressToId) {
        const isChatAtBottom = isChatScrolledBottom();

        userReplyInProgressToId = null;

        $answerToUserWr.addClass('d-none');

        resizeWrappersHeight(isChatAtBottom);
    }
}

function appendReplyToUserBlock($messageBlockToAppend, replyToUserId, messageId) {
    const replyToUserInfo = lastUserInfoByIdCache[replyToUserId];

    let replyToUserName = UNKNOWN_USER_NAME;
    let replyToUserIsAnon = true;

    if (replyToUserInfo) {
        replyToUserName = decodeURIComponent(replyToUserInfo.name);
        replyToUserIsAnon = replyToUserInfo.isAnon;
    } else {
        if (!userIdToMessageIdsWithUnknownReplyToUserName[replyToUserId]) {
            userIdToMessageIdsWithUnknownReplyToUserName[replyToUserId] = [];
        }
        userIdToMessageIdsWithUnknownReplyToUserName[replyToUserId].push(messageId);
    }

    const $messageReplyToUserBlock = $('<div class="message-reply-to-user" data-user-id="' + replyToUserId + '">');
    const $messageReplyToAnonBlock = $('<span class="message-reply-to-user-anon-pref d-none">anon&nbsp;</span>');

    const $messageReplyToUserName = $('<span class="message-reply-to-user-name">');
    $messageReplyToUserName.text(replyToUserName + ',');

    if (replyToUserIsAnon) {
        $messageReplyToAnonBlock.removeClass('d-none');
    }

    $messageReplyToUserBlock.append($messageReplyToAnonBlock);
    $messageReplyToUserBlock.append($messageReplyToUserName);

    $messageBlockToAppend.append($messageReplyToUserBlock);
}

function onRoomMessageIdClick (e) {
    if (!isMobileClientDevice && getSelectionText()) {
        stopPropagationAndDefault(e);
        return;
    }

    const $messageBlock = $(e.currentTarget).closest('.room-msg-main-wr');
    const messageId = parseInt($messageBlock.attr('data-msg-id'));
    const messageShortText = cutMessageTextForResponse(
        $messageBlock.find('.room-msg-text-inner').text()
    );

    const $anonBlock = $messageBlock.find('.room-msg-name-anon');
    const messageAuthorName = $messageBlock.find('.room-msg-name-val').text();
    const messageAuthorUserId = $messageBlock.attr('data-user-id');

    startReplyToMessage(messageId, messageShortText, messageAuthorUserId, messageAuthorName, !!$anonBlock.length);
}

function startReplyToMessage(messageId, messageShortText, messageAuthorUserId, messageAuthorName, isAnon) {
    const isChatAtBottom = isChatScrolledBottom();

    cancelReplyToUser();

    messageReplyInProgressToId = messageId;
    userReplyInProgressToId = messageAuthorUserId;

    hideFolkPicksMobile();
    //show user message block if it is minimized
    if ($userMessageContentWr.css('display') === 'none') {
        toggleUserMessageBlock();
    }

    if (isAnon) {
        $answerToMessageWr.find('.reply-to-message-anon-pref').removeClass('d-none');
    } else {
        $answerToMessageWr.find('.reply-to-message-anon-pref').addClass('d-none');
    }

    $answerToMessageWr.find('.reply-to-message-user-name').text(messageAuthorName + ': ');
    $answerToMessageWr.find('.reply-to-message-short-text').text('"' + messageShortText + '"');

    $answerToMessageWr.removeClass('d-none');

    resizeWrappersHeight(isChatAtBottom);

    $userMessageTextarea.focus();

    setTimeout(function () {
        $body[0].scrollTop = getOffsetTop($answerToMessageWr[0]);
    }, 200);
}

function cancelReplyToMessage() {
    if (messageReplyInProgressToId) {
        const isChatAtBottom = isChatScrolledBottom();

        messageReplyInProgressToId = null;
        userReplyInProgressToId = null;

        $answerToMessageWr.addClass('d-none');

        resizeWrappersHeight(isChatAtBottom);
    }
}

function appendReplyToMessageBlock($messageBlockToAppend, replyToMessageId, replyToUserId, messageId, originalMessageShortText) {
    const $replyToMessageBlock = $('<div class="message-reply-to-message">');

    const $messageReplyToIdBlock = $('<span class="reply-msg-to-id">');
    $messageReplyToIdBlock.append('>>&nbsp;#' + replyToMessageId);

    $replyToMessageBlock.append($messageReplyToIdBlock);


    const replyToUserInfo = lastUserInfoByIdCache[replyToUserId];

    let replyToUserName = UNKNOWN_USER_NAME;
    let replyToUserIsAnon = true;

    if (replyToUserInfo) {
        replyToUserName = decodeURIComponent(replyToUserInfo.name);
        replyToUserIsAnon = replyToUserInfo.isAnon;
    } else {
        if (!userIdToMessageIdsWithUnknownReplyToUserName[replyToUserId]) {
            userIdToMessageIdsWithUnknownReplyToUserName[replyToUserId] = [];
        }
        userIdToMessageIdsWithUnknownReplyToUserName[replyToUserId].push(messageId);
    }

    const $messageReplyToUserBlock = $('<div class="message-reply-to-user" data-user-id="' + replyToUserId + '">');
    const $messageReplyToAnonBlock = $('<span class="message-reply-to-user-anon-pref d-none">anon&nbsp;</span>');

    const $messageReplyToUserName = $('<span class="message-reply-to-user-name">');
    $messageReplyToUserName.text(replyToUserName + ':');

    if (replyToUserIsAnon) {
        $messageReplyToAnonBlock.removeClass('d-none');
    }

    $messageReplyToUserBlock.append($messageReplyToAnonBlock);
    $messageReplyToUserBlock.append($messageReplyToUserName);

    $replyToMessageBlock.append($messageReplyToUserBlock);


    const $repliedMessageShortText = $('<span class="original-message-short-text">');
    $repliedMessageShortText.text(originalMessageShortText);

    $replyToMessageBlock.append($repliedMessageShortText);

    $messageBlockToAppend.append($replyToMessageBlock);
}

function sendData (ws, inputStr) {
    if (!isWsOpen()) {
        showError(BUSINESS_ERRORS[ERROR_CODE_CONNECTION_ERROR].text);

        return false;
    }

    try {
        ws.send(inputStr);

        return true;
    } catch (ex) {
        showError(BUSINESS_ERRORS[ERROR_CODE_CONNECTION_ERROR].text);

        return false;
    }
}

function isWsOpen() {
    return ws.readyState === ws.OPEN
}

function addNewVisitedRoom() {
    let visitedRoomsStr = LOCAL_STORAGE.getItem(VISITED_ROOMS_LOCAL_STORAGE_KEY);

    if (!visitedRoomsStr) {
        visitedRoomsStr = "[]";
    }

    const newVisitedRooms = [];

    const visitedRooms = JSON.parse(visitedRoomsStr);
    let room = null;

    for (let i = 0; i < visitedRooms.length; i++) {
        const nextRoom = visitedRooms[i];

        if (nextRoom.roomName === ROOM_NAME) {
            room = nextRoom;
        } else {
            newVisitedRooms.push(nextRoom);
        }
    }

    if (room) {
        room.visitedAt = new Date().getTime();
    } else {
        room = {
            roomName: ROOM_NAME,
            visitedAt: new Date().getTime()
        };
    }

    newVisitedRooms.push(room);

    storeVisitedRoomsArray(newVisitedRooms);
}

function visitedRoomOnClickCallback (e) {
    let roomName = $(e.currentTarget).text();

    if (roomName.trim() === ROOM_NAME) {
        hideMyRecentRoomsPopup();
        return;
    }

    shutdownSocket();

    LOCAL_STORAGE.setItem(REDIRECT_VARIABLE_LOCAL_STORAGE_KEY, JSON.stringify({
        redirectFrom: REDIRECT_FROM_ANY_PAGE_ON_RECENT_ROOM_CLICK,
        redirectedAt: new Date().getTime(),
        roomName: roomName,
    }));

    redirectToURL("/");
}

function textareaSwitchToBigMode () {
    if (!$userMessageTextarea.isBig) {
        $userMessageTextarea.isBig = true;

        const isChatAtBottom = isChatScrolledBottom();

        $userMessageTextarea.addClass('user-message-textarea-big');
        $userMessageTextarea.removeClass('user-message-textarea-small');

        resizeWrappersHeight(isChatAtBottom && !messageEditingInProgressId);

        setTimeout(function () {
            $body[0].scrollTop = getOffsetTop($userMessageTextarea[0]);
            }, 200);
    }
}

function textareaSwitchToSmallMode () {
    if ($userMessageTextarea.isBig) {
        $userMessageTextarea.isBig = false;

        const isChatAtBottom = isChatScrolledBottom();

        $userMessageTextarea.removeClass('user-message-textarea-big');
        $userMessageTextarea.addClass('user-message-textarea-small');

        resizeWrappersHeight(isChatAtBottom);

        textareaBigModeExitedAt = new Date().getTime();
    }
}

function isUserInfoByIdCacheUpdated(newUserInfoByIdCache, oldUserInfoByIdCache) {
    if (!oldUserInfoByIdCache || (Object.keys(newUserInfoByIdCache).length !== Object.keys(oldUserInfoByIdCache).length)) {
        return true;
    }

    for (let userId in newUserInfoByIdCache) {
        const userFromNewCache = newUserInfoByIdCache[userId];
        const userFromOldCache = oldUserInfoByIdCache[userId];

        const userFoundAndIsEqual = userFromNewCache && userFromOldCache &&
            userFromNewCache.id === userFromOldCache.id &&
            userFromNewCache.name === userFromOldCache.name &&
            userFromNewCache.isAnon === userFromOldCache.isAnon &&
            userFromNewCache.isOnlineInRoom === userFromOldCache.isOnlineInRoom;

        if (!userFoundAndIsEqual) {
            return true;
        }
    }

    return false;
}

function startMessageTextSelectionMode ($messageBlock) {
    cancelMessageTextSelectionMode();

    messageTextSelectionInProgressId = parseInt($messageBlock.attr('data-msg-id'));
    messageTextSelectionInProgressIsFolkPick = !!$messageBlock.closest('.picks-messages-wr').length;

    $messageBlock.off('contextmenu touchstart touchmove touchend');

    $messageBlock.addClass('text-selection-active');

    if (isMobileClientDevice) {
        //mobile
        $messageBlock.removeClass('noselect');

        //for mobile we have to make a hack because of something that seems like mobile chrome bug:
        //if we select text in message, that is larger than its parent tag (and thus parent tag scrolls) - when selection goes out of viewport it jumps to ANY first child of body that has 'height' != 0
        //so we do all below actions and then revet back after selection finishes
        $body.css('position', 'initial');

        $userMessageContentWr.css('display', 'none');

        mainNavbarChildrenBlocks = $mainNavbarBlock.children();
        mainNavbarChildrenBlocks.detach();

        roomInfoChildrenBlocks = $roomInfoBlock.children();
        roomInfoChildrenBlocks.detach();

        userMessageWrChildrenBlocks = $userMessageWr.children();
        userMessageWrChildrenBlocks.detach();

        $folkPicksWr.detach();

        $mobileTextSelectionButtonsBlock.removeClass('d-none');

        //text selection control buttons for mobile - exists all the time and is shown/hidden

        $mobileCopyButtonBlock.off();
        $mobileCopyButtonBlock.on('mousedown', function () {
            const selectedText = getSelectionText();

            if (!selectedText) {
                return;
            }

            copyStringToClipboard(selectedText);

            cancelMessageTextSelectionMode();
        });

        $mobileCancelButtonBlock.off();
        $mobileCancelButtonBlock.on('mousedown', cancelMessageTextSelectionMode);

        $roomMessagesWr.find('.room-msg-main-wr').each((idx, message) => {
            const $message = $(message);

            if (messageTextSelectionInProgressId !== parseInt($message.attr('data-msg-id'))) {
                $message.css('opacity', '0.65');
            }
        });

        transformMobileTextSelectionCopyButton(getSelectionText());

        moveMobileTextSelectionControlsToMessage();

        resizeWrappersHeight();

        scrollToMessageDuringTextSelection($messageBlock[0]);

    } else {
        //desktop

        //text selection control buttons for desktop - are created/deleted each time in scope of message block being selected

        const $copyButtonBlock = $('<div class="text-copy-button">copy text</div>');
        $copyButtonBlock.on('mousedown', function (e) {
            const selectedText = getSelectionText();

            if (!selectedText) {
                return;
            }

            copyStringToClipboard(selectedText);

            cancelMessageTextSelectionMode($(e.currentTarget).closest('.room-msg-main-wr'), true);
        });

        const $cancelButtonBlock = $('<div class="text-cancel-button">cancel</div>');
        $cancelButtonBlock.on('mousedown', function (e) {
            cancelMessageTextSelectionMode($(e.currentTarget).closest('.room-msg-main-wr'), true);
        });

        $messageBlock.find('.text-copy-button, .text-cancel-button').remove();

        $messageBlock
            .append($copyButtonBlock, $cancelButtonBlock);

        //fix buttons position for 1st message (for desktop its possible to select text inside folk picks messages)
        if ($messageBlock.closest('.picks-messages-wr').length) {
            if (messageTextSelectionInProgressIsFolkPick &&
                messageTextSelectionInProgressId === parseInt($($folkPicksMessagesWr.children()[0]).attr('data-msg-id'))) {
                $copyButtonBlock.addClass('text-selection-picks-first-child');
                $cancelButtonBlock.addClass('text-selection-picks-first-child');
            }
        }

        adaptDesktopTextSelectionControls();
    }
}

function cancelMessageTextSelectionMode ($messageBlock, animateControls) {
    if (messageTextSelectionInProgressId) {
        clearTextSelection();

        if (!$messageBlock || !$messageBlock.length) {
            //for desktop its possible to select text inside folk picks messages
            $messageBlock = messageTextSelectionInProgressIsFolkPick
                ? findFolkPicksMessageBlockById(messageTextSelectionInProgressId)
                : findMessageBlockById(messageTextSelectionInProgressId);
        }

        if (!$messageBlock) {
            return;
        }

        messageTextSelectionInProgressId = null;

        messageTextSelectionInProgressIsFolkPick = false;

        attachLongTouchContextMenuToElement($messageBlock);

        $messageBlock.removeClass('text-selection-active');

        if (isMobileClientDevice) {
            //mobile
            $body.css('position', 'fixed');

            $messageBlock.addClass('noselect');

            $mainNavbarBlock.append(mainNavbarChildrenBlocks);
            $roomInfoBlock.append(roomInfoChildrenBlocks);

            $userMessageWr.append(userMessageWrChildrenBlocks);

            $containerMain.find('>.row').append($folkPicksWr);

            if (anyMessageIsFolkPicked()) {
                $folkPicksEmptyMessagesWr.remove();
            } else {
                $folkPicksMessagesWr.append($folkPicksEmptyMessagesWr);
            }

            $userMessageContentWr.css('display', '');

            $mobileTextSelectionButtonsBlock.addClass('d-none');
            $roomMessagesWr.css('padding-right', '');

            $mobileTextSelectionToMessageOverlayTopBlock.addClass('d-none');
            $mobileTextSelectionToMessageOverlayBotBlock.addClass('d-none');

            $roomMessagesWr.find('.room-msg-main-wr').each((idx, message) => {
                $(message).css('opacity', '');
            });

            $containerMainCenterWr
                .removeClass('col-xl-12')
                .addClass('col-xl-9');

            resizeWrappersHeight();

            scrollToElement($messageBlock[0]);

        } else {
            //desktop

            const $copyButtonBlock = $messageBlock.find('.text-copy-button').remove();
            const $cancelButtonBlock = $messageBlock.find('.text-cancel-button').remove();

            if (animateControls) {
                $copyButtonBlock.fadeTo(100, 0, function () {
                    $copyButtonBlock.remove();
                });

                $cancelButtonBlock.fadeTo(100, 0, function () {
                    $cancelButtonBlock.remove();
                });
            } else {
                $copyButtonBlock.remove();
                $cancelButtonBlock.remove();
            }
        }
    }
}

function reconnectToRoom () {
    isRoomReconnectInProgress = true;

    showSpinnerOverlay();

    cancelMessageEdit();
    cancelMessageTextSelectionMode();
    cancelReplyToUser();
    cancelReplyToMessage();

    /* Clear all variables set since last login */

    clearTimeout(wsReconnectTimeout);
    clearTimeout(onSelectionChangeTimeout);

    initializeVariables();

    /* Clear all info rendered since last login */

    $roomInfoOnlineUsers.children().remove();
    $roomWelcomeMessageUsername.find('.user-message-joined-as-block').remove();

    $userJoinedAsAnonPref.addClass('d-none');
    $userJoinedAsName.text('');
    $userJoinedAsChangeInput.val('');

    $roomInfoDescriptionEmptyText.removeClass('d-none');
    $roomInfoDescriptionEmptyTextCreator.addClass('d-none');

    $roomInfoDescriptionText.text('');
    $roomInfoUsersCount.text('X (X)');
    $roomInfoCreationTime.text('XX:XX');

    //messages in main and picks wrappers must stay until we reload page. But we cant just delete all and redraw after re-load,
    //since some new messages could come AFTER reload but BEFORE 'all-messages' command.
    //So mark them 'to-be-deleted' and remove AFTER new page is fully loaded (and even same messages are drawn once more)
    $roomMessagesWr.find(".room-msg-main-wr").each(function () {
        const $messageBlock = $(this);

        recursiveApplyFuncToDom($messageBlock, function ($elem) {
            $elem.off();
        });

        $messageBlock.addClass(CLASS_MESSAGE_TO_CLEAN_UP);
    });

    $folkPicksMessagesWr.find(".room-msg-main-wr").each(function () {
        const $messageBlock = $(this);

        recursiveApplyFuncToDom($messageBlock, function ($elem) {
            $elem.off();
        });

        $messageBlock.addClass(CLASS_MESSAGE_TO_CLEAN_UP);
    });

    /* Cancel any active UI action */

    hideGlobalTransparentOverlay();

    hideFolkPicksMobile();
    hideMenuMobile();
    hideMyRecentRoomsPopup();
    hideShareRoomPopup();

    cancelActionsUnderGlobalTransparentOverlay();

    textareaSwitchToSmallMode();

    cancelTextSearch();

    resizeWrappersHeight(true);

    wsReconnectTimeout = setTimeout(createOrJoinRoom, reconnectAttempt * 3000);

    if (reconnectAttempt < 10) {
        reconnectAttempt++;
    }
}

function findMessageBlockById (messageId) {
    return roomMessageIdToDOMElem[messageId];
}

function findFolkPicksMessageBlockById (messageId) {
    return folkPicksMessageIdToDOMElem[messageId];
}

function textSearchPrev () {
    const textLower = $searchBarInput.val().toLowerCase().trim();

    if (!textLower) {
        return;
    }

    $roomMessagesWr.find('.room-msg-main-wr.search-msg-selected').removeClass('search-msg-selected');

    //setup new 'global' search context if none in progress
    if (!currentTextSearchContext.text ||
        currentTextSearchContext.text !== textLower ||
        currentTextSearchContext.startedAt < textSearchDataChangedAt) {

        initNewTextSearch(textLower);
    }

    if (currentTextSearchContext.searchResultIdx === -1) {
        currentTextSearchContext.searchResultIdx = currentTextSearchContext.searchResultMessagesContextArr.length - 1;
    }

    highlightAllTextSearchResults(textLower);

    if (!currentTextSearchContext.searchResultMessagesContextArr.length) {
        return;
    }

    //get previous 'message' search context from 'global' one, using 'current message index'
    //consider acquired context to be 'previous', even if it wasn't
    let prevMsgSearchContext = currentTextSearchContext.searchResultMessagesContextArr[currentTextSearchContext.searchResultIdx];

    //if last search action was not in current order - update search context for all messages
    if (currentTextSearchContext.currentDirection !== 'prev') {
        currentTextSearchContext.currentDirection = 'prev';

        for (let i = 0; i < currentTextSearchContext.searchResultMessagesContextArr.length; i++) {
            const nextMsg = currentTextSearchContext.searchResultMessagesContextArr[i];

            //return 1st field down the priority circle, that has search results
            nextMsg.activeSearchArea = findPrevActiveAreaForSearch(prevMsgSearchContext);
            nextMsg.searchResultInAreaIdx = findAreaSearchResultInitialArrayIndex(nextMsg, true);
        }
    }

    //if last 'search result selection step' inside area was the last 'previous' search result in currently active 'search area' - step to new 'previous' area.
    //Or if this was last area - to new 'previous' message

    //wont be true if last step was not 'prev'
    if (prevMsgSearchContext.searchResultInAreaIdx < 0) {
        //step to new area if available
        const newPrevSearchArea = findPrevActiveAreaForSearch(prevMsgSearchContext, prevMsgSearchContext.activeSearchArea);

        if (newPrevSearchArea && prevMsgSearchContext.activeSearchArea !== TEXT_SEARCH_AREA_MSG_ID) {
            prevMsgSearchContext.activeSearchArea = newPrevSearchArea;

            prevMsgSearchContext.searchResultInAreaIdx = findAreaSearchResultInitialArrayIndex(prevMsgSearchContext, true);

        } else {
            //if not - step to new 'previous' message and its 1st priority 'previous' area

            //reset area/index for old message context
            prevMsgSearchContext.activeSearchArea = findPrevActiveAreaForSearch(prevMsgSearchContext);
            prevMsgSearchContext.searchResultInAreaIdx = findAreaSearchResultInitialArrayIndex(prevMsgSearchContext, true);

            //step to next message
            if (currentTextSearchContext.searchResultIdx === 0) {
                currentTextSearchContext.searchResultIdx = currentTextSearchContext.searchResultMessagesContextArr.length - 1;
            } else {
                currentTextSearchContext.searchResultIdx--;
            }

            prevMsgSearchContext = currentTextSearchContext.searchResultMessagesContextArr[currentTextSearchContext.searchResultIdx];

            //step to 1st available area
            prevMsgSearchContext.activeSearchArea = findPrevActiveAreaForSearch(prevMsgSearchContext);
            prevMsgSearchContext.searchResultInAreaIdx = findAreaSearchResultInitialArrayIndex(prevMsgSearchContext, true);
        }
    }

    //highlight message block
    const $prevMsgBlock = findMessageBlockById(prevMsgSearchContext.id);

    $prevMsgBlock.addClass('search-msg-selected');

    //if empty - pick default area inside block for 'prev' search
    if (!prevMsgSearchContext.activeSearchArea) {
        prevMsgSearchContext.activeSearchArea = findPrevActiveAreaForSearch(prevMsgSearchContext);

        prevMsgSearchContext.searchResultInAreaIdx = findAreaSearchResultInitialArrayIndex(prevMsgSearchContext, true);
    }

    //highlight 'prev' text match inside message
    //we now have message and its particular search area (username, message text, reply text etc.) to highlight result
    let $areaBlockToHighlight = findActiveAreaBlockInsideMessageBlock(prevMsgSearchContext.activeSearchArea, $prevMsgBlock);

    highlightTextInBlockByOccurrenceIndexBright(textLower, prevMsgSearchContext.searchResultInAreaIdx, $areaBlockToHighlight);

    scrollToElement($prevMsgBlock.find('.message-text-highlight')[0]);

    //step to 'prev' index in area results
    prevMsgSearchContext.searchResultInAreaIdx--;
}

function textSearchNext () {
    const textLower = $searchBarInput.val().toLowerCase().trim();

    if (!textLower) {
        return;
    }

    $roomMessagesWr.find('.room-msg-main-wr.search-msg-selected').removeClass('search-msg-selected');

    //setup new 'global' search context if none in progress
    if (!currentTextSearchContext.text ||
        currentTextSearchContext.text !== textLower ||
        currentTextSearchContext.startedAt < textSearchDataChangedAt) {

        initNewTextSearch(textLower);
    }

    if (currentTextSearchContext.searchResultIdx === -1) {
        currentTextSearchContext.searchResultIdx = 0;
    }

    highlightAllTextSearchResults(textLower);

    if (!currentTextSearchContext.searchResultMessagesContextArr.length) {
        return;
    }

    //get next 'message' search context from 'global' one, using 'current message index'
    //consider acquired context to be 'next', even if it wasn't
    let nextMsgSearchContext = currentTextSearchContext.searchResultMessagesContextArr[currentTextSearchContext.searchResultIdx];

    //if last search action was not in current order - update search context for all messages
    if (currentTextSearchContext.currentDirection !== 'next') {
        currentTextSearchContext.currentDirection = 'next';

        for (let i = 0; i < currentTextSearchContext.searchResultMessagesContextArr.length; i++) {
            const nextMsg = currentTextSearchContext.searchResultMessagesContextArr[i];

            nextMsg.activeSearchArea = findNextActiveAreaForSearch(nextMsgSearchContext);
            nextMsg.searchResultInAreaIdx = findAreaSearchResultInitialArrayIndex(nextMsg, false);
        }
    }

    //use findAreaSearchResultInitialArrayIndex() with isPrev=true to get max index within particular area search results array
    const maxAreaSearchResultIndex = findAreaSearchResultInitialArrayIndex(nextMsgSearchContext, true);

    //if last 'search result selection step' inside area was the last 'next' search result in currently active 'search area' - step to new 'next' area.
    //Or if this was last area - to new 'next' message

    //wont be true if last step was not 'next'
    if (nextMsgSearchContext.searchResultInAreaIdx > maxAreaSearchResultIndex) {
        const nextSearchArea = findNextActiveAreaForSearch(nextMsgSearchContext, nextMsgSearchContext.activeSearchArea);

        if (nextSearchArea && nextMsgSearchContext.activeSearchArea !== TEXT_SEARCH_AREA_TEXT) {
            nextMsgSearchContext.activeSearchArea = nextSearchArea;

            nextMsgSearchContext.searchResultInAreaIdx = findAreaSearchResultInitialArrayIndex(nextMsgSearchContext, false);

        } else {
            //if not - step to new 'next' message and its 1st priority 'next' area

            //reset area/index for old message context
            nextMsgSearchContext.activeSearchArea = findNextActiveAreaForSearch(nextMsgSearchContext);
            nextMsgSearchContext.searchResultInAreaIdx = findAreaSearchResultInitialArrayIndex(nextMsgSearchContext, false);

            if (currentTextSearchContext.searchResultIdx === currentTextSearchContext.searchResultMessagesContextArr.length - 1) {
                currentTextSearchContext.searchResultIdx = 0;
            } else {
                currentTextSearchContext.searchResultIdx++;
            }

            nextMsgSearchContext = currentTextSearchContext.searchResultMessagesContextArr[currentTextSearchContext.searchResultIdx];

            //step to 1st available area
            nextMsgSearchContext.activeSearchArea = findNextActiveAreaForSearch(nextMsgSearchContext);
            nextMsgSearchContext.searchResultInAreaIdx = findAreaSearchResultInitialArrayIndex(nextMsgSearchContext, false);
        }
    }

    //highlight message block
    const $nextMsgBlock = findMessageBlockById(nextMsgSearchContext.id);

    $nextMsgBlock.addClass('search-msg-selected');

    //if empty - pick default area inside block for 'next' search
    if (!nextMsgSearchContext.activeSearchArea) {
        nextMsgSearchContext.activeSearchArea = findNextActiveAreaForSearch(nextMsgSearchContext);

        nextMsgSearchContext.searchResultInAreaIdx = findAreaSearchResultInitialArrayIndex(nextMsgSearchContext, false);
    }

    //highlight 'next' text match inside message
    //we now have message and its particular search area (username, message text, reply text etc.) to highlight result
    let $areaBlockToHighlight = findActiveAreaBlockInsideMessageBlock(nextMsgSearchContext.activeSearchArea, $nextMsgBlock);

    highlightTextInBlockByOccurrenceIndexBright(textLower, nextMsgSearchContext.searchResultInAreaIdx, $areaBlockToHighlight);

    scrollToElement($nextMsgBlock.find('.message-text-highlight')[0]);

    //step to 'next' index in area results
    nextMsgSearchContext.searchResultInAreaIdx++;
}

/*
* text - lower case expected
 */
function initNewTextSearch(text) {
    cancelTextSearchHighlights();

    const matchingMessageSearchContextArr = [];

    for (let nextMsgIdFromSearchInfoCache in messageIdToTextSearchInfo) {
        const messageSearchInfo = messageIdToTextSearchInfo[nextMsgIdFromSearchInfoCache];

        //arrays of substring indexes
        const matchesInMessageIdText = getSubstringIndexes('#' + nextMsgIdFromSearchInfoCache, text);
        const matchesInAuthorName = getSubstringIndexes(messageSearchInfo.authorName, text);
        const matchesInMessageTimeText = getSubstringIndexes(messageSearchInfo.messageTimeText, text);

        let matchesInReplyToMessageIdText = [];
        if (messageSearchInfo.replyToMessageId) {
            matchesInReplyToMessageIdText = getSubstringIndexes('#' + messageSearchInfo.replyToMessageId, text);
        }

        const matchesInReplyToUserName = getSubstringIndexes(messageSearchInfo.replyToUserName, text);
        const matchesInReplyToMessageText = getSubstringIndexes(messageSearchInfo.replyToMessageText, text);
        const matchesInText = getSubstringIndexes(messageSearchInfo.text, text);

        if (matchesInMessageIdText.length || matchesInAuthorName.length || matchesInMessageTimeText.length ||
            matchesInReplyToMessageIdText.length || matchesInReplyToUserName.length || matchesInReplyToMessageText.length || matchesInText.length) {
            matchingMessageSearchContextArr.push({
                id: nextMsgIdFromSearchInfoCache,
                //arrays of substring indexes
                matchesInMessageIdText: matchesInMessageIdText,
                matchesInAuthorName: matchesInAuthorName,
                matchesInMessageTimeText: matchesInMessageTimeText,
                matchesInReplyToMessageIdText: matchesInReplyToMessageIdText,
                matchesInReplyToUserName: matchesInReplyToUserName,
                matchesInReplyToMessageText: matchesInReplyToMessageText,
                matchesInText: matchesInText,
                //code name of string field inside message block, which we will step through, if any matches of target string are to be found there
                activeSearchArea: null,
                searchResultInAreaIdx: -1,
            });
        }
    }

    currentTextSearchContext = {
        text: text,
        startedAt: new Date().getTime(),
        searchResultMessagesContextArr: matchingMessageSearchContextArr,
        searchResultIdx: -1,
        currentDirection: null,
        nextSearchFixLinkId: 0,
        linkOrigHrefByFixLinkId: {},
    };
}

function cancelTextSearch (needToRemoveSearchInputVal) {
    cancelTextSearchHighlights();

    hideTextSearchUI();

    if (needToRemoveSearchInputVal) {
        $searchBarInput.val('');
    }

    currentTextSearchContext = {};

    resizeWrappersHeight();
}

function highlightAllTextSearchResults (text) {
    if (isTextSearchInProgress()) {
        for (let i = 0; i < currentTextSearchContext.searchResultMessagesContextArr.length; i++) {
            const nextMsgSearchContext = currentTextSearchContext.searchResultMessagesContextArr[i];
            const $messageBlock = findMessageBlockById(nextMsgSearchContext.id);

            if (nextMsgSearchContext.matchesInMessageIdText.length) {
                highlightAllTextOccurrencesInBlockDim(text, $messageBlock.find('.room-msg-id'));
            }
            if (nextMsgSearchContext.matchesInAuthorName.length) {
                highlightAllTextOccurrencesInBlockDim(text, $messageBlock.find('.room-msg-name-val'));
            }
            if (nextMsgSearchContext.matchesInMessageTimeText.length) {
                highlightAllTextOccurrencesInBlockDim(text, $messageBlock.find('.room-msg-name-time-part'));
            }
            if (nextMsgSearchContext.matchesInReplyToMessageIdText.length) {
                highlightAllTextOccurrencesInBlockDim(text, $messageBlock.find('.reply-msg-to-id'));
            }
            if (nextMsgSearchContext.matchesInReplyToUserName.length) {
                highlightAllTextOccurrencesInBlockDim(text, $messageBlock.find('.message-reply-to-user-name'));
            }
            if (nextMsgSearchContext.matchesInReplyToMessageText.length) {
                highlightAllTextOccurrencesInBlockDim(text, $messageBlock.find('.original-message-short-text'));
            }
            if (nextMsgSearchContext.matchesInText.length) {
                highlightAllTextOccurrencesInBlockDim(text, $messageBlock.find('.room-msg-text-inner'), true);
            }
        }
    }
}

function cancelTextSearchHighlights () {
    if (isTextSearchInProgress()) {
        for (let i = 0; i < currentTextSearchContext.searchResultMessagesContextArr.length; i++) {
            const nextMsg = currentTextSearchContext.searchResultMessagesContextArr[i];

            const $nextMsgBlock = findMessageBlockById(nextMsg.id);

            cancelTextSearchHighlightsForBlock($nextMsgBlock);
        }
    }
}

function isTextSearchInProgress () {
    return !!currentTextSearchContext.searchResultMessagesContextArr;
}

function cancelTextSearchHighlightsForBlock ($messageBlock) {
    $messageBlock.removeClass('search-msg-selected');
    deHighlightTextInBlock($messageBlock.find('.room-msg-id'));
    deHighlightTextInBlock($messageBlock.find('.room-msg-name-val'));
    deHighlightTextInBlock($messageBlock.find('.room-msg-name-time-part'));
    deHighlightTextInBlock($messageBlock.find('.reply-msg-to-id'));
    deHighlightTextInBlock($messageBlock.find('.message-reply-to-user-name'));
    deHighlightTextInBlock($messageBlock.find('.original-message-short-text'));
    deHighlightTextInBlock($messageBlock.find('.room-msg-text-inner'), true);
}

function findPrevActiveAreaForSearch (messageSearchContext, oldArea) {
    //at least 1 match (search result) will always be available after search started

    //if current area passed - pick next one that has search results, if available (circle ends on unconditional return null at TEXT_SEARCH_AREA_MSG_ID)
    if (oldArea) {
        if (oldArea === TEXT_SEARCH_AREA_TEXT) {
            if (messageSearchContext.matchesInReplyToMessageText.length) {
                return TEXT_SEARCH_AREA_R_MESSAGE;
            } else if (messageSearchContext.matchesInReplyToUserName.length) {
                return TEXT_SEARCH_AREA_R_USER;
            } else if (messageSearchContext.matchesInReplyToMessageIdText.length) {
                return TEXT_SEARCH_AREA_R_MESSAGE_ID;
            } else if (messageSearchContext.matchesInMessageTimeText.length) {
                return TEXT_SEARCH_AREA_MSG_TIME;
            } else if (messageSearchContext.matchesInAuthorName.length) {
                return TEXT_SEARCH_AREA_MSG_AUTHOR;
            } else if (messageSearchContext.matchesInMessageIdText.length) {
                return TEXT_SEARCH_AREA_MSG_ID;
            } else {
                return null;
            }

        } else if (oldArea === TEXT_SEARCH_AREA_R_MESSAGE) {
            if (messageSearchContext.matchesInReplyToUserName.length) {
                return TEXT_SEARCH_AREA_R_USER;
            } else if (messageSearchContext.matchesInReplyToMessageIdText.length) {
                return TEXT_SEARCH_AREA_R_MESSAGE_ID;
            } else if (messageSearchContext.matchesInMessageTimeText.length) {
                return TEXT_SEARCH_AREA_MSG_TIME;
            } else if (messageSearchContext.matchesInAuthorName.length) {
                return TEXT_SEARCH_AREA_MSG_AUTHOR;
            } else if (messageSearchContext.matchesInMessageIdText.length) {
                return TEXT_SEARCH_AREA_MSG_ID;
            } else {
                return null;
            }

        } else if (oldArea === TEXT_SEARCH_AREA_R_USER) {
            if (messageSearchContext.matchesInReplyToMessageIdText.length) {
                return TEXT_SEARCH_AREA_R_MESSAGE_ID;
            } else if (messageSearchContext.matchesInMessageTimeText.length) {
                return TEXT_SEARCH_AREA_MSG_TIME;
            } else if (messageSearchContext.matchesInAuthorName.length) {
                return TEXT_SEARCH_AREA_MSG_AUTHOR;
            } else if (messageSearchContext.matchesInMessageIdText.length) {
                return TEXT_SEARCH_AREA_MSG_ID;
            } else {
                return null;
            }

        } else if (oldArea === TEXT_SEARCH_AREA_R_MESSAGE_ID) {
            if (messageSearchContext.matchesInMessageTimeText.length) {
                return TEXT_SEARCH_AREA_MSG_TIME;
            } else if (messageSearchContext.matchesInAuthorName.length) {
                return TEXT_SEARCH_AREA_MSG_AUTHOR;
            } else if (messageSearchContext.matchesInMessageIdText.length) {
                return TEXT_SEARCH_AREA_MSG_ID;
            } else {
                return null;
            }

        } else if (oldArea === TEXT_SEARCH_AREA_MSG_TIME) {
            if (messageSearchContext.matchesInAuthorName.length) {
                return TEXT_SEARCH_AREA_MSG_AUTHOR;
            } else if (messageSearchContext.matchesInMessageIdText.length) {
                return TEXT_SEARCH_AREA_MSG_ID;
            } else {
                return null;
            }

        } else if (oldArea === TEXT_SEARCH_AREA_MSG_AUTHOR) {
            if (messageSearchContext.matchesInMessageIdText.length) {
                return TEXT_SEARCH_AREA_MSG_ID;
            } else {
                return null;
            }

        } else if (oldArea === TEXT_SEARCH_AREA_MSG_ID) {
            return null;

        } else {
            //should not happen
            return TEXT_SEARCH_AREA_TEXT;
        }

    } else {
        //return 1st field down the priority circle, that has search results
        if (messageSearchContext.matchesInText.length) {
            return TEXT_SEARCH_AREA_TEXT;
        } else if (messageSearchContext.matchesInReplyToMessageText.length) {
            return TEXT_SEARCH_AREA_R_MESSAGE;
        } else if (messageSearchContext.matchesInReplyToUserName.length) {
            return TEXT_SEARCH_AREA_R_USER;
        } else if (messageSearchContext.matchesInReplyToMessageIdText.length) {
            return TEXT_SEARCH_AREA_R_MESSAGE_ID;
        } else if (messageSearchContext.matchesInMessageTimeText.length) {
            return TEXT_SEARCH_AREA_MSG_TIME;
        } else if (messageSearchContext.matchesInAuthorName.length) {
            return TEXT_SEARCH_AREA_MSG_AUTHOR;
        } else {
            return TEXT_SEARCH_AREA_MSG_ID;
        }
    }
}

function findNextActiveAreaForSearch (messageSearchContext, oldArea) {
    if (oldArea) {
        if (oldArea === TEXT_SEARCH_AREA_MSG_ID) {
            if (messageSearchContext.matchesInAuthorName.length) {
                return TEXT_SEARCH_AREA_MSG_AUTHOR;
            } else if (messageSearchContext.matchesInMessageTimeText.length) {
                return TEXT_SEARCH_AREA_MSG_TIME;
            } else if (messageSearchContext.matchesInReplyToMessageIdText.length) {
                return TEXT_SEARCH_AREA_R_MESSAGE_ID;
            } else if (messageSearchContext.matchesInReplyToUserName.length) {
                return TEXT_SEARCH_AREA_R_USER;
            } else if (messageSearchContext.matchesInReplyToMessageText.length) {
                return TEXT_SEARCH_AREA_R_MESSAGE;
            } else if (messageSearchContext.matchesInText.length) {
                return TEXT_SEARCH_AREA_TEXT;
            } else {
                return null;
            }

        } else if (oldArea === TEXT_SEARCH_AREA_MSG_AUTHOR) {
            if (messageSearchContext.matchesInMessageTimeText.length) {
                return TEXT_SEARCH_AREA_MSG_TIME;
            } else if (messageSearchContext.matchesInReplyToMessageIdText.length) {
                return TEXT_SEARCH_AREA_R_MESSAGE_ID;
            }  else if (messageSearchContext.matchesInReplyToUserName.length) {
                return TEXT_SEARCH_AREA_R_USER;
            } else if (messageSearchContext.matchesInReplyToMessageText.length) {
                return TEXT_SEARCH_AREA_R_MESSAGE;
            } else if (messageSearchContext.matchesInText.length) {
                return TEXT_SEARCH_AREA_TEXT;
            } else {
                return null;
            }

        } else if (oldArea === TEXT_SEARCH_AREA_MSG_TIME) {
            if (messageSearchContext.matchesInReplyToMessageIdText.length) {
                return TEXT_SEARCH_AREA_R_MESSAGE_ID;
            } else if (messageSearchContext.matchesInReplyToUserName.length) {
                return TEXT_SEARCH_AREA_R_USER;
            } else if (messageSearchContext.matchesInReplyToMessageText.length) {
                return TEXT_SEARCH_AREA_R_MESSAGE;
            } else if (messageSearchContext.matchesInText.length) {
                return TEXT_SEARCH_AREA_TEXT;
            } else {
                return null;
            }

        } else if (oldArea === TEXT_SEARCH_AREA_R_MESSAGE_ID) {
            if (messageSearchContext.matchesInReplyToUserName.length) {
                return TEXT_SEARCH_AREA_R_USER;
            } else if (messageSearchContext.matchesInReplyToMessageText.length) {
                return TEXT_SEARCH_AREA_R_MESSAGE;
            } else if (messageSearchContext.matchesInText.length) {
                return TEXT_SEARCH_AREA_TEXT;
            } else {
                return null;
            }

        } else if (oldArea === TEXT_SEARCH_AREA_R_USER) {
            if (messageSearchContext.matchesInReplyToMessageText.length) {
                return TEXT_SEARCH_AREA_R_MESSAGE;
            } else if (messageSearchContext.matchesInText.length) {
                return TEXT_SEARCH_AREA_TEXT;
            } else {
                return null;
            }

        } else if (oldArea === TEXT_SEARCH_AREA_R_MESSAGE) {
            if (messageSearchContext.matchesInText.length) {
                return TEXT_SEARCH_AREA_TEXT;
            } else {
                return null;
            }

        } else if (oldArea === TEXT_SEARCH_AREA_TEXT) {
            return null;

        }  else {
            //should not happen
            return TEXT_SEARCH_AREA_TEXT;
        }

    } else {
        if (messageSearchContext.matchesInMessageIdText.length) {
            return TEXT_SEARCH_AREA_MSG_ID;
        } else if (messageSearchContext.matchesInAuthorName.length) {
            return TEXT_SEARCH_AREA_MSG_AUTHOR;
        } else if (messageSearchContext.matchesInMessageTimeText.length) {
            return TEXT_SEARCH_AREA_MSG_TIME;
        } else if (messageSearchContext.matchesInReplyToMessageIdText.length) {
            return TEXT_SEARCH_AREA_R_MESSAGE_ID;
        } else if (messageSearchContext.matchesInReplyToUserName.length) {
            return TEXT_SEARCH_AREA_R_USER;
        } else if (messageSearchContext.matchesInReplyToMessageText.length) {
            return TEXT_SEARCH_AREA_R_MESSAGE;
        } else {
            return TEXT_SEARCH_AREA_TEXT;
        }
    }
}

function findAreaSearchResultInitialArrayIndex (msgSearchContext, isPrev) {
    if (!isPrev) {
        return 0;
    }

    if (msgSearchContext.activeSearchArea === TEXT_SEARCH_AREA_MSG_ID) {
        return msgSearchContext.matchesInMessageIdText.length - 1;

    } else if (msgSearchContext.activeSearchArea === TEXT_SEARCH_AREA_MSG_AUTHOR) {
        return msgSearchContext.matchesInAuthorName.length - 1;

    } else if (msgSearchContext.activeSearchArea === TEXT_SEARCH_AREA_MSG_TIME) {
        return msgSearchContext.matchesInMessageTimeText.length - 1;

    } else if (msgSearchContext.activeSearchArea === TEXT_SEARCH_AREA_R_MESSAGE_ID) {
        return msgSearchContext.matchesInReplyToMessageIdText.length - 1;

    } else if (msgSearchContext.activeSearchArea === TEXT_SEARCH_AREA_R_USER) {
        return msgSearchContext.matchesInReplyToUserName.length - 1;

    } else if (msgSearchContext.activeSearchArea === TEXT_SEARCH_AREA_R_MESSAGE) {
        return msgSearchContext.matchesInReplyToMessageText.length - 1;

    } else {
        //text
        return msgSearchContext.matchesInText.length - 1;
    }
}

function findActiveAreaBlockInsideMessageBlock (activeArea, $messageBlock) {
    if (activeArea === TEXT_SEARCH_AREA_MSG_ID) {
        return $messageBlock.find('.room-msg-id');

    } else if (activeArea === TEXT_SEARCH_AREA_MSG_AUTHOR) {
        return $messageBlock.find('.room-msg-name-val');

    } else if (activeArea === TEXT_SEARCH_AREA_MSG_TIME) {
        return $messageBlock.find('.room-msg-name-time-part');

    } else if (activeArea === TEXT_SEARCH_AREA_R_MESSAGE_ID) {
        return $messageBlock.find('.reply-msg-to-id');

    } else if (activeArea === TEXT_SEARCH_AREA_R_USER) {
        return $messageBlock.find('.message-reply-to-user-name');

    } else if (activeArea === TEXT_SEARCH_AREA_R_MESSAGE) {
        return $messageBlock.find('.original-message-short-text');

    } else {
        return $messageBlock.find('.room-msg-text-inner');
    }
}

function getOnlineUsersCountString () {
    const onlineUsersCount = Object.values(lastUserInfoByIdCache)
        .filter(user => !!user.isOnlineInRoom).length;

    const allUsersCount = Object.values(lastUserInfoByIdCache).length;

    return onlineUsersCount === allUsersCount
        ? onlineUsersCount + ''
        : onlineUsersCount + ' (' + allUsersCount + ')';
}

function onChangeUserName (e) {
    stopPropagationAndDefault(e);

    changeUserName();

    const isChatAtBottom = isChatScrolledBottom();

    hideUserNameChangingBlock();
    $userJoinedAsNameWr.removeClass('d-none');
    $userJoinedAsTitle.removeClass('d-none');
    resizeWrappersHeight(isChatAtBottom);

    hideGlobalTransparentOverlay();
}

function onChangeRoomDescription (e) {
    stopPropagationAndDefault(e);

    hideRoomDescriptionEditingBlock();

    hideGlobalTransparentOverlay();

    changeRoomDescription();

    resizeWrappersHeight();
}

/* User bots */

function addUserBotToCluster (botId, matchStr, matchUser, payloadStr, isEnabled) {
    clientBotsCluster[botId] = {
        isEnabled: isEnabled,
        matchStr: matchStr ? matchStr.trim() : null,
        matchUser: matchUser ? matchUser.trim() : null,
        matchStrPayload: payloadStr ? payloadStr.trim() : null,

        vars: {
            counter: 0,
        },

        onMessage: onRoomMessageBotAction
    };
}

function deleteUserBotFromCluster (botId) {
    delete clientBotsCluster[botId];
    LOCAL_STORAGE.setItem(BOTS_LIST_LOCAL_STORAGE_KEY, JSON.stringify(clientBotsCluster));
}

function onRoomMessageBotAction (botId, messageInfo) {
    const bot = clientBotsCluster[botId];
    const matchStr = bot.matchStr;
    const matchUser = bot.matchUser;

    let count = bot.vars.counter;

    let matchStrPayload = bot.matchStrPayload;

    if (
        matchStrPayload
        && (
            (matchStr && messageInfo.text.toLowerCase().includes(matchStr.toLowerCase().trim()))
            || (matchUser && messageInfo.authorName && messageInfo.authorName.toLowerCase() === matchUser.toLowerCase().trim())
        )
    ) {
        count = ++bot.vars.counter;

        matchStrPayload = matchStrPayload.replace("${autor_name}", (messageInfo.authorName ? messageInfo.authorName : 'Unknown'));
        matchStrPayload = matchStrPayload.replace("${matches_count}", count + '');

        $userMessageTextarea.val(
            "[BOT]: " + matchStrPayload
        );

        sendUserMessage();
    }
}

function onBotConfigChangedDelayedFunc (e) {
    clearTimeout(onBotConfigChangedTimeout);
    onBotConfigChangedTimeout = setTimeout(function () {
        onBotConfigChanged(e);
    }, 100);
}

function onBotConfigChanged (e) {
    const $botBlock = $(e.currentTarget).closest('.bots-list-item-wr');
    const botId = $botBlock.attr('data-bot-id');

    addUserBotToCluster(
        botId,
        $botBlock.find('.bots-item-keyphrase').val(),
        $botBlock.find('.bots-item-keyusername').val(),
        $botBlock.find('.bots-item-template').val(),
        $botBlock.find('.bots-list-item-enabled').prop('checked'),
    );

    LOCAL_STORAGE.setItem(BOTS_LIST_LOCAL_STORAGE_KEY, JSON.stringify(clientBotsCluster));
}

function onBotDeleteClick (e) {
    const $botBlock = $(e.currentTarget).closest('.bots-list-item-wr');
    const botId = $botBlock.attr('data-bot-id');

    deleteUserBotFromCluster(botId);

    $botBlock.remove();
}

function clearTextSelection () {
    if (window.getSelection) {
        if (window.getSelection().empty) {  // Chrome
            window.getSelection().empty();
        } else if (window.getSelection().removeAllRanges) {  // Firefox
            window.getSelection().removeAllRanges();
        }
    } else if (document.selection) {  // IE
        document.selection.empty();
    }
}
