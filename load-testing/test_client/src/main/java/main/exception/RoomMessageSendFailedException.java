package main.exception;

public class RoomMessageSendFailedException extends Exception {
    public RoomMessageSendFailedException() {
    }

    public RoomMessageSendFailedException(String message) {
        super(message);
    }

    public RoomMessageSendFailedException(String message, Throwable cause) {
        super(message, cause);
    }

    public RoomMessageSendFailedException(Throwable cause) {
        super(cause);
    }
}
