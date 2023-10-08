package vimebu

import (
	"bytes"
	"sync"
)

// bytesBufferPool is a simple pool to create or retrieve a [bytes.Buffer].
var bytesBufferPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

// getBuffer acquires a [bytes.Buffer] from the pool.
func getBuffer() *bytes.Buffer {
	return bytesBufferPool.Get().(*bytes.Buffer)
}

// putBuffer resets and returns a [bytes.Buffer] to the pool.
func putBuffer(b *bytes.Buffer) {
	if b == nil {
		return
	}
	b.Reset()
	bytesBufferPool.Put(b)
}
