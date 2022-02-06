package main.domain;

import com.fasterxml.jackson.annotation.JsonProperty;

import java.util.Arrays;

public class OutMessageFrame {
    @JsonProperty("c")
    private String command;
    @JsonProperty("rq")
    private String requestId;
    @JsonProperty("pd")
    private String processingDetails;
    @JsonProperty("rId")
    private String roomUUID;
    @JsonProperty("uId")
    private String userInRoomUUID;
    @JsonProperty("rCuId")
    private String roomCreatorUserInRoomUUID;
    @JsonProperty("cAt")
    private Long createdAtNano;
    @JsonProperty("bN")
    private String currentBuildNumber;

    @JsonProperty("m")
    private RoomMessage[] message;

    @JsonProperty("rU")
    private RoomUser[] roomUsers;

    public OutMessageFrame() {}

    public OutMessageFrame(String command,
                           String requestId,
                           String processingDetails,
                           String userInRoomUUID,
                           String roomCreatorUserInRoomUUID,
                           Long createdAtNano,
                           String currentBuildNumber,
                           RoomMessage[] message,
                           RoomUser[] roomUsers) {
        this.command = command;
        this.requestId = requestId;
        this.processingDetails = processingDetails;
        this.userInRoomUUID = userInRoomUUID;
        this.roomCreatorUserInRoomUUID = roomCreatorUserInRoomUUID;
        this.createdAtNano = createdAtNano;
        this.currentBuildNumber = currentBuildNumber;
        this.message = message;
        this.roomUsers = roomUsers;
    }

    public String getCommand() {
        return command;
    }

    public void setCommand(String command) {
        this.command = command;
    }

    public String getRequestId() {
        return requestId;
    }

    public void setRequestId(String requestId) {
        this.requestId = requestId;
    }

    public String getProcessingDetails() {
        return processingDetails;
    }

    public void setProcessingDetails(String processingDetails) {
        this.processingDetails = processingDetails;
    }

    public String getUserInRoomUUID() {
        return userInRoomUUID;
    }

    public void setUserInRoomUUID(String userInRoomUUID) {
        this.userInRoomUUID = userInRoomUUID;
    }

    public String getRoomCreatorUserInRoomUUID() {
        return roomCreatorUserInRoomUUID;
    }

    public void setRoomCreatorUserInRoomUUID(String roomCreatorUserInRoomUUID) {
        this.roomCreatorUserInRoomUUID = roomCreatorUserInRoomUUID;
    }

    public Long getCreatedAtNano() {
        return createdAtNano;
    }

    public void setCreatedAtNano(Long createdAtNano) {
        this.createdAtNano = createdAtNano;
    }

    public String getCurrentBuildNumber() {
        return currentBuildNumber;
    }

    public void setCurrentBuildNumber(String currentBuildNumber) {
        this.currentBuildNumber = currentBuildNumber;
    }

    public RoomMessage[] getMessage() {
        return message;
    }

    public void setMessage(RoomMessage[] message) {
        this.message = message;
    }

    public RoomUser[] getRoomUsers() {
        return roomUsers;
    }

    public void setRoomUsers(RoomUser[] roomUsers) {
        this.roomUsers = roomUsers;
    }

    public String getRoomUUID() {
        return roomUUID;
    }

    public void setRoomUUID(String roomUUID) {
        this.roomUUID = roomUUID;
    }

    @Override
    public String toString() {
        return "OutMessageFrame{" +
                "command='" + command + '\'' +
                ", requestId='" + requestId + '\'' +
                ", processingDetails='" + processingDetails + '\'' +
                ", roomUUID='" + roomUUID + '\'' +
                ", userInRoomUUID='" + userInRoomUUID + '\'' +
                ", roomCreatorUserInRoomUUID='" + roomCreatorUserInRoomUUID + '\'' +
                ", createdAtNano=" + createdAtNano +
                ", currentBuildNumber=" + currentBuildNumber +
                ", message=" + Arrays.toString(message) +
                ", roomUsers=" + Arrays.toString(roomUsers) +
                '}';
    }
}
