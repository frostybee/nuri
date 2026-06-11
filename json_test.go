package nuri

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/frostybee/nuri/ast"
)

func TestCodeToJSONRoundTrip(t *testing.T) {
	h := newTestHighlighter(t)
	data, err := h.CodeToJSON(context.Background(), "package main\n", CodeToJSONOptions{
		Lang: "go", Theme: "github-dark",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(data) {
		t.Fatal("output is not valid JSON")
	}
	var result ast.TokensResult
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(result.Tokens) == 0 {
		t.Fatal("expected at least one line of tokens")
	}
	if result.FG == "" {
		t.Error("expected non-empty FG")
	}
}

func TestCodeToJSONIndent(t *testing.T) {
	h := newTestHighlighter(t)
	data, err := h.CodeToJSON(context.Background(), "x\n", CodeToJSONOptions{
		Lang: "go", Theme: "github-dark", Indent: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	if len(s) < 10 || s[0] != '{' {
		t.Fatal("expected indented JSON object")
	}
	if !json.Valid(data) {
		t.Fatal("indented output is not valid JSON")
	}
}

func TestCodeToJSONCamelCaseKeys(t *testing.T) {
	h := newTestHighlighter(t)
	data, err := h.CodeToJSON(context.Background(), "package main\n", CodeToJSONOptions{
		Lang: "go", Theme: "github-dark",
	})
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"tokens", "fg", "bg", "themeName"} {
		if _, ok := raw[key]; !ok {
			t.Errorf("missing expected camelCase key %q", key)
		}
	}
}

func TestCodeToJSONOmitEmpty(t *testing.T) {
	h := newTestHighlighter(t)
	data, err := h.CodeToJSON(context.Background(), "package main\n", CodeToJSONOptions{
		Lang: "go", Theme: "github-dark",
	})
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"themeStyles", "themeFG", "themeBG", "themeNames"} {
		if _, ok := raw[key]; ok {
			t.Errorf("single-theme output should omit %q", key)
		}
	}
}

func TestCodeToJSONMultiTheme(t *testing.T) {
	h := newTestHighlighter(t)
	data, err := h.CodeToJSON(context.Background(), "package main\n", CodeToJSONOptions{
		Themes: map[string]string{"dark": "github-dark", "light": "github-light"},
	})
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if _, ok := raw["themeFG"]; !ok {
		t.Error("multi-theme output should include themeFG")
	}
}
