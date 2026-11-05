# skills

Personal, provider-agnostic AI skills — plain folders any agent can read.

Each skill is a directory containing a `SKILL.md` with YAML frontmatter
(`name`, `description`) and whatever supporting files it needs. The format
follows the common Agent Skills convention, so it works with Claude Code,
Codex, and any tool that reads the same layout.

## Skills

| Skill | Purpose |
|-------|---------|
| [`open-knowledge-format/`](open-knowledge-format/SKILL.md) | Read and write [Open Knowledge Format](open-knowledge-format/SPEC.md) knowledge bundles. |
| [`elevator-pitch/`](elevator-pitch/SKILL.md) | Write for busy readers: takeaway first, then assessment, then detail. |
| [`pr-review/`](pr-review/SKILL.md) | Review pull requests and write the review elevator-pitch style. |

## Installing

A skill is just its directory. "Installing" means putting that directory
where your agent looks for skills. Instructions per provider follow; more
will be added over time.

### Claude (Claude Code)

Claude Code discovers skills in `~/.claude/skills/` (available everywhere)
or `.claude/skills/` inside a project (that project only). Install a single
skill by cloning this repo and symlinking (or copying) the skill directory:

```sh
# Personal — available in every project:
mkdir -p ~/.claude/skills
ln -s ~/src/skills/open-knowledge-format ~/.claude/skills/open-knowledge-format

# ... or per-project — available in one repo only:
mkdir -p .claude/skills
ln -s ~/src/skills/open-knowledge-format .claude/skills/open-knowledge-format
```

Symlinking keeps the skill updating with `git pull`; copy instead if you want
a frozen version. Claude picks the skill up automatically when a task matches
its `description` — invoke it explicitly with `/open-knowledge-format`.

### Codex

Codex discovers personal skills in `$CODEX_HOME/skills` (normally
`~/.codex/skills`). Install the skill by cloning this repo and symlinking (or
copying) its directory:

```sh
mkdir -p "${CODEX_HOME:-$HOME/.codex}/skills"
ln -s ~/src/skills/open-knowledge-format "${CODEX_HOME:-$HOME/.codex}/skills/open-knowledge-format"
```

Symlinking keeps the skill updating with `git pull`; copy instead if you want
a frozen version. Codex picks the skill up automatically when a task matches
its `description` — invoke it explicitly with `$open-knowledge-format`.

## Using the open-knowledge-format skill

By default it **reads** knowledge from a `.okf/` directory (any other name
must be given explicitly); on explicit request it **writes** — bootstrapping
a new bundle or reconciling an existing one against reality. Independent
modules (apps, services, libraries) are encouraged to keep their own
co-located bundle rather than relying on a single one at the repo root. See
[open-knowledge-format/SKILL.md](open-knowledge-format/SKILL.md).

## Licensing

Repository content is MIT (see [LICENSE](LICENSE)). The vendored OKF
specification at [open-knowledge-format/SPEC.md](open-knowledge-format/SPEC.md) is © Google LLC under Apache-2.0
(see [open-knowledge-format/LICENSE](open-knowledge-format/LICENSE) and [open-knowledge-format/NOTICE](open-knowledge-format/NOTICE)).
