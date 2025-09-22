package runtime

type AtomCell struct {
	Captured bool
	Value    *AtomValue
}

func NewAtomCell(captured bool, value *AtomValue) *AtomCell {
	return &AtomCell{
		Captured: captured,
		Value:    value,
	}
}
