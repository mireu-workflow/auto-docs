package mcpclient

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

type MCPClientManager struct {
	mcpClient client.MCPClient
}

func NewStdioMCPClient(command string, args ...string) (*MCPClientManager, error) {
	cmd := exec.Command(command, args...)
	c, err := client.NewStdioMCPClient(cmd.Path, nil, cmd.Args[1:]...)
	if err != nil {
		return nil, fmt.Errorf("failed to launch MCP stdio client: %w", err)
	}

	return &MCPClientManager{mcpClient: c}, nil
}

func (m *MCPClientManager) Initialize(ctx context.Context) error {
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "auto-docs-client",
		Version: "1.0.0",
	}

	_, err := m.mcpClient.Initialize(ctx, initRequest)
	if err != nil {
		return fmt.Errorf("failed to initialize MCP connection: %w", err)
	}
	return nil
}

func (m *MCPClientManager) ListTools(ctx context.Context) ([]mcp.Tool, error) {
	res, err := m.mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return nil, err
	}
	return res.Tools, nil
}

func (m *MCPClientManager) CallTool(ctx context.Context, name string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	req := mcp.CallToolRequest{}
	req.Params.Name = name
	req.Params.Arguments = args
	return m.mcpClient.CallTool(ctx, req)
}

func (m *MCPClientManager) Close() error {
	return m.mcpClient.Close()
}
