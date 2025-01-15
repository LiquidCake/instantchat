package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"instantchat.rooms/instantchat/backend/internal/domain_structures"
	"instantchat.rooms/instantchat/backend/internal/util"
)

const AvgRoomMessagesTimerRunInterval = 60 * time.Second

var stopAvgRoomMessagesTimers = false

func StartAvgRoomMessagesGaugeTimer(gauge *prometheus.Gauge, activeRoomsByNameMap *domain_structures.ActiveRoomsByName) {
	util.LogInfo("Starting AvgRoomMessages gauge")

	timer := time.NewTimer(AvgRoomMessagesTimerRunInterval)

	go runAvgRoomMessagesTimer(gauge, timer, activeRoomsByNameMap)
}

func runAvgRoomMessagesTimer(gauge *prometheus.Gauge, timer *time.Timer, activeRoomsByNameMap *domain_structures.ActiveRoomsByName) {
	if stopAvgRoomMessagesTimers {
		return
	}

	//wait for timer to tick
	<-timer.C

	if stopAvgRoomMessagesTimers {
		return
	}

	activeRoomsByNameMap.Lock()

	totalMessages := 0

	for _, room := range activeRoomsByNameMap.ActiveRoomsByName {
		if !room.IsDeleted {
			totalMessages += room.RoomMessagesLen
		}
	}

	roomCount := len(activeRoomsByNameMap.ActiveRoomsByName)
	if roomCount == 0 {
		(*gauge).Set(0)
	} else {
		(*gauge).Set(float64(totalMessages / roomCount))
	}

	activeRoomsByNameMap.Unlock()

	util.LogTrace("Restarting AvgRoomMessages gauge timer")

	timer.Reset(AvgRoomMessagesTimerRunInterval)

	go runAvgRoomMessagesTimer(gauge, timer, activeRoomsByNameMap)
}

func StopAvgRoomMessagesGaugeTimer() {
	util.LogInfo("Stopping AvgRoomMessages gauge timer")

	stopAvgRoomMessagesTimers = true
}
