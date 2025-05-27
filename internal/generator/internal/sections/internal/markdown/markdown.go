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
)

type ParsedFile struct {
	FrontMatter FrontMatter
	HTML        string
}

type FrontMatter struct {
	Title     string     `json:"title"`
	CreatedAt time.Time  `json:"created-at"`
	UpdatedAt *time.Time `json:"updated-at"`
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
	if len(missingFields) > 0 {
		return fmt.Errorf(
			"missing required front matter fields: %s",
			strings.Join(missingFields, ", "),
		)
	}
	return nil
}

var markdown = goldmark.New()

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
