// HTML templates for the viewer's page shell and recursive sidebar tree.
package main

import (
	"html/template"
	"strings"
)

type pageData struct {
	Title      string
	BundleName string
	Meta       *FileMeta
	Body       template.HTML
	Tree       *DirNode
	Current    string
	Breadcrumb []breadcrumbItem
}

// treeCtx carries the current page's path alongside each subtree so the
// recursive "tree" template can mark the active entry at any depth.
type treeCtx struct {
	Node    *DirNode
	Current string
}

func mustParsePageTemplate() *template.Template {
	funcs := template.FuncMap{
		"subtree": func(n *DirNode, current string) treeCtx { return treeCtx{Node: n, Current: current} },
		"stem":    func(filename string) string { return strings.TrimSuffix(filename, ".md") },
	}
	t := template.Must(template.New("page").Funcs(funcs).Parse(pageHTML))
	t = template.Must(t.Parse(treeHTML))
	return t
}

const pageHTML = `{{define "page"}}<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Title}} &middot; {{.BundleName}}</title>
<script>(function(){try{var t=localStorage.getItem("okf-theme");if(t)document.documentElement.setAttribute("data-theme",t);}catch(e){}})();</script>
<link rel="stylesheet" href="/static/style.css">
</head>
<body>
<button class="theme-toggle" id="theme-toggle" type="button" aria-label="Toggle light/dark theme">
  <svg class="icon-sun" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><circle cx="12" cy="12" r="4"/><path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M6.34 17.66l-1.41 1.41M19.07 4.93l-1.41 1.41"/></svg>
  <svg class="icon-moon" viewBox="0 0 24 24" fill="currentColor"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/></svg>
</button>
<div class="layout">
  <nav class="sidebar">
    <div class="sidebar-header"><a href="/">{{.BundleName}}</a></div>
    {{template "tree" (subtree .Tree .Current)}}
  </nav>
  <main>
    {{if .Breadcrumb}}
    <nav class="breadcrumb">
      {{range .Breadcrumb}}{{if .Href}}<a href="{{.Href}}">{{.Name}}</a>{{else}}<span>{{.Name}}</span>{{end}} / {{end}}
    </nav>
    {{end}}
    {{if .Meta}}
    <header class="meta">
      <h1>{{.Meta.Title}}</h1>
      {{if or .Meta.Type .Meta.Resource .Meta.Timestamp .Meta.Tags}}
      <dl class="frontmatter">
        {{if .Meta.Type}}<dt>type</dt><dd>{{.Meta.Type}}</dd>{{end}}
        {{if .Meta.Resource}}<dt>resource</dt><dd>{{.Meta.Resource}}</dd>{{end}}
        {{if .Meta.Timestamp}}<dt>timestamp</dt><dd>{{.Meta.Timestamp}}</dd>{{end}}
        {{if .Meta.Tags}}<dt>tags</dt><dd>{{range .Meta.Tags}}<span class="tag">{{.}}</span>{{end}}</dd>{{end}}
      </dl>
      {{end}}
    </header>
    {{else}}
    <header class="meta"><h1>{{.Title}}</h1></header>
    {{end}}
    <article>{{.Body}}</article>
  </main>
</div>
<script src="/static/mermaid.min.js"></script>
<script src="/static/client.js"></script>
</body>
</html>
{{end}}`

const treeHTML = `{{define "tree"}}<ul class="tree">
  {{with .Node.Index}}<li class="entry entry-index{{if eq .RelPath $.Current}} active{{end}}"><a href="/view/{{.RelPath}}">{{if $.Node.RelPath}}{{$.Node.Name}}{{else}}Index{{end}}</a></li>{{end}}
  {{range .Node.Concepts}}<li class="entry entry-concept{{if eq .RelPath $.Current}} active{{end}}"><a href="/view/{{.RelPath}}" title="{{.Title}}">{{stem .Name}}</a></li>{{end}}
  {{range .Node.Dirs}}<li class="entry entry-dir"><details open><summary>{{.Name}}/</summary>{{template "tree" (subtree . $.Current)}}</details></li>{{end}}
  {{with .Node.Log}}<li class="entry entry-log{{if eq .RelPath $.Current}} active{{end}}"><a href="/view/{{.RelPath}}">[log]</a></li>{{end}}
</ul>{{end}}`
