package nuri

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/frostybee/nuri/internal/shared"
)

func newTestHighlighter(t *testing.T) *Highlighter {
	t.Helper()
	ctx := context.Background()
	h, err := New(ctx,
		WithGrammarFS(os.DirFS(shared.GrammarsDir(t))),
		WithThemeFS(os.DirFS(shared.ThemesDir(t))),
		WithPoolSize(1),
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { h.Close(ctx) })
	return h
}

func newTestHighlighterWithOpts(t *testing.T, extra ...Option) *Highlighter {
	t.Helper()
	ctx := context.Background()
	opts := append([]Option{
		WithGrammarFS(os.DirFS(shared.GrammarsDir(t))),
		WithThemeFS(os.DirFS(shared.ThemesDir(t))),
		WithPoolSize(1),
	}, extra...)
	h, err := New(ctx, opts...)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { h.Close(ctx) })
	return h
}

func TestCodeToTokensGoGitHubDark(t *testing.T) {
	h := newTestHighlighter(t)
	result, err := h.CodeToTokens(context.Background(), "package main\n", CodeToTokensOptions{
		Lang:  "go",
		Theme: "github-dark",
	})
	if err != nil {
		t.Fatal(err)
	}

	if result.FG != "#e1e4e8" {
		t.Errorf("FG = %q, want #e1e4e8", result.FG)
	}
	if result.BG != "#24292e" {
		t.Errorf("BG = %q, want #24292e", result.BG)
	}
	if result.ThemeName != "github-dark" {
		t.Errorf("ThemeName = %q, want github-dark", result.ThemeName)
	}
	if len(result.Tokens) == 0 {
		t.Fatal("no lines")
	}

	found := false
	for _, tok := range result.Tokens[0] {
		if tok.Content == "package" {
			found = true
			if tok.Color != "#f97583" {
				t.Errorf("package color = %q, want #f97583", tok.Color)
			}
		}
	}
	if !found {
		t.Error("no token with content \"package\"")
		for i, tok := range result.Tokens[0] {
			t.Logf("  [%d] %q color=%s", i, tok.Content, tok.Color)
		}
	}
}

func TestCodeToTokensGoString(t *testing.T) {
	h := newTestHighlighter(t)
	result, err := h.CodeToTokens(context.Background(), "var s = \"hello\"\n", CodeToTokensOptions{
		Lang:  "go",
		Theme: "github-dark",
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, tok := range result.Tokens[0] {
		if tok.Content == `"hello"` || tok.Content == "hello" {
			if tok.Color != "#9ecbff" {
				t.Errorf("string token %q color = %q, want #9ecbff", tok.Content, tok.Color)
			}
			return
		}
	}
	t.Log("String token not found in expected form, dumping all tokens:")
	for i, tok := range result.Tokens[0] {
		t.Logf("  [%d] %q color=%s", i, tok.Content, tok.Color)
	}
}

func TestCodeToTokensJavaScript(t *testing.T) {
	h := newTestHighlighter(t)
	result, err := h.CodeToTokens(context.Background(), "const x = 1;\n", CodeToTokensOptions{
		Lang:  "javascript",
		Theme: "github-dark",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Tokens) == 0 {
		t.Fatal("no lines")
	}

	for _, tok := range result.Tokens[0] {
		if tok.Content == "const" {
			if tok.Color != "#f97583" {
				t.Errorf("const color = %q, want #f97583", tok.Color)
			}
			return
		}
	}
	t.Error("no token with content \"const\"")
	for i, tok := range result.Tokens[0] {
		t.Logf("  [%d] %q color=%s", i, tok.Content, tok.Color)
	}
}

func TestCodeToTokensUnknownLang(t *testing.T) {
	h := newTestHighlighter(t)
	result, err := h.CodeToTokens(context.Background(), "hello world\n", CodeToTokensOptions{
		Lang:  "not-a-real-language-xyz",
		Theme: "github-dark",
	})
	if err != nil {
		t.Fatalf("expected no error for unknown lang, got: %v", err)
	}
	if len(result.Diagnostics) == 0 {
		t.Error("expected diagnostic for unknown lang")
	}
	if len(result.Tokens) == 0 {
		t.Fatal("expected plaintext tokens")
	}
	if result.Tokens[0][0].Content != "hello world" {
		t.Errorf("content = %q, want \"hello world\"", result.Tokens[0][0].Content)
	}
}

func TestCodeToTokensUnknownTheme(t *testing.T) {
	h := newTestHighlighter(t)
	_, err := h.CodeToTokens(context.Background(), "x\n", CodeToTokensOptions{
		Lang:  "go",
		Theme: "nonexistent-theme",
	})
	if err == nil {
		t.Fatal("expected error for unknown theme")
	}
	if !errors.Is(err, ErrThemeNotFound) {
		t.Errorf("error = %v, want ErrThemeNotFound", err)
	}
}

func TestCodeToTokensMultiLine(t *testing.T) {
	h := newTestHighlighter(t)
	code := "package main\n\nfunc main() {\n}\n"
	result, err := h.CodeToTokens(context.Background(), code, CodeToTokensOptions{
		Lang:  "go",
		Theme: "github-dark",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Tokens) < 3 {
		t.Fatalf("expected at least 3 lines, got %d", len(result.Tokens))
	}

	// Line 3 should have "func" keyword
	for _, tok := range result.Tokens[2] {
		if tok.Content == "func" {
			if tok.Color != "#f97583" {
				t.Errorf("func color = %q, want #f97583", tok.Color)
			}
			return
		}
	}
	t.Error("no 'func' token on line 3")
	for i, tok := range result.Tokens[2] {
		t.Logf("  [%d] %q color=%s", i, tok.Content, tok.Color)
	}
}

func TestCodeToTokensEmptyInput(t *testing.T) {
	h := newTestHighlighter(t)
	result, err := h.CodeToTokens(context.Background(), "", CodeToTokensOptions{
		Lang:  "go",
		Theme: "github-dark",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Tokens) != 0 {
		t.Errorf("expected 0 lines for empty input, got %d", len(result.Tokens))
	}
}

func TestHighlighterCloseIdempotent(t *testing.T) {
	ctx := context.Background()
	h, err := New(ctx,
		WithGrammarFS(os.DirFS(shared.GrammarsDir(t))),
		WithThemeFS(os.DirFS(shared.ThemesDir(t))),
		WithPoolSize(1),
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := h.Close(ctx); err != nil {
		t.Errorf("first Close: %v", err)
	}
	if err := h.Close(ctx); err != nil {
		t.Errorf("second Close: %v", err)
	}
}

func TestRegisterAlias(t *testing.T) {
	h := newTestHighlighter(t)
	h.RegisterAlias("golang", "go")

	result, err := h.CodeToTokens(context.Background(), "package main\n", CodeToTokensOptions{
		Lang:  "golang",
		Theme: "github-dark",
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, tok := range result.Tokens[0] {
		if tok.Content == "package" {
			if tok.Color != "#f97583" {
				t.Errorf("package color = %q, want #f97583", tok.Color)
			}
			return
		}
	}
	t.Error("alias 'golang' did not resolve to Go grammar")
}
