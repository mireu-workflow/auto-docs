package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Mireu-Lab/auto-docs/pkg/agent"
	"github.com/Mireu-Lab/auto-docs/pkg/analyzer"
	"github.com/Mireu-Lab/auto-docs/pkg/config"
	"github.com/Mireu-Lab/auto-docs/pkg/mcpserver"
	"github.com/Mireu-Lab/auto-docs/pkg/searcher"
	"github.com/Mireu-Lab/auto-docs/pkg/wiki"
)

func main() {
	mcpServerMode := flag.Bool("mcp-server", false, "Run auto-docs as an MCP Server over Stdio")
	langFlag := flag.String("lang", "", "Documentation output language (e.g. ko, en, ja, zh)")
	languageFlag := flag.String("language", "", "Alias for -lang")
	modeFlag := flag.String("mode", "", "Deployment fallback mode: 'A' (wiki-only) or 'B' (repo-docs push)")
	deployModeFlag := flag.String("deploy-mode", "", "Alias for -mode")
	flag.Parse()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Apply manual language override if specified via CLI flags
	if *langFlag != "" {
		cfg.DocsLanguage = *langFlag
	} else if *languageFlag != "" {
		cfg.DocsLanguage = *languageFlag
	}

	// Apply manual deployment mode override if specified
	if *modeFlag != "" {
		if strings.ToUpper(*modeFlag) == "B" || strings.ToLower(*modeFlag) == "repo-docs" {
			cfg.DeployMode = "B"
		} else {
			cfg.DeployMode = "A"
		}
	} else if *deployModeFlag != "" {
		if strings.ToUpper(*deployModeFlag) == "B" || strings.ToLower(*deployModeFlag) == "repo-docs" {
			cfg.DeployMode = "B"
		} else {
			cfg.DeployMode = "A"
		}
	}

	// 1. If invoked as MCP Server mode
	if *mcpServerMode {
		fmt.Fprintln(os.Stderr, "🚀 Running auto-docs MCP Server (Stdio mode)...")
		mcpServer := mcpserver.NewAutoDocsMCPServer(cfg.RepoPath)
		if err := mcpServer.ServeStdio(); err != nil {
			log.Fatalf("MCP Server error: %v", err)
		}
		return
	}

	// 2. Normal Auto-Docs Generator Mode
	fmt.Println("🚀 Starting auto-docs report generator with MCP Engine & DuckDuckGo Search...")

	codeAnalyzer := analyzer.NewCodeAnalyzer(cfg.RepoPath)
	ctx := context.Background()

	diffInfo, err := codeAnalyzer.GetGitDiff(ctx, "HEAD~1")
	if err != nil {
		log.Printf("Warning: failed to get git diff: %v", err)
	}

	var changedFiles []string
	var diffSummary string
	if diffInfo != nil {
		changedFiles = diffInfo.ChangedFiles
		diffSummary = diffInfo.DiffSummary
	}

	if len(changedFiles) > 15 {
		changedFiles = changedFiles[:15]
	}
	codeSnippets, _ := codeAnalyzer.ReadSourceFiles(changedFiles)
	overview, _ := codeAnalyzer.CollectProjectOverview()
	gitMeta := codeAnalyzer.GetGitMetadata(ctx)
	opsExecutionTime := time.Now()

	// 3. DuckDuckGo Web Search for live external documentation enrichment
	ddgSearcher := searcher.NewDuckDuckGoSearcher()
	var webResults []searcher.SearchResult
	searchQuery := "Model Context Protocol Go SDK technical documentation best practices"
	fmt.Printf("🔍 Performing DuckDuckGo search for live tech context: '%s'...\n", searchQuery)
	searchResults, searchErr := ddgSearcher.Search(ctx, searchQuery)
	if searchErr != nil {
		log.Printf("Warning: DuckDuckGo search failed: %v", searchErr)
	} else {
		webResults = searchResults
		fmt.Printf("🌐 DuckDuckGo search retrieved %d external reference results.\n", len(webResults))
	}

	// 4. Fetch Previous Wiki Documentation for incremental updates
	deployer := wiki.NewWikiDeployer(cfg)
	previousWiki, fetchErr := deployer.FetchExistingWikiContent()
	if fetchErr == nil && previousWiki != "" {
		fmt.Printf("📚 Found existing Wiki documentation (%d bytes). Reconciling previous features...\n", len(previousWiki))
	}

	// 5. Generate Report via Gemini / Gemma4 AI Agent
	if err := cfg.Validate(); err != nil {
		fmt.Printf("⚠️  Skipping AI API call: %v\n", err)
		fmt.Println("Writing default structured report template...")
		
		_ = os.MkdirAll("wiki_output", 0755)
		templateReport := fmt.Sprintf("# Auto-Docs Project Report v%s (%s)\n\n> **Project Version**: `%s`\n> **Target Branch**: `%s`\n> **Last Updated**: %s (UTC) / %s (KST)\n\n## 🎯 1. Program Intent & Implementation Purpose\n- **Design Intent**: Automated AI-driven documentation engine to eliminate manual documentation drift.\n- **Implementation Purpose**: Automatically analyze codebase and git changes to publish user manuals and specs.\n\n## 🚀 2. Implemented Features\n- Dual Mode (CLI Pipeline & Stdio MCP Server)\n- Incremental Wiki context reconciliation\n- Deploy Mode A/B fallback support\n\n## ⚡ 3. Practical Usage\n- Run CLI: `go run main.go -lang en`\n- Run MCP: `go run main.go --mcp-server`\n\n## Overview\n%s\n\n## Changed Files\n%v\n",
			gitMeta.Version,
			gitMeta.BranchName,
			gitMeta.Version,
			gitMeta.BranchName,
			opsExecutionTime.UTC().Format("2006-01-02 15:04:05 MST"),
			opsExecutionTime.In(time.FixedZone("KST", 9*3600)).Format("2006-01-02 15:04:05 MST"),
			overview,
			changedFiles,
		)
		_ = deployer.Deploy(templateReport)
		return
	}

	aiAgent := agent.NewGeminiClient(cfg)
	report, err := aiAgent.GenerateReportWithContext(ctx, overview, diffSummary, codeSnippets, webResults, previousWiki, gitMeta.BranchName, gitMeta.Version, opsExecutionTime)
	if err != nil {
		log.Printf("⚠️ Warning: AI Agent report generation failed: %v", err)
		log.Println("Deploying default structured report template as fallback...")
		_ = os.MkdirAll("wiki_output", 0755)
		templateReport := fmt.Sprintf("# Auto-Docs Project Report v%s (%s)\n\n> **Project Version**: `%s`\n> **Target Branch**: `%s`\n> **Last Updated**: %s (UTC) / %s (KST)\n\n## 🎯 1. Program Intent & Implementation Purpose\n- **Design Intent**: Automated AI-driven documentation engine to eliminate manual documentation drift.\n- **Implementation Purpose**: Automatically analyze codebase and git changes to publish user manuals and specs.\n\n## 🚀 2. Implemented Features\n- Dual Mode (CLI Pipeline & Stdio MCP Server)\n- Incremental Wiki context reconciliation\n- Deploy Mode A/B fallback support\n\n## ⚡ 3. Practical Usage\n- Run CLI: `go run main.go -lang en`\n- Run MCP: `go run main.go --mcp-server`\n\n## Overview\n%s\n\n## Changed Files\n%v\n",
			gitMeta.Version,
			gitMeta.BranchName,
			gitMeta.Version,
			gitMeta.BranchName,
			opsExecutionTime.UTC().Format("2006-01-02 15:04:05 MST"),
			opsExecutionTime.In(time.FixedZone("KST", 9*3600)).Format("2006-01-02 15:04:05 MST"),
			overview,
			changedFiles,
		)
		if err := deployer.Deploy(templateReport); err != nil {
			log.Fatalf("Failed to deploy wiki report: %v", err)
		}
		fmt.Println("✅ auto-docs processing completed with fallback report!")
		return
	}

	// 6. Deploy Report to Wiki
	if err := deployer.Deploy(report); err != nil {
		log.Fatalf("Failed to deploy wiki report: %v", err)
	}

	fmt.Println("✅ auto-docs processing completed successfully!")
}
