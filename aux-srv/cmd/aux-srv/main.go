package main

import (
	"log"

	"instantchat.rooms/instantchat/aux-srv/internal/http_server"
)

var BuildVersion = "n/a"

func main() {
	log.Printf("Starting aux-srv. App version: '%s'", BuildVersion)

	http_server.StartServer()
}
