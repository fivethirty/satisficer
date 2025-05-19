package content

import (
	"bytes"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fivethirty/static/internal/frontmatter"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
)

type Metadata struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"created-at"`
	UpdatedAt   *time.Time `json:"updated-at"`
}

func (fm *Metadata) validate() error {
	missingFields := []string{}
	if fm.Title == "" {
		missingFields = append(missingFields, "title")
	}
	if fm.Description == "" {
		missingFields = append(missingFields, "description")
	}
	if fm.CreatedAt.IsZero() {
		missingFields = append(missingFields, "created-at")
	}
	if len(missingFields) > 0 {
		return fmt.Errorf(
			"missing required front matter fields: %s",
			strings.Join(missingFields, ", "),
		)
	}
	return nil
}

type Content struct {
	RelativePath string
	Metadata     Metadata
	HTML         string
}

type Loader struct {
	goldmark goldmark.Markdown
}

func NewLoader() *Loader {
	return &Loader{
		goldmark: goldmark.New(
			goldmark.WithExtensions(
				&frontmatter.Extender{},
			),
		),
	}
}

func (l *Loader) FromDir(dir string) ([]Content, error) {
	contents := []Content{}
	err := filepath.WalkDir(dir, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".md" {
			slog.Warn("Skipping non-markdown file", "path", path)
			return nil
		}
		content, err := l.fromFile(dir, path)
		if err != nil {
			return err
		}
		contents = append(contents, *content)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return contents, nil
}

func (l *Loader) fromFile(baseDir string, path string) (*Content, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	ctx := parser.NewContext()
	buf := &bytes.Buffer{}
	err = l.goldmark.Convert(content, buf, parser.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	fm, err := frontmatter.Get(ctx)
	if err != nil {
		return nil, err
	}
	metadata := Metadata{}
	if err := fm.Decode(&metadata); err != nil {
		return nil, err
	}
	if err := metadata.validate(); err != nil {
		return nil, err
	}
	relativePath := strings.TrimPrefix(path, fmt.Sprintf("%s%s", baseDir, string(os.PathSeparator)))
	return &Content{
		RelativePath: relativePath,
		Metadata:     metadata,
		HTML:         buf.String(),
	}, nil
}
