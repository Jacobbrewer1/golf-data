package cache

import (
	"strconv"

	"github.com/serialx/hashring"
)

var _ HashBucket = new(FixedHashBucket)

// FixedHashBucket represents a mechanism which determines whether supplied keys belong to the current bucket index.
// It uses a consistent hashing ring to determine the bucket index for each key.
// The bucket index is advanced in a round-robin fashion, and the keys are distributed across the buckets based on their hash values.
type FixedHashBucket struct {
	hr    *hashring.HashRing
	index int
}

// NewFixedHashBucket instantiates a FixedHashBucket of the provided size.
func NewFixedHashBucket(size uint) *FixedHashBucket {
	h := &FixedHashBucket{
		hr:    hashring.New(make([]string, 0)),
		index: 0,
	}

	// Add nodes to the hash ring based on the size.
	for i := uint64(0); i < uint64(size); i++ { // nolint:intrange // Cannot use i range loops for uint64
		h.hr = h.hr.AddNode(strconv.FormatUint(i, 10))
	}

	return h
}

// InBucket returns whether the supplied key resides in the current bucket.
func (h *FixedHashBucket) InBucket(key string) bool {
	pos, _ := h.hr.GetNodePos(key)
	return pos == h.index
}

// Advance advances the current bucket index by 1. The index automatically wraps on index overflow.
func (h *FixedHashBucket) Advance() {
	if h.index+1 >= h.hr.Size() {
		h.index = 0
	} else {
		h.index++
	}
}
