package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL   string
	RedisURL      string
	JWTPrivateKey string
	JWTPublicKey  string
	EncryptionKey string
	Port          int
	LogLevel      string
	LogFormat     string
	DBPoolMin     int
	DBPoolMax     int
	RateLimit     int
	BufferSize    int

	GitHubClientID     string
	GitHubClientSecret string
	GitHubCallbackURL  string

	GoogleClientID      string
	GoogleCallbackURL   string
	FirebaseProjectID   string
	FirebasePrivateKey  string
	FirebaseClientEmail string

	SlackClientID     string
	SlackClientSecret string
	SlackCallbackURL  string
}

func Load() (*Config, error) {
	port, err := strconv.Atoi(getEnv("PORT", "50051"))
	if err != nil {
		return nil, fmt.Errorf("invalid PORT: %w", err)
	}
	poolMin, _ := strconv.Atoi(getEnv("DB_POOL_MIN", "2"))
	poolMax, _ := strconv.Atoi(getEnv("DB_POOL_MAX", "20"))
	rateLimit, _ := strconv.Atoi(getEnv("RATE_LIMIT_PER_MINUTE", "100"))
	bufferSize, _ := strconv.Atoi(getEnv("STREAM_BUFFER_SIZE", "100"))

	return &Config{
		DatabaseURL:   getEnv("DATABASE_URL", ""),
		RedisURL:      getEnv("REDIS_URL", ""),
		JWTPrivateKey: getEnv("JWT_PRIVATE_KEY", ""),
		JWTPublicKey:  getEnv("JWT_PUBLIC_KEY", ""),
		EncryptionKey: getEnv("ENCRYPTION_MASTER_KEY", ""),
		Port:          port,
		LogLevel:      getEnv("LOG_LEVEL", "info"),
		LogFormat:     getEnv("LOG_FORMAT", "json"),
		DBPoolMin:     poolMin,
		DBPoolMax:     poolMax,
		RateLimit:     rateLimit,
		BufferSize:    bufferSize,

		GitHubClientID:     getEnv("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
		GitHubCallbackURL:  getEnv("GITHUB_CALLBACK_URL", ""),

		GoogleClientID:      getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleCallbackURL:   getEnv("GOOGLE_CALLBACK_URL", ""),
		FirebaseProjectID:   getEnv("FIREBASE_PROJECT_ID", ""),
		FirebasePrivateKey:  getEnv("FIREBASE_PRIVATE_KEY", ""),
		FirebaseClientEmail: getEnv("FIREBASE_CLIENT_EMAIL", ""),

		SlackClientID:     getEnv("SLACK_CLIENT_ID", ""),
		SlackClientSecret: getEnv("SLACK_CLIENT_SECRET", ""),
		SlackCallbackURL:  getEnv("SLACK_CALLBACK_URL", ""),
	}, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
