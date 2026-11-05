# okf-evals

Behavioral evals for the `open-knowledge-format` skill: does an agent following `SKILL.md`
make the right judgment calls (mode selection, recency gate, cross-link
convention, frontmatter preservation), not just produce SPEC-valid output.

This is a different layer from `../scripts/validate/*_test.go`, which are
deterministic unit tests of the validator itself. These invoke a real Claude
or Codex agent per case and may cost real API usage.

## Running

```sh
go run .                                # all cases with Claude (default)
go run . -agent=codex                    # all cases with Codex
go run . -agent=codex -case bootstrap-fresh # one Codex case
go run . -no-judge                       # skip the LLM judge, report validator conformance only
go run . -keep                           # keep each case's working dir for inspection
```

Prints a JSON report per case and exits nonzero if any case fails.

## Anatomy of a case

```
cases/<name>/
  fixture/     # repo state before the task (copied to a scratch dir)
  prompt.txt   # the task, verbatim, as if the user typed it
  rubric.md    # pass/fail checks a second agent grades against
```

Each run: copies `fixture/` to a temp dir, symlinks the `open-knowledge-format` skill into the
selected provider's project-scoped skill path (`.claude/skills/open-knowledge-format` or
`.codex/skills/open-knowledge-format`), runs the case prompt against it, runs
`../scripts/validate/validate -strict .okf` on the result, and — unless `-no-judge`
— asks a fresh agent of the same provider to grade the rubric against the
working dir and the tested agent's final message. Claude cost is included in
the JSON report; Codex results omit `cost_usd` because its CLI does not expose
the same value.

`go test ./...` only checks case directories are well-formed (prompt/
rubric/fixture present); it does not invoke either agent and costs nothing.

## Adding a case

Add a directory under `cases/` with the three files above, aimed at a
specific judgment call from `SKILL.md` — e.g. "does it apply the recency
gate instead of rewriting a current subtree," "does it preserve unknown
frontmatter keys," "does it stop at the bundle root instead of the repo
root."
