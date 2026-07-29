---
title: Line Decorations
description: "Highlight, focus, and mark lines as inserted or deleted."
sidebar:
  order: 7
---

Four fields on `CodeToHTMLOptions` apply CSS classes to specific lines without magic comments.

## Line ranges

```go
html, err := h.CodeToHTML(ctx, code, nuri.CodeToHTMLOptions{
	Lang:           "go",
	Theme:          "github-dark",
	HighlightLines: []nuri.LineRange{nuri.Range(3, 5), nuri.Range(10, 10)},
	FocusLines:     nuri.Lines(1, 2, 3),
	InsertedLines:  nuri.Lines(7),
	DeletedLines:   nuri.Lines(8),
})
```

### Constructors

- `nuri.Range(start, end)` returns a `LineRange` covering lines `start` through `end` (1-based, inclusive)
- `nuri.Lines(nums...)` returns a `[]LineRange` where each number becomes a single-line range

Line numbers are 1-based throughout.

## Decoration classes

| Field | Classes added to line | Notes |
|---|---|---|
| `HighlightLines` | `highlighted` | |
| `FocusLines` | `focused` or `dimmed` | `has-focused` added to `<pre>` |
| `InsertedLines` | `diff`, `add` | |
| `DeletedLines` | `diff`, `remove` | |

## Highlight

```go
html, err := h.CodeToHTML(ctx, code, nuri.CodeToHTMLOptions{
	Lang:           "go",
	Theme:          "github-dark",
	HighlightLines: []nuri.LineRange{nuri.Range(2, 4)},
})
```

Lines 2, 3, and 4 get class `highlighted`. This is the same class produced by `[!code highlight]` and `transformers.Meta()`.

## Focus

```go
html, err := h.CodeToHTML(ctx, code, nuri.CodeToHTMLOptions{
	Lang:       "go",
	Theme:      "github-dark",
	FocusLines: nuri.Lines(3, 4, 5),
})
```

When `FocusLines` is non-empty:

- The `<pre>` element gets class `has-focused`
- Lines in the range get class `focused`
- All other lines get class `dimmed`

Style the dimmed lines with reduced opacity for a spotlight effect:

```css
.nuri.has-focused .line:not(.focused) {
  opacity: 0.3;
  transition: opacity 0.3s;
}
.nuri.has-focused:hover .line:not(.focused) {
  opacity: 1;
}
```

## Diff

```go
html, err := h.CodeToHTML(ctx, code, nuri.CodeToHTMLOptions{
	Lang:          "go",
	Theme:         "github-dark",
	InsertedLines: nuri.Lines(3),
	DeletedLines:  nuri.Lines(5),
})
```

- Inserted lines get classes `diff` and `add`
- Deleted lines get classes `diff` and `remove`

These are the same classes produced by `[!code ++]` and `[!code --]`.

```css
.nuri .line.diff.add {
  background-color: rgba(16, 185, 129, 0.14);
}
.nuri .line.diff.remove {
  background-color: rgba(244, 63, 94, 0.14);
}
```

## Interaction with transformers

Decorations and the `Notation` transformer produce the same CSS classes. Both can be used in the same call. A line matched by both `HighlightLines` and `[!code highlight]` gets the class once (`AddClass` is idempotent).

The `Meta` transformer also uses `HighlightLines` internally. It appends parsed ranges to `CodeToHTMLOptions.HighlightLines` during its `Preprocess` hook.

## Next steps

- [Custom Languages & Themes](/docs/guides/custom-languages-and-themes) to register custom grammars and themes
- [Transformers](/docs/guides/transformers) for notation comments and meta string highlighting
