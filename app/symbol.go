package main

type AtomSymbol struct {
	name     string
	global   bool
	constant bool
}

func NewAtomSymbol(name string, global bool, constant bool) *AtomSymbol {
	return &AtomSymbol{
		name:     name,
		global:   global,
		constant: constant,
	}
}
