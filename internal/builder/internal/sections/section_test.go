package sections_test

import (
	"io"
	"io/fs"
	"reflect"
	"sort"
	"testing"
	"testing/fstest"
	"time"

	"github.com/fivethirty/satisficer/internal/builder/internal/markdown"
	"github.com/fivethirty/satisficer/internal/builder/internal/sections"
)

func fakeParseFunc(_ io.Reader) (*markdown.ParsedFile, error) {
	return &markdown.ParsedFile{}, nil
}

func TestFromFS(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		contentFS fs.FS
		expected  map[string]*sections.Section
	}{
		{
			name: "can convert a filesystem to sections",
			contentFS: fstest.MapFS{
				"index.md":      &fstest.MapFile{},
				"about.md":      &fstest.MapFile{},
				"blog/index.md": &fstest.MapFile{},
				"blog/post1.md": &fstest.MapFile{},
				"blog/post2.md": &fstest.MapFile{},
			},
			expected: map[string]*sections.Section{
				"blog": {
					Index: &sections.Page{
						URL:    "blog/index.html",
						Source: "blog/index.md",
					},
					Pages: []sections.Page{
						{
							URL:    "blog/post1/index.html",
							Source: "blog/post1.md",
						},
						{
							URL:    "blog/post2/index.html",
							Source: "blog/post2.md",
						},
					},
					Files: []sections.File{},
				},
				".": {
					Index: &sections.Page{
						URL:    "index.html",
						Source: "index.md",
					},
					Pages: []sections.Page{
						{
							URL:    "about/index.html",
							Source: "about.md",
						},
					},
					Files: []sections.File{},
				},
			},
		},
		{
			name: "can ignore non-markdown files when converting",
			contentFS: fstest.MapFS{
				"index.md":   &fstest.MapFile{},
				"about.html": &fstest.MapFile{},
			},
			expected: map[string]*sections.Section{
				".": {
					Index: &sections.Page{
						URL:    "index.html",
						Source: "index.md",
					},
					Pages: []sections.Page{},
					Files: []sections.File{
						{
							URL: "about.html",
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			actual, err := sections.FromFS(test.contentFS, fakeParseFunc)
			if err != nil {
				t.Fatal(err)
			}
			if len(actual) != len(test.expected) {
				t.Fatalf("expected %d sections, got %d", len(test.expected), len(actual))
			}
			for key, expectedSection := range test.expected {
				actualSection, ok := actual[key]
				if !ok {
					t.Fatalf("expected section %q not found", key)
				}
				sortPages(actualSection.Pages)
				sortPages(expectedSection.Pages)
				sortFiles(actualSection.Files)
				sortFiles(expectedSection.Files)
				if !reflect.DeepEqual(actualSection, expectedSection) {
					t.Fatalf(
						"section %q mismatch: got %v, want %v",
						key,
						actualSection,
						expectedSection,
					)
				}
			}
		})
	}
}

func sortPages(pages []sections.Page) {
	sort.Slice(pages, func(i, j int) bool {
		return pages[i].URL < pages[j].URL
	})
}

func sortFiles(files []sections.File) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].URL < files[j].URL
	})
}

func TestPagesSorting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		section  func() sections.Pages
		expected sections.Pages
	}{
		{
			name: "can sort by title",
			section: func() sections.Pages {
				return sections.Pages{
					{Title: "B"},
					{Title: "A"},
					{Title: "C"},
				}.ByTitle()
			},
			expected: sections.Pages{
				{Title: "A"},
				{Title: "B"},
				{Title: "C"},
			},
		},
		{
			name: "can sort by created at",
			section: func() sections.Pages {
				return sections.Pages{
					{CreatedAt: time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)},
					{CreatedAt: time.Date(2021, 1, 2, 0, 0, 0, 0, time.UTC)},
					{CreatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)},
				}.ByCreatedAt()
			},
			expected: sections.Pages{
				{CreatedAt: time.Date(2021, 1, 2, 0, 0, 0, 0, time.UTC)},
				{CreatedAt: time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)},
				{CreatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)},
			},
		},
		{
			name: "reverse order",
			section: func() sections.Pages {
				return sections.Pages{
					{Title: "A"},
					{Title: "B"},
					{Title: "C"},
				}.Reverse()
			},
			expected: sections.Pages{
				{Title: "C"},
				{Title: "B"},
				{Title: "A"},
			},
		},
		{
			name: "can chain methods",
			section: func() sections.Pages {
				return sections.Pages{
					{Title: "B"},
					{Title: "A"},
					{Title: "C"},
				}.ByTitle().Reverse()
			},
			expected: sections.Pages{
				{Title: "C"},
				{Title: "B"},
				{Title: "A"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			actual := test.section()
			if !reflect.DeepEqual(actual, test.expected) {
				t.Fatalf("got section: %v, want section: %v", actual, test.expected)
			}
		})
	}
}
