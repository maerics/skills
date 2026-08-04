// HTTP routes for the viewer: page rendering with sidebar navigation, raw
// asset serving, and the path-traversal guard that keeps both under the
// bundle root.
package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type server struct {
	root string
	lc   *lifecycle
	tmpl *template.Template
}

func newServer(root string, lc *lifecycle) *server {
	return &server{root: root, lc: lc, tmpl: mustParsePageTemplate()}
}

func (s *server) mux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRoot)
	mux.HandleFunc("/view/", s.handleView)
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) { serveWS(w, r, s.lc) })
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	return mux
}

func (s *server) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if _, err := os.Stat(filepath.Join(s.root, indexFile)); err == nil {
		http.Redirect(w, r, "/view/"+indexFile, http.StatusFound)
		return
	}
	// No root index.md: synthesize a bare listing from the tree itself so
	// the bundle is still navigable, per SKILL.md's "if absent, synthesize
	// a listing" guidance.
	tree, err := buildTree(s.root)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.render(w, pageData{
		Title:      s.bundleName(),
		BundleName: s.bundleName(),
		Tree:       tree,
	})
}

func (s *server) handleView(w http.ResponseWriter, r *http.Request) {
	reqPath := strings.TrimPrefix(r.URL.Path, "/view/")
	abs, rel, err := safeJoin(s.root, reqPath)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	info, err := os.Stat(abs)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if info.IsDir() {
		if _, err := os.Stat(filepath.Join(abs, indexFile)); err == nil {
			http.Redirect(w, r, "/view/"+path.Join(rel, indexFile), http.StatusFound)
			return
		}
		http.NotFound(w, r)
		return
	}
	if !strings.HasSuffix(rel, ".md") {
		http.ServeFile(w, r, abs)
		return
	}
	s.renderPage(w, rel)
}

func (s *server) renderPage(w http.ResponseWriter, relPath string) {
	body, meta, err := renderMarkdown(s.root, relPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tree, err := buildTree(s.root)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.render(w, pageData{
		Title:      meta.Title,
		BundleName: s.bundleName(),
		Meta:       meta,
		Body:       body,
		Tree:       tree,
		Current:    relPath,
		Breadcrumb: buildBreadcrumb(s.root, relPath),
	})
}

func (s *server) render(w http.ResponseWriter, data pageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmpl.ExecuteTemplate(w, "page", data); err != nil {
		log.Println("okf-viewer: template error:", err)
	}
}

func (s *server) bundleName() string {
	base := filepath.Base(s.root)
	if base == ".okf" {
		return filepath.Base(filepath.Dir(s.root))
	}
	return base
}

// safeJoin resolves reqPath against root, guaranteeing the result cannot
// escape root via ".." components. Returns both the absolute filesystem
// path and the cleaned bundle-relative (forward-slash) path.
func safeJoin(root, reqPath string) (abs, rel string, err error) {
	reqPath = strings.TrimPrefix(reqPath, "/")
	cleaned := strings.TrimPrefix(path.Clean("/"+reqPath), "/")
	if cleaned == "." {
		cleaned = ""
	}
	abs = filepath.Join(root, filepath.FromSlash(cleaned))
	relCheck, err := filepath.Rel(root, abs)
	if err != nil || relCheck == ".." || strings.HasPrefix(relCheck, ".."+string(filepath.Separator)) {
		return "", "", fmt.Errorf("path escapes bundle root")
	}
	return abs, cleaned, nil
}

type breadcrumbItem struct {
	Name string
	Href string
}

// buildBreadcrumb turns "tables/orders.md" into Index > tables > (current
// page, not linked). An ancestor only gets a link if its index.md actually
// exists, so the breadcrumb never points at a dead route.
func buildBreadcrumb(root, relPath string) []breadcrumbItem {
	items := []breadcrumbItem{{Name: "index", Href: "/"}}
	parts := strings.Split(relPath, "/")
	cum := ""
	for i, part := range parts {
		if i == len(parts)-1 {
			break // last segment is the page itself
		}
		if cum == "" {
			cum = part
		} else {
			cum = cum + "/" + part
		}
		item := breadcrumbItem{Name: part}
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(cum), indexFile)); err == nil {
			item.Href = "/view/" + cum + "/" + indexFile
		}
		items = append(items, item)
	}
	return items
}
