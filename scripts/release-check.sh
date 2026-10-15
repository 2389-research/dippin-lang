#!/bin/sh
# ABOUTME: Pre-tag gate for a dippin-lang release: CHANGELOG entry present, docs
# ABOUTME: and website twins in sync since the last tag, no stale "current version" claims.
#
# Usage: scripts/release-check.sh vX.Y.Z
# Run on an up-to-date main (after the changelog PR merges), before `git tag`.
set -e

version="${1:-}"
case "$version" in
  v[0-9]*.[0-9]*.[0-9]*) ;;
  *) echo "usage: scripts/release-check.sh vX.Y.Z"; exit 2 ;;
esac

fail=0

# 1. The release is described in CHANGELOG.md (dippin.org/changelog is generated from it).
if grep -q "^## \[$version\]" CHANGELOG.md; then
  echo "ok: CHANGELOG.md has a [$version] entry"
else
  echo "ERROR: CHANGELOG.md has no '## [$version]' entry"
  fail=1
fi

# 2. Every docs/ change since the previous release reached its website twin.
last=$(git describe --tags --abbrev=0 2>/dev/null || true)
if [ -n "$last" ]; then
  if ! ./scripts/docs-sync-check.sh "$last"; then
    fail=1
  fi
fi

# 3. No hand-written page names an older release as the current one. Historical
#    notes ("added in v0.33.0") are fine; "currently/recommended/latest vX" is not.
stale=$(grep -nE '(currently|[Rr]ecommended|latest)[^.]{0,40}v0\.[0-9]+\.[0-9]+|@v0\.[0-9]+\.[0-9]+' \
  README.md AGENTS.md docs/*.md site/content/*.md site/static/skill.md 2>/dev/null \
  | grep -v "site/content/changelog.md" | grep -v "$version" || true)
if [ -n "$stale" ]; then
  echo "ERROR: pages claim an older release is current (use a placeholder or the releases link):"
  echo "$stale"
  fail=1
else
  echo "ok: no stale current-version claims"
fi

if [ "$fail" -ne 0 ]; then
  exit 1
fi
echo "release-check: $version is ready to tag"
