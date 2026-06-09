package nuri

import (
	"context"
	"strings"
	"testing"
)

func TestCodeToANSIGoGitHubDark(t *testing.T) {
	h := newTestHighlighter(t)
	out, err := h.CodeToANSI(context.Background(), "package main\n", CodeToANSIOptions{
		Lang:  "go",
		Theme: "github-dark",
	})
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(out, "package") {
		t.Error("missing 'package' token in output")
	}
	if !strings.Contains(out, "\033[") {
		t.Error("missing ANSI escape sequences")
	}
	if !strings.Contains(out, "\033[0m") {
		t.Error("missing reset sequences")
	}
	// "package" keyword should have color #f97583 = rgb(249,117,131).
	if !strings.Contains(out, "38;2;249;117;131") {
		t.Errorf("expected keyword color 38;2;249;117;131 in output:\n%s", out)
	}
}

func TestCodeToANSIAllDepths(t *testing.T) {
	h := newTestHighlighter(t)
	code := "const x = 1;\n"

	depths := []struct {
		name  string
		depth ColorDepth
		check string // substring that must appear in output
	}{
		{"truecolor", ColorDepthTruecolor, "38;2;"},
		{"256", ColorDepth256, "38;5;"},
		{"16", ColorDepth16, "\033["},
		{"8", ColorDepth8, "\033["},
	}

	for _, d := range depths {
		t.Run(d.name, func(t *testing.T) {
			out, err := h.CodeToANSI(context.Background(), code, CodeToANSIOptions{
				Lang:       "javascript",
				Theme:      "github-dark",
				ColorDepth: d.depth,
			})
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(out, d.check) {
				t.Errorf("depth %s: expected %q in output:\n%q", d.name, d.check, out)
			}
			if !strings.Contains(out, "const") {
				t.Errorf("depth %s: missing 'const' token", d.name)
			}
		})
	}
}

func TestCodeToANSIMultiLine(t *testing.T) {
	h := newTestHighlighter(t)
	out, err := h.CodeToANSI(context.Background(), "package main\n\nfunc main() {\n}\n", CodeToANSIOptions{
		Lang:  "go",
		Theme: "github-dark",
	})
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(out, "\n")
	if len(lines) < 4 {
		t.Errorf("expected at least 4 lines, got %d", len(lines))
	}
}

func TestCodeToANSIPagerSafety(t *testing.T) {
	h := newTestHighlighter(t)
	out, err := h.CodeToANSI(context.Background(), "package main\n\nfunc main() {\n}\n", CodeToANSIOptions{
		Lang:  "go",
		Theme: "github-dark",
	})
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(out, "\n")
	for i, line := range lines {
		if line == "" {
			continue
		}
		if strings.Contains(line, "\033[") && !strings.HasSuffix(line, "\033[0m") {
			t.Errorf("line %d has unclosed escape: %q", i+1, line)
		}
	}
}

func TestCodeToANSINoBgColor(t *testing.T) {
	h := newTestHighlighter(t)
	out, err := h.CodeToANSI(context.Background(), "package main\n", CodeToANSIOptions{
		Lang:  "go",
		Theme: "github-dark",
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "48;2;") {
		t.Errorf("background color should not appear in ANSI output:\n%s", out)
	}
}

func TestCodeToANSIUnknownLang(t *testing.T) {
	h := newTestHighlighter(t)
	out, err := h.CodeToANSI(context.Background(), "some text\n", CodeToANSIOptions{
		Lang:  "nonexistent-lang",
		Theme: "github-dark",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "some text") {
		t.Error("plaintext fallback should preserve content")
	}
}

func TestCodeToANSIBadTheme(t *testing.T) {
	h := newTestHighlighter(t)
	_, err := h.CodeToANSI(context.Background(), "x\n", CodeToANSIOptions{
		Lang:  "go",
		Theme: "nonexistent-theme",
	})
	if err == nil {
		t.Error("expected error for nonexistent theme")
	}
}
