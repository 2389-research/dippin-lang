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
# internal note): the trailer `Site-Sync: none` on the commit(s) that changed
# that docs file (works in CI; exempts only the files those commits touched),
# or SITE_SYNC=skip in the environment (local pre-commit).
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

# exempt_by_trailer FILE: true when every commit in range that touched FILE
# carries a `Site-Sync: none` trailer and the working tree adds nothing
# uncommitted to it. The trailer vouches for that commit's own docs edits
# only, never for the whole range (a release spans many PRs).
exempt_by_trailer() {
  git diff --quiet HEAD -- "$1" 2>/dev/null || return 1
  shas=$(git log --format=%H "$base"..HEAD -- "$1" 2>/dev/null)
  [ -n "$shas" ] || return 1
  for sha in $shas; do
    git log -1 --format=%B "$sha" | grep -qi '^Site-Sync: *none' || return 1
  done
  return 0
}

changed=$(git diff --name-only "$base" -- docs site/content)
stale=""
for f in $(printf '%s\n' "$changed" | grep '^docs/[^/]*\.md$' || true); do
  twin="site/content/$(basename "$f")"
  [ -f "$twin" ] || continue
  if printf '%s\n' "$changed" | grep -qx "$twin"; then
    continue
  fi
  if exempt_by_trailer "$f"; then
    echo "docs-sync-check: $f exempt (Site-Sync: none on every commit touching it)"
    continue
  fi
  stale="$stale  $f changed, but $twin did not\n"
done

if [ -n "$stale" ]; then
  echo "ERROR: docs changed without their website twins (dippin.org would go stale):"
  printf "$stale"
  echo "Carry the user-facing change into each site/content page (same facts, site style)."
  echo "Nothing to carry? Add a 'Site-Sync: none' commit trailer, or SITE_SYNC=skip locally."
  exit 1
fi
echo "docs-sync-check: OK (base $(git rev-parse --short "$base^{commit}"))"
