# tools

Development scripts and tooling for Nuri. None of these are part of the module or linked into any binary.

| Tool | What it is |
|---|---|
| `devtool/` | Go CLI for syncing grammar/theme assets, managing the provenance lockfile, and generating fidelity fixtures. See [`devtool/README.md`](devtool/README.md). |
| `genfixtures/` | Node.js fixture generator that drives real `vscode-textmate` and writes golden JSON for the fidelity test suite. See [`genfixtures/README.md`](genfixtures/README.md). |
| `compare/` | Benchmark tool that produces a speed and fidelity table comparing Nuri, Shiki, and Chroma. See [`compare/README.md`](compare/README.md). |
| `race-wsl.sh` | Bash script that runs the Go race detector inside WSL. See below. |

## race-wsl.sh

Nuri intentionally builds with `CGO_ENABLED=0`, and the Windows dev machine has no C toolchain. The Go race detector requires CGO, so race tests cannot run on Windows directly. This script runs them inside WSL Ubuntu, which has gcc installed and a user-local Go at `~/go-sdk/go`.

### Prerequisites

- WSL with Ubuntu (the distribution name used in the command below is `Ubuntu`)
- `gcc` available in the WSL environment (`sudo apt install gcc` if missing)
- Go installed at `~/go-sdk/go` inside WSL (no system-wide install needed)

### Running it

From a Windows terminal:

```
wsl -d Ubuntu -- bash /mnt/d/dev/my-repos/irosashi/tools/race-wsl.sh
```

The `.gitattributes` entry for `tools/race-wsl.sh` forces LF line endings on checkout, so no CRLF stripping is needed.

### What it runs

```bash
go test -race -count=1 $(go list ./... | grep -v internal/fidelity)
go test -race -run '^$' -bench BenchmarkConcurrent -benchtime=1x ./cmd/bench/
```

`internal/fidelity` is excluded because the fidelity suite is red by design until all held grammars reach 100%, and running 200+ fixture files under the race detector takes 10 to 30 minutes. Everything else in the module runs, including the pool under `RunParallel` and concurrent Do/panic-swap churn.

### Updating the repo path

If your repo is checked out somewhere other than `/mnt/d/dev/my-repos/irosashi`, only the `wsl` invocation needs updating. The script resolves the repo root from its own location at runtime.
