package utool_test

import (
	"testing"

	"github.com/NightmareZero/nzgoutil/utool"
)

func TestSliceRing(t *testing.T) {
	r := utool.NewSliceRing[int](2)

	// Test Add and Len
	r.Add(1)
	if r.Len() != 1 {
		t.Fatalf("Expected Len to be 1, got %d", r.Len())
	}

	r.Add(2)
	if r.Len() != 2 {
		t.Fatalf("Expected Len to be 2, got %d", r.Len())
	}

	// Test capacity doubling
	r.Add(3)
	if r.Len() != 3 {
		t.Fatalf("Expected Len to be 3, got %d", r.Len())
	}

	// Test Pop
	item, ok := r.Pop()
	if !ok || item != 1 {
		t.Fatalf("Expected Pop to return (1, true), got (%d, %t)", item, ok)
	}

	// Test Peek
	item, ok = r.Peek()
	if !ok || item != 2 {
		t.Fatalf("Expected Peek to return (2, true), got (%d, %t)", item, ok)
	}

	// Test empty Pop
	r.Pop()
	r.Pop()
	item, ok = r.Pop()
	if ok || item != 0 {
		t.Fatalf("Expected Pop to return (0, false) on empty ring, got (%d, %t)", item, ok)
	}

	// Test empty Peek
	item, ok = r.Peek()
	if ok || item != 0 {
		t.Fatalf("Expected Peek to return (0, false) on empty ring, got (%d, %t)", item, ok)
	}
}
