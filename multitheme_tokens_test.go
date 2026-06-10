package nuri

import (
	"context"
	"strings"
	"testing"
)

// goFixture is multi-line on purpose: it produces tokens whose content trims
// to "" (skipped by buildResultMulti), which is exactly the shape that
// desynced the old two-pass ThemeStyles construction.
const goFixture = `package main

import "fmt"

func main() {
	fmt.Println("hello", 42)
}
`

func TestCodeToTokensMultiTheme(t *testing.T) {
	h := newTestHighlighter(t)
	result, err := h.CodeToTokens(context.Background(), goFixture, CodeToTokensOptions{
		Lang: "go",
		Themes: map[string]string{
			"dark":  "github-dark",
			"light": "github-light",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(result.ThemeNames) != 2 || result.ThemeNames[0] != "dark" || result.ThemeNames[1] != "light" {
		t.Errorf("ThemeNames: got %v, want [dark light]", result.ThemeNames)
	}
	// Non-default themes get ThemeFG/ThemeBG entries; the default theme's
	// colors live in FG/BG.
	if result.ThemeFG["light"] == "" || result.ThemeBG["light"] == "" {
		t.Errorf("light ThemeFG/ThemeBG not populated: %q / %q",
			result.ThemeFG["light"], result.ThemeBG["light"])
	}
	if result.FG == "" || result.BG == "" {
		t.Errorf("default theme FG/BG not populated: %q / %q", result.FG, result.BG)
	}

	var rebuilt strings.Builder
	for _, line := range result.Tokens {
		for _, tok := range line {
			if tok.Color == "" {
				t.Errorf("token %q: missing default-theme color", tok.Content)
			}
			if _, ok := tok.ThemeStyles["light"]; !ok {
				t.Errorf("token %q: missing ThemeStyles[light]", tok.Content)
			}
			if _, ok := tok.ThemeStyles["dark"]; ok {
				t.Errorf("token %q: default theme must not appear in ThemeStyles", tok.Content)
			}
			rebuilt.WriteString(tok.Content)
		}
		rebuilt.WriteString("\n")
	}
	for i, srcLine := range strings.Split(strings.TrimRight(goFixture, "\n"), "\n") {
		gotLine := strings.Split(rebuilt.String(), "\n")[i]
		if gotLine != srcLine {
			t.Errorf("line %d content mismatch:\n got: %q\nwant: %q", i, gotLine, srcLine)
		}
	}
}

func TestCodeToTokensThemesWinsOverTheme(t *testing.T) {
	h := newTestHighlighter(t)
	result, err := h.CodeToTokens(context.Background(), "package main\n", CodeToTokensOptions{
		Lang:  "go",
		Theme: "github-light",
		Themes: map[string]string{
			"dark":  "github-dark",
			"light": "github-light",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.ThemeNames == nil {
		t.Error("Themes set alongside Theme: expected multi-theme result, got single-theme")
	}
	if result.ThemeName != "github-dark" {
		t.Errorf("default theme: got %q, want github-dark (first sorted key wins, Theme ignored)", result.ThemeName)
	}
}

func TestCodeToTokensEmptyThemesMap(t *testing.T) {
	h := newTestHighlighter(t)
	result, err := h.CodeToTokens(context.Background(), "package main\n", CodeToTokensOptions{
		Lang:   "go",
		Theme:  "github-dark",
		Themes: map[string]string{}, // empty non-nil map must not enter the multi path
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.ThemeNames != nil {
		t.Errorf("empty Themes map: expected single-theme result, got ThemeNames=%v", result.ThemeNames)
	}
	if result.ThemeName != "github-dark" {
		t.Errorf("theme: got %q, want github-dark", result.ThemeName)
	}

	// Same guard on the HTML path.
	html, err := h.CodeToHTML(context.Background(), "package main\n", CodeToHTMLOptions{
		Lang:   "go",
		Theme:  "github-dark",
		Themes: map[string]string{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(html, "nuri-themes") {
		t.Error("empty Themes map: CodeToHTML must render single-theme output")
	}
}

// TestCodeToTokensMultiThemeAlignment is the regression test for the
// ThemeStyles index-desync bug: styles must stay attached to the token whose
// scopes produced them, even when empty-content tokens are skipped. It
// cross-checks the multi-theme result token-for-token against two
// single-theme runs of the same input.
func TestCodeToTokensMultiThemeAlignment(t *testing.T) {
	h := newTestHighlighter(t)
	ctx := context.Background()

	multi, err := h.CodeToTokens(ctx, goFixture, CodeToTokensOptions{
		Lang: "go",
		Themes: map[string]string{
			"dark":  "github-dark",
			"light": "github-light",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	dark, err := h.CodeToTokens(ctx, goFixture, CodeToTokensOptions{Lang: "go", Theme: "github-dark"})
	if err != nil {
		t.Fatal(err)
	}
	light, err := h.CodeToTokens(ctx, goFixture, CodeToTokensOptions{Lang: "go", Theme: "github-light"})
	if err != nil {
		t.Fatal(err)
	}

	if len(multi.Tokens) != len(dark.Tokens) {
		t.Fatalf("line count: multi %d vs dark %d", len(multi.Tokens), len(dark.Tokens))
	}
	for i := range multi.Tokens {
		if len(multi.Tokens[i]) != len(dark.Tokens[i]) || len(multi.Tokens[i]) != len(light.Tokens[i]) {
			t.Fatalf("line %d token count: multi %d, dark %d, light %d",
				i, len(multi.Tokens[i]), len(dark.Tokens[i]), len(light.Tokens[i]))
		}
		for j, mt := range multi.Tokens[i] {
			dt, lt := dark.Tokens[i][j], light.Tokens[i][j]
			if mt.Content != dt.Content {
				t.Fatalf("line %d token %d content: multi %q vs dark %q", i, j, mt.Content, dt.Content)
			}
			// Default theme (sorted-first key "dark") fills the token's own fields.
			if mt.Color != dt.Color || mt.FontStyle != dt.FontStyle {
				t.Errorf("line %d token %d (%q): default style %s/%v, want dark run's %s/%v",
					i, j, mt.Content, mt.Color, mt.FontStyle, dt.Color, dt.FontStyle)
			}
			ls := mt.ThemeStyles["light"]
			if ls.Color != lt.Color || ls.FontStyle != lt.FontStyle {
				t.Errorf("line %d token %d (%q): ThemeStyles[light] %s/%v, want light run's %s/%v",
					i, j, mt.Content, ls.Color, ls.FontStyle, lt.Color, lt.FontStyle)
			}
		}
	}
}

func TestCodeToTokensMultiThemeANSI(t *testing.T) {
	h := newTestHighlighter(t)
	result, err := h.CodeToTokens(context.Background(), "\x1b[31mred\x1b[0m plain\n", CodeToTokensOptions{
		Lang: "ansi",
		Themes: map[string]string{
			"dark":  "github-dark",
			"light": "github-light",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.ThemeNames) != 2 {
		t.Errorf("ansi multi-theme: ThemeNames %v, want 2 entries", result.ThemeNames)
	}
	if len(result.Tokens) == 0 {
		t.Error("ansi multi-theme: no tokens")
	}
	for _, line := range result.Tokens {
		for _, tok := range line {
			ls, ok := tok.ThemeStyles["light"]
			if !ok {
				t.Fatalf("ansi token %q: missing ThemeStyles[light]", tok.Content)
			}
			if strings.Contains(tok.Content, "plain") {
				// Uncolored ANSI text takes each theme's default foreground.
				if ls.Color != result.ThemeFG["light"] {
					t.Errorf("plain ansi token: light color %s, want light default fg %s",
						ls.Color, result.ThemeFG["light"])
				}
				if tok.Color != result.FG {
					t.Errorf("plain ansi token: default color %s, want default fg %s", tok.Color, result.FG)
				}
			}
			if strings.Contains(tok.Content, "red") {
				// Explicit ANSI colors are theme-independent.
				if ls.Color != tok.Color {
					t.Errorf("red ansi token: light %s != default %s (ANSI colors must not vary per theme)",
						ls.Color, tok.Color)
				}
			}
		}
	}
}

func TestCodeToTokensMultiThemeUnknownLang(t *testing.T) {
	h := newTestHighlighter(t)
	result, err := h.CodeToTokens(context.Background(), "some plain text\n", CodeToTokensOptions{
		Lang: "definitely-not-a-language",
		Themes: map[string]string{
			"dark":  "github-dark",
			"light": "github-light",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.ThemeNames) != 2 {
		t.Errorf("plaintext multi-theme: ThemeNames %v, want 2 entries", result.ThemeNames)
	}
	found := false
	for _, d := range result.Diagnostics {
		if d.Kind == "unknown_lang" {
			found = true
		}
	}
	if !found {
		t.Error("plaintext fallback: expected unknown_lang diagnostic")
	}
	for _, line := range result.Tokens {
		for _, tok := range line {
			ls, ok := tok.ThemeStyles["light"]
			if !ok {
				t.Fatalf("plaintext token %q: missing ThemeStyles[light]", tok.Content)
			}
			if ls.Color != result.ThemeFG["light"] {
				t.Errorf("plaintext token: light color %s, want light default fg %s",
					ls.Color, result.ThemeFG["light"])
			}
			if tok.Color != result.FG {
				t.Errorf("plaintext token: default color %s, want default fg %s", tok.Color, result.FG)
			}
		}
	}
}
