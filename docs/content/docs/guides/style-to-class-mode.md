---
title: Style-to-Class Mode
description: "Replace inline styles with deterministic hashed class names."
sidebar:
  order: 6
---

Pass a `StyleClassMap` to replace inline color styles with CSS class names, producing a shared stylesheet for all code blocks.

## The workflow

```go
classMap := nuri.NewStyleClassMap()

html1, _ := h.CodeToHTML(ctx, goCode, nuri.CodeToHTMLOptions{
	Lang: "go", Theme: "nord", ClassMap: classMap,
})
html2, _ := h.CodeToHTML(ctx, jsCode, nuri.CodeToHTMLOptions{
	Lang: "javascript", Theme: "nord", ClassMap: classMap,
})

css := classMap.CSS()
```

The `StyleClassMap` accumulates unique style combinations across multiple `h.CodeToHTML()` calls. `classMap.CSS()` returns the combined stylesheet covering all code blocks.

## How it works

Each unique style combination gets a deterministic class name derived from an FNV-64a hash. The format is `_s_` followed by a hex string (for example, `_s_4a3f9c2b1d0e7f8a`).

Without `ClassMap`, a token `<span>` looks like:

```html
<span style="color:#ff7b72;font-style:italic">func</span>
```

With `ClassMap`, the same token becomes:

```html
<span class="_s_4a3f9c2b1d0e7f8a">func</span>
```

The `<pre>` element's background and foreground styles are also replaced with a class name. Identical style combinations across different code blocks, languages, or themes produce the same class name.

## The generated stylesheet

`classMap.CSS()` returns one rule per unique style combination, sorted by class name:

```css
._s_1a2b3c4d5e6f7890 { color: #79c0ff }
._s_4a3f9c2b1d0e7f8a { color: #ff7b72; font-style: italic }
._s_7e8f9a0b1c2d3e4f { color: #a5d6ff }
```

Embed this stylesheet in a `<style>` tag or serve it as a static CSS file.

## When to use

- **CSP policies** that forbid inline `style` attributes
- **Cacheable stylesheets** across pages that share the same theme
- **Smaller HTML** when many tokens share the same style (one class name vs. a full inline declaration on every `<span>`)

## Next steps

- [Line Decorations](/docs/guides/line-decorations) to highlight, focus, and diff lines
- [Transformers](/docs/guides/transformers) for notation comments and meta string highlighting
