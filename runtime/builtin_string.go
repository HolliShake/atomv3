package runtime

import (
	"fmt"
	"strings"
)

func string_len(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 1 {
		CleanupStack(frame, argc)
		message := fmt.Sprintf("string.len expected 1 argument, got %d", argc)
		frame.Stack.Push(NewAtomValueError(message))
		return
	}

	arg := frame.Stack.Pop()

	if !CheckType(arg, AtomTypeStr) {
		CleanupStack(frame, argc-1)
		message := fmt.Sprintf("string.len expected a string, got %s", GetTypeString(arg))
		frame.Stack.Push(NewAtomValueError(message))
		return
	}

	str := arg.Str
	frame.Stack.Push(NewAtomValueNum(float64(len(str))))
}

func string_toUpper(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 1 {
		CleanupStack(frame, argc)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "string.toUpper expected 1 argument"),
		))
		return
	}

	arg := frame.Stack.Pop()

	if !CheckType(arg, AtomTypeStr) {
		CleanupStack(frame, argc-1)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "string.toUpper expected a string"),
		))
		return
	}

	str := arg.Str
	frame.Stack.Push(NewAtomValueStr(strings.ToUpper(str)))
}

func string_toLower(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 1 {
		CleanupStack(frame, argc)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "string.toLower expected 1 argument"),
		))
		return
	}

	arg := frame.Stack.Pop()

	if !CheckType(arg, AtomTypeStr) {
		CleanupStack(frame, argc-1)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "string.toLower expected a string"),
		))
		return
	}

	str := arg.Str
	frame.Stack.Push(NewAtomValueStr(strings.ToLower(str)))
}

func string_contains(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 2 {
		CleanupStack(frame, argc)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "string.contains expected 2 arguments"),
		))
		return
	}

	arg0 := frame.Stack.GetOffset(argc, 0)
	arg1 := frame.Stack.GetOffset(argc, 1)

	if !CheckType(arg0, AtomTypeStr) {
		CleanupStack(frame, argc-2)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "string.contains expected a string"),
		))
		return
	}

	if !CheckType(arg1, AtomTypeStr) {
		CleanupStack(frame, argc-2)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "string.contains expected a string"),
		))
		return
	}

	str0 := arg0.Str
	str1 := arg1.Str

	if strings.Contains(str0, str1) {
		frame.Stack.Push(interpreter.State.TrueValue)
	} else {
		frame.Stack.Push(interpreter.State.FalseValue)
	}
}

func string_reverse(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 1 {
		CleanupStack(frame, argc)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "string.reverse expected 1 argument"),
		))
		return
	}

	arg := frame.Stack.Pop()

	if !CheckType(arg, AtomTypeStr) {
		CleanupStack(frame, argc-1)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "string.reverse expected a string"),
		))
		return
	}

	str := arg.Str
	// Use a more efficient string reversal with runes to handle UTF-8 correctly
	runes := []rune(str)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	frame.Stack.Push(NewAtomValueStr(string(runes)))
}

func string_runes(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 1 {
		CleanupStack(frame, argc)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "string.rune expected 1 argument"),
		))
		return
	}

	arg := frame.Stack.Pop()

	if !CheckType(arg, AtomTypeStr) {
		CleanupStack(frame, argc-1)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "string.rune expected a string"),
		))
		return
	}

	runes := []rune(arg.Str)
	elements := make([]*AtomValue, len(runes))
	for i, r := range runes {
		elements[i] = NewAtomValueInt(int(r))
	}
	frame.Stack.Push(NewAtomGenericValue(
		AtomTypeArray,
		NewAtomArray(elements),
	))
}

func string_bytes(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 1 {
		CleanupStack(frame, argc)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "string.bytes expected 1 argument"),
		))
		return
	}

	arg := frame.Stack.Pop()

	if !CheckType(arg, AtomTypeStr) {
		CleanupStack(frame, argc-1)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "string.bytes expected a string"),
		))
		return
	}

	bytes := []byte(arg.Str)
	elements := make([]*AtomValue, len(bytes))
	for i, b := range bytes {
		elements[i] = NewAtomValueInt(int(b))
	}
	frame.Stack.Push(NewAtomGenericValue(
		AtomTypeArray,
		NewAtomArray(elements),
	))
}

func string_format(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc < 2 {
		CleanupStack(frame, argc)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "string.format expected at least 2 arguments"),
		))
		return
	}

	format := frame.Stack.GetOffset(argc, 0)

	if !CheckType(format, AtomTypeStr) {
		CleanupStack(frame, argc)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "string.format expected a string"),
		))
		return
	}

	formatStr := format.Str
	count := strings.Count(formatStr, "{}")
	if count != argc-1 {
		CleanupStack(frame, argc)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, fmt.Sprintf("string.format expected %d arguments, got %d", count, argc-1)),
		))
		return
	}

	// Build the result string directly without creating an intermediate array
	result := formatStr
	for i := 1; i < argc; i++ {
		arg := frame.Stack.GetOffset(argc, i)
		result = strings.Replace(result, "{}", arg.String(), 1)
	}

	CleanupStack(frame, argc)
	frame.Stack.Push(NewAtomValueStr(result))
}

var EXPORT_STRING = map[string]*AtomValue{
	"len": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("string.len", 1, string_len),
	),
	"toUpper": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("string.toUpper", 1, string_toUpper),
	),
	"toLower": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("string.toLower", 1, string_toLower),
	),
	"contains": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("string.contains", 2, string_contains),
	),
	"reverse": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("string.reverse", 1, string_reverse),
	),
	"runes": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("string.rune", 1, string_runes),
	),
	"bytes": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("string.bytes", 1, string_bytes),
	),
	"format": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("string.format", Variadict, string_format),
	),
}
