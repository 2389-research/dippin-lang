#!/bin/sh
# ABOUTME: Fails when a docs/X.md page changed but its hand-written website twin
# ABOUTME: site/content/X.md did not, so dippin.org can't silently fall behind.
#
# The site pages are rewrites of docs/ (front matter, different structure,
# inline HTML), not generated copies, so every user-facing docs change has to
# be carried over by hand. This check makes forgetting that a failure.
#
# Usage: scripts/docs-sync-check.sh [base-ref]
#   base-ref defaults to the merge-base with origin/main (i.e. "this branch's
#   changes"); `just release-check` passes the last release tag instead.
#   Uncommitted changes in the working tree are included.
#
# Escape hatches, for a docs change with nothing to carry to the site (typo,
# internal note): a commit in the range with the trailer `Site-Sync: none`
# (works in CI), or SITE_SYNC=skip in the environment (local pre-commit).
set -e

if [ "${SITE_SYNC:-}" = "skip" ]; then
  echo "docs-sync-check: skipped (SITE_SYNC=skip)"
  exit 0
fi

base="${1:-}"
if [ -z "$base" ]; then
  base=$(git merge-base HEAD origin/main 2>/dev/null || git merge-base HEAD main 2>/dev/null || true)
fi
if [ -z "$base" ]; then
  echo "docs-sync-check: no base ref (no origin/main or main) — skipping"
  exit 0
fi

if git log --format=%B "$base"..HEAD 2>/dev/null | grep -qi '^Site-Sync: *none'; then
  echo "docs-sync-check: skipped (Site-Sync: none trailer in $base..HEAD)"
  exit 0
fi

changed=$(git diff --name-only "$base" -- docs site/content)
stale=""
for f in $(printf '%s\n' "$changed" | grep '^docs/[^/]*\.md$' || true); do
  twin="site/content/$(basename "$f")"
  [ -f "$twin" ] || continue
  if ! printf '%s\n' "$changed" | grep -qx "$twin"; then
    stale="$stale  $f changed, but $twin did not\n"
  fi
done

if [ -n "$stale" ]; then
  echo "ERROR: docs changed without their website twins (dippin.org would go stale):"
  printf "$stale"
  echo "Carry the user-facing change into each site/content page (same facts, site style)."
  echo "Nothing to carry? Add a 'Site-Sync: none' commit trailer, or SITE_SYNC=skip locally."
  exit 1
fi
echo "docs-sync-check: OK (base $(git rev-parse --short "$base^{commit}"))"
