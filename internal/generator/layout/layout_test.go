package layout_test

import (
	"io/fs"
	"reflect"
	"sort"
	"testing"
	"testing/fstest"

	"github.com/fivethirty/satisficer/internal/generator/layout"
)

var testFile = &fstest.MapFile{
	Data: []byte("a test file"),
}

func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		fs                fstest.MapFS
		wantTemplateNames []string
		wantStaticNames   []string
		wantError         bool
	}{
		{
			name: "can create layout with templates and static files",
			fs: fstest.MapFS{
				"index.html.tmpl":       testFile,
				"single.html.tmpl":      testFile,
				"blog/single.html.tmpl": testFile,
				"blog/index.html.tmpl":  testFile,
				"static/css/main.css":   testFile,
				"static/favicon.ico":    testFile,
			},
			wantTemplateNames: []string{
				"index.html.tmpl",
				"single.html.tmpl",
				"blog/single.html.tmpl",
				"blog/index.html.tmpl",
			},
			wantStaticNames: []string{
				"favicon.ico",
				"css/main.css",
			},
		},
		{
			name: "can create layout with only templates",
			fs: fstest.MapFS{
				"index.html.tmpl": testFile,
			},
			wantTemplateNames: []string{
				"index.html.tmpl",
			},
			wantStaticNames: []string{},
		},
		{
			name: "can create layout with only static files",
			fs: fstest.MapFS{
				"static/css/main.css": &fstest.MapFile{},
			},
			wantTemplateNames: []string{},
			wantStaticNames: []string{
				"css/main.css",
			},
		},
		{
			name:              "can create layout with no templates or static files",
			fs:                fstest.MapFS{},
			wantTemplateNames: []string{},
			wantStaticNames:   []string{},
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
				"static/index.html.tmpl":       testFile,
				"static/blog/single.html.tmpl": testFile,
				"single.html.tmpl":             testFile,
			},
			wantTemplateNames: []string{
				"single.html.tmpl",
			},
			wantStaticNames: []string{
				"index.html.tmpl",
				"blog/single.html.tmpl",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			l, err := layout.New(test.fs)
			if err != nil {
				if !test.wantError {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if test.wantError {
				t.Fatalf("expected error, got nil")
			}

			templateNames := []string{}
			for _, tmpl := range l.Template.Templates() {
				templateNames = append(templateNames, tmpl.Name())
			}

			sort.Strings(templateNames)
			sort.Strings(test.wantTemplateNames)
			if !reflect.DeepEqual(templateNames, test.wantTemplateNames) {
				t.Fatalf(
					"expected template names %v, got %v",
					test.wantTemplateNames,
					templateNames,
				)
			}

			staticNames := []string{}
			if l.Static != nil {
				err = fs.WalkDir(*l.Static, ".", func(path string, d fs.DirEntry, err error) error {
					if err != nil {
						return err
					}
					if d.IsDir() {
						return nil
					}
					staticNames = append(staticNames, path)
					return nil
				})
				if err != nil {
					t.Fatalf("error walking static dir: %v", err)
				}
			}

			sort.Strings(staticNames)
			sort.Strings(test.wantStaticNames)
			if !reflect.DeepEqual(staticNames, test.wantStaticNames) {
				t.Fatalf("expected static names %v, got %v", test.wantStaticNames, staticNames)
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
				"single.html.tmpl": testFile,
			},
			wantTemplateName: "single.html.tmpl",
		},
		{
			name:        "can fallback to root template for non-index.md page",
			contentPath: "blog/post.md",
			fs: fstest.MapFS{
				"single.html.tmpl": testFile,
			},
			wantTemplateName: "single.html.tmpl",
		},
		{
			name:        "can fallback to nearest template for non-index.md page",
			contentPath: "blog/2025/post.md",
			fs: fstest.MapFS{
				"single.html.tmpl":      testFile,
				"blog/single.html.tmpl": testFile,
			},
			wantTemplateName: "blog/single.html.tmpl",
		},
		{
			name:        "can't get template for non-index.md page when no parent single.html.tmpl",
			contentPath: "foo/bar/single.html.tmpl",
			fs:          fstest.MapFS{},
			wantErr:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			l, err := layout.New(test.fs)
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
