package utool

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type JsonObj struct {
	JsonTObj[any]
}

type JsonTObj[T any] struct {
	OrderedMap[string, T]
}

// MarshalJSON implements the json.Marshaler interface.
func (m JsonTObj[T]) MarshalJSON() ([]byte, error) {
	// Create a slice to hold the ordered entries
	orderedEntries := make([]mapEntry[string, json.RawMessage], 0, m.Len())

	// Iterate over the map and marshal each value
	m.Range(func(key string, val T) bool {
		rawVal, err := json.Marshal(val)
		if err != nil {
			return false
		}
		orderedEntries = append(orderedEntries, mapEntry[string, json.RawMessage]{key: key, val: rawVal})
		return true
	})

	// Create a buffer to write the JSON object
	buffer := bytes.NewBufferString("{")
	for i, entry := range orderedEntries {
		if i > 0 {
			buffer.WriteString(",")
		}
		buffer.WriteString(fmt.Sprintf(`"%s":%s`, entry.key, entry.val))
	}
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (m *JsonTObj[T]) UnmarshalJSON(data []byte) error {
	// Create a temporary slice to hold the ordered entries
	var tempSlice []struct {
		Key string          `json:"key"`
		Val json.RawMessage `json:"val"`
	}

	// Create a decoder to decode the JSON data
	decoder := json.NewDecoder(bytes.NewReader(data))

	// Read the opening brace
	if _, err := decoder.Token(); err != nil {
		return err
	}

	// Iterate over the JSON object
	for decoder.More() {
		// Read the key
		tok, err := decoder.Token()
		if err != nil {
			return err
		}
		key := tok.(string)

		// Read the value
		var rawVal json.RawMessage
		if err := decoder.Decode(&rawVal); err != nil {
			return err
		}

		// Append the key-value pair to the slice
		tempSlice = append(tempSlice, struct {
			Key string          `json:"key"`
			Val json.RawMessage `json:"val"`
		}{Key: key, Val: rawVal})
	}

	// Read the closing brace
	if _, err := decoder.Token(); err != nil {
		return err
	}

	// Clear the current map
	m.Clear()

	// Iterate over the temporary slice and add each entry to the JsonObj
	for _, entry := range tempSlice {
		var val T
		if err := json.Unmarshal(entry.Val, &val); err != nil {
			return err
		}
		m.Put(entry.Key, val)
	}

	return nil
}

type OrderedMap[K comparable, V any] struct {
	entrys map[K]*mapEntry[K, V]
	head   *mapEntry[K, V]
	last   *mapEntry[K, V]
}

func (m *OrderedMap[K, V]) Put(key K, val V) {
	if m.entrys == nil {
		m.entrys = make(map[K]*mapEntry[K, V])
	}
	if entry, ok := m.entrys[key]; ok {
		entry.val = val
	} else {
		entry := &mapEntry[K, V]{key: key, val: val}
		if m.head == nil {
			m.head = entry
		}
		if m.last == nil {
			m.last = entry
		} else {
			m.last.append(entry)
			m.last = entry
		}
		m.entrys[key] = entry
	}
}

func (m *OrderedMap[K, V]) Get(key K) (v V, ok bool) {
	if entry, ok := m.entrys[key]; ok {
		return entry.val, true
	}
	return v, false
}

func (m *OrderedMap[K, V]) Remove(key K) {
	if entry, ok := m.entrys[key]; ok {
		entry.remove()
		delete(m.entrys, key)
		if entry == m.head {
			m.head = entry._next
		}
		if entry == m.last {
			m.last = entry._before
		}
	}
}

func (m *OrderedMap[K, V]) Keys() []K {
	keys := make([]K, 0, len(m.entrys))
	for entry := m.head; entry != nil; entry = entry._next {
		keys = append(keys, entry.key)
	}
	return keys
}

func (m *OrderedMap[K, V]) Range(f func(key K, val V) bool) {
	for entry := m.head; entry != nil; entry = entry._next {
		if !f(entry.key, entry.val) {
			break
		}
	}
}

func (m *OrderedMap[K, V]) Len() int {
	return len(m.entrys)
}

func (m *OrderedMap[K, V]) Clear() {
	m.entrys = make(map[K]*mapEntry[K, V])
	m.head = nil
	m.last = nil
}

func (m *OrderedMap[K, V]) Clone() *OrderedMap[K, V] {
	clone := &OrderedMap[K, V]{}
	m.Range(func(key K, val V) bool {
		clone.Put(key, val)
		return true
	})
	return clone
}

func (m *OrderedMap[K, V]) Merge(om *OrderedMap[K, V]) {
	om.Range(func(key K, val V) bool {
		m.Put(key, val)
		return true
	})
}

func (m *OrderedMap[K, V]) ToList() []V {
	list := make([]V, 0, len(m.entrys))
	m.Range(func(key K, val V) bool {
		list = append(list, val)
		return true
	})
	return list
}

func (m *OrderedMap[K, V]) ToMap() map[K]V {
	mm := make(map[K]V, len(m.entrys))
	m.Range(func(key K, val V) bool {
		mm[key] = val
		return true
	})
	return mm
}

type mapEntry[K comparable, V any] struct {
	key K
	val V

	_next   *mapEntry[K, V]
	_before *mapEntry[K, V]
}

func (e *mapEntry[K, V]) next() *mapEntry[K, V] {
	return e._next
}

// 在e后面插入entry
func (e *mapEntry[K, V]) append(entry *mapEntry[K, V]) {
	next := e._next
	e._next = entry
	entry._before = e
	entry._next = next
	if next != nil {
		next._before = entry
	}
}

func (e *mapEntry[K, V]) before() *mapEntry[K, V] {
	return e._before
}

// 在e前面插入entry
func (e *mapEntry[K, V]) prepend(entry *mapEntry[K, V]) {
	before := e._before
	e._before = entry
	entry._next = e
	entry._before = before
	if before != nil {
		before._next = entry
	}
}

func (e *mapEntry[K, V]) remove() {
	if e._before != nil {
		e._before._next = e._next
	}
	if e._next != nil {
		e._next._before = e._before
	}
}
