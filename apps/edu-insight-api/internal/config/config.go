package config

import "os"

type Config struct {
	AppEnv        string
	Port          string
	DBDSN         string
	MigrationsDir string
}

func Load() Config {
	return Config{
		AppEnv:        getEnv("APP_ENV", "development"),
		Port:          getEnv("PORT", "8088"),
		DBDSN:         getEnv("DB_DSN", ""),
		MigrationsDir: getEnv("MIGRATIONS_DIR", "migrations"),
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
