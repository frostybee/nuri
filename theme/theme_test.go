package theme

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/frostybee/nuri/internal/shared"
)

func loadTestTheme(t testing.TB, name string) *Theme {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(shared.ThemesDir(t), name+".json"))
	if err != nil {
		t.Fatalf("load theme %s: %v", name, err)
	}
	th, err := Parse(data)
	if err != nil {
		t.Fatalf("parse theme %s: %v", name, err)
	}
	return th
}

func TestParseGitHubDark(t *testing.T) {
	th := loadTestTheme(t, "github-dark")

	if th.Name != "github-dark" {
		t.Errorf("Name = %q, want %q", th.Name, "github-dark")
	}
	if th.DisplayName != "GitHub Dark" {
		t.Errorf("DisplayName = %q, want %q", th.DisplayName, "GitHub Dark")
	}
	if th.Type != "dark" {
		t.Errorf("Type = %q, want %q", th.Type, "dark")
	}
	if len(th.Colors) == 0 {
		t.Fatal("Colors map is empty")
	}
	if len(th.TokenColors) == 0 {
		t.Fatal("TokenColors is empty")
	}

	// Spot-check a known rule: "keyword" → #f97583
	found := false
	for _, tc := range th.TokenColors {
		for _, s := range tc.Scopes {
			if s == "keyword" && tc.Settings.Foreground == "#f97583" {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected keyword → #f97583 rule")
	}
}

func TestParseGitHubLight(t *testing.T) {
	th := loadTestTheme(t, "github-light")
	if th.Name != "github-light" {
		t.Errorf("Name = %q, want %q", th.Name, "github-light")
	}
	if th.Type != "light" {
		t.Errorf("Type = %q, want %q", th.Type, "light")
	}
	if len(th.TokenColors) == 0 {
		t.Fatal("TokenColors is empty")
	}
}

func TestParseDracula(t *testing.T) {
	th := loadTestTheme(t, "dracula")
	if th.Name == "" {
		t.Error("Name is empty")
	}
	if len(th.TokenColors) == 0 {
		t.Fatal("TokenColors is empty")
	}
}

func TestBase(t *testing.T) {
	th := loadTestTheme(t, "github-dark")
	base := th.Base()

	if base.Foreground != "#e1e4e8" {
		t.Errorf("Base().Foreground = %q, want %q", base.Foreground, "#e1e4e8")
	}
	if base.Background != "#24292e" {
		t.Errorf("Base().Background = %q, want %q", base.Background, "#24292e")
	}
	if base.FontStyle != FontStyleNone {
		t.Errorf("Base().FontStyle = %v, want FontStyleNone", base.FontStyle)
	}
}

func TestParseScopeString(t *testing.T) {
	data := []byte(`{
		"name": "test",
		"tokenColors": [
			{"scope": "keyword", "settings": {"foreground": "#ff0000"}}
		]
	}`)
	th, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(th.TokenColors) != 1 {
		t.Fatalf("got %d rules, want 1", len(th.TokenColors))
	}
	if th.TokenColors[0].Scopes[0] != "keyword" {
		t.Errorf("scope = %q, want %q", th.TokenColors[0].Scopes[0], "keyword")
	}
}

func TestParseScopeArray(t *testing.T) {
	data := []byte(`{
		"name": "test",
		"tokenColors": [
			{"scope": ["keyword", "storage"], "settings": {"foreground": "#ff0000"}}
		]
	}`)
	th, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(th.TokenColors[0].Scopes) != 2 {
		t.Fatalf("got %d scopes, want 2", len(th.TokenColors[0].Scopes))
	}
}

func TestParseScopeCommaSeparated(t *testing.T) {
	data := []byte(`{
		"name": "test",
		"tokenColors": [
			{"scope": "keyword, storage.type", "settings": {"foreground": "#ff0000"}}
		]
	}`)
	th, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	scopes := th.TokenColors[0].Scopes
	if len(scopes) != 2 || scopes[0] != "keyword" || scopes[1] != "storage.type" {
		t.Errorf("scopes = %v, want [keyword storage.type]", scopes)
	}
}

func TestParseScopelessEntrySkipped(t *testing.T) {
	data := []byte(`{
		"name": "test",
		"tokenColors": [
			{"settings": {"foreground": "#e1e4e8"}},
			{"scope": "keyword", "settings": {"foreground": "#ff0000"}}
		]
	}`)
	th, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(th.TokenColors) != 1 {
		t.Errorf("got %d rules, want 1 (scopeless entry should be skipped)", len(th.TokenColors))
	}
}

func TestParseFontStyle(t *testing.T) {
	tests := []struct {
		input *string
		want  FontStyle
	}{
		{nil, FontStyleNotSet},
		{strptr(""), FontStyleNone},
		{strptr("italic"), FontStyleItalic},
		{strptr("bold"), FontStyleBold},
		{strptr("underline"), FontStyleUnderline},
		{strptr("strikethrough"), FontStyleStrikethrough},
		{strptr("italic bold"), FontStyleItalic | FontStyleBold},
		{strptr("italic underline"), FontStyleItalic | FontStyleUnderline},
	}
	for _, tt := range tests {
		got := parseFontStyle(tt.input)
		label := "<nil>"
		if tt.input != nil {
			label = `"` + *tt.input + `"`
		}
		if got != tt.want {
			t.Errorf("parseFontStyle(%s) = %v, want %v", label, got, tt.want)
		}
	}
}

func TestParseSemanticHighlightingIgnored(t *testing.T) {
	data := []byte(`{
		"name": "test",
		"semanticHighlighting": true,
		"semanticTokenColors": {"keyword": "#ff0000"},
		"tokenColors": [
			{"scope": "keyword", "settings": {"foreground": "#ff0000"}}
		]
	}`)
	_, err := Parse(data)
	if err != nil {
		t.Fatalf("should ignore unknown fields, got: %v", err)
	}
}

func TestNormalizeDefaultsFromColors(t *testing.T) {
	data := []byte(`{
		"name": "test",
		"colors": {
			"editor.foreground": "#aabbcc",
			"editor.background": "#112233"
		},
		"tokenColors": []
	}`)
	th, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if th.DefaultForeground != "#aabbcc" {
		t.Errorf("DefaultForeground = %q, want %q", th.DefaultForeground, "#aabbcc")
	}
	if th.DefaultBackground != "#112233" {
		t.Errorf("DefaultBackground = %q, want %q", th.DefaultBackground, "#112233")
	}
}

func TestNormalizeInvalidColorFallback(t *testing.T) {
	data := []byte(`{
		"name": "test",
		"colors": {
			"editor.foreground": "not-a-color",
			"editor.background": "#GGG"
		},
		"tokenColors": []
	}`)
	th, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if th.DefaultForeground != "#000000" {
		t.Errorf("DefaultForeground = %q, want #000000", th.DefaultForeground)
	}
	if th.DefaultBackground != "#000000" {
		t.Errorf("DefaultBackground = %q, want #000000", th.DefaultBackground)
	}
}

func TestFontStyleString(t *testing.T) {
	tests := []struct {
		fs   FontStyle
		want string
	}{
		{FontStyleNotSet, "notset"},
		{FontStyleNone, "none"},
		{FontStyleItalic, "italic"},
		{FontStyleBold, "bold"},
		{FontStyleItalic | FontStyleBold, "italic bold"},
		{FontStyleItalic | FontStyleUnderline, "italic underline"},
		{FontStyleItalic | FontStyleBold | FontStyleStrikethrough, "italic bold strikethrough"},
	}
	for _, tt := range tests {
		got := tt.fs.String()
		if got != tt.want {
			t.Errorf("FontStyle(%d).String() = %q, want %q", tt.fs, got, tt.want)
		}
	}
}

func strptr(s string) *string { return &s }
