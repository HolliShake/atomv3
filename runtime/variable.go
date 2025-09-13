package runtime

type Variable struct {
	Name   string
	Global bool
	Const  bool
	Value  *AtomValue
}
