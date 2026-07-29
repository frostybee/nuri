---
title: Types
description: "All exported types, structs, interfaces, and constants."
sidebar:
  order: 3
---

This page documents all exported types in the `nuri` package. These types are defined in the `ast` package and re-exported so consumers write `nuri.Element`, `nuri.CodeToHTMLOptions`, etc. The `theme.FontStyle` type used in token styles is documented in [Theme Package](/docs/reference/theme-package).

## Options structs

### `CodeToHTMLOptions`

```go
type CodeToHTMLOptions struct {
    Lang           string
    Theme          string
    Themes         map[string]string
    DefaultColor   *bool
    Transformers   []Transformer
    HighlightLines []LineRange
    FocusLines     []LineRange
    InsertedLines  []LineRange
    DeletedLines   []LineRange
    PreClass       string
    CodeClass      string
    PreAttrs       map[string]string
    CodeAttrs      map[string]string
    ClassMap       *StyleClassMap
    MaxLineLength  *int
    TimeoutMs      *int
}
```

Configures an `h.CodeToHTML()` call.

| Field | Type | Description |
|---|---|---|
| `Lang` | `string` | Language name (e.g. `"go"`, `"javascript"`) |
| `Theme` | `string` | Theme name for single-theme mode. Ignored when `Themes` is non-nil. |
| `Themes` | `map[string]string` | Multi-theme mode: key to theme name. The lexicographically first key is the default theme (inline styles); others produce `--nuri-{key}-{prop}` CSS variables. |
| `DefaultColor` | `*bool` | Controls inline `color:` emission in multi-theme mode. `nil` or `*true` = emit inline color for the default theme. `*false` = CSS variables only. |
| `Transformers` | `[]Transformer` | Ordered list of transformers applied during the rendering pipeline. |
| `HighlightLines` | `[]LineRange` | Lines to highlight (adds `highlighted` class). |
| `FocusLines` | `[]LineRange` | Lines to focus (adds `focused` class to target lines, `dimmed` to others, `has-focused` to the root). |
| `InsertedLines` | `[]LineRange` | Lines marked as inserted (adds `diff add` class). |
| `DeletedLines` | `[]LineRange` | Lines marked as deleted (adds `diff remove` class). |
| `PreClass` | `string` | Extra CSS class added to the `<pre>` element. |
| `CodeClass` | `string` | Extra CSS class added to the `<code>` element. |
| `PreAttrs` | `map[string]string` | Extra HTML attributes added to the `<pre>` element. |
| `CodeAttrs` | `map[string]string` | Extra HTML attributes added to the `<code>` element. |
| `ClassMap` | `*StyleClassMap` | When non-nil, replaces inline `style=` attributes with hashed class names. Call `CSS()` for the stylesheet. |
| `MaxLineLength` | `*int` | `nil` = use highlighter default. Overrides the per-line byte-length pre-filter. |
| `TimeoutMs` | `*int` | `nil` = use highlighter default. Overrides the per-line soft timeout. |

### `CodeToTokensOptions`

```go
type CodeToTokensOptions struct {
    Lang          string
    Theme         string
    Themes        map[string]string
    MaxLineLength *int
    TimeoutMs     *int
}
```

Configures an `h.CodeToTokens()` call.

| Field | Type | Description |
|---|---|---|
| `Lang` | `string` | Language name |
| `Theme` | `string` | Theme name. Ignored when `Themes` is non-empty. |
| `Themes` | `map[string]string` | Multi-theme mode. Same semantics as `CodeToHTMLOptions.Themes`. |
| `MaxLineLength` | `*int` | `nil` = use highlighter default |
| `TimeoutMs` | `*int` | `nil` = use highlighter default |

### `CodeToANSIOptions`

```go
type CodeToANSIOptions struct {
    Lang          string
    Theme         string
    ColorDepth    ColorDepth
    MaxLineLength *int
    TimeoutMs     *int
}
```

Configures an `h.CodeToANSI()` call.

| Field | Type | Description |
|---|---|---|
| `Lang` | `string` | Language name |
| `Theme` | `string` | Theme name |
| `ColorDepth` | `ColorDepth` | ANSI color mode. Zero value (`ColorDepthTruecolor`) uses 24-bit RGB. |
| `MaxLineLength` | `*int` | `nil` = use highlighter default |
| `TimeoutMs` | `*int` | `nil` = use highlighter default |

### `CodeToSVGOptions`

```go
type CodeToSVGOptions struct {
    Lang           string
    Theme          string
    MaxLineLength  *int
    TimeoutMs      *int
    FontFamily     string
    FontSize       float64
    LineHeight     float64
    PadX           float64
    PadY           float64
    TabWidth       int
    CornerRadius   float64
    ShowBackground *bool
}
```

Configures an `h.CodeToSVG()` call. SVG-specific fields use renderer defaults when zero-valued.

| Field | Type | Default | Description |
|---|---|---|---|
| `Lang` | `string` | | Language name |
| `Theme` | `string` | | Theme name |
| `MaxLineLength` | `*int` | `nil` | `nil` = use highlighter default |
| `TimeoutMs` | `*int` | `nil` | `nil` = use highlighter default |
| `FontFamily` | `string` | `"Consolas, Monaco, ..."` | CSS font-family for the SVG text |
| `FontSize` | `float64` | `14` | Font size in px |
| `LineHeight` | `float64` | `1.2` | Line height as an em multiplier |
| `PadX` | `float64` | `16` | Horizontal padding in px |
| `PadY` | `float64` | `16` | Vertical padding in px |
| `TabWidth` | `int` | `4` | Number of spaces per tab |
| `CornerRadius` | `float64` | `8` | Background rectangle corner radius in px |
| `ShowBackground` | `*bool` | `nil` | `nil` or `*true` = show background rect; `*false` = no background |

### `CodeToJSONOptions`

```go
type CodeToJSONOptions struct {
    Lang          string
    Theme         string
    Themes        map[string]string
    MaxLineLength *int
    TimeoutMs     *int
    Indent        bool
}
```

Configures an `h.CodeToJSON()` call.

| Field | Type | Description |
|---|---|---|
| `Lang` | `string` | Language name |
| `Theme` | `string` | Theme name. Ignored when `Themes` is non-empty. |
| `Themes` | `map[string]string` | Multi-theme mode. Same semantics as `CodeToHTMLOptions.Themes`. |
| `MaxLineLength` | `*int` | `nil` = use highlighter default |
| `TimeoutMs` | `*int` | `nil` = use highlighter default |
| `Indent` | `bool` | When `true`, uses 2-space indented JSON. |

### `CodeToPlainTextOptions`

```go
type CodeToPlainTextOptions struct {
    Lang          string
    Theme         string
    MaxLineLength *int
    TimeoutMs     *int
}
```

Configures an `h.CodeToPlainText()` call. `Theme` is accepted for API uniformity but has no effect on the output.

## Result types

### `TokensResult`

```go
type TokensResult struct {
    Tokens      [][]ThemedToken    `json:"tokens"`
    FG          string             `json:"fg"`
    BG          string             `json:"bg"`
    ThemeName   string             `json:"themeName"`
    Diagnostics []Diagnostic       `json:"diagnostics,omitempty"`
    ThemeFG     map[string]string  `json:"themeFG,omitempty"`
    ThemeBG     map[string]string  `json:"themeBG,omitempty"`
    ThemeNames  []string           `json:"themeNames,omitempty"`
}
```

Output of `h.CodeToTokens()` and `h.CodeToHighlightedTokens()`.

| Field | Type | Description |
|---|---|---|
| `Tokens` | `[][]ThemedToken` | Per-line token slices. Outer slice is lines, inner is tokens. |
| `FG` | `string` | Default theme foreground hex color |
| `BG` | `string` | Default theme background hex color |
| `ThemeName` | `string` | Resolved theme name |
| `Diagnostics` | `[]Diagnostic` | Non-fatal per-line degradations |
| `ThemeFG` | `map[string]string` | Per-theme default foreground. `nil` in single-theme mode. |
| `ThemeBG` | `map[string]string` | Per-theme default background. `nil` in single-theme mode. |
| `ThemeNames` | `[]string` | Sorted theme keys. `nil` in single-theme mode. |

### `ThemedToken`

```go
type ThemedToken struct {
    Content     string                `json:"content"`
    Color       string                `json:"color,omitempty"`
    BgColor     string                `json:"bgColor,omitempty"`
    FontStyle   theme.FontStyle       `json:"fontStyle"`
    Scopes      []string              `json:"scopes,omitempty"`
    ThemeStyles map[string]TokenStyle `json:"themeStyles,omitempty"`
}
```

A single token with resolved style information.

| Field | Type | Description |
|---|---|---|
| `Content` | `string` | Source text of this token |
| `Color` | `string` | Foreground hex color from the default theme |
| `BgColor` | `string` | Background hex color from the default theme |
| `FontStyle` | `theme.FontStyle` | Font style from the default theme |
| `Scopes` | `[]string` | TextMate scope stack at this token |
| `ThemeStyles` | `map[string]TokenStyle` | Per-theme styles in multi-theme mode. Keys are theme keys from `Themes`. `nil` in single-theme mode. |

### `TokenStyle`

```go
type TokenStyle struct {
    Color     string          `json:"color,omitempty"`
    BgColor   string          `json:"bgColor,omitempty"`
    FontStyle theme.FontStyle `json:"fontStyle"`
}
```

Resolved style for a single theme. Used as values in `ThemedToken.ThemeStyles`.

## Diagnostics

### `Diagnostic`

```go
type Diagnostic struct {
    Line int    `json:"line"`
    Kind string `json:"kind"`
}
```

Records a non-fatal degradation during tokenization. `Line` is 0-based.

| Kind | Trigger |
|---|---|
| `"timeout"` | Line tokenization exceeded `TimeoutMs` |
| `"too_long"` | Line byte length exceeded `MaxLineLength` |
| `"panic"` | Tokenizer panicked on a line (instance replaced) |
| `"unknown_lang"` | Language name not found; plaintext fallback used |

## HTML AST

### `Node`

```go
type Node interface {
    WriteTo(w io.Writer) (int64, error)
}
```

An HTML node that can serialize itself. Satisfies `io.WriterTo`. Both `*Element` and `*Text` implement this interface.

### `Element`

```go
type Element struct {
    Tag      string
    Classes  []string
    Styles   map[string]string
    Attrs    map[string]string
    Children []Node
}
```

An HTML element node. Styles and attributes are emitted in sorted key order for deterministic output.

| Field | Type | Description |
|---|---|---|
| `Tag` | `string` | HTML tag name (e.g. `"pre"`, `"code"`, `"span"`) |
| `Classes` | `[]string` | CSS classes emitted as `class="..."` |
| `Styles` | `map[string]string` | Inline style declarations emitted as `style="..."` |
| `Attrs` | `map[string]string` | Extra HTML attributes |
| `Children` | `[]Node` | Ordered child nodes |

**Methods:**

- `AddClass(class string)` -- appends a CSS class if not already present.
- `SetAttr(key, val string)` -- sets an HTML attribute. Initializes the `Attrs` map if nil.
- `WriteTo(w io.Writer) (int64, error)` -- serializes the element and its children to `w`. Style values and attribute values are HTML-escaped.

### `Text`

```go
type Text struct {
    Content string
}
```

A text node. Content is HTML-escaped on write (`<`, `>`, `&` become `&lt;`, `&gt;`, `&amp;`).

**Methods:**

- `WriteTo(w io.Writer) (int64, error)` -- writes the HTML-escaped content to `w`.

## Line decorations

### `LineRange`

```go
type LineRange struct {
    Start int
    End   int
}
```

A 1-based inclusive range of lines. Used in `CodeToHTMLOptions` decoration fields (`HighlightLines`, `FocusLines`, `InsertedLines`, `DeletedLines`).

**Methods:**

- `Contains(line int) bool` -- reports whether the range includes `line` (1-based).

### `Range`

```go
func Range(start, end int) LineRange
```

Returns a `LineRange` spanning from `start` to `end` (1-based, inclusive).

### `Lines`

```go
func Lines(nums ...int) []LineRange
```

Returns a slice of single-line `LineRange` values. Each `Start` and `End` equals the given line number.

### `InRanges`

```go
func InRanges(ranges []LineRange, line int) bool
```

Reports whether `line` is contained in any of the given ranges.

## Style-to-class

### `StyleClassMap`

```go
type StyleClassMap struct { /* unexported fields */ }
```

Collects unique style combinations and assigns deterministic hashed class names. Pass a shared instance across multiple `h.CodeToHTML()` calls via `CodeToHTMLOptions.ClassMap` to deduplicate styles, then call `CSS()` for the stylesheet.

**Constructor:**

```go
func NewStyleClassMap() *StyleClassMap
```

**Methods:**

- `Get(styles map[string]string) string` -- returns the class name for the given style map, creating a new hashed class if this combination is new.
- `CSS() string` -- returns the complete stylesheet. Rules are sorted by class name for deterministic output.

### `CanonicalStyles`

```go
func CanonicalStyles(styles map[string]string) string
```

Produces a deterministic string key from a style map. Keys are sorted; format is `key1:val1;key2:val2`. Used as the deduplication key in `StyleClassMap.Get`.

### `StyleHash`

```go
func StyleHash(canon string) string
```

Computes a deterministic short hash from a canonical style string using FNV-64a. Returns a string of the form `_s_{hexdigest}`.

### `StylestoCSS`

```go
func StylestoCSS(styles map[string]string) string
```

Converts a style map to a CSS rule body string. Keys are sorted; format is `key1: val1; key2: val2`.

## Theme colors

### `ThemeColors`

```go
type ThemeColors struct {
    Type                string
    Background          string
    Foreground          string
    SelectionBackground string
    LineHighlightBg     string
    Colors              map[string]string
}
```

Exposes a theme's UI colors. Returned by `h.GetThemeColors()`.

| Field | Type | Description |
|---|---|---|
| `Type` | `string` | `"light"` or `"dark"` |
| `Background` | `string` | `editor.background` color |
| `Foreground` | `string` | `editor.foreground` color |
| `SelectionBackground` | `string` | `editor.selectionBackground` color |
| `LineHighlightBg` | `string` | `editor.lineHighlightBackground` color |
| `Colors` | `map[string]string` | Full color map (`editor.*`, `terminal.*`, etc.) |

## Color depth

### `ColorDepth`

```go
type ColorDepth int
```

Specifies the ANSI color mode for terminal output.

| Constant | Value | Description |
|---|---|---|
| `ColorDepthTruecolor` | `0` | 24-bit RGB (default) |
| `ColorDepth256` | `256` | 256-color indexed |
| `ColorDepth16` | `16` | 16-color (standard + bright) |
| `ColorDepth8` | `8` | 8-color (standard only) |

## Font style

The `FontStyle` type is defined in the `theme` package as `theme.FontStyle`. It appears in `ThemedToken.FontStyle` and `TokenStyle.FontStyle`. See [Theme Package](/docs/reference/theme-package) for the type definition, constants (`FontStyleNone`, `FontStyleItalic`, `FontStyleBold`, `FontStyleUnderline`, `FontStyleStrikethrough`), and the `Has()` method.
