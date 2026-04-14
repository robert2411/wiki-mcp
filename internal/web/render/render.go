// Package render converts wiki markdown pages to HTML.
//
// It supports GFM extensions, [[wikilinks]], frontmatter metadata blocks,
// and strips .md from standard markdown link destinations.
package render

import (
	"bytes"
	"fmt"
	"html"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/robertstevens/wiki-mcp/internal/config"
	"github.com/robertstevens/wiki-mcp/internal/wiki"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	goldmarkHTML "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// RenderedPage holds the result of rendering a single wiki page.
type RenderedPage struct {
	HTML     string
	Title    string
	Metadata map[string]interface{}
	RelPath  string
}

// Renderer holds a pre-built title index and renders pages consistently.
type Renderer struct {
	titleIndex map[string]string // normalised-title → rel-path-without-.md
	md         goldmark.Markdown
}

// NewRenderer builds a Renderer for the given wiki root.
func NewRenderer(wikiPath string) (*Renderer, error) {
	idx, err := BuildTitleIndex(wikiPath)
	if err != nil {
		return nil, err
	}
	r := &Renderer{titleIndex: idx}
	r.md = buildGoldmark(r)
	return r, nil
}

// newRendererWithIndex builds a Renderer with a supplied index (for tests).
func newRendererWithIndex(idx map[string]string) *Renderer {
	r := &Renderer{titleIndex: idx}
	r.md = buildGoldmark(r)
	return r
}

func buildGoldmark(r *Renderer) goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Typographer,
			extension.Footnote,
			&wikilinkExt{renderer: r},
		),
		goldmark.WithParserOptions(
			parser.WithASTTransformers(
				util.Prioritized(&mdLinkTransformer{}, 999),
			),
		),
		goldmark.WithRendererOptions(
			goldmarkHTML.WithUnsafe(),
		),
	)
}

// RenderPage renders a single page. Builds a fresh title index on each call;
// use NewRenderer directly when rendering multiple pages.
func RenderPage(c *config.Config, relPath string) (*RenderedPage, error) {
	rdr, err := NewRenderer(c.WikiPath)
	if err != nil {
		return nil, err
	}
	absPath, err := c.ResolveWikiPath(relPath)
	if err != nil {
		return nil, err
	}
	return rdr.renderFile(absPath, relPath)
}

// RenderFile renders the page at absPath, using relPath for metadata and error messages.
func (r *Renderer) RenderFile(absPath, relPath string) (*RenderedPage, error) {
	return r.renderFile(absPath, relPath)
}

func (r *Renderer) renderFile(absPath, relPath string) (*RenderedPage, error) {
	raw, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("render %s: %w", relPath, err)
	}
	return r.renderBytes(raw, relPath)
}

func (r *Renderer) renderBytes(raw []byte, relPath string) (*RenderedPage, error) {
	meta, bodyStr := wiki.ParseFrontmatter(raw)
	if meta == nil {
		meta = map[string]any{}
	}
	body := []byte(bodyStr)

	var buf bytes.Buffer
	if err := r.md.Convert(body, &buf); err != nil {
		return nil, fmt.Errorf("goldmark convert %s: %w", relPath, err)
	}

	title := extractH1(body)
	if title == "" {
		title = filenameTitle(relPath)
	}

	var out strings.Builder
	if len(meta) > 0 {
		out.WriteString(renderFrontmatter(meta))
	}
	out.WriteString(buf.String())

	return &RenderedPage{
		HTML:     out.String(),
		Title:    title,
		Metadata: meta,
		RelPath:  relPath,
	}, nil
}

// BuildTitleIndex walks wikiPath and builds a map of normalised title → URL path.
// The URL path is the relative path with the .md extension stripped.
func BuildTitleIndex(wikiPath string) (map[string]string, error) {
	idx := map[string]string{}
	err := filepath.WalkDir(wikiPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		rel, err := filepath.Rel(wikiPath, path)
		if err != nil {
			return err
		}
		urlPath := strings.TrimSuffix(rel, ".md")

		title := h1FromFile(path)
		if title == "" {
			title = filenameTitle(rel)
		}
		idx[normaliseTitle(title)] = urlPath
		return nil
	})
	return idx, err
}

// h1FromFile reads up to 8 KB of a file to extract its H1 heading.
func h1FromFile(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	limited, err := io.ReadAll(io.LimitReader(f, 8192))
	if err != nil {
		return ""
	}
	_, bodyStr := wiki.ParseFrontmatter(limited)
	return extractH1([]byte(bodyStr))
}

var h1Re = regexp.MustCompile(`(?m)^#\s+(.+)$`)

func extractH1(md []byte) string {
	m := h1Re.FindSubmatch(md)
	if m == nil {
		return ""
	}
	return strings.TrimSpace(string(m[1]))
}

func filenameTitle(rel string) string {
	base := filepath.Base(rel)
	return strings.TrimSuffix(base, ".md")
}

func normaliseTitle(t string) string {
	return strings.ToLower(strings.TrimSpace(t))
}

func renderFrontmatter(meta map[string]interface{}) string {
	var b strings.Builder
	b.WriteString(`<div class="frontmatter">`)

	if v, ok := meta["tags"]; ok {
		b.WriteString(`<div class="frontmatter-tags">`)
		switch tv := v.(type) {
		case []interface{}:
			for _, tag := range tv {
				b.WriteString(`<span class="tag">`)
				b.WriteString(html.EscapeString(fmt.Sprintf("%v", tag)))
				b.WriteString(`</span>`)
			}
		default:
			b.WriteString(`<span class="tag">`)
			b.WriteString(html.EscapeString(fmt.Sprintf("%v", tv)))
			b.WriteString(`</span>`)
		}
		b.WriteString(`</div>`)
	}

	for k, v := range meta {
		if k == "tags" {
			continue
		}
		b.WriteString(`<div class="frontmatter-field"><span class="frontmatter-key">`)
		b.WriteString(html.EscapeString(k))
		b.WriteString(`</span><span class="frontmatter-value">`)
		b.WriteString(html.EscapeString(fmt.Sprintf("%v", v)))
		b.WriteString(`</span></div>`)
	}

	b.WriteString(`</div>`)
	return b.String()
}

// ── Wikilink extension ────────────────────────────────────────────────────────

var wikilinkKind = ast.NewNodeKind("Wikilink")

type wikilinkNode struct {
	ast.BaseInline
	target string
	alias  string
}

func (n *wikilinkNode) Kind() ast.NodeKind { return wikilinkKind }
func (n *wikilinkNode) Dump(src []byte, level int) {
	ast.DumpHelper(n, src, level, map[string]string{"target": n.target, "alias": n.alias}, nil)
}

type wikilinkExt struct {
	renderer *Renderer
}

func (e *wikilinkExt) Extend(md goldmark.Markdown) {
	md.Parser().AddOptions(
		parser.WithInlineParsers(
			util.Prioritized(&wikilinkParser{}, 100),
		),
	)
	md.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(&wikilinkRenderer{r: e.renderer}, 100),
		),
	)
}

type wikilinkParser struct{}

func (p *wikilinkParser) Trigger() []byte { return []byte{'['} }

func (p *wikilinkParser) Parse(_ ast.Node, block text.Reader, _ parser.Context) ast.Node {
	line, _ := block.PeekLine()
	if len(line) < 4 || line[0] != '[' || line[1] != '[' {
		return nil
	}
	closeIdx := bytes.Index(line[2:], []byte("]]"))
	if closeIdx < 0 {
		return nil
	}
	inner := string(line[2 : closeIdx+2])
	block.Advance(closeIdx + 4)
	target, alias, _ := strings.Cut(inner, "|")
	return &wikilinkNode{target: strings.TrimSpace(target), alias: strings.TrimSpace(alias)}
}

type wikilinkRenderer struct {
	r *Renderer
}

func (wr *wikilinkRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(wikilinkKind, wr.renderWikilink)
}

func (wr *wikilinkRenderer) renderWikilink(w util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*wikilinkNode)
	display := n.alias
	if display == "" {
		display = n.target
	}
	urlPath, ok := wr.r.titleIndex[normaliseTitle(n.target)]
	if !ok {
		_, _ = fmt.Fprintf(w, `<span class="broken-link">%s</span>`, html.EscapeString(display))
		return ast.WalkContinue, nil
	}
	_, _ = fmt.Fprintf(w, `<a href="/%s">%s</a>`, html.EscapeString(urlPath), html.EscapeString(display))
	return ast.WalkContinue, nil
}

// ── .md link-stripping transformer ────────────────────────────────────────────

type mdLinkTransformer struct{}

func (t *mdLinkTransformer) Transform(doc *ast.Document, _ text.Reader, _ parser.Context) {
	ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		link, ok := node.(*ast.Link)
		if !ok {
			return ast.WalkContinue, nil
		}
		dest := string(link.Destination)
		if strings.HasSuffix(dest, ".md") && !strings.HasPrefix(dest, "http") {
			link.Destination = []byte(strings.TrimSuffix(dest, ".md"))
		}
		return ast.WalkContinue, nil
	})
}
