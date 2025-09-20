package runtime

import (
	"bufio"
	"fmt"
	"os"
	"strings"

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

var std_readLine = NewNativeFunc("readLine", 1, func(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 1 {
		frame.Stack.Push(NewAtomValueError("readLine expects 1 argument"))
		return
	}
	fmt.Print(color.BlueString(frame.Stack.Pop().String()))
	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	if err != nil {
		frame.Stack.Push(NewAtomValueError(err.Error()))
		return
	}
	frame.Stack.Push(NewAtomValueStr(strings.TrimSpace(text)))
})

var EXPORT_STD = map[string]*AtomValue{
	"println":  NewAtomValueNativeFunc(std_println),
	"print":    NewAtomValueNativeFunc(std_print),
	"readLine": NewAtomValueNativeFunc(std_readLine),
}
