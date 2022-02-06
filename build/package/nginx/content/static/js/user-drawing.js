/* Constants */

const USER_DRAWING_TEXTURE_TOOL_DEFAULT_WIDTH = 25;
const USER_DRAWING_PENCIL_LINE_WIDTH_FOR_TEXT_INPUT = 4;

const USER_DRAWING_CANVAS_EDGE_LENGTH_SMALL = 220;
const USER_DRAWING_CANVAS_EDGE_LENGTH_MID = 400;
const USER_DRAWING_CANVAS_EDGE_LENGTH_LARGE = 700;

const USER_DRAWING_CANVAS_EDGE_LENGTH_THRESHOLD_MID_WIDTH = 500;
const USER_DRAWING_CANVAS_EDGE_LENGTH_THRESHOLD_MID_HEIGHT = 750;

const USER_DRAWING_CANVAS_EDGE_LENGTH_THRESHOLD_LARGE_WIDTH = 800;
const USER_DRAWING_CANVAS_EDGE_LENGTH_THRESHOLD_LARGE_HEIGHT = 1040;


/* Variables */

/* DOM objects references */

const $toggleUserInputDrawingButton = $('.user-message-wr-toggle-input-drawing');

const $drawingMainWr = $('.user-drawing-input-wr');

const $drawingZoomReduceBtn = $('#user-drawing-zoom-reduce');
const $drawingZoomIncreaseBtn = $('#user-drawing-zoom-increase');

const $drawingModeSelector = $('#drawing-mode-selector');
const $drawingTypeTextImg = $('.user-drawing-text-img');

const $drawingOverlay = $('.user-drawing-overlay');
const $drawingModalWr = $('.user-drawing-modal-wr');
const $drawingModalText = $('.user-drawing-modal-text');
const $drawingModalInput = $('#user-drawing-modal-input');

const $drawingModalBtn = $('#drawing-modal-btn');

const $toggleDrawingModeBtn = $('#drawing-mode');
const $clearBtn = $('#drawing-clear-canvas');
const $deleteSelectedBtn = $('#user-drawing-delete');
const $undoBtn = $('#user-drawing-undo');
const $sendBtn = $('#user-drawing-send');

const $drawingOptionsEl = $('#drawing-mode-options');
const $drawingColorEl = $('#drawing-color');
const $drawingShadowColorEl = $('#drawing-shadow-color');
const $drawingLineWidthEl = $('#drawing-line-width');
const $drawingShadowWidth = $('#drawing-shadow-width');
const $drawingShadowOffset = $('#drawing-shadow-offset');

const $drawModeTexture = $('#draw-mode-texure');

const $fabricJsLink = $('.fabric-js-link');


/* vars */

let userInputCanvas;
let lastSelectedBrush;


function initUserInputDrawing () {
    const canvasEdgeLength = pickCanvasEdgeLengthByCurrentScreenSize();

    userInputCanvas = new fabric.Canvas('user-input-canvas', {
        isDrawingMode: true,
        width:  canvasEdgeLength,
        height: canvasEdgeLength,
    });

    fabric.Object.prototype.transparentCorners = false;

    $drawingZoomReduceBtn.on('click', function () {
        let newCanvasEdgeLength;

        if (userInputCanvas.width === USER_DRAWING_CANVAS_EDGE_LENGTH_LARGE) {
            newCanvasEdgeLength = USER_DRAWING_CANVAS_EDGE_LENGTH_MID;

        } else if (userInputCanvas.width === USER_DRAWING_CANVAS_EDGE_LENGTH_MID) {
            newCanvasEdgeLength = USER_DRAWING_CANVAS_EDGE_LENGTH_SMALL;

        } else {
            return;
        }

        showDrawingModalWindow(
            'reducing canvas size will cut bottom-right part of your image, please confirm',
            'confirm',
            function () {
                changeUserInputCanvasSize(newCanvasEdgeLength);
            }
        );
    });

    $drawingZoomIncreaseBtn.on('click', function () {
        let newCanvasEdgeLength;

        if (userInputCanvas.width === USER_DRAWING_CANVAS_EDGE_LENGTH_SMALL) {
            newCanvasEdgeLength = USER_DRAWING_CANVAS_EDGE_LENGTH_MID;

        } else if (userInputCanvas.width === USER_DRAWING_CANVAS_EDGE_LENGTH_MID) {
            newCanvasEdgeLength = USER_DRAWING_CANVAS_EDGE_LENGTH_LARGE;

        } else {
            return;
        }

        if ((newCanvasEdgeLength + 50) > currentViewportWidth || (newCanvasEdgeLength + 100) > currentViewportHeight) {
            showDrawingModalWindow(
                'increased canvas size may be too large for your display, please confirm increasing the size',
                'confirm',
                function () {
                    changeUserInputCanvasSize(newCanvasEdgeLength);
                }
            );
        } else {
            changeUserInputCanvasSize(newCanvasEdgeLength);
        }
    });

    $clearBtn.on('click', function () {
        showDrawingModalWindow(
            'please confirm CLEARING picture',
            'confirm',
            function () {
                userInputCanvas.clear();
            }
        );
    });

    $sendBtn.on('click', function () {
        if (!isLoggedIn || isDrawingSendInProgress) {
            return;
        }

        isDrawingSendInProgress = true;

        const fileName = userInRoomUUID + new Date().getTime();
        const fileGroupPrefix = roomUUID;
        const imageAsBase64 = getImageBase64();

        uploadUserDrawingToFileServer(
            fileName,
            fileGroupPrefix,
            imageAsBase64,
            function () {
                sendUserDrawingMessage(fileName, fileGroupPrefix);
                hideUserInputDrawingBlock();

                isDrawingSendInProgress = false;
            },
            function () {
                showError(ERROR_NOTIFICATION_TEXT_FAILED_TO_UPLOAD_FILE);

                isDrawingSendInProgress = false;
            }
        );
    });

    $drawingTypeTextImg.on('click', function () {
        showDrawingModalWindow(
            'type your text below, click OK and then draw a line, you text will be attached to!',
            'to step 2',
            function () {
                $drawingModalInput.addClass('d-none');

                //switch to pencil
                const oldTool = $drawingModeSelector.val();
                const oldLineWidth = $drawingLineWidthEl.val();

                $drawingModeSelector.val("Pencil").change();

                $drawingLineWidthEl.val(USER_DRAWING_PENCIL_LINE_WIDTH_FOR_TEXT_INPUT);
                $($drawingLineWidthEl[0].previousSibling).text(USER_DRAWING_PENCIL_LINE_WIDTH_FOR_TEXT_INPUT);
                userInputCanvas.freeDrawingBrush.width = USER_DRAWING_PENCIL_LINE_WIDTH_FOR_TEXT_INPUT;

                userInputCanvas.once('before:path:created', function (opt) {
                    if (lastSelectedBrush !== 'PencilBrush') {
                        return;
                    }

                    let path = opt.path;
                    let pathInfo = fabric.util.getPathSegmentsInfo(path.path);
                    path.segmentsInfo = pathInfo;

                    let pathLength = pathInfo[pathInfo.length - 1].length;
                    let textStr = $drawingModalInput.val();
                    let fontSize = 2.5 * pathLength / textStr.length;
                    let text = new fabric.Text(textStr, {
                        fontSize: fontSize,
                        path: path,
                        top: path.top,
                        left: path.left
                    });

                    userInputCanvas.add(text);

                    //bring old tool back
                    bringBackOldDrawingToolAfterTextInput(oldTool, oldLineWidth);
                });
            }
        );
    });

    $drawingOverlay.on('click', function () {
        $drawingOverlay.addClass('d-none');
        $drawingModalWr.addClass('d-none');
    });

    $toggleDrawingModeBtn.on('click', function () {
        userInputCanvas.isDrawingMode = !userInputCanvas.isDrawingMode;

        if (userInputCanvas.isDrawingMode) {
            $toggleDrawingModeBtn.text('switch to selecting mode');
            $drawingOptionsEl.removeClass('d-none');

            $sendBtn.removeClass('user-drawing-send-in-selection-mode');
            $undoBtn.removeClass('d-none');
            $deleteSelectedBtn.addClass('d-none');
        } else {
            $toggleDrawingModeBtn.text('switch to drawing mode');
            $drawingOptionsEl.addClass('d-none');

            $sendBtn.addClass('user-drawing-send-in-selection-mode');
            $undoBtn.addClass('d-none');
            $deleteSelectedBtn.removeClass('d-none');
        }
    });

    $deleteSelectedBtn.on('click', function () {
        userInputCanvas.getActiveObjects().forEach((obj) => {
            userInputCanvas.remove(obj)
        });

        userInputCanvas.discardActiveObject();
        userInputCanvas.requestRenderAll();
    });

    $undoBtn.on('click', function () {
        if (userInputCanvas.getObjects().length) {
            let lastItemIndex = (userInputCanvas.getObjects().length - 1);
            let item = userInputCanvas.item(lastItemIndex);

            if (item &&
                (item.get('type') === 'path' || item.get('type') === 'group' || item.get('type') === 'text')) {
                userInputCanvas.remove(item);
                userInputCanvas.renderAll();
            }
        }
    });

    /* Declare special tools (brushes) */

    let vLinePatternBrush = new fabric.PatternBrush(userInputCanvas);

    vLinePatternBrush.getPatternSrc = function () {
        let patternCanvas = fabric.document.createElement('canvas');
        patternCanvas.width = patternCanvas.height = 10;

        let ctx = patternCanvas.getContext('2d');

        ctx.strokeStyle = this.color;
        ctx.lineWidth = 5;
        ctx.beginPath();
        ctx.moveTo(5, 0);
        ctx.lineTo(5, 10);
        ctx.closePath();
        ctx.stroke();

        return patternCanvas;
    };

    let hLinePatternBrush = new fabric.PatternBrush(userInputCanvas);
    hLinePatternBrush.getPatternSrc = function () {
        let patternCanvas = fabric.document.createElement('canvas');
        patternCanvas.width = patternCanvas.height = 10;

        let ctx = patternCanvas.getContext('2d');
        ctx.strokeStyle = this.color;
        ctx.lineWidth = 5;
        ctx.beginPath();
        ctx.moveTo(0, 5);
        ctx.lineTo(10, 5);
        ctx.closePath();
        ctx.stroke();

        return patternCanvas;
    };

    let squarePatternBrush = new fabric.PatternBrush(userInputCanvas);
    squarePatternBrush.getPatternSrc = function () {

        let squareEdgeLength = 10, squareDistance = 2;

        let patternCanvas = fabric.document.createElement('canvas');
        patternCanvas.width = patternCanvas.height = squareEdgeLength + squareDistance;
        let ctx = patternCanvas.getContext('2d');

        ctx.fillStyle = this.color;
        ctx.fillRect(0, 0, squareEdgeLength, squareEdgeLength);

        return patternCanvas;
    };

    let diamondPatternBrush = new fabric.PatternBrush(userInputCanvas);
    diamondPatternBrush.getPatternSrc = function () {

        let squareEdgeLength = 10, squareDistance = 5;
        let patternCanvas = fabric.document.createElement('canvas');
        let rect = new fabric.Rect({
            width: squareEdgeLength,
            height: squareEdgeLength,
            angle: 45,
            fill: this.color
        });

        let canvasWidth = rect.getBoundingRect().width;

        patternCanvas.width = patternCanvas.height = canvasWidth + squareDistance;
        rect.set({ left: canvasWidth / 2, top: canvasWidth / 2 });

        let ctx = patternCanvas.getContext('2d');
        rect.render(ctx);

        return patternCanvas;
    };


    let img = new Image();

    let texturePatternBrush = new fabric.PatternBrush(userInputCanvas);
    texturePatternBrush.source = img;

    img.src = $drawModeTexture.attr('data-base-src') + 'texture1.png';


    /* on tool selection -  */

    $drawingModeSelector.on('change', function () {
        let isTextureMode = false;

        //special tools
        if (this.value === 'vline') {
            userInputCanvas.freeDrawingBrush = vLinePatternBrush;
            lastSelectedBrush = 'vline';

        } else if (this.value === 'hline') {
            userInputCanvas.freeDrawingBrush = hLinePatternBrush;
            lastSelectedBrush = 'hline';

        } else if (this.value === 'square') {
            userInputCanvas.freeDrawingBrush = squarePatternBrush;
            lastSelectedBrush = 'square';

        } else if (this.value === 'diamond') {
            userInputCanvas.freeDrawingBrush = diamondPatternBrush;
            lastSelectedBrush = 'diamond';

        } else if (this.value === 'texture') {
            isTextureMode = true;
            userInputCanvas.freeDrawingBrush = texturePatternBrush;
            lastSelectedBrush = 'texture';

            $drawingLineWidthEl.val(USER_DRAWING_TEXTURE_TOOL_DEFAULT_WIDTH);
            $($drawingLineWidthEl[0].previousSibling).text(USER_DRAWING_TEXTURE_TOOL_DEFAULT_WIDTH);
            userInputCanvas.freeDrawingBrush.width = USER_DRAWING_TEXTURE_TOOL_DEFAULT_WIDTH;

        } else {
            //any other tool
            userInputCanvas.freeDrawingBrush = new fabric[this.value + 'Brush'](userInputCanvas);
            lastSelectedBrush = this.value + 'Brush';
        }

        if (!isTextureMode) {
            initBrush();
        }
    });

    //select default tool
    userInputCanvas.freeDrawingBrush = new fabric.PencilBrush(userInputCanvas);
    initBrush();

    //input callbacks

    $drawingColorEl.on('input', function () {
        let brush = userInputCanvas.freeDrawingBrush;
        brush.color = this.value;

        if (brush.getPatternSrc) {
            brush.source = brush.getPatternSrc.call(brush);
        }
    });

    $drawingShadowColorEl.on('input', function () {
        userInputCanvas.freeDrawingBrush.shadow.color = this.value;
    });

    $drawingLineWidthEl.on('input', function () {
        userInputCanvas.freeDrawingBrush.width = parseInt(this.value, 10);
        $(this.previousSibling).text(this.value);
    });

    $drawingShadowWidth.on('input', function () {
        userInputCanvas.freeDrawingBrush.shadow.blur = parseInt(this.value, 10);
        $(this.previousSibling).text(this.value);
    });

    $drawingShadowOffset.on('input', function () {
        userInputCanvas.freeDrawingBrush.shadow.offsetX = parseInt(this.value, 10);
        userInputCanvas.freeDrawingBrush.shadow.offsetY = parseInt(this.value, 10);
        $(this.previousSibling).text(this.value);
    });

    $fabricJsLink.on('click', function (e) {
        stopPropagationAndDefault(e);

        showDrawingModalWindow(
            "you clicked on external link, leading to site of javascript graphics library 'fabric.js'. Please confirm proceeding",
            'confirm',
            function () {
                window.open($fabricJsLink.attr('href'), '_blank').focus();
            }
        );
    });
}

function initBrush () {
    let brush = userInputCanvas.freeDrawingBrush;
    brush.color = $drawingColorEl.val();

    if (brush.getPatternSrc) {
        brush.source = brush.getPatternSrc.call(brush);
    }

    brush.width = parseInt($drawingLineWidthEl.val(), 10);

    brush.shadow = new fabric.Shadow({
        blur: parseInt($drawingShadowWidth.val(), 10),
        offsetX: 0,
        offsetY: 0,
        affectStroke: true,
        color: $drawingShadowColorEl.val(),
    });
}

function bringBackOldDrawingToolAfterTextInput (oldTool, oldLineWidth) {
    $drawingModeSelector.val(oldTool).change();

    $drawingLineWidthEl.val(oldLineWidth);
    $($drawingLineWidthEl[0].previousSibling).text(oldLineWidth);
    userInputCanvas.freeDrawingBrush.width = parseInt(oldLineWidth, 10);
}

function toggleUserInputDrawingUI () {
    if ($drawingMainWr.hasClass('d-none')) {
        showUserInputDrawingBlock();
    } else {
        hideUserInputDrawingBlock();
    }
}

function showUserInputDrawingBlock () {
    cancelMessageTextSelectionMode();

    showMainOverlay();

    swipingBehaviourAllowed = false;

    $drawingMainWr.removeClass('d-none');

    adjustUserDrawingInputBlockTopCoord();

    $toggleUserInputDrawingButton.addClass('user-message-wr-toggle-input-drawing-active');
}

function hideUserInputDrawingBlock () {
    $drawingMainWr.addClass('d-none');
    swipingBehaviourAllowed = true;

    $toggleUserInputDrawingButton.removeClass('user-message-wr-toggle-input-drawing-active');

    hideMainOverlay();
}

function pickCanvasEdgeLengthByCurrentScreenSize () {
    if (currentViewportWidth >= USER_DRAWING_CANVAS_EDGE_LENGTH_THRESHOLD_LARGE_WIDTH &&
        currentViewportHeight >= USER_DRAWING_CANVAS_EDGE_LENGTH_THRESHOLD_LARGE_HEIGHT) {
        return USER_DRAWING_CANVAS_EDGE_LENGTH_LARGE;

    } else if (currentViewportWidth >= USER_DRAWING_CANVAS_EDGE_LENGTH_THRESHOLD_MID_WIDTH &&
        currentViewportHeight >= USER_DRAWING_CANVAS_EDGE_LENGTH_THRESHOLD_MID_HEIGHT) {
        return USER_DRAWING_CANVAS_EDGE_LENGTH_MID;

    } else {
        return USER_DRAWING_CANVAS_EDGE_LENGTH_SMALL;
    }
}

function changeUserInputCanvasSize(newCanvasEdgeLength) {
    userInputCanvas.setHeight(newCanvasEdgeLength);
    userInputCanvas.setWidth(newCanvasEdgeLength);

    userInputCanvas.renderAll();
}

function showDrawingModalWindow (titleText, buttonText, onClickCallback) {
    $drawingModalText.text(titleText);

    $drawingOverlay.removeClass('d-none');
    $drawingModalWr.removeClass('d-none');

    $drawingModalBtn.text(buttonText);

    $drawingModalBtn.off('click');
    $drawingModalBtn.on('click', function (e) {
        onClickCallback(e);

        $drawingOverlay.addClass('d-none');
        $drawingModalWr.addClass('d-none');
    });
}

function isUserDrawingInputBlockVisible () {
    return !$drawingMainWr.hasClass('d-none');
}

function adjustUserDrawingInputBlockTopCoord () {
    let userDrawingInputBlockVisible = isUserDrawingInputBlockVisible();

    if (userDrawingInputBlockVisible) {
        if ($drawingMainWr[0].getBoundingClientRect().top < currentViewportHeight * 0.1) {
            $drawingMainWr.css('top', '5%');
            $drawingMainWr.css('transform', 'translate(-50%, 0)');
        } else {
            $drawingMainWr.css('top', '');
            $drawingMainWr.css('transform', '');
        }
    }
}

function getImageBase64 () {
    return userInputCanvas.toDataURL({multiplier: 1, format: 'png'});
}

function loadImageToCanvas (imageContentBase64) {
    fabric.Image.fromURL(imageContentBase64, function(img) {
        userInputCanvas.add(img);
        userInputCanvas.renderAll();
    });
}
