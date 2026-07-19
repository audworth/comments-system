package config

import (
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
	Addr        string
	LogLevel    string
	Env         string
	Storage     StorageType
	DatabaseURL string
}

func FromEnv() (Config, error) {
	cfg := Config{
		Addr:        ":8080",
		LogLevel:    "info",
		Env:         LocalEnv,
		Storage:     StoragePostgres,
		DatabaseURL: os.Getenv("DATABASE_URL"),
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
		if cfg.DatabaseURL == "" {
			return Config{}, fmt.Errorf("DATABASE_URL is required for %s storage", StoragePostgres)
		}
	case StorageMemory:
		panic("TODO")
	default:
		return Config{}, fmt.Errorf("unsupported STORAGE_TYPE %q", cfg.Storage)
	}

	return cfg, nil
}
