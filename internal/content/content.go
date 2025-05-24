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

	"github.com/fivethirty/satisficer/internal/frontmatter"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
)

type Contents struct {
	StaticContents   []StaticContent
	MarkdownContents []MarkdownContent
}

type StaticContent struct {
	RelativeURL string
	FilePath    string
}

type MarkdownContent struct {
	RelativeURL string
	Metadata    Metadata
	HTML        string
}

type Metadata struct {
	Title     string     `json:"title"`
	CreatedAt time.Time  `json:"created-at"`
	UpdatedAt *time.Time `json:"updated-at"`
	Template  string     `json:"template"`
}

func (fm *Metadata) validate() error {
	missingFields := []string{}
	if fm.Title == "" {
		missingFields = append(missingFields, "title")
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

type Loader struct {
	goldmark goldmark.Markdown
	inputDir string
}

func NewLoader(inputDir string) *Loader {
	return &Loader{
		goldmark: goldmark.New(
			goldmark.WithExtensions(
				&frontmatter.Extender{},
			),
		),
		inputDir: inputDir,
	}
}

var ErrLoad error = fmt.Errorf("failed to load content")

func (l *Loader) Load() (*Contents, error) {
	contents := Contents{}
	urlToFile := map[string]string{}

	err := filepath.WalkDir(l.inputDir, func(filePath string, info fs.DirEntry, err error) error {
		if info.IsDir() {
			return err
		}

		slog.Info(
			"Loading file",
			"file", filePath,
		)

		relativeURL := l.relativeURL(filePath)
		if conflictingInputPath, ok := urlToFile[relativeURL]; ok {
			slog.Error(
				"Cannot load content, duplicate relative URL",
				"file", filePath,
				"conflitingFile", conflictingInputPath,
				"relativeURL", relativeURL,
			)
			return ErrLoad
		}
		urlToFile[relativeURL] = filePath

		if filepath.Ext(filePath) == ".md" {
			content, err := l.loadMarkdown(filePath, relativeURL)
			if err != nil {
				slog.Error(
					"Failed to load markdown content",
					"path", filePath,
					"error", err,
				)
				return ErrLoad
			}
			contents.MarkdownContents = append(
				contents.MarkdownContents,
				*content,
			)
		} else {
			contents.StaticContents = append(
				contents.StaticContents,
				StaticContent{
					RelativeURL: relativeURL,
					FilePath:    filePath,
				},
			)
		}

		return err
	})
	if err != nil {
		return nil, err
	}

	return &contents, nil
}

func (l *Loader) loadMarkdown(filePath string, relativeURL string) (*MarkdownContent, error) {
	content, err := os.ReadFile(filePath)
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

	return &MarkdownContent{
		RelativeURL: relativeURL,
		Metadata:    metadata,
		HTML:        buf.String(),
	}, nil
}

func (l *Loader) relativeURL(path string) string {
	relativePath := strings.TrimPrefix(
		path,
		fmt.Sprintf("%s%s", l.inputDir, string(os.PathSeparator)),
	)

	if !strings.HasSuffix(relativePath, ".md") {
		return relativePath
	}

	var template string
	if filepath.Base(relativePath) == "index.md" {
		template = "%s.html"
	} else {
		template = "%s/index.html"
	}

	return fmt.Sprintf(
		template,
		strings.TrimSuffix(relativePath, ".md"),
	)
}
