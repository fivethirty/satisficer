package generator

import (
	"io/fs"
)

type Generator struct {
}

func New(
	layoutFS fs.FS,
	contentFS fs.FS,
) *Generator {
	return &Generator{}
}
