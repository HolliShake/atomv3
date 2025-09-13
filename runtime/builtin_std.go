package runtime

import "fmt"

var freeze = NewNativeFunc("freeze", 1, func(interpreter *AtomInterpreter, argc int) {
	obj := interpreter.popp()
	if CheckType(obj, AtomTypeObj) {
		obj.Value.(*AtomObject).Freeze = true
	} else if CheckType(obj, AtomTypeArray) {
		obj.Value.(*AtomArray).Freeze = true
	} else {
		interpreter.pushVal(NewAtomValueError("cannot freeze non-object"))
		return
	}
	interpreter.pushRef(interpreter.State.NullValue)
})

var println = NewNativeFunc("println", Variadict, func(interpreter *AtomInterpreter, argc int) {
	for i := range argc {
		fmt.Print(interpreter.popp().String())
		if i < argc-1 {
			fmt.Print(" ")
		}
	}
	fmt.Println()
	interpreter.pushRef(interpreter.State.NullValue)
})

var print = NewNativeFunc("print", Variadict, func(interpreter *AtomInterpreter, argc int) {
	for i := range argc {
		fmt.Print(interpreter.popp().String())
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
