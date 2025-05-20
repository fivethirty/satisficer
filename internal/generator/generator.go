package generator

import "github.com/fivethirty/satisficer/internal/content"

type Generator struct {
	ouputDir string
	themeDir string
	contents []content.Content
}

func New(outputDir string, themeDir string, contents []content.Content) *Generator {
	return &Generator{
		ouputDir: outputDir,
		themeDir: themeDir,
		contents: contents,
	}
}

func (g *Generator) Generate() error {
	// should we iterate over the contents or somehow make it driven by the templates?
	// like let's walk our template directory and use it to generate stuff? probably better
	return nil
}
