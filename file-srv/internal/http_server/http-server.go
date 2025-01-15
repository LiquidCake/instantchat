package http_server

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v2"
	"instantchat.rooms/instantchat/file-srv/internal/file_storage"

	"instantchat.rooms/instantchat/file-srv/internal/config"
	"instantchat.rooms/instantchat/file-srv/internal/url_preview"
	"instantchat.rooms/instantchat/file-srv/internal/util"
)

/* App configs */

// Set in yaml app-config
var HttpPort = ":8085"
var HttpTimeout = 30 * time.Second

var ShutdownWaitTimeout = 10 * time.Second

var LogMaxSizeMb = 500
var LogMaxFilesToKeep = 3
var LogMaxFileAgeDays = 60

var TextFilesEnabled = true

/* Variables */

// metrics
var filesReceived prometheus.Counter
var filesRequested prometheus.Counter

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

	router.HandleFunc("/get_url_preview", middleware(getUrlPreview, loggingWrapper))
	router.HandleFunc("/get_text_file", middleware(getTextFile, loggingWrapper))
	router.HandleFunc("/upload_text_file", middleware(uploadTextFile, loggingWrapper))

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

	// Start Server
	go func() {
		util.LogInfo("Starting Server on '%s'", HttpPort)
		if err := srv.ListenAndServeTLS("", ""); err != nil {
			log.Fatal(err)
		}
	}()

	go url_preview.StartClearOldCacheItemsFuncPeriodical()
	go file_storage.StartClearOldCacheItemsFuncPeriodical()
	go file_storage.StartDeleteOldTextFilesFuncPeriodical()

	// Graceful Shutdown
	waitForShutdown(srv)
}

/* handlers */

func getUrlPreview(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	urlToPreview := strings.TrimSpace(r.Form.Get("url_to_preview"))

	if urlToPreview != "" {
		urlPreviewInfo, err := url_preview.GetUrlPreviewInfo(urlToPreview)

		if err != nil {
			util.LogTrace("error while getting url preview for '%s': '%s'", urlToPreview, err)

			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		jsonData, err := json.Marshal(urlPreviewInfo)

		if err != nil {
			util.LogSevere("Failed to serialize structure for 'get url preview' request. err: '%s'", err)

			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)

	} else {
		util.LogWarn("got empty 'url_to_preview' param in POST body (/get_url_preview)")

		w.WriteHeader(http.StatusBadRequest)
	}
}

func getTextFile(w http.ResponseWriter, r *http.Request) {
	if !TextFilesEnabled {
		util.LogWarn("got 'get text file' request but functionality is disabled")

		w.WriteHeader(http.StatusForbidden)
		return
	}

	fileNameParams, fileNameParamsOk := r.URL.Query()["file_name"]
	fileGroupPrefixParams, fileGroupPrefixParamsOk := r.URL.Query()["file_group_prefix"]

	if !fileNameParamsOk || !fileGroupPrefixParamsOk || len(fileNameParams[0]) < 1 || len(fileGroupPrefixParams[0]) < 1 {
		util.LogSevere("param(s) are missing from request (/get_text_file)")
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	fileName := fileNameParams[0]
	fileGroupPrefix := fileGroupPrefixParams[0]

	textFileBytes, err := file_storage.ReadTextFileFromDisk(fileName, fileGroupPrefix)

	if err != nil {
		util.LogSevere("error while loading text file from disk: '%s'", err)

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write(textFileBytes)
}

func uploadTextFile(w http.ResponseWriter, r *http.Request) {
	if !TextFilesEnabled {
		util.LogWarn("got 'upload text file' request but functionality is disabled")

		w.WriteHeader(http.StatusForbidden)
		return
	}

	r.ParseForm()
	fileContent := strings.TrimSpace(r.Form.Get("file_content"))
	fileName := strings.TrimSpace(r.Form.Get("file_name"))
	fileGroupPrefix := strings.TrimSpace(r.Form.Get("file_group_prefix"))

	if fileContent == "" || fileName == "" || fileGroupPrefix == "" {
		util.LogWarn("got empty param in POST body (/upload_text_file)")

		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := file_storage.SaveTextFileToDisk(fileContent, fileName, fileGroupPrefix)

	if err != nil {
		util.LogSevere("error while saving text file to disk: '%s'", err)

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
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

	TextFilesEnabled = config.AppConfig.TextFilesEnabled

	log.Printf("app config: HttpPort='%s'", HttpPort)
	log.Printf("app config: HttpTimeout='%s'", HttpTimeout)
	log.Printf("app config: ShutdownWaitTimeout='%s'", ShutdownWaitTimeout)
	log.Printf("app config: LogMaxSizeMb='%d'", LogMaxSizeMb)
	log.Printf("app config: LogMaxFilesToKeep='%d'", LogMaxFilesToKeep)
	log.Printf("app config: LogMaxFileAgeDays='%d'", LogMaxFileAgeDays)
	log.Printf("app config: TextFilesEnabled='%t'", TextFilesEnabled)
}

func setupMetrics() {
	filesReceived = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "files_received",
		})
	prometheus.MustRegister(filesReceived)

	filesRequested = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "files_requested",
		})
	prometheus.MustRegister(filesRequested)
}
