package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
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

	GoogleClientID     string
	GoogleClientSecret string
	GoogleCallbackURL  string

	SlackClientID     string
	SlackClientSecret string
	SlackCallbackURL  string

	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string
	SMTPFromName string
	FrontendURL  string
}

// Validate ensures all required configuration values are present.
func (c *Config) Validate() error {
	var missing []string
	for _, f := range []struct {
		key string
		val string
	}{
		{"DATABASE_URL", c.DatabaseURL},
		{"REDIS_URL", c.RedisURL},
		{"JWT_PRIVATE_KEY", c.JWTPrivateKey},
		{"JWT_PUBLIC_KEY", c.JWTPublicKey},
		{"ENCRYPTION_MASTER_KEY", c.EncryptionKey},
	} {
		if f.val == "" {
			missing = append(missing, f.key)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}
	if len(c.EncryptionKey) < 32 {
		return fmt.Errorf("ENCRYPTION_MASTER_KEY must be at least 32 bytes")
	}
	return nil
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
	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))

	cfg := &Config{
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

		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleCallbackURL:  getEnv("GOOGLE_CALLBACK_URL", ""),

		SlackClientID:     getEnv("SLACK_CLIENT_ID", ""),
		SlackClientSecret: getEnv("SLACK_CLIENT_SECRET", ""),
		SlackCallbackURL:  getEnv("SLACK_CALLBACK_URL", ""),

		SMTPHost:     getEnv("SMTP_HOST", ""),
		SMTPPort:     smtpPort,
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:     getEnv("SMTP_FROM", ""),
		SMTPFromName: getEnv("SMTP_FROM_NAME", "SchemaHub"),
		FrontendURL:  getEnv("FRONTEND_URL", "http://localhost:3000"),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
