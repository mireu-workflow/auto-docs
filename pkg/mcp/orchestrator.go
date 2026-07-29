package mcp

import (
	"context"
	"fmt"

	"github.com/Mireu-Lab/auto-docs/pkg/analyzer"
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

type MCPOrchestrator struct {
	tools map[string]MCPTool
}

func NewMCPOrchestrator(analyzer *analyzer.CodeAnalyzer) *MCPOrchestrator {
	o := &MCPOrchestrator{
		tools: make(map[string]MCPTool),
	}
	o.RegisterTool(&GitMCPTool{Analyzer: analyzer})
	o.RegisterTool(&FileSystemMCPTool{Analyzer: analyzer})
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
