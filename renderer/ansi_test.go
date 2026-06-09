package renderer

import (
	"strings"
	"testing"

	"github.com/frostybee/nuri/ast"
	"github.com/frostybee/nuri/theme"
)

func renderANSI(t *testing.T, result *ast.TokensResult, depth ast.ColorDepth) string {
	t.Helper()
	var buf strings.Builder
	opts := &ast.CodeToANSIOptions{ColorDepth: depth}
	if err := RenderANSI(&buf, result, opts); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func TestRenderANSITruecolor(t *testing.T) {
	result := &ast.TokensResult{
		FG: "#e1e4e8",
		BG: "#24292e",
		Tokens: [][]ast.ThemedToken{
			{
				{Content: "package", Color: "#f97583", FontStyle: theme.FontStyleNone},
				{Content: " ", Color: "#e1e4e8"},
				{Content: "main", Color: "#e1e4e8"},
			},
		},
	}
	got := renderANSI(t, result, ast.ColorDepthTruecolor)
	want := "\033[38;2;249;117;131mpackage\033[0m\033[38;2;225;228;232m \033[0m\033[38;2;225;228;232mmain\033[0m"
	if got != want {
		t.Errorf("truecolor output:\ngot:  %q\nwant: %q", got, want)
	}
}

func TestRenderANSI256(t *testing.T) {
	result := &ast.TokensResult{
		FG: "#e1e4e8",
		Tokens: [][]ast.ThemedToken{
			{
				{Content: "hello", Color: "#ff0000"},
			},
		},
	}
	got := renderANSI(t, result, ast.ColorDepth256)
	if !strings.Contains(got, "38;5;") {
		t.Errorf("256-color output should contain '38;5;': %q", got)
	}
	if !strings.Contains(got, "hello") {
		t.Errorf("output should contain 'hello': %q", got)
	}
}

func TestRenderANSIMultiLine(t *testing.T) {
	result := &ast.TokensResult{
		FG: "#e1e4e8",
		Tokens: [][]ast.ThemedToken{
			{
				{Content: "line1", Color: "#f97583"},
			},
			{
				{Content: "line2", Color: "#f97583"},
			},
		},
	}
	got := renderANSI(t, result, ast.ColorDepthTruecolor)
	if !strings.Contains(got, "\n") {
		t.Error("multi-line output should contain newline")
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(lines))
	}
}

func TestRenderANSIFontStyles(t *testing.T) {
	result := &ast.TokensResult{
		FG: "#e1e4e8",
		Tokens: [][]ast.ThemedToken{
			{
				{Content: "bold", Color: "#ff0000", FontStyle: theme.FontStyleBold},
				{Content: "italic", Color: "#ff0000", FontStyle: theme.FontStyleItalic},
				{Content: "underline", Color: "#ff0000", FontStyle: theme.FontStyleUnderline},
				{Content: "strike", Color: "#ff0000", FontStyle: theme.FontStyleStrikethrough},
				{Content: "combo", Color: "#ff0000", FontStyle: theme.FontStyleBold | theme.FontStyleItalic},
			},
		},
	}
	got := renderANSI(t, result, ast.ColorDepthTruecolor)
	if !strings.Contains(got, "\033[1;38;2;255;0;0mbold\033[0m") {
		t.Errorf("bold not rendered correctly: %q", got)
	}
	if !strings.Contains(got, "\033[3;38;2;255;0;0mitalic\033[0m") {
		t.Errorf("italic not rendered correctly: %q", got)
	}
	if !strings.Contains(got, "\033[4;38;2;255;0;0munderline\033[0m") {
		t.Errorf("underline not rendered correctly: %q", got)
	}
	if !strings.Contains(got, "\033[9;38;2;255;0;0mstrike\033[0m") {
		t.Errorf("strikethrough not rendered correctly: %q", got)
	}
	if !strings.Contains(got, "\033[1;3;38;2;255;0;0mcombo\033[0m") {
		t.Errorf("bold+italic not rendered correctly: %q", got)
	}
}

func TestRenderANSIDefaultFG(t *testing.T) {
	result := &ast.TokensResult{
		FG: "#aabbcc",
		Tokens: [][]ast.ThemedToken{
			{
				{Content: "plain", Color: ""},
			},
		},
	}
	got := renderANSI(t, result, ast.ColorDepthTruecolor)
	if !strings.Contains(got, "38;2;170;187;204") {
		t.Errorf("default FG not applied: %q", got)
	}
}

func TestRenderANSINoBgColor(t *testing.T) {
	result := &ast.TokensResult{
		FG: "#e1e4e8",
		BG: "#24292e",
		Tokens: [][]ast.ThemedToken{
			{
				{Content: "text", Color: "#ff0000", BgColor: "#00ff00"},
			},
		},
	}
	got := renderANSI(t, result, ast.ColorDepthTruecolor)
	if strings.Contains(got, "48;2;") {
		t.Errorf("background color should be stripped: %q", got)
	}
}

func TestRenderANSIEmptyInput(t *testing.T) {
	result := &ast.TokensResult{
		FG:     "#e1e4e8",
		Tokens: [][]ast.ThemedToken{},
	}
	got := renderANSI(t, result, ast.ColorDepthTruecolor)
	if got != "" {
		t.Errorf("empty input should produce empty output, got %q", got)
	}
}

func TestRenderANSIPagerSafety(t *testing.T) {
	result := &ast.TokensResult{
		FG: "#e1e4e8",
		Tokens: [][]ast.ThemedToken{
			{{Content: "a", Color: "#ff0000"}},
			{{Content: "b", Color: "#00ff00"}},
			{{Content: "c", Color: "#0000ff"}},
		},
	}
	got := renderANSI(t, result, ast.ColorDepthTruecolor)
	lines := strings.Split(got, "\n")
	for i, line := range lines {
		if line == "" {
			continue
		}
		if strings.Contains(line, "\033[") && !strings.HasSuffix(line, ansiReset) {
			t.Errorf("line %d does not end with reset: %q", i+1, line)
		}
	}
}

func TestRenderANSINewlineInToken(t *testing.T) {
	result := &ast.TokensResult{
		FG: "#e1e4e8",
		Tokens: [][]ast.ThemedToken{
			{
				{Content: "line1\nline2", Color: "#ff0000"},
			},
		},
	}
	got := renderANSI(t, result, ast.ColorDepthTruecolor)
	if !strings.Contains(got, ansiReset+"\n") {
		t.Errorf("newline in token should have reset before it: %q", got)
	}
}
