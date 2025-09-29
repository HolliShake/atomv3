package runtime

import (
	"fmt"
	"os"
)

func file_read(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 2 {
		CleanupStack(frame, argc)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "read expects 2 arguments"),
		))
		return
	}

	arg0 := frame.Stack.GetOffset(argc, 0) // path
	arg1 := frame.Stack.GetOffset(argc, 1) // mode

	if !CheckType(arg0, AtomTypeStr) {
		CleanupStack(frame, argc)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "read expects a string path"),
		))
		return
	}

	if !CheckType(arg1, AtomTypeStr) {
		CleanupStack(frame, argc)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "read expects a string mode"),
		))
		return
	}

	path := arg0.Value.(string)
	mode := arg1.Value.(string)

	CleanupStack(frame, argc)

	content, err := os.ReadFile(path)
	if err != nil {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, err.Error()),
		))
		return
	}

	switch mode {
	case "byte[]":
		{
			bytes := []byte(content)
			elements := make([]*AtomValue, len(bytes))
			for i, b := range bytes {
				elements[i] = NewAtomValueInt(int(b))
			}
			frame.Stack.Push(NewAtomGenericValue(
				AtomTypeArray,
				NewAtomArray(elements),
			))
		}
	case "int[]":
		{
			// convert to array
			runeContent := []rune(string(content))
			elements := make([]*AtomValue, len(runeContent))
			for i, r := range runeContent {
				elements[i] = NewAtomValueInt(int(r))
			}
			frame.Stack.Push(NewAtomGenericValue(
				AtomTypeArray,
				NewAtomArray(elements),
			))
		}
	case "string":
		frame.Stack.Push(NewAtomValueStr(string(content)))
	default:
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, fmt.Sprintf("supported types are: byte[], int[], string, got: %s", mode)),
		))
	}
}

func file_write(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	cleanup := func() {
		for range argc {
			frame.Stack.Pop()
		}
	}

	if argc != 3 {
		cleanup()
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "write expects 2 arguments"),
		))
		return
	}

	arg0 := frame.Stack.GetOffset(argc, 0) // path
	arg1 := frame.Stack.GetOffset(argc, 1) // content
	arg2 := frame.Stack.GetOffset(argc, 2) // mode

	if !CheckType(arg0, AtomTypeStr) {
		cleanup()
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "write expects a string path"),
		))
		return
	}

	if !CheckType(arg1, AtomTypeStr) {
		cleanup()
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "write expects a string content"),
		))
		return
	}

	if !CheckType(arg2, AtomTypeStr) {
		cleanup()
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "write expects a string mode"),
		))
		return
	}

	path := arg0.Value.(string)
	content := arg1.Value.(string)
	mode := arg2.Value.(string)

	cleanup()

	switch mode {
	case "w":
		{
			bytes := []byte(content)
			err := os.WriteFile(path, bytes, 0644)
			if err != nil {
				frame.Stack.Push(NewAtomValueError(
					FormatError(frame, err.Error()),
				))
				return
			}

			frame.Stack.Push(NewAtomValueInt(len(bytes)))
		}
	case "a":
		{
			file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				frame.Stack.Push(NewAtomValueError(
					FormatError(frame, err.Error()),
				))
				return
			}
			defer file.Close()

			bytes := []byte(content)
			n, err := file.Write(bytes)
			if err != nil {
				frame.Stack.Push(NewAtomValueError(
					FormatError(frame, err.Error()),
				))
				return
			}

			frame.Stack.Push(NewAtomValueInt(n))
		}
	default:
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "write expects a string mode"),
		))
		return
	}
}

var EXPORT_FILE = map[string]*AtomValue{
	"read": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("read", 2, file_read),
	),
	"write": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("write", 3, file_write),
	),
}
