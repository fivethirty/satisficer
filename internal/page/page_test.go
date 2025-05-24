package page_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"testing/fstest"
	"time"

	"github.com/fivethirty/satisficer/internal/page"
)

func TestPageLoader(t *testing.T) {
	t.Parallel()

	templatesFS := fstest.MapFS{
		"default.html.tmpl": &fstest.MapFile{
			Data: []byte("{{ .Title }}"),
		},
		"other.html.tmpl": &fstest.MapFile{
			Data: []byte("{{ .Title }} but different"),
		},
	}

	tests := []struct {
		name        string
		frontMatter func() ([]byte, error)
		markdown    string
		wantPage    page.Page
		wantError   bool
	}{{
		name: "can load a valid page with minimal front matter",
		frontMatter: func() ([]byte, error) {
			return json.Marshal(
				map[string]any{
					"title":      "Test Title",
					"created-at": "2025-05-13T00:00:00Z",
				},
			)
		},
		wantPage: page.Page{
			RelativeURL: "test/index.html",
			Title:       "Test Title",
			CreatedAt:   time.Date(2025, 5, 13, 0, 0, 0, 0, time.UTC),
			UpdatedAt:   nil,
			Content:     "<h1>Test Content</h1>",
		},
		markdown: "# Test Content",
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			loader, err := page.NewPageLoader(templatesFS)
			if err != nil {
				t.Fatalf("unexpected error creating page loader: %v", err)
			}
			frontMatter, err := test.frontMatter()
			if err != nil {
				t.Fatalf("unexpected error getting frontmatter: %v", err)
			}
			fileContent := fmt.Sprintf("---\n%s\n---\n%s", frontMatter, test.markdown)
			mapFS := fstest.MapFS{
				"test.md": &fstest.MapFile{
					Data: []byte(fileContent),
				},
			}
			page, err := loader.Load(mapFS, "test.md")
			if err != nil {
				if test.wantError {
					return
				}
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
