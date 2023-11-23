package data_struct

// Queue[T] is a generic queue
type Queue[T any] struct {
	data []T
}

// NewQueue creates a new generic queue
func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{}
}

// Push adds an element to the end of the queue
func (q *Queue[T]) Push(item T) {
	q.data = append(q.data, item)
}

// Pop pops an element from the front of the queue and returns it
func (q *Queue[T]) Pop() T {
	if len(q.data) == 0 {
		var zero T
		return zero
	}
	item := q.data[0]
	q.data = q.data[1:]
	return item
}

// Unshift inserts an element at the front of the queue
func (q *Queue[T]) Unshift(item T) {
	newData := make([]T, len(q.data)+1)
	newData[0] = item
	copy(newData[1:], q.data)
	q.data = newData
}

func (q *Queue[T]) RemoveAll(match func(T) bool) {
	var newData []T
	for _, item := range q.data {
		if !match(item) {
			newData = append(newData, item)
		}
	}
	q.data = newData
}
