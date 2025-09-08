package runtime

import "fmt"

var freeze = NewNativeFunc("freeze", 1, func(interpreter *AtomInterpreter, argc int) {
	obj := interpreter.pop()
	if !CheckType(obj, AtomTypeObj) {
		interpreter.pushVal(NewAtomValueError("cannot freeze non-object"))
		return
	}
	obj.Value.(*AtomObject).Freeze = true
	interpreter.pushRef(interpreter.State.NullValue)
})

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
	"freeze":  NewAtomValueNativeFunc(freeze),
	"println": NewAtomValueNativeFunc(println),
	"print":   NewAtomValueNativeFunc(print),
}
