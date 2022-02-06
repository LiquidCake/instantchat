

/* Constants */

const MILLS_IN_HOUR = 3600000;

const ERROR_CODE_ROOM_INVALID_PASSWORD = 203;
const ERROR_CODE_ROOM_INVALID_CREDS_LENGTH = 301;
const ERROR_CODE_ROOM_INVALID_DESCRIPTION_LENGTH = 304;
const ERROR_CODE_ROOM_USER_DUPLICATION = 209;

const ERROR_CODE_INVALID_USER_NAME_LENGTH = 205;
const ERROR_CODE_ROOM_NAME_CONTAINS_BAD_CHARS = 303;
const ERROR_CODE_CONNECTION_ERROR = 102;

const ERROR_CODE_MESSAGE_IS_TOO_LONG = 207;

const ROOM_CREDS_MIN_LENGTH = 3;
const ROOM_CREDS_MAX_LENGTH = 100;

const USER_NAME_MIN_LENGTH = 1;
const USER_NAME_MAX_LENGTH = 80;

const MAX_ROOM_DESCRIPTION_LENGTH = 400;

const MAX_TEXT_MESSAGE_LENGTH = 10000;

const MESSAGE_UNAVAILABLE_PLACEHOLDER_TEXT = "/ message unavailable /";

const MESSAGE_META_MARKER_TYPE_DRAWING = "$#$meta_marker_is_drawing$#$";

const LOCAL_STORAGE = window.localStorage;

const PICK_BACKEND_ENDPOINT = "./pick_backend?roomName=";
const GET_URL_PREVIEW_ENDPOINT = "./get_url_preview";
const GET_TEXT_FILE_ENDPOINT = "./get_text_file";
const UPLOAD_TEXT_FILE_ENDPOINT = "./upload_text_file";

const URL_PARAM_BACKEND_HOST = 'backendHost';

const protocol = location.protocol.startsWith("https") ? "wss://" : "ws://";

const WS_ENDPOINT = protocol + location.host + "/ws_entry";
const WS_CLOSE_CODE_NORMAL = 1000;

const KEEP_ALIVE_BEACON = JSON.stringify({kA: "OK"});

const RECENT_ROOMS_EMPTY_BLOCK_ID = "recent-rooms-empty";

const VISITED_ROOMS_LOCAL_STORAGE_KEY = "VISITED_ROOMS";
const COOKIES_ACCEPTED_LOCAL_STORAGE_KEY = "COOKIES_ACCEPTED";
const BOTS_LIST_LOCAL_STORAGE_KEY = "BOTS_LIST";

const LAST_VISITED_ROOM_ID_LOCAL_STORAGE_KEY = "LAST_VISITED_ROOM_ID";
const LAST_UNSENT_MESSAGE_TEXT_LOCAL_STORAGE_KEY = "LAST_UNSENT_MESSAGE_TEXT";

//redirect variable passing
const REDIRECT_VARIABLE_LOCAL_STORAGE_KEY = "REDIRECT_VARIABLE";

const REDIRECT_FROM_HOME_PAGE_ROOM_FORM            = "home-page-room-form";
const REDIRECT_FROM_ROOM_PAGE_ON_ERROR             = "room-page-on-error";
const REDIRECT_FROM_ANY_PAGE_ON_RECENT_ROOM_CLICK  = "room-page-on-recent-room-click";

const REDIRECT_VARIABLE_TTL_MS = 30000;

//message codes
const ROOM_TO_HOME_PG_REDIRECT_ERROR_GENERIC                      = "room_to_home_pg_redirect_error_generic";
const ROOM_TO_HOME_PG_REDIRECT_ERROR_BAD_ROOM_NAME_LENGTH         = "room_to_home_pg_redirect_error_bad_room_name_length";
const ROOM_TO_HOME_PG_REDIRECT_ERROR_ROOM_NAME_CONTAINS_SLASH     = "room_to_home_pg_redirect_error_room_name_contains_slash";
const ROOM_TO_HOME_PG_REDIRECT_ERROR_BUSINESS                     = "room_to_home_pg_redirect_error_business";
const ROOM_TO_HOME_PG_REDIRECT_ERROR_CONNECTION                   = "room_to_home_pg_redirect_connection_error";
const ROOM_TO_HOME_PG_REDIRECT_ERROR_CONNECTION_PICK_BACKEND      = "room_to_home_pg_redirect_connection_error_pick_backend";
const ROOM_TO_HOME_PG_REDIRECT_ERROR_SESSION_COOKIE_MISSING       = "room_to_home_pg_redirect_connection_error_session_cookie_missing";

const REQUEST_PROCESSING_DETAILS_ROOM_CREATED = "room_created";
const REQUEST_PROCESSING_DETAILS_ROOM_HAS_PASSWORD = "password=true";


//messages
const REQUEST_PROCESSING_DETAILS_ROOM_CREATED_MESSAGE = "you just created a new room!";
const REQUEST_PROCESSING_DETAILS_ROOM_JOINED_MESSAGE  = "joined existing room";

const NOTIFICATION_TEXT_MESSAGE_LIMIT_APPROACHING     = "room is approaching messages limit, old messages will be removed soon";
const NOTIFICATION_TEXT_MESSAGE_LIMIT_REACHED         = "room messages limit reached, old messages were removed";

const NOTIFICATION_TEXT_USE_SHARE_BTN                 = 'to share room - please use \'share\' button';
const NOTIFICATION_TEXT_ROOM_CREATOR_SELF_VOTE        = 'as a room creator, you can \'support\' own messages - to pin them';
const NOTIFICATION_TEXT_ZOOM                          = 'use zoom to scale interface!';

const ERROR_TEXT_SESSION_COOKIE_MISSING               = "session cookie is absent. Please make sure cookies are enabled in your browser and reload page";
const ERROR_TEXT_ROOM_NAME_CONTAINS_SLASH             = "room name must NOT contain '/' symbol";
const ERROR_TEXT_REDIRECT_FROM_ROOM_PAGE_GENERIC      = "error while opening room page";

const ERROR_NOTIFICATION_TEXT_FAILED_COPY_MESSAGE_TO_CLIPBOARD = "failed to copy message text to clipboard, please do it manually";
const ERROR_NOTIFICATION_TEXT_FAILED_COPY_URL_TO_CLIPBOARD = "failed to copy room URL to clipboard, please do it manually";
const ERROR_NOTIFICATION_TEXT_FAILED_TO_UPLOAD_FILE = "failed to upload file to server, please try again in a moment";

const TOP_NOTIFICATION_SHOW_MS = 3000;
const NOTIFICATION_SHOW_MS = 7000;
const KEEPALIVE_INTERVAL_MS = 5000;


const BUSINESS_ERRORS = {
    101: {name: "WsServerError",                             code: 101, text: "server error"},
    102: {name: "WsConnectionError",                         code: 102, text: "connection error"},
    103: {name: "WsInvalidInput",                            code: 103, text: "invalid input"},

    201: {name: "WsRoomExists",                              code: 201, text: "room with this name already exists"},
    202: {name: "WsRoomNotFound",                            code: 202, text: "room not found"},
    203: {name: "WsRoomInvalidPassword",                     code: 203, text: "invalid room password"},
    204: {name: "WsRoomUserNameTaken",                       code: 204, text: "provided user name is already taken"},
    205: {name: "WsRoomUserNameValidationError",             code: 205, text: "invalid room user name length"},
    206: {name: "WsRoomNotAuthorized",                       code: 206, text: "not authorized to join this room"},
    207: {name: "WsRoomMessageTooLargeError",                code: 207, text: "message is too long"},
    208: {name: "WsRoomIsFullError",                         code: 208, text: "room is full"},
    209: {name: "WsRoomUserDuplication",                     code: 209, text: "user connected to this room from another browser tab"},

    301: {name: "WsRoomCredsValidationErrorBadLength",       code: 301, text: "invalid room name length"},
    302: {name: "WsRoomCredsValidationErrorNameForbidden",   code: 302, text: "room name is forbidden"},
    303: {name: "WsRoomCredsValidationErrorNameHasBadChars", code: 303, text: "room name contains bad characters"},
    304: {name: "WsRoomValidationErrorBadDescriptionLength", code: 304, text: "invalid room description length"},
};

const COMMANDS = {
    RoomCreateJoin: "R_C_J",
    RoomCreateJoinAuthorize: "R_C_J_AU",
    RoomCreate: "R_C",
    RoomJoin: "R_J",
    RoomChangeUserName: "R_CH_UN",
    RoomChangeDescription: "R_CH_D",
    RoomMembersChanged: "R_M_CH",

    TextMessage: "TM",
    TextMessageEdit: "TM_E",
    TextMessageDelete: "TM_D",
    TextMessageSupportOrReject: "TM_S_R",
    AllTextMessages: "ALL_TM",

    UserDrawingMessage: "DM",

    Error: "ER",
    RequestProcessed: "RP",

    NotifyMessagesLimitApproaching: "N_M_LIMIT_A",
    NotifyMessagesLimitReached: "N_M_LIMIT_R",
};

const allowedRoomNameSpecialChars = {
    "!": true,
    "@": true,
    "$": true,
    "*": true,
    "(": true,
    ")": true,
    "_": true,
    "-": true,
    ",": true,
    ".": true,
    "~": true,
    "[": true,
    "]": true,
}

const disallowedRoomNameSpecialChars = {
    "”": true,
    "#": true,
    "%": true,
    "&": true,
    "’": true,
    "+": true,
    "/": true,
    ":": true,
    ";": true,
    "<": true,
    "=": true,
    ">": true,
    "?": true,
    "\\": true,
    "^": true,
    "`": true,
    "{": true,
    "|": true,
    "}": true,
}

const receivedURLPreviewInfoItems = {};

/* Variables */

let ws;

let keepAliveTimerID = -1;

let currentBuildNumber;

document.addEventListener("DOMContentLoaded", function(event) {
    currentBuildNumber = BUILD_NUMBER;
});

function keepAlive() {
    //console.log(getCurrentTimeStr() + ": start keepAlive, timeoutId: " + keepAliveTimerID);

    if (ws) {
        switch (ws.readyState) {
            case ws.OPEN:
                ws.send(KEEP_ALIVE_BEACON);

                break
            case ws.CLOSED:
            case ws.CLOSING:
                break
        }
    }

    keepAliveTimerID = setTimeout(keepAlive, KEEPALIVE_INTERVAL_MS);
    //console.log(getCurrentTimeStr() + ": end keepAlive, timeoutId: " + keepAliveTimerID);
}

function cancelKeepAlive() {
    if (keepAliveTimerID) {
        clearTimeout(keepAliveTimerID);
    }
}

//added to break the bfcache in Firefox and Safari
window.addEventListener('unload', function(){});

window.onbeforeunload = function (e) {
    shutdownSocket();

    return null;
}
window.addEventListener(
    "beforeunload",
    function (e) {
        shutdownSocket();

        return null;
    },
    false
);

function unicodeStringLength(str) {
    return [...str].length
}

function getMillsFromNano (nanoTimestamp) {
    return new Date(Math.trunc(nanoTimestamp / 1000000)).getTime();
}

function getTimeStrFromMills (timestamp) {
    const date = new Date(timestamp);

    return formatNumberForDatetime(date.getHours()) + ":" + formatNumberForDatetime(date.getMinutes()) + ":" + formatNumberForDatetime(date.getSeconds());
}

function getDateStrFromMills (timestamp) {
    const date = new Date(timestamp);

    return formatNumberForDatetime(date.getDate()) + "-" + formatNumberForDatetime(date.getMonth() + 1) + "-" + date.getFullYear();
}

function getHoursMinutesStrFromMills (timestamp) {
    const date = new Date(timestamp);

    return formatNumberForDatetime(date.getHours()) + ":" + formatNumberForDatetime(date.getMinutes());
}

function formatNumberForDatetime(number) {
    return number < 10 ? "0" + number : number;
}

function storeVisitedRoomsArray(newVisitedRooms) {
    newVisitedRooms.sort(function (a, b) {
        if (a.visitedAt < b.visitedAt) {
            return 1;
        }
        if (a.visitedAt > b.visitedAt) {
            return -1;
        }

        return 0;
    });

    LOCAL_STORAGE.setItem(VISITED_ROOMS_LOCAL_STORAGE_KEY, JSON.stringify(newVisitedRooms));
}

function formatRoomNameInput(roomNameInput) {
    return roomNameInput.toLowerCase().trim()
        .replace(/\s+/g, '-');
}

function getUrlVars() {
    let vars = [], hash;
    let hashes = window.location.href.slice(window.location.href.indexOf('?') + 1).split('&');
    for(let i = 0; i < hashes.length; i++)
    {
        hash = hashes[i].split('=');
        vars.push(hash[0]);
        vars[hash[0]] = hash[1];
    }
    return vars;
}

function redirectToURL(url) {
    window.location.assign(url);
}

function closeWsIfExists() {
    if (ws) {
        ws.onclose = function () {};
        ws.close(WS_CLOSE_CODE_NORMAL);
    }
}

function shutdownSocket() {
    cancelKeepAlive();
    closeWsIfExists();
}

function cutMessageTextForResponse (text) {
    const textWithoutLinebreaks = text.replace(/(\r\n|\n|\r)/gm, " ");

    return textWithoutLinebreaks.length <= 33
        ? textWithoutLinebreaks
        : textWithoutLinebreaks.substring(0, 30) + "...";
}

function getCookie (name) {
    const match = document.cookie.match(new RegExp('(^| )' + name + '=([^;]+)'));
    if (match) return match[2];
}

function getSelectionText () {
    let text = "";

    if (window.getSelection) {
        text = window.getSelection().toString();
    } else if (document.selection && document.selection.type !== "Control") {
        text = document.selection.createRange().text;
    }

    return text;
}

function isCharASCII(str) {
    return /^[\x00-\x7F]*$/.test(str);
}

function stopPropagationAndDefault (e) {
    e.preventDefault();
    e.stopPropagation();
}

function getSubstringIndexes (str, substr) {
    if (!str || !substr) {
        return [];
    }

    const stringLower = str.toLowerCase();
    const substringLower = substr.toLowerCase();
    const indexesArr = [];

    let nextIndex = stringLower.indexOf(substringLower);
    let idxCount = 0;

    if (nextIndex === -1) {
        return indexesArr;
    } else {
        indexesArr.push(nextIndex);
        idxCount++;
    }

    do {
        nextIndex = stringLower.indexOf(substringLower, nextIndex + 1);

        if (nextIndex !== -1) {
            indexesArr.push(nextIndex);
            idxCount++;
        }
    } while (nextIndex !== -1);

    return indexesArr;
}

function recursiveApplyFuncToDom ($element, func) {
    func($element);

    $element.children().each(function () {
        let $currentElement = $(this);

        recursiveApplyFuncToDom($currentElement, func);
    });
}

function getAllIndexes(arr, val) {
    let indexes = [], i = -1;

    while ((i = arr.indexOf(val, i + 1)) !== -1){
        indexes.push(i);
    }

    return indexes;
}

function findAllLinksInText(text) {
    let arr = [];

    let nextMatchEndingIdx = -1;

    while (true) {
        let linkStartIdx;

        const httpLinkStartIdx = text.indexOf('http://', nextMatchEndingIdx);
        const httpsLinkStartIdx = text.indexOf('https://', nextMatchEndingIdx);

        if (httpLinkStartIdx === -1 && httpsLinkStartIdx === -1) {
            return arr;
        }

        if (httpLinkStartIdx !== -1 && httpsLinkStartIdx !== -1) {
            linkStartIdx = Math.min(httpLinkStartIdx, httpsLinkStartIdx);
        } else {
            linkStartIdx = httpLinkStartIdx !== -1 ? httpLinkStartIdx : httpsLinkStartIdx;
        }


        let linkEndIdx;

        const nextSpaceIdx = text.indexOf(' ', linkStartIdx);
        const nextQuoteIdx = text.indexOf('\'', linkStartIdx);
        const nextQuote2Idx = text.indexOf('`', linkStartIdx);
        const nextDoubleQuoteIdx = text.indexOf('"', linkStartIdx);

        const nextTabIdx = text.indexOf('\t', linkStartIdx);
        const nextNewlineIdx = text.indexOf('\n', linkStartIdx);

        linkEndIdx = findMinimalExistingStringIndex(nextSpaceIdx, nextQuoteIdx, nextQuote2Idx, nextDoubleQuoteIdx, nextTabIdx, nextNewlineIdx);

        if (linkEndIdx === -1) {
            linkEndIdx = text.length;
        }

        const linkText = text.substr(linkStartIdx , linkEndIdx - linkStartIdx);

        //avoid duplicates
        if (!arr.find(function (element) {
            return element.text === linkText;
        })) {
            arr.push({
                index: linkStartIdx,
                length: linkEndIdx - linkStartIdx,
                text: linkText,
            });
        }

        nextMatchEndingIdx = linkEndIdx;
    }
}

function findMinimalExistingStringIndex (...indexes) {
    let minIndex = null;

    for (let i = 0; i < indexes.length; i++) {
        const nextIndex = indexes[i];

        if (
            nextIndex >= 0 &&
            (minIndex === null || nextIndex < minIndex)
        ) {
            minIndex = nextIndex;
        }
    }

    return minIndex === null ? -1 : minIndex;
}

//only name length is checked here, rest is checked on backend (aux-srv)
function validateRoomName(roomName) {
    const roomNameTrimmed = roomName.trim();

    if (unicodeStringLength(roomNameTrimmed) < ROOM_CREDS_MIN_LENGTH || unicodeStringLength(roomNameTrimmed) > ROOM_CREDS_MAX_LENGTH) {
        return false;
    }

    return true;
}

function escapeRegExp (string){
    return string.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

function stringContainsEmoji (str) {
    return /\p{Emoji}/u.test(str);
}

function getUrlPreviewInfo (urlToGetPreview, callback) {
    if (receivedURLPreviewInfoItems[urlToGetPreview]) {
        setTimeout(function () {
            callback(receivedURLPreviewInfoItems[urlToGetPreview]);
        }, 1);

        return;
    }

    $.ajax({
        url: GET_URL_PREVIEW_ENDPOINT,
        type: "post",
        data: "url_to_preview=" + urlToGetPreview,
    })
        .done(function (data) {
            receivedURLPreviewInfoItems[urlToGetPreview] = data;
            callback(data);
        })
        .fail(function () {
            console.log('failed to load URL preview: ' + JSON.stringify(arguments));
        });
}

function getUserDrawingFromFileServer (fileName, fileGroupPrefix, okCallback, failCallback) {
    $.ajax({
        url: GET_TEXT_FILE_ENDPOINT + '?file_name=' + fileName + '&file_group_prefix=' + fileGroupPrefix,
        type: "get",
    })
        .done(okCallback)
        .fail(failCallback);
}

function uploadUserDrawingToFileServer (fileName, fileGroupPrefix, fileContentBase64, okCallback, failCallback) {
    $.ajax({
        url: UPLOAD_TEXT_FILE_ENDPOINT,
        type: "post",
        data: {
            file_name: fileName,
            file_group_prefix: fileGroupPrefix,
            file_content: fileContentBase64
        },
    })
        .done(okCallback)
        .fail(failCallback);
}

function saveBase64AsFile(base64, fileName) {
    const $link = $("<a>");

    //fix for FF
    $body.append($link);

    $link.attr('href', base64);
    $link.attr('download', fileName);
    $link[0].click();

    $link.remove();
}

function reloadPage () {
    location.reload();
}

function randomIntBetween (min, max) {
    return Math.floor(Math.random() * (max - min + 1) + min)
}
