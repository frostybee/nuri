package tokenizer

import (
	"fmt"
	"strings"
	"testing"
)

// DiffKind classifies a token mismatch for diagnostic purposes.
type DiffKind int

const (
	BoundaryMismatch DiffKind = iota // different start/end offsets
	ScopeMismatch                     // same span, different scopes
	MissingToken                      // token present in expected, absent in actual
	ExtraToken                        // token present in actual, absent in expected
)

func (k DiffKind) String() string {
	switch k {
	case BoundaryMismatch:
		return "BoundaryMismatch"
	case ScopeMismatch:
		return "ScopeMismatch"
	case MissingToken:
		return "MissingToken"
	case ExtraToken:
		return "ExtraToken"
	default:
		return "Unknown"
	}
}

// TokenDiff describes one difference between expected and actual tokens.
type TokenDiff struct {
	Line   int
	Kind   DiffKind
	Want   *Token
	Got    *Token
	Detail string
}

func (d TokenDiff) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "line %d: %s", d.Line, d.Kind)
	if d.Detail != "" {
		fmt.Fprintf(&sb, " — %s", d.Detail)
	}
	if d.Want != nil {
		fmt.Fprintf(&sb, "\n  want: [%d:%d] scopes=%v", d.Want.Start, d.Want.End, d.Want.Scopes)
	}
	if d.Got != nil {
		fmt.Fprintf(&sb, "\n  got:  [%d:%d] scopes=%v", d.Got.Start, d.Got.End, d.Got.Scopes)
	}
	return sb.String()
}

// CompareTokens compares expected vs actual token lines and returns
// localized diffs. Returns nil if they match.
func CompareTokens(want, got [][]Token) []TokenDiff {
	var diffs []TokenDiff
	maxLines := len(want)
	if len(got) > maxLines {
		maxLines = len(got)
	}

	for line := 0; line < maxLines; line++ {
		if line >= len(want) {
			diffs = append(diffs, TokenDiff{
				Line:   line,
				Kind:   ExtraToken,
				Detail: fmt.Sprintf("extra line %d in actual output", line),
			})
			continue
		}
		if line >= len(got) {
			diffs = append(diffs, TokenDiff{
				Line:   line,
				Kind:   MissingToken,
				Detail: fmt.Sprintf("missing line %d in actual output", line),
			})
			continue
		}

		lineDiffs := compareTokenLine(line, want[line], got[line])
		diffs = append(diffs, lineDiffs...)
	}

	return diffs
}

func compareTokenLine(line int, want, got []Token) []TokenDiff {
	var diffs []TokenDiff
	wi, gi := 0, 0

	for wi < len(want) && gi < len(got) {
		w, g := want[wi], got[gi]

		if w.Start != g.Start || w.End != g.End {
			diffs = append(diffs, TokenDiff{
				Line:   line,
				Kind:   BoundaryMismatch,
				Want:   &w,
				Got:    &g,
				Detail: fmt.Sprintf("token %d: boundary [%d:%d] vs [%d:%d]", wi, w.Start, w.End, g.Start, g.End),
			})
			return diffs // stop at first boundary mismatch — cascading diffs are noise
		}

		if !scopesEqual(w.Scopes, g.Scopes) {
			diffs = append(diffs, TokenDiff{
				Line:   line,
				Kind:   ScopeMismatch,
				Want:   &w,
				Got:    &g,
				Detail: fmt.Sprintf("token %d: scopes differ", wi),
			})
		}

		wi++
		gi++
	}

	for ; wi < len(want); wi++ {
		w := want[wi]
		diffs = append(diffs, TokenDiff{
			Line:   line,
			Kind:   MissingToken,
			Want:   &w,
			Detail: fmt.Sprintf("missing token %d", wi),
		})
	}
	for ; gi < len(got); gi++ {
		g := got[gi]
		diffs = append(diffs, TokenDiff{
			Line:   line,
			Kind:   ExtraToken,
			Got:    &g,
			Detail: fmt.Sprintf("extra token %d", gi),
		})
	}

	return diffs
}

func scopesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestCompareTokensIdentical(t *testing.T) {
	tokens := [][]Token{{
		{Scopes: []string{"source.go", "keyword"}, Start: 0, End: 3},
		{Scopes: []string{"source.go"}, Start: 3, End: 4},
	}}
	diffs := CompareTokens(tokens, tokens)
	if len(diffs) != 0 {
		t.Errorf("expected no diffs, got %d", len(diffs))
	}
}

func TestCompareTokensBoundaryMismatch(t *testing.T) {
	want := [][]Token{{
		{Scopes: []string{"source.go"}, Start: 0, End: 5},
	}}
	got := [][]Token{{
		{Scopes: []string{"source.go"}, Start: 0, End: 4},
	}}
	diffs := CompareTokens(want, got)
	if len(diffs) != 1 || diffs[0].Kind != BoundaryMismatch {
		t.Errorf("expected BoundaryMismatch, got %v", diffs)
	}
}

func TestCompareTokensScopeMismatch(t *testing.T) {
	want := [][]Token{{
		{Scopes: []string{"source.go", "keyword.var"}, Start: 0, End: 3},
	}}
	got := [][]Token{{
		{Scopes: []string{"source.go", "keyword.const"}, Start: 0, End: 3},
	}}
	diffs := CompareTokens(want, got)
	if len(diffs) != 1 || diffs[0].Kind != ScopeMismatch {
		t.Errorf("expected ScopeMismatch, got %v", diffs)
	}
}

func TestCompareTokensMissing(t *testing.T) {
	want := [][]Token{{
		{Scopes: []string{"a"}, Start: 0, End: 3},
		{Scopes: []string{"b"}, Start: 3, End: 6},
	}}
	got := [][]Token{{
		{Scopes: []string{"a"}, Start: 0, End: 3},
	}}
	diffs := CompareTokens(want, got)
	if len(diffs) != 1 || diffs[0].Kind != MissingToken {
		t.Errorf("expected MissingToken, got %v", diffs)
	}
}

func TestCompareTokensExtra(t *testing.T) {
	want := [][]Token{{
		{Scopes: []string{"a"}, Start: 0, End: 3},
	}}
	got := [][]Token{{
		{Scopes: []string{"a"}, Start: 0, End: 3},
		{Scopes: []string{"b"}, Start: 3, End: 6},
	}}
	diffs := CompareTokens(want, got)
	if len(diffs) != 1 || diffs[0].Kind != ExtraToken {
		t.Errorf("expected ExtraToken, got %v", diffs)
	}
}
