package util

import "errors"

type Stack[T any] []T

func NewStack[T any]() *Stack[T] {
	return &Stack[T]{}
}

func (s *Stack[T]) Push(value T) {
	*s = append(*s, value)
}

func (s *Stack[T]) Pop() (T, error) {
	if s.Len() == 0 {
		return *(new(T)), errors.New("Stack is empty")
	}

	index := len(*s) - 1
	element := (*s)[index]
	*s = (*s)[:index]
	return element, nil
}

func (s *Stack[T]) Top() (T, error) {
	if s.Len() == 0 {
		return *(new(T)), errors.New("Stack is empty")
	}

	return (*s)[len(*s)-1], nil
}

func (s *Stack[T]) Len() int {
	return len(*s)
}

func (s *Stack[T]) IsEmpty() bool {
	return len(*s) == 0
}
