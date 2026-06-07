package transformers_test

import (
	"context"
	"strings"
	"testing"

	nuri "github.com/frostybee/nuri"
	"github.com/frostybee/nuri/transformers"
)

func TestNotationDiffAdd(t *testing.T) {
	h := newTestHighlighter(t)
	html, err := h.CodeToHTML(context.Background(), "const x = 1; // [!code ++]\n", nuri.CodeToHTMLOptions{
		Lang:         "javascript",
		Theme:        "github-dark",
		Transformers: []nuri.Transformer{transformers.Notation()},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, "diff") || !strings.Contains(html, "add") {
		t.Errorf("missing diff/add classes:\n%s", html)
	}
}

func TestNotationDiffRemove(t *testing.T) {
	h := newTestHighlighter(t)
	html, err := h.CodeToHTML(context.Background(), "const x = 1; // [!code --]\n", nuri.CodeToHTMLOptions{
		Lang:         "javascript",
		Theme:        "github-dark",
		Transformers: []nuri.Transformer{transformers.Notation()},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, "diff") || !strings.Contains(html, "remove") {
		t.Errorf("missing diff/remove classes:\n%s", html)
	}
}

func TestNotationHighlight(t *testing.T) {
	h := newTestHighlighter(t)
	html, err := h.CodeToHTML(context.Background(), "const x = 1; // [!code highlight]\n", nuri.CodeToHTMLOptions{
		Lang:         "javascript",
		Theme:        "github-dark",
		Transformers: []nuri.Transformer{transformers.Notation()},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, "highlighted") {
		t.Errorf("missing highlighted class:\n%s", html)
	}
}

func TestNotationFocus(t *testing.T) {
	h := newTestHighlighter(t)
	html, err := h.CodeToHTML(context.Background(), "const x = 1; // [!code focus]\nconst y = 2;\n", nuri.CodeToHTMLOptions{
		Lang:         "javascript",
		Theme:        "github-dark",
		Transformers: []nuri.Transformer{transformers.Notation()},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, "focused") {
		t.Errorf("missing focused class:\n%s", html)
	}
	if !strings.Contains(html, "dimmed") {
		t.Errorf("missing dimmed class:\n%s", html)
	}
	if !strings.Contains(html, "has-focused") {
		t.Errorf("missing has-focused on pre:\n%s", html)
	}
}

func TestNotationError(t *testing.T) {
	h := newTestHighlighter(t)
	html, err := h.CodeToHTML(context.Background(), "const x = 1; // [!code error]\n", nuri.CodeToHTMLOptions{
		Lang:         "javascript",
		Theme:        "github-dark",
		Transformers: []nuri.Transformer{transformers.Notation()},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, "highlighted") || !strings.Contains(html, "error") {
		t.Errorf("missing highlighted/error classes:\n%s", html)
	}
}

func TestNotationWarning(t *testing.T) {
	h := newTestHighlighter(t)
	html, err := h.CodeToHTML(context.Background(), "const x = 1; // [!code warning]\n", nuri.CodeToHTMLOptions{
		Lang:         "javascript",
		Theme:        "github-dark",
		Transformers: []nuri.Transformer{transformers.Notation()},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, "warning") {
		t.Errorf("missing warning class:\n%s", html)
	}
}

func TestNotationCommentStripped(t *testing.T) {
	h := newTestHighlighter(t)
	html, err := h.CodeToHTML(context.Background(), "const x = 1; // [!code ++]\n", nuri.CodeToHTMLOptions{
		Lang:         "javascript",
		Theme:        "github-dark",
		Transformers: []nuri.Transformer{transformers.Notation()},
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(html, "[!code") {
		t.Errorf("magic comment not stripped:\n%s", html)
	}
}

func TestNotationSlashComment(t *testing.T) {
	h := newTestHighlighter(t)
	html, err := h.CodeToHTML(context.Background(), "x := 1 // [!code highlight]\n", nuri.CodeToHTMLOptions{
		Lang:         "go",
		Theme:        "github-dark",
		Transformers: []nuri.Transformer{transformers.Notation()},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, "highlighted") {
		t.Errorf("missing highlighted class for // comment:\n%s", html)
	}
	if strings.Contains(html, "[!code") {
		t.Errorf("annotation not stripped:\n%s", html)
	}
}

func TestNotationHashComment(t *testing.T) {
	h := newTestHighlighter(t)
	html, err := h.CodeToHTML(context.Background(), "x = 1 # [!code highlight]\n", nuri.CodeToHTMLOptions{
		Lang:         "python",
		Theme:        "github-dark",
		Transformers: []nuri.Transformer{transformers.Notation()},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, "highlighted") {
		t.Errorf("missing highlighted class for # comment:\n%s", html)
	}
}

func TestNotationMultiple(t *testing.T) {
	h := newTestHighlighter(t)
	code := "const a = 1; // [!code ++]\nconst b = 2; // [!code --]\nconst c = 3;\n"
	html, err := h.CodeToHTML(context.Background(), code, nuri.CodeToHTMLOptions{
		Lang:         "javascript",
		Theme:        "github-dark",
		Transformers: []nuri.Transformer{transformers.Notation()},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, "add") {
		t.Errorf("missing add class:\n%s", html)
	}
	if !strings.Contains(html, "remove") {
		t.Errorf("missing remove class:\n%s", html)
	}
}

func TestNotationNoAnnotations(t *testing.T) {
	h := newTestHighlighter(t)
	code := "const x = 1;\n"

	baseline, err := h.CodeToHTML(context.Background(), code, nuri.CodeToHTMLOptions{
		Lang:  "javascript",
		Theme: "github-dark",
	})
	if err != nil {
		t.Fatal(err)
	}
	withNotation, err := h.CodeToHTML(context.Background(), code, nuri.CodeToHTMLOptions{
		Lang:         "javascript",
		Theme:        "github-dark",
		Transformers: []nuri.Transformer{transformers.Notation()},
	})
	if err != nil {
		t.Fatal(err)
	}
	if baseline != withNotation {
		t.Errorf("notation transformer changed output without annotations:\nbaseline: %s\nwith:     %s", baseline, withNotation)
	}
}

func TestNotationWordHighlight(t *testing.T) {
	h := newTestHighlighter(t)
	code := "// [!code word:hello]\nconst hello = \"hello world\";\n"
	html, err := h.CodeToHTML(context.Background(), code, nuri.CodeToHTMLOptions{
		Lang:         "javascript",
		Theme:        "github-dark",
		Transformers: []nuri.Transformer{transformers.Notation()},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, "highlighted-word") {
		t.Errorf("missing highlighted-word class:\n%s", html)
	}
	if strings.Contains(html, "[!code") {
		t.Errorf("annotation not stripped:\n%s", html)
	}
}
