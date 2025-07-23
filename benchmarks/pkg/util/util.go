package util

import (
	"log"
	"math/rand"
	"strconv"

	"github.com/kkkzoz/oreo/pkg/config"
)

var letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// RandBytes fills the bytes with alphabetic characters randomly
func RandBytes(r *rand.Rand, b []byte) {
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
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
