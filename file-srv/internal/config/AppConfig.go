package config

import "time"

type AppConfigList struct {
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
}

var AppConfig AppConfigList
