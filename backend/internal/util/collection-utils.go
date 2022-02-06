package util

import (
	"instantchat.rooms/instantchat/backend/internal/domain_structures"
)

func MapKeySetsAreEqual(m1 *map[string]*domain_structures.WebSocket, m2 *map[string]*domain_structures.WebSocket) bool {
	if len(*m1) != len(*m2) {
		return false
	}

	for k, _ := range *m1 {
		if _, found := (*m2)[k]; !found {
			return false
		}
	}

	return true
}

func ArrayContainsString(source []string, val string) bool {
	for _, next := range source {
		if next == val {
			return true
		}
	}
	return false
}

func ArrayContainsInt(source []int, val int) bool {
	for _, next := range source {
		if next == val {
			return true
		}
	}
	return false
}
