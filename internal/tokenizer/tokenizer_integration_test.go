package tokenizer

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/frostybee/nuri/internal/grammar"
	"github.com/frostybee/nuri/internal/registry"
	"github.com/frostybee/nuri/internal/shared"
)

func loadRealGrammar(t *testing.T, name string) *grammar.Grammar {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(shared.GrammarsDir(t), name+".json"))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	g, err := grammar.ParseGrammar(data)
	if err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}
	return g
}

func TestGoPackageDecl(t *testing.T) {
	ctx := context.Background()
	g := loadRealGrammar(t, "go")
	onigLib := newTestOnigLib(t)

	src := []byte("package main\n")
	result, err := Tokenize(ctx, src, g, onigLib)
	if err != nil {
		t.Fatalf("Tokenize: %v", err)
	}

	tokens := result.Lines[0]
	dumpTokens(t, tokens)

	assertHasScope(t, tokens, "keyword.package.go")
	assertHasScope(t, tokens, "entity.name.type.package.go")
	assertLineReconstruction(t, src, tokens)
	assertNonOverlapping(t, tokens)
	assertTokenText(t, src, tokens, "keyword.package.go", "package")
}

func TestGoFuncDecl(t *testing.T) {
	ctx := context.Background()
	g := loadRealGrammar(t, "go")
	onigLib := newTestOnigLib(t)

	src := []byte("func main() {\n}\n")
	result, err := Tokenize(ctx, src, g, onigLib)
	if err != nil {
		t.Fatalf("Tokenize: %v", err)
	}

	tokens := result.Lines[0]
	dumpTokens(t, tokens)

	assertHasScope(t, tokens, "keyword.function.go")
	assertLineReconstruction(t, src, tokens)
	assertNonOverlapping(t, tokens)
	assertTokenText(t, src, tokens, "keyword.function.go", "func")
}

func TestGoVarDecl(t *testing.T) {
	ctx := context.Background()
	g := loadRealGrammar(t, "go")
	onigLib := newTestOnigLib(t)

	src := []byte("var x int = 42\n")
	result, err := Tokenize(ctx, src, g, onigLib)
	if err != nil {
		t.Fatalf("Tokenize: %v", err)
	}

	tokens := result.Lines[0]
	dumpTokens(t, tokens)

	assertHasScope(t, tokens, "keyword.var.go")
	assertLineReconstruction(t, src, tokens)
	assertNonOverlapping(t, tokens)
	assertTokenText(t, src, tokens, "keyword.var.go", "var")
}

func TestJavaScriptConst(t *testing.T) {
	ctx := context.Background()
	g := loadRealGrammar(t, "javascript")
	onigLib := newTestOnigLib(t)

	src := []byte("const x = 1;\n")
	result, err := Tokenize(ctx, src, g, onigLib)
	if err != nil {
		t.Fatalf("Tokenize: %v", err)
	}

	tokens := result.Lines[0]
	dumpTokens(t, tokens)

	assertHasScope(t, tokens, "storage.type.js")
	assertLineReconstruction(t, src, tokens)
	assertNonOverlapping(t, tokens)
	assertTokenText(t, src, tokens, "storage.type.js", "const")
}

func TestJavaScriptConstPI(t *testing.T) {
	ctx := context.Background()
	g := loadRealGrammar(t, "javascript")
	onigLib := newTestOnigLib(t)

	src := []byte("const PI = 3.14159;\n")
	result, err := Tokenize(ctx, src, g, onigLib)
	if err != nil {
		t.Fatalf("Tokenize: %v", err)
	}

	tokens := result.Lines[0]
	dumpTokens(t, tokens)

	assertHasScope(t, tokens, "storage.type.js")
	assertLineReconstruction(t, src, tokens)
	assertNonOverlapping(t, tokens)
	assertTokenText(t, src, tokens, "storage.type.js", "const")
}

func TestOverlappingCaptures(t *testing.T) {
	ctx := context.Background()
	g := loadMiniGrammar(t, "overlapping_captures.json")
	onigLib := newTestOnigLib(t)

	// Test 1: function declaration with captures 1, 2, 3 inside capture 0
	src := []byte("fn main(x)")
	result, err := Tokenize(ctx, src, g, onigLib)
	if err != nil {
		t.Fatalf("Tokenize: %v", err)
	}

	tokens := result.Lines[0]
	dumpTokens(t, tokens)

	assertLineReconstruction(t, src, tokens)
	assertNonOverlapping(t, tokens)
	assertTokenText(t, src, tokens, "keyword.function.test", "fn")
	assertTokenText(t, src, tokens, "entity.name.function.test", "main")
	assertTokenText(t, src, tokens, "meta.parameters.test", "(x)")

	// Test 2: assignment with captures 0 (full), 1, 2, 3
	src2 := []byte("let count = total")
	result2, err := Tokenize(ctx, src2, g, onigLib)
	if err != nil {
		t.Fatalf("Tokenize: %v", err)
	}

	tokens2 := result2.Lines[0]
	dumpTokens(t, tokens2)

	assertLineReconstruction(t, src2, tokens2)
	assertNonOverlapping(t, tokens2)
	assertTokenText(t, src2, tokens2, "storage.type.test", "let")
	assertTokenText(t, src2, tokens2, "variable.name.test", "count")
	assertTokenText(t, src2, tokens2, "variable.value.test", "total")
}

func TestNonASCIIByteOffsets(t *testing.T) {
	ctx := context.Background()
	g := loadRealGrammar(t, "go")
	onigLib := newTestOnigLib(t)

	// "変数" is 6 bytes (2 × 3-byte CJK), then " := 42"
	src := []byte("変数 := 42\n")
	result, err := Tokenize(ctx, src, g, onigLib)
	if err != nil {
		t.Fatalf("Tokenize: %v", err)
	}

	tokens := result.Lines[0]
	dumpTokens(t, tokens)

	assertLineReconstruction(t, src, tokens)
	assertNonOverlapping(t, tokens)
}

func TestEmojiByteOffsets(t *testing.T) {
	ctx := context.Background()
	g := loadMiniGrammar(t, "match_only.json")
	onigLib := newTestOnigLib(t)

	src := []byte("👋 42")
	result, err := Tokenize(ctx, src, g, onigLib)
	if err != nil {
		t.Fatalf("Tokenize: %v", err)
	}

	tokens := result.Lines[0]
	dumpTokens(t, tokens)

	assertLineReconstruction(t, src, tokens)
	assertNonOverlapping(t, tokens)
	assertTokenScope(t, tokens, "constant.numeric.test", "42")
}

func TestHTMLEscapingOffsets(t *testing.T) {
	ctx := context.Background()
	g := loadMiniGrammar(t, "match_only.json")
	onigLib := newTestOnigLib(t)

	src := []byte("</script> &amp; 42")
	result, err := Tokenize(ctx, src, g, onigLib)
	if err != nil {
		t.Fatalf("Tokenize: %v", err)
	}

	tokens := result.Lines[0]
	dumpTokens(t, tokens)

	assertLineReconstruction(t, src, tokens)
	assertNonOverlapping(t, tokens)
	assertTokenScope(t, tokens, "constant.numeric.test", "42")
}

func assertHasScope(t *testing.T, tokens []Token, scope string) {
	t.Helper()
	for _, tok := range tokens {
		for _, s := range tok.Scopes {
			if s == scope {
				return
			}
		}
	}
	t.Errorf("no token with scope %q", scope)
}

func newTestResolver(t *testing.T) *registry.Repository {
	t.Helper()
	repo, err := registry.NewRepository(os.DirFS(shared.GrammarsDir(t)))
	if err != nil {
		t.Fatalf("NewRepository: %v", err)
	}
	return repo
}

func TestHTMLCrossGrammarInclude(t *testing.T) {
	ctx := context.Background()
	repo := newTestResolver(t)
	g, err := repo.GetByScope("text.html.derivative")
	if err != nil {
		t.Fatalf("get html grammar: %v", err)
	}
	onigLib := newTestOnigLib(t)

	src := []byte("<div></div>\n")
	result, err := Tokenize(ctx, src, g, onigLib, repo)
	if err != nil {
		t.Fatalf("Tokenize: %v", err)
	}

	tokens := result.Lines[0]
	dumpTokens(t, tokens)

	assertHasScope(t, tokens, "entity.name.tag.html")
	assertLineReconstruction(t, src, tokens)
	assertNonOverlapping(t, tokens)
}

func TestHTMLEmbeddedCSS(t *testing.T) {
	t.Skip("requires html-derivative style tag resolution via text.html.basic repository chain — tracked for Phase 7 fidelity")
}

func TestHTMLEmbeddedJS(t *testing.T) {
	t.Skip("requires html-derivative script tag resolution via text.html.basic repository chain — tracked for Phase 7 fidelity")
}

func dumpTokens(t *testing.T, tokens []Token) {
	t.Helper()
	for i, tok := range tokens {
		t.Logf("  [%d] %d-%d scopes=%v", i, tok.Start, tok.End, tok.Scopes)
	}
}

// assertLineReconstruction verifies that concatenating token text exactly
// reproduces the covered portion of the source (no gaps, no overlaps).
// The source must be the full line as passed to the tokenizer (may include \n).
func assertLineReconstruction(t *testing.T, source []byte, tokens []Token) {
	t.Helper()
	if len(tokens) == 0 {
		return
	}
	start := tokens[0].Start
	end := tokens[len(tokens)-1].End
	if end > len(source) {
		t.Errorf("last token end %d exceeds source length %d", end, len(source))
		return
	}
	var reconstructed []byte
	for _, tok := range tokens {
		if tok.Start < 0 || tok.End > len(source) || tok.Start > tok.End {
			t.Errorf("invalid token offset [%d:%d] for source of length %d", tok.Start, tok.End, len(source))
			return
		}
		reconstructed = append(reconstructed, source[tok.Start:tok.End]...)
	}
	want := string(source[start:end])
	if string(reconstructed) != want {
		t.Errorf("line reconstruction failed:\n  want: %q\n  got:  %q", want, string(reconstructed))
	}
}

// assertNonOverlapping verifies that tokens are contiguous and non-overlapping:
// each token's Start must equal the previous token's End.
func assertNonOverlapping(t *testing.T, tokens []Token) {
	t.Helper()
	for i := 1; i < len(tokens); i++ {
		if tokens[i].Start != tokens[i-1].End {
			t.Errorf("token gap/overlap at [%d]-[%d]: prev.End=%d, cur.Start=%d",
				i-1, i, tokens[i-1].End, tokens[i].Start)
		}
	}
}

// assertTokenText finds the first token containing the given scope and
// asserts that its text content matches wantText.
func assertTokenText(t *testing.T, source []byte, tokens []Token, scope, wantText string) {
	t.Helper()
	for _, tok := range tokens {
		for _, s := range tok.Scopes {
			if s == scope {
				gotText := string(source[tok.Start:tok.End])
				if gotText != wantText {
					t.Errorf("token with scope %q: want text %q, got %q", scope, wantText, gotText)
				}
				return
			}
		}
	}
	t.Errorf("no token with scope %q found (looking for text %q)", scope, wantText)
}
