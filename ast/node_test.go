package ast

import (
	"strings"
	"testing"
)

func TestEscapeText(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"hello", "hello"},
		{"<b>", "&lt;b&gt;"},
		{"a & b", "a &amp; b"},
		{"<>&", "&lt;&gt;&amp;"},
		{"", ""},
		{"no specials", "no specials"},
	}
	for _, tt := range tests {
		got := escapeText(tt.in)
		if got != tt.want {
			t.Errorf("escapeText(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestEscapeTextScript(t *testing.T) {
	got := escapeText("</script>")
	want := "&lt;/script&gt;"
	if got != want {
		t.Errorf("escapeText(</script>) = %q, want %q", got, want)
	}
}

func TestEscapeAttr(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{`hello`, `hello`},
		{`say "hi"`, `say &quot;hi&quot;`},
		{`a & b`, `a &amp; b`},
		{`"&"`, `&quot;&amp;&quot;`},
		{"", ""},
	}
	for _, tt := range tests {
		got := escapeAttr(tt.in)
		if got != tt.want {
			t.Errorf("escapeAttr(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestTextWriteTo(t *testing.T) {
	txt := &Text{Content: "<b>hello</b>"}
	var buf strings.Builder
	n, err := txt.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	want := "&lt;b&gt;hello&lt;/b&gt;"
	if got != want {
		t.Errorf("Text.WriteTo = %q, want %q", got, want)
	}
	if n != int64(len(want)) {
		t.Errorf("byte count = %d, want %d", n, len(want))
	}
}

func TestElementWriteTo(t *testing.T) {
	el := &Element{
		Tag:     "span",
		Classes: []string{"line"},
		Styles:  map[string]string{"color": "#ff0000", "background-color": "#000"},
		Children: []Node{
			&Text{Content: "hello"},
		},
	}
	var buf strings.Builder
	_, err := el.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	want := `<span class="line" style="background-color:#000;color:#ff0000">hello</span>`
	if got != want {
		t.Errorf("Element.WriteTo =\n  %q\nwant\n  %q", got, want)
	}
}

func TestElementAttrs(t *testing.T) {
	el := &Element{
		Tag:   "pre",
		Attrs: map[string]string{"tabindex": "0", "data-lang": "go"},
	}
	var buf strings.Builder
	el.WriteTo(&buf)
	got := buf.String()
	want := `<pre data-lang="go" tabindex="0"></pre>`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestElementNested(t *testing.T) {
	el := &Element{
		Tag: "div",
		Children: []Node{
			&Element{Tag: "span", Children: []Node{&Text{Content: "a"}}},
			&Text{Content: "\n"},
			&Element{Tag: "span", Children: []Node{&Text{Content: "b"}}},
		},
	}
	var buf strings.Builder
	el.WriteTo(&buf)
	got := buf.String()
	want := "<div><span>a</span>\n<span>b</span></div>"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestElementEmpty(t *testing.T) {
	el := &Element{Tag: "span", Classes: []string{"line"}}
	var buf strings.Builder
	el.WriteTo(&buf)
	got := buf.String()
	want := `<span class="line"></span>`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestElementDeterministic(t *testing.T) {
	el := &Element{
		Tag:     "span",
		Classes: []string{"line"},
		Styles:  map[string]string{"color": "#aaa", "font-style": "italic", "background-color": "#000"},
		Attrs:   map[string]string{"data-z": "1", "data-a": "2"},
		Children: []Node{
			&Text{Content: "test"},
		},
	}
	var first string
	for i := 0; i < 100; i++ {
		var buf strings.Builder
		el.WriteTo(&buf)
		if i == 0 {
			first = buf.String()
		} else if buf.String() != first {
			t.Fatalf("non-deterministic output at iteration %d:\n  %q\nvs\n  %q", i, buf.String(), first)
		}
	}
}

func TestAddClass(t *testing.T) {
	el := &Element{Tag: "span", Classes: []string{"line"}}
	el.AddClass("highlighted")
	el.AddClass("highlighted")
	if len(el.Classes) != 2 {
		t.Errorf("expected 2 classes, got %d: %v", len(el.Classes), el.Classes)
	}
	if el.Classes[1] != "highlighted" {
		t.Errorf("expected 'highlighted', got %q", el.Classes[1])
	}
}

func TestSetAttr(t *testing.T) {
	el := &Element{Tag: "span"}
	el.SetAttr("data-line", "1")
	if el.Attrs["data-line"] != "1" {
		t.Errorf("attr = %q, want \"1\"", el.Attrs["data-line"])
	}
	el.SetAttr("data-line", "2")
	if el.Attrs["data-line"] != "2" {
		t.Errorf("attr = %q, want \"2\"", el.Attrs["data-line"])
	}
}

func TestStyleValueEscaped(t *testing.T) {
	el := &Element{
		Tag:    "span",
		Styles: map[string]string{"color": `#fff" onmouseover="x` + "&"},
	}
	var buf strings.Builder
	el.WriteTo(&buf)
	out := buf.String()
	if strings.Contains(out, `#fff"`) {
		t.Errorf("style value with quote broke out of the attribute: %s", out)
	}
	if !strings.Contains(out, `style="color:#fff&quot; onmouseover=&quot;x&amp;"`) {
		t.Errorf("style value not escaped as expected: %s", out)
	}
}

func TestClassValueEscaped(t *testing.T) {
	el := &Element{
		Tag:     "pre",
		Classes: []string{"nuri", `theme"with&quote`},
	}
	var buf strings.Builder
	el.WriteTo(&buf)
	out := buf.String()
	if strings.Contains(out, `theme"`) {
		t.Errorf("class value with quote broke out of the attribute: %s", out)
	}
	if !strings.Contains(out, `class="nuri theme&quot;with&amp;quote"`) {
		t.Errorf("class value not escaped as expected: %s", out)
	}
}
