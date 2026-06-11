package main

import (
	"fmt"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

func benchChroma(inputs []Input, iters int, theme string) (map[string]EngineResult, error) {
	style := styles.Get(theme)
	if style == nil {
		style = styles.Get("monokai")
	}
	formatter := html.New(html.WithClasses(false))

	results := make(map[string]EngineResult, len(inputs))
	for _, inp := range inputs {
		lexer := lexers.Get(inp.Lang)
		if lexer == nil {
			return nil, fmt.Errorf("chroma: no lexer for %s", inp.Lang)
		}
		lexer = chroma.Coalesce(lexer)

		// Cold: first call.
		var buf strings.Builder
		start := time.Now()
		iter, err := lexer.Tokenise(nil, inp.Code)
		if err != nil {
			return nil, fmt.Errorf("chroma cold tokenise %s: %w", inp.Name, err)
		}
		if err := formatter.Format(&buf, style, iter); err != nil {
			return nil, fmt.Errorf("chroma cold format %s: %w", inp.Name, err)
		}
		coldMs := float64(time.Since(start).Microseconds()) / 1000.0

		// Warm: N iterations.
		durations := make([]float64, iters)
		var totalAlloc, totalAllocs uint64
		for i := range iters {
			buf.Reset()
			var m1, m2 runtime.MemStats
			runtime.ReadMemStats(&m1)
			t0 := time.Now()
			iter, err := lexer.Tokenise(nil, inp.Code)
			if err != nil {
				return nil, fmt.Errorf("chroma warm tokenise %s: %w", inp.Name, err)
			}
			if err := formatter.Format(&buf, style, iter); err != nil {
				return nil, fmt.Errorf("chroma warm format %s: %w", inp.Name, err)
			}
			durations[i] = float64(time.Since(t0).Microseconds()) / 1000.0
			runtime.ReadMemStats(&m2)
			totalAlloc += m2.TotalAlloc - m1.TotalAlloc
			totalAllocs += m2.Mallocs - m1.Mallocs
		}
		sort.Float64s(durations)
		warmMs := durations[len(durations)/2]

		// Token/scope counting.
		iter, _ = lexer.Tokenise(nil, inp.Code)
		var tokenCount int
		typeSet := make(map[chroma.TokenType]struct{})
		for _, tok := range iter.Tokens() {
			if tok.Value == "" {
				continue
			}
			tokenCount++
			typeSet[tok.Type] = struct{}{}
		}

		results[inp.Name] = EngineResult{
			ColdMs: coldMs,
			WarmMs: warmMs,
			AllocB: int64(totalAlloc) / int64(iters),
			Allocs: int64(totalAllocs) / int64(iters),
			Tokens: tokenCount,
			Scopes: len(typeSet),
		}
	}
	return results, nil
}

func dumpChromaTokens(inputs []Input, theme string) (map[string]string, error) {
	style := styles.Get(theme)
	if style == nil {
		style = styles.Get("monokai")
	}
	dumps := make(map[string]string, len(inputs))
	for _, inp := range inputs {
		lexer := lexers.Get(inp.Lang)
		if lexer == nil {
			continue
		}
		lexer = chroma.Coalesce(lexer)
		iter, err := lexer.Tokenise(nil, inp.Code)
		if err != nil {
			return nil, fmt.Errorf("chroma tokenise %s: %w", inp.Name, err)
		}
		var buf strings.Builder
		for _, tok := range iter.Tokens() {
			if tok.Value == "" {
				continue
			}
			entry := style.Get(tok.Type)
			color := "#------"
			if entry.Colour.IsSet() {
				color = entry.Colour.String()
			}
			fs := ""
			if entry.Bold == chroma.Yes {
				fs = "[b]"
			} else if entry.Italic == chroma.Yes {
				fs = "[i]"
			} else if entry.Underline == chroma.Yes {
				fs = "[u]"
			}
			buf.WriteString(fmt.Sprintf("%-10s%-6s%s\n", color, fs, tok.Value))
		}
		dumps[inp.Lang] = buf.String()
	}
	return dumps, nil
}
