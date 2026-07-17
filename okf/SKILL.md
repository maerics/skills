---
name: okf
description: >-
  Read and write Open Knowledge Format (OKF) knowledge bundles — Google Cloud's
  vendor-neutral markdown + YAML spec for AI-agent knowledge bases. Use when
  reading project knowledge from an okf/ or .okf/ directory, answering a
  question a curated knowledge base would cover, or when asked to bootstrap,
  update, or reconcile an OKF bundle.
---

# OKF: Open Knowledge Format

OKF represents knowledge as a directory of markdown files with YAML
frontmatter — no schema registry, no tooling required. The full,
authoritative rules are in [SPEC.md](SPEC.md); read it before writing a
bundle. This skill defines two operating modes: **read** (default) and
**write** (on request).

A bundle is a tree of concept documents. Reserved filenames: `index.md`
(directory listing) and `log.md` (update history). Every other `.md` file
is a concept with a required `type` frontmatter field. Consume
**permissively**: tolerate unknown types, extra keys, broken links, and
missing index files (SPEC §9).

## Locating a bundle

Search, in order, unless the user names a path:

1. `okf/` at the project root.
2. `.okf/` at the project root.
3. A directory the user points to.

If several exist, prefer `okf/` and mention the others. The bundle root is
the directory containing the top-level `index.md` (or, absent that, the
`okf/` / `.okf/` directory itself).

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

## Write mode (on request only)

Never write a bundle unless explicitly asked. First determine state, then
pick one path:

### A. Bootstrap (no bundle exists)

1. Create the bundle directory (`okf/` unless the user chose otherwise).
2. Derive concepts from reality — the actual code, data, schemas, docs.
   Pick descriptive `type` values (`SPEC.md` §4.1) and organize into
   subdirectories that fit the domain.
3. Each concept: required `type`; recommended `title`, `description`,
   `resource` (when it maps to a real asset), `tags`, and `timestamp`
   (ISO 8601). Favor structured markdown (tables, `# Schema`, `# Examples`).
4. Cross-link with bundle-relative links (`/tables/orders.md`).
5. Write a root `index.md` (frontmatter allowed **only** here — include
   `okf_version: "0.1"`), plus per-directory `index.md` files.
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

- [ ] Every non-reserved `.md` has parseable frontmatter with non-empty `type`.
- [ ] Links are bundle-relative where practical.
- [ ] Affected `index.md` files reflect current contents.
- [ ] `log.md` has a dated entry for this change.
- [ ] Unknown pre-existing frontmatter keys preserved.
- [ ] Broad updates: recency gate applied first; current subtrees left
      untouched.

## Reference

- [SPEC.md](SPEC.md) — the full OKF v0.1 specification (authoritative).
- Terminology, conformance, citations, and versioning all live in SPEC.md;
  consult it rather than guessing.
