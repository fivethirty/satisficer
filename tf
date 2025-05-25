package page_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"testing/fstest"
	"time"

	"github.com/fivethirty/satisficer/internal/page"
)

func TestLoad(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		content   func() (string, error)
		wantPage  *page.Page
		wantError bool
	}{
		{
			name: "can load a valid page with minimal front matter",
			content: func() (string, error) {
				frontMatter, err := json.Marshal(
					map[string]any{
						"title":      "Test Title",
						"created-at": "2025-05-13T00:00:00Z",
					},
				)
				if err != nil {
					return "", err
				}
				return fmt.Sprintf("---\n%s\n---\n# TestContent", frontMatter), nil
			},
			wantPage: &page.Page{
				RelativeURL: "test/index.html",
				Title:       "Test Title",
				CreatedAt:   time.Date(2025, 5, 13, 0, 0, 0, 0, time.UTC),
				UpdatedAt:   nil,
				Content:     "<h1>Test Content</h1>\n\n",
				Template:    "default.html.tmpl",
			},
		},
		/*{
			name: "can load a page with all front matter",
			content: func() (string, error) {
				frontMatter, err := json.Marshal(
					map[string]any{
						"title":      "Test Title",
						"created-at": "2025-05-13T00:00:00Z",
						"updated-at": "2025-05-14T00:00:00Z",
						"template":   "custom.html.tmpl",
					},
				)
				if err != nil {
					return "", err
				}
				return fmt.Sprintf("---\n%s\n---\n# TestContent", frontMatter), nil
			},
			wantPage: &page.Page{
				RelativeURL: "test/index.html",
				Title:       "Test Title",
				CreatedAt:   time.Date(2025, 5, 13, 0, 0, 0, 0, time.UTC),
				UpdatedAt:   timePtr(time.Date(2025, 5, 14, 0, 0, 0, 0, time.UTC)),
				Content:     "<h1>Test Content</h1>\n",
				Template:    "custom.html.tmpl",
			},
		},*/
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			content, err := test.content()
			if err != nil {
				t.Fatalf("unexpected error getting frontmatter: %v", err)
			}
			mapFS := fstest.MapFS{
				"test.md": &fstest.MapFile{
					Data: []byte(content),
				},
			}
			p, err := page.Load(mapFS, "test.md")
			if err != nil {
				if test.wantError {
					return
				}
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(*p, *test.wantPage) {
				t.Fatalf("got page: %v, want page: %v", p, test.wantPage)
			}
		})
	}
}

func timePtr(t time.Time) *time.Time {
	return &t
}
