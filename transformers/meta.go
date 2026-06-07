package transformers

import (
	"strconv"
	"strings"

	"github.com/frostybee/nuri/ast"
)

// ParseMetaRanges parses a code-fence meta string like "{1,3-5,7}"
// into a slice of LineRanges.
func ParseMetaRanges(meta string) ([]ast.LineRange, error) {
	meta = strings.TrimSpace(meta)
	if meta == "" {
		return nil, nil
	}
	start := strings.IndexByte(meta, '{')
	end := strings.LastIndexByte(meta, '}')
	if start < 0 || end <= start {
		return nil, nil
	}
	inner := strings.TrimSpace(meta[start+1 : end])
	if inner == "" {
		return nil, nil
	}

	parts := strings.Split(inner, ",")
	ranges := make([]ast.LineRange, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if dash := strings.IndexByte(p, '-'); dash >= 0 {
			s, err := strconv.Atoi(strings.TrimSpace(p[:dash]))
			if err != nil {
				return nil, err
			}
			e, err := strconv.Atoi(strings.TrimSpace(p[dash+1:]))
			if err != nil {
				return nil, err
			}
			ranges = append(ranges, ast.LineRange{Start: s, End: e})
		} else {
			n, err := strconv.Atoi(p)
			if err != nil {
				return nil, err
			}
			ranges = append(ranges, ast.LineRange{Start: n, End: n})
		}
	}
	return ranges, nil
}

// Meta returns a Transformer that applies line highlighting
// based on a code-fence meta string (e.g. "{1,3-5}").
func Meta(meta string) ast.Transformer {
	ranges, _ := ParseMetaRanges(meta)
	return &metaTransformer{ranges: ranges}
}

type metaTransformer struct {
	ast.BaseTransformer
	ranges []ast.LineRange
}

func (m *metaTransformer) Name() string { return "meta" }

func (m *metaTransformer) Preprocess(code string, opts *ast.CodeToHTMLOptions) string {
	if len(m.ranges) > 0 {
		opts.HighlightLines = append(opts.HighlightLines, m.ranges...)
	}
	return ""
}
