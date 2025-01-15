package util

import (
	"fmt"
	"net/url"
	"strings"
	"unicode"

	"instantchat.rooms/instantchat/aux-srv/internal/config"
)

const RoomCredsMinChars = 3
const RoomCredsMaxChars = 100

var allowedRoomNameSpecialChars = []string{
	"!",
	"@",
	"$",
	"*",
	"(",
	")",
	"_",
	"-",
	",",
	".",
	"~",
	"[",
	"]",
}

func ValidateRoomName(roomName string) string {
	roomNameTrimmed := strings.TrimSpace(roomName)

	if roomNameTrimmed == "" {
		return "empty room name"
	}

	roomNameDecoded, _ := url.QueryUnescape(roomNameTrimmed)

	if len([]rune(roomNameDecoded)) < RoomCredsMinChars || len([]rune(roomNameDecoded)) > RoomCredsMaxChars {
		return fmt.Sprintf("room name length must be between %d and %d", RoomCredsMinChars, RoomCredsMaxChars)
	}

	if ArrayContains(config.AppConfig.ForbiddenRoomNames, roomNameTrimmed) {
		return "this room name is forbidden"
	}

	allCharsAllowed := true

	for _, r := range roomNameTrimmed {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && !ArrayContains(allowedRoomNameSpecialChars, string(r)) {
			allCharsAllowed = false
			break
		}
	}

	if !allCharsAllowed {
		return "invalid room name: allowed only letters, numbers " +
			"and special characters: " + strings.Join(allowedRoomNameSpecialChars, " ")
	}

	return ""
}
