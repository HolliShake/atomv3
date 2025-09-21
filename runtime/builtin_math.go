package runtime

import (
	"math"
	"math/rand"
)

var math_rand = NewNativeFunc("rand", 1, func(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
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
})

var math_abs = NewNativeFunc("abs", 1, func(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
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
})

var math_floor = NewNativeFunc("floor", 1, func(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
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
})

var math_ceil = NewNativeFunc("ceil", 1, func(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
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
})

var math_round = NewNativeFunc("round", 1, func(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
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
})

var math_pow = NewNativeFunc("pow", 2, func(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 2 {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "pow expects 2 arguments"),
		))
		return
	}

	arg := frame.Stack.Pop()
	arg2 := frame.Stack.Pop()

	if !IsNumberType(arg) || !IsNumberType(arg2) {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "pow expects number"),
		))
		return
	}

	val := CoerceToNum(arg)
	val2 := CoerceToNum(arg2)
	frame.Stack.Push(NewAtomValueNum(math.Pow(val, val2)))
})

var math_sqrt = NewNativeFunc("sqrt", 1, func(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
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
})

var math_log = NewNativeFunc("log", 1, func(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
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
})

var EXPORT_MATH = map[string]*AtomValue{
	"rand":  NewAtomValueNativeFunc(math_rand),
	"abs":   NewAtomValueNativeFunc(math_abs),
	"floor": NewAtomValueNativeFunc(math_floor),
	"ceil":  NewAtomValueNativeFunc(math_ceil),
	"round": NewAtomValueNativeFunc(math_round),
	"pow":   NewAtomValueNativeFunc(math_pow),
	"sqrt":  NewAtomValueNativeFunc(math_sqrt),
	"log":   NewAtomValueNativeFunc(math_log),
}
