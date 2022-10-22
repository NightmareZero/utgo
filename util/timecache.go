package util

import (
	"sync"
	"time"
)

var _ TimeCache[int] = &timeCache[int]{}

type TimeCache[T any] interface {
	Get() (T, error)
	Reset()
}

type timeCache[T any] struct {
	data    T
	lock    sync.Mutex
	getTime time.Time
	expire  time.Duration
	getter  func() (T, error)
	err     error
}

// Get implements TimeCache
func (tc *timeCache[T]) Get() (T, error) {
	if tc.isTimeout(tc.expire) || tc.err != nil {
		tc.lock.Lock()
		defer tc.lock.Unlock()
		tc.reget()
		return tc.data, tc.err
	} else if tc.isTimeout(tc.expire / 2) {
		go func() {
			b := tc.lock.TryLock()
			if b {
				defer tc.lock.Unlock()
				tc.reget()
			}
		}()
	}
	return tc.data, tc.err
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

func (tc *timeCache[T]) reget() {
	tc.data, tc.err = tc.getter()
	if tc.err != nil {
		return
	}
	tc.getTime = time.Now()
}

func NewTimeCache[T any](expire time.Duration, getter func() (T, error)) *timeCache[T] {
	tc := timeCache[T]{
		expire: expire,
		getter: getter,
	}

	return &tc
}
