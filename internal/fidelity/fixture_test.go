package fidelity

import (
	"path/filepath"
	"testing"
)

func TestLoadFixture(t *testing.T) {
	path := filepath.Join("testdata", "sample_fixture.json")
	f, err := LoadFixture(path)
	if err != nil {
		t.Fatalf("LoadFixture: %v", err)
	}

	if f.VsctmVersion != "9.3.2" {
		t.Errorf("VsctmVersion = %q, want %q", f.VsctmVersion, "9.3.2")
	}
	if f.Grammar != "go" {
		t.Errorf("Grammar = %q, want %q", f.Grammar, "go")
	}
	if f.Source != "package main\n" {
		t.Errorf("Source = %q, want %q", f.Source, "package main\n")
	}
	if len(f.Themes) != 1 {
		t.Fatalf("len(Themes) = %d, want 1", len(f.Themes))
	}

	tf, ok := f.Themes["github-light"]
	if !ok {
		t.Fatal("missing theme github-light")
	}
	if len(tf.Tokens) != 1 {
		t.Fatalf("len(Tokens) = %d, want 1 line", len(tf.Tokens))
	}
	if len(tf.Tokens[0]) != 3 {
		t.Fatalf("len(Tokens[0]) = %d, want 3 tokens", len(tf.Tokens[0]))
	}

	tok := tf.Tokens[0][0]
	if tok.Start != 0 || tok.End != 7 {
		t.Errorf("token 0: [%d:%d], want [0:7]", tok.Start, tok.End)
	}
	if tok.Text != "package" {
		t.Errorf("token 0: text = %q, want %q", tok.Text, "package")
	}
	if tok.Color != "#D73A49" {
		t.Errorf("token 0: color = %q, want %q", tok.Color, "#D73A49")
	}
	if len(tok.Scopes) != 2 || tok.Scopes[0] != "source.go" {
		t.Errorf("token 0: scopes = %v, want [source.go keyword.other.package.go]", tok.Scopes)
	}
	// HTML field is populated by Shiki-based generators but empty for
	// vscode-textmate-based generators (tokens are the primary output).
	// We only validate it's parseable, not that it's non-empty.
}

func TestLoadFixtures(t *testing.T) {
	fixtures, err := LoadFixtures(filepath.Join("testdata"))
	if err != nil {
		t.Fatalf("LoadFixtures: %v", err)
	}
	if len(fixtures) != 1 {
		t.Errorf("len(fixtures) = %d, want 1", len(fixtures))
	}
}

func TestLoadFixtureMissing(t *testing.T) {
	_, err := LoadFixture("nonexistent.json")
	if err == nil {
		t.Error("expected error for missing file")
	}
}
