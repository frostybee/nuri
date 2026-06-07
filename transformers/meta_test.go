package transformers_test

import (
	"context"
	"os"
	"strings"
	"testing"

	nuri "github.com/frostybee/nuri"
	"github.com/frostybee/nuri/internal/shared"
	"github.com/frostybee/nuri/transformers"
)

func newTestHighlighter(t *testing.T) *nuri.Highlighter {
	t.Helper()
	ctx := context.Background()
	h, err := nuri.New(ctx,
		nuri.WithGrammarFS(os.DirFS(shared.GrammarsDir(t))),
		nuri.WithThemeFS(os.DirFS(shared.ThemesDir(t))),
		nuri.WithPoolSize(1),
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { h.Close(ctx) })
	return h
}

func TestParseMetaRangesSingle(t *testing.T) {
	ranges, err := transformers.ParseMetaRanges("{3}")
	if err != nil {
		t.Fatal(err)
	}
	if len(ranges) != 1 || ranges[0].Start != 3 || ranges[0].End != 3 {
		t.Errorf("got %v, want [{3,3}]", ranges)
	}
}

func TestParseMetaRangesMultiple(t *testing.T) {
	ranges, err := transformers.ParseMetaRanges("{1,3-5,7}")
	if err != nil {
		t.Fatal(err)
	}
	if len(ranges) != 3 {
		t.Fatalf("got %d ranges, want 3", len(ranges))
	}
	want := []nuri.LineRange{{Start: 1, End: 1}, {Start: 3, End: 5}, {Start: 7, End: 7}}
	for i, r := range ranges {
		if r != want[i] {
			t.Errorf("range[%d] = %v, want %v", i, r, want[i])
		}
	}
}

func TestParseMetaRangesEmpty(t *testing.T) {
	for _, input := range []string{"", "{}", "  ", "no braces"} {
		ranges, err := transformers.ParseMetaRanges(input)
		if err != nil {
			t.Errorf("ParseMetaRanges(%q) error: %v", input, err)
		}
		if len(ranges) != 0 {
			t.Errorf("ParseMetaRanges(%q) = %v, want nil", input, ranges)
		}
	}
}

func TestParseMetaRangesInvalid(t *testing.T) {
	_, err := transformers.ParseMetaRanges("{abc}")
	if err == nil {
		t.Error("expected error for non-numeric input")
	}
}

func TestMetaTransformerHighlightsLines(t *testing.T) {
	h := newTestHighlighter(t)
	html, err := h.CodeToHTML(context.Background(), "a\nb\nc\nd\ne\n", nuri.CodeToHTMLOptions{
		Lang:         "go",
		Theme:        "github-dark",
		Transformers: []nuri.Transformer{transformers.Meta("{1,3-5}")},
	})
	if err != nil {
		t.Fatal(err)
	}
	count := strings.Count(html, "highlighted")
	if count != 4 {
		t.Errorf("expected 4 highlighted lines (1,3,4,5), got %d:\n%s", count, html)
	}
}
