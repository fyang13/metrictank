package snappy

import (
	"bytes"
	"encoding/binary"
	"sync"

	master "github.com/golang/snappy"
)

type BufferPool struct {
	pool sync.Pool
}

func (bp *BufferPool) Get() []byte {
	ret := bp.pool.Get()
	if ret == nil {
		return nil
	}
	return ret.([]byte)[:0]
}
func (bp *BufferPool) Put(b []byte) {
	bp.pool.Put(b)
}

var workspacePool BufferPool

var xerialHeader = []byte{130, 83, 78, 65, 80, 80, 89, 0}

// Encode encodes data as snappy with no framing header.
func Encode(src []byte) []byte {
	return master.Encode(nil, src)
}

// Decode decodes snappy data whether it is traditional unframed
// or includes the xerial framing format.
func Decode(src []byte) ([]byte, error) {
	if !bytes.Equal(src[:8], xerialHeader) {
		return master.Decode(nil, src)
	}

	dst := make([]byte, 0, len(src))

	return DecodeInto(dst, src)
}

// DecodePreAlloc decodes snappy data whether it is traditional unframed
// or includes the xerial framing format.
func DecodeInto(dst, src []byte) ([]byte, error) {
	if !bytes.Equal(src[:8], xerialHeader) {
		return master.Decode(dst[:cap(dst)], src)
	}

	if dst == nil {
		dst = bytePool.Get(len(src) * 2)
	}

	dst = dst[:0]
	var (
		pos       = uint32(16)
		max       = uint32(len(src))
		workspace = workspacePool.Get()
		err       error
	)

	for pos < max {
		size := binary.BigEndian.Uint32(src[pos : pos+4])
		pos += 4

		workspace = workspace[:cap(workspace)]
		workspace, err = master.Decode(workspace, src[pos:pos+size])
		if err != nil {
			workspacePool.Put(workspace)
			return dst, err
		}

		pos += size
		origDst := dst
		dst = append(dst, workspace...)
		if len(origDst) > 0 && &origDst[0] != &dst[0] {
			// Re-alloc, stick the old one into the pool for reuse
			bytePool.Put(origDst)
		}
	}
	workspacePool.Put(workspace)
	return dst, nil
}
