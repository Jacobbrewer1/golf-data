package slices

import (
	"sync"
)

// Set represents a mathematical set: https://en.wikipedia.org/wiki/Set_(mathematics)#
type Set[T comparable] struct {
	// mut is a read-write mutex that ensures thread-safe access to the set.
	mut *sync.RWMutex

	// items is a map where the keys represent the elements of the set,
	// and the values are always true to indicate the presence of an element.
	items map[T]bool
}

// NewSet creates a new set with the provided items.
func NewSet[T comparable](items ...T) *Set[T] {
	set := &Set[T]{
		mut:   new(sync.RWMutex),
		items: make(map[T]bool),
	}

	for _, k := range items {
		set.items[k] = true
	}

	return set
}

// Add inserts a new element into the set.
func (s *Set[T]) Add(item T) {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.items[item] = true
}

// Remove deletes an element from the set.
func (s *Set[T]) Remove(item T) {
	s.mut.Lock()
	defer s.mut.Unlock()
	delete(s.items, item)
}

// Contains checks if the set contains the specified item.
func (s *Set[T]) Contains(item T) bool {
	s.mut.RLock()
	defer s.mut.RUnlock()
	_, ok := s.items[item]
	return ok
}

// Difference returns a new set containing elements that are in the current set (A)
// but not in the provided set (B).
func (s *Set[T]) Difference(b *Set[T]) *Set[T] {
	s.mut.RLock()
	b.mut.RLock()
	defer b.mut.RUnlock()
	defer s.mut.RUnlock()

	s3 := NewSet[T]()
	for k := range s.items {
		if !b.items[k] {
			s3.Add(k)
		}
	}
	return s3
}

// Union returns a new set containing all elements that are in either the current set (A),
// the provided set (B), or both.
func (s *Set[T]) Union(b *Set[T]) *Set[T] {
	s.mut.RLock()
	b.mut.RLock()
	defer b.mut.RUnlock()
	defer s.mut.RUnlock()

	s3 := NewSet[T]()
	for k := range s.items {
		s3.Add(k)
	}

	for k := range b.items {
		s3.Add(k)
	}
	return s3
}

// Each calls the provided function on each item in the set.
func (s *Set[T]) Each(fn func(item T)) {
	s.mut.RLock()
	defer s.mut.RUnlock()

	for k := range s.items {
		fn(k)
	}
}

// Items returns the items in the set as a slice.
func (s *Set[T]) Items() []T {
	s.mut.RLock()
	defer s.mut.RUnlock()

	keys := make([]T, 0, len(s.items))
	for k := range s.items {
		keys = append(keys, k)
	}

	return keys
}
