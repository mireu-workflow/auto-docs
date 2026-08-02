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

func TestResolveWikiRepoURL(t *testing.T) {
	// Test 1: Replace gitlab-ci-token with PROJECT_ACCESS_TOKEN if present
	os.Setenv("PROJECT_ACCESS_TOKEN", "pat-secret-123")
	defer os.Unsetenv("PROJECT_ACCESS_TOKEN")

	rawURL := "https://gitlab-ci-token:[MASKED]@gitlab.com/cozwo/core-system/Auth-System.wiki.git"
	expected := "https://oauth2:pat-secret-123@gitlab.com/cozwo/core-system/Auth-System.wiki.git"
	resolved := config.ResolveWikiRepoURL(rawURL)

	if resolved != expected {
		t.Errorf("expected '%s', got '%s'", expected, resolved)
	}

	// Test 2: Auto-construct URL from GitLab CI env vars
	os.Setenv("CI_SERVER_HOST", "gitlab.example.com")
	os.Setenv("CI_PROJECT_PATH", "group/project")
	defer os.Unsetenv("CI_SERVER_HOST")
	defer os.Unsetenv("CI_PROJECT_PATH")

	expectedAuto := "https://oauth2:pat-secret-123@gitlab.example.com/group/project.wiki.git"
	resolvedAuto := config.ResolveWikiRepoURL("")

	if resolvedAuto != expectedAuto {
		t.Errorf("expected '%s', got '%s'", expectedAuto, resolvedAuto)
	}
}
