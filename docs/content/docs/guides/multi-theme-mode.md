---
title: Multi-Theme Mode
description: "Tokenize once and emit styles for multiple themes via CSS variables."
sidebar:
  order: 2
---

Multi-theme mode tokenizes code once and resolves colors from multiple themes. The default theme emits inline styles. Other themes emit CSS variables. Multi-theme is available in `CodeToHTML`, `CodeToTokens`, and `CodeToJSON`. The other output formats (`CodeToANSI`, `CodeToSVG`, `CodeToPlainText`) operate in single-theme mode only.

## Enable multi-theme

Pass a `Themes` map instead of a single `Theme`:

```go
html, err := h.CodeToHTML(ctx, code, nuri.CodeToHTMLOptions{
	Lang: "typescript",
	Themes: map[string]string{
		"dark":  "github-dark",
		"light": "github-light",
	},
})
```

The lexicographically first key (`"dark"`) becomes the default theme with inline `color:` styles. The `"light"` theme produces `--nuri-light` CSS variables on each token.

Output:

```html
<pre class="nuri nuri-themes dark light"
     style="background-color:#24292e;color:#e1e4e8;--nuri-light:#fff;--nuri-light-bg:#24292e"
     tabindex="0">
  <code>
    <span class="line">
      <span style="color:#F97583;--nuri-light:#D73A49">const</span>
      <span style="color:#E1E4E8;--nuri-light:#24292E"> x </span>
      ...
    </span>
  </code>
</pre>
```

## CSS variable naming

Variables follow the pattern `--nuri-{key}`:

On token `<span>` elements:

| Variable | Purpose |
|---|---|
| `--nuri-{key}` | Foreground color |
| `--nuri-{key}-bg` | Background color |
| `--nuri-{key}-font-style` | `italic` |
| `--nuri-{key}-font-weight` | `bold` |
| `--nuri-{key}-text-decoration` | `underline`, `line-through` |

On the `<pre>` element:

| Variable | Purpose |
|---|---|
| `--nuri-{key}` | Theme default foreground |
| `--nuri-{key}-bg` | Theme default background |

## Switch themes with CSS

Apply the non-default theme by reading its CSS variables. For a class-based toggle:

```css
html.light .nuri,
html.light .nuri span {
  color: var(--nuri-light) !important;
  background-color: var(--nuri-light-bg) !important;
  font-style: var(--nuri-light-font-style) !important;
  font-weight: var(--nuri-light-font-weight) !important;
  text-decoration: var(--nuri-light-text-decoration) !important;
}
```

For automatic switching based on system preference:

```css
@media (prefers-color-scheme: light) {
  .nuri,
  .nuri span {
    color: var(--nuri-light) !important;
    background-color: var(--nuri-light-bg) !important;
    font-style: var(--nuri-light-font-style) !important;
    font-weight: var(--nuri-light-font-weight) !important;
    text-decoration: var(--nuri-light-text-decoration) !important;
  }
}
```

## Suppress inline styles

`DefaultColor` controls whether the default theme emits inline `color:` and `background-color:` properties:

```go
f := false
html, err := h.CodeToHTML(ctx, code, nuri.CodeToHTMLOptions{
	Lang: "go",
	Themes: map[string]string{
		"dark":  "github-dark",
		"light": "github-light",
	},
	DefaultColor: &f,
})
```

| Value | Behavior |
|---|---|
| `nil` (default) | Inline styles for the default theme, CSS variables for others |
| `&true` | Same as `nil` |
| `&false` | CSS variables only; no inline `color:` or `background-color:` on any element |

Set `DefaultColor` to `&false` when both themes are activated through CSS variables and neither needs inline styles.

## Multi-theme token data

`h.CodeToTokens()` also supports multi-theme. The `TokensResult` exposes per-theme colors:

```go
result, err := h.CodeToTokens(ctx, code, nuri.CodeToTokensOptions{
	Lang: "go",
	Themes: map[string]string{
		"dark":  "github-dark",
		"light": "github-light",
	},
})

fmt.Println(result.FG)         // default theme foreground
fmt.Println(result.BG)         // default theme background
fmt.Println(result.ThemeFG)    // map["light"] → foreground
fmt.Println(result.ThemeBG)    // map["light"] → background
fmt.Println(result.ThemeNames) // ["dark", "light"]
```

| Field | Type | Description |
|---|---|---|
| `FG`, `BG` | `string` | Default theme colors |
| `ThemeFG`, `ThemeBG` | `map[string]string` | Non-default theme colors (keyed by theme key) |
| `ThemeNames` | `[]string` | All theme keys, sorted |
| `Tokens[i][j].ThemeStyles` | `map[string]TokenStyle` | Per-token styles for non-default themes |

The default theme's per-token style lives in `ThemedToken.Color`, `ThemedToken.BgColor`, and `ThemedToken.FontStyle`. Non-default themes live in `ThemedToken.ThemeStyles`.

## Next steps

- [Output Formats](/docs/guides/output-formats) for ANSI, SVG, JSON, plain text, and raw tokens
- [Style-to-Class Mode](/docs/guides/style-to-class-mode) to replace inline styles with class names
