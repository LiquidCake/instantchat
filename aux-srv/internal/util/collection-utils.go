package util

func ArrayContains(source []string, val string) bool {
	for _, next := range source {
		if next == val {
			return true
		}
	}
	return false
}
