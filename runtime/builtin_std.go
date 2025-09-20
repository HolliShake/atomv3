package runtime

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

var std_decompile = NewNativeFunc("decompile", 1, func(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 1 {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "decompile expects 1 argument"),
		))
		return
	}
	if !CheckType(frame.Stack.Peek(), AtomTypeFunc) {
		frame.Stack.Pop()
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "decompile expects a function"),
		))
		return
	}
	frame.Stack.Push(NewAtomValueStr(Decompile(frame.Stack.Pop().Value.(*AtomCode))))
})

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
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "readLine expects 1 argument"),
		))
		return
	}
	fmt.Print(color.BlueString(frame.Stack.Pop().String()))
	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	if err != nil {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, err.Error()),
		))
		return
	}
	frame.Stack.Push(NewAtomValueStr(strings.TrimSpace(text)))
})

func std_throw_error(frame *AtomCallFrame, err *AtomValue) {
	// Stack trace
	builder := strings.Builder{}

	builder.WriteString(err.String())

	if frame.Caller != nil {
		builder.WriteString("\n")
	}

	for current := frame.Caller; current != nil; current = current.Caller {
		builder.WriteString(FormatError(current, current.Fn.Value.(*AtomCode).Name))
		if current.Caller != nil {
			builder.WriteString("\n")
		}
	}
	fmt.Fprintln(os.Stderr, builder.String())
	os.Exit(1)
}

var std_throw = NewNativeFunc("throw", 1, func(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 1 {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "throw expects 1 argument"),
		))
		return
	}
	if !CheckType(frame.Stack.Peek(), AtomTypeErr) {
		frame.Stack.Pop()
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "throw expects an error"),
		))
		return
	}
	std_throw_error(frame.Caller, frame.Stack.Pop())
	frame.Stack.Push(interpreter.State.NullValue)
})

var EXPORT_STD = map[string]*AtomValue{
	"decompile": NewAtomValueNativeFunc(std_decompile),
	"println":   NewAtomValueNativeFunc(std_println),
	"print":     NewAtomValueNativeFunc(std_print),
	"readLine":  NewAtomValueNativeFunc(std_readLine),
	"throw":     NewAtomValueNativeFunc(std_throw),
}
