package config

import "os"

type Config struct {
	AppEnv string
	Port   string
	DBDSN  string
}

func Load() Config {
	return Config{
		AppEnv: getEnv("APP_ENV", "development"),
		Port:   getEnv("PORT", "8088"),
		DBDSN:  getEnv("DB_DSN", ""),
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
