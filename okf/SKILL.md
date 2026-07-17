---
name: okf
description: >-
  Read and write Open Knowledge Format (OKF) knowledge bundles — Google Cloud's
  vendor-neutral markdown + YAML spec for AI-agent knowledge bases. Use when
  reading project knowledge from a .okf/ (or explicitly named) directory,
  answering a question a curated knowledge base would cover, or when asked to
  bootstrap, update, or reconcile an OKF bundle.
---

# OKF: Open Knowledge Format

OKF represents knowledge as a directory of markdown files with YAML
frontmatter — no schema registry, no tooling required. The full,
authoritative rules are in [SPEC.md](SPEC.md); read it before writing a
bundle. This skill has exactly two operating modes: **read** (default)
and **write** (everything else).

A bundle is a tree of concept documents. Reserved filenames: `index.md`
(directory listing) and `log.md` (update history). Every other `.md` file
is a concept with a required `type` frontmatter field. Consume
**permissively**: tolerate unknown types, extra keys, broken links, and
missing index files (SPEC §9).

## Locating a bundle

The `.okf/` dir is autodetected, named or inferred locations follow.

Search **nearest-first**, unless the user names a path:

1. A co-located `.okf/` in the module/component/directory the current task
   concerns. Independent modules — apps, services, libraries — are
   encouraged to keep their own bundle so knowledge stays local to the code
   it describes and travels with it through moves and refactors.
2. Walking upward, a `.okf/` at the repository root.
3. A directory the user explicitly names.

If several bundles exist (e.g., co-located ones plus a root-level one),
treat each as fully independent: don't mix concepts or links across bundle
boundaries, and note which bundle(s) were consulted when answering.

The bundle root is the directory containing the top-level `index.md` (or,
absent that, the `.okf/` directory itself) — **never** the repository
root, even when a bundle happens to sit at or near it.

## Cross-linking

SPEC §5 permits two link forms. When writing, default to **relative**
links (§5.2: `./orders.md`, `../tables/customers.md`) — they resolve
correctly in GitHub's web UI, VS Code's explorer and markdown preview, and
plain `cat`, regardless of where the bundle sits in the repo tree.

Avoid *writing* SPEC §5.1's absolute `/`-prefixed form even though it's
spec-legal: GitHub and VS Code both resolve a leading `/` against the
repository/workspace root, not the bundle root, so it only happens to
work when the bundle root and repo root are the same directory — which
co-location (above) makes the exception, not the rule. Still tolerate it
when *reading* — a link written before this guidance, or by another
producer, is not malformed (SPEC §9).

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
5. Write a root `index.md` (frontmatter allowed **only** here). Include
   `okf_version: "0.1"` plus, as a house convention layered on SPEC's
   permissive extension rule (§4.1), standard manifest fields describing
   the bundle itself: `type` (the boundary this bundle covers — e.g.
   `service`, `library`, `app`, `team`), `title` (human-readable component
   name), and `tags` (scope limiters). `resource` is welcome too when
   there's a canonical URI for the bundle's subject (repo, package,
   service). This lets an agent that finds several co-located bundles
   tell them apart without opening each one. Also write per-directory
   `index.md` files.
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
- [ ] Links are relative (not `/`-absolute) so they resolve correctly in
      GitHub, VS Code, and other common dev tools without needing to know
      the bundle root.
- [ ] Affected `index.md` files reflect current contents.
- [ ] `log.md` has a dated entry for this change.
- [ ] Unknown pre-existing frontmatter keys preserved.
- [ ] Broad updates: recency gate applied first; current subtrees left
      untouched.
- [ ] Bundle validates cleanly (see Validation below).

## Validation

After bootstrapping or updating a bundle, or whenever asked to check/lint
one, confirm it against SPEC.md and the conventions above with the
validator script bundled in this skill:

```sh
go run <skill-base-dir>/scripts <bundle-path>
```

`<skill-base-dir>` is this skill's own directory (wherever it's installed —
`~/.claude/skills/okf`, a project's `.claude/skills/okf`, or this source
repo). The script is stdlib-only Go, so `go run` needs nothing beyond a Go
toolchain: no build step, no dependencies to fetch. It prints a JSON report
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
