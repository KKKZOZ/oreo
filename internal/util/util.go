package util

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/kkkzoz/oreo/pkg/config"
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
		log.Fatalf("ToJSONString error: %v\n", err)
	}
	return string(jsonString)
}

func ToInt(value any) int64 {
	switch v := value.(type) {
	case string:
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		return i

	case int:
		return int64(v)
	case int64:
		return v
	case float64:
		return int64(v)
	case float32:
		return int64(v)
	case uint:
		return int64(v)
	case uint32:
		return int64(v)
	case uint64:
		return int64(v)
	case byte:
		return int64(v)
	case rune:
		return int64(v)
	default:
		log.Fatalf("unsupported type: %T", value)
	}
	return 0
}

func ToString(value interface{}) string {
	switch v := value.(type) {
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', 6, 64)
	case float64:
		return strconv.FormatFloat(v, 'f', 6, 64)
	case bool:
		return strconv.FormatBool(v)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case byte:
		return strconv.Itoa(int(v))
	case rune:
		return string(v)
	case string:
		return v
	case config.State:
		return strconv.Itoa(int(v))
	default:
		log.Fatalf("ToString: Unsupported type: %T", v)
	}
	return ""
}
