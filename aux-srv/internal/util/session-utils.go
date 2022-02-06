package util

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"unicode/utf8"
)

type HttpSession struct {
	SessionUUID string `json:"sessionUUID"`
	StartedAt   string `json:"startedAt"`
}

func GetCookieValue(name string, r *http.Request) (string, error) {
	sessionCookie, err := r.Cookie(name)

	if err != nil {
		return "", err
	} else {
		return sessionCookie.Value, nil
	}
}

func DecodeSessionToken(value string, session *HttpSession) error {
	data, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		LogSevere("error decoding token: '%s' ('%s')", err, value)

		return err
	}

	err = json.Unmarshal(data, session)
	if err != nil {
		LogSevere("error unmarshalling token: '%s'", err)

		return err
	}

	return nil
}

func GetUserSession(r *http.Request, session *HttpSession) error {
	sessionCookieValue, err := GetCookieValue("session", r)

	if err != nil {
		return err
	}

	// parse cookie data
	err = DecodeSessionToken(sessionCookieValue, session)

	if err != nil {
		return err
	}

	return nil
}

// returns true if s is equal to t with ASCII case folding as defined in RFC 4790
func EqualASCIIFold(s, t string) bool {
	for s != "" && t != "" {
		sr, size := utf8.DecodeRuneInString(s)
		s = s[size:]
		tr, size := utf8.DecodeRuneInString(t)
		t = t[size:]
		if sr == tr {
			continue
		}
		if 'A' <= sr && sr <= 'Z' {
			sr = sr + 'a' - 'A'
		}
		if 'A' <= tr && tr <= 'Z' {
			tr = tr + 'a' - 'A'
		}
		if sr != tr {
			return false
		}
	}
	return s == t
}
