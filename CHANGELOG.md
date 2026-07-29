# Changelog

All notable changes to Nuri are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.1] - 2026-06-27

### Changed

- **The `<pre>` element class is now `nuri`, not `shiki`.** Single-theme output renders `class="nuri <theme-name>"` and multi-theme output renders `class="nuri nuri-themes <theme-names>"`. Stylesheets, transformers, and downstream selectors targeting `.shiki` must be updated to `.nuri`.

## [1.0.0] - 2026-06-24

Initial release.

### Added

- **TextMate grammar engine.** A port of vscode-textmate's tokenizer covering begin/end rules, begin/while rules, captures, end-pattern backreferences, the `\G` anchor, injections, embedded grammars, and `$self` / `$base` / `#repository` includes.
- **Real Oniguruma, no CGO.** The Oniguruma regex engine is compiled to WebAssembly and run through wazero, a pure Go WASM runtime. This is the same regex binary Shiki uses, so matching behavior is identical down to the edge cases. An opt-in `onig_cgo` build tag swaps in native Oniguruma for higher throughput.
- **Shiki fidelity.** 227 of 234 tested grammars produce output byte-identical to Shiki. The core bundle's 32 fidelity-tested grammars are at 100%. Golden fixtures are generated from a pinned Shiki version and recorded in `provenance.lock.json`.
- **Six output formats.** `h.CodeToHTML()`, `h.CodeToANSI()`, `h.CodeToSVG()`, `h.CodeToJSON()`, `h.CodeToPlainText()`, and `h.CodeToTokens()` for raw token access.
- **Multi-theme mode.** Tokenize once and resolve colors from several themes in a single pass. The default theme emits inline styles; the rest emit CSS variables, so theme switching is pure CSS.
- **Transformer pipeline.** Notation comments (`[!code ++]`, `[!code highlight]`, `[!code focus]`), meta string ranges (`{1,3-5}`), visible whitespace rendering, and custom hooks over the HTML AST through the `Transformer` interface.
- **Line decorations.** Per-line and per-range class, style, and attribute injection via `LineRange`, independent of the notation transformers.
- **Style-to-class mode.** Deterministic hashed class names in place of inline styles, plus a generated stylesheet shared across every code block on a page.
- **WCAG 2.1 contrast correction.** Token foreground colors are adjusted against the theme background to meet a minimum contrast ratio. Configurable with `nuri.WithMinContrast()`; the default ratio is 5.5.
- **Concurrency safety.** WASM instances are held in a bounded LIFO pool sized to `runtime.NumCPU()` by default and adjustable with `nuri.WithPoolSize()`. Highlighting from multiple goroutines is safe.
- **Tokenizer safety.** A per-line byte-length pre-filter (`nuri.WithMaxLineLength()`), a per-line soft timeout (`nuri.WithTimeoutMs()`), WASM-level regex interruption, and per-line panic recovery. Each degradation emits the line unstyled and records a `Diagnostic` rather than failing the call.
- **Language detection.** Resolution by file extension, exact filename, and first-line shebang, plus runtime alias and extension registration.
- **Two bundles.** `bundle/core` ships 38 popular languages and 65 themes in roughly 0.5 MB; `bundle/full` ships all 257 languages in roughly 1.5 MB. Assets are stored minified and gzipped behind a transparent decompressing filesystem.
- **Custom assets.** Constructors accept any `fs.FS`, so grammars and themes can come from disk, an embedded filesystem, or a test filesystem. `h.LoadLanguage()` and `h.LoadTheme()` register additional assets at runtime.
- **Standalone theme package.** `theme.Parse()`, `theme.Match()`, and `theme.Store` expose VS Code theme parsing and scope matching independently of the highlighter.
- **Compilation cache.** `nuri.WithCompilationCacheDir()` persists the ahead-of-time compiled WASM module across process invocations, removing cold-start cost for short-lived builds.

## Releasing

The changelog is maintained by hand. To cut a release:

1. Add entries under `[Unreleased]` as work lands on `main`.
2. Rename `[Unreleased]` to the new version with today's date, add a fresh empty `[Unreleased]` above it, and update the link references at the bottom of this file.
3. Mirror the new section into `docs/content/docs/changelog.md`.
4. Tag the release: `git tag -a vX.Y.Z -m "vX.Y.Z"` and push it with `git push origin vX.Y.Z`.
5. Publish the GitHub release with the notes from this file: `gh release create vX.Y.Z --notes-file <section>`.

[Unreleased]: https://github.com/frostybee/nuri/compare/v1.0.1...HEAD
[1.0.1]: https://github.com/frostybee/nuri/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/frostybee/nuri/releases/tag/v1.0.0
