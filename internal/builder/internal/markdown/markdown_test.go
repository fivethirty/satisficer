package markdown_test

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/fivethirty/satisficer/internal/builder/internal/markdown"
	"github.com/fivethirty/satisficer/internal/testutil"
)

func TestParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		markdown  string
		wantPage  *markdown.ParsedFile
		wantError bool
	}{
		{
			name: "can load a valid page with minimal front matter",
			markdown: testutil.ToContent(
				t,
				map[string]any{
					"title":     "Test Title",
					"createdAt": "2025-05-13T00:00:00Z",
				},
				"# Test Content",
			),
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
			markdown: testutil.ToContent(
				t,
				map[string]any{
					"title":     "Test Title",
					"createdAt": "2025-05-13T00:00:00Z",
					"updatedAt": "2025-05-14T00:00:00Z",
				},
				"# Test Content",
			),
			wantPage: &markdown.ParsedFile{
				FrontMatter: markdown.FrontMatter{
					Title:     "Test Title",
					CreatedAt: time.Date(2025, 5, 13, 0, 0, 0, 0, time.UTC),
					UpdatedAt: testutil.Ptr(t, time.Date(2025, 5, 14, 0, 0, 0, 0, time.UTC)),
				},
				HTML: "<h1>Test Content</h1>\n",
			},
		},
		{
			name:      "can't load a page with no front matter",
			markdown:  "# Test Content",
			wantError: true,
		},
		{
			name:      "can't load a page with empty front matter",
			markdown:  "---\n---\n# Test Content",
			wantError: true,
		},
		{
			name:      "can't load a page with invalid front matter",
			markdown:  "---\n{invalid}\n---\n# Test Content",
			wantError: true,
		},
		{
			name:      "can't load a page with wrong opening front matter delimiter",
			markdown:  "----\n{invalid}\n---\n# Test Content",
			wantError: true,
		},
		{
			name:      "can't load a page with wrong closing front matter delimiter",
			markdown:  "---\n{invalid}\n----\n# Test Content",
			wantError: true,
		},
		{
			name: "can't load a page with missing title in front matter",
			markdown: testutil.ToContent(
				t,
				map[string]any{
					"createdAt": "2025-05-13T00:00:00Z",
				},
				"# Test Content",
			),
			wantError: true,
		},
		{
			name: "can't load a page with missing createdAt in front matter",
			markdown: testutil.ToContent(
				t,
				map[string]any{
					"title": "Test Title",
				},
				"# Test Content",
			),
			wantError: true,
		},
		{
			name: "external HTTP link gets target=_blank",
			markdown: testutil.ToContent(
				t,
				map[string]any{
					"title":     "Test Title",
					"createdAt": "2025-05-13T00:00:00Z",
				},
				"[Example](http://example.com)",
			),
			wantPage: &markdown.ParsedFile{
				FrontMatter: markdown.FrontMatter{
					Title:     "Test Title",
					CreatedAt: time.Date(2025, 5, 13, 0, 0, 0, 0, time.UTC),
					UpdatedAt: nil,
				},
				HTML: `<p><a href="http://example.com" target="_blank" rel="noopener noreferrer">Example</a></p>` + "\n",
			},
		},
		{
			name: "external HTTPS link gets target=_blank",
			markdown: testutil.ToContent(
				t,
				map[string]any{
					"title":     "Test Title",
					"createdAt": "2025-05-13T00:00:00Z",
				},
				"[Example](https://example.com)",
			),
			wantPage: &markdown.ParsedFile{
				FrontMatter: markdown.FrontMatter{
					Title:     "Test Title",
					CreatedAt: time.Date(2025, 5, 13, 0, 0, 0, 0, time.UTC),
					UpdatedAt: nil,
				},
				HTML: `<p><a href="https://example.com" target="_blank" rel="noopener noreferrer">Example</a></p>` + "\n",
			},
		},
		{
			name: "internal relative link remains unchanged",
			markdown: testutil.ToContent(
				t,
				map[string]any{
					"title":     "Test Title",
					"createdAt": "2025-05-13T00:00:00Z",
				},
				"[about page](/about)",
			),
			wantPage: &markdown.ParsedFile{
				FrontMatter: markdown.FrontMatter{
					Title:     "Test Title",
					CreatedAt: time.Date(2025, 5, 13, 0, 0, 0, 0, time.UTC),
					UpdatedAt: nil,
				},
				HTML: `<p><a href="/about">about page</a></p>` + "\n",
			},
		},
		{
			name: "internal anchor link remains unchanged",
			markdown: testutil.ToContent(
				t,
				map[string]any{
					"title":     "Test Title",
					"createdAt": "2025-05-13T00:00:00Z",
				},
				"[section](#heading)",
			),
			wantPage: &markdown.ParsedFile{
				FrontMatter: markdown.FrontMatter{
					Title:     "Test Title",
					CreatedAt: time.Date(2025, 5, 13, 0, 0, 0, 0, time.UTC),
					UpdatedAt: nil,
				},
				HTML: `<p><a href="#heading">section</a></p>` + "\n",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			p, err := markdown.Parse(strings.NewReader(test.markdown))
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
