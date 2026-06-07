package transformers

import (
	"strings"

	"github.com/frostybee/nuri/ast"
)

// WhitespaceOptions configures the whitespace renderer.
type WhitespaceOptions struct {
	Tab   string // visible tab replacement (default "→")
	Space string // visible space replacement (default "·")
}

// Whitespace returns a Transformer that renders whitespace
// characters as visible symbols.
func Whitespace() ast.Transformer {
	return WhitespaceWith(WhitespaceOptions{})
}

// WhitespaceWith returns a whitespace Transformer with custom symbols.
func WhitespaceWith(opts WhitespaceOptions) ast.Transformer {
	if opts.Tab == "" {
		opts.Tab = "→"
	}
	if opts.Space == "" {
		opts.Space = "·"
	}
	return &whitespaceTransformer{tab: opts.Tab, space: opts.Space}
}

type whitespaceTransformer struct {
	ast.BaseTransformer
	tab   string
	space string
}

func (w *whitespaceTransformer) Name() string { return "whitespace" }

func (w *whitespaceTransformer) Span(el *ast.Element, line, col int, lineEl *ast.Element, tok ast.ThemedToken) *ast.Element {
	if !strings.ContainsAny(tok.Content, "\t ") {
		return nil
	}
	wrapper := &ast.Element{
		Tag:    el.Tag,
		Styles: el.Styles,
	}
	for k, v := range el.Attrs {
		wrapper.SetAttr(k, v)
	}
	wrapper.Classes = append(wrapper.Classes, el.Classes...)

	var buf strings.Builder
	for _, ch := range tok.Content {
		switch ch {
		case '\t':
			if buf.Len() > 0 {
				wrapper.Children = append(wrapper.Children, &ast.Text{Content: buf.String()})
				buf.Reset()
			}
			wrapper.Children = append(wrapper.Children, &ast.Element{
				Tag:     "span",
				Classes: []string{"ws-tab"},
				Children: []ast.Node{
					&ast.Text{Content: w.tab},
				},
			})
		case ' ':
			if buf.Len() > 0 {
				wrapper.Children = append(wrapper.Children, &ast.Text{Content: buf.String()})
				buf.Reset()
			}
			wrapper.Children = append(wrapper.Children, &ast.Element{
				Tag:     "span",
				Classes: []string{"ws-space"},
				Children: []ast.Node{
					&ast.Text{Content: w.space},
				},
			})
		default:
			buf.WriteRune(ch)
		}
	}
	if buf.Len() > 0 {
		wrapper.Children = append(wrapper.Children, &ast.Text{Content: buf.String()})
	}
	return wrapper
}
