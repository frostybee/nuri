---
title: Custom Languages & Themes
description: "Register custom grammars and themes, configure aliases, and detect languages."
sidebar:
  order: 8
---

Register grammars and themes at construction time, at runtime, or from a filesystem.

## At construction time

Pass `WithGrammar` and `WithTheme` with raw JSON bytes when creating the highlighter:

```go
grammarJSON, _ := os.ReadFile("my-language.tmLanguage.json")
themeJSON, _ := os.ReadFile("my-theme.json")

h, err := nuri.New(ctx,
	nuri.WithFS(core.FS()),
	nuri.WithGrammar("my-language", grammarJSON),
	nuri.WithTheme("my-theme", themeJSON),
	nuri.WithAlias("ml", "my-language"),
	nuri.WithExtension("mlg", "my-language"),
)
```

| Option | Description |
|---|---|
| `nuri.WithGrammar(name, data)` | Register a grammar from JSON bytes |
| `nuri.WithTheme(name, data)` | Register a theme from JSON bytes |
| `nuri.WithAlias(alias, target)` | Map an alias to a canonical language name |
| `nuri.WithExtension(ext, lang)` | Map a file extension (without dot) to a language |

## At runtime

Add grammars and themes to a running highlighter:

```go
grammarJSON, _ := os.ReadFile("hcl.tmLanguage.json")
if err := h.LoadLanguage("hcl", grammarJSON); err != nil {
	log.Fatal(err)
}

themeJSON, _ := os.ReadFile("catppuccin-latte.json")
if err := h.LoadTheme("catppuccin-latte", themeJSON); err != nil {
	log.Fatal(err)
}

h.RegisterAlias("terraform", "hcl")
h.RegisterExtension("tf", "hcl")
```

Runtime registration adds to the existing registry. Use this when loading grammars or themes on demand or from user-provided data.

## From a filesystem

Point the highlighter at directories of grammar and theme JSON files:

```go
h, err := nuri.New(ctx,
	nuri.WithGrammarFS(os.DirFS("./grammars")),
	nuri.WithThemeFS(os.DirFS("./themes")),
)
```

`nuri.WithFS()` is a shorthand when both grammars and themes live under a single root with `grammars/` and `themes/` subdirectories:

```go
h, err := nuri.New(ctx, nuri.WithFS(os.DirFS("./assets")))
```

The filesystem is scanned at construction time. Grammars and themes load lazily on first use.

## Expected JSON formats

### Grammar (TextMate format)

```json
{
  "scopeName": "source.example",
  "name": "Example Language",
  "patterns": [
    { "match": "\\b(func|return)\\b", "name": "keyword.control.example" }
  ],
  "repository": {}
}
```

`scopeName` is required. The format follows the standard TextMate grammar structure: `patterns`, `repository`, `begin`/`end`, `captures`, `injections`, and `injectionSelector`. Unknown fields are silently ignored.

### Theme (VS Code format)

```json
{
  "name": "My Theme",
  "type": "dark",
  "colors": {
    "editor.background": "#1e1e1e",
    "editor.foreground": "#d4d4d4"
  },
  "tokenColors": [
    {
      "scope": "keyword.control",
      "settings": {
        "foreground": "#569cd6",
        "fontStyle": "bold"
      }
    }
  ]
}
```

`scope` accepts a string (`"keyword.control"`) or an array (`["keyword", "storage"]`). `fontStyle` is a space-separated string: `"italic"`, `"bold"`, `"underline"`, `"strikethrough"`. The `semanticTokenColors` field (present in real VS Code themes) is silently ignored.

## Aliases

About 80 aliases are pre-registered from the upstream grammar collection. A sample:

| Alias | Canonical name |
|---|---|
| `sh`, `bash`, `zsh` | `shellscript` |
| `py` | `python` |
| `js` | `javascript` |
| `ts` | `typescript` |
| `c++` | `cpp` |
| `c#` | `csharp` |
| `rs` | `rust` |
| `md` | `markdown` |
| `yml` | `yaml` |
| `dockerfile` | `docker` |

Use `nuri.WithAlias()` or `h.RegisterAlias()` to add custom aliases.

## File extension mapping

58 extensions are pre-registered. Extensions are stored without the leading dot and matched case-insensitively. `nuri.WithExtension()` and `h.RegisterExtension()` override existing mappings for a given extension.

## Language detection

### By filename

```go
lang, found := h.DetectLanguage("src/main.go")
// lang = "go", found = true
```

`h.DetectLanguage()` accepts a filename or full path (it extracts the base name internally). It checks exact filenames first (`Makefile`, `Dockerfile`, `Gemfile`, `Rakefile`), then file extensions.

### By content

```go
lang, found := h.DetectLanguageByContent("#!/usr/bin/env python3")
// lang = "python", found = true
```

`h.DetectLanguageByContent()` runs `firstLineMatch` regular expressions from grammar JSON files against the provided string. This handles shebang lines and other first-line markers.

## Inspecting loaded languages

```go
names := h.LoadedLanguages()
// Returns sorted names of all currently cached grammars.
```

`h.LoadedLanguages()` returns the canonical names of grammars that have been loaded (either from the bundle or via `LoadLanguage`). Grammars are loaded lazily on first use, so the list grows as languages are highlighted.

## Next steps

- [Contrast Correction](/docs/guides/contrast-correction) for WCAG 2.1 automatic color adjustment
- [Performance & Concurrency](/docs/guides/performance-and-concurrency) for pool sizing and safety valves
