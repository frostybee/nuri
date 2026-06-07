package fidelity

import (
	"testing"
)

func TestCompareThemeTokensIdentical(t *testing.T) {
	tokens := [][]FixtureToken{{
		{Start: 0, End: 5, Text: "const", Scopes: []string{"source.ts", "keyword"}, Color: "#D73A49", FontStyle: 0},
		{Start: 5, End: 6, Text: " ", Scopes: []string{"source.ts"}, Color: "#24292E", FontStyle: 0},
	}}
	diffs := CompareThemeTokens(tokens, tokens)
	if len(diffs) != 0 {
		t.Errorf("expected no diffs, got %d: %v", len(diffs), diffs)
	}
}

func TestCompareThemeTokensBoundaryMismatch(t *testing.T) {
	want := [][]FixtureToken{{
		{Start: 0, End: 5, Text: "const", Scopes: []string{"source.ts"}, Color: "#D73A49"},
	}}
	got := [][]FixtureToken{{
		{Start: 0, End: 4, Text: "cons", Scopes: []string{"source.ts"}, Color: "#D73A49"},
	}}
	diffs := CompareThemeTokens(want, got)
	if len(diffs) != 1 || diffs[0].Kind != BoundaryMismatch {
		t.Errorf("expected BoundaryMismatch, got %v", diffs)
	}
}

func TestCompareThemeTokensScopeMismatch(t *testing.T) {
	want := [][]FixtureToken{{
		{Start: 0, End: 5, Text: "const", Scopes: []string{"source.ts", "keyword.var"}, Color: "#D73A49"},
	}}
	got := [][]FixtureToken{{
		{Start: 0, End: 5, Text: "const", Scopes: []string{"source.ts", "keyword.const"}, Color: "#D73A49"},
	}}
	diffs := CompareThemeTokens(want, got)
	if len(diffs) != 1 || diffs[0].Kind != ScopeMismatch {
		t.Errorf("expected ScopeMismatch, got %v", diffs)
	}
}

func TestCompareThemeTokensStyleMismatch(t *testing.T) {
	want := [][]FixtureToken{{
		{Start: 0, End: 5, Text: "const", Scopes: []string{"source.ts"}, Color: "#D73A49", FontStyle: 0},
	}}
	got := [][]FixtureToken{{
		{Start: 0, End: 5, Text: "const", Scopes: []string{"source.ts"}, Color: "#24292E", FontStyle: 0},
	}}
	diffs := CompareThemeTokens(want, got)
	if len(diffs) != 1 || diffs[0].Kind != StyleMismatch {
		t.Errorf("expected StyleMismatch, got %v", diffs)
	}
}

func TestCompareThemeTokensMissing(t *testing.T) {
	want := [][]FixtureToken{{
		{Start: 0, End: 3, Text: "var", Scopes: []string{"a"}, Color: "#000"},
		{Start: 3, End: 6, Text: " x ", Scopes: []string{"b"}, Color: "#111"},
	}}
	got := [][]FixtureToken{{
		{Start: 0, End: 3, Text: "var", Scopes: []string{"a"}, Color: "#000"},
	}}
	diffs := CompareThemeTokens(want, got)
	if len(diffs) != 1 || diffs[0].Kind != MissingToken {
		t.Errorf("expected MissingToken, got %v", diffs)
	}
}

func TestCompareThemeTokensExtra(t *testing.T) {
	want := [][]FixtureToken{{
		{Start: 0, End: 3, Text: "var", Scopes: []string{"a"}, Color: "#000"},
	}}
	got := [][]FixtureToken{{
		{Start: 0, End: 3, Text: "var", Scopes: []string{"a"}, Color: "#000"},
		{Start: 3, End: 6, Text: " x ", Scopes: []string{"b"}, Color: "#111"},
	}}
	diffs := CompareThemeTokens(want, got)
	if len(diffs) != 1 || diffs[0].Kind != ExtraToken {
		t.Errorf("expected ExtraToken, got %v", diffs)
	}
}

func TestCompareThemeTokensLineMismatch(t *testing.T) {
	want := [][]FixtureToken{
		{{Start: 0, End: 3, Text: "foo", Scopes: []string{"a"}, Color: "#000"}},
		{{Start: 0, End: 3, Text: "bar", Scopes: []string{"a"}, Color: "#000"}},
	}
	got := [][]FixtureToken{
		{{Start: 0, End: 3, Text: "foo", Scopes: []string{"a"}, Color: "#000"}},
	}
	diffs := CompareThemeTokens(want, got)
	if len(diffs) != 1 || diffs[0].Kind != MissingToken {
		t.Errorf("expected MissingToken for missing line, got %v", diffs)
	}
}

func TestCompareHTMLIdentical(t *testing.T) {
	html := "<pre><code>hello</code></pre>"
	diff := CompareHTML(html, html)
	if diff != nil {
		t.Errorf("expected nil diff, got %v", diff)
	}
}

func TestCompareHTMLDifferent(t *testing.T) {
	want := `<pre class="shiki"><code>hello</code></pre>`
	got := `<pre class="shiki"><code>world</code></pre>`
	diff := CompareHTML(want, got)
	if diff == nil {
		t.Fatal("expected diff, got nil")
	}
	if diff.Kind != HTMLMismatch {
		t.Errorf("expected HTMLMismatch, got %v", diff.Kind)
	}
}

func TestNormalizeColor(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"#D73A49", "#d73a49"},
		{"#FFF", "#ffffff"},
		{"#abc", "#aabbcc"},
		{"  #ABC  ", "#aabbcc"},
		{"#112233", "#112233"},
	}
	for _, tt := range tests {
		got := normalizeColor(tt.in)
		if got != tt.want {
			t.Errorf("normalizeColor(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
