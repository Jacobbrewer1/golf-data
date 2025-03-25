package cache

import "context"

type HashBucket interface {
	Start(ctx context.Context) error
	InBucket(key string) bool
}
