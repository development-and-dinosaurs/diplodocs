package markdown

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// Matches > [!NOTE] or > [!NOTE] Custom Title
	githubCalloutRegex = regexp.MustCompile(`^>\s*\[!(NOTE|TIP|IMPORTANT|WARNING|CAUTION)\](?:\s+(.*))?$`)

	// Matches !!! note "Optional Title"
	materialCalloutRegex = regexp.MustCompile(`^(!{3}|\?{3})\s+(\w+)(?:\s+"([^"]+)")?\s*$`)

	// Matches === "Tab Title"
	tabHeaderRegex = regexp.MustCompile(`^={3}\s+"([^"]+)"\s*$`)

	// Matches ![Alt](src){ .class }
	imgAttrRegex = regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)\{\s*\.([a-zA-Z0-9_-]+)\s*\}`)
)

type calloutMeta struct {
	kind  string
	title string
	icon  string
}

func getCalloutMeta(rawType string, customTitle string) calloutMeta {
	k := strings.ToLower(rawType)
	title := customTitle
	icon := "ℹ️"

	switch k {
	case "note", "info":
		k = "note"
		if title == "" {
			title = "Note"
		}
		icon = `<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="16" x2="12" y2="12"></line><line x1="12" y1="8" x2="12.01" y2="8"></line></svg>`
	case "tip", "success", "hint":
		k = "tip"
		if title == "" {
			title = "Tip"
		}
		icon = `<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"></path></svg>`
	case "important":
		k = "important"
		if title == "" {
			title = "Important"
		}
		icon = `<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>`
	case "warning":
		k = "warning"
		if title == "" {
			title = "Warning"
		}
		icon = `<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z"></path><line x1="12" y1="9" x2="12" y2="13"></line><line x1="12" y1="17" x2="12.01" y2="17"></line></svg>`
	case "caution", "danger", "error":
		k = "danger"
		if title == "" {
			title = "Caution"
		}
		icon = `<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="7.86 2 16.14 2 22 7.86 22 16.14 16.14 22 7.86 22 2 16.14 2 7.86 7.86 2"></polygon><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>`
	default:
		k = "note"
		if title == "" {
			title = strings.ToUpper(rawType[:1]) + rawType[1:]
		}
		icon = `<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="16" x2="12" y2="12"></line><line x1="12" y1="8" x2="12.01" y2="8"></line></svg>`
	}

	return calloutMeta{kind: k, title: title, icon: icon}
}

// PreprocessMarkdown handles GitHub Callouts, Material Admonitions, and Code Tabs.
func PreprocessMarkdown(content string) string {
	content = imgAttrRegex.ReplaceAllString(content, `<img src="$2" alt="$1" class="$3">`)
	lines := strings.Split(content, "\n")
	var result []string
	n := len(lines)
	i := 0

	for i < n {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// 1. Check for GitHub Callout: > [!NOTE]
		if matches := githubCalloutRegex.FindStringSubmatch(trimmed); len(matches) > 0 {
			rawType := matches[1]
			customTitle := strings.TrimSpace(matches[2])
			meta := getCalloutMeta(rawType, customTitle)

			// Collect subsequent lines belonging to this blockquote
			var innerLines []string
			i++
			for i < n {
				cur := lines[i]
				curTrimmed := strings.TrimSpace(cur)
				if githubCalloutRegex.MatchString(curTrimmed) {
					// New callout starts here
					break
				}
				if strings.HasPrefix(curTrimmed, ">") {
					stripped := strings.TrimPrefix(curTrimmed, ">")
					stripped = strings.TrimPrefix(stripped, " ")
					innerLines = append(innerLines, stripped)
					i++
				} else {
					break
				}
			}

			innerContent := strings.Join(innerLines, "\n")
			// Recursively preprocess inside callout
			innerContent = PreprocessMarkdown(innerContent)

			calloutHTML := fmt.Sprintf(
				"<div class=\"callout callout-%s\">\n<div class=\"callout-header\"><span class=\"callout-icon\">%s</span><span class=\"callout-title\">%s</span></div>\n<div class=\"callout-body\">\n\n%s\n\n</div>\n</div>\n",
				meta.kind, meta.icon, meta.title, innerContent,
			)
			result = append(result, calloutHTML)
			continue
		}

		// 2. Check for Material Admonition: !!! note "Optional Title"
		if matches := materialCalloutRegex.FindStringSubmatch(trimmed); len(matches) > 0 {
			rawType := matches[2]
			customTitle := matches[3]
			meta := getCalloutMeta(rawType, customTitle)

			var innerLines []string
			i++
			for i < n {
				cur := lines[i]
				// Must be indented by 4 spaces or tab
				if strings.HasPrefix(cur, "    ") {
					innerLines = append(innerLines, strings.TrimPrefix(cur, "    "))
					i++
				} else if strings.HasPrefix(cur, "\t") {
					innerLines = append(innerLines, strings.TrimPrefix(cur, "\t"))
					i++
				} else if strings.TrimSpace(cur) == "" {
					// Allow empty line if next line is indented
					if i+1 < n && (strings.HasPrefix(lines[i+1], "    ") || strings.HasPrefix(lines[i+1], "\t")) {
						innerLines = append(innerLines, "")
						i++
					} else {
						break
					}
				} else {
					break
				}
			}

			innerContent := strings.Join(innerLines, "\n")
			innerContent = PreprocessMarkdown(innerContent)

			calloutHTML := fmt.Sprintf(
				"<div class=\"callout callout-%s\">\n<div class=\"callout-header\"><span class=\"callout-icon\">%s</span><span class=\"callout-title\">%s</span></div>\n<div class=\"callout-body\">\n\n%s\n\n</div>\n</div>\n",
				meta.kind, meta.icon, meta.title, innerContent,
			)
			result = append(result, calloutHTML)
			continue
		}

		// 3. Check for Code Tabs: === "Title"
		if matches := tabHeaderRegex.FindStringSubmatch(trimmed); len(matches) > 0 {
			type tabItem struct {
				title   string
				content string
			}
			var tabs []tabItem

			for i < n {
				tMatch := tabHeaderRegex.FindStringSubmatch(strings.TrimSpace(lines[i]))
				if len(tMatch) == 0 {
					break
				}
				tabTitle := tMatch[1]
				i++

				var tabLines []string
				for i < n {
					cur := lines[i]
					if strings.HasPrefix(cur, "    ") {
						tabLines = append(tabLines, strings.TrimPrefix(cur, "    "))
						i++
					} else if strings.HasPrefix(cur, "\t") {
						tabLines = append(tabLines, strings.TrimPrefix(cur, "\t"))
						i++
					} else if strings.TrimSpace(cur) == "" {
						if i+1 < n && (strings.HasPrefix(lines[i+1], "    ") || strings.HasPrefix(lines[i+1], "\t") || tabHeaderRegex.MatchString(strings.TrimSpace(lines[i+1]))) {
							tabLines = append(tabLines, "")
							i++
						} else {
							break
						}
					} else {
						break
					}
				}

				tContent := strings.Join(tabLines, "\n")
				tContent = PreprocessMarkdown(tContent)
				tabs = append(tabs, tabItem{title: tabTitle, content: tContent})

				// Skip empty lines between tabs
				for i < n && strings.TrimSpace(lines[i]) == "" {
					if i+1 < n && tabHeaderRegex.MatchString(strings.TrimSpace(lines[i+1])) {
						i++
					} else {
						break
					}
				}
			}

			if len(tabs) > 0 {
				var tabsHTML strings.Builder
				tabsHTML.WriteString("<div class=\"code-tabs\">\n<div class=\"tab-headers\" role=\"tablist\">\n")
				for idx, tab := range tabs {
					activeClass := ""
					if idx == 0 {
						activeClass = " active"
					}
					tabsHTML.WriteString(fmt.Sprintf(
						"<button class=\"tab-button%s\" role=\"tab\" onclick=\"diplodocsSwitchTab(this, %d)\">%s</button>\n",
						activeClass, idx, tab.title,
					))
				}
				tabsHTML.WriteString("</div>\n<div class=\"tab-panels\">\n")
				for idx, tab := range tabs {
					activeClass := ""
					if idx == 0 {
						activeClass = " active"
					}
					tabsHTML.WriteString(fmt.Sprintf(
						"<div class=\"tab-panel%s\" role=\"tabpanel\">\n\n%s\n\n</div>\n",
						activeClass, tab.content,
					))
				}
				tabsHTML.WriteString("</div>\n</div>\n")
				result = append(result, tabsHTML.String())
				continue
			}
		}

		result = append(result, line)
		i++
	}

	return strings.Join(result, "\n")
}
