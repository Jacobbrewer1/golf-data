package slices

import (
	"sync"
)

// Set represents a mathematical set: https://en.wikipedia.org/wiki/Set_(mathematics)#
//
// A Set provides a thread-safe collection of unique elements of type T.
// The elements must be comparable as defined by the Go language specification.
//
// The implementation uses a map with bool values to represent set membership,
// with synchronization provided by a read-write mutex for concurrent access.
type Set[T comparable] struct {
	// mut is a read-write mutex that ensures thread-safe access to the set.
	mut *sync.RWMutex

	// items is a map where the keys represent the elements of the set,
	// and the values are always true to indicate the presence of an element.
	items map[T]bool
}

// NewSet creates a new set with the provided items.
//
// This function initializes a new instance of the Set type with the given items.
// It ensures that the set is thread-safe by initializing a read-write mutex
// and uses a map to store the unique elements.
//
// Parameters:
//   - items (...T): A variadic parameter representing the initial elements to add to the set.
//
// Returns:
//   - *Set[T]: A pointer to the newly created Set instance.
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
//
// This method ensures thread-safe access by acquiring a write lock
// before adding the specified item to the set.
//
// Parameters:
//   - item (T): The element to be added to the set.
func (s *Set[T]) Add(item T) {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.items[item] = true
}

// Remove deletes an element from the set.
//
// This method ensures thread-safe access by acquiring a write lock
// before removing the specified item from the set.
//
// Parameters:
//   - item (T): The element to be removed from the set.
func (s *Set[T]) Remove(item T) {
	s.mut.Lock()
	defer s.mut.Unlock()
	delete(s.items, item)
}

// Contains checks if the set contains the specified item.
//
// This method ensures thread-safe access by acquiring a read lock
// before checking for the presence of the item in the set.
//
// Parameters:
//   - item (T): The element to check for in the set.
//
// Returns:
//   - bool: True if the item exists in the set, otherwise false.
func (s *Set[T]) Contains(item T) bool {
	s.mut.RLock()
	defer s.mut.RUnlock()
	_, ok := s.items[item]
	return ok
}

// Difference returns a new set containing elements that are in the current set (A)
// but not in the provided set (B).
//
// This method ensures thread-safe access by acquiring read locks on both sets
// before performing the operation.
//
// Parameters:
//   - b (*Set[T]): The set to compare against.
//
// Returns:
//   - *Set[T]: A new set containing elements that are in the current set but not in set B.
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
//
// This method ensures thread-safe access by acquiring read locks on both sets
// before performing the union operation.
//
// Parameters:
//   - b (*Set[T]): The set to union with the current set.
//
// Returns:
//   - *Set[T]: A new set containing all unique elements from both sets.
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
//
// This method ensures thread-safe access by acquiring a read lock
// before iterating over the elements of the set.
//
// Parameters:
//   - fn (func(item T)): A function to be executed for each element in the set.
func (s *Set[T]) Each(fn func(item T)) {
	s.mut.RLock()
	defer s.mut.RUnlock()

	for k := range s.items {
		fn(k)
	}
}

// Items returns the items in the set as a slice.
//
// This method ensures thread-safe access by acquiring a read lock
// before iterating over the elements of the set and collecting them into a slice.
//
// Returns:
//   - []T: A slice containing all the elements in the set.
func (s *Set[T]) Items() []T {
	s.mut.RLock()
	defer s.mut.RUnlock()

	keys := make([]T, 0, len(s.items))
	for k := range s.items {
		keys = append(keys, k)
	}

	return keys
}
