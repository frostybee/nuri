package fidelity_test

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"io/fs"

	"github.com/frostybee/nuri"
	"github.com/frostybee/nuri/ast"
	"github.com/frostybee/nuri/bundle/core"
	"github.com/frostybee/nuri/bundle/full"
	"github.com/frostybee/nuri/internal/fidelity"
	"github.com/frostybee/nuri/internal/shared"
	"github.com/frostybee/nuri/theme"
)

var (
	goldenDir            = shared.FixtureGoldenDir
	goldenFullDir        = shared.FixtureGoldenFullDir
	goldenThemeStressDir = shared.FixtureGoldenThemeStressDir
	goldenAllDir         = shared.FixtureGoldenAllDir
)

// runGoldenSuite loads fixtures from dir using core.FS(), runs nuri over each, and compares.
func runGoldenSuite(t *testing.T, dir string) (*fidelity.FidelityReport, []string) {
	return runGoldenSuiteWithFS(t, dir, core.FS())
}

// runGoldenSuiteWithFS loads fixtures from dir, runs nuri over each, and compares.
// Returns the fidelity report and list of themes found.
func runGoldenSuiteWithFS(t *testing.T, dir string, fsys fs.FS) (*fidelity.FidelityReport, []string) {
	t.Helper()

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Skipf("no fixtures in %s — run tools/genfixtures first", dir)
	}

	fixtures, err := fidelity.LoadFixtures(dir)
	if err != nil {
		t.Fatalf("LoadFixtures(%s): %v", dir, err)
	}
	if len(fixtures) == 0 {
		t.Skipf("no fixtures found in %s", dir)
	}

	ctx := context.Background()
	h, err := nuri.New(ctx, nuri.WithFS(fsys))
	if err != nil {
		t.Fatalf("nuri.New: %v", err)
	}
	defer h.Close(ctx)

	var results []fidelity.TripleResult
	var themes []string
	themeSet := make(map[string]bool)

	for _, fix := range fixtures {
		for themeName, tf := range fix.Themes {
			if !themeSet[themeName] {
				themeSet[themeName] = true
				themes = append(themes, themeName)
			}
			t.Run(fix.Grammar+"/"+themeName, func(t *testing.T) {
				tr := runTriple(t, ctx, h, fix, themeName, tf)
				results = append(results, tr)
			})
		}
	}

	slices.Sort(themes)
	report := fidelity.ComputeReport(results)
	t.Logf("Fidelity (%s): %d/%d pass (%.1f%%)", dir, report.GlobalPass, report.GlobalTotal, globalRate(report)*100)
	return report, themes
}

func TestGoldenFidelity(t *testing.T) {
	runGoldenSuite(t, goldenDir)
}

func TestGoldenFidelityFull(t *testing.T) {
	runGoldenSuite(t, goldenFullDir)
}

func TestGoldenFidelityThemeStress(t *testing.T) {
	runGoldenSuite(t, goldenThemeStressDir)
}

func TestGoldenFidelityAll(t *testing.T) {
	runGoldenSuiteWithFS(t, goldenAllDir, full.FS())
}

// auxiliaryGrammars are grammars included via $include from parent grammars.
// They have no standalone samples and are not tested independently.
var auxiliaryGrammars = map[string]bool{
	"cpp-macro":                       true,
	"html-derivative":                 true,
	"vue-directives":                  true,
	"vue-html":                        true,
	"vue-interpolations":              true,
	"vue-sfc-style-variable-injection": true,
}

func TestCoreOnlyShipsGreen(t *testing.T) {
	report, _ := runGoldenSuite(t, goldenFullDir)

	coreGrammarsDir := filepath.Join(shared.RepoRoot(), shared.BundleCoreGrammars)
	entries, err := os.ReadDir(coreGrammarsDir)
	if err != nil {
		t.Fatalf("read bundle/core/grammars: %v", err)
	}

	var held []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		grammar := strings.TrimSuffix(e.Name(), ".json")
		if auxiliaryGrammars[grammar] {
			continue
		}
		score, ok := report.ByGrammar[grammar]
		if !ok {
			t.Errorf("core grammar %q has no fidelity data — missing from fixture matrix", grammar)
			continue
		}
		if score.Rate() < 1.0 {
			held = append(held, grammar)
		}
	}

	if len(held) > 0 {
		t.Errorf("core bundle contains held grammars (fidelity < 100%%): %v", held)
	}
}

func runTriple(t *testing.T, ctx context.Context, h *nuri.Highlighter, fix *fidelity.Fixture, themeName string, tf fidelity.ThemeFixture) fidelity.TripleResult {
	t.Helper()
	result := fidelity.TripleResult{
		Grammar: fix.Grammar,
		Theme:   themeName,
		Source:   fix.Grammar,
	}

	tokResult, err := h.CodeToTokens(ctx, fix.Source, ast.CodeToTokensOptions{
		Lang:  fix.Grammar,
		Theme: themeName,
	})
	if err != nil {
		t.Errorf("CodeToTokens: %v", err)
		result.Diffs = []fidelity.TokenDiff{{Kind: fidelity.MissingToken, Detail: err.Error()}}
		return result
	}

	gotTokens := themedToFixture(tokResult.Tokens)
	diffs := fidelity.CompareThemeTokens(tf.Tokens, gotTokens)
	if len(diffs) > 0 {
		for _, d := range diffs {
			t.Errorf("token diff: %s", d)
		}
	}

	htmlResult, err := h.CodeToHTML(ctx, fix.Source, ast.CodeToHTMLOptions{
		Lang:  fix.Grammar,
		Theme: themeName,
	})
	if err != nil {
		t.Errorf("CodeToHTML: %v", err)
		diffs = append(diffs, fidelity.TokenDiff{Kind: fidelity.HTMLMismatch, Detail: err.Error()})
	} else if tf.HTML != "" {
		if hd := fidelity.CompareHTML(tf.HTML, htmlResult); hd != nil {
			t.Errorf("HTML diff: %s", hd)
			diffs = append(diffs, *hd)
		}
	}

	result.Diffs = diffs
	result.Pass = len(diffs) == 0
	return result
}

func themedToFixture(tokens [][]ast.ThemedToken) [][]fidelity.FixtureToken {
	result := make([][]fidelity.FixtureToken, len(tokens))
	for i, line := range tokens {
		result[i] = make([]fidelity.FixtureToken, len(line))
		offset := 0
		for j, tok := range line {
			end := offset + len(tok.Content)
			result[i][j] = fidelity.FixtureToken{
				Start:     offset,
				End:       end,
				Text:      tok.Content,
				Scopes:    tok.Scopes,
				Color:     tok.Color,
				FontStyle: int(tok.FontStyle),
			}
			offset = end
		}
	}
	return result
}

func globalRate(r *fidelity.FidelityReport) float64 {
	if r.GlobalTotal == 0 {
		return 0
	}
	return float64(r.GlobalPass) / float64(r.GlobalTotal)
}

func TestFidelityReport(t *testing.T) {
	if _, err := os.Stat(goldenDir); os.IsNotExist(err) {
		t.Skip("no golden fixtures")
	}

	fixtures, err := fidelity.LoadFixtures(goldenDir)
	if err != nil {
		t.Fatalf("LoadFixtures: %v", err)
	}

	ctx := context.Background()
	h, err := nuri.New(ctx, nuri.WithFS(core.FS()))
	if err != nil {
		t.Fatalf("nuri.New: %v", err)
	}
	defer h.Close(ctx)

	var results []fidelity.TripleResult
	var themes []string
	themeSet := make(map[string]bool)

	for _, fix := range fixtures {
		for themeName, tf := range fix.Themes {
			if !themeSet[themeName] {
				themeSet[themeName] = true
				themes = append(themes, themeName)
			}
			tr := runTriple(t, ctx, h, fix, themeName, tf)
			results = append(results, tr)
		}
	}

	report := fidelity.ComputeReport(results)
	slices.Sort(themes)
	md := fidelity.RenderMarkdown(report, themes)

	update := false
	for _, arg := range os.Args {
		if arg == "-update" {
			update = true
			break
		}
	}

	reportPath := filepath.Join("..", "..", "FIDELITY.md")
	if update {
		if err := os.WriteFile(reportPath, []byte(md), 0644); err != nil {
			t.Fatalf("write FIDELITY.md: %v", err)
		}
		t.Logf("Updated %s", reportPath)
		return
	}

	existing, err := os.ReadFile(reportPath)
	if err != nil {
		t.Logf("FIDELITY.md not found; run with -update to generate")
		return
	}
	if string(existing) != md {
		t.Errorf("FIDELITY.md is stale; run with -update to regenerate")
	}
}

func TestThemedToFixtureFontStyle(t *testing.T) {
	tokens := [][]ast.ThemedToken{{
		{Content: "func", Color: "#D73A49", FontStyle: theme.FontStyleItalic | theme.FontStyleBold},
	}}
	result := themedToFixture(tokens)
	got := result[0][0]
	if got.FontStyle != 3 {
		t.Errorf("FontStyle = %d, want 3 (italic|bold)", got.FontStyle)
	}
	if got.Start != 0 || got.End != 4 {
		t.Errorf("offsets [%d:%d], want [0:4]", got.Start, got.End)
	}
}
