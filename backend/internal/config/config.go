// Package config loads runtime settings from environment variables.
package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// App modes (API.md §14).
const (
	ModeMock = "mock"
	ModeReal = "real"
)

// Config holds server settings.
type Config struct {
	Addr            string
	Mode            string
	MaxUploadBytes  int64
	AllowedOrigins  []string
	ShutdownTimeout time.Duration
}

// Load reads configuration from the environment, applying defaults that match
// the development environment described in API.md §2.1.
func Load() Config {
	return Config{
		Addr:            env("APP_ADDR", ":8080"),
		Mode:            normalizeMode(env("APP_MODE", ModeMock)),
		MaxUploadBytes:  envInt64("APP_MAX_UPLOAD_BYTES", 64<<20), // 64 MiB per request
		AllowedOrigins:  splitAndTrim(env("APP_ALLOWED_ORIGINS", "*")),
		ShutdownTimeout: 10 * time.Second,
	}
}

// IsMock reports whether the mock service should back the workflows.
func (c Config) IsMock() bool { return c.Mode == ModeMock }

func normalizeMode(v string) string {
	if strings.EqualFold(v, ModeReal) {
		return ModeReal
	}
	return ModeMock
}

func env(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func envInt64(key string, fallback int64) int64 {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

func splitAndTrim(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
