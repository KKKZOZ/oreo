package util

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

// GetBodyString reads the response body from the provided http.Response
// and returns it as a string.
func GetBodyString(response *http.Response) string {
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)
	return bodyString
}

// ToJSONString converts the given value to a JSON string representation.
// It uses the json.Marshal function to serialize the value into JSON format.
// If an error occurs during the serialization process, it will log the error and exit the program.
// The resulting JSON string is returned as a string.
func ToJSONString(value any) string {
	jsonString, err := json.Marshal(value)
	if err != nil {
		log.Fatal(err)
	}
	return string(jsonString)
}
