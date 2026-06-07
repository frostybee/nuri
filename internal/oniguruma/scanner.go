package oniguruma

import (
	"context"
	"encoding/binary"
	"fmt"
)

const maxResultSlots = 2 + 64*2

type Scanner struct {
	inst *instance
	ptr  uint32
}

func (s *Scanner) FindNextMatch(ctx context.Context, text []byte, pos int, options SearchOptions) (*Match, error) {
	if len(text) == 0 {
		return nil, nil
	}

	txtPtr, err := s.inst.wasmAlloc(ctx, uint32(len(text)))
	if err != nil {
		return nil, fmt.Errorf("oniguruma: alloc text: %w", err)
	}
	defer s.inst.wasmFree(ctx, txtPtr)
	s.inst.mem.Write(txtPtr, text)

	resultBufSize := uint32(maxResultSlots * 4)
	resultPtr, err := s.inst.wasmAlloc(ctx, resultBufSize)
	if err != nil {
		return nil, fmt.Errorf("oniguruma: alloc result buf: %w", err)
	}
	defer s.inst.wasmFree(ctx, resultPtr)

	results, err := s.inst.fnFindNextMatch.Call(ctx,
		uint64(s.ptr),
		uint64(txtPtr), uint64(len(text)), uint64(pos),
		uint64(resultPtr), uint64(maxResultSlots),
		uint64(options),
	)
	if err != nil {
		return nil, fmt.Errorf("oniguruma: find_next_match call: %w", err)
	}

	numCaptures := int32(results[0])
	if numCaptures < 0 {
		if numCaptures == -1 {
			return nil, nil
		}
		return nil, fmt.Errorf("oniguruma: find_next_match error code %d", numCaptures)
	}

	slotsToRead := uint32(2+numCaptures*2) * 4
	raw, ok := s.inst.mem.Read(resultPtr, slotsToRead)
	if !ok {
		return nil, fmt.Errorf("oniguruma: cannot read result buffer")
	}

	patternIndex := int(int32(binary.LittleEndian.Uint32(raw[0:])))

	captures := make([]Capture, numCaptures)
	for i := int32(0); i < numCaptures; i++ {
		off := (2 + i*2) * 4
		captures[i] = Capture{
			Start: int(int32(binary.LittleEndian.Uint32(raw[off:]))),
			End:   int(int32(binary.LittleEndian.Uint32(raw[off+4:]))),
		}
	}

	return &Match{
		Index:    patternIndex,
		Captures: captures,
	}, nil
}

func (s *Scanner) Free(ctx context.Context) {
	if s.ptr != 0 {
		s.inst.fnFreeScanner.Call(ctx, uint64(s.ptr))
		s.ptr = 0
	}
}

func (s *Scanner) FindNextMatchCtx(ctx context.Context, text []byte, startPos int, options SearchOptions) (*Match, error) {
	return s.FindNextMatch(ctx, text, startPos, options)
}

func (s *Scanner) Close() error {
	s.Free(context.Background())
	return nil
}
