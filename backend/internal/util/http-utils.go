package util

import (
	"encoding/json"
	"net/http"
	"net/url"
)

func GetRequestParamValue(r *http.Request, paramName string) string {
	urlParams, ok := r.URL.Query()[paramName]

	if ok && len(urlParams[0]) > 0 {
		return urlParams[0]
	}

	return ""
}

func GetUnescapedRequestParamValueUnsafe(r *http.Request, paramName string) string {
	//unsafe - error is not handled
	unescapedVal, _ := url.QueryUnescape(GetRequestParamValue(r, paramName))

	return unescapedVal
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
