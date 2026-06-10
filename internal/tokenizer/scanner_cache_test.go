package tokenizer

import (
	"context"
	"reflect"
	"testing"
)

// TestTokenizeTwiceSameInstanceIdentical exercises the persistent
// per-instance scanner cache: the second Tokenize on the same checked-out
// instance runs fully from cached scanners and must produce identical output.
func TestTokenizeTwiceSameInstanceIdentical(t *testing.T) {
	ctx := context.Background()
	onigLib := newTestOnigLib(t)

	cases := []struct {
		grammar string
		code    string
	}{
		{"match_only.json", "if x 42\nreturn 7"},
		{"begin_end.json", `hello "world" end`},
		{"begin_end_backref.json", `say "hello" and 'world'`},
	}

	for _, tc := range cases {
		t.Run(tc.grammar, func(t *testing.T) {
			g := loadMiniGrammar(t, tc.grammar)

			first, err := Tokenize(ctx, []byte(tc.code), g, onigLib, TokenizeOptions{})
			if err != nil {
				t.Fatalf("Tokenize 1: %v", err)
			}
			second, err := Tokenize(ctx, []byte(tc.code), g, onigLib, TokenizeOptions{})
			if err != nil {
				t.Fatalf("Tokenize 2: %v", err)
			}

			if !reflect.DeepEqual(first.Lines, second.Lines) {
				t.Errorf("token mismatch between runs on one instance:\nfirst:  %+v\nsecond: %+v",
					first.Lines, second.Lines)
			}
			if !reflect.DeepEqual(first.Diagnostics, second.Diagnostics) {
				t.Errorf("diagnostics mismatch: %+v vs %+v", first.Diagnostics, second.Diagnostics)
			}
		})
	}
}
