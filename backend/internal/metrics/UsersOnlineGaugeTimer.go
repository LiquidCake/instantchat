package metrics

import (
	"instantchat.rooms/instantchat/backend/internal/domain_structures"
	"instantchat.rooms/instantchat/backend/internal/util"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

const UsersOnlineTimerRunInterval = 30 * time.Second

var UsersOnline int64 = 0

var stopUsersOnlineTimers = false

func StartUsersOnlineGaugeTimer(
	usersOnlineGauge *prometheus.Gauge,
	avgUsersOnlineGauge *prometheus.Gauge,
	activeRoomsByNameMap *domain_structures.ActiveRoomsByName,
) {
	util.LogInfo("Starting UsersOnline gauge")

	timer := time.NewTimer(UsersOnlineTimerRunInterval)

	go runUsersOnlineTimer(usersOnlineGauge, avgUsersOnlineGauge, timer, activeRoomsByNameMap)
}

func runUsersOnlineTimer(
	usersOnlineGauge *prometheus.Gauge,
	avgUsersOnlineGauge *prometheus.Gauge,
	timer *time.Timer,
	activeRoomsByNameMap *domain_structures.ActiveRoomsByName,
) {
	//wait for timer to tick
	<-timer.C

	if stopUsersOnlineTimers {
		return
	}

	activeRoomsByNameMap.Lock()

	totalUsers := int64(0)

	for _, room := range activeRoomsByNameMap.ActiveRoomsByName {
		if !room.IsDeleted {
			totalUsers += int64(room.ActiveRoomUsersLen)
		}
	}

	(*usersOnlineGauge).Set(float64(totalUsers))

	roomCount := int64(len(activeRoomsByNameMap.ActiveRoomsByName))
	if roomCount == 0 {
		(*avgUsersOnlineGauge).Set(0)
	} else {
		(*avgUsersOnlineGauge).Set(float64(totalUsers / roomCount))
	}

	UsersOnline = totalUsers

	activeRoomsByNameMap.Unlock()

	util.LogTrace("Restarting UsersOnline gauge timer")

	timer.Reset(UsersOnlineTimerRunInterval)

	go runUsersOnlineTimer(usersOnlineGauge, avgUsersOnlineGauge, timer, activeRoomsByNameMap)
}

func StopUsersOnlineGaugeTimer() {
	util.LogInfo("Stopping UsersOnline gauge timer")

	stopUsersOnlineTimers = true
}
