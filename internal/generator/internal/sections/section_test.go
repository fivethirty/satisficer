package sections_test

import (
	"io"
	"io/fs"
	"reflect"
	"sort"
	"testing"
	"testing/fstest"
	"time"

	"github.com/fivethirty/satisficer/internal/generator/internal/sections"
	"github.com/fivethirty/satisficer/internal/generator/internal/sections/internal/markdown"
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
					Pages: []sections.Page{
						{URL: "blog/index.html"},
						{URL: "blog/post1/index.html"},
						{URL: "blog/post2/index.html"},
					},
				},
				".": {
					Pages: []sections.Page{
						{URL: "index.html"},
						{URL: "about/index.html"},
					},
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
					Pages: []sections.Page{
						{URL: "index.html"},
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

func TestSectionManipulation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		section  func() sections.Section
		expected sections.Section
	}{
		{
			name: "can sort by title",
			section: func() sections.Section {
				return sections.Section{
					Pages: []sections.Page{
						{Title: "B"},
						{Title: "A"},
						{Title: "C"},
					},
				}.ByTitle()
			},
			expected: sections.Section{
				Pages: []sections.Page{
					{Title: "A"},
					{Title: "B"},
					{Title: "C"},
				},
			},
		},
		{
			name: "can sort by created at",
			section: func() sections.Section {
				return sections.Section{
					Pages: []sections.Page{
						{CreatedAt: time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)},
						{CreatedAt: time.Date(2021, 1, 2, 0, 0, 0, 0, time.UTC)},
						{CreatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)},
					},
				}.ByCreatedAt()
			},
			expected: sections.Section{
				Pages: []sections.Page{
					{CreatedAt: time.Date(2021, 1, 2, 0, 0, 0, 0, time.UTC)},
					{CreatedAt: time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)},
					{CreatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)},
				},
			},
		},
		{
			name: "reverse order",
			section: func() sections.Section {
				return sections.Section{
					Pages: []sections.Page{
						{Title: "A"},
						{Title: "B"},
						{Title: "C"},
					},
				}.Reverse()
			},
			expected: sections.Section{
				Pages: []sections.Page{
					{Title: "C"},
					{Title: "B"},
					{Title: "A"},
				},
			},
		},
		{
			name: "can chain methods",
			section: func() sections.Section {
				return sections.Section{
					Pages: []sections.Page{
						{Title: "B"},
						{Title: "A"},
						{Title: "C"},
					},
				}.ByTitle().Reverse()
			},
			expected: sections.Section{
				Pages: []sections.Page{
					{Title: "C"},
					{Title: "B"},
					{Title: "A"},
				},
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
