
const URL_PARAMS = new URLSearchParams(window.location.search);

const URL_PARAM_BACKEND_HOST = 'backendHost';

let output;
let input;

let roomNameInput;
let roomPasswdInput;
let roomUserNameInput;

let messageTextInput;
let messageTextEditInput;
let messageTextDeleteInput;
let messageTextSupportRejectInput;

let textMessageOutput;

let srvErrorText;
let wsErrorText;


function init () {
    document.getElementById('rooms_list_link')
        .setAttribute("href", './ctrl_rooms?' + URL_PARAM_BACKEND_HOST + '=' + URL_PARAMS.get(URL_PARAM_BACKEND_HOST))

    output = document.getElementById("raw_output");
    input = document.getElementById("raw_input");

    roomNameInput = document.getElementById("r_name");
    roomPasswdInput = document.getElementById("r_passwd");
    roomUserNameInput = document.getElementById("r_user_name");

    messageTextInput = document.getElementById("text_message");
    messageTextEditInput = document.getElementById("text_message_edit");
    messageTextDeleteInput = document.getElementById("text_message_delete");
    messageTextSupportRejectInput = document.getElementById("text_message_support_reject");

    textMessageOutput = document.getElementById("text_msg_output");

    srvErrorText = document.getElementById("srv_error_txt");
    wsErrorText = document.getElementById("ws_error_txt");

    keepAlive();
}

function keepAlive () {
    console.log(getCurrentTimeStr() + ": start keepAlive, timeoutId: " + keepAliveTimerID);

    if (ws) {
        switch (ws.readyState) {
            case ws.OPEN:
                console.log(getCurrentTimeStr() + ": keepalive");
                ws.send(KEEP_ALIVE_BEACON);

                break
            case ws.CLOSED:
            case ws.CLOSING:
                console.log(getCurrentTimeStr() + ": socket closed");
                break
        }
    }

    keepAliveTimerID = setTimeout(keepAlive, KEEPALIVE_INTERVAL_MS);
    console.log(getCurrentTimeStr() + ": end keepAlive, timeoutId: " + keepAliveTimerID);
}

function cancelKeepAlive () {
    if (keepAliveTimerID) {
        clearTimeout(keepAliveTimerID);
    }
}

window.onbeforeunload = function (e) {
    cancelKeepAlive();
    if (ws) {
        ws.close(WS_CLOSE_CODE_NORMAL);
    }

    return null;
}
window.addEventListener(
    "beforeunload",
    function (e) {
        cancelKeepAlive();
        if (ws) {
            ws.close(WS_CLOSE_CODE_NORMAL);
        }

        return null;
        },
    false
);

function applyCallbacks () {
    ws.onopen = function(evt) {
        outputLog("++ OPENED" + "\n");

        sendData(ws,
            JSON.stringify({
                p: 'unknown'
            })
        );
    }

    ws.onclose = function(evt) {
        outputLog("++ CLOSED" + "\n");
    }

    ws.onmessage = function(evt) {
        let messageJson = evt.data;

        if (messageJson.startsWith("{\"c\":\"ER\"")) {
            const newErrBlock = document.createElement("div");
            newErrBlock.textContent = messageJson;

            srvErrorText.appendChild(newErrBlock);
        } else if (messageJson.startsWith("{\"c\":\"TM\"")) {
            let messageObj = JSON.parse(messageJson);

            if (messageObj.m.length < 1) {
                const newErrBlock = document.createElement("div");
                newErrBlock.textContent = "got zero size messages array";

                srvErrorText.appendChild(newErrBlock);

                return
            }

            let message = messageObj.m[0];

            const textMsgBlock = document.createElement("div");
            textMsgBlock.textContent = "[" + message.uId + " | " + getTimeStrFromNano(message.cAt) + "]: #"
                + message.id + " " + decodeURI(message.t);

            textMessageOutput.appendChild(textMsgBlock);

        } else if (messageJson.startsWith("{\"c\":\"ALL_TM\"")) {
            let messageObj = JSON.parse(messageJson);
            let roomMessages = messageObj.m;

            if (roomMessages) {
                roomMessages.forEach(function (nextMessage) {
                    const textMsgBlock = document.createElement("div");
                    textMsgBlock.textContent = "[" + nextMessage.uId + " | " + getTimeStrFromNano(nextMessage.cAt) + "]: #"
                        + nextMessage.id + " " + decodeURI(nextMessage.t);

                    textMessageOutput.appendChild(textMsgBlock);
                });
            }
        }

        outputLog("----------------------------");
        outputLog(messageJson);
    }

    ws.onerror = function(evt) {
        const newErrBlock = document.createElement("div");
        newErrBlock.textContent = evt.data;

        wsErrorText.append(newErrBlock);
    }
}

function initWebSocket () {
    if (ws) {
        ws.close(WS_CLOSE_CODE_NORMAL);
    }

    ws = new WebSocket(WS_PROTOCOL + URL_PARAMS.get(URL_PARAM_BACKEND_HOST) + "/ws_entry");

    applyCallbacks();
}

function sendData (ws, inputStr) {
    ws.send(inputStr);
    outputLog("++ SENT '" + inputStr + "', " + (inputStr ? inputStr.length : -1) + " symbols\n");

    return false;
}

function outputLog (message) {
    let d = document.createElement("div");
    d.textContent = message;

    output.prepend(document.createElement("br"));
    output.prepend(d);
}

function createOrJoinRoom () {
    initWebSocket();

    setTimeout(function () {
        sendData(ws,
            JSON.stringify({
                c: "R_C_J",
                uN: roomUserNameInput.value,
                r: {
                    n: roomNameInput.value,
                    p: roomPasswdInput.value
                }
            })
        );
    }, 100);
}

function createRoom () {
    initWebSocket();

    setTimeout(function () {
        sendData(ws,
            JSON.stringify({
                c: "R_C",
                r: {
                    n: roomNameInput.value,
                    p: roomPasswdInput.value
                }
            })
        );
    }, 100);
}

function joinRoom () {
    initWebSocket();

    setTimeout(function () {
        sendData(ws,
            JSON.stringify({
                c: "R_J",
                uN: roomUserNameInput.value,
                r: {
                    n: roomNameInput.value,
                    p: roomPasswdInput.value
                }
            })
        );
    }, 100);
}

function changeRoomUserName () {
    setTimeout(function () {
        sendData(ws,
            JSON.stringify({
                c: "R_CH_UN",
                uN: roomUserNameInput.value,
                r: {
                    n: roomNameInput.value,
                }
            })
        );
    }, 100);
}

function sendMsg () {
    setTimeout(function () {
        let messageEditId = messageTextEditInput.value;
        let messageDeleteId = messageTextDeleteInput.value;

        if (messageEditId) {
            sendData(ws,
                JSON.stringify({
                    c: "TM_E",
                    r: {
                        n: roomNameInput.value
                    },
                    m: {
                        id: parseInt(messageEditId),
                        t: messageTextInput.value
                    }
                })
            );

        } else if (messageDeleteId) {
            sendData(ws,
                JSON.stringify({
                    c: "TM_D",
                    r: {
                        n: roomNameInput.value
                    },
                    m: {
                        id: parseInt(messageDeleteId)
                    }
                })
            );

        } else {
            sendData(ws,
                JSON.stringify({
                    c: "TM",
                    r: {
                        n: roomNameInput.value
                    },
                    m: {
                        t: encodeURI(messageTextInput.value)
                    }
                })
            );
        }
    }, 100);
}

function supportRejectMsg (isSupport) {
    setTimeout(function () {
        let messageId = messageTextSupportRejectInput.value;

        sendData(ws,
            JSON.stringify({
                c: "TM_S_R",
                r: {
                    n: roomNameInput.value
                },
                srM: isSupport,
                m: {
                    id: parseInt(messageId)
                }
            })
        );
    }, 100);
}


// Callbacks

function onManualReconnect () {
    initWebSocket();

    return false;
}

function onManualSend () {
    if (!ws) {
        initWebSocket();
    }

    setTimeout(function () {
        sendData(ws, input.value)
    }, 100);

    return false;
}

function onChangeBackend () {
    const updatedQueryStr = replaceQueryParam(
        URL_PARAM_BACKEND_HOST,
        document.getElementById("backend_instance_host").value,
        window.location.search
    );

    window.location = window.location.protocol + '//' + window.location.host + window.location.pathname + updatedQueryStr;

    return false;
}


// Util

function getTimeStrFromNano (nanoTimestamp) {
    const date = new Date(Math.trunc(nanoTimestamp / 1000000));

    return date.getHours() + ":" + date.getMinutes() + ":" + date.getSeconds();
}

function getCurrentTimeStr () {
    const today = new Date();

    return today.getHours() + ":" + today.getMinutes() + ":" + today.getSeconds();
}

function getCookie (name) {
    const match = document.cookie.match(new RegExp('(^| )' + name + '=([^;]+)'));
    if (match) return match[2];
}

function setCookie (name, value) {
    document.cookie = name + '=' + value;
}

function replaceQueryParam (param, newval, search) {
    const regex = new RegExp("([?;&])" + param + "[^&;]*[;&]?");
    const query = search.replace(regex, "$1").replace(/&$/, '');

    return (query.length > 2 ? query + "&" : "?") + (newval ? param + "=" + newval : '');
}