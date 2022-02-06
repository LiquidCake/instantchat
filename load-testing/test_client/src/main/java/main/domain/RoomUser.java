package main.domain;

import com.fasterxml.jackson.annotation.JsonProperty;

public class RoomUser {
    @JsonProperty("uId")
    private String userInRoomUUID;
    @JsonProperty("n")
    private String userName;
    @JsonProperty("an")
    private Boolean isAnonName;
    @JsonProperty("o")
    private Boolean isOnlineInRoom;

    public RoomUser() {}

    public RoomUser(final String userInRoomUUID, final String userName, final Boolean isAnonName, final Boolean isOnlineInRoom) {
        this.userInRoomUUID = userInRoomUUID;
        this.userName = userName;
        this.isAnonName = isAnonName;
        this.isOnlineInRoom = isOnlineInRoom;
    }

    public String getUserInRoomUUID() {
        return userInRoomUUID;
    }

    public void setUserInRoomUUID(String userInRoomUUID) {
        this.userInRoomUUID = userInRoomUUID;
    }

    public String getUserName() {
        return userName;
    }

    public void setUserName(String userName) {
        this.userName = userName;
    }

    public Boolean getAnonName() {
        return isAnonName;
    }

    public void setAnonName(Boolean anonName) {
        isAnonName = anonName;
    }

    public Boolean getOnlineInRoom() {
        return isOnlineInRoom;
    }

    public void setOnlineInRoom(Boolean isOnlineInRoom) {
        isOnlineInRoom = isOnlineInRoom;
    }

    @Override
    public String toString() {
        return "RoomUser{" +
                "userInRoomUUID='" + userInRoomUUID + '\'' +
                ", userName='" + userName + '\'' +
                ", isAnonName=" + isAnonName +
                ", isOnlineInRoom=" + isOnlineInRoom +
                '}';
    }
}
