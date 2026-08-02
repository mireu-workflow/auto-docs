package mcp

import (
	"context"
	"fmt"

	"github.com/Mireu-Lab/auto-docs/pkg/analyzer"
	"github.com/Mireu-Lab/auto-docs/pkg/searcher"
)

type MCPTool interface {
	Name() string
	Execute(ctx context.Context, args map[string]interface{}) (interface{}, error)
}

type GitMCPTool struct {
	Analyzer *analyzer.CodeAnalyzer
}

func (g *GitMCPTool) Name() string {
	return "git_diff_analyzer"
}

func (g *GitMCPTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	baseRef, _ := args["base_ref"].(string)
	return g.Analyzer.GetGitDiff(ctx, baseRef)
}

type FileSystemMCPTool struct {
	Analyzer *analyzer.CodeAnalyzer
}

func (f *FileSystemMCPTool) Name() string {
	return "filesystem_reader"
}

func (f *FileSystemMCPTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	files, ok := args["files"].([]string)
	if !ok {
		return nil, fmt.Errorf("invalid files argument")
	}
	return f.Analyzer.ReadSourceFiles(files)
}

type ReadFileMCPTool struct {
	Analyzer *analyzer.CodeAnalyzer
}

func (r *ReadFileMCPTool) Name() string {
	return "read_file"
}

func (r *ReadFileMCPTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	path, _ := args["path"].(string)
	if path == "" {
		return nil, fmt.Errorf("path argument is required")
	}
	var startLine, endLine int
	if sl, ok := args["start_line"].(int); ok {
		startLine = sl
	}
	if el, ok := args["end_line"].(int); ok {
		endLine = el
	}
	return r.Analyzer.ReadFileContent(path, startLine, endLine)
}

type CodeSearchMCPTool struct {
	Analyzer *analyzer.CodeAnalyzer
}

func (c *CodeSearchMCPTool) Name() string {
	return "code_search"
}

func (c *CodeSearchMCPTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	query, _ := args["query"].(string)
	if query == "" {
		return nil, fmt.Errorf("query argument is required")
	}
	isRegex, _ := args["is_regex"].(bool)
	return c.Analyzer.SearchCode(ctx, query, isRegex)
}

type DuckDuckGoSearchMCPTool struct {
	Searcher *searcher.DuckDuckGoSearcher
}

func (d *DuckDuckGoSearchMCPTool) Name() string {
	return "duckduckgo_search"
}

func (d *DuckDuckGoSearchMCPTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	query, _ := args["query"].(string)
	if query == "" {
		return nil, fmt.Errorf("query argument is required")
	}
	return d.Searcher.Search(ctx, query)
}

type MCPOrchestrator struct {
	tools map[string]MCPTool
}

func NewMCPOrchestrator(analyzer *analyzer.CodeAnalyzer) *MCPOrchestrator {
	o := &MCPOrchestrator{
		tools: make(map[string]MCPTool),
	}
	o.RegisterTool(&GitMCPTool{Analyzer: analyzer})
	o.RegisterTool(&FileSystemMCPTool{Analyzer: analyzer})
	o.RegisterTool(&ReadFileMCPTool{Analyzer: analyzer})
	o.RegisterTool(&CodeSearchMCPTool{Analyzer: analyzer})
	o.RegisterTool(&DuckDuckGoSearchMCPTool{Searcher: searcher.NewDuckDuckGoSearcher()})
	return o
}

func (o *MCPOrchestrator) RegisterTool(tool MCPTool) {
	o.tools[tool.Name()] = tool
}

func (o *MCPOrchestrator) ExecuteTool(ctx context.Context, toolName string, args map[string]interface{}) (interface{}, error) {
	tool, exists := o.tools[toolName]
	if !exists {
		return nil, fmt.Errorf("tool %s not registered in MCP orchestrator", toolName)
	}
	return tool.Execute(ctx, args)
}
