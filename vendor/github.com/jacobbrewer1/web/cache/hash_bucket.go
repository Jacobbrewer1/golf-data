package cache

// HashBucket is an interface that defines a method to check if a given key belongs to the current bucket.
type HashBucket interface {
	// InBucket checks if the provided key belongs to the current bucket.
	InBucket(key string) bool
}
