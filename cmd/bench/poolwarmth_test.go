package bench_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/frostybee/nuri"
	"github.com/frostybee/nuri/internal/shared"
)

// TestPoolWarmth is the end-to-end backstop for the LIFO pool: with the
// DEFAULT pool size (runtime.NumCPU()), repeat renders of the same language
// must reuse the warm instance and its compiled-scanner cache. Under the old
// FIFO pool every sequential render rotated onto a cold instance and paid the
// full JavaScript grammar compile (~185ms) again; warm renders are ~0–2ms.
//
// The deterministic guard lives in internal/oniguruma
// (TestPoolLIFOReusesWarmInstance); this test additionally catches
// regressions anywhere on the Tokenize path that would defeat the cache.
// Bounds are generous to tolerate CI noise and Windows timer granularity.
func TestPoolWarmth(t *testing.T) {
	if testing.Short() {
		t.Skip("timing-based test; skipped under -short")
	}

	ctx := context.Background()
	jsLine := `const x = 1;`

	h, err := nuri.New(ctx,
		nuri.WithGrammarFS(os.DirFS(shared.GrammarsDir(t))),
		nuri.WithThemeFS(os.DirFS(shared.ThemesDir(t))),
	) // default pool size — exactly the configuration that masked the FIFO bug
	if err != nil {
		t.Fatal(err)
	}
	defer h.Close(ctx)

	const renders = 4
	durations := make([]time.Duration, renders)
	for i := 0; i < renders; i++ {
		start := time.Now()
		_, err := h.CodeToTokens(ctx, jsLine, nuri.CodeToTokensOptions{
			Lang:  "javascript",
			Theme: "github-dark",
		})
		if err != nil {
			t.Fatal(err)
		}
		durations[i] = time.Since(start)
		t.Logf("render %d: %v", i+1, durations[i])
	}

	warm := durations[1]
	for _, d := range durations[2:] {
		if d < warm {
			warm = d
		}
	}
	bound := durations[0] / 3
	if floor := 30 * time.Millisecond; bound < floor {
		bound = floor
	}
	if warm >= bound {
		t.Errorf("warm render took %v (cold was %v, bound %v): "+
			"repeat renders are not reusing the warm instance's scanner cache",
			warm, durations[0], bound)
	}
}
