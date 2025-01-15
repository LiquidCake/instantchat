
/* Variables */

//DOM references
const $roomNameInput = $("#room-name");
const $roomPasswdInput = $("#room-password");
const $roomUserNameInput = $("#room-user-name");

const $errorText = $("#room-error");

const $roomSuggestionsWr = $("#room-suggestions-wr");
const $roomSuggestionsTextInvalidPassword = $("#room-suggestions-text-invalid-password");
const $roomSuggestionsTextNoPassword = $("#room-suggestions-text-no-password");
const $roomSuggestions = $("#room-suggestions");


let roomNameTypingTimer = null;


function init () {
    keepAlive();

    $roomFormTitle.on('click', function () {
        if (window.chrome && window.chrome.webview) {
             window.chrome.webview.postMessage("change_window_mode_keyF1");
        }
    });

    //get redirect variable (if any)
    const redirectVariableStr = LOCAL_STORAGE.getItem(REDIRECT_VARIABLE_LOCAL_STORAGE_KEY);

    if (redirectVariableStr) {
        const redirectVariable = JSON.parse(redirectVariableStr);

        //check if redirect is from room page and if variable is not expired
        if (new Date().getTime() - redirectVariable.redirectedAt < REDIRECT_VARIABLE_TTL_MS) {

            if (redirectVariable.redirectFrom === REDIRECT_FROM_ROOM_PAGE_ON_ERROR) {
                switch (redirectVariable.error) {
                    case ROOM_TO_HOME_PG_REDIRECT_ERROR_GENERIC:
                        showErrorMessage(
                            ERROR_TEXT_REDIRECT_FROM_ROOM_PAGE_GENERIC + ": "
                            + (redirectVariable.errorDetails ? redirectVariable.errorDetails : "unknown error from room page")
                        );
                        break;

                    case ROOM_TO_HOME_PG_REDIRECT_ERROR_BAD_ROOM_NAME_LENGTH:
                        showErrorMessage(BUSINESS_ERRORS[ERROR_CODE_ROOM_INVALID_CREDS_LENGTH].text
                            + ", must be between " + ROOM_CREDS_MIN_LENGTH + " and " + ROOM_CREDS_MAX_LENGTH + " characters");
                        break;

                    case ROOM_TO_HOME_PG_REDIRECT_ERROR_ROOM_NAME_CONTAINS_SLASH:
                        showErrorMessage(ERROR_TEXT_ROOM_NAME_CONTAINS_SLASH);
                        break;

                    case ROOM_TO_HOME_PG_REDIRECT_ERROR_BUSINESS:
                        if (redirectVariable.errorBusiness.code === ERROR_CODE_ROOM_INVALID_PASSWORD) {
                            tryShowRoomSuggestions(
                                redirectVariable.errorBusiness.code,
                                redirectVariable.roomName,
                                redirectVariable.alternativeRoomNamePostfixes,
                                true
                            );
                        } else {
                            showErrorMessage(redirectVariable.errorBusiness.text);
                        }

                        break;

                    case ROOM_TO_HOME_PG_REDIRECT_ERROR_CONNECTION:
                        showErrorMessage(BUSINESS_ERRORS[ERROR_CODE_CONNECTION_ERROR].text);
                        break;

                    case ROOM_TO_HOME_PG_REDIRECT_ERROR_SESSION_COOKIE_MISSING:
                        showErrorMessage(ERROR_TEXT_SESSION_COOKIE_MISSING);
                        break;

                    case ROOM_TO_HOME_PG_REDIRECT_ERROR_CONNECTION_PICK_BACKEND:
                        showErrorMessage(
                            redirectVariable.errorDetails
                                ? redirectVariable.errorDetails
                                : BUSINESS_ERRORS[ERROR_CODE_CONNECTION_ERROR].text
                        );

                        break;
                    default:
                        showErrorMessage(
                            (redirectVariable.error ? redirectVariable.error + ": " : "")
                            + (redirectVariable.errorDetails ? redirectVariable.errorDetails : "unknown error from room page")
                        );
                }

            } else if (redirectVariable.redirectFrom === REDIRECT_FROM_ANY_PAGE_ON_RECENT_ROOM_CLICK) {
                $roomNameInput.val(redirectVariable.roomName);
                createOrJoinRoom();
            }
        }

        LOCAL_STORAGE.removeItem(REDIRECT_VARIABLE_LOCAL_STORAGE_KEY);
    }

    $roomNameInput.on('keydown', handleFormKeypress);
    $roomPasswdInput.on('keydown', handleFormKeypress);
    $roomUserNameInput.on('keydown', handleFormKeypress);

    //load visited rooms
    loadVisitedRooms(visitedRoomOnClickCallback);
}

function handleFormKeypress (e) {
    clearTimeout(roomNameTypingTimer);

    if (e.which === KEY_CODE_ENTER) {
        inPlaceFormatRoomNameInput();
        createOrJoinRoom();
    } else {
        roomNameTypingTimer = setTimeout(inPlaceFormatRoomNameInput, 500);
    }
}

function createOrJoinRoom () {
    if (roomJoinCreateBtnDisabled) {
        return;
    }

    inPlaceFormatRoomNameInput();

    $errorText.children().remove();
    $roomSuggestionsWr.addClass('d-none');
    $roomSuggestionsTextInvalidPassword.addClass('d-none');
    $roomSuggestionsTextNoPassword.addClass('d-none');
    $roomSuggestions.children().remove();

    disableCreateJoinButton(true);

    let roomName = $roomNameInput.val();

    if (!roomName || !validateRoomAndUserName(roomName, $roomPasswdInput.val(), $roomUserNameInput.val())) {
        disableCreateJoinButton(false);

        return;
    }

    showSpinnerOverlay();

    authorizeInRoom(roomName);
}

function authorizeInRoom(roomName) {
    //retrieve proper backend address for provided roomName, then open websocket to that backend and send 'room-create/join auth-only request'
    pickBackend(roomName, function (roomName, backendInstanceAddr, alternativeRoomNamePostfixes) {
        initHomepageWebSocket(roomName, backendInstanceAddr, alternativeRoomNamePostfixes, function (roomName) {
            sendData(ws,
                JSON.stringify({
                    p: 'unknown'
                })
            );

            sendData(ws,
                JSON.stringify({
                    c: COMMANDS.RoomCreateJoinAuthorize,
                    uN: $roomUserNameInput.val(),
                    r: {
                        n: roomName,
                        p: $roomPasswdInput.val()
                    }
                })
            );

            hideSpinnerOverlay();
        });
    });
}

function initHomepageWebSocket (roomName, backendInstanceAddr, alternativeRoomNamePostfixes, onOpen) {
    closeWsIfExists();

    try {
        ws = new WebSocket(WS_PROTOCOL + backendInstanceAddr + "/ws_entry");
    } catch (ex) {
        showErrorMessage(BUSINESS_ERRORS[ERROR_CODE_CONNECTION_ERROR].text);
        hideSpinnerOverlay();

        return;
    }

    ws.onopen = function () {
        onOpen(roomName);
    };

    ws.onerror = function(evt) {
        disableCreateJoinButton(false);
        hideSpinnerOverlay();

        showErrorMessage("connection error, please try again");
    }

    ws.onclose = function(evt) {
        disableCreateJoinButton(false);
        hideSpinnerOverlay();
    }

    ws.onmessage = function(evt) {
        let message = JSON.parse(evt.data);

        switch (message.c) {
            //business error
            case COMMANDS.Error:
                disableCreateJoinButton(false);
                hideSpinnerOverlay();

                const error = BUSINESS_ERRORS[message.m[0].t];

                showErrorMessage(error.text);

                if (error.code === ERROR_CODE_ROOM_INVALID_PASSWORD) {
                    tryShowRoomSuggestions(error.code, roomName, alternativeRoomNamePostfixes, false);
                }

                break;

            //request processed
            case COMMANDS.RequestProcessed:
                disableCreateJoinButton(false);

                shutdownSocket();

                //save info about room was created/joined into local storage before redirect
                const isCreatedRoom = message.pd === REQUEST_PROCESSING_DETAILS_ROOM_CREATED;
                LOCAL_STORAGE.setItem(REDIRECT_VARIABLE_LOCAL_STORAGE_KEY, JSON.stringify({
                    redirectFrom: REDIRECT_FROM_HOME_PAGE_ROOM_FORM,
                    redirectedAt: new Date().getTime(),
                    isCreatedRoom: isCreatedRoom,
                }));

                redirectToURL("/" + $roomNameInput.val());

                break;
        }
    }
}

function pickBackend (roomName, callback) {
    $.ajax(PICK_BACKEND_ENDPOINT + roomName)
        .done(function (data) {
            if (data && data.e) {
                showErrorMessage(data.e);

                disableCreateJoinButton(false);
                hideSpinnerOverlay();

                return;
            }

            if (data && data.bA) {
                callback(roomName, data.bA, data.aN);
            }
        })
        .fail(function () {
            showErrorMessage("connection, please try again (error while picking backend)");

            disableCreateJoinButton(false);
            hideSpinnerOverlay();
        });
}

function showErrorMessage (errorMsg) {
    const $newErrBlock = $("<div>");
    $newErrBlock.text(errorMsg);

    $errorText.append($newErrBlock);

    $errorText.animate({backgroundColor: "rgba(255, 94, 94, 0.5)"}, 500, function () {
        $errorText.animate({backgroundColor: "#fcfcfc"}, 500);

        scrollBlockBottom($homeContainerMainCenterWr[0]);
    });
}

//only names length are checked here, rest is checked on backend (aux-srv)
function validateRoomAndUserName(roomName, roomPassword, userName) {
    const roomNameTrimmed = roomName.trim();
    const roomPasswordTrimmed = roomPassword.trim();
    const userNameTrimmed = userName.trim();

    if (unicodeStringLength(roomNameTrimmed) < ROOM_CREDS_MIN_LENGTH || unicodeStringLength(roomNameTrimmed) > ROOM_CREDS_MAX_LENGTH
        || unicodeStringLength(roomPasswordTrimmed) > ROOM_CREDS_MAX_LENGTH) {

        showErrorMessage(BUSINESS_ERRORS[ERROR_CODE_ROOM_INVALID_CREDS_LENGTH].text
            + ", must be between " + ROOM_CREDS_MIN_LENGTH + " and " + ROOM_CREDS_MAX_LENGTH + " characters");

        return false;
    }

    if (userNameTrimmed && (
        unicodeStringLength(userNameTrimmed) < USER_NAME_MIN_LENGTH || unicodeStringLength(userNameTrimmed) > USER_NAME_MAX_LENGTH
    )) {
        showErrorMessage(BUSINESS_ERRORS[ERROR_CODE_INVALID_USER_NAME_LENGTH].text
            + ", must be between " + USER_NAME_MIN_LENGTH + " and " + USER_NAME_MAX_LENGTH + " characters");

        return false;
    }

    for (let i = 0; i < roomNameTrimmed.length; i++) {
        let nextCh = roomNameTrimmed.charAt(i);

        if (disallowedRoomNameSpecialChars[nextCh]) {
            showErrorMessage(BUSINESS_ERRORS[ERROR_CODE_ROOM_NAME_CONTAINS_BAD_CHARS].text);

            return false;
        }
    }

    return true;
}

function inPlaceFormatRoomNameInput() {
    $roomNameInput.val(formatRoomNameInput($roomNameInput.val()));
}

function tryShowRoomSuggestions(errorCode, roomName, alternativeRoomNamePostfixes, isRequestedAfterRoomPageRedirect) {
    if (alternativeRoomNamePostfixes && alternativeRoomNamePostfixes.length) {
        $roomSuggestionsWr.removeClass('d-none');

        if (isRequestedAfterRoomPageRedirect) {
            $roomSuggestionsTextNoPassword.removeClass('d-none');
        } else {
            $roomSuggestionsTextInvalidPassword.removeClass('d-none');
        }

        $roomNameInput.val(roomName);

        for (let i = 0; i < alternativeRoomNamePostfixes.length; i++) {
            const $roomSuggestionBlock = $("<span>");
            $roomSuggestionBlock.addClass("room-suggestion");
            $roomSuggestionBlock.text(roomName + "-" + alternativeRoomNamePostfixes[i]);

            $roomSuggestionBlock.on('click', function (e) {
                $roomNameInput.val($(e.target).text());
            });

            $roomSuggestions.append($roomSuggestionBlock);
        }
    }
}

function visitedRoomOnClickCallback (e) {
    if (roomJoinCreateBtnDisabled) {
        return false;
    }

    $errorText.children().remove();
    $roomSuggestionsWr.addClass('d-none');
    $roomSuggestionsTextInvalidPassword.addClass('d-none');
    $roomSuggestionsTextNoPassword.addClass('d-none');
    $roomSuggestions.children().remove();

    disableCreateJoinButton(true);

    hideMyRecentRoomsPopup();

    let roomName = $(e.currentTarget).text();

    $roomNameInput.val(roomName);

    showSpinnerOverlay();

    authorizeInRoom(roomName);
}

function sendData (ws, inputStr) {
    try {
        ws.send(inputStr);
    } catch (ex) {
        showErrorMessage(BUSINESS_ERRORS[ERROR_CODE_CONNECTION_ERROR].text);
        hideSpinnerOverlay();
    }
}
