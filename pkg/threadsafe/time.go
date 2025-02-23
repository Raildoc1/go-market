package threadsafe

import (
	"sync"
	"time"
)

type Time struct {
	time time.Time
	mux  *sync.Mutex
}

func NewTime(t time.Time) *Time {
	return &Time{
		time: t,
		mux:  &sync.Mutex{},
	}
}

func (t *Time) Get() time.Time {
	t.mux.Lock()
	defer t.mux.Unlock()
	return t.time
}

func (t *Time) Set(value time.Time) {
	t.mux.Lock()
	defer t.mux.Unlock()
	t.time = value
}

func (t *Time) SetIf(value time.Time, condition func(current time.Time) bool) {
	t.mux.Lock()
	defer t.mux.Unlock()
	if condition(t.time) {
		t.time = value
	}
}
