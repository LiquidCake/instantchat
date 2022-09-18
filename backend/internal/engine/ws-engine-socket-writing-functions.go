package engine

import (
	"encoding/json"
	"instantchat.rooms/instantchat/backend/internal/domain_structures"
	"instantchat.rooms/instantchat/backend/internal/util"
	"github.com/gorilla/websocket"
	"math/rand"
	"strconv"
	"time"
)

const MessageWriteRandomDelayMaxValueMs = 20

func clientSocketMessageWritingRoutine(clSocket *domain_structures.WebSocket) {
	for {
		select {
		case outMessageWr, ok := <-clSocket.OutMessagesGetCh:
			//!ok means 'get' side of channel pair is closed. It may be closed only from within channel pair handler routine, after 'put' side is closed at WebSocket.Terminate()
			if !ok || clSocket.IsDead() {
				//make changes to room's active clients list
				removeDeadSocketFromRoom(clSocket.RelatedRoom, clSocket)

				return
			}

			frameJson := *outMessageWr.OutMessageJson

			roomName := "n/a"

			if clSocket.RelatedRoom != nil {
				roomName = clSocket.RelatedRoom.Name
			}

			util.LogTrace("writing to socket '%s', session '%s', room '%s'",
				clSocket.SocketUUID, clSocket.SessionUUID, roomName)

			//send message to socket with small random delay to spread IO time a bit
			select {
			case <-time.After(time.Duration(rand.Int31n(MessageWriteRandomDelayMaxValueMs)) * time.Millisecond):
				writeTimeout := SocketWriteTimeout
				err := writeMessageToSocketWithTimeout(clSocket.Socket, &frameJson, &writeTimeout)

				if err != nil {
					util.LogTrace("closing socket: error writing to socket '%s', session '%s', room '%s': '%s'",
						clSocket.SocketUUID, clSocket.SessionUUID, roomName, err)

					//marking socket as dead, closing
					clSocket.Terminate()

					//then - making changes to room's active clients list
					removeDeadSocketFromRoom(clSocket.RelatedRoom, clSocket)

					return
				}
			}
		}
	}
}

/**
* This method should be called only from clientSocketMessageWritingRoutine()
* (call it from other place only if synchronous call is really required)
*/
func writeMessageToSocketWithTimeout(socket *websocket.Conn, msgBytes *[]byte, timeout *time.Duration) error {
	_ = socket.SetWriteDeadline(time.Now().Add(*timeout))
	//send message to socket
	err := socket.WriteMessage(websocket.TextMessage, *msgBytes)

	return err
}

func writeFrameToActiveRoomMembers(
	messageDispatchingFrame *domain_structures.OutMessageFrame,
	room *domain_structures.Room,
	roomActiveClientSocketsByUUID *map[string]*domain_structures.WebSocket,
) {
	if len(*roomActiveClientSocketsByUUID) <= 0 {
		util.LogTrace("skipping writeFrameToActiveRoomMembers, no active client sockets found for room '%s'", room.Id)
		return
	}

	frameJson, err := json.Marshal(*messageDispatchingFrame)
	if err != nil {
		util.LogSevere("error serializing frame to JSON. Frame: '%s', error: '%s'", *messageDispatchingFrame, err)

		return
	}

	outMessage := &domain_structures.OutMessageWrapper{
		OutMessageJson: &frameJson,
	}

	for _, clSocket := range *roomActiveClientSocketsByUUID {
		clSocket.PutMessage(outMessage)
	}
}

func writeMembersListChangedFrameToActiveRoomMembers(room *domain_structures.Room, skipSocketUUID *string) {
	room.Lock()

	var roomActiveUsersCopy []domain_structures.RoomUserDTO

	for _, user := range room.AllRoomAuthorizedUsersBySessionUUID {
		roomActiveUsersCopy = append(roomActiveUsersCopy, domain_structures.RoomUserDTO{
			UserInRoomUUID: &(*user).UserInRoomUUID,
			UserName:       &(*user).UserName,
			IsAnonName:     &(*user).IsAnonName,
			IsOnlineInRoom: isUserOnlineInRoom(room, (*user).UserInRoomUUID),
		})
	}

	createdAt := time.Now().UnixNano()

	roomMembersListChangedDispatchingFrame := &domain_structures.OutMessageFrame{
		Command: domain_structures.RoomMembersChanged,
		CreatedAtNano: &createdAt,
		ActiveRoomUsers:     &roomActiveUsersCopy,
	}

	roomActiveClientSocketsByUUID := room.CopyActiveClientSocketMapNonLocking()

	room.Unlock()


	if skipSocketUUID != nil {
		if _, found := (*roomActiveClientSocketsByUUID)[*skipSocketUUID]; found {
			delete(*roomActiveClientSocketsByUUID, *skipSocketUUID)
		}

	}

	writeFrameToActiveRoomMembers(roomMembersListChangedDispatchingFrame, room, roomActiveClientSocketsByUUID)
}

func writeRoomDescriptionChangedFrameToActiveRoomMembers(room *domain_structures.Room, newServerStatus string) {
	room.Lock()

	//find room creator user
	var roomCreatorUserInRoomUUID *string = nil
	roomCreatorUser, found := room.AllRoomAuthorizedUsersBySessionUUID[room.CreatedBySessionUUID]

	if found {
		roomCreatorUserInRoomUUID = &roomCreatorUser.UserInRoomUUID
	}

	roomDescriptionChangedDispatchingFrame := &domain_structures.OutMessageFrame{
		Command:                   domain_structures.RoomChangeDescription,
		RoomCreatorUserInRoomUUID: roomCreatorUserInRoomUUID,
		ServerStatus:              &ServerStatus,
		Message:                   &[]domain_structures.RoomMessageDTO{
                                 { Text: &room.Description },
                               },
	}

	roomActiveClientSocketsByUUID := room.CopyActiveClientSocketMapNonLocking()

	room.Unlock()

	writeFrameToActiveRoomMembers(roomDescriptionChangedDispatchingFrame, room, roomActiveClientSocketsByUUID)
}

func writeNotificationToActiveRoomMembers(
	command domain_structures.Command,
	room *domain_structures.Room,
	roomActiveClientSocketsByUUID *map[string]*domain_structures.WebSocket,
) {
	createdAt := time.Now().UnixNano()

	notificationDispatchingFrame := &domain_structures.OutMessageFrame{
		Command: command,
		CreatedAtNano: &createdAt,
	}

	writeFrameToActiveRoomMembers(notificationDispatchingFrame, room, roomActiveClientSocketsByUUID)
}

func writeErrorMessageToSocket(
	clSocket *domain_structures.WebSocket,
	error domain_structures.WsError,
	requestId *string,
) {
	doWriteErrorMessageToSocket(clSocket, error, requestId, false, nil)
}

func doWriteErrorMessageToSocket(
	clSocket *domain_structures.WebSocket,
	error domain_structures.WsError,
	requestId *string,
	isSyncWriteRequired bool,
	syncWriteTimeout *time.Duration,
) {
	createdAt := time.Now().UnixNano()
	errorCodeStr := strconv.Itoa(error.Code)

	errorFrame := domain_structures.OutMessageFrame{
		Command:       domain_structures.Error,
		CreatedAtNano: &createdAt,
		RequestId:     requestId,
		Message:       &[]domain_structures.RoomMessageDTO{
			{Text: &errorCodeStr},
		},
	}

	frameJson, err := json.Marshal(errorFrame)
	if err != nil {
		util.LogSevere("error serializing frame to JSON. Frame: '%s', error: '%s'", errorFrame, err)

		return
	}

	//only for special cases
	if isSyncWriteRequired {
		//error is ignored
		_ = writeMessageToSocketWithTimeout(clSocket.Socket, &frameJson, syncWriteTimeout)
	} else {
		clSocket.PutMessage(&domain_structures.OutMessageWrapper{
			OutMessageJson: &frameJson,
		})
	}
}

func writeRequestProcessedToSocket(clSocket *domain_structures.WebSocket, requestId *string) {
	writeRequestProcessedToSocketWithAdditInfo(clSocket, nil, requestId, nil, nil, nil, nil)
}

func writeRequestProcessedToSocketWithAdditInfo(
	clSocket *domain_structures.WebSocket,
	createdAt *int64,
	requestId *string,
	processingDetails *string,
	roomId *string,
	userInRoomId *string,
	currentBuildNumber *string,
) {
	requestProcessedFrame := domain_structures.OutMessageFrame{
		Command:            domain_structures.RequestProcessed,
		CreatedAtNano:      createdAt,
		RequestId:          requestId,
		ProcessingDetails:  processingDetails,
		RoomUUID:           roomId,
		UserInRoomUUID:     userInRoomId,
		CurrentBuildNumber: currentBuildNumber,
	}

	frameJson, err := json.Marshal(requestProcessedFrame)
	if err != nil {
		util.LogSevere("error serializing frame to JSON. Frame: '%s', error: '%s'", requestProcessedFrame, err)

		return
	}

	clSocket.PutMessage(&domain_structures.OutMessageWrapper{
		OutMessageJson: &frameJson,
	})
}

func writeAfterRoomJoinMessagesToSocket(
		roomMembersListChangedFrame *domain_structures.OutMessageFrame,
		allMessagesFrame *domain_structures.OutMessageFrame,
	    roomDescriptionFrame *domain_structures.OutMessageFrame,
		clSocket *domain_structures.WebSocket,
	) error {
	roomMembersListChangedFrameJson, err := json.Marshal(roomMembersListChangedFrame)
	if err != nil {
		util.LogSevere("error serializing frame to JSON. Frame: '%s', error: '%s'", roomMembersListChangedFrame, err)

		return err
	}
	allMessagesFrameJson, err := json.Marshal(allMessagesFrame)
	if err != nil {
		util.LogSevere("error serializing frame to JSON. Frame: '%s', error: '%s'", allMessagesFrame, err)

		return err
	}
	roomDescriptionFrameJson, err := json.Marshal(roomDescriptionFrame)
	if err != nil {
		util.LogSevere("error serializing frame to JSON. Frame: '%s', error: '%s'", roomDescriptionFrame, err)

		return err
	}

	clSocket.PutMessage(&domain_structures.OutMessageWrapper{
		OutMessageJson: &roomMembersListChangedFrameJson,
	})

	clSocket.PutMessage(&domain_structures.OutMessageWrapper{
		OutMessageJson: &allMessagesFrameJson,
	})

	clSocket.PutMessage(&domain_structures.OutMessageWrapper{
		OutMessageJson: &roomDescriptionFrameJson,
	})

	return nil
}

func removeDeadSocketFromRoom(room *domain_structures.Room, clSocket *domain_structures.WebSocket)  {
	//if this socket was connected to some room - make changes to room's active clients list
	//and notify room's members about user quiting

	if room != nil {
		room.Lock()

		//if socket is still in room
		if _, found := room.ActiveClientSocketsByUUID[clSocket.SocketUUID]; found {
			//remove dead socket from room's sockets list
			delete(room.ActiveClientSocketsByUUID, clSocket.SocketUUID)

			//check if user already reconnected to room. If not - remove user from room's users list
			userAlreadyHasOtherActiveSocket := false
			for _, activeSocket := range room.ActiveClientSocketsByUUID {
				if !activeSocket.IsDead() && activeSocket.SessionUUID == clSocket.SessionUUID {
					userAlreadyHasOtherActiveSocket = true
				}
			}

			if userAlreadyHasOtherActiveSocket {
				room.Unlock()
			} else {
				delete(room.ActiveRoomUserUUIDBySessionUUID, clSocket.SessionUUID)
				room.ActiveRoomUsersLen = len(room.ActiveRoomUserUUIDBySessionUUID)

				room.Unlock()

				writeMembersListChangedFrameToActiveRoomMembers(room, nil)
			}
		} else {
			room.Unlock()
		}
	}
}
