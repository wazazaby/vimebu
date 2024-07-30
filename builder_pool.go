package vimebu

import "sync"

const (
	// smallBufferSize is an initial allocation minimal capacity.
	smallBufferSize int = 64
)

var (
	DefaultBuilderPool = NewBuilderPool()
)

func NewBuilderPool() *BuilderPool {
	return &BuilderPool{
		pool: sync.Pool{
			New: func() any {
				return &Builder{
					buf: make([]byte, 0, smallBufferSize),
				}
			},
		},
	}
}

type BuilderPool struct {
	pool sync.Pool
}

func (p *BuilderPool) Acquire() *Builder {
	return p.pool.Get().(*Builder)
}

func (p *BuilderPool) Release(b *Builder) {
	b.Reset()
	p.pool.Put(b)
}
