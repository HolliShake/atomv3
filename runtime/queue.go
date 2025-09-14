package runtime

// Queue is a generic optimized queue implementation using a circular buffer
type Queue[T any] struct {
	items    []T
	head     int
	tail     int
	size     int
	capacity int
}

// NewQueue creates a new queue with initial capacity of 8
func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{
		items:    make([]T, 8),
		head:     0,
		tail:     0,
		size:     0,
		capacity: 8,
	}
}

// Enqueue adds a value to the end of the queue
func (q *Queue[T]) Enqueue(value T) {
	// Resize if needed
	if q.size == q.capacity {
		q.resize()
	}

	q.items[q.tail] = value
	q.tail = (q.tail + 1) % q.capacity
	q.size++
}

// Dequeue removes and returns the value at the front of the queue
func (q *Queue[T]) Dequeue() (T, bool) {
	var zero T
	if q.size == 0 {
		return zero, false
	}

	value := q.items[q.head]
	q.head = (q.head + 1) % q.capacity
	q.size--

	return value, true
}

// Peek returns the value at the front without removing it
func (q *Queue[T]) Peek() (T, bool) {
	var zero T
	if q.size == 0 {
		return zero, false
	}
	return q.items[q.head], true
}

// Len returns the number of elements in the queue
func (q *Queue[T]) Len() int {
	return q.size
}

// IsEmpty returns true if the queue is empty
func (q *Queue[T]) IsEmpty() bool {
	return q.size == 0
}

// resize doubles the capacity of the queue
func (q *Queue[T]) resize() {
	newCapacity := q.capacity * 2
	newItems := make([]T, newCapacity)

	// Copy elements maintaining order
	for i := 0; i < q.size; i++ {
		index := (q.head + i) % q.capacity
		newItems[i] = q.items[index]
	}

	q.items = newItems
	q.head = 0
	q.tail = q.size
	q.capacity = newCapacity
}

// Clear removes all elements from the queue
func (q *Queue[T]) Clear() {
	q.head = 0
	q.tail = 0
	q.size = 0
}

// AtomQueue is a type alias for backward compatibility
type AtomQueue = Queue[*AtomValue]

// NewAtomQueue creates a new queue specifically for AtomValue pointers
func NewAtomQueue() *AtomQueue {
	return NewQueue[*AtomValue]()
}
