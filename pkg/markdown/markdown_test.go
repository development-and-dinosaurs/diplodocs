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
