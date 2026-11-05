Pass only if all of the following hold:

- `.okf/index.md` exists with frontmatter including `okf_version: "0.1"` and a
  `type`.
- At least one concept doc under `.okf/` documents `add.go` / the `Add`
  function (not just restates the README verbatim).
- `.okf/log.md` exists, dated today, with an **Initialization** entry.
- Any cross-links between concept docs are relative (`./foo.md`,
  `../foo.md`), not root-absolute (`/foo.md`).
- Frontmatter on non-index concept files contains only `type` plus
  well-known OKF keys (no invented schema).
