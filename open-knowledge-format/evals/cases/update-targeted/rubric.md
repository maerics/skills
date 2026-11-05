This was a targeted ask (names b.go specifically) — per SKILL.md, only b.md
(and the shared log.md, for the dated entry) should be touched. Pass only if
all of the following hold:

- `.okf/a.md` is byte-for-byte unchanged from:

  ```
  ---
  type: function
  title: A
  resource: ../a.go
  timestamp: 2026-01-01T00:00:00-00:00
  ---

  # Schema

  `A() int` returns the constant `1`.
  ```

- `.okf/index.md` is unchanged (or, if touched, still lists both A and B with
  no factual regressions — do not fail solely for whitespace-level edits).
- `.okf/b.md` now describes B() returning `"hello there"` (not the empty
  string), and its `timestamp` was bumped past `2026-01-01T00:00:00-00:00`.
- `.okf/log.md` has a new dated entry (today) noting the B update — the
  original `2026-01-01` Initialization entry is still present.
- The bundle validates cleanly (checked separately; do not re-derive this).
