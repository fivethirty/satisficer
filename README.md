# // ---[S A T I S F I C E R]--- \\\\

Satisficer is a simple, opinionated static site generator written in Go. It does
a lot less than the competition but as a result the docs fit in a single README.
It satisfies my needs. Perhaps it will satisfy yours too.

## Features

1. Commonmark compatible Markdown support via [goldmark](https://github.com/yuin/goldmark).
2. No external dependencies besides [goldmark](https://github.com/yuin/goldmark).
3. A simple templating system using Go's `html/template`.
4. A dev server that refreshes the page when files change.

## Getting Started

```bash
go install github.com/fivethirty/satisficer@latest
```

## Usage

```bash
# Create a new site
satisficer create <project-dir>

# Run the dev server
satisficer serve <project-dir> [--port <port>]

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

The `content` directory contains a sites content.

Markdown files in `content` are processed and rendered to HTML. The result is
written to the output directory usng the following rules:

- Any file that matches `**/index.md` is rendered to `**/index.html`.
- Any other markdown file is rendered into a new subdirectory with the same
  name as the file, e.g. `about.md` is rendered to `about/index.html`.

Non-markdown files in `content` are copied to the output directory as-is.

Here's an example of a `content` directory and the resulting output:

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

