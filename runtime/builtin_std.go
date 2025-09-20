package runtime

import (
	"fmt"

	"github.com/fatih/color"
)

var std_println = NewNativeFunc("println", Variadict, func(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	for i := range argc {
		fmt.Print(color.YellowString(frame.Stack.Pop().String()))
		if i < argc-1 {
			fmt.Print(" ")
		}
	}
	fmt.Println()
	frame.Stack.Push(interpreter.State.NullValue)
})

var std_print = NewNativeFunc("print", Variadict, func(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	for i := range argc {
		fmt.Print(color.GreenString(frame.Stack.Pop().String()))
		if i < argc-1 {
			fmt.Print(" ")
		}
	}
	frame.Stack.Push(interpreter.State.NullValue)
})

var EXPORT_STD = map[string]*AtomValue{
	"println": NewAtomValueNativeFunc(std_println),
	"print":   NewAtomValueNativeFunc(std_print),
}
