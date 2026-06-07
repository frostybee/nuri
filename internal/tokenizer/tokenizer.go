package tokenizer

import (
	"bytes"
	"context"

	"github.com/frostybee/nuri/internal/grammar"
	"github.com/frostybee/nuri/internal/oniguruma"
)

// Tokenize tokenizes source code using the given grammar. The resolver is
// optional (nil skips cross-grammar includes and external injections).
func Tokenize(
	ctx context.Context,
	code []byte,
	g *grammar.Grammar,
	onigLib oniguruma.OnigLib,
	resolver ...grammar.GrammarResolver,
) (*TokenizeResult, error) {
	if len(code) == 0 {
		return &TokenizeResult{}, nil
	}

	var res grammar.GrammarResolver
	if len(resolver) > 0 {
		res = resolver[0]
	}

	injections := collectInjections(g, res)

	lines := splitLines(code)
	result := &TokenizeResult{
		Lines: make([][]Token, 0, len(lines)),
	}

	cache := newScannerCache()
	defer cache.closeAll()

	state := newStateStack(nil, g.ScopeName)

	for i, line := range lines {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		state.resetForNewLine()

		tokens, newState, err := tokenizeLine(
			ctx, line, g, onigLib, state, res, injections, cache, 0, i == 0,
		)
		if err != nil {
			return nil, err
		}

		result.Lines = append(result.Lines, tokens)
		state = newState
	}

	return result, nil
}

func splitLines(code []byte) [][]byte {
	lines := bytes.Split(code, []byte("\n"))
	// Re-add the newline to each line except the last
	for i := 0; i < len(lines)-1; i++ {
		lines[i] = append(lines[i], '\n')
	}
	// Remove trailing empty line if present
	if len(lines) > 0 && len(lines[len(lines)-1]) == 0 {
		lines = lines[:len(lines)-1]
	}
	return lines
}
