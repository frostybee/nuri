package tokenizer

// Token represents a scoped span of source text.
type Token struct {
	Scopes []string
	Start  int // UTF-8 byte offset
	End    int // UTF-8 byte offset
}

// TokenizeResult holds the output of tokenization.
type TokenizeResult struct {
	Lines       [][]Token
	Diagnostics []Diagnostic
}

// Diagnostic records a non-fatal per-line degradation.
type Diagnostic struct {
	Line int
	Kind string // "timeout" | "too_long" | "panic"
}

// lineTokenBuilder accumulates tokens for a single line.
// It uses the produce(end) pattern: the builder tracks lastEndPos
// internally and uses it as the implicit start of the next token.
// This makes overlapping tokens structurally impossible.
type lineTokenBuilder struct {
	tokens     []Token
	lastEndPos int
}

func newLineTokenBuilder(startPos int) *lineTokenBuilder {
	return &lineTokenBuilder{
		lastEndPos: startPos,
	}
}

// produce emits a token from lastEndPos to endPos with the given scopes.
// If endPos <= lastEndPos, the call is a no-op (prevents backwards/empty tokens).
func (b *lineTokenBuilder) produce(endPos int, scopes []string) {
	if endPos <= b.lastEndPos {
		return
	}
	b.tokens = append(b.tokens, Token{
		Scopes: scopes,
		Start:  b.lastEndPos,
		End:    endPos,
	})
	b.lastEndPos = endPos
}

// finalize strips the trailing newline from the last token. Lines are
// tokenized with a re-added \n (needed for regex anchors), but the
// renderer adds its own \n between line spans. Without this, every
// line break renders twice. Matches Giallo's TokenAccumulator.finalize().
func (b *lineTokenBuilder) finalize(lineLen int) {
	if len(b.tokens) == 0 || lineLen == 0 {
		return
	}
	last := &b.tokens[len(b.tokens)-1]
	if last.Start == lineLen-1 {
		b.tokens = b.tokens[:len(b.tokens)-1]
		return
	}
	if last.End == lineLen {
		last.End--
		if last.Start >= last.End {
			b.tokens = b.tokens[:len(b.tokens)-1]
		}
	}
}

func (b *lineTokenBuilder) finish() []Token {
	return b.tokens
}
