package runtime

type AtomCell struct {
	Value *AtomValue
}

func NewAtomCell(value *AtomValue) *AtomCell {
	return &AtomCell{
		Value: value,
	}
}
