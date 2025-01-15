package config

import "time"

type AppConfigList struct {
	BackendInstances       []string      `yaml:"backendInstances,flow"`
	BackendHttpSchema      string        `yaml:"backendHttpSchema"`
	ForbiddenRoomNames     []string      `yaml:"forbiddenRoomNames,flow"`
	ShutdownWaitTimeoutSec time.Duration `yaml:"shutdownWaitTimeoutSec"`
	ClientAgreementVersion string        `yaml:"clientAgreementVersion"`
	UserDrawingEnabled     bool          `yaml:"userDrawingEnabled"`

	Server struct {
		HttpTimeoutSec time.Duration `yaml:"timeoutSec"`
	} `yaml:"http"`

	Logging struct {
		LogMaxSizeMb      int `yaml:"logMaxSizeMb"`
		LogMaxFilesToKeep int `yaml:"logMaxFilesToKeep"`
		LogMaxFileAgeDays int `yaml:"logMaxFileAgeDays"`
	} `yaml:"logging"`

	Cookies struct {
		IsSecure bool `yaml:"isSecure"`
	} `yaml:"cookies"`

	MainHttpSchema string `yaml:"mainHttpSchema"`

	Domain string `yaml:"domain"`

	CtrlAuthLogin  string `yaml:"ctrlAuthLogin"`
	CtrlAuthPasswd string `yaml:"ctrlAuthPasswd"`

	UnsecureTestMode bool `yaml:"unsecureTestMode"`
}

var AppConfig AppConfigList
