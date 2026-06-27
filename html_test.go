package nuri

import (
	"context"
	"strings"
	"testing"
)

// --- CodeToHTML integration ---

func TestCodeToHTMLGoGitHubDark(t *testing.T) {
	h := newTestHighlighter(t)
	html, err := h.CodeToHTML(context.Background(), "package main\n", CodeToHTMLOptions{
		Lang:  "go",
		Theme: "github-dark",
	})
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(html, `<pre class="nuri github-dark"`) {
		t.Errorf("missing <pre> with nuri class:\n%s", html)
	}
	if !strings.Contains(html, `style="background-color:#24292e;color:#e1e4e8"`) {
		t.Errorf("missing theme styles:\n%s", html)
	}
	if !strings.Contains(html, `tabindex="0"`) {
		t.Errorf("missing tabindex:\n%s", html)
	}
	if !strings.Contains(html, `<code>`) {
		t.Errorf("missing <code>:\n%s", html)
	}
	if !strings.Contains(html, `<span class="line">`) {
		t.Errorf("missing line span:\n%s", html)
	}
	if !strings.Contains(html, `color:#f97583">package</span>`) {
		if !strings.Contains(html, `color:#F97583">package</span>`) {
			t.Errorf("expected keyword 'package' with color #f97583:\n%s", html)
		}
	}
}

func TestCodeToHTMLJavaScript(t *testing.T) {
	h := newTestHighlighter(t)
	html, err := h.CodeToHTML(context.Background(), "const x = 1;\n", CodeToHTMLOptions{
		Lang:  "javascript",
		Theme: "github-dark",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, ">const</span>") {
		t.Errorf("missing 'const' token:\n%s", html)
	}
}

func TestCodeToHTMLMultiLine(t *testing.T) {
	h := newTestHighlighter(t)
	html, err := h.CodeToHTML(context.Background(), "package main\n\nfunc main() {\n}\n", CodeToHTMLOptions{
		Lang:  "go",
		Theme: "github-dark",
	})
	if err != nil {
		t.Fatal(err)
	}

	lineCount := strings.Count(html, `<span class="line">`)
	if lineCount < 4 {
		t.Errorf("expected at least 4 line spans, got %d:\n%s", lineCount, html)
	}
	if !strings.Contains(html, "</span>\n<span class=\"line\">") {
		t.Errorf("missing newline between line spans:\n%s", html)
	}
	if !strings.Contains(html, `<span class="line"></span>`) {
		t.Errorf("missing empty line span:\n%s", html)
	}
}

func TestCodeToHTMLEscaping(t *testing.T) {
	h := newTestHighlighter(t)
	html, err := h.CodeToHTML(context.Background(), "x := \"<b>\" + \"&\" // </script>\n", CodeToHTMLOptions{
		Lang:  "go",
		Theme: "github-dark",
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(html, "</script>") {
		t.Errorf("unescaped </script> in output:\n%s", html)
	}
	if !strings.Contains(html, "&lt;b&gt;") {
		t.Errorf("expected escaped <b>:\n%s", html)
	}
	if !strings.Contains(html, "&amp;") {
		t.Errorf("expected escaped &:\n%s", html)
	}
}

func TestCodeToHTMLEmptyInput(t *testing.T) {
	h := newTestHighlighter(t)
	html, err := h.CodeToHTML(context.Background(), "", CodeToHTMLOptions{
		Lang:  "go",
		Theme: "github-dark",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, "<code></code>") {
		t.Errorf("expected empty <code>:\n%s", html)
	}
}

func TestCodeToHTMLUnknownLang(t *testing.T) {
	h := newTestHighlighter(t)
	html, err := h.CodeToHTML(context.Background(), "hello world\n", CodeToHTMLOptions{
		Lang:  "not-a-real-language-xyz",
		Theme: "github-dark",
	})
	if err != nil {
		t.Fatalf("expected no error for unknown lang, got: %v", err)
	}
	if !strings.Contains(html, "hello world") {
		t.Errorf("expected plaintext content:\n%s", html)
	}
	if !strings.Contains(html, `<span class="line">`) {
		t.Errorf("expected line span:\n%s", html)
	}
}

// --- Decorations ---

func TestHighlightLines(t *testing.T) {
	h := newTestHighlighter(t)
	html, err := h.CodeToHTML(context.Background(), "a\nb\nc\n", CodeToHTMLOptions{
		Lang:           "go",
		Theme:          "github-dark",
		HighlightLines: []LineRange{Range(1, 1), Range(3, 3)},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, `class="line highlighted"`) {
		t.Errorf("missing highlighted class:\n%s", html)
	}
	count := strings.Count(html, "highlighted")
	if count != 2 {
		t.Errorf("expected 2 highlighted lines, got %d:\n%s", count, html)
	}
}

func TestFocusLines(t *testing.T) {
	h := newTestHighlighter(t)
	html, err := h.CodeToHTML(context.Background(), "a\nb\nc\n", CodeToHTMLOptions{
		Lang:       "go",
		Theme:      "github-dark",
		FocusLines: []LineRange{Range(2, 2)},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, "has-focused") {
		t.Errorf("pre missing has-focused:\n%s", html)
	}
	if !strings.Contains(html, `class="line focused"`) {
		t.Errorf("missing focused class:\n%s", html)
	}
	if strings.Count(html, "dimmed") != 2 {
		t.Errorf("expected 2 dimmed lines:\n%s", html)
	}
}

func TestInsertedLines(t *testing.T) {
	h := newTestHighlighter(t)
	html, err := h.CodeToHTML(context.Background(), "a\nb\n", CodeToHTMLOptions{
		Lang:          "go",
		Theme:         "github-dark",
		InsertedLines: []LineRange{Range(1, 1)},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, `class="line diff add"`) {
		t.Errorf("missing diff add classes:\n%s", html)
	}
}

func TestDeletedLines(t *testing.T) {
	h := newTestHighlighter(t)
	html, err := h.CodeToHTML(context.Background(), "a\nb\n", CodeToHTMLOptions{
		Lang:         "go",
		Theme:        "github-dark",
		DeletedLines: []LineRange{Range(2, 2)},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, `class="line diff remove"`) {
		t.Errorf("missing diff remove classes:\n%s", html)
	}
}

func TestCustomPreClass(t *testing.T) {
	h := newTestHighlighter(t)
	html, err := h.CodeToHTML(context.Background(), "x\n", CodeToHTMLOptions{
		Lang:     "go",
		Theme:    "github-dark",
		PreClass: "my-pre",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, "my-pre") {
		t.Errorf("missing PreClass:\n%s", html)
	}
}

func TestCustomCodeAttrs(t *testing.T) {
	h := newTestHighlighter(t)
	html, err := h.CodeToHTML(context.Background(), "x\n", CodeToHTMLOptions{
		Lang:      "go",
		Theme:     "github-dark",
		CodeClass: "lang-go",
		CodeAttrs: map[string]string{"data-lang": "go"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, `class="lang-go"`) {
		t.Errorf("missing CodeClass:\n%s", html)
	}
	if !strings.Contains(html, `data-lang="go"`) {
		t.Errorf("missing CodeAttrs:\n%s", html)
	}
}

func TestNoDecorationsBackwardCompat(t *testing.T) {
	h := newTestHighlighter(t)
	code := "package main\n"

	html, err := h.CodeToHTML(context.Background(), code, CodeToHTMLOptions{
		Lang:  "go",
		Theme: "github-dark",
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, cls := range []string{"highlighted", "focused", "dimmed", "has-focused", "diff", "add", "remove"} {
		if strings.Contains(html, cls) {
			t.Errorf("unexpected class %q in output without decorations:\n%s", cls, html)
		}
	}
}

func TestCodeToHTMLFontStyle(t *testing.T) {
	h := newTestHighlighter(t)
	html, err := h.CodeToHTML(context.Background(), "*italic text*\n", CodeToHTMLOptions{
		Lang:  "markdown",
		Theme: "github-dark",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, `<span class="line">`) {
		t.Errorf("expected line span:\n%s", html)
	}
	t.Logf("Font style HTML:\n%s", html)
}
