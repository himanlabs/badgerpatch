package badgerpatch

import "sync"

type SequenceBuilder[T any] struct {
	mu      sync.Mutex
	returns []T
	current int
}

func NewSequence[T any](values ...T) *SequenceBuilder[T] {
	return &SequenceBuilder[T]{
		returns: values,
		current: 0,
	}
}

// Next safely returns the next item in the sequence
func (s *SequenceBuilder[T]) Next() T {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.current >= len(s.returns) {
		return s.returns[len(s.returns)-1]
	}

	val := s.returns[s.current]
	s.current++
	return val
}