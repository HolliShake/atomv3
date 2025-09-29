package runtime

import "fmt"

func string_len(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 1 {
		CleanupStack(frame, argc)
		message := fmt.Sprintf("string.len expected 1 argument, got %d", argc)
		frame.Stack.Push(NewAtomValueError(message))
		return
	}

	arg := frame.Stack.Pop()

	if !CheckType(arg, AtomTypeStr) {
		CleanupStack(frame, argc)
		message := fmt.Sprintf("string.len expected a string, got %s", GetTypeString(arg))
		frame.Stack.Push(NewAtomValueError(message))
		return
	}

	str := arg.Value.(string)
	frame.Stack.Push(NewAtomValueNum(float64(len(str))))
}

var EXPORT_STRING = map[string]*AtomValue{
	"len": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("string.len", 1, string_len),
	),
}
