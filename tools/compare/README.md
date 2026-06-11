# Nuri vs Shiki vs Chroma: Comparison Benchmark

A local measurement tool that produces a two-axis table (speed + fidelity) comparing Nuri, Shiki, and Chroma. Re-run whenever any engine releases a new version; versioned snapshots track changes over time.

## Setup

```bash
cd tools/compare
npm install        # installs shiki
go mod tidy        # fetches chroma
```

## Usage

```bash
go run . [flags]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-n` | `50` | Warm iterations per input per engine |
| `-save` | `false` | Write `results/` snapshot + regenerate `RESULTS.md` |
| `-skip-shiki` | `false` | Skip Shiki (no Node needed) |
| `-skip-chroma` | `false` | Skip Chroma |
| `-skip-nuri` | `false` | Skip Nuri |
| `-compare FILE` | | Compare against a previous snapshot JSON |
| `-theme` | `github-dark` | Theme name for all engines |

### Examples

```bash
# Quick validation (Go engines only)
go run . -skip-shiki -n 10

# Full three-engine run
go run . -n 10

# Save snapshot + RESULTS.md
go run . -n 50 -save

# Compare against a previous run
go run . -n 50 -save -compare results/2026-06-10T12-00-00/snapshot.json
```

## Output

### Console

Formatted table printed to stdout with cold/warm timing, allocations (Go engines), and token/scope counts.

### Snapshot (`-save`)

Each run creates a timestamped directory under `results/`:

```
results/2026-06-10T12-00-00/
├── snapshot.json          # machine-readable timing + fidelity
├── nuri-go.tokens         # token dump: Nuri highlighting of Go snippet
├── shiki-go.tokens        # token dump: Shiki highlighting of Go snippet
└── chroma-go.tokens       # token dump: Chroma highlighting of Go snippet
```

### Token dump format

One line per token: `{hexColor}{fontStyle}{text}`. Diff any two files to see where highlighting diverges:

```
diff results/latest/nuri-go.tokens results/latest/chroma-go.tokens
```

### RESULTS.md

Regenerated from the latest snapshot. Contains speed tables, cold start comparison, allocation breakdown, and fidelity counts.

## Interpreting results

- **Nuri vs Shiki**: Both use the same Oniguruma engine and TextMate grammars. Speed differences reflect Go vs JS runtime and implementation choices.
- **Nuri (default) vs Nuri (no-interrupt)**: The default configuration enables WASM-level regex interruption for per-line timeouts. Disabling it removes the overhead but loses timeout safety.
- **Chroma**: A completely different architecture (Pygments-model lexers, RE2 regex). Fewer token types by design (~80 vs hundreds). Speed comparisons are not directly equivalent but still useful for SSG build-time decisions.
- **Node startup**: ~50-150ms excluded from Shiki measurements (timing happens inside the process). This is a real cost for CLI tools but irrelevant for SSG long-lived processes.

## Design

- **Separate Go module**: `chroma/v2` dependencies stay out of Nuri's `go.sum`. A `replace` directive points to the local working tree.
- **Local only**: No CI workflow. Runner variability makes timing comparisons unreliable.
- **Hardcoded engines**: No plugin interface. Nuri, Shiki, and Chroma are the only three that matter; can refactor later if needed.
