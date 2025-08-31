package config

import "os"

type Config struct {
	Port       string
	PGDsn      string
	RedisAddr  string
	JWTSecret  string
	CORSOrigin string
}

func FromEnv() Config {
	return Config{
		Port:       getenv("APP_PORT", "8080"),
		PGDsn:      getenv("PG_DSN", "postgres://zodak:change_me_strong@127.0.0.1:5432/zodak?sslmode=disable"),
		RedisAddr:  getenv("REDIS_ADDR", "127.0.0.1:6379"),
		JWTSecret:  getenv("JWT_SECRET", "please_change_me"),
		CORSOrigin: getenv("CORS_ORIGIN", "https://zodak-test.ru"),
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

