package main

type AtomSymbol struct {
	Name   string
	Offset int
	Global bool
	Const  bool
}

func NewAtomSymbol(name string, offset int, global bool) *AtomSymbol {
	return &AtomSymbol{Name: name, Offset: offset, Global: global, Const: false}
}

func NewConstAtomSymbol(name string, offset int, global bool) *AtomSymbol {
	return &AtomSymbol{Name: name, Offset: offset, Global: global}
}
