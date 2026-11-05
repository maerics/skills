#!/bin/sh
# Builds git history so the recency gate has something real to compare
# against: widgets/ was last touched before its docs' log date (current,
# should be pruned); payments/ was last touched after its docs' log date
# (stale, should be reconciled).
set -eu

git init -q
git config user.email "eval@example.com"
git config user.name "Eval Bot"

export GIT_AUTHOR_DATE="2026-01-05T00:00:00"
export GIT_COMMITTER_DATE="2026-01-05T00:00:00"
git add widgets
git commit -q -m "widgets: initial pricing"

export GIT_AUTHOR_DATE="2026-02-01T00:00:00"
export GIT_COMMITTER_DATE="2026-02-01T00:00:00"
git add payments
git commit -q -m "payments: switch to percentage fee"
git add .okf
git commit -q -m "docs: initial okf bundle"
