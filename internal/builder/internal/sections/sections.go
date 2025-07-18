package sections

import (
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/fivethirty/satisficer/internal/builder/internal/markdown"
)

type Section struct {
	Current *Page
	Others  Pages
	Files   []File
}
type Pages []Page

type Page struct {
	URL       string
	Source    string
	Title     string
	CreatedAt time.Time
	UpdatedAt *time.Time
	Content   string
	Template  string
	UglyURL   bool
}

type File struct {
	URL string
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

		slog.Info("Processing file", "path", path)

		dir := filepath.Dir(path)
		if _, ok := sections[dir]; !ok {
			sections[dir] = &Section{
				Others: make([]Page, 0, 10),
				Files:  make([]File, 0, 10),
			}
		}

		if !strings.HasSuffix(path, ".md") {
			sections[dir].Files = append(sections[dir].Files, File{
				URL: path,
			})
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

		page := Page{
			URL:       url(path, parsed.FrontMatter.UglyURL),
			Source:    path,
			Title:     parsed.FrontMatter.Title,
			CreatedAt: parsed.FrontMatter.CreatedAt,
			UpdatedAt: parsed.FrontMatter.UpdatedAt,
			Content:   parsed.HTML,
			Template:  parsed.FrontMatter.Template,
			UglyURL:   parsed.FrontMatter.UglyURL,
		}

		sections[dir].Others = append(sections[dir].Others, page)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse content: %w", err)
	}
	return sections, nil
}

func url(path string, uglyURL bool) string {
	trimmed := strings.TrimSuffix(path, ".md")
	if filepath.Base(path) == "index.md" || uglyURL {
		return fmt.Sprintf("%s.html", trimmed)
	} else {
		return filepath.Join(trimmed, "index.html")
	}
}

func (p Pages) ByTitle() Pages {
	sort.SliceStable(p, func(i, j int) bool {
		return p[i].Title < p[j].Title
	})
	return p
}

func (p Pages) ByCreatedAt() Pages {
	sort.SliceStable(p, func(i, j int) bool {
		return p[i].CreatedAt.Before(p[j].CreatedAt)
	})
	return p
}

func (p Pages) Reverse() Pages {
	slices.Reverse(p)
	return p
}

func (s *Section) ForPage(page *Page) *Section {
	otherPages := make(Pages, 0, len(s.Others))
	for _, p := range s.Others {
		if p.Source != page.Source {
			otherPages = append(otherPages, p)
		}
	}

	return &Section{
		Current: page,
		Others:  otherPages,
		Files:   s.Files,
	}
}
