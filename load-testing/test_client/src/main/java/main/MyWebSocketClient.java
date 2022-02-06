package main;

import com.fasterxml.jackson.core.JsonProcessingException;
import main.domain.*;
import main.util.Constants;
import main.util.Logging;
import main.util.Util;
import org.java_websocket.client.WebSocketClient;
import org.java_websocket.drafts.Draft;
import org.java_websocket.enums.ReadyState;
import org.java_websocket.handshake.ServerHandshake;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.net.URI;
import java.util.ArrayList;
import java.util.Date;
import java.util.List;
import java.util.Map;

public class MyWebSocketClient extends WebSocketClient {
    private final Logger LOGGER = LoggerFactory.getLogger(this.getClass());

    private static final int KEEPALIVE_INTERVAL_MS = 5000;

    private final List<Exception> errors = new ArrayList<>();
    private final List<OutMessageFrame> messages = new ArrayList<>();

    private final Map<String, List<TextMessageTrackingInfo>> messageUniquePartToReceivedAt;

    private final String roomName;
    private final String roomUserId;
    private long lastKeepAliveSentAt = new Date(0).getTime();

    private volatile boolean connected;
    private volatile boolean stopSent;

    private final String keepAliveMessageJson;

    public MyWebSocketClient(final URI serverUri,
                             final Draft protocolDraft,
                             final Map<String, String> httpHeaders,
                             final int connectTimeout,
                             final String roomName,
                             final String roomUserId,
                             final Map<String, List<TextMessageTrackingInfo>> messageUniquePartToReceivedAt
    ) {
        super(serverUri, protocolDraft, httpHeaders, connectTimeout);
        this.roomName = roomName;
        this.roomUserId = roomUserId;
        this.messageUniquePartToReceivedAt = messageUniquePartToReceivedAt;

        try {
            this.keepAliveMessageJson = Constants.OBJECT_MAPPER.writeValueAsString(
                    new InMessageFrame(
                            null,
                            null,
                            null,
                            null,
                            null,
                            null,
                            "OK"
                    )
            );
        } catch (JsonProcessingException e) {
            Logging.logError(roomName, "!!! failed to parse keep-alive message", e);

            throw new RuntimeException();
        }
    }

    @Override
    public void onMessage(String message) {
        OutMessageFrame messageFrame = null;
        
        try {
            messageFrame = Constants.OBJECT_MAPPER.readValue(message, OutMessageFrame.class);

            Logging.logTrace(roomName, " ++++++ got message: " + messageFrame);
        } catch (JsonProcessingException e) {
            Logging.logError(roomName, "!!! error while parsing message from websocket", e);
        }
        
        if (messageFrame != null) {
            synchronized (messages) {
                messages.add(messageFrame);
            }

            //add incoming text messages to tracking list for each user, so later we can make sure all users received all messages and measure delay

            if (Command.TextMessage.equals(messageFrame.getCommand())) {
                final long messageReceivedAt = new Date().getTime();
                final RoomMessage roomMessage = messageFrame.getMessage()[0];

                final String messageUniquePart = roomMessage.getText()
                        .split(Constants.MESSAGE_TEXT_UNIQUE_PART_SPLITTER)[0];

                synchronized (this.messageUniquePartToReceivedAt) {
                    if (!this.messageUniquePartToReceivedAt.containsKey(messageUniquePart)) {
                        this.messageUniquePartToReceivedAt.put(messageUniquePart, new ArrayList<>());
                    }

                    final List<TextMessageTrackingInfo> messageReceptionTrackingList = this.messageUniquePartToReceivedAt.get(messageUniquePart);

                    messageReceptionTrackingList.add(
                            new TextMessageTrackingInfo(
                                    messageUniquePart,
                                    roomMessage.getId(),
                                    messageReceivedAt,
                                    roomUserId,
                                    false
                            )
                    );
                }
            }

            if (Command.AllTextMessages.equals(messageFrame.getCommand())) {
                final long messageReceivedAt = new Date().getTime();

                for (RoomMessage roomMessage : messageFrame.getMessage()) {
                    final String messageUniquePart = roomMessage.getText()
                            .split(Constants.MESSAGE_TEXT_UNIQUE_PART_SPLITTER)[0];

                    synchronized (this.messageUniquePartToReceivedAt) {
                        if (!this.messageUniquePartToReceivedAt.containsKey(messageUniquePart)) {
                            this.messageUniquePartToReceivedAt.put(messageUniquePart, new ArrayList<>());
                        }

                        final List<TextMessageTrackingInfo> messageReceptionTrackingList = this.messageUniquePartToReceivedAt.get(messageUniquePart);

                        messageReceptionTrackingList.add(
                                new TextMessageTrackingInfo(
                                        messageUniquePart,
                                        roomMessage.getId(),
                                        messageReceivedAt,
                                        roomUserId,
                                        true
                                )
                        );
                    }
                }
            }
        }
    }

    @Override
    public void onOpen(ServerHandshake handshake) {
        this.connected = true;

        startKeepAlive();

        Logging.logTrace(roomName, "$$$ opened connection");
    }

    @Override
    public void onClose(int code, String reason, boolean remote) {
        this.connected = false;

        Logging.logTrace(roomName, "$$$ closed connection");
    }

    @Override
    public void onError(final Exception e) {
        synchronized (errors) {
            errors.add(e);
        }

        Logging.logError(roomName, "!!! got error from ws", e);
    }

    public void startKeepAlive() {
        Thread thread = new Thread(() -> {
            while (connected && !stopSent) {
                if (getReadyState() == ReadyState.OPEN && new Date().getTime() - lastKeepAliveSentAt > KEEPALIVE_INTERVAL_MS) {
                    send(keepAliveMessageJson);
                    lastKeepAliveSentAt = new Date().getTime();
                } else {
                    Util.sleep(50);
                }
            }
        }, "KeepAlive-" + roomUserId);

        thread.start();
    }

    public List<Exception> getErrors() {
        synchronized (errors) {
            return new ArrayList<>(errors);
        }
    }

    public List<OutMessageFrame> getMessages() {
        synchronized (messages) {
            return new ArrayList<>(messages);
        }
    }

    public boolean isConnected() {
        return connected;
    }

    public void sendStopSignal() {
        this.stopSent = true;
    }
}
