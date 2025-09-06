package runtime

type AtomCode struct {
	File       string
	Name       string
	OpCodes    []OpCode
	Lines      []int
	LocalCount int
	Locals     []*AtomValue
}

func NewAtomCode(file, name string) *AtomCode {
	code := new(AtomCode)
	code.File = file
	code.Name = name
	code.OpCodes = make([]OpCode, 0)
	code.Lines = make([]int, 0)
	code.LocalCount = 0
	code.Locals = make([]*AtomValue, 0)
	return code
}

func (c *AtomCode) IncrementLocal() int {
	current := c.LocalCount
	c.LocalCount++
	return current
}

func (c *AtomCode) AllocateLocals() {
	c.Locals = make([]*AtomValue, c.LocalCount)
}
