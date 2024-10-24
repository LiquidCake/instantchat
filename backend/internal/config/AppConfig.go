package config

import "time"

var BuildVersion = "n/a"

type AppConfigList struct {
	ShutdownWaitTimeoutSec time.Duration `yaml:"shutdownWaitTimeoutSec"`
	ForbiddenRoomNames     []string      `yaml:"forbiddenRoomNames,flow"`
	AllowedOrigins         []string      `yaml:"allowedOrigins,flow"`

	Server struct {
		HttpPort       string        `yaml:"port"`
		HttpTimeoutSec time.Duration `yaml:"timeoutSec"`
	} `yaml:"http"`

	Logging struct {
		LogMaxSizeMb      int `yaml:"logMaxSizeMb"`
		LogMaxFilesToKeep int `yaml:"logMaxFilesToKeep"`
		LogMaxFileAgeDays int `yaml:"logMaxFileAgeDays"`
	} `yaml:"logging"`

	CtrlAuthLogin  string `yaml:"ctrlAuthLogin"`
	CtrlAuthPasswd string `yaml:"ctrlAuthPasswd"`

	Domain     string `yaml:"domain"`
	HttpSchema string `yaml:"httpSchema"`
}

var AppConfig AppConfigList
