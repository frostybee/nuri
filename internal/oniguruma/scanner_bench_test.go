package oniguruma

import (
	"context"
	"fmt"
	"testing"
)

func BenchmarkScannerCreate(b *testing.B) {
	ctx := context.Background()
	inst, err := testEngine.newInstance(ctx)
	if err != nil {
		b.Fatal(err)
	}
	defer inst.close(ctx)

	cases := []struct {
		name     string
		patterns [][]byte
	}{
		{"1_pattern", [][]byte{[]byte(`\w+`)}},
		{"5_patterns", [][]byte{
			[]byte(`\b(break|case|continue|default|defer|else|fallthrough|for|go|goto|if|range|return|select|switch)\b`),
			[]byte(`\bchan\b`),
			[]byte(`\bconst\b`),
			[]byte(`\bvar\b`),
			[]byte(`\bfunc\b`),
		}},
		{"50_patterns", makePatterns(50)},
		{"200_patterns", makePatterns(200)},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for b.Loop() {
				scanner, err := inst.NewScanner(ctx, tc.patterns)
				if err != nil {
					b.Fatal(err)
				}
				scanner.Free(ctx)
			}
		})
	}
}

func BenchmarkFindNextMatch(b *testing.B) {
	ctx := context.Background()
	inst, err := testEngine.newInstance(ctx)
	if err != nil {
		b.Fatal(err)
	}
	defer inst.close(ctx)

	shortText := []byte("func main() { var x int }")
	longText := makeLongText()

	cases := []struct {
		name     string
		patterns [][]byte
		text     []byte
	}{
		{"1_pattern_short", [][]byte{[]byte(`\w+`)}, shortText},
		{"1_pattern_long", [][]byte{[]byte(`\w+`)}, longText},
		{"5_patterns_short", [][]byte{
			[]byte(`\b(break|case|continue|for|if|return)\b`),
			[]byte(`\bfunc\b`),
			[]byte(`\bvar\b`),
			[]byte(`\bconst\b`),
			[]byte(`\btype\b`),
		}, shortText},
		{"5_patterns_long", [][]byte{
			[]byte(`\b(break|case|continue|for|if|return)\b`),
			[]byte(`\bfunc\b`),
			[]byte(`\bvar\b`),
			[]byte(`\bconst\b`),
			[]byte(`\btype\b`),
		}, longText},
		{"no_match", [][]byte{[]byte(`ZZZZZ_NEVER_MATCH`)}, shortText},
		{"utf8_multibyte", [][]byte{[]byte(`\d+`)}, []byte("変数 = 42; 結果 = 100")},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			scanner, err := inst.NewScanner(ctx, tc.patterns)
			if err != nil {
				b.Fatal(err)
			}
			defer scanner.Free(ctx)

			b.SetBytes(int64(len(tc.text)))
			b.ResetTimer()
			for b.Loop() {
				scanner.FindNextMatch(ctx, tc.text, 0, SearchOptionNone)
			}
		})
	}
}

func BenchmarkFindNextMatchIterative(b *testing.B) {
	ctx := context.Background()
	inst, err := testEngine.newInstance(ctx)
	if err != nil {
		b.Fatal(err)
	}
	defer inst.close(ctx)

	patterns := [][]byte{
		[]byte(`\b(break|case|continue|default|defer|else|fallthrough|for|go|goto|if|range|return|select|switch)\b`),
		[]byte(`\bfunc\b`),
		[]byte(`\bvar\b`),
		[]byte(`\bconst\b`),
		[]byte(`\btype\b`),
	}

	scanner, err := inst.NewScanner(ctx, patterns)
	if err != nil {
		b.Fatal(err)
	}
	defer scanner.Free(ctx)

	text := []byte("func main() { var x = 42; if x > 0 { return x } }")
	b.SetBytes(int64(len(text)))
	b.ResetTimer()

	for b.Loop() {
		pos := 0
		for {
			m, _ := scanner.FindNextMatch(ctx, text, pos, SearchOptionNone)
			if m == nil {
				break
			}
			pos = m.Captures[0].End
		}
	}
}

func BenchmarkPoolGetPut(b *testing.B) {
	ctx := context.Background()

	b.Run("uncontended", func(b *testing.B) {
		pool, err := NewPool(ctx, testEngine, 4)
		if err != nil {
			b.Fatal(err)
		}
		defer pool.Close(ctx)

		b.ResetTimer()
		for b.Loop() {
			inst, _ := pool.Get(ctx)
			pool.Put(inst)
		}
	})

	for _, goroutines := range []int{4, 16} {
		b.Run(fmt.Sprintf("contended_%dg", goroutines), func(b *testing.B) {
			pool, err := NewPool(ctx, testEngine, 4)
			if err != nil {
				b.Fatal(err)
			}
			defer pool.Close(ctx)

			b.ResetTimer()
			b.SetParallelism(goroutines)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					inst, _ := pool.Get(ctx)
					pool.Put(inst)
				}
			})
		})
	}
}

func BenchmarkPoolRoundTrip(b *testing.B) {
	ctx := context.Background()
	pool, err := NewPool(ctx, testEngine, 4)
	if err != nil {
		b.Fatal(err)
	}
	defer pool.Close(ctx)

	patterns := [][]byte{[]byte(`\bfunc\b`), []byte(`\bvar\b`), []byte(`\bconst\b`)}
	text := []byte("func main() { var x = 1 }")

	b.ResetTimer()
	for b.Loop() {
		inst, _ := pool.Get(ctx)
		scanner, _ := inst.NewScanner(ctx, patterns)
		scanner.FindNextMatch(ctx, text, 0, SearchOptionNone)
		scanner.Free(ctx)
		pool.Put(inst)
	}
}

func BenchmarkWASMCallOverhead(b *testing.B) {
	ctx := context.Background()
	inst, err := testEngine.newInstance(ctx)
	if err != nil {
		b.Fatal(err)
	}
	defer inst.close(ctx)

	b.ResetTimer()
	for b.Loop() {
		ptr, _ := inst.wasmAlloc(ctx, 64)
		inst.wasmFree(ctx, ptr)
	}
}

func makePatterns(n int) [][]byte {
	pats := make([][]byte, n)
	for i := range n {
		pats[i] = []byte(fmt.Sprintf(`\bpattern_%d\b`, i))
	}
	return pats
}

func makeLongText() []byte {
	var b []byte
	for range 20 {
		b = append(b, []byte("func processData(ctx context.Context, items []Item) error {\n\tfor i, item := range items {\n\t\tif item.Valid() {\n\t\t\tcontinue\n\t\t}\n\t\treturn fmt.Errorf(\"invalid item at %d\", i)\n\t}\n\treturn nil\n}\n")...)
	}
	return b
}
