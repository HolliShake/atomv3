package runtime

type Code struct {
	File       string
	Name       string
	OpCodes    []OpCode
	Lines      []int
	LocalCount int
	Locals     []*AtomValue
}

func NewCode(file, name string) *Code {
	code := new(Code)
	code.File = file
	code.Name = name
	code.OpCodes = make([]OpCode, 0)
	code.Lines = make([]int, 0)
	code.LocalCount = 0
	code.Locals = make([]*AtomValue, 0)
	return code
}

func (c *Code) IncrementLocal() int {
	current := c.LocalCount
	c.LocalCount++
	return current
}

func (c *Code) AllocateLocals() {
	c.Locals = make([]*AtomValue, c.LocalCount)
}
