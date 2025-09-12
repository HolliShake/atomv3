package runtime

import "fmt"

type AtomStack struct {
	Stack []*AtomValue
}

func NewAtomStack() *AtomStack {
	return &AtomStack{
		Stack: []*AtomValue{},
	}
}

func (s *AtomStack) Get(index int) *AtomValue {
	return s.Stack[index]
}

func (s *AtomStack) Push(obj *AtomValue) {
	s.Stack = append(s.Stack, obj)
}

func (s *AtomStack) Pop() *AtomValue {
	top := s.Stack[len(s.Stack)-1]
	s.Stack = s.Stack[:len(s.Stack)-1]
	return top
}

func (s *AtomStack) Peek() *AtomValue {
	return s.Stack[len(s.Stack)-1]
}

func (s *AtomStack) Len() int {
	return len(s.Stack)
}

func (s *AtomStack) Dump() {
	for _, obj := range s.Stack {
		fmt.Println(obj.String())
	}
}
