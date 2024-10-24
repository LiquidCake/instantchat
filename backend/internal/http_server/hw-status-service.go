package http_server

import (
	"errors"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"instantchat.rooms/instantchat/backend/internal/util"
)

type HwStatusInfo struct {
	AvgRecentCPUUsagePerc float64 `json:"cpu"`
	LastRamUsagePerc      float64 `json:"ram"`
	UsersOnline           int64   `json:"uo"`
	RequestedRoomFound    bool    `json:"rf"`
}

const MeasurementsMaxTicks = 3

var avgRecentCPUUsagePerc float64 = 0
var lastRamUsagePerc float64 = 0

// 12 slots for measurement each 5sec during 1 minute
var cpuMeasurementsArr [12]float64
var measurementsCounter = 0

func getCPUSample() (idle, total uint64, error error) {
	contents, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		error = err
		return
	}

	lines := strings.Split(string(contents), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)

		if fields[0] == "cpu" {
			numFields := len(fields)

			for i := 1; i < numFields; i++ {
				val, err := strconv.ParseUint(fields[i], 10, 64)
				if err != nil {
					error = err
					return
				}

				total += val // sum all types of values to get total ticks

				if i == 4 { // idle is the 5th field in the cpu line
					idle = val
				}
			}
			return
		}
	}

	error = errors.New("failed to read data from /proc/stat")
	return
}

func getRAMUsage() (available, total uint64, error error) {
	contents, err := ioutil.ReadFile("/proc/meminfo")
	if err != nil {
		error = err
		return
	}

	//make sure both fields were read successfully
	measuredFields := 0

	lines := strings.Split(string(contents), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)

		if len(fields) > 2 {
			if fields[0] == "MemTotal:" {
				total, err = strconv.ParseUint(strings.TrimSpace(fields[1]), 10, 64)
				if err != nil {
					error = err
					return
				}
				measuredFields += 1
			}

			if fields[0] == "MemAvailable:" {
				available, err = strconv.ParseUint(strings.TrimSpace(fields[1]), 10, 64)
				if err != nil {
					error = err
					return
				}
				measuredFields += 1
			}
		}
	}

	if measuredFields < 2 {
		error = errors.New("failed to read data from /proc/meminfo")
	}
	return
}

func startMeasuringHardwareStatus() {
	go func() {
		//cycle breaks on app shutdown
		for {
			//Measure CPU

			idle0, total0, err := getCPUSample()
			if err != nil {
				util.LogSevere("Failed to measure CPU (1st measurement): '%s'", err)
				avgRecentCPUUsagePerc = -1
				lastRamUsagePerc = -1

				continue
			}

			//measure CPU usage for last 5sec
			time.Sleep(3 * time.Second)

			idle1, total1, err := getCPUSample()
			if err != nil {
				util.LogSevere("Failed to measure CPU (2nd measurement): '%s'", err)
				avgRecentCPUUsagePerc = -1
				lastRamUsagePerc = -1

				continue
			}

			idleTicks := float64(idle1 - idle0)
			totalTicks := float64(total1 - total0)

			cpuUsagePerc := 100 * (totalTicks - idleTicks) / totalTicks

			//util.LogTrace("CPU usage is %f%% [busy: %f, total: %f]\n", cpuUsagePerc, totalTicks-idleTicks, totalTicks)

			//Measure RAM

			ramAvailable, ramTotal, err := getRAMUsage()
			if err != nil {
				util.LogSevere("Failed to measure RAM: '%s'", err)
				avgRecentCPUUsagePerc = -1
				lastRamUsagePerc = -1

				continue
			}

			ramUsagePerc := 100 * (ramTotal - ramAvailable) / ramTotal

			//util.LogTrace("RAM usage is %d%% [available: %d, total: %d]", ramUsagePerc, ramAvailable, ramTotal)

			cpuMeasurementsArr[measurementsCounter] = cpuUsagePerc
			lastRamUsagePerc = float64(ramUsagePerc)

			// find the average CPU load from all recent measurements
			if measurementsCounter >= MeasurementsMaxTicks {
				measurementsCounter = 0

				var sum float64 = 0
				for i := 0; i < len(cpuMeasurementsArr); i++ {
					sum += cpuMeasurementsArr[i]
				}

				avgRecentCPUUsagePerc = sum / float64(len(cpuMeasurementsArr))
			} else {
				measurementsCounter += 1
			}
		}
	}()
}
