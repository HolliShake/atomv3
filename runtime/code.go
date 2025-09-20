package runtime

type AtomCode struct {
	File        string
	Name        string
	Async       bool
	Argc        int
	Line        []AtomDebugLine
	Code        []OpCode // Instructions
	Locals      []*AtomCell
	CapturedEnv []*AtomCell
}

func NewAtomCode(file, name string, async bool, argc int) *AtomCode {
	return &AtomCode{
		File:        file,
		Name:        name,
		Async:       async,
		Argc:        argc,
		Line:        []AtomDebugLine{},
		Code:        []OpCode{},
		Locals:      []*AtomCell{},
		CapturedEnv: []*AtomCell{},
	}
}

func (c *AtomCode) HashValue() int {
	hash := uint32(0)

	// Hash the file name
	for _, b := range []byte(c.File) {
		hash = hash*31 + uint32(b)
	}

	// Hash the function name
	for _, b := range []byte(c.Name) {
		hash = hash*31 + uint32(b)
	}

	// Hash the async flag
	if c.Async {
		hash = hash*31 + 1
	} else {
		hash = hash*31 + 0
	}

	// Hash the argument count
	hash = hash*31 + uint32(c.Argc)

	// Hash the line numbers
	for _, line := range c.Line {
		hash = hash*31 + uint32(line.Line)
		hash = hash*31 + uint32(line.Address)
	}

	// Hash the code
	for _, opcode := range c.Code {
		hash = hash*31 + uint32(opcode)
	}

	return int(hash)
}
