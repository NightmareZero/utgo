package utool

import (
	"cmp"
	"slices"
	"sync"
)

type TMap[T any] struct {
	TRMap[string, T]
}

type TRMap[T any, R any] struct {
	sync.Map
}

func (t *TRMap[T, R]) Get(key T) (r R, ok bool) {
	v, ok := t.Map.Load(key)
	if !ok {
		return r, false
	}
	return v.(R), true
}

func (t *TRMap[T, R]) Exists(key T) bool {
	_, ok := t.Map.Load(key)
	return ok
}

func (t *TRMap[T, R]) Set(key T, value R) {
	t.Map.Store(key, value)
}

func (t *TRMap[T, R]) Del(key T) {
	t.Map.Delete(key)
}

// Range calls f sequentially for each key and value present in the map.
// if return false ,break range
func (t *TRMap[T, R]) Range(f func(key T, value R) bool) {
	t.Map.Range(func(key, value interface{}) bool {
		return f(key.(T), value.(R))
	})
}

func (t *TRMap[T, R]) Len() int {
	var l int
	t.Map.Range(func(key, value interface{}) bool {
		l++
		return true
	})
	return l
}

func (t *TRMap[T, R]) Keys() []T {
	var keys []T
	t.Map.Range(func(key, value interface{}) bool {
		keys = append(keys, key.(T))
		return true
	})
	return keys
}

// 将slice转换为map
func SliceToMap[T any, K comparable, V any](s []T, keyFn func(*T) K, valueFn func(*T) V) map[K]V {
	m := make(map[K]V)
	for _, item := range s {
		key := keyFn(&item)
		value := valueFn(&item)
		m[key] = value
	}
	return m
}

func RemoveDuplicates[T cmp.Ordered](elements []T) []T {
	slices.Sort(elements)
	return slices.Compact(elements)
}

// RemoveFromSlice 从slice中移除指定元素
func RemoveFromSlice[T comparable](source []T, element ...T) []T {
	result := make([]T, 0, len(source))
	for _, e := range source {
		if !slices.Contains(element, e) {
			result = append(result, e)
		}
	}
	return result
}
