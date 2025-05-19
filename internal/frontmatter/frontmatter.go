package frontmatter

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var frontMatterKey = parser.NewContextKey()

type FrontMatter struct {
	raw []byte
}

func Get(ctx parser.Context) (*FrontMatter, error) {
	frontMatter, ok := ctx.Get(frontMatterKey).(*FrontMatter)
	if !ok {
		return nil, fmt.Errorf("could not read front matter from parser context")
	}
	return frontMatter, nil
}

func (fm *FrontMatter) Decode(target any) error {
	fmt.Println("raw front matter:", string(fm.raw))
	return json.Unmarshal(fm.raw, target)
}

type Extender struct{}

func (e *Extender) Extend(md goldmark.Markdown) {
	md.Parser().AddOptions(
		parser.WithBlockParsers(
			util.Prioritized(
				&frontMatterParser{},
				0,
			),
		),
	)
}

type frontMatterParser struct{}

func (fmp *frontMatterParser) Trigger() []byte {
	return []byte{'-'}
}

func (fmp *frontMatterParser) Open(
	_ ast.Node,
	reader text.Reader,
	_ parser.Context,
) (ast.Node, parser.State) {
	if pos, _ := reader.Position(); pos != 0 {
		return nil, parser.NoChildren
	}
	line, _ := reader.PeekLine()
	if isSeparator(line) {
		return ast.NewTextBlock(), parser.NoChildren
	}
	return nil, parser.NoChildren
}

func (fmp *frontMatterParser) Continue(
	node ast.Node,
	reader text.Reader,
	_ parser.Context,
) parser.State {
	line, segment := reader.PeekLine()
	if isSeparator(line) {
		reader.Advance(segment.Len())
		return parser.Close
	}
	node.Lines().Append(segment)
	return parser.Continue | parser.NoChildren
}

func (fmp *frontMatterParser) Close(
	node ast.Node,
	reader text.Reader,
	ctx parser.Context,
) {
	lines := node.Lines()
	var buf bytes.Buffer
	for i := range lines.Len() {
		segment := lines.At(i)
		buf.Write(segment.Value(reader.Source()))
	}

	ctx.Set(frontMatterKey, &FrontMatter{
		raw: buf.Bytes(),
	})

	parent := node.Parent()
	parent.RemoveChild(parent, node)
}

func (fmp *frontMatterParser) CanInterruptParagraph() bool {
	return false
}

func (fmp *frontMatterParser) CanAcceptIndentedLine() bool {
	return false
}

func isSeparator(line []byte) bool {
	line = util.TrimRightSpace(util.TrimLeftSpace(line))
	for i := range line {
		if line[i] != '-' {
			return false
		}
	}
	return len(line) == 3
}
