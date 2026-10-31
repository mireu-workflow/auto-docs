package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Mireu-Lab/auto-docs/pkg/config"
	"github.com/Mireu-Lab/auto-docs/pkg/searcher"
)

type AIClient interface {
	GenerateReport(ctx context.Context, projectOverview string, diffSummary string, codeSnippets map[string]string, webSearchResults []searcher.SearchResult) (string, error)
	GenerateReportWithContext(ctx context.Context, projectOverview string, diffSummary string, codeSnippets map[string]string, webSearchResults []searcher.SearchResult, existingWiki string, branchName string, version string, updatedAt time.Time) (string, error)
}

type GeminiClient struct {
	cfg        *config.Config
	httpClient *http.Client
}

func NewGeminiClient(cfg *config.Config) *GeminiClient {
	return &GeminiClient{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 120 * time.Second},
	}
}

type geminiRequest struct {
	SystemInstruction *geminiContent  `json:"systemInstruction,omitempty"`
	Contents          []geminiContent `json:"contents"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text    string `json:"text"`
	Thought bool   `json:"thought,omitempty"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []geminiPart `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error,omitempty"`
}

func (g *GeminiClient) GenerateReport(ctx context.Context, projectOverview string, diffSummary string, codeSnippets map[string]string, webSearchResults []searcher.SearchResult) (string, error) {
	return g.GenerateReportWithContext(ctx, projectOverview, diffSummary, codeSnippets, webSearchResults, "", "", "", time.Now())
}

func (g *GeminiClient) GenerateReportWithContext(ctx context.Context, projectOverview string, diffSummary string, codeSnippets map[string]string, webSearchResults []searcher.SearchResult, existingWiki string, branchName string, version string, updatedAt time.Time) (string, error) {
	var promptBuilder strings.Builder
	promptBuilder.WriteString("## Goal & Persona\n")
	promptBuilder.WriteString(g.cfg.GoalPrompt)
	promptBuilder.WriteString("\n\n")

	promptBuilder.WriteString(fmt.Sprintf("## Target Language\nPlease write the entire documentation report in **%s** language.\n\n", g.cfg.DocsLanguage))

	promptBuilder.WriteString("You are an expert technical documentation writer acting under the above persona. Analyze the following source code, git diff, and DuckDuckGo live web search results, then generate a comprehensive, structured Markdown technical report suitable for a GitHub/GitLab Wiki.\n\n")
	
	promptBuilder.WriteString("## Project Overview\n")
	promptBuilder.WriteString(projectOverview)
	promptBuilder.WriteString("\n\n")

	if branchName != "" || version != "" || !updatedAt.IsZero() {
		promptBuilder.WriteString("## Execution & Ops Pipeline Metadata\n")
		if version != "" {
			promptBuilder.WriteString(fmt.Sprintf("- **Project / Release Version**: `%s`\n", version))
		}
		if branchName != "" {
			promptBuilder.WriteString(fmt.Sprintf("- **Target Git Branch**: `%s`\n", branchName))
		}
		if !updatedAt.IsZero() {
			promptBuilder.WriteString(fmt.Sprintf("- **Updated At (UTC)**: %s\n", updatedAt.UTC().Format("2006-01-02 15:04:05 MST")))
			promptBuilder.WriteString(fmt.Sprintf("- **Updated At (KST)**: %s\n", updatedAt.In(time.FixedZone("KST", 9*3600)).Format("2006-01-02 15:04:05 MST")))
		}
		promptBuilder.WriteString("\n")
	}

	if existingWiki != "" {
		promptBuilder.WriteString("## Previously Existing Wiki Documentation (Reference for Incremental Update)\n")
		promptBuilder.WriteString("> The following documentation was previously generated. You MUST review it to understand previously documented features, avoid removing valid content, reconcile similar or duplicate functionality, and perform a seamless incremental update.\n\n")
		promptBuilder.WriteString("```markdown\n")
		if len(existingWiki) > 8000 {
			promptBuilder.WriteString(existingWiki[:8000] + "\n...[truncated]...")
		} else {
			promptBuilder.WriteString(existingWiki)
		}
		promptBuilder.WriteString("\n```\n\n")
	}

	if len(webSearchResults) > 0 {
		promptBuilder.WriteString("## Live Web Search Context (DuckDuckGo Search Results)\n")
		for i, res := range webSearchResults {
			promptBuilder.WriteString(fmt.Sprintf("%d. [%s](%s)\n   Snippet: %s\n\n", i+1, res.Title, res.URL, res.Snippet))
		}
	}

	if diffSummary != "" {
		promptBuilder.WriteString("## Git Diff / Recent Changes\n```diff\n")
		// Limit diff size if too large
		if len(diffSummary) > 10000 {
			promptBuilder.WriteString(diffSummary[:10000] + "\n...[truncated]...")
		} else {
			promptBuilder.WriteString(diffSummary)
		}
		promptBuilder.WriteString("\n```\n\n")
	}

	promptBuilder.WriteString("## Key Source Code Snippets\n")
	for file, content := range codeSnippets {
		promptBuilder.WriteString(fmt.Sprintf("### File: `%s`\n```\n", file))
		if len(content) > 4000 {
			promptBuilder.WriteString(content[:4000] + "\n...[truncated]...")
		} else {
			promptBuilder.WriteString(content)
		}
		promptBuilder.WriteString("\n```\n\n")
	}

	promptBuilder.WriteString(fmt.Sprintf(`
---
### CRITICAL OUTPUT RULE:
- DO NOT output any reasoning steps, thinking notes, preambles, summaries of intent, or conversational wrapper text.
- Start IMMEDIATELY with the main Markdown title header ('# 🚀 ...') and output ONLY pure Markdown document content.
- Your entire response will be written directly into a GitHub/GitLab Wiki Markdown file.

### Required Wiki Documentation Structure:
You MUST follow the structured format below for the Wiki manual:

# 🚀 (Project Name) v(Version) ((Branch Name)) User Guide & Technical Specification

> **[Executive Summary]**: Summarize the core mission of the project and SLM / MCP integration purpose in 1-2 concise sentences.
> 
> * **Project Version**: '(Version)'
> * **Target Branch**: '(Branch Name)'
> * **Last Updated**: (Execution Timestamp UTC / KST)

## 🎯 1. Program Intent & Implementation Purpose (프로그램 의도 및 구현 목적)
- **Design Intent & Background Motivation (프로그램의 의도)**:
  - Why was this software created? What real-world developer friction, manual burden, or technical problem does it solve?
  - Background context and architectural motivation for introducing AI/SLM and MCP automation.
- **Implementation Objectives & Goals (구현 목적)**:
  - Specific engineering goals and milestones achieved by this system (e.g., zero-touch documentation, deterministic git tracking, standardized tool APIs).
  - Target audience (DevOps engineers, open-source contributors, team leads).

## 🚀 2. Implemented Features & Core Capabilities (구현 기능 명세)
- **Core Implemented Features (주요 구현 기능)**:
  - Feature 1: Comprehensive operational behavior, supported options, and data guarantees.
  - Feature 2: Architectural mechanisms and protocol support.
  - Feature 3: Error resilience, fallback mechanisms, and automation triggers.
- **Feature Matrix & Supported Platforms**:
  - Summary table comparing supported platforms (GitHub, GitLab, Local CLI, Stdio MCP).

## 🏗️ 3. System Architecture & Workflow (시스템 아키텍처 및 워크플로우)
- **Architecture Diagram**: Visualize data flow and module relationships using Mermaid ('graph TD' or 'sequenceDiagram') code block.
- **Core Components**: In-depth explanation of primary packages and operational principles based on the codebase.

## ⚡ 4. Practical Usage & Scenarios (활용 및 사용 방법)
### 4.1. Prerequisites & Environment Setup
- Runtime environment (e.g., Go 1.26+, Docker 24+), CLI tools, and required accounts/credentials.
### 4.2. Environment Variables & Configurations
- Markdown table format:
  | Variable | Required | Default | Description |
  | :--- | :---: | :--- | :--- |
### 4.3. Step-by-Step Usage Scenarios (활용 방법 시나리오)
- **Scenario A: Local CLI Execution**:
  - Command: 'go run main.go ...' or binary execution with flags ('-lang', '-mode').
- **Scenario B: Container & Docker Execution**:
  - Command: 'docker run ...' with volume mounts and environment injections.
- **Scenario C: Stdio MCP Server Integration**:
  - How to configure and consume tools in Claude Desktop, Antigravity, or other MCP hosts.
- **Scenario D: Automated CI/CD Ops Pipeline**:
  - GitHub Actions and GitLab CI automated Wiki deployment setups.

## 🔍 5. Module & API/MCP Specification (상세 모듈 및 명세)
- Detailed specification of packages, structs, major functions, CLI flags, and MCP tools defined in the codebase.
- Input/output data schemas and return patterns.

## 🔄 6. CI/CD & Automation Pipelines (CI/CD 파이프라인)
- **GitHub Actions (.github/workflows/)**: Trigger events, step-by-step build/test/deployment, and secret management (e.g., GEMINI_API_KEY, GITHUB_TOKEN).
- **GitLab CI (.gitlab-ci.yml)**: Pipeline stages (test, build-container, generate-docs, deploy) and runner operations.

## 🛠️ 7. Troubleshooting & FAQ (문제 해결 및 FAQ)
- Known error scenarios (e.g., Wiki Push HTTP 403, missing API keys, permission issues) and resolution tips (use GitHub Alerts > [!TIP] or > [!WARNING]).

---
### Additional Rules:
- Act strictly according to the defined goal & persona: "%s"
- Output Language: MUST be written entirely in '%s'.
- Provide response ONLY in Markdown Wiki format with clean headings, tables, mermaid diagrams, and code blocks.
`, g.cfg.GoalPrompt, g.cfg.DocsLanguage))

	reqBody := geminiRequest{
		SystemInstruction: &geminiContent{
			Parts: []geminiPart{
				{Text: "You are Gemma 4, an expert AI technical writer. Output ONLY pure Markdown Wiki documentation. NEVER output reasoning notes, planning bullet points, preambles, summaries of intent, or conversational wrapper text before or after the report. Start directly with '# 🚀'."},
			},
		},
		Contents: []geminiContent{
			{
				Parts: []geminiPart{
					{Text: promptBuilder.String()},
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		g.cfg.ModelName, g.cfg.GeminiAPIKey)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create http request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var gResp geminiResponse
	if err := json.Unmarshal(respBytes, &gResp); err != nil {
		return "", fmt.Errorf("failed to parse gemini response: %w", err)
	}

	if gResp.Error != nil {
		return "", fmt.Errorf("Gemini API error (%d): %s", gResp.Error.Code, gResp.Error.Message)
	}

	if len(gResp.Candidates) == 0 || len(gResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty candidate response from Gemini API")
	}

	// Find the non-thought part (Gemma 4 outputs reasoning with thought: true in earlier parts)
	var rawReport string
	parts := gResp.Candidates[0].Content.Parts
	for _, p := range parts {
		if !p.Thought && strings.TrimSpace(p.Text) != "" {
			rawReport += p.Text
		}
	}

	// Fallback to last part if all were marked thought
	if strings.TrimSpace(rawReport) == "" {
		rawReport = parts[len(parts)-1].Text
	}

	return cleanReportOutput(rawReport), nil
}

func cleanReportOutput(text string) string {
	text = strings.TrimSpace(text)

	// Remove markdown codeblock wrapper if the model accidentally wrapped the entire document in ```markdown ... ```
	if strings.HasPrefix(text, "```markdown") && strings.HasSuffix(text, "```") {
		text = strings.TrimPrefix(text, "```markdown")
		text = strings.TrimSuffix(text, "```")
		text = strings.TrimSpace(text)
	} else if strings.HasPrefix(text, "```") && strings.HasSuffix(text, "```") {
		text = strings.TrimPrefix(text, "```")
		text = strings.TrimSuffix(text, "```")
		text = strings.TrimSpace(text)
	}

	// Strip any preamble or reasoning notes before the first markdown header '# '
	if idx := strings.Index(text, "# "); idx > 0 {
		text = text[idx:]
	}

	return strings.TrimSpace(text)
}
