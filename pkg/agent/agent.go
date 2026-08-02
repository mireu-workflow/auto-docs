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
)

type AIClient interface {
	GenerateReport(ctx context.Context, projectOverview string, diffSummary string, codeSnippets map[string]string) (string, error)
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

func (g *GeminiClient) GenerateReport(ctx context.Context, projectOverview string, diffSummary string, codeSnippets map[string]string) (string, error) {
	var promptBuilder strings.Builder
	promptBuilder.WriteString("## Goal & Persona\n")
	promptBuilder.WriteString(g.cfg.GoalPrompt)
	promptBuilder.WriteString("\n\n")

	promptBuilder.WriteString(fmt.Sprintf("## Target Language\nPlease write the entire documentation report in **%s** language.\n\n", g.cfg.DocsLanguage))

	promptBuilder.WriteString("You are an expert technical documentation writer acting under the above persona. Analyze the following source code and git diff, then generate a comprehensive, structured Markdown technical report suitable for a GitHub/GitLab Wiki.\n\n")
	
	promptBuilder.WriteString("## Project Overview\n")
	promptBuilder.WriteString(projectOverview)
	promptBuilder.WriteString("\n\n")

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
- DO NOT output any reasoning steps, planning notes, preambles, metadata summaries, bulleted lists of project stack analysis, or conversational filler text.
- Start IMMEDIATELY with the main Markdown title header ('# 🚀 ...') and output ONLY pure Markdown document content.
- Your entire response will be written directly into a Wiki Markdown file.

### Report Format & Structure Requirements:
You MUST follow the exact Markdown structure below (matching the official Wiki template format):

1. **Title**: '# 🚀 {Project Name} 사용설명 가이드 (Usage Guide)' (or translated to target language)
2. **Project Summary**: 1-2 sentence high-level overview of the project & LLM + MCP engine capabilities.
3. **📌 주요 특징 (Key Features)**:
   - Bulleted/numbered list describing key features and modules found in code.
4. **📖 Wiki 문서 목차 (Navigation)**:
   - Bulleted list of documentation sub-sections and guide links.
5. **⚡ 빠른 시작 (Quick Start)**:
   - **1. 환경 변수 설정 / 주요 설정**: Markdown table with columns: '환경 변수' (Env Var / Setting), '설명' (Description), '기본값' (Default Value).
   - **2. 실행 Command**: Code blocks showing execution commands (e.g. Go, Docker, CI/CD).
6. **🏗️ 시스템 아키텍처 (Architecture)**:
   - Structured diagram (Mermaid format if applicable) or detailed system workflow explanation based on codebase.
7. **🔍 모듈 및 구현 명세 (Implementation Specification)**:
   - Review and specification of key packages, structs, functions, and modules based on the code analysis.
8. **🤝 문의 및 기여 (Contributing & Support)**:
   - Contribution guide, MR/PR submission instructions, and repository links.

### Additional Rules:
- Act strictly according to the defined goal & persona: "%s"
- Output Language: MUST be written entirely in '%s'.
- Provide response ONLY in Markdown Wiki format with clean headings, tables, and code blocks.
`, g.cfg.GoalPrompt, g.cfg.DocsLanguage))

	reqBody := geminiRequest{
		SystemInstruction: &geminiContent{
			Parts: []geminiPart{
				{Text: "You are a strict technical documentation generator. Output ONLY pure Markdown documentation. NEVER output reasoning notes, planning bullet points, preambles, summaries of intent, or conversational wrapper text before or after the report. Start directly with '# 🚀'."},
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
