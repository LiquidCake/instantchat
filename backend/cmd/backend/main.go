package main

import (
	"instantchat.rooms/instantchat/backend/internal/http_server"
	"instantchat.rooms/instantchat/backend/internal/config"
	"log"
)

var BuildVersion = "n/a"

func main() {
	log.Printf("Starting backend. App version: '%s'", BuildVersion)
	config.BuildVersion = BuildVersion

	http_server.StartServer()
}
