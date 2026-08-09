package main

import (
	"flag"
	"os"
	"testing"
)

func TestMainFlags(t *testing.T) {
	// Test --mcp-server flag parsing
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"auto-docs", "-mcp-server=false"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	mcpServerMode := flag.Bool("mcp-server", false, "Run auto-docs as an MCP Server over Stdio")
	flag.Parse()

	if *mcpServerMode {
		t.Errorf("expected mcp-server flag to be false")
	}
}
