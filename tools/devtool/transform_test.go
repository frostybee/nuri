package main

import (
	"bytes"
	"strings"
	"testing"
)

const sampleGrammar = `{
  "displayName": "Test",
  "name": "test",
  "scopeName": "source.test",
  "patterns": [
    {
      "match": "<\\w+> & friends",
      "name": "tag.test"
    }
  ],
  "repository": {
    "inner": {
      "displayName": "nested key with the same name must survive",
      "match": "1e21"
    }
  }
}`

func TestMinifyStripTopDeterminism(t *testing.T) {
	a, err := minifyStripTop([]byte(sampleGrammar), grammarDropFields...)
	if err != nil {
		t.Fatalf("minify: %v", err)
	}
	b, err := minifyStripTop([]byte(sampleGrammar), grammarDropFields...)
	if err != nil {
		t.Fatalf("minify: %v", err)
	}
	if !bytes.Equal(a, b) {
		t.Error("repeated minification differs")
	}

	crlf := strings.ReplaceAll(sampleGrammar, "\n", "\r\n")
	c, err := minifyStripTop([]byte(crlf), grammarDropFields...)
	if err != nil {
		t.Fatalf("minify CRLF: %v", err)
	}
	if !bytes.Equal(a, c) {
		t.Error("CRLF input produced different output than LF input")
	}
}

func TestMinifyStripsOnlyTopLevel(t *testing.T) {
	out, err := minifyStripTop([]byte(sampleGrammar), grammarDropFields...)
	if err != nil {
		t.Fatalf("minify: %v", err)
	}
	s := string(out)

	if strings.Contains(s, `"displayName":"Test"`) {
		t.Error("top level displayName not stripped")
	}
	if !strings.Contains(s, "nested key with the same name must survive") {
		t.Error("nested displayName was stripped, must only strip top level")
	}
	if !strings.Contains(s, `"scopeName":"source.test"`) {
		t.Error("kept fields must survive minification")
	}
}

func TestMinifyPreservesSpecialChars(t *testing.T) {
	out, err := minifyStripTop([]byte(sampleGrammar), grammarDropFields...)
	if err != nil {
		t.Fatalf("minify: %v", err)
	}
	s := string(out)

	if !strings.Contains(s, `<\\w+> & friends`) {
		t.Errorf("regex special characters were escaped or mangled:\n%s", s)
	}
	if !strings.Contains(s, `"match":"1e21"`) {
		t.Errorf("value bytes not preserved verbatim:\n%s", s)
	}
}

func TestMinifyRemovesWhitespace(t *testing.T) {
	out, err := minifyStripTop([]byte(sampleGrammar))
	if err != nil {
		t.Fatalf("minify: %v", err)
	}
	if bytes.ContainsAny(out, "\n\r\t") {
		t.Error("minified output still contains whitespace characters")
	}
}

func TestGzipRoundTripAndDeterminism(t *testing.T) {
	payload := []byte(strings.Repeat(`{"key":"value with repetition"}`, 100))

	a, err := gzipBytes(payload)
	if err != nil {
		t.Fatalf("gzip: %v", err)
	}
	b, err := gzipBytes(payload)
	if err != nil {
		t.Fatalf("gzip: %v", err)
	}
	if !bytes.Equal(a, b) {
		t.Error("repeated compression differs, header must be deterministic")
	}
	if len(a) >= len(payload) {
		t.Errorf("compressed %d >= raw %d", len(a), len(payload))
	}

	back, err := gunzipBytes(a)
	if err != nil {
		t.Fatalf("gunzip: %v", err)
	}
	if !bytes.Equal(back, payload) {
		t.Error("round trip mismatch")
	}
}
