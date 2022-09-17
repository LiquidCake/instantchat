package http_server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v2"

  "instantchat.rooms/instantchat/aux-srv/internal/load_balancing"
	"instantchat.rooms/instantchat/aux-srv/internal/config"
	"instantchat.rooms/instantchat/aux-srv/internal/templates"
	"instantchat.rooms/instantchat/aux-srv/internal/util"
)

/* Constants */

const RoomNameParam = "roomName"

/* App configs */

// Set in yaml app-config
const DefaultEnvType = "n/a"
var EnvType = DefaultEnvType

var HttpPort = ":8080"
var HttpTimeout = 30 * time.Second

var ShutdownWaitTimeout = 10 * time.Second

var LogMaxSizeMb = 500
var LogMaxFilesToKeep = 3
var LogMaxFileAgeDays = 60

var Domain = "n/a"
var HttpSchema = "https"
var CookiesIsSecure = true

var CtrlAuthLogin = "admin132"
var CtrlAuthPasswd = "password132"

/* Variables */

//metrics
var homePageRequested prometheus.Counter
var roomPageRequested prometheus.Counter
var pickBackendRequested prometheus.Counter

func StartServer() {
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

	router.HandleFunc("/", middleware(renderHomePage, loggingWrapper))
	router.HandleFunc("/about", middleware(renderAboutPage, loggingWrapper))
	router.HandleFunc("/pick_backend", middleware(pickBackendForRoom, loggingWrapper, noCacheWrapper))
	router.HandleFunc("/control_page_proxy", middleware(renderControlPageProxy, basicAuthWrapper, loggingWrapper))
	router.HandleFunc("/{query_path:.*}", middleware(renderRoomPage, loggingWrapper, noCacheWrapper))

	srv := &http.Server{
		Handler:      router,
		Addr:         HttpPort,
		ReadTimeout:  HttpTimeout,
		WriteTimeout: HttpTimeout,
	}

	// Start Server
	go func() {
		util.LogInfo("Starting Server on '%s'", HttpPort)
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	// Graceful Shutdown
	waitForShutdown(srv)
}

/* handlers */

func renderHomePage(w http.ResponseWriter, r *http.Request) {
	homePageRequested.Inc()

	err := checkAndCreateUserSession(w, r)

	vars := map[string]interface{}{
		"error": err,
		"domain": Domain,
		"httpSchema": HttpSchema,
	}

	err = templates.CompiledTemplates.ExecuteTemplate(w, "tpl-home.html", vars)

	if err != nil {
		log.Fatal("Cannot Get View ", err)
	}
}

func renderAboutPage(w http.ResponseWriter, r *http.Request) {
	vars := map[string]interface{}{
		"domain": Domain,
		"httpSchema": HttpSchema,
	}

	err := templates.CompiledTemplates.ExecuteTemplate(w, "tpl-about.html", vars)

	if err != nil {
		log.Fatal("Cannot Get View ", err)
	}
}

func renderRoomPage(w http.ResponseWriter, r *http.Request) {
	roomPageRequested.Inc()

	err := checkAndCreateUserSession(w, r)

	requestedRoom := mux.Vars(r)["query_path"]

	if strings.Contains(requestedRoom, "/") {
		err = errors.New("room_to_home_pg_redirect_error_room_name_contains_slash")
	}

	vars := map[string]interface{}{
		"requestedRoom": requestedRoom,
		"domain": Domain,
		"httpSchema": HttpSchema,
		"error": err,
	}

	err = templates.CompiledTemplates.ExecuteTemplate(w, "tpl-room.html", vars)

	if err != nil {
		log.Fatal("Cannot Get View ", err)
	}
}

func renderControlPageProxy(w http.ResponseWriter, r *http.Request) {
	vars := map[string]interface{}{
		"backendInstances": config.AppConfig.BackendInstances,
	}

	err := templates.CompiledTemplates.ExecuteTemplate(w, "tpl-control-page-proxy.html", vars)

	if err != nil {
		log.Fatal("Cannot Get View ", err)
	}
}

func pickBackendForRoom(w http.ResponseWriter, r *http.Request) {
	pickBackendRequested.Inc()

	urlParams, ok := r.URL.Query()[RoomNameParam]

	if !ok || len(urlParams[0]) < 1 {
		util.LogSevere("'%s' param is missing from request", RoomNameParam)
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	roomName := urlParams[0]

	jsonData, err := json.Marshal(load_balancing.ValidateRoomAndPickBackend(roomName))

	if err != nil {
		util.LogSevere("Failed to serialize structure for 'pick backend' request. err: '%s'", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonData)

	if err != nil {
		util.LogWarn("Failed to write response for 'pick backend' request. err: '%s'", err)
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
func noCacheWrapper(h http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=0, no-cache, no-store, must-revalidate, private")

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

func checkAndCreateUserSession(w http.ResponseWriter, r *http.Request) error {
	var session util.HttpSession

	err := util.GetUserSession(r, &session)

	if err != nil {
		//create new session cookie
		sessionUUID, err := uuid.NewUUID()

		if err != nil {
			util.LogSevere("Failed to generate UUID: '%s'", err)

			return err
		}

		var session = util.HttpSession{
			SessionUUID: sessionUUID.String(),
			StartedAt:   time.Now().String(),
		}

		util.LogInfo("Created new session: '%s', startedAt: '%s'", session.SessionUUID, session.StartedAt)

		newSessionJson, err := json.Marshal(session)

		if err != nil {
			util.LogSevere("Failed to marshal JSON: '%s'", err)

			return err
		}

		encodedSession := base64.StdEncoding.EncodeToString(newSessionJson)

		http.SetCookie(w, &http.Cookie{
			Name:  "session",
			Value: encodedSession,
			Expires: time.Now().Add(365 * 24 * time.Hour),
			HttpOnly: true,
			Secure: CookiesIsSecure,
			Domain: Domain,
			SameSite: http.SameSiteStrictMode,
		})
	}

	return nil
}

func waitForShutdown(srv *http.Server) {
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive our signal.
	<-interruptChan

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), ShutdownWaitTimeout)
	defer cancel()

	srv.Shutdown(ctx)

	util.LogInfo("Shutting down")

	os.Exit(0)
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

	/* Populate configured vars */

	//if default env var value isn't overridden from OS env variable - set it from config
	if EnvType == DefaultEnvType {
		EnvType = config.AppConfig.EnvType
	}

	HttpPort = config.AppConfig.Server.HttpPort
	HttpTimeout = config.AppConfig.Server.HttpTimeoutSec * time.Second

	ShutdownWaitTimeout = config.AppConfig.ShutdownWaitTimeoutSec * time.Second

	LogMaxSizeMb = config.AppConfig.Logging.LogMaxSizeMb
	LogMaxFilesToKeep = config.AppConfig.Logging.LogMaxFilesToKeep
	LogMaxFileAgeDays = config.AppConfig.Logging.LogMaxFileAgeDays

	if EnvType == "prod" {
		HttpSchema = config.AppConfig.MainHttpSchemaProd
		CookiesIsSecure = config.AppConfig.CookiesProd.IsSecure
	} else {
		HttpSchema = config.AppConfig.MainHttpSchemaDev
		CookiesIsSecure = config.AppConfig.CookiesDev.IsSecure
	}

	Domain = config.AppConfig.Domain

	CtrlAuthLogin = config.AppConfig.CtrlAuthLogin
	CtrlAuthPasswd = config.AppConfig.CtrlAuthPasswd

	log.Printf("app config: EnvType='%s'", EnvType)
	log.Printf("app config: HttpPort='%s'", HttpPort)
	log.Printf("app config: HttpTimeout='%s'", HttpTimeout)
	log.Printf("app config: ShutdownWaitTimeout='%s'", ShutdownWaitTimeout)
	log.Printf("app config: LogMaxSizeMb='%d'", LogMaxSizeMb)
	log.Printf("app config: LogMaxFilesToKeep='%d'", LogMaxFilesToKeep)
	log.Printf("app config: LogMaxFileAgeDays='%d'", LogMaxFileAgeDays)
	log.Printf("app config: Domain='%s'", Domain)
	log.Printf("app config: HttpSchema='%s'", HttpSchema)
	log.Printf("app config: CookiesIsSecure='%t'", CookiesIsSecure)
}

func setupMetrics() {
	homePageRequested = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "home_page_requested",
		})
	prometheus.MustRegister(homePageRequested)

	roomPageRequested = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "room_page_requested",
		})
	prometheus.MustRegister(roomPageRequested)

	pickBackendRequested = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "pick_backend_requested",
		})
	prometheus.MustRegister(pickBackendRequested)
}
