# Response to Downstream Consumer Field Report (2026-03-25)

Hey team —

Thanks for the thorough field report from the build_product.dip test writing session. All five issues you raised have been addressed. Here's what shipped and where to find it.

---

## 1. preferred_label silently ignored on freeform gates

**Fixed in v0.9.0.** The scenario key `preferred_label` (or per-node `Gate.preferred_label`) now matches against edge labels with case-insensitive substring matching. This works on both `choice` and `freeform` gates. Previously it was silently ignored — sorry about that.

**How to use it:**
```json
{
  "scenario": {
    "ApprovePlan.preferred_label": "approve"
  }
}
```

The simulator will select the edge whose label contains "approve" (case-insensitive).

---

## 2. mode:labeled / --- content blocks

**Closed as won't-fix.** Only `choice` and `freeform` modes are supported. `mode: labeled` was never implemented and the `---` content block syntax isn't part of the grammar. We've documented this in `docs/nodes.md` — use `preferred_label` on edges for human gate routing instead.

If you have existing `.dip` files using `mode: labeled`, switch them to `mode: choice` with edge labels. The behavior is equivalent.

---

## 3. immediately_after assertion

**Added in v0.9.0.** New test assertion field:

```json
{
  "expect": {
    "immediately_after": {
      "ReviewConsensus": "Postmortem"
    }
  }
}
```

This checks that `Postmortem` is the very next node after `ReviewConsensus` in the execution path — stricter than `path_contains` which only checks ordering.

---

## 4. Tool defaults mask fallback edges

**Fixed in v0.9.0.** Set an empty string in the scenario to suppress the auto-seeded `success` default:

```json
{
  "scenario": {
    "BuildGate.tool_stdout": ""
  }
}
```

This lets unconditional fallback edges fire. Documented in `docs/testing.md` under "Clearing Defaults".

---

## 5. not_visited fragility with loop-breaking

**Documented in v0.9.0.** `docs/testing.md` now has a Caveats section explaining why `not_visited` can be fragile when `MaxNodeVisits` loop-breaking kicks in (it continues execution rather than stopping, so downstream nodes may get visited unexpectedly).

**Recommendation:** Use `path_contains` for edge-routing assertions and reserve `not_visited` for nodes you're confident the simulator can't reach via any path.

---

## What else shipped since your report

- **v0.9.0**: `branch` field for targeted parallel testing, simulator walks all parallel branches
- **v0.10.0**: DIP123/124/125 tool command lint rules, brochure site
- **v0.11.0**: DIP126 subgraph ref validation, `dippin watch`, `dippin test --coverage`, tree-sitter grammar, WASM playground
- **v0.12.0**: Blog with 5 guides, SEO, two-tier nav

Upgrade with:
```sh
go install github.com/2389-research/dippin-lang/cmd/dippin@latest
```

Or Homebrew:
```sh
brew upgrade dippin
```

Let us know if any of the fixes don't work as expected with your test suites, or if you hit new issues.

— Dippin team
