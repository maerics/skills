package main

import (
	"strings"
	"testing"
)

const testBundleDir = "testdata/bundle"

func testRoot(t *testing.T) string {
	t.Helper()
	root, err := resolveBundleRoot(testBundleDir)
	if err != nil {
		t.Fatalf("resolveBundleRoot: %v", err)
	}
	return root
}

func TestResolveBundleRoot(t *testing.T) {
	root := testRoot(t)
	if !strings.HasSuffix(root, "/.okf") {
		t.Fatalf("expected root to resolve to the .okf subdirectory, got %q", root)
	}
}

func TestBuildTree(t *testing.T) {
	root := testRoot(t)
	tree, err := buildTree(root)
	if err != nil {
		t.Fatalf("buildTree: %v", err)
	}

	if tree.Index == nil || tree.Index.Title != "Test Bundle" {
		t.Fatalf("expected root index with title %q, got %+v", "Test Bundle", tree.Index)
	}
	if got, want := tree.Index.Tags, []string{"alpha", "beta"}; !equalStrings(got, want) {
		t.Fatalf("root tags = %v, want %v", got, want)
	}
	if tree.Log == nil {
		t.Fatal("expected root log.md to be found")
	}
	if len(tree.Dirs) != 1 || tree.Dirs[0].Name != "widgets" {
		t.Fatalf("expected exactly one child dir \"widgets\", got %+v", tree.Dirs)
	}

	widgets := tree.Dirs[0]
	if widgets.Index == nil {
		t.Fatal("expected widgets/index.md to be found")
	}
	if widgets.Log != nil {
		t.Fatal("did not expect a widgets/log.md")
	}
	if len(widgets.Concepts) != 1 || widgets.Concepts[0].Title != "Gadget" {
		t.Fatalf("expected one concept titled Gadget, got %+v", widgets.Concepts)
	}
	gadget := widgets.Concepts[0]
	if got, want := gadget.Tags, []string{"hardware", "widget"}; !equalStrings(got, want) {
		t.Fatalf("gadget tags = %v, want %v (block-list frontmatter parsing)", got, want)
	}
	if gadget.Resource != "https://example.com/gadget" {
		t.Fatalf("gadget resource = %q", gadget.Resource)
	}
}

func TestRenderMarkdownLinkClassification(t *testing.T) {
	root := testRoot(t)
	html, meta, err := renderMarkdown(root, "widgets/gadget.md")
	if err != nil {
		t.Fatalf("renderMarkdown: %v", err)
	}
	if meta.Title != "Gadget" {
		t.Fatalf("meta.Title = %q", meta.Title)
	}
	out := string(html)

	if !strings.Contains(out, `href="https://example.com"`) || !strings.Contains(out, `target="_blank"`) {
		t.Errorf("expected external link to keep its destination and open in a new tab, got: %s", out)
	}
	if !strings.Contains(out, `class="broken-link"`) {
		t.Errorf("expected the link to ./missing.md to be flagged broken-link, got: %s", out)
	}
	if !strings.Contains(out, `href="/view/widgets/index.md"`) {
		t.Errorf("expected the link to ./index.md to resolve to /view/widgets/index.md, got: %s", out)
	}
}

func TestRenderMarkdownMermaidPassthrough(t *testing.T) {
	root := testRoot(t)
	html, _, err := renderMarkdown(root, "index.md")
	if err != nil {
		t.Fatalf("renderMarkdown: %v", err)
	}
	out := string(html)

	if !strings.Contains(out, `<pre class="mermaid">`) {
		t.Errorf("expected a mermaid fenced block to render as <pre class=\"mermaid\">, got: %s", out)
	}
	if !strings.Contains(out, "flowchart LR") {
		t.Errorf("expected the mermaid source text to pass through unescaped-content, got: %s", out)
	}
	if !strings.Contains(out, `href="/view/widgets/index.md"`) {
		t.Errorf("expected the Sections link to resolve, got: %s", out)
	}
}

func TestSafeJoinContainsTraversal(t *testing.T) {
	root := testRoot(t)
	cases := []string{
		"../../../../etc/passwd",
		"widgets/../../../../etc/passwd",
		"/../../etc/passwd",
	}
	for _, c := range cases {
		abs, _, err := safeJoin(root, c)
		if err != nil {
			t.Errorf("safeJoin(%q) unexpectedly errored: %v", c, err)
			continue
		}
		if !strings.HasPrefix(abs, root+"/") && abs != root {
			t.Errorf("safeJoin(%q) = %q, escapes bundle root %q", c, abs, root)
		}
	}

	abs, rel, err := safeJoin(root, "widgets/gadget.md")
	if err != nil {
		t.Fatalf("safeJoin: unexpected error for an in-bundle path: %v", err)
	}
	if rel != "widgets/gadget.md" {
		t.Fatalf("rel = %q, want widgets/gadget.md", rel)
	}
	if !strings.HasSuffix(abs, "/widgets/gadget.md") {
		t.Fatalf("abs = %q", abs)
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
