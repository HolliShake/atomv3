package runtime

import (
	"fmt"
	"math"
)

func DoLoadArray(interpreter *AtomInterpreter, size int) {
	elements := []*AtomValue{}
	for range size {
		elements = append(elements, interpreter.popp())
	}
	interpreter.pushVal(
		NewAtomValueArray(elements),
	)
}

func DoLoadObject(interpreter *AtomInterpreter, size int) {
	elements := map[string]*AtomValue{}
	for range size {
		k := interpreter.popp()
		v := interpreter.popp()
		elements[k.Value.(string)] = v
	}
	interpreter.pushVal(
		NewAtomValueObject(elements),
	)
}

func DoLoadName(interpreter *AtomInterpreter, env *AtomEnv, name string) {
	value, err := env.Lookup(name)
	if err != nil {
		interpreter.pushVal(
			NewAtomValueError(err.Error()),
		)
	} else {
		interpreter.pushVal(value)
	}
}

func DoLoadModule0(interpreter *AtomInterpreter, name string) {
	module := interpreter.ModuleTable[name]
	if module == nil {
		message := fmt.Sprintf("module %s not found", name)
		interpreter.pushVal(NewAtomValueError(message))
		return
	}
	interpreter.pushVal(module)
}

func DoIndex(interpreter *AtomInterpreter, obj *AtomValue, index *AtomValue) {
	if CheckType(obj, AtomTypeStr) {
		if !IsNumberType(index) {
			message := fmt.Sprintf("cannot index type: %s with type: %s", GetTypeString(obj), GetTypeString(index))
			interpreter.pushVal(NewAtomValueError(message))
			return
		}

		r := []rune(obj.Value.(string))
		indexValue := CoerceToLong(index)
		if indexValue < 0 || indexValue >= int64(len(r)) {
			message := fmt.Sprintf("index out of bounds: %d", indexValue)
			interpreter.pushVal(NewAtomValueError(message))
			return
		}

		interpreter.pushVal(NewAtomValueStr(string(r[indexValue])))
		return

	} else if CheckType(obj, AtomTypeArray) {
		if !IsNumberType(index) {
			message := fmt.Sprintf("cannot index type: %s with type: %s", GetTypeString(obj), GetTypeString(index))
			interpreter.pushVal(NewAtomValueError(message))
			return
		}

		array := obj.Value.(*AtomArray)
		indexValue := CoerceToLong(index)

		if !array.ValidIndex(int(indexValue)) {
			message := fmt.Sprintf("index out of bounds: %d", indexValue)
			interpreter.pushVal(NewAtomValueError(message))
			return
		}

		interpreter.pushVal(array.Get(int(indexValue)))
		return

	} else if CheckType(obj, AtomTypeObj) {
		objValue := obj.Value.(*AtomObject)
		indexValue := index.String()

		value := objValue.Get(indexValue)
		if value == nil {
			interpreter.pushVal(interpreter.State.NullValue)
			return
		}

		interpreter.pushVal(value)
		return

	} else if CheckType(obj, AtomTypeClass) {
		class := obj.Value.(*AtomClass)
		if class.Proto.Value.(*AtomObject).Get(index.String()) != nil {
			interpreter.pushVal(class.Proto.Value.(*AtomObject).Get(index.String()))
			return
		}
		interpreter.pushVal(interpreter.State.NullValue)
		return

	} else if CheckType(obj, AtomTypeClassInstance) {
		classInstance := obj.Value.(*AtomClassInstance)
		property := classInstance.Property

		if property.Value.(*AtomObject).Get(index.String()) != nil {
			interpreter.pushVal(property.Value.(*AtomObject).Get(index.String()))
			return
		}
		// Is proto
		current := classInstance.Prototype.Value.(*AtomClass)
		for current != nil {
			if attribute := current.Proto.Value.(*AtomObject).Get(index.String()); attribute != nil {
				if CheckType(attribute, AtomTypeFunc) || CheckType(attribute, AtomTypeNativeFunc) {
					interpreter.pushVal(NewAtomValueMethod(property, attribute))
					return
				}
				interpreter.pushVal(attribute)
				return
			}
			if current.Base == nil {
				break
			}
			current = current.Base.Value.(*AtomClass)
		}
		interpreter.pushVal(interpreter.State.NullValue)
		return

	} else if CheckType(obj, AtomTypeEnum) {
		if !CheckType(index, AtomTypeStr) {
			message := fmt.Sprintf("cannot index type: %s with type: %s", GetTypeString(obj), GetTypeString(index))
			interpreter.pushVal(NewAtomValueError(message))
			return
		}

		enumValue := obj.Value.(*AtomObject)
		indexValue := index.String()
		value := enumValue.Get(indexValue)

		if value == nil {
			interpreter.pushVal(interpreter.State.NullValue)
			return
		}

		interpreter.pushVal(value)
		return

	} else {
		message := fmt.Sprintf("cannot index type: %s", GetTypeString(obj))
		interpreter.pushVal(NewAtomValueError(message))
		return
	}
}

func DoPluckAttribute(interpreter *AtomInterpreter, obj *AtomValue, attribute string) {
	if !CheckType(obj, AtomTypeObj) {
		message := fmt.Sprintf("cannot pluck attribute type: %s", GetTypeString(obj))
		interpreter.pushVal(NewAtomValueError(message))
		return
	}

	objValue := obj.Value.(*AtomObject)

	if value := objValue.Get(attribute); value != nil {
		interpreter.pushVal(value)
		return
	}
	interpreter.pushVal(interpreter.State.NullValue)
}

func DoSetIndex(interpreter *AtomInterpreter, obj *AtomValue, index *AtomValue) {
	cleanupStack := func(size int) {
		for range size {
			interpreter.popp()
		}
	}
	if CheckType(obj, AtomTypeArray) {
		if !IsNumberType(index) {
			cleanupStack(1)
			message := fmt.Sprintf("cannot set index type: %s with type: %s", GetTypeString(obj), GetTypeString(index))
			interpreter.pushVal(NewAtomValueError(message))
			return
		}
		array := obj.Value.(*AtomArray)
		indexValue := CoerceToLong(index)

		if array.Freeze {
			cleanupStack(2)
			message := "cannot set index on frozen array"
			interpreter.pushVal(NewAtomValueError(message))
			return
		}

		if !array.ValidIndex(int(indexValue)) {
			cleanupStack(2)
			message := fmt.Sprintf("index out of bounds: %d", indexValue)
			interpreter.pushVal(NewAtomValueError(message))
			return
		}

		array.Set(int(indexValue), interpreter.popp())
		return

	} else if CheckType(obj, AtomTypeObj) {
		if obj.Value.(*AtomObject).Freeze {
			cleanupStack(2) // includes duplicate obj
			message := "cannot set index on frozen object"
			interpreter.pushVal(NewAtomValueError(message))
			return
		}

		objValue := obj.Value.(*AtomObject)
		indexValue := index.String()
		objValue.Set(indexValue, interpreter.popp())
		return

	} else {
		cleanupStack(2)
		message := fmt.Sprintf("cannot set index type: %s", GetTypeString(obj))
		interpreter.pushVal(NewAtomValueError(message))
		return
	}
}

func DoLoadFunction(interpreter *AtomInterpreter, offset int) {
	fn := interpreter.State.FunctionTable.Get(offset)
	interpreter.pushVal(fn)
}

func DoMakeClass(interpreter *AtomInterpreter, name string, size int) {
	elements := map[string]*AtomValue{}

	for range size {
		k := interpreter.popp()
		v := interpreter.popp()
		elements[k.Value.(string)] = v
	}

	interpreter.pushVal(NewAtomValueClass(
		name,
		nil,
		NewAtomValueObject(elements),
	))
}

func DoExtendClass(interpreter *AtomInterpreter, cls *AtomValue, ext *AtomValue) {
	clsValue := cls.Value.(*AtomClass)
	clsValue.Base = ext
}

func DoMakeEnum(interpreter *AtomInterpreter, env *AtomEnv, size int) {
	elements := map[string]*AtomValue{}
	valueHashes := map[int]bool{}

	for range size {
		k := interpreter.popp()
		v := interpreter.popp()
		key := k.Value.(string)

		valueHash := v.HashValue()
		if valueHashes[valueHash] {
			elements[key] = NewAtomValueError(fmt.Sprintf("duplicate value in enum (%s)", v.String()))
		} else {
			elements[key] = v
			valueHashes[valueHash] = true
		}
	}

	interpreter.pushVal(NewAtomValueEnum(elements))
}

func DoCallConstructor(interpreter *AtomInterpreter, env *AtomEnv, cls *AtomValue, argc int) {
	cleanupStack := func() {
		for range argc {
			interpreter.popp()
		}
	}
	if !CheckType(cls, AtomTypeClass) {
		cleanupStack()
		message := GetTypeString(cls) + " is not a constructor"
		interpreter.pushVal(NewAtomValueError(message))
		return
	}
	atomClass := cls.Value.(*AtomClass)
	properties := atomClass.Proto.Value.(*AtomObject)

	this := NewAtomValueObject(map[string]*AtomValue{})

	// Check if has initializer
	if initializer := properties.Get("init"); initializer == nil {
		interpreter.pushVal(
			NewAtomValueClassInstance(cls, this),
		)
	} else {
		DoCallInit(interpreter, env, cls, initializer, this, 1+argc)
	}
}

func DoCallInit(interpreter *AtomInterpreter, env *AtomEnv, cls *AtomValue, fn *AtomValue, this *AtomValue, argc int) {
	cleanupStack := func() {
		for range argc {
			interpreter.popp()
		}
	}

	// Push this
	interpreter.pushVal(this)

	if CheckType(fn, AtomTypeFunc) {
		code := fn.Value.(*AtomCode)
		if argc != code.Argc {
			cleanupStack()
			message := fmt.Sprintf("Error: argument count mismatch, expected %d, got %d", code.Argc, argc)
			interpreter.pushVal(NewAtomValueError(message))
			return
		}

		newEnv := NewAtomEnv(env)

		interpreter.executeFrame(NewAtomCallFrame(
			fn,
			newEnv,
			0,
		))

		// Pop return
		interpreter.popp()
		interpreter.pushVal(
			NewAtomValueClassInstance(cls, this),
		)

	} else if CheckType(fn, AtomTypeNativeFunc) {
		nativeFunc := fn.Value.(NativeFunc)
		if nativeFunc.Paramc != argc && nativeFunc.Paramc != Variadict {
			cleanupStack()
			message := "Error: argument count mismatch"
			interpreter.pushVal(NewAtomValueError(message))
			return
		}

		nativeFunc.Callable(interpreter, argc)

		// Pop return
		interpreter.popp()
		interpreter.pushVal(
			NewAtomValueClassInstance(cls, this),
		)

	} else {
		cleanupStack()
		message := fmt.Sprintf("Error: %s is not a function", GetTypeString(fn))
		interpreter.pushVal(NewAtomValueError(message))
	}
}

func DoCall(interpreter *AtomInterpreter, env *AtomEnv, fn *AtomValue, argc int) {
	cleanupStack := func() {
		for range argc {
			interpreter.popp()
		}
	}

	if CheckType(fn, AtomTypeMethod) {
		method := fn.Value.(*AtomMethod)

		// Push this
		interpreter.pushVal(method.This)

		fn = method.Fn
		argc++
	}

	if CheckType(fn, AtomTypeFunc) {
		code := fn.Value.(*AtomCode)
		if argc != code.Argc {
			cleanupStack()
			message := fmt.Sprintf("Error: argument count mismatch, expected %d, got %d", code.Argc, argc)
			interpreter.pushVal(NewAtomValueError(message))
			return
		}

		newEnv := NewAtomEnv(env)

		interpreter.executeFrame(NewAtomCallFrame(
			fn,
			newEnv,
			0,
		))

	} else if CheckType(fn, AtomTypeNativeFunc) {
		nativeFunc := fn.Value.(NativeFunc)
		if nativeFunc.Paramc != argc && nativeFunc.Paramc != Variadict {
			cleanupStack()
			message := fmt.Sprintf("Error: argument count mismatch, expected %d, got %d", nativeFunc.Paramc, argc)
			interpreter.pushVal(NewAtomValueError(message))
			return
		}
		nativeFunc.Callable(interpreter, argc)

	} else {
		cleanupStack()
		message := fmt.Sprintf("Error: %s is not a function", GetTypeString(fn))
		interpreter.pushVal(NewAtomValueError(message))
	}
}

func DoNeg(interpreter *AtomInterpreter, val *AtomValue) {
	if !IsNumberType(val) {
		message := fmt.Sprintf("Error: cannot negate type: %s", GetTypeString(val))
		interpreter.pushVal(NewAtomValueError(message))
		return
	}
	interpreter.pushVal(NewAtomValueNum(-CoerceToNum(val)))
}

func DoNot(interpreter *AtomInterpreter, val *AtomValue) {
	if !CoerceToBool(val) {
		interpreter.pushVal(interpreter.State.TrueValue)
		return
	}
	interpreter.pushVal(interpreter.State.FalseValue)
}

func DoPos(interpreter *AtomInterpreter, val *AtomValue) {
	if !IsNumberType(val) {
		message := fmt.Sprintf("Error: cannot pos type: %s", GetTypeString(val))
		interpreter.pushVal(NewAtomValueError(message))
		return
	}
	interpreter.pushVal(NewAtomValueNum(CoerceToNum(val)))
}

func DoAddition(interpreter *AtomInterpreter, val0 *AtomValue, val1 *AtomValue) {
	// Fast path for integers
	if CheckType(val0, AtomTypeInt) && CheckType(val1, AtomTypeInt) {
		// Use XOR trick to detect overflow
		a := CoerceToInt(val0)
		b := CoerceToInt(val1)
		sum := a + b
		if ((a ^ sum) & (b ^ sum)) < 0 {
			// Overflow occurred, promote to double
			interpreter.pushVal(NewAtomValueNum(float64(a) + float64(b)))
			return
		}
		interpreter.pushVal(NewAtomValueInt(int(sum)))
		return
	}

	// Fast path for strings
	if CheckType(val0, AtomTypeStr) && CheckType(val1, AtomTypeStr) {
		lhs := val0.Value.(string)
		rhs := val1.Value.(string)
		result := lhs + rhs
		interpreter.pushVal(NewAtomValueStr(result))
		return
	}

	// Check if both values are numbers (int or float)
	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := fmt.Sprintf("Error: cannot add types: %s and %s", GetTypeString(val0), GetTypeString(val1))
		interpreter.pushVal(NewAtomValueError(message))
		return
	}

	// Fallback path using coercion
	lhsValue := CoerceToNum(val0)
	rhsValue := CoerceToNum(val1)
	result := lhsValue + rhsValue

	// Try to preserve integer types if possible
	if IsInteger(result) && result <= math.MaxInt32 && result >= math.MinInt32 {
		interpreter.pushVal(NewAtomValueInt(int(result)))
		return
	}
	interpreter.pushVal(NewAtomValueNum(result))
}

func DoDivision(interpreter *AtomInterpreter, val0 *AtomValue, val1 *AtomValue) {
	// Fast path for integers
	if CheckType(val0, AtomTypeInt) && CheckType(val1, AtomTypeInt) {
		a := CoerceToInt(val0)
		b := CoerceToInt(val1)
		if b == 0 {
			message := "Error: division by zero"
			interpreter.pushVal(NewAtomValueError(message))
			return
		}
		result := a / b
		interpreter.pushVal(NewAtomValueInt(int(result)))
		return
	}

	// Check if both values are numbers (int or float)
	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := fmt.Sprintf("Error: cannot divide types: %s and %s", GetTypeString(val0), GetTypeString(val1))
		interpreter.pushVal(NewAtomValueError(message))
		return
	}

	// Fallback path using coercion
	lhsValue := CoerceToNum(val0)
	rhsValue := CoerceToNum(val1)
	if rhsValue == 0 {
		message := "division by zero"
		interpreter.pushVal(NewAtomValueError(message))
		return
	}
	result := lhsValue / rhsValue

	// Try to preserve integer types if possible
	if IsInteger(result) && result <= math.MaxInt32 && result >= math.MinInt32 {
		interpreter.pushVal(NewAtomValueInt(int(result)))
		return
	}
	interpreter.pushVal(NewAtomValueNum(result))
}

func DoModulus(interpreter *AtomInterpreter, val0 *AtomValue, val1 *AtomValue) {
	// Fast path for integers
	if CheckType(val0, AtomTypeInt) && CheckType(val1, AtomTypeInt) {
		a := CoerceToInt(val0)
		b := CoerceToInt(val1)
		if b == 0 {
			message := "Error: division by zero"
			interpreter.pushVal(NewAtomValueError(message))
			return
		}
		result := a % b
		interpreter.pushVal(NewAtomValueInt(int(result)))
		return
	}

	// Check if both values are numbers (int or float)
	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := fmt.Sprintf("Error: cannot modulo types: %s and %s", GetTypeString(val0), GetTypeString(val1))
		interpreter.pushVal(NewAtomValueError(message))
		return
	}

	// Fallback path using coercion
	lhsValue := CoerceToNum(val0)
	rhsValue := CoerceToNum(val1)
	if rhsValue == 0 {
		message := "Error: division by zero"
		interpreter.pushVal(NewAtomValueError(message))
		return
	}
	result := math.Mod(lhsValue, rhsValue)

	// Try to preserve integer types if possible
	if IsInteger(result) && result <= math.MaxInt32 && result >= math.MinInt32 {
		interpreter.pushVal(NewAtomValueInt(int(result)))
		return
	}
	interpreter.pushVal(NewAtomValueNum(result))
}

func DoMultiplication(interpreter *AtomInterpreter, val0 *AtomValue, val1 *AtomValue) {
	// Fast path for integers
	if CheckType(val0, AtomTypeInt) && CheckType(val1, AtomTypeInt) {
		lhs := CoerceToInt(val0)
		rhs := CoerceToInt(val1)

		// Check for overflow using int64 arithmetic
		result := int64(lhs) * int64(rhs)
		if result >= math.MinInt32 && result <= math.MaxInt32 {
			interpreter.pushVal(NewAtomValueInt(int(result)))
			return
		}
		// Overflow occurred, promote to float64
		interpreter.pushVal(NewAtomValueNum(float64(result)))
		return
	}

	// Check if both values are numbers (int or float)
	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := fmt.Sprintf("Error: cannot multiply types: %s and %s", GetTypeString(val0), GetTypeString(val1))
		interpreter.pushVal(NewAtomValueError(message))
		return
	}

	// Fallback path using coercion
	lhsValue := CoerceToNum(val0)
	rhsValue := CoerceToNum(val1)
	result := lhsValue * rhsValue

	// Try to preserve integer types if possible
	if IsInteger(result) && result <= math.MaxInt32 && result >= math.MinInt32 {
		interpreter.pushVal(NewAtomValueInt(int(result)))
		return
	}
	interpreter.pushVal(NewAtomValueNum(result))
}

func DoSubtraction(interpreter *AtomInterpreter, val0 *AtomValue, val1 *AtomValue) {
	// Fast path for integers
	if CheckType(val0, AtomTypeInt) && CheckType(val1, AtomTypeInt) {
		a := CoerceToInt(val0)
		b := CoerceToInt(val1)
		diff := a - b
		if ((a ^ b) & (a ^ diff)) < 0 {
			// Overflow occurred, promote to double
			interpreter.pushVal(NewAtomValueNum(float64(a) - float64(b)))
			return
		}
		interpreter.pushVal(NewAtomValueInt(int(diff)))
		return
	}

	// Check if both values are numbers (int or float)
	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := fmt.Sprintf("Error: cannot subtract types: %s and %s", GetTypeString(val0), GetTypeString(val1))
		interpreter.pushVal(NewAtomValueError(message))
		return
	}

	// Fallback path using coercion
	lhsValue := CoerceToNum(val0)
	rhsValue := CoerceToNum(val1)
	result := lhsValue - rhsValue

	// Try to preserve integer types if possible
	if IsInteger(result) && result <= math.MaxInt32 && result >= math.MinInt32 {
		interpreter.pushVal(NewAtomValueInt(int(result)))
		return
	}
	interpreter.pushVal(NewAtomValueNum(result))
}

func DoAnd(interpreter *AtomInterpreter, val0 *AtomValue, val1 *AtomValue) {
	// Fast path for integers
	if CheckType(val0, AtomTypeInt) && CheckType(val1, AtomTypeInt) {
		a := val0.Value.(int32)
		b := val1.Value.(int32)
		result := a & b
		interpreter.pushVal(NewAtomValueInt(int(result)))
		return
	}

	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := fmt.Sprintf("Error: cannot bitwise and type(s) %s and %s", GetTypeString(val0), GetTypeString(val1))
		interpreter.pushVal(NewAtomValueError(message))
		return
	}

	lhsValue := CoerceToLong(val0)
	rhsValue := CoerceToLong(val1)
	result := lhsValue & rhsValue

	// Check if result can be represented as an int
	if result >= math.MinInt32 && result <= math.MaxInt32 {
		interpreter.pushVal(NewAtomValueInt(int(result)))
	} else {
		interpreter.pushVal(NewAtomValueNum(float64(result)))
	}
}

func DoOr(interpreter *AtomInterpreter, val0 *AtomValue, val1 *AtomValue) {
	// Fast path for integers
	if CheckType(val0, AtomTypeInt) && CheckType(val1, AtomTypeInt) {
		a := val0.Value.(int32)
		b := val1.Value.(int32)
		result := a | b
		interpreter.pushVal(NewAtomValueInt(int(result)))
		return
	}

	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := fmt.Sprintf("Error: cannot bitwise or type(s) %s and %s", GetTypeString(val0), GetTypeString(val1))
		interpreter.pushVal(NewAtomValueError(message))
		return
	}

	lhsValue := CoerceToLong(val0)
	rhsValue := CoerceToLong(val1)
	result := lhsValue | rhsValue

	// Check if result can be represented as an int
	if result >= math.MinInt32 && result <= math.MaxInt32 {
		interpreter.pushVal(NewAtomValueInt(int(result)))
	} else {
		interpreter.pushVal(NewAtomValueNum(float64(result)))
	}
}

func DoShiftLeft(interpreter *AtomInterpreter, val0 *AtomValue, val1 *AtomValue) {
	// Fast path for integers
	if CheckType(val0, AtomTypeInt) && CheckType(val1, AtomTypeInt) {
		a := CoerceToInt(val0)
		b := CoerceToInt(val1)
		result := a << b
		interpreter.pushVal(NewAtomValueInt(int(result)))
		return
	}

	// Check if both values are numbers (int or float)
	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := fmt.Sprintf("Error: cannot shift left types: %s and %s", GetTypeString(val0), GetTypeString(val1))
		interpreter.pushVal(NewAtomValueError(message))
		return
	}

	// Fallback path using coercion
	lhsValue := CoerceToNum(val0)
	rhsValue := CoerceToNum(val1)
	result := int64(lhsValue) << int64(rhsValue)

	// Check if result can be represented as an int
	if result >= math.MinInt32 && result <= math.MaxInt32 {
		interpreter.pushVal(NewAtomValueInt(int(result)))
		return
	}
	interpreter.pushVal(NewAtomValueNum(float64(result)))
}

func DoShiftRight(interpreter *AtomInterpreter, val0 *AtomValue, val1 *AtomValue) {
	// Fast path for integers
	if CheckType(val0, AtomTypeInt) && CheckType(val1, AtomTypeInt) {
		a := CoerceToInt(val0)
		b := CoerceToInt(val1)
		result := a >> b
		interpreter.pushVal(NewAtomValueInt(int(result)))
		return
	}

	// Check if both values are numbers (int or float)
	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := fmt.Sprintf("Error: cannot shift right types: %s and %s", GetTypeString(val0), GetTypeString(val1))
		interpreter.pushVal(NewAtomValueError(message))
		return
	}

	// Fallback path using coercion
	lhsValue := CoerceToNum(val0)
	rhsValue := CoerceToNum(val1)
	result := int64(lhsValue) >> int64(rhsValue)

	// Try to preserve integer types if possible
	if result >= math.MinInt32 && result <= math.MaxInt32 {
		interpreter.pushVal(NewAtomValueInt(int(result)))
		return
	}
	interpreter.pushVal(NewAtomValueNum(float64(result)))
}

func DoXor(interpreter *AtomInterpreter, val0 *AtomValue, val1 *AtomValue) {
	// Fast path for integers
	if CheckType(val0, AtomTypeInt) && CheckType(val1, AtomTypeInt) {
		a := val0.Value.(int32)
		b := val1.Value.(int32)
		result := a ^ b
		interpreter.pushVal(NewAtomValueInt(int(result)))
		return
	}

	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := fmt.Sprintf("Error: cannot bitwise xor type(s) %s and %s", GetTypeString(val0), GetTypeString(val1))
		interpreter.pushVal(NewAtomValueError(message))
		return
	}

	lhsValue := CoerceToLong(val0)
	rhsValue := CoerceToLong(val1)
	result := lhsValue ^ rhsValue

	// Check if result can be represented as an int
	if result >= math.MinInt32 && result <= math.MaxInt32 {
		interpreter.pushVal(NewAtomValueInt(int(result)))
	} else {
		interpreter.pushVal(NewAtomValueNum(float64(result)))
	}
}

func DoCmpEq(interpreter *AtomInterpreter, val0 *AtomValue, val1 *AtomValue) {
	if IsNumberType(val0) && IsNumberType(val1) {
		lhsValue := CoerceToLong(val0)
		rhsValue := CoerceToLong(val1)
		if lhsValue == rhsValue {
			interpreter.pushVal(interpreter.State.TrueValue)
			return
		}
		interpreter.pushVal(interpreter.State.FalseValue)
		return
	}

	if CheckType(val0, AtomTypeStr) && CheckType(val1, AtomTypeStr) {
		lhsStr := val0.Value.(string)
		rhsStr := val1.Value.(string)
		if lhsStr == rhsStr {
			interpreter.pushVal(interpreter.State.TrueValue)
			return
		}
		interpreter.pushVal(interpreter.State.FalseValue)
		return
	}

	if CheckType(val0, AtomTypeNull) && CheckType(val1, AtomTypeNull) {
		interpreter.pushVal(interpreter.State.TrueValue)
		return
	}

	// For other types, use simple reference equality for now
	if val0.HashValue() == val1.HashValue() {
		interpreter.pushVal(interpreter.State.TrueValue)
		return
	}

	interpreter.pushVal(interpreter.State.FalseValue)
}

func DoCmpGt(interpreter *AtomInterpreter, val0 *AtomValue, val1 *AtomValue) {
	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := fmt.Sprintf("Error: cannot compare greater than type(s) %s and %s", GetTypeString(val0), GetTypeString(val1))
		interpreter.pushVal(NewAtomValueError(message))
		return
	}

	// Coerce to long to avoid floating point comparisons
	lhsValue := CoerceToLong(val0)
	rhsValue := CoerceToLong(val1)

	// Compare the long values
	if lhsValue > rhsValue {
		interpreter.pushVal(interpreter.State.TrueValue)
		return
	}
	interpreter.pushVal(interpreter.State.FalseValue)
}

func DoCmpGte(interpreter *AtomInterpreter, val0 *AtomValue, val1 *AtomValue) {
	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := fmt.Sprintf("Error: cannot compare greater than or equal to type(s) %s and %s", GetTypeString(val0), GetTypeString(val1))
		interpreter.pushVal(NewAtomValueError(message))
		return
	}

	// Coerce to long to avoid floating point comparisons
	lhsValue := CoerceToLong(val0)
	rhsValue := CoerceToLong(val1)

	// Compare the long values
	if lhsValue >= rhsValue {
		interpreter.pushVal(interpreter.State.TrueValue)
		return
	}
	interpreter.pushVal(interpreter.State.FalseValue)
}

func DoCmpLt(interpreter *AtomInterpreter, val0 *AtomValue, val1 *AtomValue) {
	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := fmt.Sprintf("Error: cannot compare less than type(s) %s and %s", GetTypeString(val0), GetTypeString(val1))
		interpreter.pushVal(NewAtomValueError(message))
		return
	}

	// Coerce to long to avoid floating point comparisons
	lhsValue := CoerceToLong(val0)
	rhsValue := CoerceToLong(val1)

	// Compare the long values
	if lhsValue < rhsValue {
		interpreter.pushVal(interpreter.State.TrueValue)
		return
	}
	interpreter.pushVal(interpreter.State.FalseValue)
}

func DoCmpLte(interpreter *AtomInterpreter, val0 *AtomValue, val1 *AtomValue) {
	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := fmt.Sprintf("Error: cannot compare less than or equal to type(s) %s and %s", GetTypeString(val0), GetTypeString(val1))
		interpreter.pushVal(NewAtomValueError(message))
		return
	}

	// Coerce to long to avoid floating point comparisons
	lhsValue := CoerceToLong(val0)
	rhsValue := CoerceToLong(val1)

	// Compare the long values
	if lhsValue <= rhsValue {
		interpreter.pushVal(interpreter.State.TrueValue)
		return
	}
	interpreter.pushVal(interpreter.State.FalseValue)
}

func DoCmpNe(interpreter *AtomInterpreter, val0 *AtomValue, val1 *AtomValue) {
	if IsNumberType(val0) && IsNumberType(val1) {
		lhsValue := CoerceToLong(val0)
		rhsValue := CoerceToLong(val1)
		if lhsValue != rhsValue {
			interpreter.pushVal(interpreter.State.TrueValue)
			return
		}
		interpreter.pushVal(interpreter.State.FalseValue)
		return
	}

	if CheckType(val0, AtomTypeStr) && CheckType(val1, AtomTypeStr) {
		lhsStr := val0.Value.(string)
		rhsStr := val1.Value.(string)
		if lhsStr != rhsStr {
			interpreter.pushVal(interpreter.State.TrueValue)
			return
		}
		interpreter.pushVal(interpreter.State.FalseValue)
		return
	}

	if CheckType(val0, AtomTypeNull) || CheckType(val1, AtomTypeNull) {
		interpreter.pushVal(interpreter.State.FalseValue)
		return
	}

	// For other types, use simple reference equality for now
	if val0.HashValue() != val1.HashValue() {
		interpreter.pushVal(interpreter.State.TrueValue)
		return
	}

	interpreter.pushVal(interpreter.State.FalseValue)
}
