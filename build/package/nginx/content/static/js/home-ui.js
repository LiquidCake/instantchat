/* DOM objects references */

const $pageHomeMainWr = $('.page-home-main-wr');
const $homeContainerMainCenterWr = $('.page-home-container-main-center-wr');

const $roomFormTitle = $('.room-form-title');

const $roomJoinCreateBtn = $('#room-join-create');

/* vars */

let onResizeTimeout;

/* visibility flags */

let isInMenuMobileMode = false;

let roomJoinCreateBtnDisabled = false;

function initHomePageUI () {
    commonPageUIInit();

    onResizeActions();
    $window.on('resize', onResizeDelayedFunc);

    /* Set Menu mobile and Recent rooms toggle behaviour */

    $pageCommonMobileMenuToggle.on('click', showMenuMobile);

    $recentRoomsLink.on('click', function () {
        hideMenuMobile();
        showMyRecentRoomsPopup();
    });

    $recentRoomsClose.on('click', function () {
        hideMyRecentRoomsPopup();
    });

    $preferencesLink.on('click', function () {
        hideMenuMobile();
        showPreferencesPopup();
    });

    $preferencesClose.on('click', function () {
        hidePreferencesPopup();
    });

    $mainOverlay.on('click', function () {
        hideMenuMobile();
        hideMyRecentRoomsPopup();
        hidePreferencesPopup();
    });

    $document.on('keydown', function(e) {
        if (e.which === KEY_CODE_ESCAPE) {
            hideMyRecentRoomsPopup();
            hidePreferencesPopup();
        }
    });

    const xwiper = new Xwiper(document);
    xwiper.onSwipeRight(showMenuMobileOnSwipe);
}

function onResizeDelayedFunc () {
    clearTimeout(onResizeTimeout);
    onResizeTimeout = setTimeout(onResizeActions, 50);
}

function onResizeActions () {
    storeCurrentViewportDimensions();

    if (isOnMobileScreenSize()) {
        //transition from desktop to mobile
        if (!isInMenuMobileMode) {
            //hide desktop sidebar
            $sidebarWr.removeClass('d-md-block');
            $pageHomeMainWr.removeClass('col-md-10').addClass('col-md-12');
        }

        isInMenuMobileMode = true;
    } else {
        //transition from mobile to desktop
        if (isInMenuMobileMode) {
            hideMenuMobile();
        }

        //show desktop sidebar
        $sidebarWr.addClass('d-md-block');
        $pageHomeMainWr.removeClass('col-md-12').addClass('col-md-10');


        isInMenuMobileMode = false;
    }
}

function disableCreateJoinButton (disable) {
    if (disable) {
        roomJoinCreateBtnDisabled = true;
        $roomJoinCreateBtn.css("opacity", "0.6")
    } else {
        roomJoinCreateBtnDisabled = false;
        $roomJoinCreateBtn.css("opacity", "1")
    }
}
