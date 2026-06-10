package tokenizer

import (
	"context"
	"testing"

	"github.com/frostybee/nuri/internal/grammar"
)

// TestMemoHitEndRulePath enters the same BeginEndRule twice on one line, so
// the second entry's rule context is a memo hit and the EndRule match must
// dispatch against the live frame's EndRule, not the cached fill-time one.
func TestMemoHitEndRulePath(t *testing.T) {
	ctx := context.Background()
	g := loadMiniGrammar(t, "begin_end.json")
	onigLib := newTestOnigLib(t)

	result, err := Tokenize(ctx, []byte(`a "x" b "y" c`), g, onigLib, TokenizeOptions{})
	if err != nil {
		t.Fatalf("Tokenize: %v", err)
	}

	tokens := result.Lines[0]
	dumpTokens(t, tokens)

	var stringRanges [][2]int
	for _, tok := range tokens {
		for _, s := range tok.Scopes {
			if s == "string.quoted.double.test" {
				stringRanges = append(stringRanges, [2]int{tok.Start, tok.End})
				break
			}
		}
	}
	// Both quoted regions ("x" spans 2..5, "y" spans 8..11) must carry the
	// string scope: open punctuation, content, close punctuation each.
	if len(stringRanges) < 6 {
		t.Errorf("expected >= 6 string-scoped tokens across both strings, got %d (%v)",
			len(stringRanges), stringRanges)
	}
	var coveredFirst, coveredSecond bool
	for _, r := range stringRanges {
		if r[0] >= 2 && r[1] <= 5 {
			coveredFirst = true
		}
		if r[0] >= 8 && r[1] <= 11 {
			coveredSecond = true
		}
	}
	if !coveredFirst || !coveredSecond {
		t.Errorf("expected string scope on both quoted regions, got ranges %v", stringRanges)
	}
}

// TestMemoDistinctEndPatternsDistinctEntries verifies that two blocks whose
// begin captures resolve to different end patterns get distinct memo entries
// and distinct scanners, while a repeat of the same end pattern memo-hits.
func TestMemoDistinctEndPatternsDistinctEntries(t *testing.T) {
	ctx := context.Background()
	g := loadMiniGrammar(t, "begin_end_backref.json")
	onigLib := newTestOnigLib(t)

	var beginEnd *grammar.BeginEndRule
	for _, r := range g.Patterns {
		if be, ok := r.(*grammar.BeginEndRule); ok {
			beginEnd = be
			break
		}
	}
	if beginEnd == nil {
		t.Fatal("grammar has no BeginEndRule")
	}

	memo := newCompileMemo(onigLib, nil)

	endDouble := &grammar.EndRule{ID: 1000001, Parent: beginEnd, EndPattern: `\"`, EndCaptures: beginEnd.EndCaptures}
	endSingle := &grammar.EndRule{ID: 1000002, Parent: beginEnd, EndPattern: `'`, EndCaptures: beginEnd.EndCaptures}

	e1 := memo.getOrCompile(ctx, beginEnd, beginEnd.Patterns, g, endDouble)
	if e1.err != nil {
		t.Fatalf("compile double-quote context: %v", e1.err)
	}
	e2 := memo.getOrCompile(ctx, beginEnd, beginEnd.Patterns, g, endSingle)
	if e2.err != nil {
		t.Fatalf("compile single-quote context: %v", e2.err)
	}
	if e1 == e2 {
		t.Error("different end patterns must produce distinct memo entries")
	}
	if e1.scanner == e2.scanner {
		t.Error("different end patterns must produce distinct scanners")
	}

	// A separate EndRule instance with the SAME pattern must memo-hit.
	endDouble2 := &grammar.EndRule{ID: 1000003, Parent: beginEnd, EndPattern: `\"`, EndCaptures: beginEnd.EndCaptures}
	e3 := memo.getOrCompile(ctx, beginEnd, beginEnd.Patterns, g, endDouble2)
	if e3 != e1 {
		t.Error("same rule context + same end pattern must return the same memo entry")
	}
}

// TestMemoEmptyEndPatternDistinctFromNoEnd pins the hasEnd key component:
// a backref-resolved empty end pattern must not collide with "no end rule".
func TestMemoEmptyEndPatternDistinctFromNoEnd(t *testing.T) {
	ctx := context.Background()
	g := loadMiniGrammar(t, "begin_end_backref.json")
	onigLib := newTestOnigLib(t)

	var beginEnd *grammar.BeginEndRule
	for _, r := range g.Patterns {
		if be, ok := r.(*grammar.BeginEndRule); ok {
			beginEnd = be
			break
		}
	}
	if beginEnd == nil {
		t.Fatal("grammar has no BeginEndRule")
	}

	memo := newCompileMemo(onigLib, nil)

	noEnd := memo.getOrCompile(ctx, beginEnd, beginEnd.Patterns, g, nil)
	emptyEnd := memo.getOrCompile(ctx, beginEnd, beginEnd.Patterns, g,
		&grammar.EndRule{ID: 1000004, Parent: beginEnd, EndPattern: "", EndCaptures: beginEnd.EndCaptures})

	if noEnd == emptyEnd {
		t.Error("empty end pattern must not collide with no-end-rule context")
	}
	if len(emptyEnd.rules) != len(noEnd.rules)+1 {
		t.Errorf("empty-end entry should have one extra rule: got %d vs %d",
			len(emptyEnd.rules), len(noEnd.rules))
	}
}

// TestMemoCaptureRootStable verifies capture retokenization reuses a stable
// root rule per CaptureRule, so the compile memo hits on hot capture paths.
func TestMemoCaptureRootStable(t *testing.T) {
	onigLib := newTestOnigLib(t)
	memo := newCompileMemo(onigLib, nil)

	cr := &grammar.CaptureRule{
		ID:       7,
		Name:     "test.capture",
		Patterns: []grammar.Rule{&grammar.MatchRule{ID: 8, Match: `\w+`, Name: "word.test"}},
	}

	r1 := memo.captureRoot(cr)
	r2 := memo.captureRoot(cr)
	if r1 != r2 {
		t.Error("captureRoot must return a stable pointer per CaptureRule")
	}
	if len(r1.Patterns) != 1 || r1.Patterns[0] != cr.Patterns[0] {
		t.Error("captureRoot must wrap the capture rule's patterns")
	}
}
