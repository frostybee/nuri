package oniguruma

import (
	"context"
	"fmt"

	"github.com/frostybee/nuri/resources/wasm"
	"github.com/tetratelabs/wazero"
)

type Engine struct {
	runtime  wazero.Runtime
	compiled wazero.CompiledModule
}

func NewEngine(ctx context.Context) (*Engine, error) {
	cfg := wazero.NewRuntimeConfig().WithCloseOnContextDone(true)
	rt := wazero.NewRuntimeWithConfig(ctx, cfg)

	_, err := rt.NewHostModuleBuilder("env").
		NewFunctionBuilder().
		WithFunc(func(ctx context.Context, idx int32) {}).
		WithParameterNames("memory_index").
		Export("emscripten_notify_memory_growth").
		Instantiate(ctx)
	if err != nil {
		rt.Close(ctx)
		return nil, fmt.Errorf("oniguruma: stub env imports: %w", err)
	}

	compiled, err := rt.CompileModule(ctx, wasm.OnigWasm)
	if err != nil {
		rt.Close(ctx)
		return nil, fmt.Errorf("oniguruma: compile wasm: %w", err)
	}

	return &Engine{
		runtime:  rt,
		compiled: compiled,
	}, nil
}

func (e *Engine) Close(ctx context.Context) error {
	return e.runtime.Close(ctx)
}

func (e *Engine) newInstance(ctx context.Context) (*instance, error) {
	mod, err := e.runtime.InstantiateModule(ctx, e.compiled, wazero.NewModuleConfig())
	if err != nil {
		return nil, fmt.Errorf("oniguruma: instantiate module: %w", err)
	}

	inst := &instance{
		module:         mod,
		mem:            mod.Memory(),
		fnMalloc:       mod.ExportedFunction("malloc"),
		fnFree:         mod.ExportedFunction("free"),
		fnInit:         mod.ExportedFunction("onig_scanner_init"),
		fnCreateScanner: mod.ExportedFunction("create_onig_scanner"),
		fnFindNextMatch: mod.ExportedFunction("find_next_match"),
		fnFreeScanner:  mod.ExportedFunction("free_onig_scanner"),
		fnGetLastError: mod.ExportedFunction("get_last_onig_error"),
	}

	results, err := inst.fnInit.Call(ctx)
	if err != nil {
		mod.Close(ctx)
		return nil, fmt.Errorf("oniguruma: onig_scanner_init call: %w", err)
	}
	if results[0] != 0 {
		mod.Close(ctx)
		return nil, fmt.Errorf("oniguruma: onig_scanner_init returned error %d", results[0])
	}

	return inst, nil
}
