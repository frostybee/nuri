package nuri

import (
	"context"
	"testing"
)

func TestCodeToPlainTextReconstruct(t *testing.T) {
	h := newTestHighlighter(t)
	code := "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"hello\")\n}"
	got, err := h.CodeToPlainText(context.Background(), code+"\n", CodeToPlainTextOptions{
		Lang: "go", Theme: "github-dark",
	})
	if err != nil {
		t.Fatal(err)
	}
	// The tokenizer strips the trailing newline from the last line,
	// so the reconstructed text omits it. This matches Shiki behavior.
	if got != code {
		t.Errorf("plaintext output does not match source\ngot:  %q\nwant: %q", got, code)
	}
}

func TestCodeToPlainTextMultiLine(t *testing.T) {
	h := newTestHighlighter(t)
	got, err := h.CodeToPlainText(context.Background(), "a\nb\nc\n", CodeToPlainTextOptions{
		Lang: "go", Theme: "github-dark",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := "a\nb\nc"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestCodeToPlainTextEmpty(t *testing.T) {
	h := newTestHighlighter(t)
	got, err := h.CodeToPlainText(context.Background(), "", CodeToPlainTextOptions{
		Lang: "go", Theme: "github-dark",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestCodeToPlainTextUnknownLang(t *testing.T) {
	h := newTestHighlighter(t)
	got, err := h.CodeToPlainText(context.Background(), "hello world\n", CodeToPlainTextOptions{
		Lang: "nonexistent", Theme: "github-dark",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != "hello world" {
		t.Errorf("got %q, want %q", got, "hello world")
	}
}
