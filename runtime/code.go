package runtime

type AtomCode struct {
	File string
	Name string
	Argc int
	Code []OpCode // Instructions
}

func NewAtomCode(file, name string, argc int) *AtomCode {
	return &AtomCode{
		File: file,
		Name: name,
		Argc: argc,
		Code: []OpCode{},
	}
}

func (c *AtomCode) HashValue() int {
	hash := 0

	// Hash the file name
	for _, b := range []byte(c.File) {
		hash = hash*31 + int(b)
	}

	// Hash the function name
	for _, b := range []byte(c.Name) {
		hash = hash*31 + int(b)
	}

	// Hash the argument count
	hash = hash*31 + c.Argc

	// Hash the code length
	hash = hash*31 + len(c.Code)

	// Hash the code
	for _, opcode := range c.Code {
		hash = hash*31 + int(opcode)
	}

	return hash
}
