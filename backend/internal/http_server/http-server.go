package http_server

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v2"
	"instantchat.rooms/instantchat/backend/internal/domain_structures"
	"instantchat.rooms/instantchat/backend/internal/metrics"

	"instantchat.rooms/instantchat/backend/internal/config"
	"instantchat.rooms/instantchat/backend/internal/engine"
	"instantchat.rooms/instantchat/backend/internal/templates"
	"instantchat.rooms/instantchat/backend/internal/util"
)

/* Constants */

const BackendHostURLParam = "backendHost"
const RoomNameURLParam = "roomName"
const DeleteRoomNameURLParam = "deleteRoomName"
const DirectMessageTextURLParam = "m"
const DirectMessagesRoomPasswordURLParam = "p"
const DirectMessagesLimitParam = "l"
const DirectMessagesIdParam = "id"
const DirectMessagesQuiteModeParam = "quite"
const DirectMessagesResponseFormatParam = "format"
const CtrlCommandURLParam = "ctrlCommand"

const CtrlCommandNotifyShutdown = "notify_shutdown"
const CtrlCommandNotifyRestart = "notify_restart"

var BackendInstanceNum string

/* App configs */

// Set in yaml app-config
var HttpPort = ":8080"
var HttpTimeout = 30 * time.Second

var ShutdownWaitTimeout = 10 * time.Second

var LogMaxSizeMb = 500
var LogMaxFilesToKeep = 3
var LogMaxFileAgeDays = 60

var CtrlAuthLogin = "admin132"
var CtrlAuthPasswd = "password132"

var Domain = "n/a"
var HttpSchema = "n/a"

type CtrlCommandResponse struct {
	Command      string `json:"command"`
	Result       string `json:"result"`
	ErrorMessage string `json:"errorMessage"`
}

func StartServer() {
	rand.Seed(time.Now().UnixNano())

	// Read app configs
	loadAppConfigs()

	// Configure Logging
	LOG_DIR := os.Getenv("LOG_DIR")
	LOG_FILE_NAME := os.Getenv("LOG_FILE_NAME")
	if LOG_DIR != "" && LOG_FILE_NAME != "" {
		logFile := &lumberjack.Logger{
			Filename:   LOG_DIR + "/" + LOG_FILE_NAME,
			MaxSize:    LogMaxSizeMb,
			MaxBackups: LogMaxFilesToKeep,
			MaxAge:     LogMaxFileAgeDays,
			Compress:   true, // disabled by default
		}

		mw := io.MultiWriter(os.Stdout, logFile)
		log.SetOutput(mw)

		log.Printf("File logger initialized: %s/%s", LOG_DIR, LOG_FILE_NAME)
	} else {
		log.Println("File logger skipped, no env vars LOG_DIR, LOG_FILE_NAME")
	}

	LOG_LEVEL := os.Getenv("LOG_LEVEL")
	if LOG_LEVEL == "" {
		LOG_LEVEL = "INFO"
	}

	switch LOG_LEVEL {
	case "TRACE":
		util.CurrentLogLevel = util.Trace
		break
	case "INFO":
		util.CurrentLogLevel = util.Info
		break
	case "WARN":
		util.CurrentLogLevel = util.Warn
		break
	case "SEVERE":
		util.CurrentLogLevel = util.Severe
		break
	}

	util.LogInfo("Log level: '%s'", LOG_LEVEL)

	// Setup metrics
	setupMetrics()

	// Create Server and Route Handlers
	router := mux.NewRouter()

	router.HandleFunc("/metrics", promhttp.Handler().ServeHTTP)

	router.HandleFunc("/ctrl", middleware(ctrlHandler, basicAuthWrapper, loggingWrapper))
	router.HandleFunc("/ctrl_rooms", middleware(roomsCtrlHandler, basicAuthWrapper))
	router.HandleFunc("/ctrl_command", middleware(ctrlCommandHandler, basicAuthWrapper))

	router.HandleFunc("/ws_entry", middleware(websocketHandler, loggingWrapper))
	router.HandleFunc("/direct_sending", middleware(directlySendRoomMessageHandler, loggingWrapper))
	router.HandleFunc("/direct_retrieval", middleware(directlyRetrieveRoomMessagesHandler, loggingWrapper))
	router.HandleFunc("/hw", middleware(hwStatusHandler, loggingWrapper))

	cert, err := tls.LoadX509KeyPair("/etc/ssl/ssl-bundle.crt", "/etc/ssl/cert.key")

	if err != nil {
		log.Fatal(err)
	}

	srv := &http.Server{
		Handler:      router,
		Addr:         HttpPort,
		ReadTimeout:  HttpTimeout,
		WriteTimeout: HttpTimeout,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
	}

	BackendInstanceNum = os.Getenv("INSTANCE_NUM")
	util.LogInfo("Backend instance num: '%s'", BackendInstanceNum)

	// Start Server
	go func() {
		util.LogInfo("Starting Server on '%s'", HttpPort)
		if err := srv.ListenAndServeTLS("", ""); err != nil {
			log.Fatal(err)
		}
	}()

	startMeasuringHardwareStatus()

	// Graceful Shutdown
	waitForShutdown(srv)
}

/* handlers */

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	engine.WsEntry(w, r)
}

func directlyRetrieveRoomMessagesHandler(w http.ResponseWriter, r *http.Request) {
	var responseTextBytes []byte

	responseFormat := util.GetUnescapedRequestParamValueUnsafe(r, DirectMessagesResponseFormatParam)

	roomName := util.GetUnescapedRequestParamValueUnsafe(r, RoomNameURLParam)
	roomName = strings.ToLower(roomName)

	if roomName == "" {
		responseTextBytes = util.BuildDirectRoomMessagesErrorResponse("error: bad room name", responseFormat)

	} else {
		roomPassword := util.GetUnescapedRequestParamValueUnsafe(r, DirectMessagesRoomPasswordURLParam)

		messagesLimitVal := util.GetUnescapedRequestParamValueUnsafe(r, DirectMessagesLimitParam)

		messagesLimit, err := strconv.Atoi(messagesLimitVal)
		if err != nil {
			messagesLimit = 0
		}

		messageIdVal := util.GetUnescapedRequestParamValueUnsafe(r, DirectMessagesIdParam)

		messageId, err := strconv.Atoi(messageIdVal)
		if err != nil {
			messageId = 0
		}

		quiteMode := util.GetUnescapedRequestParamValueUnsafe(r, DirectMessagesQuiteModeParam) == "true"

		responseTextBytes = engine.RetrieveRoomMessagesDirectly(
			roomName, roomPassword, messagesLimit, int64(messageId), responseFormat, quiteMode)
	}

	writeDirectMessagesResponse(w, responseTextBytes, "directly retrieve messages", responseFormat)
}

func directlySendRoomMessageHandler(w http.ResponseWriter, r *http.Request) {
	var responseTextBytes []byte

	responseFormat := util.GetUnescapedRequestParamValueUnsafe(r, DirectMessagesResponseFormatParam)

	roomName := util.GetUnescapedRequestParamValueUnsafe(r, RoomNameURLParam)
	roomName = strings.ToLower(roomName)

	if roomName == "" {
		responseTextBytes = util.BuildDirectRoomMessagesErrorResponse("error: bad room name", responseFormat)

	} else {
		roomPassword := util.GetUnescapedRequestParamValueUnsafe(r, DirectMessagesRoomPasswordURLParam)

		messageText := util.GetRequestParamValue(r, DirectMessageTextURLParam)
		messageText = url.QueryEscape(messageText) //properly escape input (we are always storing escaped message text)

		if len(strings.TrimSpace(messageText)) == 0 {
			responseTextBytes = util.BuildDirectRoomMessagesErrorResponse("error: empty message", responseFormat)

		} else if len(messageText) >= util.MaxMessageLength {
			responseTextBytes = util.BuildDirectRoomMessagesErrorResponse("error: message is too long", responseFormat)

		} else {
			responseTextBytes = engine.SendRoomMessageDirectly(roomName, roomPassword, messageText, responseFormat)
		}
	}

	writeDirectMessagesResponse(w, responseTextBytes, "directly send messages", responseFormat)
}

func writeDirectMessagesResponse(w http.ResponseWriter, responseTextBytes []byte, requestType string, responseFormat string) {
	var contentType = "text/plain; charset=utf-8"

	if responseFormat == "json" {
		contentType = "application/json"
	}

	w.Header().Set("Content-Type", contentType)
	_, err := w.Write(responseTextBytes)

	if err != nil {
		util.LogWarn("Failed to write response for '%s' request. err: '%s'", requestType, err)
	}
}

func hwStatusHandler(w http.ResponseWriter, r *http.Request) {
	requestedRoomFound := false

	urlParams, ok := r.URL.Query()[RoomNameURLParam]

	if ok && len(urlParams[0]) > 0 {
		requestedRoomFound = engine.ActiveRoomsByNameMap.Contains(urlParams[0])
	}

	jsonData, err := json.Marshal(HwStatusInfo{
		avgRecentCPUUsagePerc,
		lastRamUsagePerc,
		metrics.UsersOnline,
		requestedRoomFound,
	})

	if err != nil {
		util.LogSevere("Failed to serialize structure for 'hw status' request. err: '%s'", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonData)

	if err != nil {
		util.LogSevere("Failed to write response for 'hw status' request. err: '%s'", err)
	} else {
		util.LogTrace("returning hw status: '%s'", jsonData)
	}
}

func ctrlHandler(w http.ResponseWriter, r *http.Request) {
	var session util.HttpSession
	errorStr := ""

	util.LogInfo("ctrl page requested. UUID: '%s'", session)

	err := util.GetUserSession(r, &session)

	if err != nil {
		errorStr = "ERROR: please get session cookie from home/room page"
	}

	renderCtrlPage(w, r, errorStr)
}

func roomsCtrlHandler(w http.ResponseWriter, r *http.Request) {
	//delete room
	urlParams, ok := r.URL.Query()[DeleteRoomNameURLParam]

	if ok && len(urlParams[0]) > 0 {
		roomName := urlParams[0]

		room := engine.ActiveRoomsByNameMap.Get(roomName)

		room.Lock()

		util.LogInfo("Manually deleting room '%s' / '%s'", room.Id, room.Name)
		room.ActiveRoomUsersLen = 0
		room.IsDeleted = true

		room.Unlock()

		engine.ActiveRoomsByNameMap.Delete(room.Name)
		engine.RoomsOnlineGauge.Dec()
	}

	//draw page

	activeRoomsByName, _ := engine.ActiveRoomsByNameMap.CopyActiveRoomsByNameMap()

	var activeRooms []domain_structures.RoomCtrlInfo

	for _, room := range *activeRoomsByName {
		if room.IsDeleted {
			continue
		}

		activeRooms = append(activeRooms, domain_structures.RoomCtrlInfo{
			Id:                 room.Id,
			Name:               room.Name,
			StartedAt:          time.Unix(0, room.StartedAt).String(),
			ActiveRoomUsersNum: room.ActiveRoomUsersLen,
		})
	}

	roomsOrder, err := util.GetCookieValue("X-rooms_order", r)

	if err != nil {
		roomsOrder = "by_name"
	}

	var roomsOrderFunc func(i int, j int) bool

	switch roomsOrder {
	case "by_id":
		roomsOrderFunc = func(i int, j int) bool {
			return activeRooms[i].Id < activeRooms[j].Id
		}
		break

	case "by_name":
		roomsOrderFunc = func(i int, j int) bool {
			return activeRooms[i].Name < activeRooms[j].Name
		}
		break

	case "by_started_at":
		roomsOrderFunc = func(i int, j int) bool {
			return activeRooms[i].StartedAt < activeRooms[j].StartedAt
		}
		break

	case "active_room_users_num":
		roomsOrderFunc = func(i int, j int) bool {
			return activeRooms[i].ActiveRoomUsersNum < activeRooms[j].ActiveRoomUsersNum
		}
		break
	}

	sort.SliceStable(activeRooms, roomsOrderFunc)

	renderRoomsCtrlPage(w, &activeRooms)
}

func renderCtrlPage(w http.ResponseWriter, r *http.Request, errorStr string) {
	urlParams, ok := r.URL.Query()[BackendHostURLParam]

	if !ok || len(urlParams[0]) < 1 {
		util.LogSevere("'%s' param is missing from request", BackendHostURLParam)
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	backendInstanceHost := urlParams[0]

	vars := map[string]interface{}{
		"backendInstanceNum":  BackendInstanceNum,
		"backendInstanceHost": backendInstanceHost,
		"error":               errorStr,
		//reading prometheus metrics is expensive but for admin page its ok
		"activeRooms": testutil.ToFloat64(engine.RoomsOnlineGauge),
		"activeUsers": testutil.ToFloat64(engine.UsersOnlineGauge),
	}

	err := templates.CompiledTemplates.ExecuteTemplate(w, "tpl-ctrl.html", vars)

	if err != nil {
		log.Fatal("Cannot Get View ", err)
	}
}

func renderRoomsCtrlPage(w http.ResponseWriter, activeRooms *[]domain_structures.RoomCtrlInfo) {
	vars := map[string]interface{}{
		"activeRooms": *activeRooms,
	}

	err := templates.CompiledTemplates.ExecuteTemplate(w, "tpl-rooms-ctrl.html", vars)

	if err != nil {
		log.Fatal("Cannot Get View ", err)
	}
}

func ctrlCommandHandler(w http.ResponseWriter, r *http.Request) {
	//delete room
	urlParams, ok := r.URL.Query()[CtrlCommandURLParam]

	if ok && len(urlParams[0]) > 0 {
		ctrlCommand := urlParams[0]

		errorMessage := ""
		result := ""

		switch ctrlCommand {
		case CtrlCommandNotifyShutdown:
			engine.SendControlCommandServerStatusChanged(util.ServerStatusShuttingDown)
			result = "ok"
			break
		case CtrlCommandNotifyRestart:
			engine.SendControlCommandServerStatusChanged(util.ServerStatusRestarting)
			result = "ok"
			break
		default:
			errorMessage = "command_not_found"
		}

		jsonData, err := json.Marshal(CtrlCommandResponse{
			ctrlCommand,
			result,
			errorMessage,
		})

		if err != nil {
			util.LogSevere("Failed to serialize structure for 'ctrl command' request. err: '%s'", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", HttpSchema+"://"+Domain)
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		_, err = w.Write(jsonData)

		if err != nil {
			util.LogSevere("Failed to write response for 'ctrl command' request. err: '%s'", err)
		} else {
			util.LogTrace("returning ctrl command: '%s'", jsonData)
		}
	}
}

/* middleware */

// middleware interface for chaining middleware for single routes. Functions are simple HTTP handlers (w http.ResponseWriter, r *http.Request)
func middleware(handler http.HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}

	return handler
}

// loggingWrapper - middleware that catches panics in handlers, logs the stack trace and and serves a HTTP 500 error
func loggingWrapper(h http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				http.Error(w, "HTTP 500: internal server error", http.StatusInternalServerError)
				util.LogSevere("handler HTTP 500 - log wrapper: %s\n%s\n", err, debug.Stack())
			}
		}()

		h.ServeHTTP(w, r)
	}
}

// basicAuthWrapper - middleware that wraps handler with basic auth
func basicAuthWrapper(h http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

		s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
		if len(s) != 2 {
			http.Error(w, "Not authorized", util.HTTP_STATUS_UNAUTHORIZED)
			return
		}

		b, err := base64.StdEncoding.DecodeString(s[1])
		if err != nil {
			http.Error(w, err.Error(), util.HTTP_STATUS_UNAUTHORIZED)
			return
		}

		pair := strings.SplitN(string(b), ":", 2)
		if len(pair) != 2 {
			http.Error(w, "Not authorized", util.HTTP_STATUS_UNAUTHORIZED)
			return
		}

		if pair[0] != CtrlAuthLogin || pair[1] != CtrlAuthPasswd {
			http.Error(w, "Not authorized", util.HTTP_STATUS_UNAUTHORIZED)
			return
		}

		h.ServeHTTP(w, r)
	}
}

func waitForShutdown(srv *http.Server) {
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive our signal.
	<-interruptChan

	engine.StopSocketHouseKeeperRoutines()

	metrics.StopAvgRoomMessagesGaugeTimer()
	metrics.StopUsersOnlineGaugeTimer()

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), ShutdownWaitTimeout)
	defer cancel()

	srv.Shutdown(ctx)

	util.LogInfo("Shutting down")

	os.Exit(0)
}

// logs in this method are written to stdout so they get into container logs
// but not into application log file that is not yet initialized
func loadAppConfigs() {
	appConfigFile, _ := filepath.Abs("app-config.yml")
	yamlFile, err := os.ReadFile(appConfigFile)
	if err != nil {
		log.Printf("[SEVERE] Failed to read app config file: '%s'", appConfigFile)
		panic(err)
	}
	err = yaml.Unmarshal(yamlFile, &config.AppConfig)
	if err != nil {
		log.Printf("[SEVERE] Failed to parse app config file: '%s'", appConfigFile)
		panic(err)
	}

	// populate configured vars
	HttpPort = config.AppConfig.Server.HttpPort
	HttpTimeout = config.AppConfig.Server.HttpTimeoutSec * time.Second

	ShutdownWaitTimeout = config.AppConfig.ShutdownWaitTimeoutSec * time.Second

	LogMaxSizeMb = config.AppConfig.Logging.LogMaxSizeMb
	LogMaxFilesToKeep = config.AppConfig.Logging.LogMaxFilesToKeep
	LogMaxFileAgeDays = config.AppConfig.Logging.LogMaxFileAgeDays

	Domain = config.AppConfig.Domain
	HttpSchema = config.AppConfig.HttpSchema

	envCtrlAuthLogin := os.Getenv("CTRL_AUTH_LOGIN")
	if envCtrlAuthLogin != "" {
		config.AppConfig.CtrlAuthLogin = envCtrlAuthLogin

		log.Printf("CtrlAuthLogin is overridden using env variable CTRL_AUTH_LOGIN")
	}

	envCtrlAuthPasswd := os.Getenv("CTRL_AUTH_PASSWD")
	if envCtrlAuthPasswd != "" {
		config.AppConfig.CtrlAuthPasswd = envCtrlAuthPasswd

		log.Printf("CtrlAuthPasswd is overridden using env variable CTRL_AUTH_PASSWD")
	}

	CtrlAuthLogin = config.AppConfig.CtrlAuthLogin
	CtrlAuthPasswd = config.AppConfig.CtrlAuthPasswd

	log.Printf("app config: HttpPort='%s'", HttpPort)
	log.Printf("app config: HttpTimeout='%s'", HttpTimeout)
	log.Printf("app config: HttpSchema='%s'", HttpSchema)
	log.Printf("app config: Domain='%s'", Domain)
	log.Printf("app config: ShutdownWaitTimeout='%s'", ShutdownWaitTimeout)
	log.Printf("app config: LogMaxSizeMb='%d'", LogMaxSizeMb)
	log.Printf("app config: LogMaxFilesToKeep='%d'", LogMaxFilesToKeep)
	log.Printf("app config: LogMaxFileAgeDays='%d'", LogMaxFileAgeDays)
}

func setupMetrics() {
	engine.RoomsOnlineGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "rooms_online",
		})
	prometheus.MustRegister(engine.RoomsOnlineGauge)

	engine.UsersOnlineGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "users_online",
		})
	prometheus.MustRegister(engine.UsersOnlineGauge)

	engine.AvgUsersOnlineGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "avg_users_online",
		})
	prometheus.MustRegister(engine.AvgUsersOnlineGauge)

	engine.AvgMessagesPerRoomGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "avg_messages_per_room",
		})
	prometheus.MustRegister(engine.AvgMessagesPerRoomGauge)

	metrics.StartUsersOnlineGaugeTimer(&engine.UsersOnlineGauge, &engine.AvgUsersOnlineGauge, &engine.ActiveRoomsByNameMap)
	metrics.StartAvgRoomMessagesGaugeTimer(&engine.AvgMessagesPerRoomGauge, &engine.ActiveRoomsByNameMap)
}
