package nuri

import (
	"io/fs"
	"runtime"
)

type options struct {
	grammarFS     fs.FS
	themeFS       fs.FS
	poolSize      int
	maxLineLength int
	timeoutMs     int
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

func defaultOptions() options {
	return options{
		poolSize: runtime.NumCPU(),
	}
}
