package config_test

import (
	"os"
	"testing"

	"github.com/Mireu-Lab/auto-docs/pkg/config"
)

func TestLoadConfig(t *testing.T) {
	os.Setenv("GEMINI_API_KEY", "test-api-key")
	defer os.Unsetenv("GEMINI_API_KEY")

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.GeminiAPIKey != "test-api-key" {
		t.Errorf("expected GeminiAPIKey 'test-api-key', got '%s'", cfg.GeminiAPIKey)
	}

	if err := cfg.Validate(); err != nil {
		t.Errorf("config validation failed: %v", err)
	}
}
