# wasm-build

Build tooling for `resources/wasm/onig.wasm`, the [Oniguruma](https://github.com/kkos/oniguruma) regex engine compiled to WebAssembly. Nuri embeds this binary and runs it under [wazero](https://wazero.io), which is what gives the tokenizer bug-for-bug regex compatibility with Shiki without requiring CGO.

The built artifact is committed to the repository. You do not need any of this to build, test, or consume Nuri. Use it only when updating the Oniguruma version, changing the C scanner wrapper, or reproducing the binary from source.

## Files

| File | Purpose |
|---|---|
| `Dockerfile` | Pinned Emscripten image, Oniguruma source build, and the `emcc` link step. |
| `build.sh` | One-command wrapper: builds the image, extracts `onig.wasm`, copies it into `resources/wasm/`. |
| `onig_scanner.c` | C wrapper exposing the multi-pattern scanner API that the Go engine calls. |

## Prerequisites

Docker is the only requirement. Emscripten and the Oniguruma sources are pulled inside the container, so no local emsdk or C toolchain is needed.

`build.sh` is a bash script. On Windows, run it from WSL or Git Bash, or use the manual Docker commands below.

## Building

```bash
cd wasm-build
./build.sh
```

The script runs four steps:

```bash
docker build -t nuri-onig-build .
docker create --name nuri-onig-extract nuri-onig-build
docker cp nuri-onig-extract:/build/onig.wasm ../resources/wasm/onig.wasm
docker rm nuri-onig-extract
```

The first build takes several minutes because Oniguruma is configured and compiled from source inside the image. Later builds reuse the Docker layer cache and only redo the `emcc` step if `onig_scanner.c` changed.

The result is roughly 470 KB.

## What the Dockerfile does

| Step | Detail |
|---|---|
| Base image | `emscripten/emsdk:3.1.61` |
| Oniguruma source | Release tarball `v6.9.10` from GitHub releases |
| Configure | `emconfigure ./configure --enable-posix-api=no --disable-shared` |
| Compile | `emmake make`, producing `src/.libs/libonig.a` |
| Link | `emcc` against `onig_scanner.c` and the static archive |

Both versions are pinned deliberately. Changing either changes the SHA256 recorded in the provenance lockfile, so treat a version bump as a reviewed change, not an incidental one.

The `emcc` flags:

| Flag | Why |
|---|---|
| `-O2` | Optimization level. |
| `-s STANDALONE_WASM=1` | Emit a standalone module with no JavaScript glue. wazero has no JS runtime to provide it. |
| `--no-entry` | The module is a library, not a program. There is no `main`. |
| `-s EXPORTED_FUNCTIONS=[...]` | The scanner ABI plus `_malloc`, `_free`, and `__initialize`. See below. |
| `-s ALLOW_MEMORY_GROWTH=1` | Grammar patterns and source text are copied into linear memory and can exceed the initial allocation. |
| `-Ionig-6.9.10/src` | Header path for `oniguruma.h`. |

## Exported ABI

These exports are the contract between the WASM module and `internal/oniguruma/`. Changing a signature here means changing `engine.go`, `instance.go`, and `scanner.go` in lockstep.

| Export | C signature | Behavior |
|---|---|---|
| `onig_scanner_init` | `int onig_scanner_init(void)` | Called once per instance at instantiation. Initializes Oniguruma with UTF-8 encoding and sets the match stack limit to 100000. Returns 0 on success. |
| `create_onig_scanner` | `uintptr_t create_onig_scanner(const char *patterns_buf, const int *lengths, int count)` | Compiles `count` patterns from one concatenated buffer. Returns an opaque handle, or 0 on failure (call `get_last_onig_error` for the reason). |
| `find_next_match` | `int find_next_match(uintptr_t scanner, const char *str, int str_len, int start_pos, int *result_buf, int result_buf_size, int options)` | Searches every pattern and keeps the leftmost match, with ties broken by lowest pattern index. Writes `[patternIndex, numRegs, beg0, end0, beg1, end1, ...]` into `result_buf`. Returns `numRegs`, `-1` for no match, `-2` for an allocation or buffer failure. |
| `free_onig_scanner` | `void free_onig_scanner(uintptr_t scanner)` | Frees the compiled patterns and the handle. |
| `get_last_onig_error` | `int get_last_onig_error(char *buf, int buf_size)` | Copies the last Oniguruma error message (up to 255 bytes) into `buf` and returns its length. This is what surfaces invalid regex syntax in a grammar as a real message instead of an opaque failure. |
| `malloc` / `free` | Emscripten runtime | The Go side manages linear-memory buffers through these. Pattern and length buffers are allocated and freed around each `create_onig_scanner` call (`instance.go`); the source-text and result buffers are kept and grown across calls (`scanner.go`). |
| `__initialize` | Emscripten runtime | The standalone reactor initializer. Nothing in `internal/oniguruma/` looks it up or calls it, so it is exported for the runtime's benefit at instantiation rather than for the Go code. |

The leftmost-wins tie-break in `find_next_match` (`onig_scanner.c`) is tokenization semantics, not an optimization detail. Which rule wins when several patterns match at the same position determines the resulting token stream, so this comparison is not safe to "clean up" without regenerating and re-checking the fidelity fixtures.

### Imports

The module imports one symbol, `env.emscripten_notify_memory_growth`. The Go engine registers an empty host stub for it (`internal/oniguruma/engine.go`) because wazero handles memory growth natively. If a change to the `emcc` flags introduces any additional import, a matching stub must be added there or module instantiation fails outright.

## After rebuilding

1. Run the engine and tokenizer tests: `go test ./internal/oniguruma/... ./internal/tokenizer/...`
2. Run the fidelity gate: `go test ./internal/fidelity/... -run TestGoldenFidelity`
3. Regenerate the provenance lockfile: `go run ./tools/devtool lock`
4. Commit `resources/wasm/onig.wasm` and `provenance.lock.json` in the same change.

Step 3 is not optional. `provenance.lock.json` pins the SHA256 of `onig.wasm`, and `go run ./tools/devtool verify` (which CI runs) fails when the committed hash does not match the binary on disk.

## Bumping the Oniguruma version

Edit the tarball URL in the `RUN curl` line, then update the two `onig-6.9.10` path references in the same Dockerfile (the `cd` in the build step and the `-I` include path in the `emcc` step). Rebuild, then follow the steps above.

The Dockerfile is the only place the Oniguruma version is recorded. `provenance.lock.json` pins the SHA256 of the built binary but not the version that produced it, and the `ONIGURUMA (WASM BINARY)` block in `THIRD-PARTY-NOTICE` (generated by `tools/devtool/notices.go`) records the file path, source URL, license, and copyright without a version. Nothing will flag a mismatch between the Dockerfile and a binary built from different sources, so keep the two committed together.

Because Oniguruma is the regex engine, a version bump can shift tokenization for grammars that rely on edge-case regex behavior. Run the full fidelity matrix, not just the core one, before committing:

```bash
go run ./tools/devtool generate -- --config matrix.full.config.json
go test ./internal/fidelity/... -run TestGoldenFidelityFull -count=1
```

Fidelity is measured against vscode-textmate, which reaches Oniguruma through the `vscode-oniguruma` npm package (pinned at 1.7.0 in `vscode-textmate/`). A version bump here moves Nuri's engine independently of the reference implementation's, so drift that is not a bug in Nuri can still surface as a failing golden test.
