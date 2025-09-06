package runtime

type AtomCode struct {
	File   string
	Name   string
	Argc   int
	LCount int         // Local count
	Env0   []*AtomCell // Local environment
	Code   []OpCode
}

func NewAtomCode(file, name string, argc int) *AtomCode {
	code := new(AtomCode)
	code.File = file
	code.Name = name
	code.Argc = argc
	code.LCount = 0
	code.Env0 = make([]*AtomCell, 0)
	code.Code = make([]OpCode, 0)
	return code
}

func (c *AtomCode) IncrementLocal() int {
	current := c.LCount
	c.LCount++
	return current
}

func (c *AtomCode) AllocateLocals() {
	c.Env0 = make([]*AtomCell, c.LCount)
	for i := 0; i < c.LCount; i++ {
		c.Env0[i] = NewAtomCell(nil)
	}
}
