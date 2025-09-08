package runtime

import "fmt"

var println = NewNativeFunc("println", Variadict, func(interpreter *AtomInterpreter, argc int) {
	for i := range argc {
		fmt.Print(interpreter.pop().String())
		if i < argc-1 {
			fmt.Print(" ")
		}
	}
	fmt.Println()
	interpreter.pushRef(interpreter.State.NullValue)
})

var print = NewNativeFunc("print", Variadict, func(interpreter *AtomInterpreter, argc int) {
	for i := range argc {
		fmt.Print(interpreter.pop().String())
		if i < argc-1 {
			fmt.Print(" ")
		}
	}
	interpreter.pushRef(interpreter.State.NullValue)
})

var EXPORT_STD = map[string]*AtomValue{
	"println": NewAtomValueNativeFunc(println),
	"print":   NewAtomValueNativeFunc(print),
}
