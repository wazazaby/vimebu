package vimebu

import (
	"sync"
)

type buffer struct {
	f []byte
}

// bytesBufferPool is a simple pool to create or retrieve a [bytes.Buffer].
var bytesBufferPool = sync.Pool{
	New: func() any {
		return &buffer{
			f: make([]byte, 0, 64),
		}
	},
}

// getBuffer acquires a [bytes.Buffer] from the pool.
func getBuffer() *buffer {
	return bytesBufferPool.Get().(*buffer)
}

// putBuffer resets and returns a [bytes.Buffer] to the pool.
func putBuffer(b *buffer) {
	if b == nil {
		return
	}
	b.f = b.f[:0]
	bytesBufferPool.Put(b)
}
