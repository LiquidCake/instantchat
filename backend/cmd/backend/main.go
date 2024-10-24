package main

import (
	"log"

	"instantchat.rooms/instantchat/backend/internal/config"
	"instantchat.rooms/instantchat/backend/internal/http_server"
)

var BuildVersion = "n/a"

func main() {
	log.Printf("Starting backend. App version: '%s'", BuildVersion)
	config.BuildVersion = BuildVersion

	http_server.StartServer()
}
