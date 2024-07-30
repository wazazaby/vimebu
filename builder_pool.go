package vimebu

import "github.com/wazazaby/gs"

var (
	DefaultBuilderPool = NewBuilderPool()
)

func NewBuilderPool() *BuilderPool {
	return &BuilderPool{
		pool: gs.NewPool(func() *Builder {
			return new(Builder)
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
