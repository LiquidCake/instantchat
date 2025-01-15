package load_balancing

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"instantchat.rooms/instantchat/aux-srv/internal/config"
	"instantchat.rooms/instantchat/aux-srv/internal/util"
)

type PickBackendResponse struct {
	BackendInstanceAddr          string   `json:"bA"`
	ErrorMessage                 string   `json:"e"`
	AlternativeRoomNamePostfixes []string `json:"aN"`
}

type RoomNameCacheItem struct {
	roomNameLock        sync.Mutex
	lastCheckTimestamp  int64
	backendInstanceAddr string
}

// backend response format
type HwStatusInfo struct {
	AvgRecentCPUUsagePerc float64 `json:"cpu"`
	LastRamUsagePerc      float64 `json:"ram"`
	UsersOnline           int64   `json:"uo"`
	RequestedRoomFound    bool    `json:"rf"`
}

/* Constants */

const RoomNameCacheTTL = 1 * time.Minute

const BackendUnavailableNextCheckDelay = 10 * time.Second

const MaxUsersOnlinePerBackend = 100000
const ConcurrentBackendUsersNotificationThreshold = 100

const MaxBackendRecentCPULoadPercentUntilWarning = 90
const MaxBackendRecentRAMUsagePercentUntilWarning = 85

const BackendHwEndpointPath = "hw"
const BackendHwEndpointRoomNameURLParam = "roomName"

const MaxAlternativeRoomNamePostfixes = 4

/* Variables */

// room name tracking cache - holds room name lock, lastCheckTimestamp, picked backend instance
var roomNamesCache = map[string]*RoomNameCacheItem{}
var roomNamesMutex = sync.Mutex{}

var unavailableBackendsTracking = make(map[string]int64)
var unavailableBackendsTrackingMutex = sync.Mutex{}

var backendHwStatusClient *http.Client = nil

func InitBackendHwStatusHttpClient(unsecureTestMode bool) {
	var tlsConfig *tls.Config = nil

	if unsecureTestMode {
		log.Printf("WARNING: unsecure http client enabled")

		tlsConfig = &tls.Config{
			InsecureSkipVerify: true,
			MinVersion:         tls.VersionTLS10,
		}
	}

	backendHwStatusClient = &http.Client{
		Timeout: 4 * time.Second,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: 2 * time.Second,
			}).DialContext,

			TLSHandshakeTimeout:   3 * time.Second,
			ExpectContinueTimeout: 3 * time.Second,
			ResponseHeaderTimeout: 5 * time.Second,
			IdleConnTimeout:       90 * time.Second,
			MaxConnsPerHost:       0,
			MaxIdleConns:          0,
			MaxIdleConnsPerHost:   20,
			TLSClientConfig:       tlsConfig,
		},
	}
}

func ValidateRoomAndPickBackend(roomName string) PickBackendResponse {
	/* Validate room name */

	errorMessage := util.ValidateRoomName(roomName)

	if errorMessage != "" {
		return PickBackendResponse{
			BackendInstanceAddr: "",
			ErrorMessage:        errorMessage,
		}
	}

	//make copy of unavailableBackendsTracking to use during this check (cuz concurrent routines may also use this map)
	//Update original map with possible new values after this routine ends
	//(a bit of de-sync between copies of this map that belongs to different concurrent routines should not be an issue)
	unavailableBackendsTrackingCopy := make(map[string]int64)

	unavailableBackendsTrackingMutex.Lock()
	for k, v := range unavailableBackendsTracking {
		unavailableBackendsTrackingCopy[k] = v
	}
	unavailableBackendsTrackingMutex.Unlock()

	//put newly found dead backends here, after this routine exists this map's entries should be added to original "unavailableBackendsTracking" map
	newlyFoundUnavailableBackends := make(map[string]int64)

	/* Pick backend for room name:
	find room cache item, lock it, get picked backend instance
	or if none picked yet - do pick and thus "load" cache */

	roomNamesMutex.Lock()

	knownExistingRoomNames := gatherKnownExistingRoomNames()

	cacheItem, found := roomNamesCache[roomName]

	if !found {
		cacheItem = &RoomNameCacheItem{
			roomNameLock:        sync.Mutex{},
			lastCheckTimestamp:  time.Now().UnixNano(),
			backendInstanceAddr: "",
		}

		roomNamesCache[roomName] = cacheItem
	}

	roomNamesMutex.Unlock()

	cacheItem.roomNameLock.Lock()
	defer func() {
		cacheItem.roomNameLock.Unlock()

		//after this routine ends and all locks it uses are cancelled - take lock for original newlyFoundUnavailableBackends map
		//and update it with all dead backends, found during current run
		unavailableBackendsTrackingMutex.Lock()
		for instance, lastCheckedAt := range newlyFoundUnavailableBackends {
			existingLastCheckedAt, ok := unavailableBackendsTracking[instance]

			if !ok || lastCheckedAt > existingLastCheckedAt {
				unavailableBackendsTracking[instance] = lastCheckedAt
			}
		}
		unavailableBackendsTrackingMutex.Unlock()
	}()

	/* If backend was already picked for this room but last "room exists on backend" check was too long ago - check if room still exists on that backend */

	roomDisappearedFromBackend := false

	if cacheItem.backendInstanceAddr != "" && time.Now().UnixNano()-cacheItem.lastCheckTimestamp > RoomNameCacheTTL.Nanoseconds() {
		hwStatusResponse, err := getHwInfoStatusFromBackend(cacheItem.backendInstanceAddr, roomName)

		if err != nil {
			hwStatusResponse, err = getHwInfoStatusFromBackend(cacheItem.backendInstanceAddr, roomName)

			if err != nil {
				util.LogSevere("ERROR: After 2nd attempt - failed to query backend '%s' for 'hw status'. Error: %s",
					cacheItem.backendInstanceAddr, err)

				newlyFoundUnavailableBackends[cacheItem.backendInstanceAddr] = time.Now().UnixNano()
				go util.NotifyBackendUnavailable(cacheItem.backendInstanceAddr)
			}
		}

		gotEmptyValues := err == nil && (hwStatusResponse.AvgRecentCPUUsagePerc == -1 || hwStatusResponse.LastRamUsagePerc == -1)

		if gotEmptyValues {
			util.LogSevere("ERROR: backend '%s' failed to calculate current CPU('%d') or RAM('%d') load",
				cacheItem.backendInstanceAddr, hwStatusResponse.AvgRecentCPUUsagePerc, hwStatusResponse.LastRamUsagePerc)
		}

		//if room still exists on backend - just update last check timestamp
		if err == nil {
			if hwStatusResponse.RequestedRoomFound {
				cacheItem.lastCheckTimestamp = time.Now().UnixNano()
			} else {
				//if either failed to query backend or got -1 values or room disappeared from it - pick new backend for this room
				roomDisappearedFromBackend = true
			}
		}
	}

	/* If backend was not picked yet for this room - pick least-loaded backend instance and save to cache */

	if cacheItem.backendInstanceAddr == "" || roomDisappearedFromBackend {
		//if failed to query some backend - it just wont be in this map
		hwStatusResponsesByInstance := queryAllBackendsForHwStatus(
			roomName, &unavailableBackendsTrackingCopy, &newlyFoundUnavailableBackends)

		if len(hwStatusResponsesByInstance) == 0 {
			util.LogSevere("Error while querying backends for for 'hw status', all requests failed")

			return PickBackendResponse{
				BackendInstanceAddr: "",
				ErrorMessage:        "server error, please try again a bit later ('cannot pick backend')",
			}
		}

		roomAlreadyExists := false

		for backendInstance, hwStatus := range hwStatusResponsesByInstance {
			//if room already exists on this backend
			if hwStatus.RequestedRoomFound {
				roomAlreadyExists = true

				cacheItem.backendInstanceAddr = backendInstance
				cacheItem.lastCheckTimestamp = time.Now().UnixNano()

				break
			}
		}

		if !roomAlreadyExists {
			leastLoadedBackendInstance, err := findLeastLoadedBackend(&hwStatusResponsesByInstance)

			if err != nil {
				util.LogSevere("Error while looking for least-loaded backend: '%s'", err)

				return PickBackendResponse{
					BackendInstanceAddr: "",
					ErrorMessage:        fmt.Sprintf("server error, please try again a bit later ('%s')", err),
				}
			}

			cacheItem.backendInstanceAddr = leastLoadedBackendInstance
			cacheItem.lastCheckTimestamp = time.Now().UnixNano()
		}
	}

	// returning copy of string from within cache item while it is under lock, so thread-safe
	return PickBackendResponse{
		BackendInstanceAddr:          cacheItem.backendInstanceAddr,
		ErrorMessage:                 "",
		AlternativeRoomNamePostfixes: findAlternativeRoomNamePostfixes(roomName, knownExistingRoomNames),
	}
}

// this is a simple method that just checks if room is known to exist on some backend at some point.
// used for direct message retrieval with loose room/backend relation guaranty
func GetRoomBackend(roomName string) PickBackendResponse {
	roomNamesMutex.Lock()

	cacheItem, ok := roomNamesCache[roomName]

	roomNamesMutex.Unlock()

	if !ok {
		return PickBackendResponse{
			BackendInstanceAddr:          "",
			ErrorMessage:                 "failed to identify backend for requested room",
			AlternativeRoomNamePostfixes: []string{},
		}
	}

	cacheItem.roomNameLock.Lock()
	defer cacheItem.roomNameLock.Unlock()

	// returning copy of string from within cache item while it is under lock, so thread-safe
	return PickBackendResponse{
		BackendInstanceAddr:          cacheItem.backendInstanceAddr,
		ErrorMessage:                 "",
		AlternativeRoomNamePostfixes: []string{},
	}
}

// creates list of postfixes to propose alternative room names. E.g. if "myroom" is taken - propose "myroom-2", "myroom-100" etc.
func findAlternativeRoomNamePostfixes(roomName string, knownExistingRoomNames []string) []string {
	alternativeRoomNamePostfixes := make([]string, 0)
	freePostfixesFound := 0

	for i := 2; i < 10; i++ {
		iStr := strconv.Itoa(i)

		if !util.ArrayContains(knownExistingRoomNames, roomName+"-"+iStr) {
			alternativeRoomNamePostfixes = append(alternativeRoomNamePostfixes, iStr)
			freePostfixesFound++
		}

		if freePostfixesFound >= (MaxAlternativeRoomNamePostfixes / 2) {
			break
		}
	}

	for i := 100; i < 9999999; i++ {
		iStr := strconv.Itoa(i)

		if !util.ArrayContains(knownExistingRoomNames, roomName+"-"+iStr) {
			alternativeRoomNamePostfixes = append(alternativeRoomNamePostfixes, iStr)
			freePostfixesFound++
		}

		if freePostfixesFound >= MaxAlternativeRoomNamePostfixes {
			break
		}

	}

	return alternativeRoomNamePostfixes
}

func queryAllBackendsForHwStatus(
	roomName string,
	unavailableBackendsTrackingCopy *map[string]int64,
	newlyFoundUnavailableBackends *map[string]int64,
) map[string]HwStatusInfo {
	newlyFoundUnavailableBackendsMutex := sync.Mutex{}

	hwStatusResponsesByInstance := make(map[string]HwStatusInfo, 0)
	hwStatusResponsesMutex := sync.Mutex{}

	waitGroup := sync.WaitGroup{}

	for _, backendInstance := range config.AppConfig.BackendInstances {
		lastUnavailableTimestamp, exists := (*unavailableBackendsTrackingCopy)[backendInstance]
		//if this backend was unavailable recently - skip further checks until delay time passes
		if exists && time.Now().UnixNano()-lastUnavailableTimestamp < BackendUnavailableNextCheckDelay.Nanoseconds() {
			continue
		}

		waitGroup.Add(1)

		backendInstance := backendInstance

		go func() {
			defer waitGroup.Done()

			hwStatusResponse, err := getHwInfoStatusFromBackend(backendInstance, roomName)

			if err != nil {
				hwStatusResponse, err = getHwInfoStatusFromBackend(backendInstance, roomName)

				if err != nil {
					util.LogSevere("ERROR: After 2nd attempt - failed to query backend '%s' for 'hw status'. Error: %s",
						backendInstance, err)

					newlyFoundUnavailableBackendsMutex.Lock()
					(*newlyFoundUnavailableBackends)[backendInstance] = time.Now().UnixNano()
					newlyFoundUnavailableBackendsMutex.Unlock()

					go util.NotifyBackendUnavailable(backendInstance)
				}
			}

			//skip backend from result map if failed to query it or got -1 values

			gotEmptyValues := err == nil &&
				(hwStatusResponse.AvgRecentCPUUsagePerc == -1 || hwStatusResponse.LastRamUsagePerc == -1)

			if gotEmptyValues {
				util.LogSevere("ERROR: backend '%s' failed to calculate current CPU('%d') or RAM('%d') load",
					backendInstance, hwStatusResponse.AvgRecentCPUUsagePerc, hwStatusResponse.LastRamUsagePerc)
			}

			//check backend stats, notify
			isBackendInFatalState := performBackendStatsChecks(backendInstance, &hwStatusResponse)

			if err == nil && !gotEmptyValues && !isBackendInFatalState {
				hwStatusResponsesMutex.Lock()
				hwStatusResponsesByInstance[backendInstance] = hwStatusResponse
				hwStatusResponsesMutex.Unlock()
			}
		}()
	}

	waitGroup.Wait()

	return hwStatusResponsesByInstance
}

func performBackendStatsChecks(backendInstance string, hwStatusResponse *HwStatusInfo) bool {
	isFatal := false

	//max users online is a general limit + safeguard for reaching OS file descriptors limit (it doesnt count exactly sockets but should be enough)
	tooManyUsersOnBackend := hwStatusResponse.UsersOnline >= MaxUsersOnlinePerBackend

	if tooManyUsersOnBackend {
		util.LogSevere("ERROR: backend '%s' has >= maximum allowed users connected: %d", backendInstance, MaxUsersOnlinePerBackend)

		go util.NotifyBackendIsFull(backendInstance)

		isFatal = true
	}

	if hwStatusResponse.UsersOnline >= ConcurrentBackendUsersNotificationThreshold {
		util.LogInfo("Concurrent users threshold reached (%s) on backend '%s', sending email notification",
			ConcurrentBackendUsersNotificationThreshold, backendInstance)

		go util.NotifyUsersCountReached(backendInstance, ConcurrentBackendUsersNotificationThreshold)
	}

	if hwStatusResponse.AvgRecentCPUUsagePerc >= MaxBackendRecentCPULoadPercentUntilWarning {
		util.LogWarn("WARN: backend '%s' has >= %d percent recent CPU load", backendInstance, MaxBackendRecentCPULoadPercentUntilWarning)

		go util.NotifyBackendCpuOverload(backendInstance)
	}

	if hwStatusResponse.LastRamUsagePerc >= MaxBackendRecentRAMUsagePercentUntilWarning {
		util.LogWarn("WARN: backend '%s' has >= %d percent recent RAM usage", backendInstance, MaxBackendRecentRAMUsagePercentUntilWarning)

		go util.NotifyBackendRamOverload(backendInstance)
	}

	return isFatal
}

func getHwInfoStatusFromBackend(backendInstance string, roomName string) (HwStatusInfo, error) {
	r, err := backendHwStatusClient.Get(fmt.Sprintf("%s://%s/%s?%s=%s",
		config.AppConfig.BackendHttpSchema,
		backendInstance,
		BackendHwEndpointPath,
		BackendHwEndpointRoomNameURLParam,
		roomName,
	))

	hwStatusResponse := HwStatusInfo{}

	if err != nil {
		util.LogWarn("Failed to query backend '%s' for 'hw status'. Error: %s", backendInstance, err)

		return hwStatusResponse, err
	} else {
		defer r.Body.Close()

		decoder := json.NewDecoder(r.Body)
		err = decoder.Decode(&hwStatusResponse)

		if err != nil {
			util.LogWarn("Failed to parse 'hw status' response from backend '%s'. Error: %s", backendInstance, err)

			return hwStatusResponse, err
		}

		util.LogTrace("Got HW status response from backend '%s'. avg CPU: '%f', RAM: '%f', usersOnline: '%d'",
			backendInstance, hwStatusResponse.AvgRecentCPUUsagePerc, hwStatusResponse.LastRamUsagePerc, hwStatusResponse.UsersOnline)

		return hwStatusResponse, nil
	}
}

func findLeastLoadedBackend(hwStatusResponsesByInstance *map[string]HwStatusInfo) (string, error) {
	//list is sorted by RAM, then if equal - by CPU
	//RAM values are ceiled to 10th during sorting, so e.g. 16 becomes 20 etc.
	//needed to average RAM load a bit and allow picking node with lower CPU between several nodes with close RAM load values
	backendInfoPairList := sortMapByValue(hwStatusResponsesByInstance)

	var loadMarginStep float64 = 10

	//look for node with less load (both RAM and CPU) than current loadMarginStep value, then increment margin step and look further.
	//So e.g. for loadMarginStep 10 - look for node with load RAM<=10% / CPU<=10%, for loadMarginStep 20 - RAM<=20% / CPU<=20% etc.
	// this cycle is needed to find more balanced value, e.g. RAM-6/CPU-60 will be sorted to lower position in list but if there is also value like RAM-20/CPU-20 - we will prefer it to 6/60

	//adjustment margin to allow higher CPU value to pass the check. Needed to prioritize e.g. RAM-10%/CPU-90% over RAM-80%/CPU-10%
	//(if lower RAM is considered more important. Currently it isn't)
	const CPUAdjustmentMargin = 0

	var leastLoadedNode *Pair

outer:
	for {
		if loadMarginStep > 100 {
			break
		}

		for _, p := range backendInfoPairList {
			if p.Value.LastRamUsagePerc <= loadMarginStep && p.Value.AvgRecentCPUUsagePerc <= (loadMarginStep+CPUAdjustmentMargin) {
				leastLoadedNode = &p
				break outer
			}
		}

		loadMarginStep += 10
	}

	if leastLoadedNode != nil {
		return leastLoadedNode.Key, nil
	} else {
		return "", errors.New("failed to find least-loaded backend")
	}
}

// must be executed under room names cache lock
func gatherKnownExistingRoomNames() []string {
	knownExistingRoomNames := make([]string, 0, len(roomNamesCache))

	for roomNameKey := range roomNamesCache {
		knownExistingRoomNames = append(knownExistingRoomNames, roomNameKey)
	}

	return knownExistingRoomNames
}
