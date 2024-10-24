package domain_structures

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// socket
type WebSocket struct {
	sync.Mutex
	Socket              *websocket.Conn
	SocketUUID          string //id of this socket
	SessionUUID         string //user secret token
	LastKeepAliveSignal int64
	OutMessagesPutCh    chan<- *OutMessageWrapper //input side of channel pair - here engine puts new messages that must be sent to user
	OutMessagesGetCh    <-chan *OutMessageWrapper //output side of channel pair - here 'client socket message writing routine' takes messages and sends to user

	RelatedRoom *Room //room to which user joined (if any)
	isDead      bool
}

func (s *WebSocket) PutMessage(message *OutMessageWrapper) {
	s.Lock()
	defer s.Unlock()

	if !s.isDead {
		s.OutMessagesPutCh <- message
	}
}

func (s *WebSocket) Terminate() {
	s.Lock()

	if s.isDead {
		s.Unlock()

		return
	}

	s.isDead = true
	s.Unlock()

	s.Socket.Close()
	close(s.OutMessagesPutCh)
}

func (s *WebSocket) IsDead() bool {
	s.Lock()
	defer s.Unlock()

	return s.isDead
}

// message frame
type InMessageFrame struct {
	Command   Command     `json:"c"`
	RequestId *string     `json:"rq"`
	Message   RoomMessage `json:"m"`

	Room     RoomInfo `json:"r"`
	UserName string   `json:"uN"`

	//for supporing/rejecting message
	SupportOrRejectMessage bool `json:"srM"`
	//for keep-alive messages
	KeepAliveBeacon string `json:"kA"`
}

type OutMessageFrame struct {
	Command           Command           `json:"c"`
	RequestId         *string           `json:"rq,omitempty"`
	ProcessingDetails *string           `json:"pd,omitempty"`
	Message           *[]RoomMessageDTO `json:"m,omitempty"`

	//for returning current room context info
	RoomUUID                  *string `json:"rId,omitempty"`
	UserInRoomUUID            *string `json:"uId,omitempty"`
	RoomCreatorUserInRoomUUID *string `json:"rCuId,omitempty"`
	CreatedAtNano             *int64  `json:"cAt,omitempty"`

	//for returning all-time users list
	AllRoomUsers *[]RoomUserDTO `json:"rU,omitempty"`

	CurrentBuildNumber *string `json:"bN,omitempty"`
	ServerStatus       *string `json:"sS,omitempty"`
}

// struct to distribute message between all client socket routines
type OutMessageWrapper struct {
	OutMessageJson *[]byte
	Room           *Room
}

// message frame
type InitFrame struct {
	Platform *string `json:"p,omitempty"`
}

// commands are markers of action being performed - either incoming from user or returning to user
type Command string

const (
	RoomCreateJoin          Command = "R_C_J"
	RoomCreateJoinAuthorize Command = "R_C_J_AU"
	RoomCreate              Command = "R_C"
	RoomJoin                Command = "R_J"
	RoomChangeDescription   Command = "R_CH_D"
	RoomChangeUserName      Command = "R_CH_UN"
	RoomMembersChanged      Command = "R_M_CH"

	TextMessage                Command = "TM"
	TextMessageEdit            Command = "TM_E"
	TextMessageDelete          Command = "TM_D"
	TextMessageSupportOrReject Command = "TM_S_R"
	AllTextMessages            Command = "ALL_TM"

	UserDrawingMessage Command = "DM"

	Error            Command = "ER"
	RequestProcessed Command = "RP"

	NotifyMessagesLimitApproaching Command = "N_M_LIMIT_A"
	NotifyMessagesLimitReached     Command = "N_M_LIMIT_R"
)

/* rooms */

type RoomInfo struct {
	Name     string `json:"n"`
	Password string `json:"p"`
}

type RoomUser struct {
	UserInRoomUUID string `json:"uId"` //user id in scope of room (public)
	UserName       string `json:"n"`
	IsAnonName     bool   `json:"an"`
}

type RoomUserDTO struct {
	UserInRoomUUID *string `json:"uId"`
	UserName       *string `json:"n"`
	IsAnonName     *bool   `json:"an"`
	IsOnlineInRoom *bool   `json:"o"`
}

type RoomMessage struct {
	Id               int64   `json:"id"`
	Text             string  `json:"t"`
	SupportedCount   int     `json:"sC"`
	RejectedCount    int     `json:"rC"`
	LastEditedAt     *int64  `json:"lE"`
	LastVotedAt      *int64  `json:"lV"`
	ReplyToUserId    *string `json:"rU"`
	ReplyToMessageId *int64  `json:"rM"`
	UserInRoomUUID   string  `json:"uId"`
	CreatedAtSec     int64   `json:"cAt"` //! timestamp in seconds
}

// DTO is needed for sending RoomMessage to users, because some fields may be empty in case of different Commands,
// and we need pointer structure fields to omit them
type RoomMessageDTO struct {
	Id               *int64  `json:"id,omitempty"`
	Text             *string `json:"t,omitempty"`
	SupportedCount   *int    `json:"sC,omitempty"`
	RejectedCount    *int    `json:"rC,omitempty"`
	LastEditedAt     *int64  `json:"lE,omitempty"`
	LastVotedAt      *int64  `json:"lV,omitempty"`
	ReplyToUserId    *string `json:"rU,omitempty"`
	ReplyToMessageId *int64  `json:"rM,omitempty"`
	UserInRoomUUID   *string `json:"uId,omitempty"`
	CreatedAtSec     *int64  `json:"cAt,omitempty"`
}

type RoomCtrlInfo struct {
	Id                 string `json:"name"`
	Name               string `json:"id"`
	StartedAt          string `json:"startedAt"`
	ActiveRoomUsersNum int    `json:"activeRoomUsersNum"`
}

type RoomMessageVotes struct {
	SupportVotesBySessionUUID map[string]bool
	RejectVotesBySessionUUID  map[string]bool
}

// room
type Room struct {
	sync.Mutex
	IsDeleted            bool
	Id                   string
	Name                 string
	PasswordHash         string
	Description          string
	CreatedBySessionUUID string
	StartedAt            int64
	LastActiveAt         int64
	NextMessageId        int64 //gets incremented for every next message

	AllRoomAuthorizedUsersBySessionUUID map[string]*RoomUser //all users that were authorized for this room at any point - entries are NOT deleted
	ActiveRoomUserUUIDBySessionUUID     map[string]string    //user-in-room-UUID for users, authorized and active (joined room) at this moment (i.e. has alive socket) - entries are deleted once user's connection (socket) dies
	ActiveRoomUsersLen                  int
	ActiveClientSocketsByUUID           map[string]*WebSocket //sockets of each active user

	RoomMessages            map[int64]*RoomMessage //all room messages
	RoomMessagesLen         int
	MessageVotesByMessageId map[int64]*RoomMessageVotes //user votes (support/reject) for messages or this room
}

func (r *Room) CopyActiveClientSocketMap() (*map[string]*WebSocket, int64) {
	r.Lock()
	defer r.Unlock()

	activeClientSocketsByUUIDCopy := make(map[string]*WebSocket, len(r.ActiveClientSocketsByUUID))

	for k, v := range r.ActiveClientSocketsByUUID {
		activeClientSocketsByUUIDCopy[k] = v
	}

	return &activeClientSocketsByUUIDCopy, time.Now().UnixNano()
}

func (r *Room) CopyActiveClientSocketMapNonLocking() *map[string]*WebSocket {
	activeClientSocketsByUUIDCopy := make(map[string]*WebSocket, len(r.ActiveClientSocketsByUUID))

	for k, v := range r.ActiveClientSocketsByUUID {
		activeClientSocketsByUUIDCopy[k] = v
	}

	return &activeClientSocketsByUUIDCopy
}

//active rooms map

type ActiveRoomsByName struct {
	sync.Mutex
	ActiveRoomsByName map[string]*Room
}

func (a *ActiveRoomsByName) Get(name string) *Room {
	a.Lock()
	defer a.Unlock()

	return a.ActiveRoomsByName[name]
}

func (a *ActiveRoomsByName) Contains(name string) bool {
	a.Lock()
	defer a.Unlock()

	if _, found := a.ActiveRoomsByName[name]; found {
		return true
	} else {
		return false
	}
}

func (a *ActiveRoomsByName) ContainsNonLocking(name string) bool {
	if _, found := a.ActiveRoomsByName[name]; found {
		return true
	} else {
		return false
	}
}

func (a *ActiveRoomsByName) Delete(name string) {
	a.Lock()
	defer a.Unlock()

	delete(a.ActiveRoomsByName, name)
}

func (a *ActiveRoomsByName) CopyActiveRoomsByNameMap() (*map[string]*Room, int64) {
	a.Lock()
	defer a.Unlock()

	activeRoomsByNameCopy := make(map[string]*Room)

	for k, v := range a.ActiveRoomsByName {
		activeRoomsByNameCopy[k] = v
	}

	return &activeRoomsByNameCopy, time.Now().UnixNano()
}

/* Business errors */

type WsError struct {
	Code int
	Name string
	Text string
}

var WsServerError = WsError{Name: "WsServerError", Code: 101, Text: "server error"}
var WsConnectionError = WsError{Name: "WsConnectionError", Code: 102, Text: "connection error"}
var WsInvalidInput = WsError{Name: "WsInvalidInput", Code: 103, Text: "invalid input"}

var WsRoomExists = WsError{Name: "WsRoomExists", Code: 201, Text: "room with this name already exists"}
var WsRoomNotFound = WsError{Name: "WsRoomNotFound", Code: 202, Text: "room not found"}
var WsRoomInvalidPassword = WsError{Name: "WsRoomInvalidPassword", Code: 203, Text: "invalid room password"}
var WsRoomUserNameTaken = WsError{Name: "WsRoomUserNameTaken", Code: 204, Text: "provided user name is already taken"}
var WsRoomUserNameValidationError = WsError{Name: "WsRoomUserNameValidationError", Code: 205, Text: "invalid room user name length"}
var WsRoomNotAuthorized = WsError{Name: "WsRoomNotAuthorized", Code: 206, Text: "not authorized to join this room"}
var WsRoomMessageTooLargeError = WsError{Name: "WsRoomMessageTooLargeError", Code: 207, Text: "message is too long"}
var WsRoomIsFullError = WsError{Name: "WsRoomIsFullError", Code: 208, Text: "room is full"}
var WsRoomUserDuplication = WsError{Name: "WsRoomUserDuplication", Code: 209, Text: "user connected to this room from another browser tab"}

var WsRoomCredsValidationErrorBadLength = WsError{Name: "WsRoomCredsValidationErrorBadLength", Code: 301, Text: "invalid room name length"}
var WsRoomCredsValidationErrorNameForbidden = WsError{Name: "WsRoomCredsValidationErrorNameForbidden", Code: 302, Text: "room name is forbidden"}
var WsRoomCredsValidationErrorNameHasBadChars = WsError{Name: "WsRoomCredsValidationErrorNameHasBadChars", Code: 303, Text: "room name contains bad characters"}
var WsRoomValidationErrorBadDescriptionLength = WsError{Name: "WsRoomValidationErrorBadDescriptionLength", Code: 304, Text: "invalid room description length"}
