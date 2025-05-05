package cache

// HashBucket is an interface that defines a method to check if a given key belongs to the current bucket.
//
// This interface is designed to be used outside the package and allows for future implementations
// of different caching mechanisms to conform to this interface.
type HashBucket interface { // nolint:iface // Used outside the package and allows for future different caches to take to the interface
	// InBucket checks if the provided key belongs to the current bucket.
	//
	// Parameters:
	//   - key: A string representing the key to check.
	//
	// Returns:
	//   - A boolean value indicating whether the key is in the current bucket.
	InBucket(key string) bool
}
