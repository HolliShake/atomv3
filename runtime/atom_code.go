package runtime

type AtomCode struct {
	File   string
	Name   string
	Argc   int
	LCount int         // Local count
	CCount int         // Capture count
	Env0   []*AtomCell // Local environment
	Env1   []*AtomCell // Capture environment
	Code   []OpCode
}

func NewAtomCode(file, name string, argc int) *AtomCode {
	code := new(AtomCode)
	code.File = file
	code.Name = name
	code.Argc = argc
	code.LCount = 0
	code.CCount = 0
	code.Env0 = make([]*AtomCell, 0)
	code.Env1 = make([]*AtomCell, 0)
	code.Code = make([]OpCode, 0)
	return code
}

func (c *AtomCode) IncrementLocal() int {
	current := c.LCount
	c.LCount++
	return current
}

func (c *AtomCode) IncrementCapture() int {
	current := c.CCount
	c.CCount++
	return current
}

func (c *AtomCode) AllocateLocals() {
	c.Env0 = make([]*AtomCell, c.LCount)
	for i := 0; i < c.LCount; i++ {
		c.Env0[i] = NewAtomCell(nil)
	}
}

func (c *AtomCode) AllocateCaptures() {
	c.Env1 = make([]*AtomCell, c.CCount)
	for i := 0; i < c.CCount; i++ {
		c.Env1[i] = NewAtomCell(nil)
	}
}
