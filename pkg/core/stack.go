package core

// Push adds an element to the top of the stack
func (s *Stack[T]) Push(v T) {
	*s = append(*s, v)
}

// Pop removes and returns the top element of the stack
func (s *Stack[T]) Pop() (T, bool) {
	if len(*s) == 0 {
		var zero T // Return zero value if stack is empty
		return zero, false
	}
	top := (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return top, true
}

// Peek returns the top element without removing it
func (s *Stack[T]) Peek() (T, bool) {
	if len(*s) == 0 {
		var zero T
		return zero, false
	}
	return (*s)[len(*s)-1], true
}

// IsEmpty checks if the stack is empty
func (s *Stack[T]) IsEmpty() bool {
	return len(*s) == 0
}

// Size returns the number of elements in the stack
func (s *Stack[T]) Size() int {
	return len(*s)
}
