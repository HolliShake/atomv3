package runtime

import (
	"os"
	"os/exec"
)

func os_exit(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 1 {
		CleanupStack(frame, argc)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "exit expects 1 argument"),
		))
		return
	}

	code := CoerceToLong(frame.Stack.Pop())
	os.Exit(int(code))
	frame.Stack.Push(interpreter.State.NullValue)
}

func os_exec(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 1 {
		CleanupStack(frame, argc)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "exec expects 1 argument"),
		))
		return
	}

	if !CheckType(frame.Stack.Peek(), AtomTypeStr) {
		CleanupStack(frame, argc)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "exec expects a string"),
		))
		return
	}

	cmd := frame.Stack.Pop().Str
	err := exec.Command(cmd).Run()

	if err != nil {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, err.Error()),
		))
		return
	}
	frame.Stack.Push(interpreter.State.NullValue)
}

var EXPORT_OS = map[string]*AtomValue{
	"failure": NewAtomValueInt(1),
	"success": NewAtomValueInt(0),
	"exit": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("exit", 1, os_exit),
	),
	"exec": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("exec", 1, os_exec),
	),
}
