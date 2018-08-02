package snappy

import (
	"math"
	"sync"
)

// BytePool is safe to use from multiple go-routines for buffer reuse
var bytePool = NewSizeTieredBufferPool()

// SizeTieredBufferPool attempts to match requested sizes up
// with buffers large enough to handle them, without being wasteful.
type SizeTieredBufferPool struct {
	pools []*sync.Pool
}

// NewSizeTieredBufferPool will return a default bufferpool designed
// to hold items ranging from small (<1kb) to large (>1mb).
func NewSizeTieredBufferPool() SizeTieredBufferPool {
	return NewSizeTieredBufferPoolWithNBuckets(10)
}

// NewSizeTieredBufferPoolWithNBuckets will create a bufferpool with
// the specified number of buckets. More buckets means this pool can
// handle larger buffers gracefully.
// numBuckets cannot be greater than 30 or less than 2
func NewSizeTieredBufferPoolWithNBuckets(numBuckets int) SizeTieredBufferPool {
	if numBuckets > 30 {
		numBuckets = 30
	} else if numBuckets < 2 {
		numBuckets = 2
	}
	pools := make([]*sync.Pool, numBuckets)
	for i := range pools {
		tierSize := 1 << uint32(10+i)
		pools[i] = &sync.Pool{
			New: func() interface{} {
				return make([]byte, 0, tierSize)
			},
		}
	}
	return SizeTieredBufferPool{pools: pools}
}

// Get returns a buffer at least as large as 'size'.
func (s *SizeTieredBufferPool) Get(size int) []byte {
	bucket := s.getBucketForPow2(math.Ceil(math.Log2(float64(size))))
	return s.pools[bucket].Get().([]byte)[:0]
}

// Put returns buf to the proper size bucket for reuse
func (s *SizeTieredBufferPool) Put(buf []byte) {
	bucket := s.getBucketForPow2(math.Floor(math.Log2(float64(cap(buf)))))
	s.pools[bucket].Put(buf)
}

func (s *SizeTieredBufferPool) getBucketForPow2(exponent float64) int {
	bucket := int(exponent) - 10
	if bucket < 0 {
		bucket = 0
	} else if bucket >= len(s.pools) {
		bucket = len(s.pools) - 1
	}
	return bucket
}
