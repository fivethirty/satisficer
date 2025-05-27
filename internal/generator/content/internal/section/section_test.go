package section_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/fivethirty/satisficer/internal/generator/content/internal/section"
)

func TestSectionManipulation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		section  func() section.Section
		expected section.Section
	}{
		{
			name: "can sort by title",
			section: func() section.Section {
				return section.Section{
					Pages: []section.Page{
						{Title: "B"},
						{Title: "A"},
						{Title: "C"},
					},
				}.ByTitle()
			},
			expected: section.Section{
				Pages: []section.Page{
					{Title: "A"},
					{Title: "B"},
					{Title: "C"},
				},
			},
		},
		{
			name: "can sort by created at",
			section: func() section.Section {
				return section.Section{
					Pages: []section.Page{
						{CreatedAt: time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)},
						{CreatedAt: time.Date(2021, 1, 2, 0, 0, 0, 0, time.UTC)},
						{CreatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)},
					},
				}.ByCreatedAt()
			},
			expected: section.Section{
				Pages: []section.Page{
					{CreatedAt: time.Date(2021, 1, 2, 0, 0, 0, 0, time.UTC)},
					{CreatedAt: time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)},
					{CreatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)},
				},
			},
		},
		{
			name: "reverse order",
			section: func() section.Section {
				return section.Section{
					Pages: []section.Page{
						{Title: "A"},
						{Title: "B"},
						{Title: "C"},
					},
				}.Reverse()
			},
			expected: section.Section{
				Pages: []section.Page{
					{Title: "C"},
					{Title: "B"},
					{Title: "A"},
				},
			},
		},
		{
			name: "can chain methods",
			section: func() section.Section {
				return section.Section{
					Pages: []section.Page{
						{Title: "B"},
						{Title: "A"},
						{Title: "C"},
					},
				}.ByTitle().Reverse()
			},
			expected: section.Section{
				Pages: []section.Page{
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
