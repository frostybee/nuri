package nuri

import (
	"context"
	"testing"
)

func TestDetectLanguageByExtension(t *testing.T) {
	h := newTestHighlighter(t)
	tests := []struct {
		filename string
		want     string
	}{
		{"main.go", "go"},
		{"app.py", "python"},
		{"index.js", "javascript"},
		{"lib.ts", "typescript"},
		{"style.css", "css"},
		{"data.json", "json"},
		{"page.html", "html"},
		{"page.htm", "html"},
		{"lib.rs", "rust"},
		{"main.c", "c"},
		{"header.h", "c"},
		{"main.cpp", "cpp"},
		{"app.java", "java"},
		{"app.rb", "ruby"},
		{"script.sh", "shellscript"},
		{"config.yaml", "yaml"},
		{"config.yml", "yaml"},
		{"config.toml", "toml"},
		{"app.lua", "lua"},
		{"query.sql", "sql"},
		{"readme.md", "markdown"},
		{"app.swift", "swift"},
		{"run.bat", "bat"},
	}
	for _, tt := range tests {
		lang, ok := h.DetectLanguage(tt.filename)
		if !ok {
			t.Errorf("DetectLanguage(%q) = not found, want %q", tt.filename, tt.want)
			continue
		}
		if lang != tt.want {
			t.Errorf("DetectLanguage(%q) = %q, want %q", tt.filename, lang, tt.want)
		}
	}
}

func TestDetectLanguageByFilename(t *testing.T) {
	h := newTestHighlighter(t)
	tests := []struct {
		filename string
		want     string
	}{
		{"Makefile", "make"},
		{"makefile", "make"},
		{"GNUmakefile", "make"},
		{"Dockerfile", "docker"},
		{"Gemfile", "ruby"},
		{"Rakefile", "ruby"},
	}
	for _, tt := range tests {
		lang, ok := h.DetectLanguage(tt.filename)
		if !ok {
			t.Errorf("DetectLanguage(%q) = not found, want %q", tt.filename, tt.want)
			continue
		}
		if lang != tt.want {
			t.Errorf("DetectLanguage(%q) = %q, want %q", tt.filename, lang, tt.want)
		}
	}
}

func TestDetectLanguageByPath(t *testing.T) {
	h := newTestHighlighter(t)
	tests := []struct {
		path string
		want string
	}{
		{"/home/user/project/main.go", "go"},
		{"src/components/App.tsx", "tsx"},
		{"C:\\Users\\dev\\script.py", "python"},
	}
	for _, tt := range tests {
		lang, ok := h.DetectLanguage(tt.path)
		if !ok {
			t.Errorf("DetectLanguage(%q) = not found, want %q", tt.path, tt.want)
			continue
		}
		if lang != tt.want {
			t.Errorf("DetectLanguage(%q) = %q, want %q", tt.path, lang, tt.want)
		}
	}
}

func TestDetectLanguageUnknown(t *testing.T) {
	h := newTestHighlighter(t)
	_, ok := h.DetectLanguage("data.xyz")
	if ok {
		t.Error("DetectLanguage(data.xyz) should return false for unknown extension")
	}
	_, ok = h.DetectLanguage("noextension")
	if ok {
		t.Error("DetectLanguage(noextension) should return false for no extension")
	}
}

func TestDetectLanguageCaseInsensitive(t *testing.T) {
	h := newTestHighlighter(t)
	tests := []string{"FILE.PY", "Main.GO", "app.Js"}
	for _, f := range tests {
		_, ok := h.DetectLanguage(f)
		if !ok {
			t.Errorf("DetectLanguage(%q) should match case-insensitively", f)
		}
	}
}

func TestRegisterExtensionOverride(t *testing.T) {
	h := newTestHighlighter(t)
	h.RegisterExtension("h", "cpp")
	lang, ok := h.DetectLanguage("header.h")
	if !ok || lang != "cpp" {
		t.Errorf("after RegisterExtension, DetectLanguage(header.h) = %q,%v, want cpp,true", lang, ok)
	}
}

func TestWithExtensionOption(t *testing.T) {
	ctx := context.Background()
	h := newTestHighlighterWithOpts(t, WithExtension("myext", "python"))
	defer h.Close(ctx)

	lang, ok := h.DetectLanguage("file.myext")
	if !ok || lang != "python" {
		t.Errorf("WithExtension: DetectLanguage(file.myext) = %q,%v, want python,true", lang, ok)
	}
}
