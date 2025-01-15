package util

import (
	"encoding/json"
	"net/http"
	"net/url"
)

func GetUnescapedParamValueUnsafe(r *http.Request, paramName string) string {
	urlParams, ok := r.URL.Query()[paramName]

	if ok && len(urlParams[0]) > 0 {
		//unsafe - error is not handled
		unescapedVal, _ := url.QueryUnescape(urlParams[0])

		return unescapedVal
	}

	return ""
}

func BuildDirectRoomMessagesErrorResponse(errorMessage string, responseFormat string) []byte {
	if responseFormat == "json" {
		responseJsonStr := map[string]string{
			"error": errorMessage,
		}

		jsonData, err := json.Marshal(responseJsonStr)

		if err != nil {
			LogSevere("Failed to serialize 'direct room messages error response'. err: '%s'", err)

			return []byte("{\"error\": \"failed to serialize json\"}")
		}

		return jsonData

	} else {
		return []byte(errorMessage)
	}
}
