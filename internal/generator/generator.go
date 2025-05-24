package generator

import (
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
