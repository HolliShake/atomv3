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

var EXPORT_OBJECT = map[string]*AtomValue{
	"freeze": NewAtomValueNativeFunc(obj_freeze),
}
