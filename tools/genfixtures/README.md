# genfixtures

`genfixtures` is the fixture generator for Nuri's fidelity test suite. It drives real
`vscode-textmate` (the same engine Shiki uses) over committed source samples, captures
token streams per line with resolved theme colors, and writes JSON fixtures to
`internal/fidelity/testdata/`. The Go fidelity runner then diffs Nuri's output against
these fixtures to measure correctness.

Regenerating fixtures is a deliberate, reviewed step. CI regenerates from the pinned
Shiki version and asserts `git diff --exit-code testdata/` to catch any drift.

## When to regenerate fixtures

Run this tool only when there is a concrete reason to change the committed fixtures.
The four situations that warrant regeneration are:

1. **Adding a grammar.** A new grammar is being added to `bundle/core` or `bundle/full`.
   Add it to the relevant config and regenerate to produce its golden fixture.

2. **Updating the pinned Shiki or grammar version.** The `grammars-themes` or
   `vscode-textmate` sibling repository is updated to a newer commit. Regenerate to
   capture any upstream tokenizer or grammar changes, then review the diff before committing.

3. **Changing the fixture schema.** The shape of the JSON output changes (for example,
   a new field is added to each token object). Regenerate all affected configs so the
   Go fidelity runner and the fixtures stay in sync.

4. **Adding a theme to a config.** A theme is added to one of the config files.
   Regenerate that config to produce the new theme columns in existing fixtures.

Do not regenerate as part of routine development or to silence a fidelity test failure.
A failing fidelity test means Nuri's output diverged from Shiki's. The correct response
is to fix the tokenizer or theme resolver, not to update the fixtures.

## Prerequisites

No `npm install` is required inside this directory. Two sibling repositories must be
present alongside the `irosashi` repo root:

| Repository | Expected path | Purpose |
|---|---|---|
| `vscode-textmate` | `../../../vscode-textmate` | TextMate tokenizer (must be built: `out/src/main.js` must exist) |
| `grammars-themes` | `../../../grammars-themes` | TextMate grammar JSON files, theme JSON files, and source samples |

The generator loads these via `require()` at startup. If either sibling is absent or
`vscode-textmate` has not been built, the script will exit with a module resolution error.

## Running the generator

```bash
# Generate fixtures using the default config (32 core grammars, 2 themes)
npm run generate

# Generate fixtures using a specific config file
node generate.mjs --config matrix.full.config.json
node generate.mjs --config matrix.all.config.json
node generate.mjs --config matrix.theme-stress.config.json
```

Each run prints a summary line per fixture file and a final count of generated vs skipped.
A grammar is skipped when no matching sample file is found in `samplesDir` or its scope
name cannot be resolved from the loaded grammar set.

## Running the unit tests

```bash
npm test
```

This runs `utf16_test.mjs`, which tests `buildUtf16ToUtf8Map`. That function converts
UTF-16 offsets per source line (as used internally by `vscode-textmate`) into UTF-8 byte
offsets used by the Go tokenizer.

## Config files

Four configs are included for common scenarios:

| File | Grammars | Themes | Output directory |
|---|---|---|---|
| `matrix.config.json` | 32 core grammars | github-light, github-dark | `internal/fidelity/testdata/golden` |
| `matrix.full.config.json` | 32 core grammars | 8 themes | `internal/fidelity/testdata/golden-full` |
| `matrix.all.config.json` | 234 grammars (full set) | github-light, github-dark | `internal/fidelity/testdata/golden-all` |
| `matrix.theme-stress.config.json` | typescript, markdown, html | 65 themes | `internal/fidelity/testdata/golden-theme-stress` |

### Config schema

```json
{
  "themes":     ["<theme-name>", "..."],
  "grammars":   ["<grammar-name>", "..."],
  "samplesDir": "<relative path to samples directory>",
  "outputDir":  "<relative path for generated fixture files>"
}
```

All paths are relative to the config file location. Theme and grammar names correspond to
filenames (without extension) inside the `grammars-themes` repository.

## Output fixture format

Each fixture file is named `{grammar}__{sampleBasename}.json` and contains:

```json
{
  "vsctmVersion": "9.x.x",
  "grammar": "typescript",
  "grammarSourceHash": "sha256:<hex>",
  "source": "<full source text of the sample file>",
  "themes": {
    "github-dark": {
      "tokens": [
        [
          {
            "start": 0,
            "end": 6,
            "text": "import",
            "scopes": ["source.ts", "keyword.control.import.ts"],
            "color": "#cf222e",
            "fontStyle": 0
          }
        ]
      ],
      "html": ""
    }
  }
}
```

The outer `tokens` array has one entry per source line. Each entry is an array of token
objects for that line. All `start` and `end` values are UTF-8 byte offsets. The `html`
field is reserved for future use and is always an empty string.

`fontStyle` is a bitmask encoded by `vscode-textmate`: bit 0 = italic, bit 1 = bold,
bit 2 = underline, bit 3 = strikethrough.

## File overview

| File | Purpose |
|---|---|
| `generate.mjs` | Main entry point; reads config, tokenizes samples, writes fixtures |
| `lib/textmate-common.mjs` | Shared setup: loads `vscode-textmate`, builds the grammar registry, loads themes |
| `utf16_test.mjs` | Unit tests for the UTF-16 to UTF-8 offset conversion map |
| `matrix.config.json` | Default config: 32 core grammars, 2 themes |
| `matrix.full.config.json` | Full theme matrix: 32 core grammars, 8 themes |
| `matrix.all.config.json` | All grammars: 234 grammars, 2 themes |
| `matrix.theme-stress.config.json` | Theme stress: 3 grammars, 65 themes |
