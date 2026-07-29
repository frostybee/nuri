---
title: Configuration
description: "Every constructor option for nuri.New()."
sidebar:
  order: 2
---

This page documents every `Option` passed to `nuri.New()`. For per-call options structs, see [Types](/docs/reference/types).

## Filesystem options

### `nuri.WithFS`

```go
func WithFS(fsys fs.FS) Option
```

Sets both the grammar and theme filesystems from a single `fs.FS` containing `grammars/` and `themes/` subdirectories. This is the intended way to use the bundle packages.

**Default:** `nil`

```go
h, err := nuri.New(ctx, nuri.WithFS(core.FS()))
```

### `nuri.WithGrammarFS`

```go
func WithGrammarFS(fsys fs.FS) Option
```

Sets the filesystem for grammar JSON files. Use when grammar and theme files live in separate locations.

**Default:** `nil`

### `nuri.WithThemeFS`

```go
func WithThemeFS(fsys fs.FS) Option
```

Sets the filesystem for theme JSON files. Use when grammar and theme files live in separate locations.

**Default:** `nil`

## Registration options

### `nuri.WithGrammar`

```go
func WithGrammar(name string, data []byte) Option
```

Registers a custom grammar from raw TextMate grammar JSON bytes at construction time. Returns an error from `nuri.New()` if the JSON is malformed or the grammar fails validation. Multiple calls register multiple grammars.

For runtime registration after construction, use `h.LoadLanguage()`.

### `nuri.WithTheme`

```go
func WithTheme(name string, data []byte) Option
```

Registers a custom theme from raw VS Code theme JSON bytes at construction time. Returns an error from `nuri.New()` if the JSON is malformed or the theme fails validation. Multiple calls register multiple themes.

For runtime registration after construction, use `h.LoadTheme()`.

### `nuri.WithAlias`

```go
func WithAlias(alias, target string) Option
```

Maps a language alias to a canonical name at construction time. For example, `nuri.WithAlias("sh", "shellscript")` makes `Lang: "sh"` resolve to the `shellscript` grammar.

For runtime registration after construction, use `h.RegisterAlias()`.

### `nuri.WithExtension`

```go
func WithExtension(ext, lang string) Option
```

Maps a file extension (without the leading dot) to a language name at construction time. Overrides any existing mapping for that extension.

For runtime registration after construction, use `h.RegisterExtension()`.

## Performance options

For rationale and architecture details, see [Performance & Concurrency](/docs/guides/performance-and-concurrency).

### `nuri.WithPoolSize`

```go
func WithPoolSize(n int) Option
```

Sets the maximum number of WASM instances in the pool. Instances are created lazily on demand. The pool uses LIFO reuse to favor warm instances with cached scanners.

**Default:** `runtime.NumCPU()` (clamped to minimum 1)

### `nuri.WithMaxLineLength`

```go
func WithMaxLineLength(n int) Option
```

Sets the byte-length threshold for per-line pre-filtering. Lines exceeding this length are emitted as a single unstyled token with a `"too_long"` diagnostic. The check runs before any WASM call.

**Default:** `0` (no limit)

### `nuri.WithTimeoutMs`

```go
func WithTimeoutMs(ms int) Option
```

Sets the per-line soft timeout in milliseconds. Lines whose tokenization exceeds this duration are stopped early; partial tokens are preserved and a `"timeout"` diagnostic is recorded.

No-op when built with `//go:build onig_cgo`.

**Default:** `0` (no timeout)

### `nuri.WithCompilationCacheDir`

```go
func WithCompilationCacheDir(dir string) Option
```

Enables an on-disk cache for the AOT-compiled `onig.wasm` module. The directory is created if needed and may be shared between processes. It only grows when the embedded WASM module changes.

**Default:** `""` (no cache; module is compiled on every process start)

### `nuri.WithRegexInterruption`

```go
func WithRegexInterruption(enabled bool) Option
```

Toggles WASM-level regex interruption. When enabled, wazero compiles interrupt checkpoints into the WASM JIT so a runaway regex can be stopped mid-search via context cancellation. The cost is roughly 3x lower throughput.

**Default:** `true`

## Rendering options

### `nuri.WithMinContrast`

```go
func WithMinContrast(ratio float64) Option
```

Sets the minimum WCAG 2.1 contrast ratio between syntax token foreground colors and the editor background. Tokens that fail the check are adjusted at theme load time with zero per-token cost during highlighting.

See [Contrast Correction](/docs/guides/contrast-correction) for the adjustment algorithm and WCAG levels.

**Default:** `5.5` (WCAG AA enhanced). Set to `0` to disable.

### `nuri.WithDefaults`

```go
func WithDefaults(defaults CodeToHTMLOptions) Option
```

Sets default `CodeToHTMLOptions` applied to every `h.CodeToHTML()` call. Per-call non-zero fields override these defaults.

Merge behavior: for each field, a non-zero or non-nil per-call value wins. Slice fields (decoration ranges like `HighlightLines`, `FocusLines`, `InsertedLines`, `DeletedLines`) are cloned after merging to prevent aliasing between calls.

```go
h, err := nuri.New(ctx,
	nuri.WithFS(core.FS()),
	nuri.WithDefaults(nuri.CodeToHTMLOptions{
		Theme: "github-dark",
		Transformers: []nuri.Transformer{
			transformers.Notation(),
		},
	}),
)

html, err := h.CodeToHTML(ctx, code, nuri.CodeToHTMLOptions{
	Lang: "go",
})
```

In this example, the per-call options specify only `Lang`. The default `Theme` and `Transformers` are applied automatically.

**Default:** `nil` (no defaults; every call must specify its own options)
