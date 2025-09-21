package runtime

import "math/rand"

var math_rand = NewNativeFunc("rand", 1, func(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 1 {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "rand expects 1 argument"),
		))
		return
	}

	arg := frame.Stack.Pop()

	if !CheckType(arg, AtomTypeInt) {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "rand expects integer"),
		))
		return
	}

	val := CoerceToLong(arg)
	frame.Stack.Push(NewAtomValueInt(rand.Intn(int(val))))
})

var EXPORT_MATH = map[string]*AtomValue{
	"rand": NewAtomValueNativeFunc(math_rand),
}
