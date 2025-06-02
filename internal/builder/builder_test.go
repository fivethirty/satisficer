package builder_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/fivethirty/satisficer/internal/builder"
	"github.com/fivethirty/satisficer/internal/testutil"
)

func projectFS(t *testing.T, layoutFS fstest.MapFS, contentFS fstest.MapFS) fs.FS {
	t.Helper()

	result := fstest.MapFS{}

	// Create the layout files in the temporary directory
	for path, file := range layoutFS {
		result[filepath.Join(builder.LayoutDir, path)] = file
	}

	for path, file := range contentFS {
		result[filepath.Join(builder.ContentDir, path)] = file
	}

	return result
}

func TestGenerate(t *testing.T) {
	t.Parallel()

	var (
		indexFile = &fstest.MapFile{
			Data: []byte(
				testutil.ToContent(
					t,
					map[string]any{
						"title":     "Home Page",
						"createdAt": "2025-05-13T00:00:00Z",
					},
					"# Welcome to the Home Page",
				),
			),
		}
		indexTemplateFile = &fstest.MapFile{
			Data: []byte("{{ .Index.Title }}"),
		}
		pageFile = &fstest.MapFile{
			Data: []byte(
				testutil.ToContent(
					t,
					map[string]any{
						"title":     "Home Page",
						"createdAt": "2025-05-13T00:00:00Z",
					},
					"# Welcome to Some Other Page",
				),
			),
		}
		pageTemplateFile = &fstest.MapFile{
			Data: []byte("{{ .Title }}"),
		}
		simpleLayoutFS = fstest.MapFS{
			"index.html.tmpl": indexTemplateFile,
			"page.html.tmpl":  pageTemplateFile,
		}
		staticFile = &fstest.MapFile{
			Data: []byte("static content"),
		}
	)

	tests := []struct {
		name      string
		projectFS fs.FS
		layoutFS  fstest.MapFS
		contentFS fstest.MapFS
		wantPaths []string
		wantError bool
	}{
		{
			name:     "can generate a single page site",
			layoutFS: simpleLayoutFS,
			contentFS: fstest.MapFS{
				"index.md": indexFile,
			},
			wantPaths: []string{"index.html"},
		},
		{
			name:     "can generate a multi-page site",
			layoutFS: simpleLayoutFS,
			contentFS: fstest.MapFS{
				"index.md": indexFile,
				"page.md":  pageFile,
			},
			wantPaths: []string{"index.html", "page/index.html"},
		},
		{
			name:     "can generate a site with many directories",
			layoutFS: simpleLayoutFS,
			contentFS: fstest.MapFS{
				"index.md":            indexFile,
				"page.md":             pageFile,
				"blog/index.md":       indexFile,
				"blog/post1.md":       pageFile,
				"blog/post2.md":       pageFile,
				"blog/post3/index.md": indexFile,
			},
			wantPaths: []string{
				"index.html",
				"page/index.html",
				"blog/index.html",
				"blog/post1/index.html",
				"blog/post2/index.html",
				"blog/post3/index.html",
			},
		},
		{
			name:     "won't return an error if two pages have the same URL",
			layoutFS: simpleLayoutFS,
			contentFS: fstest.MapFS{
				"page.md":       pageFile,
				"page/index.md": indexFile,
			},
			wantPaths: []string{"page/index.html"},
		},
		{
			name: "can generate a site with static layout files",
			layoutFS: fstest.MapFS{
				"index.html.tmpl": indexTemplateFile,
				"page.html.tmpl":  pageTemplateFile,
				"static/main.js":  staticFile,
			},
			contentFS: fstest.MapFS{
				"index.md": indexFile,
				"page.md":  pageFile,
			},
			wantPaths: []string{"index.html", "page/index.html", "static/main.js"},
		},
		{
			name:     "can generage a site with static content files",
			layoutFS: simpleLayoutFS,
			contentFS: fstest.MapFS{
				"index.md":      indexFile,
				"page.md":       pageFile,
				"page/logo.png": staticFile,
			},
			wantPaths: []string{"index.html", "page/index.html", "page/logo.png"},
		},
		{
			name:     "returns an error if can't find template",
			layoutFS: fstest.MapFS{},
			contentFS: fstest.MapFS{
				"index.md": indexFile,
			},
			wantError: true,
		},
		{
			name:     "returns an error if can't parse content",
			layoutFS: simpleLayoutFS,
			contentFS: fstest.MapFS{
				"index.md": &fstest.MapFile{
					Data: []byte("invalid content"),
				},
			},
			wantError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()
			pfs := projectFS(t, test.layoutFS, test.contentFS)
			b, err := builder.New(pfs)
			if err != nil {
				t.Fatal(err)
			}
			err = b.Build(dir)
			if err != nil {
				if !test.wantError {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if test.wantError {
				t.Fatal("expected an error but got none")
			}

			actualPaths := testutil.SortedPaths(t, os.DirFS(dir))
			sort.Strings(actualPaths)
			sort.Strings(test.wantPaths)
			if !reflect.DeepEqual(actualPaths, test.wantPaths) {
				t.Fatalf("expected paths %v, got %v", test.wantPaths, actualPaths)
			}
		})
	}
}

func TestPageTemplateRendering(t *testing.T) {
	t.Parallel()

	var sb strings.Builder
	sb.WriteString("{{ .URL}}\n")
	sb.WriteString("{{ .Source}}\n")
	sb.WriteString("{{ .Title}}\n")
	sb.WriteString("{{ .CreatedAt}}\n")
	sb.WriteString("{{ .UpdatedAt}}\n")
	sb.WriteString("{{ .Content}}\n")
	pageTemplate := sb.String()

	layoutFS := fstest.MapFS{
		"page.html.tmpl": {
			Data: []byte(pageTemplate),
		},
	}

	const (
		mdPath   = "page.md"
		htmlPath = "page/index.html"
	)

	tests := []struct {
		name      string
		content   fstest.MapFile
		wantLines []string
	}{
		{
			name: "renders page template with all fields",
			content: fstest.MapFile{
				Data: []byte(
					testutil.ToContent(
						t,
						map[string]any{
							"title":     "Test Page",
							"createdAt": "2025-05-13T00:00:00Z",
							"updatedAt": "2025-05-14T00:00:00Z",
						},
						"# Test Page Content",
					),
				),
			},
			wantLines: []string{
				htmlPath,
				mdPath,
				"Test Page",
				"2025-05-13 00:00:00 +0000 UTC",
				"2025-05-14 00:00:00 +0000 UTC",
				"<h1>Test Page Content</h1>",
			},
		},
		{
			name: "renders page template with only required fields",
			content: fstest.MapFile{
				Data: []byte(
					testutil.ToContent(
						t,
						map[string]any{
							"title":     "Test Page",
							"createdAt": "2025-05-13T00:00:00Z",
						},
						"# Test Page Content",
					),
				),
			},
			wantLines: []string{
				htmlPath,
				mdPath,
				"Test Page",
				"2025-05-13 00:00:00 +0000 UTC",
				"<nil>",
				"<h1>Test Page Content</h1>",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()
			contentFS := fstest.MapFS{
				mdPath: &test.content,
			}
			pfs := projectFS(t, layoutFS, contentFS)
			b, err := builder.New(pfs)
			if err != nil {
				t.Fatal(err)
			}
			err = b.Build(dir)
			if err != nil {
				t.Fatal(err)
			}

			actualFile := filepath.Join(dir, htmlPath)
			actualContent, err := os.ReadFile(actualFile)
			if err != nil {
				t.Fatal(err)
			}

			lines := strings.Split(strings.TrimSpace(string(actualContent)), "\n")
			if !reflect.DeepEqual(lines, test.wantLines) {
				t.Fatalf("expected lines %v, got %v", test.wantLines, lines)
			}
		})
	}
}

func TestIndexTemplateRendering(t *testing.T) {
	t.Parallel()

	var sb strings.Builder
	sb.WriteString("{{ .Index.URL }}\n")
	sb.WriteString("{{ .Index.Source }}\n")
	sb.WriteString("{{ .Index.Title }}\n")
	sb.WriteString("{{ .Index.CreatedAt }}\n")
	sb.WriteString("{{ .Index.UpdatedAt }}\n")
	sb.WriteString("{{ .Index.Content }}\n")
	sb.WriteString("{{ range .Pages }}")
	sb.WriteString("{{ .URL }}\n")
	sb.WriteString("{{ .Source }}\n")
	sb.WriteString("{{ .Title }}\n")
	sb.WriteString("{{ .CreatedAt }}\n")
	sb.WriteString("{{ .UpdatedAt }}\n")
	sb.WriteString("{{ .Content }}\n")
	sb.WriteString("{{ end }}")
	sb.WriteString("{{ range .Files }}")
	sb.WriteString("{{ .URL }}\n")
	sb.WriteString("{{ end }}")
	indexTemplate := sb.String()

	pageTemplate := "{{ .Content }}\n"

	layoutFS := fstest.MapFS{
		"index.html.tmpl": {
			Data: []byte(indexTemplate),
		},
		"page.html.tmpl": {
			Data: []byte(pageTemplate),
		},
	}

	const (
		mdPath   = "index.md"
		htmlPath = "index.html"
	)

	tests := []struct {
		name      string
		contentFS fstest.MapFS
		wantLines []string
	}{
		{
			name: "renders index template",
			contentFS: fstest.MapFS{
				mdPath: &fstest.MapFile{
					Data: []byte(
						testutil.ToContent(
							t,
							map[string]any{
								"title":     "Home Page",
								"createdAt": "2025-05-13T00:00:00Z",
								"updatedAt": "2025-05-14T00:00:00Z",
							},
							"# Welcome to the Home Page",
						),
					),
				},
				"page1.md": &fstest.MapFile{
					Data: []byte(
						testutil.ToContent(
							t,
							map[string]any{
								"title":     "Page 1",
								"createdAt": "2025-05-15T00:00:00Z",
								"updatedAt": "2025-05-16T00:00:00Z",
							},
							"# Content of Page 1",
						),
					),
				},
				"page2.md": &fstest.MapFile{
					Data: []byte(
						testutil.ToContent(
							t,
							map[string]any{
								"title":     "Page 2",
								"createdAt": "2025-05-17T00:00:00Z",
								"updatedAt": "2025-05-18T00:00:00Z",
							},
							"# Content of Page 2",
						),
					),
				},
				"main.js": &fstest.MapFile{
					Data: []byte("console.log('Hello, world!');"),
				},
			},
			wantLines: []string{
				htmlPath,
				mdPath,
				"Home Page",
				"2025-05-13 00:00:00 +0000 UTC",
				"2025-05-14 00:00:00 +0000 UTC",
				"<h1>Welcome to the Home Page</h1>",
				"",
				"page1/index.html",
				"page1.md",
				"Page 1",
				"2025-05-15 00:00:00 +0000 UTC",
				"2025-05-16 00:00:00 +0000 UTC",
				"<h1>Content of Page 1</h1>",
				"",
				"page2/index.html",
				"page2.md",
				"Page 2",
				"2025-05-17 00:00:00 +0000 UTC",
				"2025-05-18 00:00:00 +0000 UTC",
				"<h1>Content of Page 2</h1>",
				"",
				"main.js",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()
			pfs := projectFS(t, layoutFS, test.contentFS)
			g, err := builder.New(pfs)
			if err != nil {
				t.Fatal(err)
			}
			err = g.Build(dir)
			if err != nil {
				t.Fatal(err)
			}

			actualFile := filepath.Join(dir, htmlPath)
			actualContent, err := os.ReadFile(actualFile)
			if err != nil {
				t.Fatal(err)
			}

			lines := strings.Split(strings.TrimSpace(string(actualContent)), "\n")
			if !reflect.DeepEqual(lines, test.wantLines) {
				t.Fatalf("expected lines %v, got %v", test.wantLines, lines)
			}
		})
	}
}
