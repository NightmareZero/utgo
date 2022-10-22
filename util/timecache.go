package util

import (
	"sync"
	"time"
)

var _ TimeCache[int] = &timeCache[int]{}

type TimeCache[T any] interface {
	Get() T
	Reset()
}

type timeCache[T any] struct {
	data    T
	lock    sync.Mutex
	getTime time.Time
	expire  time.Duration
	getter  func() T
}

// Get implements TimeCache
func (tc *timeCache[T]) Get() T {
	if tc.isTimeout(tc.expire) {
		tc.lock.Lock()
		defer tc.lock.Unlock()
		tc.data = tc.getter()
		tc.getTime = time.Now()
		return tc.data
	} else if tc.isTimeout(tc.expire / 2) {
		go func() {
			b := tc.lock.TryLock()
			if b {
				defer tc.lock.Unlock()
				tc.data = tc.getter()
				tc.getTime = time.Now()
			}
		}()
	}
	return tc.data
}

// Reset implements TimeCache
func (tc *timeCache[T]) Reset() {
	tc.lock.Lock()
	defer tc.lock.Unlock()
	tc.getTime = time.Now().Add(-tc.expire)
}

func (tc *timeCache[T]) isTimeout(exp time.Duration) bool {
	return tc.getTime.Add(tc.expire).Before(time.Now())
}

func NewTimeCache[T any](expire time.Duration, getter func() T) *timeCache[T] {
	tc := timeCache[T]{
		expire: expire,
		getter: getter,
	}

	return &tc
}
