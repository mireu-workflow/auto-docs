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

// ReadFileContent reads a specific file with optional line range (1-indexed, 0 means all lines).
func (a *CodeAnalyzer) ReadFileContent(relPath string, startLine, endLine int) (string, error) {
	fullPath := filepath.Join(a.RepoPath, relPath)
	info, err := os.Stat(fullPath)
	if err != nil {
		return "", fmt.Errorf("file not found: %s", relPath)
	}
	if info.IsDir() {
		return "", fmt.Errorf("path is a directory: %s", relPath)
	}
	if info.Size() > 1024*1024 {
		return "", fmt.Errorf("file size too large (>1MB): %s", relPath)
	}

	contentBytes, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(string(contentBytes), "\n")
	if startLine <= 0 {
		startLine = 1
	}
	if endLine <= 0 || endLine > len(lines) {
		endLine = len(lines)
	}
	if startLine > len(lines) {
		return "", fmt.Errorf("startLine %d exceeds line count %d", startLine, len(lines))
	}

	var sb strings.Builder
	for i := startLine - 1; i < endLine; i++ {
		sb.WriteString(fmt.Sprintf("%4d | %s\n", i+1, lines[i]))
	}
	return sb.String(), nil
}

// SearchCode searches for a query string or regex pattern across repository files using git grep.
func (a *CodeAnalyzer) SearchCode(ctx context.Context, query string, isRegex bool) (string, error) {
	if strings.TrimSpace(query) == "" {
		return "", fmt.Errorf("query cannot be empty")
	}

	args := []string{"grep", "-n", "-I"}
	if isRegex {
		args = append(args, "-E")
	} else {
		args = append(args, "-F")
	}
	args = append(args, query)

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = a.RepoPath
	out, err := cmd.Output()
	if err != nil {
		// git grep returns exit code 1 if no matches found
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return "No matches found for query: " + query, nil
		}
		return "", fmt.Errorf("failed to search code: %w", err)
	}

	result := string(out)
	lines := strings.Split(result, "\n")
	if len(lines) > 200 {
		result = strings.Join(lines[:200], "\n") + fmt.Sprintf("\n... [truncated %d total matches]", len(lines))
	}

	return result, nil
}

type GitMetadata struct {
	BranchName string
	CommitHash string
	Version    string
	Author     string
}

// GetGitMetadata retrieves the current branch name, commit hash, version/tag, and author.
func (a *CodeAnalyzer) GetGitMetadata(ctx context.Context) GitMetadata {
	meta := GitMetadata{
		BranchName: "unknown",
		CommitHash: "unknown",
		Version:    "dev",
	}

	// 1. Branch name from CI environment or git command
	if b := os.Getenv("GITHUB_REF_NAME"); b != "" {
		meta.BranchName = b
	} else if b := os.Getenv("CI_COMMIT_REF_NAME"); b != "" {
		meta.BranchName = b
	} else {
		cmdBranch := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
		cmdBranch.Dir = a.RepoPath
		if out, err := cmdBranch.Output(); err == nil {
			meta.BranchName = strings.TrimSpace(string(out))
		}
	}

	// 2. Commit hash
	if c := os.Getenv("GITHUB_SHA"); c != "" {
		meta.CommitHash = c
	} else if c := os.Getenv("CI_COMMIT_SHA"); c != "" {
		meta.CommitHash = c
	} else {
		cmdHash := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
		cmdHash.Dir = a.RepoPath
		if out, err := cmdHash.Output(); err == nil {
			meta.CommitHash = strings.TrimSpace(string(out))
		}
	}

	// 3. Version / Release Tag
	if v := os.Getenv("VERSION"); v != "" {
		meta.Version = v
	} else if v := os.Getenv("RELEASE_VERSION"); v != "" {
		meta.Version = v
	} else if v := os.Getenv("CI_COMMIT_TAG"); v != "" {
		meta.Version = v
	} else {
		cmdTag := exec.CommandContext(ctx, "git", "describe", "--tags", "--always")
		cmdTag.Dir = a.RepoPath
		if out, err := cmdTag.Output(); err == nil {
			trimmed := strings.TrimSpace(string(out))
			if trimmed != "" {
				meta.Version = trimmed
			}
		}
	}

	// 4. Author / Actor
	if act := os.Getenv("GITHUB_ACTOR"); act != "" {
		meta.Author = act
	} else if act := os.Getenv("GITLAB_USER_LOGIN"); act != "" {
		meta.Author = act
	}

	return meta
}

