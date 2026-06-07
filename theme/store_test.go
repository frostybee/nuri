package theme

import (
	"testing"
	"testing/fstest"
)

func TestStoreGet(t *testing.T) {
	fsys := fstest.MapFS{
		"mytheme.json": &fstest.MapFile{Data: []byte(`{
			"name": "mytheme",
			"type": "dark",
			"colors": {"editor.foreground": "#ffffff", "editor.background": "#000000"},
			"tokenColors": [
				{"scope": "keyword", "settings": {"foreground": "#ff0000"}}
			]
		}`)},
	}

	s := NewStore(fsys)
	th, err := s.Get("mytheme")
	if err != nil {
		t.Fatal(err)
	}
	if th.Name != "mytheme" {
		t.Errorf("Name = %q, want %q", th.Name, "mytheme")
	}
}

func TestStoreCaching(t *testing.T) {
	fsys := fstest.MapFS{
		"mytheme.json": &fstest.MapFile{Data: []byte(`{
			"name": "mytheme",
			"tokenColors": []
		}`)},
	}

	s := NewStore(fsys)
	t1, err := s.Get("mytheme")
	if err != nil {
		t.Fatal(err)
	}
	t2, err := s.Get("mytheme")
	if err != nil {
		t.Fatal(err)
	}
	if t1 != t2 {
		t.Error("expected same pointer on second Get (caching)")
	}
}

func TestStoreRegister(t *testing.T) {
	s := NewStore(nil)
	data := []byte(`{
		"name": "registered",
		"type": "light",
		"colors": {"editor.foreground": "#333333"},
		"tokenColors": []
	}`)
	if err := s.Register("custom", data); err != nil {
		t.Fatal(err)
	}
	th, err := s.Get("custom")
	if err != nil {
		t.Fatal(err)
	}
	if th.Name != "registered" {
		t.Errorf("Name = %q, want %q", th.Name, "registered")
	}
}

func TestStoreRegisterInvalidJSON(t *testing.T) {
	s := NewStore(nil)
	err := s.Register("bad", []byte(`not json`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestStoreRegisterOverwritesCache(t *testing.T) {
	s := NewStore(nil)
	v1 := []byte(`{"name": "v1", "tokenColors": []}`)
	v2 := []byte(`{"name": "v2", "tokenColors": []}`)

	if err := s.Register("t", v1); err != nil {
		t.Fatal(err)
	}
	t1, _ := s.Get("t")
	if t1.Name != "v1" {
		t.Fatalf("Name = %q, want v1", t1.Name)
	}

	if err := s.Register("t", v2); err != nil {
		t.Fatal(err)
	}
	t2, _ := s.Get("t")
	if t2.Name != "v2" {
		t.Errorf("Name = %q after re-register, want v2", t2.Name)
	}
}

func TestStoreNotFound(t *testing.T) {
	s := NewStore(nil)
	_, err := s.Get("nonexistent")
	if err == nil {
		t.Error("expected error for missing theme")
	}
}

func TestStoreLoadedThemes(t *testing.T) {
	s := NewStore(nil)

	got := s.LoadedThemes()
	if len(got) != 0 {
		t.Errorf("expected empty, got %v", got)
	}

	_ = s.Register("beta", []byte(`{"name":"b","tokenColors":[]}`))
	_ = s.Register("alpha", []byte(`{"name":"a","tokenColors":[]}`))
	s.Get("beta")
	s.Get("alpha")

	got = s.LoadedThemes()
	if len(got) != 2 || got[0] != "alpha" || got[1] != "beta" {
		t.Errorf("LoadedThemes = %v, want [alpha beta]", got)
	}
}
