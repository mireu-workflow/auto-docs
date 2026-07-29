package analyzer

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type GitDiffInfo struct {
	ChangedFiles []string
	DiffSummary  string
}

type CodeAnalyzer struct {
	RepoPath string
}

func NewCodeAnalyzer(repoPath string) *CodeAnalyzer {
	return &CodeAnalyzer{RepoPath: repoPath}
}

func (a *CodeAnalyzer) GetGitDiff(ctx context.Context, baseRef string) (*GitDiffInfo, error) {
	if baseRef == "" {
		baseRef = "HEAD~1"
	}

	cmdFiles := exec.CommandContext(ctx, "git", "diff", "--name-only", baseRef, "HEAD")
	cmdFiles.Dir = a.RepoPath
	outFiles, err := cmdFiles.Output()
	if err != nil {
		// Fallback if HEAD~1 fails (e.g. initial commit or shallow clone)
		cmdFiles = exec.CommandContext(ctx, "git", "ls-files")
		cmdFiles.Dir = a.RepoPath
		outFiles, _ = cmdFiles.Output()
	}

	files := strings.Split(strings.TrimSpace(string(outFiles)), "\n")
	var validFiles []string
	for _, f := range files {
		f = strings.TrimSpace(f)
		if f != "" {
			validFiles = append(validFiles, f)
		}
	}

	cmdDiff := exec.CommandContext(ctx, "git", "diff", baseRef, "HEAD")
	cmdDiff.Dir = a.RepoPath
	outDiff, _ := cmdDiff.Output()

	return &GitDiffInfo{
		ChangedFiles: validFiles,
		DiffSummary:  string(outDiff),
	}, nil
}

func (a *CodeAnalyzer) ReadSourceFiles(files []string) (map[string]string, error) {
	result := make(map[string]string)
	for _, relPath := range files {
		fullPath := filepath.Join(a.RepoPath, relPath)
		info, err := os.Stat(fullPath)
		if err != nil || info.IsDir() {
			continue
		}
		// Skip large or non-source files
		if info.Size() > 500*1024 {
			continue
		}
		content, err := os.ReadFile(fullPath)
		if err == nil {
			result[relPath] = string(content)
		}
	}
	return result, nil
}

func (a *CodeAnalyzer) CollectProjectOverview() (string, error) {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Repository Path: %s\n\n", a.RepoPath))

	// Get file structure
	cmdTree := exec.Command("git", "ls-files")
	cmdTree.Dir = a.RepoPath
	outTree, err := cmdTree.Output()
	if err == nil {
		builder.WriteString("=== Project Files ===\n")
		builder.WriteString(string(outTree))
		builder.WriteString("\n")
	}

	return builder.String(), nil
}
