package bench_test

import (
	"context"
	"os"
	"testing"

	"github.com/frostybee/nuri"
	"github.com/frostybee/nuri/internal/shared"
)

// BenchmarkSequentialDefaultPool measures sequential renders against a
// highlighter with the DEFAULT pool size. The other benchmarks in this
// package pin WithPoolSize(1), which is precisely what masked the FIFO
// rotation bug: with one instance every render is warm. This benchmark
// reproduces the real sequential-consumer shape (SSG/CLI): under the old
// FIFO pool it sat at ~185ms/op (every op cold); under the LIFO pool only
// the first op compiles and steady state is sub-millisecond.
func BenchmarkSequentialDefaultPool(b *testing.B) {
	ctx := context.Background()
	h, err := nuri.New(ctx,
		nuri.WithGrammarFS(os.DirFS(shared.GrammarsDir(b))),
		nuri.WithThemeFS(os.DirFS(shared.ThemesDir(b))),
	)
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { h.Close(ctx) })

	code := benchJSCode
	b.SetBytes(int64(len(code)))
	b.ResetTimer()
	for b.Loop() {
		_, err := h.CodeToTokens(ctx, code, nuri.CodeToTokensOptions{
			Lang:  "javascript",
			Theme: "github-dark",
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}
