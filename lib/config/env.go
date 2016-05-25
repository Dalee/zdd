package config

import (
	"os"
	"strings"
)

func GetEnvMap() map[string]string {
	env := make(map[string]string)
	envList := os.Environ()
	for _, envLine := range envList {
		split := strings.Split(envLine, "=")
		env[split[0]] = split[1]
	}

	return env
}

func GetEnv(key string) string {
	return os.Getenv(key)
}

func GetEnvDefault(key string, defaultValue string) string {
	val := GetEnv(key)
	if val == "" {
		val = defaultValue
	}

	return val
}