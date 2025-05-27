package generator

import (
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/fivethirty/satisficer/internal/generator/layout"
	"github.com/fivethirty/satisficer/internal/generator/markdown"
	"github.com/fivethirty/satisficer/internal/generator/section"
)

type Generator struct {
	layoutFS  fs.FS
	contentFS fs.FS
	buildDir  string
}

func New(
	layoutFS fs.FS,
	contentFS fs.FS,
	outputDir string,
) (*Generator, error) {
	info, err := os.Stat(outputDir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("output dir is not a directory: %s", outputDir)
	}

	return &Generator{
		layoutFS:  layoutFS,
		contentFS: contentFS,
		buildDir:  outputDir,
	}, nil
}

func (g *Generator) Generate() error {
	slog.Info("Generating site...")
	slog.Info("Reading layout files...")
	l, err := layout.New(g.layoutFS)
	if err != nil {
		return err
	}
	slog.Info("Writing static layout files...")
	if err := g.copyFS(l.Static, "**/*"); err != nil {
		return err
	}
	slog.Info("Writing static content files...")
	if err := g.copyFS(g.contentFS, "**/!(*.md)"); err != nil {
		return err
	}
	slog.Info("Reading markdown content...")
	sections, err := g.sections()
	if err != nil {
		return err
	}
	fmt.Println("Sections:", sections)
	slog.Info("Writing markdown content...")
	return nil
}

func (g *Generator) copyFS(fsys fs.FS, glob string) error {
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		ok, err := filepath.Match(glob, path)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		// xxx: log src -> dest here
		return g.copyFile(fsys, path)
	})
	if err != nil {
		return err
	}
	return nil
}

// this should be write file or something that works with sections
func (g *Generator) copyFile(fsys fs.FS, path string) error {
	src, err := fsys.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = src.Close() }()

	destPath := filepath.Join(g.buildDir, path)
	if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
		return err
	}

	dest, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer func() { _ = dest.Close() }()

	if _, err := io.Copy(dest, src); err != nil {
		return err
	}
	return nil
}

func (g *Generator) sections() (map[string]*section.Section, error) {
	sections := map[string]*section.Section{}
	err := fs.WalkDir(g.contentFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".md" {
			return nil
		}
		file, err := g.contentFS.Open(path)
		if err != nil {
			return err
		}
		content, err := markdown.Parse(file)
		if err != nil {
			return err
		}
		page := section.Page{
			URL:       g.url(path),
			Title:     content.Title,
			CreatedAt: content.CreatedAt,
			UpdatedAt: content.UpdatedAt,
			Content:   content.HTML,
		}
		dir := filepath.Dir(path)
		if _, ok := sections[dir]; !ok {
			sections[dir] = &section.Section{
				Pages: []section.Page{},
			}
		}
		sections[dir].Pages = append(sections[dir].Pages, page)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return sections, nil
}

func (g *Generator) generateSection(section *section.Section, l *layout.Layout) error {
	for _, page := range section.Pages {
		tmpl, err := l.TemplateForContent(page.URL)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(page.URL), os.ModePerm); err != nil {
			return err
		}
		f, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()

		if err := tmpl.Execute(f, page); err != nil {
			return fmt.Errorf("failed to execute template for %s: %w", page.URL, err)
		}
	}
	return nil
}

func (g *Generator) url(path string) string {
	trimmed := strings.TrimSuffix(path, ".md")
	var base string
	if filepath.Base(path) == "index.md" {
		base = fmt.Sprintf("%s.html", trimmed)
	} else {
		base = filepath.Join(trimmed, "index.html")
	}

	return filepath.Join(g.buildDir, base)
}
