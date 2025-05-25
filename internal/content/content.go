package content

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

type Content struct {
	Title            string
	CreatedAt        time.Time
	UpdatedAt        *time.Time
	TemplateOverride string
	HTML             string
}

type frontMatter struct {
	Title     string     `json:"title"`
	CreatedAt time.Time  `json:"created-at"`
	UpdatedAt *time.Time `json:"updated-at"`
	Template  string     `json:"template"`
}

func (fm *frontMatter) validate() error {
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

func New(reader io.Reader) (*Content, error) {
	pf, err := readPageFile(reader)
	if err != nil {
		return nil, err
	}

	fm := frontMatter{}
	if err := json.Unmarshal(pf.frontMatter, &fm); err != nil {
		return nil, fmt.Errorf("error unmarshalling front matter: %w", err)
	}
	if err := fm.validate(); err != nil {
		return nil, fmt.Errorf("error validating front matter: %w", err)
	}

	buf := &bytes.Buffer{}
	if err := markdown.Convert(pf.content, buf); err != nil {
		return nil, fmt.Errorf("error converting markdown: %w", err)
	}

	template := fm.Template

	return &Content{
		Title:            fm.Title,
		CreatedAt:        fm.CreatedAt,
		UpdatedAt:        fm.UpdatedAt,
		HTML:             buf.String(),
		TemplateOverride: template,
	}, nil
}

var frontMatterDelimiter = []byte{'-', '-', '-'}

type pageFile struct {
	frontMatter []byte
	content     []byte
}

func readPageFile(reader io.Reader) (*pageFile, error) {
	frontMatter := []byte{}
	content := []byte{}

	scanner := bufio.NewScanner(reader)
	scanner.Scan()
	firstLine := scanner.Bytes()
	inFrontMatter := bytes.Equal(firstLine, frontMatterDelimiter)
	if !inFrontMatter {
		return nil, fmt.Errorf("error parsing front matter")
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
		return nil, fmt.Errorf("error reading file: %w", err)
	}
	return &pageFile{
		frontMatter: frontMatter,
		content:     content,
	}, nil
}

func appendLine(b []byte, line string) []byte {
	b = append(b, line...)
	b = append(b, '\n')
	return b
}
