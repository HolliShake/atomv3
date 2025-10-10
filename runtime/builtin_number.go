package runtime

import "strconv"

func number_parseInt(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 1 {
		CleanupStack(frame, argc)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "Error: parseInt expects 1 argument"),
		))
		return
	}

	value := frame.Stack.Pop()

	if !CheckType(value, AtomTypeStr) {
		CleanupStack(frame, argc-1)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "Error: parseInt expects a string"),
		))
		return
	}

	str := value.Str

	intValue, err := strconv.Atoi(str)
	if err != nil {
		CleanupStack(frame, argc-1)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "Error: parseInt expects a valid integer"),
		))
		return
	}
	frame.Stack.Push(NewAtomValueInt(intValue))
}

func number_parseFloat(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 1 {
		CleanupStack(frame, argc)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "Error: parseFloat expects 1 argument"),
		))
		return
	}

	value := frame.Stack.Pop()

	if !CheckType(value, AtomTypeStr) {
		CleanupStack(frame, argc-1)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "Error: parseFloat expects a string"),
		))
		return
	}

	str := value.Str

	floatValue, err := strconv.ParseFloat(str, 64)
	if err != nil {
		CleanupStack(frame, argc-1)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "Error: parseFloat expects a valid float"),
		))
		return
	}
	frame.Stack.Push(NewAtomValueNum(floatValue))
}

func number_toString(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 1 {
		CleanupStack(frame, argc)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "Error: toString expects 1 argument"),
		))
		return
	}

	value := frame.Stack.Pop()

	if !CheckType(value, AtomTypeInt) && !CheckType(value, AtomTypeNum) {
		CleanupStack(frame, argc-1)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "Error: toString expects a number"),
		))
		return
	}

	num := CoerceToNum(value)
	frame.Stack.Push(NewAtomValueStr(strconv.FormatFloat(num, 'g', -1, 64)))
}

// cast to int
func number_int(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 1 {
		CleanupStack(frame, argc)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "Error: int expects 1 argument"),
		))
		return
	}

	value := frame.Stack.Pop()

	if !IsNumberType(value) {
		CleanupStack(frame, argc-1)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "Error: int expects a number"),
		))
		return
	}

	num := CoerceToInt(value)
	frame.Stack.Push(NewAtomValueInt(int(num)))
}

// cast to num
func number_num(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 1 {
		CleanupStack(frame, argc)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "Error: num expects 1 argument"),
		))
		return
	}

	value := frame.Stack.Pop()

	if !IsNumberType(value) {
		CleanupStack(frame, argc-1)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "Error: num expects a number"),
		))
		return
	}

	num := CoerceToNum(value)
	frame.Stack.Push(NewAtomValueNum(float64(num)))
}

func number_bigInt(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 1 {
		CleanupStack(frame, argc)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "Error: bigInt expects 1 argument"),
		))
		return
	}

	value := frame.Stack.Pop()

	if !IsNumberType(value) {
		CleanupStack(frame, argc-1)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "Error: bigInt expects a number"),
		))
		return
	}

	num := CoerceToBigInt(value)
	frame.Stack.Push(NewAtomValueBigInt(num))
}

var EXPORT_NUMBER = map[string]*AtomValue{
	"parseInt": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("number.parseInt", 1, number_parseInt),
	),
	"parseFloat": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("number.parseFloat", 1, number_parseFloat),
	),
	"toString": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("number.toString", 1, number_toString),
	),
	"int": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("number.int", 1, number_int),
	),
	"num": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("number.num", 1, number_num),
	),
	"bigInt": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("number.bigInt", 1, number_bigInt),
	),
}
