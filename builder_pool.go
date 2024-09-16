package vimebu

import "sync"

const (
	// smallBufferSize is an initial allocation minimal capacity.
	smallBufferSize int = 64
)

var (
	DefaultBuilderPool = NewBuilderPool()
)

// NewBuilderPool creates a new [BuilderPool] instance.
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

// BuilderPool is a strongly typed wrapper around a [sync.Pool], specifically used to
// store and retrieve [Builder] instances.
type BuilderPool struct {
	pool sync.Pool
}

// Acquire returns an empty [Builder] instance from the specified pool.
//
// Release the [Builder] with [BuilderPool.Release] after the [Builder] is no longer needed.
// This allows reducing GC load.
func (p *BuilderPool) Acquire() *Builder {
	return p.pool.Get().(*Builder)
}

// AcquireBuilder returns an empty [Builder] instance from the default pool.
//
// Release the [Builder] with [ReleaseBuilder] after the [Builder] is no longer needed.
// This allows reducing GC load.
func AcquireBuilder() *Builder {
	return DefaultBuilderPool.Acquire()
}

// Release releases the [Builder] acquired via [BuilderPool.Acquire] to the specified pool.
//
// The released [Builder] mustn't be used after releasing it, otherwise data races
// may occur.
func (p *BuilderPool) Release(b *Builder) {
	b.Reset()
	p.pool.Put(b)
}

// Release releases the [Builder] acquired via [AcquireBuilder] to the default pool.
//
// The released [Builder] mustn't be used after releasing it, otherwise data races
// may occur.
func ReleaseBuilder(b *Builder) {
	DefaultBuilderPool.Release(b)
}

// Metric acquires and returns a zeroed-out [Builder] instance from the
// specified pool and sets the metric's name.
func (p *BuilderPool) Metric(name string, options ...BuilderOption) *Builder {
	b := p.Acquire()
	b.pool = p
	return b.Metric(name, options...)
}
