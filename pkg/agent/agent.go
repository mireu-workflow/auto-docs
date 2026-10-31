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
	GenerateReportWithContext(ctx context.Context, projectOverview string, diffSummary string, codeSnippets map[string]string, webSearchResults []searcher.SearchResult, existingWiki string, branchName string, updatedAt time.Time) (string, error)
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
	Text string `json:"text"`
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
	return g.GenerateReportWithContext(ctx, projectOverview, diffSummary, codeSnippets, webSearchResults, "", "", time.Now())
}

func (g *GeminiClient) GenerateReportWithContext(ctx context.Context, projectOverview string, diffSummary string, codeSnippets map[string]string, webSearchResults []searcher.SearchResult, existingWiki string, branchName string, updatedAt time.Time) (string, error) {
	var promptBuilder strings.Builder
	promptBuilder.WriteString("## Goal & Persona\n")
	promptBuilder.WriteString(g.cfg.GoalPrompt)
	promptBuilder.WriteString("\n\n")

	promptBuilder.WriteString(fmt.Sprintf("## Target Language\nPlease write the entire documentation report in **%s** language.\n\n", g.cfg.DocsLanguage))

	promptBuilder.WriteString("You are an expert technical documentation writer acting under the above persona. Analyze the following source code, git diff, and DuckDuckGo live web search results, then generate a comprehensive, structured Markdown technical report suitable for a GitHub/GitLab Wiki.\n\n")
	
	promptBuilder.WriteString("## Project Overview\n")
	promptBuilder.WriteString(projectOverview)
	promptBuilder.WriteString("\n\n")

	if branchName != "" || !updatedAt.IsZero() {
		promptBuilder.WriteString("## Execution & Ops Pipeline Metadata\n")
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

### Required Wiki Documentation Structure (위키 설명서 필수 구조):
You MUST follow the structured format below for the Wiki manual:

# 🚀 (Project Name) 사용설명서 & 시스템 명세서

> **[개요 요약]**: 프로젝트의 핵심 역할과 SLM / MCP 연동 목적을 1~2문장으로 요약하여 작성합니다.
> 
> * **최종 업데이트 일시 (Last Updated)**: (Execution Timestamp KST / UTC)
> * **대상 브랜치 (Git Branch)**: '(Branch Name)'

## 📖 1. 개요 및 주요 기능 (Overview & Key Features)
- **프로젝트 개요**: 시스템의 핵심 목표 및 주요 활용 시나리오.
- **주요 기능 (Key Features)**:
  - 기능 1: 상세 설명
  - 기능 2: 상세 설명
  - 기능 3: 상세 설명

## 🏗️ 2. 시스템 아키텍처 & 워크플로우 (Architecture & Workflow)
- **아키텍처 구조도**: mermaid ('graph TD' 또는 'sequenceDiagram') 코드 블록으로 데이터 흐름과 모듈 간 관계를 시각화합니다.
- **핵심 컴포넌트**: 소스코드 기반 주요 구성 요소 및 동작 원리 설명.

## ⚡ 3. 빠른 시작 가이드 (Quick Start Guide)
### 3.1. 사전 요구사항 (Prerequisites)
- 런타임 환경 (Go 1.26+, Docker 24+), CLI 도구 및 필요 계정/권한.
### 3.2. 환경 변수 및 설정 (Environment Variables)
- 마크다운 표 형식:
  | 환경 변수명 | 필수 여부 | 기본값 | 설명 |
  | :--- | :---: | :--- | :--- |
### 3.3. 실행 방법 (Usage Commands)
- **CLI 실행**: 'go run main.go ...' 또는 바이너리 실행 커맨드.
- **Docker 실행**: 'docker run ...' 명령 커맨드.
- **MCP Server 실행 (지원 시)**: MCP stdio 실행 커맨드.

## 🔍 4. 상세 모듈 및 API/MCP 명세 (Module & API Specification)
- 소스코드에 정의된 주요 패키지, 구조체(Struct), 주요 함수(Function), CLI 플래그(Flags) 및 MCP 도구(Tools) 상세 명세.
- 입력(Input)/출력(Output) 데이터 형식 및 반환 패턴.

## 🔄 5. CI/CD & 자동화 파이프라인 (CI/CD Pipeline)
- **GitHub Actions (.github/workflows/)**: 트리거 조건, 단계별 빌드/테스트/배포 및 Secret 설정 (GEMINI_API_KEY, WIKI_REPO_URL, GCP_SA_KEY).
- **GitLab CI (.gitlab-ci.yml)**: 파이프라인 스테이지 (test, build-container, generate-docs, deploy) 구성 및 Runner 동작 원리.

## 🛠️ 6. 문제 해결 및 FAQ (Troubleshooting & FAQ)
- 실행 중 발생 가능한 예외 사례 (예: Wiki Push HTTP 403, API Key 미설정, GCP 권한 오차) 및 해결 방법 (GitHub Alert 상자 > [!TIP] 또는 > [!WARNING] 활용).

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

	rawReport := gResp.Candidates[0].Content.Parts[0].Text
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
