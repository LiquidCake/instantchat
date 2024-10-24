package engine

import (
	"errors"
	"net/url"
	"sort"

	"instantchat.rooms/instantchat/backend/internal/domain_structures"
)

const MaxRoomDescriptionLength = 400

func shrinkRoomMessagesMap(room *domain_structures.Room) {
	//sort existing messages by id to be able to cut older half
	messagesArray := make([]*domain_structures.RoomMessage, len(room.RoomMessages))

	i := 0
	for _, message := range room.RoomMessages {
		messagesArray[i] = message
		i++
	}
	//may be slow but shrinking operation is rare
	sort.Slice(messagesArray, func(i, j int) bool {
		return messagesArray[i].Id < messagesArray[j].Id
	})

	newMessagesArray := make([]*domain_structures.RoomMessage, 0)
	newMessagesArray = append([]*domain_structures.RoomMessage{}, messagesArray[RoomMessagesLimit/2:]...)

	newRoomMessagesMap := make(map[int64]*domain_structures.RoomMessage, len(newMessagesArray))

	for i := 0; i < len(newMessagesArray); i++ {
		msg := newMessagesArray[i]
		newRoomMessagesMap[msg.Id] = msg
	}

	room.RoomMessages = newRoomMessagesMap
	room.RoomMessagesLen = len(room.RoomMessages)
}

func copyMessageAsDTO(orig *domain_structures.RoomMessage) domain_structures.RoomMessageDTO {
	//safe copy of current message state
	messageSafeCopy := domain_structures.RoomMessage{
		Id:               orig.Id,
		Text:             orig.Text,
		SupportedCount:   orig.SupportedCount,
		RejectedCount:    orig.RejectedCount,
		LastEditedAt:     orig.LastEditedAt,
		LastVotedAt:      orig.LastVotedAt,
		ReplyToUserId:    orig.ReplyToUserId,
		ReplyToMessageId: orig.ReplyToMessageId,
		UserInRoomUUID:   orig.UserInRoomUUID,
		CreatedAtSec:     orig.CreatedAtSec,
	}

	return domain_structures.RoomMessageDTO{
		Id:               &messageSafeCopy.Id,
		Text:             &messageSafeCopy.Text,
		SupportedCount:   &messageSafeCopy.SupportedCount,
		RejectedCount:    &messageSafeCopy.RejectedCount,
		LastEditedAt:     messageSafeCopy.LastEditedAt,
		LastVotedAt:      messageSafeCopy.LastVotedAt,
		ReplyToUserId:    messageSafeCopy.ReplyToUserId,
		ReplyToMessageId: messageSafeCopy.ReplyToMessageId,
		UserInRoomUUID:   &messageSafeCopy.UserInRoomUUID,
		CreatedAtSec:     &messageSafeCopy.CreatedAtSec,
	}
}

func copyEditedMessageAsDTO(orig *domain_structures.RoomMessage) domain_structures.RoomMessageDTO {
	//safe copy of current message state
	messageSafeCopy := domain_structures.RoomMessage{
		Id:               orig.Id,
		Text:             orig.Text,
		LastEditedAt:     orig.LastEditedAt,
		ReplyToUserId:    orig.ReplyToUserId,
		ReplyToMessageId: orig.ReplyToMessageId,
	}

	return domain_structures.RoomMessageDTO{
		Id:               &messageSafeCopy.Id,
		Text:             &messageSafeCopy.Text,
		LastEditedAt:     messageSafeCopy.LastEditedAt,
		ReplyToUserId:    messageSafeCopy.ReplyToUserId,
		ReplyToMessageId: messageSafeCopy.ReplyToMessageId,
	}
}

func copyVotedMessageAsDTO(orig *domain_structures.RoomMessage) domain_structures.RoomMessageDTO {
	//safe copy of current message state
	messageSafeCopy := domain_structures.RoomMessage{
		Id:             orig.Id,
		SupportedCount: orig.SupportedCount,
		RejectedCount:  orig.RejectedCount,
	}

	return domain_structures.RoomMessageDTO{
		Id:             &messageSafeCopy.Id,
		SupportedCount: &messageSafeCopy.SupportedCount,
		RejectedCount:  &messageSafeCopy.RejectedCount,
	}
}

func copyAllRoomMessagesAsDTOArray(roomMessages *map[int64]*domain_structures.RoomMessage) *[]domain_structures.RoomMessageDTO {
	dtoArray := make([]domain_structures.RoomMessageDTO, len(*roomMessages))

	i := 0
	for _, orig := range *roomMessages {
		//safe copy of current message state
		messageSafeCopy := domain_structures.RoomMessage{
			Id:               orig.Id,
			Text:             orig.Text,
			SupportedCount:   orig.SupportedCount,
			RejectedCount:    orig.RejectedCount,
			LastEditedAt:     orig.LastEditedAt,
			LastVotedAt:      orig.LastVotedAt,
			ReplyToUserId:    orig.ReplyToUserId,
			ReplyToMessageId: orig.ReplyToMessageId,
			UserInRoomUUID:   orig.UserInRoomUUID,
			CreatedAtSec:     orig.CreatedAtSec,
		}

		dtoArray[i] = domain_structures.RoomMessageDTO{
			Id:               &messageSafeCopy.Id,
			Text:             &messageSafeCopy.Text,
			SupportedCount:   &messageSafeCopy.SupportedCount,
			RejectedCount:    &messageSafeCopy.RejectedCount,
			LastEditedAt:     messageSafeCopy.LastEditedAt,
			LastVotedAt:      messageSafeCopy.LastVotedAt,
			ReplyToUserId:    messageSafeCopy.ReplyToUserId,
			ReplyToMessageId: messageSafeCopy.ReplyToMessageId,
			UserInRoomUUID:   &messageSafeCopy.UserInRoomUUID,
			CreatedAtSec:     &messageSafeCopy.CreatedAtSec,
		}

		i++
	}

	return &dtoArray
}

func findSocketBySessionUUID(clientSocketsByUUID *map[string]*domain_structures.WebSocket, sessionUUID string) *domain_structures.WebSocket {
	for _, socket := range *clientSocketsByUUID {
		if socket.SessionUUID == sessionUUID {
			return socket
		}
	}

	return nil
}

// equivalent of buffered channel with unlimited size
func makeFlexibleChannelPair() (chan<- *domain_structures.OutMessageWrapper, <-chan *domain_structures.OutMessageWrapper) {
	in := make(chan *domain_structures.OutMessageWrapper)
	out := make(chan *domain_structures.OutMessageWrapper)

	go func() {
		var inQueue []*domain_structures.OutMessageWrapper

		outCh := func() chan *domain_structures.OutMessageWrapper {
			if len(inQueue) == 0 {
				return nil
			}

			return out
		}

		curVal := func() *domain_structures.OutMessageWrapper {
			if len(inQueue) == 0 {
				return nil
			}

			return inQueue[0]
		}

	loop:
		for {
			select {
			case val, ok := <-in:
				if !ok {
					break loop
				} else {
					inQueue = append(inQueue, val)
				}
			case outCh() <- curVal():
				inQueue = inQueue[1:]
			}
		}

		close(out)
	}()

	return in, out
}

func validateRoomDescription(roomDescription string) error {
	roomDescriptionDecoded, _ := url.QueryUnescape(roomDescription)

	if len([]rune(roomDescriptionDecoded)) > MaxRoomDescriptionLength {
		return errors.New("invalid room description length")
	}
	return nil
}

func isUserOnlineInRoom(room *domain_structures.Room, userInRoomUUID string) *bool {
	isUserOnlineInRoom := false

	for _, activeUserInRoomUUID := range room.ActiveRoomUserUUIDBySessionUUID {
		if activeUserInRoomUUID == userInRoomUUID {
			isUserOnlineInRoom = true
			break
		}
	}

	return &isUserOnlineInRoom
}
