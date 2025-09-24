package runtime

type AtomNativeFunc struct {
	Name     string
	Paramc   int
	Callable func(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int)
}

const Variadict = -1

func NewNativeFunc(name string, paramc int, callable func(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int)) *AtomNativeFunc {
	return &AtomNativeFunc{
		Name:     name,
		Paramc:   paramc,
		Callable: callable,
	}
}

func DefineModule(interpreter *AtomInterpreter, name string, values map[string]*AtomValue) {
	values["__name__"] = NewAtomValueStr(name)
	interpreter.ModuleTable[name] = NewAtomValueObject(values)
}
