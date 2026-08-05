package config

import (
	"fmt"
	"os"

	"github.com/mpaverini/budget-back/internal/indicator"
)

// AuthMode values.
const (
	AuthModeFirebase = "firebase"
	AuthModeDev      = "dev"
)

type Config struct {
	Port                    string
	DatabaseURL             string
	FirebaseCredentialsFile string
	IPCSeriesID             string
	AuthMode                string
}

func Load() (Config, error) {
	cfg := Config{
		Port:                    getenv("PORT", "8080"),
		DatabaseURL:             os.Getenv("DATABASE_URL"),
		FirebaseCredentialsFile: os.Getenv("FIREBASE_CREDENTIALS_FILE"),
		IPCSeriesID:             getenv("INDICATOR_IPC_SERIES_ID", indicator.DefaultIPCSeriesID),
		AuthMode:                getenv("AUTH_MODE", AuthModeFirebase),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.AuthMode != AuthModeFirebase && cfg.AuthMode != AuthModeDev {
		return Config{}, fmt.Errorf("AUTH_MODE must be %q or %q, got %q", AuthModeFirebase, AuthModeDev, cfg.AuthMode)
	}
	if cfg.AuthMode == AuthModeFirebase && cfg.FirebaseCredentialsFile == "" {
		return Config{}, fmt.Errorf("FIREBASE_CREDENTIALS_FILE is required when AUTH_MODE=%s", AuthModeFirebase)
	}

	return cfg, nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
