package util

import (
	"log"
)

var CurrentLogLevel LogLevel

type LogLevel int

const (
	Trace  LogLevel = 0
	Info   LogLevel = 1
	Warn   LogLevel = 2
	Severe LogLevel = 3
)

func LogTrace(format string, args ...interface{}) {
	if CurrentLogLevel <= Trace {
		log.Printf("[TRACE] "+format, args...)
	}
}

func LogInfo(format string, args ...interface{}) {
	if CurrentLogLevel <= Info {
		log.Printf("[INFO] "+format, args...)
	}
}

func LogWarn(format string, args ...interface{}) {
	if CurrentLogLevel <= Warn {
		log.Printf("[WARN] "+format, args...)
	}
}

func LogSevere(format string, args ...interface{}) {
	if CurrentLogLevel <= Severe {
		log.Printf("[SEVERE] "+format, args...)
	}
}
