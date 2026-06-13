export PATH=$HOME/go-sdk/go/bin:$PATH
export CGO_ENABLED=1
cd "$(dirname "${BASH_SOURCE[0]}")/.."
go version
echo "=== RACE TESTS ==="
go test -race -count=1 $(go list ./... | grep -v internal/fidelity)
echo "RACE_TESTS_EXIT=$?"
echo "=== RACE BENCH ==="
go test -race -run '^$' -bench BenchmarkConcurrent -benchtime=1x ./cmd/bench/
echo "RACE_BENCH_EXIT=$?"
echo "WSL RACE DONE"
