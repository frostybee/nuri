package nuri

import (
	"context"
	"os"
	"testing"

	"github.com/frostybee/nuri/internal/shared"
)

var (
	benchGoCode = `package main

import "fmt"

func main() {
	for i := range 10 {
		if i%2 == 0 {
			fmt.Println(i, "is even")
		}
	}
}
`
	benchJSCode = `const express = require('express');
const app = express();

app.get('/api/users/:id', async (req, res) => {
  try {
    const user = await db.findById(req.params.id);
    res.json({ user, timestamp: Date.now() });
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});

app.listen(3000);
`
	benchHTMLCode = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Test</title>
    <style>
        body { font-family: sans-serif; color: #333; }
        .highlight { background: yellow; }
    </style>
</head>
<body>
    <div id="app">
        <h1>Hello World</h1>
    </div>
</body>
</html>
`
)

func newBenchHighlighter(b *testing.B) *Highlighter {
	b.Helper()
	ctx := context.Background()
	h, err := New(ctx,
		WithGrammarFS(os.DirFS(shared.GrammarsDir(b))),
		WithThemeFS(os.DirFS(shared.ThemesDir(b))),
		WithPoolSize(1),
	)
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { h.Close(ctx) })
	return h
}

func BenchmarkNew(b *testing.B) {
	ctx := context.Background()
	grammarFS := os.DirFS(shared.GrammarsDir(b))
	themeFS := os.DirFS(shared.ThemesDir(b))

	for b.Loop() {
		h, err := New(ctx,
			WithGrammarFS(grammarFS),
			WithThemeFS(themeFS),
			WithPoolSize(1),
		)
		if err != nil {
			b.Fatal(err)
		}
		h.Close(ctx)
	}
}

func BenchmarkCodeToTokens(b *testing.B) {
	cases := []struct {
		name string
		lang string
		code string
	}{
		{"Go", "go", benchGoCode},
		{"JavaScript", "javascript", benchJSCode},
		{"HTML", "html", benchHTMLCode},
	}

	h := newBenchHighlighter(b)
	ctx := context.Background()

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			b.SetBytes(int64(len(tc.code)))
			b.ResetTimer()
			for b.Loop() {
				_, err := h.CodeToTokens(ctx, tc.code, CodeToTokensOptions{
					Lang:  tc.lang,
					Theme: "github-dark",
				})
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkCodeToHTML(b *testing.B) {
	cases := []struct {
		name string
		lang string
		code string
	}{
		{"Go", "go", benchGoCode},
		{"JavaScript", "javascript", benchJSCode},
		{"HTML", "html", benchHTMLCode},
	}

	h := newBenchHighlighter(b)
	ctx := context.Background()

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			b.SetBytes(int64(len(tc.code)))
			b.ResetTimer()
			for b.Loop() {
				_, err := h.CodeToHTML(ctx, tc.code, CodeToHTMLOptions{
					Lang:  tc.lang,
					Theme: "github-dark",
				})
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
