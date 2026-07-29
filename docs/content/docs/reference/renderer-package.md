---
title: Renderer Package
description: "Lower-level rendering functions for custom pipelines."
sidebar:
  order: 6
---

This page documents the `renderer` package (`github.com/frostybee/nuri/renderer`). Most consumers use the highlighter methods (`h.CodeToHTML()`, `h.CodeToANSI()`, etc.), which wrap these functions. Import the `renderer` package directly when building a custom rendering pipeline or a downstream library like Kazari.

## `renderer.BuildTree`

```go
func BuildTree(result *ast.TokensResult, opts *ast.CodeToHTMLOptions) *ast.Element
```

Converts a `*TokensResult` into an HTML element tree. This is the main HTML construction function called by `h.CodeToHTML()` between the `Tokens` and `Postprocess` transformer hooks.

**Single-theme output:** The `<pre>` element gets classes `["nuri", themeName]` with inline `background-color` and `color` styles.

**Multi-theme output:** The `<pre>` element gets classes `["nuri", "nuri-themes", key1, key2, ...]` with CSS variables `--nuri-{key}-bg` and `--nuri-{key}` for each non-default theme.

The `<pre>` element always has `tabindex="0"`. Lines are `<span class="line">` elements separated by `"\n"` text nodes. Per-token byte offsets (UTF-8) are accumulated and passed to `Transformer.Span` as the `col` parameter.

The function applies all decoration classes (`highlighted`, `focused`/`dimmed`, `diff add`/`diff remove`), runs transformer hooks (`Span`, `Line`, `Code`, `Pre`, `Root`), and replaces inline styles with hashed class names when `opts.ClassMap` is non-nil.

**Returns:** The root `*ast.Element`, serializable via `WriteTo(w io.Writer)`.

## `renderer.TokenStyles`

```go
func TokenStyles(tok ast.ThemedToken) map[string]string
```

Converts a token's resolved style into an inline CSS property map for single-theme mode. Used internally by `BuildTree` for each `<span>` element.

**Possible keys:** `"color"`, `"background-color"`, `"font-style"`, `"font-weight"`, `"text-decoration"`. Returns an empty map if the token has no styling.

## `renderer.TokenStylesMulti`

```go
func TokenStylesMulti(tok ast.ThemedToken, defaultColor bool) map[string]string
```

Converts a token's resolved style into a combined CSS property map for multi-theme mode. When `defaultColor` is true, emits standard inline properties (`color`, `background-color`, font properties) for the default theme. For each non-default theme key in `tok.ThemeStyles`, emits CSS custom properties:

- `--nuri-{key}` for foreground color
- `--nuri-{key}-bg` for background color
- `--nuri-{key}-font-style`, `--nuri-{key}-font-weight`, `--nuri-{key}-text-decoration` for font style

## `renderer.FontStyleCSS`

```go
func FontStyleCSS(fs theme.FontStyle) map[string]string
```

Translates a `theme.FontStyle` bitmask into CSS properties.

| Flag | CSS property | CSS value |
|---|---|---|
| `FontStyleItalic` | `font-style` | `italic` |
| `FontStyleBold` | `font-weight` | `bold` |
| `FontStyleUnderline` | `text-decoration` | `underline` |
| `FontStyleStrikethrough` | `text-decoration` | `line-through` |

When both underline and strikethrough are set, `text-decoration` is `"underline line-through"`.

**Returns:** `nil` when `fs <= 0` (covers both `FontStyleNone` and `FontStyleNotSet`).

## `renderer.RenderANSI`

```go
func RenderANSI(w io.Writer, result *ast.TokensResult, opts *ast.CodeToANSIOptions) error
```

Writes ANSI SGR escape sequences to `w`. Applies foreground color and font style (bold, italic, underline, strikethrough) per token. Resets after each styled token with `\033[0m`.

`opts.ColorDepth` controls how hex colors map to ANSI codes:

| Depth | SGR format |
|---|---|
| `ColorDepthTruecolor` (0) | `38;2;r;g;b` (24-bit RGB) |
| `ColorDepth256` (256) | `38;5;n` (nearest 256-color match) |
| `ColorDepth16` (16) | 30-37, 90-97 (nearest standard/bright) |
| `ColorDepth8` (8) | 30-37 (nearest standard color) |

Tokens without an explicit foreground color fall back to `result.FG`. Background colors from tokens are not emitted. Lines are separated by `"\n"` with no trailing newline.

**Returns:** The first `io.Writer` error encountered, or `nil`.

## `renderer.RenderSVG`

```go
func RenderSVG(w io.Writer, result *ast.TokensResult, opts *ast.CodeToSVGOptions) error
```

Writes a self-contained SVG document to `w`. Each line becomes a `<text>` element; styled tokens become `<tspan>` children with `fill`, `font-weight`, `font-style`, and `text-decoration` attributes.

Spaces are replaced with `&#160;` to preserve whitespace. Tabs are expanded to `opts.TabWidth` non-breaking spaces. Tokens whose color equals the default foreground omit the `fill` attribute (inherited from the parent `<g>`).

SVG-specific options (`FontFamily`, `FontSize`, `LineHeight`, `PadX`, `PadY`, `CornerRadius`, `ShowBackground`) apply renderer defaults when zero-valued. See [Types](/docs/reference/types) for the `CodeToSVGOptions` field table.

**Returns:** The first write error, or `nil`.

## `renderer.RenderPlainText`

```go
func RenderPlainText(w io.Writer, result *ast.TokensResult) error
```

Writes raw token content to `w` with no formatting. Lines are joined with `"\n"`. All color, font style, and scope information is ignored.

**Returns:** The first write error, or `nil`.
