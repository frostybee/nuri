package nuri

import (
	"context"
	"encoding/xml"
	"strings"
	"testing"
)

func TestCodeToSVGBasic(t *testing.T) {
	h := newTestHighlighter(t)
	svg, err := h.CodeToSVG(context.Background(), "package main\n", CodeToSVGOptions{
		Lang: "go", Theme: "github-dark",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(svg, "<svg") {
		t.Error("output should start with <svg")
	}
	if !strings.Contains(svg, `xmlns="http://www.w3.org/2000/svg"`) {
		t.Error("missing xmlns attribute")
	}
	if !strings.Contains(svg, "</svg>") {
		t.Error("missing closing </svg>")
	}
}

func TestCodeToSVGBackground(t *testing.T) {
	h := newTestHighlighter(t)
	svg, err := h.CodeToSVG(context.Background(), "x\n", CodeToSVGOptions{
		Lang: "go", Theme: "github-dark",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(svg, "<rect") {
		t.Error("expected background rect")
	}
	if !strings.Contains(svg, `rx="8"`) {
		t.Error("expected corner radius on background rect")
	}
}

func TestCodeToSVGNoBackground(t *testing.T) {
	h := newTestHighlighter(t)
	f := false
	svg, err := h.CodeToSVG(context.Background(), "x\n", CodeToSVGOptions{
		Lang: "go", Theme: "github-dark", ShowBackground: &f,
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(svg, "<rect") {
		t.Error("expected no background rect when ShowBackground is false")
	}
}

func TestCodeToSVGEscaping(t *testing.T) {
	h := newTestHighlighter(t)
	svg, err := h.CodeToSVG(context.Background(), `x := "<b>" + "&"`, CodeToSVGOptions{
		Lang: "go", Theme: "github-dark",
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(svg, "<b>") {
		t.Error("unescaped <b> in SVG output")
	}
	if !strings.Contains(svg, "&lt;b&gt;") {
		t.Error("expected escaped <b>")
	}
	if !strings.Contains(svg, "&amp;") {
		t.Error("expected escaped &")
	}
}

func TestCodeToSVGMultiLine(t *testing.T) {
	h := newTestHighlighter(t)
	svg, err := h.CodeToSVG(context.Background(), "a\nb\nc\n", CodeToSVGOptions{
		Lang: "go", Theme: "github-dark",
	})
	if err != nil {
		t.Fatal(err)
	}
	count := strings.Count(svg, "<text ")
	if count < 3 {
		t.Errorf("expected at least 3 <text> elements for 3+ lines, got %d", count)
	}
}

func TestCodeToSVGCustomOptions(t *testing.T) {
	h := newTestHighlighter(t)
	svg, err := h.CodeToSVG(context.Background(), "x\n", CodeToSVGOptions{
		Lang: "go", Theme: "github-dark",
		FontFamily: "Fira Code", FontSize: 16,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(svg, "Fira Code") {
		t.Error("expected custom font family in output")
	}
	if !strings.Contains(svg, `font-size="16px"`) {
		t.Error("expected custom font size in output")
	}
}

func TestCodeToSVGValidXML(t *testing.T) {
	h := newTestHighlighter(t)
	svg, err := h.CodeToSVG(context.Background(), "package main\n\nfunc main() {\n}\n", CodeToSVGOptions{
		Lang: "go", Theme: "github-dark",
	})
	if err != nil {
		t.Fatal(err)
	}
	decoder := xml.NewDecoder(strings.NewReader(svg))
	for {
		_, err := decoder.Token()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			t.Fatalf("SVG is not well-formed XML: %v", err)
		}
	}
}

func TestCodeToSVGEmpty(t *testing.T) {
	h := newTestHighlighter(t)
	svg, err := h.CodeToSVG(context.Background(), "", CodeToSVGOptions{
		Lang: "go", Theme: "github-dark",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(svg, "<svg") {
		t.Error("empty input should still produce valid SVG")
	}
}
