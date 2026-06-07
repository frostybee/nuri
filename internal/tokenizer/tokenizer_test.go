package tokenizer

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/frostybee/nuri/internal/grammar"
	"github.com/frostybee/nuri/internal/oniguruma"
	"github.com/frostybee/nuri/internal/shared"
)

var testEngine *oniguruma.Engine

func TestMain(m *testing.M) {
	ctx := context.Background()
	eng, err := oniguruma.NewEngine(ctx)
	if err != nil {
		panic("NewEngine: " + err.Error())
	}
	testEngine = eng
	defer eng.Close(ctx)
	m.Run()
}

func newTestOnigLib(t *testing.T) oniguruma.OnigLib {
	t.Helper()
	ctx := context.Background()
	pool, err := oniguruma.NewPool(ctx, testEngine, 1)
	if err != nil {
		t.Fatalf("NewPool: %v", err)
	}
	inst, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	t.Cleanup(func() {
		pool.Put(inst)
		pool.Close(ctx)
	})
	return inst
}

func loadMiniGrammar(t *testing.T, name string) *grammar.Grammar {
	t.Helper()
	data, err := os.ReadFile("testdata/mini/" + name)
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	g, err := grammar.ParseGrammar(data)
	if err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}
	return g
}

func TestMatchOnly(t *testing.T) {
	ctx := context.Background()
	g := loadMiniGrammar(t, "match_only.json")
	onigLib := newTestOnigLib(t)

	result, err := Tokenize(ctx, []byte("if x 42"), g, onigLib)
	if err != nil {
		t.Fatalf("Tokenize: %v", err)
	}
	if len(result.Lines) != 1 {
		t.Fatalf("lines: got %d, want 1", len(result.Lines))
	}

	tokens := result.Lines[0]
	t.Logf("tokens: %d", len(tokens))
	for i, tok := range tokens {
		t.Logf("  [%d] %d-%d scopes=%v", i, tok.Start, tok.End, tok.Scopes)
	}

	// Should have tokens for "if", " ", "x", " ", "42"
	assertTokenScope(t, tokens, "keyword.control.test", "if")
	assertTokenScope(t, tokens, "variable.other.test", "x")
	assertTokenScope(t, tokens, "constant.numeric.test", "42")
}

func TestBeginEnd(t *testing.T) {
	ctx := context.Background()
	g := loadMiniGrammar(t, "begin_end.json")
	onigLib := newTestOnigLib(t)

	result, err := Tokenize(ctx, []byte(`hello "world" end`), g, onigLib)
	if err != nil {
		t.Fatalf("Tokenize: %v", err)
	}

	tokens := result.Lines[0]
	t.Logf("tokens: %d", len(tokens))
	for i, tok := range tokens {
		t.Logf("  [%d] %d-%d scopes=%v", i, tok.Start, tok.End, tok.Scopes)
	}

	// Should see the string scope on "world" content
	var foundString bool
	for _, tok := range tokens {
		for _, s := range tok.Scopes {
			if s == "string.quoted.double.test" {
				foundString = true
			}
		}
	}
	if !foundString {
		t.Error("expected to find string.quoted.double.test scope")
	}
}

func TestEmptyInput(t *testing.T) {
	ctx := context.Background()
	g := loadMiniGrammar(t, "match_only.json")
	onigLib := newTestOnigLib(t)

	result, err := Tokenize(ctx, []byte(""), g, onigLib)
	if err != nil {
		t.Fatalf("Tokenize: %v", err)
	}
	if len(result.Lines) != 0 {
		t.Errorf("lines: got %d, want 0", len(result.Lines))
	}
}

func TestMultiLine(t *testing.T) {
	ctx := context.Background()
	g := loadMiniGrammar(t, "match_only.json")
	onigLib := newTestOnigLib(t)

	result, err := Tokenize(ctx, []byte("if x\nreturn 42"), g, onigLib)
	if err != nil {
		t.Fatalf("Tokenize: %v", err)
	}
	if len(result.Lines) != 2 {
		t.Fatalf("lines: got %d, want 2", len(result.Lines))
	}

	// Line 1 should have "if" and "x"
	assertTokenScope(t, result.Lines[0], "keyword.control.test", "if")
	// Line 2 should have "return" and "42"
	assertTokenScope(t, result.Lines[1], "keyword.control.test", "return")
	assertTokenScope(t, result.Lines[1], "constant.numeric.test", "42")
}

func TestRealGoGrammar(t *testing.T) {
	ctx := context.Background()

	data, err := os.ReadFile(filepath.Join(shared.GrammarsDir(t), "go.json"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	g, err := grammar.ParseGrammar(data)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	onigLib := newTestOnigLib(t)

	result, err := Tokenize(ctx, []byte("package main\n"), g, onigLib)
	if err != nil {
		t.Fatalf("Tokenize: %v", err)
	}
	if len(result.Lines) == 0 {
		t.Fatal("no lines")
	}

	tokens := result.Lines[0]
	t.Logf("tokens: %d", len(tokens))
	for i, tok := range tokens {
		t.Logf("  [%d] %d-%d scopes=%v", i, tok.Start, tok.End, tok.Scopes)
	}

	if len(tokens) == 0 {
		t.Fatal("no tokens")
	}
}

func TestGAnchor(t *testing.T) {
	ctx := context.Background()
	g := loadMiniGrammar(t, "g_anchor.json")
	onigLib := newTestOnigLib(t)

	result, err := Tokenize(ctx, []byte("- hello world"), g, onigLib)
	if err != nil {
		t.Fatalf("Tokenize: %v", err)
	}

	tokens := result.Lines[0]
	dumpTokens(t, tokens)

	// The \G-anchored pattern should match "hello" (first word after begin)
	// but "world" should get variable.other.test (not first-word)
	assertTokenScope(t, tokens, "entity.name.first-word.test", "hello")
}

func TestBeginEndBackref(t *testing.T) {
	ctx := context.Background()
	g := loadMiniGrammar(t, "begin_end_backref.json")
	onigLib := newTestOnigLib(t)

	result, err := Tokenize(ctx, []byte(`say "hello" and 'world'`), g, onigLib)
	if err != nil {
		t.Fatalf("Tokenize: %v", err)
	}

	tokens := result.Lines[0]
	dumpTokens(t, tokens)

	// Both double-quoted and single-quoted strings should be recognized
	var foundString int
	for _, tok := range tokens {
		for _, s := range tok.Scopes {
			if s == "string.quoted.test" {
				foundString++
				break
			}
		}
	}
	if foundString < 2 {
		t.Errorf("expected at least 2 tokens with string.quoted.test scope, got %d", foundString)
	}
}

func TestSelfInjection(t *testing.T) {
	ctx := context.Background()
	g := loadMiniGrammar(t, "injection_self.json")
	onigLib := newTestOnigLib(t)

	result, err := Tokenize(ctx, []byte("if x # comment"), g, onigLib)
	if err != nil {
		t.Fatalf("Tokenize: %v", err)
	}
	if len(result.Lines) != 1 {
		t.Fatalf("lines: got %d, want 1", len(result.Lines))
	}

	tokens := result.Lines[0]
	dumpTokens(t, tokens)

	assertTokenScope(t, tokens, "keyword.test", "if")
	assertTokenScope(t, tokens, "comment.line.test", "# comment")
}

func TestPickBestMatch(t *testing.T) {
	makeMatch := func(start int) *matchResult {
		return &matchResult{
			match: &oniguruma.Match{
				Captures: []oniguruma.Capture{{Start: start, End: start + 2}},
			},
			rule: &grammar.MatchRule{},
		}
	}

	tests := []struct {
		name        string
		grammar     *matchResult
		injection   *matchResult
		priority    grammar.Priority
		wantNil     bool
		wantFromInj bool
	}{
		{"both nil", nil, nil, grammar.PriorityNone, true, false},
		{"grammar only", makeMatch(0), nil, grammar.PriorityNone, false, false},
		{"injection only", nil, makeMatch(0), grammar.PriorityNone, false, true},
		{"injection earlier", makeMatch(5), makeMatch(2), grammar.PriorityNone, false, true},
		{"grammar earlier", makeMatch(2), makeMatch(5), grammar.PriorityNone, false, false},
		{"tie grammar wins default", makeMatch(3), makeMatch(3), grammar.PriorityNone, false, false},
		{"tie grammar wins R:", makeMatch(3), makeMatch(3), grammar.PriorityRight, false, false},
		{"tie injection wins L:", makeMatch(3), makeMatch(3), grammar.PriorityLeft, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pickBestMatch(tt.grammar, tt.injection, tt.priority)
			if tt.wantNil {
				if got != nil {
					t.Fatal("expected nil")
				}
				return
			}
			if got == nil {
				t.Fatal("unexpected nil")
			}
			if tt.wantFromInj && got != tt.injection {
				t.Error("expected injection match to win")
			}
			if !tt.wantFromInj && got != tt.grammar {
				t.Error("expected grammar match to win")
			}
		})
	}
}

func TestInjectionPriorityLeft(t *testing.T) {
	ctx := context.Background()
	g := loadMiniGrammar(t, "injection_priority_left.json")
	onigLib := newTestOnigLib(t)

	result, err := Tokenize(ctx, []byte("hello"), g, onigLib)
	if err != nil {
		t.Fatalf("Tokenize: %v", err)
	}

	tokens := result.Lines[0]
	dumpTokens(t, tokens)

	// L: injection wins ties — "hello" should get injected.word.test
	assertTokenScope(t, tokens, "injected.word.test", "hello")
}

func TestInjectionPriorityRight(t *testing.T) {
	ctx := context.Background()
	g := loadMiniGrammar(t, "injection_priority_right.json")
	onigLib := newTestOnigLib(t)

	result, err := Tokenize(ctx, []byte("hello"), g, onigLib)
	if err != nil {
		t.Fatalf("Tokenize: %v", err)
	}

	tokens := result.Lines[0]
	dumpTokens(t, tokens)

	// R: grammar wins ties — "hello" should get word.test, NOT injected.word.test
	assertTokenScope(t, tokens, "word.test", "hello")
	assertNoTokenScope(t, tokens, "injected.word.test")
}

func TestInjectionNegation(t *testing.T) {
	ctx := context.Background()
	g := loadMiniGrammar(t, "injection_negation.json")
	onigLib := newTestOnigLib(t)

	result, err := Tokenize(ctx, []byte("if 42 // 99"), g, onigLib)
	if err != nil {
		t.Fatalf("Tokenize: %v", err)
	}

	tokens := result.Lines[0]
	dumpTokens(t, tokens)

	// 42 outside comment → injected.number.test
	assertTokenScope(t, tokens, "injected.number.test", "42")
	// 99 inside comment → NOT injected (negated scope -comment)
	for _, tok := range tokens {
		hasComment := false
		hasInjected := false
		for _, s := range tok.Scopes {
			if strings.Contains(s, "comment") {
				hasComment = true
			}
			if s == "injected.number.test" {
				hasInjected = true
			}
		}
		if hasComment && hasInjected {
			t.Error("injection should not fire inside comment scope")
		}
	}
}

func TestInjectionBeginEnd(t *testing.T) {
	ctx := context.Background()
	g := loadMiniGrammar(t, "injection_begin_end.json")
	onigLib := newTestOnigLib(t)

	result, err := Tokenize(ctx, []byte("hello { 42 }"), g, onigLib)
	if err != nil {
		t.Fatalf("Tokenize: %v", err)
	}

	tokens := result.Lines[0]
	dumpTokens(t, tokens)

	assertTokenScope(t, tokens, "block.injected.test", "{")
	assertTokenScope(t, tokens, "number.injected.test", "42")
}

func TestInjectionBeginEndMultiLine(t *testing.T) {
	ctx := context.Background()
	g := loadMiniGrammar(t, "injection_begin_end.json")
	onigLib := newTestOnigLib(t)

	result, err := Tokenize(ctx, []byte("hello {\n42\n}"), g, onigLib)
	if err != nil {
		t.Fatalf("Tokenize: %v", err)
	}

	if len(result.Lines) != 3 {
		t.Fatalf("lines: got %d, want 3", len(result.Lines))
	}

	// Line 2 should have 42 with number.injected.test inside block.injected.test
	tokens := result.Lines[1]
	dumpTokens(t, tokens)
	assertTokenScope(t, tokens, "number.injected.test", "42")
}

func assertNoTokenScope(t *testing.T, tokens []Token, scope string) {
	t.Helper()
	for _, tok := range tokens {
		for _, s := range tok.Scopes {
			if s == scope {
				t.Errorf("unexpected scope %q found", scope)
				return
			}
		}
	}
}

func assertTokenScope(t *testing.T, tokens []Token, scopeSuffix, textContent string) {
	t.Helper()
	for _, tok := range tokens {
		for _, s := range tok.Scopes {
			if strings.HasSuffix(s, scopeSuffix) || s == scopeSuffix {
				return
			}
		}
	}
	t.Errorf("no token with scope %q found (looking for text %q)", scopeSuffix, textContent)
}
