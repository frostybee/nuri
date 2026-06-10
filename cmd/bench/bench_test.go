package bench_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/frostybee/nuri"
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
	benchTSCode = `interface User {
  id: number;
  name: string;
  email?: string;
}

function greet<T extends User>(user: T): string {
  return "Hello, " + user.name;
}

const users: User[] = [{ id: 1, name: "Alice" }];
`
	benchMarkdownCode = "# Heading\n\nA paragraph with **bold**, *italic*, and `code`.\n\n- Item one\n- Item two\n\n```go\nfunc main() {\n    fmt.Println(\"Hello\")\n}\n```\n"
)

type benchInput struct {
	name string
	lang string
	code string
}

var smallInputs = []benchInput{
	{"Go", "go", benchGoCode},
	{"JavaScript", "javascript", benchJSCode},
	{"HTML", "html", benchHTMLCode},
	{"TypeScript", "typescript", benchTSCode},
	{"Markdown", "markdown", benchMarkdownCode},
}

var langSamples = []struct {
	lang       string
	sampleFile string
	smallCode  string
}{
	{"go", "go.sample", benchGoCode},
	{"javascript", "javascript.sample", benchJSCode},
	{"html", "html.sample", benchHTMLCode},
	{"typescript", "typescript.sample", benchTSCode},
	{"markdown", "markdown.sample", benchMarkdownCode},
}

func newBenchHighlighter(b *testing.B) *nuri.Highlighter {
	b.Helper()
	return newBenchHighlighterPooled(b, 1)
}

func newBenchHighlighterPooled(b *testing.B, poolSize int) *nuri.Highlighter {
	b.Helper()
	ctx := context.Background()
	h, err := nuri.New(ctx,
		nuri.WithGrammarFS(os.DirFS(shared.GrammarsDir(b))),
		nuri.WithThemeFS(os.DirFS(shared.ThemesDir(b))),
		nuri.WithPoolSize(poolSize),
	)
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { h.Close(ctx) })
	return h
}

func makeLargeBenchInput(code string, targetBytes int) string {
	var sb strings.Builder
	for sb.Len() < targetBytes {
		sb.WriteString(code)
	}
	return sb.String()
}

func benchInputsForLang(lang, sampleFile string) []benchInput {
	dir := filepath.Join(shared.RepoRoot(), shared.SubmoduleSamplesDir)
	if _, err := os.Stat(dir); err != nil {
		return nil
	}
	data, err := os.ReadFile(filepath.Join(dir, sampleFile))
	if err != nil {
		return nil
	}
	medium := string(data)
	return []benchInput{
		{lang + "/Medium", lang, medium},
		{lang + "/Large", lang, makeLargeBenchInput(medium, 50*1024)},
	}
}

// --- Migrated benchmarks ---

func BenchmarkNew(b *testing.B) {
	ctx := context.Background()
	grammarFS := os.DirFS(shared.GrammarsDir(b))
	themeFS := os.DirFS(shared.ThemesDir(b))

	for b.Loop() {
		h, err := nuri.New(ctx,
			nuri.WithGrammarFS(grammarFS),
			nuri.WithThemeFS(themeFS),
			nuri.WithPoolSize(1),
		)
		if err != nil {
			b.Fatal(err)
		}
		h.Close(ctx)
	}
}

func BenchmarkCodeToTokens(b *testing.B) {
	h := newBenchHighlighter(b)
	ctx := context.Background()

	for _, tc := range smallInputs {
		b.Run(tc.name, func(b *testing.B) {
			b.SetBytes(int64(len(tc.code)))
			b.ResetTimer()
			for b.Loop() {
				_, err := h.CodeToTokens(ctx, tc.code, nuri.CodeToTokensOptions{
					Lang:  tc.lang,
					Theme: "github-dark",
				})
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}

	for _, ls := range langSamples {
		for _, bi := range benchInputsForLang(ls.lang, ls.sampleFile) {
			b.Run(bi.name, func(b *testing.B) {
				b.SetBytes(int64(len(bi.code)))
				b.ResetTimer()
				for b.Loop() {
					_, err := h.CodeToTokens(ctx, bi.code, nuri.CodeToTokensOptions{
						Lang:  bi.lang,
						Theme: "github-dark",
					})
					if err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}

func BenchmarkCodeToHighlightedTokens(b *testing.B) {
	h := newBenchHighlighter(b)
	ctx := context.Background()

	for _, tc := range smallInputs {
		b.Run(tc.name, func(b *testing.B) {
			b.SetBytes(int64(len(tc.code)))
			b.ResetTimer()
			for b.Loop() {
				_, err := h.CodeToHighlightedTokens(ctx, tc.code, nuri.CodeToTokensOptions{
					Lang:  tc.lang,
					Theme: "github-dark",
				})
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}

	for _, ls := range langSamples {
		for _, bi := range benchInputsForLang(ls.lang, ls.sampleFile) {
			b.Run(bi.name, func(b *testing.B) {
				b.SetBytes(int64(len(bi.code)))
				b.ResetTimer()
				for b.Loop() {
					_, err := h.CodeToHighlightedTokens(ctx, bi.code, nuri.CodeToTokensOptions{
						Lang:  bi.lang,
						Theme: "github-dark",
					})
					if err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}

func BenchmarkCodeToHTML(b *testing.B) {
	h := newBenchHighlighter(b)
	ctx := context.Background()

	for _, tc := range smallInputs {
		b.Run(tc.name, func(b *testing.B) {
			b.SetBytes(int64(len(tc.code)))
			b.ResetTimer()
			for b.Loop() {
				_, err := h.CodeToHTML(ctx, tc.code, nuri.CodeToHTMLOptions{
					Lang:  tc.lang,
					Theme: "github-dark",
				})
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}

	for _, ls := range langSamples {
		for _, bi := range benchInputsForLang(ls.lang, ls.sampleFile) {
			b.Run(bi.name, func(b *testing.B) {
				b.SetBytes(int64(len(bi.code)))
				b.ResetTimer()
				for b.Loop() {
					_, err := h.CodeToHTML(ctx, bi.code, nuri.CodeToHTMLOptions{
						Lang:  bi.lang,
						Theme: "github-dark",
					})
					if err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}

func BenchmarkCodeToANSI(b *testing.B) {
	h := newBenchHighlighter(b)
	ctx := context.Background()

	for _, tc := range smallInputs {
		b.Run(tc.name, func(b *testing.B) {
			b.SetBytes(int64(len(tc.code)))
			b.ResetTimer()
			for b.Loop() {
				_, err := h.CodeToANSI(ctx, tc.code, nuri.CodeToANSIOptions{
					Lang:  tc.lang,
					Theme: "github-dark",
				})
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}

	for _, ls := range langSamples {
		for _, bi := range benchInputsForLang(ls.lang, ls.sampleFile) {
			b.Run(bi.name, func(b *testing.B) {
				b.SetBytes(int64(len(bi.code)))
				b.ResetTimer()
				for b.Loop() {
					_, err := h.CodeToANSI(ctx, bi.code, nuri.CodeToANSIOptions{
						Lang:  bi.lang,
						Theme: "github-dark",
					})
					if err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}

// --- New benchmarks ---

func BenchmarkCodeToHTMLMultiTheme(b *testing.B) {
	h := newBenchHighlighter(b)
	ctx := context.Background()

	themes := map[string]string{
		"dark":  "github-dark",
		"light": "github-light",
	}

	cases := []benchInput{
		{"Go", "go", benchGoCode},
		{"JavaScript", "javascript", benchJSCode},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			b.SetBytes(int64(len(tc.code)))
			b.ResetTimer()
			for b.Loop() {
				_, err := h.CodeToHTML(ctx, tc.code, nuri.CodeToHTMLOptions{
					Lang:   tc.lang,
					Themes: themes,
				})
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkConcurrent(b *testing.B) {
	h := newBenchHighlighterPooled(b, 4)
	ctx := context.Background()

	cases := []struct {
		name string
		fn   func() error
		size int64
	}{
		{
			name: "CodeToTokens/Go",
			fn: func() error {
				_, err := h.CodeToTokens(ctx, benchGoCode, nuri.CodeToTokensOptions{
					Lang: "go", Theme: "github-dark",
				})
				return err
			},
			size: int64(len(benchGoCode)),
		},
		{
			name: "CodeToHTML/JavaScript",
			fn: func() error {
				_, err := h.CodeToHTML(ctx, benchJSCode, nuri.CodeToHTMLOptions{
					Lang: "javascript", Theme: "github-dark",
				})
				return err
			},
			size: int64(len(benchJSCode)),
		},
		{
			name: "CodeToANSI/Go",
			fn: func() error {
				_, err := h.CodeToANSI(ctx, benchGoCode, nuri.CodeToANSIOptions{
					Lang: "go", Theme: "github-dark",
				})
				return err
			},
			size: int64(len(benchGoCode)),
		},
	}

	for _, tc := range cases {
		tc := tc
		b.Run(tc.name, func(b *testing.B) {
			b.SetBytes(tc.size)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					if err := tc.fn(); err != nil {
						b.Fatal(err)
					}
				}
			})
		})
	}
}

func BenchmarkDetectLanguage(b *testing.B) {
	h := newBenchHighlighter(b)

	cases := []struct {
		name     string
		filename string
	}{
		{"Extension", "main.go"},
		{"Path", "/home/user/project/src/components/App.tsx"},
		{"Filename", "Dockerfile"},
		{"Unknown", "data.xyz"},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for b.Loop() {
				h.DetectLanguage(tc.filename)
			}
		})
	}
}
