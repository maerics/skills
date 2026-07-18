---
name: okf
description: >-
  Read and write Open Knowledge Format (OKF) knowledge bundles — Google Cloud's
  vendor-neutral markdown + YAML spec for AI-agent knowledge bases. Use when
  reading project knowledge from a .okf/ (or explicitly named) directory,
  answering a question a curated knowledge base would cover, when asked to
  bootstrap, update, or reconcile an OKF bundle, or — proactively, without
  being asked — after making any code/doc change in a project that already
  has an OKF bundle, so the bundle stays in sync with what it describes.
---

# OKF: Open Knowledge Format

The full, authoritative rules are in [SPEC.md](SPEC.md); read it before
writing a bundle. This skill has exactly two operating modes: **read**
(default) and **write** (everything else).

A bundle is a tree of concept documents. Reserved filenames: `index.md`
(directory listing) and `log.md` (update history). Every other `.md` file
is a concept with a required `type` frontmatter field. Consume
**permissively**: tolerate unknown types, extra keys, broken links, and
missing index files (SPEC §9).

## Locating a bundle

If the user names a path, use it. Otherwise search **nearest-first**: a
co-located `.okf/` in the directory the task concerns (independent
modules — apps, services, libraries — are encouraged to keep their own
bundle so knowledge travels with the code it describes), then walk
upward to a `.okf/` at the repository root.

Treat multiple bundles as fully independent — don't mix concepts or
links across them, and note which bundle(s) you consulted.

Bundle root = the directory containing the top-level `index.md`, or the
`.okf/` directory itself if there's no index — **never** the repo root,
even when a bundle sits at or near it.

## Cross-linking

SPEC §5 permits relative links (`./orders.md`) and absolute
bundle-relative links (`/tables/orders.md`). Default to **relative**
when writing: GitHub, VS Code, and `cat` resolve a leading `/` against
the repo root, not the bundle root, so the absolute form only works by
coincidence when they're the same directory — which co-location (above)
makes the exception, not the rule. Still tolerate absolute links when
reading; a differently-formed link isn't malformed (SPEC §9).

## Read mode (default)

Triggered when the user asks something a curated knowledge base would answer,
or references project knowledge. Do **not** slurp the whole tree.

1. Read the bundle-root `index.md` for progressive disclosure. If absent,
   synthesize a listing by scanning immediate children's frontmatter.
2. Follow the hierarchy and markdown links toward the relevant concept(s);
   open only what the question needs.
3. Answer from concept bodies. Honor `resource` URIs and `# Citations` as
   sources; surface them when they back a claim.
4. Treat a broken link as not-yet-written knowledge, not an error.

## Write mode (everything else)

Any ask that isn't answerable by reading falls here — including bootstrap,
update, and reconcile requests. Determine state and pick one path
yourself; don't ask the user to choose between bootstrap and update:

### A. Bootstrap (no bundle exists)

1. Create the bundle directory (`.okf/` unless the user explicitly chose
   otherwise).
2. Derive concepts from reality — the actual code, data, schemas, docs.
   Pick descriptive `type` values (`SPEC.md` §4.1) and organize into
   subdirectories that fit the domain.
3. Each concept: required `type`; recommended `title`, `description`,
   `resource` (when it maps to a real asset), `tags`, and `timestamp`
   (ISO 8601). Favor structured markdown (tables, `# Schema`, `# Examples`).
4. Cross-link with relative links (`../tables/orders.md`, `./orders.md`) —
   see Cross-linking above.
5. Write a root `index.md` (frontmatter allowed **only** here):
   `okf_version: "0.1"`, `type` (the boundary this bundle covers — e.g.
   `service`, `library`, `app`, `team`), `title`, `tags` (scope
   limiters), and `resource` when there's a canonical URI for the
   bundle's subject. Also write per-directory `index.md` files.
6. Write `log.md` with today's date and an `**Initialization**` entry.

### B. Update (bundle exists)

First classify the ask:

- **Targeted** (user names a specific concept, resource, or change): skip
  the gate below entirely. Locate or create that one concept, check only
  its own `timestamp` (or its directory's `log.md`) against the resource it
  describes, and write it. Nothing else needs to be opened.
- **Broad** ("reconcile", "is this up to date", "sync the bundle", or
  similar with no named target): apply the recency gate.

**Recency gate** (broad asks only) — walk the bundle top-down, one
directory at a time, so cost scales with what's actually stale, not with
bundle size:

1. At each directory, read its nearest `log.md` heading date (log.md MAY
   exist at any level — §7). No `log.md` at that level counts as unknown:
   treat it as stale and descend.
2. Compare that date against the last material change to what the
   directory's concepts describe:
   - If a concept maps to a path in this repo, compare with `git log -1
     --date=short -- <source-path>`.
   - If a concept has no git-trackable source (an external resource, a
     process, a third-party system), there is no cheap way to verify —
     trust its own `timestamp`/`log.md` history rather than forcing a
     rewrite you can't actually check.
3. **If the directory's log date is ≥ every checked source's last change,
   it is current: prune.** Do not open its concept files, do not descend
   into its subdirectories, and don't mention it in the report.
4. **Otherwise, descend and reconcile only that subtree**: update the
   concept docs whose subjects changed, preserve unknown frontmatter keys
   when round-tripping (§4.1), refresh that directory's `index.md`, bump
   each touched concept's `timestamp`, and append a dated entry
   (`**Update**`, `**Creation**`, `**Deprecation**`) to that directory's
   `log.md`.

If the repo is not under git, skip the git comparisons and rely on
`log.md` dates and `timestamp` fields alone throughout.

### Write checklist

- [ ] Broad updates: recency gate applied first; current subtrees left
      untouched.
- [ ] Unknown pre-existing frontmatter keys preserved when round-tripping.
- [ ] Bundle validates cleanly (see Validation below) — the validator
      covers frontmatter, `type`, links, index files, and `log.md`.

## Validation

After bootstrapping or updating a bundle, or whenever asked to check/lint
one, confirm it against SPEC.md and the conventions above with the
validator script bundled in this skill:

```sh
<skill-base-dir>/scripts/validate [-strict] <bundle-path>
```

`<skill-base-dir>` is this skill's own directory (wherever it's installed —
`~/.claude/skills/okf`, a project's `.claude/skills/okf`, or this source
repo). `validate` is a thin shell wrapper around `go run .`, so it works
from any working directory and needs nothing but a Go toolchain — no build
step, no dependencies to fetch. It prints a JSON report
(`conformant`, `counts`, `findings`) and exits non-zero when the bundle has
SPEC violations (errors) or, with `-strict`, when it merely has
house-convention warnings (unrelated to genuine SPEC conformance). Each
finding cites the SPEC section it's based on.

## Reference

- [SPEC.md](SPEC.md) — the full OKF v0.1 specification (authoritative).
- Terminology, conformance, citations, and versioning all live in SPEC.md;
  consult it rather than guessing.
- [scripts/README.md](scripts/README.md) — the bundle validator invoked
  above: full check list, output shape, exit codes.
