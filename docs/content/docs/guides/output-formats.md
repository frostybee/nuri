---
title: Output Formats
description: "HTML, ANSI, SVG, JSON, plain text, and raw tokens."
sidebar:
  order: 3
---

Nuri provides six output methods. All follow the same pattern: pass code and an options struct, get formatted output.

## HTML

```go
html, err := h.CodeToHTML(ctx, `fmt.Println("hello")`, nuri.CodeToHTMLOptions{
	Lang:  "go",
	Theme: "github-dark",
})
```

Returns `(string, error)`. Produces a `<pre><code>` element with inline color styles.

`CodeToHTMLOptions` is the most feature-rich options struct. Beyond `Lang` and `Theme`, it supports:

- `Themes` for [multi-theme mode](/docs/guides/multi-theme-mode)
- `Transformers` for [notation comments, meta ranges, and custom hooks](/docs/guides/transformers)
- `HighlightLines`, `FocusLines`, `InsertedLines`, `DeletedLines` for [line decorations](/docs/guides/line-decorations)
- `ClassMap` for [style-to-class mode](/docs/guides/style-to-class-mode)
- `PreClass`, `CodeClass`, `PreAttrs`, `CodeAttrs` for custom attributes on the wrapper elements

Use `nuri.WithDefaults()` to set site-wide defaults applied to every `CodeToHTML` call. Per-call non-zero fields override the defaults:

```go
h, _ := nuri.New(ctx,
	nuri.WithFS(core.FS()),
	nuri.WithDefaults(nuri.CodeToHTMLOptions{
		Theme:    "github-dark",
		PreClass: "code-block",
	}),
)

// Only Lang is needed per call; Theme and PreClass carry over.
html, _ := h.CodeToHTML(ctx, code, nuri.CodeToHTMLOptions{Lang: "go"})
```

## ANSI

```go
ansi, err := h.CodeToANSI(ctx, `fmt.Println("hello")`, nuri.CodeToANSIOptions{
	Lang:  "go",
	Theme: "github-dark",
})
fmt.Print(ansi)
```

Returns `(string, error)`. Produces ANSI escape sequences for terminal display. Use `fmt.Print`, not `fmt.Println`, since the output already contains newlines.

### Color depth

`ColorDepth` controls the ANSI color encoding:

| Constant | Value | Description |
|---|---|---|
| `nuri.ColorDepthTruecolor` | `0` | 24-bit RGB (default) |
| `nuri.ColorDepth256` | `256` | 256-color indexed palette |
| `nuri.ColorDepth16` | `16` | 16-color (standard + bright) |
| `nuri.ColorDepth8` | `8` | 8-color (standard only) |

```go
ansi, err := h.CodeToANSI(ctx, code, nuri.CodeToANSIOptions{
	Lang:       "go",
	Theme:      "github-dark",
	ColorDepth: nuri.ColorDepth256,
})
```

Indexed modes use perceptual nearest-color matching to map 24-bit theme colors to the available palette.

### ANSI as an input language

Set `Lang: "ansi"` to tokenize text containing ANSI escape sequences. This parses existing terminal output and re-highlights it through a theme:

```go
html, err := h.CodeToHTML(ctx, ansiInput, nuri.CodeToHTMLOptions{
	Lang:  "ansi",
	Theme: "github-dark",
})
```

## SVG

```go
svg, err := h.CodeToSVG(ctx, `fmt.Println("hello")`, nuri.CodeToSVGOptions{
	Lang:         "go",
	Theme:        "github-dark",
	FontSize:     16,
	CornerRadius: 12,
})
```

Returns `(string, error)`. Produces a self-contained SVG document. Useful for generating images of code snippets without a browser.

### SVG options

| Field | Type | Default | Description |
|---|---|---|---|
| `FontFamily` | `string` | `"Consolas, Monaco, ..."` | Font stack |
| `FontSize` | `float64` | `14` | Font size in px |
| `LineHeight` | `float64` | `1.2` | Line height multiplier |
| `PadX` | `float64` | `16` | Horizontal padding in px |
| `PadY` | `float64` | `16` | Vertical padding in px |
| `TabWidth` | `int` | `4` | Spaces per tab character |
| `CornerRadius` | `float64` | `8` | Border radius in px |
| `ShowBackground` | `*bool` | `nil` (show) | Set to `&false` to omit the background rect |

Zero values use the defaults listed above.

## JSON

```go
data, err := h.CodeToJSON(ctx, `fmt.Println("hello")`, nuri.CodeToJSONOptions{
	Lang:   "go",
	Theme:  "github-dark",
	Indent: true,
})
```

Returns `([]byte, error)`. Marshals the `TokensResult` struct directly to JSON. Set `Indent: true` for pretty-printed output with 2-space indentation.

The JSON structure mirrors `TokensResult`:

```json
{
  "tokens": [[{"content": "fmt", "color": "#E1E4E8", ...}, ...]],
  "fg": "#e1e4e8",
  "bg": "#24292e",
  "themeName": "github-dark"
}
```

`CodeToJSONOptions` supports `Themes` for multi-theme output. When set, the JSON includes `themeFG`, `themeBG`, `themeNames`, and per-token `themeStyles` fields.

## Plain text

```go
text, err := h.CodeToPlainText(ctx, code, nuri.CodeToPlainTextOptions{
	Lang:  "go",
	Theme: "github-dark",
})
```

Returns `(string, error)`. Reconstructs the source text from the token stream with no styling. Newlines separate lines; there is no trailing newline.

Use this to extract text content from a highlighting pipeline or to verify tokenizer coverage (the output should match the input).

## Raw tokens

```go
result, err := h.CodeToTokens(ctx, `fmt.Println("hello")`, nuri.CodeToTokensOptions{
	Lang:  "go",
	Theme: "github-dark",
})

for _, line := range result.Tokens {
	for _, tok := range line {
		fmt.Printf("%s → %s\n", tok.Content, tok.Color)
	}
}
```

Returns `(*TokensResult, error)`. Gives programmatic access to the token grid with resolved colors and scope names.

Each `ThemedToken` has:

| Field | Type | Description |
|---|---|---|
| `Content` | `string` | Token text |
| `Color` | `string` | Foreground color |
| `BgColor` | `string` | Background color (if set by theme) |
| `FontStyle` | `theme.FontStyle` | Bold, italic, underline, strikethrough bitmask |
| `Scopes` | `[]string` | TextMate scope names |

`CodeToTokensOptions` supports `Themes` for multi-theme resolution. Non-default theme styles appear in `ThemedToken.ThemeStyles`.

`h.CodeToHighlightedTokens()` is an alias for `h.CodeToTokens()`, retained for Shiki API compatibility.

## Per-call overrides

All options structs accept `MaxLineLength *int` and `TimeoutMs *int` as per-call overrides for the constructor defaults. Pass `nil` to use the highlighter's default, or a pointer to an `int` to override for a single call.

## Next steps

- [Transformers](/docs/guides/transformers) for notation comments, meta ranges, and custom hooks (HTML only)
- [Multi-Theme Mode](/docs/guides/multi-theme-mode) for CSS variable emission and light/dark switching
