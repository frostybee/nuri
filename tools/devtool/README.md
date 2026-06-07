# devtool

Cross-platform Go CLI for managing Nuri's grammar/theme assets and generating fidelity test fixtures.

## Prerequisites

- Go 1.25+
- Node.js 22+ (for fixture generation only)
- Git submodules initialized (`git submodule update --init --recursive`)

## Commands

### sync

Copies grammar and theme JSON files from the `grammars-themes` submodule into `bundle/core/` and `bundle/full/`.

```
go run ./tools/devtool sync
```

- Core grammars are defined in `bundle/core/grammars.txt` (one name per line)
- All grammars from the submodule go into `bundle/full/grammars/`
- Themes are copied into both `bundle/core/themes/` and `bundle/full/themes/`
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

### all

Runs `sync` then `generate`.

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
