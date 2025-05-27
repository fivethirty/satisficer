package content

import "github.com/fivethirty/satisficer/internal/generator/content/internal/section"

type Content struct {
	sections map[string]*section.Section
}
