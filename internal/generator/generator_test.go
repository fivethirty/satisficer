package generator_test

import (
	"io/fs"
	"os"
	"reflect"
	"sort"
	"testing"
	"testing/fstest"

	"github.com/fivethirty/satisficer/internal/generator"
	"github.com/fivethirty/satisficer/internal/generator/internal/testutil"
)

func TestGenerate(t *testing.T) {
	t.Parallel()

	var (
		indexFile = &fstest.MapFile{
			Data: []byte(
				testutil.ToContent(
					t,
					map[string]any{
						"title":      "Home Page",
						"created-at": "2025-05-13T00:00:00Z",
					},
					"# Welcome to the Home Page",
				),
			),
		}
		indexTemplateFile = &fstest.MapFile{
			Data: []byte("{{ .IndexPage.Title }}"),
		}
		pageFile = &fstest.MapFile{
			Data: []byte(
				testutil.ToContent(
					t,
					map[string]any{
						"title":      "Home Page",
						"created-at": "2025-05-13T00:00:00Z",
					},
					"# Welcome to Some Other Page",
				),
			),
		}
		pageTemplateFile = &fstest.MapFile{
			Data: []byte("{{ .Page.Title }}"),
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
		layoutFS  fs.FS
		contentFS fs.FS
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
			wantPaths: []string{"index.html", "page/index.html", "main.js"},
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
			g, err := generator.New(test.layoutFS, test.contentFS, dir)
			if err != nil {
				t.Fatal(err)
			}
			err = g.Generate()
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
