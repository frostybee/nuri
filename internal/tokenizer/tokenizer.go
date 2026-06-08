package tokenizer

import (
	"bytes"
	"context"
	"time"

	"github.com/frostybee/nuri/internal/grammar"
	"github.com/frostybee/nuri/internal/oniguruma"
)

// panicOnLineHook is a test-only hook for panic recovery testing.
// When >= 0, Tokenize panics before tokenizing that line index.
var panicOnLineHook = -1

// TokenizeOptions carries per-call safety configuration.
type TokenizeOptions struct {
	MaxLineLength int // 0 = no limit; lines exceeding this are emitted unstyled
	TimeoutMs     int // 0 = no timeout; per-line soft timeout in milliseconds
}

// Tokenize tokenizes source code using the given grammar. The resolver is
// optional (nil skips cross-grammar includes and external injections).
func Tokenize(
	ctx context.Context,
	code []byte,
	g *grammar.Grammar,
	onigLib oniguruma.OnigLib,
	opts TokenizeOptions,
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

		// MaxLineLength pre-filter: skip tokenization for oversized lines.
		if opts.MaxLineLength > 0 {
			contentLen := len(line)
			if contentLen > 0 && line[contentLen-1] == '\n' {
				contentLen--
			}
			if contentLen > opts.MaxLineLength {
				tok := Token{Scopes: state.scopeSlice(), Start: 0, End: contentLen}
				result.Lines = append(result.Lines, []Token{tok})
				result.Diagnostics = append(result.Diagnostics, Diagnostic{Line: i, Kind: "too_long"})
				continue
			}
		}

		// Compute per-line deadline for soft timeout.
		var deadline time.Time
		if opts.TimeoutMs > 0 {
			deadline = time.Now().Add(time.Duration(opts.TimeoutMs) * time.Millisecond)
		}

		var (
			tokens       []Token
			newState     *StateStack
			stoppedEarly bool
			lineErr      error
			panicked     bool
		)
		func() {
			defer func() {
				if r := recover(); r != nil {
					panicked = true
				}
			}()
			if panicOnLineHook == i {
				panic("test panic on line")
			}
			tokens, newState, stoppedEarly, lineErr = tokenizeLine(
				ctx, line, g, onigLib, state, res, injections, cache, 0, i == 0, deadline,
			)
		}()

		if panicked {
			contentLen := len(line)
			if contentLen > 0 && line[contentLen-1] == '\n' {
				contentLen--
			}
			tokens = []Token{{Scopes: state.scopeSlice(), Start: 0, End: contentLen}}
			newState = state
			result.Diagnostics = append(result.Diagnostics, Diagnostic{Line: i, Kind: "panic"})
		} else if lineErr != nil {
			return nil, lineErr
		} else if stoppedEarly {
			result.Diagnostics = append(result.Diagnostics, Diagnostic{Line: i, Kind: "timeout"})
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
