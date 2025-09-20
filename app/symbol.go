package main

import runtime "dev.runtime"

type AtomSymbol struct {
	name     string
	global   bool
	constant bool
	index    int
	cell     *runtime.AtomCell
}

func NewAtomSymbol(name string, global bool, constant bool, index int, cell *runtime.AtomCell) *AtomSymbol {
	return &AtomSymbol{
		name:     name,
		global:   global,
		constant: constant,
		index:    index,
		cell:     cell,
	}
}
