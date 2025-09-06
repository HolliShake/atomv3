package runtime

import "math"

func DoMultiplication(intereter *AtomInterpreter, val0 *AtomValue, val1 *AtomValue) {
	// Fast path for integers
	if CheckType(val0, AtomTypeInt) && CheckType(val1, AtomTypeInt) {
		lhs := CoerceToInt(val0)
		rhs := CoerceToInt(val1)

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
	if !IsNumberType(val0) || !IsNumberType(val1) {
		// TODO: Implement proper error handling instead of panic
		panic("cannot multiply types: " + GetTypeString(val0) + " and " + GetTypeString(val1))
	}

	// Fallback path using coercion
	lhsValue := CoerceToNum(val0)
	rhsValue := CoerceToNum(val1)
	result := lhsValue * rhsValue

	// Try to preserve integer types if possible
	if IsInteger(result) && result <= math.MaxInt32 && result >= math.MinInt32 {
		intereter.EvaluationStack.Push(NewAtomValueInt(int(result)))
		return
	}
	intereter.EvaluationStack.Push(NewAtomValueNum(result))
}

func DoDivision(intereter *AtomInterpreter, val0 *AtomValue, val1 *AtomValue) {
	// Fast path for integers
	if CheckType(val0, AtomTypeInt) && CheckType(val1, AtomTypeInt) {
		a := CoerceToInt(val0)
		b := CoerceToInt(val1)
		if b == 0 {
			// TODO: Implement proper error handling instead of panic
			panic("division by zero")
		}
		result := a / b
		intereter.EvaluationStack.Push(NewAtomValueInt(int(result)))
		return
	}

	// Check if both values are numbers (int or float)
	if !IsNumberType(val0) || !IsNumberType(val1) {
		// TODO: Implement proper error handling instead of panic
		panic("cannot divide types: " + GetTypeString(val0) + " and " + GetTypeString(val1))
	}

	// Fallback path using coercion
	lhsValue := CoerceToNum(val0)
	rhsValue := CoerceToNum(val1)
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

func DoModulus(intereter *AtomInterpreter, val0 *AtomValue, val1 *AtomValue) {
	// Fast path for integers
	if CheckType(val0, AtomTypeInt) && CheckType(val1, AtomTypeInt) {
		a := CoerceToInt(val0)
		b := CoerceToInt(val1)
		if b == 0 {
			// TODO: Implement proper error handling instead of panic
			panic("division by zero")
		}
		result := a % b
		intereter.EvaluationStack.Push(NewAtomValueInt(int(result)))
		return
	}

	// Check if both values are numbers (int or float)
	if !IsNumberType(val0) || !IsNumberType(val1) {
		// TODO: Implement proper error handling instead of panic
		panic("cannot modulo types: " + GetTypeString(val0) + " and " + GetTypeString(val1))
	}

	// Fallback path using coercion
	lhsValue := CoerceToNum(val0)
	rhsValue := CoerceToNum(val1)
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

func DoAddition(intereter *AtomInterpreter, val0 *AtomValue, val1 *AtomValue) {
	// Fast path for integers
	if CheckType(val0, AtomTypeInt) && CheckType(val1, AtomTypeInt) {
		// Use XOR trick to detect overflow
		a := CoerceToInt(val0)
		b := CoerceToInt(val1)
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
	if CheckType(val0, AtomTypeStr) && CheckType(val1, AtomTypeStr) {
		lhs := val0.Value.(string)
		rhs := val1.Value.(string)
		result := lhs + rhs
		intereter.EvaluationStack.Push(NewAtomValueStr(result))
		return
	}

	// Check if both values are numbers (int or float)
	if !IsNumberType(val0) || !IsNumberType(val1) {
		// TODO: Implement proper error handling instead of panic
		panic("cannot add types: " + GetTypeString(val0) + " and " + GetTypeString(val1))
	}

	// Fallback path using coercion
	lhsValue := CoerceToNum(val0)
	rhsValue := CoerceToNum(val1)
	result := lhsValue + rhsValue

	// Try to preserve integer types if possible
	if IsInteger(result) && result <= math.MaxInt32 && result >= math.MinInt32 {
		intereter.EvaluationStack.Push(NewAtomValueInt(int(result)))
		return
	}
	intereter.EvaluationStack.Push(NewAtomValueNum(result))
}

func DoSubtraction(intereter *AtomInterpreter, val0 *AtomValue, val1 *AtomValue) {
	// Fast path for integers
	if CheckType(val0, AtomTypeInt) && CheckType(val1, AtomTypeInt) {
		a := CoerceToInt(val0)
		b := CoerceToInt(val1)
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
	if !IsNumberType(val0) || !IsNumberType(val1) {
		// TODO: Implement proper error handling instead of panic
		panic("cannot subtract types: " + GetTypeString(val0) + " and " + GetTypeString(val1))
	}

	// Fallback path using coercion
	lhsValue := CoerceToNum(val0)
	rhsValue := CoerceToNum(val1)
	result := lhsValue - rhsValue

	// Try to preserve integer types if possible
	if IsInteger(result) && result <= math.MaxInt32 && result >= math.MinInt32 {
		intereter.EvaluationStack.Push(NewAtomValueInt(int(result)))
		return
	}
	intereter.EvaluationStack.Push(NewAtomValueNum(result))
}
