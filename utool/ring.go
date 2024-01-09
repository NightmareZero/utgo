package utool

type SliceRing[T any] struct {
	data     []T
	capacity int
	start    int
	end      int
}

func NewSliceRing[T any](capacity int) *SliceRing[T] {
	return &SliceRing[T]{
		data:     make([]T, capacity),
		capacity: capacity,
	}
}

func (r *SliceRing[T]) Len() int {
	if r.end >= r.start {
		return r.end - r.start
	}
	return r.capacity - r.start + r.end
}

func (r *SliceRing[T]) Add(item T) {
	// buffer is full, double the capacity
	if r.end == (r.start-1+r.capacity)%r.capacity {
		// double the capacity
		newData := make([]T, r.capacity*2)
		oldLen := r.Len()
		if r.start < r.end {
			copy(newData, r.data[r.start:r.end])
		} else {
			copy(newData, r.data[r.start:r.capacity])
			copy(newData[r.capacity-r.start:], r.data[:r.end])
		}
		r.start = 0
		r.end = oldLen
		r.capacity *= 2
		r.data = newData
	}
	// not full, add item
	r.data[r.end] = item
	r.end = (r.end + 1) % r.capacity
}

func (r *SliceRing[T]) Pop() (item T, ok bool) {
	if r.start == r.end {
		return item, false // buffer is empty
	}
	item = r.data[r.start]
	r.start = (r.start + 1) % r.capacity
	return item, true
}

func (r *SliceRing[T]) Peek() (item T, ok bool) {
	if r.start == r.end {
		return item, false // buffer is empty
	}
	return r.data[r.start], true
}
