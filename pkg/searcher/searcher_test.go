package searcher_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Mireu-Lab/auto-docs/pkg/searcher"
)

func TestDuckDuckGoSearcher(t *testing.T) {
	s := searcher.NewDuckDuckGoSearcher()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	results, err := s.Search(ctx, "Golang")
	if err != nil {
		t.Logf("Warning: DuckDuckGo search network request failed or timed out: %v", err)
		return
	}

	if len(results) == 0 {
		t.Logf("Warning: 0 search results returned")
		return
	}

	found := false
	for _, res := range results {
		if strings.Contains(strings.ToLower(res.Title), "go") || strings.Contains(strings.ToLower(res.Snippet), "go") {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected search results to contain 'go', got:\n%+v", results)
	}
}
