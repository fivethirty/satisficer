package generator

import (
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/fivethirty/satisficer/internal/generator/layout"
	"github.com/fivethirty/satisficer/internal/generator/markdown"
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
	contents, err := g.generateContent()
	if err != nil {
		return fmt.Errorf("failed to generate content: %w", err)
	}
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
		// log src -> dest here
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

func (g *Generator) generateContent() ([]*markdown.Content, error) {
	contents := []*markdown.Content{}
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
		contents = append(contents, content)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return contents, nil
}
