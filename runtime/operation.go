package runtime

import (
	"fmt"
	"math"
)

func DoMakeModule(interpreter *AtomInterpreter, frame *AtomCallFrame, size int) {
	elements := map[string]*AtomValue{}
	for range size {
		k := frame.Stack.Pop()
		v := frame.Stack.Pop()
		elements[k.String()] = v
	}
	frame.Stack.Push(NewAtomGenericValue(
		AtomTypeObj,
		NewAtomObject(elements),
	))
}

func DoLoadInt(frame *AtomCallFrame, value int) {
	frame.Stack.Push(NewAtomValueInt(value))
}

func DoLoadNum(frame *AtomCallFrame, value float64) {
	frame.Stack.Push(NewAtomValueNum(value))
}

func DoLoadStr(frame *AtomCallFrame, value string) {
	frame.Stack.Push(NewAtomValueStr(value))
}

func DoLoadBool(interpreter *AtomInterpreter, frame *AtomCallFrame, value int) {
	if value != 0 {
		frame.Stack.Push(interpreter.State.TrueValue)
	} else {
		frame.Stack.Push(interpreter.State.FalseValue)
	}
}

func DoLoadNull(interpreter *AtomInterpreter, frame *AtomCallFrame) {
	frame.Stack.Push(interpreter.State.NullValue)
}

func DoLoadBase(interpreter *AtomInterpreter, frame *AtomCallFrame) {
	// Get base from the "self"
	self := frame.Env.Get("self")
	if self == nil {
		frame.Stack.Push(NewAtomValueError(
			FormatError(frame, "self not defined, cannot get base"),
		))
		return
	}
	base := self.Value.(*AtomClassInstance).Prototype
	if base == nil {
		frame.Stack.Push(interpreter.State.NullValue)
		return
	}
	frame.Stack.Push(base.Value.(*AtomClass).Base)
}

func DoLoadArray(frame *AtomCallFrame, size int) {
	cleanup := func() {
		for range size {
			frame.Stack.Pop()
		}
	}

	elements := []*AtomValue{}
	for i := range size {
		elements = append(elements, frame.Stack.GetOffset(size, i))
	}
	cleanup()
	frame.Stack.Push(NewAtomGenericValue(
		AtomTypeArray,
		NewAtomArray(elements),
	))
}

func DoLoadObject(frame *AtomCallFrame, size int) {
	elements := map[string]*AtomValue{}
	for range size {
		k := frame.Stack.Pop()
		v := frame.Stack.Pop()
		elements[k.String()] = v
	}

	frame.Stack.Push(NewAtomGenericValue(
		AtomTypeObj,
		NewAtomObject(elements),
	))
}

func DoLoadName(frame *AtomCallFrame, index string) {
	// Local?
	if frame.Env.Has(index) {
		frame.Stack.Push(frame.Env.Get(index))
		return
	}
	std_throw_error(frame, NewAtomValueError(
		FormatError(frame, fmt.Sprintf("name '%s' not found", index)),
	))
}

func DoLoadModule(interpreter *AtomInterpreter, frame *AtomCallFrame, name string) {
	module := interpreter.ModuleTable[name]
	if module == nil {
		message := FormatError(frame, fmt.Sprintf("module %s not found", name))
		frame.Stack.Push(NewAtomValueError(message))
		return
	}
	frame.Stack.Push(module)
}

func DoLoadFunction(interpreter *AtomInterpreter, frame *AtomCallFrame, offset int) {
	// Consider everything in the function table is a closure
	templateFn := interpreter.State.FunctionTable.Get(offset)
	templateCode := templateFn.Value.(*AtomCode)
	templateLocals := templateCode.Line
	templateFile := templateCode.File
	templateName := templateCode.Name
	templateAsync := templateCode.Async
	templateArgc := templateCode.Argc

	newCode := NewAtomCode(templateFile, templateName, templateAsync, templateArgc)
	newCode.Line = templateLocals
	newCode.Code = templateCode.Code
	newCode.Capture = frame.Env

	fn := NewAtomGenericValue(AtomTypeFunc, newCode)
	frame.Stack.Push(fn)
}

func DoMakeClass(interpreter *AtomInterpreter, frame *AtomCallFrame, name string, size int) {
	elements := map[string]*AtomValue{}

	for range size {
		k := frame.Stack.Pop()
		v := frame.Stack.Pop()
		elements[k.String()] = v
	}

	frame.Stack.Push(NewAtomGenericValue(
		AtomTypeClass,
		NewAtomClass(
			name,
			nil,
			NewAtomGenericValue(
				AtomTypeObj,
				NewAtomObject(elements),
			),
		),
	))
}

func DoExtendClass(cls *AtomValue, ext *AtomValue) {
	clsValue := cls.Value.(*AtomClass)
	clsValue.Base = ext
}

func DoMakeEnum(frame *AtomCallFrame, size int) {
	elements := map[string]*AtomValue{}
	valueHashes := map[int]bool{}

	for range size {
		k := frame.Stack.Pop()
		v := frame.Stack.Pop()
		key := k.String()

		valueHash := v.HashValue()
		if valueHashes[valueHash] {
			elements[key] = NewAtomValueError(fmt.Sprintf("duplicate value in enum (%s)", v.String()))
		} else {
			elements[key] = v
			valueHashes[valueHash] = true
		}
	}

	frame.Stack.Push(NewAtomGenericValue(
		AtomTypeEnum,
		NewAtomObject(elements),
	))
}

func DoCallConstructor(interpreter *AtomInterpreter, frame *AtomCallFrame, cls *AtomValue, argc int) {
	cleanupStack := func() {
		for range argc {
			frame.Stack.Pop()
		}
	}
	if !CheckType(cls, AtomTypeClass) {
		cleanupStack()
		message := FormatError(frame, GetTypeString(cls)+" is not a constructor")
		frame.Stack.Push(NewAtomValueError(message))
		return
	}

	atomClass := cls.Value.(*AtomClass)

	// Create this
	this := NewAtomGenericValue(
		AtomTypeClassInstance,
		NewAtomClassInstance(cls, NewAtomGenericValue(
			AtomTypeObj,
			NewAtomObject(map[string]*AtomValue{}),
		)),
	)

	// Walk up the inheritance chain to collect all initializers
	var initializers []*AtomValue
	currentClass := atomClass

	for currentClass != nil {
		properties := currentClass.Proto.Value.(*AtomObject)
		if initializer := properties.Get("init"); initializer != nil {
			initializers = append(initializers, initializer)
		}
		if currentClass.Base == nil {
			break
		}
		currentClass = currentClass.Base.Value.(*AtomClass)
	}

	// Call initializers from base to derived (reverse order)
	if len(initializers) == 0 {
		frame.Stack.Push(
			this,
		)
	} else {
		// Call the most derived initializer (last in the slice)
		// The inheritance chain should be handled by the language design,
		// not by calling multiple initializers
		DoCallInit(interpreter, frame, initializers[0], this, 1+argc)
	}
}

func DoCall(interpreter *AtomInterpreter, frame *AtomCallFrame, fn *AtomValue, argc int) {
	if CheckType(fn, AtomTypeMethod) {
		method := fn.Value.(*AtomMethod)
		/*
			    [
			      arg1, <- top
				  arg2,
				  ...
			    ]
				  to
				[
			      arg1, <- top
				  arg2,
				  ...,
				  this
			    ]
		*/

		argc++

		tmpStack := make([]*AtomValue, argc)
		tmpStack[0] = method.This
		for i := 1; i < argc; i++ {
			tmpStack[i] = frame.Stack.GetOffset(argc, i)
		}
		for i := 0; i < argc-1; i++ {
			frame.Stack.Pop()
		}
		for _, item := range tmpStack {
			frame.Stack.Push(item)
		}

		fn = method.Fn

	} else if CheckType(fn, AtomTypeNativeMethod) {
		method := fn.Value.(*AtomNativeMethod)
		/*
			    [
			      arg1, <- top
				  arg2,
				  ...
			    ]
				  to
				[
			      arg1, <- top
				  arg2,
				  ...,
				  this
			    ]
		*/

		argc++

		tmpStack := make([]*AtomValue, argc)
		tmpStack[0] = method.This
		for i := 1; i < argc; i++ {
			tmpStack[i] = frame.Stack.GetOffset(argc, i)
		}
		for i := 0; i < argc-1; i++ {
			frame.Stack.Pop()
		}
		for _, item := range tmpStack {
			frame.Stack.Push(item)
		}
	}

	cleanupStack := func() {
		for range argc {
			frame.Stack.Pop()
		}
	}

	if CheckType(fn, AtomTypeFunc) {
		code := fn.Value.(*AtomCode)
		if argc != code.Argc {
			cleanupStack()
			message := FormatError(frame, fmt.Sprintf("Error: argument count mismatch, expected %d, got %d", code.Argc, argc))
			frame.Stack.Push(NewAtomValueError(message))
			return
		}

		// Create a frame for the function
		newFrame := NewAtomCallFrame(frame, fn, 0)
		newFrame.Stack.Copy(frame.Stack, argc)
		cleanupStack()
		interpreter.ExecuteFrame(newFrame)

		// For async functions, the frame will push a promise to the stack
		// For non-async functions, the frame will push the return value to the stack

	} else if CheckType(fn, AtomTypeNativeFunc) {
		nativeFunc := fn.Value.(*AtomNativeFunc)
		if nativeFunc.Paramc != argc && nativeFunc.Paramc != Variadict {
			cleanupStack()
			message := FormatError(frame, fmt.Sprintf("Error: argument count mismatch, expected %d, got %d", nativeFunc.Paramc, argc))
			frame.Stack.Push(NewAtomValueError(message))
			return
		}

		nativeFunc.Callable(interpreter, frame, argc)

	} else if CheckType(fn, AtomTypeNativeMethod) {
		nativeMethod := fn.Value.(*AtomNativeMethod)
		if nativeMethod.Paramc != argc && nativeMethod.Paramc != Variadict {
			cleanupStack()
			message := FormatError(frame, fmt.Sprintf("Error: argument count mismatch, expected %d, got %d", nativeMethod.Paramc, argc))
			frame.Stack.Push(NewAtomValueError(message))
			return
		}
		nativeMethod.Callable(interpreter, frame, argc)

	} else {
		cleanupStack()
		message := FormatError(frame, fmt.Sprintf("Error: %s is not a function", GetTypeString(fn)))
		frame.Stack.Push(NewAtomValueError(message))
	}
}

func DoCallInit(interpreter *AtomInterpreter, frame *AtomCallFrame, fn *AtomValue, this *AtomValue, argc int) {
	cleanupStack := func() {
		for range argc {
			frame.Stack.Pop()
		}
	}

	tmpStack := make([]*AtomValue, argc)
	tmpStack[0] = this
	for i := 1; i < argc; i++ {
		tmpStack[i] = frame.Stack.GetOffset(argc, i)
	}
	for i := 0; i < argc-1; i++ {
		frame.Stack.Pop()
	}
	for _, item := range tmpStack {
		frame.Stack.Push(item)
	}

	if CheckType(fn, AtomTypeFunc) {
		code := fn.Value.(*AtomCode)
		if argc != code.Argc {
			cleanupStack()
			message := FormatError(frame, fmt.Sprintf("Error: argument count mismatch, expected %d, got %d", code.Argc, argc))
			frame.Stack.Push(NewAtomValueError(message))
			return
		}

		newFrame := NewAtomCallFrame(frame, fn, 0)
		newFrame.Stack.Copy(frame.Stack, argc)
		cleanupStack()
		interpreter.ExecuteFrame(newFrame)

		// Pop return
		frame.Stack.Pop()
		// Push this
		frame.Stack.Push(this)

	} else if CheckType(fn, AtomTypeNativeFunc) {
		nativeFunc := fn.Value.(*AtomNativeFunc)
		if nativeFunc.Paramc != argc && nativeFunc.Paramc != Variadict {
			cleanupStack()
			message := FormatError(frame, "Error: argument count mismatch")
			frame.Stack.Push(NewAtomValueError(message))
			return
		}

		nativeFunc.Callable(interpreter, frame, argc)

		// Pop return
		frame.Stack.Pop()
		// Push this
		frame.Stack.Push(this)

	} else {
		cleanupStack()
		message := FormatError(frame, fmt.Sprintf("Error: %s is not a function", GetTypeString(fn)))
		frame.Stack.Push(NewAtomValueError(message))
	}
}

func DoNot(interpreter *AtomInterpreter, frame *AtomCallFrame, val *AtomValue) {
	if !CoerceToBool(val) {
		frame.Stack.Push(interpreter.State.TrueValue)
		return
	}
	frame.Stack.Push(interpreter.State.FalseValue)
}

func DoNeg(frame *AtomCallFrame, val *AtomValue) {
	if !IsNumberType(val) {
		message := FormatError(frame, fmt.Sprintf("Error: cannot negate type: %s", GetTypeString(val)))
		frame.Stack.Push(NewAtomValueError(message))
		return
	}
	frame.Stack.Push(NewAtomValueNum(-CoerceToNum(val)))
}

func DoPos(frame *AtomCallFrame, val *AtomValue) {
	if !IsNumberType(val) {
		message := FormatError(frame, fmt.Sprintf("Error: cannot pos type: %s", GetTypeString(val)))
		frame.Stack.Push(NewAtomValueError(message))
		return
	}
	frame.Stack.Push(NewAtomValueNum(CoerceToNum(val)))
}

func DoTypeof(frame *AtomCallFrame, val *AtomValue) {
	frame.Stack.Push(NewAtomValueStr(GetTypeString(val)))
}

func DoIndex(interpreter *AtomInterpreter, frame *AtomCallFrame, obj *AtomValue, index *AtomValue) {
	if CheckType(obj, AtomTypeStr) {
		if !IsNumberType(index) {
			message := FormatError(frame, fmt.Sprintf("cannot index type: %s with type: %s", GetTypeString(obj), GetTypeString(index)))
			frame.Stack.Push(NewAtomValueError(message))
			return
		}

		r := []rune(obj.Value.(string))
		indexValue := CoerceToLong(index)
		if indexValue < 0 || indexValue >= int64(len(r)) {
			message := FormatError(frame, fmt.Sprintf("index out of bounds: %d", indexValue))
			frame.Stack.Push(NewAtomValueError(message))
			return
		}

		frame.Stack.Push(NewAtomValueStr(string(r[indexValue])))
		return

	} else if CheckType(obj, AtomTypeArray) {
		if method := index.String(); CheckType(index, AtomTypeStr) && IsArrayMethod(method) {
			frame.Stack.Push(NewAtomGenericValue(
				AtomTypeNativeMethod,
				ArrayGetMethod(obj, method),
			))
			return
		}
		if !IsNumberType(index) {
			message := FormatError(frame, fmt.Sprintf("cannot index type: %s with type: %s", GetTypeString(obj), GetTypeString(index)))
			frame.Stack.Push(NewAtomValueError(message))
			return
		}

		array := obj.Value.(*AtomArray)
		indexValue := CoerceToLong(index)

		if !array.ValidIndex(int(indexValue)) {
			message := FormatError(frame, fmt.Sprintf("index out of bounds: %d", indexValue))
			frame.Stack.Push(NewAtomValueError(message))
			return
		}

		frame.Stack.Push(array.Get(int(indexValue)))
		return

	} else if CheckType(obj, AtomTypeObj) {
		objValue := obj.Value.(*AtomObject)
		indexValue := index.String()

		value := objValue.Get(indexValue)
		if value == nil {
			frame.Stack.Push(interpreter.State.NullValue)
			return
		}

		frame.Stack.Push(value)
		return

	} else if CheckType(obj, AtomTypeClass) {
		class := obj.Value.(*AtomClass)

		for class != nil {
			if value := class.Proto.Value.(*AtomObject).Get(index.String()); value != nil {
				frame.Stack.Push(value)
				return
			}
			if class.Base != nil {
				class = class.Base.Value.(*AtomClass)
			} else {
				break
			}
		}

		frame.Stack.Push(interpreter.State.NullValue)
		return

	} else if CheckType(obj, AtomTypeClassInstance) {
		classInstance := obj.Value.(*AtomClassInstance)
		property := classInstance.Property

		// Direct property?
		if property.Value.(*AtomObject).Get(index.String()) != nil {
			frame.Stack.Push(property.Value.(*AtomObject).Get(index.String()))
			return
		}

		// Is prototype?
		current := classInstance.Prototype.Value.(*AtomClass)
		for current != nil {
			// Class direct prototype?
			if attribute := current.Proto.Value.(*AtomObject).Get(index.String()); attribute != nil {
				if CheckType(attribute, AtomTypeFunc) {
					frame.Stack.Push(NewAtomGenericValue(
						AtomTypeMethod,
						NewAtomMethod(obj, attribute),
					))
					return
				}
				frame.Stack.Push(attribute)
				return
			}
			if current.Base == nil {
				break
			}
			current = current.Base.Value.(*AtomClass)
		}
		frame.Stack.Push(interpreter.State.NullValue)
		return

	} else if CheckType(obj, AtomTypeEnum) {
		if !CheckType(index, AtomTypeStr) {
			message := FormatError(frame, fmt.Sprintf("cannot index type: %s with type: %s", GetTypeString(obj), GetTypeString(index)))
			frame.Stack.Push(NewAtomValueError(message))
			return
		}

		enumValue := obj.Value.(*AtomObject)
		indexValue := index.String()
		value := enumValue.Get(indexValue)

		if value == nil {
			frame.Stack.Push(interpreter.State.NullValue)
			return
		}

		frame.Stack.Push(value)
		return

	} else {
		message := FormatError(frame, fmt.Sprintf("cannot index type: %s", GetTypeString(obj)))
		frame.Stack.Push(NewAtomValueError(message))
		return
	}
}

func DoPluckAttribute(interpreter *AtomInterpreter, frame *AtomCallFrame, obj *AtomValue, attribute string) {
	if !CheckType(obj, AtomTypeObj) {
		message := FormatError(frame, fmt.Sprintf("cannot pluck attribute type: %s", GetTypeString(obj)))
		frame.Stack.Push(NewAtomValueError(message))
		return
	}

	objValue := obj.Value.(*AtomObject)

	if value := objValue.Get(attribute); value != nil {
		frame.Stack.Push(value)
		return
	}
	frame.Stack.Push(interpreter.State.NullValue)
}

func DoMultiplication(frame *AtomCallFrame, val0 *AtomValue, val1 *AtomValue) {
	// Fast path for integers
	if CheckType(val0, AtomTypeInt) && CheckType(val1, AtomTypeInt) {
		lhs := CoerceToInt(val0)
		rhs := CoerceToInt(val1)

		// Check for overflow using int64 arithmetic
		result := int64(lhs) * int64(rhs)
		if result >= math.MinInt32 && result <= math.MaxInt32 {
			frame.Stack.Push(NewAtomValueInt(int(result)))
			return
		}
		// Overflow occurred, promote to float64
		frame.Stack.Push(NewAtomValueNum(float64(result)))
		return
	}

	// Check if both values are numbers (int or float)
	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := FormatError(frame, fmt.Sprintf("Error: cannot multiply types: %s and %s", GetTypeString(val0), GetTypeString(val1)))
		frame.Stack.Push(NewAtomValueError(message))
		return
	}

	// Fallback path using coercion
	lhsValue := CoerceToNum(val0)
	rhsValue := CoerceToNum(val1)
	result := lhsValue * rhsValue

	// Try to preserve integer types if possible
	if IsInteger(result) && result <= math.MaxInt32 && result >= math.MinInt32 {
		frame.Stack.Push(NewAtomValueInt(int(result)))
		return
	}
	frame.Stack.Push(NewAtomValueNum(result))
}

func DoDivision(frame *AtomCallFrame, val0 *AtomValue, val1 *AtomValue) {
	// Fast path for integers
	if CheckType(val0, AtomTypeInt) && CheckType(val1, AtomTypeInt) {
		a := CoerceToInt(val0)
		b := CoerceToInt(val1)
		if b == 0 {
			message := FormatError(frame, "division by zero")
			frame.Stack.Push(NewAtomValueError(message))
			return
		}
		result := a / b
		frame.Stack.Push(NewAtomValueInt(int(result)))
		return
	}

	// Check if both values are numbers (int or float)
	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := FormatError(frame, fmt.Sprintf("Error: cannot divide types: %s and %s", GetTypeString(val0), GetTypeString(val1)))
		frame.Stack.Push(NewAtomValueError(message))
		return
	}

	// Fallback path using coercion
	lhsValue := CoerceToNum(val0)
	rhsValue := CoerceToNum(val1)
	if rhsValue == 0 {
		message := FormatError(frame, "division by zero")
		frame.Stack.Push(NewAtomValueError(message))
		return
	}
	result := lhsValue / rhsValue

	// Try to preserve integer types if possible
	if IsInteger(result) && result <= math.MaxInt32 && result >= math.MinInt32 {
		frame.Stack.Push(NewAtomValueInt(int(result)))
		return
	}
	frame.Stack.Push(NewAtomValueNum(result))
}

func DoModulus(frame *AtomCallFrame, val0 *AtomValue, val1 *AtomValue) {
	// Fast path for integers
	if CheckType(val0, AtomTypeInt) && CheckType(val1, AtomTypeInt) {
		a := CoerceToInt(val0)
		b := CoerceToInt(val1)
		if b == 0 {
			message := FormatError(frame, "Error: division by zero")
			frame.Stack.Push(NewAtomValueError(message))
			return
		}
		result := a % b
		frame.Stack.Push(NewAtomValueInt(int(result)))
		return
	}

	// Check if both values are numbers (int or float)
	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := FormatError(frame, fmt.Sprintf("Error: cannot modulo types: %s and %s", GetTypeString(val0), GetTypeString(val1)))
		frame.Stack.Push(NewAtomValueError(message))
		return
	}

	// Fallback path using coercion
	lhsValue := CoerceToNum(val0)
	rhsValue := CoerceToNum(val1)
	if rhsValue == 0 {
		message := FormatError(frame, "Error: division by zero")
		frame.Stack.Push(NewAtomValueError(message))
		return
	}
	result := math.Mod(lhsValue, rhsValue)

	// Try to preserve integer types if possible
	if IsInteger(result) && result <= math.MaxInt32 && result >= math.MinInt32 {
		frame.Stack.Push(NewAtomValueInt(int(result)))
		return
	}
	frame.Stack.Push(NewAtomValueNum(result))
}

func DoAddition(frame *AtomCallFrame, val0 *AtomValue, val1 *AtomValue) {
	// Fast path for integers
	if CheckType(val0, AtomTypeInt) && CheckType(val1, AtomTypeInt) {
		// Use XOR trick to detect overflow
		a := CoerceToInt(val0)
		b := CoerceToInt(val1)
		sum := a + b
		if ((a ^ sum) & (b ^ sum)) < 0 {
			// Overflow occurred, promote to double
			frame.Stack.Push(NewAtomValueNum(float64(a) + float64(b)))
			return
		}
		frame.Stack.Push(NewAtomValueInt(int(sum)))
		return
	}

	// Fast path for strings
	if CheckType(val0, AtomTypeStr) && CheckType(val1, AtomTypeStr) {
		lhs := val0.Value.(string)
		rhs := val1.Value.(string)
		result := lhs + rhs
		frame.Stack.Push(NewAtomValueStr(result))
		return
	}

	if CheckType(val0, AtomTypeStr) || CheckType(val1, AtomTypeStr) {
		lhs := val0.String()
		rhs := val1.String()
		result := lhs + rhs
		frame.Stack.Push(NewAtomValueStr(result))
		return
	}

	// Check if both values are numbers (int or float)
	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := FormatError(frame, fmt.Sprintf("Error: cannot add types: %s and %s", GetTypeString(val0), GetTypeString(val1)))
		frame.Stack.Push(NewAtomValueError(message))
		return
	}

	// Fallback path using coercion
	lhsValue := CoerceToNum(val0)
	rhsValue := CoerceToNum(val1)
	result := lhsValue + rhsValue

	// Try to preserve integer types if possible
	if IsInteger(result) && result <= math.MaxInt32 && result >= math.MinInt32 {
		frame.Stack.Push(NewAtomValueInt(int(result)))
		return
	}
	frame.Stack.Push(NewAtomValueNum(result))
}

func DoSubtraction(frame *AtomCallFrame, val0 *AtomValue, val1 *AtomValue) {
	// Fast path for integers
	if CheckType(val0, AtomTypeInt) && CheckType(val1, AtomTypeInt) {
		a := CoerceToInt(val0)
		b := CoerceToInt(val1)
		diff := a - b
		if ((a ^ b) & (a ^ diff)) < 0 {
			// Overflow occurred, promote to double
			frame.Stack.Push(NewAtomValueNum(float64(a) - float64(b)))
			return
		}
		frame.Stack.Push(NewAtomValueInt(int(diff)))
		return
	}

	// Check if both values are numbers (int or float)
	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := FormatError(frame, fmt.Sprintf("Error: cannot subtract types: %s and %s", GetTypeString(val0), GetTypeString(val1)))
		frame.Stack.Push(NewAtomValueError(message))
		return
	}

	// Fallback path using coercion
	lhsValue := CoerceToNum(val0)
	rhsValue := CoerceToNum(val1)
	result := lhsValue - rhsValue

	// Try to preserve integer types if possible
	if IsInteger(result) && result <= math.MaxInt32 && result >= math.MinInt32 {
		frame.Stack.Push(NewAtomValueInt(int(result)))
		return
	}
	frame.Stack.Push(NewAtomValueNum(result))
}

func DoShiftLeft(frame *AtomCallFrame, val0 *AtomValue, val1 *AtomValue) {
	// Fast path for integers
	if CheckType(val0, AtomTypeInt) && CheckType(val1, AtomTypeInt) {
		a := CoerceToInt(val0)
		b := CoerceToInt(val1)
		result := a << b
		frame.Stack.Push(NewAtomValueInt(int(result)))
		return
	}

	// Check if both values are numbers (int or float)
	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := FormatError(frame, fmt.Sprintf("Error: cannot shift left types: %s and %s", GetTypeString(val0), GetTypeString(val1)))
		frame.Stack.Push(NewAtomValueError(message))
		return
	}

	// Fallback path using coercion
	lhsValue := CoerceToNum(val0)
	rhsValue := CoerceToNum(val1)
	result := int64(lhsValue) << int64(rhsValue)

	// Check if result can be represented as an int
	if result >= math.MinInt32 && result <= math.MaxInt32 {
		frame.Stack.Push(NewAtomValueInt(int(result)))
		return
	}
	frame.Stack.Push(NewAtomValueNum(float64(result)))
}

func DoShiftRight(frame *AtomCallFrame, val0 *AtomValue, val1 *AtomValue) {
	// Fast path for integers
	if CheckType(val0, AtomTypeInt) && CheckType(val1, AtomTypeInt) {
		a := CoerceToInt(val0)
		b := CoerceToInt(val1)
		result := a >> b
		frame.Stack.Push(NewAtomValueInt(int(result)))
		return
	}

	// Check if both values are numbers (int or float)
	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := FormatError(frame, fmt.Sprintf("Error: cannot shift right types: %s and %s", GetTypeString(val0), GetTypeString(val1)))
		frame.Stack.Push(NewAtomValueError(message))
		return
	}

	// Fallback path using coercion
	lhsValue := CoerceToNum(val0)
	rhsValue := CoerceToNum(val1)
	result := int64(lhsValue) >> int64(rhsValue)

	// Try to preserve integer types if possible
	if result >= math.MinInt32 && result <= math.MaxInt32 {
		frame.Stack.Push(NewAtomValueInt(int(result)))
		return
	}
	frame.Stack.Push(NewAtomValueNum(float64(result)))
}

func DoCmpLt(interpreter *AtomInterpreter, frame *AtomCallFrame, val0 *AtomValue, val1 *AtomValue) {
	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := FormatError(frame, fmt.Sprintf("Error: cannot compare less than type(s) %s and %s", GetTypeString(val0), GetTypeString(val1)))
		frame.Stack.Push(NewAtomValueError(message))
		return
	}

	// Coerce to long to avoid floating point comparisons
	lhsValue := CoerceToLong(val0)
	rhsValue := CoerceToLong(val1)

	// Compare the long values
	if lhsValue < rhsValue {
		frame.Stack.Push(interpreter.State.TrueValue)
		return
	}
	frame.Stack.Push(interpreter.State.FalseValue)
}

func DoCmpLte(interpreter *AtomInterpreter, frame *AtomCallFrame, val0 *AtomValue, val1 *AtomValue) {
	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := FormatError(frame, fmt.Sprintf("Error: cannot compare less than or equal to type(s) %s and %s", GetTypeString(val0), GetTypeString(val1)))
		frame.Stack.Push(NewAtomValueError(message))
		return
	}

	// Coerce to long to avoid floating point comparisons
	lhsValue := CoerceToLong(val0)
	rhsValue := CoerceToLong(val1)

	// Compare the long values
	if lhsValue <= rhsValue {
		frame.Stack.Push(interpreter.State.TrueValue)
		return
	}
	frame.Stack.Push(interpreter.State.FalseValue)
}

func DoCmpGt(interpreter *AtomInterpreter, frame *AtomCallFrame, val0 *AtomValue, val1 *AtomValue) {
	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := FormatError(frame, fmt.Sprintf("Error: cannot compare greater than type(s) %s and %s", GetTypeString(val0), GetTypeString(val1)))
		frame.Stack.Push(NewAtomValueError(message))
		return
	}

	// Coerce to long to avoid floating point comparisons
	lhsValue := CoerceToLong(val0)
	rhsValue := CoerceToLong(val1)

	// Compare the long values
	if lhsValue > rhsValue {
		frame.Stack.Push(interpreter.State.TrueValue)
		return
	}
	frame.Stack.Push(interpreter.State.FalseValue)
}

func DoCmpGte(interpreter *AtomInterpreter, frame *AtomCallFrame, val0 *AtomValue, val1 *AtomValue) {
	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := FormatError(frame, fmt.Sprintf("Error: cannot compare greater than or equal to type(s) %s and %s", GetTypeString(val0), GetTypeString(val1)))
		frame.Stack.Push(NewAtomValueError(message))
		return
	}

	// Coerce to long to avoid floating point comparisons
	lhsValue := CoerceToLong(val0)
	rhsValue := CoerceToLong(val1)

	// Compare the long values
	if lhsValue >= rhsValue {
		frame.Stack.Push(interpreter.State.TrueValue)
		return
	}
	frame.Stack.Push(interpreter.State.FalseValue)
}

func DoCmpEq(interpreter *AtomInterpreter, frame *AtomCallFrame, val0 *AtomValue, val1 *AtomValue) {
	if IsNumberType(val0) && IsNumberType(val1) {
		lhsValue := CoerceToLong(val0)
		rhsValue := CoerceToLong(val1)
		if lhsValue == rhsValue {
			frame.Stack.Push(interpreter.State.TrueValue)
			return
		}
		frame.Stack.Push(interpreter.State.FalseValue)
		return
	}

	if CheckType(val0, AtomTypeStr) && CheckType(val1, AtomTypeStr) {
		lhsStr := val0.Value.(string)
		rhsStr := val1.Value.(string)
		if lhsStr == rhsStr {
			frame.Stack.Push(interpreter.State.TrueValue)
			return
		}
		frame.Stack.Push(interpreter.State.FalseValue)
		return
	}

	if CheckType(val0, AtomTypeNull) && CheckType(val1, AtomTypeNull) {
		frame.Stack.Push(interpreter.State.TrueValue)
		return
	}

	// For other types, use simple reference equality for now
	if val0.HashValue() == val1.HashValue() || val0 == val1 {
		frame.Stack.Push(interpreter.State.TrueValue)
		return
	}

	frame.Stack.Push(interpreter.State.FalseValue)
}

func DoCmpNe(interpreter *AtomInterpreter, frame *AtomCallFrame, val0 *AtomValue, val1 *AtomValue) {
	if IsNumberType(val0) && IsNumberType(val1) {
		lhsValue := CoerceToLong(val0)
		rhsValue := CoerceToLong(val1)
		if lhsValue == rhsValue {
			frame.Stack.Push(interpreter.State.FalseValue)
			return
		}
		frame.Stack.Push(interpreter.State.TrueValue)
		return
	}

	if CheckType(val0, AtomTypeStr) && CheckType(val1, AtomTypeStr) {
		lhsStr := val0.Value.(string)
		rhsStr := val1.Value.(string)
		if lhsStr == rhsStr {
			frame.Stack.Push(interpreter.State.FalseValue)
			return
		}
		frame.Stack.Push(interpreter.State.TrueValue)
		return
	}

	if CheckType(val0, AtomTypeNull) && CheckType(val1, AtomTypeNull) {
		frame.Stack.Push(interpreter.State.FalseValue)
		return
	}

	// For other types, use simple reference equality for now
	if val0.HashValue() == val1.HashValue() || val0 == val1 {
		frame.Stack.Push(interpreter.State.FalseValue)
		return
	}

	frame.Stack.Push(interpreter.State.TrueValue)
}

func DoAnd(frame *AtomCallFrame, val0 *AtomValue, val1 *AtomValue) {
	// Fast path for integers
	if CheckType(val0, AtomTypeInt) && CheckType(val1, AtomTypeInt) {
		a := val0.Value.(int32)
		b := val1.Value.(int32)
		result := a & b
		frame.Stack.Push(NewAtomValueInt(int(result)))
		return
	}

	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := FormatError(frame, fmt.Sprintf("Error: cannot bitwise and type(s) %s and %s", GetTypeString(val0), GetTypeString(val1)))
		frame.Stack.Push(NewAtomValueError(message))
		return
	}

	lhsValue := CoerceToLong(val0)
	rhsValue := CoerceToLong(val1)
	result := lhsValue & rhsValue

	// Check if result can be represented as an int
	if result >= math.MinInt32 && result <= math.MaxInt32 {
		frame.Stack.Push(NewAtomValueInt(int(result)))
	} else {
		frame.Stack.Push(NewAtomValueNum(float64(result)))
	}
}

func DoOr(frame *AtomCallFrame, val0 *AtomValue, val1 *AtomValue) {
	// Fast path for integers
	if CheckType(val0, AtomTypeInt) && CheckType(val1, AtomTypeInt) {
		a := val0.Value.(int32)
		b := val1.Value.(int32)
		result := a | b
		frame.Stack.Push(NewAtomValueInt(int(result)))
		return
	}

	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := FormatError(frame, fmt.Sprintf("Error: cannot bitwise or type(s) %s and %s", GetTypeString(val0), GetTypeString(val1)))
		frame.Stack.Push(NewAtomValueError(message))
		return
	}

	lhsValue := CoerceToLong(val0)
	rhsValue := CoerceToLong(val1)
	result := lhsValue | rhsValue

	// Check if result can be represented as an int
	if result >= math.MinInt32 && result <= math.MaxInt32 {
		frame.Stack.Push(NewAtomValueInt(int(result)))
	} else {
		frame.Stack.Push(NewAtomValueNum(float64(result)))
	}
}

func DoXor(frame *AtomCallFrame, val0 *AtomValue, val1 *AtomValue) {
	// Fast path for integers
	if CheckType(val0, AtomTypeInt) && CheckType(val1, AtomTypeInt) {
		a := val0.Value.(int32)
		b := val1.Value.(int32)
		result := a ^ b
		frame.Stack.Push(NewAtomValueInt(int(result)))
		return
	}

	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := FormatError(frame, fmt.Sprintf("Error: cannot bitwise xor type(s) %s and %s", GetTypeString(val0), GetTypeString(val1)))
		frame.Stack.Push(NewAtomValueError(message))
		return
	}

	lhsValue := CoerceToLong(val0)
	rhsValue := CoerceToLong(val1)
	result := lhsValue ^ rhsValue

	// Check if result can be represented as an int
	if result >= math.MinInt32 && result <= math.MaxInt32 {
		frame.Stack.Push(NewAtomValueInt(int(result)))
	} else {
		frame.Stack.Push(NewAtomValueNum(float64(result)))
	}
}

func DoStoreModule(interpreter *AtomInterpreter, frame *AtomCallFrame, name string) {
	module := frame.Stack.Pop()
	module.Value.(*AtomObject).Set("__name__", NewAtomValueStr(name))
	interpreter.ModuleTable[name] = module
}

func DoInitLocal(interpreter *AtomInterpreter, frame *AtomCallFrame, name string, value *AtomValue) {
	frame.Env.Put(name, value)
}

func DoStoreLocal(interpreter *AtomInterpreter, frame *AtomCallFrame, name string, value *AtomValue) {
	// Local?
	if frame.Env.Has(name) {
		frame.Env.Set(name, value)
		return
	}
	panic("Not handled properly!!")
}

func DoSetIndex(interpreter *AtomInterpreter, frame *AtomCallFrame, obj *AtomValue, index *AtomValue) {
	cleanupStack := func(size int) {
		for range size {
			frame.Stack.Pop()
		}
	}
	if CheckType(obj, AtomTypeArray) {
		if !IsNumberType(index) {
			cleanupStack(1)
			message := FormatError(frame, fmt.Sprintf("cannot set index type: %s with type: %s", GetTypeString(obj), GetTypeString(index)))
			frame.Stack.Push(NewAtomValueError(message))
			return
		}
		array := obj.Value.(*AtomArray)
		indexValue := CoerceToLong(index)

		if array.Freeze {
			cleanupStack(2)
			message := FormatError(frame, "cannot set index on frozen array")
			frame.Stack.Push(NewAtomValueError(message))
			return
		}

		if !array.ValidIndex(int(indexValue)) {
			cleanupStack(2)
			message := FormatError(frame, fmt.Sprintf("index out of bounds: %d", indexValue))
			frame.Stack.Push(NewAtomValueError(message))
			return
		}

		array.Set(int(indexValue), frame.Stack.Pop())
		return

	} else if CheckType(obj, AtomTypeObj) {
		if obj.Value.(*AtomObject).Freeze {
			cleanupStack(2) // includes duplicate obj
			message := FormatError(frame, "cannot set index on frozen object")
			frame.Stack.Push(NewAtomValueError(message))
			return
		}

		objValue := obj.Value.(*AtomObject)
		indexValue := index.String()
		objValue.Set(indexValue, frame.Stack.Pop())
		return

	} else if CheckType(obj, AtomTypeClass) {
		class := obj.Value.(*AtomClass)
		class.Proto.Value.(*AtomObject).Set(index.String(), frame.Stack.Pop())
		return
	} else if CheckType(obj, AtomTypeClassInstance) {
		classInstance := obj.Value.(*AtomClassInstance)
		classInstance.Property.Value.(*AtomObject).Set(index.String(), frame.Stack.Pop())
		return
	} else {
		cleanupStack(2)
		message := FormatError(frame, fmt.Sprintf("cannot set index type: %s", GetTypeString(obj)))
		frame.Stack.Push(NewAtomValueError(message))
		return
	}
}

func DoInc(frame *AtomCallFrame, val *AtomValue) {
	if !IsNumberType(val) {
		message := FormatError(frame, fmt.Sprintf("Error: cannot increment type: %s", GetTypeString(val)))
		frame.Stack.Push(NewAtomValueError(message))
		return
	}

	if CheckType(val, AtomTypeInt) {
		// Fast path for integers
		a := CoerceToInt(val)
		result := a + 1
		// Check for overflow
		if (a > 0 && result < 0) || (a < 0 && result > 0) {
			// Overflow occurred, promote to double
			frame.Stack.Push(NewAtomValueNum(float64(a) + 1))
			return
		}
		frame.Stack.Push(NewAtomValueInt(int(result)))
		return
	}

	// Fallback path using coercion
	numValue := CoerceToNum(val)
	result := numValue + 1

	// Try to preserve integer types if possible
	if IsInteger(result) && result <= math.MaxInt32 && result >= math.MinInt32 {
		frame.Stack.Push(NewAtomValueInt(int(result)))
		return
	}
	frame.Stack.Push(NewAtomValueNum(result))
}

func DoDec(frame *AtomCallFrame, val *AtomValue) {
	if !IsNumberType(val) {
		message := FormatError(frame, fmt.Sprintf("Error: cannot decrement type: %s", GetTypeString(val)))
		frame.Stack.Push(NewAtomValueError(message))
		return
	}

	if CheckType(val, AtomTypeInt) {
		// Fast path for integers
		a := CoerceToInt(val)
		result := a - 1
		// Check for overflow
		if (a > 0 && result < 0) || (a < 0 && result > 0) {
			// Overflow occurred, promote to double
			frame.Stack.Push(NewAtomValueNum(float64(a) - 1))
			return
		}
		frame.Stack.Push(NewAtomValueInt(int(result)))
		return
	}

	// Fallback path using coercion
	numValue := CoerceToNum(val)
	result := numValue - 1

	// Try to preserve integer types if possible
	if IsInteger(result) && result <= math.MaxInt32 && result >= math.MinInt32 {
		frame.Stack.Push(NewAtomValueInt(int(result)))
		return
	}
	frame.Stack.Push(NewAtomValueNum(result))
}

func DoRot2(frame *AtomCallFrame) {
	// [A, B] -> [B, A]
	A := frame.Stack.Pop()
	B := frame.Stack.Pop()
	frame.Stack.Push(A)
	frame.Stack.Push(B)
}

func DoRot3(frame *AtomCallFrame) {
	// [A, B, C] -> [C, A, B]
	C := frame.Stack.Pop()
	B := frame.Stack.Pop()
	A := frame.Stack.Pop()
	frame.Stack.Push(C)
	frame.Stack.Push(A)
	frame.Stack.Push(B)
}

func DoRot4(frame *AtomCallFrame) {
	// [A, B, C, D] -> [D, A, B, C]
	D := frame.Stack.Pop()
	C := frame.Stack.Pop()
	B := frame.Stack.Pop()
	A := frame.Stack.Pop()
	frame.Stack.Push(D)
	frame.Stack.Push(A)
	frame.Stack.Push(B)
	frame.Stack.Push(C)
}
