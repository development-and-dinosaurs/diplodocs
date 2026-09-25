package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/development-and-dinosaurs/diplodocs/pkg/builder"
	"github.com/development-and-dinosaurs/diplodocs/pkg/config"
	"github.com/development-and-dinosaurs/diplodocs/pkg/navigation"
	"github.com/development-and-dinosaurs/diplodocs/pkg/scaffold"
	"github.com/development-and-dinosaurs/diplodocs/pkg/server"
)

const version = "0.1.0"

const dinoBanner = `
  __               _ _       _                 
 / /_   __ _ _   _(_) | ___ | | ___   __ _ ___ 
| '_ \ / _' | | | | | |/ _ \| |/ _ \ / _' / __|
| (_) | (_| | |_| | | | (_) | | (_) | (_| \__ \
 \___/ \__,_|\__,_|_|_|\___/|_|\___/ \__, |___/
                                     |___/     
        Colossal Power, Zero Configuration 🦕
`

func printHelp() {
	fmt.Print(dinoBanner)
	fmt.Print(`Usage:
  diplodocs <command> [arguments]

Commands:
  new [dir]      Scaffold a new documentation project with sample docs
  dev            Start local development server with instant live reload
  build          Compile documentation into high performance static HTML
  check          Scan documentation for broken internal links and images
  migrate [dir]  Convert an existing MkDocs project (mkdocs.yml) to Diplodocs
  version        Print Diplodocs version

Flags for 'dev':
  --port <int>   Port to listen on (default: 8080)

Flags for 'build':
  --dir <string> Root project directory (default: current dir)

Examples:
  diplodocs new my-docs
  diplodocs dev
  diplodocs build
`)
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(0)
	}

	command := os.Args[1]

	switch command {
	case "version", "-v", "--version":
		fmt.Printf("diplodocs version %s\n", version)

	case "help", "-h", "--help":
		printHelp()

	case "new":
		targetDir := "."
		if len(os.Args) >= 3 {
			targetDir = os.Args[2]
		}
		fmt.Printf("🦕 Scaffolding new Diplodocs project in '%s'...\n", targetDir)
		if err := scaffold.Project(targetDir); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error scaffolding project: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ Project ready! Get started by running:\n\n")
		if targetDir != "." {
			fmt.Printf("   cd %s\n", targetDir)
		}
		fmt.Printf("   diplodocs dev\n\n")

	case "dev":
		devCmd := flag.NewFlagSet("dev", flag.ExitOnError)
		port := devCmd.Int("port", 8080, "Dev server port")
		dir := devCmd.String("dir", ".", "Project root directory")
		_ = devCmd.Parse(os.Args[2:])

		fmt.Print(dinoBanner)
		srv, err := server.NewDevServer(*dir, *port)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to start dev server: %v\n", err)
			os.Exit(1)
		}
		if err := srv.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Server error: %v\n", err)
			os.Exit(1)
		}

	case "build":
		buildCmd := flag.NewFlagSet("build", flag.ExitOnError)
		dir := buildCmd.String("dir", ".", "Project root directory")
		_ = buildCmd.Parse(os.Args[2:])

		fmt.Println("🦕 Diplodocs building static site...")
		stats, err := builder.Build(*dir, false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Build error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✨ Built %d pages to '%s' in %v\n", stats.PageCount, stats.OutDir, stats.Duration)

	case "check":
		checkCmd := flag.NewFlagSet("check", flag.ExitOnError)
		dir := checkCmd.String("dir", ".", "Project root directory")
		_ = checkCmd.Parse(os.Args[2:])

		runCheck(*dir)

	case "migrate":
		migrateCmd := flag.NewFlagSet("migrate", flag.ExitOnError)
		dir := migrateCmd.String("dir", ".", "Project root directory containing mkdocs.yml")
		_ = migrateCmd.Parse(os.Args[2:])

		targetDir := *dir
		if len(migrateCmd.Args()) > 0 {
			targetDir = migrateCmd.Args()[0]
		}

		fmt.Printf("🦕 Searching for MkDocs configuration in '%s'...\n", targetDir)
		var mkPath string
		for _, name := range []string{"mkdocs.yml", "mkdocs.yaml"} {
			p := filepath.Join(targetDir, name)
			if _, err := os.Stat(p); err == nil {
				mkPath = p
				break
			}
		}
		if mkPath == "" {
			fmt.Fprintf(os.Stderr, "❌ No mkdocs.yml or mkdocs.yaml found in '%s'\n", targetDir)
			os.Exit(1)
		}

		cfg, err := config.LoadFromMkDocs(mkPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to parse %s: %v\n", mkPath, err)
			os.Exit(1)
		}

		data, err := config.ExportTOML(cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to generate TOML: %v\n", err)
			os.Exit(1)
		}

		targetToml := filepath.Join(targetDir, "diplodocs.toml")
		if err := os.WriteFile(targetToml, data, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to write %s: %v\n", targetToml, err)
			os.Exit(1)
		}

		fmt.Printf("✨ Successfully migrated '%s' to '%s'!\n", mkPath, targetToml)
		fmt.Printf("   Site: %s\n", cfg.Site.Name)
		if len(cfg.Nav) > 0 {
			fmt.Printf("   Navigation: %d top-level items migrated\n", len(cfg.Nav))
		}
		if len(cfg.ExtraCSS) > 0 {
			fmt.Printf("   Extra CSS: %s\n", strings.Join(cfg.ExtraCSS, ", "))
		}
		fmt.Printf("\nYou can now run:\n   diplodocs dev --dir %s\n", targetDir)

	default:
		fmt.Fprintf(os.Stderr, "Unknown command '%s'. Run 'diplodocs help' for usage.\n", command)
		os.Exit(1)
	}
}

var mdLinkRegex = regexp.MustCompile(`\[.*?\]\((.*?)\)`)

func runCheck(projectRoot string) {
	fmt.Println("🦕 Checking documentation links and integrity...")
	cfg, err := config.LoadConfig(projectRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error loading config: %v\n", err)
		os.Exit(1)
	}

	docsDir := filepath.Join(projectRoot, cfg.DocsDir)
	_, flatPages, err := navigation.BuildNavTree(docsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error scanning docs: %v\n", err)
		os.Exit(1)
	}

	brokenLinks := 0
	totalLinks := 0

	for _, page := range flatPages {
		filePath := filepath.Join(docsDir, page.RelPath)
		content, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		matches := mdLinkRegex.FindAllStringSubmatch(string(content), -1)
		for _, m := range matches {
			rawLink := m[1]
			// Ignore web URLs, mailto, and anchor links
			if strings.HasPrefix(rawLink, "http://") || strings.HasPrefix(rawLink, "https://") || strings.HasPrefix(rawLink, "mailto:") || strings.HasPrefix(rawLink, "#") {
				continue
			}

			totalLinks++
			// Clean hash fragment from relative link
			targetPath := strings.Split(rawLink, "#")[0]
			if targetPath == "" {
				continue
			}

			// Target relative to page directory
			resolved := filepath.Join(filepath.Dir(filePath), targetPath)
			if _, err := os.Stat(resolved); os.IsNotExist(err) {
				// Try with .md extension if missing
				if _, err2 := os.Stat(resolved + ".md"); os.IsNotExist(err2) {
					fmt.Printf("⚠️  Broken link in %s -> '%s'\n", page.RelPath, rawLink)
					brokenLinks++
				}
			}
		}
	}

	if brokenLinks == 0 {
		fmt.Printf("✅ Checked %d internal links across %d pages. No broken links found!\n", totalLinks, len(flatPages))
	} else {
		fmt.Printf("❌ Found %d broken links!\n", brokenLinks)
		os.Exit(1)
	}
}
