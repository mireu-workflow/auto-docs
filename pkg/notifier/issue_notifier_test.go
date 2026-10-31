package notifier_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Mireu-Lab/auto-docs/pkg/analyzer"
	"github.com/Mireu-Lab/auto-docs/pkg/config"
	"github.com/Mireu-Lab/auto-docs/pkg/notifier"
)

func TestFailureNotifier_Creation(t *testing.T) {
	cfg := &config.Config{
		WikiType: "unknown-platform",
	}
	codeAnalyzer := analyzer.NewCodeAnalyzer(".")
	fn := notifier.NewFailureNotifier(cfg, codeAnalyzer)

	// Since unknown platform has no CLI, it should gracefully skip without crashing
	err := fn.NotifyWikiFailure(context.Background(), errors.New("test simulated wiki deployment error"))
	if err != nil {
		t.Errorf("expected no error when CLI is missing, got: %v", err)
	}
}
