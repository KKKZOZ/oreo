package util

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-errors/errors"
	"github.com/oreo-dtx-lab/oreo/pkg/config"
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

// ToInt converts a value of any type to an int64.
// If the value is a string, it attempts to parse it as an int64 using strconv.ParseInt.
// If the value is an int, int64, float64, float32, uint, uint32, uint64, byte, or rune, it converts it to int64.
// If the value is of any other type, it logs a fatal error.
// Returns the converted int64 value.
func ToInt(value any) int64 {
	switch v := value.(type) {
	case string:
		if v == "" {
			return 0
		}
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

// ToString converts a value to its string representation.
// It supports conversion for various types including int, int64, float32, float64, bool, uint, uint32, uint64, byte, rune, string, []byte, and config.State.
// If the value is of an unsupported type, it will log a fatal error.
// The function returns the string representation of the value.
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
	case []byte:
		return string(v)
	case config.State:
		return strconv.Itoa(int(v))
	default:
		log.Fatalf("ToString: Unsupported type: %T", v)
	}
	return ""
}

func RetryHelper(maxRetryTimes int, retryInterval time.Duration, fn func() error) error {
	var err error
	for i := 0; i < maxRetryTimes; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		time.Sleep(retryInterval)
	}
	return errors.New("reached maximum retry limit, last error: " + err.Error())
}

func AddToString(s string, i int) string {
	num := ToInt(s) + ToInt(i)
	return ToString(num)
}

func FormatErrorStack(stackError *errors.Error) string {
	return strings.Replace(stackError.ErrorStack(), "\\n", "\n", -1)
}
