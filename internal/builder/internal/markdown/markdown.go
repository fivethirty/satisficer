package markdown

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type ParsedFile struct {
	FrontMatter FrontMatter
	HTML        string
}

type FrontMatter struct {
	Title     string     `json:"title"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt"`
	Template  string     `json:"template"`
}

func (fm *FrontMatter) validate() error {
	missingFields := []string{}
	if fm.Title == "" {
		missingFields = append(missingFields, "title")
	}
	if fm.CreatedAt.IsZero() {
		missingFields = append(missingFields, "created-at")
	}
	if fm.Template == "" {
		missingFields = append(missingFields, "template")
	}
	if len(missingFields) > 0 {
		return fmt.Errorf(
			"missing required front matter fields: %s",
			strings.Join(missingFields, ", "),
		)
	}
	return nil
}

type externalLinkTransformer struct{}

func (t *externalLinkTransformer) Transform(
	node *ast.Document,
	reader text.Reader,
	pc parser.Context,
) {
	_ = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering || n.Kind() != ast.KindLink {
			return ast.WalkContinue, nil
		}

		link := n.(*ast.Link)
		dest := string(link.Destination)

		if strings.HasPrefix(dest, "http://") || strings.HasPrefix(dest, "https://") {
			link.SetAttribute([]byte("target"), []byte("_blank"))
			link.SetAttribute([]byte("rel"), []byte("noopener noreferrer"))
		}

		return ast.WalkContinue, nil
	})
}

var markdown = goldmark.New(
	goldmark.WithParserOptions(
		parser.WithASTTransformers(
			util.Prioritized(&externalLinkTransformer{}, 100),
		),
	),
)

func Parse(reader io.Reader) (*ParsedFile, error) {
	pf, err := readPageFile(reader)
	if err != nil {
		return nil, err
	}

	parsedFile := &ParsedFile{}

	if err := json.Unmarshal(pf.frontMatter, &parsedFile.FrontMatter); err != nil {
		return nil, err
	}
	if err := parsedFile.FrontMatter.validate(); err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	if err := markdown.Convert(pf.content, buf); err != nil {
		return nil, err
	}
	parsedFile.HTML = buf.String()

	return parsedFile, nil
}

var frontMatterDelimiter = []byte{'-', '-', '-'}

type rawFile struct {
	frontMatter []byte
	content     []byte
}

func readPageFile(reader io.Reader) (*rawFile, error) {
	frontMatter := make([]byte, 0, 1024)
	content := make([]byte, 0, 1024)

	scanner := bufio.NewScanner(reader)
	scanner.Scan()
	firstLine := scanner.Bytes()
	inFrontMatter := bytes.Equal(firstLine, frontMatterDelimiter)
	if !inFrontMatter {
		return nil, fmt.Errorf("could not find front matter")
	}
	for scanner.Scan() {
		line := scanner.Bytes()
		if inFrontMatter && bytes.Equal(line, frontMatterDelimiter) {
			inFrontMatter = false
			continue
		}

		if inFrontMatter {
			frontMatter = appendLine(frontMatter, string(line))
		} else {
			content = appendLine(content, string(line))
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return &rawFile{
		frontMatter: frontMatter,
		content:     content,
	}, nil
}

func appendLine(b []byte, line string) []byte {
	b = append(b, line...)
	b = append(b, '\n')
	return b
}
