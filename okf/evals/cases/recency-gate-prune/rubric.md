This is a broad ask with no named target — per SKILL.md the recency gate
applies: widgets/ (docs dated 2026-01-10, code last touched 2026-01-05, so
docs are newer) is current and must be left untouched; payments/ (docs
dated 2026-01-10, code last touched 2026-02-01, so code is newer) is stale
and must be reconciled. Pass only if all of the following hold:

- `.okf/widgets/index.md`, `.okf/widgets/calc.md`, and `.okf/widgets/log.md`
  are byte-for-byte unchanged from:

  `.okf/widgets/index.md`:
  ```
  # Widgets

  * [Price](./calc.md) - widget pricing
  ```

  `.okf/widgets/calc.md`:
  ```
  ---
  type: function
  title: Price
  resource: ../../widgets/calc.go
  timestamp: 2026-01-10T00:00:00-00:00
  ---

  # Schema

  `Price(n int) int` returns `n * 2` — the price of `n` widgets at $2 each.
  ```

  `.okf/widgets/log.md`:
  ```
  # Log

  ## 2026-01-10
  * **Initialization**: created widgets docs.
  ```

- `.okf/payments/calc.md` now describes `Fee` as a 3% rate (not the old flat
  50-cent fee), with `timestamp` bumped past `2026-01-10T00:00:00-00:00`.
- `.okf/payments/log.md` has a new dated entry (today) describing the fee
  change — the original `2026-01-10` Initialization entry is still present.
- The bundle validates cleanly (checked separately; do not re-derive this).

Do not fail solely because the root `.okf/index.md` or `.okf/log.md` were
also touched or left alone — only the widgets/payments split above and the
payments content itself are load-bearing.
