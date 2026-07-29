---
title: Transformers
description: "Built-in transformers for notation comments, meta ranges, and visible whitespace."
sidebar:
  order: 4
---

Transformers modify the HTML rendering pipeline. Pass them in `CodeToHTMLOptions.Transformers`:

```go
import "github.com/frostybee/nuri/transformers"

html, err := h.CodeToHTML(ctx, code, nuri.CodeToHTMLOptions{
	Lang:  "go",
	Theme: "github-dark",
	Transformers: []nuri.Transformer{
		transformers.Notation(),
		transformers.Meta("{1,3-5}"),
		transformers.Whitespace(),
	},
})
```

Transformers run in slice order and apply to HTML output only.

## Notation

`transformers.Notation()` parses magic comments in the source code and applies CSS classes to the corresponding lines. The magic comments are stripped from the output.

```go
html, err := h.CodeToHTML(ctx, `
func greet(name string) {
	fmt.Println("hello, " + name) // [!code highlight]
}
`, nuri.CodeToHTMLOptions{
	Lang:  "go",
	Theme: "github-dark",
	Transformers: []nuri.Transformer{transformers.Notation()},
})
```

### Supported annotations

| Comment | Classes added to line |
|---|---|
| `[!code ++]` | `diff`, `add` |
| `[!code --]` | `diff`, `remove` |
| `[!code highlight]` | `highlighted` |
| `[!code focus]` | `focused` |
| `[!code error]` | `highlighted`, `error` |
| `[!code warning]` | `highlighted`, `warning` |
| `[!code word:WORD]` | Wraps each occurrence of WORD in `<span class="highlighted-word">` |

The `[!code word:WORD]` annotation does not add a class to the line. Instead, it wraps every matching occurrence of the word within token spans.

### Comment stripping

The annotation comment is removed from the output. If the entire line is a comment containing only the annotation (no other code), the line itself is removed and subsequent line numbers shift down.

### Focus behavior

When any `[!code focus]` annotation is present:

- The `<pre>` element gets class `has-focused`
- The annotated line gets class `focused`
- Every other line gets class `dimmed`

Style `has-focused .line:not(.focused)` with reduced opacity to create a spotlight effect.

## Meta string highlighting

`transformers.Meta()` highlights lines based on a code-fence meta string:

```go
html, err := h.CodeToHTML(ctx, code, nuri.CodeToHTMLOptions{
	Lang:  "go",
	Theme: "github-dark",
	Transformers: []nuri.Transformer{
		transformers.Meta("{1,3-5}"),
	},
})
```

The range expression `{1,3-5}` highlights lines 1, 3, 4, and 5. Highlighted lines get class `highlighted`, the same class produced by `[!code highlight]`.

The `Meta` transformer works by appending parsed ranges to `CodeToHTMLOptions.HighlightLines` during the Preprocess hook. The renderer applies the class.

### Standalone parsing

`transformers.ParseMetaRanges()` is exported for consumers who need the range parsing without the transformer:

```go
ranges, err := transformers.ParseMetaRanges("{1,3-5}")
```

Returns `([]nuri.LineRange, error)`. Returns `nil, nil` for empty input or input without `{}`.

## Visible whitespace

`transformers.Whitespace()` replaces tab and space characters with visible symbols:

```go
html, err := h.CodeToHTML(ctx, code, nuri.CodeToHTMLOptions{
	Lang:  "go",
	Theme: "github-dark",
	Transformers: []nuri.Transformer{transformers.Whitespace()},
})
```

| Character | Symbol | CSS class |
|---|---|---|
| Tab | `→` | `ws-tab` |
| Space | `·` | `ws-space` |

Each replaced character is wrapped in a `<span>` with the corresponding class.

### Custom symbols

Override the default symbols with `transformers.WhitespaceWith()`:

```go
transformers.WhitespaceWith(transformers.WhitespaceOptions{
	Tab:   "⇥",
	Space: "⋅",
})
```

## Next steps

- [Custom Transformers](/docs/guides/custom-transformers) to write your own transformer hooks
- [Line Decorations](/docs/guides/line-decorations) for highlight, focus, and diff without magic comments
- [Style-to-Class Mode](/docs/guides/style-to-class-mode) to replace inline styles with class names
