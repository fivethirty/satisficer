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
file as follows. All fields except `updatedAt` are required.

```markdown
---
{
    "title": "My Cool Page",
    "description": "It's so cool.",
    "createdAt": "2023-06-09T12:00:00Z",
    "updatedAt": "2023-06-09T12:00:00Z",
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

There are two types of templates in Satisficer: pages and indexes. Page
templates are always named `page.html.tmpl` and are used to render any piece of
markdown content not named `index.md`. Index templates are named
`index.html.tmpl` and are used to render only files named `index.md`.

Any piece of markdown content is rendered using a single template chosen by
starting in the subdirectory of the `layout` directory that matches the
subdirectory of the markdown file in `content`. If no such subdirectory exists
or if the subdirectory does not contain a suitable template, Satisficer will
look at each parent directory in turn until it finds a suitable template. If no
template is found, Satisficer will exit with an error when building.

If that all sounds very confusing, here's a simple example. Assume we have a
piece of content in `content/subdir/about.md`. Satisficer will first look to see
if there is a template `layout/subdir/page.html.tmpl`. If that does not exist,
it will look for `layout/page.html.tmpl`. If that does not exist, it will exit
with an error.

If we had a piece of content in `content/subdir/index.md`, Satisficer would do
the exact same thing, but it would look for `index.html.tmpl` at each step
instead.

##### page.html.tmpl

Page templates are used to render individual pieces of content. Each page
template is passed a `Page` struct that contains the following fields:

```go
type Page struct {
	URL       string
    // the path to the markdown file from which this page was generated
    // relative to the content directory
	Source    string
	Title     string
	CreatedAt time.Time
	UpdatedAt *time.Time
    // the content of the markdown file, rendered to HTML
	Content   string
}
```

A simple page template might look like this:

```html
<html>
<head>
    <title>{{ .Title }}</title>
</head>
<body>
    <header>
        <h1>{{ .Title }}</h1>
    </header>
    <main>
        {{ .Content }}
    </main>
    <footer>
        <p>Created at: {{ .CreatedAt.Format "2006-01-02" }}</p>
        {{ if .UpdatedAt }}
            <p>Updated at: {{ .UpdatedAt.Format "2006-01-02" }}</p>
        {{ end }}
    </footer>
</body>
</html>
```

##### index.html.tmpl

Index templates are used to render information about a whole directory of
content. Each index template is passed a `Section` struct that contains the
following fields:

```go
type Section struct {
    // the Index page itself
	Index *Page
    // all of the other pages in the section
	Pages Pages
    // all of the files in the section that are not markdown content
	Files []File
}
type Pages []Page

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

To sort the pages descending by title, you would use:

```go
{{ range .Pages.ByTitle.Reverse }}
    <a href="{{ .URL }}">{{ .Title }}</a>
{{ end }}
```

A simple index template might look like this:

```html
<html>
<head>
    <title>{{ .Index.Title }}</title>
</head>
<body>
    <header>
        <h1>{{ .Index.Title }}</h1>
        <p>{{ .Index.Description }}</p>
    </header>
    <main>
        {{ range .Pages.ByCreatedAt.Reverse }}
            <article>
                <h2><a href="{{ .URL }}">{{ .Title }}</a></h2>
                <p>Created at: {{ .CreatedAt.Format "2006-01-02" }}</p>
                {{ if .UpdatedAt }}
                    <p>Updated at: {{ .UpdatedAt.Format "2006-01-02" }}</p>
                {{ end }}
            </article>
        {{ end }}
    </main>
</body>
