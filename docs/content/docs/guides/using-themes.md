---
title: Using Themes
description: "Select a theme for syntax highlighting and access theme colors."
sidebar:
  order: 1
---

Pass a theme name in `CodeToHTMLOptions.Theme` to color tokens from a VS Code theme.

## Select a theme

```go
html, err := h.CodeToHTML(ctx, code, nuri.CodeToHTMLOptions{
	Lang:  "typescript",
	Theme: "github-dark",
})
```

The theme resolves colors from a VS Code theme JSON bundled in the filesystem. Themes load lazily on first use.

Swap the name to change the color scheme:

```go
html, err := h.CodeToHTML(ctx, code, nuri.CodeToHTMLOptions{
	Lang:  "typescript",
	Theme: "dracula",
})
```

The same theme name works across all output methods (`h.CodeToANSI()`, `h.CodeToSVG()`, etc.).

## Available themes

65 themes ship in both the core and full bundles. A sample of the included names:

andromeeda, ayu-dark, catppuccin-mocha, dark-plus, dracula, everforest-dark, github-dark, github-light, min-dark, monokai, nord, one-dark-pro, rose-pine, slack-dark, solarized-dark, tokyo-night, vitesse-dark, and more.

For the complete list, see [Bundled Languages & Themes](/docs/reference/bundled-languages-and-themes).

`h.LoadedThemes()` returns the names of themes that have been loaded into the cache so far:

```go
names := h.LoadedThemes()
fmt.Println(names) // e.g. ["github-dark"]
```

This reflects cached themes, not the full set of available themes. A theme enters the cache the first time it is used in a highlighting call or loaded explicitly.

## Theme colors

`h.GetThemeColors()` returns the UI colors for a loaded theme. Use these to style code block chrome (title bars, copy buttons, terminal frames) consistently with the highlighted code:

```go
colors, err := h.GetThemeColors("github-dark")
if err != nil {
	log.Fatal(err)
}
fmt.Println(colors.Type)       // "dark"
fmt.Println(colors.Background) // "#24292e"
fmt.Println(colors.Foreground) // "#e1e4e8"
```

| Field | Type | Description |
|---|---|---|
| `Type` | `string` | `"dark"` or `"light"` |
| `Background` | `string` | Editor background color |
| `Foreground` | `string` | Editor foreground color |
| `SelectionBackground` | `string` | Selection highlight color |
| `LineHighlightBg` | `string` | Current-line highlight color |
| `Colors` | `map[string]string` | Full VS Code color map (`editor.*`, `terminal.*`, etc.) |

The `Colors` map contains all color keys defined in the theme JSON. Access specific keys when the top-level fields are not enough:

```go
tabBorder := colors.Colors["editorGroupHeader.tabsBorder"]
```

## Contrast correction

Nuri adjusts token foreground colors that fail a WCAG 2.1 contrast check against the theme background. The default ratio is 5.5 (WCAG AA enhanced). Adjusted themes are cached, so there is no per-token cost. `h.GetThemeColors()` returns contrast-adjusted colors.

See the [Contrast Correction](/docs/guides/contrast-correction) guide to configure or disable this behavior.

## Next steps

- [Multi-Theme Mode](/docs/guides/multi-theme-mode) for light/dark switching with CSS variables
- [Output Formats](/docs/guides/output-formats) for ANSI, SVG, JSON, and more
