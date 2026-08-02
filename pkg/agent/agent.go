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
	Contents []geminiContent `json:"contents"`
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
### Report Requirements:
1. Act strictly according to the defined goal & persona: "%s"
2. Write a comprehensive User Manual (사용설명서) and Implementation Specification (구현 명세서) based on the current codebase.
3. Include Executive Summary, Architecture & Flow Overview, Module/API Specifications, Setup & Usage Guide.
4. Output Language: MUST be written entirely in '%s'.
5. Provide response ONLY in Markdown Wiki format with clean headings and code blocks.
`, g.cfg.GoalPrompt, g.cfg.DocsLanguage))

	reqBody := geminiRequest{
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

	return gResp.Candidates[0].Content.Parts[0].Text, nil
}
