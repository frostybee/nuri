package registry

import (
	"os"
	"testing"
	"testing/fstest"

	"github.com/frostybee/nuri/internal/shared"
)

func TestGetByName(t *testing.T) {
	repo := newTestRepo(t)

	g, err := repo.Get("go")
	if err != nil {
		t.Fatalf("Get(go): %v", err)
	}
	if g.ScopeName != "source.go" {
		t.Errorf("scopeName: got %q, want %q", g.ScopeName, "source.go")
	}
}

func TestGetByScope(t *testing.T) {
	repo := newTestRepo(t)

	g, err := repo.GetByScope("source.go")
	if err != nil {
		t.Fatalf("GetByScope(source.go): %v", err)
	}
	if g.ScopeName != "source.go" {
		t.Errorf("scopeName: got %q, want %q", g.ScopeName, "source.go")
	}
}

func TestCaching(t *testing.T) {
	repo := newTestRepo(t)

	g1, err := repo.Get("go")
	if err != nil {
		t.Fatalf("first Get: %v", err)
	}
	g2, err := repo.Get("go")
	if err != nil {
		t.Fatalf("second Get: %v", err)
	}
	if g1 != g2 {
		t.Error("expected same pointer from cached result")
	}
}

func TestGetByNameAndScopeReturnSame(t *testing.T) {
	repo := newTestRepo(t)

	g1, err := repo.Get("go")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	g2, err := repo.GetByScope("source.go")
	if err != nil {
		t.Fatalf("GetByScope: %v", err)
	}
	if g1 != g2 {
		t.Error("Get and GetByScope should return the same cached grammar")
	}
}

func TestNotFound(t *testing.T) {
	repo := newTestRepo(t)

	_, err := repo.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent grammar")
	}

	_, err = repo.GetByScope("source.nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent scope")
	}
}

func TestRegister(t *testing.T) {
	repo, err := NewRepository(nil)
	if err != nil {
		t.Fatalf("NewRepository: %v", err)
	}

	data := []byte(`{
		"scopeName": "source.custom",
		"name": "custom",
		"patterns": [{"match": "\\w+", "name": "word"}]
	}`)

	if err := repo.Register("custom", data); err != nil {
		t.Fatalf("Register: %v", err)
	}

	g, err := repo.Get("custom")
	if err != nil {
		t.Fatalf("Get(custom): %v", err)
	}
	if g.ScopeName != "source.custom" {
		t.Errorf("scopeName: got %q", g.ScopeName)
	}

	g2, err := repo.GetByScope("source.custom")
	if err != nil {
		t.Fatalf("GetByScope: %v", err)
	}
	if g2.ScopeName != "source.custom" {
		t.Errorf("scopeName via scope: got %q", g2.ScopeName)
	}
}

func TestMapFS(t *testing.T) {
	fsys := fstest.MapFS{
		"test-lang.json": &fstest.MapFile{
			Data: []byte(`{
				"scopeName": "source.test",
				"name": "test-lang",
				"patterns": [{"match": "hello", "name": "greeting"}]
			}`),
		},
	}

	repo, err := NewRepository(fsys)
	if err != nil {
		t.Fatalf("NewRepository: %v", err)
	}

	g, err := repo.Get("test-lang")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if g.ScopeName != "source.test" {
		t.Errorf("scopeName: got %q", g.ScopeName)
	}

	g2, err := repo.GetByScope("source.test")
	if err != nil {
		t.Fatalf("GetByScope: %v", err)
	}
	if g != g2 {
		t.Error("expected same pointer")
	}
}

func TestNilFS(t *testing.T) {
	repo, err := NewRepository(nil)
	if err != nil {
		t.Fatalf("NewRepository(nil): %v", err)
	}

	_, err = repo.Get("anything")
	if err == nil {
		t.Fatal("expected error with nil fs")
	}
}

func newTestRepo(t *testing.T) *Repository {
	t.Helper()
	repo, err := NewRepository(os.DirFS(shared.GrammarsDir(t)))
	if err != nil {
		t.Fatalf("NewRepository: %v", err)
	}
	return repo
}
