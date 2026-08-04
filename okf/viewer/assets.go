// Static assets embedded into the binary so the viewer works fully
// offline with no build step or external requests.
package main

import (
	"embed"
	"io/fs"
)

//go:embed static/style.css static/client.js static/mermaid.min.js
var embeddedStatic embed.FS

var staticFS fs.FS

func init() {
	sub, err := fs.Sub(embeddedStatic, "static")
	if err != nil {
		panic(err)
	}
	staticFS = sub
}
