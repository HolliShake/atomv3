package runtime

import (
	"fmt"
	"math"
)

func DoLoadModule0(intereter *AtomInterpreter, name string) {
	module := intereter.ModuleTable[name]
	if module == nil {
		message := fmt.Sprintf("module %s not found", name)
		intereter.pushVal(NewAtomValueError(message))
		return
	}
	intereter.pushRef(module)
}

func DoIndex(intereter *AtomInterpreter, obj *AtomValue, index *AtomValue) {
	if CheckType(obj, AtomTypeArray) {
		if !IsNumberType(index) {
			message := fmt.Sprintf("cannot index type: %s with type: %s", GetTypeString(obj), GetTypeString(index))
			intereter.pushVal(NewAtomValueError(message))
			return
		}

		array := obj.Value.(*AtomArray)
		indexValue := CoerceToLong(index)

		if !array.ValidIndex(int(indexValue)) {
			message := fmt.Sprintf("index out of bounds: %d", indexValue)
			intereter.pushVal(NewAtomValueError(message))
			return
		}

		intereter.pushVal(array.Get(int(indexValue)))
		return

	} else if CheckType(obj, AtomTypeObj) {
		objValue := obj.Value.(*AtomObject)
		indexValue := index.String()

		value := objValue.Get(indexValue)
		if value == nil {
			intereter.pushRef(intereter.State.NullValue)
			return
		}

		intereter.pushVal(value)
		return

	} else {
		message := fmt.Sprintf("cannot index type: %s", GetTypeString(obj))
		intereter.pushVal(NewAtomValueError(message))
		return
	}
}

func DoPluckAttribute(intereter *AtomInterpreter, obj *AtomValue, attribute string) {
	if !CheckType(obj, AtomTypeObj) {
		message := fmt.Sprintf("cannot pluck attribute type: %s", GetTypeString(obj))
		intereter.pushVal(NewAtomValueError(message))
		return
	}

	objValue := obj.Value.(*AtomObject)

	if value := objValue.Get(attribute); value != nil {
		intereter.pushVal(value)
		return
	}
	intereter.pushRef(intereter.State.NullValue)
}

func DoCall(intereter *AtomInterpreter, funcValue *AtomValue, argc int) {
	cleanupStack := func() {
		for range argc {
			intereter.pop()
		}
	}
	if CheckType(funcValue, AtomTypeFunc) {
		code := funcValue.Value.(*AtomCode)
		if argc != code.Argc {
			cleanupStack()
			message := "argument count mismatch"
			intereter.pushVal(NewAtomValueError(message))
			return
		}
		intereter.executeFrame(funcValue, 0)

	} else if CheckType(funcValue, AtomTypeNativeFunc) {
		nativeFunc := funcValue.Value.(NativeFunc)
		if nativeFunc.Paramc != argc && nativeFunc.Paramc != Variadict {
			cleanupStack()
			message := "argument count mismatch"
			intereter.pushVal(NewAtomValueError(message))
			return
		}
		nativeFunc.Callable(intereter, argc)

	} else {
		cleanupStack()
		message := "not a function " + GetTypeString(funcValue)
		intereter.pushVal(NewAtomValueError(message))
	}
}

func DoNot(intereter *AtomInterpreter, val *AtomValue) {
	if !CoerceToBool(val) {
		intereter.pushRef(intereter.State.TrueValue)
		return
	}
	intereter.pushRef(intereter.State.FalseValue)
}

func DoPos(intereter *AtomInterpreter, val *AtomValue) {
	if !IsNumberType(val) {
		message := fmt.Sprintf("cannot pos type: %s", GetTypeString(val))
		intereter.pushVal(NewAtomValueError(message))
		return
	}
	intereter.pushVal(NewAtomValueNum(CoerceToNum(val)))
}

func DoNeg(intereter *AtomInterpreter, val *AtomValue) {
	if !IsNumberType(val) {
		message := fmt.Sprintf("cannot negate type: %s", GetTypeString(val))
		intereter.pushVal(NewAtomValueError(message))
		return
	}
	intereter.pushVal(NewAtomValueNum(-CoerceToNum(val)))
}

func DoMultiplication(intereter *AtomInterpreter, val0 *AtomValue, val1 *AtomValue) {
	// Fast path for integers
	if CheckType(val0, AtomTypeInt) && CheckType(val1, AtomTypeInt) {
		lhs := CoerceToInt(val0)
		rhs := CoerceToInt(val1)

		// Check for overflow using int64 arithmetic
		result := int64(lhs) * int64(rhs)
		if result >= math.MinInt32 && result <= math.MaxInt32 {
			intereter.pushVal(NewAtomValueInt(int(result)))
			return
		}
		// Overflow occurred, promote to float64
		intereter.pushVal(NewAtomValueNum(float64(result)))
		return
	}

	// Check if both values are numbers (int or float)
	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := fmt.Sprintf("cannot multiply types: %s and %s", GetTypeString(val0), GetTypeString(val1))
		intereter.pushVal(NewAtomValueError(message))
		return
	}

	// Fallback path using coercion
	lhsValue := CoerceToNum(val0)
	rhsValue := CoerceToNum(val1)
	result := lhsValue * rhsValue

	// Try to preserve integer types if possible
	if IsInteger(result) && result <= math.MaxInt32 && result >= math.MinInt32 {
		intereter.pushVal(NewAtomValueInt(int(result)))
		return
	}
	intereter.pushVal(NewAtomValueNum(result))
}

func DoDivision(intereter *AtomInterpreter, val0 *AtomValue, val1 *AtomValue) {
	// Fast path for integers
	if CheckType(val0, AtomTypeInt) && CheckType(val1, AtomTypeInt) {
		a := CoerceToInt(val0)
		b := CoerceToInt(val1)
		if b == 0 {
			message := "division by zero"
			intereter.pushVal(NewAtomValueError(message))
			return
		}
		result := a / b
		intereter.pushVal(NewAtomValueInt(int(result)))
		return
	}

	// Check if both values are numbers (int or float)
	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := fmt.Sprintf("cannot divide types: %s and %s", GetTypeString(val0), GetTypeString(val1))
		intereter.pushVal(NewAtomValueError(message))
		return
	}

	// Fallback path using coercion
	lhsValue := CoerceToNum(val0)
	rhsValue := CoerceToNum(val1)
	if rhsValue == 0 {
		message := "division by zero"
		intereter.pushVal(NewAtomValueError(message))
		return
	}
	result := lhsValue / rhsValue

	// Try to preserve integer types if possible
	if IsInteger(result) && result <= math.MaxInt32 && result >= math.MinInt32 {
		intereter.pushVal(NewAtomValueInt(int(result)))
		return
	}
	intereter.pushVal(NewAtomValueNum(result))
}

func DoModulus(intereter *AtomInterpreter, val0 *AtomValue, val1 *AtomValue) {
	// Fast path for integers
	if CheckType(val0, AtomTypeInt) && CheckType(val1, AtomTypeInt) {
		a := CoerceToInt(val0)
		b := CoerceToInt(val1)
		if b == 0 {
			message := "division by zero"
			intereter.pushVal(NewAtomValueError(message))
			return
		}
		result := a % b
		intereter.pushVal(NewAtomValueInt(int(result)))
		return
	}

	// Check if both values are numbers (int or float)
	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := fmt.Sprintf("cannot modulo types: %s and %s", GetTypeString(val0), GetTypeString(val1))
		intereter.pushVal(NewAtomValueError(message))
		return
	}

	// Fallback path using coercion
	lhsValue := CoerceToNum(val0)
	rhsValue := CoerceToNum(val1)
	if rhsValue == 0 {
		message := "division by zero"
		intereter.pushVal(NewAtomValueError(message))
		return
	}
	result := math.Mod(lhsValue, rhsValue)

	// Try to preserve integer types if possible
	if IsInteger(result) && result <= math.MaxInt32 && result >= math.MinInt32 {
		intereter.pushVal(NewAtomValueInt(int(result)))
		return
	}
	intereter.pushVal(NewAtomValueNum(result))
}

func DoAddition(intereter *AtomInterpreter, val0 *AtomValue, val1 *AtomValue) {
	// Fast path for integers
	if CheckType(val0, AtomTypeInt) && CheckType(val1, AtomTypeInt) {
		// Use XOR trick to detect overflow
		a := CoerceToInt(val0)
		b := CoerceToInt(val1)
		sum := a + b
		if ((a ^ sum) & (b ^ sum)) < 0 {
			// Overflow occurred, promote to double
			intereter.pushVal(NewAtomValueNum(float64(a) + float64(b)))
			return
		}
		intereter.pushVal(NewAtomValueInt(int(sum)))
		return
	}

	// Fast path for strings
	if CheckType(val0, AtomTypeStr) && CheckType(val1, AtomTypeStr) {
		lhs := val0.Value.(string)
		rhs := val1.Value.(string)
		result := lhs + rhs
		intereter.pushVal(NewAtomValueStr(result))
		return
	}

	// Check if both values are numbers (int or float)
	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := fmt.Sprintf("cannot add types: %s and %s", GetTypeString(val0), GetTypeString(val1))
		intereter.pushVal(NewAtomValueError(message))
		return
	}

	// Fallback path using coercion
	lhsValue := CoerceToNum(val0)
	rhsValue := CoerceToNum(val1)
	result := lhsValue + rhsValue

	// Try to preserve integer types if possible
	if IsInteger(result) && result <= math.MaxInt32 && result >= math.MinInt32 {
		intereter.pushVal(NewAtomValueInt(int(result)))
		return
	}
	intereter.pushVal(NewAtomValueNum(result))
}

func DoSubtraction(intereter *AtomInterpreter, val0 *AtomValue, val1 *AtomValue) {
	// Fast path for integers
	if CheckType(val0, AtomTypeInt) && CheckType(val1, AtomTypeInt) {
		a := CoerceToInt(val0)
		b := CoerceToInt(val1)
		diff := a - b
		if ((a ^ b) & (a ^ diff)) < 0 {
			// Overflow occurred, promote to double
			intereter.pushVal(NewAtomValueNum(float64(a) - float64(b)))
			return
		}
		intereter.pushVal(NewAtomValueInt(int(diff)))
		return
	}

	// Check if both values are numbers (int or float)
	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := fmt.Sprintf("cannot subtract types: %s and %s", GetTypeString(val0), GetTypeString(val1))
		intereter.pushVal(NewAtomValueError(message))
		return
	}

	// Fallback path using coercion
	lhsValue := CoerceToNum(val0)
	rhsValue := CoerceToNum(val1)
	result := lhsValue - rhsValue

	// Try to preserve integer types if possible
	if IsInteger(result) && result <= math.MaxInt32 && result >= math.MinInt32 {
		intereter.pushVal(NewAtomValueInt(int(result)))
		return
	}
	intereter.pushVal(NewAtomValueNum(result))
}

func DoShiftLeft(intereter *AtomInterpreter, val0 *AtomValue, val1 *AtomValue) {
	// Fast path for integers
	if CheckType(val0, AtomTypeInt) && CheckType(val1, AtomTypeInt) {
		a := CoerceToInt(val0)
		b := CoerceToInt(val1)
		result := a << b
		intereter.pushVal(NewAtomValueInt(int(result)))
		return
	}

	// Check if both values are numbers (int or float)
	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := fmt.Sprintf("cannot shift left types: %s and %s", GetTypeString(val0), GetTypeString(val1))
		intereter.pushVal(NewAtomValueError(message))
		return
	}

	// Fallback path using coercion
	lhsValue := CoerceToNum(val0)
	rhsValue := CoerceToNum(val1)
	result := int64(lhsValue) << int64(rhsValue)

	// Check if result can be represented as an int
	if result >= math.MinInt32 && result <= math.MaxInt32 {
		intereter.pushVal(NewAtomValueInt(int(result)))
		return
	}
	intereter.pushVal(NewAtomValueNum(float64(result)))
}

func DoShiftRight(intereter *AtomInterpreter, val0 *AtomValue, val1 *AtomValue) {
	// Fast path for integers
	if CheckType(val0, AtomTypeInt) && CheckType(val1, AtomTypeInt) {
		a := CoerceToInt(val0)
		b := CoerceToInt(val1)
		result := a >> b
		intereter.pushVal(NewAtomValueInt(int(result)))
		return
	}

	// Check if both values are numbers (int or float)
	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := fmt.Sprintf("cannot shift right types: %s and %s", GetTypeString(val0), GetTypeString(val1))
		intereter.pushVal(NewAtomValueError(message))
		return
	}

	// Fallback path using coercion
	lhsValue := CoerceToNum(val0)
	rhsValue := CoerceToNum(val1)
	result := int64(lhsValue) >> int64(rhsValue)

	// Try to preserve integer types if possible
	if result >= math.MinInt32 && result <= math.MaxInt32 {
		intereter.pushVal(NewAtomValueInt(int(result)))
		return
	}
	intereter.pushVal(NewAtomValueNum(float64(result)))
}

// Comparison operations

func DoCmpLt(intereter *AtomInterpreter, val0 *AtomValue, val1 *AtomValue) {
	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := fmt.Sprintf("cannot compare less than type(s) %s and %s", GetTypeString(val0), GetTypeString(val1))
		intereter.pushVal(NewAtomValueError(message))
		return
	}

	// Coerce to long to avoid floating point comparisons
	lhsValue := CoerceToLong(val0)
	rhsValue := CoerceToLong(val1)

	// Compare the long values
	if lhsValue < rhsValue {
		intereter.pushRef(intereter.State.TrueValue)
		return
	}
	intereter.pushRef(intereter.State.FalseValue)
}

func DoCmpLte(intereter *AtomInterpreter, val0 *AtomValue, val1 *AtomValue) {
	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := fmt.Sprintf("cannot compare less than or equal to type(s) %s and %s", GetTypeString(val0), GetTypeString(val1))
		intereter.pushVal(NewAtomValueError(message))
		return
	}

	// Coerce to long to avoid floating point comparisons
	lhsValue := CoerceToLong(val0)
	rhsValue := CoerceToLong(val1)

	// Compare the long values
	if lhsValue <= rhsValue {
		intereter.pushRef(intereter.State.TrueValue)
		return
	}
	intereter.pushRef(intereter.State.FalseValue)
}

func DoCmpGt(intereter *AtomInterpreter, val0 *AtomValue, val1 *AtomValue) {
	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := fmt.Sprintf("cannot compare greater than type(s) %s and %s", GetTypeString(val0), GetTypeString(val1))
		intereter.pushVal(NewAtomValueError(message))
		return
	}

	// Coerce to long to avoid floating point comparisons
	lhsValue := CoerceToLong(val0)
	rhsValue := CoerceToLong(val1)

	// Compare the long values
	if lhsValue > rhsValue {
		intereter.pushRef(intereter.State.TrueValue)
		return
	}
	intereter.pushRef(intereter.State.FalseValue)
}

func DoCmpGte(intereter *AtomInterpreter, val0 *AtomValue, val1 *AtomValue) {
	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := fmt.Sprintf("cannot compare greater than or equal to type(s) %s and %s", GetTypeString(val0), GetTypeString(val1))
		intereter.pushVal(NewAtomValueError(message))
		return
	}

	// Coerce to long to avoid floating point comparisons
	lhsValue := CoerceToLong(val0)
	rhsValue := CoerceToLong(val1)

	// Compare the long values
	if lhsValue >= rhsValue {
		intereter.pushRef(intereter.State.TrueValue)
		return
	}
	intereter.pushRef(intereter.State.FalseValue)
}

func DoCmpEq(intereter *AtomInterpreter, val0 *AtomValue, val1 *AtomValue) {
	if IsNumberType(val0) && IsNumberType(val1) {
		lhsValue := CoerceToLong(val0)
		rhsValue := CoerceToLong(val1)
		if lhsValue == rhsValue {
			intereter.pushRef(intereter.State.TrueValue)
			return
		}
		intereter.pushRef(intereter.State.FalseValue)
		return
	}

	if CheckType(val0, AtomTypeStr) && CheckType(val1, AtomTypeStr) {
		lhsStr := val0.Value.(string)
		rhsStr := val1.Value.(string)
		if lhsStr == rhsStr {
			intereter.pushRef(intereter.State.TrueValue)
			return
		}
		intereter.pushRef(intereter.State.FalseValue)
		return
	}

	if CheckType(val0, AtomTypeNull) && CheckType(val1, AtomTypeNull) {
		intereter.pushRef(intereter.State.TrueValue)
		return
	}

	// For other types, use simple reference equality for now
	if val0 == val1 {
		intereter.pushRef(intereter.State.TrueValue)
		return
	}

	intereter.pushRef(intereter.State.FalseValue)
}

func DoCmpNe(intereter *AtomInterpreter, val0 *AtomValue, val1 *AtomValue) {
	if IsNumberType(val0) && IsNumberType(val1) {
		lhsValue := CoerceToLong(val0)
		rhsValue := CoerceToLong(val1)
		if lhsValue != rhsValue {
			intereter.pushRef(intereter.State.TrueValue)
			return
		}
		intereter.pushRef(intereter.State.FalseValue)
		return
	}

	if CheckType(val0, AtomTypeStr) && CheckType(val1, AtomTypeStr) {
		lhsStr := val0.Value.(string)
		rhsStr := val1.Value.(string)
		if lhsStr != rhsStr {
			intereter.pushRef(intereter.State.TrueValue)
			return
		}
		intereter.pushRef(intereter.State.FalseValue)
		return
	}

	if CheckType(val0, AtomTypeNull) || CheckType(val1, AtomTypeNull) {
		intereter.pushRef(intereter.State.FalseValue)
		return
	}

	// For other types, use simple reference equality for now
	if val0 != val1 {
		intereter.pushRef(intereter.State.TrueValue)
		return
	}

	intereter.pushRef(intereter.State.FalseValue)
}

// Bitwise operations

func DoAnd(intereter *AtomInterpreter, val0 *AtomValue, val1 *AtomValue) {
	// Fast path for integers
	if CheckType(val0, AtomTypeInt) && CheckType(val1, AtomTypeInt) {
		a := val0.Value.(int32)
		b := val1.Value.(int32)
		result := a & b
		intereter.pushVal(NewAtomValueInt(int(result)))
		return
	}

	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := fmt.Sprintf("cannot bitwise and type(s) %s and %s", GetTypeString(val0), GetTypeString(val1))
		intereter.pushVal(NewAtomValueError(message))
		return
	}

	lhsValue := CoerceToLong(val0)
	rhsValue := CoerceToLong(val1)
	result := lhsValue & rhsValue

	// Check if result can be represented as an int
	if result >= math.MinInt32 && result <= math.MaxInt32 {
		intereter.pushVal(NewAtomValueInt(int(result)))
	} else {
		intereter.pushVal(NewAtomValueNum(float64(result)))
	}
}

func DoOr(intereter *AtomInterpreter, val0 *AtomValue, val1 *AtomValue) {
	// Fast path for integers
	if CheckType(val0, AtomTypeInt) && CheckType(val1, AtomTypeInt) {
		a := val0.Value.(int32)
		b := val1.Value.(int32)
		result := a | b
		intereter.pushVal(NewAtomValueInt(int(result)))
		return
	}

	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := fmt.Sprintf("cannot bitwise or type(s) %s and %s", GetTypeString(val0), GetTypeString(val1))
		intereter.pushVal(NewAtomValueError(message))
		return
	}

	lhsValue := CoerceToLong(val0)
	rhsValue := CoerceToLong(val1)
	result := lhsValue | rhsValue

	// Check if result can be represented as an int
	if result >= math.MinInt32 && result <= math.MaxInt32 {
		intereter.pushVal(NewAtomValueInt(int(result)))
	} else {
		intereter.pushVal(NewAtomValueNum(float64(result)))
	}
}

func DoXor(intereter *AtomInterpreter, val0 *AtomValue, val1 *AtomValue) {
	// Fast path for integers
	if CheckType(val0, AtomTypeInt) && CheckType(val1, AtomTypeInt) {
		a := val0.Value.(int32)
		b := val1.Value.(int32)
		result := a ^ b
		intereter.pushVal(NewAtomValueInt(int(result)))
		return
	}

	if !IsNumberType(val0) || !IsNumberType(val1) {
		message := fmt.Sprintf("cannot bitwise xor type(s) %s and %s", GetTypeString(val0), GetTypeString(val1))
		intereter.pushVal(NewAtomValueError(message))
		return
	}

	lhsValue := CoerceToLong(val0)
	rhsValue := CoerceToLong(val1)
	result := lhsValue ^ rhsValue

	// Check if result can be represented as an int
	if result >= math.MinInt32 && result <= math.MaxInt32 {
		intereter.pushVal(NewAtomValueInt(int(result)))
	} else {
		intereter.pushVal(NewAtomValueNum(float64(result)))
	}
}

func DoSetIndex(intereter *AtomInterpreter, obj *AtomValue, index *AtomValue) {
	cleanupStack := func(size int) {
		for range size {
			intereter.pop()
		}
	}
	if CheckType(obj, AtomTypeArray) {
		if !IsNumberType(index) {
			cleanupStack(1)
			message := fmt.Sprintf("cannot set index type: %s with type: %s", GetTypeString(obj), GetTypeString(index))
			intereter.pushVal(NewAtomValueError(message))
			return
		}
		array := obj.Value.(*AtomArray)
		indexValue := CoerceToLong(index)

		if array.Freeze {
			cleanupStack(2)
			message := "cannot set index on frozen array"
			intereter.pushVal(NewAtomValueError(message))
			return
		}

		if !array.ValidIndex(int(indexValue)) {
			cleanupStack(2)
			message := fmt.Sprintf("index out of bounds: %d", indexValue)
			intereter.pushVal(NewAtomValueError(message))
			return
		}

		array.Set(int(indexValue), intereter.pop())
		return

	} else if CheckType(obj, AtomTypeObj) {
		if obj.Value.(*AtomObject).Freeze {
			cleanupStack(2) // includes duplicate obj
			message := "cannot set index on frozen object"
			intereter.pushVal(NewAtomValueError(message))
			return
		}

		objValue := obj.Value.(*AtomObject)
		indexValue := index.String()
		objValue.Set(indexValue, intereter.pop())
		return

	} else {
		cleanupStack(2)
		message := fmt.Sprintf("cannot set index type: %s", GetTypeString(obj))
		intereter.pushVal(NewAtomValueError(message))
		return
	}
}
