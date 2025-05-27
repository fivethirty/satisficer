package sections

import (
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/fivethirty/satisficer/internal/generator/internal/markdown"
)

type Page struct {
	URL       string
	Source    string
	Title     string
	CreatedAt time.Time
	UpdatedAt *time.Time
}

type Section struct {
	Pages []Page
}

type ParseFunc func(io.Reader) (*markdown.ParsedFile, error)

func FromFS(contentFS fs.FS, parse ParseFunc) (map[string]*Section, error) {
	sections := make(map[string]*Section)
	err := fs.WalkDir(contentFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		file, err := contentFS.Open(path)
		if err != nil {
			return err
		}

		parsed, err := parse(file)
		if err != nil {
			return err
		}

		dir := filepath.Dir(path)
		if _, ok := sections[dir]; !ok {
			sections[dir] = &Section{
				Pages: make([]Page, 0, 10),
			}
		}
		sections[dir].Pages = append(sections[dir].Pages, Page{
			URL:       url(path),
			Source:    path,
			Title:     parsed.FrontMatter.Title,
			CreatedAt: parsed.FrontMatter.CreatedAt,
			UpdatedAt: parsed.FrontMatter.UpdatedAt,
		})

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse content: %w", err)
	}
	return sections, nil
}

func url(path string) string {
	trimmed := strings.TrimSuffix(path, ".md")
	if filepath.Base(path) == "index.md" {
		return fmt.Sprintf("%s.html", trimmed)
	} else {
		return filepath.Join(trimmed, "index.html")
	}
}

func (s Section) ByTitle() Section {
	sort.SliceStable(s.Pages, func(i, j int) bool {
		return s.Pages[i].Title < s.Pages[j].Title
	})
	return s
}

func (s Section) ByCreatedAt() Section {
	sort.SliceStable(s.Pages, func(i, j int) bool {
		return s.Pages[i].CreatedAt.Before(s.Pages[j].CreatedAt)
	})
	return s
}

func (s Section) Reverse() Section {
	slices.Reverse(s.Pages)
	return s
}
