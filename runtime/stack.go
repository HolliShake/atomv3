package runtime

import "fmt"

type AtomStack struct {
	Index int
	Stack []*AtomValue
}

func NewAtomStack() *AtomStack {
	return &AtomStack{
		Index: -1,
		Stack: make([]*AtomValue, 1000),
	}
}

func (s *AtomStack) SetIndex(index int) {
	if index >= 0 && index < len(s.Stack) {
		s.Index = index
	}
}

func (s *AtomStack) Get(index int) *AtomValue {
	return s.Stack[index]
}

func (s *AtomStack) Push(obj *AtomValue) {
	s.Index++
	s.Stack[s.Index] = obj
}

func (s *AtomStack) Pop() *AtomValue {
	if s.Index < 0 {
		return nil
	}
	top := s.Stack[s.Index]
	s.Stack[s.Index] = nil
	s.Index--
	return top
}

func (s *AtomStack) Peek() *AtomValue {
	if s.Index >= 0 {
		return s.Stack[s.Index]
	}
	return nil
}

func (s *AtomStack) Len() int {
	return s.Index + 1
}

func (s *AtomStack) Dump() {
	for i := 0; i <= s.Index; i++ {
		obj := s.Stack[i]
		marker := ""
		if i == s.Index {
			marker = " <- current"
		}
		fmt.Printf("[%d] %s%s\n", i, obj.String(), marker)
	}
}
