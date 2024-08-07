package vimebu

import (
	"github.com/wazazaby/gs"
)

const (
	// smallBufferSize is an initial allocation minimal capacity.
	smallBufferSize = 64
)

var (
	DefaultBuilderPool = NewBuilderPool()
)

func NewBuilderPool() *BuilderPool {
	return &BuilderPool{
		pool: gs.NewPool(func() *Builder {
			return &Builder{
				buf: make([]byte, 0, smallBufferSize),
			}
		}),
	}
}

type BuilderPool struct {
	pool gs.Pool[*Builder]
}

func (p *BuilderPool) Acquire() *Builder {
	return p.pool.Get()
}

func (p *BuilderPool) Release(b *Builder) {
	b.Reset()
	p.pool.Put(b)
}
