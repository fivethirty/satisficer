package layout_test

import (
	"reflect"
	"sort"
	"testing"
	"testing/fstest"

	"github.com/fivethirty/satisficer/internal/generator/internal/layout"
	"github.com/fivethirty/satisficer/internal/testutil"
)

var testFile = &fstest.MapFile{
	Data: []byte("a test file"),
}

func TestFromFS(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		fs                fstest.MapFS
		wantTemplatePaths []string
		wantStaticPaths   []string
		wantError         bool
	}{
		{
			name: "can create layout with templates and static files",
			fs: fstest.MapFS{
				"index.html.tmpl":      testFile,
				"page.html.tmpl":       testFile,
				"blog/page.html.tmpl":  testFile,
				"blog/index.html.tmpl": testFile,
				"static/css/main.css":  testFile,
				"static/favicon.ico":   testFile,
			},
			wantTemplatePaths: []string{
				"index.html.tmpl",
				"page.html.tmpl",
				"blog/page.html.tmpl",
				"blog/index.html.tmpl",
			},
			wantStaticPaths: []string{
				"favicon.ico",
				"css/main.css",
			},
		},
		{
			name: "can create layout with only templates",
			fs: fstest.MapFS{
				"index.html.tmpl": testFile,
			},
			wantTemplatePaths: []string{
				"index.html.tmpl",
			},
			wantStaticPaths: []string{},
		},
		{
			name: "can create layout with only static files",
			fs: fstest.MapFS{
				"static/css/main.css": &fstest.MapFile{},
			},
			wantTemplatePaths: []string{},
			wantStaticPaths: []string{
				"css/main.css",
			},
		},
		{
			name:              "can create layout with no templates or static files",
			fs:                fstest.MapFS{},
			wantTemplatePaths: []string{},
			wantStaticPaths:   []string{},
		},
		{
			name: "can't create layout if static dir is not a directory",
			fs: fstest.MapFS{
				"static": testFile,
			},
			wantError: true,
		},
		{
			name: "should treat .html.tmpl files in static dir as static files",
			fs: fstest.MapFS{
				"static/index.html.tmpl":     testFile,
				"static/blog/page.html.tmpl": testFile,
				"page.html.tmpl":             testFile,
			},
			wantTemplatePaths: []string{
				"page.html.tmpl",
			},
			wantStaticPaths: []string{
				"index.html.tmpl",
				"blog/page.html.tmpl",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			l, err := layout.FromFS(test.fs)
			if err != nil {
				if !test.wantError {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if test.wantError {
				t.Fatalf("expected error, got nil")
			}

			templateNames := make([]string, 0, len(test.wantTemplatePaths))
			for _, tmpl := range l.Templates.Templates() {
				templateNames = append(templateNames, tmpl.Name())
			}

			sort.Strings(templateNames)
			sort.Strings(test.wantTemplatePaths)
			if !reflect.DeepEqual(templateNames, test.wantTemplatePaths) {
				t.Fatalf(
					"expected template names %v, got %v",
					test.wantTemplatePaths,
					templateNames,
				)
			}

			staticPaths := []string{}
			if l.Static != nil {
				staticPaths = testutil.SortedPaths(t, l.Static)
			}

			sort.Strings(test.wantStaticPaths)
			if !reflect.DeepEqual(staticPaths, test.wantStaticPaths) {
				t.Fatalf("expected static names %v, got %v", test.wantStaticPaths, staticPaths)
			}
		})
	}
}

func TestTemplateForContent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		contentPath      string
		fs               fstest.MapFS
		wantTemplateName string
		wantErr          bool
	}{
		{
			name:        "can get template for index.md",
			contentPath: "index.md",
			fs: fstest.MapFS{
				"index.html.tmpl": testFile,
			},
			wantTemplateName: "index.html.tmpl",
		},
		{
			name:        "can fallback to root template for index.md",
			contentPath: "foo/index.md",
			fs: fstest.MapFS{
				"index.html.tmpl": testFile,
			},
			wantTemplateName: "index.html.tmpl",
		},
		{
			name:        "can fallback to nearest template for index.md",
			contentPath: "foo/bar/index.md",
			fs: fstest.MapFS{
				"index.html.tmpl":     testFile,
				"foo/index.html.tmpl": testFile,
			},
			wantTemplateName: "foo/index.html.tmpl",
		},
		{
			name:        "can't get template for index.md when no parent index.html.tmpl",
			contentPath: "index.md",
			fs: fstest.MapFS{
				"foo/index.html.tmpl": testFile,
			},
			wantErr: true,
		},
		{
			name:        "can get template for non-index.md page",
			contentPath: "about.md",
			fs: fstest.MapFS{
				"page.html.tmpl": testFile,
			},
			wantTemplateName: "page.html.tmpl",
		},
		{
			name:        "can fallback to root template for non-index.md page",
			contentPath: "blog/post.md",
			fs: fstest.MapFS{
				"page.html.tmpl": testFile,
			},
			wantTemplateName: "page.html.tmpl",
		},
		{
			name:        "can fallback to nearest template for non-index.md page",
			contentPath: "blog/2025/post.md",
			fs: fstest.MapFS{
				"page.html.tmpl":      testFile,
				"blog/page.html.tmpl": testFile,
			},
			wantTemplateName: "blog/page.html.tmpl",
		},
		{
			name:        "can't get template for non-index.md page when no parent page.html.tmpl",
			contentPath: "foo/bar/page.html.tmpl",
			fs:          fstest.MapFS{},
			wantErr:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			l, err := layout.FromFS(test.fs)
			if err != nil {
				t.Fatalf("failed to create templates: %v", err)
			}

			tmpl, err := l.TemplateForContent(test.contentPath)
			if err != nil {
				if !test.wantErr {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if test.wantErr {
				t.Fatalf("expected error, got nil")
			}
			if test.wantTemplateName != tmpl.Name() {
				t.Fatalf("expected template name %s, got %s", test.wantTemplateName, tmpl.Name())
			}
		})
	}
}
