package slices

import (
	"sync"
)

// Set represents a mathematical set: https://en.wikipedia.org/wiki/Set_(mathematics)#
type Set[T comparable] struct {
	mut   sync.Mutex
	items map[T]bool
}

func NewSet[T comparable](items ...T) *Set[T] {
	set := &Set[T]{
		items: make(map[T]bool),
	}

	for _, k := range items {
		set.items[k] = true
	}

	return set
}

// Add adds an item to the set.
func (s *Set[T]) Add(item T) {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.items[item] = true
}

// Remove removes an item from the set.
func (s *Set[T]) Remove(item T) {
	s.mut.Lock()
	defer s.mut.Unlock()
	delete(s.items, item)
}

// Contains returns whether the set contains item.
func (s *Set[T]) Contains(item T) bool {
	s.mut.Lock()
	defer s.mut.Unlock()
	_, ok := s.items[item]
	return ok
}

// Difference returns the set of all things that belong to A, but not B.
func (s *Set[T]) Difference(b *Set[T]) *Set[T] {
	s3 := NewSet[T]()
	s.mut.Lock()
	b.mut.Lock()
	defer b.mut.Unlock()
	defer s.mut.Unlock()

	for k := range s.items {
		if !b.items[k] {
			s3.Add(k)
		}
	}

	return s3
}

// Union returns the set of all things that belong in A, in B or in both.
func (s *Set[T]) Union(b *Set[T]) *Set[T] {
	s3 := NewSet[T]()
	s.mut.Lock()
	b.mut.Lock()
	defer b.mut.Unlock()
	defer s.mut.Unlock()

	for k := range s.items {
		s3.Add(k)
	}

	for k := range b.items {
		s3.Add(k)
	}

	return s3
}

// Each calls fn on each item of the set.
func (s *Set[T]) Each(fn func(item T)) {
	s.mut.Lock()
	defer s.mut.Unlock()

	for k := range s.items {
		fn(k)
	}
}

// Items returns the items in the set as a slice.
func (s *Set[T]) Items() []T {
	s.mut.Lock()
	defer s.mut.Unlock()

	keys := make([]T, 0, len(s.items))
	for k := range s.items {
		keys = append(keys, k)
	}

	return keys
}
