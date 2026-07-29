---
title: Performance & Concurrency
description: "Pool sizing, compilation cache, per-line safety valves, and regex interruption."
sidebar:
  order: 10
---

Nuri runs Oniguruma compiled to WASM. Performance depends on the bounded instance pool, per-instance scanner caching, and configurable safety limits for pathological input.

## Instance pool

Nuri maintains a bounded pool of WASM instances. Each highlighting call borrows an instance, tokenizes, and returns it. The pool is concurrency-safe: multiple goroutines can highlight in parallel up to the pool size.

```go
h, err := nuri.New(ctx,
	nuri.WithFS(core.FS()),
	nuri.WithPoolSize(4),
)
```

Key behaviors:

- **LIFO reuse.** The pool returns the most-recently-used instance. Each instance keeps a cache of compiled regex scanners, so LIFO reuse favors the warmest instance. A FIFO rotation would spread calls across cold instances and pay the full grammar-compile cost on each.
- **Lazy creation.** Instances are created on first demand, not pre-allocated. A sequential workload reuses one warm instance. A workload with W concurrent tokenizations creates and warms W instances.
- **Blocking on exhaustion.** When all instances are checked out, the next call blocks until one is returned. The block respects context cancellation.
- **Memory.** Scales with instances actually created multiplied by distinct grammars used. Setting a pool size above expected concurrency costs nothing until that concurrency materializes.

The default pool size is `runtime.NumCPU()`. For a web server handling concurrent requests, match the pool size to expected parallel highlighting calls. For a CLI that highlights one file at a time, the default of 1 active instance is sufficient.

## Scanner cache

Each WASM instance keeps an unbounded cache of compiled regex scanners, keyed by the concatenated pattern bytes. The cache lives for the instance's lifetime and is never evicted (matching Shiki's behavior). This is why LIFO matters: the most-recently-used instance already has compiled scanners for recently highlighted grammars, avoiding a recompile that can cost ~185ms for complex grammars like JavaScript.

## Compilation cache

By default, `nuri.New()` AOT-compiles the embedded `onig.wasm` module on every process start. For short-lived processes (CLI tools, SSG builds), enable an on-disk compilation cache to skip this step on subsequent runs:

```go
h, err := nuri.New(ctx,
	nuri.WithFS(core.FS()),
	nuri.WithCompilationCacheDir("/tmp/nuri-cache"),
)
```

The directory is created if it does not exist. It may be shared between processes and only grows when `onig.wasm` itself changes. The cached artifact is ~470KB.

SSG pipelines that highlight hundreds of files per build benefit from `nuri.WithCompilationCacheDir` to eliminate cold starts across invocations.

## Per-line safety valves

Two options protect against pathological input: lines that are too long for practical tokenization and lines that trigger catastrophic regex backtracking. Both are non-fatal. Tokenization continues for subsequent lines, and a `Diagnostic` is recorded.

### Maximum line length

```go
h, err := nuri.New(ctx,
	nuri.WithFS(core.FS()),
	nuri.WithMaxLineLength(10000),
)
```

Lines exceeding this byte count are emitted as a single unstyled token with a `"too_long"` diagnostic. The check runs before any WASM call, so it is effectively free. Set this for pipelines that may encounter minified CSS or JavaScript.

### Per-line timeout

```go
h, err := nuri.New(ctx,
	nuri.WithFS(core.FS()),
	nuri.WithTimeoutMs(500),
)
```

Lines whose tokenization exceeds this duration are stopped early. Partial tokens are preserved and a `"timeout"` diagnostic is recorded. The timed-out WASM instance is replaced with a fresh one; subsequent lines tokenize normally.

`nuri.WithTimeoutMs` is a no-op when built with `//go:build onig_cgo` (native Oniguruma calls cannot be interrupted).

### Per-call overrides

Both options accept per-call overrides through pointer fields on all options structs. A nil pointer uses the highlighter default; a non-nil pointer overrides it:

```go
maxLine := 5000
html, err := h.CodeToHTML(ctx, code, nuri.CodeToHTMLOptions{
	Lang:          "css",
	Theme:         "github-dark",
	MaxLineLength: &maxLine,
})
```

## Regex interruption

```go
h, err := nuri.New(ctx,
	nuri.WithFS(core.FS()),
	nuri.WithRegexInterruption(false),
)
```

When enabled (the default), wazero compiles interrupt checkpoints into the WASM JIT. A runaway regex match can be stopped mid-search when the caller's context is cancelled. The cost is roughly 3x lower throughput compared to disabled.

When disabled, runaway patterns are bounded only by the Go-side per-line timeout (checked between scan positions) and Oniguruma's internal match stack limit.

The default is ON. Safety over speed. Disable it for batch processing with trusted input where throughput matters and context-based cancellation is not needed.

## Options summary

| Option | Type | Default | Description |
|---|---|---|---|
| `nuri.WithPoolSize(n)` | `int` | `runtime.NumCPU()` | Maximum WASM instances in the pool |
| `nuri.WithMaxLineLength(n)` | `int` | `0` (no limit) | Byte-length pre-filter per line |
| `nuri.WithTimeoutMs(ms)` | `int` | `0` (no timeout) | Per-line soft timeout in milliseconds |
| `nuri.WithCompilationCacheDir(dir)` | `string` | `""` (no cache) | On-disk AOT compilation cache directory |
| `nuri.WithRegexInterruption(enabled)` | `bool` | `true` | WASM-level interrupt checkpoints |

## Next steps

- [Highlighter API](/docs/reference/highlighter-api) for the full method reference
- [Configuration](/docs/reference/configuration) for all constructor options
- [Architecture](/docs/internals/architecture) for the WASM pool and engine internals
