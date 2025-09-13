package runtime

type AtomClassInstance struct {
	Prototype *AtomValue // AtomClass
	Property  *AtomValue // AtomObject
}

type AtomMethod struct {
	This *AtomValue
	Fn   *AtomValue
}

func NewAtomClassInstance(prototype, property *AtomValue) *AtomClassInstance {
	return &AtomClassInstance{
		Prototype: prototype,
		Property:  property,
	}
}

func NewAtomMethod(this *AtomValue, fn *AtomValue) *AtomMethod {
	return &AtomMethod{
		This: this,
		Fn:   fn,
	}
}
