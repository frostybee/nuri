package main

import (
	"context"
	"fmt"
	"runtime"
	"sort"
	"time"

	"github.com/frostybee/nuri"
	"github.com/frostybee/nuri/ast"
	"github.com/frostybee/nuri/bundle/core"
	"github.com/frostybee/nuri/theme"
)

func benchNuri(inputs []Input, iters int, theme string, interruption bool) (map[string]EngineResult, error) {
	ctx := context.Background()
	h, err := nuri.New(ctx,
		nuri.WithFS(core.FS()),
		nuri.WithPoolSize(1),
		nuri.WithRegexInterruption(interruption),
	)
	if err != nil {
		return nil, fmt.Errorf("nuri.New: %w", err)
	}
	defer h.Close(ctx)

	results := make(map[string]EngineResult, len(inputs))
	for _, inp := range inputs {
		opts := ast.CodeToHTMLOptions{Lang: inp.Lang, Theme: theme}

		// Cold: first call includes grammar compile + theme resolve.
		start := time.Now()
		_, err := h.CodeToHTML(ctx, inp.Code, opts)
		coldMs := float64(time.Since(start).Microseconds()) / 1000.0
		if err != nil {
			return nil, fmt.Errorf("nuri cold %s: %w", inp.Name, err)
		}

		// Warm: N iterations, collect durations for median.
		durations := make([]float64, iters)
		var totalAlloc, totalAllocs uint64
		for i := range iters {
			var m1, m2 runtime.MemStats
			runtime.ReadMemStats(&m1)
			t0 := time.Now()
			_, err := h.CodeToHTML(ctx, inp.Code, opts)
			durations[i] = float64(time.Since(t0).Microseconds()) / 1000.0
			runtime.ReadMemStats(&m2)
			if err != nil {
				return nil, fmt.Errorf("nuri warm %s iter %d: %w", inp.Name, i, err)
			}
			totalAlloc += m2.TotalAlloc - m1.TotalAlloc
			totalAllocs += m2.Mallocs - m1.Mallocs
		}
		sort.Float64s(durations)
		warmMs := durations[len(durations)/2]

		// Fidelity: token + scope counting.
		tokResult, err := h.CodeToTokens(ctx, inp.Code, ast.CodeToTokensOptions{Lang: inp.Lang, Theme: theme})
		if err != nil {
			return nil, fmt.Errorf("nuri tokens %s: %w", inp.Name, err)
		}
		var tokenCount int
		scopeSet := make(map[string]struct{})
		for _, line := range tokResult.Tokens {
			tokenCount += len(line)
			for _, tok := range line {
				for _, s := range tok.Scopes {
					scopeSet[s] = struct{}{}
				}
			}
		}

		results[inp.Name] = EngineResult{
			ColdMs: coldMs,
			WarmMs: warmMs,
			AllocB: int64(totalAlloc) / int64(iters),
			Allocs: int64(totalAllocs) / int64(iters),
			Tokens: tokenCount,
			Scopes: len(scopeSet),
		}
	}
	return results, nil
}

func dumpNuriTokens(inputs []Input, theme string) (map[string]string, error) {
	ctx := context.Background()
	h, err := nuri.New(ctx,
		nuri.WithFS(core.FS()),
		nuri.WithPoolSize(1),
	)
	if err != nil {
		return nil, fmt.Errorf("nuri.New: %w", err)
	}
	defer h.Close(ctx)

	dumps := make(map[string]string, len(inputs))
	for _, inp := range inputs {
		tokResult, err := h.CodeToTokens(ctx, inp.Code, ast.CodeToTokensOptions{Lang: inp.Lang, Theme: theme})
		if err != nil {
			return nil, fmt.Errorf("nuri tokens %s: %w", inp.Name, err)
		}
		dumps[inp.Lang] = formatTokenDump(tokResult.Tokens)
	}
	return dumps, nil
}

func formatTokenDump(lines [][]ast.ThemedToken) string {
	var buf []byte
	for _, line := range lines {
		for _, tok := range line {
			color := tok.Color
			if color == "" {
				color = "#------"
			}
			style := fontStyleAbbrev(tok.FontStyle)
			buf = append(buf, fmt.Sprintf("%-10s%-6s%s\n", color, style, tok.Content)...)
		}
	}
	return string(buf)
}

func fontStyleAbbrev(fs theme.FontStyle) string {
	switch {
	case fs&1 != 0:
		return "[i]"
	case fs&2 != 0:
		return "[b]"
	case fs&4 != 0:
		return "[u]"
	case fs&8 != 0:
		return "[s]"
	default:
		return ""
	}
}
