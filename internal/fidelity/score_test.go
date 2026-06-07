package fidelity

import (
	"testing"
)

func TestComputeReport(t *testing.T) {
	results := []TripleResult{
		{Grammar: "go", Theme: "github-light", Source: "go.go", Pass: true},
		{Grammar: "go", Theme: "github-dark", Source: "go.go", Pass: true},
		{Grammar: "typescript", Theme: "github-light", Source: "ts.ts", Pass: true},
		{Grammar: "typescript", Theme: "github-dark", Source: "ts.ts", Pass: false, Diffs: []TokenDiff{{Kind: StyleMismatch}}},
	}

	report := ComputeReport(results)

	if report.GlobalPass != 3 || report.GlobalTotal != 4 {
		t.Errorf("global: %d/%d, want 3/4", report.GlobalPass, report.GlobalTotal)
	}

	goScore := report.ByGrammar["go"]
	if goScore.Pass != 2 || goScore.Total != 2 {
		t.Errorf("go: %d/%d, want 2/2", goScore.Pass, goScore.Total)
	}

	tsScore := report.ByGrammar["typescript"]
	if tsScore.Pass != 1 || tsScore.Total != 2 {
		t.Errorf("typescript: %d/%d, want 1/2", tsScore.Pass, tsScore.Total)
	}

	lightScore := report.ByTheme["github-light"]
	if lightScore.Pass != 2 || lightScore.Total != 2 {
		t.Errorf("github-light: %d/%d, want 2/2", lightScore.Pass, lightScore.Total)
	}

	darkScore := report.ByTheme["github-dark"]
	if darkScore.Pass != 1 || darkScore.Total != 2 {
		t.Errorf("github-dark: %d/%d, want 1/2", darkScore.Pass, darkScore.Total)
	}
}

func TestComputeReportEmpty(t *testing.T) {
	report := ComputeReport(nil)
	if report.GlobalPass != 0 || report.GlobalTotal != 0 {
		t.Errorf("expected 0/0, got %d/%d", report.GlobalPass, report.GlobalTotal)
	}
}

func TestGrammarScoreRate(t *testing.T) {
	s := GrammarScore{Pass: 3, Total: 4}
	if r := s.Rate(); r != 0.75 {
		t.Errorf("Rate() = %f, want 0.75", r)
	}
}

func TestGrammarScoreRateZero(t *testing.T) {
	s := GrammarScore{Pass: 0, Total: 0}
	if r := s.Rate(); r != 0 {
		t.Errorf("Rate() = %f, want 0", r)
	}
}
