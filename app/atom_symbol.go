package main

type AtomSymbol struct {
	Name    string
	Offset  int
	Global  bool
	Const   bool
	Capture bool
}

func NewAtomSymbol(name string, offset int, global bool) *AtomSymbol {
	return &AtomSymbol{Name: name, Offset: offset, Global: global, Const: false}
}

func NewConstAtomSymbol(name string, offset int, global bool) *AtomSymbol {
	return &AtomSymbol{Name: name, Offset: offset, Global: global, Const: true}
}

func NewCaptureAtomSymbol(name string, offset int, global bool, isConst bool, isConstCapture bool) *AtomSymbol {
	return &AtomSymbol{Name: name, Offset: offset, Global: global, Const: isConst, Capture: isConstCapture}
}
