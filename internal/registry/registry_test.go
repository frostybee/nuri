package registry

import (
	"errors"
	"os"
	"testing"

	"github.com/frostybee/nuri/internal/shared"
)

func TestGetGrammar(t *testing.T) {
	grammarFS := os.DirFS(shared.GrammarsDir(t))
	r, err := New(grammarFS, nil)
	if err != nil {
		t.Fatal(err)
	}
	g, err := r.GetGrammar("go")
	if err != nil {
		t.Fatal(err)
	}
	if g.ScopeName != "source.go" {
		t.Errorf("ScopeName = %q, want source.go", g.ScopeName)
	}
}

func TestGetGrammarAlias(t *testing.T) {
	grammarFS := os.DirFS(shared.GrammarsDir(t))
	r, err := New(grammarFS, nil)
	if err != nil {
		t.Fatal(err)
	}
	r.RegisterAlias("golang", "go")
	g, err := r.GetGrammar("golang")
	if err != nil {
		t.Fatal(err)
	}
	if g.ScopeName != "source.go" {
		t.Errorf("ScopeName = %q, want source.go", g.ScopeName)
	}
}

func TestGetGrammarNotFound(t *testing.T) {
	r, err := New(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = r.GetGrammar("nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrLanguageNotFound) {
		t.Errorf("error = %v, want ErrLanguageNotFound", err)
	}
}

func TestGetTheme(t *testing.T) {
	themeFS := os.DirFS(shared.ThemesDir(t))
	r, err := New(nil, themeFS)
	if err != nil {
		t.Fatal(err)
	}
	th, err := r.GetTheme("github-dark")
	if err != nil {
		t.Fatal(err)
	}
	if th.Name != "github-dark" {
		t.Errorf("Name = %q, want github-dark", th.Name)
	}
}

func TestGetThemeNotFound(t *testing.T) {
	r, err := New(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = r.GetTheme("nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrThemeNotFound) {
		t.Errorf("error = %v, want ErrThemeNotFound", err)
	}
}

func TestRegisterGrammar(t *testing.T) {
	r, err := New(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	data := []byte(`{"scopeName": "source.test", "name": "Test", "patterns": []}`)
	if err := r.RegisterGrammar("test", data); err != nil {
		t.Fatal(err)
	}
	g, err := r.GetGrammar("test")
	if err != nil {
		t.Fatal(err)
	}
	if g.ScopeName != "source.test" {
		t.Errorf("ScopeName = %q, want source.test", g.ScopeName)
	}
}

func TestLoadedLanguages(t *testing.T) {
	r, err := New(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := r.LoadedLanguages(); len(got) != 0 {
		t.Errorf("expected empty, got %v", got)
	}
	r.RegisterGrammar("beta", []byte(`{"scopeName":"s.b","patterns":[]}`))
	r.RegisterGrammar("alpha", []byte(`{"scopeName":"s.a","patterns":[]}`))
	r.GetGrammar("beta")
	r.GetGrammar("alpha")

	got := r.LoadedLanguages()
	if len(got) != 2 || got[0] != "alpha" || got[1] != "beta" {
		t.Errorf("LoadedLanguages = %v, want [alpha beta]", got)
	}
}
