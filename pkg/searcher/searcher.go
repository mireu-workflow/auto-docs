package searcher

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type SearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

type DuckDuckGoSearcher struct {
	client *http.Client
}

func NewDuckDuckGoSearcher() *DuckDuckGoSearcher {
	return &DuckDuckGoSearcher{
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (d *DuckDuckGoSearcher) Search(ctx context.Context, query string) ([]SearchResult, error) {
	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	form := url.Values{}
	form.Set("q", query)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://html.duckduckgo.com/html/", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("DuckDuckGo HTTP status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	htmlContent := string(body)
	return d.parseHTMLResults(htmlContent), nil
}

func (d *DuckDuckGoSearcher) parseHTMLResults(html string) []SearchResult {
	var results []SearchResult

	// Match result link: <a class="result__a" href="...">Title</a>
	reLink := regexp.MustCompile(`<a[^>]*class="result__a"[^>]*href="([^"]+)"[^>]*>(.*?)</a>`)
	// Match snippet: <a class="result__snippet"[^>]*>(.*?)</a> or <td class="result-snippet"[^>]*>(.*?)</td>
	reSnippet := regexp.MustCompile(`class="result__snippet"[^>]*>(.*?)</a>`)

	linkMatches := reLink.FindAllStringSubmatch(html, 10)
	snippetMatches := reSnippet.FindAllStringSubmatch(html, 10)

	for i, match := range linkMatches {
		if len(match) < 3 {
			continue
		}
		rawURL := match[1]
		title := stripHTML(match[2])

		// Decode DDG redirect URL if needed (e.g. //duckduckgo.com/l/?uddg=...)
		finalURL := rawURL
		if strings.Contains(rawURL, "uddg=") {
			if u, err := url.Parse(rawURL); err == nil {
				if actual := u.Query().Get("uddg"); actual != "" {
					finalURL = actual
				}
			}
		}

		var snippet string
		if i < len(snippetMatches) && len(snippetMatches[i]) >= 2 {
			snippet = stripHTML(snippetMatches[i][1])
		}

		results = append(results, SearchResult{
			Title:   strings.TrimSpace(title),
			URL:     strings.TrimSpace(finalURL),
			Snippet: strings.TrimSpace(snippet),
		})
	}

	return results
}

func stripHTML(input string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	cleaned := re.ReplaceAllString(input, "")
	cleaned = strings.ReplaceAll(cleaned, "&quot;", "\"")
	cleaned = strings.ReplaceAll(cleaned, "&amp;", "&")
	cleaned = strings.ReplaceAll(cleaned, "&lt;", "<")
	cleaned = strings.ReplaceAll(cleaned, "&gt;", ">")
	cleaned = strings.ReplaceAll(cleaned, "&#x27;", "'")
	cleaned = strings.ReplaceAll(cleaned, "&nbsp;", " ")
	return strings.TrimSpace(cleaned)
}
