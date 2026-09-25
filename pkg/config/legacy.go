package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/development-and-dinosaurs/diplodocs/pkg/navigation"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

// ParseLegacyNav transforms legacy YAML navigation structures into NavItemConfig.
func ParseLegacyNav(items []interface{}) []navigation.NavItemConfig {
	var result []navigation.NavItemConfig
	for _, it := range items {
		switch v := it.(type) {
		case string:
			_, clean := navigation.CleanName(v)
			result = append(result, navigation.NavItemConfig{
				Title: clean,
				Path:  v,
			})
		case map[string]interface{}:
			for key, val := range v {
				title := key
				switch target := val.(type) {
				case string:
					result = append(result, navigation.NavItemConfig{
						Title: title,
						Path:  target,
					})
				case []interface{}:
					children := ParseLegacyNav(target)
					result = append(result, navigation.NavItemConfig{
						Title:    title,
						Children: children,
					})
				}
			}
		case map[interface{}]interface{}:
			for key, val := range v {
				title := fmt.Sprint(key)
				switch target := val.(type) {
				case string:
					result = append(result, navigation.NavItemConfig{
						Title: title,
						Path:  target,
					})
				case []interface{}:
					children := ParseLegacyNav(target)
					result = append(result, navigation.NavItemConfig{
						Title:    title,
						Children: children,
					})
				}
			}
		}
	}
	return result
}


// LoadFromLegacyYAML parses legacy YAML configuration and constructs a Config struct.
func LoadFromLegacyYAML(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", configPath, err)
	}

	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing YAML in %s: %w", configPath, err)
	}

	dirName := filepath.Dir(configPath)
	cfg := DefaultConfig(dirName)

	if name, ok := raw["site_name"].(string); ok && name != "" {
		cfg.Site.Name = name
		cfg.Theme.LogoText = "🦕 " + name
	}
	if desc, ok := raw["site_description"].(string); ok {
		cfg.Site.Description = desc
	}
	if repo, ok := raw["repo_url"].(string); ok {
		cfg.Site.RepoURL = repo
	}
	if url, ok := raw["site_url"].(string); ok {
		cfg.Site.BaseURL = url
	}
	if docsDir, ok := raw["docs_dir"].(string); ok && docsDir != "" {
		cfg.DocsDir = docsDir
	}
	if siteDir, ok := raw["site_dir"].(string); ok && siteDir != "" {
		cfg.OutDir = siteDir
	}

	// Extra CSS
	if cssList, ok := raw["extra_css"].([]interface{}); ok {
		for _, css := range cssList {
			if cssStr, ok := css.(string); ok && cssStr != "" {
				cfg.ExtraCSS = append(cfg.ExtraCSS, cssStr)
			}
		}
	}

	// Extra JS
	if jsList, ok := raw["extra_javascript"].([]interface{}); ok {
		for _, js := range jsList {
			if jsStr, ok := js.(string); ok && jsStr != "" {
				cfg.ExtraJS = append(cfg.ExtraJS, jsStr)
			}
		}
	}

	// Navigation
	if navList, ok := raw["nav"].([]interface{}); ok {
		cfg.Nav = ParseLegacyNav(navList)
	}

	return cfg, nil
}


// ExportTOML marshals a Config into formatted TOML bytes.
func ExportTOML(cfg *Config) ([]byte, error) {
	return toml.Marshal(cfg)
}
