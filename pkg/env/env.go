package env

import (
	"fmt"
	"os"
)

func GetOrDefault(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

func GetOrDie(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	panic(fmt.Errorf("env for key=%s not found", key))
}
