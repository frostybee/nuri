package ast

import (
	"fmt"
	"hash/fnv"
	"slices"
	"strings"

	"github.com/frostybee/nuri/theme"
)

// CodeToTokensOptions configures a CodeToTokens call.
type CodeToTokensOptions struct {
	Lang  string // language name (e.g. "go", "javascript")
	Theme string // theme name (e.g. "github-dark")

	MaxLineLength *int // nil = use highlighter default; per-line byte-length pre-filter
	TimeoutMs     *int // nil = use highlighter default; per-line soft timeout in ms
}

// TokenStyle holds resolved style for a single theme.
type TokenStyle struct {
	Color     string
	BgColor   string
	FontStyle theme.FontStyle
}

// ThemedToken represents a single token with resolved style information.
type ThemedToken struct {
	Content   string
	Color     string          // resolved foreground hex color (default theme)
	BgColor   string          // resolved background hex color (default theme, often empty)
	FontStyle theme.FontStyle // bitmask (default theme): Italic, Bold, Underline, Strikethrough
	Scopes    []string        // TextMate scope stack (outermost first), nil for plaintext

	// ThemeStyles holds per-theme styles in multi-theme mode.
	// Keys are theme keys from CodeToHTMLOptions.Themes (e.g. "dark").
	// nil in single-theme mode.
	ThemeStyles map[string]TokenStyle
}

// TokensResult is the output of CodeToTokens.
type TokensResult struct {
	Tokens      [][]ThemedToken
	FG          string // theme default foreground
	BG          string // theme default background
	ThemeName   string
	Diagnostics []Diagnostic

	// Multi-theme: per-theme default colors. nil in single-theme mode.
	ThemeFG    map[string]string
	ThemeBG    map[string]string
	ThemeNames []string // sorted theme keys (for deterministic output)
}

// Diagnostic records a non-fatal degradation during tokenization.
type Diagnostic struct {
	Line int
	Kind string // "timeout" | "too_long" | "panic" | "unknown_lang"
}

// CodeToHTMLOptions configures a CodeToHTML call.
type CodeToHTMLOptions struct {
	Lang  string
	Theme string // single-theme mode

	// Multi-theme: key → theme name (e.g. {"light":"github-light","dark":"github-dark"}).
	// The lexicographically first key is the default theme (inline styles);
	// others produce CSS variables (--nuri-{key}-{prop}).
	// When non-nil, Theme is ignored.
	Themes map[string]string

	// DefaultColor controls inline color emission in multi-theme mode.
	// nil or *true = emit inline color: for the default theme.
	// *false = emit only CSS variables, no inline styles on tokens.
	DefaultColor *bool

	Transformers   []Transformer
	HighlightLines []LineRange
	FocusLines     []LineRange
	InsertedLines  []LineRange
	DeletedLines   []LineRange
	PreClass       string
	CodeClass      string
	PreAttrs       map[string]string
	CodeAttrs      map[string]string

	// ClassMap replaces inline styles with hashed class names when non-nil.
	// The map accumulates across multiple CodeToHTML calls; call CSS() for the stylesheet.
	ClassMap *StyleClassMap

	MaxLineLength *int // nil = use highlighter default; per-line byte-length pre-filter
	TimeoutMs     *int // nil = use highlighter default; per-line soft timeout in ms
}

// StyleClassMap collects unique style combinations and assigns deterministic
// class names. Pass a shared instance across multiple CodeToHTML calls to
// deduplicate styles across code blocks, then call CSS() for the stylesheet.
type StyleClassMap struct {
	byCanon map[string]string
	rules   map[string]string
}

// NewStyleClassMap creates an empty StyleClassMap.
func NewStyleClassMap() *StyleClassMap {
	return &StyleClassMap{
		byCanon: make(map[string]string),
		rules:   make(map[string]string),
	}
}

// Get returns the class name for the given style map, creating one if needed.
func (m *StyleClassMap) Get(styles map[string]string) string {
	canon := CanonicalStyles(styles)
	if cls, ok := m.byCanon[canon]; ok {
		return cls
	}
	cls := StyleHash(canon)
	m.byCanon[canon] = cls
	m.rules[cls] = StylestoCSS(styles)
	return cls
}

// CSS returns the complete stylesheet with rules sorted by class name.
func (m *StyleClassMap) CSS() string {
	names := make([]string, 0, len(m.rules))
	for name := range m.rules {
		names = append(names, name)
	}
	slices.Sort(names)

	var sb strings.Builder
	for _, name := range names {
		fmt.Fprintf(&sb, ".%s { %s }\n", name, m.rules[name])
	}
	return sb.String()
}

// CanonicalStyles produces a deterministic string key from a style map.
func CanonicalStyles(styles map[string]string) string {
	keys := make([]string, 0, len(styles))
	for k := range styles {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	var sb strings.Builder
	for i, k := range keys {
		if i > 0 {
			sb.WriteByte(';')
		}
		sb.WriteString(k)
		sb.WriteByte(':')
		sb.WriteString(styles[k])
	}
	return sb.String()
}

// StyleHash computes a deterministic short hash from a canonical style string.
func StyleHash(canon string) string {
	h := fnv.New64a()
	h.Write([]byte(canon))
	return fmt.Sprintf("_s_%x", h.Sum64())
}

// StylestoCSS converts a style map to a CSS rule body string.
func StylestoCSS(styles map[string]string) string {
	keys := make([]string, 0, len(styles))
	for k := range styles {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	var sb strings.Builder
	for i, k := range keys {
		if i > 0 {
			sb.WriteString("; ")
		}
		sb.WriteString(k)
		sb.WriteString(": ")
		sb.WriteString(styles[k])
	}
	return sb.String()
}
