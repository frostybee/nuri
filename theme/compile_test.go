package theme

import (
	"reflect"
	"strings"
	"sync"
	"testing"
)

// TestScoreCompiledEquivalence pins the compiled path to the per-call
// parsing path over a matrix of selector and stack edge cases.
func TestScoreCompiledEquivalence(t *testing.T) {
	selectors := []string{
		"",
		"   ",
		"\t\n",
		"keyword",
		"keyword.control",
		"source.go keyword",
		"source.go meta.function entity.name",
		" padded ",
		"a\tb",
		"keyword.",
		".",
		"notfound",
		"entity.name source.go", // wrong order — must not match
	}
	stacks := [][]string{
		nil,
		{},
		{"source.go"},
		{"source.go", "meta.function.declaration.go", "entity.name.function.go"},
		{"text.html.markdown", "markup.fenced_code.block.markdown", "source.js",
			"meta.function.js", "meta.definition.function.js", "entity.name.function.js"},
	}

	for _, sel := range selectors {
		for _, stack := range stacks {
			c := compileSelector(sel)
			gotScore, gotOK := scoreCompiled(c.parts, c.scopeDepth, stack)
			wantScore, wantOK := scoreSelector(sel, stack)
			if gotOK != wantOK || gotScore != wantScore {
				t.Errorf("selector %q stack %v: compiled=(%+v,%v) fallback=(%+v,%v)",
					sel, stack, gotScore, gotOK, wantScore, wantOK)
			}
		}
	}
}

// TestParsePopulatesCompiled verifies Parse pre-compiles every selector.
func TestParsePopulatesCompiled(t *testing.T) {
	data := []byte(`{
		"name": "test",
		"tokenColors": [
			{"scope": "keyword, source.go string", "settings": {"foreground": "#aaaaaa"}},
			{"scope": ["entity.name.function", "meta.function entity"], "settings": {"foreground": "#bbbbbb"}}
		]
	}`)
	th, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	for ri, rule := range th.TokenColors {
		if len(rule.compiled) != len(rule.Scopes) {
			t.Fatalf("rule %d: compiled len %d != scopes len %d", ri, len(rule.compiled), len(rule.Scopes))
		}
		for i, sel := range rule.Scopes {
			c := rule.compiled[i]
			if c.source != sel {
				t.Errorf("rule %d sel %d: source %q != scope %q", ri, i, c.source, sel)
			}
			if !reflect.DeepEqual(c.parts, strings.Fields(sel)) {
				t.Errorf("rule %d sel %d: parts %v != Fields(%q)", ri, i, c.parts, sel)
			}
			wantDepth := 0
			if parts := strings.Fields(sel); len(parts) > 0 {
				wantDepth = strings.Count(parts[len(parts)-1], ".") + 1
			}
			if c.scopeDepth != wantDepth {
				t.Errorf("rule %d sel %d: scopeDepth %d, want %d", ri, i, c.scopeDepth, wantDepth)
			}
		}
	}
}

// TestHandBuiltThemeFallback verifies a Theme constructed without Parse
// (compiled == nil) matches identically to its Parse-built equivalent.
func TestHandBuiltThemeFallback(t *testing.T) {
	data := []byte(`{
		"name": "test",
		"tokenColors": [
			{"scope": "keyword", "settings": {"foreground": "#aaaaaa"}},
			{"scope": "source.go keyword.control", "settings": {"foreground": "#bbbbbb"}},
			{"scope": "string", "settings": {"foreground": "#cccccc", "fontStyle": "italic"}}
		]
	}`)
	parsed, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	handBuilt := &Theme{
		TokenColors: []TokenColor{
			{Scopes: []string{"keyword"}, Settings: TokenSettings{Foreground: "#aaaaaa", FontStyle: FontStyleNotSet}},
			{Scopes: []string{"source.go keyword.control"}, Settings: TokenSettings{Foreground: "#bbbbbb", FontStyle: FontStyleNotSet}},
			{Scopes: []string{"string"}, Settings: TokenSettings{Foreground: "#cccccc", FontStyle: FontStyleItalic}},
			{Scopes: nil, Settings: TokenSettings{Foreground: "#dddddd", FontStyle: FontStyleNotSet}},
		},
	}

	stacks := [][]string{
		{"source.go", "keyword.control.go"},
		{"source.go", "string.quoted.double.go"},
		{"source.js", "keyword.operator.js"},
		{"source.go"},
	}
	for _, stack := range stacks {
		got := handBuilt.Match(stack)
		want := parsed.Match(stack)
		if got != want {
			t.Errorf("stack %v: hand-built %+v != parsed %+v", stack, got, want)
		}
	}
}

// TestInPlaceScopeMutation verifies the per-selector source guard: editing
// Scopes after Parse must take effect rather than serving the stale
// pre-compiled selector.
func TestInPlaceScopeMutation(t *testing.T) {
	data := []byte(`{
		"name": "test",
		"tokenColors": [
			{"scope": "keyword", "settings": {"foreground": "#aaaaaa"}}
		]
	}`)
	th, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	if got := th.Match([]string{"string.quoted"}); got.Foreground == "#aaaaaa" {
		t.Fatal("string.quoted should not match selector keyword")
	}

	th.TokenColors[0].Scopes[0] = "string"

	if got := th.Match([]string{"string.quoted"}); got.Foreground != "#aaaaaa" {
		t.Errorf("after mutation, string.quoted should match: got %+v", got)
	}
	if got := th.Match([]string{"keyword.control"}); got.Foreground == "#aaaaaa" {
		t.Errorf("after mutation, keyword.control should no longer match: got %+v", got)
	}
}

// TestMatchConcurrent is a race-detector smoke test: Match on a shared
// parsed theme must be safe from many goroutines (compiled data is written
// only inside Parse, before the theme escapes).
func TestMatchConcurrent(t *testing.T) {
	data := []byte(`{
		"name": "test",
		"tokenColors": [
			{"scope": "keyword", "settings": {"foreground": "#aaaaaa"}},
			{"scope": "source.go string", "settings": {"foreground": "#bbbbbb"}}
		]
	}`)
	th, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	stacks := [][]string{
		{"source.go", "keyword.control.go"},
		{"source.go", "string.quoted.double.go"},
	}
	want := []TokenSettings{th.Match(stacks[0]), th.Match(stacks[1])}

	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 500; i++ {
				k := i % len(stacks)
				if got := th.Match(stacks[k]); got != want[k] {
					t.Errorf("concurrent Match mismatch: got %+v, want %+v", got, want[k])
					return
				}
			}
		}()
	}
	wg.Wait()
}
