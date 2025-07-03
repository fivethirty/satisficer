# Satisficer

![test status](https://github.com/fivethirty/satisficer/actions/workflows/test.yml/badge.svg)
![lint status](https://github.com/fivethirty/satisficer/actions/workflows/lint.yml/badge.svg)


Satisficer is a simple, opinionated static site generator written in Go. It does
a lot less than the competition but as a result the docs fit in a single `README.md`.
It satisfies my needs. Perhaps it will satisfy yours too.

## Features

1. Commonmark compatible Markdown support via [goldmark](https://github.com/yuin/goldmark).
2. No external dependencies besides [goldmark](https://github.com/yuin/goldmark).
3. A simple templating system using Go's `html/template`.
4. A dev server that refreshes the page when files change.

## Getting Started

## Installation

### From Source
```bash
go install github.com/fivethirty/satisficer@latest
```

### Pre-built Binaries
Download from [GitHub Releases](https://github.com/fivethirty/satisficer/releases) or:

```bash
curl -L https://github.com/fivethirty/satisficer/releases/latest/download/satisficer-linux-amd64.tar.gz | tar -xz
```

## Usage

```bash
# Create a new site
satisficer create <project-dir>

# Run the dev server
satisficer serve <project-dir> [-p <port>]

# Build the site
satisficer build <project-dir> <output-dir>
```

## Documentation

Satisficer expects the following project directory structure:

```
├── content
├── layout
│   ├── static
```

### Content

The `content` directory contains a site's content.

#### Markdown Content

All markdown content must contain a JSON front matter block at the top of the
file as follows. All fields except `updatedAt` and `uglyURL` are required.

```markdown
---
{
    "title": "My Cool Page",
    "description": "It's so cool.",
    "createdAt": "2023-06-09T12:00:00Z",
    "updatedAt": "2023-06-09T12:00:00Z",
    "template": "custom.html.tmpl",
    "uglyURL": false
}
---
# Cool Page
```

When building a site, Satisficer renders markdown to HTML using the templates in
the `layout` directory and places the results in the output directory according
to the following logic:

- Any file that matches `content/**/index.md` is rendered to `<output>/**/index.html`.
- Any other markdown file is rendered into `index.html` in a new subdirectory
  with the same name as the file. For example, `content/about.md` is rendered to
  `<output>/about/index.html`.
- Pages with `"uglyURL": true` in frontmatter are rendered as direct `.html` files
  instead of subdirectories. For example, `content/about.md` with `uglyURL: true`
  is rendered to `<output>/about.html`.

#### Non-Markdown Content

Non-markdown files in `content` are copied directly to the output directory.


#### Example Directory Structure

```
content/
├── about.md
├── logo.png
├── index.md
└── posts
    ├── post1.md
    └── post2.md

output/
├── about
│   └── index.html
├── index.html
├── logo.png
└── posts
    ├── post1
    │   └── index.html
    └── post2
        └── index.html
```
### Layout

The `layout` directory contains both static assets and templates used to
render the markdown content in `content` into HTML.

#### Static Assets

The `layout/static` directory contains static assets that are copied directly
into `<output>/static`. This is useful for Favicons, CSS files, etc.

Note that content in `content/static` will be copied to `<output>/static` as
well. As this might cause conflicts, it is recommended to only use one of the
two.


#### Templates

Templates are Go `html/template` files that are used to render the HTML
generated from markdown content into pages.

Each markdown file must specify which template to use via the `template` field
in its frontmatter. The template file will be loaded from the `layout`
directory.

For example, a file with `"template": "page.html.tmpl"` will use the template
located at `layout/page.html.tmpl`.

All templates receive a `Section` struct that contains the current page being
rendered and all other pages in the same directory. This allows templates to
render individual pages or create section listings as needed.

```go
type Section struct {
	Current *Page    // The page being rendered
	Others  []Page   // All other pages in the same directory
	Files   []File   // Non-markdown files in the directory
}

type Page struct {
	URL       string
	Source    string  // Path to the markdown file
	Title     string
	CreatedAt time.Time
	UpdatedAt *time.Time
	Content   string  // Rendered HTML content
}

type File struct {
	URL string
}
```

It is often useful to order pages in a section by various fields. To do this,
Satisficer provides a number of chainable methods on `Pages`:

```go
func (p Pages) ByTitle() Pages
func (p Pages) ByCreatedAt() Pages
func (p Pages) ByUpdatedAt() Pages
func (p Pages) Reverse() Pages
```

A template using both `Current` page and `Others` pages might look like this:

```html
<html>
<head>
    <title>{{ .Current.Title }}</title>
</head>
<body>
    <header>
        <h1>{{ .Current.Title }}</h1>
    </header>
    <main>
        {{ .Current.Content }}
        <h2>Other Pages</h2>
        {{ range .Others.ByCreatedAt.Reverse }}
            <article>
                <h3><a href="{{ .URL }}">{{ .Title }}</a></h3>
                <p>Created at: {{ .CreatedAt.Format "2006-01-02" }}</p>
            </article>
        {{ end }}
    </main>
</body>
</html>
```
