package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

type Config struct {
	GeminiAPIKey string
	ModelName    string
	RepoPath     string
	WikiRepoURL  string
	WikiType     string // "github" or "gitlab"
	GitUsername  string
	GitEmail     string
	DocsLanguage string // "ko", "en", etc.
	GoalPrompt   string // Persona & Mission goal prompt
	DeployMode   string // "A" ("wiki-only") or "B" ("repo-docs")
}

func LoadConfig() (*Config, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("AI_STUDIO_API_KEY")
	}

	modelName := os.Getenv("GEMINI_MODEL")
	if modelName == "" {
		modelName = "gemma-4-26b-a4b-it"
	}

	repoPath := os.Getenv("REPO_PATH")
	if repoPath == "" {
		repoPath = "."
	}

	wikiRepoURL := os.Getenv("WIKI_REPO_URL")
	wikiType := os.Getenv("WIKI_TYPE")
	if wikiType == "" {
		wikiType = "github"
	}

	docsLanguage := os.Getenv("DOCS_LANGUAGE")
	if docsLanguage == "" {
		docsLanguage = os.Getenv("LANGUAGE")
	}
	if docsLanguage == "" {
		docsLanguage = "ko"
	}

	defaultGoalPrompt := "당신은 소스코드와 Git Diff를 분석하여 개발자 및 사용자를 위한 명확하고 완전한 '시스템 사용설명서 및 기술 명세서'를 작성하는 Gemma 4 AI 테크니컬 라이터입니다. 구현된 소스코드 구조, API/MCP 인터페이스, 환경 변수, 실행 명령어를 체계적인 Markdown Wiki 형식으로 정리하여 작성해야 합니다."
	goalPrompt := os.Getenv("GOAL_PROMPT")
	if goalPrompt == "" {
		goalPrompt = os.Getenv("CUSTOM_GOAL_PROMPT")
	}
	if goalPrompt == "" {
		goalPrompt = defaultGoalPrompt
	}

	gitUser := os.Getenv("GIT_USERNAME")
	if gitUser == "" {
		if os.Getenv("GITHUB_ACTIONS") == "true" || os.Getenv("GITHUB_REPOSITORY") != "" {
			actor := os.Getenv("GITHUB_ACTOR")
			if actor != "" {
				gitUser = actor
			} else {
				gitUser = "github-actions[bot]"
			}
		} else if os.Getenv("GITLAB_CI") == "true" || os.Getenv("CI_PROJECT_PATH") != "" {
			user := os.Getenv("GITLAB_USER_LOGIN")
			if user != "" {
				gitUser = user
			} else {
				gitUser = "gitlab-ci-bot"
			}
		} else {
			gitUser = "auto-docs[bot]"
		}
	}

	gitEmail := os.Getenv("GIT_EMAIL")
	if gitEmail == "" {
		if os.Getenv("GITHUB_ACTIONS") == "true" || os.Getenv("GITHUB_REPOSITORY") != "" {
			gitEmail = "github-actions[bot]@users.noreply.github.com"
		} else if os.Getenv("GITLAB_CI") == "true" || os.Getenv("CI_PROJECT_PATH") != "" {
			gitEmail = "gitlab-ci@example.com"
		} else {
			gitEmail = "auto-docs[bot]@users.noreply.github.com"
		}
	}

	deployMode := strings.ToLower(os.Getenv("DEPLOY_MODE"))
	if deployMode == "" {
		deployMode = strings.ToLower(os.Getenv("DEPLOY_FALLBACK_MODE"))
	}
	if deployMode == "b" || deployMode == "repo-docs" || deployMode == "docs" {
		deployMode = "B"
	} else {
		deployMode = "A" // Default: Wiki-only
	}

	resolvedWikiURL := ResolveWikiRepoURL(wikiRepoURL)

	return &Config{
		GeminiAPIKey: apiKey,
		ModelName:    modelName,
		RepoPath:     repoPath,
		WikiRepoURL:  resolvedWikiURL,
		WikiType:     wikiType,
		GitUsername:  gitUser,
		GitEmail:     gitEmail,
		DocsLanguage: docsLanguage,
		GoalPrompt:   goalPrompt,
		DeployMode:   deployMode,
	}, nil
}

// ResolveWikiRepoURL automatically resolves or sanitizes the Wiki Repository URL.
// It replaces restricted tokens (like gitlab-ci-token) with write access tokens if available,
// injects GitHub or GitLab tokens into raw HTTP URLs, or auto-constructs the URL from CI environment variables.
func ResolveWikiRepoURL(rawURL string) string {
	ghToken := os.Getenv("GITHUB_TOKEN")
	if ghToken == "" {
		ghToken = os.Getenv("GH_TOKEN")
	}
	if ghToken == "" {
		ghToken = os.Getenv("PERSONAL_ACCESS_TOKEN")
	}
	if ghToken == "" {
		ghToken = os.Getenv("PAT")
	}

	glTokens := []string{
		os.Getenv("PROJECT_ACCESS_TOKEN"),
		os.Getenv("GITLAB_TOKEN"),
		os.Getenv("GITLAB_WIKI_TOKEN"),
		os.Getenv("PERSONAL_ACCESS_TOKEN"),
	}
	var glToken string
	for _, t := range glTokens {
		if strings.TrimSpace(t) != "" {
			glToken = strings.TrimSpace(t)
			break
		}
	}

	ciHost := os.Getenv("CI_SERVER_HOST")
	ciPath := os.Getenv("CI_PROJECT_PATH")
	ciJobToken := os.Getenv("CI_JOB_TOKEN")
	ghRepo := os.Getenv("GITHUB_REPOSITORY")

	// 1. If rawURL is provided explicitly
	if rawURL != "" {
		// If rawURL uses gitlab-ci-token and a write token is available, replace it.
		if glToken != "" && (strings.Contains(rawURL, "gitlab-ci-token") || strings.Contains(rawURL, "gitlab-ci-token:[MASKED]")) {
			re := regexp.MustCompile(`https://gitlab-ci-token:[^@]*@`)
			if re.MatchString(rawURL) {
				return re.ReplaceAllString(rawURL, fmt.Sprintf("https://oauth2:%s@", glToken))
			}
		}

		// Inject GitHub Token if rawURL is standard github.com HTTPS without auth credentials
		if ghToken != "" && strings.HasPrefix(rawURL, "https://github.com/") {
			return strings.Replace(rawURL, "https://github.com/", fmt.Sprintf("https://x-access-token:%s@github.com/", ghToken), 1)
		}

		// Inject GitLab Token if rawURL is standard gitlab HTTPS without auth credentials
		if strings.HasPrefix(rawURL, "https://gitlab.") || strings.Contains(rawURL, "gitlab") {
			if !strings.Contains(rawURL, "@") {
				if glToken != "" {
					return strings.Replace(rawURL, "https://", fmt.Sprintf("https://oauth2:%s@", glToken), 1)
				} else if ciJobToken != "" {
					return strings.Replace(rawURL, "https://", fmt.Sprintf("https://gitlab-ci-token:%s@", ciJobToken), 1)
				}
			}
		}

		return rawURL
	}

	// 2. Auto-detect GitLab CI environment if WIKI_REPO_URL is empty
	if ciHost != "" && ciPath != "" {
		if glToken != "" {
			return fmt.Sprintf("https://oauth2:%s@%s/%s.wiki.git", glToken, ciHost, ciPath)
		} else if ciJobToken != "" {
			fmt.Println("⚠️  [auto-docs] Warning: WIKI_REPO_URL not set. Auto-generated URL using CI_JOB_TOKEN. Note: CI_JOB_TOKEN may lack Wiki push permissions (HTTP 403). Set PROJECT_ACCESS_TOKEN or GITLAB_TOKEN in CI/CD variables for write access.")
			return fmt.Sprintf("https://gitlab-ci-token:%s@%s/%s.wiki.git", ciJobToken, ciHost, ciPath)
		}
	}

	// 3. Auto-detect GitHub Actions environment if WIKI_REPO_URL is empty
	if ghRepo != "" && ghToken != "" {
		return fmt.Sprintf("https://x-access-token:%s@github.com/%s.wiki.git", ghToken, ghRepo)
	}

	return ""
}

func (c *Config) Validate() error {
	if c.GeminiAPIKey == "" {
		return fmt.Errorf("GEMINI_API_KEY or AI_STUDIO_API_KEY environment variable is required")
	}
	return nil
}
