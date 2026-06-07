//go:build generate

package oniguruma

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/frostybee/nuri/internal/shared"
)

type grammarResult struct {
	Name               string
	Total              int
	Passed             int
	Failed             int
	Failures           []string
	BackrefEndPatterns []string
}

func TestGenerateGrammarReport(t *testing.T) {
	grammars, err := filepath.Glob(filepath.Join(shared.GrammarsDir(t), "*.json"))
	if err != nil || len(grammars) == 0 {
		t.Fatal("no grammar files found — run: git submodule update --init")
	}

	sort.Strings(grammars)
	ctx := context.Background()
	inst := newTestInstance(t)

	var results []grammarResult
	totalPatterns := 0
	totalPassed := 0
	totalFailed := 0
	totalBackref := 0

	for _, path := range grammars {
		patterns := loadGrammarPatterns(t, path)
		name := strings.TrimSuffix(filepath.Base(path), ".json")

		var failures []string
		var backrefEnds []string
		passed := 0
		for _, pat := range patterns {
			scanner, err := inst.NewScanner(ctx, [][]byte{[]byte(pat)})
			if err != nil {
				if isBackrefEndPattern(pat) {
					backrefEnds = append(backrefEnds, pat)
				} else {
					failures = append(failures, pat)
				}
				continue
			}
			scanner.Free(ctx)
			passed++
		}

		results = append(results, grammarResult{
			Name:               name,
			Total:              len(patterns),
			Passed:             passed,
			Failed:             len(failures),
			Failures:           failures,
			BackrefEndPatterns: backrefEnds,
		})

		totalPatterns += len(patterns)
		totalPassed += passed
		totalFailed += len(failures)
		totalBackref += len(backrefEnds)
	}

	var b strings.Builder

	b.WriteString("# Oniguruma WASM Scanner — Grammar Compile Report\n\n")
	b.WriteString("Compile smoke test: every regex pattern extracted from TextMate grammar\n")
	b.WriteString("files is compiled through the production Scanner API (Oniguruma WASM + wazero).\n\n")

	b.WriteString("## Summary\n\n")
	b.WriteString(fmt.Sprintf("- **Grammars tested:** %d\n", len(results)))
	b.WriteString(fmt.Sprintf("- **Total unique patterns:** %d\n", totalPatterns))
	b.WriteString(fmt.Sprintf("- **Compiled successfully:** %d", totalPassed))
	if totalPatterns > 0 {
		pct := float64(totalPassed) / float64(totalPatterns) * 100
		b.WriteString(fmt.Sprintf(" (%.1f%%)", pct))
	}
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("- **Backreference end-patterns (expected):** %d\n", totalBackref))
	b.WriteString(fmt.Sprintf("- **Genuine failures:** %d\n", totalFailed))
	b.WriteString("\n")

	if totalBackref > 0 {
		b.WriteString("> **Note on backreference end-patterns:** TextMate `begin`/`end` rules use\n")
		b.WriteString("> backreferences (`\\1`, `\\2`, ...) in the `end` pattern to refer to capture\n")
		b.WriteString("> groups from the `begin` pattern. These are resolved at runtime by the\n")
		b.WriteString("> tokenizer (which substitutes the captured text before compiling). They are\n")
		b.WriteString("> not valid standalone regexes, so Oniguruma correctly rejects them. This is\n")
		b.WriteString("> expected behavior.\n\n")
	}

	b.WriteString("## Per-Grammar Results\n\n")
	b.WriteString("| Grammar | Patterns | Passed | Backref | Failed |\n")
	b.WriteString("|---|---|---|---|---|\n")
	for _, r := range results {
		status := ""
		if r.Failed > 0 {
			status = " **!!**"
		}
		b.WriteString(fmt.Sprintf("| %s | %d | %d | %d | %d%s |\n",
			r.Name, r.Total, r.Passed, len(r.BackrefEndPatterns), r.Failed, status))
	}
	b.WriteString("\n")

	if totalFailed > 0 {
		b.WriteString("## Genuine Failures\n\n")
		for _, r := range results {
			if r.Failed == 0 {
				continue
			}
			b.WriteString(fmt.Sprintf("### %s (%d/%d failed)\n\n", r.Name, r.Failed, r.Total))
			for _, f := range r.Failures {
				b.WriteString(fmt.Sprintf("- `%s`\n", f))
			}
			b.WriteString("\n")
		}
	} else {
		b.WriteString("## Genuine Failures\n\n")
		b.WriteString("None — all patterns compiled successfully (excluding expected backreference end-patterns).\n")
	}

	if totalBackref > 0 {
		b.WriteString("\n## Backreference End-Patterns (expected, not failures)\n\n")
		b.WriteString("<details>\n<summary>")
		b.WriteString(fmt.Sprintf("%d patterns across %d grammars", totalBackref, countGrammarsWithBackrefs(results)))
		b.WriteString("</summary>\n\n")
		for _, r := range results {
			if len(r.BackrefEndPatterns) == 0 {
				continue
			}
			b.WriteString(fmt.Sprintf("**%s** (%d)\n", r.Name, len(r.BackrefEndPatterns)))
			for _, f := range r.BackrefEndPatterns {
				b.WriteString(fmt.Sprintf("- `%s`\n", f))
			}
			b.WriteString("\n")
		}
		b.WriteString("</details>\n")
	}

	report := b.String()

	if err := os.WriteFile("GRAMMAR_REPORT.md", []byte(report), 0644); err != nil {
		t.Fatalf("writing report: %v", err)
	}

	t.Logf("Report written to GRAMMAR_REPORT.md")
	t.Logf("Grammars: %d | Patterns: %d | Passed: %d | Backref: %d | Genuine failures: %d",
		len(results), totalPatterns, totalPassed, totalBackref, totalFailed)
}

func countGrammarsWithBackrefs(results []grammarResult) int {
	n := 0
	for _, r := range results {
		if len(r.BackrefEndPatterns) > 0 {
			n++
		}
	}
	return n
}
