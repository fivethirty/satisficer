package layout

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"strings"
	"text/template"
)

type Layout struct {
	Static    fs.FS
	Templates *template.Template
}

const StaticDir = "static"

func FromFS(fsys fs.FS) (*Layout, error) {
	info, err := fs.Stat(fsys, StaticDir)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	if err == nil && !info.IsDir() {
		return nil, fmt.Errorf("static directory %s is not a directory", StaticDir)
	}

	var static fs.FS
	if err == nil {
		sub, err := fs.Sub(fsys, StaticDir)
		if err != nil {
			return nil, err
		}
		static = sub
	}

	tmpl, err := templates(fsys)
	if err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}

	return &Layout{
		Static:    static,
		Templates: tmpl,
	}, nil
}

func templates(fsys fs.FS) (*template.Template, error) {
	tmpl := template.New("")
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if path == StaticDir {
				return fs.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".html.tmpl") {
			return nil
		}

		slog.Info("Loading template file", "path", path)

		file, err := fsys.Open(path)
		if err != nil {
			return err
		}
		defer func() { _ = file.Close() }()

		bytes, err := io.ReadAll(file)
		if err != nil {
			return err
		}
		_, err = tmpl.New(path).Parse(string(bytes))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return tmpl, nil
}

func (t *Layout) TemplateForContent(
	contentPath string,
	templateFile string,
) (*template.Template, error) {
	tmpl := t.Templates.Lookup(templateFile)
	if tmpl == nil {
		return nil, fmt.Errorf("template %s for %s not found", templateFile, contentPath)
	}
	return tmpl, nil
}
