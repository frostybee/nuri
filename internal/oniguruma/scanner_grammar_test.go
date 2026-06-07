package oniguruma

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/frostybee/nuri/internal/shared"
)

var regexFields = map[string]bool{
	"match": true,
	"begin": true,
	"end":   true,
	"while": true,
}

func extractAllPatterns(grammar any) []string {
	seen := make(map[string]bool)
	var result []string

	var walk func(v any)
	walk = func(v any) {
		switch val := v.(type) {
		case map[string]any:
			for key, child := range val {
				if regexFields[key] {
					if s, ok := child.(string); ok && s != "" && !seen[s] {
						seen[s] = true
						result = append(result, s)
					}
				}
				walk(child)
			}
		case []any:
			for _, item := range val {
				walk(item)
			}
		}
	}

	walk(grammar)
	return result
}

func loadGrammarPatterns(t *testing.T, path string) []string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	var grammar any
	if err := json.Unmarshal(data, &grammar); err != nil {
		t.Fatalf("parsing %s: %v", path, err)
	}
	return extractAllPatterns(grammar)
}

var backrefPattern = regexp.MustCompile(`\\[1-9]`)

func isBackrefEndPattern(pat string) bool {
	return backrefPattern.MatchString(pat)
}

func TestGrammarPatternCount(t *testing.T) {
	grammars, err := filepath.Glob(filepath.Join(shared.GrammarsDir(t), "*.json"))
	if err != nil || len(grammars) == 0 {
		t.Fatal("no grammar files found — run: git submodule update --init")
	}

	total := 0
	for _, path := range grammars {
		patterns := loadGrammarPatterns(t, path)
		name := filepath.Base(path)
		t.Logf("%-30s %d patterns", name, len(patterns))
		total += len(patterns)
	}
	t.Logf("%-30s %d patterns", "TOTAL", total)

	if total < 100 {
		t.Errorf("suspiciously low pattern count (%d) — extraction may be broken", total)
	}
}

func TestGrammarCompile(t *testing.T) {
	grammars, err := filepath.Glob(filepath.Join(shared.GrammarsDir(t), "*.json"))
	if err != nil || len(grammars) == 0 {
		t.Fatal("no grammar files found — run: git submodule update --init")
	}

	ctx := context.Background()

	for _, path := range grammars {
		patterns := loadGrammarPatterns(t, path)
		name := filepath.Base(path)

		t.Run(name, func(t *testing.T) {
			inst := newTestInstance(t)
			var failed []string
			var backrefs []string

			for _, pat := range patterns {
				scanner, err := inst.NewScanner(ctx, [][]byte{[]byte(pat)})
				if err != nil {
					if isBackrefEndPattern(pat) {
						backrefs = append(backrefs, pat)
					} else {
						failed = append(failed, pat)
					}
					continue
				}
				scanner.Free(ctx)
			}

			if len(failed) > 0 {
				t.Errorf("%d/%d genuine failures (plus %d backref end-patterns):",
					len(failed), len(patterns), len(backrefs))
				for i, f := range failed {
					if i >= 10 {
						t.Errorf("  ... and %d more", len(failed)-10)
						break
					}
					t.Errorf("  %s", f)
				}
			} else {
				t.Logf("%d/%d compiled (%d backref end-patterns expected to fail)",
					len(patterns)-len(backrefs), len(patterns), len(backrefs))
			}
		})
	}
}
