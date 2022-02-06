package main.exception;

public class RoomCreationFailedException extends Exception {
    public RoomCreationFailedException() {
    }

    public RoomCreationFailedException(String message) {
        super(message);
    }

    public RoomCreationFailedException(String message, Throwable cause) {
        super(message, cause);
    }

    public RoomCreationFailedException(Throwable cause) {
        super(cause);
    }
}
