package config

import (
	"errors"
	"fmt"
	"os"
)

const LocalEnv = "local"

type StorageType string

const (
	StoragePostgres StorageType = "postgres"
	StorageMemory   StorageType = "memory"
)

type Config struct {
	Addr     string
	LogLevel string
	Env      string
	Storage  StorageType
	DB       string
	RedisURL string
}

func FromEnv() (Config, error) {
	cfg := Config{
		Addr:     ":8080",
		LogLevel: "info",
		Env:      LocalEnv,
		Storage:  StoragePostgres,
		DB:       os.Getenv("DATABASE_URL"),
		RedisURL: os.Getenv("REDIS_URL"),
	}

	if cfg.RedisURL == "" {
		return Config{}, errors.New("REDIS_URL is required")
	}

	if addr := os.Getenv("HTTP_ADDRESS"); addr != "" {
		cfg.Addr = addr
	}
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		cfg.LogLevel = level
	}
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		cfg.Env = env
	}
	if storage := os.Getenv("STORAGE_TYPE"); storage != "" {
		cfg.Storage = StorageType(storage)
	}

	switch cfg.Storage {
	case StoragePostgres:
		if cfg.DB == "" {
			return Config{}, errors.New("DATABASE_URL is required for postgres")
		}
	case StorageMemory:
		panic("TODO")
	default:
		return Config{}, fmt.Errorf("unsupported STORAGE_TYPE %q", cfg.Storage)
	}

	return cfg, nil
}
