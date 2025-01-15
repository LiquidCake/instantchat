package util

import (
	"crypto/tls"
	"fmt"
	"sync"
	"time"

	"github.com/go-mail/mail"
)

const FromEmailAddr = "myinstantchat.adm@example.com"
const FromEmailPasswd = "maypassword"
const ToEmailAddr = "myinstantchat.adm@example.com"

const SMTPServerAddr = "smtp.example.com"
const SMTPServerPort = "587"

const BackendReportingDelay = 3 * time.Minute

var commonMutex = sync.Mutex{} //should be ok to use one for all maps

var backendUnavailableReportedAtTracking = make(map[string]int64)
var backendFullReportedAtTracking = make(map[string]int64)
var backendCpuOverloadReportedAtTracking = make(map[string]int64)
var backendRamOverloadReportedAtTracking = make(map[string]int64)
var backendConcurrentUsersReportedAtTracking = make(map[string]int64)

func NotifyBackendUnavailable(backendInstance string) {
	commonMutex.Lock()
	lastReportedTimestamp, exists := backendUnavailableReportedAtTracking[backendInstance]

	if exists && time.Now().UnixNano()-lastReportedTimestamp < BackendReportingDelay.Nanoseconds() {
		commonMutex.Unlock()

		LogTrace("Backend '%s' unavailable. Skipping email report - already reported recently", backendInstance)

		return
	}

	backendUnavailableReportedAtTracking[backendInstance] = time.Now().UnixNano()
	commonMutex.Unlock()

	now := time.Now()
	formattedTime := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d",
		now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second())

	err := send(fmt.Sprintf("[%s] Backend UNAVAILABLE: '%s'", formattedTime, backendInstance))

	if err != nil {
		LogSevere("Error sending email report (backend '%s' unavailable). Error: '%s'", backendInstance, err)
	} else {
		LogInfo("Sent email report - backend '%s' unavailable", backendInstance)
	}
}

func NotifyBackendIsFull(backendInstance string) {
	commonMutex.Lock()
	lastReportedTimestamp, exists := backendFullReportedAtTracking[backendInstance]

	if exists && time.Now().UnixNano()-lastReportedTimestamp < BackendReportingDelay.Nanoseconds() {
		commonMutex.Unlock()

		LogTrace("Backend '%s' full. Skipping email report - already reported recently", backendInstance)

		return
	}

	backendFullReportedAtTracking[backendInstance] = time.Now().UnixNano()
	commonMutex.Unlock()

	now := time.Now()
	formattedTime := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d",
		now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second())

	err := send(fmt.Sprintf("[%s] Backend is FULL: '%s'", formattedTime, backendInstance))

	if err != nil {
		LogSevere("Error sending email report (backend '%s' is full). Error: '%s'", backendInstance, err)
	} else {
		LogInfo("Sent email report - backend '%s' is full", backendInstance)
	}
}

func NotifyBackendCpuOverload(backendInstance string) {
	commonMutex.Lock()
	lastReportedTimestamp, exists := backendCpuOverloadReportedAtTracking[backendInstance]

	if exists && time.Now().UnixNano()-lastReportedTimestamp < BackendReportingDelay.Nanoseconds() {
		commonMutex.Unlock()

		LogTrace("Backend '%s' has high recent CPU usage. Skipping email report - already reported recently", backendInstance)

		return
	}

	backendCpuOverloadReportedAtTracking[backendInstance] = time.Now().UnixNano()
	commonMutex.Unlock()

	now := time.Now()
	formattedTime := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d",
		now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second())

	err := send(fmt.Sprintf("[%s] Backend has high recent CPU usage: '%s'", formattedTime, backendInstance))

	if err != nil {
		LogSevere("Error sending email report (backend '%s' has high recent CPU usage). Error: '%s'", backendInstance, err)
	} else {
		LogInfo("Sent email report - backend '%s' has high recent CPU usage", backendInstance)
	}
}

func NotifyBackendRamOverload(backendInstance string) {
	commonMutex.Lock()
	lastReportedTimestamp, exists := backendRamOverloadReportedAtTracking[backendInstance]

	if exists && time.Now().UnixNano()-lastReportedTimestamp < BackendReportingDelay.Nanoseconds() {
		commonMutex.Unlock()

		LogTrace("Backend '%s' has high recent RAM usage. Skipping email report - already reported recently", backendInstance)

		return
	}

	backendRamOverloadReportedAtTracking[backendInstance] = time.Now().UnixNano()
	commonMutex.Unlock()

	now := time.Now()
	formattedTime := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d",
		now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second())

	err := send(fmt.Sprintf("[%s] Backend has high recent RAM usage: '%s'", formattedTime, backendInstance))

	if err != nil {
		LogSevere("Error sending email report (backend '%s' has high recent RAM usage). Error: '%s'", backendInstance, err)
	} else {
		LogInfo("Sent email report - backend '%s' has high recent RAM usage", backendInstance)
	}
}

func NotifyUsersCountReached(backendInstance string, usersCount int) {
	commonMutex.Lock()
	lastReportedTimestamp, exists := backendConcurrentUsersReportedAtTracking[backendInstance]

	if exists && time.Now().UnixNano()-lastReportedTimestamp < BackendReportingDelay.Nanoseconds() {
		commonMutex.Unlock()

		LogTrace("Backend '%s' reached '%d' concurrent users. Skipping email report - already reported recently",
			backendInstance, usersCount)

		return
	}

	backendConcurrentUsersReportedAtTracking[backendInstance] = time.Now().UnixNano()
	commonMutex.Unlock()

	now := time.Now()
	formattedTime := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d",
		now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second())

	err := send(fmt.Sprintf("['%s'] Backend reached '%d' concurrent users: '%s'",
		formattedTime, usersCount, backendInstance))

	if err != nil {
		LogSevere("Error sending email report (backend '%s' reached '%d' concurrent users). Error: '%s'",
			backendInstance, usersCount, err)
	} else {
		LogInfo("Sent email report - backend '%s' reached '%d' concurrent users", backendInstance, usersCount)
	}
}

func send(body string) error {
	dialer := mail.NewDialer(SMTPServerAddr, 587, FromEmailAddr, FromEmailPasswd)
	dialer.TLSConfig = &tls.Config{InsecureSkipVerify: true} //for some reason go+docker alpine image just wont recognize google's CA

	message := mail.NewMessage()
	message.SetHeader("From", FromEmailAddr)
	message.SetHeader("To", ToEmailAddr)
	message.SetHeader("Subject", "Backend issue")
	message.SetBody("text/plain", body)

	err := dialer.DialAndSend(message)

	return err
}
