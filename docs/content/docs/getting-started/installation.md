---
title: Installation
description: "Install Nuri and choose a grammar bundle."
sidebar:
  order: 1
---

## Requirements

Go 1.25 or later.

## Install

```bash
go get github.com/frostybee/nuri
```

## Choose a bundle

Nuri ships two bundle packages. Both include the same 65 themes. Import one to embed grammar and theme assets into the binary:

### Core bundle

```go
import "github.com/frostybee/nuri/bundle/core"
```

38 languages, ~0.5 MB embedded. Covers the most common web and backend languages:

bat, c, cpp, csharp, css, docker, go, graphql, html, java, javascript, json, jsonc, jsx, kotlin, lua, markdown, php, python, ruby, rust, scss, shellscript, sql, svelte, swift, toml, tsx, typescript, vue, xml, yaml, and injection helpers (html-derivative, cpp-macro, regex, sassdoc, jsdoc, css-additions).

### Full bundle

```go
import "github.com/frostybee/nuri/bundle/full"
```

257 languages, ~1.5 MB embedded. Every grammar from [shikijs/textmate-grammars-themes](https://github.com/shikijs/textmate-grammars-themes). Use this when language coverage matters more than binary size.

### Passing the bundle to the highlighter

Both bundles export an `FS()` function that returns an `fs.FS`. Pass it to `nuri.WithFS()` when creating the highlighter:

```go
h, err := nuri.New(ctx, nuri.WithFS(core.FS()))
```

`nuri.WithFS()` is the intended way to use the bundle packages. It sets both the grammar and theme filesystems from a single `fs.FS` that contains `grammars/` and `themes/` subdirectories.

## Custom assets

The constructor accepts any `fs.FS`, so `os.DirFS()` works for loading grammars and themes from disk at runtime. See the [Custom Languages & Themes](/docs/guides/custom-languages-and-themes) guide for details.

## Next steps

- [Quick Start](/docs/getting-started/quick-start) to create a highlighter and render HTML
- [Using Themes](/docs/guides/using-themes) to select and switch themes
- [Output Formats](/docs/guides/output-formats) for HTML, ANSI, SVG, and more
