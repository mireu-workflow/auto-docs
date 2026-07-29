package config

import (
	"fmt"
	"os"
)

type Config struct {
	GeminiAPIKey string
	ModelName    string
	RepoPath     string
	WikiRepoURL  string
	WikiType     string // "github" or "gitlab"
	GitUsername  string
	GitEmail     string
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

	return &Config{
		GeminiAPIKey: apiKey,
		ModelName:    modelName,
		RepoPath:     repoPath,
		WikiRepoURL:  wikiRepoURL,
		WikiType:     wikiType,
		GitUsername:  os.Getenv("GIT_USERNAME"),
		GitEmail:     os.Getenv("GIT_EMAIL"),
	}, nil
}

func (c *Config) Validate() error {
	if c.GeminiAPIKey == "" {
		return fmt.Errorf("GEMINI_API_KEY or AI_STUDIO_API_KEY environment variable is required")
	}
	return nil
}
