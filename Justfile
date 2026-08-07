# dippin-lang development tasks

# List available recipes
default:
    @just --list

# Build the dippin binary
build: gen-spec
    go build -ldflags "-X main.version=dev -X main.commit=$(git rev-parse --short HEAD) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o dippin ./cmd/dippin/

# Install dippin globally to $GOBIN (injects version info)
install:
    go install -ldflags "-X main.version=dev -X main.commit=$(git rev-parse --short HEAD) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" ./cmd/dippin/

# Generate the language spec from source docs
gen-spec:
    ./scripts/gen-spec.sh

# Run all tests
test:
    go test ./... -count=1

# Run all tests with race detector
test-race:
    go test ./... -count=1 -race

# Run tests for a specific package (e.g., just test-pkg validator)
test-pkg pkg:
    go test ./{{pkg}}/... -count=1 -v

# Run go vet
vet:
    go vet ./...

# Run golangci-lint (same as CI)
lint-go:
    golangci-lint run --timeout=10m

# Check formatting (exit 1 if unformatted)
fmt-check:
    @unformatted=$(gofmt -l . 2>&1); \
    if [ -n "$unformatted" ]; then \
        echo "Files not gofmt'd:"; \
        echo "$unformatted"; \
        exit 1; \
    fi

# Format all Go files
fmt:
    gofmt -w .

# Validate all example .dip files
validate-examples: build
    @for f in examples/*.dip; do \
        ./dippin validate "$f" > /dev/null 2>&1 || { \
            echo "FAIL: $f"; \
            ./dippin validate "$f" 2>&1; \
            exit 1; \
        }; \
    done
    @echo "All examples valid."

# Lint all example .dip files
lint-examples: build
    @for f in examples/*.dip; do \
        echo "--- $f ---"; \
        ./dippin lint "$f" 2>&1 || true; \
    done

# Pack every example via dipx and verify round-trip
pack-examples: build
    #!/usr/bin/env bash
    set -euo pipefail
    for f in $(find examples -name '*.dip'); do
        echo "Packing $f"
        ./dippin pack -o "$(mktemp -u).dipx" "$f"
    done

# Validate that docs/GRAMMAR.ebnf is valid W3C EBNF (the bottlecaps/XML-spec
# dialect): `name ::= ...` productions, balanced groups, closed comments/strings/
# char-classes, and all nonterminals defined.
validate-grammar:
    go test ./ebnf/ -count=1
    @echo "GRAMMAR.ebnf is valid W3C EBNF."

# Check cyclomatic complexity (max 5 per function, excluding tests)
complexity:
    @violations=$( gocyclo -over 5 . | grep -v _test.go ); \
    if [ -n "$violations" ]; then \
        echo "Cyclomatic complexity violations (max 5):"; \
        echo "$violations"; \
        exit 1; \
    fi
    @violations=$( gocognit -over 7 . | grep -v _test.go ); \
    if [ -n "$violations" ]; then \
        echo "Cognitive complexity violations (max 7):"; \
        echo "$violations"; \
        exit 1; \
    fi
    @echo "Complexity OK."

# Regenerate spec from source docs (always fresh)
spec-check:
    @./scripts/gen-spec.sh
    @echo "Spec regenerated."

# Run release invariant checks (git tracking, freshness, tarball build)
releasecheck:
    go test ./releasecheck/ -count=1 -race

# Run the full pre-commit check suite (mirrors CI exactly)
check: spec-check build vet fmt-check lint-go test-race releasecheck complexity validate-examples validate-grammar pack-examples tree-sitter-test
    @echo "All checks passed."

# Generate test coverage report (excludes untestable files: main.go, cmd_lsp.go)
cover:
    go test ./... -coverprofile=cover.out
    grep -v -E '(cmd/dippin/main\.go|cmd/dippin/cmd_lsp\.go)' cover.out > cover_filtered.out || true
    go tool cover -func=cover_filtered.out

# Open coverage in browser
cover-html:
    go test ./... -coverprofile=cover.out
    grep -v -E '(cmd/dippin/main\.go|cmd/dippin/cmd_lsp\.go)' cover.out > cover_filtered.out || true
    go tool cover -html=cover_filtered.out

# Install the pre-commit hook
setup-hooks:
    ln -sf ../../scripts/pre-commit .git/hooks/pre-commit
    chmod +x .git/hooks/pre-commit
    @echo "Pre-commit hook installed."

# Tag a semver release and push (e.g., just release v0.8.0 "add CLI integration tests")
release tag msg:
    git tag -a {{tag}} -m "{{msg}}"
    git push origin {{tag}}

# Start Hugo dev server (builds WASM and changelog first)
site-serve: build wasm changelog-md
    cd site && hugo serve

# Regenerate site/content/changelog.md from CHANGELOG.md
changelog-md:
    ./scripts/gen-changelog.sh
    @echo "Generated site/content/changelog.md"

# Build the Hugo site for production
site-build: build wasm changelog-md
    cp docs/generated-spec.md site/static/llms-full.txt
    cd site && hugo --minify

# Build WASM binary for the browser playground
wasm:
    GOOS=js GOARCH=wasm go build -o site/static/dippin.wasm ./cmd/wasm/
    cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" site/static/wasm_exec.js
    @echo "WASM built: site/static/dippin.wasm"

# Regenerate tree-sitter parser from grammar.js
tree-sitter-generate:
    cd editors/tree-sitter-dippin && npx tree-sitter generate

# Run tree-sitter corpus tests (after regenerate)
tree-sitter-test: tree-sitter-generate
    cd editors/tree-sitter-dippin && npx tree-sitter test

# Clean build artifacts
clean:
    rm -f dippin cover.out cover_filtered.out cover_check.out site/static/dippin.wasm site/static/wasm_exec.js
    rm -rf site/public
    @echo "Cleaned."

# Report catalog entries overdue for price re-verification (as_of age)
check-prices:
    go run ./cmd/pricing-sync check

# Diff the catalog against models.dev and report candidate changes (report-only)
sync-prices:
    go run ./cmd/pricing-sync sync
