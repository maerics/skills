// Bundle walking, frontmatter extraction, and markdown rendering for the
// viewer. Parsing here is deliberately lenient (best-effort display) rather
// than the strict conformance checking scripts/ does — a malformed or
// unconventional bundle should still render as much as it can, never error
// out or crash.
package main

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	gmhtml "github.com/yuin/goldmark/renderer/html"
	gmtext "github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

const (
	indexFile = "index.md"
	logFile   = "log.md"
)

// FileMeta is the best-effort-parsed frontmatter of a single concept,
// index, or log file, enough to display in navigation and page headers.
type FileMeta struct {
	RelPath   string // forward-slash path relative to the bundle root
	Name      string // base filename
	Title     string
	Type      string
	Tags      []string
	Resource  string
	Timestamp string
}

// DirNode is one directory of the bundle tree, used to render the sidebar.
type DirNode struct {
	RelPath  string // forward-slash path relative to the bundle root; "" for root
	Name     string
	Index    *FileMeta
	Log      *FileMeta
	Concepts []*FileMeta
	Dirs     []*DirNode
}

// resolveBundleRoot resolves path to a bundle root directory: if
// <path>/.okf is a directory, that is the bundle root; otherwise path
// itself is treated as the bundle root. Mirrors scripts/bundle.go.
func resolveBundleRoot(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", path)
	}
	okfDir := filepath.Join(path, ".okf")
	if st, err := os.Stat(okfDir); err == nil && st.IsDir() {
		return okfDir, nil
	}
	return path, nil
}

// buildTree walks root into an in-memory DirNode tree. Rebuilt per request
// by callers that need a fresh view of the bundle — cheap enough for the
// bundle sizes this tool targets, and avoids file-watching complexity.
func buildTree(root string) (*DirNode, error) {
	return buildDir(root, "")
}

func buildDir(root, relDir string) (*DirNode, error) {
	abs := filepath.Join(root, filepath.FromSlash(relDir))
	entries, err := os.ReadDir(abs)
	if err != nil {
		return nil, err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })

	node := &DirNode{RelPath: relDir, Name: filepath.Base(abs)}
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		relPath := name
		if relDir != "" {
			relPath = relDir + "/" + name
		}
		if e.IsDir() {
			child, err := buildDir(root, relPath)
			if err != nil {
				return nil, err
			}
			node.Dirs = append(node.Dirs, child)
			continue
		}
		if !strings.HasSuffix(name, ".md") {
			continue
		}
		meta, err := readMeta(root, relPath)
		if err != nil {
			continue // unreadable file: skip it, don't fail the whole tree
		}
		switch name {
		case indexFile:
			node.Index = meta
		case logFile:
			node.Log = meta
		default:
			node.Concepts = append(node.Concepts, meta)
		}
	}
	return node, nil
}

func readMeta(root, relPath string) (*FileMeta, error) {
	raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relPath)))
	if err != nil {
		return nil, err
	}
	fm, _ := splitFrontmatter(raw)
	meta := &FileMeta{RelPath: relPath, Name: filepath.Base(relPath)}
	if fm != nil {
		vals, tags := parseFrontmatterFlat(fm)
		meta.Type = vals["type"]
		meta.Title = vals["title"]
		meta.Resource = vals["resource"]
		meta.Timestamp = vals["timestamp"]
		meta.Tags = tags
	}
	if meta.Title == "" {
		meta.Title = meta.Name
	}
	return meta, nil
}

// splitFrontmatter splits raw into a leading YAML frontmatter block (nil if
// absent or never closed) and the remaining body.
func splitFrontmatter(raw []byte) (fm, body []byte) {
	lines := strings.Split(string(raw), "\n")
	if len(lines) == 0 || strings.TrimRight(lines[0], "\r") != "---" {
		return nil, raw
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimRight(lines[i], "\r") == "---" {
			return []byte(strings.Join(lines[1:i], "\n")), []byte(strings.Join(lines[i+1:], "\n"))
		}
	}
	// Opened but never closed: lenient fallback, treat as no frontmatter.
	return nil, raw
}

var fmKeyRE = regexp.MustCompile(`^([A-Za-z0-9_-]+):(.*)$`)

// parseFrontmatterFlat scans a frontmatter block for flat "key: value"
// pairs and, specifically, a "tags" list (inline "[a, b]" or a block of
// "- item" lines). It is intentionally minimal: the viewer only needs
// enough to label pages, not to validate the bundle.
func parseFrontmatterFlat(fm []byte) (vals map[string]string, tags []string) {
	vals = map[string]string{}
	lastKey := ""
	for _, line := range strings.Split(string(fm), "\n") {
		trimmed := strings.TrimRight(line, "\r")
		stripped := strings.TrimSpace(trimmed)
		if stripped == "" || strings.HasPrefix(stripped, "#") {
			continue
		}
		if !strings.HasPrefix(trimmed, " ") && !strings.HasPrefix(trimmed, "\t") {
			m := fmKeyRE.FindStringSubmatch(trimmed)
			if m == nil {
				continue
			}
			key := strings.ToLower(m[1])
			val := strings.TrimSpace(m[2])
			vals[key] = strings.Trim(val, `"'`)
			lastKey = key
			if key == "tags" && val != "" {
				tags = parseInlineList(val)
			}
			continue
		}
		if lastKey != "tags" {
			continue
		}
		item := strings.TrimSpace(trimmed)
		item = strings.TrimPrefix(item, "-")
		item = strings.Trim(strings.TrimSpace(item), `"'`)
		if item != "" {
			tags = append(tags, item)
		}
	}
	return vals, tags
}

func parseInlineList(v string) []string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "[")
	v = strings.TrimSuffix(v, "]")
	if v == "" {
		return nil
	}
	var out []string
	for _, p := range strings.Split(v, ",") {
		p = strings.Trim(strings.TrimSpace(p), `"'`)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// renderMarkdown renders relPath's body to HTML, rewriting in-bundle
// relative links to viewer routes and flagging links that don't resolve.
func renderMarkdown(root, relPath string) (template.HTML, *FileMeta, error) {
	raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relPath)))
	if err != nil {
		return "", nil, err
	}
	fm, body := splitFrontmatter(raw)
	meta := &FileMeta{RelPath: relPath, Name: filepath.Base(relPath)}
	if fm != nil {
		vals, tags := parseFrontmatterFlat(fm)
		meta.Type = vals["type"]
		meta.Title = vals["title"]
		meta.Resource = vals["resource"]
		meta.Timestamp = vals["timestamp"]
		meta.Tags = tags
	}
	if meta.Title == "" {
		meta.Title = meta.Name
	}

	curDir := path.Dir(filepath.ToSlash(relPath))
	if curDir == "." {
		curDir = ""
	}

	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithASTTransformers(util.Prioritized(&linkTransformer{root: root, curDir: curDir}, 100)),
		),
		goldmark.WithRendererOptions(
			gmhtml.WithUnsafe(),
			renderer.WithNodeRenderers(util.Prioritized(fencedCodeRenderer{}, 10)),
		),
	)

	var buf bytes.Buffer
	if err := md.Convert(body, &buf); err != nil {
		return "", meta, err
	}
	return template.HTML(buf.String()), meta, nil
}

// linkTransformer rewrites markdown link destinations after parsing:
// external links get target=_blank; relative links that resolve inside the
// bundle become /view/<relpath> routes; anything else (a link that escapes
// the bundle root, or simply doesn't resolve) is defanged rather than
// dropped, per SPEC §9 tolerating broken links.
type linkTransformer struct {
	root   string // bundle root, absolute filesystem path
	curDir string // forward-slash dir of the file being rendered, relative to root
}

var externalLinkRE = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.-]*:`)

func (t *linkTransformer) Transform(doc *ast.Document, reader gmtext.Reader, pc parser.Context) {
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if link, ok := n.(*ast.Link); ok {
			t.rewrite(link)
		}
		return ast.WalkContinue, nil
	})
}

func (t *linkTransformer) rewrite(link *ast.Link) {
	dest := string(link.Destination)
	if dest == "" || strings.HasPrefix(dest, "#") {
		return
	}
	if externalLinkRE.MatchString(dest) {
		link.SetAttributeString("target", []byte("_blank"))
		link.SetAttributeString("rel", []byte("noopener"))
		return
	}

	target, frag := dest, ""
	if i := strings.IndexByte(target, '#'); i >= 0 {
		frag = target[i:]
		target = target[:i]
	}

	var resolved string
	if strings.HasPrefix(target, "/") {
		resolved = strings.TrimPrefix(target, "/")
	} else {
		resolved = path.Join(t.curDir, target)
	}
	resolved = path.Clean(resolved)
	if resolved == "." {
		resolved = ""
	}
	if resolved == ".." || strings.HasPrefix(resolved, "../") {
		t.markBroken(link)
		return
	}

	abs := filepath.Join(t.root, filepath.FromSlash(resolved))
	info, err := os.Stat(abs)
	if err != nil {
		t.markBroken(link)
		return
	}
	if info.IsDir() {
		idx := filepath.Join(abs, indexFile)
		if st, err2 := os.Stat(idx); err2 != nil || st.IsDir() {
			t.markBroken(link)
			return
		}
		resolved = path.Join(resolved, indexFile)
	}

	link.Destination = []byte("/view/" + resolved + frag)
}

func (t *linkTransformer) markBroken(link *ast.Link) {
	link.Destination = []byte("#")
	link.SetAttributeString("class", []byte("broken-link"))
}

// fencedCodeRenderer overrides goldmark's default fenced-code-block
// rendering so a ```mermaid block becomes <pre class="mermaid"> (what
// mermaid.js's default selector expects) instead of <pre><code
// class="language-mermaid">. Every other language renders the same as
// goldmark's default.
type fencedCodeRenderer struct{}

func (fencedCodeRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindFencedCodeBlock, renderFencedCodeBlock)
}

func renderFencedCodeBlock(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	node := n.(*ast.FencedCodeBlock)
	language := string(node.Language(source))
	isMermaid := language == "mermaid"

	switch {
	case isMermaid:
		_, _ = w.WriteString(`<pre class="mermaid">`)
	case language != "":
		_, _ = w.WriteString(`<pre><code class="language-`)
		_, _ = w.WriteString(language)
		_, _ = w.WriteString(`">`)
	default:
		_, _ = w.WriteString(`<pre><code>`)
	}

	lines := node.Lines()
	for i := 0; i < lines.Len(); i++ {
		line := lines.At(i)
		_, _ = w.Write(util.EscapeHTML(line.Value(source)))
	}

	if isMermaid {
		_, _ = w.WriteString("</pre>\n")
	} else {
		_, _ = w.WriteString("</code></pre>\n")
	}
	return ast.WalkContinue, nil
}
