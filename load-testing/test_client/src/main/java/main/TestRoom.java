package main;

import com.fasterxml.uuid.Generators;
import main.domain.*;
import main.exception.RoomCreationFailedException;
import main.util.Constants;
import main.util.Logging;
import main.util.Util;
import org.apache.commons.collections4.CollectionUtils;
import org.apache.commons.lang3.StringUtils;
import org.java_websocket.drafts.Draft_6455;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.net.URI;
import java.util.*;
import java.util.stream.Collectors;

public class TestRoom {
    private final Logger LOGGER = LoggerFactory.getLogger(this.getClass());


    private static final int ROOM_USERS_TOTAL_COUNT = Main.ROOM_USERS_ADDING_STEPS_NUM * Main.ROOM_USERS_ADDING_AMOUNT_PER_STEP;



    private String roomName;
    private final List<TestUser> testUsers = new ArrayList<>();
    //keeps message test unique identifier (taken from message body) and timestamp of when message was initially sent (send acknowledgement)
    //needed to track if message was dispatched to all users
    private final Map<String, Long> allSentRoomMessageUniquePartToSentAt = new HashMap<>();

    //keeps message test unique identifier (taken from message body) and timestamp of when message was received by each user
    private final Map<String, List<TextMessageTrackingInfo>> allReceivedRoomMessageUniquePartToReceivedAt = new HashMap<>();


    public void startTestRoom(final int roomNum) throws Exception {
        /* Create new room */

        this.roomName = "room-" + Generators.timeBasedGenerator().generate().toString();

        LOGGER.info(String.format("=== creating new room #%d / '%s'", roomNum, roomName));

        try {
            createRoom(this.roomName, roomNum);
        } catch (final Exception e) {
            Logging.logError(this.roomName, String.format("!!! failed to create room #%d", roomNum), e);
            throw e;
        }


        /* Create and start initial set of users (e.g. each 20sec add 10 users, 5 times total) */

        int nextRoomUserNumber = 0;

        //make ROOM_USERS_ADDING_STEPS_NUM iterations of adding test users to room
        for (int i = 0; i < Main.ROOM_USERS_ADDING_STEPS_NUM; i++) {

            //add ROOM_USERS_ADDING_AMOUNT_PER_STEP test users to room
            for (int j = 0; j < Main.ROOM_USERS_ADDING_AMOUNT_PER_STEP; j++) {
                final TestUser nextTestUser = new TestUser(this.allSentRoomMessageUniquePartToSentAt, this.allReceivedRoomMessageUniquePartToReceivedAt);
                this.testUsers.add(nextTestUser);

                final int userNum = nextRoomUserNumber;

                new Thread(() -> {
                    try {
                        nextTestUser.startTestUser(this.roomName, userNum);
                    } catch (Exception e) {
                        Logging.logError(this.roomName, String.format("!!! failed to start test user# '%d'", userNum));
                    }
                }).start();

                nextRoomUserNumber++;
            }

            Util.sleep(Main.ROOM_USERS_ADDING_STEP_DELAY_MS);
        }

        Logging.logInfo(this.roomName, String.format("=== room is fully populated (%d users). Will be active for next %ds",
                Main.ROOM_USERS_ADDING_STEPS_NUM * Main.ROOM_USERS_ADDING_AMOUNT_PER_STEP, Main.ROOM_LIFE_SPAN_MS / 1000));


        //let room exist for ROOM_LIFE_SPAN_MS more minutes
        Util.sleep(Main.ROOM_LIFE_SPAN_MS);


        /* Stop test users */

        for (TestUser testUser : this.testUsers) {
            testUser.sendStopSignal();
        }

        int stoppedChecks = 0;
        boolean allUsersStopped = false;

        while (stoppedChecks++ < 50) {
            if (this.testUsers.stream().allMatch(TestUser::isStopped)) {
                allUsersStopped = true;
            } else {
                Util.sleep(1000);
            }
        }

        if (!allUsersStopped) {
            final String errorMsg = String.format("failed to ensure all test users stopped after waiting: %s", this.testUsers);

            LOGGER.error(errorMsg);

            throw new RuntimeException(errorMsg);
        }


        /* Log results */

        final Set<Long> allReceivedMessageDelays = allReceivedRoomMessageUniquePartToReceivedAt.values()
                .stream()
                .flatMap(Collection::stream)
                .map(info -> info.isReceivedOnRoomLogin()
                        ? 0
                        : (info.getMessageReceivedAt() - this.allSentRoomMessageUniquePartToSentAt.get(info.getMessageUniquePart()))
                )
                .collect(Collectors.toSet());

        final long maxMessageDelay =
                Math.round(allReceivedMessageDelays.stream()
                        .mapToDouble(d -> d)
                        .max()
                        .orElse(-1));

        final long avgMessageDelay =
                Math.round(allReceivedMessageDelays.stream()
                        .mapToDouble(d -> d)
                        .average()
                        .orElse(-1));

        Logging.logInfo(this.roomName, String.format("## total messages sent to room: %d. Max delay: %dms, avg delay: %dms",
                this.allSentRoomMessageUniquePartToSentAt.size(),
                maxMessageDelay, avgMessageDelay
        ));


        final Map<String, Long> allSentRoomMessagesUniquePartToSentAtSorted = Util.sortByValue(this.allSentRoomMessageUniquePartToSentAt);

        //for each message sent to this room - find tracking of its retrieval by each user. Make sure all users got message and log delay between sent and receive time
        allSentRoomMessagesUniquePartToSentAtSorted.forEach((messageUniquePart, sentAt) -> {

            if (this.allReceivedRoomMessageUniquePartToReceivedAt.containsKey(messageUniquePart)) {
                final List<TextMessageTrackingInfo> messageTrackings = this.allReceivedRoomMessageUniquePartToReceivedAt.get(messageUniquePart);

                final int messageReceivedByUsersCount = messageTrackings.size();

                //all messages must be received by all users. Except for case when we add some users after room's messages list reached size limit and was already cut
                // - then earlier messages wont be delivered to users that joined after cut
                // (e.g. if limit is 1000 - messages with id < 500 wont get to users joined after 1st cut, messages with id < 1000 wont get to users joined after 2nd cut etc.)
                if (messageReceivedByUsersCount < ROOM_USERS_TOTAL_COUNT) {
                    Logging.logError(this.roomName, String.format("!! message was received by '%d' users instead of expected '%d'. MessageId: '%d'",
                            messageReceivedByUsersCount, ROOM_USERS_TOTAL_COUNT, messageTrackings.get(0).getMessageId()));
                }

                final Set<Long> messageDeliveryDelays = messageTrackings.stream()
                        .map(info -> info.isReceivedOnRoomLogin()
                                ? 0
                                : (info.getMessageReceivedAt() - sentAt)
                        )
                        .collect(Collectors.toSet());

                Logging.logInfo(this.roomName, String.format("# message sent at %s was received by '%s' users. Timings under: %dms, Avg: %dms",
                        Logging.getLogDateString(sentAt),
                        messageReceivedByUsersCount,
                        Collections.max(messageDeliveryDelays),
                        Math.round(messageDeliveryDelays.stream()
                                .mapToDouble(d -> d)
                                .average()
                                .orElse(-1))
                ));

            } else {
                Logging.logError(this.roomName, String.format("!! message was not received by any user: '%s'", messageUniquePart));
            }
        });

        testUsers.forEach(testUser -> {
            final List<Exception> wsErrors = testUser.getWsErrors();

            if (!CollectionUtils.isEmpty(wsErrors)) {
                Logging.logError(this.roomName, String.format("!! got websocket errors for user '%s': %s", testUser.getUserName(), wsErrors));
            }
        });

        Logging.logInfo(this.roomName, String.format("=== shutting down room #%d / '%s'", roomNum, this.roomName));
    }

    public void createRoom(final String newRoomName, final int roomNum) throws Exception {
        /* Query home page to create user session */

        String userSessionToken = null;
        final Map<String, List<String>> homePageHeaders;

        try {
            homePageHeaders = HttpFunctions.getHomePage();
        } catch (final Exception e) {
            Logging.logError(newRoomName, "!!! creating room failed. Failed to request home page");

            throw new RoomCreationFailedException(e);
        }

        final List<String> cookies = homePageHeaders.get("Set-Cookie");

        if (cookies != null) {
            userSessionToken = cookies.stream()
                    .filter(cookie -> cookie.startsWith(Constants.USER_SESSION_COOKIE_NAME + "="))
                    .map(cookie -> cookie.substring((Constants.USER_SESSION_COOKIE_NAME + "=").length()))
                    .findAny().orElse(null);
        }

        if (StringUtils.isBlank(userSessionToken)) {
            Logging.logError(newRoomName, "!!! creating room failed. Got empty user session cookie");

            throw new RoomCreationFailedException();
        }


        /* Query aux-srv to pick backend for this room */

        final PickBackendResponse pickBackendResponse;

        try {
            pickBackendResponse = HttpFunctions.getBackendInstanceByRoomName(newRoomName);
        } catch (final Exception e) {
            Logging.logError(newRoomName, String.format("!!! creating room failed. Failed to request aux-srv 'pick backend instance' for roomName '%s', Exception: '%s'", newRoomName, e.getMessage()));

            throw new RoomCreationFailedException(e);
        }

        final String roomBackendInstance = pickBackendResponse.getBackendInstanceAddr();


        /* Create room */

        final String userName = "user-room-creator-" + Generators.timeBasedGenerator().generate().toString();
        final String createRoomRequestId = Generators.timeBasedGenerator().generate().toString();

        final Map<String, String> wsHeaders = new HashMap<>();
        wsHeaders.put("origin", Constants.HTTP_PROTOCOL + Constants.SERVER_ROOT_ADDR);
        wsHeaders.put("Cookie", String.format("%s=%s; ", Constants.USER_SESSION_COOKIE_NAME, userSessionToken));

        MyWebSocketClient wsClient = new MyWebSocketClient(
                new URI(Constants.WS_ENDPOINT + roomBackendInstance),
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
            } else if (tries >= 100) {
                Logging.logError(newRoomName, "!!! creating room failed. Didn't connect ws after waiting");

                wsClient.closeBlocking();

                throw new RoomCreationFailedException();
            } else {
                tries += 1;
                Util.sleep(100);
            }
        }

        //send create message to websocket
        try {
            wsClient.send(Constants.OBJECT_MAPPER.writeValueAsString(
                    new InMessageFrame(
                            Command.RoomCreateJoin,
                            createRoomRequestId,
                            new RoomInfo(
                                    newRoomName,
                                    Constants.ROOM_PASSWORD
                            ),
                            userName,
                            null,
                            null,
                            null
                    )
            ));
        } catch (final Exception e) {
            Logging.logError(newRoomName, String.format("!!! creating room '%s' failed on backend '%s'", newRoomName, roomBackendInstance), e);

            wsClient.closeBlocking();

            throw new RoomCreationFailedException(e);
        }

        //create room, wait for 'request processed' ack
        while (true) {
            int tries = 0;

            if (Util.findMessageByExpectedCommandAndRequestId(Command.RequestProcessed, createRoomRequestId, wsClient.getMessages())
                    .isPresent()) {
                Logging.logInfo(newRoomName, String.format("=== created room #%d '%s' on backend '%s'", roomNum, newRoomName, roomBackendInstance));

                wsClient.closeBlocking();

                return;
            } else if (tries >= 100) {
                Logging.logError(newRoomName, "!!! creating room failed. Haven't got 'request processed' ack after waiting");

                wsClient.closeBlocking();

                throw new RoomCreationFailedException();
            } else {
                tries += 1;
                Util.sleep(100);
            }
        }
    }
}
