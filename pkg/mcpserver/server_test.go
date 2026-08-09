package mcpserver_test

import (
	"testing"

	"github.com/Mireu-Lab/auto-docs/pkg/mcpserver"
)

func TestNewAutoDocsMCPServer(t *testing.T) {
	tmpDir := t.TempDir()
	server := mcpserver.NewAutoDocsMCPServer(tmpDir)
	if server == nil {
		t.Fatalf("expected non-nil AutoDocsMCPServer")
	}

	mcpServer := server.GetServer()
	if mcpServer == nil {
		t.Errorf("expected non-nil MCPServer from GetServer()")
	}
}
