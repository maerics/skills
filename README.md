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

## Using a skill

Point your agent at this repository (or copy a skill directory into your
project). An agent reads the skill's `SKILL.md` and applies it when the
`description` matches the task.

For `okf`: by default it **reads** knowledge from an `okf/` or `.okf/`
directory; on explicit request it **writes** — bootstrapping a new bundle or
reconciling an existing one against reality. See [okf/SKILL.md](okf/SKILL.md).

## Licensing

Repository content is MIT (see [LICENSE](LICENSE)). The vendored OKF
specification at [okf/SPEC.md](okf/SPEC.md) is © Google LLC under Apache-2.0
(see [okf/LICENSE](okf/LICENSE) and [okf/NOTICE](okf/NOTICE)).
