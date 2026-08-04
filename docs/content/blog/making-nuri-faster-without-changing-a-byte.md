---
title: "Making a WASM-Powered Syntax Highlighter 13–42× Faster Without Changing a Single Output Byte"
date: 2026-08-01
description: "The story of taking Nuri's CodeToHTML from 76ms to 5.8ms while a golden-fixture suite guaranteed zero output changes, ending with the profiler catching my own diagnosis being wrong."
tags: [performance, internals]
---

*Nuri is a pure Go port of [Shiki](https://shiki.style), the TextMate grammar syntax highlighter
used by VS Code. This post is the story of taking its `CodeToHTML` from 76ms to 5.8ms for a small
Go snippet, and from 658ms to 15.5ms for TypeScript, while a golden-fixture suite guaranteed that
not one output byte changed. It ends with an epilogue in which my own freshly written diagnosis
turned out to be wrong, proving the post's first lesson twice.*

*New to Nuri? The [introduction post](/blog/introducing-nuri/) covers what it is, why it exists,
and the other four problems that made the port hard.*

---

## Why Nuri was slow on purpose

Nuri's whole value proposition is *fidelity*. TextMate grammars are riddled with Oniguruma-specific
regex behavior: backreferences into begin matches, `\G` anchors, possessive quantifiers, subtle
leftmost-match tie-breaking. The only way to match VS Code's output bug for bug is to use the
real Oniguruma. So Nuri embeds `onig.wasm` (the same ~470KB binary Shiki uses, Oniguruma compiled to
WebAssembly) and runs it inside [wazero](https://wazero.io), a pure Go WASM runtime. No CGO, no
native dependencies, byte-identical behavior to Shiki.

The price showed up in the benchmarks:

| Input | Time per `CodeToHTML` | Allocations |
|---|---|---|
| 170-byte Go snippet | 75.7ms | 12,407 |
| Small TypeScript snippet | 658ms | 19,402 |
| 50KB Go file | 2.84s | 3.3M (266MB!) |

For Nuri's primary consumer, a static site generator with ~300 code blocks per build, that's 10–15
seconds of highlighting alone. Unusable.

## Distrust the existing diagnosis

Two prior analysis documents predated this work, and verifying them against the source came before touching any code.
**Both misattributed the cost.**

One blamed FFI overhead, claiming each Go→WASM call cost ~0.15ms, so ~500 calls per block explained
the 72ms. Plausible, since FFI boundaries are a classic villain. But wazero in compiled mode
dispatches a call in tens of *nanoseconds*. Five hundred calls is microseconds, three orders of
magnitude short of explaining anything.

The other claimed scanners were being recompiled at every byte position. Half right: a scanner cache
existed and prevented recompiles *within* one tokenize call, but the cache itself was created and
destroyed *per call*.

After that, the rule was to read the code every diagnosis pointed to and measure before touching anything:

```
BenchmarkScannerCreate/1_pattern      6.3µs
BenchmarkScannerCreate/5_patterns     168µs
BenchmarkScannerCreate/50_patterns    442µs
BenchmarkScannerCreate/200_patterns   1.79ms   ← the smoking gun
BenchmarkFindNextMatch (per call)     5.1µs, 27 allocs
```

Compiling a 200-pattern scanner costs **1.79 milliseconds**. The TypeScript grammar's root context
alone has hundreds of enormous patterns, and a TextMate tokenizer needs a compiled scanner for
*every distinct rule context* it enters. Recompile all of those for every code block and you've
found your 658ms. Regex *compilation*, not regex *execution*, not FFI.

## Cache lifetime is a design axis (75ms → 6.7ms)

The scanner cache wasn't missing; it was scoped wrong. It lived for one `Tokenize()` call:

```go
cache := newScannerCache()
defer cache.closeAll()   // every compiled regex, discarded per code block
```

The fix moves it onto the WASM instance itself, which lives in a pool for the lifetime of the process.
Highlight a thousand Go files and the Go grammar's contexts compile exactly once per pooled
instance. This mirrors what vscode-textmate does: compiled rules live as long as the grammar
registry, and are never evicted.

Hash collisions get verified, not assumed away. The cache key is a 64-bit FNV-1a hash of the
pattern set. A process-lifetime cache shared across all grammars multiplies collision exposure, and
a silent collision wouldn't crash; it would *quietly return the wrong scanner* and corrupt
highlighting. So each bucket entry stores the exact pattern bytes plus per-pattern lengths, and
lookups memcmp on hit. The lengths matter: `["ab","c"]` and `["a","bc"]` join to identical bytes.
One equality check per lookup buys "collisions are impossible" instead of "collisions are unlikely."
A test forces a collision by overriding the hash function with `func(...) uint64 { return 42 }` and
proves both entries still resolve correctly.

Because ownership moved, lifecycle had to move too. Cached scanners are owned by the instance, and
callers must never close them. When a pathological regex poisons an instance (timeout or panic), the
pool tears down the whole WASM module and swaps in a fresh one; the cache dies with it, no cleanup
code needed. Unbounded growth is by design and documented; Shiki never evicts either.

Result: Go 75.7ms → 6.7ms, **TypeScript 658ms → 18ms**, JavaScript 550ms → 34ms. One change, 11–36×.

## Stop redoing per-block work per position (3× memory)

With compilation amortized, the per-position loop was next. At *every scan position*, the tokenizer:

- re-flattened the active rule set (recursive include resolution with a fresh `visited` map and
  full repository-map copies),
- re-built the pattern list and re-hashed every pattern byte only to *look up* the cached scanner,
- re-did all of it again for every matching injection grammar.

That's why a 170-byte snippet allocated 12,000 times.

The fix is a per-call memo: one map lookup per position instead of re-flatten and re-hash. The
cache key is where it gets tricky:

```go
type memoKey struct {
    rule       grammar.Rule     // identity of the current context; nil = root
    compileG   *grammar.Grammar // which grammar resolves includes (cross-grammar embeds!)
    hasEnd     bool             // ← can't be folded into endPattern
    endPattern string           // end patterns embed backrefs resolved from begin captures
    applyLast  bool             // applyEndPatternLast changes pattern ordering
}
```

Every field exists because of a real near miss:

- `hasEnd` is separate because backref resolution can legitimately produce an *empty* end pattern,
  which must not collide with "this context has no end rule."
- The end rule is inserted into the cached slice **once, as a copy**. The old code appended it per
  position; against a now-shared cached backing array, that would silently shift the index→rule
  mapping. Caching turns a harmless `append` into a corruption bug. Lifetimes change what's safe.
- Capture re-tokenization built a fresh root rule object per call, which would have made the memo
  miss on the hottest recursive paths. A side map interns one stable root per capture rule.

Compile *errors* are memoized too, because "identical output" includes identical degradation.

Result on top of Fix 1: allocations roughly halved again, and bytes per op dropped 3× (the 50KB Go
file went from 266MB allocated to 59MB).

## Make the boundary boring (27 allocs/match → 5)

Each match call still did the full FFI dance: allocate WASM memory for the text, copy the text in,
allocate a result buffer, call, read, free both. Five boundary calls and a full text copy per match,
~µs each. Small, but multiplied by every match of every line.

Three changes, all enabled by one fact verified against wazero's source: **`memory.grow` never
relocates existing allocations**, so WASM-side pointers are stable.

1. A persistent text buffer with pinning. The instance keeps the WASM-side buffer plus a Go-side
   reference to the last-uploaded slice. The upload is skipped when the incoming text is a prefix of
   what's already there:

   ```go
   pinned := len(inst.curText) > 0 &&
       unsafe.SliceData(text) == unsafe.SliceData(inst.curText) &&
       len(text) <= len(inst.curText)
   ```

   Why is comparing pointers sound? Because `curText` *is* the pin: while the instance holds that
   live reference, the Go allocator cannot reuse the backing array's address for a new object. Same
   address genuinely means same memory. And the prefix rule isn't merely an optimization for
   repeated calls: TextMate capture re-tokenization recursively scans *prefixes of the current
   line*, so every level of that recursion now uploads nothing at all.

2. A one-time result buffer (520 bytes, allocated once per instance) instead of alloc/free per call.

3. `CallWithStack` with a reused `[8]uint64` instead of the allocating `Call` API.

Per-call cost: 5.1µs/27 allocs/1120B → **3.9µs/5 allocs/224B**. I also deliberately did *not* pool
the `Match` objects: the tokenizer mutates capture slices and holds two candidate matches across a
recursive call, so reuse would be a correctness bug. Knowing when to stop is part of the fix.

## The cheap stuff you find along the way

- Capture-text extraction ran per match even when nothing could consume it. The consumers are
  no-ops unless the pattern has a `\N` backref or the scope name has `$`, both knowable at *parse
  time*, so a per-rule flag now gates it. Skipping work is output-identical when the work was a no-op.
- `splitLines` copied every line of every file (`bytes.Split` plus a per-line `append('\n')`).
  Manual index slicing returns views: zero copies.
- And an honest-to-goodness bug: the multi-theme path called `splitLines` on the *entire source*
  inside the per-line loop, which is O(lines²), and then used the result only as `_ = line`. Deleted.

## The bottleneck I compiled in

After all of that, small-Go sat at ~6ms. Good, but the plan had a contingency: if the WASM boundary
still dominated, rebuild the C wrapper layer. Before scoping that, one long-carried config flag
deserved a benchmark first: wazero's `WithCloseOnContextDone(true)`, which compiles context-cancellation
checkpoints into all WASM execution so a runaway regex can be interrupted mid-search.

| Config | Go | TypeScript |
|---|---|---|
| Interruption on (default) | 5.6ms | 16.0ms |
| Interruption off | **1.6ms** | **3.9ms** |

**3.5–4.1×.** The single biggest remaining cost wasn't FFI, wasn't Nuri's Go code, wasn't even
Oniguruma. It was the safety instrumentation compiled into the WASM module. The planned
WASM-rewrite contingency died right there: it would have attacked a boundary that was already cheap.

I kept the default **on**, because safety defaults should survive performance work, and exposed
`WithRegexInterruption(false)` for trusted-input builds (an SSG highlighting its own repository
doesn't need mid-regex interruption; a Go-side per-line timeout and Oniguruma's internal backtrack
limit still apply). The decision I *didn't* make is as documented as the ones I did: flipping the
default now has recorded evidence waiting for it.

## Proving "zero output change"

Every optimization above was made under one gate: Nuri's golden-fixture suite, where expected output
is generated by running real Shiki and committed. But the baseline wasn't 100% green. A
few grammars had known, pre-existing fidelity gaps (436/468 passing on the full matrix at the
time; 454/468 today, with 227 of 234 grammars byte-identical).

So "all tests pass" was useless as a gate. The gate became **"the failure set is byte-identical to
baseline"**: capture every token-diff line at baseline, sort, and `diff` against the same extraction
after each step. 348 diff lines before, the same 348 after: same passes, same failures, same bytes.
A fuzzer (`FuzzTokenize`, arbitrary UTF-8, no-panic and token-coverage invariants) backstopped the
cases no fixture covers.

An imperfect-but-frozen oracle is still a great regression oracle. You don't need green tests to
refactor safely; you need *stable, comparable* output.

## War stories (Windows edition)

- `go test -cpuprofile` crashed the Go runtime: `runtime.sigprof` faults while unwinding a thread
  executing wazero's JIT-compiled code on Windows. Deterministic, within a second. The fallback was
  allocation profiles (no signal unwinding) plus benchmark deltas. Allocation profiles turned out to
  be the better storyteller anyway: they're what flagged both the per-call watchdog goroutines and
  the next target.
- `-race` needs CGO, and the machine had no C toolchain, which is exactly the kind of machine a
  no-CGO library gets developed on. Race coverage moved to Linux CI.
- The first iteration of every post-fix benchmark is now ~3× slower than steady state: that's the
  persistent cache warming up. Worth knowing which number you're quoting. For an SSG doing 300
  blocks, steady state is the honest one.

## Final numbers

| Input | Before | After | Speedup | allocs/op |
|---|---|---|---|---|
| Go snippet | 75.7ms | 5.8ms | 13× | 12,407 → 1,941 (6.4×) |
| TypeScript snippet | 658ms | 15.5ms | 42× | 19,402 → 3,126 (6.2×) |
| JavaScript snippet | 550ms | 29.6ms | 19× | 31,213 → 3,624 (8.6×) |
| HTML snippet | 349ms | 18.8ms | 19× | 18,197 → 2,776 (6.6×) |
| Markdown snippet | 129ms | 5.3ms | 24× | 5,447 → 2,152 (2.5×) |
| 50KB Go file | 2.84s | 2.32s | 1.2× | 3.33M → 241k (14×); 266MB → 20.3MB |

(With interruption disabled, the Go snippet is 1.6ms, a 47× total.) The 300-block SSG build:
from 10–15 seconds of highlighting to roughly one, and well under that with the instance pool's
parallelism or interruption disabled.

One row moved the wrong way since this work first landed.
The Markdown number was 3.0ms when the optimization pass finished. The day after, a correctness
fix aligned Nuri's begin/while end-of-line semantics with vscode-textmate's, which made
begin/while contexts (lists, blockquotes) do the full nested tokenization they were previously
skipping. Markdown got slower because it started doing *more correct work*, and nine grammars
(markdown among them, plus make, org, rst, and others) flipped from failing to byte-identical.
When fidelity and speed conflict, fidelity wins,
and the benchmark table takes the hit in public.

## My own diagnosis was wrong, again

The first draft of this post ended by naming the next bottleneck: with everything above fixed, 71%
of remaining allocation objects came from one function, `strings.Fields`, which I confidently
attributed to scope-name splitting in the tokenizer's stack. I'd seen it at the top of
`pprof -top`, the tokenizer splits scope names, case closed.

Then I went to fix it and ran one more command first:

```
$ go tool pprof -peek strings.Fields mem.out
                       1960651   100% |  theme.scoreSelector
```

**One hundred percent of the calls came from the theme matcher**, not the tokenizer. The function
I'd blamed had a no-space fast path and accounted for 2.4% of objects. The real culprit:
`Theme.Match` runs once per token, loops ~100 theme rules, and re-split each rule's *static*
selector string ("source.go keyword") with `strings.Fields` on every single call. Thousands of
identical splits of strings that never change after the theme is parsed.

The fix is the same shape as Fix 1: move work to the right lifetime. Selectors now pre-compile
once at theme parse time (split parts and specificity depth cached on each rule, with a per-selector
source-string guard so hand-built or post-parse-mutated themes still work via the old path, and no
mutation during `Match` so concurrent readers stay race-free). `Theme.Match`: 5.9µs and 74 allocs →
**1.5µs and zero allocs**. That's where the final numbers in the table above come from:
allocations across the pipeline dropped another 3–6× on top of everything else, and
`strings.Fields` vanished from the profile entirely.

I made the exact mistake the top of this post
warns about, *while writing the post*. `pprof -top` tells you **what** allocates; only the call
graph (`-peek`, `-traces`) tells you **who's responsible**. Flat percentages plus a plausible story
is how misattributions are born. A misattribution written into a planning doc would have sent
the next optimization pass to the wrong file. One command would have caught it. Eventually, one
command did.

You don't run out of bottlenecks; you keep promoting the next one. Verify each one as if it
were the first.

## The checklist version

1. Verify the diagnosis against the code before believing it. Both prior analyses named
   plausible villains; both were wrong. Micro-benchmark the suspected cost: "1.79ms per compile"
   ended the debate instantly.
2. Cache lifetime is a design decision, not an implementation detail. The cache existed; it was
   scoped to the wrong lifetime. Promoting it was worth 11–36×, more than everything else combined.
3. When you extend a cache's lifetime, re-audit everything it touches. Hash collisions go from
   unlikely per call to compounding per process (verify bytes). A shared backing array turns a safe
   `append` into index corruption (copy at fill).
4. Make memo keys earn every field with a counterexample. Empty-after-resolution end patterns,
   pattern-ordering flags, unstable temporary objects: each one was a would-be cache-poisoning bug.
5. Pin, don't copy, and write down why the pin is sound. A live Go reference makes
   same-address-means-same-data true. One pointer comparison eliminated a text upload per match and
   made recursive prefix scans free.
6. Benchmark your safety instrumentation. Nuri's was 70% of the remaining runtime. Keep the safe
   default, expose the informed opt-out, record the evidence.
7. Freeze an oracle, even an imperfect one. Byte-diffing 348 known failure lines against
   baseline is what let every step ship the same day it was written.
8. `pprof -top` names the allocator; `-peek` names the culprit. A flat percentage plus a
   plausible story is how misattributions are born; I wrote one into this very post and nearly
   optimized the wrong file. Check the call graph before you believe your own diagnosis.

---

*Benchmarks: Intel i9-10850K, Windows 10, Go 1.26, wazero v1.12.0, single pooled WASM instance,
default options (regex interruption and WCAG contrast correction on except where noted). "Before"
numbers are the 2026-06-10 baseline; "After" numbers were re-measured 2026-08-02 on current main.
All numbers are steady-state medians of 5 runs of `BenchmarkCodeToHTML` with `-benchmem`.*
