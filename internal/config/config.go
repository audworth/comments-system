package config

import (
	"fmt"
	"os"
	"strconv"
)

const LocalEnv = "local"

type Config struct {
	Addr                 string
	LogLevel             string
	Env                  string
	QueryComplexityLimit int
}

func FromEnv() (Config, error) {
	config := Config{
		Addr:                 ":8080",
		LogLevel:             "info",
		Env:                  LocalEnv,
		QueryComplexityLimit: 10000,
	}

	if addr := os.Getenv("HTTP_ADDRESS"); addr != "" {
		config.Addr = addr
	}
	if logLVL := os.Getenv("LOG_LEVEL"); logLVL != "" {
		config.LogLevel = logLVL
	}
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		config.Env = env
	}
	if queryLimit := os.Getenv("QUERY_COMPLEXITY_LIMIT:"); queryLimit != "" {
		limit, err := strconv.Atoi(queryLimit)
		if err != nil {
			return Config{}, fmt.Errorf("parse QUERY_COMPLEXITY_LIMIT: %w", err)
		}
		if limit < 1 {
			return Config{}, fmt.Errorf("QUERY_COMPLEXITY_LIMIT must be positive")
		}
		config.QueryComplexityLimit = limit
	}

	return config, nil
}
