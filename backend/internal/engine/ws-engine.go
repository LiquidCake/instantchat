package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
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

const RoomMessagesLimit = 500

var RoomMessagesLimitApproachingWarningBreakpoints = []int{460, 480, 495}

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
				newRoom, err := createRoom(inFrame.Room.Name, inFrame.Room.Password, clSocket.SessionUUID)

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

			newRoom, err := createRoom(inFrame.Room.Name, inFrame.Room.Password, clSocket.SessionUUID)

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
			newRoomMessage := addNewMessageToRoom(
				room,
				userInRoomUUID,
				message.Text, //NOTE: we are expecting message text to come url-escaped
				message.ReplyToUserId,
				message.ReplyToMessageId,
			)

			//check room messages amount. if reached limit - cut messages list in half
			lowestMessageIdAfterShrink := checkLimitAndShrinkMessagesList(room)

			messageDispatchingFrame := &domain_structures.OutMessageFrame{
				Command: domain_structures.TextMessage,
				Message: &[]domain_structures.RoomMessageDTO{copyMessageAsDTO(newRoomMessage)},
			}

			//make copy of active client sockets connected to this room while under lock.
			//After unlock - initial list may be updated at any point by parallel routines
			roomActiveClientSocketsByUUID := room.CopyActiveClientSocketMapNonLocking()

			room.Unlock()

			//schedule sending new message to all active users, respond OK to user immediately
			scheduleSendingNewMessageToActiveUsers(
				room,
				messageDispatchingFrame,
				roomActiveClientSocketsByUUID,
				lowestMessageIdAfterShrink,
			)

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

			existingMessage.Text = message.Text //NOTE: we are expecting message text to come url-escaped
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
			newRoomMessage := addNewMessageToRoom(
				room,
				userInRoomUUID,
				message.Text,
				message.ReplyToUserId,
				message.ReplyToMessageId,
			)

			//check room messages amount. if reached limit - cut messages list in half
			lowestMessageIdAfterShrink := checkLimitAndShrinkMessagesList(room)

			messageDispatchingFrame := &domain_structures.OutMessageFrame{
				Command: domain_structures.UserDrawingMessage,
				Message: &[]domain_structures.RoomMessageDTO{copyMessageAsDTO(newRoomMessage)},
			}

			//make copy of active client sockets connected to this room while under lock.
			//After unlock - initial list may be updated at any point by parallel routines
			roomActiveClientSocketsByUUID := room.CopyActiveClientSocketMapNonLocking()

			room.Unlock()

			//schedule sending new message to all active users, respond OK to user immediately
			scheduleSendingNewMessageToActiveUsers(
				room,
				messageDispatchingFrame,
				roomActiveClientSocketsByUUID,
				lowestMessageIdAfterShrink,
			)

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

func createRoom(roomName string, roomPassword string, createdBySessionUUID string) (*domain_structures.Room, error) {
	nameTrimmed := strings.TrimSpace(roomName)
	passwordTrimmed := strings.TrimSpace(roomPassword)

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

	room := &domain_structures.Room{
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
	}

	addTechnicalUsersToRoom(room)

	return room, nil
}

func addTechnicalUsersToRoom(room *domain_structures.Room) {
	//technical user for directly sent messages (via http)
	externalUser := &domain_structures.RoomUser{}

	externalUser.UserInRoomUUID = ExternalUserUUID
	externalUser.UserName = ExternalUserName
	externalUser.IsAnonName = false

	room.AllRoomAuthorizedUsersBySessionUUID[ExternalUserSessionUUID] = externalUser
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
	allRoomUsersCopy := copyAllRoomUsersList(room)

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
		Command:            domain_structures.RoomMembersChanged,
		CreatedAtNano:      &roomDataCopiedAt,
		AllRoomUsers:       allRoomUsersCopy,
		CurrentBuildNumber: &config.BuildVersion,
	}

	//send all room messages to user
	sort.Slice(*allRoomMessagesDTOCopy, func(i, j int) bool {
		return *((*allRoomMessagesDTOCopy)[i].Id) < *((*allRoomMessagesDTOCopy)[j].Id)
	})

	allMessagesFrame := domain_structures.OutMessageFrame{
		Command:            domain_structures.AllTextMessages,
		Message:            allRoomMessagesDTOCopy,
		CreatedAtNano:      &roomDataCopiedAt,
		CurrentBuildNumber: &config.BuildVersion,
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

// call only under room lock
func addNewMessageToRoom(
	room *domain_structures.Room,
	userInRoomUUID string,
	messageText string,
	replyToUserId *string,
	replyToMessageId *int64,

) *domain_structures.RoomMessage {
	newRoomMessage := &domain_structures.RoomMessage{
		Id:               room.NextMessageId,
		Text:             messageText,
		SupportedCount:   0,
		RejectedCount:    0,
		UserInRoomUUID:   userInRoomUUID,
		CreatedAtSec:     time.Now().Unix(),
		ReplyToUserId:    replyToUserId,
		ReplyToMessageId: replyToMessageId,
	}

	room.NextMessageId += 1

	room.RoomMessages[newRoomMessage.Id] = newRoomMessage
	room.RoomMessagesLen = len(room.RoomMessages)

	return newRoomMessage
}

func scheduleSendingNewMessageToActiveUsers(
	room *domain_structures.Room,
	messageDispatchingFrame *domain_structures.OutMessageFrame,
	roomActiveClientSocketsByUUID *map[string]*domain_structures.WebSocket,
	lowestMessageIdAfterShrink int64,
) {
	//schedule actual message sending
	writeFrameToActiveRoomMembers(messageDispatchingFrame, room, roomActiveClientSocketsByUUID)

	//notify users if messages shrink happened or is approaching
	if lowestMessageIdAfterShrink != int64(-1) {
		lowestMessageIdAfterShrinkStr := strconv.FormatInt(lowestMessageIdAfterShrink, 10)

		writeNotificationToActiveRoomMembers(domain_structures.NotifyMessagesLimitReached, room, roomActiveClientSocketsByUUID, &lowestMessageIdAfterShrinkStr)

	} else if util.ArrayContainsInt(RoomMessagesLimitApproachingWarningBreakpoints, room.RoomMessagesLen) {
		writeNotificationToActiveRoomMembers(domain_structures.NotifyMessagesLimitApproaching, room, roomActiveClientSocketsByUUID, nil)
	}
}

// must be executed under room lock
func checkLimitAndShrinkMessagesList(room *domain_structures.Room) int64 {
	if room.RoomMessagesLen >= RoomMessagesLimit {
		return shrinkRoomMessagesMap(room)
	} else {
		return -1
	}
}

// side method for room messages retrieval - used for direct http requests
func RetrieveRoomMessagesDirectly(roomName string, roomPassword string, messagesLimit int,
	targetMessageId int64, responseFormat string, quiteMode bool) []byte {
	room := ActiveRoomsByNameMap.Get(roomName)
	newRoomCreated := false

	if room == nil {
		err := createRoomByDirectMessageFlow(roomName, roomPassword, ExternalUserSessionUUID)

		if err != "" {
			util.LogTrace("failed to create room '%s' for dirrect message flow. Erorr: %s", roomName, err)

			return util.BuildDirectRoomMessagesErrorResponse(
				fmt.Sprintf("error: failed to create room - %s", err), responseFormat)
		}

		room = ActiveRoomsByNameMap.Get(roomName)
		newRoomCreated = true
	}

	room.Lock()

	//corner case, should not happen realistically
	if room.IsDeleted {
		room.Unlock()

		util.LogWarn("failed to send direct message - room '%s' is already deleted", roomName)

		return util.BuildDirectRoomMessagesErrorResponse("error: internal error", responseFormat)
	}

	roomHasPassword := room.PasswordHash != ""

	if roomHasPassword {
		err := hasher.CheckHashEquality(room.PasswordHash, roomPassword)

		if err != nil {
			room.Unlock()

			return util.BuildDirectRoomMessagesErrorResponse(
				"error: wrong room password (use URL param 'p=myPassword')", responseFormat)
		}
	}

	allRoomMessagesDTOCopy := copyAllRoomMessagesAsDTOArray(&room.RoomMessages)

	allRoomUsersCopy := copyAllRoomUsersList(room)
	userNameByUserInRoomUUID := make(map[string]string)

	for _, user := range *allRoomUsersCopy {
		userNameByUserInRoomUUID[*user.UserInRoomUUID] = *user.UserName
	}

	room.Unlock()

	sort.Slice(*allRoomMessagesDTOCopy, func(i, j int) bool {
		return *((*allRoomMessagesDTOCopy)[i].Id) < *((*allRoomMessagesDTOCopy)[j].Id)
	})

	messagesToReturn := allRoomMessagesDTOCopy
	messagesToReturnLen := room.RoomMessagesLen

	//if user requested messages starting from particular id - return only those
	if messagesToReturnLen > 0 && targetMessageId > 0 {
		//to begin with - consider last message in array to be closest to target
		//normally later we will find better candidate
		closestMessageIdx := messagesToReturnLen - 1

		for i, message := range *messagesToReturn {
			if *message.Id >= targetMessageId {
				closestMessageIdx = i
				break
			}
		}

		messagesTail := (*messagesToReturn)[closestMessageIdx:]
		messagesToReturn = &messagesTail
		messagesToReturnLen = len(messagesTail)
	}

	//if user requested limited response - return only tail of messages array
	if messagesToReturnLen > 0 && messagesLimit > 0 && messagesLimit < messagesToReturnLen {
		tailBeginingIndex := messagesToReturnLen - messagesLimit

		messagesTail := (*messagesToReturn)[tailBeginingIndex:]
		messagesToReturn = &messagesTail
		messagesToReturnLen = len(messagesTail)
	}

	if responseFormat == "json" {
		responseJsonStr := map[string]interface{}{
			"createdNewRoom":         newRoomCreated,
			"messagesCount":          messagesToReturnLen,
			"totalRoomMessagesCount": room.RoomMessagesLen,
		}

		messagesArray := make([]map[string]interface{}, messagesToReturnLen)

		for i, message := range *messagesToReturn {
			messageId := *message.Id
			unescapedMessageText, err := url.QueryUnescape(*message.Text)

			if err != nil {
				unescapedMessageText = fmt.Sprintf("system: failed to unescape message: %s", err)
			}

			userName, found := userNameByUserInRoomUUID[*message.UserInRoomUUID]

			if !found {
				userName = "unknown"
			}

			userName, err = url.QueryUnescape(userName)
			if err != nil {
				userName = "unknown"
			}

			messagesArray[i] = map[string]interface{}{
				"id":       messageId,
				"text":     unescapedMessageText,
				"userName": userName,
			}
		}

		responseJsonStr["messages"] = messagesArray

		jsonData, err := json.Marshal(responseJsonStr)

		if err != nil {
			util.LogSevere("Failed to serialize 'direct room messages response'. err: '%s'", err)

			return []byte("{\"error\": \"failed to serialize json\"}")
		}

		return jsonData

	} else {
		var sb strings.Builder

		if !quiteMode {
			sb.WriteString("URL params\n")
			sb.WriteString("- send room password: 'p=myPassword'\n")
			sb.WriteString("- start from message id: 'id=8' to get only messages starting from id 8 (or closest)\n")
			sb.WriteString("- limit messages: 'l=5' to get only 5 latest messages\n")
			sb.WriteString("- format response as json: 'format=json'\n")
			sb.WriteString("- dont send this help text: 'quite=true'\n")
			sb.WriteString("\n")
		}

		sb.WriteString(fmt.Sprintf("Showing %d of %d universally accessible chat room messages below this line\n\n",
			messagesToReturnLen, room.RoomMessagesLen))

		if newRoomCreated {
			sb.WriteString("system: you have just created this room\n")
		}

		for _, message := range *messagesToReturn {
			messageId := *message.Id
			unescapedMessageText, err := url.QueryUnescape(*message.Text)

			if err != nil {
				unescapedMessageText = fmt.Sprintf("system: failed to unescape message: %s", err)
			}

			userName, found := userNameByUserInRoomUUID[*message.UserInRoomUUID]

			if !found {
				userName = "unknown"
			}

			userName, err = url.QueryUnescape(userName)
			if err != nil {
				userName = "unknown"
			}

			sb.WriteString(fmt.Sprintf("#%d %s: %s\n", messageId, userName, unescapedMessageText))
		}

		sb.WriteString("\n")

		return []byte(sb.String())
	}
}

// side method for sending room message - used for direct http requests
func SendRoomMessageDirectly(roomName string, roomPassword string, message string, responseFormat string) []byte {
	room := ActiveRoomsByNameMap.Get(roomName)
	newRoomCreated := false

	if room == nil {
		err := createRoomByDirectMessageFlow(roomName, roomPassword, ExternalUserSessionUUID)

		if err != "" {
			util.LogTrace("failed to create room '%s' for dirrect message flow. Erorr: %s", roomName, err)

			return util.BuildDirectRoomMessagesErrorResponse(
				fmt.Sprintf("error: failed to create room - %s", err), responseFormat)
		}

		room = ActiveRoomsByNameMap.Get(roomName)
		newRoomCreated = true
	}

	roomHasPassword := room.PasswordHash != ""

	if roomHasPassword {
		err := hasher.CheckHashEquality(room.PasswordHash, roomPassword)

		if err != nil {
			util.LogInfo("failed to send direct message - wrong password for room '%s'", room.Name)

			return util.BuildDirectRoomMessagesErrorResponse(
				"error: wrong room password (use HTTP param &p=myPassword)", responseFormat)
		}
	}

	room.Lock()

	room.LastActiveAt = time.Now().UnixNano()

	if room.IsDeleted {
		room.Unlock()

		util.LogInfo("failed to send direct message - room '%s' is deleted", room.Name)

		return util.BuildDirectRoomMessagesErrorResponse("error: room is deleted", responseFormat)
	}

	util.LogTrace("sending direct message of len '%d' to room '%s' / '%s'", len(message), room.Id, room.Name)

	//transform message and add to room messages array
	newRoomMessage := addNewMessageToRoom(
		room,
		ExternalUserUUID,
		message,
		nil,
		nil,
	)

	//check room messages amount. if reached limit - cut messages list in half
	lowestMessageIdAfterShrink := checkLimitAndShrinkMessagesList(room)

	messageDispatchingFrame := &domain_structures.OutMessageFrame{
		Command: domain_structures.TextMessage,
		Message: &[]domain_structures.RoomMessageDTO{copyMessageAsDTO(newRoomMessage)},
	}

	//make copy of active client sockets connected to this room while under lock.
	//After unlock - initial list may be updated at any point by parallel routines
	roomActiveClientSocketsByUUID := room.CopyActiveClientSocketMapNonLocking()

	room.Unlock()

	//schedule sending new message to all active users, respond OK to user immediately
	scheduleSendingNewMessageToActiveUsers(
		room,
		messageDispatchingFrame,
		roomActiveClientSocketsByUUID,
		lowestMessageIdAfterShrink,
	)

	if responseFormat == "json" {
		responseJsonStr := map[string]interface{}{
			"responseText":   "message sent",
			"createdNewRoom": newRoomCreated,
		}

		jsonData, _ := json.Marshal(responseJsonStr)

		return jsonData

	} else {
		var sb strings.Builder

		if newRoomCreated {
			sb.WriteString("Created new room!\n\n")
		}

		sb.WriteString("message sent\n")

		return []byte(sb.String())
	}
}

// side method direct message flow (http requests)
func createRoomByDirectMessageFlow(roomName string, roomPassword string, createdBySessionUUID string) string {
	ActiveRoomsByNameMap.Lock()

	if ActiveRoomsByNameMap.ContainsNonLocking(roomName) {
		ActiveRoomsByNameMap.Unlock()

		return ""
	}

	newRoom, err := createRoom(roomName, roomPassword, createdBySessionUUID)

	if err != nil {
		ActiveRoomsByNameMap.Unlock()

		if err == RoomCredsValidationErrorInvalidLength {
			util.LogTrace("invalid room credentials length: '%s'", roomName)
			return "error: invalid room credentials length"

		} else if err == RoomCredsValidationErrorNameForbidden {
			util.LogTrace("room name is forbidden: '%s'", roomName)
			return "error: room name is forbidden"

		} else if err == RoomCredsValidationErrorNameHasBadChars {
			util.LogTrace("room name contains bad characters: '%s'", roomName)
			return "error: room name contains bad characters"

		} else {
			util.LogSevere("error while creating room: '%s'", err)
			return "error: error while creating room"
		}
	}

	ActiveRoomsByNameMap.ActiveRoomsByName[newRoom.Name] = newRoom
	RoomsOnlineGauge.Inc()

	ActiveRoomsByNameMap.Unlock()

	util.LogInfo("room '%s' / '%s' created by direct message flow", newRoom.Id, newRoom.Name)

	StartSocketHouseKeeper(newRoom)

	return ""
}
