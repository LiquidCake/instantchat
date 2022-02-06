package engine

import (
	"instantchat.rooms/instantchat/backend/internal/domain_structures"
	"github.com/gorilla/websocket"
	"sync"
	"time"

	"instantchat.rooms/instantchat/backend/internal/util"
)

const HouseKeeperRunInterval = 60 * time.Second
const RoomEmptyTTL = 5 * time.Minute
const RoomInactiveTTL = 12 * time.Hour

var stopTimers = false

func StartSocketHouseKeeper(room *domain_structures.Room) {
	util.LogInfo("Starting SocketHouseKeeper for room '%s' / '%s'", room.Id, room.Name)

	timer := time.NewTimer(HouseKeeperRunInterval)

	go runTimer(room, timer)
}

func runTimer(room *domain_structures.Room, timer *time.Timer) {
	//wait for timer to tick
	<-timer.C

	if stopTimers || room.IsDeleted {
		return
	}

	//check sockets connected to room

	roomActiveClientSocketsByUUID, _ := room.CopyActiveClientSocketMap()

	if len(*roomActiveClientSocketsByUUID) <= 0 {
		//if room is empty, old enough and has been inactive for enough time - delete it
		if (time.Now().UnixNano() - room.StartedAt) >= RoomEmptyTTL.Nanoseconds() &&
			(time.Now().UnixNano() - room.LastActiveAt) >= RoomInactiveTTL.Nanoseconds() {
			deleted := tryDeleteEmptyRoom(room, roomActiveClientSocketsByUUID)

			if deleted {
				util.LogInfo("SocketHouseKeeper deleted old empty room '%s' / '%s'. Exiting", room.Id, room.Name)

				return
			}
		}
	} else {
		//check active sockets health
		var foundDeadSocketsById = make(map[string]*domain_structures.WebSocket)
		foundDeadSocketsByIdMutex := sync.Mutex{}

		waitGroup := sync.WaitGroup{}

		//check each socket and close if it is dead
		for _, clSocket := range *roomActiveClientSocketsByUUID {
			waitGroup.Add(1)

			clSocket := clSocket

			go func() {
				defer waitGroup.Done()

				var lastKeepAliveSec = (time.Now().UnixNano() - clSocket.LastKeepAliveSignal) / time.Second.Nanoseconds()

				if clSocket.IsDead() {
					util.LogTrace("SocketHouseKeeper found dead socket. Last keep alive ago: '%d's Room '%s' / '%s', socket '%s'",
						lastKeepAliveSec, room.Id, room.Name, clSocket.SocketUUID)

					foundDeadSocketsByIdMutex.Lock()
					foundDeadSocketsById[clSocket.SocketUUID] = clSocket
					foundDeadSocketsByIdMutex.Unlock()

					return
				}

				err := clSocket.Socket.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(SocketWriteTimeout))

				if err == nil {
					util.LogTrace("SocketHouseKeeper socket OK. Last keep alive ago: '%d's Room '%s' / '%s', socket '%s'",
						lastKeepAliveSec, room.Id, room.Name, clSocket.SocketUUID)
				} else {
					util.LogTrace("SocketHouseKeeper checked and found dead socket. Last keep alive ago: '%d's Room '%s' / '%s', socket '%s'",
						lastKeepAliveSec, room.Id, room.Name, clSocket.SocketUUID)

					clSocket.Terminate()

					foundDeadSocketsByIdMutex.Lock()
					foundDeadSocketsById[clSocket.SocketUUID] = clSocket
					foundDeadSocketsByIdMutex.Unlock()
				}
			}()
		}

		waitGroup.Wait()

		//if all sockets disconnected AND room has been inactive for enough time - delete room (wont delete if new sockets opened during check)
		if len(foundDeadSocketsById) == len(*roomActiveClientSocketsByUUID) &&
			(time.Now().UnixNano() - room.LastActiveAt) >= RoomInactiveTTL.Nanoseconds() {

			deleted := tryDeleteEmptyRoom(room, roomActiveClientSocketsByUUID)
			if deleted {
				util.LogInfo("SocketHouseKeeper deleted empty room '%s' / '%s'. Exiting", room.Id, room.Name)

				return
			}
		}

		//if found dead sockets - clean them from room and write "MembersCountChanged" message to active room members
		if len(foundDeadSocketsById) > 0 {
			membersListChanged := false

			room.Lock()

			for _, clSocket := range foundDeadSocketsById {
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

					if !userAlreadyHasOtherActiveSocket {
						delete(room.ActiveRoomUserUUIDBySessionUUID, clSocket.SessionUUID)
						room.ActiveRoomUsersLen = len(room.ActiveRoomUserUUIDBySessionUUID)

						membersListChanged = true
					}
				}
			}

			room.Unlock()

			if membersListChanged {
				writeMembersListChangedFrameToActiveRoomMembers(room, nil)
			}
		}
	}

	util.LogTrace("Restarting SocketHouseKeeper for room '%s' / '%s'", room.Id, room.Name)

	timer.Reset(HouseKeeperRunInterval)

	go runTimer(room, timer)
}

func tryDeleteEmptyRoom(room *domain_structures.Room, roomActiveClientSocketsByUUID *map[string]*domain_structures.WebSocket) bool {
	room.Lock()
	//check no new sockets were added to room while we were checking original list
	if util.MapKeySetsAreEqual(roomActiveClientSocketsByUUID, &room.ActiveClientSocketsByUUID) {
		room.ActiveRoomUsersLen = 0
		room.IsDeleted = true
		room.Unlock()

		ActiveRoomsByNameMap.Delete(room.Name)
		RoomsOnlineGauge.Dec()

		return true
	} else {
		room.Unlock()

		return false
	}
}

func StopSocketHouseKeeperRoutines() {
	util.LogInfo("Stopping StopSocketHouseKeeperRoutines")

	stopTimers = true
}
