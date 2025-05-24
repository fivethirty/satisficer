package content_test

import (
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/fivethirty/satisficer/internal/content"
)

func TestFromFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		files       map[string]string
		wantContent func(dir string) content.Contents
		wantError   bool
	}{
		{
			name: "loads files",
			files: map[string]string{
				"index.md": `
					---
					{
						"title": "Index",
						"created-at": "2025-05-13T0:00:00Z",
						"updated-at": "2025-05-14T0:00:00Z",
						"template": "foo.html.tmpl"
					}
					---
					# Index Content
				`,
				"test1.md": `
					---
					{
						"title": "Test Title 1",
						"created-at": "2025-05-13T0:00:00Z"
					}
					---
					# Test Content 1
				`,
				"foo/test2.md": `
					---
					{
						"title": "Test Title 2",
						"created-at": "2025-05-15T0:00:00Z",
						"updated-at": "2025-05-16T0:00:00Z"
					}
					---
					# Test Content 2
				`,
				"foo/test3.txt": "hello!",
			},
			wantContent: func(inputDir string) content.Contents {
				return content.Contents{
					MarkdownContents: []content.MarkdownContent{
						{
							RelativeURL: "index.html",
							Metadata: content.Metadata{
								Title:     "Index",
								CreatedAt: time.Date(2025, 5, 13, 0, 0, 0, 0, time.UTC),
								UpdatedAt: timePtr(time.Date(2025, 5, 14, 0, 0, 0, 0, time.UTC)),
								Template:  "foo.html.tmpl",
							},
							HTML: "<h1>Index Content</h1>\n",
						},
						{
							RelativeURL: "test1/index.html",
							Metadata: content.Metadata{
								Title:     "Test Title 1",
								CreatedAt: time.Date(2025, 5, 13, 0, 0, 0, 0, time.UTC),
							},
							HTML: "<h1>Test Content 1</h1>\n",
						},
						{
							RelativeURL: "foo/test2/index.html",
							Metadata: content.Metadata{
								Title:     "Test Title 2",
								CreatedAt: time.Date(2025, 5, 15, 0, 0, 0, 0, time.UTC),
								UpdatedAt: timePtr(time.Date(2025, 5, 16, 0, 0, 0, 0, time.UTC)),
							},
							HTML: "<h1>Test Content 2</h1>\n",
						},
					},
					StaticContents: []content.StaticContent{
						{
							RelativeURL: "foo/test3.txt",
							FilePath:    filepath.Join(inputDir, "foo/test3.txt"),
						},
					},
				}
			},
		},
		{
			name: "returns error if two files have conflicting output paths",
			files: map[string]string{
				"test.md": `
					---
					{
						"title": "Test Title",
						"created-at": "2025-05-13T0:00:00Z",
						"updated-at": "2025-05-14T0:00:00Z"
					}
					---
					# Test Content
				`,
				"test/index.html": "hello",
			},
			wantError: true,
		},
		{
			name: "returns error if missing title in front matter",
			files: map[string]string{
				"test.md": `
					---
					{
						"created-at": "2025-05-15T0:00:00Z",
						"updated-at": "2025-05-16T0:00:00Z"
					}
					---
					# Test Content
				`,
			},
			wantError: true,
		},
		{
			name: "returns error if missing created-at in front matter",
			files: map[string]string{
				"test.md": `
					---
					{
						"title": "Test Title",
						"description": "Test Description",
						"updated-at": "2025-05-16T0:00:00Z"
					}
					---
					# Test Content
				`,
			},
			wantError: true,
		},
		{
			name: "returns error if front matter is not valid JSON",
			files: map[string]string{
				"test.md": `
					---
					{
						"title": "Test Title",
					---
					# Test Content
				`,
			},
			wantError: true,
		},
		{
			name: "returns error if empty front matter",
			files: map[string]string{
				"test.md": `
					---
					---
					# Test Content
				`,
			},
			wantError: true,
		},
		{
			name: "returns error if no front matter",
			files: map[string]string{
				"test.md": `
					# Test Content
				`,
			},
			wantError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			dir := fmt.Sprintf("content_test-%d", rand.Int())
			path, err := os.MkdirTemp("", dir)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = os.RemoveAll(dir) }()

			for name, content := range test.files {
				content = trimLines(content)
				err := os.MkdirAll(filepath.Dir(filepath.Join(path, name)), os.ModePerm)
				if err != nil {
					t.Fatal(err)
				}
				err = os.WriteFile(filepath.Join(path, name), []byte(content), os.ModePerm)
				if err != nil {
					t.Fatal(err)
				}
			}
			loader := content.NewLoader(path)
			actual, err := loader.Load()
			if err != nil {
				if !test.wantError {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			expected := test.wantContent(path)
			sortMarkdownContent(actual.MarkdownContents)
			sortMarkdownContent(expected.MarkdownContents)
			sortStaticContent(actual.StaticContents)
			sortStaticContent(expected.StaticContents)
			if !reflect.DeepEqual(&expected, actual) {
				t.Fatalf("expected %v, got %v", expected, actual)
			}
		})
	}
}

func sortMarkdownContent(contents []content.MarkdownContent) {
	sort.Slice(contents, func(i, j int) bool {
		return contents[i].RelativeURL < contents[j].RelativeURL
	})
}

func sortStaticContent(contents []content.StaticContent) {
	sort.Slice(contents, func(i, j int) bool {
		return contents[i].RelativeURL < contents[j].RelativeURL
	})
}

func trimLines(s string) string {
	s = strings.TrimSpace(s)
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	return strings.Join(lines, "\n")
}

func timePtr(t time.Time) *time.Time {
