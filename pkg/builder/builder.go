package builder

import (
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/development-and-dinosaurs/diplodocs/pkg/agent"
	"github.com/development-and-dinosaurs/diplodocs/pkg/config"
	"github.com/development-and-dinosaurs/diplodocs/pkg/markdown"
	"github.com/development-and-dinosaurs/diplodocs/pkg/navigation"
	"github.com/development-and-dinosaurs/diplodocs/pkg/search"
	"github.com/development-and-dinosaurs/diplodocs/pkg/theme"
)

// BuildStats captures performance and file statistics.
type BuildStats struct {
	PageCount int
	Duration  time.Duration
	OutDir    string
}

// Compute relative root path based on URL nesting level.
func computeRootPath(url string) string {
	dir := filepath.Dir(url)
	if dir == "." || dir == "" {
		return ""
	}
	parts := strings.Split(filepath.ToSlash(dir), "/")
	var up []string
	for range parts {
		up = append(up, "..")
	}
	return strings.Join(up, "/") + "/"
}

// CopyDir recursively copies a directory tree.
func copyDir(src string, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	for _, entry := range entries {
		s := filepath.Join(src, entry.Name())
		d := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			if err := copyDir(s, d); err != nil {
				return err
			}
		} else {
			in, err := os.Open(s)
			if err != nil {
				return err
			}
			defer in.Close()

			out, err := os.Create(d)
			if err != nil {
				return err
			}
			defer out.Close()

			if _, err := io.Copy(out, in); err != nil {
				return err
			}
		}
	}
	return nil
}

func extractGitHubRepo(repoURL string) string {
	u := strings.TrimSuffix(repoURL, "/")
	u = strings.TrimSuffix(u, ".git")
	idx := strings.Index(u, "github.com/")
	if idx != -1 {
		return u[idx+len("github.com/"):]
	}
	return ""
}

// Build compiles the documentation site from projectRoot into cfg.OutDir.
func Build(projectRoot string, isDev bool) (*BuildStats, error) {
	start := time.Now()

	cfg, err := config.LoadConfig(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	repoOwnerRepo := extractGitHubRepo(cfg.Site.RepoURL)

	docsDir := filepath.Join(projectRoot, cfg.DocsDir)
	if _, err := os.Stat(docsDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("docs directory '%s' does not exist. Run 'diplodocs new' to scaffold a project", docsDir)
	}

	outDir := filepath.Join(projectRoot, cfg.OutDir)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return nil, fmt.Errorf("creating output directory: %w", err)
	}

	// 1. Initialize engines
	themeMgr, err := theme.NewManager()
	if err != nil {
		return nil, fmt.Errorf("initializing theme: %w", err)
	}

	mdEngine := markdown.NewEngine()
	searchIdx := search.NewIndex()

	// 2. Extract theme assets
	if err := themeMgr.WriteAssets(outDir); err != nil {
		return nil, fmt.Errorf("writing theme assets: %w", err)
	}

	// 3. Copy user static files and directories (skip .md and .markdown)
	_ = filepath.Walk(docsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || path == docsDir {
			return nil
		}
		rel, err := filepath.Rel(docsDir, path)
		if err != nil || strings.HasPrefix(filepath.Base(path), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		dest := filepath.Join(outDir, rel)
		if info.IsDir() {
			return os.MkdirAll(dest, 0755)
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".md" || ext == ".markdown" {
			return nil
		}

		in, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer in.Close()

		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return nil
		}

		out, err := os.Create(dest)
		if err != nil {
			return nil
		}
		defer out.Close()

		_, _ = io.Copy(out, in)
		return nil
	})

	// 4. Scan navigation (use config.Nav if defined, else scan filesystem)
	var navTree *navigation.NavItem
	var flatPages []*navigation.NavItem
	if len(cfg.Nav) > 0 {
		navTree, flatPages, err = navigation.BuildNavTreeFromConfig(docsDir, cfg.Nav)
	} else {
		navTree, flatPages, err = navigation.BuildNavTree(docsDir)
	}
	if err != nil {
		return nil, fmt.Errorf("building navigation tree: %w", err)
	}

	if len(flatPages) == 0 {
		return nil, fmt.Errorf("no markdown files found in '%s'", docsDir)
	}

	// 5. Render each page
	for _, page := range flatPages {
		srcPath := filepath.Join(docsDir, page.RelPath)
		rawBytes, err := os.ReadFile(srcPath)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", srcPath, err)
		}

		res, err := mdEngine.Render(rawBytes, page.Title)
		if err != nil {
			return nil, fmt.Errorf("rendering %s: %w", srcPath, err)
		}

		// Update page title if changed by frontmatter/H1
		page.Title = res.Title

		// Register into search index
		searchIdx.Add(res.Title, page.URL, filepath.Dir(page.RelPath), res.PlainTxt)

		// Destination HTML file
		destPath := filepath.Join(outDir, page.URL)
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return nil, fmt.Errorf("creating page dir %s: %w", destPath, err)
		}

		destFile, err := os.Create(destPath)
		if err != nil {
			return nil, fmt.Errorf("creating %s: %w", destPath, err)
		}

		var editURL string
		if cfg.Features.ShowEditLink && repoOwnerRepo != "" {
			editURL = fmt.Sprintf("https://github.com/%s/edit/%s/%s/%s",
				repoOwnerRepo,
				cfg.Site.EditBranch,
				cfg.DocsDir,
				page.RelPath,
			)
		}

		pageData := &theme.PageData{
			Site:           cfg.Site,
			Theme:          cfg.Theme,
			Features:       cfg.Features,
			PageTitle:      res.Title,
			CurrentURL:     page.URL,
			Content:        template.HTML(res.HTML),
			TOC:            res.TOC,
			NavTree:        navTree,
			PrevPage:       page.Prev,
			NextPage:       page.Next,
			RootPath:       computeRootPath(page.URL),
			IsDev:          isDev,
			RepoOwnerRepo:  repoOwnerRepo,
			EditURL:        editURL,
			ReadingTimeMin: res.ReadingTimeMin,
			WordCount:      res.WordCount,
			ExtraCSS:       cfg.ExtraCSS,
			ExtraJS:        cfg.ExtraJS,
		}

		if err := themeMgr.Render(destFile, pageData); err != nil {
			destFile.Close()
			return nil, fmt.Errorf("rendering template for %s: %w", page.URL, err)
		}
		destFile.Close()
	}

	// 6. Write search index if enabled
	if cfg.Features.Search {
		indexPath := filepath.Join(outDir, "search-index.json")
		if err := searchIdx.Save(indexPath); err != nil {
			return nil, fmt.Errorf("saving search index: %w", err)
		}
	}

	// 7. Write AI Agent context (llms.txt & llms-full.txt) if enabled
	if cfg.Features.LLMsTxt {
		if err := agent.WriteAgentFiles(outDir, cfg, docsDir, flatPages); err != nil {
			return nil, fmt.Errorf("writing agent context: %w", err)
		}
	}

	return &BuildStats{
		PageCount: len(flatPages),
		Duration:  time.Since(start),
		OutDir:    outDir,
	}, nil
}
