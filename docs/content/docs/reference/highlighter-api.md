---
title: Highlighter API
description: "Every exported method on the Highlighter type."
sidebar:
  order: 1
---

This page documents every method on `*nuri.Highlighter`. For constructor options, see [Configuration](/docs/reference/configuration). For options structs and result types, see [Types](/docs/reference/types).

## Constructor and lifecycle

### `nuri.New`

```go
func New(ctx context.Context, opts ...Option) (*Highlighter, error)
```

AOT-compiles the embedded `onig.wasm` module, instantiates the WASM instance pool, and builds the grammar/theme registry. Context is used for WASM compilation and pool setup.

Returns an error if engine creation, pool creation, or any custom grammar/theme registration fails. On partial failure, already-created resources are closed before returning.

```go
h, err := nuri.New(ctx,
	nuri.WithFS(core.FS()),
	nuri.WithPoolSize(4),
)
if err != nil {
	log.Fatal(err)
}
defer h.Close(ctx)
```

### `h.Close`

```go
func (h *Highlighter) Close(ctx context.Context) error
```

Releases all WASM resources owned by the highlighter. Safe to call multiple times (guarded by `sync.Once`). Returns the joined pool and engine close errors; both close operations are attempted even if the first fails.

## Highlighting methods

### `h.CodeToHTML`

```go
func (h *Highlighter) CodeToHTML(
	ctx context.Context,
	code string,
	opts CodeToHTMLOptions,
) (string, error)
```

Tokenizes source code, resolves colors from the theme, runs the transformer pipeline, and returns an HTML string.

If `WithDefaults` was set at construction time, per-call non-zero fields override the defaults.

**Pipeline order:**

1. `Preprocess` hook on each transformer (may rewrite `code` or mutate `opts`)
2. Tokenize (single-theme or multi-theme depending on `opts.Themes`)
3. `Tokens` hook on each transformer (may rewrite the token stream)
4. Build HTML AST via `renderer.BuildTree`
5. Serialize to HTML via `WriteTo`
6. `Postprocess` hook on each transformer (may rewrite the final HTML string)

**Returns:** HTML string with `<pre><code>` containing styled `<span>` elements.

**Errors:** Returns an error on WASM trap or context cancellation. Unknown languages produce a plaintext fallback with a `"unknown_lang"` diagnostic (non-fatal).

### `h.CodeToTokens`

```go
func (h *Highlighter) CodeToTokens(
	ctx context.Context,
	code string,
	opts CodeToTokensOptions,
) (*TokensResult, error)
```

Tokenizes source code and resolves per-token colors from the specified theme. Returns structured token data without rendering to any output format.

**Routing:**

- `len(opts.Themes) > 0`: multi-theme mode. Tokenizes once, resolves all themes from the same token stream. `opts.Theme` is ignored.
- `opts.Lang == "ansi"`: routes to the ANSI tokenizer (no TextMate grammar).
- Unknown `opts.Lang`: returns a plaintext fallback with a `"unknown_lang"` diagnostic. Not a hard error.

**Returns:** `*TokensResult` with per-line token slices, theme foreground/background, and any non-fatal diagnostics (`"timeout"`, `"too_long"`, `"panic"`, `"unknown_lang"`).

**Errors:** Returns an error on WASM trap or context cancellation.

### `h.CodeToHighlightedTokens`

```go
func (h *Highlighter) CodeToHighlightedTokens(
	ctx context.Context,
	code string,
	opts CodeToTokensOptions,
) (*TokensResult, error)
```

Equivalent to `h.CodeToTokens`. Delegates directly to it. The name reflects that the returned tokens already carry resolved color information, including contrast adjustment if `nuri.WithMinContrast` was set.

### `h.CodeToANSI`

```go
func (h *Highlighter) CodeToANSI(
	ctx context.Context,
	code string,
	opts CodeToANSIOptions,
) (string, error)
```

Tokenizes source code and renders as ANSI escape sequences for terminal display.

**Color depth:** Controlled by `opts.ColorDepth`. The default (`ColorDepthTruecolor`, value `0`) emits 24-bit RGB sequences. Other constants: `ColorDepth256`, `ColorDepth16`, `ColorDepth8`.

**Returns:** A string with ANSI escape codes. Each line is terminated by a reset sequence.

**Errors:** Same as `h.CodeToTokens`.

### `h.CodeToSVG`

```go
func (h *Highlighter) CodeToSVG(
	ctx context.Context,
	code string,
	opts CodeToSVGOptions,
) (string, error)
```

Tokenizes source code and renders as a self-contained SVG document. SVG-specific layout options (`FontFamily`, `FontSize`, `LineHeight`, `PadX`, `PadY`, `CornerRadius`, `ShowBackground`) are documented in [Types](/docs/reference/types).

**Returns:** A complete SVG document as a string.

**Errors:** Same as `h.CodeToTokens`.

### `h.CodeToJSON`

```go
func (h *Highlighter) CodeToJSON(
	ctx context.Context,
	code string,
	opts CodeToJSONOptions,
) ([]byte, error)
```

Tokenizes source code and returns the `*TokensResult` serialized to JSON bytes. Supports multi-theme mode via `opts.Themes`.

When `opts.Indent` is true, the output uses `json.MarshalIndent` with 2-space indentation. Otherwise, `json.Marshal` produces compact JSON.

**Returns:** JSON bytes representing the token result.

**Errors:** Same as `h.CodeToTokens`, plus any JSON marshaling error.

### `h.CodeToPlainText`

```go
func (h *Highlighter) CodeToPlainText(
	ctx context.Context,
	code string,
	opts CodeToPlainTextOptions,
) (string, error)
```

Tokenizes source code and returns the concatenated token content with no formatting or color information.

**Returns:** Plain text string.

**Errors:** Same as `h.CodeToTokens`.

## Runtime registration

### `h.LoadLanguage`

```go
func (h *Highlighter) LoadLanguage(name string, data []byte) error
```

Registers a grammar from raw TextMate grammar JSON bytes at runtime. Equivalent to `nuri.WithGrammar` but usable after construction.

**Returns:** An error if the JSON is malformed or the grammar fails validation.

### `h.LoadTheme`

```go
func (h *Highlighter) LoadTheme(name string, data []byte) error
```

Registers a theme from raw VS Code theme JSON bytes at runtime. Equivalent to `nuri.WithTheme` but usable after construction.

**Returns:** An error if the JSON is malformed or the theme fails validation.

### `h.RegisterAlias`

```go
func (h *Highlighter) RegisterAlias(alias, target string)
```

Maps a language alias to a canonical name at runtime. Equivalent to `nuri.WithAlias` but usable after construction.

```go
h.RegisterAlias("sh", "shellscript")
```

A non-existent target is silently recorded. It resolves to the plaintext fallback at highlight time.

### `h.RegisterExtension`

```go
func (h *Highlighter) RegisterExtension(ext, lang string)
```

Maps a file extension (without the leading dot) to a language name, overriding any existing mapping for that extension. Equivalent to `nuri.WithExtension` but usable after construction.

```go
h.RegisterExtension("mjs", "javascript")
```

## Introspection

### `h.LoadedLanguages`

```go
func (h *Highlighter) LoadedLanguages() []string
```

Returns the names of all currently cached grammars. This includes grammars loaded on demand from the bundle filesystem and those registered explicitly via `h.LoadLanguage` or `nuri.WithGrammar`.

### `h.LoadedThemes`

```go
func (h *Highlighter) LoadedThemes() []string
```

Returns the names of all currently cached themes.

### `h.GetThemeColors`

```go
func (h *Highlighter) GetThemeColors(themeName string) (ThemeColors, error)
```

Returns the UI colors for a loaded theme. If `nuri.WithMinContrast` was set, the returned colors reflect the contrast-adjusted theme (the same adjustment applied during highlighting).

Intended for building code block chrome (title bars, copy buttons, terminal frames) styled consistently with the highlighted output.

**Returns:** A `ThemeColors` struct with `Type` (`"light"` or `"dark"`), `Background`, `Foreground`, `SelectionBackground`, `LineHighlightBg`, and the full `Colors` map.

**Errors:** Returns `nuri.ErrThemeNotFound` if the theme is not loaded.

```go
colors, err := h.GetThemeColors("github-dark")
if err != nil {
	log.Fatal(err)
}
fmt.Println(colors.Background) // "#24292e"
```

## Language detection

### `h.DetectLanguage`

```go
func (h *Highlighter) DetectLanguage(filename string) (string, bool)
```

Resolves a language name from a filename or path. Checks exact filenames first (e.g. `"Makefile"`, `"Dockerfile"`), then file extensions. The returned name is usable with all `CodeTo*` methods.

**Returns:** The language name and `true` if matched; `""` and `false` if no match.

```go
lang, ok := h.DetectLanguage("main.go")
if ok {
	html, err := h.CodeToHTML(ctx, code, nuri.CodeToHTMLOptions{
		Lang:  lang,
		Theme: "github-dark",
	})
}
```

### `h.DetectLanguageByContent`

```go
func (h *Highlighter) DetectLanguageByContent(firstLine string) (string, bool)
```

Resolves a language from the first line of file content. Intended for shebang detection (e.g. `"#!/usr/bin/env python3"` resolves to `"python"`).

**Returns:** The language name and `true` if matched; `""` and `false` if no match.
