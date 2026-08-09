package wiki_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Mireu-Lab/auto-docs/pkg/config"
	"github.com/Mireu-Lab/auto-docs/pkg/wiki"
)

func TestWikiDeployer_LocalFallback(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, err := os.Getwd()
	if err == nil {
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(oldWd) }()
	}

	cfg := &config.Config{
		WikiRepoURL: "",
	}
	deployer := wiki.NewWikiDeployer(cfg)
	_ = os.MkdirAll("wiki_output", 0755)

	reportContent := "# Test Report\n\nThis is a test wiki report."
	err = deployer.Deploy(reportContent)
	if err != nil {
		t.Fatalf("unexpected error deploying local wiki report: %v", err)
	}

	outputFile := filepath.Join("wiki_output", "Home.md")
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read local wiki output file: %v", err)
	}

	if string(content) != reportContent {
		t.Errorf("expected '%s', got '%s'", reportContent, string(content))
	}
}

func TestWikiDeployer_InvalidRepoURL(t *testing.T) {
	cfg := &config.Config{
		WikiRepoURL: "https://invalid-host-99999.example.invalid/repo.wiki.git",
		GitUsername: "bot",
		GitEmail:    "bot@example.com",
	}
	deployer := wiki.NewWikiDeployer(cfg)

	err := deployer.Deploy("# Report")
	if err == nil {
		t.Errorf("expected error when cloning invalid git wiki URL, got nil")
	}
}
