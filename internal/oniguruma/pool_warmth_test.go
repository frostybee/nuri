package oniguruma

import (
	"context"
	"testing"
)

// TestPoolLIFOReusesWarmInstance is the regression guard for the LIFO pool
// design. Compiled scanners are cached per instance, so sequential borrows
// must keep landing on the same (warm) instance: two sequential Do calls
// asking for the same pattern set must receive the exact same cached scanner.
// Under a FIFO pool this fails deterministically — the second Do rotates onto
// a different, cold instance and compiles a new scanner.
func TestPoolLIFOReusesWarmInstance(t *testing.T) {
	ctx := context.Background()
	pool, err := NewPool(ctx, testEngine, 4)
	if err != nil {
		t.Fatalf("NewPool: %v", err)
	}
	defer pool.Close(ctx)

	patterns := [][]byte{[]byte(`\bfunc\b`), []byte(`\d+`)}

	scanners := make([]OnigScanner, 2)
	for i := range scanners {
		err := pool.Do(ctx, func(lib OnigLib) error {
			s, err := lib.GetOrCreateScannerCtx(ctx, patterns)
			if err != nil {
				return err
			}
			scanners[i] = s
			return nil
		})
		if err != nil {
			t.Fatalf("Do %d: %v", i, err)
		}
	}

	if scanners[0] != scanners[1] {
		t.Error("sequential Do calls hit different instances: " +
			"expected LIFO reuse of the warm instance and its cached scanner")
	}
}
