package runtime

import "math"

func DoMultiplication(intereter *Interpreter, value1 *AtomValue, value2 *AtomValue) {
	// Fast path for integers
	if CheckType(value1, AtomTypeInt) && CheckType(value2, AtomTypeInt) {
		lhs := CoerceToInt(value1)
		rhs := CoerceToInt(value2)

		// Check for overflow using int64 arithmetic
		result := int64(lhs) * int64(rhs)
		if result >= math.MinInt32 && result <= math.MaxInt32 {
			intereter.EvaluationStack.Push(NewAtomValueInt(int(result)))
			return
		}
		// Overflow occurred, promote to float64
		intereter.EvaluationStack.Push(NewAtomValueNum(float64(result)))
		return
	}

	// Check if both values are numbers (int or float)
	if !IsNumberType(value1) || !IsNumberType(value2) {
		// TODO: Implement proper error handling instead of panic
		panic("cannot multiply types: " + getTypeString(value1) + " and " + getTypeString(value2))
	}

	// Fallback path using coercion
	lhsValue := CoerceToNum(value1)
	rhsValue := CoerceToNum(value2)
	result := lhsValue * rhsValue

	// Try to preserve integer types if possible
	if IsInteger(result) && result <= math.MaxInt32 && result >= math.MinInt32 {
		intereter.EvaluationStack.Push(NewAtomValueInt(int(result)))
		return
	}
	intereter.EvaluationStack.Push(NewAtomValueNum(result))
}

func DoDivision(intereter *Interpreter, value1 *AtomValue, value2 *AtomValue) {
	// Fast path for integers
	if CheckType(value1, AtomTypeInt) && CheckType(value2, AtomTypeInt) {
		a := CoerceToInt(value1)
		b := CoerceToInt(value2)
		if b == 0 {
			// TODO: Implement proper error handling instead of panic
			panic("division by zero")
		}
		result := a / b
		intereter.EvaluationStack.Push(NewAtomValueInt(int(result)))
		return
	}

	// Check if both values are numbers (int or float)
	if !IsNumberType(value1) || !IsNumberType(value2) {
		// TODO: Implement proper error handling instead of panic
		panic("cannot divide types: " + getTypeString(value1) + " and " + getTypeString(value2))
	}

	// Fallback path using coercion
	lhsValue := CoerceToNum(value1)
	rhsValue := CoerceToNum(value2)
	if rhsValue == 0 {
		// TODO: Implement proper error handling instead of panic
		panic("division by zero")
	}
	result := lhsValue / rhsValue

	// Try to preserve integer types if possible
	if IsInteger(result) && result <= math.MaxInt32 && result >= math.MinInt32 {
		intereter.EvaluationStack.Push(NewAtomValueInt(int(result)))
		return
	}
	intereter.EvaluationStack.Push(NewAtomValueNum(result))
}

func DoModulus(intereter *Interpreter, value1 *AtomValue, value2 *AtomValue) {
	// Fast path for integers
	if CheckType(value1, AtomTypeInt) && CheckType(value2, AtomTypeInt) {
		a := CoerceToInt(value1)
		b := CoerceToInt(value2)
		if b == 0 {
			// TODO: Implement proper error handling instead of panic
			panic("division by zero")
		}
		result := a % b
		intereter.EvaluationStack.Push(NewAtomValueInt(int(result)))
		return
	}

	// Check if both values are numbers (int or float)
	if !IsNumberType(value1) || !IsNumberType(value2) {
		// TODO: Implement proper error handling instead of panic
		panic("cannot modulo types: " + getTypeString(value1) + " and " + getTypeString(value2))
	}

	// Fallback path using coercion
	lhsValue := CoerceToNum(value1)
	rhsValue := CoerceToNum(value2)
	if rhsValue == 0 {
		// TODO: Implement proper error handling instead of panic
		panic("division by zero")
	}
	result := math.Mod(lhsValue, rhsValue)

	// Try to preserve integer types if possible
	if IsInteger(result) && result <= math.MaxInt32 && result >= math.MinInt32 {
		intereter.EvaluationStack.Push(NewAtomValueInt(int(result)))
		return
	}
	intereter.EvaluationStack.Push(NewAtomValueNum(result))
}

func DoAddition(intereter *Interpreter, value1 *AtomValue, value2 *AtomValue) {
	// Fast path for integers
	if CheckType(value1, AtomTypeInt) && CheckType(value2, AtomTypeInt) {
		// Use XOR trick to detect overflow
		a := CoerceToInt(value1)
		b := CoerceToInt(value2)
		sum := a + b
		if ((a ^ sum) & (b ^ sum)) < 0 {
			// Overflow occurred, promote to double
			intereter.EvaluationStack.Push(NewAtomValueNum(float64(a) + float64(b)))
			return
		}
		intereter.EvaluationStack.Push(NewAtomValueInt(int(sum)))
		return
	}

	// Fast path for strings
	if CheckType(value1, AtomTypeStr) && CheckType(value2, AtomTypeStr) {
		lhs := value1.Value.(string)
		rhs := value2.Value.(string)
		result := lhs + rhs
		intereter.EvaluationStack.Push(NewAtomValueStr(result))
		return
	}

	// Check if both values are numbers (int or float)
	if !IsNumberType(value1) || !IsNumberType(value2) {
		// TODO: Implement proper error handling instead of panic
		panic("cannot add types: " + getTypeString(value1) + " and " + getTypeString(value2))
	}

	// Fallback path using coercion
	lhsValue := CoerceToNum(value1)
	rhsValue := CoerceToNum(value2)
	result := lhsValue + rhsValue

	// Try to preserve integer types if possible
	if IsInteger(result) && result <= math.MaxInt32 && result >= math.MinInt32 {
		intereter.EvaluationStack.Push(NewAtomValueInt(int(result)))
		return
	}
	intereter.EvaluationStack.Push(NewAtomValueNum(result))
}

func DoSubtraction(intereter *Interpreter, value1 *AtomValue, value2 *AtomValue) {
	// Fast path for integers
	if CheckType(value1, AtomTypeInt) && CheckType(value2, AtomTypeInt) {
		a := CoerceToInt(value1)
		b := CoerceToInt(value2)
		diff := a - b
		if ((a ^ b) & (a ^ diff)) < 0 {
			// Overflow occurred, promote to double
			intereter.EvaluationStack.Push(NewAtomValueNum(float64(a) - float64(b)))
			return
		}
		intereter.EvaluationStack.Push(NewAtomValueInt(int(diff)))
		return
	}

	// Check if both values are numbers (int or float)
	if !IsNumberType(value1) || !IsNumberType(value2) {
		// TODO: Implement proper error handling instead of panic
		panic("cannot subtract types: " + getTypeString(value1) + " and " + getTypeString(value2))
	}

	// Fallback path using coercion
	lhsValue := CoerceToNum(value1)
	rhsValue := CoerceToNum(value2)
	result := lhsValue - rhsValue

	// Try to preserve integer types if possible
	if IsInteger(result) && result <= math.MaxInt32 && result >= math.MinInt32 {
		intereter.EvaluationStack.Push(NewAtomValueInt(int(result)))
		return
	}
	intereter.EvaluationStack.Push(NewAtomValueNum(result))
}

func getTypeString(value *AtomValue) string {
	switch value.Type {
	case AtomTypeInt:
		return "int"
	case AtomTypeNum:
		return "number"
	case AtomTypeBool:
		return "bool"
	case AtomTypeStr:
		return "string"
	case AtomTypeNull:
		return "null"
	case AtomTypeObj:
		return "object"
	case AtomTypeFunc:
		return "function"
	default:
		return "unknown"
	}
}
