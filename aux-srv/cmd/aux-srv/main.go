package main

import (
	"instantchat.rooms/instantchat/aux-srv/internal/http_server"
	"log"
)

var BuildVersion = "n/a"

var BuildEnv = http_server.DefaultEnvType

func main() {
	log.Printf("Starting aux-srv. Env: '%s', App version: '%s'", BuildEnv, BuildVersion)

	http_server.EnvType = BuildEnv

	http_server.StartServer()
}
