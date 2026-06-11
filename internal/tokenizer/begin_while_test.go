package tokenizer

import (
	"context"
	"testing"
)

func lineHasScope(tokens []Token, scope string) bool {
	for _, tok := range tokens {
		for _, s := range tok.Scopes {
			if s == scope {
				return true
			}
		}
	}
	return false
}

// TestBeginWhileCapturedEOL pins the BeginCapturedEOL semantics: a begin
// pattern that consumes the sentinel newline must seed anchorPosition = 0
// on continuation lines, so a while pattern starting with \G matches at
// the start of every continuation line (vscode-textmate
// tokenizeString.ts line 181 computes the flag on unclamped indices).
func TestBeginWhileCapturedEOL(t *testing.T) {
	ctx := context.Background()
	g := loadMiniGrammar(t, "begin_while_eol.json")
	onigLib := newTestOnigLib(t)

	src := ">>>\n.one\n.two\nthree"
	result, err := Tokenize(ctx, []byte(src), g, onigLib, TokenizeOptions{})
	if err != nil {
		t.Fatalf("Tokenize: %v", err)
	}
	if len(result.Lines) != 4 {
		t.Fatalf("lines: got %d, want 4", len(result.Lines))
	}
	for i, line := range result.Lines {
		for j, tok := range line {
			t.Logf("line %d [%d] %d-%d scopes=%v", i, j, tok.Start, tok.End, tok.Scopes)
		}
	}

	// Continuation lines starting with "." satisfy the \G while pattern
	// and must stay inside the block.
	if !lineHasScope(result.Lines[1], "meta.block.test") {
		t.Error("line 1 (.one): expected meta.block.test scope, while condition with \\G failed")
	}
	if !lineHasScope(result.Lines[2], "meta.block.test") {
		t.Error("line 2 (.two): expected meta.block.test scope, while condition with \\G failed")
	}

	// A line not matching the while pattern pops the block.
	if lineHasScope(result.Lines[3], "meta.block.test") {
		t.Error("line 3 (three): block should have popped, meta.block.test scope must be gone")
	}
	if !lineHasScope(result.Lines[3], "other.word.test") {
		t.Error("line 3 (three): expected other.word.test scope after block popped")
	}
}

// TestBeginWhileNoCapturedEOL is the negative case: when the begin match
// stops before the sentinel newline, BeginCapturedEOL stays false,
// anchorPosition starts at -1 on the next line, and a \G anchored while
// pattern must NOT match.
func TestBeginWhileNoCapturedEOL(t *testing.T) {
	ctx := context.Background()
	g := loadMiniGrammar(t, "begin_while_no_eol.json")
	onigLib := newTestOnigLib(t)

	src := "<<< x\n.one"
	result, err := Tokenize(ctx, []byte(src), g, onigLib, TokenizeOptions{})
	if err != nil {
		t.Fatalf("Tokenize: %v", err)
	}
	if len(result.Lines) != 2 {
		t.Fatalf("lines: got %d, want 2", len(result.Lines))
	}
	for i, line := range result.Lines {
		for j, tok := range line {
			t.Logf("line %d [%d] %d-%d scopes=%v", i, j, tok.Start, tok.End, tok.Scopes)
		}
	}

	// Begin matched mid line, so the block scope applies on line 0.
	if !lineHasScope(result.Lines[0], "meta.block.test") {
		t.Error("line 0: expected meta.block.test scope from begin match")
	}

	// Without BeginCapturedEOL the \G anchor has no anchor position on
	// the continuation line, the while check fails, and the block pops.
	if lineHasScope(result.Lines[1], "meta.block.test") {
		t.Error("line 1 (.one): block should have popped, \\G must not match without captured EOL")
	}
}
