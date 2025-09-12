package runtime

type AtomClassInstance struct {
	Prototype *AtomValue // AtomClass
	Property  *AtomValue // AtomObject
}
