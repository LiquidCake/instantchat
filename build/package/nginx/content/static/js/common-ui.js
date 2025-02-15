/* Constants */
const KEY_CODE_ENTER = 13;
const KEY_CODE_ESCAPE = 27;
const KEY_CODE_ARROW_UP = 38;

const DESKTOP_MENU_SCREEN_MIN_WIDTH_PX = 992;

const MENU_HIDDEN_LOCAL_STORAGE_KEY = 'MENU_HIDDEN';
const FOLK_PICKS_HIDDEN_LOCAL_STORAGE_KEY = 'FOLK_PICKS_HIDDEN';
const SHOW_FULL_SMILES_LIST_LOCAL_STORAGE_KEY = 'SHOW_FULL_SMILES_LIST';
const ZOOM_NOTIF_LAST_SHOWN_AT_LOCAL_STORAGE_KEY = 'ZOOM_NOTIF_LAST_SHOWN_AT';
const UI_THEME_LOCAL_STORAGE_KEY = 'UI_THEME';

const UI_THEME_LIGHT = 'light';
const UI_THEME_DARK  = 'dark';

/* DOM objects references */

const $window = $(window);
const $document = $(document);
const $body = $('body');

const $pageContentHolder = $('#page-content-holder');

const $mainOverlay = $('.main-overlay');

const $spinnerOverlay = $('.spinner-overlay');
const $spinnerOverlayMidContentWr = $('.spinner-overlay-mid-content-wr');

const $pageInitOverlay = $('.page-init-overlay');

const $mainHeader = $('.main-navbar');
const $mainWr = $('.main-wr');

const $sidebarWr = $('.main-sidebar-wr');
const $sidebarMobileWr = $('.main-sidebar-mobile-wr');
const $recentRoomsLink = $('.main-sidebar-recent-rooms');
const $recentRoomsClose = $('.recent-rooms-close');
const $recentRoomsPopup = $('.recent-rooms-popup');

const $preferencesLink = $('.main-sidebar-preferences');
const $preferencesClose = $('.preferences-close');
const $preferencesPopup = $('.preferences-popup');

const $preferencesKeyRecordInput = $('#preferences-key-record-input');
const $preferencesKeyRecordBtn = $('#preferences-key-record-btn');

const $preferencesKeyThemeLight = $('#preferences-theme-light-btn');
const $preferencesKeyThemeDark = $('#preferences-theme-dark-btn');

const $preferencesModeKeyWr = $('.preferences-mode-key-wr');

const $pageCommonMobileMenuToggle = $('.page-common-toggle-mobile-menu');

const $recentRoomsWr = $("#recent-rooms-wr");

const $cookieWarning = $(".cookie-warning");
const $cookieWarningAgree = $(".cookie-warning-agree");

/* Variables */

let currentViewportWidth;
let currentViewportHeight;

let swipingBehaviourAllowed = true;

let preferencesKeyRecordingInProgress = false;
let preferencesSelectedKeyCode = null;

let selectedUITheme = null;

function commonPageUIInit () {
    const cookiesAccepted = LOCAL_STORAGE.getItem(COOKIES_ACCEPTED_LOCAL_STORAGE_KEY);

    if (!cookiesAccepted) {
        showCookieWarning();
    }

    const selectedUIThemeVal = LOCAL_STORAGE.getItem(UI_THEME_LOCAL_STORAGE_KEY);

    if (selectedUIThemeVal) {
        selectedUITheme = selectedUIThemeVal;
    } else {
        const osIsInDarkTheme = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches;

        selectedUITheme = osIsInDarkTheme ? UI_THEME_DARK : UI_THEME_LIGHT;
    }

    initUIColorTheme();

    $preferencesKeyThemeLight.on('click', function () {
        selectedUITheme = UI_THEME_LIGHT;
        LOCAL_STORAGE.setItem(UI_THEME_LOCAL_STORAGE_KEY, UI_THEME_LIGHT);

        initUIColorTheme();
    });

    $preferencesKeyThemeDark.on('click', function () {
        selectedUITheme = UI_THEME_DARK;
        LOCAL_STORAGE.setItem(UI_THEME_LOCAL_STORAGE_KEY, UI_THEME_DARK);

        initUIColorTheme();
    });

    $preferencesKeyRecordInput.on('focusin', function () {
        preferencesKeyRecordingInProgress = true;

        $preferencesKeyRecordBtn.removeClass('background-greyed');
        $preferencesKeyRecordInput.addClass('preferences-key-record-input-active');
    });

    $preferencesKeyRecordInput.on('keydown', function (e) {
        stopPropagationAndDefault(e);

        //on keydown (!) event codes will be same as in win32 libs
        preferencesSelectedKeyCode = e.keyCode;

        let keyName = keycodesToNames[preferencesSelectedKeyCode];

        if (keyName === undefined) {
            keyName = "unknown";
        }

        $preferencesKeyRecordInput.val(keyName);
    });

    //handle window mode change key actions

    let storedChangeWindowModeKey = LOCAL_STORAGE.getItem(WEBVIEW_CHANGE_WINDOW_MODE_KEY_LOCAL_STORAGE_KEY);

    if (!storedChangeWindowModeKey) {
        storedChangeWindowModeKey = WEBVIEW_CHANGE_WINDOW_MODE_KEY_DEFAULT;
    }

    LOCAL_STORAGE.setItem(WEBVIEW_CHANGE_WINDOW_MODE_KEY_LOCAL_STORAGE_KEY, storedChangeWindowModeKey);

    postMessageToWebview(WEB_COMMAND_CHANGE_WINDOW_MODE_KEY + storedChangeWindowModeKey);

    preferencesSelectedKeyCode = storedChangeWindowModeKey;

    const keyName = keycodesToNames[storedChangeWindowModeKey];
    $preferencesKeyRecordInput.val(keyName);
    changeRoomInfoOverlayKeyIndicator(keyName);

    $preferencesKeyRecordBtn.on('click', function () {
        if (preferencesKeyRecordingInProgress) {
            const keyName = keycodesToNames[preferencesSelectedKeyCode];

            //unknown key
            if (keyName === undefined) {
                return;
            }

            LOCAL_STORAGE.setItem(WEBVIEW_CHANGE_WINDOW_MODE_KEY_LOCAL_STORAGE_KEY, preferencesSelectedKeyCode);

            postMessageToWebview(WEB_COMMAND_CHANGE_WINDOW_MODE_KEY + preferencesSelectedKeyCode);

            $preferencesKeyRecordInput.removeClass('preferences-key-record-input-active');

            $preferencesKeyRecordInput.val(keyName);
            changeRoomInfoOverlayKeyIndicator(keyName);

            setShowChangeOverlayModeFlag(true);
            tryShowRoomInfoOverlayKeyWr();

            $preferencesKeyRecordBtn.addClass('background-greyed');

            preferencesKeyRecordingInProgress = false;
        }
    });

    //check global var set by client app (if any)
    if (window.clientType === 'win') {
        $preferencesModeKeyWr.removeClass('d-none');
    }
}

function isMobileDevice () {
    let check = false;
    (function (a) {if(/(android|bb\d+|meego).+mobile|avantgo|bada\/|blackberry|blazer|compal|elaine|fennec|hiptop|iemobile|ip(hone|od)|iris|kindle|lge |maemo|midp|mmp|mobile.+firefox|netfront|opera m(ob|in)i|palm( os)?|phone|p(ixi|re)\/|plucker|pocket|psp|series(4|6)0|symbian|treo|up\.(browser|link)|vodafone|wap|windows ce|xda|xiino/i.test(a)||/1207|6310|6590|3gso|4thp|50[1-6]i|770s|802s|a wa|abac|ac(er|oo|s\-)|ai(ko|rn)|al(av|ca|co)|amoi|an(ex|ny|yw)|aptu|ar(ch|go)|as(te|us)|attw|au(di|\-m|r |s )|avan|be(ck|ll|nq)|bi(lb|rd)|bl(ac|az)|br(e|v)w|bumb|bw\-(n|u)|c55\/|capi|ccwa|cdm\-|cell|chtm|cldc|cmd\-|co(mp|nd)|craw|da(it|ll|ng)|dbte|dc\-s|devi|dica|dmob|do(c|p)o|ds(12|\-d)|el(49|ai)|em(l2|ul)|er(ic|k0)|esl8|ez([4-7]0|os|wa|ze)|fetc|fly(\-|_)|g1 u|g560|gene|gf\-5|g\-mo|go(\.w|od)|gr(ad|un)|haie|hcit|hd\-(m|p|t)|hei\-|hi(pt|ta)|hp( i|ip)|hs\-c|ht(c(\-| |_|a|g|p|s|t)|tp)|hu(aw|tc)|i\-(20|go|ma)|i230|iac( |\-|\/)|ibro|idea|ig01|ikom|im1k|inno|ipaq|iris|ja(t|v)a|jbro|jemu|jigs|kddi|keji|kgt( |\/)|klon|kpt |kwc\-|kyo(c|k)|le(no|xi)|lg( g|\/(k|l|u)|50|54|\-[a-w])|libw|lynx|m1\-w|m3ga|m50\/|ma(te|ui|xo)|mc(01|21|ca)|m\-cr|me(rc|ri)|mi(o8|oa|ts)|mmef|mo(01|02|bi|de|do|t(\-| |o|v)|zz)|mt(50|p1|v )|mwbp|mywa|n10[0-2]|n20[2-3]|n30(0|2)|n50(0|2|5)|n7(0(0|1)|10)|ne((c|m)\-|on|tf|wf|wg|wt)|nok(6|i)|nzph|o2im|op(ti|wv)|oran|owg1|p800|pan(a|d|t)|pdxg|pg(13|\-([1-8]|c))|phil|pire|pl(ay|uc)|pn\-2|po(ck|rt|se)|prox|psio|pt\-g|qa\-a|qc(07|12|21|32|60|\-[2-7]|i\-)|qtek|r380|r600|raks|rim9|ro(ve|zo)|s55\/|sa(ge|ma|mm|ms|ny|va)|sc(01|h\-|oo|p\-)|sdk\/|se(c(\-|0|1)|47|mc|nd|ri)|sgh\-|shar|sie(\-|m)|sk\-0|sl(45|id)|sm(al|ar|b3|it|t5)|so(ft|ny)|sp(01|h\-|v\-|v )|sy(01|mb)|t2(18|50)|t6(00|10|18)|ta(gt|lk)|tcl\-|tdg\-|tel(i|m)|tim\-|t\-mo|to(pl|sh)|ts(70|m\-|m3|m5)|tx\-9|up(\.b|g1|si)|utst|v400|v750|veri|vi(rg|te)|vk(40|5[0-3]|\-v)|vm40|voda|vulc|vx(52|53|60|61|70|80|81|83|85|98)|w3c(\-| )|webc|whit|wi(g |nc|nw)|wmlb|wonu|x700|yas\-|your|zeto|zte\-/i.test(a.substr(0,4))) check = true;})(navigator.userAgent||navigator.vendor||window.opera);
    return check;
}

function animateBlockForAttention ($elem, isSuccess, duration) {
    if ($elem.attr('data-is-animation-in-progress') !== 'true') {
        $elem.attr('data-is-animation-in-progress', true);

        const origBgColor = $elem.css('background-color');
        const toColor = isSuccess ? "#75de94" : "#cf4141";

        $elem.animate({backgroundColor: toColor}, duration || 500, function () {
            $elem.animate({backgroundColor: origBgColor}, duration || 500, function () {
                $elem.removeAttr('data-is-animation-in-progress');
                $elem.css('background-color', '');
            });
        });
    }
}

function showMenuMobile () {
    $sidebarMobileWr.css('display', 'block');
    showMainOverlay();
}

function hideMenuMobile () {
    if ($sidebarMobileWr.css('display') !== 'none') {
        $sidebarMobileWr.css('display', 'none');
        hideMainOverlay();
    }
}

function showMyRecentRoomsPopup () {
    cancelMessageTextSelectionMode();

    $recentRoomsPopup.css('display', 'block');
    showMainOverlay();
}

function hideMyRecentRoomsPopup () {
    if ($recentRoomsPopup.css('display') !== 'none') {
        $recentRoomsPopup.css('display', 'none');
        hideMainOverlay();
    }
}

function showPreferencesPopup () {
    cancelMessageTextSelectionMode();

    $preferencesPopup.css('display', 'block');
    showMainOverlay();
}

function hidePreferencesPopup () {
    if ($preferencesPopup.css('display') !== 'none') {
        $preferencesPopup.css('display', 'none');
        hideMainOverlay();
    }
}

function showMainOverlay () {
    $mainOverlay.removeClass('d-none');
}

function hideMainOverlay () {
    $mainOverlay.addClass('d-none');
}

function showMenuMobileOnSwipe () {
    if (isOnMobileScreenSize()) {
        hideMyRecentRoomsPopup();
        showMenuMobile();
    }
}

function scrollToRepliedMessage () {
    const replyMsgToId = $(this).closest('.room-msg-main-wr').attr('data-reply-to-msg-id');

    scrollToTargetMsg($roomMessagesWr.find('div[data-msg-id=' + replyMsgToId + ']'));
}

function scrollToTargetMsg ($targetMsgElement) {
    scrollToElement($targetMsgElement[0]);

    if ($targetMsgElement.attr('data-is-animation-in-progress') !== 'true') {
        $targetMsgElement.attr('data-is-animation-in-progress', true);

        $targetMsgElement.animate({backgroundColor: selectedUITheme === UI_THEME_DARK ? '#1c5e2f' : '#91e6a9'}, 700,
        function () {
            $targetMsgElement.animate({backgroundColor: selectedUITheme === UI_THEME_DARK ? '#272826' : '#fff'}, 700,
            function () {
                $targetMsgElement.css('background-color', '');
                $targetMsgElement.removeAttr('data-is-animation-in-progress');
            });
        });
    }
}

function scrollToElement (element) {
    if (element) {
        if (isMobileClientDevice) {
            element.scrollIntoView(true);  //passing params object doesnt work on touch event currently
        } else {
            element.scrollIntoView({block: 'center', behavior: 'smooth'});
        }
    }
}

class Xwiper {
    constructor(element) {
        this.options = {threshold: 120, verticalThreshold: 150, passive: false};

        this.element = null;
        this.touchStartX = 0;
        this.touchStartY = 0;
        this.touchEndX = 0;
        this.touchEndY = 0;
        this.onSwipeLeftAgent = null;
        this.onSwipeRightAgent = null;
        this.onSwipeUpAgent = null;
        this.onSwipeDownAgent = null;
        this.onTapAgent = null;

        this.onTouchStart = this.onTouchStart.bind(this);
        this.onTouchEnd = this.onTouchEnd.bind(this);
        this.onSwipeLeft = this.onSwipeLeft.bind(this);
        this.onSwipeRight = this.onSwipeRight.bind(this);
        this.onSwipeUp = this.onSwipeUp.bind(this);
        this.onSwipeDown = this.onSwipeDown.bind(this);
        this.onTap = this.onTap.bind(this);
        this.destroy = this.destroy.bind(this);
        this.handleGesture = this.handleGesture.bind(this);

        let eventOptions = this.options.passive ? { passive: true } : false;

        this.element = (element instanceof EventTarget) ? element : document.querySelector(element);

        this.element.addEventListener('touchstart', this.onTouchStart, eventOptions);
        this.element.addEventListener('touchend', this.onTouchEnd, eventOptions);
    }

    onTouchStart(event) {
        this.touchStartX = event.changedTouches[0].screenX;
        this.touchStartY = event.changedTouches[0].screenY;
    }

    onTouchEnd(event) {
        this.touchEndX = event.changedTouches[0].screenX;
        this.touchEndY = event.changedTouches[0].screenY;

        if (swipingBehaviourAllowed) {
            this.handleGesture();
        }
    }

    onSwipeLeft(func) {
        this.onSwipeLeftAgent = func;
    }
    onSwipeRight(func) {
        this.onSwipeRightAgent = func;
    }
    onSwipeUp(func) {
        this.onSwipeUpAgent = func;
    }
    onSwipeDown(func) {
        this.onSwipeDownAgent = func;
    }
    onTap(func) {
        this.onTapAgent = func;
    }

    destroy() {
        this.element.removeEventListener('touchstart', this.onTouchStart);
        this.element.removeEventListener('touchend', this.onTouchEnd);
    }

    handleGesture() {
        /**
         * swiped left
         */
        if (this.touchEndX + this.options.threshold <= this.touchStartX
            && Math.abs(this.touchEndY - this.touchStartY) < this.options.verticalThreshold) {
            this.onSwipeLeftAgent && this.onSwipeLeftAgent();
            return 'left';
        }

        /**
         * swiped right
         */
        if (this.touchEndX - this.options.threshold >= this.touchStartX
            && Math.abs(this.touchEndY - this.touchStartY) < this.options.verticalThreshold) {
            this.onSwipeRightAgent && this.onSwipeRightAgent();
            return 'right';
        }

        /**
         * swiped up
         */
        if (this.touchEndY + this.options.threshold <= this.touchStartY) {
            this.onSwipeUpAgent && this.onSwipeUpAgent();
            return 'up';
        }

        /**
         * swiped down
         */
        if (this.touchEndY - this.options.threshold >= this.touchStartY) {
            this.onSwipeDownAgent && this.onSwipeDownAgent();
            return 'down';
        }

        /**
         * tap
         */
        if (this.touchEndY === this.touchStartY) {
            this.onTapAgent && this.onTapAgent();
            return 'tap';
        }
    }
}

function storeCurrentViewportDimensions () {
    currentViewportWidth = $window.width();
    currentViewportHeight = $window.height();
}

function loadVisitedRooms(roomLinkOnClickCallback) {

    $recentRoomsWr.children().remove();
    let visitedRoomsStr = LOCAL_STORAGE.getItem(VISITED_ROOMS_LOCAL_STORAGE_KEY);

    if (!visitedRoomsStr) {
        visitedRoomsStr = "[]";
    }

    const visitedRooms = JSON.parse(visitedRoomsStr);

    if (visitedRooms.length) {
        for (let i = 0; i < visitedRooms.length; i++) {
            const room = visitedRooms[i];

            drawRecentRoom($recentRoomsWr, room, roomLinkOnClickCallback);
        }
    } else {
        $recentRoomsWr.append('<p id="' + RECENT_ROOMS_EMPTY_BLOCK_ID + '">no rooms visited yet</p>');
    }
}

function drawRecentRoom ($recentRoomsWr, room, roomLinkOnClickCallback) {
    const recentRoomsEmptyBlock = $(RECENT_ROOMS_EMPTY_BLOCK_ID);
    if (recentRoomsEmptyBlock.length) {
        recentRoomsEmptyBlock.remove();
    }

    const $roomBlock = $('<div class="recent-room-wr">');
    const $roomLink = $('<a href="javascript:void(0);" class="recent-room-title">');
    $roomLink.text(room.roomName);

    $roomLink.on('click', roomLinkOnClickCallback);

    const $roomRemoveLink = $('<a href="javascript:void(0);" title="remove" class="recent-rooms-remove" data-room-name="' + room.roomName + '">');
    $roomRemoveLink.append($('<img class="image-theme-light" src="/static/' + BUILD_NUMBER + '/img/trash-2.svg" alt="remove"/>'));
    $roomRemoveLink.append($('<img class="image-theme-dark" src="/static/' + BUILD_NUMBER + '/img/trash-2-yellow.svg" alt="remove"/>'));

    $roomRemoveLink.on('click', function (e) {
        const removeLink = $(e.currentTarget);
        const roomName = removeLink.attr("data-room-name");

        let visitedRoomsStr = LOCAL_STORAGE.getItem(VISITED_ROOMS_LOCAL_STORAGE_KEY);

        if (!visitedRoomsStr) {
            visitedRoomsStr = "[]";
        }

        const newVisitedRooms = [];

        const visitedRooms = JSON.parse(visitedRoomsStr);

        for (let i = 0; i < visitedRooms.length; i++) {
            const room = visitedRooms[i];

            if (room.roomName !== roomName) {
                newVisitedRooms.push(room);
            }
        }

        storeVisitedRoomsArray(newVisitedRooms);

        removeLink.parent().remove();

        if (!$recentRoomsWr.children().length) {
            $recentRoomsWr.append('<p id="' + RECENT_ROOMS_EMPTY_BLOCK_ID + '">no rooms visited yet</p>');
        }
    });

    const $visitedAt = $('<p class="recent-room-last-visited" title="last visited">');
    $visitedAt.append($('<span class="recent-room-last-visited-date">' + getDateStrFromMills(room.visitedAt) + '</span>'));
    $visitedAt.append($('<span class="recent-room-last-visited-time">' + getTimeStrFromMills(room.visitedAt) + '</span>'));

    const $clearfix = $('<div class="clearfix">');

    $roomBlock.append($roomLink);
    $roomBlock.append($roomRemoveLink);
    $roomBlock.append($visitedAt);
    $roomBlock.append($clearfix);

    $recentRoomsWr.append($roomBlock);

    redrawThemeableImages();
}

function getClickCoordinatesInsideElement(element, event) {
    //click coordinates inside element
    const rect = element.getBoundingClientRect();
    const x = event.clientX - rect.left;
    const y = event.clientY - rect.top;

    return {x: x, y: y, rect: rect};
}

function getOffsetTop (elem) {
    let offsetTop = 0;

    do {
        if (!isNaN(elem.offsetTop)) {
            offsetTop += elem.offsetTop;
        }
    } while (elem = elem.offsetParent);

    return offsetTop;
}

function getTextSelectionStartNode () {
    if (window.getSelection().rangeCount > 0) {
        const range = window.getSelection().getRangeAt(0);

        return range.startContainer;
    }

    return null;
}

function isOnMobileScreenSize () {
    return currentViewportWidth < DESKTOP_MENU_SCREEN_MIN_WIDTH_PX;
}

function showSpinnerOverlay() {
    if ($spinnerOverlay.hasClass('d-none')) {
        $spinnerOverlay.removeClass('d-none');
    }
}

function hideSpinnerOverlay() {
    if (!$spinnerOverlay.hasClass('d-none')) {
        $spinnerOverlay.addClass('d-none');
    }
}

function scrollBlockBottom (block) {
    block.scrollIntoView(false);
}

function setShowChangeOverlayModeFlag (flag) {
    LOCAL_STORAGE.setItem(SHOW_WINDOW_MODE_KEY_ALERT_LOCAL_STORAGE_KEY, flag);
}

function getShowChangeOverlayModeFlag () {
    return LOCAL_STORAGE.getItem(SHOW_WINDOW_MODE_KEY_ALERT_LOCAL_STORAGE_KEY) === 'true';
}

function initUIColorTheme() {
    const $themeableBlocks = $(
        '.room-messages-wr, ' +
        '.room-messages-init-msg, ' +
        '.user-message-textarea-wr, ' +
        '.room-info-addit-mid-right-labels, ' +
        '.main-navbar, ' +
        '.room-info-addit-mid-right-buttons-leave, ' +
        '.main-sidebar-top-item-1, ' +
        '.picks-wr, ' +
        '.picks-messages-wr, ' +
        '.picks-bot-wr, ' +
        '.main-sidebar-wr, ' +
        '.share-room-btn, ' +
        '.recent-rooms-popup, ' +
        '.preferences-popup, ' +
        '#recent-rooms-wr, ' +
        '#preferences-wr, ' +
        '#client-agreement-wr, ' +
        '.references-block-wr' +
        '.room-form-join-create, ' +
        '.room-form-row, ' +
        '.home-common-room-suggestions-wr, ' +
        '.page-about-content-wr, ' +
        '.page-home-container-main-center-wr, ' +
        '.main-sidebar-mobile-wr, ' +
        '.cookie-warning, ' +
        '.cookie-warning-agree, ' +
        '.client-agreement-popup, ' +
        '.room-title-link-change-input, ' +
        '.room-info-addit-alt-share-manual-label, ' +
        '.room-info-addit-share-btn, ' +
        '.room-descr-title, ' +
        '.room-descr-text, ' +
        '.room-descr, ' +
        '.room-info-online-users, ' +
        '.room-info-online-title, ' +
        '#room-descr-change-link, ' +
        '#room-descr-change-input, ' +
        '.share-room-wr, ' +
        '.share-room-manual-url, ' +
        '#user-message-joined-as-change-link, ' +
        '#user-message-joined-as-change-input, ' +
        '.user-message-send-btn, ' +
        '.message-edit-cancel, ' +
        '.user-message-wr-toggle-folk-picks, ' +
        '.user-message-wr-toggle-mobile-menu, ' +
        '.room-info-overlay-key-wr, ' +
        '.user-message-collapse-compact-toggle-label, ' +
        '.search-bar-input, ' +
        '.room-info-wr-main-navbar, ' +
        '.bots-main-wr, ' +
        '.chat-scroll-overlay'
    );

    //light is default
    if (selectedUITheme === UI_THEME_DARK) {
        $body.addClass('theme-dark');
        $themeableBlocks.addClass('theme-dark');
    } else {
        $body.removeClass('theme-dark');
        $themeableBlocks.removeClass('theme-dark');
    }

    redrawThemeableImages();
}

function redrawThemeableImages () {
    if (selectedUITheme === UI_THEME_DARK) {
        $('.image-theme-light').addClass('d-none');
        $('.image-theme-dark').removeClass('d-none');
    } else {
        $('.image-theme-light').removeClass('d-none');
        $('.image-theme-dark').addClass('d-none');
    }
}

function showCookieWarning () {
    showCookieWarningPopup();
}

function showCookieWarningPopup () {
    $cookieWarning.css('display', 'block');
    $cookieWarningAgree.on('click', function () {
        LOCAL_STORAGE.setItem(COOKIES_ACCEPTED_LOCAL_STORAGE_KEY, new Date().getTime() + '');
        $cookieWarning.css('display', 'none');
    });
}

function activateUserMessageInputAfterWindowWasActivated () {
    //stub for non-room page
}

function deactivateUserMessageInputAfterWindowWasDeactivated () {
    //stub for non-room page
}

function changeRoomInfoOverlayKeyIndicator (newKeyName) {
    //stub for non-room page
}

function tryShowRoomInfoOverlayKeyWr () {
    //stub for non-room page
}

function cancelMessageTextSelectionMode () {
    //stub for non-room page
}