---
title: Shiki Compatibility
description: "What Nuri ports from vscode-textmate, how it differs, and the upstream source references."
sidebar:
  order: 2
---

Nuri is a Go port of Shiki's core pipeline. It re-implements the vscode-textmate TextMate grammar engine in Go and runs the same Oniguruma regex engine (compiled to WASM via wazero). The goal is byte-for-byte identical token output for every supported grammar and theme.

This page describes what was ported, what differs by design, and which upstream repositories pin the behavior.

## Ported features

### Rule model

All six vscode-textmate rule types are implemented in `internal/grammar/rules.go`:

- **MatchRule**: single-match pattern with numbered capture groups.
- **BeginEndRule**: paired span rule with `begin`/`end` patterns, `beginCaptures`/`endCaptures`, `contentName`, and `applyEndPatternLast`.
- **BeginWhileRule**: per-line while-condition rule with `begin`/`while` patterns and `whileCaptures`.
- **EndRule** / **WhileRule**: created at match time from their parent rules, carrying the resolved pattern after backref substitution.
- **IncludeRule**: handles `$self`, `$base`, `#key`, `source.scope`, and `source.scope#key`.
- **CollectionRule**: container with a local `repository` that shadows the parent during include resolution.

Backref substitution (`\1`-`\9`) in end and while patterns is implemented in `internal/grammar/compile.go`. Scope name substitution (`$1`, `${1:/downcase}`, `${1:/upcase}`) mirrors vscode-textmate's `RegexSource.replaceCaptures`. If `beginCaptures` is absent, `captures` applies to the begin match (vscode-textmate behavior).

### Include resolution

All four include forms are supported: `$self`, `$base`, `#key`, and cross-grammar includes (`source.scope`, `source.scope#key`). The resolver in `internal/grammar/compile.go` tracks a visited set for cycle detection (`ErrGrammarCycle`) and enforces a depth limit of 256 (`ErrGrammarDepth`). Rules whose sub-patterns all fail to resolve are skipped, matching vscode-textmate's `hasMissingPatterns` check.

### Injection selectors

The full injection selector grammar is implemented in `internal/grammar/selector.go`:

- Comma-separated composite selectors
- `L:` and `R:` priority prefixes (left injection wins on tie vs. grammar rule wins)
- Negation (`-`) expressions
- Parenthesized groups with `|` alternatives
- Dot-boundary prefix scope matching
- Positive expressions matched as an ordered subsequence of the scope stack (right-to-left), matching vscode-textmate's `nameMatcher`

The injection selector parser was ported from Phiki (PHP), adapted to Go.

### Per-line tokenization

The tokenizer state machine in `internal/tokenizer/line.go` mirrors vscode-textmate's `_tokenizeString`:

- A sentinel `\n` is appended to every line before scanning, matching vscode-textmate's behavior (grammar.ts:380). Anchors like `$` and lookaheads in end patterns interact with this newline.
- `\G` anchor tracking via `anchorPosition`. `SearchOptionNotBeginPosition` is set when the scan position differs from the anchor position.
- `isFirstLine` drives `SearchOptionNotBeginString` (first line has no preceding text, so `^` matches at position 0).
- `capturedEOL` flag: set when a begin match consumes the full scan string, seeding `anchorPosition = 0` for while-rule checking on the next line.
- Four infinite-loop guards, each citing the corresponding vscode-textmate line numbers in code comments.
- `pickBestMatch`: injection vs. grammar rule competition. Earlier start position wins. On a tie, `L:` priority lets the injection win; otherwise the grammar rule wins.

### While-condition checking

`checkWhileConditions` in `internal/tokenizer/line.go` mirrors vscode-textmate's `_checkWhileConditions`: it collects all while-rule frames bottom-to-top, checks in outermost-first order, pops all frames up to and including a failing frame, emits tokens for while captures, and re-seeds the anchor position from the while-match end.

### Capture re-tokenization

Captures with nested `Patterns` are recursively tokenized via `retokenizeCapture` in `internal/tokenizer/captures.go`. The line is truncated to the capture end (preserving absolute byte offsets, not a substring), matching vscode-textmate's approach. Unmatched optional capture groups (`Start < 0`) are skipped. When a cross-grammar rule has captures with patterns, includes resolve against the rule's grammar, not the root grammar.

### State stack

The `StateStack` in `internal/tokenizer/state.go` is a per-frame linked list carrying the rule, content grammar (for cross-grammar contexts), scope names, end/while rules, anchor position, and enter position. `scopeSlice()` walks all frames and splits space-separated scope names into individual entries, matching vscode-textmate's behavior for multi-word scope names.

### Theme resolution

VS Code theme parsing is ported in the `theme/` package. Scope matching uses the same specificity scoring as vscode-textmate: stack depth (deeper match wins), dot-segment count (more specific selector wins), and parent-scope part count (contextual match wins). The `FontStyle` bitmask (italic=1, bold=2, underline=4, strikethrough=8) uses the same encoding as vscode-textmate's binary token format.

### HTML output and transformers

HTML rendering produces the same `<pre><code><span>` structure as Shiki's `codeToHtml`. Multi-theme CSS variable emission (`--nuri-{key}`, `--nuri-{key}-bg`) parallels Shiki's multi-theme output. The `styleToClass` mode produces deterministic hashed class names.

The transformer hook pipeline matches `@shikijs/transformers`: notation annotations (`[!code ++]`, `[!code highlight]`, `[!code focus]`, etc.), meta string ranges, and visible whitespace rendering are all ported with the same hook sequence (Preprocess, Tokens, Span, Line, Code, Pre, Root, Postprocess).

## What differs

### Go API

Shiki uses JavaScript async callbacks and returns HAST (hypertext AST). Nuri uses `context.Context` for cancellation, `fs.FS` for asset loading, `error` returns, and `io.WriterTo` for output serialization. The lifecycle is `nuri.New()` / `h.Close()` instead of `createHighlighter`.

### Concurrency model

Shiki runs in Node.js (single-threaded event loop). Nuri uses a bounded LIFO pool of WASM instances for safe concurrent highlighting. See [Architecture](/docs/internals/architecture) for pool design details.

### Safety features

Nuri adds four safety mechanisms not present in Shiki:

- Per-line soft timeout (`WithTimeoutMs`): on expiry, the WASM instance is poisoned and replaced; the line emits as a single unstyled token with a `Diagnostic`.
- Maximum line length pre-filter (`WithMaxLineLength`): lines over the threshold skip tokenization entirely, catching minified one-liners before the timeout path.
- Panic recovery with instance poison-swap: a panic during tokenization emits a `Diagnostic` and continues on the next line.
- WCAG 2.1 contrast correction (`WithMinContrast`): automatic foreground color adjustment at theme load time.

### No `codeToHast`

Shiki returns a HAST tree via `codeToHast`. Nuri returns a Go `*Element` tree with an equivalent structure. Consumers serialize it via `WriteTo(io.Writer)`.

### No semantic tokens

VS Code's LSP-driven `semanticTokenColors` field in themes is parsed and silently ignored. Nuri does not implement semantic token highlighting.

### Whole-buffer API only

`h.CodeToTokens()` tokenizes the entire input. There is no incremental `TokenizeLine` entry point. The internal `StateStack` supports per-line re-entry, but no public API exposes it. This may be added in a future version if there is demand.

### Deterministic output

`Element.Styles` and `Element.Attrs` maps are emitted in sorted key order. Shiki's output order depends on JavaScript engine map iteration order, which is not guaranteed stable across engines or versions.

### `$base` simplification

`$base` currently resolves like `$self`, returning the root grammar's patterns. In vscode-textmate, `$base` refers to the grammar from which the current grammar was included as an embedded language. This is a known simplification documented in `internal/tokenizer/memo.go`.

### Additional output formats

Nuri adds ANSI terminal (with 8/16/256/truecolor depth), SVG, JSON, and plain text output formats beyond Shiki's HTML, HAST, and raw tokens.

## Upstream references

Three upstream repositories define Nuri's behavior. All three are pinned in `provenance.lock.json`.

**vscode-textmate** (TypeScript) is the reference tokenizer implementation. The tokenizer state machine in `internal/tokenizer/` is a direct port. Code comments throughout the tokenizer cite specific vscode-textmate line numbers for each ported behavior. The kotlin-textmate pure Kotlin port served as a useful cross-reference during development.

**shikijs/textmate-grammars-themes** provides all 257 grammars and 65 themes. Per-file SHA-256 hashes in the provenance lockfile verify that the working tree matches the pinned version.

**onig.wasm** is the Oniguruma C library compiled to WebAssembly (~400KB). Its SHA-256 is pinned in the provenance lockfile. This is the same binary Shiki uses, giving bug-for-bug regex compatibility.

The lockfile ensures that a `go test` run and a Node.js Shiki run use exactly the same grammars, themes, and regex engine. See [Fidelity Testing](/docs/internals/fidelity-testing) for how the lockfile is verified in CI.
