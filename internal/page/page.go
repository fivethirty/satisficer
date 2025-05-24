package page

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/yuin/goldmark"
)

type Page struct {
	RelativeURL string
	Title       string
	CreatedAt   time.Time
	UpdatedAt   *time.Time
	Content     string
	template    *template.Template
}

func (p *Page) Render() (string, error) {
	if p.template == nil {
		return "", fmt.Errorf("template is nil")
	}
	// make this an actual output buffer
	var output strings.Builder
	if err := p.template.Execute(&output, p); err != nil {
		return "", fmt.Errorf("error executing template: %w", err)
	}
	return output.String(), nil
}

type PageLoader struct {
	markdown  goldmark.Markdown
	templates map[string]*template.Template
}

func NewPageLoader(templatesFS fs.FS) (*PageLoader, error) {
	templates, err := loadTemplates(templatesFS)
	if err != nil {
		return nil, fmt.Errorf("NewPageLoader: %w", err)
	}

	return &PageLoader{
		markdown:  goldmark.New(),
		templates: templates,
	}, nil
}

var ErrLoadingTemplates = fmt.Errorf("error loading templates")

const defaultTemplate = "default.html.tmpl"

func loadTemplates(fsys fs.FS) (map[string]*template.Template, error) {
	templates := map[string]*template.Template{}
	err := fs.WalkDir(fsys, ".", func(path string, info fs.DirEntry, err error) error {
		if info.IsDir() {
			return err
		}
		if !strings.HasSuffix(path, ".html.tmpl") {
			slog.Warn("Skipping non-template file", "file", path)
			return err
		}
		slog.Info("Loading template", "file", path)
		tmpl, err := template.ParseFS(fsys, path)
		if err != nil {
			slog.Error("Error parsing template", "file", path, "error", err)
			return ErrLoadingTemplates
		}
		templates[path] = tmpl
		return nil
	})
	if err != nil {
		return nil, err
	}
	if templates[defaultTemplate] == nil {
		return nil, fmt.Errorf("missing default template: %s", defaultTemplate)
	}
	return templates, nil
}

type frontMatter struct {
	Title     string     `json:"title"`
	CreatedAt time.Time  `json:"created-at"`
	UpdatedAt *time.Time `json:"updated-at"`
	Template  string     `json:"template"`
}

func (fm *frontMatter) validate() error {
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

func (g *PageLoader) Load(fsys fs.FS, path string) (*Page, error) {
	if !strings.HasSuffix(path, ".md") {
		return nil, fmt.Errorf("file is not a markdown file")
	}

	pf, err := readPageFile(fsys, path)
	if err != nil {
		return nil, err
	}

	fm := frontMatter{}
	if err := json.Unmarshal(pf.frontMatter, &fm); err != nil {
		return nil, fmt.Errorf("error unmarshalling front matter: %w", err)
	}
	if err := fm.validate(); err != nil {
		return nil, fmt.Errorf("error validating front matter: %w", err)
	}

	buf := &bytes.Buffer{}
	if err := g.markdown.Convert(pf.content, buf); err != nil {
		return nil, fmt.Errorf("error converting markdown: %w", err)
	}

	template, ok := g.templates[fm.Template]
	if !ok {
		template = g.templates[defaultTemplate]
	}

	return &Page{
		RelativeURL: relativeURL(path),
		Title:       fm.Title,
		CreatedAt:   fm.CreatedAt,
		UpdatedAt:   fm.UpdatedAt,
		Content:     buf.String(),
		template:    template,
	}, nil
}

type pageFile struct {
	frontMatter []byte
	content     []byte
}

var frontMatterDelimiter = []byte{'-', '-', '-'}

func readPageFile(fsys fs.FS, path string) (*pageFile, error) {
	frontMatter := []byte{}
	content := []byte{}
	file, err := fsys.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	firstLine := scanner.Bytes()
	inFrontMatter := bytes.Equal(firstLine, frontMatterDelimiter)
	if !inFrontMatter {
		return nil, fmt.Errorf("error parsing front matter")
	}
	for scanner.Scan() {
		line := scanner.Bytes()
		if inFrontMatter && bytes.Equal(line, frontMatterDelimiter) {
			inFrontMatter = false
			continue
		}

		if inFrontMatter {
			frontMatter = appendLine(frontMatter, string(line))
		} else {
			content = appendLine(content, string(line))
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}
	return &pageFile{
		frontMatter: frontMatter,
		content:     content,
	}, nil
}

func appendLine(b []byte, line string) []byte {
	b = append(b, line...)
	b = append(b, '\n')
	return b
}

func relativeURL(path string) string {
	var template string
	if filepath.Base(path) == "index.md" {
		template = "%s.html"
	} else {
		template = "%s/index.html"
	}
	return fmt.Sprintf(template, strings.TrimSuffix(path, ".md"))
}
