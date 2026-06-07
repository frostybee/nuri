package full_test

import (
	"context"
	"io/fs"
	"testing"

	"github.com/frostybee/nuri"
	"github.com/frostybee/nuri/bundle/full"
)

func TestFullBundleFS(t *testing.T) {
	fsys := full.FS()

	data, err := fs.ReadFile(fsys, "grammars/go.json")
	if err != nil {
		t.Fatalf("read grammar: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("empty grammar file")
	}

	data, err = fs.ReadFile(fsys, "themes/github-dark.json")
	if err != nil {
		t.Fatalf("read theme: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("empty theme file")
	}
}

func TestFullBundleHighlight(t *testing.T) {
	ctx := context.Background()
	h, err := nuri.New(ctx,
		nuri.WithFS(full.FS()),
		nuri.WithPoolSize(1),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer h.Close(ctx)

	html, err := h.CodeToHTML(ctx, "fn main() {}\n", nuri.CodeToHTMLOptions{
		Lang:  "rust",
		Theme: "dracula",
	})
	if err != nil {
		t.Fatal(err)
	}
	if html == "" {
		t.Error("empty HTML output")
	}
}

func TestFullBundleHasMoreThanCore(t *testing.T) {
	ctx := context.Background()
	h, err := nuri.New(ctx,
		nuri.WithFS(full.FS()),
		nuri.WithPoolSize(1),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer h.Close(ctx)

	result, err := h.CodeToTokens(ctx, "module Main where", nuri.CodeToTokensOptions{
		Lang:  "haskell",
		Theme: "github-dark",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Tokens) == 0 {
		t.Error("expected tokens for haskell (full bundle)")
	}
}

func TestFullBundleGrammarCount(t *testing.T) {
	fsys := full.FS()
	entries, err := fs.ReadDir(fsys, "grammars")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) < 200 {
		t.Errorf("expected 200+ grammars in full bundle, got %d", len(entries))
	}
}
