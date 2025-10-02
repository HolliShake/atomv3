package runtime

func obj_freeze(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 1 {
		CleanupStack(frame, argc)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "freeze expects 1 argument"),
		))
		return
	}

	obj := frame.Stack.Pop()

	if CheckType(obj, AtomTypeObj) {
		obj.Obj.(*AtomObject).Freeze = true
	} else if CheckType(obj, AtomTypeArray) {
		obj.Obj.(*AtomArray).Freeze = true
	} else {
		CleanupStack(frame, argc-1)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "cannot freeze non-object"),
		))
		return
	}
	frame.Stack.Push(obj)
}

func obj_keys(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 1 {
		CleanupStack(frame, argc)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "keys expects 1 argument"),
		))
		return
	}

	obj := frame.Stack.Pop()

	if CheckType(obj, AtomTypeObj) {
		keys := []*AtomValue{}
		for key := range obj.Obj.(*AtomObject).Elements {
			keys = append(keys, NewAtomValueStr(key))
		}
		frame.Stack.Push(NewAtomGenericValue(
			AtomTypeArray,
			NewAtomArray(keys),
		))
	} else {
		CleanupStack(frame, argc-1)
		frame.Stack.Push(NewAtomValueError(FormatError(frame, "cannot keys non-object")))
	}
}

func obj_values(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	if argc != 1 {
		CleanupStack(frame, argc)
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "values expects 1 argument"),
		))
		return
	}

	obj := frame.Stack.Pop()

	if CheckType(obj, AtomTypeObj) {
		values := []*AtomValue{}
		for _, value := range obj.Obj.(*AtomObject).Elements {
			values = append(values, value)
		}
		frame.Stack.Push(NewAtomGenericValue(
			AtomTypeArray,
			NewAtomArray(values),
		))
	} else {
		CleanupStack(frame, argc-1)
		frame.Stack.Push(NewAtomValueError(FormatError(frame, "cannot values non-object")))
	}
}

var EXPORT_OBJECT = map[string]*AtomValue{
	"freeze": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("freeze", 1, obj_freeze),
	),
	"keys": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("keys", 1, obj_keys),
	),
	"values": NewAtomGenericValue(
		AtomTypeNativeFunc,
		NewNativeFunc("values", 1, obj_values),
	),
}
