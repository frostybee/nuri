---
title: "Changelog"
description: "Release notes and version history for Nuri."
sidebar:
  order: 5
---

Every released version of Nuri and what changed in it. Entries follow [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and version numbers follow [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

The canonical copy lives in [`CHANGELOG.md`](https://github.com/frostybee/nuri/blob/main/CHANGELOG.md) at the repository root. Downloadable archives and release notes are on the [GitHub releases page](https://github.com/frostybee/nuri/releases).

## v1.0.1

Released 2026-06-27.

### Changed

- **The `<pre>` element class is now `nuri`, not `shiki`.** Single-theme output renders `class="nuri <theme-name>"` and multi-theme output renders `class="nuri nuri-themes <theme-names>"`. Stylesheets, transformers, and downstream selectors targeting `.shiki` must be updated to `.nuri`. See [Output Formats](/docs/guides/output-formats).

## v1.0.0

Released 2026-06-24. Initial release.

### Added

- **TextMate grammar engine.** A port of vscode-textmate's tokenizer covering begin/end rules, begin/while rules, captures, end-pattern backreferences, the `\G` anchor, injections, embedded grammars, and `$self` / `$base` / `#repository` includes. See [Shiki Compatibility](/docs/internals/shiki-compatibility).
- **Real Oniguruma, no CGO.** The Oniguruma regex engine is compiled to WebAssembly and run through wazero, a pure Go WASM runtime. This is the same regex binary Shiki uses, so matching behavior is identical down to the edge cases. An opt-in `onig_cgo` build tag swaps in native Oniguruma for higher throughput. See [Architecture](/docs/internals/architecture).
- **Shiki fidelity.** 227 of 234 tested grammars produce output byte-identical to Shiki. The core bundle's 32 fidelity-tested grammars are at 100%. Golden fixtures are generated from a pinned Shiki version and recorded in `provenance.lock.json`. See [Fidelity Testing](/docs/internals/fidelity-testing).
- **Six output formats.** `h.CodeToHTML()`, `h.CodeToANSI()`, `h.CodeToSVG()`, `h.CodeToJSON()`, `h.CodeToPlainText()`, and `h.CodeToTokens()` for raw token access. See [Output Formats](/docs/guides/output-formats).
- **Multi-theme mode.** Tokenize once and resolve colors from several themes in a single pass. The default theme emits inline styles; the rest emit CSS variables, so theme switching is pure CSS. See [Multi-Theme Mode](/docs/guides/multi-theme-mode).
- **Transformer pipeline.** Notation comments (`[!code ++]`, `[!code highlight]`, `[!code focus]`), meta string ranges (`{1,3-5}`), visible whitespace rendering, and custom hooks over the HTML AST through the `Transformer` interface. See [Transformers](/docs/guides/transformers) and [Custom Transformers](/docs/guides/custom-transformers).
- **Line decorations.** Per-line and per-range class, style, and attribute injection via `LineRange`, independent of the notation transformers. See [Line Decorations](/docs/guides/line-decorations).
- **Style-to-class mode.** Deterministic hashed class names in place of inline styles, plus a generated stylesheet shared across every code block on a page. See [Style-to-Class Mode](/docs/guides/style-to-class-mode).
- **WCAG 2.1 contrast correction.** Token foreground colors are adjusted against the theme background to meet a minimum contrast ratio. Configurable with `nuri.WithMinContrast()`; the default ratio is 5.5. See [Contrast Correction](/docs/guides/contrast-correction).
- **Concurrency safety.** WASM instances are held in a bounded LIFO pool sized to `runtime.NumCPU()` by default and adjustable with `nuri.WithPoolSize()`. Highlighting from multiple goroutines is safe. See [Performance and Concurrency](/docs/guides/performance-and-concurrency).
- **Tokenizer safety.** A per-line byte-length pre-filter (`nuri.WithMaxLineLength()`), a per-line soft timeout (`nuri.WithTimeoutMs()`), WASM-level regex interruption, and per-line panic recovery. Each degradation emits the line unstyled and records a `Diagnostic` rather than failing the call. See [Errors and Diagnostics](/docs/reference/errors-and-diagnostics).
- **Language detection.** Resolution by file extension, exact filename, and first-line shebang, plus runtime alias and extension registration. See [Custom Languages and Themes](/docs/guides/custom-languages-and-themes).
- **Two bundles.** `bundle/core` ships 38 popular languages and 65 themes in roughly 0.5 MB; `bundle/full` ships all 257 languages in roughly 1.5 MB. Assets are stored minified and gzipped behind a transparent decompressing filesystem. See [Bundled Languages and Themes](/docs/reference/bundled-languages-and-themes).
- **Custom assets.** Constructors accept any `fs.FS`, so grammars and themes can come from disk, an embedded filesystem, or a test filesystem. `h.LoadLanguage()` and `h.LoadTheme()` register additional assets at runtime. See [Custom Languages and Themes](/docs/guides/custom-languages-and-themes).
- **Standalone theme package.** `theme.Parse()`, `theme.Match()`, and `theme.Store` expose VS Code theme parsing and scope matching independently of the highlighter. See [Theme Package](/docs/reference/theme-package).
- **Compilation cache.** `nuri.WithCompilationCacheDir()` persists the ahead-of-time compiled WASM module across process invocations, removing cold-start cost for short-lived builds. See [Configuration](/docs/reference/configuration).
