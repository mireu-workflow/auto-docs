package wiki

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Mireu-Lab/auto-docs/pkg/config"
)

type WikiDeployer struct {
	cfg *config.Config
}

func NewWikiDeployer(cfg *config.Config) *WikiDeployer {
	return &WikiDeployer{cfg: cfg}
}

func (w *WikiDeployer) Deploy(reportContent string) error {
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
		return fmt.Errorf("git push failed: %s (%w)", string(out), err)
	}

	fmt.Println("[WikiDeployer] Wiki successfully updated and deployed!")
	return nil
}
