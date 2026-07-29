package mcpserver

import (
	"context"
	"fmt"

	"github.com/Mireu-Lab/auto-docs/pkg/analyzer"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type AutoDocsMCPServer struct {
	server   *server.MCPServer
	analyzer *analyzer.CodeAnalyzer
}

func NewAutoDocsMCPServer(repoPath string) *AutoDocsMCPServer {
	codeAnalyzer := analyzer.NewCodeAnalyzer(repoPath)
	s := server.NewMCPServer(
		"auto-docs-mcp-server",
		"1.0.0",
		server.WithLogging(),
	)

	s.AddTool(
		mcp.NewTool("git_diff_analyzer",
			mcp.WithDescription("Get git diff summary and list of changed files from repository"),
			mcp.WithString("base_ref", mcp.Description("Git base ref (default: HEAD~1)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args, _ := req.Params.Arguments.(map[string]interface{})
			var baseRef string
			if args != nil {
				baseRef, _ = args["base_ref"].(string)
			}
			diffInfo, err := codeAnalyzer.GetGitDiff(ctx, baseRef)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Changed Files: %v\n\nDiff Summary:\n%s", diffInfo.ChangedFiles, diffInfo.DiffSummary)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("filesystem_reader",
			mcp.WithDescription("Read contents of source code files in repository"),
			mcp.WithString("files", mcp.Description("Comma-separated file paths relative to repository root")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args, _ := req.Params.Arguments.(map[string]interface{})
			var filesStr string
			if args != nil {
				filesStr, _ = args["files"].(string)
			}
			var fileList []string
			if filesStr != "" {
				fileList = append(fileList, filesStr)
			}
			snippets, err := codeAnalyzer.ReadSourceFiles(fileList)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			var res string
			for f, content := range snippets {
				res += fmt.Sprintf("=== File: %s ===\n%s\n\n", f, content)
			}
			return mcp.NewToolResultText(res), nil
		},
	)

	return &AutoDocsMCPServer{
		server:   s,
		analyzer: codeAnalyzer,
	}
}

func (s *AutoDocsMCPServer) ServeStdio() error {
	return server.ServeStdio(s.server)
}

func (s *AutoDocsMCPServer) GetServer() *server.MCPServer {
	return s.server
}
