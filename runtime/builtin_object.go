package runtime

var obj_freeze = NewNativeFunc("freeze", 1, func(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	obj := frame.Stack.Pop()
	if CheckType(obj, AtomTypeObj) {
		obj.Value.(*AtomObject).Freeze = true
	} else if CheckType(obj, AtomTypeArray) {
		obj.Value.(*AtomArray).Freeze = true
	} else {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "cannot freeze non-object"),
		))
		return
	}
	frame.Stack.Push(obj)
})

var obj_keys = NewNativeFunc("keys", 1, func(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	obj := frame.Stack.Pop()
	if CheckType(obj, AtomTypeObj) {
		keys := []*AtomValue{}
		for key := range obj.Value.(*AtomObject).Elements {
			keys = append(keys, NewAtomValueStr(key))
		}
		frame.Stack.Push(NewAtomValueArray(keys))
	} else {
		frame.Stack.Push(NewAtomValueError(FormatError(frame, "cannot keys non-object")))
	}
})

var obj_values = NewNativeFunc("values", 1, func(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	obj := frame.Stack.Pop()
	if CheckType(obj, AtomTypeObj) {
		values := []*AtomValue{}
		for _, value := range obj.Value.(*AtomObject).Elements {
			values = append(values, value)
		}
		frame.Stack.Push(NewAtomValueArray(values))
	} else {
		frame.Stack.Push(NewAtomValueError(FormatError(frame, "cannot values non-object")))
	}
})

var EXPORT_OBJECT = map[string]*AtomValue{
	"freeze": NewAtomValueNativeFunc(obj_freeze),
	"keys":   NewAtomValueNativeFunc(obj_keys),
	"values": NewAtomValueNativeFunc(obj_values),
}
