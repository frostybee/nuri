package tokenizer

import (
	"context"
	"fmt"
	"hash/fnv"

	"github.com/frostybee/nuri/internal/grammar"
	"github.com/frostybee/nuri/internal/oniguruma"
)

// matchResult holds the winning match from a pattern search.
type matchResult struct {
	match       *oniguruma.Match
	rule        grammar.Rule
	ruleGrammar *grammar.Grammar
}

// scannerCache caches compiled WASM scanners keyed by pattern set hash.
// Scoped to a single Tokenize() call — must not outlive the OnigLib instance.
type scannerCache struct {
	cache map[uint64]oniguruma.OnigScanner
}

func newScannerCache() *scannerCache {
	return &scannerCache{cache: make(map[uint64]oniguruma.OnigScanner)}
}

func (sc *scannerCache) closeAll() {
	for _, s := range sc.cache {
		s.Close()
	}
	sc.cache = nil
}

// hashPatterns computes an FNV-1a hash of the pattern byte slices.
func hashPatterns(patterns [][]byte) uint64 {
	h := fnv.New64a()
	for _, p := range patterns {
		h.Write(p)
		h.Write([]byte{0}) // separator
	}
	return h.Sum64()
}

// getOrCreate returns a cached scanner for the given patterns, or creates one.
func (sc *scannerCache) getOrCreate(
	ctx context.Context,
	onigLib oniguruma.OnigLib,
	patterns [][]byte,
) (oniguruma.OnigScanner, error) {
	key := hashPatterns(patterns)
	if scanner, ok := sc.cache[key]; ok {
		return scanner, nil
	}
	scanner, err := onigLib.NewScannerCtx(ctx, patterns)
	if err != nil {
		return nil, err
	}
	sc.cache[key] = scanner
	return scanner, nil
}

// findNextMatch uses the cache to get a scanner and returns
// the leftmost match, or nil if no pattern matches.
func findNextMatch(
	ctx context.Context,
	onigLib oniguruma.OnigLib,
	compiled *grammar.CompileResult,
	line []byte,
	pos int,
	cache *scannerCache,
	options oniguruma.SearchOptions,
) (*matchResult, error) {
	if len(compiled.Rules) == 0 {
		return nil, nil
	}

	patterns := make([][]byte, len(compiled.Rules))
	for i, cr := range compiled.Rules {
		patterns[i] = cr.Pattern
	}

	scanner, err := cache.getOrCreate(ctx, onigLib, patterns)
	if err != nil {
		return nil, fmt.Errorf("tokenizer: create scanner: %w", err)
	}

	m, err := scanner.FindNextMatchCtx(ctx, line, pos, options)
	if err != nil {
		return nil, fmt.Errorf("tokenizer: find match: %w", err)
	}
	if m == nil {
		return nil, nil
	}

	return &matchResult{
		match:       m,
		rule:        compiled.Rules[m.Index].Rule,
		ruleGrammar: compiled.Rules[m.Index].Grammar,
	}, nil
}
