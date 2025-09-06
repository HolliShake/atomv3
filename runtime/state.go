package runtime

type AtomState struct {
	FunctionTable *AtomStack
}

func NewAtomState() *AtomState {
	return &AtomState{
		FunctionTable: NewAtomStack(),
	}
}

func (s *AtomState) SaveFunction(obj *AtomValue) {
	s.FunctionTable.Push(obj)
}
