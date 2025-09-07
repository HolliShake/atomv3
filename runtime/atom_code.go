package runtime

type AtomCode struct {
	File string
	Name string
	Argc int
	Env0 []*AtomCell // Local environment
	Code []OpCode
}

func NewAtomCode(file, name string, argc int) *AtomCode {
	code := new(AtomCode)
	code.File = file
	code.Name = name
	code.Argc = argc
	code.Env0 = make([]*AtomCell, 0)
	code.Code = make([]OpCode, 0)
	return code
}

func (c *AtomCode) IncrementLocal() int {
	count := len(c.Env0)
	c.Env0 = append(c.Env0, NewAtomCell(nil))
	return count
}
