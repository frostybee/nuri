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

func TestIndexFastPath(t *testing.T) {
	// Grammar bodies are deliberately invalid JSON: if construction tries
	// to read them the metadata probe fails and no lookups would exist, so
	// passing lookups prove the index alone built the tables.
	fsys := fstest.MapFS{
		"index.json": &fstest.MapFile{
			Data: []byte(`{
				"version": 1,
				"grammars": {
					"alpha": {"scopeName": "source.alpha", "fileTypes": ["al"], "firstLineMatch": "^#!alpha"},
					"beta":  {"scopeName": "source.beta", "injectTo": ["source.alpha"]}
				}
			}`),
		},
		"alpha.json": &fstest.MapFile{Data: []byte(`not json`)},
		"beta.json":  &fstest.MapFile{Data: []byte(`not json`)},
	}

	repo, err := NewRepository(fsys)
	if err != nil {
		t.Fatalf("NewRepository: %v", err)
	}

	if name, ok := repo.scopeIndex["source.alpha"]; !ok || name != "alpha" {
		t.Errorf("scopeIndex[source.alpha] = %q, %v", name, ok)
	}
	if name, ok := repo.DetectByFilename("main.al"); !ok || name != "alpha" {
		t.Errorf("DetectByFilename(main.al) = %q, %v", name, ok)
	}
	if name, ok := repo.DetectByFirstLine("#!alpha"); !ok || name != "alpha" {
		t.Errorf("DetectByFirstLine = %q, %v", name, ok)
	}
	if injectors := repo.injectionIndex["source.alpha"]; len(injectors) != 1 || injectors[0] != "beta" {
		t.Errorf("injectionIndex[source.alpha] = %v", injectors)
	}

	// Full parsing stays lazy: the invalid body surfaces only on Get.
	if _, err := repo.Get("alpha"); err == nil {
		t.Error("Get(alpha) should fail on the invalid grammar body")
	}
}

func TestIndexMalformedFallsBack(t *testing.T) {
	for name, indexData := range map[string]string{
		"invalid json":        `{not json`,
		"unsupported version": `{"version": 2, "grammars": {"ghost": {"scopeName": "source.ghost"}}}`,
	} {
		t.Run(name, func(t *testing.T) {
			fsys := fstest.MapFS{
				"index.json": &fstest.MapFile{Data: []byte(indexData)},
				"real.json": &fstest.MapFile{
					Data: []byte(`{"scopeName": "source.real", "name": "real", "patterns": []}`),
				},
			}

			repo, err := NewRepository(fsys)
			if err != nil {
				t.Fatalf("NewRepository: %v", err)
			}
			if name, ok := repo.scopeIndex["source.real"]; !ok || name != "real" {
				t.Errorf("fallback scan should index real.json: got %q, %v", name, ok)
			}
			if _, ok := repo.scopeIndex["source.ghost"]; ok {
				t.Error("unsupported index version must not populate lookups")
			}
		})
	}
}

func TestFallbackScanSkipsIndexFile(t *testing.T) {
	fsys := fstest.MapFS{
		// Malformed index forces the fallback scan, which must not treat
		// index.json as a grammar.
		"index.json": &fstest.MapFile{Data: []byte(`{not json`)},
		"real.json": &fstest.MapFile{
			Data: []byte(`{"scopeName": "source.real", "name": "real", "patterns": []}`),
		},
	}

	repo, err := NewRepository(fsys)
	if err != nil {
		t.Fatalf("NewRepository: %v", err)
	}
	for scope, name := range repo.scopeIndex {
		if name == "index" {
			t.Errorf("scan indexed index.json as a grammar (scope %q)", scope)
		}
	}
}

func TestIndexInjectorOrderDeterministic(t *testing.T) {
	grammarBody := func(scope string) []byte {
		return []byte(`{"scopeName": "` + scope + `", "name": "x", "patterns": []}`)
	}
	fsys := fstest.MapFS{
		"index.json": &fstest.MapFile{
			Data: []byte(`{
				"version": 1,
				"grammars": {
					"zeta":  {"scopeName": "source.zeta", "injectTo": ["text.target"]},
					"alpha": {"scopeName": "source.alpha", "injectTo": ["text.target"]}
				}
			}`),
		},
		"alpha.json": &fstest.MapFile{Data: grammarBody("source.alpha")},
		"zeta.json":  &fstest.MapFile{Data: grammarBody("source.zeta")},
	}

	repo, err := NewRepository(fsys)
	if err != nil {
		t.Fatalf("NewRepository: %v", err)
	}

	injectors, err := repo.GetInjectors("text.target")
	if err != nil {
		t.Fatalf("GetInjectors: %v", err)
	}
	if len(injectors) != 2 {
		t.Fatalf("GetInjectors count = %d, want 2", len(injectors))
	}
	if injectors[0].ScopeName != "source.alpha" || injectors[1].ScopeName != "source.zeta" {
		t.Errorf("injector order = [%s, %s], want sorted grammar name order",
			injectors[0].ScopeName, injectors[1].ScopeName)
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
