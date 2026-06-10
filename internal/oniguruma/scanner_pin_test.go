package oniguruma

import (
	"context"
	"reflect"
	"testing"
)

// referenceMatch runs the same search on a fresh instance + fresh scanner,
// giving an oracle untouched by buffer reuse.
func referenceMatch(t *testing.T, patterns [][]byte, text []byte, pos int, options SearchOptions) *Match {
	t.Helper()
	ctx := context.Background()
	inst := newTestInstance(t)
	s, err := inst.NewScanner(ctx, patterns)
	if err != nil {
		t.Fatalf("reference NewScanner: %v", err)
	}
	defer s.Free(ctx)
	m, err := s.FindNextMatch(ctx, text, pos, options)
	if err != nil {
		t.Fatalf("reference FindNextMatch: %v", err)
	}
	return m
}

// TestFindNextMatchPinSequence drives the pinned-upload path through
// full → prefix → shifted prefix → full → same-length-different-array
// transitions and compares every result against a fresh-instance oracle.
func TestFindNextMatchPinSequence(t *testing.T) {
	ctx := context.Background()
	inst := newTestInstance(t)

	patterns := [][]byte{[]byte(`\w+`)}
	scanner, err := inst.NewScanner(ctx, patterns)
	if err != nil {
		t.Fatalf("NewScanner: %v", err)
	}
	defer scanner.Free(ctx)

	full := []byte("alpha beta gamma")
	// Same length as full, different backing array, different content —
	// catches a stale pin that skips the upload it must not skip.
	other := []byte("12345 67890 88888")[:len(full)]

	steps := []struct {
		name string
		text []byte
		pos  int
	}{
		{"full", full, 0},
		{"prefix", full[:10], 0},
		{"prefix_later_pos", full[:10], 6},
		{"full_again", full, 6},
		{"same_len_other_array", other, 0},
		{"back_to_full", full, 0},
	}

	for _, st := range steps {
		got, err := scanner.FindNextMatch(ctx, st.text, st.pos, SearchOptionNone)
		if err != nil {
			t.Fatalf("%s: FindNextMatch: %v", st.name, err)
		}
		want := referenceMatch(t, patterns, st.text, st.pos, SearchOptionNone)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s: got %+v, want %+v", st.name, got, want)
		}
	}
}

// TestFindNextMatchBufferGrowth crosses the text-buffer growth boundary
// (initial 4096 bytes) and verifies results stay correct after realloc.
func TestFindNextMatchBufferGrowth(t *testing.T) {
	ctx := context.Background()
	inst := newTestInstance(t)

	patterns := [][]byte{[]byte(`needle\d+`)}
	scanner, err := inst.NewScanner(ctx, patterns)
	if err != nil {
		t.Fatalf("NewScanner: %v", err)
	}
	defer scanner.Free(ctx)

	small := []byte("pad needle1 pad")
	big := make([]byte, 10000)
	for i := range big {
		big[i] = 'x'
	}
	copy(big[9000:], []byte("needle22"))

	for _, st := range []struct {
		name string
		text []byte
	}{
		{"small_before_growth", small},
		{"big_forces_growth", big},
		{"small_after_growth", small},
	} {
		got, err := scanner.FindNextMatch(ctx, st.text, 0, SearchOptionNone)
		if err != nil {
			t.Fatalf("%s: %v", st.name, err)
		}
		want := referenceMatch(t, patterns, st.text, 0, SearchOptionNone)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s: got %+v, want %+v", st.name, got, want)
		}
	}
}

// TestFindNextMatchSharedAcrossScanners verifies that two scanners on one
// instance share the pinned text buffer correctly (the pin belongs to the
// instance, not the scanner).
func TestFindNextMatchSharedAcrossScanners(t *testing.T) {
	ctx := context.Background()
	inst := newTestInstance(t)

	sWord, err := inst.NewScanner(ctx, [][]byte{[]byte(`[a-z]+`)})
	if err != nil {
		t.Fatalf("NewScanner word: %v", err)
	}
	defer sWord.Free(ctx)
	sNum, err := inst.NewScanner(ctx, [][]byte{[]byte(`\d+`)})
	if err != nil {
		t.Fatalf("NewScanner num: %v", err)
	}
	defer sNum.Free(ctx)

	text := []byte("abc 123 def")

	m1, err := sWord.FindNextMatch(ctx, text, 0, SearchOptionNone)
	if err != nil || m1 == nil {
		t.Fatalf("word scan: m=%v err=%v", m1, err)
	}
	// Same text bytes on the second scanner — upload is skipped, result
	// must still be from the correct (current) text.
	m2, err := sNum.FindNextMatch(ctx, text, 0, SearchOptionNone)
	if err != nil || m2 == nil {
		t.Fatalf("num scan: m=%v err=%v", m2, err)
	}
	if m1.Captures[0].Start != 0 || m1.Captures[0].End != 3 {
		t.Errorf("word match: got [%d:%d], want [0:3]", m1.Captures[0].Start, m1.Captures[0].End)
	}
	if m2.Captures[0].Start != 4 || m2.Captures[0].End != 7 {
		t.Errorf("num match: got [%d:%d], want [4:7]", m2.Captures[0].Start, m2.Captures[0].End)
	}
}
