package http_server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"instantchat.rooms/instantchat/backend/internal/domain_structures"
	"instantchat.rooms/instantchat/backend/internal/metrics"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strings"
	"syscall"
	"time"

	"instantchat.rooms/instantchat/backend/internal/config"
	"instantchat.rooms/instantchat/backend/internal/engine"
	"instantchat.rooms/instantchat/backend/internal/templates"
	"instantchat.rooms/instantchat/backend/internal/util"
)

/* Constants */

const BackendHostURLParam = "backendHost"
const RoomNameURLParam = "roomName"
const DeleteRoomNameURLParam = "deleteRoomName"

var defaultPorts = map[string]string{"http": "80", "https": "443"}

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
	router.HandleFunc("/ws_entry", middleware(websocketHandler, loggingWrapper))
	router.HandleFunc("/hw", middleware(hwStatusHandler, loggingWrapper))

	srv := &http.Server{
		Handler:      router,
		Addr:         HttpPort,
		ReadTimeout:  HttpTimeout,
		WriteTimeout: HttpTimeout,
	}

	BackendInstanceNum = os.Getenv("INSTANCE_NUM")
	util.LogInfo("Backend instance num: '%s'", BackendInstanceNum)

	// Start Server
	go func() {
		util.LogInfo("Starting Server on '%s'", HttpPort)
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	startMeasuringHardwareStatus()

	// Graceful Shutdown
	waitForShutdown(srv)
}

/* handlers */

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	if !checkSameOrigin(r) {
		http.Error(w, "bad origin", http.StatusForbidden)
		util.LogSevere("got request from bad origin: '%s'", r.Header["Origin"])

		return
	}

	engine.WsEntry(w, r)
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

	err := util.GetUserSession(r, &session)

	util.LogInfo("ctrl page requested. UUID: '%s'", session)

	if err != nil {
		//create new session cookie
		sessionUUID, err := uuid.NewUUID()

		if err != nil {
			util.LogSevere("Failed to generate UUID: '%s'", err)

			renderCtrlPage(w, r, "Internal error")

			return
		}

		var session = util.HttpSession{
			SessionUUID: sessionUUID.String(),
			StartedAt:   time.Now().String(),
		}

		util.LogInfo("Created new session: '%s', startedAt: '%s'", session.SessionUUID, session.StartedAt)

		newSessionJson, err := json.Marshal(session)

		if err != nil {
			util.LogSevere("Failed to marshal JSON: '%s'", err)

			renderCtrlPage(w, r, "Internal error")

			return
		}

		encodedSession := base64.StdEncoding.EncodeToString(newSessionJson)

		http.SetCookie(w, &http.Cookie{
			Name:  "session",
			Value: encodedSession,
			Expires: time.Now().Add(365 * 24 * time.Hour),
		})
	}

	renderCtrlPage(w, r, "")
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
			Id:                     room.Id,
			Name:                   room.Name,
			StartedAt:              time.Unix(0, room.StartedAt).String(),
			ActiveRoomUsersNum:     room.ActiveRoomUsersLen,
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

	renderRoomsCtrlPage(w, r, &activeRooms, "")
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
		"activeRooms":         testutil.ToFloat64(engine.RoomsOnlineGauge),
		"activeUsers":         testutil.ToFloat64(engine.UsersOnlineGauge),
	}

	err := templates.CompiledTemplates.ExecuteTemplate(w, "tpl-ctrl.html", vars)

	if err != nil {
		log.Fatal("Cannot Get View ", err)
	}
}

func renderRoomsCtrlPage(w http.ResponseWriter, r *http.Request, activeRooms *[]domain_structures.RoomCtrlInfo, errorStr string) {
	vars := map[string]interface{}{
		"activeRooms": *activeRooms,
	}

	err := templates.CompiledTemplates.ExecuteTemplate(w, "tpl-rooms-ctrl.html", vars)

	if err != nil {
		log.Fatal("Cannot Get View ", err)
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
			http.Error(w, "Not authorized", 401)
			return
		}

		b, err := base64.StdEncoding.DecodeString(s[1])
		if err != nil {
			http.Error(w, err.Error(), 401)
			return
		}

		pair := strings.SplitN(string(b), ":", 2)
		if len(pair) != 2 {
			http.Error(w, "Not authorized", 401)
			return
		}

		if pair[0] != CtrlAuthLogin || pair[1] != CtrlAuthPasswd {
			http.Error(w, "Not authorized", 401)
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

func checkSameOrigin(r *http.Request) bool {
	origin := r.Header["Origin"]
	if len(origin) == 0 {
		return true
	}

	u, err := url.Parse(origin[0])
	if err != nil {
		return false
	}

	if util.EqualASCIIFold(u.Host, r.Host) {
		return true
	}

	defaultPort, ok := defaultPorts[u.Scheme]
	if !ok {
		return false
	}

	host, port, err := net.SplitHostPort(u.Host)
	if err == nil {
		return port == defaultPort && util.EqualASCIIFold(host, r.Host)
	}

	host, port, err = net.SplitHostPort(r.Host)
	if err == nil {
		return port == defaultPort && util.EqualASCIIFold(u.Host, host)
	}

	return false
}

func loadAppConfigs() {
	appConfigFile, _ := filepath.Abs("app-config.yml")
	yamlFile, err := ioutil.ReadFile(appConfigFile)
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

	CtrlAuthLogin = config.AppConfig.CtrlAuthLogin
	CtrlAuthPasswd = config.AppConfig.CtrlAuthPasswd

	log.Printf("app config: HttpPort='%s'", HttpPort)
	log.Printf("app config: HttpTimeout='%s'", HttpTimeout)
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
