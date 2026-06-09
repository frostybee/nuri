package renderer

import (
	"math"
	"testing"

	"github.com/frostybee/nuri/ast"
)

func TestParseHex(t *testing.T) {
	tests := []struct {
		hex        string
		r, g, b    uint8
		ok         bool
	}{
		{"#ff0000", 255, 0, 0, true},
		{"#00ff00", 0, 255, 0, true},
		{"#0000ff", 0, 0, 255, true},
		{"#f97583", 249, 117, 131, true},
		{"#000000", 0, 0, 0, true},
		{"#ffffff", 255, 255, 255, true},
		{"#f00", 255, 0, 0, true},           // short form
		{"#ff000080", 255, 0, 0, true},       // with alpha (ignored)
		{"#f008", 255, 0, 0, true},           // short with alpha (ignored)
		{"", 0, 0, 0, false},                 // empty
		{"ff0000", 0, 0, 0, false},           // missing #
		{"#gg0000", 0, 0, 0, false},          // invalid hex
		{"#ff00", 255, 255, 0, true},          // 4-char short form with alpha
	}
	for _, tt := range tests {
		r, g, b, ok := parseHex(tt.hex)
		if ok != tt.ok || r != tt.r || g != tt.g || b != tt.b {
			t.Errorf("parseHex(%q) = (%d,%d,%d,%v), want (%d,%d,%d,%v)",
				tt.hex, r, g, b, ok, tt.r, tt.g, tt.b, tt.ok)
		}
	}
}

func TestColorDistance(t *testing.T) {
	// Same color = 0.
	if d := colorDistance(255, 0, 0, 255, 0, 0); d != 0 {
		t.Errorf("same color distance = %f, want 0", d)
	}
	// Black vs white should be large.
	d := colorDistance(0, 0, 0, 255, 255, 255)
	if d < 500 {
		t.Errorf("black-white distance = %f, expected > 500", d)
	}
	// Red vs green should be > red vs dark-red.
	dRG := colorDistance(255, 0, 0, 0, 255, 0)
	dRR := colorDistance(255, 0, 0, 200, 0, 0)
	if dRR >= dRG {
		t.Errorf("red-darkred (%f) should be < red-green (%f)", dRR, dRG)
	}
}

func TestPaletteTruecolor(t *testing.T) {
	p := newPalette(ast.ColorDepthTruecolor)
	fg := p.resolveFG("#f97583")
	if fg != "38;2;249;117;131" {
		t.Errorf("truecolor FG = %q, want %q", fg, "38;2;249;117;131")
	}
	bg := p.resolveBG("#24292e")
	if bg != "48;2;36;41;46" {
		t.Errorf("truecolor BG = %q, want %q", bg, "48;2;36;41;46")
	}
}

func TestPalette256(t *testing.T) {
	p := newPalette(ast.ColorDepth256)
	// Pure red should map to index 196 (38;5;196) = rgb(255,0,0).
	fg := p.resolveFG("#ff0000")
	if fg != "38;5;196" {
		t.Errorf("256-color red FG = %q, want %q", fg, "38;5;196")
	}
	// Pure white should map to index 231 (38;5;231) = rgb(255,255,255).
	fg = p.resolveFG("#ffffff")
	if fg != "38;5;231" {
		t.Errorf("256-color white FG = %q, want %q", fg, "38;5;231")
	}
}

func TestPalette16(t *testing.T) {
	p := newPalette(ast.ColorDepth16)
	// Pure red is perceptually closest to standard red (31), not bright (91).
	fg := p.resolveFG("#ff0000")
	if fg != "31" {
		t.Errorf("16-color red FG = %q, want %q", fg, "31")
	}
	// Bright red hex should map to bright red (91).
	fg = p.resolveFG("#f14c4c")
	if fg != "91" {
		t.Errorf("16-color bright red FG = %q, want %q", fg, "91")
	}
}

func TestPalette8(t *testing.T) {
	p := newPalette(ast.ColorDepth8)
	// Pure black should map to code 30.
	fg := p.resolveFG("#000000")
	if fg != "30" {
		t.Errorf("8-color black FG = %q, want %q", fg, "30")
	}
}

func TestPaletteCache(t *testing.T) {
	p := newPalette(ast.ColorDepth256)
	fg1 := p.resolveFG("#f97583")
	fg2 := p.resolveFG("#f97583")
	if fg1 != fg2 {
		t.Errorf("cached result mismatch: %q vs %q", fg1, fg2)
	}
}

func TestPaletteEmpty(t *testing.T) {
	p := newPalette(ast.ColorDepthTruecolor)
	if s := p.resolveFG(""); s != "" {
		t.Errorf("empty hex FG = %q, want empty", s)
	}
	if s := p.resolveBG(""); s != "" {
		t.Errorf("empty hex BG = %q, want empty", s)
	}
}

func TestPaletteInvalidHex(t *testing.T) {
	p := newPalette(ast.ColorDepthTruecolor)
	if s := p.resolveFG("not-a-color"); s != "" {
		t.Errorf("invalid hex FG = %q, want empty", s)
	}
}

func TestAnsi256PaletteSize(t *testing.T) {
	if len(ansi256Palette) != 256 {
		t.Errorf("ansi256Palette has %d entries, want 256", len(ansi256Palette))
	}
}

func TestAnsi256CubeLevels(t *testing.T) {
	// Index 16 = rgb(0,0,0), the first cube entry.
	e := ansi256Palette[16]
	if e.r != 0 || e.g != 0 || e.b != 0 {
		t.Errorf("index 16 = (%d,%d,%d), want (0,0,0)", e.r, e.g, e.b)
	}
	// Index 231 = rgb(255,255,255), the last cube entry.
	e = ansi256Palette[231]
	if e.r != 255 || e.g != 255 || e.b != 255 {
		t.Errorf("index 231 = (%d,%d,%d), want (255,255,255)", e.r, e.g, e.b)
	}
}

func TestAnsi256Grayscale(t *testing.T) {
	// Index 232 = #080808.
	e := ansi256Palette[232]
	if e.r != 8 || e.g != 8 || e.b != 8 {
		t.Errorf("index 232 = (%d,%d,%d), want (8,8,8)", e.r, e.g, e.b)
	}
	// Index 255 = #eeeeee.
	e = ansi256Palette[255]
	if e.r != 238 || e.g != 238 || e.b != 238 {
		t.Errorf("index 255 = (%d,%d,%d), want (238,238,238)", e.r, e.g, e.b)
	}
}

func TestColorDistanceSymmetric(t *testing.T) {
	d1 := colorDistance(100, 50, 200, 50, 100, 150)
	d2 := colorDistance(50, 100, 150, 100, 50, 200)
	if math.Abs(d1-d2) > 0.001 {
		t.Errorf("distance not symmetric: %f vs %f", d1, d2)
	}
}
