package core_test

import (
	"context"
	"io/fs"
	"testing"

	"github.com/frostybee/nuri"
	"github.com/frostybee/nuri/bundle/core"
)

func TestCoreBundleFS(t *testing.T) {
	fsys := core.FS()

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

func TestCoreBundleHighlight(t *testing.T) {
	ctx := context.Background()
	h, err := nuri.New(ctx,
		nuri.WithFS(core.FS()),
		nuri.WithPoolSize(1),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer h.Close(ctx)

	html, err := h.CodeToHTML(ctx, "const x = 1;\n", nuri.CodeToHTMLOptions{
		Lang:  "js",
		Theme: "github-dark",
	})
	if err != nil {
		t.Fatal(err)
	}
	if html == "" {
		t.Error("empty HTML output")
	}
}

func TestCoreBundleAliases(t *testing.T) {
	ctx := context.Background()
	h, err := nuri.New(ctx,
		nuri.WithFS(core.FS()),
		nuri.WithPoolSize(1),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer h.Close(ctx)

	aliases := []struct {
		alias string
		code  string
	}{
		{"js", "const x = 1;"},
		{"ts", "const x: number = 1;"},
		{"py", "x = 42"},
		{"bash", "echo hello"},
		{"yml", "key: value"},
		{"md", "# Hello"},
		{"rs", "fn main() {}"},
		{"rb", "puts 'hello'"},
	}

	for _, tc := range aliases {
		t.Run(tc.alias, func(t *testing.T) {
			result, err := h.CodeToTokens(ctx, tc.code, nuri.CodeToTokensOptions{
				Lang:  tc.alias,
				Theme: "github-dark",
			})
			if err != nil {
				t.Fatalf("alias %q: %v", tc.alias, err)
			}
			if len(result.Tokens) == 0 {
				t.Errorf("alias %q: no tokens", tc.alias)
			}
		})
	}
}
