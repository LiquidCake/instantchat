package main;

import com.fasterxml.uuid.Generators;
import main.domain.*;
import main.exception.RoomJoinFailedException;
import main.exception.RoomMessageSendFailedException;
import main.util.Constants;
import main.util.Logging;
import main.util.Util;
import org.apache.commons.lang3.StringUtils;
import org.java_websocket.drafts.Draft_6455;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.net.URI;
import java.util.*;

public class TestUser {
    private final Logger LOGGER = LoggerFactory.getLogger(this.getClass());
    //each message will have unique prepended to it
    private static final String MESSAGE_TEXT_TPL = Constants.MESSAGE_TEXT_UNIQUE_PART_SPLITTER +  "test text message aasdasd asdsadada dadadadsa d " +
            "adsa sad adsadsadsadasasdasdasdasd 1111111111111 " +
            "asdsasdaasdasdsadasdasdsadas dsadsa !@#$_@!(@!$(! asdsad";

    private String roomName;
    private String userName;
    private String userSessionToken;

    private volatile boolean stopSent;
    private volatile boolean stopped;

    private final Map<String, Long> allSentRoomMessageUniquePartToSentAt;
    private final Map<String, List<TextMessageTrackingInfo>> allReceivedRoomMessageUniquePartToReceivedAt;

    private MyWebSocketClient wsClient;

    public TestUser(
            final Map<String, Long> allSentRoomMessageUniquePartToSentAt, 
            final Map<String, List<TextMessageTrackingInfo>> allReceivedRoomMessageUniquePartToReceivedAt
    ) {
        this.allSentRoomMessageUniquePartToSentAt = allSentRoomMessageUniquePartToSentAt;
        this.allReceivedRoomMessageUniquePartToReceivedAt = allReceivedRoomMessageUniquePartToReceivedAt;
    }

    public void startTestUser(final String roomName, final int userRoomNumber) throws Exception {
        this.roomName = roomName;
        this.userName = String.format("user-%d-%s", userRoomNumber, Generators.timeBasedGenerator().generate().toString());

        this.stopSent = false;
        this.stopped = false;

        /* Query home page to create user session */

        try {
            final Map<String, List<String>> homePageHeaders = HttpFunctions.getHomePage();
            final List<String> cookies = homePageHeaders.get("Set-Cookie");

            if (cookies != null) {
                userSessionToken = cookies.stream()
                        .filter(cookie -> cookie.startsWith(Constants.USER_SESSION_COOKIE_NAME + "="))
                        .map(cookie -> cookie.substring((Constants.USER_SESSION_COOKIE_NAME + "=").length()))
                        .findAny().orElse(null);
            }

            if (StringUtils.isBlank(userSessionToken)) {
                Logging.logError(roomName, String.format("!!! failed starting test user '%s' for room '%s'. Got empty user session cookie", this.userName, this.roomName));

                throw new RoomJoinFailedException();
            }
        } catch (final Exception e) {
            Logging.logError(roomName, String.format("!!! failed starting test user '%s' for room '%s'. Failed to request home page", this.userName, this.roomName));

            throw new RoomJoinFailedException(e);
        }

        /* Query aux-srv to pick backend for this room */

        final PickBackendResponse pickBackendResponse;

        try {
            pickBackendResponse = HttpFunctions.getBackendInstanceByRoomName(this.roomName);
        } catch (final Exception e) {
            Logging.logError(roomName, String.format("!!! failed starting test user '%s' for room '%s'. Failed to request aux-srv 'pick backend instance'", this.userName, this.roomName));

            throw new RoomJoinFailedException(e);
        }

        final String roomBackendInstance = pickBackendResponse.getBackendInstanceAddr();


        /* Join room */

        final String joinRoomRequestId = Generators.timeBasedGenerator().generate().toString();

        final Map<String, String> wsHeaders = new HashMap<>();
        wsHeaders.put("origin", Constants.HTTP_PROTOCOL + Constants.SERVER_ROOT_ADDR);
        wsHeaders.put("Cookie", String.format("%s=%s; ", Constants.USER_SESSION_COOKIE_NAME, userSessionToken));

        this.wsClient = new MyWebSocketClient(
                new URI(Constants.WS_PROTOCOL + roomBackendInstance + Constants.WS_ENDPOINT),
                new Draft_6455(),
                wsHeaders,
                Constants.SOCKET_TIMEOUT_MS,
                this.roomName,
                userName,
                this.allReceivedRoomMessageUniquePartToReceivedAt
        );

        wsClient.connect();

        //wait for connect
        while (true) {
            int tries = 0;

            if (wsClient.isConnected()) {
                break;
            } else if (tries >= 50) {
                Logging.logError(roomName, String.format("!!! failed starting test user '%s' for room '%s'. Didn't connect ws after waiting", this.userName, this.roomName));

                wsClient.closeBlocking();

                return;
            } else {
                tries += 1;
                Util.sleep(100);
            }
        }

        // send join message to websocket
        try {
            wsClient.send(Constants.OBJECT_MAPPER.writeValueAsString(
                    new InMessageFrame(
                            Command.RoomCreateJoin,
                            joinRoomRequestId,
                            new RoomInfo(
                                    this.roomName,
                                    Constants.ROOM_PASSWORD
                            ),
                            userName,
                            null,
                            null,
                            null
                    )
            ));
        } catch (final Exception e) {
            Logging.logError(roomName, String.format("!!! failed to start test user (join room) '%s' on backend '%s'", this.roomName, roomBackendInstance), e);

            wsClient.closeBlocking();

            throw new RoomJoinFailedException(e);
        }

        //join room, wait for 'request processed' ack
        while (true) {
            int tries = 0;

            if (Util.findMessageByExpectedCommandAndRequestId(Command.RequestProcessed, joinRoomRequestId, wsClient.getMessages())
                    .isPresent()) {
                Logging.logTrace(roomName, String.format("--- started test user '%s' for room '%s' on backend '%s'", userName, this.roomName, roomBackendInstance));

                break;
            } else if (tries >= 50) {
                Logging.logError(roomName, "!!! failed to start test user (join room). Haven't got 'request processed' ack after waiting");

                wsClient.closeBlocking();

                throw new RoomJoinFailedException();
            } else {
                tries += 1;
                Util.sleep(100);
            }
        }

        //send text messages until stopped

        while (!stopSent) {
            final String nextTextMessageRequestId = Generators.timeBasedGenerator().generate().toString();
            final String messageUniquePart = Generators.timeBasedGenerator().generate().toString();

            //send new text message to websocket
            try {
                //track message send event
                synchronized (this.allSentRoomMessageUniquePartToSentAt) {
                    this.allSentRoomMessageUniquePartToSentAt.put(messageUniquePart, new Date().getTime());
                }

                wsClient.send(Constants.OBJECT_MAPPER.writeValueAsString(
                        new InMessageFrame(
                                Command.TextMessage,
                                nextTextMessageRequestId,
                                new RoomInfo(
                                        this.roomName,
                                        Constants.ROOM_PASSWORD
                                ),
                                null,
                                new RoomMessage(
                                        messageUniquePart + MESSAGE_TEXT_TPL
                                ),
                                null,
                                null
                        )
                ));

            } catch (final Exception e) {
                Logging.logError(roomName, String.format("!!! failed to send text message, room '%s', backend '%s'", this.roomName, roomBackendInstance), e);

                wsClient.closeBlocking();

                throw new RoomMessageSendFailedException(e);
            }

            //wait for 'request processed' ack
            while (true) {
                int tries = 0;

                final Optional<OutMessageFrame> messageFrameOpt = Util.findMessageByExpectedCommandAndRequestId(
                        Command.RequestProcessed, nextTextMessageRequestId, wsClient.getMessages()
                );

                if (messageFrameOpt.isPresent()) {
                    Logging.logTrace(roomName, String.format("--- sent text message. User '%s', room '%s', backend '%s'", this.userName, this.roomName, roomBackendInstance));

                    break;
                } else if (tries >= 50) {
                    Logging.logError(roomName, "!!! failed to send text message. Haven't got 'request processed' ack after waiting");

                    wsClient.closeBlocking();

                    throw new RoomMessageSendFailedException();
                } else {
                    tries += 1;
                    Util.sleep(100);
                }
            }

            Util.sleep(
                    new Random().nextInt(Constants.SEND_MESSAGE_DELAY_MS) + 2000
            );
        }

        wsClient.closeBlocking();

        this.stopped = true;
    }

    public List<Exception> getWsErrors() {
        return this.wsClient.getErrors();
    }

    public void sendStopSignal() {
        this.stopSent = true;
        this.wsClient.sendStopSignal();
    }

    public boolean isStopped() {
        return stopped;
    }

    public String getUserName() {
        return userName;
    }

    @Override
    public String toString() {
        return "TestUser{" +
                "roomName='" + roomName + '\'' +
                ", userName='" + userName + '\'' +
                ", stopped=" + stopped +
                '}';
    }
}
