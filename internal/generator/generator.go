package generator

import (
	"os"
	"path/filepath"

	"github.com/fivethirty/satisficer/internal/content"
)

type Generator struct {
	ouputDir string
	themeDir string
	contents *content.Contents
}

func New(outputDir string, themeDir string, contents *content.Contents) *Generator {
	return &Generator{
		ouputDir: outputDir,
		themeDir: themeDir,
		contents: contents,
	}
}

func (g *Generator) Generate() error {
	if err := os.RemoveAll(g.ouputDir); err != nil {
		return err
	}
	for _, staticContent := range g.contents.StaticContents {
		srcPath := staticContent.Path
		destPath := filepath.Join(g.ouputDir, staticContent.RelativeInputPath)
		destDir := filepath.Dir(destPath)
		if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}
