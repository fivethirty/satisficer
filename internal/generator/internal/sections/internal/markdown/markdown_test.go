package markdown_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/fivethirty/satisficer/internal/generator/internal/sections/internal/markdown"
)

func TestParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		markdown  func() (string, error)
		wantPage  *markdown.ParsedFile
		wantError bool
	}{
		{
			name: "can load a valid page with minimal front matter",
			markdown: func() (string, error) {
				return toContent(
					t,
					map[string]any{
						"title":      "Test Title",
						"created-at": "2025-05-13T00:00:00Z",
					},
					"# Test Content",
				)
			},
			wantPage: &markdown.ParsedFile{
				FrontMatter: markdown.FrontMatter{
					Title:     "Test Title",
					CreatedAt: time.Date(2025, 5, 13, 0, 0, 0, 0, time.UTC),
					UpdatedAt: nil,
				},
				HTML: "<h1>Test Content</h1>\n",
			},
		},
		{
			name: "can load a page with all front matter",
			markdown: func() (string, error) {
				return toContent(
					t,
					map[string]any{
						"title":      "Test Title",
						"created-at": "2025-05-13T00:00:00Z",
						"updated-at": "2025-05-14T00:00:00Z",
						"template":   "custom.html.tmpl",
					},
					"# Test Content",
				)
			},
			wantPage: &markdown.ParsedFile{
				FrontMatter: markdown.FrontMatter{
					Title:     "Test Title",
					CreatedAt: time.Date(2025, 5, 13, 0, 0, 0, 0, time.UTC),
					UpdatedAt: timePtr(t, time.Date(2025, 5, 14, 0, 0, 0, 0, time.UTC)),
					Template:  "custom.html.tmpl",
				},
				HTML: "<h1>Test Content</h1>\n",
			},
		},
		{
			name: "can't load a page with no front matter",
			markdown: func() (string, error) {
				return "# Test Content", nil
			},
			wantError: true,
		},
		{
			name: "can't load a page with empty front matter",
			markdown: func() (string, error) {
				return "---\n---\n# Test Content", nil
			},
			wantError: true,
		},
		{
			name: "can't load a page with invalid front matter",
			markdown: func() (string, error) {
				return "---\n{invalid}\n---\n# Test Content", nil
			},
			wantError: true,
		},
		{
			name: "can't load a page with wrong opening front matter delimiter",
			markdown: func() (string, error) {
				return "----\n{invalid}\n---\n# Test Content", nil
			},
			wantError: true,
		},
		{
			name: "can't load a page with wrong closing front matter delimiter",
			markdown: func() (string, error) {
				return "---\n{invalid}\n----\n# Test Content", nil
			},
			wantError: true,
		},
		{
			name: "can't load a page with missing title in front matter",
			markdown: func() (string, error) {
				return toContent(
					t,
					map[string]any{
						"created-at": "2025-05-13T00:00:00Z",
					},
					"# Test Content",
				)
			},
			wantError: true,
		},
		{
			name: "can't load a page with missing created-at in front matter",
			markdown: func() (string, error) {
				return toContent(
					t,
					map[string]any{
						"title": "Test Title",
					},
					"# Test Content",
				)
			},
			wantError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			md, err := test.markdown()
			if err != nil {
				t.Fatal(err)
			}
			p, err := markdown.Parse(strings.NewReader(md))
			if err != nil {
				if !test.wantError {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if test.wantError {
				t.Fatalf("expected error, got nil")
			}
			if !reflect.DeepEqual(p, test.wantPage) {
				t.Fatalf("got page: %v, want page: %v", p, test.wantPage)
			}
		})
	}
}

func timePtr(t *testing.T, tt time.Time) *time.Time {
	t.Helper()
	return &tt
}

func toContent(t *testing.T, fmMap map[string]any, markdown string) (string, error) {
	t.Helper()
	frontMatter, err := json.Marshal(fmMap)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("---\n%s\n---\n%s", frontMatter, markdown), nil
}
