package nuri

import (
	"bytes"
	"context"
	"errors"
	"slices"
	"sort"
	"strings"
	"sync"

	"github.com/frostybee/nuri/ast"
	"github.com/frostybee/nuri/internal/registry"
	"github.com/frostybee/nuri/internal/oniguruma"
	"github.com/frostybee/nuri/internal/tokenizer"
	"github.com/frostybee/nuri/renderer"
	"github.com/frostybee/nuri/theme"
)

// Re-export ast types so consumers can write nuri.Element, etc.
type (
	Element            = ast.Element
	Text               = ast.Text
	Node               = ast.Node
	Transformer        = ast.Transformer
	BaseTransformer    = ast.BaseTransformer
	ThemedToken        = ast.ThemedToken
	TokenStyle         = ast.TokenStyle
	TokensResult       = ast.TokensResult
	Diagnostic         = ast.Diagnostic
	CodeToTokensOptions = ast.CodeToTokensOptions
	CodeToHTMLOptions  = ast.CodeToHTMLOptions
	LineRange          = ast.LineRange
	StyleClassMap      = ast.StyleClassMap
)

// Re-export ast functions.
var (
	Range           = ast.Range
	Lines           = ast.Lines
	NewStyleClassMap = ast.NewStyleClassMap
)

// Highlighter is the main entry point for syntax highlighting.
// It owns WASM resources and must be closed when no longer needed.
type Highlighter struct {
	eng       *oniguruma.Engine
	pool      *oniguruma.Pool
	reg       *registry.Registry
	closeOnce sync.Once
}

// New compiles the WASM engine, instantiates a pool of WASM instances,
// and builds the grammar/theme registry. It does real work and may fail.
func New(ctx context.Context, opts ...Option) (*Highlighter, error) {
	cfg := defaultOptions()
	for _, o := range opts {
		o(&cfg)
	}
	if cfg.poolSize < 1 {
		cfg.poolSize = 1
	}

	eng, err := oniguruma.NewEngine(ctx)
	if err != nil {
		return nil, err
	}

	pool, err := oniguruma.NewPool(ctx, eng, cfg.poolSize)
	if err != nil {
		eng.Close(ctx)
		return nil, err
	}

	reg, err := registry.New(cfg.grammarFS, cfg.themeFS)
	if err != nil {
		pool.Close(ctx)
		eng.Close(ctx)
		return nil, err
	}

	for alias, target := range defaultAliases {
		reg.RegisterAlias(alias, target)
	}

	return &Highlighter{
		eng:  eng,
		pool: pool,
		reg:  reg,
	}, nil
}

// Close releases all WASM resources. Safe to call multiple times.
func (h *Highlighter) Close(ctx context.Context) error {
	var closeErr error
	h.closeOnce.Do(func() {
		poolErr := h.pool.Close(ctx)
		engErr := h.eng.Close(ctx)
		closeErr = errors.Join(poolErr, engErr)
	})
	return closeErr
}

// CodeToTokens tokenizes source code and resolves each token's color
// from the specified theme.
func (h *Highlighter) CodeToTokens(
	ctx context.Context,
	code string,
	opts ast.CodeToTokensOptions,
) (*ast.TokensResult, error) {
	thm, err := h.reg.GetTheme(opts.Theme)
	if err != nil {
		return nil, err
	}

	g, langErr := h.reg.GetGrammar(opts.Lang)
	if langErr != nil && !errors.Is(langErr, ErrLanguageNotFound) {
		return nil, langErr
	}

	if g == nil {
		return h.plaintextFallback(code, thm), nil
	}

	var tokResult *tokenizer.TokenizeResult
	doErr := h.pool.Do(ctx, func(lib oniguruma.OnigLib) error {
		var tokenizeErr error
		tokResult, tokenizeErr = tokenizer.Tokenize(ctx, []byte(code), g, lib, h.reg)
		return tokenizeErr
	})
	if doErr != nil {
		return nil, doErr
	}

	return h.buildResult(code, tokResult, thm), nil
}

// CodeToHTML tokenizes source code, resolves colors from the theme,
// and renders the result as an HTML string.
func (h *Highlighter) CodeToHTML(
	ctx context.Context,
	code string,
	opts ast.CodeToHTMLOptions,
) (string, error) {
	for _, tr := range opts.Transformers {
		if s := tr.Preprocess(code, &opts); s != "" {
			code = s
		}
	}

	var result *ast.TokensResult
	var err error
	if opts.Themes != nil {
		result, err = h.codeToTokensMulti(ctx, code, opts.Lang, opts.Themes)
	} else {
		result, err = h.CodeToTokens(ctx, code, ast.CodeToTokensOptions{
			Lang:  opts.Lang,
			Theme: opts.Theme,
		})
	}
	if err != nil {
		return "", err
	}

	tokens := result.Tokens
	for _, tr := range opts.Transformers {
		if t := tr.Tokens(tokens); t != nil {
			tokens = t
		}
	}
	result.Tokens = tokens

	tree := renderer.BuildTree(result, &opts)

	var buf strings.Builder
	tree.WriteTo(&buf)
	html := buf.String()

	for _, tr := range opts.Transformers {
		if s := tr.Postprocess(html, &opts); s != "" {
			html = s
		}
	}

	return html, nil
}

func (h *Highlighter) buildResult(
	code string,
	tokResult *tokenizer.TokenizeResult,
	thm *theme.Theme,
) *ast.TokensResult {
	lines := splitLines([]byte(code))
	result := &ast.TokensResult{
		Tokens:    make([][]ast.ThemedToken, len(tokResult.Lines)),
		FG:        thm.DefaultForeground,
		BG:        thm.DefaultBackground,
		ThemeName: thm.Name,
	}

	for i, tokLine := range tokResult.Lines {
		themed := make([]ast.ThemedToken, len(tokLine))
		var line []byte
		if i < len(lines) {
			line = lines[i]
		}
		for j, tok := range tokLine {
			var content string
			if line != nil && tok.Start >= 0 && tok.End <= len(line) {
				content = string(line[tok.Start:tok.End])
			}
			ts := thm.Match(tok.Scopes)
			color := ts.Foreground
			if color == "" {
				color = thm.DefaultForeground
			}
			fontStyle := ts.FontStyle
			if fontStyle == theme.FontStyleNotSet {
				fontStyle = theme.FontStyleNone
			}
			themed[j] = ast.ThemedToken{
				Content:   content,
				Color:     color,
				BgColor:   ts.Background,
				FontStyle: fontStyle,
				Scopes:    tok.Scopes,
			}
		}
		result.Tokens[i] = themed
	}

	for _, d := range tokResult.Diagnostics {
		result.Diagnostics = append(result.Diagnostics, ast.Diagnostic{
			Line: d.Line,
			Kind: d.Kind,
		})
	}

	return result
}

func (h *Highlighter) codeToTokensMulti(
	ctx context.Context,
	code string,
	lang string,
	themes map[string]string,
) (*ast.TokensResult, error) {
	keys := make([]string, 0, len(themes))
	for k := range themes {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	thms := make(map[string]*theme.Theme, len(keys))
	for _, k := range keys {
		thm, err := h.reg.GetTheme(themes[k])
		if err != nil {
			return nil, err
		}
		thms[k] = thm
	}

	defaultKey := keys[0]
	defaultThm := thms[defaultKey]

	g, langErr := h.reg.GetGrammar(lang)
	if langErr != nil && !errors.Is(langErr, ErrLanguageNotFound) {
		return nil, langErr
	}

	if g == nil {
		result := h.plaintextFallback(code, defaultThm)
		h.addMultiThemeInfo(result, keys, thms, defaultKey)
		return result, nil
	}

	var tokResult *tokenizer.TokenizeResult
	doErr := h.pool.Do(ctx, func(lib oniguruma.OnigLib) error {
		var tokenizeErr error
		tokResult, tokenizeErr = tokenizer.Tokenize(ctx, []byte(code), g, lib, h.reg)
		return tokenizeErr
	})
	if doErr != nil {
		return nil, doErr
	}

	result := h.buildResult(code, tokResult, defaultThm)

	nonDefaultKeys := slices.Delete(slices.Clone(keys), 0, 1)
	for i, tokLine := range tokResult.Lines {
		lines := splitLines([]byte(code))
		var line []byte
		if i < len(lines) {
			line = lines[i]
		}
		for j, tok := range tokLine {
			styles := make(map[string]ast.TokenStyle, len(nonDefaultKeys))
			for _, k := range nonDefaultKeys {
				ts := thms[k].Match(tok.Scopes)
				color := ts.Foreground
				if color == "" {
					color = thms[k].DefaultForeground
				}
				fontStyle := ts.FontStyle
				if fontStyle == theme.FontStyleNotSet {
					fontStyle = theme.FontStyleNone
				}
				styles[k] = ast.TokenStyle{
					Color:     color,
					BgColor:   ts.Background,
					FontStyle: fontStyle,
				}
			}
			_ = line
			result.Tokens[i][j].ThemeStyles = styles
		}
	}

	h.addMultiThemeInfo(result, keys, thms, defaultKey)
	return result, nil
}

func (h *Highlighter) addMultiThemeInfo(
	result *ast.TokensResult,
	keys []string,
	thms map[string]*theme.Theme,
	defaultKey string,
) {
	result.ThemeNames = keys
	result.ThemeFG = make(map[string]string, len(keys))
	result.ThemeBG = make(map[string]string, len(keys))
	for _, k := range keys {
		if k == defaultKey {
			continue
		}
		result.ThemeFG[k] = thms[k].DefaultForeground
		result.ThemeBG[k] = thms[k].DefaultBackground
	}
}

func (h *Highlighter) plaintextFallback(code string, thm *theme.Theme) *ast.TokensResult {
	raw := splitLines([]byte(code))
	themed := make([][]ast.ThemedToken, len(raw))
	for i, line := range raw {
		text := string(line)
		if len(text) > 0 && text[len(text)-1] == '\n' {
			text = text[:len(text)-1]
		}
		if text == "" && i == len(raw)-1 {
			continue
		}
		themed[i] = []ast.ThemedToken{{
			Content:   text,
			Color:     thm.DefaultForeground,
			FontStyle: theme.FontStyleNone,
		}}
	}
	return &ast.TokensResult{
		Tokens:    themed,
		FG:        thm.DefaultForeground,
		BG:        thm.DefaultBackground,
		ThemeName: thm.Name,
		Diagnostics: []ast.Diagnostic{{
			Line: 0,
			Kind: "unknown_lang",
		}},
	}
}

// LoadLanguage registers a grammar from raw JSON bytes.
func (h *Highlighter) LoadLanguage(name string, data []byte) error {
	return h.reg.RegisterGrammar(name, data)
}

// LoadTheme registers a theme from raw JSON bytes.
func (h *Highlighter) LoadTheme(name string, data []byte) error {
	return h.reg.RegisterTheme(name, data)
}

// RegisterAlias maps a language alias to a canonical name.
func (h *Highlighter) RegisterAlias(alias, target string) {
	h.reg.RegisterAlias(alias, target)
}

// LoadedLanguages returns the names of all currently cached grammars.
func (h *Highlighter) LoadedLanguages() []string {
	return h.reg.LoadedLanguages()
}

// LoadedThemes returns the names of all currently cached themes.
func (h *Highlighter) LoadedThemes() []string {
	return h.reg.LoadedThemes()
}

func splitLines(code []byte) [][]byte {
	lines := bytes.Split(code, []byte("\n"))
	for i := 0; i < len(lines)-1; i++ {
		lines[i] = append(lines[i], '\n')
	}
	if len(lines) > 0 && len(lines[len(lines)-1]) == 0 {
		lines = lines[:len(lines)-1]
	}
	return lines
}
