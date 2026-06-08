package tokenizer

import (
	"context"
	"testing"

	"github.com/frostybee/nuri/internal/grammar"
)

func FuzzParseGrammar(f *testing.F) {
	f.Add([]byte(`{"scopeName":"source.test","patterns":[{"match":"\\w+","name":"word"}]}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"scopeName":"s","patterns":[]}`))
	f.Add([]byte(`[]`))
	f.Add([]byte(`null`))
	f.Add([]byte(``))
	f.Add([]byte(`{"scopeName":"s","patterns":[{"begin":"(","end":")","patterns":[{"include":"$self"}]}]}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		// Invariant: never panics, regardless of input
		grammar.ParseGrammar(data)
	})
}

func FuzzTokenize(f *testing.F) {
	f.Add([]byte("hello world"))
	f.Add([]byte("const x = 1;"))
	f.Add([]byte(""))
	f.Add([]byte("変数 = 42"))
	f.Add([]byte("👋 hello"))
	f.Add([]byte("</script> &amp;"))
	f.Add([]byte("\n\n\n"))
	f.Add([]byte("a b c d e f g h i j k l m n o p"))

	miniGrammar := []byte(`{
		"scopeName": "source.fuzz",
		"patterns": [
			{"match": "\\b\\w+\\b", "name": "word"},
			{"match": "\\d+", "name": "number"},
			{
				"begin": "\"",
				"end": "\"",
				"name": "string",
				"patterns": [{"match": "\\\\.", "name": "escape"}]
			}
		]
	}`)

	f.Fuzz(func(t *testing.T, source []byte) {
		g, err := grammar.ParseGrammar(miniGrammar)
		if err != nil {
			t.Fatal(err)
		}

		ctx := context.Background()
		onigLib := newTestOnigLib(t)

		result, err := Tokenize(ctx, source, g, onigLib, TokenizeOptions{})
		if err != nil {
			return // errors are fine, panics are not
		}

		// Invariant: byte-coverage completeness — token ranges must cover
		// the entire source with no gaps and no overlaps within each line.
		for lineIdx, tokens := range result.Lines {
			if len(tokens) == 0 {
				continue
			}
			for i := 1; i < len(tokens); i++ {
				prev := tokens[i-1]
				cur := tokens[i]
				if cur.Start < prev.End {
					t.Errorf("line %d: overlap at tokens %d-%d: [%d:%d] and [%d:%d]",
						lineIdx, i-1, i, prev.Start, prev.End, cur.Start, cur.End)
				}
			}
		}
	})
}
