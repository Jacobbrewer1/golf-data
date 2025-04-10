package cache

type HashBucket interface { // nolint:iface // Used outside the package and allows for future different caches to take to the interface
	InBucket(key string) bool
}
