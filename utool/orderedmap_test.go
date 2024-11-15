package utila

import (
	"encoding/json"
	"testing"
)

func TestJsonObj_MarshalJSON(t *testing.T) {
	obj := NewJsonObj()
	obj.Put("name", "Alice")
	obj.Put("age", 30)
	obj.Put("city", "Wonderland")

	data, err := json.Marshal(obj)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	expected := `{"name":"Alice","age":30,"city":"Wonderland"}`
	if string(data) != expected {
		t.Errorf("Expected %s, got %s", expected, string(data))
	}
}

func TestJsonObj_UnmarshalJSON(t *testing.T) {
	data := `{"name":"Alice","age":30,"city":"Wonderland"}`
	var obj JsonObj

	err := json.Unmarshal([]byte(data), &obj)
	if err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	if name, ok := obj.Get("name"); !ok || name != "Alice" {
		t.Errorf("Expected name to be Alice, got %v", name)
	}
	if age, ok := obj.Get("age"); !ok || age != float64(30) {
		t.Errorf("Expected age to be 30, got %v", age)
	}
	if city, ok := obj.Get("city"); !ok || city != "Wonderland" {
		t.Errorf("Expected city to be Wonderland, got %v", city)
	}
	if obj.Keys()[0] != "name" || obj.Keys()[1] != "age" || obj.Keys()[2] != "city" {
		t.Errorf("Expected keys to be name, age, city, but %+v", obj.Keys())
	}
}

func TestOrderedMap_Put(t *testing.T) {
	m := NewOrderedMap[string, any]()
	m.Put("key1", "value1")
	m.Put("key2", "value2")

	if val, ok := m.Get("key1"); !ok || val != "value1" {
		t.Errorf("Expected key1 to be value1, got %v", val)
	}
	if val, ok := m.Get("key2"); !ok || val != "value2" {
		t.Errorf("Expected key2 to be value2, got %v", val)
	}
}

func TestOrderedMap_Get(t *testing.T) {
	m := NewOrderedMap[string, any]()
	m.Put("key1", "value1")

	if val, ok := m.Get("key1"); !ok || val != "value1" {
		t.Errorf("Expected key1 to be value1, got %v", val)
	}

	if _, ok := m.Get("key2"); ok {
		t.Errorf("Expected key2 to be absent")
	}
}

func TestOrderedMap_Remove(t *testing.T) {
	m := NewOrderedMap[string, any]()
	m.Put("key1", "value1")
	m.Put("key2", "value2")
	m.Remove("key1")
	m.Put("key3", "value3")
	m.Remove("key3")
	m.Remove("key2")

	if m.Len() != 0 {
		t.Errorf("Expected map to be empty")
	}

	if _, ok := m.Get("key1"); ok {
		t.Errorf("Expected key1 to be removed")
	}
}

func TestOrderedMap_Keys(t *testing.T) {
	m := NewOrderedMap[string, any]()
	m.Put("key1", "value1")
	m.Put("key2", "value2")

	keys := m.Keys()
	expected := []string{"key1", "key2"}

	for i, key := range keys {
		if key != expected[i] {
			t.Errorf("Expected key %s, got %s", expected[i], key)
		}
	}
}

func TestOrderedMap_Range(t *testing.T) {
	m := NewOrderedMap[string, any]()
	m.Put("key1", "value1")
	m.Put("key2", "value2")

	count := 0
	m.Range(func(key string, val any) bool {
		count++
		return true
	})

	if count != 2 {
		t.Errorf("Expected range to iterate over 2 elements, got %d", count)
	}
}

func TestOrderedMap_Len(t *testing.T) {
	m := NewOrderedMap[string, any]()
	m.Put("key1", "value1")
	m.Put("key2", "value2")

	if length := m.Len(); length != 2 {
		t.Errorf("Expected length to be 2, got %d", length)
	}
}

func TestOrderedMap_Clear(t *testing.T) {
	m := NewOrderedMap[string, any]()
	m.Put("key1", "value1")
	m.Clear()

	if length := m.Len(); length != 0 {
		t.Errorf("Expected length to be 0 after clear, got %d", length)
	}
}

func TestOrderedMap_Clone(t *testing.T) {
	m := NewOrderedMap[string, any]()
	m.Put("key1", "value1")
	clone := m.Clone()

	if val, ok := clone.Get("key1"); !ok || val != "value1" {
		t.Errorf("Expected key1 to be value1 in clone, got %v", val)
	}
}

func TestOrderedMap_Merge(t *testing.T) {
	m1 := NewOrderedMap[string, any]()
	m1.Put("key1", "value1")

	m2 := NewOrderedMap[string, any]()
	m2.Put("key2", "value2")

	m1.Merge(m2)

	if val, ok := m1.Get("key2"); !ok || val != "value2" {
		t.Errorf("Expected key2 to be value2 after merge, got %v", val)
	}
}

func TestOrderedMap_ToList(t *testing.T) {
	m := NewOrderedMap[string, any]()
	m.Put("key1", "value1")
	m.Put("key2", "value2")

	list := m.ToList()
	expected := []any{"value1", "value2"}

	for i, val := range list {
		if val != expected[i] {
			t.Errorf("Expected value %v, got %v", expected[i], val)
		}
	}
}

func TestOrderedMap_ToMap(t *testing.T) {
	m := NewOrderedMap[string, any]()
	m.Put("key1", "value1")
	m.Put("key2", "value2")

	mm := m.ToMap()
	expected := map[string]any{"key1": "value1", "key2": "value2"}

	for key, val := range expected {
		if mm[key] != val {
			t.Errorf("Expected key %s to be %v, got %v", key, val, mm[key])
		}
	}
}

func TestMapEntry_InsertAfter(t *testing.T) {
	entry1 := &mapEntry[int, string]{key: 1, val: "one"}
	entry2 := &mapEntry[int, string]{key: 2, val: "two"}
	entry3 := &mapEntry[int, string]{key: 3, val: "three"}

	// Insert entry2 after entry1
	entry1.append(entry2)
	if entry2._before != entry1 {
		t.Errorf("Expected entry2.before to be entry1, got %v", entry2._before)
	}
	if entry1._next != entry2 {
		t.Errorf("Expected entry1.next to be entry2, got %v", entry1._next)
	}

	// Insert entry3 after entry2
	entry2.append(entry3)
	if entry3._before != entry2 {
		t.Errorf("Expected entry3.before to be entry2, got %v", entry3._before)
	}
	if entry2._next != entry3 {
		t.Errorf("Expected entry2.next to be entry3, got %v", entry2._next)
	}

	// Verify the order: entry1 -> entry2 -> entry3
	if entry1._next != entry2 || entry2._next != entry3 {
		t.Errorf("Expected order to be entry1 -> entry2 -> entry3")
	}
	if entry3._before != entry2 || entry2._before != entry1 {
		t.Errorf("Expected order to be entry1 <- entry2 <- entry3")
	}
}

func TestMapEntry_InsertBefore(t *testing.T) {
	entry1 := &mapEntry[int, string]{key: 1, val: "one"}
	entry2 := &mapEntry[int, string]{key: 2, val: "two"}
	entry3 := &mapEntry[int, string]{key: 3, val: "three"}

	// Insert entry2 before entry1
	entry1.prepend(entry2)
	if entry1._before != entry2 {
		t.Errorf("Expected entry1.before to be entry2, got %v", entry1._before)
	}
	if entry2._next != entry1 {
		t.Errorf("Expected entry2.next to be entry1, got %v", entry2._next)
	}

	// Insert entry3 before entry2
	entry2.prepend(entry3)
	if entry2._before != entry3 {
		t.Errorf("Expected entry2.before to be entry3, got %v", entry2._before)
	}
	if entry3._next != entry2 {
		t.Errorf("Expected entry3.next to be entry2, got %v", entry3._next)
	}

	// Verify the order: entry3 -> entry2 -> entry1
	if entry3._next != entry2 || entry2._next != entry1 {
		t.Errorf("Expected order to be entry3 -> entry2 -> entry1")
	}
	if entry1._before != entry2 || entry2._before != entry3 {
		t.Errorf("Expected order to be entry3 <- entry2 <- entry1")
	}
}

func TestMapEntry_NextEntry(t *testing.T) {
	entry1 := &mapEntry[string, int]{key: "key1", val: 1}
	entry2 := &mapEntry[string, int]{key: "key2", val: 2}

	entry1._next = entry2

	if next := entry1.next(); next != entry2 {
		t.Errorf("Expected next entry to be entry2, got %v", next)
	}
}

func TestMapEntry_BeforeEntry(t *testing.T) {
	entry1 := &mapEntry[string, int]{key: "key1", val: 1}
	entry2 := &mapEntry[string, int]{key: "key2", val: 2}

	entry2._before = entry1

	if before := entry2.before(); before != entry1 {
		t.Errorf("Expected before entry to be entry1, got %v", before)
	}
}

func TestMapEntry_Remove(t *testing.T) {
	entry1 := &mapEntry[string, int]{key: "key1", val: 1}
	entry2 := &mapEntry[string, int]{key: "key2", val: 2}
	entry3 := &mapEntry[string, int]{key: "key3", val: 3}

	entry1._next = entry2
	entry2._before = entry1
	entry2._next = entry3
	entry3._before = entry2

	entry2.remove()

	if entry1._next != entry3 {
		t.Errorf("Expected entry1.next to be entry3, got %v", entry1._next)
	}
	if entry3._before != entry1 {
		t.Errorf("Expected entry3.before to be entry1, got %v", entry3._before)
	}
}
