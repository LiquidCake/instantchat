package config

import "time"

type AppConfigList struct {
	EnvType                string        `yaml:"envType"`
	BackendInstances       []string      `yaml:"backendInstances,flow"`
	BackendHttpSchema      string        `yaml:"backendHttpSchema"`
	ForbiddenRoomNames     []string      `yaml:"forbiddenRoomNames,flow"`
	ShutdownWaitTimeoutSec time.Duration `yaml:"shutdownWaitTimeoutSec"`

	Server struct {
		HttpPort       string        `yaml:"port"`
		HttpTimeoutSec time.Duration `yaml:"timeoutSec"`
	} `yaml:"http"`

	Logging struct {
		LogMaxSizeMb      int `yaml:"logMaxSizeMb"`
		LogMaxFilesToKeep int `yaml:"logMaxFilesToKeep"`
		LogMaxFileAgeDays int `yaml:"logMaxFileAgeDays"`
	} `yaml:"logging"`

	CookiesProd struct {
		IsSecure bool `yaml:"isSecure"`
	} `yaml:"cookiesProd"`

	CookiesDev struct {
		IsSecure bool `yaml:"isSecure"`
	} `yaml:"cookiesDev"`

	Domain string `yaml:"domain"`

	MainHttpSchemaProd string `yaml:"mainHttpSchemaProd"`
	MainHttpSchemaDev  string `yaml:"mainHttpSchemaDev"`

	CtrlAuthLogin  string `yaml:"ctrlAuthLogin"`
	CtrlAuthPasswd string `yaml:"ctrlAuthPasswd"`
}

var AppConfig AppConfigList
