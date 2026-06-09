package nuri

import (
	"io/fs"
	"runtime"

	"github.com/frostybee/nuri/ast"
)

type grammarEntry struct {
	name string
	data []byte
}

type themeEntry struct {
	name string
	data []byte
}

type aliasEntry struct {
	alias  string
	target string
}

type extensionEntry struct {
	ext  string
	lang string
}

type options struct {
	grammarFS     fs.FS
	themeFS       fs.FS
	poolSize      int
	maxLineLength int
	timeoutMs     int
	grammars      []grammarEntry
	themes        []themeEntry
	aliases       []aliasEntry
	extensions    []extensionEntry
	defaults      *ast.CodeToHTMLOptions
}

// Option configures a Highlighter.
type Option func(*options)

// WithGrammarFS sets the filesystem for grammar JSON files.
func WithGrammarFS(fsys fs.FS) Option {
	return func(o *options) { o.grammarFS = fsys }
}

// WithThemeFS sets the filesystem for theme JSON files.
func WithThemeFS(fsys fs.FS) Option {
	return func(o *options) { o.themeFS = fsys }
}

// WithFS sets both the grammar and theme filesystems from a single fs.FS
// that contains "grammars/" and "themes/" subdirectories. This is the
// intended way to use the bundle packages.
func WithFS(fsys fs.FS) Option {
	return func(o *options) {
		o.grammarFS, _ = fs.Sub(fsys, "grammars")
		o.themeFS, _ = fs.Sub(fsys, "themes")
	}
}

// WithPoolSize sets the number of WASM instances in the pool.
// Defaults to runtime.NumCPU().
func WithPoolSize(n int) Option {
	return func(o *options) { o.poolSize = n }
}

// WithMaxLineLength sets the byte-length threshold for per-line pre-filtering.
// Lines exceeding this are emitted as a single unstyled token with a "too_long"
// diagnostic. 0 means no limit (the default).
func WithMaxLineLength(n int) Option {
	return func(o *options) { o.maxLineLength = n }
}

// WithTimeoutMs sets the per-line soft timeout in milliseconds. Lines whose
// tokenization exceeds this are stopped early; partial tokens are preserved and
// a "timeout" diagnostic is recorded. 0 means no timeout (the default).
func WithTimeoutMs(ms int) Option {
	return func(o *options) { o.timeoutMs = ms }
}

// WithGrammar registers a custom grammar from JSON bytes at construction time.
func WithGrammar(name string, data []byte) Option {
	return func(o *options) {
		o.grammars = append(o.grammars, grammarEntry{name, data})
	}
}

// WithTheme registers a custom theme from JSON bytes at construction time.
func WithTheme(name string, data []byte) Option {
	return func(o *options) {
		o.themes = append(o.themes, themeEntry{name, data})
	}
}

// WithAlias registers a language alias at construction time (e.g. "sh" -> "shellscript").
func WithAlias(alias, target string) Option {
	return func(o *options) {
		o.aliases = append(o.aliases, aliasEntry{alias, target})
	}
}

// WithExtension maps a file extension (without dot) to a language name at
// construction time. Overrides any existing mapping for that extension.
func WithExtension(ext, lang string) Option {
	return func(o *options) {
		o.extensions = append(o.extensions, extensionEntry{ext, lang})
	}
}

// WithDefaults sets default CodeToHTMLOptions applied to every CodeToHTML call.
// Per-call options override these defaults (non-zero values win).
func WithDefaults(defaults CodeToHTMLOptions) Option {
	return func(o *options) {
		o.defaults = &defaults
	}
}

func defaultOptions() options {
	return options{
		poolSize: runtime.NumCPU(),
	}
}
