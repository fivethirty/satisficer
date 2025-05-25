package layout

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"path/filepath"
	"strings"
	"text/template"
)

type Layout struct {
	Static   *fs.FS
	Template *template.Template
}

var errNewTemplates = fmt.Errorf("error reading template files")

const staticDir = "static"

func New(fsys fs.FS) (*Layout, error) {
	info, err := fs.Stat(fsys, staticDir)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("error checking static directory: %w", err)
	}
	if err == nil && !info.IsDir() {
		return nil, fmt.Errorf("static directory %s is not a directory", staticDir)
	}

	var static *fs.FS
	if err == nil {
		sub, err := fs.Sub(fsys, staticDir)
		if err != nil {
			return nil, fmt.Errorf("error creating static subdirectory fs: %w", err)
		}
		static = &sub
	}

	tmpl, err := templates(fsys)
	if err != nil {
		return nil, err
	}

	return &Layout{
		Static:   static,
		Template: tmpl,
	}, nil
}

func templates(fsys fs.FS) (*template.Template, error) {
	tmpl := template.New("")
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			if path == staticDir {
				return fs.SkipDir
			}
			return err
		}
		if !strings.HasSuffix(path, ".html.tmpl") {
			return err
		}

		slog.Info("Loading template file", "path", path)

		file, err := fsys.Open(path)
		if err != nil {
			slog.Error("Error opening template file", "path", path, "err", err)
			return errNewTemplates
		}
		bytes := []byte{}
		_, err = file.Read(bytes)
		if err != nil {
			slog.Error("Error reading template file", "path", path, "err", err)
			return errNewTemplates
		}
		_, err = tmpl.New(path).Parse(string(bytes))
		if err != nil {
			slog.Error("Error parsing template file", "path", path, "err", err)
			return errNewTemplates
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return tmpl, nil
}

func (t *Layout) TemplateForContent(contentPath string) (*template.Template, error) {
	base := filepath.Base(contentPath)
	targetBase := "single.html.tmpl"
	if base == "index.md" {
		targetBase = "index.html.tmpl"
	}
	dir := filepath.Dir(contentPath)
	target := filepath.Join(dir, targetBase)
	tmpl := t.Template.Lookup(target)
	for tmpl == nil && dir != "." {
		dir = filepath.Dir(dir)
		target = filepath.Join(dir, targetBase)
		tmpl = t.Template.Lookup(target)
	}
	if tmpl == nil {
		return nil, fmt.Errorf("template for %s not found", contentPath)
	}
	return tmpl, nil
}
