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
		CleanupStack(frame, argc)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "decompile expects 1 argument"),
		))
		return
	}
	if !CheckType(frame.Stack.Peek(), AtomTypeFunc) {
		CleanupStack(frame, argc)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "decompile expects a function"),
		))
		return
	}
	frame.Stack.Push(NewAtomValueStr(Decompile(frame.Stack.Pop().Obj.(*AtomCode))))
}

func std_println(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	writer := bufio.NewWriter(os.Stdout)
	for i := range argc {
		top := frame.Stack.GetOffset(argc, i)
		fmt.Fprint(writer, color.YellowString(top.String()))
		if i < argc-1 {
			fmt.Fprint(writer, " ")
		}
	}

	fmt.Fprintln(writer)
	writer.Flush()

	CleanupStack(frame, argc)
	frame.Stack.Push(interpreter.State.NullValue)
}

func std_print(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	writer := bufio.NewWriter(os.Stdout)
	for i := range argc {
		top := frame.Stack.GetOffset(argc, i)
		fmt.Fprint(writer, color.YellowString(top.String()))
		if i < argc-1 {
			fmt.Fprint(writer, " ")
		}
	}

	fmt.Fprint(writer)
	writer.Flush()

	CleanupStack(frame, argc)
	frame.Stack.Push(interpreter.State.NullValue)
}

func std_clear(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 0 {
		CleanupStack(frame, argc)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "clear expects 0 arguments"),
		))
		return
	}
	writer := bufio.NewWriter(os.Stdout)
	fmt.Fprint(writer, "\033[H\033[2J")
	writer.Flush()
	frame.Stack.Push(interpreter.State.NullValue)
}

func std_readLine(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 1 {
		CleanupStack(frame, argc)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "readLine expects 1 argument"),
		))
		return
	}
	fmt.Print(color.BlueString(frame.Stack.Pop().String()))
	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	if err != nil {
		CleanupStack(frame, argc-1)
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

	builder.WriteString(
		FormatError(frame, err.String()),
	)
	if frame.Caller != nil {
		builder.WriteString("\n")
	}

	for current := frame.Caller; current != nil; current = current.Caller {
		builder.WriteString(FormatError(current, current.Fn.Obj.(*AtomCode).Name))
		if current.Caller != nil {
			builder.WriteString("\n")
		}
	}
	fmt.Fprintln(os.Stderr, color.RedString(builder.String()))
	os.Exit(1)
}

func std_throw(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 1 {
		CleanupStack(frame, argc)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "throw expects 1 argument"),
		))
		return
	}
	std_throw_error(frame, frame.Stack.Pop())
	frame.Stack.Push(interpreter.State.NullValue)
}

func std_epoch(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 0 {
		CleanupStack(frame, argc)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "epoch expects 0 arguments"),
		))
		return
	}
	frame.Stack.Push(NewAtomValueNum(float64(time.Now().Unix())))
}

func std_sleep(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 1 {
		CleanupStack(frame, argc)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "sleep expects 1 argument"),
		))
		return
	}
	val := frame.Stack.Pop()
	if !CheckType(val, AtomTypeInt) && !CheckType(val, AtomTypeNum) {
		CleanupStack(frame, argc-1)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "sleep expects a number"),
		))
		return
	}
	// arg should be in milliseconds
	num := CoerceToNum(val)
	time.Sleep(time.Duration(num) * time.Millisecond)
	frame.Stack.Push(interpreter.State.NullValue)
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
	"clear": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("clear", 0, std_clear),
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
	"sleep": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("sleep", 1, std_sleep),
	),
}
