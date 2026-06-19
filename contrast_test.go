package nuri

import (
	"math"
	"testing"
)

func TestParseHexColor(t *testing.T) {
	tests := []struct {
		input      string
		r, g, b    float64
		ok         bool
	}{
		{"#ffffff", 1, 1, 1, true},
		{"#000000", 0, 0, 0, true},
		{"#ff0000", 1, 0, 0, true},
		{"#fff", 1, 1, 1, true},
		{"#000", 0, 0, 0, true},
		{"#f00", 1, 0, 0, true},
		{"#ff000080", 1, 0, 0, true}, // alpha ignored
		{"#f008", 1, 0, 0, true},     // 4-char with alpha
		{"invalid", 0, 0, 0, false},
		{"", 0, 0, 0, false},
		{"#gg0000", 0, 0, 0, false},
		{"#12345", 0, 0, 0, false}, // 5 hex chars invalid
	}

	for _, tt := range tests {
		r, g, b, ok := parseHexColor(tt.input)
		if ok != tt.ok {
			t.Errorf("parseHexColor(%q): ok = %v, want %v", tt.input, ok, tt.ok)
			continue
		}
		if !ok {
			continue
		}
		if math.Abs(r-tt.r) > 0.01 || math.Abs(g-tt.g) > 0.01 || math.Abs(b-tt.b) > 0.01 {
			t.Errorf("parseHexColor(%q) = (%f, %f, %f), want (%f, %f, %f)", tt.input, r, g, b, tt.r, tt.g, tt.b)
		}
	}
}

func TestRelativeLuminance(t *testing.T) {
	tests := []struct {
		r, g, b float64
		want    float64
	}{
		{1, 1, 1, 1.0},
		{0, 0, 0, 0.0},
	}
	for _, tt := range tests {
		got := relativeLuminance(tt.r, tt.g, tt.b)
		if math.Abs(got-tt.want) > 0.001 {
			t.Errorf("relativeLuminance(%f, %f, %f) = %f, want %f", tt.r, tt.g, tt.b, got, tt.want)
		}
	}
}

func TestContrastRatio(t *testing.T) {
	tests := []struct {
		fg, bg string
		min    float64
		max    float64
	}{
		{"#000000", "#ffffff", 20.9, 21.1},
		{"#ffffff", "#000000", 20.9, 21.1}, // symmetric
		{"#ffffff", "#ffffff", 0.99, 1.01},
		{"#777777", "#ffffff", 4.4, 4.6},
	}
	for _, tt := range tests {
		got := contrastRatio(tt.fg, tt.bg)
		if got < tt.min || got > tt.max {
			t.Errorf("contrastRatio(%q, %q) = %f, want [%f, %f]", tt.fg, tt.bg, got, tt.min, tt.max)
		}
	}
}

func TestContrastRatio_InvalidInput(t *testing.T) {
	got := contrastRatio("invalid", "#ffffff")
	if got != 1 {
		t.Errorf("contrastRatio with invalid input = %f, want 1", got)
	}
}

func TestAdjustForeground_AlreadyMeetsContrast(t *testing.T) {
	result := adjustForeground("#000000", "#ffffff", 5.5)
	if result != "#000000" {
		t.Errorf("adjustForeground should return original color when contrast is met, got %q", result)
	}
}

func TestAdjustForeground_LowContrastOnLight(t *testing.T) {
	// Light yellow on white: very poor contrast
	result := adjustForeground("#f0e8b0", "#ffffff", 5.5)
	if result == "#f0e8b0" {
		t.Error("adjustForeground should have changed the low-contrast color")
	}
	ratio := contrastRatio(result, "#ffffff")
	if ratio < 5.5 {
		t.Errorf("adjusted color %q has contrast %f, want >= 5.5", result, ratio)
	}
}

func TestAdjustForeground_LowContrastOnDark(t *testing.T) {
	// Dark color on dark background
	result := adjustForeground("#1a1a2e", "#0d1117", 5.5)
	if result == "#1a1a2e" {
		t.Error("adjustForeground should have changed the low-contrast color")
	}
	ratio := contrastRatio(result, "#0d1117")
	if ratio < 5.5 {
		t.Errorf("adjusted color %q has contrast %f, want >= 5.5", result, ratio)
	}
}

func TestAdjustForeground_InvalidInput(t *testing.T) {
	result := adjustForeground("invalid", "#ffffff", 5.5)
	if result != "invalid" {
		t.Errorf("adjustForeground should return original on invalid input, got %q", result)
	}
}

func TestAdjustForeground_MinimalShift(t *testing.T) {
	// A color that barely fails contrast should be minimally adjusted
	adjusted := adjustForeground("#767676", "#ffffff", 4.5)
	ratio := contrastRatio(adjusted, "#ffffff")
	if ratio < 4.5 {
		t.Errorf("adjusted color %q has contrast %f, want >= 4.5", adjusted, ratio)
	}
}
