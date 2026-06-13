# devtool

Cross-platform Go CLI for managing Nuri's grammar/theme assets and generating fidelity test fixtures.

## Prerequisites

- Go 1.25+
- Node.js 22+ (for fixture generation only)
- Git submodules initialized (`git submodule update --init --recursive`)

## Typical workflows

**Pulling in upstream grammar or theme changes.** Update the `grammars-themes` submodule, then run `sync`. That one command re-minifies and re-compresses the assets, regenerates the lockfile, and rewrites `THIRD-PARTY-NOTICE`. Commit `bundle/`, `provenance.lock.json`, and `THIRD-PARTY-NOTICE` together as a single change.

**Regenerating fidelity fixtures.** Run `generate` after any change to the tokenizer, grammar loading, or theme resolver that might shift token output. Pass `-- --config matrix.full.config.json` (or another config) when you need coverage beyond the default matrix. Fixtures are committed, so a clean `git diff` after regeneration means nothing changed.

**Full refresh.** Run `all` when you want to pull upstream changes and regenerate fixtures in one shot. It is just `sync` followed by `generate`.

**Checking provenance in CI.** Run `verify` to confirm the working tree matches the committed lockfile. It catches submodule drift, edited assets, a stale index, and a missing `THIRD-PARTY-NOTICE` all in one pass.

**Lockfile or NOTICE out of date without a full sync.** If you edited assets by hand, pinned the submodule to a new commit manually, or added a Go dependency, run `lock` and then `notices` to bring those files back in sync without re-running `sync`.

## Commands

### sync

Copies grammar and theme JSON files from the `grammars-themes` submodule into `bundle/core/` and `bundle/full/`, then automatically runs `lock` and `notices`.

```
go run ./tools/devtool sync
```

- Core grammars are defined in `bundle/core/grammars.txt` (one name per line)
- All grammars from the submodule go into `bundle/full/grammars/`
- Themes are copied into both `bundle/core/themes/` and `bundle/full/themes/`
- Assets are minified, pruned of unused fields, and gzip-compressed (`.json.gz`)
- A metadata index (`grammars/index.json.gz`) is regenerated for each bundle
- After syncing, commit any changes to `bundle/` (required for `go:embed`)

### generate

Builds vscode-textmate and generates golden test fixtures for the fidelity test suite.

```
go run ./tools/devtool generate
```

This runs three steps:
1. `git submodule update --init --recursive`
2. `npm install && npm run compile` in `vscode-textmate/`
3. `npm ci && node generate.mjs` in `tools/genfixtures/`

Pass extra arguments after `--` to forward them to `generate.mjs`:

```
go run ./tools/devtool generate -- --config matrix.full.config.json
```

### lock

Generates `provenance.lock.json` at the repo root. Records the current submodule commits and URLs, SHA256 hashes of every asset in `bundle/full/`, and the SHA256 of `onig.wasm`.

```
go run ./tools/devtool lock
```

`sync` calls this automatically. Run it manually after editing assets by hand or updating submodule pins without a full sync.

### notices

Generates `THIRD-PARTY-NOTICE` at the repo root from the upstream NOTICE files in the `grammars-themes` submodule and the versions of Go module dependencies (`wazero`, `golang.org/x/term`, `golang.org/x/sys`). Requires `provenance.lock.json` to exist (run `lock` first if it is missing).

```
go run ./tools/devtool notices
```

`sync` calls this automatically.

### verify

Checks `provenance.lock.json` against the current repo state and exits non-zero on any mismatch.

```
go run ./tools/devtool verify
```

Checks performed:
- Submodule commits and URLs match the lockfile
- SHA256 of `onig.wasm` matches
- SHA256 of every file in `bundle/full/grammars/` and `bundle/full/themes/` matches
- Every file in `bundle/core/grammars/` is byte-identical to its counterpart in `bundle/full/`
- Grammar metadata indexes (`index.json.gz`) are consistent with the files on disk
- `THIRD-PARTY-NOTICE` exists and is non-empty

Run this in CI after pulling to confirm the working tree matches the committed lockfile.

### all

Runs `sync` (which includes `lock` and `notices`) then `generate`.

```
go run ./tools/devtool all
```

## Running Fidelity Tests

After generating fixtures, run the Go fidelity tests to compare Nuri's output against the vscode-textmate reference:

```bash
# Verify fixture JSON format loads correctly
go test ./internal/fidelity/... -run "TestLoadFixture$|TestLoadFixtures$" -v

# Core matrix: compare Nuri vs vscode-textmate (10 grammars, 2 themes)
go run ./tools/devtool generate
go test ./internal/fidelity/... -run TestGoldenFidelity -v -count=1

# Full matrix (32 grammars, 8 themes)
go run ./tools/devtool generate -- --config matrix.full.config.json
go test ./internal/fidelity/... -run TestGoldenFidelityFull -v -count=1

# Theme stress matrix (3 grammars, 65 themes)
go run ./tools/devtool generate -- --config matrix.theme-stress.config.json
go test ./internal/fidelity/... -run TestGoldenFidelityThemeStress -v -count=1
```

## Submodules

| Submodule | Source | Purpose |
|-----------|--------|---------|
| `grammars-themes/` | [shikijs/textmate-grammars-themes](https://github.com/shikijs/textmate-grammars-themes) | Grammar JSONs, theme JSONs, code samples |
| `vscode-textmate/` | [microsoft/vscode-textmate](https://github.com/microsoft/vscode-textmate) | Reference tokenizer for generating golden fixtures |

Both are pinned to specific commits. To update:

```
cd grammars-themes && git pull origin main && cd ..
cd vscode-textmate && git pull origin main && cd ..
git add grammars-themes vscode-textmate
git commit -m "update submodules"
```

Then re-run `sync` and `generate` to propagate changes.
