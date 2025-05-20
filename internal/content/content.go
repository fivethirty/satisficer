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
	DestinationPath string
	SourcePath      string
}

type MarkdownContent struct {
	DestinationPath string
	Metadata        Metadata
	HTML            string
}

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

type Loader struct {
	goldmark  goldmark.Markdown
	inputDir  string
	outputDir string
}

func NewLoader(inputDir string, outputDir string) *Loader {
	return &Loader{
		goldmark: goldmark.New(
			goldmark.WithExtensions(
				&frontmatter.Extender{},
			),
		),
		inputDir:  inputDir,
		outputDir: outputDir,
	}
}

var ErrLoad error = fmt.Errorf("failed to load content")

func (l *Loader) Load() (*Contents, error) {
	contents := Contents{}
	outputToInput := map[string]string{}

	err := filepath.WalkDir(l.inputDir, func(inputPath string, info fs.DirEntry, err error) error {
		if info.IsDir() {
			return err
		}

		slog.Info(
			"Loading content",
			"content", inputPath,
		)

		outputPath := l.toOutputPath(inputPath)
		if conflictingInputPath, ok := outputToInput[outputPath]; ok {
			slog.Error(
				"Cannot load content, duplicate output path",
				"content", inputPath,
				"conflict", conflictingInputPath,
				"path", outputPath,
			)
			return ErrLoad
		}
		outputToInput[outputPath] = inputPath

		if filepath.Ext(inputPath) == ".md" {
			content, err := l.loadMarkdown(inputPath, outputPath)
			if err != nil {
				slog.Error(
					"Failed to load markdown content",
					"path", inputPath,
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
					SourcePath:      inputPath,
					DestinationPath: outputPath,
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

func (l *Loader) loadMarkdown(inputPath string, outputPath string) (*MarkdownContent, error) {
	content, err := os.ReadFile(inputPath)
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
		DestinationPath: outputPath,
		Metadata:        metadata,
		HTML:            buf.String(),
	}, nil
}

func (l *Loader) toOutputPath(path string) string {
	relativePath := strings.TrimPrefix(
		path,
		fmt.Sprintf("%s%s", l.inputDir, string(os.PathSeparator)),
	)

	directCopyPath := filepath.Join(l.outputDir, relativePath)

	if !strings.HasSuffix(directCopyPath, ".md") {
		return directCopyPath
	}

	var template string
	if filepath.Base(directCopyPath) == "index.md" {
		template = "%s.html"
	} else {
		template = "%s/index.html"
	}

	return fmt.Sprintf(
		template,
		strings.TrimSuffix(directCopyPath, ".md"),
	)
}
