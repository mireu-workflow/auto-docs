package analyzer_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Mireu-Lab/auto-docs/pkg/analyzer"
)

func TestReadFileContent(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")
	content := "line1\nline2\nline3\nline4\nline5"
	_ = os.WriteFile(filePath, []byte(content), 0644)

	ca := analyzer.NewCodeAnalyzer(tmpDir)

	// Test reading full file
	res, err := ca.ReadFileContent("test.txt", 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res, "line1") || !strings.Contains(res, "line5") {
		t.Errorf("expected all lines, got:\n%s", res)
	}

	// Test line range
	resRange, err := ca.ReadFileContent("test.txt", 2, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(resRange, "line1") || !strings.Contains(resRange, "line2") || !strings.Contains(resRange, "line4") {
		t.Errorf("expected lines 2..4, got:\n%s", resRange)
	}
}

func TestSearchCode(t *testing.T) {
	ca := analyzer.NewCodeAnalyzer(".")
	ctx := context.Background()

	// Search for 'CodeAnalyzer' symbol in this repo
	res, err := ca.SearchCode(ctx, "CodeAnalyzer", false)
	if err != nil {
		t.Fatalf("unexpected error searching code: %v", err)
	}
	if !strings.Contains(res, "analyzer.go") {
		t.Errorf("expected search result to contain analyzer.go, got:\n%s", res)
	}
}
