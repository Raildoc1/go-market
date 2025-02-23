package threadsafe

import "sync"

type HashSet[T int | string] struct {
	inner map[T]struct{}
	mux   *sync.Mutex
}

func NewHashSet[T int | string]() *HashSet[T] {
	return &HashSet[T]{
		inner: make(map[T]struct{}),
		mux:   &sync.Mutex{},
	}
}

func (h *HashSet[T]) Add(item T) bool {
	h.mux.Lock()
	defer h.mux.Unlock()
	if _, ok := h.inner[item]; ok {
		return false
	}
	h.inner[item] = struct{}{}
	return true
}

func (h *HashSet[T]) Remove(item T) bool {
	h.mux.Lock()
	defer h.mux.Unlock()
	if _, ok := h.inner[item]; !ok {
		return false
	}
	delete(h.inner, item)
	return true
}

func (h *HashSet[T]) Contains(item T) bool {
	h.mux.Lock()
	defer h.mux.Unlock()
	_, ok := h.inner[item]
	return ok
}
