---
title: Theme Package
description: "The standalone theme package for parsing, matching, and caching VS Code themes."
sidebar:
  order: 5
---

This page documents the `theme` package (`github.com/frostybee/nuri/theme`), which is exported for standalone use. Most consumers interact with themes through the highlighter methods (`h.GetThemeColors()`, the `Theme` field on options structs). Import the `theme` package directly when direct theme parsing, scope matching, or a custom theme cache is needed.

## `theme.Parse`

```go
func Parse(data []byte) (*Theme, error)
```

Parses a VS Code theme JSON file into a `*Theme` with pre-compiled selectors for fast scope matching.

Scopes in the JSON may appear as a comma-separated string, a plain string, or a `[]string` array; all forms are normalized into `[]string`. Rules with an empty or missing scope are skipped.

The `semanticTokenColors` field (present in real VS Code themes for LSP-driven highlighting) is silently ignored. Unknown JSON fields are also ignored.

**Returns:** An error on JSON parse failure. Errors include the `tokenColors` array index for rule-level failures (e.g. `"theme: tokenColors[3]: ..."`).

## `theme.Theme`

```go
type Theme struct {
    Name              string
    DisplayName       string
    Type              string
    Colors            map[string]string
    TokenColors       []TokenColor
    DefaultForeground string
    DefaultBackground string
}
```

A parsed VS Code color theme.

| Field | Description |
|---|---|
| `Name` | Internal name from the JSON `"name"` key |
| `DisplayName` | Human-facing name from the JSON `"displayName"` key |
| `Type` | `"dark"` or `"light"` |
| `Colors` | Full VS Code color map (`editor.*`, `terminal.*`, etc.) |
| `TokenColors` | Ordered token-color rules, preserving JSON order |
| `DefaultForeground` | Resolved from `Colors["editor.foreground"]`; falls back to `"#000000"` |
| `DefaultBackground` | Resolved from `Colors["editor.background"]`; falls back to `"#000000"` |

### `t.Base`

```go
func (t *Theme) Base() TokenSettings
```

Returns the default foreground and background as a `TokenSettings` with `FontStyle` set to `FontStyleNone`. Use this as the baseline before overlaying per-token results from `t.Match()`.

### `t.Match`

```go
func (t *Theme) Match(scopes []string) TokenSettings
```

Resolves the style for a token with the given scope stack. `scopes[0]` is the root scope (e.g. `"source.go"`), `scopes[len-1]` is the most specific (e.g. `"keyword.control.go"`).

Each property (foreground, background, fontStyle) is resolved independently from the highest-scoring matching rule. The scoring algorithm mirrors vscode-textmate:

1. **Stack depth:** deeper match wins (a rule matching at `scopes[4]` beats one matching at `scopes[2]`)
2. **Dot-segment count:** `keyword.control.go` beats `keyword` (more specific selector)
3. **Parent-scope parts:** `source.go keyword` beats `keyword` (contextual match)

**Returns:** `TokenSettings` where unmatched properties are zero values (`""` for colors, `FontStyleNotSet` for font style). Merge with `t.Base()` to fill in defaults.

Safe for concurrent use on a `*Theme` returned by `theme.Parse()`.

## `theme.TokenColor`

```go
type TokenColor struct {
    Scopes   []string
    Settings TokenSettings
}
```

A single rule mapping scope selectors to style settings. `Scopes` contains the selectors this rule applies to (e.g. `["keyword", "keyword.control"]`). When any selector matches the scope stack, `Settings` is considered for that property.

## `theme.TokenSettings`

```go
type TokenSettings struct {
    Foreground string
    Background string
    FontStyle  FontStyle
}
```

Style properties for a scope match. An empty string for `Foreground` or `Background` means the rule does not set that property. `FontStyleNotSet` means the rule does not set a font style.

## `theme.FontStyle`

```go
type FontStyle int8
```

A bitmask representing text decoration styles. Multiple flags can be combined with bitwise OR (e.g. `FontStyleItalic | FontStyleBold`).

| Constant | Value | Meaning |
|---|---|---|
| `FontStyleNotSet` | `-1` | No font-style information in the rule (sentinel) |
| `FontStyleNone` | `0` | Explicitly no decoration |
| `FontStyleItalic` | `1` | Italic |
| `FontStyleBold` | `2` | Bold |
| `FontStyleUnderline` | `4` | Underline |
| `FontStyleStrikethrough` | `8` | Strikethrough |

### `fs.Has`

```go
func (fs FontStyle) Has(flag FontStyle) bool
```

Reports whether `flag` is set in `fs`. Guard against `FontStyleNotSet` before calling, since the negative sentinel participates in bitwise operations.

### `fs.String`

```go
func (fs FontStyle) String() string
```

Returns a human-readable representation: `"notset"`, `"none"`, or a space-joined list of active flags (e.g. `"italic bold"`).

## `theme.Store`

```go
type Store struct { /* unexported fields */ }
```

A thread-safe cache for parsed themes backed by an `fs.FS`. Themes are loaded lazily on first access via `s.Get()` and cached for subsequent calls. The internal lock uses double-checked read/write locking.

### `theme.NewStore`

```go
func NewStore(fsys fs.FS) *Store
```

Creates a `Store` backed by the given filesystem. When `fsys` is non-nil, `s.Get()` reads `name + ".json"` from it for any uncached theme. Pass `nil` for a store with only explicitly registered themes.

### `s.Get`

```go
func (s *Store) Get(name string) (*Theme, error)
```

Returns a parsed theme by name, loading and caching it on first access.

**Lookup order:**

1. In-memory cache (read lock)
2. Registered bytes (added via `s.Register()`)
3. Filesystem: reads `name + ".json"` from the backing `fs.FS`

**Returns:** An error if the theme is not found in any source or if `theme.Parse()` fails.

### `s.Register`

```go
func (s *Store) Register(name string, data []byte) error
```

Adds raw theme JSON under the given name. Validates that `data` is well-formed JSON; returns an error if not. Overwrites any previously cached theme with the same name, so the next `s.Get()` re-parses from the new bytes. The theme is not parsed at registration time.

### `s.LoadedThemes`

```go
func (s *Store) LoadedThemes() []string
```

Returns the sorted names of all currently cached themes (those retrieved via `s.Get()` at least once). Does not include themes that were registered but not yet accessed.
