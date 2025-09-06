package runtime

type AtomCode struct {
	File   string
	Name   string
	Argc   int
	LCount int          // Local count
	CCount int          // Capture count
	Env0   []*AtomValue // Local environment
	Env1   []*AtomValue // Capture environment
	Code   []OpCode
}

func NewAtomCode(file, name string, argc int) *AtomCode {
	code := new(AtomCode)
	code.File = file
	code.Name = name
	code.Argc = argc
	code.LCount = 0
	code.CCount = 0
	code.Env0 = make([]*AtomValue, 0)
	code.Env1 = make([]*AtomValue, 0)
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
	c.Env0 = make([]*AtomValue, c.LCount)
}

func (c *AtomCode) AllocateCaptures() {
	c.Env1 = make([]*AtomValue, c.CCount)
}
