package runtime

type AtomState struct {
	FunctionTable *AtomStack
	NullValue     *AtomValue
	FalseValue    *AtomValue
	TrueValue     *AtomValue
}

func NewAtomState() *AtomState {
	return &AtomState{
		FunctionTable: NewAtomStack(),
		NullValue:     NewAtomValueNull(),
		FalseValue:    NewAtomValueFalse(),
		TrueValue:     NewAtomValueTrue(),
	}
}

func (s *AtomState) SaveFunction(obj *AtomValue) int {
	s.FunctionTable.Push(obj)
	return s.FunctionTable.Len() - 1
}
