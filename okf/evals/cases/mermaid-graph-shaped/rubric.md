This system is genuinely graph-shaped (a linear 3-stage pipeline) — per the
SKILL.md mermaid guidance, a diagram earns its place here. Pass only if all
of the following hold:

- Somewhere in the bundle (most likely an overview/index concept doc),
  there is a fenced ```mermaid code block.
- That diagram depicts the actual flow: Ingest -> Transform -> Load, in
  that order, with no fabricated stages or edges not present in the
  fixture.
- The diagram supplements prose describing the pipeline rather than
  replacing all explanation of it (concept docs still have real content,
  not just a diagram).
- The bundle validates cleanly (checked separately; do not re-derive this).

Do not fail solely because of where in the bundle the diagram lives, its
exact mermaid diagram type (`flowchart`, `graph`, etc.), or minor styling.
