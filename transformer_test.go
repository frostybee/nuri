package nuri

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

// Compile-time check: BaseTransformer satisfies Transformer.
var _ Transformer = BaseTransformer{}

func TestBaseTransformerInterface(t *testing.T) {
	var tr Transformer = BaseTransformer{}
	if tr.Name() != "" {
		t.Error("expected empty name")
	}
}

// --- Preprocess ---

type preprocessTransformer struct {
	BaseTransformer
}

func (preprocessTransformer) Name() string { return "preprocess" }
func (preprocessTransformer) Preprocess(code string, opts *CodeToHTMLOptions) string {
	return strings.ReplaceAll(code, "REPLACE_ME", "replaced")
}

func TestTransformerPreprocess(t *testing.T) {
	h := newTestHighlighter(t)
	html, err := h.CodeToHTML(context.Background(), "REPLACE_ME\n", CodeToHTMLOptions{
		Lang:         "go",
		Theme:        "github-dark",
		Transformers: []Transformer{preprocessTransformer{}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(html, "REPLACE_ME") {
		t.Error("preprocess hook did not modify source")
	}
	if !strings.Contains(html, "replaced") {
		t.Errorf("expected 'replaced' in output:\n%s", html)
	}
}

// --- Line ---

type lineAttrTransformer struct {
	BaseTransformer
}

func (lineAttrTransformer) Name() string { return "line-attr" }
func (lineAttrTransformer) Line(el *Element, line int) *Element {
	el.SetAttr("data-line", fmt.Sprintf("%d", line))
	return el
}

func TestTransformerLine(t *testing.T) {
	h := newTestHighlighter(t)
	html, err := h.CodeToHTML(context.Background(), "a\nb\n", CodeToHTMLOptions{
		Lang:         "go",
		Theme:        "github-dark",
		Transformers: []Transformer{lineAttrTransformer{}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, `data-line="1"`) {
		t.Errorf("missing data-line=\"1\":\n%s", html)
	}
	if !strings.Contains(html, `data-line="2"`) {
		t.Errorf("missing data-line=\"2\":\n%s", html)
	}
}

// --- Span ---

type spanClassTransformer struct {
	BaseTransformer
}

func (spanClassTransformer) Name() string { return "span-class" }
func (spanClassTransformer) Span(el *Element, line, col int, lineEl *Element, tok ThemedToken) *Element {
	el.AddClass("token")
	return el
}

func TestTransformerSpan(t *testing.T) {
	h := newTestHighlighter(t)
	html, err := h.CodeToHTML(context.Background(), "const x = 1;\n", CodeToHTMLOptions{
		Lang:         "javascript",
		Theme:        "github-dark",
		Transformers: []Transformer{spanClassTransformer{}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, `class="token"`) {
		t.Errorf("span missing 'token' class:\n%s", html)
	}
}

// --- Span byte col ---

type spanByteColRecorder struct {
	BaseTransformer
	cols [][]int
}

func (r *spanByteColRecorder) Name() string { return "col-recorder" }
func (r *spanByteColRecorder) Span(el *Element, line, col int, lineEl *Element, tok ThemedToken) *Element {
	for len(r.cols) < line {
		r.cols = append(r.cols, nil)
	}
	r.cols[line-1] = append(r.cols[line-1], col)
	return nil
}

func TestTransformerSpanByteCol(t *testing.T) {
	h := newTestHighlighter(t)
	recorder := &spanByteColRecorder{}
	_, err := h.CodeToHTML(context.Background(), "ab\ncd\n", CodeToHTMLOptions{
		Lang:         "go",
		Theme:        "github-dark",
		Transformers: []Transformer{recorder},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(recorder.cols) < 2 {
		t.Fatalf("expected at least 2 lines, got %d", len(recorder.cols))
	}
	// First token on each line should start at col 0.
	if recorder.cols[0][0] != 0 {
		t.Errorf("line 1 first col = %d, want 0", recorder.cols[0][0])
	}
	if recorder.cols[1][0] != 0 {
		t.Errorf("line 2 first col = %d, want 0", recorder.cols[1][0])
	}
}

// --- Code ---

type codeClassTransformer struct {
	BaseTransformer
}

func (codeClassTransformer) Name() string { return "code-class" }
func (codeClassTransformer) Code(el *Element) *Element {
	el.AddClass("custom-code")
	return el
}

func TestTransformerCode(t *testing.T) {
	h := newTestHighlighter(t)
	html, err := h.CodeToHTML(context.Background(), "x\n", CodeToHTMLOptions{
		Lang:         "go",
		Theme:        "github-dark",
		Transformers: []Transformer{codeClassTransformer{}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, `class="custom-code"`) {
		t.Errorf("code element missing class:\n%s", html)
	}
}

// --- Pre ---

type preClassTransformer struct {
	BaseTransformer
}

func (preClassTransformer) Name() string { return "pre-class" }
func (preClassTransformer) Pre(el *Element) *Element {
	el.AddClass("custom-pre")
	return el
}

func TestTransformerPre(t *testing.T) {
	h := newTestHighlighter(t)
	html, err := h.CodeToHTML(context.Background(), "x\n", CodeToHTMLOptions{
		Lang:         "go",
		Theme:        "github-dark",
		Transformers: []Transformer{preClassTransformer{}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, "custom-pre") {
		t.Errorf("pre element missing class:\n%s", html)
	}
}

// --- Postprocess ---

type postprocessTransformer struct {
	BaseTransformer
}

func (postprocessTransformer) Name() string { return "postprocess" }
func (postprocessTransformer) Postprocess(html string, opts *CodeToHTMLOptions) string {
	return html + "<!-- postprocessed -->"
}

func TestTransformerPostprocess(t *testing.T) {
	h := newTestHighlighter(t)
	html, err := h.CodeToHTML(context.Background(), "x\n", CodeToHTMLOptions{
		Lang:         "go",
		Theme:        "github-dark",
		Transformers: []Transformer{postprocessTransformer{}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(html, "<!-- postprocessed -->") {
		t.Errorf("postprocess hook not applied:\n%s", html)
	}
}

// --- Chaining ---

type addAttrTransformer struct {
	BaseTransformer
	key, val string
}

func (a addAttrTransformer) Name() string { return "add-attr-" + a.key }
func (a addAttrTransformer) Pre(el *Element) *Element {
	el.SetAttr(a.key, a.val)
	return el
}

func TestTransformerChaining(t *testing.T) {
	h := newTestHighlighter(t)
	html, err := h.CodeToHTML(context.Background(), "x\n", CodeToHTMLOptions{
		Lang:  "go",
		Theme: "github-dark",
		Transformers: []Transformer{
			addAttrTransformer{key: "data-a", val: "1"},
			addAttrTransformer{key: "data-b", val: "2"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, `data-a="1"`) {
		t.Errorf("first transformer attr missing:\n%s", html)
	}
	if !strings.Contains(html, `data-b="2"`) {
		t.Errorf("second transformer attr missing:\n%s", html)
	}
}

// --- Nil return ---

type nilReturnTransformer struct {
	BaseTransformer
}

func (nilReturnTransformer) Name() string                                              { return "nil" }
func (nilReturnTransformer) Span(el *Element, line, col int, lineEl *Element, tok ThemedToken) *Element {
	return nil
}
func (nilReturnTransformer) Line(el *Element, line int) *Element { return nil }
func (nilReturnTransformer) Code(el *Element) *Element           { return nil }
func (nilReturnTransformer) Pre(el *Element) *Element            { return nil }

func TestTransformerNilReturn(t *testing.T) {
	h := newTestHighlighter(t)
	// Get baseline without transformers.
	baseline, err := h.CodeToHTML(context.Background(), "package main\n", CodeToHTMLOptions{
		Lang:  "go",
		Theme: "github-dark",
	})
	if err != nil {
		t.Fatal(err)
	}
	// With nil-returning transformer should produce identical output.
	got, err := h.CodeToHTML(context.Background(), "package main\n", CodeToHTMLOptions{
		Lang:         "go",
		Theme:        "github-dark",
		Transformers: []Transformer{nilReturnTransformer{}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != baseline {
		t.Errorf("nil-return transformer changed output:\ngot:  %s\nwant: %s", got, baseline)
	}
}

// --- Backward compatibility ---

func TestNoTransformersBackwardCompat(t *testing.T) {
	h := newTestHighlighter(t)
	code := "package main\n\nfunc main() {\n}\n"

	// Phase 4 path: no transformers, no decorations.
	html, err := h.CodeToHTML(context.Background(), code, CodeToHTMLOptions{
		Lang:  "go",
		Theme: "github-dark",
	})
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(html, `<pre class="shiki github-dark"`) {
		t.Errorf("missing <pre> with shiki class:\n%s", html)
	}
	if !strings.Contains(html, `<code>`) {
		t.Errorf("missing plain <code>:\n%s", html)
	}
	if !strings.Contains(html, `<span class="line">`) {
		t.Errorf("missing line span:\n%s", html)
	}
}
