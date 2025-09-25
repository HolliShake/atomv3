package runtime

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
)

func std_decompile(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
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
}

func std_println(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	writer := bufio.NewWriter(os.Stdout)
	for i := range argc {
		fmt.Fprint(writer, color.YellowString(frame.Stack.Pop().String()))
		if i < argc-1 {
			fmt.Fprint(writer, " ")
		}
	}
	fmt.Fprintln(writer)
	writer.Flush()
	frame.Stack.Push(interpreter.State.NullValue)
}

func std_print(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	writer := bufio.NewWriter(os.Stdout)
	for i := range argc {
		fmt.Fprint(writer, color.YellowString(frame.Stack.Pop().String()))
		if i < argc-1 {
			fmt.Fprint(writer, " ")
		}
	}
	fmt.Fprint(writer)
	writer.Flush()
	frame.Stack.Push(interpreter.State.NullValue)
}

func std_readLine(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
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
}

func std_throw_error(frame *AtomCallFrame, err *AtomValue) {
	// Stack trace
	builder := strings.Builder{}
	builder.WriteByte('\n')

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

func std_throw(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
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
}

func std_epoch(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	frame.Stack.Push(NewAtomValueNum(float64(time.Now().Unix())))
}

var EXPORT_STD = map[string]*AtomValue{
	"decompile": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("decompile", 1, std_decompile),
	),
	"println": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("println", Variadict, std_println),
	),
	"print": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("print", Variadict, std_print),
	),
	"readLine": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("readLine", 1, std_readLine),
	),
	"throw": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("throw", 1, std_throw),
	),
	"epoch": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("epoch", 0, std_epoch),
	),
}
