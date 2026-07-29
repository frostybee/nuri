---
title: Custom Transformers
description: "Write custom transformers to modify the HTML rendering pipeline."
sidebar:
  order: 5
---

Implement the `Transformer` interface to hook into the HTML rendering pipeline at nine lifecycle points.

## The Transformer interface

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

Return conventions:

- `Preprocess` / `Postprocess`: return `""` to leave unchanged, non-empty string to replace
- `Tokens`: return `nil` to leave unchanged, non-nil slice to replace
- `Span` / `Line` / `Code` / `Pre` / `Root`: return `nil` to leave unchanged, non-nil `*Element` to replace

## Hook execution order

1. **Preprocess** -- before tokenization. Receives the raw source and a pointer to the merged options. Can modify `code` (by returning a non-empty string) and mutate `opts` (the pointer allows side effects).
2. Tokenization runs on the (possibly replaced) source.
3. **Tokens** -- after tokenization. Receives the 2-D token grid (`[][]ThemedToken`, outer = lines, inner = tokens per line).
4. **Span** -- per token `<span>`, left-to-right within each line. `line` is 1-based. `col` is the UTF-8 byte offset within the line (0-based). `lineEl` is the containing `<span class="line">`.
5. **Line** -- per `<span class="line">`, after all its span hooks. `line` is 1-based.
6. **Code** -- once, on the `<code>` element.
7. **Pre** -- once, on the `<pre>` element.
8. **Root** -- once, on the top-level element (currently `<pre>`).
9. Serialization writes the tree to an HTML string.
10. **Postprocess** -- after serialization. Receives the final HTML string.

When multiple transformers are passed, each hook runs all transformers in slice order. For element hooks (Span through Root), the last non-nil return wins.

## BaseTransformer

Embed `nuri.BaseTransformer` to get no-op defaults for all hooks, then override only the hooks needed:

```go
type lineNumbers struct {
	nuri.BaseTransformer
}

func (lineNumbers) Name() string { return "line-numbers" }

func (lineNumbers) Line(el *nuri.Element, line int) *nuri.Element {
	el.SetAttr("data-line", strconv.Itoa(line))
	return el
}
```

Use it in the `Transformers` slice:

```go
html, err := h.CodeToHTML(ctx, code, nuri.CodeToHTMLOptions{
	Lang:         "go",
	Theme:        "github-dark",
	Transformers: []nuri.Transformer{lineNumbers{}},
})
```

Each `<span class="line">` in the output now carries a `data-line` attribute.

## The HTML AST

Transformer hooks receive and return `*nuri.Element` nodes. The AST has two node types:

### Element

```go
type Element struct {
	Tag      string
	Classes  []string
	Styles   map[string]string
	Attrs    map[string]string
	Children []Node
}
```

Helper methods:

- `el.AddClass(class)` -- appends a CSS class if not already present
- `el.SetAttr(key, val)` -- sets an HTML attribute (initializes the map if nil)

`Styles` and `Attrs` maps are emitted in sorted key order for deterministic output.

### Text

```go
type Text struct {
	Content string
}
```

`Content` is HTML-escaped (`<`, `>`, `&`) on write.

### Node

```go
type Node interface {
	WriteTo(w io.Writer) (int64, error)
}
```

Both `*Element` and `*Text` satisfy `Node` and `io.WriterTo`.

## Example: wrapping in a container

The `Root` hook runs after the tree is fully built. Use it to wrap the output in a container element:

```go
type codeBlock struct {
	nuri.BaseTransformer
}

func (codeBlock) Name() string { return "code-block" }

func (codeBlock) Root(el *nuri.Element) *nuri.Element {
	return &nuri.Element{
		Tag:      "div",
		Classes:  []string{"code-block"},
		Children: []nuri.Node{el},
	}
}
```

The output becomes `<div class="code-block"><pre class="nuri ...">...</pre></div>`.

## Example: modifying tokens

The `Tokens` hook runs after tokenization but before tree building. Use it to filter or transform the raw token data:

```go
type trimTrailing struct {
	nuri.BaseTransformer
}

func (trimTrailing) Name() string { return "trim-trailing" }

func (trimTrailing) Tokens(tokens [][]nuri.ThemedToken) [][]nuri.ThemedToken {
	for i, line := range tokens {
		if len(line) > 0 {
			last := &tokens[i][len(line)-1]
			last.Content = strings.TrimRight(last.Content, " \t")
		}
	}
	return tokens
}
```

## Next steps

- [Transformers](/docs/guides/transformers) for the built-in notation, meta, and whitespace transformers
- [Types](/docs/reference/types) for the full reference on Element, Text, Node, and all AST types
- [Transformer API](/docs/reference/transformer-api) for complete hook signatures and built-in constructor details
