package main.domain;

public class TextMessageTrackingInfo {
    private final String messageUniquePart;
    private final long messageId;
    private final long messageReceivedAt;
    private final String userName;
    private final boolean receivedOnRoomLogin;

    public TextMessageTrackingInfo(final String messageUniquePart, final long messageId, final long messageReceivedAt, final String userName, final boolean receivedOnRoomLogin) {
        this.messageUniquePart = messageUniquePart;
        this.messageId = messageId;
        this.messageReceivedAt = messageReceivedAt;
        this.userName = userName;
        this.receivedOnRoomLogin = receivedOnRoomLogin;
    }

    public String getMessageUniquePart() {
        return messageUniquePart;
    }

    public long getMessageId() {
        return messageId;
    }

    public long getMessageReceivedAt() {
        return messageReceivedAt;
    }

    public String getUserName() {
        return userName;
    }

    public boolean isReceivedOnRoomLogin() {
        return receivedOnRoomLogin;
    }

    @Override
    public String toString() {
        return "TextMessageTrackingInfo{" +
                "messageUniquePart='" + messageUniquePart + '\'' +
                ", messageId=" + messageId +
                ", messageReceivedAt=" + messageReceivedAt +
                ", userName='" + userName + '\'' +
                ", receivedOnRoomLogin=" + receivedOnRoomLogin +
                '}';
    }
}
