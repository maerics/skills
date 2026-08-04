// HTML templates for the viewer's page shell and recursive sidebar tree.
package main

import "html/template"

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
<link rel="stylesheet" href="/static/style.css">
</head>
<body>
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
  {{with .Node.Log}}<li class="entry entry-log{{if eq .RelPath $.Current}} active{{end}}"><a href="/view/{{.RelPath}}">Log</a></li>{{end}}
  {{range .Node.Concepts}}<li class="entry entry-concept{{if eq .RelPath $.Current}} active{{end}}"><a href="/view/{{.RelPath}}">{{.Title}}</a></li>{{end}}
  {{range .Node.Dirs}}<li class="entry entry-dir"><details open><summary>{{.Name}}/</summary>{{template "tree" (subtree . $.Current)}}</details></li>{{end}}
</ul>{{end}}`
