# skills

Personal, provider-agnostic AI skills — plain folders any agent can read.

Each skill is a directory containing a `SKILL.md` with YAML frontmatter
(`name`, `description`) and whatever supporting files it needs. The format
follows the common Agent Skills convention, so it works with Claude Code and
any tool that reads the same layout.

## Skills

| Skill | Purpose |
|-------|---------|
| [`okf/`](okf/SKILL.md) | Read and write [Open Knowledge Format](okf/SPEC.md) knowledge bundles. |

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
ln -s ~/src/skills/okf ~/.claude/skills/okf

# ... or per-project — available in one repo only:
mkdir -p .claude/skills
ln -s ~/src/skills/okf .claude/skills/okf
```

Symlinking keeps the skill updating with `git pull`; copy instead if you want
a frozen version. Claude picks the skill up automatically when a task matches
its `description` — invoke it explicitly with `/okf`.

### Other providers

_Coming soon._ The layout follows the common Agent Skills convention (a
directory with a `SKILL.md` carrying `name`/`description` frontmatter), so any
tool that reads that convention can consume these skills the same way.

## Using the okf skill

By default it **reads** knowledge from an `okf/` or `.okf/` directory; on
explicit request it **writes** — bootstrapping a new bundle or reconciling an
existing one against reality. See [okf/SKILL.md](okf/SKILL.md).

## Licensing

Repository content is MIT (see [LICENSE](LICENSE)). The vendored OKF
specification at [okf/SPEC.md](okf/SPEC.md) is © Google LLC under Apache-2.0
(see [okf/LICENSE](okf/LICENSE) and [okf/NOTICE](okf/NOTICE)).
