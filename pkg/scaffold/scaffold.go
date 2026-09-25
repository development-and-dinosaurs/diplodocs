package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func b(s string) string {
	return strings.ReplaceAll(s, "§", "`")
}

// Project scaffolds a starter diplodocs project.
func Project(targetDir string) error {
	if targetDir == "" {
		targetDir = "."
	}

	projectName := filepath.Base(targetDir)
	if projectName == "." || projectName == "/" {
		projectName = "Diplodocs Site"
	}

	// 1. Create directory structure
	docsDir := filepath.Join(targetDir, "docs")
	getStartedDir := filepath.Join(docsDir, "01-get-started")
	guidesDir := filepath.Join(docsDir, "02-guides")

	for _, d := range []string{getStartedDir, guidesDir} {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("creating directory %s: %w", d, err)
		}
	}

	// 2. diplodocs.toml
	tomlContent := fmt.Sprintf(`[site]
name = "%s"
description = "High performance documentation built with Diplodocs"
author = "Diplodocus"
repo_url = "https://github.com/diplodocs/diplodocs"

[theme]
accent = "emerald"      # emerald, ocean, amber, rose, purple
dark_mode = "auto"       # auto, light, dark
logo_text = "🦕 %s"

[features]
search = true
mermaid = true
copy_button = true
llms_txt = true
`, projectName, projectName)

	tomlPath := filepath.Join(targetDir, "diplodocs.toml")
	if _, err := os.Stat(tomlPath); os.IsNotExist(err) {
		if err := os.WriteFile(tomlPath, []byte(tomlContent), 0644); err != nil {
			return err
		}
	}

	// 3. docs/index.md
	indexContent := b(`# Welcome to Diplodocs 🦕

> **Diplodocs** is an opinionated, batteries-included documentation engine.
> Colossal power, zero configuration.

---

## Why Diplodocs?

Some documentation engines require wrangling Python environments, pip packages, virtualenvs, and 60+ lines of YAML configuration just to get a basic site running.

Diplodocs replaces the dependency sprawl and plugin clutter with a single fast binary — zero Python runtime, zero npm packages, and zero boilerplate:

- 🚀 **Zero-Config Defaults:** Point Diplodocs at your §docs/§ folder and start.
- ⚡ **Native Modern Markdown:** GitHub callouts, code tabs, and mermaid diagrams work without third-party plugins.
- 📂 **Filesystem Navigation:** Automatic section hierarchies with natural ordering.
- 🤖 **Agent-Ready (§/llms.txt§):** Generates structured context files for AI coding assistants.
- 🔍 **Instant Offline Search:** Pre-indexed fuzzy search out of the box with §Ctrl+K§.

---

## Interactive Feature Preview

### Callouts / Admonitions

> [!NOTE]
> This note rendered with standard GitHub Markdown syntax. No special plugins required.

> [!TIP]
> Prefer three-bang syntax? Diplodocs also natively supports §!!! note "Title"§ admonitions without needing third-party extensions.

> [!WARNING]
> Breaking changes or critical warnings grab attention immediately.

### Code Tabs

=== "Go"

    §§§go
    package main

    import "fmt"

    func main() {
        fmt.Println("Hello from Diplodocs! 🦕")
    }
    §§§

=== "Python"

    §§§python
    def main():
        print("Hello from Diplodocs! 🦕")

    if __name__ == "__main__":
        main()
    §§§

=== "JavaScript"

    §§§javascript
    console.log("Hello from Diplodocs! 🦕");
    §§§
`)
	_ = os.WriteFile(filepath.Join(docsDir, "index.md"), []byte(indexContent), 0644)

	// 4. docs/01-get-started/01-installation.md
	installContent := b(`# Installation

Diplodocs is distributed as a single static binary. You don't need Python or Node.js installed to use it.

## Quick Install

§§§bash
# Download and install diplodocs binary
go install diplodocs/cmd/diplodocs@latest
§§§

## Verify Installation

§§§bash
diplodocs version
§§§
`)
	_ = os.WriteFile(filepath.Join(getStartedDir, "01-installation.md"), []byte(installContent), 0644)

	// 5. docs/01-get-started/02-quickstart.md
	quickstartContent := b(`# Quickstart Guide

Get up and running with a complete documentation site in seconds.

## 1. Create a Project

§§§bash
diplodocs new my-docs
cd my-docs
§§§

## 2. Start Dev Server

§§§bash
diplodocs dev
§§§

Open [http://localhost:8080](http://localhost:8080) in your browser. Any edit to §docs/§ triggers an instant hot-reload.

## 3. Build for Production

§§§bash
diplodocs build
§§§

Your static documentation site is ready in §dist/§. Deploy it to GitHub Pages, Netlify, Vercel, or S3.
`)
	_ = os.WriteFile(filepath.Join(getStartedDir, "02-quickstart.md"), []byte(quickstartContent), 0644)

	// 6. docs/02-guides/01-diagrams.md
	diagramsContent := b(`# Diagrams & Visuals

Diplodocs includes native support for Mermaid.js diagrams.

## Flowchart Example

§§§mermaid
graph TD
    A[Markdown Docs] --> B(Diplodocs Engine)
    B --> C{Output}
    C --> D[Static HTML]
    C --> E[Search Index]
    C --> F[llms.txt for AI]
§§§

## Architecture Sequence

§§§mermaid
sequenceDiagram
    autonumber
    User->>Diplodocs: diplodocs dev
    Diplodocs->>Browser: Serve on localhost:8080
    User->>Editor: Edit docs/guide.md
    Diplodocs->>Browser: SSE live reload
§§§
`)
	_ = os.WriteFile(filepath.Join(guidesDir, "01-diagrams.md"), []byte(diagramsContent), 0644)

	return nil
}
