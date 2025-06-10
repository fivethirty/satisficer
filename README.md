# Satisficer

Satisficer is a simple, opinionated static site generator written in Go. It does
a lot less than the competition but as a result the docs fit on a single page.
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
<project-dir>
├── content
├── layouts
│   ├── static
```
