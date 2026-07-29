---
title: Nuri
description: "Pure Go port of Shiki, the TextMate grammar-based syntax highlighter used by VS Code."
sidebar:
  order: 0
  label: "Introduction"
---

Nuri is a pure Go port of [Shiki](https://shiki.style), the TextMate grammar-based syntax highlighter used by VS Code. It runs the real Oniguruma regex engine compiled to WASM via [wazero](https://wazero.io), a pure Go WASM runtime. No CGO, no Node.js, no subprocess.

257 languages, 65 themes, full TextMate grammar support. 227 of 234 tested grammars (97%) produce output byte-identical to Shiki; the core bundle's 32 fidelity-tested grammars are at 100%.

## Minimal example

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/frostybee/nuri"
	"github.com/frostybee/nuri/bundle/core"
)

func main() {
	ctx := context.Background()

	h, err := nuri.New(ctx, nuri.WithFS(core.FS()))
	if err != nil {
		log.Fatal(err)
	}
	defer h.Close(ctx)

	html, err := h.CodeToHTML(ctx, `fmt.Println("hello, world")`, nuri.CodeToHTMLOptions{
		Lang:  "go",
		Theme: "github-dark",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(html)
}
```

Output:

```html
<pre class="nuri github-dark" style="background-color:#24292e;color:#e1e4e8" tabindex="0"><code><span class="line"><span style="color:#E1E4E8">fmt.</span><span style="color:#79B8FF">Println</span><span style="color:#E1E4E8">(</span><span style="color:#9ECBFF">"hello, world"</span><span style="color:#E1E4E8">)</span></span></code></pre>
```

## What Nuri provides

- **Six output formats** via `h.CodeToHTML()`, `h.CodeToANSI()`, `h.CodeToSVG()`, `h.CodeToJSON()`, `h.CodeToPlainText()`, and `h.CodeToTokens()`
- **Multi-theme mode** that tokenizes once and resolves multiple themes through CSS variables
- **Transformer pipeline** with built-in support for notation comments (`[!code ++]`, `[!code highlight]`), meta string ranges, visible whitespace, and custom hooks
- **Style-to-class mode** replacing inline styles with deterministic hashed class names
- **WCAG 2.1 contrast correction** that adjusts foreground colors against the theme background (default ratio 5.5)
- **Concurrency-safe** bounded pool of WASM instances for parallel highlighting
- **Language detection** by file extension, exact filename, or first-line shebang
- **Two bundle sizes**: `bundle/core` (38 languages, ~0.5 MB) and `bundle/full` (257 languages, ~1.5 MB)

## Where to go next

- [Installation](/docs/getting-started/installation) and [Quick Start](/docs/getting-started/quick-start) to set up and run Nuri
- [Guides](/docs/guides/using-themes) for task-oriented walkthroughs: themes, output formats, transformers, line decorations, and more
- [Reference](/docs/reference/highlighter-api) for complete API documentation: every method, option, and type
