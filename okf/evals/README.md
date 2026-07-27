# okf-evals

Behavioral evals for the `okf` skill: does an agent following `SKILL.md`
make the right judgment calls (mode selection, recency gate, cross-link
convention, frontmatter preservation), not just produce SPEC-valid output.

This is a different layer from `../scripts/*_test.go`, which are
deterministic unit tests of the validator itself. These invoke a real
`claude -p` agent per case and cost real API usage.

## Running

```sh
go run .                       # all cases
go run . -case bootstrap-fresh # one case
go run . -no-judge              # skip the LLM judge, report validator conformance only
go run . -keep                  # keep each case's working dir for inspection
```

Prints a JSON report per case and exits nonzero if any case fails.

## Anatomy of a case

```
cases/<name>/
  fixture/     # repo state before the task (copied to a scratch dir)
  prompt.txt   # the task, verbatim, as if the user typed it
  rubric.md    # pass/fail checks a second agent grades against
```

Each run: copies `fixture/` to a temp dir, symlinks the `okf` skill into
`.claude/skills/okf` there (so it resolves like it would in a real repo),
runs `claude -p "$(cat prompt.txt)"` against it, runs
`../scripts/validate -strict .okf` on the result, and — unless `-no-judge`
— asks a fresh agent to grade the rubric against the working dir and the
tested agent's final message.

`go test ./...` only checks case directories are well-formed (prompt/
rubric/fixture present); it does not invoke `claude` and costs nothing.

## Adding a case

Add a directory under `cases/` with the three files above, aimed at a
specific judgment call from `SKILL.md` — e.g. "does it apply the recency
gate instead of rewriting a current subtree," "does it preserve unknown
frontmatter keys," "does it stop at the bundle root instead of the repo
root."
