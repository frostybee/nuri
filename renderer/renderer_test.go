package renderer

import (
	"testing"

	"github.com/frostybee/nuri/theme"
)

func TestFontStyleCSS(t *testing.T) {
	tests := []struct {
		fs   theme.FontStyle
		want map[string]string
	}{
		{theme.FontStyleNone, nil},
		{theme.FontStyleNotSet, nil},
		{theme.FontStyleItalic, map[string]string{"font-style": "italic"}},
		{theme.FontStyleBold, map[string]string{"font-weight": "bold"}},
		{theme.FontStyleUnderline, map[string]string{"text-decoration": "underline"}},
		{theme.FontStyleStrikethrough, map[string]string{"text-decoration": "line-through"}},
		{theme.FontStyleItalic | theme.FontStyleBold, map[string]string{
			"font-style": "italic", "font-weight": "bold",
		}},
		{theme.FontStyleUnderline | theme.FontStyleStrikethrough, map[string]string{
			"text-decoration": "underline line-through",
		}},
	}
	for _, tt := range tests {
		got := FontStyleCSS(tt.fs)
		if len(got) != len(tt.want) {
			t.Errorf("FontStyleCSS(%v) = %v, want %v", tt.fs, got, tt.want)
			continue
		}
		for k, v := range tt.want {
			if got[k] != v {
				t.Errorf("FontStyleCSS(%v)[%q] = %q, want %q", tt.fs, k, got[k], v)
			}
		}
	}
}
