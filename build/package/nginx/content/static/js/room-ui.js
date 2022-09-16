/* Constants */

const DESKTOP_FOLK_PICKS_SCREEN_MIN_WIDTH_PX = 1200;
const DESKTOP_ROOM_INFO_SHOW_SCREEN_MIN_HEIGHT_PX = 900;
const MOBILE_CONTAINER_MAIN_LARGE_PADDING_STARTS_FROM_WIDTH_PX = 768;
const NAVBAR_COMPACT_MODE_MAX_HEIGHT_PX = 500;

//on mobiles, calculating current scroll state of block is unstable, so this margin allows some safety
const CHAT_SCROLL_MEASUREMENTS_MARGIN = 5;

const SELECTION_CONTROLS_ADAPTION_OFFSET_FROM_MSG_TOP_PX = 2;

const ROOM_TITLE_LINK_CHANGE_INPUT_TOUCHEND_DELAY = 200;

const UI_THEME_LIGHT = 'light';
const UI_THEME_DARK  = 'dark';

/* Variables */

/* DOM objects references */

const $mainSidebarWr = $('.main-sidebar-wr');

const $navbarMainWr = $('.navbar-brand-wr');
const $mainNavbarBlock = $('.main-navbar-page-room');
const $navbarRoomAndLogoWr = $('.navbar-room-and-logo-wr');

const $roomTitleLinks = $('.room-title-link');
const $roomTitleLinksRefreshLink = $('.room-title-link-refresh-link');
const $roomTitleLinkNames = $('.room-title-link-name');
const $roomTitleContentWr = $('.room-title-content-wr');

const $roomTitleLinkChangeInput = $roomTitleContentWr.find('.room-title-link-change-input');
const $roomTitleLinkChangeLinks = $('.room-title-link-change-link');

const $topNotifBlock = $('.top-notif-msg');

const $messageContextMenu = $('.message-context-menu');
const $messageContextMenuEditButton = $messageContextMenu.find('#message-tooltip-edit');
const $messageContextMenuDeleteButton = $messageContextMenu.find('#message-tooltip-delete');
const $messageContextMenuCopyButton = $messageContextMenu.find('#message-tooltip-copy');
const $messageContextMenuSelectButton = $messageContextMenu.find('#message-tooltip-select');
const $messageContextMenuGotoButton = $messageContextMenu.find('#message-tooltip-goto');
const $messageContextMenuRespondMessageButton = $messageContextMenu.find('#message-tooltip-respond-message');
const $messageContextMenuRespondUserButton = $messageContextMenu.find('#message-tooltip-respond-user');
const $messageContextMenuDownloadDrawingButton = $messageContextMenu.find('#message-tooltip-download-image');

const $globalTransparentOverlay = $('.global-transparent-overlay');

const $spinnerOverlayVersionChangedWr = $('.build-version-changed-wr');
const $spinnerOverlayRoomUserDuplicationWr = $('.room-user-duplication-wr');

const $containerMain = $('.container-main');
const $containerMainCenterWr = $('.container-main-center-wr');

const $mobileTextSelectionButtonsBlock = $('.mobile-text-selection-buttons');
const $mobileCopyButtonBlock = $('.mobile-text-copy-button');
const $mobileCancelButtonBlock = $('.mobile-text-cancel-button');
const $mobileTextSelectionToMessageOverlayTopBlock = $('.mobile-text-selection-scroll-overlay-top');
const $mobileTextSelectionToMessageOverlayBotBlock = $('.mobile-text-selection-scroll-overlay-bot');

const $mobileCopyButtonIcon = $('.mobile-text-copy-button-icon');
const $mobileCopyButtonIconGreyed = $('.mobile-text-copy-button-icon-greyed');


//folk picks block exists as single one for desktop and mobile, and gets transformed on screen size change
const $folkPicksWr = $('.picks-wr');
const $folkPicksTitle = $('.picks-wr-title');
const $folkPicksMessagesWr = $('.picks-messages-wr');
const $folkPicksBotWr = $('.picks-bot-wr');

const $roomInfoBlock = $('.room-info-block');
const $roomInfoWr = $('.room-info-wr');
const $roomInfoOnlineUsers = $('.room-info-online-users');
const $roomInfoWrMainNavbar = $('.room-info-wr-main-navbar');
const $roomInfoCollapseWr = $('.room-info-addit-bot-collapse-wr');
const $roomInfoCollapseShareImg = $('.room-info-addit-bot-collapse-share-img');
const $roomInfoCreationTime = $('.room-info-addit-creation-time');
const $roomInfoUsersCount = $('.room-info-addit-users-count');
const $roomInfoLeaveRoomBtn = $('.room-info-addit-mid-right-buttons-leave');
const $roomInfoShareRoomBtn = $('.room-info-addit-share-btn');

const $roomInfoDescription = $('.room-descr');
const $roomInfoDescriptionText = $('.room-descr-text');
const $roomInfoDescriptionEmptyWr = $('.room-descr-text-empty-wr');
const $roomInfoDescriptionEmptyText = $('.room-descr-text-empty');
const $roomInfoDescriptionEmptyTextCreator = $('.room-descr-text-empty-creator');

const $roomInfoDescriptionCreatorChangeWr = $('.room-descr-creator-change-wr');
const $roomInfoDescriptionCreatorChangeInput = $('#room-descr-change-input');
const $roomInfoDescriptionCreatorChangeLink = $('#room-descr-change-link');

const $roomMessagesWr = $('.room-messages-wr');
const $chatScrollOverlay = $('.chat-scroll-overlay');

const $userMessageWr = $('.user-message-wr');
const $userMessageContentWr = $userMessageWr.find('.user-message-content-wr');
const $userMessageErrorBlock = $userMessageWr.find('.user-message-wr-error-msg');
const $userMessageNotifBlock = $userMessageWr.find('.user-message-wr-notif-msg');
const $userMessageToggleMenu = $userMessageWr.find('.user-message-wr-toggle-mobile-menu');
const $userMessageToggleFolkPicks = $userMessageWr.find('.user-message-wr-toggle-folk-picks');
const $userMessageTextarea = $userMessageWr.find('.user-message-textarea');

const $userMessageCollapseWr = $userMessageWr.find('.user-message-wr-collapse-wr');
const $userMessageCollapseWrSendingIndication = $userMessageWr.find('.user-message-wr-collapse-sending-indication');

const $toggleSearchButton = $('.user-message-wr-toggle-search');
const $searchBarWr = $('.search-bar-wr');
const $searchBarInput = $('.search-bar-input');
const $searchBarPrevButton = $('.search-bar-prev-btn');
const $searchBarNextButton = $('.search-bar-next-btn');

const $answerToUserWr = $('.answer-to-user-wr');
const $answerToUserCloseButton = $('.answer-to-user-close');

const $answerToMessageWr = $('.answer-to-message-wr');
const $answerToMessageCloseButton = $('.answer-to-message-close');

const $sendMessageButton = $('.user-message-send-btn');

const $userJoinedAsNameWr = $('.user-message-joined-as-block');
const $userJoinedAsTitle = $('.user-message-joined-as-title');

const $userJoinedAsAnonPref = $('.user-message-joined-as-block-anon-pref');
const $userJoinedAsName = $('.user-message-joined-as-name');

const $userJoinedAsChangeWr = $('.user-message-joined-as-change-wr');
const $userJoinedAsChangeInput = $('#user-message-joined-as-change-input');
const $userJoinedAsChangeLink = $('#user-message-joined-as-change-link');

const $messageEditCancelButton = $('.message-edit-cancel');

const $shareRoomWr = $('.share-room-wr');
const $shareRoomBtn = $('.share-room-btn');
const $shareRoomAlert = $('.share-room-copied-alert');
const $shareRoomManualURL = $('.share-room-manual-url');
const $shareRoomHasPassword = $('.share-room-has-password');

const $roomWelcomeMessageUsername = $('.room-messages-init-msg-username');

const $folkPicksEmptyMessagesWr = $('.page-room-empty-picks-wr');

const $drawingFullSizeViewImgBlock = $('#drawing-full-size-view');

//smiles
const $toggleSmilesButton = $('.user-message-wr-toggle-smiles');

const $smilesMainWr = $('.smiles-main-wr');
const $smilesListShortWr = $('.smiles-list-wr-short');
const $smilesListFullWr = $('.smiles-list-wr-full');
const $smilesCtrlShowFullFlag = $('#smiles-ctrl-show-full');

const $toggleBotsButton = $('.user-message-wr-toggle-bots');
const $botsWr = $('.bots-main-wr');
const $botsListWr = $('#bots-list-wr');
const $botsWrClose = $('.bots-wr-close');
const $botsShowExampleBtn = $('.bots-list-example-show');
const $botsAddNewBtn = $('.bots-add-new-bnt');
const $botsExampleWr = $('.bot-example-wr');
const $botsEmptyBotForCloningBlock = $('.empty-bot-dom-for-cloning');


/* vars */

let isMobileClientDevice;
let onResizeTimeout;
let onBotConfigChangedTimeout;
let onMessagesWrapperScrollTimeout;

let selectedUITheme = UI_THEME_LIGHT;

let topNotificationsQueue = [];

let notificationHideTimeout = -1;
let topNotificationHideTimeout = -1;

let needToFocusRoomChangeInput = false;

let isNavbarInBigMode = true;

let roomTitleLinkChangeInputLastTouchEndAt = 0;

//store DOM elements for period when they must be detached from DOM
let mainNavbarChildrenBlocks;
let roomInfoChildrenBlocks;
let userMessageWrChildrenBlocks;

/* visibility flags */

let menuHiddenDesktop = false;
let folkPicksHiddenDesktop = false;

let isInMenuMobileMode = false;
let isInFolkPicksMobileMode = false;


function initRoomPageUI () {
    commonPageUIInit();
    storeCurrentViewportDimensions();

    /* initial page setup */

    //replace address bar value in case of direct room access by url with spaces
    window.history.replaceState("str", document.title, window.location.href.replace(/(%20)+/g, '-'));

    isMobileClientDevice = isMobileDevice();

    //set formatted room name to array of room links.room-msg-id
    $roomTitleLinkNames.text(ROOM_NAME);

    /* Set UI preferences */

    const folkPicksHiddenSettingStr = LOCAL_STORAGE.getItem(FOLK_PICKS_HIDDEN_LOCAL_STORAGE_KEY);

    if (!folkPicksHiddenSettingStr) {
        LOCAL_STORAGE.setItem(FOLK_PICKS_HIDDEN_LOCAL_STORAGE_KEY, folkPicksHiddenDesktop);
    } else {
        folkPicksHiddenDesktop = folkPicksHiddenSettingStr === 'true';
    }

    const menuHiddenSettingStr = LOCAL_STORAGE.getItem(MENU_HIDDEN_LOCAL_STORAGE_KEY);

    if (!menuHiddenSettingStr) {
        LOCAL_STORAGE.setItem(MENU_HIDDEN_LOCAL_STORAGE_KEY, menuHiddenDesktop);
    } else {
        menuHiddenDesktop = menuHiddenSettingStr === 'true';
    }

    const showFullSmilesListSettingStr = LOCAL_STORAGE.getItem(SHOW_FULL_SMILES_LIST_LOCAL_STORAGE_KEY);

    if (!showFullSmilesListSettingStr) {
        LOCAL_STORAGE.setItem(SHOW_FULL_SMILES_LIST_LOCAL_STORAGE_KEY, false);
    } else {
        if (showFullSmilesListSettingStr === 'true') {
            toggleFullSmilesList(true);
        } else {
            toggleFullSmilesList(false);
        }
    }

    $smilesCtrlShowFullFlag.on('change', function () {
        toggleFullSmilesList($(this).prop('checked'));
    });

    //send notification if non-standard chars in room name
    for (let i = 0; i < ROOM_NAME.length; i++) {
        let nextCh = ROOM_NAME.charAt(i);

        if (!isCharASCII(nextCh) && !allowedRoomNameSpecialChars[nextCh]) {
            showTopNotification(NOTIFICATION_TEXT_USE_SHARE_BTN, TOP_NOTIFICATION_SHOW_MS * 3, true);

            break;
        }
    }

    /* Resize blocks height (e.g. room messages wrapper to fill area between top and bottom neighbours) */
    $window.on("load", onResizeActions);
    $window.on('resize', onResizeDelayedFunc);

    $mainSidebarWr.on('contextmenu', function (e) {
        stopPropagationAndDefault(e);
    });
    $folkPicksWr.on('contextmenu', function (e) {
        const $clickedBlock = $(e.target);

        //if clicked on link - do nothing
        if (!$clickedBlock.hasClass('message-highlight-link')) {
            stopPropagationAndDefault(e);
        }

        $messageContextMenu.addClass('d-none');
    });
    $folkPicksWr.on('click', function () {
        //normally if context menu is opened - there will be global overlay to cancel it. But for mobile folk picks
        //it is impossible to show global overlay as mobile folk picks block is already position absolute
        if (!$messageContextMenu.hasClass('d-none')) {
            $messageContextMenu.addClass('d-none');
        }
    });
    $userMessageCollapseWr.on('contextmenu', function (e) {
        stopPropagationAndDefault(e);
    });

    $containerMainCenterWr.on('contextmenu', function (e) {
        const $clickedBlock = $(e.target);

        const $messageBlock = $clickedBlock.closest('.room-msg-main-wr');

        //if clicked on link OR click inside message block - do nothing, allow to propagate
        if ($clickedBlock.hasClass('message-highlight-link') || $messageBlock.length) {

            return;
        }

        const mainHeaderHeightPx = $mainHeader.css('display') !== 'none' ? $mainHeader.outerHeight(true) : 0;
        const roomInfoCollapseWrBottomEdgePositionPx =
            mainHeaderHeightPx
            + $roomInfoBlock.outerHeight(true);

        const userMessageWrHeightPx = $userMessageWr.outerHeight(true);

        //click coordinates inside element
        const clickCoordsInsideElementPair = getClickCoordinatesInsideElement($body[0], e);
        //if click happened near 'messages' block (not above or below) - prevent context menu
        if (clickCoordsInsideElementPair.y > roomInfoCollapseWrBottomEdgePositionPx
            && clickCoordsInsideElementPair.y < (currentViewportHeight - userMessageWrHeightPx)) {
            e.preventDefault();
        }
    });

    /* Set visibility behaviour for Room Info block */

    $roomInfoCollapseWr.on('contextmenu', function (e) {
        stopPropagationAndDefault(e);
    });

    $roomInfoCollapseWr.on('click', toggleRoomInfoBlock);
    $roomTitleLinks.on('click', toggleRoomInfoBlock);

    $roomTitleLinksRefreshLink.on('click', function () {
        reloadPage();
    });

    //show or hide room info block
    if (isMobileClientDevice) {
        //hide Room Info on mobile by default
        $roomInfoWr.css('display', 'none');
        $roomInfoCollapseShareImg.removeClass('d-none');
        $roomTitleLinks.addClass('room-title-link-padded-right');

        resizeWrappersHeight(true);

    } else {
        //hide Room Info on desktop if height is low
        if (currentViewportHeight < DESKTOP_ROOM_INFO_SHOW_SCREEN_MIN_HEIGHT_PX) {
            $roomInfoWr.css('display', 'none');
            $roomInfoCollapseShareImg.removeClass('d-none');
            $roomTitleLinks.addClass('room-title-link-padded-right');
        } else {
            $roomTitleLinks.removeClass('room-title-link-padded-right');

            showRoomNameChangeInput();
        }

        $roomInfoCollapseWr.hover(
            function () {
                if ($roomInfoWr.css('display') !== 'none') {
                    $roomInfoWr.css('opacity', '0.6');
                }
            },
            function () {
                // on mouseout - reset
                if ($roomInfoWr.css('display') !== 'none') {
                    $roomInfoWr.fadeTo( "fast" , 1);
                }
            }
        );
    }


    /* Set user message UI behaviour */

    $userMessageErrorBlock.on('mousedown', function () {
        if (!$userMessageErrorBlock.isFading) {
            $userMessageErrorBlock.isFading = true;
            $userMessageErrorBlock.toggle('fade', function () {
                $userMessageErrorBlock.isFading = false;
            });
        }
    });

    $userMessageNotifBlock.on('mousedown', function () {
        if (!$userMessageNotifBlock.isFading) {
            $userMessageNotifBlock.isFading = true;
            $userMessageNotifBlock.toggle('fade', function () {
                $userMessageNotifBlock.isFading = false;
            });
        }
    });

    $topNotifBlock.on('mousedown', function () {
        if (!$topNotifBlock.isFading) {
            $topNotifBlock.isFading = true;
            $topNotifBlock.toggle('fade', function () {
                $topNotifBlock.isFading = false;

                setTimeout(showNextQueuedTopNotification, 500);
            });
        }
    });

    //if applicable - show 'zoom' top notification
    if (!isMobileClientDevice) {
        const zoomNotificationLastShownAt = LOCAL_STORAGE.getItem(ZOOM_NOTIF_LAST_SHOWN_AT_LOCAL_STORAGE_KEY);
        const currentTimeMills = new Date().getTime();

        if (
            !zoomNotificationLastShownAt
            || currentTimeMills > (parseInt(zoomNotificationLastShownAt) + (MILLS_IN_HOUR * 72))
        ) {
            setTimeout(function () {
                LOCAL_STORAGE.setItem(ZOOM_NOTIF_LAST_SHOWN_AT_LOCAL_STORAGE_KEY, currentTimeMills);

                showTopNotification(NOTIFICATION_TEXT_ZOOM, TOP_NOTIFICATION_SHOW_MS * 3, false);
            }, 20000);
        }
    }

    if (!isMobileClientDevice) {
        $userMessageCollapseWr.hover(
            function () {
                if ($userMessageWr.css('display') !== 'none') {
                    $userMessageWr.css('opacity', '0.6');
                }
            },
            function () {
                // on mouseout - reset
                if ($userMessageWr.css('display') !== 'none') {
                    $userMessageWr.fadeTo("fast", 1);
                }
            }
        );
    }

    $userMessageCollapseWr.on('click', toggleUserMessageBlock);


    /* Set Menu and Folk Picks blocks toggle behaviour */

    $userMessageToggleMenu.on('mousedown', toggleMenu);

    $userMessageToggleFolkPicks.on('mousedown', function () {
        cancelMessageTextSelectionMode();

        if (currentViewportWidth < DESKTOP_FOLK_PICKS_SCREEN_MIN_WIDTH_PX) {
            showFolkPicksMobile();
        } else {
            if (!folkPicksHiddenDesktop) {
                folkPicksHiddenDesktop = true;
                LOCAL_STORAGE.setItem(FOLK_PICKS_HIDDEN_LOCAL_STORAGE_KEY, folkPicksHiddenDesktop);

                hideFolkPicksDesktop();
            } else {
                folkPicksHiddenDesktop = false;
                LOCAL_STORAGE.setItem(FOLK_PICKS_HIDDEN_LOCAL_STORAGE_KEY, folkPicksHiddenDesktop);

                showFolkPicksDesktop();
            }
        }
    });

    bindMainOverlay();


    /* Set Recent rooms toggle behaviour */

    $recentRoomsLink.on('click', function () {
        hideMenuMobile();
        hideFolkPicksMobile();
        showMyRecentRoomsPopup();
    });

    $recentRoomsClose.on('click', function () {
        hideMyRecentRoomsPopup();
    });

    $botsWrClose.on('click', function () {
        hideBotsUI();
    });

    //bind mobile swipe to menu/folk picks opening
    const xwiper = new Xwiper(document);
    xwiper.onSwipeRight(function () {
        if (
            !isSwipedOnRoomChangeInput()
            && isOnMobileScreenSize()
            && !messageTextSelectionInProgressId
            && !isAnyPopupOpened()
        ) {
            $messageContextMenu.addClass('d-none');

            hideMyRecentRoomsPopup();
            hideFolkPicksMobile();

            if ($spinnerOverlay.hasClass('d-none')) {
                showMenuMobileInRoom();
            }
        }
    });
    xwiper.onSwipeLeft(function () {
        if (
            !isSwipedOnRoomChangeInput()
            && currentViewportWidth < DESKTOP_FOLK_PICKS_SCREEN_MIN_WIDTH_PX
            && !messageTextSelectionInProgressId
            && !isAnyPopupOpened())
         {
            $messageContextMenu.addClass('d-none');

            hideMyRecentRoomsPopup();
            hideMenuMobile();

            showFolkPicksMobile();
        }
    });

    /* Set Messages UI behaviour */

    $chatScrollOverlay.on('mousedown', function () {
        scrollChatBottom();
    });

    $roomMessagesWr.on('scroll', function () {
        tryShowChatScrollOverlay();

        if (!isMobileClientDevice) {
            adaptDesktopTextSelectionControlsDelayedFunc();
        }
    });

    $roomMessagesWr.on('contextmenu', function (e) {
        const $clickedBlock = $(e.target);

        const $messageBlock = $clickedBlock.closest('.room-msg-main-wr');

        //if clicked NOT on link and click happened NOT inside message block - prevent it
        if (!$clickedBlock.hasClass('message-highlight-link') && !$messageBlock.length) {
            stopPropagationAndDefault(e);
        }
    });

    $globalTransparentOverlay.on('click contextmenu', function (e) {
        const isChatAtBottom = isChatScrolledBottom();

        $globalTransparentOverlay.addClass('d-none');

        cancelActionsUnderGlobalTransparentOverlay();

        resizeWrappersHeight(isChatAtBottom);

        stopPropagationAndDefault(e);
        return false;
    });

    /* Setup share behaviour */

    $roomInfoShareRoomBtn.on('click', function () {
        if (isMobileClientDevice && navigator.share) {
            navigator.share({
                title: ROOM_NAME + ' - Instantchat',
                url: $shareRoomManualURL.text()

            }).catch(function () {
                $shareRoomWr.css('display', 'block');
                showMainOverlay();
            });
        } else {
            $shareRoomWr.css('display', 'block');
            showMainOverlay();
        }
    });

    $shareRoomBtn.on('click', copyRoomFastLink);

    $roomInfoDescriptionCreatorChangeInput.on('focusin', function () {
        $roomInfoDescription.css('border-color', '#3b3b3b');
    });

    $roomInfoDescriptionCreatorChangeInput.on('focusout', function () {
        $roomInfoDescription.css('border-color', '#FCFCFC');
    });

    $toggleSearchButton.on('mousedown', toggleTextSearchUI);

    $toggleBotsButton.on('mousedown', toggleBotsUI);

    $toggleSmilesButton.on('mousedown', toggleSmilesUI);

    $toggleUserInputDrawingButton.on('mousedown', toggleUserInputDrawingUI);

    $mobileTextSelectionToMessageOverlayTopBlock.on('click', function () {
        const $messageBlock = findMessageBlockById(messageTextSelectionInProgressId);

        if ($messageBlock) {
            scrollToMessageDuringTextSelection($messageBlock[0]);
        }
    });
    $mobileTextSelectionToMessageOverlayBotBlock.on('click', function () {
        const $messageBlock = findMessageBlockById(messageTextSelectionInProgressId);

        if ($messageBlock) {
            scrollToMessageDuringTextSelection($messageBlock[0]);
        }
    });

    $spinnerOverlayVersionChangedWr.on('click', function () {
        reloadPage();
    });
    $spinnerOverlayRoomUserDuplicationWr.on('click', function () {
        reloadPage();
    });

    $roomTitleLinkChangeInput.val(ROOM_NAME);

    $roomTitleLinkChangeLinks.on('click', navigateViaRoomTitleChangeInput);

    $roomTitleLinkChangeInput.on('keydown', function(e) {
        e.stopPropagation();

        if (e.which === KEY_CODE_ENTER) {
            navigateViaRoomTitleChangeInput();
        }
    });

    $roomTitleLinkChangeInput.on('click', function(e) {
        needToFocusRoomChangeInput = true;

        e.stopPropagation();
    });

    $roomTitleLinkChangeInput.on('touchend', function(e) {
        roomTitleLinkChangeInputLastTouchEndAt = new Date().getTime();
    });

    $roomTitleLinkChangeInput.on('input', changeRoomTitleRoomChangeControls);

    $smilesListShortWr.on('click', onSmileClick);
    $smilesListFullWr.on('click', onSmileClick);

    if (isMobileClientDevice) {
        $drawingMainWr.addClass('user-drawing-input-wr-mobile');
    }

    initUIColorTheme();

    onResizeActions();
}

function navigateViaRoomTitleChangeInput () {
    const newRoomName = $roomTitleLinkChangeInput.val().trim();

    if (newRoomName !== ROOM_NAME && validateRoomName(newRoomName)) {
        redirectToURL("/" + newRoomName);
    }
}

function showUserNameChangingBlock() {
    cancelMessageEdit();

    $userJoinedAsNameWr.addClass('d-none');
    $userJoinedAsTitle.addClass('d-none');

    $userJoinedAsChangeWr.removeClass('d-none');
    $userJoinedAsChangeInput.focus();

    showGlobalTransparentOverlay();

    resizeWrappersHeight(true);
}

function hideUserNameChangingBlock() {
    $userJoinedAsChangeWr.addClass('d-none');
    setUserNameChangingInputToCurrentValue();
}

function tryShowChatScrollOverlay () {
    if (isChatScrolledBottom()) {
        $chatScrollOverlay.addClass('d-none');
    } else {
        $chatScrollOverlay.removeClass('d-none');
    }
}

function redrawScreenSizeDependentSideBlocks () {
    if (isOnMobileScreenSize()) {
        if (!isInMenuMobileMode) {
            //transition from desktop to mobile
            //hide desktop sidebar
            $sidebarWr.removeClass('d-md-block');
            $navbarMainWr.removeClass('d-lg-block');

            $navbarRoomAndLogoWr.removeClass('col-lg-10');
            $navbarRoomAndLogoWr.addClass('col-lg-12');

            $mainWr.removeClass('col-md-10').addClass('col-md-12');
        }

        isInMenuMobileMode = true;
    } else {
        if (isInMenuMobileMode) {
            //transition from mobile to desktop
            hideMenuMobile();
        }

        //now in desktop mode - show or hide sidebar menu block respective to its desktop flag
        if (menuHiddenDesktop) {
            $sidebarWr.removeClass('d-md-block');
            $navbarMainWr.removeClass('d-lg-block');

            $navbarRoomAndLogoWr.removeClass('col-lg-10');
            $navbarRoomAndLogoWr.addClass('col-lg-12');

            $mainWr.removeClass('col-md-10').addClass('col-md-12');
        } else {
            //show desktop sidebar
            $sidebarWr.addClass('d-md-block');

            $navbarRoomAndLogoWr.removeClass('col-lg-12');
            $navbarRoomAndLogoWr.addClass('col-lg-10');

            $navbarMainWr.addClass('d-lg-block');

            $mainWr.removeClass('col-md-12').addClass('col-md-10');
        }

        isInMenuMobileMode = false;
    }

    if (currentViewportWidth < DESKTOP_FOLK_PICKS_SCREEN_MIN_WIDTH_PX) {
        if (!isInFolkPicksMobileMode) {
            //transition from desktop to mobile
            //hide desktop folk picks
            $folkPicksWr.css('display', 'none');
            folkPicksTransformMobile();
        }

        isInFolkPicksMobileMode = true;
    } else {
        if (isInFolkPicksMobileMode) {
            //transition from mobile to desktop
            hideFolkPicksMobile();
            folkPicksTransformDesktop();
        }

        //now in desktop mode - show or hide folk picks block respective to its desktop flag
        if (folkPicksHiddenDesktop) {
            $folkPicksWr.css('display', 'none');
            $containerMainCenterWr.removeClass('col-xl-9').addClass('col-xl-12');
        } else {
            //show desktop folk picks
            $folkPicksWr.css('display', 'block');
            $containerMainCenterWr.removeClass('col-xl-12').addClass('col-xl-9');
        }

        isInFolkPicksMobileMode = false;
    }

    //set visibility for one of navbar variants
    if (currentViewportHeight > NAVBAR_COMPACT_MODE_MAX_HEIGHT_PX) {
        if (!isNavbarInBigMode) {
            changeNavbarMode(true);
        }
    } else {
        if (isNavbarInBigMode) {
            changeNavbarMode(false);
        }
    }

    if (needToFocusRoomChangeInput) {
        needToFocusRoomChangeInput = false;

        $roomTitleLinkChangeInput.focus();
    }
}

function changeNavbarMode (isBig) {
    if (isBig) {
        isNavbarInBigMode = true;

        $roomInfoBlock.addClass('room-info-block-navbar-mode-big');

        $mainNavbarBlock.removeClass('d-none');
        $roomInfoWrMainNavbar.addClass('d-none');

        $mainNavbarBlock.find('.room-title-link').append($roomTitleLinkChangeInput);
    } else {
        isNavbarInBigMode = false;

        $roomInfoBlock.removeClass('room-info-block-navbar-mode-big');

        $mainNavbarBlock.addClass('d-none');
        $roomInfoWrMainNavbar.removeClass('d-none');

        $roomInfoWrMainNavbar.find('.room-title-link').append($roomTitleLinkChangeInput);
    }
}

function toggleMenu () {
    cancelMessageTextSelectionMode();

    if (isOnMobileScreenSize()) {
        if ($spinnerOverlay.hasClass('d-none')) {
            showMenuMobileInRoom();
        }
    } else {
        if (!menuHiddenDesktop) {
            menuHiddenDesktop = true;
            LOCAL_STORAGE.setItem(MENU_HIDDEN_LOCAL_STORAGE_KEY, menuHiddenDesktop);

            hideMenuDesktop();
        } else {
            menuHiddenDesktop = false;
            LOCAL_STORAGE.setItem(MENU_HIDDEN_LOCAL_STORAGE_KEY, menuHiddenDesktop);

            showMenuDesktop();
        }
    }
}

function showMenuDesktop () {
    $sidebarWr.hide("slide", { duration: 0, direction: "left", complete: function () {
            $sidebarWr.addClass('d-md-block');
            $mainWr.removeClass('col-md-12').addClass('col-md-10');

            $navbarRoomAndLogoWr.removeClass('col-lg-12');
            $navbarRoomAndLogoWr.addClass('col-lg-10');

            $navbarMainWr.addClass('d-lg-block');

            resizeWrappersHeight();
        }}, 0);
}

function hideMenuDesktop () {
    $sidebarWr.hide("slide", { duration: 200, direction: "left", complete: function () {
            $sidebarWr.removeClass('d-md-block');
            $mainWr.removeClass('col-md-10').addClass('col-md-12');
            $navbarMainWr.removeClass('d-lg-block');

            $navbarRoomAndLogoWr.removeClass('col-lg-10');
            $navbarRoomAndLogoWr.addClass('col-lg-12');

            resizeWrappersHeight();
            animateBlockForAttention($userMessageToggleMenu, true, 700);
        }}, 200);
}

function showFolkPicksDesktop () {
    $folkPicksWr.show("slide", { duration: 0, direction: "right", complete: function () {
            $containerMainCenterWr.removeClass('col-xl-12').addClass('col-xl-9');
            resizeWrappersHeight();
        }}, 0);
}

function hideFolkPicksDesktop () {
    $folkPicksWr.hide("slide", { duration: 200, direction: "right", complete: function () {
            $containerMainCenterWr.removeClass('col-xl-9').addClass('col-xl-12');
            resizeWrappersHeight();
            animateBlockForAttention($userMessageToggleFolkPicks, true, 700);
        }}, 200);
}

function showFolkPicksMobile() {
    if ($spinnerOverlay.hasClass('d-none')) {
        $folkPicksWr.css('display', 'block');
        showMainOverlay();

        hideMobileKeyboard();

        resizeWrappersHeight();
    }
}

function showMenuMobileInRoom() {
     hideMobileKeyboard();

     showMenuMobile();
}

function hideFolkPicksMobile() {
    if (isInFolkPicksMobileMode && $folkPicksWr.css('display') !== 'none') {
        $folkPicksWr.css('display', 'none');
        hideMainOverlay();
    }
}

function folkPicksTransformMobile () {
    $containerMainCenterWr.removeClass('col-xl-9').addClass('col-xl-12');

    $folkPicksWr.removeClass('col-xl-3');
    $folkPicksWr.css('position', 'fixed');
    $folkPicksWr.css('border-radius', '2px');
    $folkPicksWr.css('width', '75%');
    $folkPicksWr.css('z-index', '120');
}

function folkPicksTransformDesktop () {
    $containerMainCenterWr.removeClass('col-xl-12').addClass('col-xl-9');

    $folkPicksWr.addClass('col-xl-3');
    $folkPicksWr.css('position', 'inherit');
    $folkPicksWr.css('border-radius', '0');
    $folkPicksWr.css('width', 'auto');
    $folkPicksWr.css('z-index', 'initial');
}

function scrollChatBottom () {
    $roomMessagesWr.scrollTop(999999999); //just some max_value, because scrollHeight behaves weird on mobiles
}

/* Resize blocks height (e.g. room messages wrapper to fill area between top and bottom neighbours) */
function resizeWrappersHeight (needToScrollChatToBottom) {
    //during message text selection on mobile - user message wrapper will be hidden, so here - just remove manual height from message wrappers
    if (messageTextSelectionInProgressId && isMobileClientDevice) {
        $roomMessagesWr.css('height', '');
        $folkPicksMessagesWr.css('height', '');

        return;
    }

    const mainHeaderHeightPx = $mainHeader.css('display') !== 'none' ? $mainHeader.outerHeight(true) : 0;

    //resize Room Messages wrapper
    const roomInfoCollapseWrBottomEdgePositionPx =
        mainHeaderHeightPx
        + $roomInfoBlock.outerHeight(true);

    const userMessageWrHeightPx = $userMessageWr.outerHeight(true);

    const roomMessagesWrHeightPx =
        Math.floor((currentViewportHeight - roomInfoCollapseWrBottomEdgePositionPx - userMessageWrHeightPx));

    $roomMessagesWr.css('height', roomMessagesWrHeightPx + 'px');

    //resize Folk Picks Messages wrapper
    const folkPicksTitleBottomEdgePositionPx =
        ($mainNavbarBlock.css('display') !== 'none' ? $mainNavbarBlock.outerHeight(true) : 0)
        + $folkPicksTitle.outerHeight(true);

    const folkPicksBotWrHeightPx = $folkPicksBotWr.css('display') !== 'none' ? $folkPicksBotWr.outerHeight(true) : 0;

    const $folkPicksMessagesWrHeightPx =
        Math.floor((currentViewportHeight - folkPicksTitleBottomEdgePositionPx - folkPicksBotWrHeightPx));

    $folkPicksMessagesWr.css('height', $folkPicksMessagesWrHeightPx + 'px');

    if (needToScrollChatToBottom) {
        scrollChatBottom();
    }
}

function copyRoomFastLink () {
    const roomUrl = $shareRoomManualURL.text();

    copyStringToClipboard(roomUrl,
        function () {
            $shareRoomAlert.css('display', 'block');

            $shareRoomAlert.animate({backgroundColor: "#fff"}, 300, function () {
                $shareRoomAlert.animate({backgroundColor: "#487d58"}, 300);
            });
        },
        function () {
            showError(ERROR_NOTIFICATION_TEXT_FAILED_COPY_URL_TO_CLIPBOARD);
        }
    );
}

function copyStringToClipboard (strToCopy, successCallback, errorCallback) {
    const textarea = document.createElement("textarea");

    textarea.style.position = "fixed";  // Prevent scrolling to bottom of page in Microsoft Edge.
    textarea.textContent = strToCopy;

    document.body.appendChild(textarea);
    textarea.focus();
    textarea.select();

    try {
        if (!document.execCommand('copy')) {
            if (errorCallback) {
                errorCallback();
            }
        } else {
            if (successCallback) {
                successCallback();
            }
        }
    } catch (err) {
        if (errorCallback) {
            errorCallback();
        }
    } finally {
        document.body.removeChild(textarea);
    }
}

function toggleRoomInfoBlock(e) {
    stopPropagationAndDefault(e);

    cancelMessageTextSelectionMode();
    scrollChatBottom();

    //hide user message block until animation is complete
    $userMessageWr.css('visibility', 'hidden');
    $roomInfoWrMainNavbar.css('visibility', 'hidden');

    $roomInfoWr.slideToggle('fast', function () {
        // Animation complete
        $userMessageWr.css('visibility', 'visible');
        $roomInfoWrMainNavbar.css('visibility', 'visible');

        //if current toggling means collapsing
        if ($roomInfoWr.css('display') === 'none') {
            $roomInfoCollapseShareImg.removeClass('d-none');

            hideRoomNameChangeInput();
        } else {
            if (isMobileClientDevice) {
                changeNavbarMode(currentViewportHeight > NAVBAR_COMPACT_MODE_MAX_HEIGHT_PX);
            }

            $roomInfoCollapseShareImg.addClass('d-none');

            showRoomNameChangeInput();
        }

        setTimeout(function () {
            resizeWrappersHeight(true);
        }, 50);
    });
}

function toggleUserMessageBlock() {
    cancelMessageTextSelectionMode();
    scrollChatBottom();

    $userMessageContentWr.slideToggle('fast', function () {
        resizeWrappersHeight(true);
    });
}

function showError(message, notificationShowTimeMs) {
    $userMessageErrorBlock.text(message);
    $userMessageErrorBlock.css('background-color', '#fff');
    $userMessageErrorBlock.css('display', 'block');
    $userMessageErrorBlock.animate({backgroundColor: "#ff5e5e"}, 500);

    if (notificationShowTimeMs) {
        setTimeout(function () {
            if ($userMessageErrorBlock.css('display') === 'block' && !$userMessageErrorBlock.isFading) {
                $userMessageErrorBlock.isFading = true;
                $userMessageErrorBlock.toggle('fade', function () {
                    $userMessageErrorBlock.isFading = false;
                });
            }
        }, notificationShowTimeMs);
    }
}

function showNotification(message, notificationShowTimeMs) {
    clearTimeout(notificationHideTimeout);

    $userMessageNotifBlock.text(message);
    $userMessageNotifBlock.css('background-color', '#fff');
    $userMessageNotifBlock.css('display', 'block');
    $userMessageNotifBlock.animate({backgroundColor: "#22b14c"}, 500);

    if (notificationShowTimeMs) {
        notificationHideTimeout = setTimeout(function () {
            if ($userMessageNotifBlock.css('display') === 'block' && !$userMessageNotifBlock.isFading) {
                $userMessageNotifBlock.isFading = true;
                $userMessageNotifBlock.toggle('fade', function () {
                    $userMessageNotifBlock.isFading = false;
                });
            }
        }, notificationShowTimeMs);
    }
}

/*
* 2nd notification type, drawn on top of room title - for less important events
* */
function showTopNotification(message, notificationShowTimeMs, showNow) {
    if ($topNotifBlock.css('display') === 'block' && !showNow) {
        topNotificationsQueue.push({message: message, notificationShowTimeMs: notificationShowTimeMs});

        return;
    }

    clearTimeout(topNotificationHideTimeout);

    $topNotifBlock.text(message);
    $topNotifBlock.css('background-color', '#fff');
    $topNotifBlock.css('display', 'block');
    $topNotifBlock.animate({backgroundColor: "#4b7d5a"}, 500);

    if (notificationShowTimeMs) {
        topNotificationHideTimeout = setTimeout(function () {
            $topNotifBlock.css('display', 'none');

            setTimeout(showNextQueuedTopNotification, 500);

        }, notificationShowTimeMs);
    }
}

function onResizeDelayedFunc () {
    clearTimeout(onResizeTimeout);
    onResizeTimeout = setTimeout(onResizeActions, 50);
}

function showNextQueuedTopNotification () {
    if (topNotificationsQueue.length) {
        const nextNotifInfo = topNotificationsQueue[0];
        topNotificationsQueue.shift();

        showTopNotification(nextNotifInfo.message, nextNotifInfo.notificationShowTimeMs, false);
    }
}

function onResizeActions () {
    storeCurrentViewportDimensions();

    adjustUserDrawingInputBlockTopCoord();

    const isChatAtBottom = isChatScrolledBottom();

    if (!$messageContextMenu.hasClass('d-none')) {
        $messageContextMenu.addClass('d-none');
    }

    if (isMobileClientDevice) {
        moveMobileTextSelectionControlsToMessage();
    }

    redrawScreenSizeDependentSideBlocks();
    resizeWrappersHeight(isChatAtBottom);

    tryShowChatScrollOverlay();
}

function createTextMessageDom(textMessage, userName, isAnon, originalMessageShortText) {
    const messageId = textMessage.id;
    const messageUserId = textMessage.uId;

    const $messageMainWrBlock = $('<div class="room-msg-main-wr box-shadow">');
    $messageMainWrBlock.attr('data-msg-id', messageId);
    $messageMainWrBlock.attr('data-user-id', messageUserId);

    if (isMobileClientDevice) {
        $messageMainWrBlock.addClass('noselect');
    }

    const $messageWrBlock = $('<div class="room-msg-wr">');
    $messageWrBlock.append('<div class="message-marks-wr">');


    /* Message id / name / time */

    const $messageNameWrBlock = $('<div class="room-msg-name-wr">');

    const $messageIdWrBlock = $('<p class="room-msg-id-wr font-weight-bold">');

    const $messageIdBlock = $('<span class="room-msg-id">');
    $messageIdBlock.text('#' + messageId);

    const $messageNameAnonBlock = $('<span class="room-msg-name-anon font-weight-bold">anon&nbsp;</span>');
    if (!isAnon) {
        $messageNameAnonBlock.addClass('d-none');
    }

    const $messageNameBlock = $('<p class="room-msg-name font-weight-bold">');
    const $messageNameValBlock = $('<span class="room-msg-name-val">');
    $messageNameValBlock.text(userName);

    $messageNameBlock.append($messageNameValBlock);

    $messageIdWrBlock.append($messageIdBlock);
    $messageIdWrBlock.append($messageNameAnonBlock);
    $messageIdWrBlock.append($messageNameBlock);
    $messageIdWrBlock.append($('<div class="clearfix"></div>'));

    const $messageNameTimeBlock = $('<div class="room-msg-name-time">');
    const msgTimeInMills = textMessage.cAt * 1000;

    $messageNameTimeBlock.append(
        $('<span class="room-msg-name-time-date-part">' + getDateStrFromMills(msgTimeInMills) + ', </span>'),
        $('<span class="room-msg-name-time-part">' + getTimeStrFromMills(msgTimeInMills) + '</span>'),
    );

    if (textMessage.lE) {
        $messageNameTimeBlock.append($('<span class="message-edited-title">(edited)</span>'));
    }

    $messageNameWrBlock.append($messageIdWrBlock);
    $messageNameWrBlock.append($messageNameTimeBlock);


    /* Message body */

    const $messageBodyBlock = $('<div class="room-msg-body">');
    const $messageTextBlock = $('<p class="room-msg-text card-text">');

    //if message has reply to message or user
    if (textMessage.rM) {
        appendReplyToMessageBlock($messageTextBlock, textMessage.rM, textMessage.rU, textMessage.id, originalMessageShortText);
        $messageMainWrBlock.attr('data-reply-to-msg-id', textMessage.rM);
        $messageMainWrBlock.attr('data-reply-to-user-id', textMessage.rU);

    } else if (textMessage.rU) {
        appendReplyToUserBlock($messageTextBlock, textMessage.rU, textMessage.id);
        $messageMainWrBlock.attr('data-reply-to-user-id', textMessage.rU);
    }

    const $messageTextInnerBlock = $('<span class="room-msg-text-inner">');
    $messageTextInnerBlock.text(decodeURIComponent(textMessage.t));

    $messageTextBlock.append($messageTextInnerBlock);

    //voting buttons
    const $messageButtonsBlock = $('<div class="room-msg-buttons">');
    const $messageButtonsSupportBlock = $('<a href="javascript:void(0);" class="room-msg-buttons-support font-weight-bold">');
    $messageButtonsSupportBlock
        .append($('<span class="room-msg-buttons-support-text">support&nbsp;</span>'))
        .append($('<span class="room-msg-buttons-support-text-short">S:&nbsp;</span>'))
        .append($('<span class="room-msg-buttons-support-val">' + textMessage.sC + '</span>'));
    const $messageButtonsRejectBlock = $('<a href="javascript:void(0);" class="room-msg-buttons-reject font-weight-bold">');
    $messageButtonsRejectBlock
        .append($('<span class="room-msg-buttons-reject-text">reject&nbsp;</span>'))
        .append($('<span class="room-msg-buttons-reject-text-short">R:&nbsp;</span>'))
        .append($('<span class="room-msg-buttons-reject-val">' + textMessage.rC + '</span>'));

    $messageButtonsBlock.append($messageButtonsSupportBlock);
    $messageButtonsBlock.append($messageButtonsRejectBlock);

    $messageBodyBlock.append($messageTextBlock);
    $messageBodyBlock.append($messageButtonsBlock);
    $messageBodyBlock.append($('<div class="clearfix"></div>'));

    $messageWrBlock.append($messageNameWrBlock);
    $messageWrBlock.append($messageBodyBlock);
    $messageWrBlock.append($('<div class="clearfix"></div>'));

    $messageMainWrBlock.append($messageWrBlock);

    return $messageMainWrBlock;
}

function setRoomInfo (createdAtMills, usersCount) {
    $roomInfoCreationTime.text(getHoursMinutesStrFromMills(createdAtMills));
    $roomInfoUsersCount.text(usersCount);
}

function hideShareRoomPopup() {
    if ($shareRoomWr.css('display') !== 'none') {
        $shareRoomWr.css('display', 'none');
        hideMainOverlay();
    }

    $shareRoomAlert.css('display', 'none');
}

function isChatScrolledBottom () {
    return Math.abs(($roomMessagesWr[0].scrollHeight - $roomMessagesWr.scrollTop()) - $roomMessagesWr.outerHeight())
        < CHAT_SCROLL_MEASUREMENTS_MARGIN
}

function showGlobalTransparentOverlay() {
    $globalTransparentOverlay.removeClass('d-none');
}

function hideGlobalTransparentOverlay() {
    $globalTransparentOverlay.addClass('d-none');
}

function showRoomDescriptionEditingBlock () {
    $roomInfoDescriptionCreatorChangeWr.removeClass('d-none');
    $roomInfoDescriptionCreatorChangeInput.focus();
    resizeWrappersHeight(true);

    showGlobalTransparentOverlay();
}

function hideRoomDescriptionEditingBlock () {
    if (!$roomInfoDescriptionCreatorChangeWr.hasClass('d-none')) {
        $roomInfoDescriptionCreatorChangeWr.addClass('d-none');
    }
}

function hideMobileKeyboard() {
    $userMessageTextarea.focusout();
    document.activeElement.blur();
    $("input").blur();
}

function sortMainMessages () {
    $roomMessagesWr.find('.room-msg-main-wr').sort(function(a, b) {
        return +a.dataset.msgId - +b.dataset.msgId;
    })
        .appendTo($roomMessagesWr);
}

function sortFolkPics () {
    $folkPicksMessagesWr.find('.room-msg-main-wr').sort(function(a, b) {
        return +a.dataset.msgId - +b.dataset.msgId;
    })
        .appendTo($folkPicksMessagesWr);
}

function cancelActionsUnderGlobalTransparentOverlay () {
    //cancel context menu
    if (!$messageContextMenu.hasClass('d-none')) {
        $messageContextMenu.addClass('d-none');
    }

    //cancel user name editing block
    if (!$userJoinedAsChangeWr.hasClass('d-none')) {
        hideUserNameChangingBlock();
        $userJoinedAsNameWr.removeClass('d-none');
        $userJoinedAsTitle.removeClass('d-none');
    }

    //cancel room description editing block
    hideRoomDescriptionEditingBlock();

    hideDrawingFullSizeView();

    //hide smiles block
    hideSmilesUI();
}

function showDrawingFullSizeView(imgBase64Content) {
    $drawingFullSizeViewImgBlock.attr('src', imgBase64Content);
    $drawingFullSizeViewImgBlock.removeClass('d-none');
}

function hideDrawingFullSizeView() {
    if (!$drawingFullSizeViewImgBlock.hasClass('d-none')) {
        $drawingFullSizeViewImgBlock.addClass('d-none');
    }
}

function toggleSmilesUI () {
    if ($smilesMainWr.hasClass('d-none')) {
        showSmilesUI();
    } else {
        hideSmilesUI();
    }
}

function showSmilesUI () {
    cancelMessageTextSelectionMode();

    $smilesMainWr.removeClass('d-none');
    $toggleSmilesButton.addClass('user-message-wr-toggle-smiles-active');

    showGlobalTransparentOverlay();
}

function hideSmilesUI () {
    $smilesMainWr.addClass('d-none');
    $toggleSmilesButton.removeClass('user-message-wr-toggle-smiles-active');

    hideGlobalTransparentOverlay();
}

function toggleFullSmilesList (showFull) {
    $smilesCtrlShowFullFlag.prop('checked', showFull);
    LOCAL_STORAGE.setItem(SHOW_FULL_SMILES_LIST_LOCAL_STORAGE_KEY, showFull);

    if (showFull) {
        $smilesMainWr.removeClass('smiles-mode-short').addClass('smiles-mode-full');
        $smilesListShortWr.addClass('d-none');
        $smilesListFullWr.removeClass('d-none');
    } else {
        $smilesMainWr.addClass('smiles-mode-short').removeClass('smiles-mode-full');
        $smilesListShortWr.removeClass('d-none');
        $smilesListFullWr.addClass('d-none');
    }
}

function onSmileClick(e) {
    const $clickedSmile = $(e.target);

    if ($clickedSmile.length && $clickedSmile.hasClass('smile-wr')) {
        $userMessageTextarea.val($userMessageTextarea.val() + $clickedSmile.text());

        stopPropagationAndDefault(e);
    }
}

function toggleTextSearchUI () {
    const isChatAtBottom = isChatScrolledBottom();

    if ($searchBarWr.hasClass('d-none')) {
        showTextSearchUI();
    } else {
        hideTextSearchUI();

        cancelTextSearch();
    }

    resizeWrappersHeight(isChatAtBottom);
}

function showTextSearchUI () {
    $searchBarWr.removeClass('d-none');

    $toggleSearchButton.addClass('user-message-wr-toggle-search-active');

    setTimeout(function () {
        $searchBarInput.focus();
    }, 50);
}

function hideTextSearchUI () {
    $searchBarWr.addClass('d-none');

    $toggleSearchButton.removeClass('user-message-wr-toggle-search-active');
}

function toggleBotsUI () {
    if ($botsWr.hasClass('d-none')) {
        showBotsUI();
    } else {
        hideBotsUI();
    }
}

function showBotsUI () {
    cancelMessageTextSelectionMode();

    showMainOverlay();

    swipingBehaviourAllowed = false;

    $botsWr.removeClass('d-none');
    $toggleBotsButton.addClass('user-message-wr-toggle-bots-active');
}

function hideBotsUI () {
    if ($botsWr.hasClass('d-none')) {
        return;
    }

    $botsWr.addClass('d-none');
    $toggleBotsButton.removeClass('user-message-wr-toggle-bots-active');

    swipingBehaviourAllowed = true;

    hideMainOverlay();
}

function highlightTextInBlockByOccurrenceIndexBright (text, textIdx, $messageBlock) {
    const elem = $messageBlock.find('.message-text-highlight-dim')
        .get(textIdx);

    if (elem) {
        const $elem = $(elem);
        $elem.removeClass('message-text-highlight-dim');
        $elem.addClass('message-text-highlight');
    }
}

function highlightAllTextOccurrencesInBlockDim (text, $textBlock, isInnerMessageTextBlock) {
    const thisMessageSearchFixLinkIds = [];

    if (isInnerMessageTextBlock) {
        for (let i = 0; i < $textBlock.children().length; i++) {
            const $nextChild = $($textBlock.children().get(i));

            if ($nextChild[0].tagName.toLowerCase() === 'a') {
                const $nextChildLink = $nextChild;
                const nextLinkId = currentTextSearchContext.nextSearchFixLinkId++;

                thisMessageSearchFixLinkIds.push(nextLinkId);

                currentTextSearchContext.linkOrigHrefByFixLinkId[nextLinkId] = $nextChildLink.attr('href');

                //if this link has preview loaded (i.e. it is an 'outer' link, which contains real link inside as 'inner' as well as other preview info)
                //- then swap outer link with inner
                if ($nextChildLink.hasClass('message-link-preview')) {
                    const $innerLink = $nextChildLink.find('.message-highlight-link');
                    $nextChildLink.replaceWith($innerLink);
                }
            }
        }
    }

    if (thisMessageSearchFixLinkIds.length) {
        $textBlock.attr('data-search-fix-link-ids', thisMessageSearchFixLinkIds.join(","));
    }

    //replace all text blocks inside $textBlock with just text until search is over - to be able to highlight searched text without issues with child tags
    $textBlock.text($textBlock.text());

    const innerHTML = $textBlock.html();

    $textBlock.html(
        innerHTML.replace(new RegExp(escapeRegExp(text), 'ig'), function(str) {
            const $highlightBlock = $("<span class='message-text-highlight-dim'></span>");
            $highlightBlock.text(str);

            return $highlightBlock.get(0).outerHTML;
        })
    );
}

function deHighlightTextInBlock ($textBlock, isInnerMessageTextBlock) {
    //replace all text blocks inside $textBlock with just text, to clear search highlighting
    $textBlock.text($textBlock.text());

    if (isInnerMessageTextBlock && $textBlock.attr('data-search-fix-link-ids')) {
        const thisMessageSearchFixLinkIds = $textBlock.attr('data-search-fix-link-ids')
            .split(',').map(function (val) {
                return parseInt(val, 10);
            });

        const allLinksInfos = thisMessageSearchFixLinkIds.map(function (linkSearchFixId) {
            return {text: currentTextSearchContext.linkOrigHrefByFixLinkId[linkSearchFixId]};
        });

        //turn (back) all links text occurrences from message into tags
        findLinksTextAndTurnIntoLinks(allLinksInfos, $textBlock, true);
    }
}

function turnTextIntoLink (textPlaceholder, text, $textBlock) {
    const innerHTML = $textBlock.html();

    $textBlock.html(
        innerHTML.replace(new RegExp(escapeRegExp(textPlaceholder), 'ig'), function() {
            const $highlightLinkBlock = $('<a class="message-highlight-link"></a>');
            $highlightLinkBlock.text(text);
            $highlightLinkBlock.attr('href', text);

            return $highlightLinkBlock.get(0).outerHTML;
        })
    );

    $textBlock.find('a.message-highlight-link').off('click').on('click', function (e) {
        stopPropagationAndDefault(e);
        window.open($(this).attr('href'), '_blank');
    });
}

function getOriginalMessageShortText (originalMessageId, allTextMessages) {
    let originalMessageShortText = MESSAGE_UNAVAILABLE_PLACEHOLDER_TEXT;

    //if this is an initial page load (allTextMessages is passed) - look for original message in allTextMessages. Else - look in DOM
    if (allTextMessages) {
        for (let i = 0; i < allTextMessages.length; i++) {
            const nextTextMessage = allTextMessages[i];

            if (nextTextMessage.id === originalMessageId) {
                originalMessageShortText = cutMessageTextForResponse(
                    decodeURIComponent(nextTextMessage.t)
                );

                break;
            }
        }
    } else {
        const $originalMessageBlock = findMessageBlockById(originalMessageId);
        if ($originalMessageBlock) {
            originalMessageShortText = cutMessageTextForResponse(
                $originalMessageBlock.find('.room-msg-text-inner').text()
            );
        }
    }

    return originalMessageShortText;
}


function adaptDesktopTextSelectionControlsDelayedFunc () {
    clearTimeout(onMessagesWrapperScrollTimeout);
    onMessagesWrapperScrollTimeout = setTimeout(adaptDesktopTextSelectionControls, 25);
}

function adaptDesktopTextSelectionControls() {
    //if any message is in text-selection mode - check if top part of message visible (not scrolled out of screen)
    //if at least some part is scrolled up and NOT visible - move text selection controls down
    if (messageTextSelectionInProgressId) {
        const $messageBlock = findMessageBlockById(messageTextSelectionInProgressId);
        if (!$messageBlock.length) {
            return;
        }

        const $textSelectionControlButtons = $messageBlock.find('.text-copy-button, .text-cancel-button');

        const mainHeaderHeightPx = $mainHeader.css('display') !== 'none' ? $mainHeader.outerHeight(true) : 0;

        //find on which height 'messages list wrapper' begins
        const roomInfoCollapseWrBottomEdgePositionPx =
            mainHeaderHeightPx
            + $roomInfoBlock.outerHeight(true);

        const clRect = $messageBlock[0].getBoundingClientRect();

        //find Y position of message block top edge relative to viewport top edge
        const messageBlockTopEdgeYCoord = clRect.top;
        //find Y position of message block bottom edge relative to viewport top edge
        const messageBlockBotEdgeYCoord = clRect.bottom;

        //subtract 'messages list wrapper' top edge position from message block top edge position, to get difference
        //(can be negative)
        const topEdgeYCoordRelativeToTopOfMessagesArea = messageBlockTopEdgeYCoord - roomInfoCollapseWrBottomEdgePositionPx;

        //subtract 'messages list wrapper' top edge position from message block bottom edge position, to get difference
        //(can be negative)
        const botEdgeYCoordRelativeToTopOfMessagesArea = messageBlockBotEdgeYCoord - roomInfoCollapseWrBottomEdgePositionPx;

        const selectionControlsHeightPx = 30;

        //if 'messages list wrapper' is scrolled so current message block is only partially visible:
        //i.e. true when message top edge is above 'messages list wrapper's top edge (or close to it), but message bot edge is still below
        if (topEdgeYCoordRelativeToTopOfMessagesArea < selectionControlsHeightPx &&
            botEdgeYCoordRelativeToTopOfMessagesArea > selectionControlsHeightPx) {
            //if block's top is still visible but we just dont have a room for controls on top of it - just set top=0 + some offset
            if (topEdgeYCoordRelativeToTopOfMessagesArea > 0) {
                $textSelectionControlButtons.css('top', SELECTION_CONTROLS_ADAPTION_OFFSET_FROM_MSG_TOP_PX + 'px');
            } else {
                $textSelectionControlButtons.css('top', (Math.abs(topEdgeYCoordRelativeToTopOfMessagesArea) + SELECTION_CONTROLS_ADAPTION_OFFSET_FROM_MSG_TOP_PX) + 'px');
            }
        } else {
            $textSelectionControlButtons.css('top', '');
        }
    }
}

function scrollToMessageDuringTextSelection (element) {
    $body[0].scrollTo(window.scrollX, element.offsetTop - 50);
}

function transformDesktopTextSelectionCopyButton (currentlySelectedText) {
    if (currentlySelectedText) {
        $('.text-copy-button').removeClass('text-copy-button-greyed');
    } else {
        $('.text-copy-button').addClass('text-copy-button-greyed');
    }
}

function transformMobileTextSelectionCopyButton (currentlySelectedText) {
    if (currentlySelectedText) {
        $mobileCopyButtonIcon.removeClass('d-none');
        $mobileCopyButtonIconGreyed.addClass('d-none');
    } else {
        $mobileCopyButtonIcon.addClass('d-none');
        $mobileCopyButtonIconGreyed.removeClass('d-none');
    }
}

function anyMessageIsFolkPicked () {
    return !!$folkPicksMessagesWr.children('.room-msg-main-wr').length;
}

function showRoomNameChangeInput () {
    $roomTitleLinks.off('click');

    $roomTitleContentWr.find('.room-title-link-name, .room-title-pref').addClass('d-none');
    $roomTitleLinkChangeInput.removeClass('d-none');
    $roomTitleLinkChangeLinks.removeClass('d-none');

    $roomTitleLinks.removeClass('room-title-link-padded-right');

    $roomTitleLinkChangeInput.val(ROOM_NAME);

    changeRoomTitleRoomChangeControls();
}

function hideRoomNameChangeInput () {
    $roomTitleLinks.on('click', toggleRoomInfoBlock);

    $roomTitleContentWr.find('.room-title-link-name, .room-title-pref').removeClass('d-none');
    $roomTitleLinkChangeInput.addClass('d-none');
    $roomTitleLinkChangeLinks.addClass('d-none');

    $roomTitleLinks.addClass('room-title-link-padded-right');

    $roomTitleLinkChangeInput.val(ROOM_NAME);
}

function changeRoomTitleRoomChangeControls () {
    if ($roomTitleLinkChangeInput.val().trim() === ROOM_NAME) {
        $roomTitleLinkChangeLinks.find('.room-title-link-change-img').addClass('d-none');
        $roomTitleLinkChangeLinks.find('.room-title-link-change-img-greyed').removeClass('d-none');
    } else {
        $roomTitleLinkChangeLinks.find('.room-title-link-change-img').removeClass('d-none');
        $roomTitleLinkChangeLinks.find('.room-title-link-change-img-greyed').addClass('d-none');
    }
}

function initUIColorTheme() {
    //light is default

    //TODO this is kind of POC for now
    if (selectedUITheme === UI_THEME_DARK) {
        $body.addClass('gm-elite');

        $(
            '.room-messages-wr, ' +
            '.room-messages-init-msg, ' +
            '.user-message-textarea-wr, ' +
            '.room-descr-text, ' +
            '.room-info-addit-mid-right-labels, ' +
            '.room-msg-name-val, ' +
            '.room-info-online-user-name, ' +
            '.main-navbar, ' +
            '.room-info-addit-mid-right-buttons-leave, ' +
            '.main-sidebar-top-item-1, ' +
            '.picks-wr, ' +
            '.picks-messages-wr, ' +
            '.picks-bot-wr, ' +
            '.main-sidebar-wr'
        ).addClass('gm-elite');
    }
}

function loadUrlPreview (linkURL, $textMessage, messageId, needToScrollChatToBottom) {
    getUrlPreviewInfo(linkURL, function (urlInfoResp) {
        if (!urlInfoResp) {
            return;
        }

        loadUrlPreviewIntoMessageBlock(urlInfoResp, $textMessage, needToScrollChatToBottom);

        //since this code is executed after async backend call returns - we expect that message block was already added to both main and folk-picks lists (if applicable).
        const $textMessageInFolkPicks = findFolkPicksMessageBlockById(messageId);
        if ($textMessageInFolkPicks) {
            loadUrlPreviewIntoMessageBlock(urlInfoResp, $textMessageInFolkPicks, false);
        }
    });
}

function loadUrlPreviewIntoMessageBlock (urlInfoResp, $textMessageBlock, needToScrollChatToBottom) {
    //1st highlighted link block is searched here again, because jquery seems to replace link node with its copy,
    //after parent's inner html is changed (which happens after getUrlPreviewInfo() is requested but before server responds)
    const $linkBlock = $($textMessageBlock.find('a.message-highlight-link').get(0));
    const $linkBlockInner = $linkBlock.clone();

    let $image = null;

    $linkBlockInner.off('click').on('click', function (e) {
        stopPropagationAndDefault(e);
        window.open($(this).attr('href'), '_blank');
    });

    //else - this is web page URL
    const imageURLBase64 = urlInfoResp.b64;

    const URL = urlInfoResp.u;
    const title = urlInfoResp.t;
    const description = urlInfoResp.d;
    const webPageLogoImageUrl = urlInfoResp.i;

    if (title || description || webPageLogoImageUrl || imageURLBase64) {
        $linkBlock.addClass('message-link-preview');

        $linkBlock.text('');
        $linkBlock.append($linkBlockInner);
    }

    if (title) {
        const $title = $('<h5 class="message-link-preview-title"></h5>');
        $title.text(title);

        $linkBlock.append($title);
    }

    if (webPageLogoImageUrl) {
        if (webPageLogoImageUrl.trim().startsWith('file')) {
            console.log('security issue: "file://..." passed from og:image for URL: ' + URL);

            return;
        }

        const $logoImage = $('<img class="message-link-preview-logo-img" />');
        $logoImage.attr('src', webPageLogoImageUrl);
        $logoImage.attr('title', (title ? title : ''));
        $logoImage.attr('alt', (title ? title : ''));

        $linkBlock.append($logoImage);

        $image = $logoImage;
    }

    if (imageURLBase64) {
        const $previewImage = $('<img class="message-link-preview-image" />');
        $previewImage.attr('src', imageURLBase64);

        $linkBlock.append($previewImage);

        $image = $previewImage;
    }

    if (needToScrollChatToBottom && $image) {
        $image.on('load', scrollChatBottom);
    }
}

function findLinksTextAndTurnIntoLinks (allLinksInfos, $textMessageBlock, messageId, needToScrollChatToBottom) {
    const firstLinkInRealOrder = allLinksInfos.length > 0 ? allLinksInfos[0] : null;
    //sort so in case of equal links, one of which is longer than other - longer one will be on top
    const allLinksInfosSorted = allLinksInfos.sort((a, b) => (a.text > b.text) ? -1 : 1);

    //iterate 2 times because in case of 2 equal links, one of which is just longer than other - we have to replace them separately

    //1st iteration - swap link text occurrences with placeholders
    for (let i = 0; i < allLinksInfosSorted.length; i++) {
        const linkInfo = allLinksInfosSorted[i];
        const innerHTML = $textMessageBlock.html();

        $textMessageBlock.html(
            innerHTML.replace(new RegExp(escapeRegExp(linkInfo.text.replaceAll('&', '&amp;')), 'ig'), 'link-placeholder-' + i)
        );
    }

    //2nd iteration - swap placeholders with links
    for (let i = 0; i < allLinksInfosSorted.length; i++) {
        const linkInfo = allLinksInfosSorted[i];

        turnTextIntoLink('link-placeholder-' + i, linkInfo.text, $textMessageBlock);
    }

    if (firstLinkInRealOrder) {
        setTimeout(function () {
            loadUrlPreview(firstLinkInRealOrder.text, $textMessageBlock, messageId, needToScrollChatToBottom);
        }, randomIntBetween(1, 200));
    }
}

function isAnyPopupOpened () {
    return !$drawingFullSizeViewImgBlock.hasClass('d-none')
        || !$botsWr.hasClass('d-none')
        || !$userJoinedAsChangeWr.hasClass('d-none')
        || !$roomInfoDescriptionCreatorChangeWr.hasClass('d-none')
    ;
}

function toggleBotExample () {
    if ($botsExampleWr.hasClass('d-none')) {
        showBotExample();
    } else {
        hideBotExample();
    }
}

function showBotExample () {
    $botsExampleWr.removeClass('d-none');
    $botsShowExampleBtn.text('hide bot example');
}

function hideBotExample () {
    $botsExampleWr.addClass('d-none');
    $botsShowExampleBtn.text('show bot example');
}

function isSwipedOnRoomChangeInput () {
    return new Date().getTime() <= (roomTitleLinkChangeInputLastTouchEndAt + ROOM_TITLE_LINK_CHANGE_INPUT_TOUCHEND_DELAY);
}

function showVersionChangedSpinner () {
    $spinnerOverlayVersionChangedWr.removeClass('d-none');
}

function unBindMainOverlay () {
    $mainOverlay.off('click');
}

function bindMainOverlay () {
    $mainOverlay.on('click', function () {
        $messageContextMenu.addClass('d-none');
        hideFolkPicksMobile();
        hideMenuMobile();
        hideMyRecentRoomsPopup();
        hideShareRoomPopup();
        hideUserInputDrawingBlock();
        hideBotsUI();
        //main overlay will be hidden as a result of one of above actions, so no need to hide it explicitly
    });
}
