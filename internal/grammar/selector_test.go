package grammar

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/frostybee/nuri/internal/shared"
)

func TestParseSelectorSimple(t *testing.T) {
	sel, err := ParseSelector("L:source.js")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(sel.Composites) != 1 {
		t.Fatalf("composites: got %d, want 1", len(sel.Composites))
	}
	if sel.Composites[0].Priority != PriorityLeft {
		t.Errorf("priority: got %d, want Left", sel.Composites[0].Priority)
	}
	if len(sel.Composites[0].Expressions) != 1 {
		t.Fatalf("expressions: got %d", len(sel.Composites[0].Expressions))
	}
	if sel.Composites[0].Expressions[0].Path.Scope != "source.js" {
		t.Errorf("scope: got %q", sel.Composites[0].Expressions[0].Path.Scope)
	}
}

func TestParseSelectorWithNegation(t *testing.T) {
	sel, _ := ParseSelector("L:text.html -comment.block")
	c := sel.Composites[0]
	if len(c.Expressions) != 2 {
		t.Fatalf("expressions: got %d, want 2", len(c.Expressions))
	}
	if c.Expressions[0].Negate {
		t.Error("first expression should not be negated")
	}
	if !c.Expressions[1].Negate {
		t.Error("second expression should be negated")
	}
}

func TestParseSelectorComma(t *testing.T) {
	sel, _ := ParseSelector("L:source.ts, L:source.js, L:source.coffee")
	if len(sel.Composites) != 3 {
		t.Fatalf("composites: got %d, want 3", len(sel.Composites))
	}
}

func TestParseSelectorGroup(t *testing.T) {
	sel, _ := ParseSelector("L:(meta.script.svelte | meta.style.svelte) (meta.lang.ts | meta.lang.typescript) - (meta source)")
	c := sel.Composites[0]
	if len(c.Expressions) < 2 {
		t.Fatalf("expressions: got %d, want >= 2", len(c.Expressions))
	}
	if c.Expressions[0].Group == nil {
		t.Error("first expression should be a group")
	}
}

func TestParseSelectorRightPriority(t *testing.T) {
	sel, _ := ParseSelector("R:source.python")
	if sel.Composites[0].Priority != PriorityRight {
		t.Errorf("priority: got %d, want Right", sel.Composites[0].Priority)
	}
}

func TestParseSelectorNoPriority(t *testing.T) {
	sel, _ := ParseSelector("source.python")
	if sel.Composites[0].Priority != PriorityNone {
		t.Errorf("priority: got %d, want None", sel.Composites[0].Priority)
	}
}

func TestSelectorMatchesBasic(t *testing.T) {
	sel, _ := ParseSelector("L:source.js")
	tests := []struct {
		scopes []string
		want   bool
	}{
		{[]string{"source.js"}, true},
		{[]string{"source.js", "meta.function"}, true},
		{[]string{"source.python"}, false},
		{[]string{"text.html", "source.js.embedded"}, true},
	}
	for _, tc := range tests {
		got, _ := sel.Matches(tc.scopes)
		if got != tc.want {
			t.Errorf("Matches(%v): got %v, want %v", tc.scopes, got, tc.want)
		}
	}
}

func TestSelectorMatchesNegation(t *testing.T) {
	sel, _ := ParseSelector("L:text.html -comment.block")
	tests := []struct {
		scopes []string
		want   bool
	}{
		{[]string{"text.html", "meta.tag"}, true},
		{[]string{"text.html", "comment.block"}, false},
		{[]string{"text.html.basic", "comment.block.html"}, false},
		{[]string{"text.html"}, true},
	}
	for _, tc := range tests {
		got, _ := sel.Matches(tc.scopes)
		if got != tc.want {
			t.Errorf("Matches(%v): got %v, want %v", tc.scopes, got, tc.want)
		}
	}
}

func TestSelectorMatchesGroup(t *testing.T) {
	sel, _ := ParseSelector("L:(source.ts | source.js)")
	tests := []struct {
		scopes []string
		want   bool
	}{
		{[]string{"source.ts"}, true},
		{[]string{"source.js"}, true},
		{[]string{"source.python"}, false},
	}
	for _, tc := range tests {
		got, _ := sel.Matches(tc.scopes)
		if got != tc.want {
			t.Errorf("Matches(%v): got %v, want %v", tc.scopes, got, tc.want)
		}
	}
}

func TestParseAllGrammarInjectionSelectors(t *testing.T) {
	grammars, err := filepath.Glob(filepath.Join(shared.GrammarsDir(t), "*.json"))
	if err != nil || len(grammars) == 0 {
		t.Fatal("no grammar files found")
	}

	totalSelectors := 0
	for _, path := range grammars {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		var raw struct {
			Injections map[string]json.RawMessage `json:"injections"`
		}
		if err := json.Unmarshal(data, &raw); err != nil {
			continue
		}
		for selectorStr := range raw.Injections {
			_, err := ParseSelector(selectorStr)
			if err != nil {
				t.Errorf("%s: selector %q: %v", filepath.Base(path), selectorStr, err)
			}
			totalSelectors++
		}
	}
	t.Logf("parsed %d injection selectors across all grammars", totalSelectors)
}

func TestSelectorNegatedGroupSubsequence(t *testing.T) {
	sel, err := ParseSelector("L:meta.script.svelte - meta.lang - (meta source)")
	if err != nil {
		t.Fatalf("ParseSelector: %v", err)
	}

	tests := []struct {
		scopes []string
		want   bool
	}{
		{[]string{"source.svelte", "meta.script.svelte"}, true},
		{[]string{"source.svelte", "meta.script.svelte", "meta.embedded.block.svelte", "source.js"}, false},
		{[]string{"source.svelte", "meta.script.svelte", "meta.lang.ts"}, false},
	}

	for _, tt := range tests {
		matches, _ := sel.Matches(tt.scopes)
		if matches != tt.want {
			t.Errorf("Matches(%v) = %v, want %v", tt.scopes, matches, tt.want)
		}
	}
}
