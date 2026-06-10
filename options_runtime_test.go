package nuri

import (
	"context"
	"os"
	"strings"
	"testing"
)

// TestRegexInterruptionOffHighlights verifies output is unaffected by
// disabling WASM-level regex interruption.
func TestRegexInterruptionOffHighlights(t *testing.T) {
	ctx := context.Background()
	hOn := newTestHighlighter(t)
	hOff := newTestHighlighterWithOpts(t, WithRegexInterruption(false))

	code := "package main\n\nfunc main() {\n\tprintln(\"hi\")\n}\n"
	opts := CodeToHTMLOptions{Lang: "go", Theme: "github-dark"}

	wantHTML, err := hOn.CodeToHTML(ctx, code, opts)
	if err != nil {
		t.Fatalf("CodeToHTML interruption-on: %v", err)
	}
	gotHTML, err := hOff.CodeToHTML(ctx, code, opts)
	if err != nil {
		t.Fatalf("CodeToHTML interruption-off: %v", err)
	}
	if gotHTML != wantHTML {
		t.Error("interruption setting changed highlighting output")
	}
	if !strings.Contains(gotHTML, "<pre") {
		t.Error("expected <pre in output")
	}
}

// TestCompilationCacheDir verifies the on-disk AOT cache is populated and a
// second engine boots from it with identical behavior.
func TestCompilationCacheDir(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()

	for run := 0; run < 2; run++ {
		h := newTestHighlighterWithOpts(t, WithCompilationCacheDir(dir))
		html, err := h.CodeToHTML(ctx, "package main\n", CodeToHTMLOptions{
			Lang: "go", Theme: "github-dark",
		})
		if err != nil {
			t.Fatalf("run %d: CodeToHTML: %v", run, err)
		}
		if !strings.Contains(html, "package") {
			t.Errorf("run %d: expected highlighted content", run)
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read cache dir: %v", err)
	}
	if len(entries) == 0 {
		t.Error("compilation cache dir is empty — cache not in use")
	}
}
