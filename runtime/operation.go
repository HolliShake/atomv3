package runtime

import (
	"fmt"
	"math"
)

func DoLoadArray(frame *AtomCallFrame, size int) {
	elements := []*AtomValue{}
	for range size {
		elements = append(elements, frame.Stack.Pop())
	}
	frame.Stack.Push(
		NewAtomValueArray(elements),
	)
}

func DoLoadObject(frame *AtomCallFrame, size int) {
	elements := map[string]*AtomValue{}
	for range size {
		k := frame.Stack.Pop()
		v := frame.Stack.Pop()
		elements[k.Value.(string)] = v
	}
	frame.Stack.Push(
		NewAtomValueObject(elements),
	)
}

func DoLoadName(frame *AtomCallFrame, env *AtomEnv, name string) {
	value, err := env.Lookup(name)
	if err != nil {
		frame.Stack.Push(
			NewAtomValueError(err.Error()),
		)
	} else {
		frame.Stack.Push(value)
	}
}

func DoLoadModule0(interpreter *AtomInterpreter, frame *AtomCallFrame, name string) {
	module := interpreter.ModuleTable[name]
	if module == nil {
		message := fmt.Sprintf("module %s not found", name)
		frame.Stack.Push(NewAtomValueError(message))
		return
	}
	frame.Stack.Push(module)
}

func DoIndex(interpreter *AtomInterpreter, frame *AtomCallFrame, obj *AtomValue, index *AtomValue) {
	if CheckType(obj, AtomTypeStr) {
		if !IsNumberType(index) {
			message := fmt.Sprintf("cannot index type: %s with type: %s", GetTypeString(obj), GetTypeString(index))
			frame.Stack.Push(NewAtomValueError(message))
			return
		}

		r := []rune(obj.Value.(string))
		indexValue := CoerceToLong(index)
		if indexValue < 0 || indexValue >= int64(len(r)) {
			message := fmt.Sprintf("index out of bounds: %d", indexValue)
			frame.Stack.Push(NewAtomValueError(message))
			return
		}

		frame.Stack.Push(NewAtomValueStr(string(r[indexValue])))
		return

	} else if CheckType(obj, AtomTypeArray) {
		if !IsNumberType(index) {
			message := fmt.Sprintf("cannot index type: %s with type: %s", GetTypeString(obj), GetTypeString(index))
			frame.Stack.Push(NewAtomValueError(message))
			return
		}

		array := obj.Value.(*AtomArray)
		indexValue := CoerceToLong(index)

		if !array.ValidIndex(int(indexValue)) {
			message := fmt.Sprintf("index out of bounds: %d", indexValue)
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
		if class.Proto.Value.(*AtomObject).Get(index.String()) != nil {
			frame.Stack.Push(class.Proto.Value.(*AtomObject).Get(index.String()))
			return
		}
		frame.Stack.Push(interpreter.State.NullValue)
		return

	} else if CheckType(obj, AtomTypeClassInstance) {
		classInstance := obj.Value.(*AtomClassInstance)
		property := classInstance.Property

		if property.Value.(*AtomObject).Get(index.String()) != nil {
			frame.Stack.Push(property.Value.(*AtomObject).Get(index.String()))
			return
		}
		// Is proto
		current := classInstance.Prototype.Value.(*AtomClass)
		for current != nil {
			if attribute := current.Proto.Value.(*AtomObject).Get(index.String()); attribute != nil {
				if CheckType(attribute, AtomTypeFunc) || CheckType(attribute, AtomTypeNativeFunc) {
					frame.Stack.Push(NewAtomValueMethod(property, attribute))
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
			message := fmt.Sprintf("cannot index type: %s with type: %s", GetTypeString(obj), GetTypeString(index))
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
		message := fmt.Sprintf("cannot index type: %s", GetTypeString(obj))
		frame.Stack.Push(NewAtomValueError(message))
		return
	}
}

func DoPluckAttribute(interpreter *AtomInterpreter, frame *AtomCallFrame, obj *AtomValue, attribute string) {
	if !CheckType(obj, AtomTypeObj) {
		message := fmt.Sprintf("cannot pluck attribute type: %s", GetTypeString(obj))
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

func DoSetIndex(interpreter *AtomInterpreter, frame *AtomCallFrame, obj *AtomValue, index *AtomValue) {
	cleanupStack := func(size int) {
		for range size {
			frame.Stack.Pop()
		}
	}
	if CheckType(obj, AtomTypeArray) {
		if !IsNumberType(index) {
			cleanupStack(1)
			message := fmt.Sprintf("cannot set index type: %s with type: %s", GetTypeString(obj), GetTypeString(index))
			frame.Stack.Push(NewAtomValueError(message))
			return
		}
		array := obj.Value.(*AtomArray)
		indexValue := CoerceToLong(index)

		if array.Freeze {
			cleanupStack(2)
			message := "cannot set index on frozen array"
			frame.Stack.Push(NewAtomValueError(message))
			return
		}

		if !array.ValidIndex(int(indexValue)) {
			cleanupStack(2)
			message := fmt.Sprintf("index out of bounds: %d", indexValue)
			frame.Stack.Push(NewAtomValueError(message))
			return
		}

		array.Set(int(indexValue), frame.Stack.Pop())
		return

	} else if CheckType(obj, AtomTypeObj) {
		if obj.Value.(*AtomObject).Freeze {
			cleanupStack(2) // includes duplicate obj
			message := "cannot set index on frozen object"
			frame.Stack.Push(NewAtomValueError(message))
			return
		}

		objValue := obj.Value.(*AtomObject)
		indexValue := index.String()
		objValue.Set(indexValue, frame.Stack.Pop())
		return

	} else {
		cleanupStack(2)
		message := fmt.Sprintf("cannot set index type: %s", GetTypeString(obj))
		frame.Stack.Push(NewAtomValueError(message))
		return
	}
}

func DoLoadFunction(interpreter *AtomInterpreter, frame *AtomCallFrame, offset int) {
	fn := interpreter.State.FunctionTable.Get(offset)
	frame.Stack.Push(fn)
}

func DoMakeClass(interpreter *AtomInterpreter, frame *AtomCallFrame, name string, size int) {
	elements := map[string]*AtomValue{}

	for range size {
		k := frame.Stack.Pop()
		v := frame.Stack.Pop()
		elements[k.Value.(string)] = v
	}

	frame.Stack.Push(NewAtomValueClass(
		name,
		nil,
		NewAtomValueObject(elements),
	))
}

func DoExtendClass(cls *AtomValue, ext *AtomValue) {
	clsValue := cls.Value.(*AtomClass)
	clsValue.Base = ext
}

func DoMakeEnum(frame *AtomCallFrame, env *AtomEnv, size int) {
	elements := map[string]*AtomValue{}
	valueHashes := map[int]bool{}

	for range size {
		k := frame.Stack.Pop()
		v := frame.Stack.Pop()
		key := k.Value.(string)

		valueHash := v.HashValue()
		if valueHashes[valueHash] {
			elements[key] = NewAtomValueError(fmt.Sprintf("duplicate value in enum (%s)", v.String()))
		} else {
			elements[key] = v
			valueHashes[valueHash] = true
		}
	}

	frame.Stack.Push(NewAtomValueEnum(elements))
}

func DoCallConstructor(interpreter *AtomInterpreter, frame *AtomCallFrame, env *AtomEnv, cls *AtomValue, argc int) {
	cleanupStack := func() {
		for range argc {
			frame.Stack.Pop()
		}
	}
	if !CheckType(cls, AtomTypeClass) {
		cleanupStack()
		message := GetTypeString(cls) + " is not a constructor"
		frame.Stack.Push(NewAtomValueError(message))
		return
	}
	atomClass := cls.Value.(*AtomClass)

	// Create the instance object first
	this := NewAtomValueObject(map[string]*AtomValue{})

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
			NewAtomValueClassInstance(cls, this),
		)
	} else {
		// Call the most derived initializer (last in the slice)
		// The inheritance chain should be handled by the language design,
		// not by calling multiple initializers
		DoCallInit(interpreter, frame, env, cls, initializers[len(initializers)-1], this, 1+argc)
	}
}

func DoCallInit(interpreter *AtomInterpreter, frame *AtomCallFrame, env *AtomEnv, cls *AtomValue, fn *AtomValue, this *AtomValue, argc int) {
	cleanupStack := func() {
		for range argc {
			frame.Stack.Pop()
		}
	}

	// Push this
	frame.Stack.Push(this)

	if CheckType(fn, AtomTypeFunc) {
		code := fn.Value.(*AtomCode)
		if argc != code.Argc {
			cleanupStack()
			message := fmt.Sprintf("Error: argument count mismatch, expected %d, got %d", code.Argc, argc)
			frame.Stack.Push(NewAtomValueError(message))
			return
		}

		newFrame := NewAtomCallFrame(frame, fn, NewAtomEnv(env), 0)
		newFrame.Stack.Copy(frame.Stack, argc)
		interpreter.ExecuteFrame(newFrame)

		// Pop return
		frame.Stack.Pop()
		frame.Stack.Push(
			NewAtomValueClassInstance(cls, this),
		)

	} else if CheckType(fn, AtomTypeNativeFunc) {
		nativeFunc := fn.Value.(NativeFunc)
		if nativeFunc.Paramc != argc && nativeFunc.Paramc != Variadict {
			cleanupStack()
			message := "Error: argument count mismatch"
			frame.Stack.Push(NewAtomValueError(message))
			return
		}

		nativeFunc.Callable(interpreter, frame, argc)

		// Pop return
		frame.Stack.Pop()
		frame.Stack.Push(
			NewAtomValueClassInstance(cls, this),
		)

	} else {
		cleanupStack()
		message := fmt.Sprintf("Error: %s is not a function", GetTypeString(fn))
		frame.Stack.Push(NewAtomValueError(message))
	}
}

func DoCall(interpreter *AtomInterpreter, frame *AtomCallFrame, env *AtomEnv, fn *AtomValue, argc int) {
	cleanupStack := func() {
		for range argc {
			frame.Stack.Pop()
		}
	}

	if CheckType(fn, AtomTypeMethod) {
		method := fn.Value.(*AtomMethod)
		frame.Stack.Push(method.This)
		fn = method.Fn
		argc++
	}

	if CheckType(fn, AtomTypeFunc) {
		code := fn.Value.(*AtomCode)
		if argc != code.Argc {
			cleanupStack()
			message := fmt.Sprintf("Error: argument count mismatch, expected %d, got %d", code.Argc, argc)
			frame.Stack.Push(NewAtomValueError(message))
			return
		}

		// Create a frame for the function
		newFrame := NewAtomCallFrame(frame, fn, NewAtomEnv(env), 0)
		newFrame.Stack.Copy(frame.Stack, argc)
		interpreter.ExecuteFrame(newFrame)

		// For async functions, the frame will push a promise to the stack
		// For non-async functions, the frame will push the return value to the stack

	} else if CheckType(fn, AtomTypeNativeFunc) {
		nativeFunc := fn.Value.(NativeFunc)
		if nativeFunc.Paramc != argc && nativeFunc.Paramc != Variadict {
			cleanupStack()
			message := fmt.Sprintf("Error: argument count mismatch, expected %d, got %d", nativeFunc.Paramc, argc)
			frame.Stack.Push(NewAtomValueError(message))
			return
		}

		nativeFunc.Callable(interpreter, frame, argc)

	} else {
		cleanupStack()
		message := fmt.Sprintf("Error: %s is not a function", GetTypeString(fn))
		frame.Stack.Push(NewAtomValueError(message))
	}
}

func DoNeg(frame *AtomCallFrame, val *AtomValue) {
	if !IsNumberType(val) {
		message := fmt.Sprintf("Error: cannot negate type: %s", GetTypeString(val))
		frame.Stack.Push(NewAtomValueError(message))
		return
	}
	frame.Stack.Push(NewAtomValueNum(-CoerceToNum(val)))
}

func DoNot(interpreter *AtomInterpreter, frame *AtomCallFrame, val *AtomValue) {
	if !CoerceToBool(val) {
		frame.Stack.Push(interpreter.State.TrueValue)
		return
	}
	frame.Stack.Push(interpreter.State.FalseValue)
}

func DoPos(frame *AtomCallFrame, val *AtomValue) {
	if !IsNumberType(val) {
		message := fmt.Sprintf("Error: cannot pos type: %s", GetTypeString(val))
		frame.Stack.Push(NewAtomValueError(message))
		return
	}
	frame.Stack.Push(NewAtomValueNum(CoerceToNum(val)))
}

func DoTypeof(frame *AtomCallFrame, val *AtomValue) {
	frame.Stack.Push(NewAtomValueStr(GetTypeString(val)))
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
		message := fmt.Sprintf("Error: cannot add types: %s and %s", GetTypeString(val0), GetTypeString(val1))
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

func DoDivision(frame *AtomCallFrame, val0 *AtomValue, val1 *AtomValue) {
	// Fast path for integers
	if CheckType(val0, AtomTypeInt) && CheckType(val1, AtomTypeInt) {
		a := CoerceToInt(val0)
		b := CoerceToInt(val1)
		if b == 0 {
			message := "Error: division by zero"
			frame.Stack.Push(NewAtomValueError(message))
			return
		}
		result := a / b
		frame.Stack.Push(NewAtomValueInt(int(result)))
		return
	}

	// Check if both values are numbers (int or float)
	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := fmt.Sprintf("Error: cannot divide types: %s and %s", GetTypeString(val0), GetTypeString(val1))
		frame.Stack.Push(NewAtomValueError(message))
		return
	}

	// Fallback path using coercion
	lhsValue := CoerceToNum(val0)
	rhsValue := CoerceToNum(val1)
	if rhsValue == 0 {
		message := "division by zero"
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
			message := "Error: division by zero"
			frame.Stack.Push(NewAtomValueError(message))
			return
		}
		result := a % b
		frame.Stack.Push(NewAtomValueInt(int(result)))
		return
	}

	// Check if both values are numbers (int or float)
	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := fmt.Sprintf("Error: cannot modulo types: %s and %s", GetTypeString(val0), GetTypeString(val1))
		frame.Stack.Push(NewAtomValueError(message))
		return
	}

	// Fallback path using coercion
	lhsValue := CoerceToNum(val0)
	rhsValue := CoerceToNum(val1)
	if rhsValue == 0 {
		message := "Error: division by zero"
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
		message := fmt.Sprintf("Error: cannot multiply types: %s and %s", GetTypeString(val0), GetTypeString(val1))
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
		message := fmt.Sprintf("Error: cannot subtract types: %s and %s", GetTypeString(val0), GetTypeString(val1))
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
		message := fmt.Sprintf("Error: cannot bitwise and type(s) %s and %s", GetTypeString(val0), GetTypeString(val1))
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
		message := fmt.Sprintf("Error: cannot bitwise or type(s) %s and %s", GetTypeString(val0), GetTypeString(val1))
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
		message := fmt.Sprintf("Error: cannot shift left types: %s and %s", GetTypeString(val0), GetTypeString(val1))
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
		message := fmt.Sprintf("Error: cannot shift right types: %s and %s", GetTypeString(val0), GetTypeString(val1))
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
		message := fmt.Sprintf("Error: cannot bitwise xor type(s) %s and %s", GetTypeString(val0), GetTypeString(val1))
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

func DoCmpGt(interpreter *AtomInterpreter, frame *AtomCallFrame, val0 *AtomValue, val1 *AtomValue) {
	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := fmt.Sprintf("Error: cannot compare greater than type(s) %s and %s", GetTypeString(val0), GetTypeString(val1))
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
		message := fmt.Sprintf("Error: cannot compare greater than or equal to type(s) %s and %s", GetTypeString(val0), GetTypeString(val1))
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

func DoCmpLt(interpreter *AtomInterpreter, frame *AtomCallFrame, val0 *AtomValue, val1 *AtomValue) {
	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := fmt.Sprintf("Error: cannot compare less than type(s) %s and %s", GetTypeString(val0), GetTypeString(val1))
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
		message := fmt.Sprintf("Error: cannot compare less than or equal to type(s) %s and %s", GetTypeString(val0), GetTypeString(val1))
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

func DoCmpNe(interpreter *AtomInterpreter, frame *AtomCallFrame, val0 *AtomValue, val1 *AtomValue) {
	if IsNumberType(val0) && IsNumberType(val1) {
		lhsValue := CoerceToLong(val0)
		rhsValue := CoerceToLong(val1)
		if lhsValue != rhsValue {
			frame.Stack.Push(interpreter.State.TrueValue)
			return
		}
		frame.Stack.Push(interpreter.State.FalseValue)
		return
	}

	if CheckType(val0, AtomTypeStr) && CheckType(val1, AtomTypeStr) {
		lhsStr := val0.Value.(string)
		rhsStr := val1.Value.(string)
		if lhsStr != rhsStr {
			frame.Stack.Push(interpreter.State.TrueValue)
			return
		}
		frame.Stack.Push(interpreter.State.FalseValue)
		return
	}

	if CheckType(val0, AtomTypeNull) && CheckType(val1, AtomTypeNull) {
		frame.Stack.Push(interpreter.State.FalseValue)
		return
	}

	// For different types or other cases, they are not equal
	if val0.Type != val1.Type {
		frame.Stack.Push(interpreter.State.TrueValue)
		return
	}

	// For other types, use simple reference equality for now
	if val0.HashValue() != val1.HashValue() {
		frame.Stack.Push(interpreter.State.TrueValue)
		return
	}

	frame.Stack.Push(interpreter.State.FalseValue)
}
