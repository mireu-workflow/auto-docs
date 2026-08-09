package mcp_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Mireu-Lab/auto-docs/pkg/analyzer"
	"github.com/Mireu-Lab/auto-docs/pkg/mcp"
)

func TestMCPOrchestrator(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	os.WriteFile(testFile, []byte("package main\n\nfunc Hello() string {\n\treturn \"world\"\n}\n"), 0644)

	codeAnalyzer := analyzer.NewCodeAnalyzer(tmpDir)
	orchestrator := mcp.NewMCPOrchestrator(codeAnalyzer)
	ctx := context.Background()

	// Test 1: ReadFile tool
	res, err := orchestrator.ExecuteTool(ctx, "read_file", map[string]interface{}{
		"path":       "test.go",
		"start_line": 1,
		"end_line":   5,
	})
	if err != nil {
		t.Fatalf("read_file failed: %v", err)
	}
	contentStr, ok := res.(string)
	if !ok || contentStr == "" {
		t.Errorf("expected non-empty string from read_file tool")
	}

	// Test 2: FileSystemReader tool with invalid args error
	_, err = orchestrator.ExecuteTool(ctx, "filesystem_reader", map[string]interface{}{
		"files": "not-a-slice",
	})
	if err == nil {
		t.Errorf("expected error for invalid files argument")
	}

	// Test 3: Unknown tool error
	_, err = orchestrator.ExecuteTool(ctx, "unknown_tool", nil)
	if err == nil {
		t.Errorf("expected error for unknown tool")
	}
}
