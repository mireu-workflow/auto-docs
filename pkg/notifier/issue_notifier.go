package notifier

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Mireu-Lab/auto-docs/pkg/analyzer"
	"github.com/Mireu-Lab/auto-docs/pkg/config"
)

type FailureNotifier struct {
	cfg      *config.Config
	analyzer *analyzer.CodeAnalyzer
}

func NewFailureNotifier(cfg *config.Config, analyzer *analyzer.CodeAnalyzer) *FailureNotifier {
	return &FailureNotifier{
		cfg:      cfg,
		analyzer: analyzer,
	}
}

// NotifyWikiFailure creates an automated issue on GitHub or GitLab when wiki deployment fails.
func (n *FailureNotifier) NotifyWikiFailure(ctx context.Context, deployErr error) error {
	gitMeta := n.analyzer.GetGitMetadata(ctx)
	now := time.Now()

	title := fmt.Sprintf("fix(wiki): automated alert - wiki deployment failed on %s (%s)", gitMeta.BranchName, gitMeta.Version)
	body := fmt.Sprintf(`* **Project Manager**: limmireu1214@gmail.com
* **Implementation Owner**: limmireu1214@gmail.com

### ⚠️ Automated Operational Alert: Wiki Deployment Failed

The auto-docs automated Wiki deployment failed during execution.

#### 📌 Context & Metadata
- **Git Branch**: %s
- **Commit SHA**: %s
- **Version / Tag**: %s
- **Execution Timestamp**: %s (UTC) / %s (KST)
- **Target Wiki URL**: %s
- **Deploy Mode**: %s

#### 🔍 Error Details
`+"```"+`
%v
`+"```"+`

#### 🛠️ Recommended Action
1. Check repository permissions (e.g. Ensure `+"`contents: write`"+` is granted in GitHub Actions).
2. For GitHub Wiki, ensure the initial Wiki Home page has been initialized in repository settings.
3. For GitLab, verify that `+"`PROJECT_ACCESS_TOKEN`"+` or `+"`GITLAB_TOKEN`"+` has `+"`write_repository`"+` scope.
`,
		gitMeta.BranchName,
		gitMeta.CommitHash,
		gitMeta.Version,
		now.UTC().Format("2006-01-02 15:04:05 MST"),
		now.In(time.FixedZone("KST", 9*3600)).Format("2006-01-02 15:04:05 MST"),
		n.cfg.WikiRepoURL,
		n.cfg.DeployMode,
		deployErr,
	)

	platform := strings.ToLower(n.cfg.WikiType)

	// 1. Try GitHub CLI (gh)
	if platform == "github" || (platform == "" && os.Getenv("GITHUB_ACTIONS") == "true") {
		if _, err := exec.LookPath("gh"); err == nil {
			cmd := exec.CommandContext(ctx, "gh", "issue", "create",
				"--title", title,
				"--body", body,
				"--label", "bug",
				"--assignee", "Mireu-Lab",
			)
			out, err := cmd.CombinedOutput()
			if err != nil {
				// Retry without label if label 'bug' doesn't exist
				cmdFallback := exec.CommandContext(ctx, "gh", "issue", "create",
					"--title", title,
					"--body", body,
					"--assignee", "Mireu-Lab",
				)
				outFallback, errFallback := cmdFallback.CombinedOutput()
				if errFallback == nil {
					fmt.Printf("📢 [Notifier] Successfully created failure issue on GitHub: %s\n", strings.TrimSpace(string(outFallback)))
					return nil
				}
				return fmt.Errorf("gh issue create failed: %s (%w)", string(out), err)
			}
			fmt.Printf("📢 [Notifier] Successfully created failure issue on GitHub: %s\n", strings.TrimSpace(string(out)))
			return nil
		}
	}

	// 2. Try GitLab CLI (glab)
	if platform == "gitlab" || (platform == "" && os.Getenv("GITLAB_CI") == "true") {
		if _, err := exec.LookPath("glab"); err == nil {
			cmd := exec.CommandContext(ctx, "glab", "issue", "create",
				"--title", title,
				"--description", body,
				"--label", "bug",
				"--assignee", "mireu-lab",
			)
			out, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("glab issue create failed: %s (%w)", string(out), err)
			}
			fmt.Printf("📢 [Notifier] Successfully created failure issue on GitLab: %s\n", strings.TrimSpace(string(out)))
			return nil
		}
	}

	fmt.Printf("ℹ️ [Notifier] Neither gh nor glab CLI is configured for issue alerting. Logging failure directly.\n")
	return nil
}
