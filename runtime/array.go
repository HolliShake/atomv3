package runtime

import (
	"fmt"
	"slices"
)

type AtomArray struct {
	Freeze   bool
	Elements []*AtomValue
}

func NewAtomArray(elements []*AtomValue) *AtomArray {
	return &AtomArray{Elements: elements, Freeze: false}
}

func (a *AtomArray) Get(index int) *AtomValue {
	return a.Elements[index]
}

func (a *AtomArray) Set(index int, value *AtomValue) {
	a.Elements[index] = value
}

func (a *AtomArray) ValidIndex(index int) bool {
	return index >= 0 && index < len(a.Elements)
}

func (a *AtomArray) Len() int {
	return len(a.Elements)
}

func (a *AtomArray) HashValue() int {
	hash := 0
	for _, element := range a.Elements {
		hash = hash*31 + element.HashValue()
	}
	return hash
}

// Pre-computed slice for method lookup optimization
var arrayMethods = []string{
	"all",    // done
	"any",    //
	"length", // done
	"peek",   //
	"pop",    //
	"push",   // done
	"select", // done
	"where",  // done
	"each",   // done
}

func IsArrayMethod(methodName string) bool {
	return slices.Contains(arrayMethods, methodName)
}

func ArrayAll(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	this := frame.Stack.Pop()
	callback := frame.Stack.Pop()

	// Fast path validation
	if argc != 2 {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, fmt.Sprintf("all expects 2 arguments, got %d", argc)),
		))
		return
	}
	if !CheckType(this, AtomTypeArray) {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "all expects array"),
		))
		return
	}
	if !CheckType(callback, AtomTypeFunc) {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "all expects function"),
		))
		return
	}

	array := this.Value.(*AtomArray)
	elements := array.Elements

	// Early exit for empty arrays
	if len(elements) == 0 {
		frame.Stack.Push(interpreter.State.TrueValue)
		return
	}

	// Process elements with early exit on false
	for _, element := range elements {
		frame.Stack.Push(element)
		DoCall(interpreter, frame, callback, 1)
		result := frame.Stack.Pop()
		if !CoerceToBool(result) {
			frame.Stack.Push(interpreter.State.FalseValue)
			return
		}
	}

	frame.Stack.Push(interpreter.State.TrueValue)
}

func ArrayLength(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	this := frame.Stack.Pop()

	if argc != 1 {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, fmt.Sprintf("length expects 1 arguments, got %d", argc)),
		))
		return
	}
	if !CheckType(this, AtomTypeArray) {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "length expects array"),
		))
		return
	}

	frame.Stack.Push(NewAtomValueInt(this.Value.(*AtomArray).Len()))
}

func ArrayWhere(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	this := frame.Stack.Pop()
	callback := frame.Stack.Pop()

	if argc != 2 {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, fmt.Sprintf("where expects 2 arguments, got %d", argc)),
		))
		return
	}
	if !CheckType(this, AtomTypeArray) {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "where expects array"),
		))
		return
	}
	if !CheckType(callback, AtomTypeFunc) {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "where expects function"),
		))
		return
	}

	array := this.Value.(*AtomArray)
	elements := array.Elements

	resultElements := []*AtomValue{}

	for _, element := range elements {
		frame.Stack.Push(element)
		DoCall(interpreter, frame, callback, 1)
		if CoerceToBool(frame.Stack.Pop()) {
			resultElements = append(resultElements, element)
		}
	}
	frame.Stack.Push(NewAtomValueArray(resultElements))
}

func ArrayPush(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	this := frame.Stack.Pop()
	value := frame.Stack.Pop()

	if argc != 2 {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, fmt.Sprintf("push expects 1 argument, got %d", argc)),
		))
		return
	}
	if !CheckType(this, AtomTypeArray) {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "push expects array"),
		))
		return
	}

	array := this.Value.(*AtomArray)
	array.Elements = append(array.Elements, value)
	frame.Stack.Push(this)
}

func ArraySelect(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	this := frame.Stack.Pop()
	callback := frame.Stack.Pop()

	if argc != 2 {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, fmt.Sprintf("select expects 2 arguments, got %d", argc)),
		))
		return
	}
	if !CheckType(this, AtomTypeArray) {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "select expects array"),
		))
		return
	}
	if !CheckType(callback, AtomTypeFunc) {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "select expects function"),
		))
		return
	}

	array := this.Value.(*AtomArray)
	sourceElements := array.Elements

	// Pre-allocate result slice with known capacity
	elements := make([]*AtomValue, 0, len(sourceElements))

	for index, element := range sourceElements {
		frame.Stack.Push(NewAtomValueInt(index))
		frame.Stack.Push(element)
		DoCall(interpreter, frame, callback, 2)
		elements = append(elements, frame.Stack.Pop())
	}

	frame.Stack.Push(NewAtomValueArray(elements))
}

func ArrayEach(interpreter *AtomInterpreter, frame *AtomCallFrame, argc int) {
	this := frame.Stack.Pop()
	callback := frame.Stack.Pop()

	if argc != 2 {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, fmt.Sprintf("each expects 2 arguments, got %d", argc)),
		))
		return
	}
	if !CheckType(this, AtomTypeArray) {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "each expects array"),
		))
		return
	}
	if !CheckType(callback, AtomTypeFunc) {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "each expects function"),
		))
		return
	}

	array := this.Value.(*AtomArray)
	elements := array.Elements

	for index, element := range elements {
		frame.Stack.Push(NewAtomValueInt(index))
		frame.Stack.Push(element)
		DoCall(interpreter, frame, callback, 2)
		frame.Stack.Pop()
	}
	frame.Stack.Push(interpreter.State.NullValue)
}

func ArrayGetMethod(this *AtomValue, name string) *AtomNativeMethod {
	switch name {
	case "all":
		// arguments(2): this, callback
		return NewAtomNativeMethod(name, 2, this, ArrayAll)
	case "length":
		// arguments(1): this
		return NewAtomNativeMethod(name, 1, this, ArrayLength)
	case "where":
		// arguments(2): this, callback
		return NewAtomNativeMethod(name, 2, this, ArrayWhere)
	case "select":
		// arguments(2): this, callback
		return NewAtomNativeMethod(name, 2, this, ArraySelect)
	case "push":
		// arguments(2): this, value
		return NewAtomNativeMethod(name, 2, this, ArrayPush)
	case "each":
		// arguments(2): this, callback
		return NewAtomNativeMethod(name, 2, this, ArrayEach)
	default:
		panic("Not found!")
	}
}
