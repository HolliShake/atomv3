package runtime

type AtomCell struct {
	Value *AtomValue
}

func NewAtomCell(value *AtomValue) *AtomCell {
	return &AtomCell{Value: value}
}

func (c *AtomCell) Get() *AtomValue {
	return c.Value
}

func (c *AtomCell) Set(value *AtomValue) {
	c.Value = value
}
