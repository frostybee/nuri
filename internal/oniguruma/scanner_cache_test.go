package oniguruma

import (
	"context"
	"testing"
)

func TestGetOrCreateScannerReturnsSameScanner(t *testing.T) {
	ctx := context.Background()
	inst := newTestInstance(t)

	patterns := [][]byte{[]byte(`\w+`), []byte(`\d+`)}
	s1, err := inst.GetOrCreateScannerCtx(ctx, patterns)
	if err != nil {
		t.Fatalf("GetOrCreateScannerCtx 1: %v", err)
	}

	// Rebuild the pattern set with fresh backing arrays — must still hit.
	rebuilt := [][]byte{[]byte(`\w+`), []byte(`\d+`)}
	s2, err := inst.GetOrCreateScannerCtx(ctx, rebuilt)
	if err != nil {
		t.Fatalf("GetOrCreateScannerCtx 2: %v", err)
	}
	if s1 != s2 {
		t.Error("same pattern set should return the identical cached scanner")
	}

	// A different set must produce a different scanner.
	s3, err := inst.GetOrCreateScannerCtx(ctx, [][]byte{[]byte(`\w+`)})
	if err != nil {
		t.Fatalf("GetOrCreateScannerCtx 3: %v", err)
	}
	if s3 == s1 {
		t.Error("different pattern set must not share a scanner")
	}
}

func TestGetOrCreateScannerSplitAmbiguity(t *testing.T) {
	ctx := context.Background()
	inst := newTestInstance(t)

	// ["ab","c"] and ["a","bc"] join to the same bytes — the per-pattern
	// length check must keep them distinct.
	s1, err := inst.GetOrCreateScannerCtx(ctx, [][]byte{[]byte("ab"), []byte("c")})
	if err != nil {
		t.Fatalf("GetOrCreateScannerCtx 1: %v", err)
	}
	s2, err := inst.GetOrCreateScannerCtx(ctx, [][]byte{[]byte("a"), []byte("bc")})
	if err != nil {
		t.Fatalf("GetOrCreateScannerCtx 2: %v", err)
	}
	if s1 == s2 {
		t.Error("pattern sets with identical joined bytes must not collide")
	}
}

func TestGetOrCreateScannerForcedCollision(t *testing.T) {
	ctx := context.Background()
	inst := newTestInstance(t)

	orig := hashPatternSet
	hashPatternSet = func([][]byte) uint64 { return 42 }
	defer func() { hashPatternSet = orig }()

	pa := [][]byte{[]byte(`alpha`)}
	pb := [][]byte{[]byte(`beta`)}

	sa, err := inst.GetOrCreateScannerCtx(ctx, pa)
	if err != nil {
		t.Fatalf("GetOrCreateScannerCtx alpha: %v", err)
	}
	sb, err := inst.GetOrCreateScannerCtx(ctx, pb)
	if err != nil {
		t.Fatalf("GetOrCreateScannerCtx beta: %v", err)
	}
	if sa == sb {
		t.Fatal("colliding hash bucket returned the wrong scanner")
	}

	// Both bucket entries must resolve correctly on re-lookup.
	if got, _ := inst.GetOrCreateScannerCtx(ctx, pa); got != sa {
		t.Error("re-lookup of alpha returned a different scanner")
	}
	if got, _ := inst.GetOrCreateScannerCtx(ctx, pb); got != sb {
		t.Error("re-lookup of beta returned a different scanner")
	}

	// And each scanner must match its own pattern.
	m, err := sa.FindNextMatchCtx(ctx, []byte("xx alpha xx"), 0, SearchOptionNone)
	if err != nil || m == nil {
		t.Fatalf("alpha scanner: m=%v err=%v", m, err)
	}
	m, err = sb.FindNextMatchCtx(ctx, []byte("xx beta xx"), 0, SearchOptionNone)
	if err != nil || m == nil {
		t.Fatalf("beta scanner: m=%v err=%v", m, err)
	}
	if m.Captures[0].Start != 3 {
		t.Errorf("beta match start: got %d, want 3", m.Captures[0].Start)
	}
}

func TestGetOrCreateScannerErrorNotCached(t *testing.T) {
	ctx := context.Background()
	inst := newTestInstance(t)

	bad := [][]byte{[]byte(`(?P<`)}
	if _, err := inst.GetOrCreateScannerCtx(ctx, bad); err == nil {
		t.Fatal("expected error for invalid pattern")
	}
	// Failed compiles must not leave a bucket entry behind.
	if _, err := inst.GetOrCreateScannerCtx(ctx, bad); err == nil {
		t.Fatal("expected error on second attempt too")
	}

	good := [][]byte{[]byte(`\w+`)}
	s, err := inst.GetOrCreateScannerCtx(ctx, good)
	if err != nil {
		t.Fatalf("valid pattern after failures: %v", err)
	}
	m, err := s.FindNextMatchCtx(ctx, []byte("ok"), 0, SearchOptionNone)
	if err != nil || m == nil {
		t.Fatalf("expected match, got m=%v err=%v", m, err)
	}
}

func TestPoolDoPanicSwapKeepsCacheUsable(t *testing.T) {
	ctx := context.Background()
	pool, err := NewPool(ctx, testEngine, 1)
	if err != nil {
		t.Fatalf("NewPool: %v", err)
	}
	defer pool.Close(ctx)

	patterns := [][]byte{[]byte(`\d+`)}

	// Warm the cache, then panic to poison the instance.
	err = pool.Do(ctx, func(lib OnigLib) error {
		if _, err := lib.GetOrCreateScannerCtx(ctx, patterns); err != nil {
			return err
		}
		panic("test panic after caching")
	})
	if err == nil {
		t.Fatal("expected error from panicking Do")
	}

	// The fresh swapped-in instance must compile and match from scratch.
	var matched bool
	err = pool.Do(ctx, func(lib OnigLib) error {
		s, err := lib.GetOrCreateScannerCtx(ctx, patterns)
		if err != nil {
			return err
		}
		m, err := s.FindNextMatchCtx(ctx, []byte("a 42"), 0, SearchOptionNone)
		if err != nil {
			return err
		}
		matched = m != nil
		return nil
	})
	if err != nil {
		t.Fatalf("Do after panic swap: %v", err)
	}
	if !matched {
		t.Error("expected match on fresh instance after swap")
	}
}
