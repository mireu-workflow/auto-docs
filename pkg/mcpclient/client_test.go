package mcpclient_test

import (
	"testing"

	"github.com/Mireu-Lab/auto-docs/pkg/mcpclient"
)

func TestNewStdioMCPClient_InvalidCommand(t *testing.T) {
	// Attempt to start a non-existent binary to test error handling
	_, err := mcpclient.NewStdioMCPClient("non_existent_binary_12345")
	if err == nil {
		t.Errorf("expected error when launching non-existent binary, got nil")
	}
}
