package builder

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"text/template"

	"github.com/fivethirty/satisficer/internal/builder/internal/layout"
	"github.com/fivethirty/satisficer/internal/builder/internal/markdown"
	"github.com/fivethirty/satisficer/internal/builder/internal/sections"
	"github.com/fivethirty/satisficer/internal/fsutil"
)

type Builder struct {
	contentFS fs.FS
	layoutFS  fs.FS
}

const (
	LayoutDir  = "layout"
	ContentDir = "content"
)

func New(projectFS fs.FS) (*Builder, error) {
	layoutFS, err := fs.Sub(projectFS, LayoutDir)
	if err != nil {
		return nil, err
	}
	contentFS, err := fs.Sub(projectFS, ContentDir)
	if err != nil {
		return nil, err
	}

	return &Builder{
		contentFS: contentFS,
		layoutFS:  layoutFS,
	}, nil
}

func valiateBuildDir(buildDir string) error {
	info, err := os.Stat(buildDir)
	if info != nil && !info.IsDir() {
		return fmt.Errorf("not a directory: %s", buildDir)
	}
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (b *Builder) Build(buildDir string) error {
	slog.Info("Building project", "outputDir", buildDir)
	if err := valiateBuildDir(buildDir); err != nil {
		return err
	}

	slog.Info("Loading layout...")
	l, err := layout.FromFS(b.layoutFS)
	if err != nil {
		return err
	}

	slog.Info("Generating content...")
	s, err := sections.FromFS(b.contentFS, markdown.Parse)
	if err != nil {
		return err
	}

	if l.Static != nil {
		slog.Info("Writing static layout files...")
		err = fsutil.CopyFS(l.Static, filepath.Join(buildDir, layout.StaticDir))
		if err != nil {
			return err
		}
	} else {
		slog.Info("No static layout files found, skipping...")
	}

	slog.Info("Writing content...")
	for _, section := range s {
		if err := b.writeSection(section, l, buildDir); err != nil {
			return err
		}
	}

	slog.Info("Project built successfully", "outputDir", buildDir)
	return nil
}

func (b *Builder) writeSection(s *sections.Section, l *layout.Layout, buildDir string) error {
	for _, file := range s.Files {
		slog.Info("Copying file", "path", file.URL)
		if err := fsutil.CopyFile(b.contentFS, file.URL, buildDir); err != nil {
			return err
		}
	}

	allPages := s.Others
	for _, page := range allPages {
		slog.Info("Generating page", "path", page.URL, "from", page.Source)
		tmpl, err := l.TemplateForContent(page.Source, page.Template)
		if err != nil {
			return err
		}

		sectionForPage := s.ForPage(&page)

		path := filepath.Join(buildDir, page.URL)
		if err := writeContent(tmpl, sectionForPage, path); err != nil {
			return err
		}
	}
	return nil
}

func writeContent(tmpl *template.Template, data any, path string) error {
	dest, err := fsutil.CreateFile(path)
	if err != nil {
		return err
	}
	defer func() { _ = dest.Close() }()
	return tmpl.Execute(dest, data)
}
