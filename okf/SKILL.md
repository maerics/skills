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

### B. Update (bundle exists) — apply the recency gate first

Before re-deriving anything, decide whether the bundle is already current:

1. Read the newest date heading in the relevant `log.md`.
2. Check git history for the bundle path: `git log -1 --format=%cd
   --date=short -- <bundle-path>` and compare against changes to the source
   it describes (e.g. `git log -1 --date=short -- <source-path>`).
3. **If the bundle was reconciled at or after the last material change to
   reality** (log date is recent and ≥ the source's last change), it is
   current: apply only the specific edit the user asked for, or report that
   nothing needs doing. Do not rebuild.
4. **Otherwise reconcile**: update the concept docs whose subjects changed,
   preserve unknown frontmatter keys when round-tripping (§4.1), refresh the
   affected `index.md` entries, bump each touched concept's `timestamp`, and
   append a dated `log.md` entry (`**Update**`, `**Creation**`,
   `**Deprecation**`) describing what changed.

If the repo is not under git, rely on `log.md` dates and the `timestamp`
fields alone for the recency judgment.

### Write checklist

- [ ] Every non-reserved `.md` has parseable frontmatter with non-empty `type`.
- [ ] Links are bundle-relative where practical.
- [ ] Affected `index.md` files reflect current contents.
- [ ] `log.md` has a dated entry for this change.
- [ ] Unknown pre-existing frontmatter keys preserved.

## Reference

- [SPEC.md](SPEC.md) — the full OKF v0.1 specification (authoritative).
- Terminology, conformance, citations, and versioning all live in SPEC.md;
  consult it rather than guessing.
