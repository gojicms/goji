package utils

import (
	"os"
	"reflect"
	"strconv"
)

func GetEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func OrDefault[T any](value T, defaultValue T) T {
	switch v := any(value).(type) {
	case string:
		if v == "" {
			return defaultValue
		}
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		if reflect.ValueOf(v).IsZero() {
			return defaultValue
		}
	default:
		if reflect.ValueOf(value).IsZero() {
			return defaultValue
		}
	}
	return value
}

func Stoid(value string, defaultValue int) int {
	v, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return v
}

type Object = map[string]interface{}
