package nuri

import (
	"context"
	"strings"
	"testing"
)

func TestMultiThemeBasic(t *testing.T) {
	h := newTestHighlighter(t)
	html, err := h.CodeToHTML(context.Background(), "package main\n", CodeToHTMLOptions{
		Lang: "go",
		Themes: map[string]string{
			"dark":  "github-dark",
			"light": "github-light",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// "dark" is lexicographically first → default theme (inline color).
	// "light" → CSS variable.
	if !strings.Contains(html, `color:`) {
		t.Errorf("missing inline color for default theme:\n%s", html)
	}
	if !strings.Contains(html, `--nuri-light`) {
		t.Errorf("missing CSS variable for light theme:\n%s", html)
	}
}

func TestMultiThemePreElement(t *testing.T) {
	h := newTestHighlighter(t)
	html, err := h.CodeToHTML(context.Background(), "x\n", CodeToHTMLOptions{
		Lang: "go",
		Themes: map[string]string{
			"dark":  "github-dark",
			"light": "github-light",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, `nuri-themes`) {
		t.Errorf("pre missing nuri-themes class:\n%s", html)
	}
	if !strings.Contains(html, `dark`) {
		t.Errorf("pre missing dark theme class:\n%s", html)
	}
	if !strings.Contains(html, `light`) {
		t.Errorf("pre missing light theme class:\n%s", html)
	}
	if !strings.Contains(html, `--nuri-light-bg`) {
		t.Errorf("pre missing CSS variable for light theme background:\n%s", html)
	}
	if !strings.Contains(html, `--nuri-light:`) {
		t.Errorf("pre missing CSS variable for light theme foreground:\n%s", html)
	}
}

func TestMultiThemeDefaultColorFalse(t *testing.T) {
	h := newTestHighlighter(t)
	f := false
	html, err := h.CodeToHTML(context.Background(), "package main\n", CodeToHTMLOptions{
		Lang: "go",
		Themes: map[string]string{
			"dark":  "github-dark",
			"light": "github-light",
		},
		DefaultColor: &f,
	})
	if err != nil {
		t.Fatal(err)
	}
	// Should have CSS variables but no inline color on token spans.
	if !strings.Contains(html, `--nuri-`) {
		t.Errorf("missing CSS variables:\n%s", html)
	}
	// Check token spans don't have inline color (but pre might still not have it).
	// Extract token spans: between <code> and </code>.
	codeStart := strings.Index(html, "<code>")
	codeEnd := strings.Index(html, "</code>")
	if codeStart < 0 || codeEnd < 0 {
		t.Fatalf("missing code element:\n%s", html)
	}
	codeHTML := html[codeStart:codeEnd]
	if strings.Contains(codeHTML, `"color:`) {
		t.Errorf("token spans should not have inline color when DefaultColor=false:\n%s", codeHTML)
	}
}

func TestMultiThemeThreeThemes(t *testing.T) {
	h := newTestHighlighter(t)
	html, err := h.CodeToHTML(context.Background(), "const x = 1;\n", CodeToHTMLOptions{
		Lang: "javascript",
		Themes: map[string]string{
			"dark":    "github-dark",
			"light":   "github-light",
			"dracula": "dracula",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// "dark" is first sorted → default (inline).
	// "dracula" and "light" → CSS variables.
	if !strings.Contains(html, `--nuri-dracula`) {
		t.Errorf("missing CSS variable for dracula theme:\n%s", html)
	}
	if !strings.Contains(html, `--nuri-light`) {
		t.Errorf("missing CSS variable for light theme:\n%s", html)
	}
}

func TestMultiThemeFontStyle(t *testing.T) {
	h := newTestHighlighter(t)
	// Markdown italic should produce font-style differences across themes.
	html, err := h.CodeToHTML(context.Background(), "*italic*\n", CodeToHTMLOptions{
		Lang: "markdown",
		Themes: map[string]string{
			"dark":  "github-dark",
			"light": "github-light",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// Just verify it doesn't crash and produces valid multi-theme output.
	if !strings.Contains(html, `nuri-themes`) {
		t.Errorf("missing nuri-themes class:\n%s", html)
	}
	t.Logf("Font style multi-theme HTML:\n%s", html)
}

func TestMultiThemeDeterministic(t *testing.T) {
	h := newTestHighlighter(t)
	opts := CodeToHTMLOptions{
		Lang: "go",
		Themes: map[string]string{
			"dark":  "github-dark",
			"light": "github-light",
		},
	}
	var first string
	for i := 0; i < 100; i++ {
		html, err := h.CodeToHTML(context.Background(), "package main\n", opts)
		if err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			first = html
		} else if html != first {
			t.Fatalf("non-deterministic output at iteration %d:\n  %s\nvs\n  %s", i, html, first)
		}
	}
}

func TestMultiThemeBackwardCompat(t *testing.T) {
	h := newTestHighlighter(t)
	code := "package main\n"

	// Single-theme path (Phase 5).
	single, err := h.CodeToHTML(context.Background(), code, CodeToHTMLOptions{
		Lang:  "go",
		Theme: "github-dark",
	})
	if err != nil {
		t.Fatal(err)
	}

	// nil Themes should behave identically to single-theme.
	if strings.Contains(single, "nuri-themes") {
		t.Errorf("single-theme output should not have nuri-themes class:\n%s", single)
	}
	if strings.Contains(single, "--nuri-") {
		t.Errorf("single-theme output should not have CSS variables:\n%s", single)
	}
}
