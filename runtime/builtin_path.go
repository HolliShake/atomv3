package runtime

import (
	"os"
	"path/filepath"
)

func path_cwd(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 0 {
		CleanupStack(frame, argc)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "cwd expects 0 arguments"),
		))
		return
	}
	wd, err := os.Getwd()
	if err != nil {
		CleanupStack(frame, argc)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "cwd failed to get working directory"),
		))
		return
	}
	frame.Stack.Push(NewAtomValueStr(wd))
}

func path_join(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 2 {
		CleanupStack(frame, argc)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "join expects at least 2 arguments"),
		))
		return
	}
	parts := []string{}
	for i := range argc {
		parts = append(parts, frame.Stack.GetOffset(argc, i).String())
	}
	CleanupStack(frame, argc)
	frame.Stack.Push(NewAtomValueStr(filepath.Join(parts...)))
}

func path_isDir(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 1 {
		CleanupStack(frame, argc)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "isDir expects 1 argument"),
		))
		return
	}

	stat, err := os.Stat(frame.Stack.Pop().String())
	if err != nil {
		CleanupStack(frame, argc-1)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "isDir failed to get stat"),
		))
		return
	}

	if stat.IsDir() {
		frame.Stack.Push(interpreter.State.TrueValue)
	} else {
		frame.Stack.Push(interpreter.State.FalseValue)
	}
}

func path_isFile(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 1 {
		CleanupStack(frame, argc)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "isFile expects 1 argument"),
		))
		return
	}

	stat, err := os.Stat(frame.Stack.Pop().String())
	if err != nil {
		CleanupStack(frame, argc-1)
		frame.Stack.Push(interpreter.State.FalseValue)
		return
	}

	if !stat.IsDir() {
		frame.Stack.Push(interpreter.State.TrueValue)
	} else {
		frame.Stack.Push(interpreter.State.FalseValue)
	}
}

func path_exists(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 1 {
		CleanupStack(frame, argc)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "exists expects 1 argument"),
		))
		return
	}

	_, err := os.Stat(frame.Stack.Pop().String())
	if err != nil {
		frame.Stack.Push(interpreter.State.FalseValue)
	} else {
		frame.Stack.Push(interpreter.State.TrueValue)
	}
}

var EXPORT_PATH = map[string]*AtomValue{
	"cwd": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("cwd", 0, path_cwd),
	),
	"join": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("join", Variadict, path_join),
	),
	"isDir": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("isDir", 1, path_isDir),
	),
	"isFile": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("isFile", 1, path_isFile),
	),
	"exists": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("exists", 1, path_exists),
	),
}
