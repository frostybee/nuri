package ansi

import (
	"testing"

	"github.com/frostybee/nuri/theme"
)

func TestBasicForeground(t *testing.T) {
	lines := Tokenize("\x1b[31mred\x1b[0m")
	if len(lines) != 1 {
		t.Fatalf("lines: got %d, want 1", len(lines))
	}
	if len(lines[0]) != 1 {
		t.Fatalf("tokens: got %d, want 1", len(lines[0]))
	}
	tok := lines[0][0]
	if tok.Content != "red" {
		t.Errorf("content: got %q, want %q", tok.Content, "red")
	}
	if tok.Style.FG != "#cd3131" {
		t.Errorf("FG: got %q, want %q", tok.Style.FG, "#cd3131")
	}
}

func TestBoldAndColor(t *testing.T) {
	lines := Tokenize("\x1b[1;32mbold green\x1b[0m")
	if len(lines) != 1 || len(lines[0]) != 1 {
		t.Fatalf("expected 1 line, 1 token")
	}
	tok := lines[0][0]
	if tok.Content != "bold green" {
		t.Errorf("content: got %q", tok.Content)
	}
	if tok.Style.FG != "#0dbc79" {
		t.Errorf("FG: got %q, want #0dbc79", tok.Style.FG)
	}
	if !tok.Style.FontStyle.Has(theme.FontStyleBold) {
		t.Error("expected bold font style")
	}
}

func TestBrightForeground(t *testing.T) {
	lines := Tokenize("\x1b[91mbright red\x1b[0m")
	if len(lines) != 1 || len(lines[0]) != 1 {
		t.Fatal("expected 1 line, 1 token")
	}
	if lines[0][0].Style.FG != "#f14c4c" {
		t.Errorf("FG: got %q, want #f14c4c", lines[0][0].Style.FG)
	}
}

func TestBackground(t *testing.T) {
	lines := Tokenize("\x1b[41mred bg\x1b[0m")
	if len(lines) != 1 || len(lines[0]) != 1 {
		t.Fatal("expected 1 line, 1 token")
	}
	if lines[0][0].Style.BG != "#cd3131" {
		t.Errorf("BG: got %q, want #cd3131", lines[0][0].Style.BG)
	}
}

func TestColor256(t *testing.T) {
	lines := Tokenize("\x1b[38;5;196mred\x1b[0m")
	if len(lines) != 1 || len(lines[0]) != 1 {
		t.Fatal("expected 1 line, 1 token")
	}
	if lines[0][0].Style.FG != "#ff0000" {
		t.Errorf("FG: got %q, want #ff0000", lines[0][0].Style.FG)
	}
}

func TestColor256Grayscale(t *testing.T) {
	lines := Tokenize("\x1b[38;5;240mgray\x1b[0m")
	if len(lines) != 1 || len(lines[0]) != 1 {
		t.Fatal("expected 1 line, 1 token")
	}
	// Index 240 = (240-232)*10+8 = 88 → #585858
	if lines[0][0].Style.FG != "#585858" {
		t.Errorf("FG: got %q, want #585858", lines[0][0].Style.FG)
	}
}

func TestTruecolor(t *testing.T) {
	lines := Tokenize("\x1b[38;2;255;128;0morange\x1b[0m")
	if len(lines) != 1 || len(lines[0]) != 1 {
		t.Fatal("expected 1 line, 1 token")
	}
	if lines[0][0].Style.FG != "#ff8000" {
		t.Errorf("FG: got %q, want #ff8000", lines[0][0].Style.FG)
	}
}

func TestStateAcrossLines(t *testing.T) {
	lines := Tokenize("\x1b[31mline1\nstill red\x1b[0m")
	if len(lines) != 2 {
		t.Fatalf("lines: got %d, want 2", len(lines))
	}
	if lines[0][0].Style.FG != "#cd3131" {
		t.Errorf("line 0 FG: got %q, want #cd3131", lines[0][0].Style.FG)
	}
	if lines[1][0].Style.FG != "#cd3131" {
		t.Errorf("line 1 FG: got %q, want #cd3131 (state should carry)", lines[1][0].Style.FG)
	}
}

func TestReset(t *testing.T) {
	lines := Tokenize("\x1b[31mred\x1b[0mnormal")
	if len(lines) != 1 {
		t.Fatalf("lines: got %d, want 1", len(lines))
	}
	if len(lines[0]) != 2 {
		t.Fatalf("tokens: got %d, want 2", len(lines[0]))
	}
	if lines[0][0].Style.FG != "#cd3131" {
		t.Errorf("token 0 FG: got %q, want #cd3131", lines[0][0].Style.FG)
	}
	if lines[0][1].Style.FG != "" {
		t.Errorf("token 1 FG: got %q, want empty (reset)", lines[0][1].Style.FG)
	}
	if lines[0][1].Content != "normal" {
		t.Errorf("token 1 content: got %q, want %q", lines[0][1].Content, "normal")
	}
}

func TestNoEscapes(t *testing.T) {
	lines := Tokenize("plain text")
	if len(lines) != 1 || len(lines[0]) != 1 {
		t.Fatal("expected 1 line, 1 token")
	}
	if lines[0][0].Content != "plain text" {
		t.Errorf("content: got %q", lines[0][0].Content)
	}
	if lines[0][0].Style.FG != "" || lines[0][0].Style.BG != "" {
		t.Error("expected no colors on plain text")
	}
}

func TestInvalidCodes(t *testing.T) {
	lines := Tokenize("\x1b[999mtext")
	if len(lines) != 1 || len(lines[0]) != 1 {
		t.Fatal("expected 1 line, 1 token")
	}
	if lines[0][0].Content != "text" {
		t.Errorf("content: got %q, want %q", lines[0][0].Content, "text")
	}
}

func TestMixed(t *testing.T) {
	lines := Tokenize("hello \x1b[32mworld\x1b[0m!")
	if len(lines) != 1 {
		t.Fatalf("lines: got %d, want 1", len(lines))
	}
	if len(lines[0]) != 3 {
		t.Fatalf("tokens: got %d, want 3", len(lines[0]))
	}
	if lines[0][0].Content != "hello " || lines[0][0].Style.FG != "" {
		t.Errorf("token 0: got %q FG=%q", lines[0][0].Content, lines[0][0].Style.FG)
	}
	if lines[0][1].Content != "world" || lines[0][1].Style.FG != "#0dbc79" {
		t.Errorf("token 1: got %q FG=%q", lines[0][1].Content, lines[0][1].Style.FG)
	}
	if lines[0][2].Content != "!" || lines[0][2].Style.FG != "" {
		t.Errorf("token 2: got %q FG=%q", lines[0][2].Content, lines[0][2].Style.FG)
	}
}

func TestItalicUnderlineStrikethrough(t *testing.T) {
	lines := Tokenize("\x1b[3;4;9mstyled\x1b[0m")
	if len(lines) != 1 || len(lines[0]) != 1 {
		t.Fatal("expected 1 line, 1 token")
	}
	fs := lines[0][0].Style.FontStyle
	if !fs.Has(theme.FontStyleItalic) {
		t.Error("expected italic")
	}
	if !fs.Has(theme.FontStyleUnderline) {
		t.Error("expected underline")
	}
	if !fs.Has(theme.FontStyleStrikethrough) {
		t.Error("expected strikethrough")
	}
}

func TestEmptyInput(t *testing.T) {
	lines := Tokenize("")
	if len(lines) != 0 {
		t.Errorf("expected 0 lines for empty input, got %d", len(lines))
	}
}

func TestOnlyEscapes(t *testing.T) {
	lines := Tokenize("\x1b[31m\x1b[0m")
	if len(lines) != 0 {
		t.Errorf("expected 0 lines for escape-only input, got %d", len(lines))
	}
}

func TestMultipleLines(t *testing.T) {
	lines := Tokenize("line1\nline2\nline3")
	if len(lines) != 3 {
		t.Fatalf("lines: got %d, want 3", len(lines))
	}
	for i, want := range []string{"line1", "line2", "line3"} {
		if len(lines[i]) != 1 || lines[i][0].Content != want {
			t.Errorf("line %d: got %q, want %q", i, lines[i][0].Content, want)
		}
	}
}

func TestTruecolorBackground(t *testing.T) {
	lines := Tokenize("\x1b[48;2;0;128;255mblue bg\x1b[0m")
	if len(lines) != 1 || len(lines[0]) != 1 {
		t.Fatal("expected 1 line, 1 token")
	}
	if lines[0][0].Style.BG != "#0080ff" {
		t.Errorf("BG: got %q, want #0080ff", lines[0][0].Style.BG)
	}
}
