package oniguruma

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type Pool struct {
	eng *Engine
	ch  chan *instance
	all []*instance
	mu  sync.Mutex
}

func NewPool(ctx context.Context, eng *Engine, size int) (*Pool, error) {
	if size < 1 {
		return nil, fmt.Errorf("oniguruma: pool size must be >= 1")
	}

	p := &Pool{
		eng: eng,
		ch:  make(chan *instance, size),
		all: make([]*instance, 0, size),
	}

	for i := 0; i < size; i++ {
		inst, err := eng.newInstance(ctx)
		if err != nil {
			p.Close(ctx)
			return nil, fmt.Errorf("oniguruma: create pool instance %d: %w", i, err)
		}
		p.all = append(p.all, inst)
		p.ch <- inst
	}

	return p, nil
}

func (p *Pool) Get(ctx context.Context) (*instance, error) {
	select {
	case inst := <-p.ch:
		return inst, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (p *Pool) Put(inst *instance) {
	p.ch <- inst
}

func (p *Pool) Swap(ctx context.Context, old *instance) error {
	fresh, err := p.eng.newInstance(ctx)
	if err != nil {
		return fmt.Errorf("oniguruma: create replacement instance: %w", err)
	}

	p.mu.Lock()
	for i, inst := range p.all {
		if inst == old {
			p.all[i] = fresh
			break
		}
	}
	p.mu.Unlock()

	old.close(ctx)
	p.ch <- fresh
	return nil
}

// Do borrows an instance from the pool, passes it to fn as OnigLib, and
// returns it when fn completes. If fn panics, the instance is poisoned and
// replaced with a fresh one.
func (p *Pool) Do(ctx context.Context, fn func(OnigLib) error) (retErr error) {
	inst, err := p.Get(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			inst.poisoned = true
			retErr = fmt.Errorf("oniguruma: panic in Do: %v", r)
		}
		if inst.poisoned {
			if swapErr := p.Swap(ctx, inst); swapErr != nil {
				retErr = errors.Join(retErr, swapErr)
			}
		} else {
			p.Put(inst)
		}
	}()

	return fn(inst)
}

func (p *Pool) Close(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var firstErr error
	for _, inst := range p.all {
		if err := inst.close(ctx); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	p.all = nil
	return firstErr
}
