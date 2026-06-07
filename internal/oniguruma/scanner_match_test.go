package oniguruma

import (
	"context"
	"testing"
)

func newTestScanner(t *testing.T, patterns ...string) (*instance, *Scanner) {
	t.Helper()
	ctx := context.Background()
	inst := newTestInstance(t)
	pats := make([][]byte, len(patterns))
	for i, p := range patterns {
		pats[i] = []byte(p)
	}
	scanner, err := inst.NewScanner(ctx, pats)
	if err != nil {
		t.Fatalf("NewScanner: %v", err)
	}
	t.Cleanup(func() { scanner.Free(ctx) })
	return inst, scanner
}

func TestRealGrammarPatterns(t *testing.T) {
	tests := []struct {
		name       string
		grammar    string
		pattern    string
		input      string
		startPos   int
		wantMatch  bool
		wantGroups []Capture
	}{
		{
			name:    "Go/language_constants_nil",
			grammar: "go.json > repository > language_constants",
			pattern: `\b(?:(true|false)|(nil)|(iota))\b`,
			input:   `if ready && count > 0 { return nil }`,
			wantMatch: true,
			wantGroups: []Capture{
				{31, 34},   // group 0: "nil"
				{-1, -1},   // group 1: (true|false) — unmatched
				{31, 34},   // group 2: (nil)
				{-1, -1},   // group 3: (iota) — unmatched
			},
		},
		{
			name:    "Go/language_constants_true",
			grammar: "go.json > repository > language_constants",
			pattern: `\b(?:(true|false)|(nil)|(iota))\b`,
			input:   `var ok = true`,
			wantMatch: true,
			wantGroups: []Capture{
				{9, 13},    // group 0: "true"
				{9, 13},    // group 1: (true|false)
				{-1, -1},   // group 2: (nil) — unmatched
				{-1, -1},   // group 3: (iota) — unmatched
			},
		},
		{
			name:    "Go/line_comment",
			grammar: "go.json > begin pattern for comments",
			pattern: `//`,
			input:   `x := 42 // comment here`,
			wantMatch: true,
			wantGroups: []Capture{
				{8, 10},
			},
		},
		{
			name:    "JS/trycatch_keywords",
			grammar: "javascript.json > repository > control-statement",
			pattern: `(?<![_$[:alnum:]])(?:(?<=\.\.\.)|(?<!\.))(catch|finally|throw|try)(?![_$[:alnum:]])(?:(?=\.\.\.)|(?!\.))`,
			input:   `try { fetch() } catch (e) { throw e }`,
			wantMatch: true,
			wantGroups: []Capture{
				{0, 3}, // group 0: "try"
				{0, 3}, // group 1: capture "try"
			},
		},
		{
			name:    "JS/jsx_assignment_operator",
			grammar: "javascript.json > repository > jsx-tag-attribute-assignment",
			pattern: `=(?=\s*(?:'|"|\{|/\*|//|\n))`,
			input:   `className="active"`,
			wantMatch: true,
			wantGroups: []Capture{
				{9, 10},
			},
		},
		{
			name:    "Python/builtin_print",
			grammar: "python.json > repository > builtin-functions",
			pattern: "(?x)\n  (?<!\\.) \\b(\n    __import__ | abs | aiter | all | any | anext | ascii | bin\n    | breakpoint | callable | chr | compile | copyright | credits\n    | delattr | dir | divmod | enumerate | eval | exec | exit\n    | filter | format | getattr | globals | hasattr | hash | help\n    | hex | id | input | isinstance | issubclass | iter | len\n    | license | locals | map | max | memoryview | min | next\n    | oct | open | ord | pow | print | quit | range | reload | repr\n    | reversed | round | setattr | sorted | sum | vars | zip\n  )\\b\n",
			input:   `result = print("hello")`,
			wantMatch: true,
			wantGroups: []Capture{
				{9, 14},  // group 0: "print"
				{9, 14},  // group 1: capture "print"
			},
		},
		{
			name:    "Python/builtin_no_method_call",
			grammar: "python.json > repository > builtin-functions",
			pattern: "(?x)\n  (?<!\\.) \\b(\n    __import__ | abs | aiter | all | any | anext | ascii | bin\n    | breakpoint | callable | chr | compile | copyright | credits\n    | delattr | dir | divmod | enumerate | eval | exec | exit\n    | filter | format | getattr | globals | hasattr | hash | help\n    | hex | id | input | isinstance | issubclass | iter | len\n    | license | locals | map | max | memoryview | min | next\n    | oct | open | ord | pow | print | quit | range | reload | repr\n    | reversed | round | setattr | sorted | sum | vars | zip\n  )\\b\n",
			input:   `obj.print("hello")`,
			wantMatch: false,
		},
		{
			name:    "Python/decorator",
			grammar: "python.json > repository > decorator",
			pattern: "(?x)\n  ^\\s*\n  ((@)) \\s* (?=[[:alpha:]]\\w*)\n",
			input:   "@staticmethod\ndef hello():",
			wantMatch: true,
			wantGroups: []Capture{
				{0, 1},  // group 0: "@"
				{0, 1},  // group 1: outer capture
				{0, 1},  // group 2: inner capture
			},
		},
		{
			name:    "Go/no_match_keyword_in_identifier",
			grammar: "go.json > repository > keywords",
			pattern: `\b(break|case|continue|default|defer|else|fallthrough|for|go|goto|if|range|return|select|switch)\b`,
			input:   `forLoop := 1`,
			wantMatch: false,
		},
		{
			name:    "Go/keyword_in_source",
			grammar: "go.json > repository > keywords",
			pattern: `\b(break|case|continue|default|defer|else|fallthrough|for|go|goto|if|range|return|select|switch)\b`,
			input:   `for i := range items {`,
			wantMatch: true,
			wantGroups: []Capture{
				{0, 3}, // group 0: "for"
				{0, 3}, // group 1: capture "for"
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			_, scanner := newTestScanner(t, tc.pattern)

			m, err := scanner.FindNextMatch(ctx, []byte(tc.input), tc.startPos, SearchOptionNone)
			if err != nil {
				t.Fatalf("FindNextMatch: %v", err)
			}

			if tc.wantMatch && m == nil {
				t.Fatalf("expected match but got none")
			}
			if !tc.wantMatch && m != nil {
				t.Fatalf("expected no match but got index=%d captures=%v", m.Index, m.Captures)
			}
			if !tc.wantMatch {
				return
			}

			if tc.wantGroups != nil {
				if len(m.Captures) != len(tc.wantGroups) {
					t.Fatalf("expected %d groups, got %d", len(tc.wantGroups), len(m.Captures))
				}
				for i, want := range tc.wantGroups {
					got := m.Captures[i]
					if got.Start != want.Start || got.End != want.End {
						matched := ""
						if got.Start >= 0 && got.End >= 0 && got.End <= len(tc.input) {
							matched = tc.input[got.Start:got.End]
						}
						t.Errorf("group %d: got [%d:%d] %q, want [%d:%d]",
							i, got.Start, got.End, matched, want.Start, want.End)
					}
				}
			}
		})
	}
}

func TestRealGrammarMultiPatternScan(t *testing.T) {
	ctx := context.Background()

	goKeywordPatterns := []struct {
		name    string
		pattern string
	}{
		{"keyword.control.go", `\b(break|case|continue|default|defer|else|fallthrough|for|go|goto|if|range|return|select|switch)\b`},
		{"keyword.channel.go", `\bchan\b`},
		{"keyword.const.go", `\bconst\b`},
		{"keyword.var.go", `\bvar\b`},
		{"keyword.function.go", `\bfunc\b`},
		{"keyword.interface.go", `\binterface\b`},
	}

	patterns := make([]string, len(goKeywordPatterns))
	for i, p := range goKeywordPatterns {
		patterns[i] = p.pattern
	}

	_, scanner := newTestScanner(t, patterns...)
	input := []byte(`func main() { var x = "hello" }`)

	m, err := scanner.FindNextMatch(ctx, input, 0, SearchOptionNone)
	if err != nil {
		t.Fatalf("FindNextMatch: %v", err)
	}
	if m == nil {
		t.Fatal("no pattern matched")
	}

	// "func" starts at byte 0, pattern index 4 (keyword.function.go)
	if goKeywordPatterns[m.Index].name != "keyword.function.go" {
		t.Errorf("expected leftmost match to be keyword.function.go, got %s (index %d)",
			goKeywordPatterns[m.Index].name, m.Index)
	}
	if m.Captures[0].Start != 0 {
		t.Errorf("expected match at byte 0, got %d", m.Captures[0].Start)
	}
	if m.Captures[0].End != 4 {
		t.Errorf("expected match end at byte 4, got %d", m.Captures[0].End)
	}
	t.Logf("leftmost winner: %s at byte %d", goKeywordPatterns[m.Index].name, m.Captures[0].Start)
}
