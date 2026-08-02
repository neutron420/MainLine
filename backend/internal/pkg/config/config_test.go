package config

import "testing"

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("DB_POOL_MIN", "")
	t.Setenv("DB_POOL_MAX", "")
	t.Setenv("RATE_LIMIT_PER_MINUTE", "")
	t.Setenv("STREAM_BUFFER_SIZE", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("LOG_FORMAT", "")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("REDIS_URL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Port != 50051 {
		t.Errorf("Port = %d, want 50051", cfg.Port)
	}
	if cfg.DBPoolMin != 2 || cfg.DBPoolMax != 20 {
		t.Errorf("pool = %d/%d, want 2/20", cfg.DBPoolMin, cfg.DBPoolMax)
	}
	if cfg.RateLimit != 100 {
		t.Errorf("RateLimit = %d, want 100", cfg.RateLimit)
	}
	if cfg.BufferSize != 100 {
		t.Errorf("BufferSize = %d, want 100", cfg.BufferSize)
	}
	if cfg.LogLevel != "info" || cfg.LogFormat != "json" {
		t.Errorf("log = %s/%s, want info/json", cfg.LogLevel, cfg.LogFormat)
	}
}

func TestLoad_Overrides(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("DB_POOL_MIN", "5")
	t.Setenv("DB_POOL_MAX", "50")
	t.Setenv("RATE_LIMIT_PER_MINUTE", "250")
	t.Setenv("STREAM_BUFFER_SIZE", "500")
	t.Setenv("DATABASE_URL", "postgres://localhost/app")
	t.Setenv("REDIS_URL", "redis://localhost:6379")
	t.Setenv("ENCRYPTION_MASTER_KEY", "0123456789abcdef0123456789abcdef")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("LOG_FORMAT", "text")
	t.Setenv("GITHUB_CLIENT_ID", "gh_id")
	t.Setenv("GOOGLE_CLIENT_ID", "g_id")
	t.Setenv("SLACK_CLIENT_ID", "s_id")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Port != 9090 {
		t.Errorf("Port = %d, want 9090", cfg.Port)
	}
	if cfg.DBPoolMin != 5 || cfg.DBPoolMax != 50 {
		t.Errorf("pool = %d/%d, want 5/50", cfg.DBPoolMin, cfg.DBPoolMax)
	}
	if cfg.RateLimit != 250 || cfg.BufferSize != 500 {
		t.Errorf("rate/buffer = %d/%d, want 250/500", cfg.RateLimit, cfg.BufferSize)
	}
	if cfg.DatabaseURL != "postgres://localhost/app" {
		t.Errorf("DatabaseURL = %q", cfg.DatabaseURL)
	}
	if cfg.RedisURL != "redis://localhost:6379" {
		t.Errorf("RedisURL = %q", cfg.RedisURL)
	}
	if cfg.EncryptionKey != "0123456789abcdef0123456789abcdef" {
		t.Errorf("EncryptionKey not loaded")
	}
	if cfg.LogLevel != "debug" || cfg.LogFormat != "text" {
		t.Errorf("log = %s/%s, want debug/text", cfg.LogLevel, cfg.LogFormat)
	}
	if cfg.GitHubClientID != "gh_id" || cfg.GoogleClientID != "g_id" || cfg.SlackClientID != "s_id" {
		t.Error("OAuth client IDs not loaded")
	}
}

func TestLoad_InvalidPort(t *testing.T) {
	t.Setenv("PORT", "not-a-number")
	if _, err := Load(); err == nil {
		t.Error("Load() with invalid PORT = nil error, want error")
	}
}

func TestGetEnv(t *testing.T) {
	t.Setenv("SCHEMAHUB_TEST_VAR", "set")
	if got := getEnv("SCHEMAHUB_TEST_VAR", "fallback"); got != "set" {
		t.Errorf("getEnv(set) = %q, want set", got)
	}
	if got := getEnv("SCHEMAHUB_UNSET_VAR", "fallback"); got != "fallback" {
		t.Errorf("getEnv(unset) = %q, want fallback", got)
	}
}
