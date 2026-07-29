---
title: Architecture
description: "The three-layer stack, WASM pool, grammar registry, and asset filesystem."
sidebar:
  order: 1
---

This page explains how Nuri is structured internally: the WASM engine, the instance pool, the grammar registry, and the asset filesystem. It is intended for contributors, downstream library authors, and anyone curious about the design tradeoffs.

Nuri mirrors Shiki's layered architecture but replaces Node.js and native Oniguruma with a pure Go stack built on wazero.

## Three-layer stack

```
nuri (public API)
├── internal/registry    (grammar + theme resolution)
├── internal/tokenizer   (TextMate state machine)
└── internal/oniguruma   (WASM engine + pool)
      └── onig.wasm      (Oniguruma compiled to WebAssembly)
```

The public `nuri` package exposes `Highlighter`, `New`, and the `CodeTo*` methods. Everything below it is `internal/`. The `theme/` package is the one exception: it is exported for standalone theme parsing.

Every highlighting call follows the same path: borrow a WASM instance from the pool, run the tokenizer (which calls Oniguruma through the instance), resolve token styles against the theme, build the output, return the instance.

## WASM engine

The engine implementation lives in `internal/oniguruma/engine.go`.

`NewEngine` performs the expensive one-time work: AOT-compiling the embedded `onig.wasm` binary (~470KB) into a native module via wazero. The compiled module is stored in `Engine.compiled`. Creating instances from this compiled module is cheap (~5ms). The raw WASM binary is available as `wasm.OnigWasm` (from `github.com/frostybee/nuri/resources/wasm`) for consumers who need direct access to the embedded artifact.

The wazero runtime config auto-selects the compiler backend on amd64 and arm64, falling back to the interpreter on unsupported platforms. This happens inside `wazero.NewRuntimeConfig()` with no configuration needed.

Before compiling, the engine registers a host module stub for `emscripten_notify_memory_growth`. The Emscripten-compiled `onig.wasm` imports this symbol; the stub body is empty because wazero handles memory growth natively.

Two optional behaviors are configured at engine creation:

- **Context interruption** (`WithCloseOnContextDone`): When enabled (the default), wazero compiles interrupt checkpoints into the WASM JIT. A cancelled context can stop a runaway regex mid-search. The cost is roughly 3x lower throughput.
- **Compilation cache** (`WithCompilationCacheDir`): Persists the JIT-compiled artifact to disk via `wazero.NewCompilationCacheWithDir`. Subsequent process starts skip AOT compilation entirely.

## Instance pool

The pool implementation lives in `internal/oniguruma/pool.go`.

Each WASM instance is an independent instantiation of the compiled module, with its own linear memory and Oniguruma state. Instances are not thread-safe. The pool manages bounded concurrent access.

### LIFO reuse

The pool uses a LIFO (last-in, first-out) idle stack, not FIFO. Each instance keeps a per-instance cache of compiled regex scanners. Compiling a grammar's regex patterns into an Oniguruma scanner costs roughly 185ms for a complex grammar like JavaScript. LIFO reuse means sequential callers always get the most-recently-used (warmest) instance, which already has compiled scanners for recently highlighted grammars. A FIFO rotation would spread calls across all instances, each of which sees each grammar infrequently and pays the full compile cost.

Under concurrent workloads with W simultaneous borrowers, exactly W instances stay warm. Additional pool capacity is never touched until concurrency grows.

### Why not `sync.Pool`

`sync.Pool` evicts entries under GC pressure. Re-creating an evicted WASM instance requires re-instantiating the module and re-compiling all cached scanners, which is expensive. `sync.Pool` also cannot be drained deterministically at shutdown, preventing clean resource cleanup.

### Semaphore channel

Bounded concurrency uses a buffered channel filled with `size` tokens at construction:

```go
sem: make(chan struct{}, size)
```

`Get` acquires a token via `select` (context-aware). `Put` pushes the instance onto the idle stack *before* releasing the token back to `sem`. This ordering guarantees that a goroutine woken by the token always finds a non-empty idle stack and gets the warm instance rather than creating a cold one.

### Lazy creation

`NewPool` does zero WASM work. Instances are created on demand inside `Get` when the idle stack is empty. A sequential workload creates exactly one instance. Memory scales with instances actually created multiplied by distinct grammars used.

### Poison and swap

When a timeout or panic occurs during tokenization, the WASM instance may be in a corrupted state. The `Do` method wraps the borrow-use-return cycle with `defer`+`recover()`. A panic marks the instance as `poisoned = true` and routes through `Swap` instead of `Put`.

`Swap` creates a fresh instance, removes the poisoned one from the `all` tracking slice, pushes the fresh instance onto the idle stack, and releases the semaphore token. If fresh instance creation fails, the slot is still released. The next `Get` creates a new instance lazily. Pool capacity is never permanently lost.

## Scanner cache

The scanner cache implementation lives in `internal/oniguruma/instance.go`.

Each instance maintains a `map[uint64][]cachedScanner` keyed by FNV-64a hash of the concatenated pattern bytes. Collisions are handled with separate chaining: each bucket holds a slice of entries, and a full byte comparison (`joined` + `lens` fields) verifies the match before returning a cached scanner.

The cache is unbounded by design, matching Shiki and vscode-textmate behavior. The number of unique scanner pattern sets is bounded by the grammar set (a finite number of rule contexts). The cache is destroyed when the instance is closed or poison-swapped.

When a cache miss occurs, `GetOrCreateScannerCtx` allocates pattern and length buffers in WASM memory, calls `create_onig_scanner`, frees the temporary allocations, and returns a `Scanner` wrapping the WASM scanner pointer. If `create_onig_scanner` returns a null pointer, `get_last_onig_error` retrieves Oniguruma's own error message (e.g. invalid regex syntax) for actionable diagnostics.

## Text upload optimization

The scanner implementation lives in `internal/oniguruma/scanner.go`.

Every `FindNextMatch` call needs the source text available in WASM linear memory. Within a single line's tokenization pass, the tokenizer calls `FindNextMatch` multiple times: once for the main scan, then for capture-group sub-ranges. These sub-ranges share the same Go-side backing array.

A pin check avoids redundant WASM memory writes:

```go
pinned := len(inst.curText) > 0 &&
    unsafe.SliceData(text) == unsafe.SliceData(inst.curText) &&
    len(text) <= len(inst.curText)
```

When pinned, the text is already in WASM memory and the write is skipped. When not pinned, the text buffer grows exponentially (doubling, minimum 4096 bytes) via WASM `malloc`, and the bytes are written into WASM memory.

A pre-allocated result buffer (130 uint32 slots: 1 pattern index, 1 match count, 64 capture pairs) is reused across all calls within an instance. `CallWithStack` is used throughout to avoid allocating a `[]uint64` argument slice on every WASM call.

## Grammar registry

The registry implementation lives in `internal/registry/`.

`Registry` is the unified resolver wrapping a `Repository` (grammar loading and metadata) and a `theme.Store` (theme parsing and caching). It adds an alias resolution layer: before any grammar lookup, the `aliases` map substitutes canonical names (e.g. `"js"` becomes `"javascript"`).

### Index file fast path

Bundle packages ship a generated `index.json.gz` in their `grammars/` directory, presented as `index.json` by the asset filesystem. This file contains metadata for every grammar (`GrammarMeta`: scope name, file types, injection targets, first-line match pattern). When the index loads successfully, the directory scan is entirely skipped. Constructing a `Repository` over the full 257-grammar bundle reads exactly one file.

If no index is present (e.g. a consumer-supplied `os.DirFS`), `buildIndexes` falls back to a full directory scan, reading a small metadata probe from each `.json` file without fully parsing the grammar.

### Lazy grammar loading

Individual grammars are parsed on first access via `Repository.Get` and cached in memory for the process lifetime. The double-checked RWMutex pattern allows concurrent reads while serializing cache misses.

### Injection resolution

The injection index maps target scope names to lists of grammar names that inject into them. This is populated at index-build time (not at grammar-parse time). When the tokenizer encounters a scope with registered injectors, it calls `GetInjectors` to retrieve and lazily load the injecting grammars.

`Registry` satisfies `grammar.GrammarResolver` via `GetGrammarByScope`, enabling the tokenizer to resolve cross-grammar includes (e.g. Markdown loading an embedded JavaScript grammar by scope name `source.js`).

## Asset filesystem

The asset filesystem implementation lives in `internal/assetfs/assetfs.go`.

Bundle packages embed grammar and theme JSON as `.json.gz` files to reduce binary size. The `assetFS` wrapper provides transparent gzip decompression, presenting each `.json.gz` file at its virtual `.json` path. All callers above (registry, theme store) use plain JSON paths with no awareness of compression.

The wrapper implements `fs.FS`, `fs.ReadFileFS`, `fs.ReadDirFS`, and `fs.SubFS`. `Open` and `ReadFile` try the plain name on the underlying `embed.FS` first, then fall back to `name + ".gz"` with in-memory decompression. `ReadDir` strips `.gz` suffixes, deduplicates against plain-file twins, and re-sorts.

Decompression is eager (full file into memory) and per-call, with no caching inside `assetFS`. This is intentional: callers cache parsed results at a higher layer, so raw bytes are read at most once per grammar or theme. Adding a byte-level cache would be redundant.

## Design decisions

**No CGO.** Wazero provides a pure Go WASM runtime with zero external dependencies. The `onig.wasm` binary is the real Oniguruma C library, giving bug-for-bug regex compatibility with Shiki without requiring a C toolchain on the consumer's machine. A `//go:build onig_cgo` escape hatch exists for native throughput in environments where CGO is available.

**Compile-once + pool.** AOT compilation is expensive but happens once. Instance creation from the compiled module is cheap. The pool amortizes the compilation cost and provides bounded concurrent access to non-thread-safe WASM instances.

**LIFO over FIFO.** Scanner cache locality is the deciding factor. A FIFO pool would rotate sequential callers across cold instances, paying ~185ms per grammar compile miss. LIFO ensures the warmest instance is always reused first.

**Bounded pool over `sync.Pool`.** Deterministic resource management, predictable memory usage, and clean shutdown. `sync.Pool` eviction under GC would force expensive re-instantiation.

**Lazy loading.** Construction reads one metadata index file. Grammars are parsed on demand. A highlighter that processes only Go code never parses the JavaScript grammar, even if both are in the bundle.

**Eager decompression, no byte cache.** The simplest correct approach. Callers cache the parsed result (grammar struct, theme struct), so raw bytes are decompressed at most once per grammar or theme per process lifetime.
