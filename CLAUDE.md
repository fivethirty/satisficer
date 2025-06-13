# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Satisficer is a simple, opinionated static site generator written in Go. It converts Markdown content with JSON front matter into HTML using Go templates.

## Commands

### Development
- `go run .` - Run the CLI directly from source
- `go test ./...` - Run all tests
- `go test ./internal/package/...` - Run tests for a specific package
- `go build .` - Build the binary
- `go install .` - Install the binary

### CLI Usage
- `satisficer create <project-dir>` - Create a new site
- `satisficer serve <project-dir>` - Run development server (default port 8080)  
- `satisficer build <project-dir> <output-dir>` - Build static site

## Architecture

### Core Components
- **Builder** (`internal/builder/`): Core build logic that processes content and templates
- **Creator** (`internal/creator/`): Scaffolds new projects with starter templates
- **Server** (`internal/server/`): Development server with file watching and auto-reload
- **Commands** (`internal/commands/`): CLI command parsing and execution

### Key Subsystems
- **Layout** (`internal/builder/internal/layout/`): Template parsing and rendering
- **Markdown** (`internal/builder/internal/markdown/`): Markdown processing with goldmark
- **Sections** (`internal/builder/internal/sections/`): Content organization and indexing
- **Handler** (`internal/server/internal/handler/`): HTTP request handling for dev server
- **Watcher** (`internal/server/internal/watcher/`): File system watching for auto-reload

### Content Processing Flow
1. Content is read from `content/` directory
2. Markdown files with JSON front matter are parsed
3. Templates from `layout/` are matched to content based on directory structure
4. Pages are rendered using Go's `html/template` package
5. Static assets are copied from `layout/static/` and `content/`

### Template System
- Page templates (`page.html.tmpl`): Render individual content files
- Index templates (`index.html.tmpl`): Render directory listings with `Section` data
- Template resolution follows directory hierarchy (subdirectory → parent → root)

### Dependencies
- Only external dependency: `github.com/yuin/goldmark` for Markdown processing
- Uses Go standard library for templating, HTTP server, and file operations