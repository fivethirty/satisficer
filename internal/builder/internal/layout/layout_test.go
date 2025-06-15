package layout_test

import (
	"reflect"
	"sort"
	"testing"
	"testing/fstest"

	"github.com/fivethirty/satisficer/internal/builder/internal/layout"
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
		templateFile     string
		fs               fstest.MapFS
		wantTemplateName string
		wantErr          bool
	}{
		{
			name:         "finds template when it exists",
			contentPath:  "about.md",
			templateFile: "page.html.tmpl",
			fs: fstest.MapFS{
				"page.html.tmpl": testFile,
			},
			wantTemplateName: "page.html.tmpl",
		},
		{
			name:         "returns error when template not found",
			contentPath:  "about.md",
			templateFile: "missing.html.tmpl",
			fs:           fstest.MapFS{},
			wantErr:      true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			l, err := layout.FromFS(test.fs)
			if err != nil {
				t.Fatalf("failed to create templates: %v", err)
			}

			tmpl, err := l.TemplateForContent(test.contentPath, test.templateFile)
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
