package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

func getBodyString(response *http.Response) string {
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)
	return bodyString
}

func toJSONString(value any) string {
	jsonString, err := json.Marshal(value)
	if err != nil {
		log.Fatal(err)
	}
	return string(jsonString)
}
