package nuri

import (
	"io/fs"
	"runtime"
)

type options struct {
	grammarFS fs.FS
	themeFS   fs.FS
	poolSize  int
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

func defaultOptions() options {
	return options{
		poolSize: runtime.NumCPU(),
	}
}
