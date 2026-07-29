---
title: Quick Start
description: "Create a highlighter and render syntax-highlighted HTML."
sidebar:
  order: 2
---

## Create a highlighter

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

`nuri.New()` compiles the WASM engine and builds the grammar/theme registry. `h.Close(ctx)` releases WASM resources; always defer it.

## Output

The call produces a self-contained `<pre>` element with inline color styles:

```html
<pre class="nuri github-dark" style="background-color:#24292e;color:#e1e4e8" tabindex="0">
  <code>
    <span class="line">
      <span style="color:#E1E4E8">fmt.</span>
      <span style="color:#79B8FF">Println</span>
      <span style="color:#E1E4E8">(</span>
      <span style="color:#9ECBFF">"hello, world"</span>
      <span style="color:#E1E4E8">)</span>
    </span>
  </code>
</pre>
```

Each token gets a `<span>` with the resolved color from the theme. Lines are wrapped in `<span class="line">`.

## Next steps

- [Using Themes](/docs/guides/using-themes) to select and switch between themes
- [Output Formats](/docs/guides/output-formats) for ANSI, SVG, JSON, plain text, and raw tokens
- [Transformers](/docs/guides/transformers) for line highlighting, diff annotations, and custom hooks
