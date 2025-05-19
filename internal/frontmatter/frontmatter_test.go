package frontmatter_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/fivethirty/satisficer/internal/frontmatter"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
)

type testFrontMatter struct {
	Title string `json:"title"`
}

func TestFrontMatter(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		input           string
		want            testFrontMatter
		wantParseError  bool
		wantGetError    bool
		wantDecodeError bool
	}{
		{
			name: "reads valid JSON front matter",
			input: `
				---
				{
					"title": "Test"
				}
				---
				# Content
			`,
			want: testFrontMatter{
				Title: "Test",
			},
		},
		{
			name: "returns error decoding invalid JSON front matter",
			input: `
				---
				{
					"title": "Test"
				---
				# Content
			`,
			wantDecodeError: true,
		},
		{
			name: "returns error decoding empty front matter",
			input: `
				---
				---
				# Content
			`,
			wantDecodeError: true,
		},
		{
			name: "returns error getting front matter from context with no front matter",
			input: `
				# Content
			`,
			wantGetError: true,
		},
		{
			name: "returns error getting front matter if first delimiter is not 3 dashes",
			input: `
				----
				{
					"title": "Test"
				}
				---
				# Content
			`,
			wantGetError: true,
		},
		{
			name: "returns error decoding front matter if second delimiter is not 3 dashes",
			input: `
				---
				{
					"title": "Test"
				}
				--
				# Content
			`,
			wantDecodeError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			md := goldmark.New(
				goldmark.WithExtensions(&frontmatter.Extender{}),
			)
			var buf bytes.Buffer
			ctx := parser.NewContext()
			input := strings.TrimSpace(test.input)
			err := md.Convert([]byte(input), &buf, parser.WithContext(ctx))
			if err != nil {
				t.Fatalf("failed to convert markdown: %v", err)
			}
			fm, err := frontmatter.Get(ctx)
			if err != nil {
				if test.wantGetError {
					return
				}
				t.Fatalf("failed to get front matter: %v", err)
			}
			fmStruct := testFrontMatter{}
			err = fm.Decode(&fmStruct)
			if err != nil {
				if test.wantDecodeError {
					return
				}
				t.Fatalf("failed to decode front matter: %v", err)
			}
			if fmStruct != test.want {
				t.Errorf("got %v, want %v", fmStruct, test.want)
			}
		})
	}
}
