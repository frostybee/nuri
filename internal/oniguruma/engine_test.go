package oniguruma

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

var testEngine *Engine

func TestMain(m *testing.M) {
	ctx := context.Background()

	eng, err := NewEngine(ctx)
	if err != nil {
		panic("NewEngine: " + err.Error())
	}
	testEngine = eng
	defer eng.Close(ctx)

	m.Run()
}

func newTestInstance(t *testing.T) *instance {
	t.Helper()
	inst, err := testEngine.newInstance(context.Background())
	if err != nil {
		t.Fatalf("newInstance: %v", err)
	}
	t.Cleanup(func() { inst.close(context.Background()) })
	return inst
}

func TestEngineRoundTrip(t *testing.T) {
	ctx := context.Background()
	inst := newTestInstance(t)

	scanner, err := inst.NewScanner(ctx, [][]byte{[]byte(`hello`)})
	if err != nil {
		t.Fatalf("NewScanner: %v", err)
	}
	defer scanner.Free(ctx)

	m, err := scanner.FindNextMatch(ctx, []byte("say hello world"), 0, SearchOptionNone)
	if err != nil {
		t.Fatalf("FindNextMatch: %v", err)
	}
	if m == nil {
		t.Fatal("expected a match")
	}
	if m.Index != 0 {
		t.Errorf("pattern index: got %d, want 0", m.Index)
	}
	assertCap(t, m.Captures, 0, 4, 9)
}

func TestMultiPatternLeftmostMatch(t *testing.T) {
	ctx := context.Background()
	inst := newTestInstance(t)

	patterns := [][]byte{
		[]byte(`\d+`),
		[]byte(`world`),
		[]byte(`hello`),
		[]byte(`hel`),
	}
	scanner, err := inst.NewScanner(ctx, patterns)
	if err != nil {
		t.Fatalf("NewScanner: %v", err)
	}
	defer scanner.Free(ctx)

	text := []byte("hello world 42")
	m, err := scanner.FindNextMatch(ctx, text, 0, SearchOptionNone)
	if err != nil {
		t.Fatalf("FindNextMatch: %v", err)
	}
	if m == nil {
		t.Fatal("expected a match")
	}
	// "hello" and "hel" both start at 0; "hello" (index 2) wins by earlier index
	if m.Index != 2 {
		t.Errorf("winner index: got %d, want 2 (hello)", m.Index)
	}
	assertCap(t, m.Captures, 0, 0, 5)
}

func TestMultiPatternLeftmostTieBreak(t *testing.T) {
	ctx := context.Background()
	inst := newTestInstance(t)

	// Both match at position 0; pattern 0 should win the tie
	patterns := [][]byte{
		[]byte(`hel`),
		[]byte(`hello`),
	}
	scanner, err := inst.NewScanner(ctx, patterns)
	if err != nil {
		t.Fatalf("NewScanner: %v", err)
	}
	defer scanner.Free(ctx)

	m, err := scanner.FindNextMatch(ctx, []byte("hello"), 0, SearchOptionNone)
	if err != nil {
		t.Fatalf("FindNextMatch: %v", err)
	}
	if m == nil {
		t.Fatal("expected a match")
	}
	if m.Index != 0 {
		t.Errorf("winner index: got %d, want 0 (hel, first pattern)", m.Index)
	}
}

func TestCaptureGroups(t *testing.T) {
	ctx := context.Background()
	inst := newTestInstance(t)

	scanner, err := inst.NewScanner(ctx, [][]byte{[]byte(`(\w+)\s+(\w+)`)})
	if err != nil {
		t.Fatalf("NewScanner: %v", err)
	}
	defer scanner.Free(ctx)

	m, err := scanner.FindNextMatch(ctx, []byte("hello world"), 0, SearchOptionNone)
	if err != nil {
		t.Fatalf("FindNextMatch: %v", err)
	}
	if m == nil {
		t.Fatal("expected a match")
	}
	if len(m.Captures) != 3 {
		t.Fatalf("captures: got %d, want 3", len(m.Captures))
	}
	assertCap(t, m.Captures, 0, 0, 11)
	assertCap(t, m.Captures, 1, 0, 5)
	assertCap(t, m.Captures, 2, 6, 11)
}

func TestBackreference(t *testing.T) {
	ctx := context.Background()
	inst := newTestInstance(t)

	scanner, err := inst.NewScanner(ctx, [][]byte{[]byte(`(["']).*?\1`)})
	if err != nil {
		t.Fatalf("NewScanner: %v", err)
	}
	defer scanner.Free(ctx)

	m, err := scanner.FindNextMatch(ctx, []byte("she said 'hi' ok"), 0, SearchOptionNone)
	if err != nil {
		t.Fatalf("FindNextMatch: %v", err)
	}
	if m == nil {
		t.Fatal("expected a match")
	}
	assertCap(t, m.Captures, 0, 9, 13)
	assertCap(t, m.Captures, 1, 9, 10)
}

func TestLookahead(t *testing.T) {
	ctx := context.Background()
	inst := newTestInstance(t)

	scanner, err := inst.NewScanner(ctx, [][]byte{[]byte(`\w+(?=\()`)})
	if err != nil {
		t.Fatalf("NewScanner: %v", err)
	}
	defer scanner.Free(ctx)

	m, err := scanner.FindNextMatch(ctx, []byte("foo(bar)"), 0, SearchOptionNone)
	if err != nil {
		t.Fatalf("FindNextMatch: %v", err)
	}
	if m == nil {
		t.Fatal("expected a match")
	}
	assertCap(t, m.Captures, 0, 0, 3)
}

func TestLookbehind(t *testing.T) {
	ctx := context.Background()
	inst := newTestInstance(t)

	scanner, err := inst.NewScanner(ctx, [][]byte{[]byte(`(?<=\.)\w+`)})
	if err != nil {
		t.Fatalf("NewScanner: %v", err)
	}
	defer scanner.Free(ctx)

	m, err := scanner.FindNextMatch(ctx, []byte("obj.method"), 0, SearchOptionNone)
	if err != nil {
		t.Fatalf("FindNextMatch: %v", err)
	}
	if m == nil {
		t.Fatal("expected a match")
	}
	assertCap(t, m.Captures, 0, 4, 10)
}

func TestGAnchor(t *testing.T) {
	ctx := context.Background()
	inst := newTestInstance(t)

	scanner, err := inst.NewScanner(ctx, [][]byte{[]byte(`\G\w`)})
	if err != nil {
		t.Fatalf("NewScanner: %v", err)
	}
	defer scanner.Free(ctx)

	text := []byte("abc")

	m, err := scanner.FindNextMatch(ctx, text, 0, SearchOptionNone)
	if err != nil {
		t.Fatalf("FindNextMatch at 0: %v", err)
	}
	if m == nil {
		t.Fatal("expected match at 0")
	}
	assertCap(t, m.Captures, 0, 0, 1)

	m, err = scanner.FindNextMatch(ctx, text, 1, SearchOptionNone)
	if err != nil {
		t.Fatalf("FindNextMatch at 1: %v", err)
	}
	if m == nil {
		t.Fatal("expected match at 1")
	}
	assertCap(t, m.Captures, 0, 1, 2)
}

func TestUTF8Multibyte(t *testing.T) {
	ctx := context.Background()
	inst := newTestInstance(t)

	scanner, err := inst.NewScanner(ctx, [][]byte{[]byte(`\d+`)})
	if err != nil {
		t.Fatalf("NewScanner: %v", err)
	}
	defer scanner.Free(ctx)

	text := []byte("変数 = 42")
	m, err := scanner.FindNextMatch(ctx, text, 0, SearchOptionNone)
	if err != nil {
		t.Fatalf("FindNextMatch: %v", err)
	}
	if m == nil {
		t.Fatal("expected a match")
	}
	matched := string(text[m.Captures[0].Start:m.Captures[0].End])
	if matched != "42" {
		t.Errorf("expected %q, got %q", "42", matched)
	}
	if m.Captures[0].Start < 6 {
		t.Errorf("byte offset %d too small — offsets may be rune-based", m.Captures[0].Start)
	}
}

func TestUTF8Emoji(t *testing.T) {
	ctx := context.Background()
	inst := newTestInstance(t)

	scanner, err := inst.NewScanner(ctx, [][]byte{[]byte(`.+`)})
	if err != nil {
		t.Fatalf("NewScanner: %v", err)
	}
	defer scanner.Free(ctx)

	text := []byte("hi 👋 ok")
	m, err := scanner.FindNextMatch(ctx, text, 0, SearchOptionNone)
	if err != nil {
		t.Fatalf("FindNextMatch: %v", err)
	}
	if m == nil {
		t.Fatal("expected a match")
	}
	if m.Captures[0].Start != 0 {
		t.Errorf("expected start=0, got %d", m.Captures[0].Start)
	}
	if m.Captures[0].End != len(text) {
		t.Errorf("expected end=%d, got %d", len(text), m.Captures[0].End)
	}
}

func TestNoMatch(t *testing.T) {
	ctx := context.Background()
	inst := newTestInstance(t)

	scanner, err := inst.NewScanner(ctx, [][]byte{[]byte(`\d+`)})
	if err != nil {
		t.Fatalf("NewScanner: %v", err)
	}
	defer scanner.Free(ctx)

	m, err := scanner.FindNextMatch(ctx, []byte("hello"), 0, SearchOptionNone)
	if err != nil {
		t.Fatalf("FindNextMatch: %v", err)
	}
	if m != nil {
		t.Errorf("expected no match, got %+v", m)
	}
}

func TestUnmatchedOptionalGroup(t *testing.T) {
	ctx := context.Background()
	inst := newTestInstance(t)

	scanner, err := inst.NewScanner(ctx, [][]byte{[]byte(`(\w+)(?:\s+(\d+))?`)})
	if err != nil {
		t.Fatalf("NewScanner: %v", err)
	}
	defer scanner.Free(ctx)

	m, err := scanner.FindNextMatch(ctx, []byte("hello"), 0, SearchOptionNone)
	if err != nil {
		t.Fatalf("FindNextMatch: %v", err)
	}
	if m == nil {
		t.Fatal("expected a match")
	}
	if len(m.Captures) != 3 {
		t.Fatalf("captures: got %d, want 3", len(m.Captures))
	}
	assertCap(t, m.Captures, 0, 0, 5)
	assertCap(t, m.Captures, 1, 0, 5)
	// Group 2 is unmatched
	assertCap(t, m.Captures, 2, -1, -1)
}

func TestEmptyText(t *testing.T) {
	ctx := context.Background()
	inst := newTestInstance(t)

	scanner, err := inst.NewScanner(ctx, [][]byte{[]byte(`\w+`)})
	if err != nil {
		t.Fatalf("NewScanner: %v", err)
	}
	defer scanner.Free(ctx)

	m, err := scanner.FindNextMatch(ctx, []byte{}, 0, SearchOptionNone)
	if err != nil {
		t.Fatalf("FindNextMatch: %v", err)
	}
	if m != nil {
		t.Errorf("expected nil match for empty text, got %+v", m)
	}
}

func TestScannerIterative(t *testing.T) {
	ctx := context.Background()
	inst := newTestInstance(t)

	scanner, err := inst.NewScanner(ctx, [][]byte{[]byte(`\w+`)})
	if err != nil {
		t.Fatalf("NewScanner: %v", err)
	}
	defer scanner.Free(ctx)

	text := []byte("foo bar baz")
	var words []string
	pos := 0
	for {
		m, err := scanner.FindNextMatch(ctx, text, pos, SearchOptionNone)
		if err != nil {
			t.Fatalf("FindNextMatch at %d: %v", pos, err)
		}
		if m == nil {
			break
		}
		words = append(words, string(text[m.Captures[0].Start:m.Captures[0].End]))
		pos = m.Captures[0].End
	}
	if len(words) != 3 || words[0] != "foo" || words[1] != "bar" || words[2] != "baz" {
		t.Errorf("words: got %v, want [foo bar baz]", words)
	}
}

func TestEmptyPatternZeroWidthMatch(t *testing.T) {
	ctx := context.Background()
	inst := newTestInstance(t)

	scanner, err := inst.NewScanner(ctx, [][]byte{[]byte("")})
	if err != nil {
		t.Fatalf("NewScanner with empty pattern: %v", err)
	}
	defer scanner.Free(ctx)

	m, err := scanner.FindNextMatch(ctx, []byte("hello"), 0, SearchOptionNone)
	if err != nil {
		t.Fatalf("FindNextMatch: %v", err)
	}
	if m == nil {
		t.Fatal("expected a zero-width match for empty pattern")
	}
	if m.Captures[0].Start != 0 || m.Captures[0].End != 0 {
		t.Errorf("expected [0:0], got [%d:%d]", m.Captures[0].Start, m.Captures[0].End)
	}
}

func TestInvalidPattern(t *testing.T) {
	ctx := context.Background()
	inst := newTestInstance(t)

	_, err := inst.NewScanner(ctx, [][]byte{[]byte(`(?P<`)})
	if err == nil {
		t.Fatal("expected error for invalid pattern")
	}
}

func TestPoolGetPut(t *testing.T) {
	ctx := context.Background()
	pool, err := NewPool(ctx, testEngine, 2)
	if err != nil {
		t.Fatalf("NewPool: %v", err)
	}
	defer pool.Close(ctx)

	inst1, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get 1: %v", err)
	}
	inst2, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get 2: %v", err)
	}

	pool.Put(inst1)
	pool.Put(inst2)

	inst3, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get 3: %v", err)
	}
	pool.Put(inst3)
}

func TestPoolConcurrency(t *testing.T) {
	ctx := context.Background()
	poolSize := 4
	pool, err := NewPool(ctx, testEngine, poolSize)
	if err != nil {
		t.Fatalf("NewPool: %v", err)
	}
	defer pool.Close(ctx)

	var wg sync.WaitGroup
	goroutines := 16
	iterations := 10

	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				inst, err := pool.Get(ctx)
				if err != nil {
					t.Errorf("Get: %v", err)
					return
				}

				scanner, err := inst.NewScanner(ctx, [][]byte{[]byte(`\w+`)})
				if err != nil {
					t.Errorf("NewScanner: %v", err)
					pool.Put(inst)
					return
				}

				m, err := scanner.FindNextMatch(ctx, []byte("test"), 0, SearchOptionNone)
				if err != nil {
					t.Errorf("FindNextMatch: %v", err)
				}
				if m == nil {
					t.Error("expected match")
				}

				scanner.Free(ctx)
				pool.Put(inst)
			}
		}()
	}

	wg.Wait()
}

func TestPoolSwap(t *testing.T) {
	ctx := context.Background()
	pool, err := NewPool(ctx, testEngine, 1)
	if err != nil {
		t.Fatalf("NewPool: %v", err)
	}
	defer pool.Close(ctx)

	inst, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	inst.poisoned = true
	err = pool.Swap(ctx, inst)
	if err != nil {
		t.Fatalf("Swap: %v", err)
	}

	fresh, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get after swap: %v", err)
	}
	if fresh.poisoned {
		t.Error("replacement instance should not be poisoned")
	}
	pool.Put(fresh)
}

func TestPoolDo(t *testing.T) {
	ctx := context.Background()
	pool, err := NewPool(ctx, testEngine, 1)
	if err != nil {
		t.Fatalf("NewPool: %v", err)
	}
	defer pool.Close(ctx)

	var matched bool
	err = pool.Do(ctx, func(lib OnigLib) error {
		scanner, err := lib.NewScannerCtx(ctx, [][]byte{[]byte(`\w+`)})
		if err != nil {
			return err
		}
		defer scanner.Close()
		m, err := scanner.FindNextMatchCtx(ctx, []byte("hello"), 0, SearchOptionNone)
		if err != nil {
			return err
		}
		matched = m != nil
		return nil
	})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if !matched {
		t.Error("expected match inside Do")
	}

	// Verify instance was returned — a second Do should work.
	err = pool.Do(ctx, func(lib OnigLib) error { return nil })
	if err != nil {
		t.Fatalf("second Do: %v", err)
	}
}

func TestPoolDoPanicRecovery(t *testing.T) {
	ctx := context.Background()
	pool, err := NewPool(ctx, testEngine, 1)
	if err != nil {
		t.Fatalf("NewPool: %v", err)
	}
	defer pool.Close(ctx)

	// Do with a panic — should recover and swap the instance.
	err = pool.Do(ctx, func(lib OnigLib) error {
		panic("test panic")
	})
	if err == nil {
		t.Fatal("expected error from panicking Do")
	}

	// Pool should still be usable — the poisoned instance was swapped.
	var matched bool
	err = pool.Do(ctx, func(lib OnigLib) error {
		scanner, err := lib.NewScannerCtx(ctx, [][]byte{[]byte(`\d+`)})
		if err != nil {
			return err
		}
		defer scanner.Close()
		m, err := scanner.FindNextMatchCtx(ctx, []byte("42"), 0, SearchOptionNone)
		if err != nil {
			return err
		}
		matched = m != nil
		return nil
	})
	if err != nil {
		t.Fatalf("Do after panic recovery: %v", err)
	}
	if !matched {
		t.Error("expected match after panic recovery")
	}
}

func TestPoolDrain(t *testing.T) {
	ctx := context.Background()
	pool, err := NewPool(ctx, testEngine, 3)
	if err != nil {
		t.Fatalf("NewPool: %v", err)
	}

	// Get all instances and put them back
	insts := make([]*instance, 3)
	for i := 0; i < 3; i++ {
		insts[i], err = pool.Get(ctx)
		if err != nil {
			t.Fatalf("Get %d: %v", i, err)
		}
	}
	for _, inst := range insts {
		pool.Put(inst)
	}

	err = pool.Close(ctx)
	if err != nil {
		t.Errorf("Close: %v", err)
	}
}

func assertCap(t *testing.T, captures []Capture, group, wantStart, wantEnd int) {
	t.Helper()
	if group >= len(captures) {
		t.Fatalf("group %d: not present (only %d captures)", group, len(captures))
	}
	got := captures[group]
	if got.Start != wantStart || got.End != wantEnd {
		t.Errorf("group %d: got [%d:%d], want [%d:%d]", group, got.Start, got.End, wantStart, wantEnd)
	}
}

func TestPoolLIFOOrder(t *testing.T) {
	ctx := context.Background()
	pool, err := NewPool(ctx, testEngine, 2)
	if err != nil {
		t.Fatalf("NewPool: %v", err)
	}
	defer pool.Close(ctx)

	inst1, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get 1: %v", err)
	}
	inst2, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get 2: %v", err)
	}

	pool.Put(inst1)
	pool.Put(inst2)

	// LIFO: the most recently returned instance comes back first.
	got, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get 3: %v", err)
	}
	if got != inst2 {
		t.Error("expected LIFO order: Get should return the most recently Put instance")
	}
	pool.Put(got)
}

func TestPoolLazyCreation(t *testing.T) {
	ctx := context.Background()
	pool, err := NewPool(ctx, testEngine, 4)
	if err != nil {
		t.Fatalf("NewPool: %v", err)
	}
	defer pool.Close(ctx)

	pool.mu.Lock()
	created := len(pool.all)
	pool.mu.Unlock()
	if created != 0 {
		t.Fatalf("NewPool created %d instances; want 0 (lazy)", created)
	}

	inst, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	pool.mu.Lock()
	created = len(pool.all)
	pool.mu.Unlock()
	if created != 1 {
		t.Errorf("after one Get: %d instances created; want 1", created)
	}
	pool.Put(inst)
}

func TestPoolGetCtxCancelled(t *testing.T) {
	ctx := context.Background()
	pool, err := NewPool(ctx, testEngine, 1)
	if err != nil {
		t.Fatalf("NewPool: %v", err)
	}
	defer pool.Close(ctx)

	// Hold the only instance so the next Get must wait on the semaphore.
	inst, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	_, err = pool.Get(cancelled)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Get with cancelled ctx: got %v, want context.Canceled", err)
	}
	pool.Put(inst)
}

func TestPoolGetAfterClose(t *testing.T) {
	ctx := context.Background()
	pool, err := NewPool(ctx, testEngine, 2)
	if err != nil {
		t.Fatalf("NewPool: %v", err)
	}
	if err := pool.Close(ctx); err != nil {
		t.Fatalf("Close: %v", err)
	}

	_, err = pool.Get(ctx)
	if err == nil {
		t.Fatal("Get after Close: expected error, got instance")
	}
}

func TestPoolGetCreateFailureReturnsToken(t *testing.T) {
	ctx := context.Background()

	// A private, already-closed engine makes newInstance fail.
	eng, err := NewEngine(ctx)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	eng.Close(ctx)

	pool, err := NewPool(ctx, eng, 1)
	if err != nil {
		t.Fatalf("NewPool: %v", err)
	}
	defer pool.Close(ctx)

	if _, err := pool.Get(ctx); err == nil {
		t.Fatal("Get with closed engine: expected creation error")
	}

	// If Get leaked its semaphore token on the failure path, this second Get
	// would block until the deadline instead of failing fast on creation.
	bounded, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_, err = pool.Get(bounded)
	if err == nil {
		t.Fatal("second Get: expected creation error")
	}
	if errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("second Get blocked on the semaphore: token leaked by the failure path")
	}
}

func TestPoolConcurrentDoWithPanics(t *testing.T) {
	ctx := context.Background()
	pool, err := NewPool(ctx, testEngine, 3)
	if err != nil {
		t.Fatalf("NewPool: %v", err)
	}
	defer pool.Close(ctx)

	var wg sync.WaitGroup
	for g := 0; g < 12; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < 5; i++ {
				panicking := (g+i)%3 == 0
				err := pool.Do(ctx, func(lib OnigLib) error {
					if panicking {
						panic("test panic")
					}
					scanner, err := lib.GetOrCreateScannerCtx(ctx, [][]byte{[]byte(`\w+`)})
					if err != nil {
						return err
					}
					_, err = scanner.FindNextMatchCtx(ctx, []byte("hello"), 0, SearchOptionNone)
					return err
				})
				if panicking && err == nil {
					t.Error("Do with panicking fn: expected error")
				}
				if !panicking && err != nil {
					t.Errorf("Do: %v", err)
				}
			}
		}(g)
	}
	wg.Wait()

	// The pool must remain fully usable after the panic/swap churn.
	err = pool.Do(ctx, func(lib OnigLib) error {
		scanner, err := lib.GetOrCreateScannerCtx(ctx, [][]byte{[]byte(`world`)})
		if err != nil {
			return err
		}
		m, err := scanner.FindNextMatchCtx(ctx, []byte("hello world"), 0, SearchOptionNone)
		if err != nil {
			return err
		}
		if m == nil {
			return errors.New("expected a match")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Do after churn: %v", err)
	}
}
