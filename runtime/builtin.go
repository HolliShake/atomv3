package runtime

type AtomNativeFunc struct {
	Name     string
	Paramc   int
	Callable func(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int)
}

type AtomNativeModuleCallback func() *AtomValue

const Variadict = -1

func NewNativeFunc(name string, paramc int, callable func(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int)) *AtomNativeFunc {
	return &AtomNativeFunc{
		Name:     name,
		Paramc:   paramc,
		Callable: callable,
	}
}

func DefineModule(interpreter *AtomInterpreter, name string, callback AtomNativeModuleCallback) {
	interpreter.ModuleTable[name] = callback
}
