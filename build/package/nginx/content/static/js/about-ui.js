/* DOM objects references */

const $pageAboutMainWr = $('.page-about-main-wr');

/* vars */

let onResizeTimeout;

/* visibility flags */

let isInMenuMobileMode = false;

function initAboutPageUI () {
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

    $mainOverlay.on('click', function () {
        hideMenuMobile();
        hideMyRecentRoomsPopup();
    });

    const xwiper = new Xwiper(document);
    xwiper.onSwipeRight(showMenuMobileOnSwipe);

    //load visited rooms
    loadVisitedRooms(function (e) {
        let roomName = $(e.currentTarget).text();

        LOCAL_STORAGE.setItem(REDIRECT_VARIABLE_LOCAL_STORAGE_KEY, JSON.stringify({
            redirectFrom: REDIRECT_FROM_ANY_PAGE_ON_RECENT_ROOM_CLICK,
            redirectedAt: new Date().getTime(),
            roomName: roomName,
        }));

        redirectToURL("/");
    });

    $document.on('keydown', function(e) {
        if (e.which === KEY_CODE_ESCAPE) {
            hideMyRecentRoomsPopup();
        }
    });
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
            $pageAboutMainWr.removeClass('col-md-10').addClass('col-md-12');
        }

        isInMenuMobileMode = true;
    } else {
        //transition from mobile to desktop
        if (isInMenuMobileMode) {
            hideMenuMobile();
        }

        //show desktop sidebar
        $sidebarWr.addClass('d-md-block');
        $pageAboutMainWr.removeClass('col-md-12').addClass('col-md-10');


        isInMenuMobileMode = false;
    }
}
