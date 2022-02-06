package util

import (
	"fmt"
	"net/smtp"
	"time"
)

const FromEmailAddr = "some@gmail.com"
const FromEmailPasswd = "password"
const ToEmailAddr = "some@gmail.com"

const SMTPServerAddr = "smtp.gmail.com"
const SMTPServerPort = "587"

const BackendUnavailableReportingDelay = 5 * time.Minute

var backendUnavailableReportedAtTracking = make(map[string]int64)
var backendFullReportedAtTracking = make(map[string]int64)

func NotifyBackendUnavailable(backendInstance string) {
	lastReportedTimestamp, exists := backendUnavailableReportedAtTracking[backendInstance]

	if exists && time.Now().UnixNano()-lastReportedTimestamp < BackendUnavailableReportingDelay.Nanoseconds() {
		LogTrace("Backend '%s' unavailable. Skipping email report - already reported recently", backendInstance)

		return
	}

	now := time.Now()
	formattedTime := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d",
		now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second())

	err := send(fmt.Sprintf("[%s] Backend UNAVAILABLE: '%s'", formattedTime, backendInstance))

	if err != nil {
		LogSevere("Error sending email report (backend '%s' unavailable). Error: '%s'", backendInstance, err)
	} else {
		LogInfo("Sent email report - backend '%s' unavailable", backendInstance)

		backendUnavailableReportedAtTracking[backendInstance] = time.Now().UnixNano()
	}
}

func NotifyBackendIsFull(backendInstance string) {
	lastReportedTimestamp, exists := backendFullReportedAtTracking[backendInstance]

	if exists && time.Now().UnixNano()-lastReportedTimestamp < BackendUnavailableReportingDelay.Nanoseconds() {
		LogTrace("Backend '%s' full. Skipping email report - already reported recently", backendInstance)

		return
	}

	now := time.Now()
	formattedTime := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d",
		now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second())

	err := send(fmt.Sprintf("[%s] Backend is FULL: '%s'", formattedTime, backendInstance))

	if err != nil {
		LogSevere("Error sending email report (backend '%s' is full). Error: '%s'", backendInstance, err)
	} else {
		LogInfo("Sent email report - backend '%s' is full", backendInstance)

		backendFullReportedAtTracking[backendInstance] = time.Now().UnixNano()
	}
}

func send(body string) error {
	msg := "From: " + FromEmailAddr + "\n" +
		"To: " + ToEmailAddr + "\n" +
		"Subject: Backend unavailable\n\n" +
		body

	err := smtp.SendMail(
		SMTPServerAddr+":"+SMTPServerPort,
		smtp.PlainAuth("", FromEmailAddr, FromEmailPasswd, SMTPServerAddr),
		FromEmailAddr, []string{ToEmailAddr}, []byte(msg),
	)

	return err
}
