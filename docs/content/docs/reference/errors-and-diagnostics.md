---
title: Errors & Diagnostics
description: "Sentinel errors, error wrapping, and non-fatal diagnostic reporting."
sidebar:
  order: 8
---

Nuri distinguishes between fatal errors (returned as `error`) and non-fatal degradations (recorded as `Diagnostic` entries in the result). Fatal errors stop the highlighting call. Non-fatal degradations allow tokenization to continue for subsequent lines.

## Sentinel errors

```go
var (
    ErrLanguageNotFound = registry.ErrLanguageNotFound
    ErrThemeNotFound    = registry.ErrThemeNotFound
    ErrGrammarCycle     = grammar.ErrGrammarCycle
    ErrGrammarDepth     = grammar.ErrGrammarDepth
)
```

All four are wrapped with `%w` by the functions that return them. Use `errors.Is` to check:

```go
if errors.Is(err, nuri.ErrThemeNotFound) {
    // theme name not in registry
}
```

### `nuri.ErrLanguageNotFound`

The grammar name is not in the registry. In `h.CodeToTokens()` and `h.CodeToHTML()`, this triggers a plaintext fallback with a `"unknown_lang"` diagnostic rather than a hard error. Use `errors.Is` to promote it to a hard error when strict behavior is needed.

### `nuri.ErrThemeNotFound`

The theme name is not in the registry. Returned by `h.GetThemeColors()` and by the internal theme resolver used by all highlighting methods.

### `nuri.ErrGrammarCycle`

Grammar include resolution detected a cycle (e.g. grammar A includes grammar B, which includes grammar A). Returned during tokenization when the grammar dependency graph contains a loop.

### `nuri.ErrGrammarDepth`

Grammar include resolution exceeded the depth limit. Returned when deeply nested grammar includes exceed the maximum allowed depth, even if no cycle exists.

## Fatal vs non-fatal

**Fatal errors** are returned as the `error` value from highlighting methods:

- WASM trap (runtime failure inside the Oniguruma module)
- Context cancellation (`ctx.Done()` before or during tokenization)
- Malformed grammar or theme JSON (from `h.LoadLanguage()` or `h.LoadTheme()`)
- Theme not found (from `h.GetThemeColors()`)

**Non-fatal degradations** are recorded in `TokensResult.Diagnostics`. The highlighting call succeeds, and the affected line is emitted as a single unstyled token or with partial styling. Tokenization continues for subsequent lines.

## Diagnostics

```go
type Diagnostic struct {
    Line int    `json:"line"`
    Kind string `json:"kind"`
}
```

`Line` is 0-based.

| Kind | Trigger | What happens |
|---|---|---|
| `"timeout"` | Line tokenization exceeded `TimeoutMs` | Partial tokens preserved; WASM instance replaced with a fresh one |
| `"too_long"` | Line byte length exceeded `MaxLineLength` | Single unstyled token emitted; no WASM call made |
| `"panic"` | Tokenizer panicked on a line | Instance marked poisoned and replaced |
| `"unknown_lang"` | Language name not found in registry | Entire input rendered as plaintext |

## Checking diagnostics

```go
result, err := h.CodeToTokens(ctx, code, nuri.CodeToTokensOptions{
    Lang:  "go",
    Theme: "github-dark",
})
if err != nil {
    log.Fatal(err)
}
for _, d := range result.Diagnostics {
    log.Printf("line %d: %s", d.Line, d.Kind)
}
```

`h.CodeToHTML()` applies the same diagnostic model internally. The diagnostics are not directly accessible from the HTML output, but the affected lines appear as unstyled `<span>` elements in the rendered HTML.
