package runtime

type AtomCode struct {
	File string
	Name string
	Argc int
	Env0 []*AtomCell // Local environment
	Code []OpCode
}

func NewAtomCode(file, name string, argc int) *AtomCode {
	return &AtomCode{
		File: file,
		Name: name,
		Argc: argc,
		Env0: []*AtomCell{},
		Code: []OpCode{},
	}
}

func (c *AtomCode) IncrementLocal() int {
	count := len(c.Env0)
	c.Env0 = append(c.Env0, NewAtomCell(nil))
	return count
}
