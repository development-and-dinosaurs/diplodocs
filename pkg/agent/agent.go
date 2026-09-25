package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/development-and-dinosaurs/diplodocs/pkg/config"
	"github.com/development-and-dinosaurs/diplodocs/pkg/navigation"
)

// DocPage holds information about a page for LLM generation.
type DocPage struct {
	Title   string
	URL     string
	Content string
}

// GenerateLLMsTxt creates an llms.txt standard file for AI consumption.
func GenerateLLMsTxt(cfg *config.Config, pages []*navigation.NavItem) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s\n\n", cfg.Site.Name))
	if cfg.Site.Description != "" {
		sb.WriteString(fmt.Sprintf("> %s\n\n", cfg.Site.Description))
	}

	sb.WriteString("## Documentation Index\n\n")
	for _, page := range pages {
		sb.WriteString(fmt.Sprintf("- [%s](%s)\n", page.Title, page.URL))
	}

	sb.WriteString("\n## Full Documentation Context\n\n")
	sb.WriteString("- [Full Markdown Document](llms-full.txt): Complete concatenated documentation in a single file.\n")

	return sb.String()
}

// GenerateLLMsFullTxt aggregates all markdown pages into a single document.
func GenerateLLMsFullTxt(cfg *config.Config, docsDir string, pages []*navigation.NavItem) (string, error) {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s — Complete Documentation\n\n", cfg.Site.Name))
	if cfg.Site.Description != "" {
		sb.WriteString(fmt.Sprintf("> %s\n\n", cfg.Site.Description))
	}
	sb.WriteString("---\n\n")

	for _, page := range pages {
		fullPath := filepath.Join(docsDir, page.RelPath)
		raw, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		sb.WriteString(fmt.Sprintf("## %s\n\n", page.Title))
		sb.WriteString(fmt.Sprintf("*Source: %s*\n\n", page.RelPath))
		sb.WriteString(strings.TrimSpace(string(raw)))
		sb.WriteString("\n\n---\n\n")
	}

	return sb.String(), nil
}

// WriteAgentFiles writes llms.txt and llms-full.txt to the output directory.
func WriteAgentFiles(outDir string, cfg *config.Config, docsDir string, pages []*navigation.NavItem) error {
	llmsTxt := GenerateLLMsTxt(cfg, pages)
	if err := os.WriteFile(filepath.Join(outDir, "llms.txt"), []byte(llmsTxt), 0644); err != nil {
		return fmt.Errorf("writing llms.txt: %w", err)
	}

	fullTxt, err := GenerateLLMsFullTxt(cfg, docsDir, pages)
	if err != nil {
		return fmt.Errorf("generating llms-full.txt: %w", err)
	}
	if err := os.WriteFile(filepath.Join(outDir, "llms-full.txt"), []byte(fullTxt), 0644); err != nil {
		return fmt.Errorf("writing llms-full.txt: %w", err)
	}

	return nil
}
