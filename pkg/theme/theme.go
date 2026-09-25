package theme

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"

	"github.com/development-and-dinosaurs/diplodocs/pkg/config"
	"github.com/development-and-dinosaurs/diplodocs/pkg/markdown"
	"github.com/development-and-dinosaurs/diplodocs/pkg/navigation"
)

//go:embed templates/* assets/*
var themeFS embed.FS

// PageData is passed to the HTML template for rendering.
type PageData struct {
	Site           config.SiteConfig
	Theme          config.ThemeConfig
	Features       config.FeaturesConfig
	PageTitle      string
	CurrentURL     string
	Content        template.HTML
	TOC            []markdown.TOCItem
	NavTree        *navigation.NavItem
	PrevPage       *navigation.NavItem
	NextPage       *navigation.NavItem
	RootPath       string
	IsDev          bool
	RepoOwnerRepo  string
	EditURL        string
	ReadingTimeMin int
	WordCount      int
	ExtraCSS       []string
	ExtraJS        []string
}

// Manager handles template execution and asset distribution.
type Manager struct {
	tmpl *template.Template
}

// NewManager loads embedded templates.
func NewManager() (*Manager, error) {
	tmpl, err := template.ParseFS(themeFS, "templates/page.html")
	if err != nil {
		return nil, fmt.Errorf("parsing embedded template: %w", err)
	}
	return &Manager{tmpl: tmpl}, nil
}

// Render writes the populated HTML template to w.
func (m *Manager) Render(w io.Writer, data *PageData) error {
	return m.tmpl.Execute(w, data)
}

// WriteAssets extracts embedded CSS and JS files to destDir.
func (m *Manager) WriteAssets(destDir string) error {
	assetsDir := filepath.Join(destDir, "assets")
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		return fmt.Errorf("creating assets dir: %w", err)
	}

	assetFiles := []string{
		"assets/diplodocs.css",
		"assets/diplodocs.js",
	}

	for _, rel := range assetFiles {
		content, err := themeFS.ReadFile(rel)
		if err != nil {
			return fmt.Errorf("reading embedded asset %s: %w", rel, err)
		}
		target := filepath.Join(destDir, rel)
		if err := os.WriteFile(target, content, 0644); err != nil {
			return fmt.Errorf("writing asset %s: %w", target, err)
		}
	}

	return nil
}
