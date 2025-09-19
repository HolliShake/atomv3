package runtime

import (
	"fmt"

	"github.com/fatih/color"
)

var freeze = NewNativeFunc("freeze", 1, func(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	obj := frame.Stack.Pop()
	if CheckType(obj, AtomTypeObj) {
		obj.Value.(*AtomObject).Freeze = true
	} else if CheckType(obj, AtomTypeArray) {
		obj.Value.(*AtomArray).Freeze = true
	} else {
		frame.Stack.Push(NewAtomValueError("cannot freeze non-object"))
		return
	}
	frame.Stack.Push(obj)
})

var println = NewNativeFunc("println", Variadict, func(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	for i := range argc {
		fmt.Print(color.YellowString(frame.Stack.Pop().String()))
		if i < argc-1 {
			fmt.Print(" ")
		}
	}
	fmt.Println()
	frame.Stack.Push(interpreter.State.NullValue)
})

var print = NewNativeFunc("print", Variadict, func(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	for i := range argc {
		fmt.Print(color.GreenString(frame.Stack.Pop().String()))
		if i < argc-1 {
			fmt.Print(" ")
		}
	}
	frame.Stack.Push(interpreter.State.NullValue)
})

var EXPORT_STD = map[string]*AtomValue{
	"freeze":  NewAtomValueNativeFunc(freeze),
	"println": NewAtomValueNativeFunc(println),
	"print":   NewAtomValueNativeFunc(print),
}
