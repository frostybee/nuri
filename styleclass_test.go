package nuri

import (
	"context"
	"strings"
	"testing"

	"github.com/frostybee/nuri/ast"
)

func TestStyleHashDeterministic(t *testing.T) {
	styles := map[string]string{"color": "#ff0000", "font-style": "italic"}
	first := ast.CanonicalStyles(styles)
	h := ast.StyleHash(first)
	for i := 0; i < 100; i++ {
		c := ast.CanonicalStyles(styles)
		if got := ast.StyleHash(c); got != h {
			t.Fatalf("hash changed at iteration %d: %s vs %s", i, got, h)
		}
	}
}

func TestStyleHashDifferentInputs(t *testing.T) {
	a := ast.StyleHash(ast.CanonicalStyles(map[string]string{"color": "#ff0000"}))
	b := ast.StyleHash(ast.CanonicalStyles(map[string]string{"color": "#00ff00"}))
	if a == b {
		t.Errorf("different styles produced same hash: %s", a)
	}
}

func TestStyleHashKeyOrder(t *testing.T) {
	a := ast.CanonicalStyles(map[string]string{"color": "#ff0000", "font-style": "italic"})
	b := ast.CanonicalStyles(map[string]string{"font-style": "italic", "color": "#ff0000"})
	if a != b {
		t.Errorf("canonical strings differ for same styles: %q vs %q", a, b)
	}
}

func TestStyleClassMapGet(t *testing.T) {
	m := NewStyleClassMap()
	s1 := map[string]string{"color": "#ff0000"}
	s2 := map[string]string{"color": "#00ff00"}

	c1a := m.Get(s1)
	c1b := m.Get(s1)
	c2 := m.Get(s2)

	if c1a != c1b {
		t.Errorf("same styles got different classes: %s vs %s", c1a, c1b)
	}
	if c1a == c2 {
		t.Errorf("different styles got same class: %s", c1a)
	}
	if !strings.HasPrefix(c1a, "_s_") {
		t.Errorf("class missing _s_ prefix: %s", c1a)
	}
}

func TestStyleClassMapCSS(t *testing.T) {
	m := NewStyleClassMap()
	m.Get(map[string]string{"color": "#ff0000"})
	m.Get(map[string]string{"color": "#00ff00", "font-style": "italic"})

	css := m.CSS()
	if !strings.Contains(css, "color: #ff0000") {
		t.Errorf("CSS missing first rule:\n%s", css)
	}
	if !strings.Contains(css, "color: #00ff00") {
		t.Errorf("CSS missing second rule:\n%s", css)
	}
	if !strings.Contains(css, "font-style: italic") {
		t.Errorf("CSS missing font-style:\n%s", css)
	}
	lines := strings.Split(strings.TrimSpace(css), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 CSS rules, got %d:\n%s", len(lines), css)
	}
	if len(lines) == 2 && lines[0] > lines[1] {
		t.Errorf("CSS rules not sorted:\n%s", css)
	}
}

func TestStyleToClassIntegration(t *testing.T) {
	h := newTestHighlighter(t)
	cm := NewStyleClassMap()
	html, err := h.CodeToHTML(context.Background(), "package main\n", CodeToHTMLOptions{
		Lang:     "go",
		Theme:    "github-dark",
		ClassMap: cm,
	})
	if err != nil {
		t.Fatal(err)
	}
	codeStart := strings.Index(html, "<code>")
	codeEnd := strings.Index(html, "</code>")
	if codeStart < 0 || codeEnd < 0 {
		t.Fatalf("missing code element:\n%s", html)
	}
	codeHTML := html[codeStart:codeEnd]
	if strings.Contains(codeHTML, `style="`) {
		t.Errorf("token spans should not have inline styles:\n%s", codeHTML)
	}
	if !strings.Contains(codeHTML, `class="_s_`) {
		t.Errorf("token spans should have _s_ class:\n%s", codeHTML)
	}
	css := cm.CSS()
	if css == "" {
		t.Error("ClassMap.CSS() returned empty string")
	}
	if !strings.Contains(css, "color:") {
		t.Errorf("CSS missing color rules:\n%s", css)
	}
}

func TestStyleToClassMultiBlock(t *testing.T) {
	h := newTestHighlighter(t)
	cm := NewStyleClassMap()

	_, err := h.CodeToHTML(context.Background(), "package main\n", CodeToHTMLOptions{
		Lang: "go", Theme: "github-dark", ClassMap: cm,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = h.CodeToHTML(context.Background(), "const x = 1;\n", CodeToHTMLOptions{
		Lang: "javascript", Theme: "github-dark", ClassMap: cm,
	})
	if err != nil {
		t.Fatal(err)
	}

	css := cm.CSS()
	lines := strings.Split(strings.TrimSpace(css), "\n")
	if len(lines) < 2 {
		t.Errorf("expected at least 2 CSS rules from two blocks, got %d:\n%s", len(lines), css)
	}
}

func TestStyleToClassWithMultiTheme(t *testing.T) {
	h := newTestHighlighter(t)
	cm := NewStyleClassMap()
	html, err := h.CodeToHTML(context.Background(), "package main\n", CodeToHTMLOptions{
		Lang: "go",
		Themes: map[string]string{
			"dark": "github-dark", "light": "github-light",
		},
		ClassMap: cm,
	})
	if err != nil {
		t.Fatal(err)
	}
	css := cm.CSS()
	if !strings.Contains(css, "--nuri-light") {
		t.Errorf("CSS missing multi-theme variables:\n%s", css)
	}
	if strings.Contains(html, `style="`) {
		t.Errorf("should not have inline styles:\n%s", html)
	}
}

func TestStyleToClassBackwardCompat(t *testing.T) {
	h := newTestHighlighter(t)
	baseline, err := h.CodeToHTML(context.Background(), "package main\n", CodeToHTMLOptions{
		Lang: "go", Theme: "github-dark",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(baseline, `style="`) {
		t.Errorf("baseline should have inline styles:\n%s", baseline)
	}
	if strings.Contains(baseline, `_s_`) {
		t.Errorf("baseline should not have _s_ classes:\n%s", baseline)
	}
}
