package fidelity

import (
	"strings"
	"testing"
)

func TestRenderMarkdown(t *testing.T) {
	results := []TripleResult{
		{Grammar: "go", Theme: "github-light", Source: "go.go", Pass: true},
		{Grammar: "go", Theme: "github-dark", Source: "go.go", Pass: true},
		{Grammar: "typescript", Theme: "github-light", Source: "ts.ts", Pass: true},
		{Grammar: "typescript", Theme: "github-dark", Source: "ts.ts", Pass: false},
	}
	report := ComputeReport(results)
	md := RenderMarkdown(report, []string{"github-light", "github-dark"})

	if !strings.Contains(md, "# Fidelity Report") {
		t.Error("missing header")
	}
	if !strings.Contains(md, "3 / 4 pass") {
		t.Errorf("missing overall stats, got:\n%s", md)
	}
	if !strings.Contains(md, "| go |") {
		t.Error("missing go row")
	}
	if !strings.Contains(md, "| typescript |") {
		t.Error("missing typescript row")
	}
	if !strings.Contains(md, "shipping") {
		t.Error("missing 'shipping' status")
	}
	if !strings.Contains(md, "held") {
		t.Error("missing 'held' status")
	}
}

func TestRenderMarkdownEmpty(t *testing.T) {
	report := ComputeReport(nil)
	md := RenderMarkdown(report, []string{"github-light"})
	if !strings.Contains(md, "0 / 0 pass") {
		t.Errorf("unexpected output for empty report: %s", md)
	}
}
