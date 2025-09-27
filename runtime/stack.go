package runtime

import "fmt"

type AtomStack struct {
	Stack []*AtomValue
}

func NewAtomStack() *AtomStack {
	return &AtomStack{
		Stack: make([]*AtomValue, 0, 8), // Pre-allocate with capacity
	}
}

func (s *AtomStack) Get(index int) *AtomValue {
	return s.Stack[index]
}

func (s *AtomStack) GetOffset(offset int, index int) *AtomValue {
	return s.Stack[(len(s.Stack)-offset)+index]
}

func (s *AtomStack) SetOffset(offset int, index int, value *AtomValue) {
	s.Stack[(len(s.Stack)-offset)+index] = value
}

func (s *AtomStack) Copy(Stack *AtomStack, size int) {
	for i := range size {
		s.Stack = append(s.Stack, Stack.Stack[len(Stack.Stack)-size+i])
	}
}

func (s *AtomStack) Clear() {
	s.Stack = make([]*AtomValue, 0, 8)
}

func (s *AtomStack) Push(obj *AtomValue) {
	s.Stack = append(s.Stack, obj)
}

func (s *AtomStack) Pop() *AtomValue {
	if len(s.Stack) == 0 {
		panic("Pop on empty stack")
	}
	top := s.Stack[len(s.Stack)-1]
	s.Stack = s.Stack[:len(s.Stack)-1]
	return top
}

func (s *AtomStack) Peek() *AtomValue {
	if len(s.Stack) == 0 {
		panic("Peek on empty stack")
	}
	return s.Stack[len(s.Stack)-1]
}

func (s *AtomStack) Len() int {
	return len(s.Stack)
}

func (s *AtomStack) IsEmpty() bool {
	return len(s.Stack) == 0
}

func (s *AtomStack) Dump() {
	for i := 0; i < len(s.Stack); i++ {
		obj := s.Stack[i]
		marker := ""
		if i == len(s.Stack)-1 {
			marker = " <- top"
		}
		fmt.Printf("[%d] %s%s\n", i, obj.String(), marker)
	}
}
