package main.domain;

import com.fasterxml.jackson.annotation.JsonProperty;

public class RoomMessage {
    @JsonProperty("id")
    private Long id;
    @JsonProperty("t")
    private String text;
    @JsonProperty("sC")
    private Integer supportedCount;
    @JsonProperty("rC")
    private Integer rejectedCount;
    @JsonProperty("lE")
    private Long lastEditedAt;
    @JsonProperty("lV")
    private Long lastVotedAt;
    @JsonProperty("rU")
    private String replyToUserId;
    @JsonProperty("rM")
    private Long replyToMessageId;
    @JsonProperty("uId")
    private String userInRoomUUID;
    @JsonProperty("cAt")
    private Long createdAtNano;

    public RoomMessage() {}

    public RoomMessage(final String text) {
        this.text = text;
    }

    public RoomMessage(final Long id,
                       final String text,
                       final Integer supportedCount,
                       final Integer rejectedCount,
                       final Long lastEditedAt,
                       final Long lastVotedAt,
                       final String replyToUserId,
                       final Long replyToMessageId,
                       final String userInRoomUUID,
                       final Long createdAtNano) {
        this.id = id;
        this.text = text;
        this.supportedCount = supportedCount;
        this.rejectedCount = rejectedCount;
        this.lastEditedAt = lastEditedAt;
        this.lastVotedAt = lastVotedAt;
        this.replyToUserId = replyToUserId;
        this.replyToMessageId = replyToMessageId;
        this.userInRoomUUID = userInRoomUUID;
        this.createdAtNano = createdAtNano;
    }

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getText() {
        return text;
    }

    public void setText(String text) {
        this.text = text;
    }

    public Integer getSupportedCount() {
        return supportedCount;
    }

    public void setSupportedCount(Integer supportedCount) {
        this.supportedCount = supportedCount;
    }

    public Integer getRejectedCount() {
        return rejectedCount;
    }

    public void setRejectedCount(Integer rejectedCount) {
        this.rejectedCount = rejectedCount;
    }

    public Long getLastEditedAt() {
        return lastEditedAt;
    }

    public void setLastEdited(Long lastEditedAt) {
        this.lastEditedAt = lastEditedAt;
    }

    public Long getLastVotedAt() {
        return lastVotedAt;
    }

    public void setLastVotedAt(Long lastVotedAt) {
        this.lastVotedAt = lastVotedAt;
    }

    public String getReplyToUserId() {
        return replyToUserId;
    }

    public void setReplyToUserId(String replyToUserId) {
        this.replyToUserId = replyToUserId;
    }

    public Long getReplyToMessageId() {
        return replyToMessageId;
    }

    public void setReplyToMessageId(Long replyToMessageId) {
        this.replyToMessageId = replyToMessageId;
    }

    public String getUserInRoomUUID() {
        return userInRoomUUID;
    }

    public void setUserInRoomUUID(String userInRoomUUID) {
        this.userInRoomUUID = userInRoomUUID;
    }

    public Long getCreatedAtNano() {
        return createdAtNano;
    }

    public void setCreatedAtNano(Long createdAtNano) {
        this.createdAtNano = createdAtNano;
    }

    @Override
    public String toString() {
        return "RoomMessage{" +
                "id=" + id +
                ", text='" + text + '\'' +
                ", supportedCount=" + supportedCount +
                ", rejectedCount=" + rejectedCount +
                ", lastEditedAt=" + lastEditedAt +
                ", lastVotedAt=" + lastVotedAt +
                ", replyToUserId=" + replyToUserId +
                ", replyToMessageId=" + replyToMessageId +
                ", userInRoomUUID='" + userInRoomUUID + '\'' +
                ", createdAtNano=" + createdAtNano +
                '}';
    }
}
