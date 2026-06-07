package transformers

import (
	"regexp"
	"strings"

	"github.com/frostybee/nuri/ast"
)

var notationRe = regexp.MustCompile(`\[!code\s+([\w:+\-]+)\]`)

// Notation returns a Transformer that parses magic comments
// like [!code ++], [!code highlight], [!code focus], etc. It strips the
// comment and applies corresponding classes to line elements.
func Notation() ast.Transformer {
	return &notationTransformer{
		lineClasses: make(map[int][]string),
	}
}

type notationTransformer struct {
	ast.BaseTransformer
	lineClasses    map[int][]string
	hasFocus       bool
	wordHighlights []wordHighlight
}

type wordHighlight struct {
	word string
}

func (n *notationTransformer) Name() string { return "notation" }

func (n *notationTransformer) Tokens(tokens [][]ast.ThemedToken) [][]ast.ThemedToken {
	result := make([][]ast.ThemedToken, 0, len(tokens))
	for lineIdx, line := range tokens {
		lineNum := lineIdx + 1
		annotIdx, annot := n.findAnnotation(line)
		if annot == "" {
			result = append(result, line)
			continue
		}

		classes := n.annotationClasses(annot)
		if len(classes) > 0 {
			n.lineClasses[lineNum] = classes
		}

		stripped := n.stripAnnotation(line, annotIdx)
		if n.isEmptyLine(stripped) {
			shifted := make(map[int][]string, len(n.lineClasses))
			for k, v := range n.lineClasses {
				if k == lineNum {
					continue
				}
				if k > lineNum {
					shifted[k-1] = v
				} else {
					shifted[k] = v
				}
			}
			n.lineClasses = shifted
			continue
		}
		result = append(result, stripped)
	}
	return result
}

func (n *notationTransformer) findAnnotation(line []ast.ThemedToken) (int, string) {
	for i := len(line) - 1; i >= 0; i-- {
		if loc := notationRe.FindStringIndex(line[i].Content); loc != nil {
			match := notationRe.FindStringSubmatch(line[i].Content)
			return i, match[1]
		}
	}
	return -1, ""
}

func (n *notationTransformer) annotationClasses(annot string) []string {
	switch annot {
	case "++":
		return []string{"diff", "add"}
	case "--":
		return []string{"diff", "remove"}
	case "highlight":
		return []string{"highlighted"}
	case "focus":
		n.hasFocus = true
		return []string{"focused"}
	case "error":
		return []string{"highlighted", "error"}
	case "warning":
		return []string{"highlighted", "warning"}
	default:
		if strings.HasPrefix(annot, "word:") {
			word := annot[5:]
			if word != "" {
				n.wordHighlights = append(n.wordHighlights, wordHighlight{word: word})
			}
		}
		return nil
	}
}

func (n *notationTransformer) stripAnnotation(line []ast.ThemedToken, annotIdx int) []ast.ThemedToken {
	tok := line[annotIdx]
	loc := notationRe.FindStringIndex(tok.Content)
	before := tok.Content[:loc[0]]
	after := tok.Content[loc[1]:]

	before = strings.TrimRight(before, " ")
	remaining := strings.TrimSpace(before + after)

	if isCommentPrefix(remaining) || remaining == "" {
		result := make([]ast.ThemedToken, 0, len(line))
		for i, t := range line {
			if i == annotIdx {
				continue
			}
			if i == annotIdx-1 && strings.TrimSpace(t.Content) == "" {
				continue
			}
			result = append(result, t)
		}
		if len(result) > 0 {
			last := &result[len(result)-1]
			last.Content = strings.TrimRight(last.Content, " ")
		}
		return result
	}

	cleaned := before + after
	result := make([]ast.ThemedToken, len(line))
	copy(result, line)
	result[annotIdx].Content = cleaned
	return result
}

func isCommentPrefix(s string) bool {
	s = strings.TrimSpace(s)
	return s == "//" || s == "#" || s == "/*" || s == "*/" || s == "/* */"
}

func (n *notationTransformer) isEmptyLine(line []ast.ThemedToken) bool {
	for _, tok := range line {
		if strings.TrimSpace(tok.Content) != "" {
			return false
		}
	}
	return true
}

func (n *notationTransformer) Line(el *ast.Element, line int) *ast.Element {
	if classes, ok := n.lineClasses[line]; ok {
		for _, c := range classes {
			el.AddClass(c)
		}
	}
	if n.hasFocus {
		if _, ok := n.lineClasses[line]; !ok || !containsClass(n.lineClasses[line], "focused") {
			el.AddClass("dimmed")
		}
	}
	return el
}

func (n *notationTransformer) Pre(el *ast.Element) *ast.Element {
	if n.hasFocus {
		el.AddClass("has-focused")
	}
	return el
}

func (n *notationTransformer) Span(el *ast.Element, line, col int, lineEl *ast.Element, tok ast.ThemedToken) *ast.Element {
	if len(n.wordHighlights) == 0 {
		return nil
	}
	for _, wh := range n.wordHighlights {
		if strings.Contains(tok.Content, wh.word) {
			return n.wrapWord(el, tok, wh.word)
		}
	}
	return nil
}

func (n *notationTransformer) wrapWord(el *ast.Element, tok ast.ThemedToken, word string) *ast.Element {
	parts := strings.SplitAfter(tok.Content, word)
	wrapper := &ast.Element{
		Tag:    el.Tag,
		Styles: el.Styles,
	}

	for i, part := range parts {
		if part == "" {
			continue
		}
		if i < len(parts)-1 && strings.HasSuffix(part, word) {
			before := part[:len(part)-len(word)]
			if before != "" {
				wrapper.Children = append(wrapper.Children, &ast.Text{Content: before})
			}
			highlight := &ast.Element{
				Tag:     "span",
				Classes: []string{"highlighted-word"},
			}
			highlight.Children = append(highlight.Children, &ast.Text{Content: word})
			wrapper.Children = append(wrapper.Children, highlight)
		} else {
			wrapper.Children = append(wrapper.Children, &ast.Text{Content: part})
		}
	}

	return wrapper
}

func containsClass(classes []string, class string) bool {
	for _, c := range classes {
		if c == class {
			return true
		}
	}
	return false
}
