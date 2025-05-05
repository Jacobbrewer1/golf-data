package cache

import (
	"strconv"

	"github.com/serialx/hashring"
)

// Ensures that FixedHashBucket implements the HashBucket interface.
//
// This line is a compile-time check to guarantee that FixedHashBucket
// satisfies all the methods defined in the HashBucket interface.
var _ HashBucket = new(FixedHashBucket)

// FixedHashBucket represents a mechanism for determining whether supplied keys belong to the current bucket index.
// It uses a consistent hashing ring to distribute keys across buckets based on their hash values.
// The bucket index is advanced in a round-robin fashion, ensuring even distribution of keys over time.
type FixedHashBucket struct {
	// hr is the consistent hashing ring used to map keys to bucket indices.
	hr *hashring.HashRing

	// index is the current bucket index. It determines which bucket is active for key assignment.
	index int
}

// NewFixedHashBucket creates and initializes a new FixedHashBucket with the specified size.
// It sets up a consistent hashing ring and assigns nodes to it based on the provided size.
//
// Parameters:
//   - size: The number of nodes to add to the hash ring.
//
// Returns:
//   - A pointer to the newly created FixedHashBucket instance.
func NewFixedHashBucket(size uint) *FixedHashBucket {
	h := &FixedHashBucket{
		hr:    hashring.New(make([]string, 0)), // Initialize an empty hash ring.
		index: 0,                               // Start with the bucket index set to 0.
	}

	// Add nodes to the hash ring, based on the size, where each node is represented as a stringified number.
	for i := uint64(0); i < uint64(size); i++ { // nolint:intrange // Avoid range loops for uint64.
		h.hr = h.hr.AddNode(strconv.FormatUint(i, 10))
	}

	return h
}

// InBucket determines if the given key belongs to the current bucket index.
//
// The function uses the consistent hashing ring to find the position of the key
// and checks if it matches the current bucket index.
//
// Parameters:
//   - key: The key to check, represented as a string.
//
// Returns:
//   - A boolean value indicating whether the key resides in the current bucket.
func (h *FixedHashBucket) InBucket(key string) bool {
	pos, _ := h.hr.GetNodePos(key)
	return pos == h.index
}

// Advance increments the bucket index to point to the next bucket in the hash ring.
// If the index exceeds the size of the hash ring, it wraps around to 0.
//
// This ensures that the bucket index cycles through all available buckets in a round-robin fashion.
func (h *FixedHashBucket) Advance() {
	if h.index+1 >= h.hr.Size() {
		h.index = 0
	} else {
		h.index++
	}
}
