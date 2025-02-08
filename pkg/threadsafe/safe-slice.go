package threadsafe

import "sync"

type SafeSlice[T any] struct {
	inner []T
	mux   *sync.RWMutex
}

func NewSafeSlice[T any](size int) *SafeSlice[T] {
	return &SafeSlice[T]{
		inner: make([]T, size),
		mux:   &sync.RWMutex{},
	}
}

func (s *SafeSlice[T]) Size() int {
	s.mux.RLock()
	defer s.mux.RUnlock()
	return len(s.inner)
}

func (s *SafeSlice[T]) Clear() {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.inner = make([]T, 0)
}

func (s *SafeSlice[T]) Append(v T) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.inner = append(s.inner, v)
}

func (s *SafeSlice[T]) RemoveAt(i int) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.inner = append(s.inner[:i], s.inner[i+1:]...)
}
