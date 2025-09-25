package runtime

import "os"

func os_exit(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 1 {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "exit expects 1 argument"),
		))
		return
	}

	code := CoerceToLong(frame.Stack.Pop())
	os.Exit(int(code))
	frame.Stack.Push(interpreter.State.NullValue)
}

var EXPORT_OS = map[string]*AtomValue{
	"failure": NewAtomValueInt(1),
	"success": NewAtomValueInt(0),
	"exit": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("exit", 1, os_exit),
	),
}
