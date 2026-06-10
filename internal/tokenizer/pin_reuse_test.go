package tokenizer

import (
	"bytes"
	"context"
	"reflect"
	"testing"
)

// TestRepeatedIdenticalLines verifies that identical lines tokenize
// identically — each line uploads a fresh scanLine allocation, so this
// exercises re-upload after the previous line's pin goes stale.
func TestRepeatedIdenticalLines(t *testing.T) {
	ctx := context.Background()
	g := loadMiniGrammar(t, "begin_end.json")
	onigLib := newTestOnigLib(t)

	line := `say "hi" end`
	code := bytes.Repeat([]byte(line+"\n"), 5)

	result, err := Tokenize(ctx, code, g, onigLib, TokenizeOptions{})
	if err != nil {
		t.Fatalf("Tokenize: %v", err)
	}
	if len(result.Lines) != 5 {
		t.Fatalf("lines: got %d, want 5", len(result.Lines))
	}
	for i := 1; i < len(result.Lines); i++ {
		if !reflect.DeepEqual(result.Lines[0], result.Lines[i]) {
			t.Errorf("line %d tokens differ from line 0:\n0: %+v\n%d: %+v",
				i, result.Lines[0], i, result.Lines[i])
		}
	}
}

// TestCaptureRecursionPrefixReuse exercises the capture-retokenization path,
// which searches prefixes of the already-uploaded scanLine (free under the
// pin) and then returns to full-line scanning.
func TestCaptureRecursionPrefixReuse(t *testing.T) {
	ctx := context.Background()
	g := loadMiniGrammar(t, "overlapping_captures.json")
	onigLib := newTestOnigLib(t)

	result, err := Tokenize(ctx, []byte("foo bar baz\nfoo bar baz"), g, onigLib, TokenizeOptions{})
	if err != nil {
		t.Fatalf("Tokenize: %v", err)
	}
	if len(result.Lines) != 2 {
		t.Fatalf("lines: got %d, want 2", len(result.Lines))
	}
	if !reflect.DeepEqual(result.Lines[0], result.Lines[1]) {
		t.Errorf("identical lines tokenized differently:\n0: %+v\n1: %+v",
			result.Lines[0], result.Lines[1])
	}
	if len(result.Lines[0]) == 0 {
		t.Error("expected tokens on line 0")
	}
}
