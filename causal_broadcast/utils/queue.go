package utils

import "sync"

// Queue is a FIFO queue(circular buffer)
type Queue[T any] struct {
	*sync.Mutex
	next, idx int
	stop      bool
	q         []T
	buf       T
}

// Make returns a new queue with backing slice of length `n`
func Make[T any](n int) *Queue[T] {
	return &Queue{
		Mutex: &sync.Mutex{},
		q:     make([]T, n),
	}
}

func (eq Queue[T any]) Len() int {
	return len(eq.q)
}

func (eq Queue[T any]) Cap() int {
	return cap(eq.q)
}

// Grow increases the size of the queue by `n` elements
func (eq *Queue[T any]) Grow(n int) {
	return
}

// Push puts an item at the next vacant slot of the queue.
func (eq *Queue[T any]) Push(e T) bool {
	eq.Lock()
	defer eq.Unlock()

	if eq.stop {
		// if queue is closed but idx and next are not
		// pointing to the same element, open queue
		if eq.next > 0 && eq.idx != eq.next {
			eq.stop = false
		} else {
			return false
		}
	}

	// queue is open for reading
	if eq.idx > 0 {
		if eq.idx == eq.next {
			// if where we are reading the next item is same
			// as where we put the next item, close queue and return.
			eq.stop = true
			return false
		} else if eq.idx == len(eq.q)-1 {
			// when appending to the end of the queue,
			// append and check if there's vacancy at the
			// start of the queue.
			eq.q[eq.idx] = e
			if eq.next > 0 {
				// continue appending if there's vacancy.
				eq.reset()
			} else {
				// stop if there's no vacancy at beginning of queue.
				eq.stop = true
			}
			return true
		}
	}

	eq.q[eq.idx] = e
	eq.idx = (eq.idx + 1) % len(eq.q)

	// if after appending eq.idx becomes equal to eq.next
	// halt the next append.
	if eq.idx > 0 && eq.idx == eq.next {
		eq.stop = true
	}

	return true
}

func (eq *Queue[T any]) reset() {
	eq.stop = false
	eq.idx = (eq.idx + 1) % len(eq.q)

	// funny story:
	//	i was trying to lock this function and defer the unlock
	//	after execution, fucking naive right?. The methods calling this
	//	method are already locking the struct before accessing it so
	//	locking it only creates a deadlock or livelock?. One of them.
	//	pretty sure its a deadlock.
}

// Pop returns the next item in the queue.
func (eq *Queue[T any]) Pop() (T, bool) {
	eq.Lock()
	defer eq.Unlock()

	// set the next reading point in the queue
	move := func() {
		eq.buf = eq.q[eq.next]
		eq.next = (eq.next + 1) % len(eq.q)
	}

	if eq.stop {
		if eq.idx == eq.next {
			move()
			eq.stop = false
			return eq.buf, true
		} else if eq.idx == len(eq.q)-1 {
			// insertion at the end, vacancy at the start
			eq.reset()
		} else if eq.next != eq.idx {
			// if idx and next are not the same, open queue for pushing.
			eq.reset()
		}
	}

	move()
	return eq.buf, true
}

// Retry returns the most recently read element from the queue.
func (eq *Queue[T any]) Retry() T {
	return eq.buf
}
