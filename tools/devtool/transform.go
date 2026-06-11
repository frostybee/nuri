package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
)

// grammarDropFields are top level grammar JSON keys that the runtime
// parser provably never reads. The grammar parser consumes scopeName,
// name, patterns, repository, injections, injectTo and injectionSelector;
// the registry probe additionally reads fileTypes and firstLineMatch.
var grammarDropFields = []string{"displayName"}

// themeDropFields are top level theme JSON keys the theme parser ignores.
// semanticTokenColors and semanticHighlighting drive LSP based semantic
// highlighting, which nuri does not implement. Theme displayName is kept
// because theme.Theme.DisplayName is public API.
var themeDropFields = []string{"semanticTokenColors", "semanticHighlighting"}

// minifyStripTop removes the given top level keys and minifies. The top
// level round trips through map[string]json.RawMessage, so top level key
// order becomes alphabetical (semantically a no op), while every kept value
// is preserved verbatim minus insignificant whitespace: the encoder
// compacts RawMessage bytes without re decoding them, so string escapes and
// number literals stay byte exact. SetEscapeHTML(false) keeps the
// characters < > & unmangled inside regex patterns. CR and LF are JSON
// whitespace, so CRLF and LF checkouts of the submodule produce identical
// output.
func minifyStripTop(data []byte, drop ...string) ([]byte, error) {
	var top map[string]json.RawMessage
	if err := json.Unmarshal(data, &top); err != nil {
		return nil, err
	}
	for _, key := range drop {
		delete(top, key)
	}
	return marshalCompact(top)
}

// marshalCompact marshals v without HTML escaping and without the trailing
// newline json.Encoder appends.
func marshalCompact(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(buf.Bytes(), []byte("\n")), nil
}

// gzipBytes compresses with gzip.BestCompression and a default header.
// The output is deterministic for a given Go version: a zero ModTime
// writes a zero MTIME field, empty Name and Comment are omitted, and the
// OS byte defaults to 255 (unknown).
func gzipBytes(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return nil, err
	}
	if _, err := w.Write(data); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func gunzipBytes(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	out, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	if err := r.Close(); err != nil {
		return nil, err
	}
	return out, nil
}
