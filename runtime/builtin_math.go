package runtime

import (
	"math"
	"math/rand"
)

func math_rand(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 1 {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "rand expects 1 argument"),
		))
		return
	}

	arg := frame.Stack.Pop()

	if !IsNumberType(arg) {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "rand expects integer"),
		))
		return
	}

	val := CoerceToLong(arg)
	frame.Stack.Push(NewAtomValueInt(rand.Intn(int(val))))
}

func math_abs(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 1 {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "abs expects 1 argument"),
		))
		return
	}

	arg := frame.Stack.Pop()

	if !IsNumberType(arg) {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "abs expects number"),
		))
		return
	}

	val := CoerceToNum(arg)
	frame.Stack.Push(NewAtomValueNum(math.Abs(val)))
}

func math_floor(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 1 {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "floor expects 1 argument"),
		))
		return
	}

	arg := frame.Stack.Pop()

	if !IsNumberType(arg) {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "floor expects number"),
		))
		return
	}

	val := CoerceToNum(arg)
	frame.Stack.Push(NewAtomValueNum(math.Floor(val)))
}

func math_ceil(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 1 {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "ceil expects 1 argument"),
		))
		return
	}

	arg := frame.Stack.Pop()

	if !IsNumberType(arg) {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "ceil expects number"),
		))
		return
	}

	val := CoerceToNum(arg)
	frame.Stack.Push(NewAtomValueNum(math.Ceil(val)))
}

func math_round(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 1 {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "round expects 1 argument"),
		))
		return
	}

	arg := frame.Stack.Pop()

	if !IsNumberType(arg) {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "round expects number"),
		))
		return
	}

	val := CoerceToNum(arg)
	frame.Stack.Push(NewAtomValueNum(math.Round(val)))
}

func math_pow(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	cleanup := func() {
		for range argc {
			frame.Stack.Pop()
		}
	}

	if argc != 2 {
		cleanup()
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "pow expects 2 arguments"),
		))
		return
	}

	arg1 := frame.Stack.GetOffset(argc, 0)
	arg2 := frame.Stack.GetOffset(argc, 1)

	if !IsNumberType(arg1) || !IsNumberType(arg2) {
		cleanup()
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "pow expects number"),
		))
		return
	}

	val1 := CoerceToNum(arg1)
	val2 := CoerceToNum(arg2)

	cleanup()
	frame.Stack.Push(NewAtomValueNum(math.Pow(val1, val2)))
}

func math_sqrt(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 1 {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "sqrt expects 1 argument"),
		))
		return
	}

	arg := frame.Stack.Pop()

	if !IsNumberType(arg) {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "sqrt expects number"),
		))
		return
	}

	val := CoerceToNum(arg)
	frame.Stack.Push(NewAtomValueNum(math.Sqrt(val)))
}

func math_log(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 1 {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "log expects 1 argument"),
		))
		return
	}

	arg := frame.Stack.Pop()

	if !IsNumberType(arg) {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "log expects number"),
		))
		return
	}

	val := CoerceToNum(arg)
	frame.Stack.Push(NewAtomValueNum(math.Log(val)))
}

func math_cos(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 1 {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "cos expects 1 argument"),
		))
		return
	}

	arg := frame.Stack.Pop()

	if !IsNumberType(arg) {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "cos expects number"),
		))
		return
	}

	val := CoerceToNum(arg)
	frame.Stack.Push(NewAtomValueNum(math.Cos(val)))
}

func math_sin(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 1 {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "sin expects 1 argument"),
		))
		return
	}

	arg := frame.Stack.Pop()
	if !IsNumberType(arg) {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "sin expects number"),
		))
		return
	}

	val := CoerceToNum(arg)
	frame.Stack.Push(NewAtomValueNum(math.Sin(val)))
}

var EXPORT_MATH = map[string]*AtomValue{
	"rand": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("rand", 1, math_rand),
	),
	"abs": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("abs", 1, math_abs),
	),
	"floor": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("floor", 1, math_floor),
	),
	"ceil": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("ceil", 1, math_ceil),
	),
	"round": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("round", 1, math_round),
	),
	"pow": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("pow", 2, math_pow),
	),
	"sqrt": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("sqrt", 1, math_sqrt),
	),
	"log": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("log", 1, math_log),
	),
	"cos": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("cos", 1, math_cos),
	),
	"sin": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("sin", 1, math_sin),
	),
}
