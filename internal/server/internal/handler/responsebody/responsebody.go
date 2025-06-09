package responsebody

import (
	_ "embed"
	"regexp"
)

var (
	//go:embed html/reload.html
	reloadHTML string
	ignored    = regexp.MustCompile(`(?s)^(?:\s+|<!--.*?-->|<\?.*?\?>)*`)
	tags       = []*regexp.Regexp{
		regexp.MustCompile(`(?is)^<!doctype\s[^>]*>`),
		regexp.MustCompile(`(?is)^<html(?:\s[^>]*)?>`),
		regexp.MustCompile(`(?is)^<head(?:\s[^>]*)?>`),
	}
)

func WithReloadHTML(base []byte) []byte {
	i := 0
	for _, tag := range tags {
		i += len(ignored.Find(base[i:]))
		i += len(tag.Find(base[i:]))
	}

	c := make([]byte, len(base)+len(reloadHTML))
	copy(c, base[:i])
	copy(c[i:], reloadHTML)
	copy(c[i+len(reloadHTML):], base[i:])
	return c
}
