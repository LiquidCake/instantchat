package main.exception;

public class RoomJoinFailedException extends Exception {
    public RoomJoinFailedException() {
    }

    public RoomJoinFailedException(String message) {
        super(message);
    }

    public RoomJoinFailedException(String message, Throwable cause) {
        super(message, cause);
    }

    public RoomJoinFailedException(Throwable cause) {
        super(cause);
    }
}
