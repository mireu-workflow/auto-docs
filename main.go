package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
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
	flag.Parse()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
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
		templateReport := fmt.Sprintf("# Auto-Docs Project Report\n\n> **최종 업데이트 일시**: %s (KST)\n> **대상 브랜치**: `%s`\n\n## Overview\n%s\n\n## Changed Files\n%v\n",
			opsExecutionTime.In(time.FixedZone("KST", 9*3600)).Format("2006-01-02 15:04:05 MST"),
			gitMeta.BranchName,
			overview,
			changedFiles,
		)
		_ = deployer.Deploy(templateReport)
		return
	}

	aiAgent := agent.NewGeminiClient(cfg)
	report, err := aiAgent.GenerateReportWithContext(ctx, overview, diffSummary, codeSnippets, webResults, previousWiki, gitMeta.BranchName, opsExecutionTime)
	if err != nil {
		log.Printf("⚠️ Warning: AI Agent report generation failed: %v", err)
		log.Println("Deploying default structured report template as fallback...")
		_ = os.MkdirAll("wiki_output", 0755)
		templateReport := fmt.Sprintf("# Auto-Docs Project Report\n\n> **최종 업데이트 일시**: %s (KST)\n> **대상 브랜치**: `%s`\n\n## Overview\n%s\n\n## Changed Files\n%v\n",
			opsExecutionTime.In(time.FixedZone("KST", 9*3600)).Format("2006-01-02 15:04:05 MST"),
			gitMeta.BranchName,
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
