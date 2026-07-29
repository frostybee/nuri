---
title: Fidelity Testing
description: "How Nuri measures and gates correctness against real Shiki output."
sidebar:
  order: 3
---

Fidelity is the core value proposition: Nuri must produce identical output to real Shiki. The testing system generates reference token streams from vscode-textmate (Node.js), compares Nuri's output byte-for-byte, classifies any mismatches to isolate which component has the bug, and gates CI on the results.

Ground truth is always real Shiki, not hand-written expectations.

## How it works

Two steps:

1. `tools/genfixtures` (a Node.js CLI) drives real vscode-textmate over committed source samples, captures per-token colors and scopes, and writes JSON fixtures to `internal/fidelity/testdata/`.
2. `internal/fidelity/golden_test.go` loads those fixtures, runs Nuri over the same source with contrast correction disabled (`WithMinContrast(0)`), and diffs the output token by token.

The fixture generator is the only Shiki/Node.js dependency. It is development tooling only, never part of the Go module or binary. Fixtures are committed to the repository. Failing a fidelity test means Nuri's output diverged from Shiki's. The correct response is to fix the tokenizer or theme resolver, not to update the fixtures.

## Fixture format

Each fixture is a JSON file named `{grammar}__{sampleBasename}.json`:

```go
type Fixture struct {
    VsctmVersion      string                    // vscode-textmate npm version
    Grammar           string                    // canonical grammar name
    GrammarSourceHash string                    // SHA-256 of the raw sample file
    Source            string                    // full source text
    Themes            map[string]ThemeFixture   // theme name to fixture
}

type ThemeFixture struct {
    Tokens [][]FixtureToken
    HTML   string
}

type FixtureToken struct {
    Start     int        // UTF-8 byte offset from line start
    End       int        // UTF-8 byte offset from line start
    Text      string     // exact token text
    Scopes    []string   // TextMate scope stack, outermost first
    Color     string     // resolved foreground hex (e.g. "#D73A49")
    FontStyle int        // bitmask: 1=italic, 2=bold, 4=underline, 8=strikethrough
}
```

Token `Start` and `End` are UTF-8 byte offsets, not rune counts. The fixture generator translates vscode-textmate's internal UTF-16 offsets to UTF-8 during generation. The `HTML` field is reserved for future use and is always empty in generated fixtures.

## Fixture generator

The generator lives in `tools/genfixtures/`. It is a Node.js ES module that requires two sibling repositories: `vscode-textmate` (built) and `textmate-grammars-themes` (source samples and grammar/theme JSON). Both are pinned by the provenance lockfile.

For each grammar, the generator:

1. Locates the sample file by trying `{grammar}.sample`, `{grammar}.txt`, and subdirectory variants.
2. SHA-256 hashes the raw sample for the `GrammarSourceHash` field.
3. Loads the grammar into the vscode-textmate registry.
4. For each theme: sets the theme, calls `tokenizeLine` (scope names) and `tokenizeLine2` (binary-encoded color and fontStyle), correlates the two results, and converts UTF-16 offsets to UTF-8 bytes.
5. Writes the fixture JSON with themes sorted alphabetically.

The Go wrapper `tools/devtool generate` initializes submodules, builds vscode-textmate if needed, and invokes `node generate.mjs`. Config flags after `--` select the fixture matrix.

Regeneration is a deliberate, reviewed step. Regenerate only when adding a grammar to a bundle, updating the pinned Shiki or grammar version, changing the fixture schema, or adding a theme to a config.

## Fixture matrices

Four configurations control which grammar/theme combinations are tested:

| Config | Grammars | Themes | CI trigger |
|---|---|---|---|
| `matrix.config.json` | 32 core | 2 (github-dark, github-light) | Every push and PR |
| `matrix.full.config.json` | 32 core | 8 | Push to main |
| `matrix.all.config.json` | 234 | 2 | Push to main |
| `matrix.theme-stress.config.json` | 3 (ts, md, html) | 65 | Nightly (informational) |

Each config produces fixtures in a separate `testdata/` subdirectory: `golden`, `golden-full`, `golden-all`, `golden-theme-stress`.

## Mismatch classification

The comparison engine in `internal/fidelity/compare.go` classifies mismatches into six `DiffKind` values:

| Kind | What it means | Which component to investigate |
|---|---|---|
| `BoundaryMismatch` | Token start/end byte offsets differ | Tokenizer state machine (`internal/tokenizer/`) |
| `ScopeMismatch` | Same byte span, different scope stack | Tokenizer or grammar registry (include resolution, backref, `\G` anchor) |
| `StyleMismatch` | Same span and scopes, different color or fontStyle | Theme resolver (`theme/matcher.go`) |
| `MissingToken` | Token in expected, absent in actual | Tokenizer (omitted output) |
| `ExtraToken` | Token in actual, absent in expected | Tokenizer (extra output) |
| `HTMLMismatch` | HTML output strings differ | Renderer or serializer |

`CompareThemeTokens` walks expected and actual token arrays in lockstep. `BoundaryMismatch` causes an early return from the line comparison because all subsequent tokens are misaligned. `ScopeMismatch` and `StyleMismatch` advance both pointers and continue checking the rest of the line. Color comparison is case-insensitive and normalizes 3-digit hex shorthand to 6-digit before comparing.

`CompareHTML` finds the first divergent byte and provides a 30-byte context window on both sides for human-readable diagnostics.

## Score and report

The unit of measurement is a **triple**: one (grammar, theme) pair from a single fixture. A triple passes if and only if `CompareThemeTokens` returns zero diffs.

`ComputeReport` in `internal/fidelity/score.go` aggregates triples into a `FidelityReport` with per-grammar scores, per-theme scores, and a global pass rate. `GrammarScore.Rate()` returns a `float64` in [0, 1].

`RenderMarkdown` in `internal/fidelity/report.go` produces `FIDELITY.md`. Each grammar row shows a pass/fail indicator per theme column and a status:

- **shipping**: all theme columns pass. The grammar is eligible for `bundle/core`.
- **held**: any theme column fails. The grammar stays out of core until it reaches 100%.

Current state: 64/64 pass (100%) across the 32 core grammars and 2 themes.

`TestFidelityReport` reads the current `FIDELITY.md` and fails if it is stale. Regenerate with:

```bash
go test ./internal/fidelity/... -run TestFidelityReport -args -update
```

## CI gating

Four test functions in `internal/fidelity/golden_test.go` drive the fixture suites:

- `TestGoldenFidelity` runs the core matrix (32 grammars, 2 themes). Gate for every push and PR.
- `TestGoldenFidelityFull` runs the extended matrix (32 grammars, 8 themes). Gate for push to main.
- `TestGoldenFidelityAll` runs all grammars (234 grammars, 2 themes) using `full.FS()`. Gate for push to main.
- `TestGoldenFidelityThemeStress` runs the theme stress test (3 grammars, 65 themes). Nightly, informational (`continue-on-error`).

`TestCoreOnlyShipsGreen` is a separate gate that fails if any grammar in `bundle/core/grammars/` has a fidelity rate below 100% in the current fixture matrix. This prevents held grammars from shipping in the core bundle.

CI also runs two checks before tests:

**Provenance verification**: `go run ./tools/devtool verify` computes SHA-256 hashes for all grammar files in the working tree and compares them against `provenance.lock.json`. Any mismatch fails the build before tests run.

**Fixture freshness**: CI regenerates fixtures from the pinned versions and asserts `git diff --exit-code testdata/golden/`. A non-empty diff means the committed fixtures do not match the pinned Shiki version.

The `full-gate.yml` workflow auto-commits an updated `FIDELITY.md` with `[skip ci]` when the report changes after a push to main.

## Provenance lockfile

`provenance.lock.json` pins three categories of upstream inputs:

**Submodule commits**: the exact git SHA for `shikijs/textmate-grammars-themes` (all grammars and themes) and `microsoft/vscode-textmate` (the reference tokenizer).

**WASM artifact hash**: SHA-256 of `resources/wasm/onig.wasm`, the Oniguruma binary.

**Per-file hashes**: SHA-256 for every `.json.gz` grammar and theme file in both bundles. These verify that the working tree matches the pinned submodule commit without requiring a full submodule checkout.

`go run ./tools/devtool verify` re-derives all hashes from disk and compares them against the lockfile. It also verifies that core bundle grammars are byte-identical to their full bundle counterparts and that the metadata index is consistent. Regenerate the lockfile with `go run ./tools/devtool lock` after any grammar or theme update.
