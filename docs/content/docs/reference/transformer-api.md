---
title: Transformer API
description: "The Transformer interface, BaseTransformer, and built-in transformer constructors."
sidebar:
  order: 4
---

This page documents the `Transformer` interface, the `BaseTransformer` embed pattern, and the built-in transformer constructors in the `transformers` package. For usage examples, see [Transformers](/docs/guides/transformers) and [Custom Transformers](/docs/guides/custom-transformers).

## `Transformer` interface

```go
type Transformer interface {
    Name() string
    Preprocess(code string, opts *CodeToHTMLOptions) string
    Tokens(tokens [][]ThemedToken) [][]ThemedToken
    Span(el *Element, line, col int, lineEl *Element, tok ThemedToken) *Element
    Line(el *Element, line int) *Element
    Code(el *Element) *Element
    Pre(el *Element) *Element
    Root(el *Element) *Element
    Postprocess(html string, opts *CodeToHTMLOptions) string
}
```

Hooks are called in the order listed during a single `h.CodeToHTML()` call.

| Hook | Called | Return `nil` / `""` means |
|---|---|---|
| `Name` | Once, for identification | (returns an identifier string) |
| `Preprocess` | Before tokenization | Leave source unchanged |
| `Tokens` | After tokenization, before rendering | Keep original token matrix |
| `Span` | Per token `<span>` element | Keep original element |
| `Line` | Per line `<span>` element | Keep original element |
| `Code` | On the `<code>` element | Keep original element |
| `Pre` | On the `<pre>` element | Keep original element |
| `Root` | On the root wrapper element | Keep original element |
| `Postprocess` | After HTML serialization | Leave HTML unchanged |

**`Span` parameters:** `line` is 1-based. `col` is the UTF-8 byte offset within the line (0-based). `lineEl` is the containing line element. `tok` is the themed token being rendered.

**`Preprocess` and `Postprocess`** receive a mutable `*CodeToHTMLOptions` pointer, allowing transformers to modify options (e.g. appending to `HighlightLines`).

## `BaseTransformer`

```go
type BaseTransformer struct{}
```

Provides no-op defaults for all `Transformer` hooks. Embed it in custom transformers to override only the hooks needed:

```go
type myTransformer struct {
    nuri.BaseTransformer
}

func (t *myTransformer) Name() string { return "my-transformer" }

func (t *myTransformer) Line(el *nuri.Element, line int) *nuri.Element {
    el.SetAttr("data-line", strconv.Itoa(line))
    return el
}
```

All `BaseTransformer` methods return `""` (string hooks) or `nil` (element hooks), signaling no change.

## Built-in transformers

### `transformers.Notation`

```go
func Notation() ast.Transformer
```

Parses magic comments in source code and applies CSS classes to the corresponding line elements. The magic comment and any surrounding bare comment markers are stripped from the output.

**Annotations:**

| Magic comment | Effect |
|---|---|
| `[!code ++]` | Adds `diff add` class to the line |
| `[!code --]` | Adds `diff remove` class to the line |
| `[!code highlight]` | Adds `highlighted` class to the line |
| `[!code focus]` | Adds `focused` to the line, `dimmed` to all other lines, `has-focused` to the `<pre>` element |
| `[!code error]` | Adds `highlighted error` classes to the line |
| `[!code warning]` | Adds `highlighted warning` classes to the line |
| `[!code word:TEXT]` | Wraps every occurrence of TEXT in `<span class="highlighted-word">` |

**Comment stripping:** After removing the `[!code ...]` pattern, if the remaining token content is only a bare comment marker (`//`, `#`, `/*`, `*/`, `/* */`), the token is dropped along with any preceding whitespace-only token. Lines that become entirely whitespace after stripping are eliminated from the output, and subsequent line numbers are renumbered.

**Hooks used:** `Name` (`"notation"`), `Tokens`, `Span` (word highlights only), `Line`, `Pre` (focus mode).

### `transformers.Meta`

```go
func Meta(meta string) ast.Transformer
```

Applies line highlighting based on a code-fence meta string. Parses `meta` using `ParseMetaRanges` and appends the resulting ranges to `opts.HighlightLines` in the `Preprocess` hook. The renderer then applies the `highlighted` class to matching lines.

Parse errors are silently ignored. The transformer becomes a no-op if no valid ranges are found.

**Hooks used:** `Name` (`"meta"`), `Preprocess`.

```go
html, err := h.CodeToHTML(ctx, code, nuri.CodeToHTMLOptions{
    Lang:  "go",
    Theme: "github-dark",
    Transformers: []nuri.Transformer{
        transformers.Meta("{1,3-5}"),
    },
})
```

### `transformers.Whitespace`

```go
func Whitespace() ast.Transformer
```

Renders whitespace characters as visible symbols. Each tab is replaced with `<span class="ws-tab">→</span>` and each space with `<span class="ws-space">·</span>`.

**Hooks used:** `Name` (`"whitespace"`), `Span`.

### `transformers.WhitespaceWith`

```go
func WhitespaceWith(opts WhitespaceOptions) ast.Transformer
```

Same as `transformers.Whitespace()` with custom replacement symbols. Zero-valued fields fall back to the defaults.

### `WhitespaceOptions`

```go
type WhitespaceOptions struct {
    Tab   string
    Space string
}
```

| Field | Type | Default | Description |
|---|---|---|---|
| `Tab` | `string` | `"→"` | Visible replacement for tab characters |
| `Space` | `string` | `"·"` | Visible replacement for space characters |

### `transformers.ParseMetaRanges`

```go
func ParseMetaRanges(meta string) ([]ast.LineRange, error)
```

Parses a code-fence meta string like `"{1,3-5,7}"` into a slice of `LineRange` values. Extracts the content between the first `{` and last `}`, splits on `,`, and parses each token as either a single line number or a `start-end` range.

**Returns:**

- `[]LineRange, nil` on success
- `nil, nil` for empty input, missing braces, or empty content between braces
- `nil, error` on non-numeric content

## Hook coverage

| Transformer | `Preprocess` | `Tokens` | `Span` | `Line` | `Pre` |
|---|---|---|---|---|---|
| `Notation()` | | yes | yes | yes | yes |
| `Meta(s)` | yes | | | | |
| `Whitespace()` | | | yes | | |

All built-in transformers use `BaseTransformer` for hooks not listed above. `Code`, `Root`, and `Postprocess` are not used by any built-in transformer.
