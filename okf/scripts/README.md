# okf-validate

Validates an OKF (Open Knowledge Format) knowledge bundle against
[SPEC.md](../SPEC.md) and the [okf skill](../SKILL.md)'s house directives,
printing a JSON report of any findings. Stdlib-only Go — no dependencies,
no build step.

## Usage

```
go run . [-strict] [bundle-path]

Arguments:
  bundle-path    Bundle root, or a directory containing .okf/. Defaults to ".".

Flags:
  -strict        Also fail (nonzero exit) when warnings are present.
```

Run from this directory (`okf/scripts/`), or point `go run` at it from
anywhere: `go run <skill-base-dir>/scripts <bundle-path>`.

## Output

A single JSON document:

```jsonc
{
  "bundle": "path/to/bundle",
  "conformant": true,           // true iff there are zero errors
  "counts": { "files": 12, "errors": 0, "warnings": 1 },
  "findings": [
    {
      "code": "absolute-link",
      "level": "warn",           // "error" or "warn"
      "spec": "SPEC §5.1 vs. SKILL.md",
      "path": "commands/index.md",
      "line": 9,
      "message": "link target \"/index.md\" uses the absolute bundle-relative form; the skill prefers relative links"
    }
  ]
}
```

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Conformant (no errors), and — under `-strict` — no warnings either. |
| `1` | At least one error, or an operational failure (bad path, I/O error). |
| `2` | Only reachable with `-strict`: no errors, but at least one warning. |

## Checks

Only rules that map to a MUST/MUST-NOT sentence in SPEC.md are errors;
everything else is a warning drawn from SPEC's own soft guidance or the
skill's house conventions (SPEC §9 explicitly forbids rejecting a bundle
over any of these on its own).

| Code | Level | Spec | What it flags |
|---|---|---|---|
| `frontmatter-missing` | error | §4.1, §9.1 | A non-reserved `.md` file has no frontmatter block. |
| `frontmatter-invalid` | error | §4.1, §9.1 | The frontmatter block is unclosed, or contains an unparseable line. |
| `type-missing` | error | §4.1, §9.2 | Frontmatter has no non-empty `type` field. |
| `type-invalid` | error | §4.1, §9.2 | `type` is a list/mapping instead of a scalar string. |
| `index-frontmatter` | error | §6, §9.3, §11 | A non-root `index.md` has a frontmatter block (only the bundle-root `index.md` may). |
| `log-frontmatter` | error | §7, §9.3 (extension*) | A `log.md` has a frontmatter block. |
| `log-bad-date-heading` | error | §7, §9.3 | A `log.md` `##` heading isn't a valid ISO 8601 `YYYY-MM-DD` date. |
| `index-no-sections` | warn | §6 | `index.md` body has no `#` heading sections. |
| `log-not-newest-first` | warn | §7 | `log.md` date headings aren't in newest-first order. |
| `absolute-link` | warn | §5.1 vs. SKILL.md | A link uses the `/`-prefixed absolute form; the skill prefers relative links. |
| `broken-link` | warn | §5.3 | A link target doesn't resolve to a file in the bundle. |
| `root-index-missing-field` | warn | SKILL.md | Root `index.md` frontmatter is missing `okf_version`, `type`, `title`, or `tags`. |
| `root-index-missing` | warn | SKILL.md | No `index.md` at the bundle root. |
| `log-missing` | warn | SKILL.md | No `log.md` at the bundle root. |

\* SPEC.md never explicitly bans frontmatter on `log.md` the way it does for
`index.md` — this is a deliberate extension applied for consistency, since
both are reserved structural files under §3.1. See the code comment on
`checkLogFile` in `checks.go`.

**Explicitly not checked**, per SPEC §9's list of things a bundle must not
be rejected for: unknown `type` values, unknown frontmatter keys, missing
optional fields on ordinary concepts, missing non-root `index.md`/`log.md`,
index bullet-line formatting, the `log.md` bold-word convention, and
citation formatting.

## Known limitations

- Frontmatter is parsed with a minimal, hand-rolled scanner (flat
  `key: value` pairs, inline or block lists) rather than a full YAML
  parser — sufficient for OKF's flat frontmatter shape, but it won't catch
  every YAML grammar error.
- Links are found with a line-based regex, not a full CommonMark parser, so
  a link split across lines or one that appears inside a fenced code block
  can be missed.

## Development

```sh
make test   # go fmt, go vet, go test ./...
```

Fixture bundles for each rule live under `testdata/`.
