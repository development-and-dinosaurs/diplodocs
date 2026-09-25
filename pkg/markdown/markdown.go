package markdown

import (
	"bytes"
	"fmt"
	"html"
	"regexp"
	"strings"

	"github.com/alecthomas/chroma/v2"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	goldmarkhtml "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// TOCItem represents a section header for "On this page" navigation.
type TOCItem struct {
	Level int    `json:"level"`
	Title string `json:"title"`
	ID    string `json:"id"`
}

// RenderResult contains the generated HTML, page title, TOC, plain text, and metadata.
type RenderResult struct {
	HTML           string
	Title          string
	TOC            []TOCItem
	PlainTxt       string
	WordCount      int
	ReadingTimeMin int
	Meta           map[string]interface{}
}

// CustomCodeBlockRenderer intercepts fenced code blocks for Chroma highlighting & Mermaid.
type CustomCodeBlockRenderer struct {
	goldmarkhtml.Config
}

func (r *CustomCodeBlockRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindFencedCodeBlock, r.renderFencedCodeBlock)
}

func (r *CustomCodeBlockRenderer) renderFencedCodeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.FencedCodeBlock)
	if !entering {
		return ast.WalkContinue, nil
	}

	lang := string(n.Language(source))
	lang = strings.TrimSpace(lang)

	var code bytes.Buffer
	for i := 0; i < n.Lines().Len(); i++ {
		line := n.Lines().At(i)
		code.Write(line.Value(source))
	}
	codeStr := code.String()

	// Mermaid diagrams
	if strings.EqualFold(lang, "mermaid") {
		w.WriteString("<div class=\"mermaid-wrapper\"><pre class=\"mermaid\">")
		w.WriteString(html.EscapeString(codeStr))
		w.WriteString("</pre></div>\n")
		return ast.WalkContinue, nil
	}

	displayLang := lang
	if displayLang == "" {
		displayLang = "text"
	}

	// Syntax highlight with Chroma
	lexer := lexers.Get(lang)
	if lexer == nil {
		lexer = lexers.Analyse(codeStr)
	}
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	iterator, err := lexer.Tokenise(nil, codeStr)
	if err != nil {
		w.WriteString(fmt.Sprintf("<pre><code>%s</code></pre>", html.EscapeString(codeStr)))
		return ast.WalkContinue, nil
	}

	formatter := chromahtml.New(
		chromahtml.WithClasses(true),
		chromahtml.TabWidth(4),
	)

	var highlighted bytes.Buffer
	// Use fallback style as token class names are generated
	_ = formatter.Format(&highlighted, styles.Fallback, iterator)

	w.WriteString(fmt.Sprintf("<div class=\"code-block\" data-lang=\"%s\">\n", html.EscapeString(displayLang)))
	w.WriteString("  <div class=\"code-header\">\n")
	w.WriteString(fmt.Sprintf("    <span class=\"code-lang\">%s</span>\n", html.EscapeString(displayLang)))
	w.WriteString("    <button class=\"copy-button\" onclick=\"diplodocsCopyCode(this)\" title=\"Copy code\" aria-label=\"Copy code\">\n")
	w.WriteString("      <svg class=\"copy-icon\" width=\"15\" height=\"15\" viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><rect width=\"14\" height=\"14\" x=\"8\" y=\"8\" rx=\"2\" ry=\"2\"/><path d=\"M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2\"/></svg>\n")
	w.WriteString("      <span class=\"copy-text\">Copy</span>\n")
	w.WriteString("    </button>\n")
	w.WriteString("  </div>\n")
	w.WriteString("  <div class=\"code-content\">\n")
	w.WriteString(highlighted.String())
	w.WriteString("  </div>\n")
	w.WriteString("</div>\n")

	return ast.WalkContinue, nil
}

// Engine wraps goldmark and provides high-level markdown rendering.
type Engine struct {
	gm goldmark.Markdown
}

// NewEngine initializes the goldmark markdown engine.
func NewEngine() *Engine {
	gm := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Table,
			extension.Strikethrough,
			extension.TaskList,
			meta.Meta,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			goldmarkhtml.WithUnsafe(),
			renderer.WithNodeRenderers(
				util.Prioritized(&CustomCodeBlockRenderer{}, 100),
			),
		),
	)

	return &Engine{gm: gm}
}

var stripHTMLRegex = regexp.MustCompile(`<[^>]*>`)

// Render converts markdown source into HTML and extracts TOC and metadata.
func (e *Engine) Render(raw []byte, defaultTitle string) (*RenderResult, error) {
	// 1. Run Preprocessor for Callouts and Tabs
	preprocessed := PreprocessMarkdown(string(raw))
	source := []byte(preprocessed)

	// 2. Parse AST
	context := parser.NewContext()
	reader := text.NewReader(source)
	doc := e.gm.Parser().Parse(reader, parser.WithContext(context))

	// Extract TOC and primary title from AST
	var toc []TOCItem
	extractedTitle := ""

	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if heading, ok := n.(*ast.Heading); ok {
			var textBuf bytes.Buffer
			for child := heading.FirstChild(); child != nil; child = child.NextSibling() {
				if t, ok := child.(*ast.Text); ok {
					textBuf.Write(t.Segment.Value(source))
				} else if str, ok := child.(*ast.String); ok {
					textBuf.Write(str.Value)
				}
			}
			titleText := strings.TrimSpace(textBuf.String())

			if heading.Level == 1 && extractedTitle == "" {
				extractedTitle = titleText
			}

			if heading.Level >= 2 && heading.Level <= 3 {
				var idStr string
				if rawID, found := heading.AttributeString("id"); found {
					idStr = string(rawID.([]byte))
				}
				if idStr == "" {
					idStr = slugify(titleText)
				}
				toc = append(toc, TOCItem{
					Level: heading.Level,
					Title: titleText,
					ID:    idStr,
				})
			}
		}
		return ast.WalkContinue, nil
	})

	// 3. Render HTML
	var htmlBuf bytes.Buffer
	if err := e.gm.Renderer().Render(&htmlBuf, source, doc); err != nil {
		return nil, fmt.Errorf("rendering markdown: %w", err)
	}

	// Extract Frontmatter metadata
	metaData := meta.Get(context)
	if metaData != nil {
		if metaTitle, ok := metaData["title"].(string); ok && metaTitle != "" {
			extractedTitle = metaTitle
		}
	}

	finalTitle := extractedTitle
	if finalTitle == "" {
		finalTitle = defaultTitle
	}

	// Generate clean plain text for search indexing
	plain := stripHTMLRegex.ReplaceAllString(htmlBuf.String(), " ")
	words := strings.Fields(plain)
	wordCount := len(words)
	readingTime := (wordCount + 180) / 200
	if readingTime < 1 {
		readingTime = 1
	}
	plain = strings.Join(words, " ")

	return &RenderResult{
		HTML:           htmlBuf.String(),
		Title:          finalTitle,
		TOC:            toc,
		PlainTxt:       plain,
		WordCount:      wordCount,
		ReadingTimeMin: readingTime,
		Meta:           metaData,
	}, nil
}

var nonWordRegex = regexp.MustCompile(`[^\w\s-]`)
var whitespaceRegex = regexp.MustCompile(`[\s_]+`)

func slugify(s string) string {
	s = strings.ToLower(s)
	s = nonWordRegex.ReplaceAllString(s, "")
	s = whitespaceRegex.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}
