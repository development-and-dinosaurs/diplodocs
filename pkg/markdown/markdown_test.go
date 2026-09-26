package markdown

import (
	"strings"
	"testing"
)

func TestPreprocessorCallouts(t *testing.T) {
	input := `
> [!NOTE]
> This is a crucial note for the user.
`
	output := PreprocessMarkdown(input)
	if !strings.Contains(output, "callout-note") {
		t.Errorf("expected callout-note in output, got:\n%s", output)
	}
	if !strings.Contains(output, "This is a crucial note") {
		t.Errorf("expected body text in output, got:\n%s", output)
	}
}

func TestPreprocessorTabs(t *testing.T) {
	input := `
=== "Go"
    package main
    func main() {}

=== "Python"
    print("hello")
`
	output := PreprocessMarkdown(input)
	if !strings.Contains(output, "code-tabs") {
		t.Errorf("expected code-tabs in output, got:\n%s", output)
	}
	if !strings.Contains(output, "Go") || !strings.Contains(output, "Python") {
		t.Errorf("expected tab titles, got:\n%s", output)
	}
}

func TestEngineRender(t *testing.T) {
	engine := NewEngine()
	src := `---
title: "Custom Title"
---

# Page Header

Welcome to Diplodocs.

## Getting Started

Here is how you start.

` + "```go\nfunc Hello() {}\n```" + `

` + "```mermaid\ngraph TD;\nA-->B;\n```"

	res, err := engine.Render([]byte(src), "Fallback")
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	if res.Title != "Custom Title" {
		t.Errorf("expected Title 'Custom Title', got '%s'", res.Title)
	}

	if len(res.TOC) != 1 {
		t.Fatalf("expected 1 TOC item, got %d", len(res.TOC))
	}
	if res.TOC[0].Title != "Getting Started" {
		t.Errorf("expected TOC title 'Getting Started', got '%s'", res.TOC[0].Title)
	}

	if !strings.Contains(res.HTML, "mermaid") {
		t.Errorf("expected mermaid container in output HTML")
	}

	if !strings.Contains(res.HTML, "code-block") {
		t.Errorf("expected code-block in output HTML")
	}

	if !strings.Contains(res.PlainTxt, "Welcome to Diplodocs") {
		t.Errorf("expected clean plain text, got:\n%s", res.PlainTxt)
	}
}

func TestRewriteMarkdownLink(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"quick-start.md", "quick-start.html"},
		{"quick-start.markdown", "quick-start.html"},
		{"quick-start.md#usage", "quick-start.html#usage"},
		{"./cli.md?v=2#flags", "./cli.html?v=2#flags"},
		{"../guides/diagrams.md", "../guides/diagrams.html"},
		{"index.md", "index.html"},
		{"https://github.com/foo/bar.md", "https://github.com/foo/bar.md"},
		{"http://example.com/test.md#anchor", "http://example.com/test.md#anchor"},
		{"mailto:user@domain.md", "mailto:user@domain.md"},
		{"#heading-only", "#heading-only"},
		{"image.png", "image.png"},
		{"archive.tar.gz", "archive.tar.gz"},
	}

	for _, tt := range tests {
		got := RewriteMarkdownLink(tt.input)
		if got != tt.expected {
			t.Errorf("RewriteMarkdownLink(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}

func TestMarkdownLinkRewritingEndToEnd(t *testing.T) {
	engine := NewEngine()
	src := `Check out the [Gradle Plugin](gradle-plugin.md), the [CLI Guide](cli.md#flags), or [External](https://example.com/doc.md).
Also raw HTML: <a href="library.md#api">Library API</a>.`

	res, err := engine.Render([]byte(src), "Links")
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	if !strings.Contains(res.HTML, `href="gradle-plugin.html"`) {
		t.Errorf("expected gradle-plugin.html in HTML, got:\n%s", res.HTML)
	}
	if !strings.Contains(res.HTML, `href="cli.html#flags"`) {
		t.Errorf("expected cli.html#flags in HTML, got:\n%s", res.HTML)
	}
	if !strings.Contains(res.HTML, `href="https://example.com/doc.md"`) {
		t.Errorf("expected untouched external link in HTML, got:\n%s", res.HTML)
	}
	if !strings.Contains(res.HTML, `href="library.html#api"`) {
		t.Errorf("expected raw HTML rewritten to library.html#api, got:\n%s", res.HTML)
	}
}
