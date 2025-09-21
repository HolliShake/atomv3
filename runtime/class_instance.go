package runtime

type AtomClassInstance struct {
	Prototype *AtomValue // AtomClass
	Property  *AtomValue // AtomObject
}

type AtomMethod struct {
	This *AtomValue
	Fn   *AtomValue
}

type AtomNativeMethod struct {
	Name     string
	Paramc   int
	This     *AtomValue
	Callable func(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int)
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

func NewAtomNativeMethod(name string, paramc int, this *AtomValue, callable func(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int)) *AtomNativeMethod {
	return &AtomNativeMethod{
		Name:     name,
		Paramc:   paramc,
		This:     this,
		Callable: callable,
	}
}
