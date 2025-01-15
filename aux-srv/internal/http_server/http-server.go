package http_server

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
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

	"instantchat.rooms/instantchat/aux-srv/internal/config"
	"instantchat.rooms/instantchat/aux-srv/internal/load_balancing"
	"instantchat.rooms/instantchat/aux-srv/internal/templates"
	"instantchat.rooms/instantchat/aux-srv/internal/util"
)

/* Constants */

const RoomNameURLParam = "roomName"
const DirectMessageTextURLParam = "m"
const DirectMessagesRoomPasswordURLParam = "p"
const DirectMessagesLimitParam = "l"
const DirectMessagesIdParam = "id"
const DirectMessagesQuiteModeParam = "quite"
const DirectMessagesResponseFormatParam = "format"

const WinAppVersion = "1"

/* App configs */

var HttpTimeout = 30 * time.Second

var ShutdownWaitTimeout = 10 * time.Second

var LogMaxSizeMb = 500
var LogMaxFilesToKeep = 3
var LogMaxFileAgeDays = 60

var Domain = "n/a"
var HttpSchema = "https"
var CookiesIsSecure = true
var ClientAgreementVersion = "0"
var UserDrawingEnabled = true

var CtrlAuthLogin = "admin132"
var CtrlAuthPasswd = "password132"

var UnsecureTestMode = false

/* Variables */

// metrics
var homePageRequested prometheus.Counter
var roomPageRequested prometheus.Counter
var pickBackendRequested prometheus.Counter

var backendDirectCallClient *http.Client = nil

func initDirectCallHttpClient(unsecureTestMode bool) {
	var tlsConfig *tls.Config = nil

	if unsecureTestMode {
		log.Printf("WARNING: unsecure http client enabled")

		tlsConfig = &tls.Config{
			InsecureSkipVerify: true,
			MinVersion:         tls.VersionTLS10,
		}
	}

	backendDirectCallClient = &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: 30 * time.Second,
			}).DialContext,

			TLSHandshakeTimeout:   15 * time.Second,
			ExpectContinueTimeout: 15 * time.Second,
			ResponseHeaderTimeout: 15 * time.Second,
			IdleConnTimeout:       90 * time.Second,
			MaxConnsPerHost:       0,
			MaxIdleConns:          0,
			MaxIdleConnsPerHost:   20,
			TLSClientConfig:       tlsConfig,
		},
	}
}

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

	router.HandleFunc("/", middleware(renderHomePageHandler, loggingWrapper))
	router.HandleFunc("/about", middleware(renderAboutPageHandler, loggingWrapper))
	router.HandleFunc("/terms-of-service", middleware(renderTermsOfServiceHandler, loggingWrapper))
	router.HandleFunc("/privacy-policy", middleware(renderPrivacyPolicyHandler, loggingWrapper))
	router.HandleFunc("/connection-methods", middleware(renderConnectionMethodsPageHandler, loggingWrapper))
	router.HandleFunc("/universal-access", middleware(renderUniversalAccessPageHandler, loggingWrapper))
	router.HandleFunc("/app-win-version", middleware(returnAppWinVersionHandler, loggingWrapper))
	router.HandleFunc("/pick_backend", middleware(pickBackendForRoomHandler, loggingWrapper, noCacheWrapper))
	router.HandleFunc("/control_page_proxy", middleware(renderControlPageProxyHandler, basicAuthWrapper, loggingWrapper))
	router.HandleFunc("/r/{query_path:.*}", middleware(directlyRetrieveRoomMessagesHandler, loggingWrapper, noCacheWrapper))
	router.HandleFunc("/s/{query_path:.*}", middleware(directlySendRoomMessagesHandler, loggingWrapper, noCacheWrapper))
	router.HandleFunc("/{query_path:.*}", middleware(renderRoomPageHandler, loggingWrapper, noCacheWrapper))

	cert, err := tls.LoadX509KeyPair("/etc/ssl/ssl-bundle.crt", "/etc/ssl/cert.key")

	if err != nil {
		log.Fatal(err)
	}

	// HTTPS Server (main)
	srv := &http.Server{
		Handler: router,
		//Addr:         ":443",
		ReadTimeout:  HttpTimeout,
		WriteTimeout: HttpTimeout,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
	}

	go func() {
		util.LogInfo("Starting TLS Server")
		if err := srv.ListenAndServeTLS("", ""); err != nil {
			log.Fatal(err)
		}
	}()

	// HTTP Server (auxilary)
	httpSrv := &http.Server{
		Handler: router,
		//Addr:         ":80",
		ReadTimeout:  HttpTimeout,
		WriteTimeout: HttpTimeout,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
	}

	go func() {
		util.LogInfo("Starting HTTP Server")
		if err := httpSrv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	// Graceful Shutdown
	waitForShutdown(srv)
}

/* handlers */

func renderHomePageHandler(w http.ResponseWriter, r *http.Request) {
	homePageRequested.Inc()

	err := checkAndCreateUserSession(w, r)

	vars := map[string]interface{}{
		"error":      err,
		"domain":     Domain,
		"httpSchema": HttpSchema,
	}

	err = templates.CompiledHomeTemplate.Execute(w, vars)

	if err != nil {
		log.Fatal("Cannot Get View ", err)
	}
}

func renderAboutPageHandler(w http.ResponseWriter, r *http.Request) {
	vars := map[string]interface{}{
		"domain":     Domain,
		"httpSchema": HttpSchema,
	}

	err := templates.CompiledAboutTemplate.Execute(w, vars)

	if err != nil {
		log.Fatal("Cannot Get View ", err)
	}
}

func renderTermsOfServiceHandler(w http.ResponseWriter, r *http.Request) {
	vars := map[string]interface{}{
		"domain":     Domain,
		"httpSchema": HttpSchema,
	}

	err := templates.CompiledTermsOfServiceTemplate.Execute(w, vars)

	if err != nil {
		log.Fatal("Cannot Get View ", err)
	}
}

func renderPrivacyPolicyHandler(w http.ResponseWriter, r *http.Request) {
	vars := map[string]interface{}{
		"domain":     Domain,
		"httpSchema": HttpSchema,
	}

	err := templates.CompiledPrivacyPolicyTemplate.Execute(w, vars)

	if err != nil {
		log.Fatal("Cannot Get View ", err)
	}
}

func renderConnectionMethodsPageHandler(w http.ResponseWriter, r *http.Request) {
	vars := map[string]interface{}{
		"domain":     Domain,
		"httpSchema": HttpSchema,
	}

	err := templates.CompiledConnectionMethodsTemplate.Execute(w, vars)

	if err != nil {
		log.Fatal("Cannot Get View ", err)
	}
}

func renderUniversalAccessPageHandler(w http.ResponseWriter, r *http.Request) {
	vars := map[string]interface{}{
		"domain":     Domain,
		"httpSchema": HttpSchema,
	}

	err := templates.CompiledUniversalAccessTemplate.Execute(w, vars)

	if err != nil {
		log.Fatal("Cannot Get View ", err)
	}
}

func returnAppWinVersionHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/text")
	w.Write([]byte(WinAppVersion))
}

func renderRoomPageHandler(w http.ResponseWriter, r *http.Request) {
	roomPageRequested.Inc()

	err := checkAndCreateUserSession(w, r)

	requestedRoom := mux.Vars(r)["query_path"]

	if strings.Contains(requestedRoom, "/") {
		err = errors.New("room_to_home_pg_redirect_error_room_name_contains_slash")
	}

	vars := map[string]interface{}{
		"requestedRoom":          requestedRoom,
		"domain":                 Domain,
		"httpSchema":             HttpSchema,
		"clientAgreementVersion": ClientAgreementVersion,
		"userDrawingEnabled":     UserDrawingEnabled,
		"error":                  err,
	}

	err = templates.CompiledRoomTemplate.Execute(w, vars)

	if err != nil {
		log.Fatal("Cannot Get View ", err)
	}
}

func directlyRetrieveRoomMessagesHandler(w http.ResponseWriter, r *http.Request) {
	responseFormat := util.GetUnescapedParamValueUnsafe(r, DirectMessagesResponseFormatParam)
	responseFormat = strings.TrimSpace(responseFormat)

	responseTextBytes := directlyRetrieveRoomMessages(r, responseFormat)

	var contentType = "text/plain; charset=utf-8"

	if responseFormat == "json" {
		contentType = "application/json"
	}

	w.Header().Set("Content-Type", contentType)
	_, err := w.Write(responseTextBytes)

	if err != nil {
		util.LogWarn("Failed to write response for 'directly retrieve messages' request. err: '%s'", err)
	}
}

func directlyRetrieveRoomMessages(r *http.Request, responseFormat string) []byte {
	requestedRoom := strings.ToLower(mux.Vars(r)["query_path"])
	requestedRoom = strings.TrimSpace(requestedRoom)

	pickBackendRequested.Inc()

	validateRoomAndPickBackendResponse := load_balancing.ValidateRoomAndPickBackend(requestedRoom)
	pickBackendError := validateRoomAndPickBackendResponse.ErrorMessage

	if pickBackendError != "" {
		util.LogWarn("Room name validation error for 'directly retrieve messages': '%s'", pickBackendError)

		return util.BuildDirectRoomMessagesErrorResponse(fmt.Sprintf("error: %s", pickBackendError), responseFormat)
	}

	requestedRoom, _ = url.QueryUnescape(requestedRoom)

	pickBackendResponse := load_balancing.GetRoomBackend(requestedRoom)
	backendInstanceAddr := pickBackendResponse.BackendInstanceAddr

	if pickBackendResponse.BackendInstanceAddr == "" {
		util.LogWarn("Failed to pick backend instance for 'directly retrieve messages': '%s'", pickBackendResponse.ErrorMessage)

		return util.BuildDirectRoomMessagesErrorResponse("error: room not found", responseFormat)
	}

	roomPassword := util.GetUnescapedParamValueUnsafe(r, DirectMessagesRoomPasswordURLParam)
	roomPassword = strings.TrimSpace(roomPassword)

	messageLimit := util.GetUnescapedParamValueUnsafe(r, DirectMessagesLimitParam)
	messageLimit = strings.TrimSpace(messageLimit)

	messageId := util.GetUnescapedParamValueUnsafe(r, DirectMessagesIdParam)
	messageId = strings.TrimSpace(messageId)

	quiteMode := util.GetUnescapedParamValueUnsafe(r, DirectMessagesQuiteModeParam)
	quiteMode = strings.TrimSpace(quiteMode)

	requestURL := fmt.Sprintf("%s://%s/direct_retrieval?roomName=%s&p=%s&l=%s&id=%s&quite=%s&format=%s",
		config.AppConfig.BackendHttpSchema, backendInstanceAddr,
		url.QueryEscape(requestedRoom), url.QueryEscape(roomPassword),
		url.QueryEscape(messageLimit), url.QueryEscape(messageId), url.QueryEscape(quiteMode), url.QueryEscape(responseFormat))

	backendResponse, err := backendDirectCallClient.Get(requestURL)

	if err != nil {
		util.LogSevere("Failed to query backend '%s' room '%s' for 'directly retrieve messages': '%s'",
			backendInstanceAddr, requestedRoom, err)

		return util.BuildDirectRoomMessagesErrorResponse("error: failed to retrieve room messages - internal error", responseFormat)
	}

	defer backendResponse.Body.Close()

	if backendResponse.StatusCode != 200 {
		util.LogSevere("Got error from backend '%s' room '%s' for 'directly retrieve messages'. Status: '%d'",
			backendInstanceAddr, requestedRoom, backendResponse.StatusCode)

		return util.BuildDirectRoomMessagesErrorResponse("error: failed to retrieve room messages - internal error", responseFormat)
	}

	respBody, err := io.ReadAll(backendResponse.Body)

	if err != nil {
		util.LogSevere("Failed to read response from backend '%s' room '%s' for 'directly retrieve messages': '%s'",
			backendInstanceAddr, requestedRoom, err)

		return util.BuildDirectRoomMessagesErrorResponse("error: failed to retrieve room messages - internal error", responseFormat)
	}

	return respBody
}

func directlySendRoomMessagesHandler(w http.ResponseWriter, r *http.Request) {
	responseFormat := util.GetUnescapedParamValueUnsafe(r, DirectMessagesResponseFormatParam)
	responseFormat = strings.TrimSpace(responseFormat)

	responseTextBytes := directlySendRoomMessages(r, responseFormat)

	var contentType = "text/plain; charset=utf-8"

	if responseFormat == "json" {
		contentType = "application/json"
	}

	w.Header().Set("Content-Type", contentType)
	_, err := w.Write(responseTextBytes)

	if err != nil {
		util.LogWarn("Failed to write response for 'retrieve direct messages' request. err: '%s'", err)
	}
}

func directlySendRoomMessages(r *http.Request, responseFormat string) []byte {
	requestedRoom := strings.ToLower(mux.Vars(r)["query_path"])
	requestedRoom = strings.TrimSpace(requestedRoom)

	pickBackendRequested.Inc()

	validateRoomAndPickBackendResponse := load_balancing.ValidateRoomAndPickBackend(requestedRoom)
	pickBackendError := validateRoomAndPickBackendResponse.ErrorMessage

	if pickBackendError != "" {
		util.LogWarn("Room name validation error for 'directly send message': '%s'", pickBackendError)

		return util.BuildDirectRoomMessagesErrorResponse(fmt.Sprintf("error: %s", pickBackendError), responseFormat)
	}

	roomPassword := util.GetUnescapedParamValueUnsafe(r, DirectMessagesRoomPasswordURLParam)
	roomPassword = strings.TrimSpace(roomPassword)

	messageText := util.GetUnescapedParamValueUnsafe(r, DirectMessageTextURLParam)
	messageText = strings.TrimSpace(messageText)

	if len(strings.TrimSpace(messageText)) == 0 {
		return util.BuildDirectRoomMessagesErrorResponse("error: empty message (use URL param 'm=myMessage')", responseFormat)

	} else if len(messageText) >= util.MaxMessageLength {
		return util.BuildDirectRoomMessagesErrorResponse("error: message is too long", responseFormat)
	}

	requestedRoom, _ = url.QueryUnescape(requestedRoom)

	pickBackendResponse := load_balancing.GetRoomBackend(requestedRoom)
	backendInstanceAddr := pickBackendResponse.BackendInstanceAddr

	if pickBackendResponse.BackendInstanceAddr == "" {
		util.LogWarn("Failed to pick backend instance for 'directly send message': '%s'", pickBackendResponse.ErrorMessage)

		return util.BuildDirectRoomMessagesErrorResponse("error: room not found", responseFormat)
	}

	requestURL := fmt.Sprintf("%s://%s/direct_sending?roomName=%s&m=%s&p=%s&format=%s",
		config.AppConfig.BackendHttpSchema, backendInstanceAddr,
		url.QueryEscape(requestedRoom), url.QueryEscape(messageText), url.QueryEscape(roomPassword), url.QueryEscape(responseFormat))

	backendResponse, err := backendDirectCallClient.Get(requestURL)

	if err != nil {
		util.LogSevere("Failed to query backend '%s' room '%s' for 'directly send message': '%s'",
			backendInstanceAddr, requestedRoom, err)

		return util.BuildDirectRoomMessagesErrorResponse("error: failed to send room message - internal error", responseFormat)
	}

	defer backendResponse.Body.Close()

	if backendResponse.StatusCode != 200 {
		util.LogSevere("Got error from backend '%s' room '%s' for 'directly send message'. Status: '%d'",
			backendInstanceAddr, requestedRoom, backendResponse.StatusCode)

		return util.BuildDirectRoomMessagesErrorResponse("error: failed to send room message - internal error", responseFormat)
	}

	respBody, err := io.ReadAll(backendResponse.Body)

	if err != nil {
		util.LogSevere("Failed to read response from backend '%s' room '%s' for 'directly send message': '%s'",
			backendInstanceAddr, requestedRoom, err)

		return util.BuildDirectRoomMessagesErrorResponse("error: failed to send room message - internal error", responseFormat)
	}

	return respBody
}

func renderControlPageProxyHandler(w http.ResponseWriter, r *http.Request) {
	vars := map[string]interface{}{
		"backendInstances": config.AppConfig.BackendInstances,
	}

	err := templates.CompiledRoomCtrlPageProxyTemplate.Execute(w, vars)

	if err != nil {
		log.Fatal("Cannot Get View ", err)
	}
}

func pickBackendForRoomHandler(w http.ResponseWriter, r *http.Request) {
	pickBackendRequested.Inc()

	urlParams, ok := r.URL.Query()[RoomNameURLParam]

	if !ok || len(urlParams[0]) < 1 {
		util.LogSevere("'%s' param is missing from request", RoomNameURLParam)
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
			Name:     "session",
			Value:    encodedSession,
			Expires:  time.Now().Add(10 * 365 * 24 * time.Hour),
			HttpOnly: true,
			Secure:   CookiesIsSecure,
			Domain:   Domain,
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

	/* Populate configured vars */

	HttpTimeout = config.AppConfig.Server.HttpTimeoutSec * time.Second

	ShutdownWaitTimeout = config.AppConfig.ShutdownWaitTimeoutSec * time.Second

	LogMaxSizeMb = config.AppConfig.Logging.LogMaxSizeMb
	LogMaxFilesToKeep = config.AppConfig.Logging.LogMaxFilesToKeep
	LogMaxFileAgeDays = config.AppConfig.Logging.LogMaxFileAgeDays

	HttpSchema = config.AppConfig.MainHttpSchema
	CookiesIsSecure = config.AppConfig.Cookies.IsSecure

	Domain = config.AppConfig.Domain
	UserDrawingEnabled = config.AppConfig.UserDrawingEnabled
	ClientAgreementVersion = config.AppConfig.ClientAgreementVersion

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

	UnsecureTestMode = config.AppConfig.UnsecureTestMode

	log.Printf("app config: HttpTimeout='%s'", HttpTimeout)
	log.Printf("app config: ShutdownWaitTimeout='%s'", ShutdownWaitTimeout)
	log.Printf("app config: LogMaxSizeMb='%d'", LogMaxSizeMb)
	log.Printf("app config: LogMaxFilesToKeep='%d'", LogMaxFilesToKeep)
	log.Printf("app config: LogMaxFileAgeDays='%d'", LogMaxFileAgeDays)
	log.Printf("app config: HttpSchema='%s'", HttpSchema)
	log.Printf("app config: CookiesIsSecure='%t'", CookiesIsSecure)
	log.Printf("app config: Domain='%s'", Domain)
	log.Printf("app config: UserDrawingEnabled='%t'", UserDrawingEnabled)
	log.Printf("app config: ClientAgreementVersion='%s'", ClientAgreementVersion)
	log.Printf("app config: UnsecureTestMode='%t'", UnsecureTestMode)

	//init http clients
	initDirectCallHttpClient(UnsecureTestMode)
	load_balancing.InitBackendHwStatusHttpClient(UnsecureTestMode)
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
