package generator

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/fivethirty/satisficer/internal/generator/internal/fswriter"
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
	slog.Info("Loading layout files...")
	l, err := layout.FromFS(g.layoutFS)
	if err != nil {
		return err
	}

	slog.Info("Parsing content files...")
	s, err := sections.FromFS(g.contentFS, markdown.Parse)
	if err != nil {
		return err
	}

	slog.Info("Copying static layout files...")
	err = fswriter.CopyFilteredFS(l.Static, g.buildDir, fswriter.AllPathFilterFunc)
	if err != nil {
		return err
	}

	// don't need to do this if we have the source i think?
	slog.Info("Copying static content files...")
	err = fswriter.CopyFilteredFS(
		g.contentFS,
		g.buildDir,
		func(path string) bool {
			return filepath.Ext(path) != ".md"
		},
	)
	if err != nil {
		return err
	}

	slog.Info("Generating HTML files...")
	for _, section := range s {
		for _, page := range section.Pages {
			slog.Info("Generating page", "url", page.URL)
			tmpl, err := l.TemplateForContent(page.Source)
			if err != nil {
				return err
			}
			destPath := filepath.Join(g.buildDir, page.URL)
			if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
				return err
			}
			tmpl.ExecuteToFile(destPath, page)
		}
	}

	return nil
}
