package registry

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"slices"
	"strings"
	"sync"

	"github.com/frostybee/nuri/internal/grammar"
)

type Repository struct {
	fsys           fs.FS
	mu             sync.RWMutex
	grammars       map[string]*grammar.Grammar // name → parsed grammar
	scopeIndex     map[string]string           // scopeName → name
	registered     map[string][]byte           // name → raw JSON (programmatic)
	injectionIndex map[string][]string         // targetScope → []grammarNames that inject into it
}

func NewRepository(fsys fs.FS) (*Repository, error) {
	r := &Repository{
		fsys:           fsys,
		grammars:       make(map[string]*grammar.Grammar),
		scopeIndex:     make(map[string]string),
		registered:     make(map[string][]byte),
		injectionIndex: make(map[string][]string),
	}
	if err := r.buildScopeIndex(); err != nil {
		return nil, err
	}
	r.buildInjectionIndex()
	return r, nil
}

func (r *Repository) Get(name string) (*grammar.Grammar, error) {
	r.mu.RLock()
	if g, ok := r.grammars[name]; ok {
		r.mu.RUnlock()
		return g, nil
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	if g, ok := r.grammars[name]; ok {
		return g, nil
	}

	data, err := r.readGrammarData(name)
	if err != nil {
		return nil, fmt.Errorf("grammar %q: %w", name, err)
	}

	g, err := grammar.ParseGrammar(data)
	if err != nil {
		return nil, fmt.Errorf("grammar %q: %w", name, err)
	}

	r.grammars[name] = g
	return g, nil
}

func (r *Repository) GetByScope(scope string) (*grammar.Grammar, error) {
	r.mu.RLock()
	name, ok := r.scopeIndex[scope]
	r.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("no grammar for scope %q", scope)
	}
	return r.Get(name)
}

func (r *Repository) Register(name string, data []byte) error {
	var probe struct {
		ScopeName string   `json:"scopeName"`
		InjectTo  []string `json:"injectTo"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return fmt.Errorf("register %q: %w", name, err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.registered[name] = data
	delete(r.grammars, name)
	if probe.ScopeName != "" {
		r.scopeIndex[probe.ScopeName] = name
	}
	for _, target := range probe.InjectTo {
		r.injectionIndex[target] = append(r.injectionIndex[target], name)
	}
	return nil
}

// GetGrammarByScope satisfies grammar.GrammarResolver.
func (r *Repository) GetGrammarByScope(scope string) (*grammar.Grammar, error) {
	return r.GetByScope(scope)
}

// LoadedGrammars returns the names of all currently cached grammars, sorted.
func (r *Repository) LoadedGrammars() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.grammars))
	for name := range r.grammars {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

func (r *Repository) readGrammarData(name string) ([]byte, error) {
	if data, ok := r.registered[name]; ok {
		return data, nil
	}
	if r.fsys == nil {
		return nil, fmt.Errorf("not found")
	}
	return fs.ReadFile(r.fsys, name+".json")
}

func (r *Repository) buildScopeIndex() error {
	if r.fsys == nil {
		return nil
	}

	entries, err := fs.ReadDir(r.fsys, ".")
	if err != nil {
		return nil
	}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".json")

		data, err := fs.ReadFile(r.fsys, e.Name())
		if err != nil {
			continue
		}
		var probe struct {
			ScopeName string `json:"scopeName"`
		}
		if err := json.Unmarshal(data, &probe); err != nil || probe.ScopeName == "" {
			continue
		}
		r.scopeIndex[probe.ScopeName] = name
	}
	return nil
}

func (r *Repository) buildInjectionIndex() {
	if r.fsys == nil {
		return
	}

	entries, err := fs.ReadDir(r.fsys, ".")
	if err != nil {
		return
	}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".json")

		data, err := fs.ReadFile(r.fsys, e.Name())
		if err != nil {
			continue
		}
		var probe struct {
			InjectTo []string `json:"injectTo"`
		}
		if err := json.Unmarshal(data, &probe); err != nil || len(probe.InjectTo) == 0 {
			continue
		}
		for _, target := range probe.InjectTo {
			r.injectionIndex[target] = append(r.injectionIndex[target], name)
		}
	}
}

// GetInjectors returns parsed grammars that inject into the given target scope.
func (r *Repository) GetInjectors(targetScope string) ([]*grammar.Grammar, error) {
	r.mu.RLock()
	names := r.injectionIndex[targetScope]
	r.mu.RUnlock()

	if len(names) == 0 {
		return nil, nil
	}

	grammars := make([]*grammar.Grammar, 0, len(names))
	for _, name := range names {
		g, err := r.Get(name)
		if err != nil {
			continue
		}
		grammars = append(grammars, g)
	}
	return grammars, nil
}
