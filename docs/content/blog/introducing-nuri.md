---
title: "Introducing Nuri: VS Code-Accurate Syntax Highlighting in Pure Go"
date: 2026-08-02
description: "Why I ported Shiki's TextMate engine to Go without CGO, and the five hardest problems I hit along the way."
tags: [announcement, internals]
---

Nuri (塗り, "painting") is a pure Go port of [Shiki](https://shiki.style), the TextMate grammar-based syntax highlighter used by VS Code: 257 languages, 65+ themes, and hundreds of hierarchical token scopes, with no CGO and no Node.js. Today the docs site and the library are public, so this is the introduction post: what Nuri is, why I built it, and the five problems that turned out to be the hard part.

## Why not Chroma, why not Shiki

Go already has a good syntax highlighter: Chroma. But Chroma plays a different game. It uses Pygments-model lexers with roughly 80 token types, while a TextMate grammar produces hundreds of hierarchical scopes per language: `meta.function.go` inside `source.go`, `punctuation.definition.string.begin` distinct from the string body. VS Code themes are written against those scopes. Feed them Chroma tokens and most of the theme's rules never fire; the output is recognizably "the same colors, fewer of them."

The other option is running real Shiki, which means a Node.js runtime in your build pipeline: a subprocess per render or a long-lived sidecar, plus a non-Go API surface. For a Go static site generator, both costs are structural.

Nuri closes the gap from the Go side: the actual TextMate engine, the actual Oniguruma regex library, VS Code's actual themes, behind an ordinary Go API.

```go
h, err := nuri.New(ctx, nuri.WithFS(core.FS()))
if err != nil {
    log.Fatal(err)
}
defer h.Close(ctx)

html, err := h.CodeToHTML(ctx, `fmt.Println("hello, world")`, nuri.CodeToHTMLOptions{
    Lang:  "go",
    Theme: "github-dark",
})
```

Output:

```html
<pre class="nuri github-dark" style="background-color:#24292e;color:#e1e4e8" tabindex="0"><code><span class="line"><span style="color:#E1E4E8">fmt.</span><span style="color:#79B8FF">Println</span><span style="color:#E1E4E8">(</span><span style="color:#9ECBFF">"hello, world"</span><span style="color:#E1E4E8">)</span></span></code></pre>
```

That was the pitch. Here is what it took.

## Fidelity has to be a number

A port that produces "pretty close" output has no reason to exist, because close-enough highlighting is what the alternatives already offer. So I made fidelity a number instead of a claim. A Node.js generator runs real vscode-textmate (the engine inside VS Code) over sample files for every grammar and records every token's byte offsets, scopes, and resolved colors. Those recordings are committed as golden fixtures, and Nuri's output is diffed against them token by token. Every divergence is classified as a boundary mismatch, scope mismatch, extra token, or missing token, which tells you *which component* has the bug before you open a single file.

A provenance lockfile pins the exact upstream grammar commit, per-file SHA256 hashes, and the Shiki version that generated the fixtures, so "regenerate the goldens" is a deliberate, reviewed act rather than a way to accidentally grade your own homework.

Today, 227 of 234 tested grammars (97%) produce output byte-identical to Shiki, 454 of 468 test cases. The core bundle's 32 fidelity-tested grammars are at 100%: a grammar does not go into `bundle/core` until it is green. (The bundle's other 6 grammars are injection helpers, covered through their parents.)

## The shared engine bought me nothing

I assumed early on that both sides run real Oniguruma, so the regexes match identically, so the output matches. The regexes *do* match identically. The tokenizer wrapped around them is where every single fidelity bug lived. vscode-textmate's state machine is a decade of accumulated, undocumented semantics, and each one I missed showed up as a wrong color somewhere.

Some examples from the sessions that closed those gaps:

- Scope-name backreferences: grammar scope names can embed capture references like `meta.tag.structure.$2.start.html`, resolving `$2` from the match. Miss that and every HTML tag carries a literal `$2` in its scope.
- Capture re-tokenization: a capture group can have its own patterns, which means recursively re-tokenizing a *prefix of the current line* with truncated bounds and preserved absolute offsets. Rewriting Nuri's to mirror vscode-textmate's approach took HTML from 89 diffs to 8 in one commit, most of it CSS inside `<style>` tags.
- `\G` anchor control: TextMate grammars use `\G` to anchor patterns at the previous match's end, and the engine must selectively disable `\A` and `\G` per search. I passed Oniguruma's native option flags, then burned an afternoon on every pattern failing because I had used vscode-textmate's bit positions (9 and 11) where Oniguruma wanted its own (22 and 24). The regex engine was correct; my two constants were not.
- Loop protection, `isFirstLine`, while-condition ordering: four separate not-advancing guards, each restoring different state, each observable in some grammar's output.

Markdown was the boss fight: begin/*while* rules driving per-line continuation, embedded languages in fenced blocks, and at its worst 146 diffs on a single fixture. The lesson I kept re-learning: with a fidelity oracle, "roughly the same algorithm" is not a category that exists. Either the state machine matches or the bytes do not.

## Real Oniguruma without CGO

TextMate grammars depend on Oniguruma-specific behavior: backreferences into begin matches, lookbehind, possessive quantifiers, leftmost-match tie-breaking. Reimplementing that on Go's regexp (RE2, no backreferences at all) was never on the table, and CGO would cost the "go get and it works" property that makes a Go library pleasant to consume.

Nuri embeds Oniguruma itself compiled to WebAssembly, the same ~470KB `onig.wasm` approach Shiki uses, and runs it inside [wazero](https://wazero.io), a pure Go WASM runtime with zero dependencies. Bug-for-bug engine compatibility, no C toolchain anywhere.

What WASM takes in exchange is control. Go cannot kill a running WASM call, so a catastrophically backtracking regex would hang a goroutine forever with no `context.Context` to save you. Nuri's answer is layered: wazero compiles cancellation checkpoints directly into the WASM machine code (`WithRegexInterruption`, on by default), a per-line timeout cancels through them, a max-line-length pre-filter catches minified one-liners before they reach the engine, and a poisoned instance (timeout or panic) is torn down and replaced rather than trusted again. Each degradation is non-fatal: the line renders unstyled and the incident lands in a `Diagnostic` the caller can inspect.

## Concurrency Shiki never needed

Shiki runs on a single-threaded JavaScript runtime. A Go library gets called from every goroutine at once, and WASM instances are not thread-safe, so concurrency had to be designed in, not bolted on.

Nuri compiles the WASM module once and maintains a bounded pool of instances, created lazily on demand. The pool is LIFO, and that is load-bearing: each instance carries its own compiled-scanner cache, so returning the most-recently-used instance means grammars are already compiled where you land. A FIFO variant I tried first rotated sequential callers across cold instances and paid the full grammar-compile cost (~185ms for JavaScript) on nearly every call. And it is deliberately not `sync.Pool`, which evicts under GC pressure; here an eviction means re-instantiating a WASM module and recompiling every cached scanner, and a shutdown path that cannot be drained deterministically.

## Making it fast without changing a byte

With correctness pinned by the fixture suite, performance was the last mountain. The short version: the first numbers were unusable (658ms to highlight one TypeScript snippet), and the causes were almost never what prior analysis claimed they were. The scanner cache existed but lived for one call instead of the process. The per-position loop re-derived work that could be memoized. The FFI boundary allocated where it could pin. And after all of that, the single biggest remaining cost turned out to be the safety instrumentation I had compiled in.

The full story, including the part where my own freshly written diagnosis was wrong and a profiler one-liner caught it, is its own post: [making Nuri 13–42× faster without changing a single output byte](/blog/making-nuri-faster-without-changing-a-byte/). The fixture suite is what made that work safe: every optimization shipped under a gate of "the failure set is byte-identical to baseline."

Today a small Go or Markdown snippet highlights in under 6ms warm (TypeScript, the heaviest common grammar, in ~15ms), and a 300-block site build spends roughly a second on highlighting with default settings. The instance pool's parallelism and disabling interruption checkpoints both cut that further; the performance post has the numbers.

## Current state

- 227/234 grammars byte-identical to Shiki; the core bundle's 32 fidelity-tested grammars at 100%
- Six output formats: HTML, ANSI, SVG, JSON, plain text, raw tokens
- Multi-theme mode, transformers, line decorations, style-to-class mode, WCAG contrast correction
- Two bundles: `bundle/core` (38 languages, ~0.5 MB) and `bundle/full` (all 257 languages, ~1.5 MB)
- MIT licensed, `go get github.com/frostybee/nuri`

If you want to try it, start with [Installation](/docs/getting-started/installation) and the [Quick Start](/docs/getting-started/quick-start). If you want to know how it works, the [internals section](/docs/internals/architecture) covers the architecture, the [Shiki compatibility page](/docs/internals/shiki-compatibility) maps what is ported and what deliberately differs, and the [fidelity methodology](/docs/internals/fidelity-testing) explains how the 97% is measured, including the seven grammars still failing and why.
