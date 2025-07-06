package responsebody_test

import (
	_ "embed"
	"fmt"
	"testing"

	"github.com/fivethirty/satisficer/internal/server/internal/handler/responsebody"
)

//go:embed html/reload.html
var reloadHTML string

func TestWithReloadHTML(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "inject into empty input",
			input:    "",
			expected: reloadHTML,
		},
		{
			name:     "inject before text content",
			input:    "some random text.",
			expected: fmt.Sprintf("%ssome random text.", reloadHTML),
		},
		{
			name:     "inject before non-considred HTML elements",
			input:    "<h1>welcome!</h1>",
			expected: fmt.Sprintf("%s<h1>welcome!</h1>", reloadHTML),
		},
		{
			name:  "inject after head",
			input: "<!doctype html><html><head>other stuff",
			expected: fmt.Sprintf(
				"<!doctype html><html><head>%sother stuff",
				reloadHTML,
			),
		},
		{
			name:  "inject after head when html missing",
			input: "<!doctype html><head>other stuff",
			expected: fmt.Sprintf(
				"<!doctype html><head>%sother stuff",
				reloadHTML,
			),
		},
		{
			name:  "inject after head when html missing",
			input: "<!doctype html><head>other stuff",
			expected: fmt.Sprintf(
				"<!doctype html><head>%sother stuff",
				reloadHTML,
			),
		},
		{
			name:  "inject after head when doctype and html missing",
			input: "<head>other stuff",
			expected: fmt.Sprintf(
				"<head>%sother stuff",
				reloadHTML,
			),
		},
		{
			name:  "inject after html when head missing",
			input: "<!doctype html><html>other stuff",
			expected: fmt.Sprintf(
				"<!doctype html><html>%sother stuff",
				reloadHTML,
			),
		},
		{
			name:  "inject after html when doctype and head missing",
			input: "<html>other stuff",
			expected: fmt.Sprintf(
				"<html>%sother stuff",
				reloadHTML,
			),
		},
		{
			name:  "ignore case",
			input: "<!DOCTYPE html><HTML><HEAD>other stuff",
			expected: fmt.Sprintf(
				"<!DOCTYPE html><HTML><HEAD>%sother stuff",
				reloadHTML,
			),
		},
		{
			name:  "skip comments",
			input: "<!doctype html><!-- comment --><html><head>other stuff",
			expected: fmt.Sprintf(
				"<!doctype html><!-- comment --><html><head>%sother stuff",
				reloadHTML,
			),
		},
		{
			name:  "skip whitespace",
			input: "<!doctype html>   \t  <html>   \t  <head>   \t  other stuff",
			expected: fmt.Sprintf(
				"<!doctype html>   \t  <html>   \t  <head>%s   \t  other stuff",
				reloadHTML,
			),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			result := responsebody.WithReloadHTML([]byte(test.input))
			if string(result) != test.expected {
				t.Errorf("expected %q, got %q", test.expected, result)
			}
		})
	}
}
