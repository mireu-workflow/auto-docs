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

	s.AddTool(
		mcp.NewTool("read_file",
			mcp.WithDescription("Read exact content and line numbers of a specific source code file in repository"),
			mcp.WithString("path", mcp.Required(), mcp.Description("File path relative to repository root")),
			mcp.WithNumber("start_line", mcp.Description("Starting line number (1-indexed, optional)")),
			mcp.WithNumber("end_line", mcp.Description("Ending line number (1-indexed, optional)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args, _ := req.Params.Arguments.(map[string]interface{})
			if args == nil {
				return mcp.NewToolResultError("missing arguments"), nil
			}
			path, _ := args["path"].(string)
			if path == "" {
				return mcp.NewToolResultError("path argument is required"), nil
			}
			var startLine, endLine int
			if sl, ok := args["start_line"].(float64); ok {
				startLine = int(sl)
			}
			if el, ok := args["end_line"].(float64); ok {
				endLine = int(el)
			}

			content, err := codeAnalyzer.ReadFileContent(path, startLine, endLine)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(content), nil
		},
	)

	s.AddTool(
		mcp.NewTool("code_search",
			mcp.WithDescription("Search codebase for keywords, function/type names, or regex patterns using git grep"),
			mcp.WithString("query", mcp.Required(), mcp.Description("Search keyword or regex query pattern")),
			mcp.WithBoolean("is_regex", mcp.Description("Whether to treat query as extended regex (default: false)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args, _ := req.Params.Arguments.(map[string]interface{})
			if args == nil {
				return mcp.NewToolResultError("missing arguments"), nil
			}
			query, _ := args["query"].(string)
			if query == "" {
				return mcp.NewToolResultError("query argument is required"), nil
			}
			isRegex, _ := args["is_regex"].(bool)

			matches, err := codeAnalyzer.SearchCode(ctx, query, isRegex)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(matches), nil
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
