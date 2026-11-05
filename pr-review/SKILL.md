---
name: pr-review
description: >-
  Review a pull/merge request (GitHub, GitLab, or similar) and write the
  review using the elevator-pitch style. Use when asked to review a PR, MR,
  or diff for someone else, or to draft review comments.
---

# PR Review

Write up the review following the [elevator-pitch](../elevator-pitch/SKILL.md)
skill. This skill covers what to judge.

## Process

1. **Understand intent.** Read the title, description, and linked issues.
   What is this change trying to do, and why now?
2. **Read the diff against that intent.** Does it do what it claims, and
   nothing else? Unexplained scope creep is a finding.
3. **Judge in priority order:**
   1. Correctness: bugs, edge cases, error handling, concurrency.
   2. Security: secrets, injection, authz, unsafe input, data exposure.
   3. Design: right place and abstraction? Simpler alternative? Reuse of
      existing code?
   4. Tests: do they prove the change works and guard against regression?
   5. Operability: migrations, rollout and rollback, observability,
      breaking changes.
   6. Readability: naming, dead code, comments that should or shouldn't
      exist.
4. **Verify before claiming.** Check the surrounding code, run the tests, or
   reproduce the issue when you can. Mark anything unverified as a question,
   not a defect.

## Output

- **Takeaway:** one of *Approve*, *Approve with nits*, *Request changes*,
  or *Needs discussion*, plus the single most important reason.
- **Assessment:** 1-3 sentences on overall quality, risk, and fit.
- **Findings:** most severe first. Give each a file:line, the problem, a
  concrete failure scenario, and a suggested fix. Label each one
  `blocking`, `suggestion`, `nit`, or `question`.
- Leave out praise filler. Credit what's good only when it's notable, in one
  line.

## Etiquette

- Critique the code, not the author. Ask "What happens if X is nil?" rather
  than "You forgot nil."
- Keep comments actionable. If you can't suggest a fix, ask a question.
- Never post comments, approve, or request changes without explicit
  confirmation. Draft first.
