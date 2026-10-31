package wiki

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Mireu-Lab/auto-docs/pkg/config"
)

type WikiDeployer struct {
	cfg *config.Config
}

func NewWikiDeployer(cfg *config.Config) *WikiDeployer {
	return &WikiDeployer{cfg: cfg}
}

func (w *WikiDeployer) Deploy(reportContent string) error {
	// 1. If WikiType is gitlab and glab CLI is installed, attempt glab API deployment first
	if strings.ToLower(w.cfg.WikiType) == "gitlab" {
		if err := w.deployWithGlabCLI(reportContent); err == nil {
			fmt.Println("[WikiDeployer] Successfully deployed Wiki via glab CLI!")
			return nil
		} else {
			fmt.Printf("[WikiDeployer] glab CLI deployment notice: %v. Falling back to git repository deployment...\n", err)
		}
	}

	// 2. If WikiType is github and gh CLI is available, setup gh auth for git operations
	if strings.ToLower(w.cfg.WikiType) == "github" {
		w.setupGitHubCLIAuth()
	}

	// 3. Deploy via Git Wiki Repository
	if w.cfg.WikiRepoURL == "" {
		fmt.Println("[WikiDeployer] WIKI_REPO_URL is not set. Saving report locally to ./wiki_output/Home.md")
		return os.WriteFile("wiki_output/Home.md", []byte(reportContent), 0644)
	}

	tmpDir, err := os.MkdirTemp("", "auto-docs-wiki-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	fmt.Printf("[WikiDeployer] Cloning Wiki repo: %s\n", w.cfg.WikiRepoURL)
	cmdClone := exec.Command("git", "clone", w.cfg.WikiRepoURL, tmpDir)
	if out, err := cmdClone.CombinedOutput(); err != nil {
		return fmt.Errorf("git clone failed: %s (%w)", string(out), err)
	}

	// Configure Git user
	if w.cfg.GitUsername != "" {
		exec.Command("git", "-C", tmpDir, "config", "user.name", w.cfg.GitUsername).Run()
	}
	if w.cfg.GitEmail != "" {
		exec.Command("git", "-C", tmpDir, "config", "user.email", w.cfg.GitEmail).Run()
	}

	// Write Home.md
	homePath := filepath.Join(tmpDir, "Home.md")
	if err := os.WriteFile(homePath, []byte(reportContent), 0644); err != nil {
		return fmt.Errorf("failed to write Home.md: %w", err)
	}

	// Update _Sidebar.md for GitHub Wiki
	sidebarPath := filepath.Join(tmpDir, "_Sidebar.md")
	sidebarContent := "## Navigation\n\n- [[Home]]\n- [[Architecture Overview|Home#architecture--flow-overview]]\n- [[API Specifications|Home#api--module-specifications-derived-from-code]]\n"
	_ = os.WriteFile(sidebarPath, []byte(sidebarContent), 0644)

	// Git Commit & Push
	exec.Command("git", "-C", tmpDir, "add", ".").Run()
	cmdCommit := exec.Command("git", "-C", tmpDir, "commit", "-m", "docs: auto-generated wiki report by auto-docs")
	_ = cmdCommit.Run()

	cmdPush := exec.Command("git", "-C", tmpDir, "push", "origin", "HEAD")
	if out, err := cmdPush.CombinedOutput(); err != nil {
		outStr := string(out)
		if strings.Contains(outStr, "403") || strings.Contains(outStr, "You are not allowed to write") {
			return fmt.Errorf("git push failed (403 Forbidden): %s\n💡 [auto-docs] Tip: GitLab CI_JOB_TOKEN does not have Wiki write permissions by default. Please set PROJECT_ACCESS_TOKEN or GITLAB_TOKEN in CI/CD Variables with write_repository scope.", outStr)
		}
		return fmt.Errorf("git push failed: %s (%w)", outStr, err)
	}

	fmt.Println("[WikiDeployer] Wiki successfully updated and deployed via Git repository!")
	return nil
}

// deployWithGlabCLI attempts to create or update GitLab Wiki pages using `glab api`
func (w *WikiDeployer) deployWithGlabCLI(content string) error {
	if _, err := exec.LookPath("glab"); err != nil {
		return fmt.Errorf("glab CLI not installed")
	}

	projectPath := os.Getenv("CI_PROJECT_PATH")
	if projectPath == "" && w.cfg.WikiRepoURL != "" {
		projectPath = ExtractProjectPathFromURL(w.cfg.WikiRepoURL)
	}
	if projectPath == "" {
		return fmt.Errorf("gitlab project path could not be determined")
	}

	encodedProject := url.PathEscape(projectPath)

	// Try to update existing Home page first via PUT /projects/:id/wikis/home
	updateEndpoint := fmt.Sprintf("projects/%s/wikis/home", encodedProject)
	cmdUpdate := exec.Command("glab", "api", "-X", "PUT", updateEndpoint, "-F", "title=home", "-F", fmt.Sprintf("content=%s", content))
	out, err := cmdUpdate.CombinedOutput()
	if err == nil {
		return nil
	}

	// If page does not exist (404), create it via POST /projects/:id/wikis
	if strings.Contains(string(out), "404") || strings.Contains(string(out), "Not Found") {
		createEndpoint := fmt.Sprintf("projects/%s/wikis", encodedProject)
		cmdCreate := exec.Command("glab", "api", "-X", "POST", createEndpoint, "-F", "title=home", "-F", fmt.Sprintf("content=%s", content))
		if createOut, createErr := cmdCreate.CombinedOutput(); createErr == nil {
			return nil
		} else {
			return fmt.Errorf("glab api create failed: %s (%w)", string(createOut), createErr)
		}
	}

	return fmt.Errorf("glab api update failed: %s (%w)", string(out), err)
}

// setupGitHubCLIAuth configures git credential helper or token with `gh auth setup-git` if gh is present
func (w *WikiDeployer) setupGitHubCLIAuth() {
	if _, err := exec.LookPath("gh"); err == nil {
		cmdAuth := exec.Command("gh", "auth", "setup-git")
		_ = cmdAuth.Run()
	}
}

// ExtractProjectPathFromURL extracts group/project from URL like https://gitlab.com/group/repo.wiki.git
func ExtractProjectPathFromURL(rawURL string) string {
	parts := strings.Split(rawURL, "/")
	if len(parts) >= 2 {
		last := parts[len(parts)-1]
		last = strings.TrimSuffix(last, ".wiki.git")
		last = strings.TrimSuffix(last, ".git")
		secondLast := parts[len(parts)-2]
		// In case user:pass@host was part of secondLast, clean it
		if strings.Contains(secondLast, "@") {
			atIdx := strings.LastIndex(secondLast, "@")
			secondLast = secondLast[atIdx+1:]
		}
		if secondLast != "" && last != "" && !strings.Contains(secondLast, ":") {
			return fmt.Sprintf("%s/%s", secondLast, last)
		}
	}
	return ""
}
