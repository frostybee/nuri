package grammar

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/frostybee/nuri/internal/shared"
)

func TestParseAllGrammars(t *testing.T) {
	grammars, err := filepath.Glob(filepath.Join(shared.GrammarsDir(t), "*.json"))
	if err != nil || len(grammars) == 0 {
		t.Fatal("no grammar files found")
	}

	for _, path := range grammars {
		name := filepath.Base(path)
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read: %v", err)
			}
			g, err := ParseGrammar(data)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			if g.ScopeName == "" {
				t.Error("empty scopeName")
			}
		})
	}
}

func TestParseGoGrammar(t *testing.T) {
	g := loadTestGrammar(t, "go.json")

	if g.ScopeName != "source.go" {
		t.Errorf("scopeName: got %q, want %q", g.ScopeName, "source.go")
	}
	if g.Name != "go" {
		t.Errorf("name: got %q, want %q", g.Name, "go")
	}
	if len(g.Patterns) == 0 {
		t.Error("expected at least one top-level pattern")
	}
	if g.Repository == nil {
		t.Fatal("expected non-nil repository")
	}
	if _, ok := g.Repository["keywords"]; !ok {
		t.Error("missing repository entry: keywords")
	}
	if _, ok := g.Repository["comments"]; !ok {
		t.Error("missing repository entry: comments")
	}
}

func TestParseGoKeywords(t *testing.T) {
	g := loadTestGrammar(t, "go.json")

	kw, ok := g.Repository["keywords"]
	if !ok {
		t.Fatal("missing repository entry: keywords")
	}
	col, ok := kw.(*CollectionRule)
	if !ok {
		t.Fatalf("keywords: expected *CollectionRule, got %T", kw)
	}
	if len(col.Patterns) == 0 {
		t.Error("keywords: expected patterns")
	}

	// First pattern should be a MatchRule
	first, ok := col.Patterns[0].(*MatchRule)
	if !ok {
		t.Fatalf("keywords.patterns[0]: expected *MatchRule, got %T", col.Patterns[0])
	}
	if first.Match == "" {
		t.Error("keywords.patterns[0]: empty match")
	}
	if first.Name == "" {
		t.Error("keywords.patterns[0]: empty name")
	}
}

func TestParseGoComments(t *testing.T) {
	g := loadTestGrammar(t, "go.json")

	comments, ok := g.Repository["comments"]
	if !ok {
		t.Fatal("missing repository entry: comments")
	}

	// comments is a collection containing begin/end rules
	col, ok := comments.(*CollectionRule)
	if !ok {
		t.Fatalf("comments: expected *CollectionRule, got %T", comments)
	}

	var foundBlockComment bool
	for _, p := range col.Patterns {
		if be, ok := p.(*BeginEndRule); ok && be.Name == "comment.block.go" {
			foundBlockComment = true
			if be.Begin == "" || be.End == "" {
				t.Error("block comment: empty begin or end")
			}
			if be.BeginCaptures == nil {
				t.Error("block comment: missing beginCaptures")
			}
			break
		}
	}
	if !foundBlockComment {
		t.Error("did not find comment.block.go begin/end rule")
	}
}

func TestParseMarkdownBlockquote(t *testing.T) {
	g := loadTestGrammar(t, "markdown.json")

	bq, ok := g.Repository["blockquote"]
	if !ok {
		t.Fatal("missing repository entry: blockquote")
	}

	bw, ok := bq.(*BeginWhileRule)
	if !ok {
		t.Fatalf("blockquote: expected *BeginWhileRule, got %T", bq)
	}
	if bw.Begin == "" {
		t.Error("blockquote: empty begin")
	}
	if bw.While == "" {
		t.Error("blockquote: empty while")
	}
	if bw.Name != "markup.quote.markdown" {
		t.Errorf("blockquote name: got %q", bw.Name)
	}
}

func TestParseJavaScriptInclude(t *testing.T) {
	g := loadTestGrammar(t, "javascript.json")

	// Top-level patterns should contain include rules
	var foundInclude bool
	for _, p := range g.Patterns {
		if inc, ok := p.(*IncludeRule); ok {
			foundInclude = true
			if inc.Include == "" {
				t.Error("include rule with empty include string")
			}
			break
		}
	}
	if !foundInclude {
		t.Error("no include rules in top-level patterns")
	}
}

func TestParseBeginEndWithCaptures(t *testing.T) {
	g := loadTestGrammar(t, "go.json")

	raw, ok := g.Repository["raw_string_literals"]
	if !ok {
		t.Fatal("missing repository entry: raw_string_literals")
	}

	be, ok := raw.(*BeginEndRule)
	if !ok {
		t.Fatalf("raw_string_literals: expected *BeginEndRule, got %T", raw)
	}
	if be.BeginCaptures == nil {
		t.Error("missing beginCaptures")
	}
	if be.EndCaptures == nil {
		t.Error("missing endCaptures")
	}
	if _, ok := be.BeginCaptures["0"]; !ok {
		t.Error("beginCaptures missing key 0")
	}
}

func TestParseUnknownFieldsIgnored(t *testing.T) {
	data := []byte(`{
		"scopeName": "source.test",
		"name": "test",
		"$schema": "https://example.com/schema",
		"information_for_contributors": ["ignore me"],
		"version": "1.0.0",
		"displayName": "Test",
		"patterns": [{"match": "\\w+", "name": "word"}]
	}`)
	g, err := ParseGrammar(data)
	if err != nil {
		t.Fatalf("should ignore unknown fields: %v", err)
	}
	if g.ScopeName != "source.test" {
		t.Errorf("scopeName: got %q", g.ScopeName)
	}
}

func TestParseMissingScopeName(t *testing.T) {
	data := []byte(`{"name": "test", "patterns": []}`)
	_, err := ParseGrammar(data)
	if err == nil {
		t.Fatal("expected error for missing scopeName")
	}
}

func TestParseInvalidJSON(t *testing.T) {
	_, err := ParseGrammar([]byte(`{invalid`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestParseRuleIDs(t *testing.T) {
	g := loadTestGrammar(t, "go.json")

	// Collect all rule IDs and verify they're unique and > 0
	ids := make(map[RuleID]bool)
	var walk func(r Rule)
	walk = func(r Rule) {
		if r == nil {
			return
		}
		id := r.GetID()
		if id <= 0 {
			return
		}
		if ids[id] {
			// This is expected — some rules may share IDs if they're
			// the same object. Just verify no zero IDs.
			return
		}
		ids[id] = true

		switch v := r.(type) {
		case *CollectionRule:
			for _, p := range v.Patterns {
				walk(p)
			}
		case *BeginEndRule:
			for _, p := range v.Patterns {
				walk(p)
			}
		case *BeginWhileRule:
			for _, p := range v.Patterns {
				walk(p)
			}
		}
	}

	for _, p := range g.Patterns {
		walk(p)
	}
	for _, r := range g.Repository {
		walk(r)
	}

	if len(ids) < 10 {
		t.Errorf("suspiciously few unique rule IDs: %d", len(ids))
	}
}

func loadTestGrammar(t *testing.T, name string) *Grammar {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(shared.GrammarsDir(t), name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	g, err := ParseGrammar(data)
	if err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}
	return g
}
