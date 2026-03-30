package pool

import "sync"

type Resettable interface {
	Reset()
}

type Pool[T Resettable] struct {
	pool []T
	new  func() T
	mu   sync.Mutex
}

func New[T Resettable](newFunc func() T) *Pool[T] {
	return &Pool[T]{
		pool: []T{},
		new:  newFunc,
	}
}

func (p *Pool[T]) Get() T {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.pool) > 0 {
		last := len(p.pool) - 1
		obj := p.pool[last]
		p.pool = p.pool[:last]
		return obj
	}
	return p.new()
}

func (p *Pool[T]) Put(obj T) {
	obj.Reset()
	p.mu.Lock()
	p.pool = append(p.pool, obj)
	p.mu.Unlock()
}
