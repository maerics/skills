---
name: open-knowledge-format
description: >-
  Read and write Open Knowledge Format (OKF) knowledge bundles — the
  vendor-neutral markdown + YAML spec for AI-agent knowledge bases. Use when
  reading project knowledge from a .okf/ (or explicitly named) directory,
  answering a question a curated knowledge base would cover, when asked to
  show/display/browse/view a bundle (launches the local viewer), when asked to
  bootstrap, update, or reconcile an OKF bundle, or — proactively, without
  being asked — after making any code/doc change in a project that already
  has an OKF bundle, so the bundle stays in sync with what it describes.
---

# OKF: Open Knowledge Format

The full, authoritative OKF v0.3 rules are in [SPEC.md](SPEC.md); read it
before writing a bundle (§13 summarizes changes from v0.2). The canonical
source is
https://github.com/open-knowledge-format/open-knowledge-format/blob/main/SPEC.md.
This skill has exactly two operating modes: **read**
(default) and **write** (everything else).

A bundle is a tree of concept documents. Reserved filenames: `index.md`
(directory listing) and `log.md` (update history). A `.md` file is a
concept only if its frontmatter has a non-empty `type`; other `.md` files
(`README.md`, `AGENTS.md`) are skipped, non-`.md` files are assets, and
dot-paths are not part of the bundle (SPEC §3.2). Consume
**permissively**: tolerate unknown types, extra keys, broken links, and
missing index files (SPEC §11.2).

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

## Links and path fields

SPEC §6 has one rule for both: a value starting with `/` resolves from
the bundle root; anything else resolves from the containing file's
directory. Both forms are valid and consumers must accept both.

- **Body links: write relative** (`./orders.md`, `../tables/orders.md`).
  GitHub, VS Code, and `cat` resolve a leading `/` against the repo root,
  not the bundle root, so the absolute form only works by coincidence —
  which co-location (above) makes the exception, not the rule.
- **Path fields: write in-bundle values with a leading `/`**
  (`superseded_by: /metrics/gross-margin.md`). Path fields are `resource`,
  `sources[].resource`, `superseded_by`, `computation`,
  `executor.resource`, and `attester.resource` (§6.2). Frontmatter isn't
  rendered as links, so the repo-root problem doesn't apply, and this
  matches the spec's examples.
- **Reading a v0.2-or-older bundle** (or one with no `okf_version`): a
  bare path field that doesn't resolve from its file's directory may be
  resolved from the bundle root instead (§12.1). This is the main v0.3
  breaking change.

## Read mode (default)

Triggered when the user asks something a curated knowledge base would answer,
or references project knowledge. Do **not** slurp the whole tree.

**Exception — "show"/"display"/"browse"/"view" the bundle**: launch the
local viewer instead of reading and summarizing files yourself (see
[scripts/viewer/README.md](scripts/viewer/README.md)):

```sh
<skill-base-dir>/scripts/viewer/view <bundle-path>
```

This is a direct request for the UI, not a question to answer from
content — don't substitute a text summary for it. For every other
read-mode ask, proceed as below and just mention the viewer as an option.

1. Read the bundle-root `index.md` for progressive disclosure. If absent,
   synthesize a listing by scanning immediate children's frontmatter.
2. Follow the hierarchy and markdown links toward the relevant concept(s);
   open only what the question needs.
3. Answer from concept bodies. Honor `sources` frontmatter and keyed
   footnote attribution; for legacy v0.1 documents, also read `# Citations`.
   Surface sources when they back a claim.
4. Treat a broken link as not-yet-written knowledge, not an error.
5. Weigh lifecycle and trust signals when they're present (§5):
   - `status: deprecated` is not current; follow `superseded_by`. An
     unknown `status` value is not `stable` either.
   - Past `stale_after` means "needs re-checking", not "false". An
     unparseable `stale_after` is never fresh.
   - Trust tier counts only **current** `verified` events (`at` >=
     `generated.at`): any `human:` actor => human-reviewed, otherwise
     machine-confirmed, none => unverified. Older events are history.

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
   `resource` (when it maps to a real asset), and `tags`. Record authorship
   and edit time with `generated: { by: <actor>, at: "<RFC 3339 datetime
   with offset>" }`; use actor forms from SPEC §7 (`<producer>/<version>`,
   `human:<id>`, `team:<id>`). Add `sources` (each entry needs `resource`)
   with stable `id` values and keyed footnotes when claims derive from
   references. Use `verified` (each event needs `by` and `at`), `status`,
   and `stale_after` when useful; never invent a `stale_after` date. Name
   any custom keys `x-<producer>-<name>`. Quote values that look like
   numbers or booleans. Favor structured markdown (tables, `# Schema`,
   `# Examples`); the body dialect is CommonMark + GFM tables + footnotes.
4. Cross-link per Links and path fields above: relative body links,
   `/`-prefixed in-bundle path fields.
5. Write a root `index.md` with frontmatter `okf_version: "0.3"`, `title`,
   and a one-sentence `description` of the bundle's scope and when to use
   it; add `license` (SPDX id or path) if known (§8.1). Write
   per-directory `index.md` files with no frontmatter. Each index groups
   links under descriptive headings and
   includes concept descriptions. When an
   index's concepts form a real graph (a pipeline's stage order, a state
   machine, service dependencies), add a fenced `mermaid` block there
   that draws it — do this, don't just describe it in prose:

   ~~~
   ```mermaid
   flowchart LR
     Ingest --> Transform --> Load
   ```
   ~~~

   When the concepts are independent (no shared order or dependency
   between them), skip the diagram entirely — forcing one onto unrelated
   concepts is worse than not having one.
6. Write `log.md` (no frontmatter) with today's date and an
   `**Initialization**` entry.

### B. Update (bundle exists)

First classify the ask:

- **Targeted** (user names a specific concept, resource, or change): skip
  the gate below entirely. Locate or create that one concept, check only
  its own `generated.at` (or legacy `timestamp`; otherwise its directory's
  `log.md`) against the resource it describes, and write it. Nothing else
  needs to be opened.
- **Broad** ("reconcile", "is this up to date", "sync the bundle", or
  similar with no named target): apply the recency gate.

**Recency gate** (broad asks only) — walk the bundle top-down, one
directory at a time, so cost scales with what's actually stale, not with
bundle size:

1. At each directory, read its nearest `log.md` heading date (log.md MAY
   exist at any level — §9). No `log.md` at that level counts as unknown:
   treat it as stale and descend.
2. Compare that date against the last material change to what the
   directory's concepts describe:
   - If a concept maps to a path in this repo, compare with `git log -1
     --date=short -- <source-path>`.
   - If a concept has no git-trackable source (an external resource, a
     process, a third-party system), there is no cheap way to verify —
     trust its own `generated.at`/`timestamp`/`log.md` history rather than
     forcing a rewrite you can't actually check.
3. **If the directory's log date is ≥ every checked source's last change,
   it is current: prune.** Do not open its concept files, do not descend
   into its subdirectories, and don't mention it in the report.
4. **Otherwise, descend and reconcile only that subtree**: update the
   concept docs whose subjects changed, preserve unknown frontmatter keys
   when round-tripping (§4.1), refresh that directory's `index.md` while
   keeping the root `index.md` frontmatter (§8.1), bump each touched
   concept's `generated.at`, and prepend a dated entry
   (`**Update**`, `**Creation**`, `**Deprecation**`) to that directory's
   `log.md` — newest first (SPEC §9), above any existing entries. Compare
   source changes with `generated.at`, falling back to legacy `timestamp`
   for v0.1 concepts. Bump `generated.at` only for meaningful content
   changes — not formatting, whitespace, or link repair — and preserve the
   original `generated.by` when the same producer makes the change. Don't
   re-serialize values you didn't change, and don't delete old `verified`
   events; they simply stop counting (§5.3).

**Retire or move** a concept by keeping the file with `status:
deprecated` and `superseded_by: /<new path>.md`; on a move, leave that
short deprecated stub at the old path so existing links keep working
(§5.4).

**Migrating an older bundle**: when you touch a bundle whose root
`index.md` declares `okf_version` below `"0.3"` (or none), bump it to
`"0.3"`, add `title`/`description` if missing, rewrite bare in-bundle path
fields that only resolve from the bundle root to the `/` form, and log it
as an `**Update**`. The validator's `path-field-root-relative` and
`okf-version-old` warnings find these.

If the repo is not under git, skip the git comparisons and rely on
`log.md` dates and `generated.at` fields (or legacy `timestamp`) alone.

### Write checklist

- [ ] Broad updates: recency gate applied first; current subtrees left
      untouched.
- [ ] Unknown pre-existing frontmatter keys preserved when round-tripping.
- [ ] v0.3 fields are used correctly: `generated.at` for the last
      meaningful change, `sources` for provenance, `verified` events with
      both `by` and `at`, timestamps with an explicit offset.
- [ ] In-bundle path fields use a leading `/`; body links are relative.
- [ ] Run the bundled structural validator (see Validation below); it
      does not check every v0.3 field or attested computation contract.

## Validation

After bootstrapping or updating a bundle, or whenever asked to check/lint
one, run the bundled structural validator:

```sh
<skill-base-dir>/scripts/validate/validate [-strict] <bundle-path>
```

`<skill-base-dir>` is this skill's own directory (wherever it's installed —
`~/.claude/skills/open-knowledge-format`, `${CODEX_HOME:-$HOME/.codex}/skills/open-knowledge-format`, or this source
repo). `validate` is a thin shell wrapper around `go run .`, so it works
from any working directory and needs nothing but a Go toolchain — no build
step, no dependencies to fetch. It prints a JSON report
(`conformant`, `counts`, `findings`) and exits non-zero when the bundle has
structural errors detected by its current checks or, with `-strict`, when
it finds house-convention warnings. It is not a complete v0.3 validator:
in particular, it does not validate provenance, trust, lifecycle, or
attested computation fields. Treat its report as a partial check, not proof
of full SPEC conformance.

## Reference

- [SPEC.md](SPEC.md) — the full OKF v0.3 specification.
- Terminology, conformance, citations, and versioning all live in SPEC.md;
  consult it rather than guessing.
- [scripts/validate/README.md](scripts/validate/README.md) — the bundled
  structural validator: check list, output shape, and exit codes.
- [scripts/viewer/README.md](scripts/viewer/README.md) — a local browser
  viewer for a bundle: `<skill-base-dir>/scripts/viewer/view <bundle-path>`.
  Launch it directly on "show"/"display"/"browse"/"view" requests (see Read
  mode above); for other read requests, just mention it as an option rather
  than launching it unprompted, since opening a browser window is a UI side
  effect. It opens a browser itself and shuts its own server down when no
  tab has it open, so there's no process to manage afterward.
