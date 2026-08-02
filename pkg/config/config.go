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
}

func LoadConfig() (*Config, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("AI_STUDIO_API_KEY")
	}

	modelName := os.Getenv("GEMINI_MODEL")
	if modelName == "" {
		modelName = "gemma-4"
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

	defaultGoalPrompt := "우리는 해당 코드를 보고 실행 여부를 떠나서, 현재 구현되어있는 코드를 기반으로 review하는 주니어 엔지니어이다. 우리는 구현되어있는 코드를 기반으로, 사용설명서 및 구현 명세서를 작성하여 wiki에다가 올려야하는 임무가 있다."
	goalPrompt := os.Getenv("GOAL_PROMPT")
	if goalPrompt == "" {
		goalPrompt = os.Getenv("CUSTOM_GOAL_PROMPT")
	}
	if goalPrompt == "" {
		goalPrompt = defaultGoalPrompt
	}

	resolvedWikiURL := ResolveWikiRepoURL(wikiRepoURL)

	return &Config{
		GeminiAPIKey: apiKey,
		ModelName:    modelName,
		RepoPath:     repoPath,
		WikiRepoURL:  resolvedWikiURL,
		WikiType:     wikiType,
		GitUsername:  os.Getenv("GIT_USERNAME"),
		GitEmail:     os.Getenv("GIT_EMAIL"),
		DocsLanguage: docsLanguage,
		GoalPrompt:   goalPrompt,
	}, nil
}

// ResolveWikiRepoURL automatically resolves or sanitizes the Wiki Repository URL.
// It replaces restricted tokens (like gitlab-ci-token) with write access tokens if available,
// or auto-constructs the URL from GitLab CI / GitHub Actions environment variables.
func ResolveWikiRepoURL(rawURL string) string {
	writeTokens := []string{
		os.Getenv("PROJECT_ACCESS_TOKEN"),
		os.Getenv("GITLAB_TOKEN"),
		os.Getenv("GITLAB_WIKI_TOKEN"),
		os.Getenv("PERSONAL_ACCESS_TOKEN"),
	}
	var token string
	for _, t := range writeTokens {
		if strings.TrimSpace(t) != "" {
			token = strings.TrimSpace(t)
			break
		}
	}

	ciHost := os.Getenv("CI_SERVER_HOST")
	ciPath := os.Getenv("CI_PROJECT_PATH")
	ciJobToken := os.Getenv("CI_JOB_TOKEN")

	// 1. If rawURL is provided explicitly
	if rawURL != "" {
		// If rawURL uses gitlab-ci-token and a write token is available, replace it.
		if token != "" && (strings.Contains(rawURL, "gitlab-ci-token") || strings.Contains(rawURL, "gitlab-ci-token:[MASKED]")) {
			re := regexp.MustCompile(`https://gitlab-ci-token:[^@]*@`)
			if re.MatchString(rawURL) {
				return re.ReplaceAllString(rawURL, fmt.Sprintf("https://oauth2:%s@", token))
			}
		}
		return rawURL
	}

	// 2. Auto-detect GitLab CI environment if WIKI_REPO_URL is empty
	if ciHost != "" && ciPath != "" {
		if token != "" {
			return fmt.Sprintf("https://oauth2:%s@%s/%s.wiki.git", token, ciHost, ciPath)
		} else if ciJobToken != "" {
			fmt.Println("⚠️  [auto-docs] Warning: WIKI_REPO_URL not set. Auto-generated URL using CI_JOB_TOKEN. Note: CI_JOB_TOKEN may lack Wiki push permissions (HTTP 403). Set PROJECT_ACCESS_TOKEN or GITLAB_TOKEN in CI/CD variables for write access.")
			return fmt.Sprintf("https://gitlab-ci-token:%s@%s/%s.wiki.git", ciJobToken, ciHost, ciPath)
		}
	}

	// 3. Auto-detect GitHub Actions environment if WIKI_REPO_URL is empty
	ghRepo := os.Getenv("GITHUB_REPOSITORY")
	ghToken := os.Getenv("GITHUB_TOKEN")
	if ghToken == "" {
		ghToken = os.Getenv("GH_TOKEN")
	}
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
