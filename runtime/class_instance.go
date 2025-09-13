package runtime

type AtomClassInstance struct {
	Prototype *AtomValue // AtomClass
	Property  *AtomValue // AtomObject
}

func NewAtomClassInstance(prototype, property *AtomValue) *AtomClassInstance {
	return &AtomClassInstance{
		Prototype: prototype,
		Property:  property,
	}
}
