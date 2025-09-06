package runtime

type AtomStack struct {
	stack []*AtomValue
}

func NewAtomStack() *AtomStack {
	return &AtomStack{
		stack: make([]*AtomValue, 0),
	}
}

func (s *AtomStack) Push(obj *AtomValue) {
	s.stack = append(s.stack, obj)
}

func (s *AtomStack) Pop() *AtomValue {
	top := s.stack[len(s.stack)-1]
	s.stack = s.stack[:len(s.stack)-1]
	return top
}

func (s *AtomStack) Peek() *AtomValue {
	return s.stack[len(s.stack)-1]
}

func (s *AtomStack) Len() int {
	return len(s.stack)
}
