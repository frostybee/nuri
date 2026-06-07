package oniguruma

import (
	"context"
	"encoding/binary"
	"fmt"

	"github.com/tetratelabs/wazero/api"
)

type instance struct {
	module   api.Module
	mem      api.Memory
	poisoned bool

	fnMalloc        api.Function
	fnFree          api.Function
	fnInit          api.Function
	fnCreateScanner api.Function
	fnFindNextMatch api.Function
	fnFreeScanner   api.Function
	fnGetLastError  api.Function
}

func (inst *instance) close(ctx context.Context) error {
	return inst.module.Close(ctx)
}

func (inst *instance) wasmAlloc(ctx context.Context, size uint32) (uint32, error) {
	results, err := inst.fnMalloc.Call(ctx, uint64(size))
	if err != nil {
		return 0, fmt.Errorf("malloc(%d): %w", size, err)
	}
	ptr := uint32(results[0])
	if ptr == 0 {
		return 0, fmt.Errorf("malloc(%d) returned null", size)
	}
	return ptr, nil
}

func (inst *instance) wasmFree(ctx context.Context, ptr uint32) {
	inst.fnFree.Call(ctx, uint64(ptr))
}

func (inst *instance) lastError(ctx context.Context) string {
	const bufSize = 256
	ptr, err := inst.wasmAlloc(ctx, bufSize)
	if err != nil {
		return "(cannot read error)"
	}
	defer inst.wasmFree(ctx, ptr)

	results, err := inst.fnGetLastError.Call(ctx, uint64(ptr), uint64(bufSize))
	if err != nil || results[0] == 0 {
		return "(cannot read error)"
	}
	length := uint32(results[0])
	buf, ok := inst.mem.Read(ptr, length)
	if !ok {
		return "(cannot read error)"
	}
	return string(buf)
}

func (inst *instance) NewScanner(ctx context.Context, patterns [][]byte) (*Scanner, error) {
	if len(patterns) == 0 {
		return nil, fmt.Errorf("oniguruma: no patterns provided")
	}

	totalLen := 0
	for _, p := range patterns {
		totalLen += len(p)
	}

	patBuf := make([]byte, totalLen)
	lengths := make([]int32, len(patterns))
	off := 0
	for i, p := range patterns {
		copy(patBuf[off:], p)
		lengths[i] = int32(len(p))
		off += len(p)
	}

	patPtr, err := inst.wasmAlloc(ctx, uint32(totalLen))
	if err != nil {
		return nil, fmt.Errorf("oniguruma: alloc patterns: %w", err)
	}
	inst.mem.Write(patPtr, patBuf)

	lenBufSize := uint32(len(patterns) * 4)
	lenPtr, err := inst.wasmAlloc(ctx, lenBufSize)
	if err != nil {
		inst.wasmFree(ctx, patPtr)
		return nil, fmt.Errorf("oniguruma: alloc lengths: %w", err)
	}
	lenBuf := make([]byte, lenBufSize)
	for i, l := range lengths {
		binary.LittleEndian.PutUint32(lenBuf[i*4:], uint32(l))
	}
	inst.mem.Write(lenPtr, lenBuf)

	results, err := inst.fnCreateScanner.Call(ctx,
		uint64(patPtr), uint64(lenPtr), uint64(len(patterns)))
	inst.wasmFree(ctx, patPtr)
	inst.wasmFree(ctx, lenPtr)

	if err != nil {
		return nil, fmt.Errorf("oniguruma: create_onig_scanner call: %w", err)
	}
	scannerPtr := uint32(results[0])
	if scannerPtr == 0 {
		errMsg := inst.lastError(ctx)
		return nil, fmt.Errorf("oniguruma: create_onig_scanner failed: %s", errMsg)
	}

	return &Scanner{
		inst: inst,
		ptr:  scannerPtr,
	}, nil
}

func (inst *instance) NewScannerCtx(ctx context.Context, patterns [][]byte) (OnigScanner, error) {
	return inst.NewScanner(ctx, patterns)
}

func (inst *instance) Close() error {
	return inst.close(context.Background())
}
