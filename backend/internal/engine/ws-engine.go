package engine

import (
	"errors"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"instantchat.rooms/instantchat/backend/internal/config"
	"instantchat.rooms/instantchat/backend/internal/domain_structures"

	"instantchat.rooms/instantchat/backend/internal/util"
)

/* Constants */

const SocketWriteTimeout = time.Second * 30
const SocketReadTimeout = time.Hour * 1
const SocketReadLimitBytes = 50000

const RoomCredsMinChars = 3
const RoomCredsMaxChars = 100

const RoomMaxUsersLimit = 100

var RoomCredsValidationErrorInvalidLength = errors.New("invalid room credentials length")
var RoomCredsValidationErrorNameForbidden = errors.New("room name forbidden")
var RoomCredsValidationErrorNameHasBadChars = errors.New("room name contains bad characters")

var allowedRoomNameSpecialChars = []string{
	"!",
	"@",
	"$",
	"*",
	"(",
	")",
	"_",
	"-",
	",",
	".",
	"~",
	"[",
	"]",
}

const RoomMessagesLimit = 1000

var RoomMessagesLimitApproachingWarningBreakpoints = []int{950, 975, 990}

/* Variables */

var hasher = util.PasswordHash{}

var wsUpgrader = websocket.Upgrader{
	HandshakeTimeout:  SocketWriteTimeout,
	EnableCompression: true,

	CheckOrigin: func(r *http.Request) bool {
		originHeader, found := r.Header["Origin"]
		return found && len(originHeader) == 1 && util.ArrayContainsString(config.AppConfig.AllowedOrigins, originHeader[0])
	},
}

var ServerStatus = util.ServerStatusOnline

// global rooms in-memory storage
var ActiveRoomsByNameMap = domain_structures.ActiveRoomsByName{
	ActiveRoomsByName: make(map[string]*domain_structures.Room),
}

// metrics
var RoomsOnlineGauge prometheus.Gauge
var UsersOnlineGauge prometheus.Gauge
var AvgUsersOnlineGauge prometheus.Gauge
var AvgMessagesPerRoomGauge prometheus.Gauge

func WsEntry(w http.ResponseWriter, r *http.Request) {
	var session util.HttpSession

	err := util.GetUserSession(r, &session)

	if err != nil {
		util.LogWarn("session cookie not found: '%s'", err)

		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	util.LogTrace("upgrading client for session '%s'", session.SessionUUID)

	//start websocket session through current TCP socket
	socketConn, err := wsUpgrader.Upgrade(w, r, nil)

	if err != nil {
		util.LogSevere("error while upgrading socket to ws: '%s'", err)

		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	if err := socketConn.SetReadDeadline(time.Now().Add(SocketReadTimeout)); err != nil {
		util.LogSevere("error setting read deadline to socket. error: '%s'", err)

		socketConn.Close()

		return
	}

	var initFrame domain_structures.InitFrame

	err = socketConn.ReadJSON(&initFrame)

	if err != nil {
		util.LogWarn("failed to read init frame: '%s'", err)

		socketConn.Close()

		return
	}

	//equivalent of buffered channel with unlimited size. Engine puts new messages that must be sent to user into 'put' channel
	//and then 'client socket message writing routine' takes messages from 'get' channel and sends them to user
	//Required since user potentially can receive messages slower than they are sent to room, and in this case we have no choice but to accumulate them
	outMessagesPutCh, outMessagesGetCh := makeFlexibleChannelPair()

	clSocket := &domain_structures.WebSocket{
		Socket:              socketConn,
		SocketUUID:          "",
		SessionUUID:         session.SessionUUID,
		LastKeepAliveSignal: time.Now().UnixNano(),
		OutMessagesPutCh:    outMessagesPutCh,
		OutMessagesGetCh:    outMessagesGetCh,
	}

	//start routine that waits for messages to send via channel
	go clientSocketMessageWritingRoutine(clSocket)

	defer clSocket.Terminate()

	//set read limit to restrict large messages
	socketConn.SetReadLimit(SocketReadLimitBytes)

	socketUUID, err := uuid.NewUUID()

	if err != nil {
		util.LogSevere("failed to generate UUID: '%s'", err)
		writeErrorMessageToSocket(clSocket, domain_structures.WsServerError, nil)

		return
	}

	clSocket.SocketUUID = socketUUID.String()

	for {
		var inFrame domain_structures.InMessageFrame

		err = clSocket.Socket.ReadJSON(&inFrame)

		if err != nil {
			if err == websocket.ErrReadLimit {
				util.LogSevere("error WsRoomMessageTooLargeError - got too large message. session '%s', error: '%s'", clSocket.SessionUUID, err)

				writeErrorMessageToSocket(clSocket, domain_structures.WsRoomMessageTooLargeError, nil)
			} else if strings.Contains(err.Error(), "invalid character") {
				util.LogSevere("error WsInvalidInput while parsing ws message. session '%s', error: '%s'", clSocket.SessionUUID, err)

				writeErrorMessageToSocket(clSocket, domain_structures.WsInvalidInput, nil)
			} else {
				util.LogTrace("error WsConnectionError (generic) while reading ws. session '%s', error: '%s'", clSocket.SessionUUID, err)

				writeErrorMessageToSocket(clSocket, domain_structures.WsConnectionError, nil)
			}

			return
		}

		/* Check if this is KeepAlive signal */

		if inFrame.KeepAliveBeacon == "OK" {
			clSocket.LastKeepAliveSignal = time.Now().UnixNano()
			util.LogTrace("KeepAlive: socket: '%s', session '%s'", clSocket.SocketUUID, clSocket.SessionUUID)

			continue
		}

		/* Process request as command */

		util.LogTrace("Got command from session '%s': '%s'", clSocket.SessionUUID, string(inFrame.Command))

		switch inFrame.Command {
		case domain_structures.RoomCreateJoin:
			fallthrough
		case domain_structures.RoomCreateJoinAuthorize:
			ActiveRoomsByNameMap.Lock()

			existingRoom := ActiveRoomsByNameMap.ActiveRoomsByName[inFrame.Room.Name]

			//log into room
			if existingRoom != nil {
				ActiveRoomsByNameMap.Unlock()

				logIntoRoom(existingRoom, clSocket, &inFrame, false)

			} else {
				//create room
				newRoom, err := createRoom(&inFrame, clSocket.SessionUUID)

				if err != nil {
					ActiveRoomsByNameMap.Unlock()

					if err == RoomCredsValidationErrorInvalidLength {
						util.LogTrace("invalid room credentials length: '%s'", inFrame.Room.Name)

						writeErrorMessageToSocket(clSocket, domain_structures.WsRoomCredsValidationErrorBadLength, inFrame.RequestId)
					} else if err == RoomCredsValidationErrorNameForbidden {
						util.LogTrace("room name is forbidden: '%s'", inFrame.Room.Name)

						writeErrorMessageToSocket(clSocket, domain_structures.WsRoomCredsValidationErrorNameForbidden, inFrame.RequestId)
					} else if err == RoomCredsValidationErrorNameHasBadChars {
						util.LogTrace("room name contains bad characters: '%s'", inFrame.Room.Name)

						writeErrorMessageToSocket(clSocket, domain_structures.WsRoomCredsValidationErrorNameHasBadChars, inFrame.RequestId)
					} else {
						util.LogSevere("error while creating room: '%s'", err)

						writeErrorMessageToSocket(clSocket, domain_structures.WsServerError, inFrame.RequestId)
					}

					continue
				}

				ActiveRoomsByNameMap.ActiveRoomsByName[newRoom.Name] = newRoom
				RoomsOnlineGauge.Inc()

				ActiveRoomsByNameMap.Unlock()

				util.LogInfo("user '%s' created room '%s' / '%s'", clSocket.SessionUUID, newRoom.Id, newRoom.Name)

				StartSocketHouseKeeper(newRoom)

				//log into room
				logIntoRoom(newRoom, clSocket, &inFrame, true)
			}

		case domain_structures.RoomCreate:
			ActiveRoomsByNameMap.Lock()

			if ActiveRoomsByNameMap.ContainsNonLocking(inFrame.Room.Name) {
				ActiveRoomsByNameMap.Unlock()

				util.LogWarn("failed to create room - already exists: '%s'", inFrame.Room.Name)
				writeErrorMessageToSocket(clSocket, domain_structures.WsRoomExists, inFrame.RequestId)

				continue
			}

			newRoom, err := createRoom(&inFrame, clSocket.SessionUUID)

			if err != nil {
				ActiveRoomsByNameMap.Unlock()

				if err == RoomCredsValidationErrorInvalidLength {
					util.LogTrace("invalid room credentials length: '%s'", inFrame.Room.Name)

					writeErrorMessageToSocket(clSocket, domain_structures.WsRoomCredsValidationErrorBadLength, inFrame.RequestId)
				} else if err == RoomCredsValidationErrorNameForbidden {
					util.LogTrace("room name is forbidden: '%s'", inFrame.Room.Name)

					writeErrorMessageToSocket(clSocket, domain_structures.WsRoomCredsValidationErrorNameForbidden, inFrame.RequestId)
				} else if err == RoomCredsValidationErrorNameHasBadChars {
					util.LogTrace("room name contains bad characters: '%s'", inFrame.Room.Name)

					writeErrorMessageToSocket(clSocket, domain_structures.WsRoomCredsValidationErrorNameHasBadChars, inFrame.RequestId)
				} else {
					util.LogSevere("error while creating room: '%s'", err)

					writeErrorMessageToSocket(clSocket, domain_structures.WsServerError, inFrame.RequestId)
				}

				continue
			}

			ActiveRoomsByNameMap.ActiveRoomsByName[newRoom.Name] = newRoom
			RoomsOnlineGauge.Inc()

			ActiveRoomsByNameMap.Unlock()

			util.LogInfo("user '%s' created room '%s' / '%s'", clSocket.SessionUUID, newRoom.Id, newRoom.Name)

			StartSocketHouseKeeper(newRoom)

			writeRequestProcessedToSocket(clSocket, inFrame.RequestId)

		case domain_structures.RoomJoin:
			room := ActiveRoomsByNameMap.Get(inFrame.Room.Name)

			if room == nil {
				util.LogTrace("room '%s' not found", inFrame.Room.Name)
				writeErrorMessageToSocket(clSocket, domain_structures.WsRoomNotFound, inFrame.RequestId)

				continue
			}

			logIntoRoom(room, clSocket, &inFrame, false)

		case domain_structures.RoomChangeUserName:
			room := ActiveRoomsByNameMap.Get(inFrame.Room.Name)

			if room == nil {
				util.LogTrace("room '%s' not found", inFrame.Room.Name)
				writeErrorMessageToSocket(clSocket, domain_structures.WsRoomNotFound, inFrame.RequestId)

				continue
			}

			room.Lock()

			room.LastActiveAt = time.Now().UnixNano()

			if room.IsDeleted {
				room.Unlock()

				util.LogInfo("failed to change name for user '%s' - room '%s' was deleted", clSocket.SessionUUID, inFrame.Room.Name)
				writeErrorMessageToSocket(clSocket, domain_structures.WsRoomNotFound, inFrame.RequestId)

				continue
			}

			existingRoomUser, userFound := room.AllRoomAuthorizedUsersBySessionUUID[clSocket.SessionUUID]

			if !userFound {
				room.Unlock()

				util.LogInfo("failed to change room user name - user '%s' not authorized for room '%s'", clSocket.SessionUUID, inFrame.Room.Name)
				writeErrorMessageToSocket(clSocket, domain_structures.WsRoomNotAuthorized, inFrame.RequestId)

				continue
			}

			trimmedRoomUserName := strings.TrimSpace(inFrame.UserName)

			//validate provided user name or pick anonymous one
			roomUserName, isAnon, err := validateOrPickRoomUserName(trimmedRoomUserName, room)

			if err != nil {
				room.Unlock()

				if err == ProvidedNameTaken {
					util.LogTrace("room UserName '%s' already taken. Room: '%s'", trimmedRoomUserName, room.Id)
					writeErrorMessageToSocket(clSocket, domain_structures.WsRoomUserNameTaken, inFrame.RequestId)

					continue
				} else if err == BadNameLength {
					util.LogTrace("room UserName '%s' has wrong length. Room: '%s'", trimmedRoomUserName, room.Id)
					writeErrorMessageToSocket(clSocket, domain_structures.WsRoomUserNameValidationError, inFrame.RequestId)

					continue
				}
			}

			existingRoomUser.UserName = roomUserName
			existingRoomUser.IsAnonName = isAnon

			room.Unlock()

			writeMembersListChangedFrameToActiveRoomMembers(room, nil)

			writeRequestProcessedToSocket(clSocket, inFrame.RequestId)

		case domain_structures.RoomChangeDescription:
			room := ActiveRoomsByNameMap.Get(inFrame.Room.Name)

			if room == nil {
				util.LogTrace("room '%s' not found", inFrame.Room.Name)
				writeErrorMessageToSocket(clSocket, domain_structures.WsRoomNotFound, inFrame.RequestId)

				continue
			}

			room.Lock()

			room.LastActiveAt = time.Now().UnixNano()

			if room.IsDeleted {
				room.Unlock()

				util.LogInfo("failed to change room description for user '%s' - room '%s' was deleted", clSocket.SessionUUID, inFrame.Room.Name)
				writeErrorMessageToSocket(clSocket, domain_structures.WsRoomNotFound, inFrame.RequestId)

				continue
			}

			_, userFound := room.ActiveRoomUserUUIDBySessionUUID[clSocket.SessionUUID]

			if !userFound {
				room.Unlock()

				util.LogInfo("failed to change room description - user '%s' not active for room '%s'", clSocket.SessionUUID, inFrame.Room.Name)
				writeErrorMessageToSocket(clSocket, domain_structures.WsRoomNotAuthorized, inFrame.RequestId)

				continue
			}

			clientSocketForThisRoom, socketFound := room.ActiveClientSocketsByUUID[clSocket.SocketUUID]

			if !socketFound || clientSocketForThisRoom.IsDead() {
				room.Unlock()

				util.LogInfo("failed to change room description - socket '%s' not active for room '%s'", clSocket.SocketUUID, inFrame.Room.Name)
				writeErrorMessageToSocket(clSocket, domain_structures.WsConnectionError, inFrame.RequestId)

				continue
			}

			if clSocket.SessionUUID != room.CreatedBySessionUUID {
				room.Unlock()

				util.LogWarn("failed to change room description - user '%s' is not a creator of room '%d'", clSocket.SessionUUID, inFrame.Room.Name)
				writeErrorMessageToSocket(clSocket, domain_structures.WsInvalidInput, inFrame.RequestId)

				continue
			}

			trimmedNewDescription := strings.TrimSpace(inFrame.Message.Text)

			err := validateRoomDescription(trimmedNewDescription)

			if err != nil {
				room.Unlock()

				util.LogTrace("failed to change room description - invalid length. Room: '%s'", room.Name)
				writeErrorMessageToSocket(clSocket, domain_structures.WsRoomValidationErrorBadDescriptionLength, inFrame.RequestId)

				continue
			}

			util.LogTrace("user '%s' is changing description of room '%s'", clSocket.SessionUUID, room.Name)

			room.Description = trimmedNewDescription

			room.Unlock()

			writeRoomDescriptionChangedFrameToActiveRoomMembers(room, ServerStatus)

			writeRequestProcessedToSocket(clSocket, inFrame.RequestId)

		case domain_structures.TextMessage:
			room := ActiveRoomsByNameMap.Get(inFrame.Room.Name)

			if room == nil {
				util.LogInfo("failed to send message for user '%s' - room '%s' not found", clSocket.SessionUUID, inFrame.Room.Name)
				writeErrorMessageToSocket(clSocket, domain_structures.WsRoomNotFound, inFrame.RequestId)

				continue
			}

			room.Lock()

			room.LastActiveAt = time.Now().UnixNano()

			if room.IsDeleted {
				room.Unlock()

				util.LogInfo("failed to send message for user '%s' - room '%s' was deleted", clSocket.SessionUUID, inFrame.Room.Name)
				writeErrorMessageToSocket(clSocket, domain_structures.WsRoomNotFound, inFrame.RequestId)

				continue
			}

			userInRoomUUID, userFound := room.ActiveRoomUserUUIDBySessionUUID[clSocket.SessionUUID]

			if !userFound {
				room.Unlock()

				util.LogInfo("failed to send message - user '%s' not active for room '%s'", clSocket.SessionUUID, inFrame.Room.Name)
				writeErrorMessageToSocket(clSocket, domain_structures.WsRoomNotAuthorized, inFrame.RequestId)

				continue
			}

			clientSocketForThisRoom, socketFound := room.ActiveClientSocketsByUUID[clSocket.SocketUUID]

			if !socketFound || clientSocketForThisRoom.IsDead() {
				room.Unlock()

				util.LogInfo("failed to send message - socket '%s' not active for room '%s'", clSocket.SocketUUID, inFrame.Room.Name)
				writeErrorMessageToSocket(clSocket, domain_structures.WsConnectionError, inFrame.RequestId)

				continue
			}

			message := inFrame.Message

			util.LogTrace("user '%s' is sending message of len '%d' to room '%s' / '%s'",
				clSocket.SessionUUID, len(message.Text), room.Id, room.Name)

			//transform message and add to room messages array
			newRoomMessage := &domain_structures.RoomMessage{
				Id:               room.NextMessageId,
				Text:             message.Text,
				SupportedCount:   0,
				RejectedCount:    0,
				UserInRoomUUID:   userInRoomUUID,
				CreatedAtSec:     time.Now().Unix(),
				ReplyToUserId:    message.ReplyToUserId,
				ReplyToMessageId: message.ReplyToMessageId,
			}

			room.NextMessageId += 1

			room.RoomMessages[newRoomMessage.Id] = newRoomMessage
			room.RoomMessagesLen = len(room.RoomMessages)

			//check room messages amount. If approaching limit - notify users (later), if reached limit - cut messages array in half
			roomMessagesCurrentAmount := room.RoomMessagesLen

			if roomMessagesCurrentAmount == RoomMessagesLimit {
				shrinkRoomMessagesMap(room)
			}

			messageDispatchingFrame := &domain_structures.OutMessageFrame{
				Command: domain_structures.TextMessage,
				Message: &[]domain_structures.RoomMessageDTO{copyMessageAsDTO(newRoomMessage)},
			}

			//make copy of active client sockets connected to this room while under lock.
			//After unlock - initial list may be updated at any point by parallel routines
			roomActiveClientSocketsByUUID := room.CopyActiveClientSocketMapNonLocking()

			room.Unlock()

			//send new message to all active users, respond OK to user immediately
			writeFrameToActiveRoomMembers(messageDispatchingFrame, room, roomActiveClientSocketsByUUID)

			if roomMessagesCurrentAmount == RoomMessagesLimit {
				writeNotificationToActiveRoomMembers(domain_structures.NotifyMessagesLimitReached, room, roomActiveClientSocketsByUUID)
			} else if util.ArrayContainsInt(RoomMessagesLimitApproachingWarningBreakpoints, roomMessagesCurrentAmount) {
				writeNotificationToActiveRoomMembers(domain_structures.NotifyMessagesLimitApproaching, room, roomActiveClientSocketsByUUID)
			}

			writeRequestProcessedToSocket(clSocket, inFrame.RequestId)

		case domain_structures.TextMessageEdit:
			room := ActiveRoomsByNameMap.Get(inFrame.Room.Name)

			if room == nil {
				util.LogInfo("failed to edit message for user '%s' - room '%s' not found", clSocket.SessionUUID, inFrame.Room.Name)
				writeErrorMessageToSocket(clSocket, domain_structures.WsRoomNotFound, inFrame.RequestId)

				continue
			}

			room.Lock()

			room.LastActiveAt = time.Now().UnixNano()

			if room.IsDeleted {
				room.Unlock()

				util.LogInfo("failed to edit message for user '%s' - room '%s' was deleted", clSocket.SessionUUID, inFrame.Room.Name)
				writeErrorMessageToSocket(clSocket, domain_structures.WsRoomNotFound, inFrame.RequestId)

				continue
			}

			userInRoomUUID, userFound := room.ActiveRoomUserUUIDBySessionUUID[clSocket.SessionUUID]

			if !userFound {
				room.Unlock()

				util.LogInfo("failed to edit message - user '%s' not active for room '%s'", clSocket.SessionUUID, inFrame.Room.Name)
				writeErrorMessageToSocket(clSocket, domain_structures.WsRoomNotAuthorized, inFrame.RequestId)

				continue
			}

			clientSocketForThisRoom, socketFound := room.ActiveClientSocketsByUUID[clSocket.SocketUUID]

			if !socketFound || clientSocketForThisRoom.IsDead() {
				room.Unlock()

				util.LogInfo("failed to edit message - socket '%s' not active for room '%s'", clSocket.SocketUUID, inFrame.Room.Name)
				writeErrorMessageToSocket(clSocket, domain_structures.WsConnectionError, inFrame.RequestId)

				continue
			}

			message := inFrame.Message

			existingMessage, messageFound := room.RoomMessages[message.Id]

			if !messageFound {
				room.Unlock()

				writeRequestProcessedToSocket(clSocket, inFrame.RequestId)

				continue
			}

			if existingMessage.UserInRoomUUID != userInRoomUUID {
				room.Unlock()

				util.LogWarn("failed to edit message - user '%s' is not an author of message '%d' for room '%s'",
					clSocket.SessionUUID, existingMessage.Id, inFrame.Room.Name)
				writeErrorMessageToSocket(clSocket, domain_structures.WsInvalidInput, inFrame.RequestId)

				continue
			}

			util.LogTrace("user '%s' is editing message '%d' in room '%s' / '%s'",
				clSocket.SessionUUID, existingMessage.Id, room.Id, room.Name)

			existingMessage.Text = message.Text
			existingMessage.ReplyToUserId = message.ReplyToUserId
			existingMessage.ReplyToMessageId = message.ReplyToMessageId

			lastEditedAt := time.Now().UnixNano()
			existingMessage.LastEditedAt = &lastEditedAt

			messageEditDispatchingFrame := &domain_structures.OutMessageFrame{
				Command:       domain_structures.TextMessageEdit,
				Message:       &[]domain_structures.RoomMessageDTO{copyEditedMessageAsDTO(existingMessage)},
				CreatedAtNano: &lastEditedAt,
			}

			//make copy of active client sockets connected to this room while under lock.
			//After unlock - initial list may be updated at any point by parallel routines
			roomActiveClientSocketsByUUID := room.CopyActiveClientSocketMapNonLocking()

			room.Unlock()

			//send new message to all active users, respond OK to user immediately
			writeFrameToActiveRoomMembers(messageEditDispatchingFrame, room, roomActiveClientSocketsByUUID)

			writeRequestProcessedToSocket(clSocket, inFrame.RequestId)

		case domain_structures.TextMessageDelete:
			room := ActiveRoomsByNameMap.Get(inFrame.Room.Name)

			if room == nil {
				util.LogInfo("failed to delete message for user '%s' - room '%s' not found", clSocket.SessionUUID, inFrame.Room.Name)
				writeErrorMessageToSocket(clSocket, domain_structures.WsRoomNotFound, inFrame.RequestId)

				continue
			}

			room.Lock()

			room.LastActiveAt = time.Now().UnixNano()

			if room.IsDeleted {
				room.Unlock()

				util.LogInfo("failed to delete message for user '%s' - room '%s' was deleted", clSocket.SessionUUID, inFrame.Room.Name)
				writeErrorMessageToSocket(clSocket, domain_structures.WsRoomNotFound, inFrame.RequestId)

				continue
			}

			userInRoomUUID, userFound := room.ActiveRoomUserUUIDBySessionUUID[clSocket.SessionUUID]

			if !userFound {
				room.Unlock()

				util.LogInfo("failed to delete message - user '%s' not active for room '%s'", clSocket.SessionUUID, inFrame.Room.Name)
				writeErrorMessageToSocket(clSocket, domain_structures.WsRoomNotAuthorized, inFrame.RequestId)

				continue
			}

			clientSocketForThisRoom, socketFound := room.ActiveClientSocketsByUUID[clSocket.SocketUUID]

			if !socketFound || clientSocketForThisRoom.IsDead() {
				room.Unlock()

				util.LogInfo("failed to delete message - socket '%s' not active for room '%s'", clSocket.SocketUUID, inFrame.Room.Name)
				writeErrorMessageToSocket(clSocket, domain_structures.WsConnectionError, inFrame.RequestId)

				continue
			}

			message := inFrame.Message

			existingMessage, messageFound := room.RoomMessages[message.Id]

			if !messageFound {
				room.Unlock()

				writeRequestProcessedToSocket(clSocket, inFrame.RequestId)

				continue
			}

			if existingMessage.UserInRoomUUID != userInRoomUUID {
				room.Unlock()

				util.LogWarn("failed to delete message - user '%s' is not an author of message '%d' for room '%s'",
					clSocket.SessionUUID, existingMessage.Id, inFrame.Room.Name)
				writeErrorMessageToSocket(clSocket, domain_structures.WsInvalidInput, inFrame.RequestId)

				continue
			}

			util.LogTrace("user '%s' is deleting message '%d' in room '%s' / '%s'",
				clSocket.SessionUUID, existingMessage.Id, room.Id, room.Name)

			delete(room.RoomMessages, existingMessage.Id)

			room.RoomMessagesLen = len(room.RoomMessages)

			messageDeleteDispatchingFrame := &domain_structures.OutMessageFrame{
				Command: domain_structures.TextMessageDelete,
				Message: &[]domain_structures.RoomMessageDTO{
					{Id: &existingMessage.Id}, //no need in safe copy because Id field wont change
				},
			}

			//make copy of active client sockets connected to this room while under lock.
			//After unlock - initial list may be updated at any point by parallel routines
			roomActiveClientSocketsByUUID := room.CopyActiveClientSocketMapNonLocking()

			room.Unlock()

			//send new message to all active users, respond OK to user immediately
			writeFrameToActiveRoomMembers(messageDeleteDispatchingFrame, room, roomActiveClientSocketsByUUID)

			writeRequestProcessedToSocket(clSocket, inFrame.RequestId)

		case domain_structures.TextMessageSupportOrReject:
			isSupport := inFrame.SupportOrRejectMessage

			room := ActiveRoomsByNameMap.Get(inFrame.Room.Name)

			if room == nil {
				util.LogInfo("failed to support/reject ('%v') message for user '%s' - room '%s' not found", isSupport, clSocket.SessionUUID, inFrame.Room.Name)
				writeErrorMessageToSocket(clSocket, domain_structures.WsRoomNotFound, inFrame.RequestId)

				continue
			}

			room.Lock()

			room.LastActiveAt = time.Now().UnixNano()

			if room.IsDeleted {
				room.Unlock()

				util.LogInfo("failed to support/reject ('%v') message for user '%s' - room '%s' was deleted", isSupport, clSocket.SessionUUID, inFrame.Room.Name)
				writeErrorMessageToSocket(clSocket, domain_structures.WsRoomNotFound, inFrame.RequestId)

				continue
			}

			userInRoomUUID, userFound := room.ActiveRoomUserUUIDBySessionUUID[clSocket.SessionUUID]

			if !userFound {
				room.Unlock()

				util.LogInfo("failed to support/reject ('%v') message - user '%s' not active for room '%s'", isSupport, clSocket.SessionUUID, inFrame.Room.Name)
				writeErrorMessageToSocket(clSocket, domain_structures.WsRoomNotAuthorized, inFrame.RequestId)

				continue
			}

			clientSocketForThisRoom, socketFound := room.ActiveClientSocketsByUUID[clSocket.SocketUUID]

			if !socketFound || clientSocketForThisRoom.IsDead() {
				room.Unlock()

				util.LogInfo("failed to support/reject ('%v') message - socket '%s' not active for room '%s'", isSupport, clSocket.SocketUUID, inFrame.Room.Name)
				writeErrorMessageToSocket(clSocket, domain_structures.WsConnectionError, inFrame.RequestId)

				continue
			}

			message := inFrame.Message

			existingMessage, messageFound := room.RoomMessages[message.Id]

			if !messageFound {
				room.Unlock()

				writeRequestProcessedToSocket(clSocket, inFrame.RequestId)

				continue
			}
			//not allowed for same user that created message, except for room creator
			if existingMessage.UserInRoomUUID == userInRoomUUID && room.CreatedBySessionUUID != clSocket.SessionUUID {
				room.Unlock()

				util.LogWarn("failed to support/reject ('%v') message - user '%s' is an author of message '%d'", isSupport, clSocket.SocketUUID, existingMessage.Id)
				writeRequestProcessedToSocket(clSocket, inFrame.RequestId)

				continue
			}

			util.LogTrace("user '%s' is supporting/rejecting ('%v') message '%d' in room '%s' / '%s'",
				isSupport, clSocket.SessionUUID, existingMessage.Id, room.Id, room.Name)

			messageVotes, votesInitialized := room.MessageVotesByMessageId[existingMessage.Id]

			if !votesInitialized {
				messageVotes = &domain_structures.RoomMessageVotes{
					SupportVotesBySessionUUID: make(map[string]bool),
					RejectVotesBySessionUUID:  make(map[string]bool),
				}
				room.MessageVotesByMessageId[existingMessage.Id] = messageVotes
			}

			userSupports, supportRecordFound := messageVotes.SupportVotesBySessionUUID[clSocket.SessionUUID]
			userRejects, rejectRecordFound := messageVotes.RejectVotesBySessionUUID[clSocket.SessionUUID]

			//if user chosen support - check if user already supports this message. If no - add one to support, else - cancel support
			if isSupport {
				if !supportRecordFound || !userSupports {
					messageVotes.SupportVotesBySessionUUID[clSocket.SessionUUID] = true
					existingMessage.SupportedCount = existingMessage.SupportedCount + 1
				} else {
					//if user already supports this message - cancel support
					messageVotes.SupportVotesBySessionUUID[clSocket.SessionUUID] = false
					existingMessage.SupportedCount = existingMessage.SupportedCount - 1
				}

				if rejectRecordFound && userRejects {
					messageVotes.RejectVotesBySessionUUID[clSocket.SessionUUID] = false
					existingMessage.RejectedCount = existingMessage.RejectedCount - 1
				}
			} else {
				//same logic for rejecting

				if !rejectRecordFound || !userRejects {
					messageVotes.RejectVotesBySessionUUID[clSocket.SessionUUID] = true
					existingMessage.RejectedCount = existingMessage.RejectedCount + 1
				} else {
					//if user already rejects this message - cancel reject
					messageVotes.RejectVotesBySessionUUID[clSocket.SessionUUID] = false
					existingMessage.RejectedCount = existingMessage.RejectedCount - 1
				}

				if supportRecordFound && userSupports {
					messageVotes.SupportVotesBySessionUUID[clSocket.SessionUUID] = false
					existingMessage.SupportedCount = existingMessage.SupportedCount - 1
				}
			}

			lastVotedAt := time.Now().UnixNano()
			existingMessage.LastVotedAt = &lastVotedAt

			messageSupportDispatchingFrame := &domain_structures.OutMessageFrame{
				Command:       domain_structures.TextMessageSupportOrReject,
				Message:       &[]domain_structures.RoomMessageDTO{copyVotedMessageAsDTO(existingMessage)},
				CreatedAtNano: &lastVotedAt,
			}

			//make copy of active client sockets connected to this room while under lock.
			//After unlock - initial list may be updated at any point by parallel routines
			roomActiveClientSocketsByUUID := room.CopyActiveClientSocketMapNonLocking()

			room.Unlock()

			//send new message to all active users, respond OK to user immediately
			writeFrameToActiveRoomMembers(messageSupportDispatchingFrame, room, roomActiveClientSocketsByUUID)

			writeRequestProcessedToSocket(clSocket, inFrame.RequestId)

		case domain_structures.UserDrawingMessage:
			room := ActiveRoomsByNameMap.Get(inFrame.Room.Name)

			if room == nil {
				util.LogInfo("failed to send drawing message for user '%s' - room '%s' not found", clSocket.SessionUUID, inFrame.Room.Name)
				writeErrorMessageToSocket(clSocket, domain_structures.WsRoomNotFound, inFrame.RequestId)

				continue
			}

			room.Lock()

			room.LastActiveAt = time.Now().UnixNano()

			if room.IsDeleted {
				room.Unlock()

				util.LogInfo("failed to send drawing message for user '%s' - room '%s' was deleted", clSocket.SessionUUID, inFrame.Room.Name)
				writeErrorMessageToSocket(clSocket, domain_structures.WsRoomNotFound, inFrame.RequestId)

				continue
			}

			userInRoomUUID, userFound := room.ActiveRoomUserUUIDBySessionUUID[clSocket.SessionUUID]

			if !userFound {
				room.Unlock()

				util.LogInfo("failed to send drawing message - user '%s' not active for room '%s'", clSocket.SessionUUID, inFrame.Room.Name)
				writeErrorMessageToSocket(clSocket, domain_structures.WsRoomNotAuthorized, inFrame.RequestId)

				continue
			}

			clientSocketForThisRoom, socketFound := room.ActiveClientSocketsByUUID[clSocket.SocketUUID]

			if !socketFound || clientSocketForThisRoom.IsDead() {
				room.Unlock()

				util.LogInfo("failed to send drawing message - socket '%s' not active for room '%s'", clSocket.SocketUUID, inFrame.Room.Name)
				writeErrorMessageToSocket(clSocket, domain_structures.WsConnectionError, inFrame.RequestId)

				continue
			}

			message := inFrame.Message

			util.LogTrace("user '%s' is sending drawing message to room '%s' / '%s'", clSocket.SessionUUID, room.Id, room.Name)

			//transform message and add to room messages array
			newRoomMessage := &domain_structures.RoomMessage{
				Id:               room.NextMessageId,
				Text:             message.Text,
				SupportedCount:   0,
				RejectedCount:    0,
				UserInRoomUUID:   userInRoomUUID,
				CreatedAtSec:     time.Now().Unix(),
				ReplyToUserId:    message.ReplyToUserId,
				ReplyToMessageId: message.ReplyToMessageId,
			}

			room.NextMessageId += 1

			room.RoomMessages[newRoomMessage.Id] = newRoomMessage
			room.RoomMessagesLen = len(room.RoomMessages)

			//check room messages amount. If approaching limit - notify users (later), if reached limit - cut messages array in half
			roomMessagesCurrentAmount := room.RoomMessagesLen

			if roomMessagesCurrentAmount == RoomMessagesLimit {
				shrinkRoomMessagesMap(room)
			}

			messageDispatchingFrame := &domain_structures.OutMessageFrame{
				Command: domain_structures.UserDrawingMessage,
				Message: &[]domain_structures.RoomMessageDTO{copyMessageAsDTO(newRoomMessage)},
			}

			//make copy of active client sockets connected to this room while under lock.
			//After unlock - initial list may be updated at any point by parallel routines
			roomActiveClientSocketsByUUID := room.CopyActiveClientSocketMapNonLocking()

			room.Unlock()

			//send new message to all active users, respond OK to user immediately
			writeFrameToActiveRoomMembers(messageDispatchingFrame, room, roomActiveClientSocketsByUUID)

			if roomMessagesCurrentAmount == RoomMessagesLimit {
				writeNotificationToActiveRoomMembers(domain_structures.NotifyMessagesLimitReached, room, roomActiveClientSocketsByUUID)
			} else if util.ArrayContainsInt(RoomMessagesLimitApproachingWarningBreakpoints, roomMessagesCurrentAmount) {
				writeNotificationToActiveRoomMembers(domain_structures.NotifyMessagesLimitApproaching, room, roomActiveClientSocketsByUUID)
			}

			writeRequestProcessedToSocket(clSocket, inFrame.RequestId)
		}
	}
}

func SendControlCommandServerStatusChanged(newServerStatus string) {
	var activeRoomsCopy []*domain_structures.Room

	ActiveRoomsByNameMap.Lock()

	//change server status string under rooms list lock
	ServerStatus = newServerStatus
	//collect list of existing rooms to notify about server status change
	for _, room := range ActiveRoomsByNameMap.ActiveRoomsByName {
		if !room.IsDeleted {
			activeRoomsCopy = append(activeRoomsCopy, room)
		}
	}

	ActiveRoomsByNameMap.Unlock()

	for _, room := range activeRoomsCopy {
		if !room.IsDeleted {
			writeRoomDescriptionChangedFrameToActiveRoomMembers(room, ServerStatus)
		}
	}
}

func createRoom(frame *domain_structures.InMessageFrame, createdBySessionUUID string) (*domain_structures.Room, error) {
	nameTrimmed := strings.TrimSpace(frame.Room.Name)
	passwordTrimmed := strings.TrimSpace(frame.Room.Password)

	nameDecoded, _ := url.QueryUnescape(nameTrimmed)
	passwordDecoded, _ := url.QueryUnescape(passwordTrimmed)

	if len([]rune(nameDecoded)) < RoomCredsMinChars ||
		len([]rune(nameDecoded)) > RoomCredsMaxChars || len([]rune(passwordDecoded)) > RoomCredsMaxChars {

		return nil, RoomCredsValidationErrorInvalidLength
	}

	if util.ArrayContainsString(config.AppConfig.ForbiddenRoomNames, nameTrimmed) {
		return nil, RoomCredsValidationErrorNameForbidden
	}

	for _, r := range nameTrimmed {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && !util.ArrayContainsString(allowedRoomNameSpecialChars, string(r)) {

			return nil, RoomCredsValidationErrorNameHasBadChars
		}
	}

	newRoomUUID, err := uuid.NewUUID()

	if err != nil {
		return nil, err
	}

	passwordHash := ""

	if passwordTrimmed != "" {
		passwordHash, err = hasher.GenerateHashFromString(passwordTrimmed)

		if err != nil {
			return nil, err
		}
	}

	roomCreatedAt := time.Now().UnixNano()

	return &domain_structures.Room{
		IsDeleted:                           false,
		Id:                                  newRoomUUID.String(),
		Name:                                nameTrimmed,
		PasswordHash:                        passwordHash,
		Description:                         "",
		CreatedBySessionUUID:                createdBySessionUUID,
		StartedAt:                           roomCreatedAt,
		LastActiveAt:                        roomCreatedAt,
		NextMessageId:                       1,
		AllRoomAuthorizedUsersBySessionUUID: make(map[string]*domain_structures.RoomUser),
		ActiveRoomUserUUIDBySessionUUID:     make(map[string]string),
		ActiveRoomUsersLen:                  0,
		ActiveClientSocketsByUUID:           make(map[string]*domain_structures.WebSocket),
		RoomMessages:                        make(map[int64]*domain_structures.RoomMessage),
		RoomMessagesLen:                     0,
		MessageVotesByMessageId:             make(map[int64]*domain_structures.RoomMessageVotes),
	}, nil
}

func logIntoRoom(
	room *domain_structures.Room,
	clSocket *domain_structures.WebSocket,
	frame *domain_structures.InMessageFrame,
	createdRoom bool,
) {
	room.Lock()

	room.LastActiveAt = time.Now().UnixNano()

	if room.IsDeleted {
		room.Unlock()

		util.LogInfo("failed to login user '%s' - room '%s' was deleted", clSocket.SessionUUID, room.Name)
		writeErrorMessageToSocket(clSocket, domain_structures.WsRoomNotFound, frame.RequestId)

		return
	}

	if room.ActiveRoomUsersLen >= RoomMaxUsersLimit {
		room.Unlock()

		util.LogTrace("room '%s' is full (%s)", room.Name, room.Id)
		writeErrorMessageToSocket(clSocket, domain_structures.WsRoomIsFullError, frame.RequestId)

		return
	}

	//check if user logged into this room at some point, save new active user into room (or possibly restore from existing authorizations list)

	var roomUser *domain_structures.RoomUser

	trimmedRoomUserName := strings.TrimSpace(frame.UserName)

	existingAuthorization, alreadyAuthorized := room.AllRoomAuthorizedUsersBySessionUUID[clSocket.SessionUUID]

	roomHasPassword := room.PasswordHash != ""

	//check password only if room has one and user either haven't authorized yet or already authorized but passed some password again
	if roomHasPassword && (!alreadyAuthorized || frame.Room.Password != "") {
		err := hasher.CheckHashEquality(room.PasswordHash, frame.Room.Password)

		if err != nil {
			room.Unlock()

			util.LogTrace("incorrect password while joining room '%s': '%s'", room.Id, err)
			writeErrorMessageToSocket(clSocket, domain_structures.WsRoomInvalidPassword, frame.RequestId)

			return
		}
	}

	if alreadyAuthorized {
		roomUser = existingAuthorization

		//if user provided user name an it is different from previous one
		if trimmedRoomUserName != "" && existingAuthorization.UserName != trimmedRoomUserName {
			//validate provided user name or pick anonymous one
			roomUserName, isAnon, err := validateOrPickRoomUserName(trimmedRoomUserName, room)

			if err != nil {
				room.Unlock()

				if err == ProvidedNameTaken {
					util.LogTrace("room UserName '%s' already taken. Room: '%s'", trimmedRoomUserName, room.Id)
					writeErrorMessageToSocket(clSocket, domain_structures.WsRoomUserNameTaken, frame.RequestId)

					return
				} else if err == BadNameLength {
					util.LogTrace("room UserName '%s' has wrong length. Room: '%s'", trimmedRoomUserName, room.Id)
					writeErrorMessageToSocket(clSocket, domain_structures.WsRoomUserNameValidationError, frame.RequestId)

					return
				}
			}

			roomUser.UserName = roomUserName
			roomUser.IsAnonName = isAnon
		}

	} else {
		roomUser = &domain_structures.RoomUser{}

		//generate new room user id
		newUserInRoomUUID, err := uuid.NewUUID()

		if err != nil {
			room.Unlock()

			util.LogSevere("failed to generate UUID for room user: '%s'", err)
			writeErrorMessageToSocket(clSocket, domain_structures.WsServerError, frame.RequestId)

			return
		}

		roomUser.UserInRoomUUID = newUserInRoomUUID.String()

		//validate provided user name or pick anonymous one
		roomUserName, isAnon, err := validateOrPickRoomUserName(trimmedRoomUserName, room)

		if err != nil {
			room.Unlock()

			if err == ProvidedNameTaken {
				util.LogTrace("room UserName '%s' already taken. Room: '%s'", trimmedRoomUserName, room.Id)
				writeErrorMessageToSocket(clSocket, domain_structures.WsRoomUserNameTaken, frame.RequestId)

				return
			} else if err == BadNameLength {
				util.LogTrace("room UserName '%s' has wrong length. Room: '%s'", trimmedRoomUserName, room.Id)
				writeErrorMessageToSocket(clSocket, domain_structures.WsRoomUserNameValidationError, frame.RequestId)

				return
			}
		}

		roomUser.UserName = roomUserName
		roomUser.IsAnonName = isAnon
	}

	var requestProcessingDetails string
	if createdRoom {
		requestProcessingDetails = "room_created"
	} else {
		requestProcessingDetails = "room_joined"
	}

	requestProcessingDetails += ";"

	if roomHasPassword {
		requestProcessingDetails += "password=true"
	} else {
		requestProcessingDetails += "password=false"
	}

	/* Put (or re-put) user into room */

	room.AllRoomAuthorizedUsersBySessionUUID[clSocket.SessionUUID] = roomUser

	//if this request if from home page - just authorize user, will fully join room later
	if frame.Command == domain_structures.RoomCreateJoinAuthorize {
		room.Unlock()

		writeRequestProcessedToSocketWithAdditInfo(clSocket, &room.StartedAt, frame.RequestId, &requestProcessingDetails, &room.Id, &roomUser.UserInRoomUUID, nil)

		return
	}

	room.ActiveRoomUserUUIDBySessionUUID[clSocket.SessionUUID] = roomUser.UserInRoomUUID

	room.ActiveRoomUsersLen = len(room.ActiveRoomUserUUIDBySessionUUID)

	//in case alive socket already exists for user - remove it from room, and later close (after unlocking room)
	existingSocket := findSocketBySessionUUID(&room.ActiveClientSocketsByUUID, clSocket.SessionUUID)

	if existingSocket != nil {
		delete(room.ActiveClientSocketsByUUID, existingSocket.SocketUUID)

		//synchronously write error message to socket, then close it. Small timeout since this call is still under room lock
		writeTimeout := time.Second * 2
		doWriteErrorMessageToSocket(existingSocket, domain_structures.WsRoomUserDuplication, frame.RequestId, true, &writeTimeout)

		existingSocket.Terminate()
	}

	//put new socket into room
	room.ActiveClientSocketsByUUID[clSocket.SocketUUID] = clSocket
	clSocket.RelatedRoom = room

	util.LogInfo("user '%s'/'%s' joined room '%s' / '%s'", clSocket.SessionUUID, roomUser.UserInRoomUUID, room.Id, room.Name)

	//copy all room messages
	allRoomMessagesDTOCopy := copyAllRoomMessagesAsDTOArray(&room.RoomMessages)

	//copy room active users list
	var roomActiveUsersCopy []domain_structures.RoomUserDTO

	for _, user := range room.AllRoomAuthorizedUsersBySessionUUID {
		roomActiveUsersCopy = append(roomActiveUsersCopy, domain_structures.RoomUserDTO{
			UserInRoomUUID: &(*user).UserInRoomUUID,
			UserName:       &(*user).UserName,
			IsAnonName:     &(*user).IsAnonName,
			IsOnlineInRoom: isUserOnlineInRoom(room, (*user).UserInRoomUUID),
		})
	}

	//find room creator user
	var roomCreatorUserInRoomUUID *string = nil
	roomCreatorUser, found := room.AllRoomAuthorizedUsersBySessionUUID[room.CreatedBySessionUUID]

	if found {
		roomCreatorUserInRoomUUID = &roomCreatorUser.UserInRoomUUID
	}

	roomDataCopiedAt := time.Now().UnixNano()

	roomDescriptionSafeCopy := room.Description

	room.Unlock()

	go writeMembersListChangedFrameToActiveRoomMembers(room, &clSocket.SocketUUID)

	/* Send current room members list and all messages existing to this moment to new user */

	//send room users list (separately for newly-joined user, to pass it pseudo-synchronously)
	roomMembersListChangedFrame := domain_structures.OutMessageFrame{
		Command:       domain_structures.RoomMembersChanged,
		CreatedAtNano: &roomDataCopiedAt,
		AllRoomUsers:  &roomActiveUsersCopy,
	}

	//send all room messages to user
	sort.Slice(*allRoomMessagesDTOCopy, func(i, j int) bool {
		return *((*allRoomMessagesDTOCopy)[i].Id) < *((*allRoomMessagesDTOCopy)[j].Id)
	})

	allMessagesFrame := domain_structures.OutMessageFrame{
		Command:       domain_structures.AllTextMessages,
		Message:       allRoomMessagesDTOCopy,
		CreatedAtNano: &roomDataCopiedAt,
	}

	roomDescriptionFrame := domain_structures.OutMessageFrame{
		Command:                   domain_structures.RoomChangeDescription,
		CreatedAtNano:             &roomDataCopiedAt,
		RoomCreatorUserInRoomUUID: roomCreatorUserInRoomUUID,
		ServerStatus:              &ServerStatus,
		Message: &[]domain_structures.RoomMessageDTO{
			{Text: &roomDescriptionSafeCopy},
		},
	}

	if err := writeAfterRoomJoinMessagesToSocket(&roomMembersListChangedFrame, &allMessagesFrame, &roomDescriptionFrame, clSocket); err == nil {
		writeRequestProcessedToSocketWithAdditInfo(clSocket, &room.StartedAt, frame.RequestId, &requestProcessingDetails, &room.Id, &roomUser.UserInRoomUUID, &config.BuildVersion)
	} else {
		writeErrorMessageToSocket(clSocket, domain_structures.WsServerError, frame.RequestId)
	}
}
