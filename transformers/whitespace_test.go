package transformers_test

import (
	"context"
	"strings"
	"testing"

	nuri "github.com/frostybee/nuri"
	"github.com/frostybee/nuri/transformers"
)

func TestWhitespaceTabs(t *testing.T) {
	h := newTestHighlighter(t)
	html, err := h.CodeToHTML(context.Background(), "\tx := 1\n", nuri.CodeToHTMLOptions{
		Lang:         "go",
		Theme:        "github-dark",
		Transformers: []nuri.Transformer{transformers.Whitespace()},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, `class="ws-tab"`) {
		t.Errorf("missing ws-tab class:\n%s", html)
	}
	if !strings.Contains(html, "→") {
		t.Errorf("missing tab symbol:\n%s", html)
	}
}

func TestWhitespaceSpaces(t *testing.T) {
	h := newTestHighlighter(t)
	html, err := h.CodeToHTML(context.Background(), "x = 1\n", nuri.CodeToHTMLOptions{
		Lang:         "go",
		Theme:        "github-dark",
		Transformers: []nuri.Transformer{transformers.Whitespace()},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, `class="ws-space"`) {
		t.Errorf("missing ws-space class:\n%s", html)
	}
	if !strings.Contains(html, "·") {
		t.Errorf("missing space symbol:\n%s", html)
	}
}

func TestWhitespaceMixed(t *testing.T) {
	h := newTestHighlighter(t)
	html, err := h.CodeToHTML(context.Background(), "\t x\n", nuri.CodeToHTMLOptions{
		Lang:         "go",
		Theme:        "github-dark",
		Transformers: []nuri.Transformer{transformers.Whitespace()},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, "ws-tab") {
		t.Errorf("missing tab:\n%s", html)
	}
	if !strings.Contains(html, "ws-space") {
		t.Errorf("missing space:\n%s", html)
	}
}

func TestWhitespaceCustomChars(t *testing.T) {
	h := newTestHighlighter(t)
	html, err := h.CodeToHTML(context.Background(), "\tx = 1\n", nuri.CodeToHTMLOptions{
		Lang:  "go",
		Theme: "github-dark",
		Transformers: []nuri.Transformer{transformers.WhitespaceWith(transformers.WhitespaceOptions{
			Tab:   "⇥",
			Space: "␣",
		})},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, "⇥") {
		t.Errorf("missing custom tab symbol:\n%s", html)
	}
	if !strings.Contains(html, "␣") {
		t.Errorf("missing custom space symbol:\n%s", html)
	}
}

func TestWhitespaceNoWhitespace(t *testing.T) {
	h := newTestHighlighter(t)
	code := "x\n"

	baseline, err := h.CodeToHTML(context.Background(), code, nuri.CodeToHTMLOptions{
		Lang:  "go",
		Theme: "github-dark",
	})
	if err != nil {
		t.Fatal(err)
	}
	withWS, err := h.CodeToHTML(context.Background(), code, nuri.CodeToHTMLOptions{
		Lang:         "go",
		Theme:        "github-dark",
		Transformers: []nuri.Transformer{transformers.Whitespace()},
	})
	if err != nil {
		t.Fatal(err)
	}
	if baseline != withWS {
		t.Errorf("whitespace transformer changed output without whitespace:\nbaseline: %s\nwith:     %s", baseline, withWS)
	}
}
