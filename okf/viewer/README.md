# okf-viewer

Serves an OKF (Open Knowledge Format) knowledge bundle as a local,
navigable website and opens it in a browser — with `index.md` and `log.md`
given special treatment (pinned, badged) but still listed like any other
file. The server manages its own lifecycle: it opens the browser itself,
and shuts itself down once no browser tab has it open, so there's nothing
to clean up manually.

## Usage

```
./view [-port N] [-no-open] [bundle-path]

Arguments:
  bundle-path    Bundle root, or a directory containing .okf/. Defaults to ".".

Flags:
  -port N        Port to listen on. Defaults to 0 (OS-assigned).
  -no-open       Don't launch a browser automatically.
```

`view` is a shell wrapper around `go run .` that resolves its own location,
so it works from anywhere: `<skill-base-dir>/viewer/view <bundle-path>`.
Equivalent without the wrapper: `go run .` from this directory, or
`go run <skill-base-dir>/viewer <bundle-path>` from anywhere.

The server binds to `127.0.0.1` only — never `0.0.0.0` — so it's never
reachable off the local machine; there's no authentication because none is
needed for a loopback-only server.

## Lifecycle

- The server picks an OS-assigned free port, prints its URL, and (unless
  `-no-open`) opens it in the default browser.
- Each open tab holds one websocket connection at `/ws` for as long as the
  page is open. When the last one closes, a short grace period (a few
  seconds, to ride out normal page-to-page navigation) runs before the
  server actually shuts down.
- If no browser ever connects at all — auto-open failed, a headless
  session, or the tab was simply never opened — a longer idle timeout (a
  few minutes) shuts the server down too, so a forgotten viewer never
  lingers as an orphan process.
- `Ctrl+C` (or `SIGTERM`) shuts it down immediately.

## Rendering

Markdown is rendered with [goldmark](https://github.com/yuin/goldmark)
(GFM tables, fenced code, autolinks, strikethrough) — the one dependency in
this module, and the one deliberate departure from the rest of this skill's
zero-dependency tooling; hand-rolling GFM-quality markdown rendering isn't
practical to do well. ```` ```mermaid ```` fenced blocks render as actual
diagrams via a vendored, embedded copy of
[mermaid.js](https://mermaid.js.org) (`static/mermaid.min.js`, bundled with
`go:embed` so the viewer works fully offline).

Frontmatter parsing here is deliberately lenient — best-effort display,
never rejects a bundle — unlike `../scripts/`'s strict conformance checks.
Run `../scripts/validate` separately if you want conformance checking.

Links are rewritten after parsing:

- External (`http(s)://`, `mailto:`, ...) links open in a new tab.
- Relative links that resolve to a file inside the bundle become in-app
  `/view/...` routes.
- Anything else — a broken link, or one that would escape the bundle root —
  is tolerated per SPEC §9: rendered as plain, inert text rather than a
  dead link, never a crash or a 404 page.

## Security

- Binds to `127.0.0.1` only.
- Every request path is resolved and checked against the bundle root before
  any file is read (`safeJoin` in `server.go`), so a request can't read
  files outside the bundle via `../` segments.
- Bundle content is rendered with `goldmark`'s unsafe-HTML passthrough
  enabled, on the assumption that a bundle's own markdown is trusted
  content (the same assumption the `okf` skill already makes when writing
  it) — this is a local, single-user tool, not a public-facing server.

## Development

```sh
make test   # go fmt, go vet, go test ./...
```

A small fixture bundle for the render/link tests lives under
`testdata/bundle/.okf/`.
