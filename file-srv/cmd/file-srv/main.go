package main

import (
	"log"

	"instantchat.rooms/instantchat/file-srv/internal/http_server"
)

var BuildVersion = "n/a"

func main() {
	log.Printf("Starting file-srv. App version: '%s'", BuildVersion)

	http_server.StartServer()
}
