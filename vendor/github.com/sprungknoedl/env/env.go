// Package env is a simple configuration solution for building 12 factor apps that
// are configured via environment variables.
package env

import (
	"os"
	"strconv"
	"strings"
)

var defaults = map[string]interface{}{}

func normalize(key string) string {
	key = strings.ToUpper(key)
	key = strings.Replace(key, ".", "_", -1)
	key = strings.Replace(key, "-", "_", -1)
	return key
}

// SetDefault sets the default value for this key.
func SetDefault(key string, value interface{}) {
	key = normalize(key)
	defaults[key] = value
}

// GetString returns the value associated with the key as a string.
func GetString(key string) string {
	key = normalize(key)
	str := os.Getenv(key)
	if str != "" {
		return str
	}

	str, _ = defaults[key].(string)
	return str
}

// GetInt returns the value associated with the key as an integer.
func GetInt(key string) int {
	key = normalize(key)
	str := os.Getenv(key)
	num, err := strconv.Atoi(str)
	if str != "" && err == nil {
		return num
	}

	num, _ = defaults[key].(int)
	return num
}

// GetBool returns the value associated with the key as a boolean.
func GetBool(key string) bool {
	key = normalize(key)
	str := os.Getenv(key)
	boo, err := strconv.ParseBool(str)
	if str != "" && err == nil {
		return boo
	}

	boo, _ = defaults[key].(bool)
	return boo
}
