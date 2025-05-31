package generator

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"text/template"

	"github.com/fivethirty/satisficer/internal/fsutil"
	"github.com/fivethirty/satisficer/internal/generator/internal/layout"
	"github.com/fivethirty/satisficer/internal/generator/internal/markdown"
	"github.com/fivethirty/satisficer/internal/generator/internal/sections"
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
	if info != nil && !info.IsDir() {
		return nil, fmt.Errorf("output dir is not a directory: %s", outputDir)
	}
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return &Generator{
		layoutFS:  layoutFS,
		contentFS: contentFS,
		buildDir:  outputDir,
	}, nil
}

func (g *Generator) Generate() error {
	slog.Info("Loading layout...")
	l, err := layout.FromFS(g.layoutFS)
	if err != nil {
		return err
	}

	slog.Info("Generating content...")
	s, err := sections.FromFS(g.contentFS, markdown.Parse)
	if err != nil {
		return err
	}

	if l.Static != nil {
		slog.Info("Writing static layout files...")
		err = fsutil.CopyFS(l.Static, g.buildDir)
		if err != nil {
			return err
		}
	} else {
		slog.Info("No static layout files found, skipping...")
	}

	slog.Info("Writing content...")
	for _, section := range s {
		for _, file := range section.Files {
			slog.Info("Copying file", "path", file.URL)
			if err := fsutil.CopyFile(g.contentFS, file.URL, g.buildDir); err != nil {
				return err
			}
		}
		if section.Index != nil {
			slog.Info(
				"Generating index page",
				"path",
				section.Index.URL,
				"from",
				section.Index.Source,
			)
			tmpl, err := l.TemplateForContent(section.Index.Source)
			if err != nil {
				return err
			}
			if err := g.writeContent(tmpl, section, section.Index.URL); err != nil {
				return err
			}
		}
		for _, page := range section.Pages {
			slog.Info("Generating page", "path", page.URL, "from", page.Source)
			tmpl, err := l.TemplateForContent(page.Source)
			if err != nil {
				return err
			}
			if err := g.writeContent(tmpl, page, page.URL); err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *Generator) writeContent(tmpl *template.Template, data any, path string) error {
	dest, err := fsutil.CreateFile(filepath.Join(g.buildDir, path))
	if err != nil {
		return err
	}
	defer func() { _ = dest.Close() }()
	return tmpl.Execute(dest, data)
}
