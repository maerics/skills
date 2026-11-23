# okf-validate

Validates an OKF (Open Knowledge Format) knowledge bundle against
[SPEC.md](../../SPEC.md) and the [open-knowledge-format skill](../../SKILL.md)'s house directives,
printing a JSON report of any findings. Stdlib-only Go — no dependencies,
no build step.

## Usage

```
./validate [-strict] [bundle-path]

Arguments:
  bundle-path    Bundle root, or a directory containing .okf/. Defaults to ".".

Flags:
  -strict        Also fail (nonzero exit) when warnings are present.
```

`validate` is a shell wrapper around `go run .` that resolves its own
location, so it works from anywhere: `<skill-base-dir>/scripts/validate/validate
<bundle-path>`. Equivalent without the wrapper: `go run .` from this
directory, or `go run <skill-base-dir>/scripts/validate <bundle-path>` from anywhere.

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
      "spec": "SPEC §6.1 vs. SKILL.md",
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

Only rules that break SPEC §11.1 bundle conformance are errors;
everything else is a warning drawn from SPEC's own soft guidance or the
skill's house conventions (SPEC §11.2 forbids rejecting a bundle over any
of these on its own).

| Code | Level | Spec | What it flags |
|---|---|---|---|
| `frontmatter-invalid` | error | §4, §11.1 | The frontmatter block is unclosed, or contains an unparseable line. |
| `index-frontmatter` | error | §8 | A non-root `index.md` has a frontmatter block (only the bundle-root `index.md` may). |
| `log-frontmatter` | error | §9 | A `log.md` has a frontmatter block. |
| `log-bad-date-heading` | error | §9 | A `log.md` `##` heading isn't a valid ISO 8601 `YYYY-MM-DD` date. |
| `frontmatter-missing` | warn | §3.2 | A non-reserved `.md` file has no frontmatter block, so it is not a concept. |
| `type-missing` | warn | §3.2, §4.1 | Frontmatter has no non-empty `type` field, so the file is not a concept. |
| `type-invalid` | warn | §3.2, §4.1 | `type` is a list/mapping instead of a scalar string. |
| `index-no-sections` | warn | §8 | `index.md` body has no `#` heading sections. |
| `log-not-newest-first` | warn | §9 | `log.md` date headings aren't in newest-first order. |
| `absolute-link` | warn | §6.1 vs. SKILL.md | A body link uses the `/`-prefixed form; the skill prefers relative links. |
| `broken-link` | warn | §6.1 | A link target doesn't resolve to a file in the bundle. |
| `path-field-root-relative` | warn | §6.2, §13.1 | A top-level `resource`, `superseded_by`, or `computation` without a leading `/` resolves only from the bundle root (v0.2 style). |
| `okf-version-old` | warn | §12, §13 | Root `index.md` declares an `okf_version` other than `0.3`. |
| `root-index-missing-field` | warn | §8.1 | Root `index.md` frontmatter is missing `okf_version`, `title`, or `description`. |
| `root-index-missing` | warn | SKILL.md | No `index.md` at the bundle root. |
| `log-missing` | warn | SKILL.md | No `log.md` at the bundle root. |

**Explicitly not checked**, per SPEC §11.2's list of things a bundle must not
be rejected for: unknown `type` values, unknown frontmatter keys, missing
optional fields on ordinary concepts, missing non-root `index.md`/`log.md`,
index bullet-line formatting, the `log.md` bold-word convention, and
citation formatting.

## Known limitations

- Frontmatter is parsed with a minimal, hand-rolled scanner (flat
  `key: value` pairs, inline or block lists) rather than a full YAML
  parser — sufficient for OKF's flat frontmatter shape, but it won't catch
  every YAML grammar error.
- Path fields nested inside mappings or lists (`sources[].resource`,
  `executor.resource`, `attester.resource`) are not checked.
- Links are found with a line-based regex, not a full CommonMark parser, so
  a link split across lines or one that appears inside a fenced code block
  can be missed.

## Development

```sh
make test   # go fmt, go vet, go test ./...
```

Fixture bundles for each rule live under `testdata/`.
