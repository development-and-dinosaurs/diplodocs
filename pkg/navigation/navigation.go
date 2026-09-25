package navigation

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// NavItemConfig represents an explicit navigation item from config.
type NavItemConfig struct {
	Title    string          `toml:"title" yaml:"title"`
	Path     string          `toml:"path,omitempty" yaml:"path,omitempty"`
	Children []NavItemConfig `toml:"children,omitempty" yaml:"children,omitempty"`
}

// NavItem represents a single page or directory in the navigation tree.
type NavItem struct {
	Title    string     `json:"title"`
	RelPath  string     `json:"relPath"`  // e.g., "01-intro/02-quickstart.md"
	URL      string     `json:"url"`      // e.g., "01-intro/02-quickstart.html"
	IsIndex  bool       `json:"isIndex"`  // true if index.md or README.md
	Children []*NavItem `json:"children"` // sub-pages or sub-sections
	Prev     *NavItem   `json:"-"`
	Next     *NavItem   `json:"-"`
}

var orderPrefixRegex = regexp.MustCompile(`^(\d+)[-_.\s]+(.*)$`)

// CleanName removes leading numeric prefixes and converts kebab/snake-case to Title Case.
func CleanName(name string) (order int, cleanTitle string) {
	name = strings.TrimSuffix(name, filepath.Ext(name))
	order = 999999

	matches := orderPrefixRegex.FindStringSubmatch(name)
	if len(matches) == 3 {
		if num, err := strconv.Atoi(matches[1]); err == nil {
			order = num
		}
		name = matches[2]
	}

	words := strings.Fields(strings.ReplaceAll(strings.ReplaceAll(name, "-", " "), "_", " "))
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
		}
	}
	cleanTitle = strings.Join(words, " ")
	if cleanTitle == "" {
		cleanTitle = "Home"
	}
	return order, cleanTitle
}

// ExtractTitle inspects frontmatter or the first H1 header in a markdown file.
func ExtractTitle(filePath string, fallback string) string {
	f, err := os.Open(filePath)
	if err != nil {
		return fallback
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	inFrontmatter := false
	frontmatterChecked := false
	lineCount := 0

	for scanner.Scan() && lineCount < 60 {
		lineCount++
		line := strings.TrimSpace(scanner.Text())

		if line == "---" || line == "+++" {
			if !inFrontmatter && !frontmatterChecked && lineCount == 1 {
				inFrontmatter = true
				continue
			} else if inFrontmatter {
				inFrontmatter = false
				frontmatterChecked = true
				continue
			}
		}

		if inFrontmatter {
			if strings.HasPrefix(strings.ToLower(line), "title:") {
				val := strings.TrimSpace(line[6:])
				val = strings.Trim(val, `"'`)
				if val != "" {
					return val
				}
			}
			continue
		}

		// Check for Markdown H1: # Title
		if strings.HasPrefix(line, "# ") {
			title := strings.TrimSpace(line[2:])
			if title != "" {
				return title
			}
		}
	}

	return fallback
}

type fileEntry struct {
	isDir    bool
	name     string
	fullPath string
	order    int
	title    string
}

// BuildNavTree scans docsDir and constructs a sorted hierarchy of NavItems.
func BuildNavTree(docsDir string) (*NavItem, []*NavItem, error) {
	rootItem := &NavItem{
		Title:    "Documentation",
		RelPath:  "",
		URL:      "",
		Children: make([]*NavItem, 0),
	}

	var flatPages []*NavItem

	var walkDir func(currentPath string, relDir string) ([]*NavItem, error)
	walkDir = func(currentPath string, relDir string) ([]*NavItem, error) {
		entries, err := os.ReadDir(currentPath)
		if err != nil {
			return nil, err
		}

		var items []*fileEntry
		for _, entry := range entries {
			name := entry.Name()
			if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") {
				continue
			}

			full := filepath.Join(currentPath, name)
			if entry.IsDir() {
				order, title := CleanName(name)
				items = append(items, &fileEntry{
					isDir:    true,
					name:     name,
					fullPath: full,
					order:    order,
					title:    title,
				})
			} else {
				ext := strings.ToLower(filepath.Ext(name))
				if ext == ".md" || ext == ".markdown" {
					order, defaultTitle := CleanName(name)
					title := ExtractTitle(full, defaultTitle)
					items = append(items, &fileEntry{
						isDir:    false,
						name:     name,
						fullPath: full,
						order:    order,
						title:    title,
					})
				}
			}
		}

		// Sort by numeric prefix order, then alphabetically
		sort.SliceStable(items, func(i, j int) bool {
			// index.md or README.md always first in its folder
			iIsIndex := !items[i].isDir && (strings.EqualFold(items[i].name, "index.md") || strings.EqualFold(items[i].name, "readme.md"))
			jIsIndex := !items[j].isDir && (strings.EqualFold(items[j].name, "index.md") || strings.EqualFold(items[j].name, "readme.md"))
			if iIsIndex != jIsIndex {
				return iIsIndex
			}

			if items[i].order != items[j].order {
				return items[i].order < items[j].order
			}
			return items[i].name < items[j].name
		})

		var res []*NavItem
		for _, it := range items {
			if it.isDir {
				subRel := filepath.Join(relDir, it.name)
				children, err := walkDir(it.fullPath, subRel)
				if err != nil {
					return nil, err
				}
				if len(children) > 0 {
					sectionNode := &NavItem{
						Title:    it.title,
						RelPath:  subRel,
						Children: children,
					}
					// If the first child is an index, the section link can point to it
					for _, ch := range children {
						if ch.IsIndex {
							sectionNode.URL = ch.URL
							break
						}
					}
					res = append(res, sectionNode)
				}
			} else {
				relPath := filepath.Join(relDir, it.name)
				isIndex := strings.EqualFold(it.name, "index.md") || strings.EqualFold(it.name, "readme.md")

				// Compute URL
				urlPath := strings.TrimSuffix(relPath, filepath.Ext(relPath)) + ".html"
				urlPath = filepath.ToSlash(urlPath)

				node := &NavItem{
					Title:   it.title,
					RelPath: filepath.ToSlash(relPath),
					URL:     urlPath,
					IsIndex: isIndex,
				}
				res = append(res, node)
				flatPages = append(flatPages, node)
			}
		}

		return res, nil
	}

	children, err := walkDir(docsDir, "")
	if err != nil {
		return nil, nil, err
	}
	rootItem.Children = children

	// Link sequential Prev and Next pages
	for i := 0; i < len(flatPages); i++ {
		if i > 0 {
			flatPages[i].Prev = flatPages[i-1]
		}
		if i < len(flatPages)-1 {
			flatPages[i].Next = flatPages[i+1]
		}
	}

	return rootItem, flatPages, nil
}

// BuildNavTreeFromConfig builds navigation tree from explicit configuration items.
func BuildNavTreeFromConfig(docsDir string, configs []NavItemConfig) (*NavItem, []*NavItem, error) {
	rootItem := &NavItem{
		Title:    "Documentation",
		Children: make([]*NavItem, 0),
	}
	var flatPages []*NavItem

	var processItems func(items []NavItemConfig) ([]*NavItem, error)
	processItems = func(items []NavItemConfig) ([]*NavItem, error) {
		var res []*NavItem
		for _, it := range items {
			if len(it.Children) > 0 {
				children, err := processItems(it.Children)
				if err != nil {
					return nil, err
				}
				sectionNode := &NavItem{
					Title:    it.Title,
					Children: children,
				}
				for _, ch := range children {
					if ch.IsIndex || sectionNode.URL == "" {
						sectionNode.URL = ch.URL
					}
				}
				res = append(res, sectionNode)
			} else if it.Path != "" {
				fullPath := filepath.Join(docsDir, it.Path)
				title := it.Title
				if title == "" {
					_, defaultTitle := CleanName(it.Path)
					title = ExtractTitle(fullPath, defaultTitle)
				}
				isIndex := strings.EqualFold(filepath.Base(it.Path), "index.md") || strings.EqualFold(filepath.Base(it.Path), "readme.md")
				urlPath := strings.TrimSuffix(it.Path, filepath.Ext(it.Path)) + ".html"
				urlPath = filepath.ToSlash(urlPath)

				node := &NavItem{
					Title:   title,
					RelPath: filepath.ToSlash(it.Path),
					URL:     urlPath,
					IsIndex: isIndex,
				}
				res = append(res, node)
				flatPages = append(flatPages, node)
			}
		}
		return res, nil
	}

	children, err := processItems(configs)
	if err != nil {
		return nil, nil, err
	}
	rootItem.Children = children

	// Link sequential Prev and Next pages
	for i := 0; i < len(flatPages); i++ {
		if i > 0 {
			flatPages[i].Prev = flatPages[i-1]
		}
		if i < len(flatPages)-1 {
			flatPages[i].Next = flatPages[i+1]
		}
	}

	return rootItem, flatPages, nil
}
