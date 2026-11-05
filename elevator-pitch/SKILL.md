---
name: elevator-pitch
description: >-
  Structure writing for a busy reader who may stop at any moment: punchy
  takeaway first, then a brief assessment, then detail only as needed. Use
  when drafting anything another person will read and act on - PR reviews,
  status updates, proposals, summaries, recommendations, emails, Slack
  messages - or when asked for an "elevator pitch", "TL;DR", or "bottom line".
---

# Elevator Pitch

Imagine an elevator ride with someone you need to convince. You don't know
which floor they get off at. Every stop must leave them with the most
important thing you could have said by then.

## Structure

1. **Takeaway** - one line. The verdict, ask, or conclusion. If they read
   nothing else, this is enough to act on.
2. **Assessment** - 1-3 sentences. Why, the main tradeoff or risk, and what
   happens next.
3. **Elaboration** - only if it's needed. Supporting points, ordered by
   importance. Use bullets, and keep each one self-contained.
4. **Details** - only on request or when required to act. Evidence,
   references, edge cases.

Each layer must stand on its own. A reader who stops after any layer should
come away with a correct, if less complete, picture.

## Rules

- Lead with the conclusion, not the context. Don't say "I looked at X and...".
  Say "X is ready." or "Don't ship X yet."
- Make the takeaway falsifiable and specific. "Approve; one blocking bug in
  retry logic" beats "Looks mostly good with some concerns."
- Make any ask explicit: who needs to do what, and by when if it matters.
- Cut throat-clearing: greetings, restating the question, hedges, summaries
  of what was just said.
- Order by importance, not by when you discovered it.
- Stop when the reader has what they need. Leaving something out is fine.

## Example

> **Ship it after one fix:** the cache key ignores `tenant_id`, so tenants
> can read each other's data.
>
> Otherwise the change is small, well tested, and does what the description
> says. The fix is a one-liner (details below).
>
> - `cache.go:42` - include `tenant_id` in `key()`.
> - Nit: `fetchAll` could reuse `fetchPage`.
