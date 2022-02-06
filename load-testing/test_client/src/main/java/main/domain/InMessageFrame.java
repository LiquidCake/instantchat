package main.domain;

import com.fasterxml.jackson.annotation.JsonProperty;

public class InMessageFrame {
    @JsonProperty("c")
    private String command;
    @JsonProperty("rq")
    private String requestId;
    @JsonProperty("r")
    private RoomInfo roomInfo;
    @JsonProperty("uN")
    private String userName;

    @JsonProperty("m")
    private RoomMessage message;

    @JsonProperty("srM")
    private Boolean supportMessage;

    @JsonProperty("kA")
    private String keepAliveBeacon;

    public InMessageFrame() {}

    public InMessageFrame(String command,
                          String requestId,
                          RoomInfo roomInfo,
                          String userName,
                          RoomMessage message,
                          Boolean supportMessage,
                          String keepAliveBeacon) {
        this.command = command;
        this.requestId = requestId;
        this.roomInfo = roomInfo;
        this.userName = userName;
        this.message = message;
        this.supportMessage = supportMessage;
        this.keepAliveBeacon = keepAliveBeacon;
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

    public RoomInfo getRoomInfo() {
        return roomInfo;
    }

    public void setRoomInfo(RoomInfo roomInfo) {
        this.roomInfo = roomInfo;
    }

    public String getUserName() {
        return userName;
    }

    public void setUserName(String userName) {
        this.userName = userName;
    }

    public RoomMessage getMessage() {
        return message;
    }

    public void setMessage(RoomMessage message) {
        this.message = message;
    }

    public Boolean getSupportMessage() {
        return supportMessage;
    }

    public void setSupportMessage(Boolean supportMessage) {
        this.supportMessage = supportMessage;
    }

    public String getKeepAliveBeacon() {
        return keepAliveBeacon;
    }

    public void setKeepAliveBeacon(String keepAliveBeacon) {
        this.keepAliveBeacon = keepAliveBeacon;
    }

    @Override
    public String toString() {
        return "InMessageFrame{" +
                "command='" + command + '\'' +
                ", requestId='" + requestId + '\'' +
                ", roomInfo=" + roomInfo +
                ", userName='" + userName + '\'' +
                ", message=" + message +
                ", supportMessage=" + supportMessage +
                ", keepAliveBeacon='" + keepAliveBeacon + '\'' +
                '}';
    }
}
