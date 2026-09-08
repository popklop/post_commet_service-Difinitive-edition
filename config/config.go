package config

import (
	"os"
)

type Config struct {
	StorageType string
	PostgresDSN string
}

func Load() Config {
	storage := os.Getenv("STORAGE_TYPE")
	if storage == "" {
		storage = "memory"
	}

	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:pass@localhost:5432/postcommozon?sslmode=disable"
	}

	return Config{
		StorageType: storage,
		PostgresDSN: dsn,
	}
}
