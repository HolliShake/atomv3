package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	runtime "dev.runtime"
)

func Eval(compiler *AtomCompile, ast *AtomAst) *runtime.AtomValue {
	switch ast.AstType {
	case AstTypeInt:
		var intValue int
		var err error

		str := ast.Str0
		if after, ok := strings.CutPrefix(str, "0x"); ok {
			val, parseErr := strconv.ParseInt(after, 16, 32)
			intValue = int(val)
			err = parseErr
		} else if after, ok := strings.CutPrefix(str, "0o"); ok {
			val, parseErr := strconv.ParseInt(after, 8, 32)
			intValue = int(val)
			err = parseErr
		} else if after, ok := strings.CutPrefix(str, "0b"); ok {
			val, parseErr := strconv.ParseInt(after, 2, 32)
			intValue = int(val)
			err = parseErr
		} else {
			intValue, err = strconv.Atoi(str)
		}

		if err != nil {
			Error(
				compiler.parser.tokenizer.file,
				compiler.parser.tokenizer.data,
				"Invalid integer",
				ast.Position,
			)
		}
		return runtime.NewAtomValueInt(intValue)

	case AstTypeNum:
		var numValue float64
		var err error

		numValue, err = strconv.ParseFloat(ast.Str0, 64)
		if err != nil {
			Error(
				compiler.parser.tokenizer.file,
				compiler.parser.tokenizer.data,
				"Invalid number",
				ast.Position,
			)
		}
		return runtime.NewAtomValueNum(numValue)

	case AstTypeStr:
		return runtime.NewAtomValueStr(ast.Str0)

	case AstTypeBool:
		if ast.Str0 == "true" {
			return compiler.state.TrueValue
		}
		return compiler.state.FalseValue

	case AstTypeNull:
		return compiler.state.NullValue

	case AstTypeBinaryAdd:
		{
			lhs := Eval(compiler, ast.Ast0)
			rhs := Eval(compiler, ast.Ast1)

			// Fast path for integers
			if runtime.CheckType(lhs, runtime.AtomTypeInt) && runtime.CheckType(rhs, runtime.AtomTypeInt) {
				// Use XOR trick to detect overflow
				a := runtime.CoerceToInt(lhs)
				b := runtime.CoerceToInt(rhs)
				sum := a + b
				if ((a ^ sum) & (b ^ sum)) < 0 {
					// Overflow occurred, promote to double
					return runtime.NewAtomValueNum(float64(a) + float64(b))
				}
				return runtime.NewAtomValueInt(int(sum))
			}

			// Fast path for strings
			if runtime.CheckType(lhs, runtime.AtomTypeStr) && runtime.CheckType(rhs, runtime.AtomTypeStr) {
				lhsStr := lhs.Str
				rhsStr := rhs.Str
				result := lhsStr + rhsStr
				return runtime.NewAtomValueStr(result)
			}

			if runtime.CheckType(lhs, runtime.AtomTypeStr) || runtime.CheckType(rhs, runtime.AtomTypeStr) {
				lhsStr := lhs.String()
				rhsStr := rhs.String()
				result := lhsStr + rhsStr
				return runtime.NewAtomValueStr(result)
			}

			// Check if both values are numbers (int or float)
			if !runtime.IsNumberType(lhs) || !runtime.IsNumberType(rhs) {
				Error(
					compiler.parser.tokenizer.file,
					compiler.parser.tokenizer.data,
					fmt.Sprintf("Error: cannot add types: %s and %s", runtime.GetTypeString(lhs), runtime.GetTypeString(rhs)),
					ast.Position,
				)
			}

			// Fallback path using coercion
			lhsValue := runtime.CoerceToNum(lhs)
			rhsValue := runtime.CoerceToNum(rhs)
			result := lhsValue + rhsValue

			// Try to preserve integer types if possible
			if runtime.IsInteger(result) && result <= math.MaxInt32 && result >= math.MinInt32 {
				return runtime.NewAtomValueInt(int(result))
			}
			return runtime.NewAtomValueNum(result)
		}

	case AstTypeBinaryMul:
		{
			lhs := Eval(compiler, ast.Ast0)
			rhs := Eval(compiler, ast.Ast1)

			// Fast path for integers
			if runtime.CheckType(lhs, runtime.AtomTypeInt) && runtime.CheckType(rhs, runtime.AtomTypeInt) {
				lhsInt := runtime.CoerceToInt(lhs)
				rhsInt := runtime.CoerceToInt(rhs)

				// Check for overflow using int64 arithmetic
				result := int64(lhsInt) * int64(rhsInt)
				if result >= math.MinInt32 && result <= math.MaxInt32 {
					return runtime.NewAtomValueInt(int(result))
				}
				// Overflow occurred, promote to float64
				return runtime.NewAtomValueNum(float64(result))
			}

			// Check if both values are numbers (int or float)
			if !runtime.IsNumberType(lhs) || !runtime.IsNumberType(rhs) {
				Error(
					compiler.parser.tokenizer.file,
					compiler.parser.tokenizer.data,
					fmt.Sprintf("Error: cannot multiply types: %s and %s", runtime.GetTypeString(lhs), runtime.GetTypeString(rhs)),
					ast.Position,
				)
			}

			// Fallback path using coercion
			lhsValue := runtime.CoerceToNum(lhs)
			rhsValue := runtime.CoerceToNum(rhs)
			result := lhsValue * rhsValue

			// Try to preserve integer types if possible
			if runtime.IsInteger(result) && result <= math.MaxInt32 && result >= math.MinInt32 {
				return runtime.NewAtomValueInt(int(result))
			}
			return runtime.NewAtomValueNum(result)
		}

	case AstTypeBinarySub:
		{
			lhs := Eval(compiler, ast.Ast0)
			rhs := Eval(compiler, ast.Ast1)

			// Fast path for integers
			if runtime.CheckType(lhs, runtime.AtomTypeInt) && runtime.CheckType(rhs, runtime.AtomTypeInt) {
				a := runtime.CoerceToInt(lhs)
				b := runtime.CoerceToInt(rhs)
				diff := a - b
				if ((a ^ b) & (a ^ diff)) < 0 {
					// Overflow occurred, promote to double
					return runtime.NewAtomValueNum(float64(a) - float64(b))
				}
				return runtime.NewAtomValueInt(int(diff))
			}

			// Check if both values are numbers (int or float)
			if !runtime.IsNumberType(lhs) || !runtime.IsNumberType(rhs) {
				Error(
					compiler.parser.tokenizer.file,
					compiler.parser.tokenizer.data,
					fmt.Sprintf("Error: cannot subtract types: %s and %s", runtime.GetTypeString(lhs), runtime.GetTypeString(rhs)),
					ast.Position,
				)
			}

			// Fallback path using coercion
			lhsValue := runtime.CoerceToNum(lhs)
			rhsValue := runtime.CoerceToNum(rhs)
			result := lhsValue - rhsValue

			// Try to preserve integer types if possible
			if runtime.IsInteger(result) && result <= math.MaxInt32 && result >= math.MinInt32 {
				return runtime.NewAtomValueInt(int(result))
			}
			return runtime.NewAtomValueNum(result)
		}

	case AstTypeBinaryDiv:
		{
			lhs := Eval(compiler, ast.Ast0)
			rhs := Eval(compiler, ast.Ast1)

			// Check if both values are numbers (int or float)
			if !runtime.IsNumberType(lhs) || !runtime.IsNumberType(rhs) {
				Error(
					compiler.parser.tokenizer.file,
					compiler.parser.tokenizer.data,
					fmt.Sprintf("Error: cannot divide types: %s and %s", runtime.GetTypeString(lhs), runtime.GetTypeString(rhs)),
					ast.Position,
				)
			}

			// Always use floating point division
			lhsValue := runtime.CoerceToNum(lhs)
			rhsValue := runtime.CoerceToNum(rhs)
			if rhsValue == 0 {
				Error(
					compiler.parser.tokenizer.file,
					compiler.parser.tokenizer.data,
					"division by zero",
					ast.Position,
				)
			}
			result := lhsValue / rhsValue

			return runtime.NewAtomValueNum(result)
		}

	case AstTypeBinaryMod:
		{
			lhs := Eval(compiler, ast.Ast0)
			rhs := Eval(compiler, ast.Ast1)

			// Check if both values are numbers (int or float)
			if !runtime.IsNumberType(lhs) || !runtime.IsNumberType(rhs) {
				Error(
					compiler.parser.tokenizer.file,
					compiler.parser.tokenizer.data,
					fmt.Sprintf("Error: cannot modulo types: %s and %s", runtime.GetTypeString(lhs), runtime.GetTypeString(rhs)),
					ast.Position,
				)
			}

			// Always use floating point modulo
			lhsValue := runtime.CoerceToNum(lhs)
			rhsValue := runtime.CoerceToNum(rhs)
			if rhsValue == 0 {
				Error(
					compiler.parser.tokenizer.file,
					compiler.parser.tokenizer.data,
					"Error: division by zero",
					ast.Position,
				)
			}
			result := math.Mod(lhsValue, rhsValue)

			return runtime.NewAtomValueNum(result)
		}

	case AstTypeBinaryShiftLeft:
		{
			lhs := Eval(compiler, ast.Ast0)
			rhs := Eval(compiler, ast.Ast1)

			// Fast path for integers
			if runtime.CheckType(lhs, runtime.AtomTypeInt) && runtime.CheckType(rhs, runtime.AtomTypeInt) {
				a := runtime.CoerceToInt(lhs)
				b := runtime.CoerceToInt(rhs)
				result := a << b
				return runtime.NewAtomValueInt(int(result))
			}

			// Check if both values are numbers (int or float)
			if !runtime.IsNumberType(lhs) || !runtime.IsNumberType(rhs) {
				Error(
					compiler.parser.tokenizer.file,
					compiler.parser.tokenizer.data,
					fmt.Sprintf("Error: cannot shift left types: %s and %s", runtime.GetTypeString(lhs), runtime.GetTypeString(rhs)),
					ast.Position,
				)
			}

			// Fallback path using coercion
			lhsValue := runtime.CoerceToLong(lhs)
			rhsValue := runtime.CoerceToLong(rhs)
			result := lhsValue << rhsValue

			// Check if result can be represented as an int
			if result >= math.MinInt32 && result <= math.MaxInt32 {
				return runtime.NewAtomValueInt(int(result))
			}
			return runtime.NewAtomValueNum(float64(result))
		}

	case AstTypeBinaryShiftRight:
		{
			lhs := Eval(compiler, ast.Ast0)
			rhs := Eval(compiler, ast.Ast1)

			// Fast path for integers
			if runtime.CheckType(lhs, runtime.AtomTypeInt) && runtime.CheckType(rhs, runtime.AtomTypeInt) {
				a := runtime.CoerceToInt(lhs)
				b := runtime.CoerceToInt(rhs)
				result := a >> b
				return runtime.NewAtomValueInt(int(result))
			}

			// Check if both values are numbers (int or float)
			if !runtime.IsNumberType(lhs) || !runtime.IsNumberType(rhs) {
				Error(
					compiler.parser.tokenizer.file,
					compiler.parser.tokenizer.data,
					fmt.Sprintf("Error: cannot shift right types: %s and %s", runtime.GetTypeString(lhs), runtime.GetTypeString(rhs)),
					ast.Position,
				)
			}

			// Fallback path using coercion
			lhsValue := runtime.CoerceToLong(lhs)
			rhsValue := runtime.CoerceToLong(rhs)
			result := lhsValue >> rhsValue

			// Try to preserve integer types if possible
			if result >= math.MinInt32 && result <= math.MaxInt32 {
				return runtime.NewAtomValueInt(int(result))
			}
			return runtime.NewAtomValueNum(float64(result))
		}

	case AstTypeBinaryGreaterThan:
		{
			lhs := Eval(compiler, ast.Ast0)
			rhs := Eval(compiler, ast.Ast1)

			if !runtime.IsNumberType(lhs) || !runtime.IsNumberType(rhs) {
				Error(
					compiler.parser.tokenizer.file,
					compiler.parser.tokenizer.data,
					fmt.Sprintf("Error: cannot compare greater than type(s) %s and %s", runtime.GetTypeString(lhs), runtime.GetTypeString(rhs)),
					ast.Position,
				)
			}

			// Coerce to long to avoid floating point comparisons
			lhsValue := runtime.CoerceToLong(lhs)
			rhsValue := runtime.CoerceToLong(rhs)

			// Compare the long values
			if lhsValue > rhsValue {
				return compiler.state.TrueValue
			}
			return compiler.state.FalseValue
		}

	case AstTypeBinaryGreaterThanEqual:
		{
			lhs := Eval(compiler, ast.Ast0)
			rhs := Eval(compiler, ast.Ast1)

			if !runtime.IsNumberType(lhs) || !runtime.IsNumberType(rhs) {
				Error(
					compiler.parser.tokenizer.file,
					compiler.parser.tokenizer.data,
					fmt.Sprintf("Error: cannot compare greater than or equal to type(s) %s and %s", runtime.GetTypeString(lhs), runtime.GetTypeString(rhs)),
					ast.Position,
				)
			}

			// Coerce to long to avoid floating point comparisons
			lhsValue := runtime.CoerceToLong(lhs)
			rhsValue := runtime.CoerceToLong(rhs)

			// Compare the long values
			if lhsValue >= rhsValue {
				return compiler.state.TrueValue
			}
			return compiler.state.FalseValue
		}

	case AstTypeBinaryLessThan:
		{
			lhs := Eval(compiler, ast.Ast0)
			rhs := Eval(compiler, ast.Ast1)

			if !runtime.IsNumberType(lhs) || !runtime.IsNumberType(rhs) {
				Error(
					compiler.parser.tokenizer.file,
					compiler.parser.tokenizer.data,
					fmt.Sprintf("Error: cannot compare less than type(s) %s and %s", runtime.GetTypeString(lhs), runtime.GetTypeString(rhs)),
					ast.Position,
				)
			}

			// Coerce to long to avoid floating point comparisons
			lhsValue := runtime.CoerceToLong(lhs)
			rhsValue := runtime.CoerceToLong(rhs)

			// Compare the long values
			if lhsValue < rhsValue {
				return compiler.state.TrueValue
			}
			return compiler.state.FalseValue
		}

	case AstTypeBinaryLessThanEqual:
		{
			lhs := Eval(compiler, ast.Ast0)
			rhs := Eval(compiler, ast.Ast1)

			if !runtime.IsNumberType(lhs) || !runtime.IsNumberType(rhs) {
				Error(
					compiler.parser.tokenizer.file,
					compiler.parser.tokenizer.data,
					fmt.Sprintf("Error: cannot compare less than or equal to type(s) %s and %s", runtime.GetTypeString(lhs), runtime.GetTypeString(rhs)),
					ast.Position,
				)
			}

			// Coerce to long to avoid floating point comparisons
			lhsValue := runtime.CoerceToLong(lhs)
			rhsValue := runtime.CoerceToLong(rhs)

			// Compare the long values
			if lhsValue <= rhsValue {
				return compiler.state.TrueValue
			}
			return compiler.state.FalseValue
		}

	case AstTypeBinaryEqual:
		{
			lhs := Eval(compiler, ast.Ast0)
			rhs := Eval(compiler, ast.Ast1)

			if runtime.IsNumberType(lhs) && runtime.IsNumberType(rhs) {
				lhsValue := runtime.CoerceToLong(lhs)
				rhsValue := runtime.CoerceToLong(rhs)
				if lhsValue == rhsValue {
					return compiler.state.TrueValue
				}
				return compiler.state.FalseValue
			}

			if runtime.CheckType(lhs, runtime.AtomTypeStr) && runtime.CheckType(rhs, runtime.AtomTypeStr) {
				lhsStr := lhs.Str
				rhsStr := rhs.Str
				if lhsStr == rhsStr {
					return compiler.state.TrueValue
				}
				return compiler.state.FalseValue
			}

			if runtime.CheckType(lhs, runtime.AtomTypeNull) && runtime.CheckType(rhs, runtime.AtomTypeNull) {
				return compiler.state.TrueValue
			}

			// For other types, use simple reference equality for now
			if lhs.HashValue() == rhs.HashValue() || lhs == rhs {
				return compiler.state.TrueValue
			}

			return compiler.state.FalseValue
		}

	case AstTypeBinaryNotEqual:
		{
			lhs := Eval(compiler, ast.Ast0)
			rhs := Eval(compiler, ast.Ast1)

			if runtime.IsNumberType(lhs) && runtime.IsNumberType(rhs) {
				lhsValue := runtime.CoerceToLong(lhs)
				rhsValue := runtime.CoerceToLong(rhs)
				if lhsValue != rhsValue {
					return compiler.state.TrueValue
				}
				return compiler.state.FalseValue
			}

			if runtime.CheckType(lhs, runtime.AtomTypeStr) && runtime.CheckType(rhs, runtime.AtomTypeStr) {
				lhsStr := lhs.Str
				rhsStr := rhs.Str
				if lhsStr != rhsStr {
					return compiler.state.TrueValue
				}
				return compiler.state.FalseValue
			}

			if runtime.CheckType(lhs, runtime.AtomTypeNull) && runtime.CheckType(rhs, runtime.AtomTypeNull) {
				return compiler.state.FalseValue
			}

			// For different types or other cases, they are not equal
			if lhs.Type != rhs.Type {
				return compiler.state.TrueValue
			}

			// For other types, use simple reference equality for now
			if lhs.HashValue() != rhs.HashValue() {
				return compiler.state.TrueValue
			}

			return compiler.state.FalseValue
		}

	case AstTypeBinaryAnd:
		{
			lhs := Eval(compiler, ast.Ast0)
			rhs := Eval(compiler, ast.Ast1)

			// Fast path for integers
			if runtime.CheckType(lhs, runtime.AtomTypeInt) && runtime.CheckType(rhs, runtime.AtomTypeInt) {
				a := lhs.I32
				b := rhs.I32
				result := a & b
				return runtime.NewAtomValueInt(int(result))
			}

			if !runtime.IsNumberType(lhs) || !runtime.IsNumberType(rhs) {
				Error(
					compiler.parser.tokenizer.file,
					compiler.parser.tokenizer.data,
					fmt.Sprintf("Error: cannot bitwise and type(s) %s and %s", runtime.GetTypeString(lhs), runtime.GetTypeString(rhs)),
					ast.Position,
				)
			}

			lhsValue := runtime.CoerceToLong(lhs)
			rhsValue := runtime.CoerceToLong(rhs)
			result := lhsValue & rhsValue

			// Check if result can be represented as an int
			if result >= math.MinInt32 && result <= math.MaxInt32 {
				return runtime.NewAtomValueInt(int(result))
			} else {
				return runtime.NewAtomValueNum(float64(result))
			}
		}

	case AstTypeBinaryOr:
		{
			lhs := Eval(compiler, ast.Ast0)
			rhs := Eval(compiler, ast.Ast1)

			// Fast path for integers
			if runtime.CheckType(lhs, runtime.AtomTypeInt) && runtime.CheckType(rhs, runtime.AtomTypeInt) {
				a := lhs.I32
				b := rhs.I32
				result := a | b
				return runtime.NewAtomValueInt(int(result))
			}

			if !runtime.IsNumberType(lhs) || !runtime.IsNumberType(rhs) {
				Error(
					compiler.parser.tokenizer.file,
					compiler.parser.tokenizer.data,
					fmt.Sprintf("Error: cannot bitwise or type(s) %s and %s", runtime.GetTypeString(lhs), runtime.GetTypeString(rhs)),
					ast.Position,
				)
			}

			lhsValue := runtime.CoerceToLong(lhs)
			rhsValue := runtime.CoerceToLong(rhs)
			result := lhsValue | rhsValue

			// Check if result can be represented as an int
			if result >= math.MinInt32 && result <= math.MaxInt32 {
				return runtime.NewAtomValueInt(int(result))
			} else {
				return runtime.NewAtomValueNum(float64(result))
			}
		}

	case AstTypeBinaryXor:
		{
			lhs := Eval(compiler, ast.Ast0)
			rhs := Eval(compiler, ast.Ast1)

			// Fast path for integers
			if runtime.CheckType(lhs, runtime.AtomTypeInt) && runtime.CheckType(rhs, runtime.AtomTypeInt) {
				a := lhs.I32
				b := rhs.I32
				result := a ^ b
				return runtime.NewAtomValueInt(int(result))
			}

			if !runtime.IsNumberType(lhs) || !runtime.IsNumberType(rhs) {
				Error(
					compiler.parser.tokenizer.file,
					compiler.parser.tokenizer.data,
					fmt.Sprintf("Error: cannot bitwise xor type(s) %s and %s", runtime.GetTypeString(lhs), runtime.GetTypeString(rhs)),
					ast.Position,
				)
			}

			lhsValue := runtime.CoerceToLong(lhs)
			rhsValue := runtime.CoerceToLong(rhs)
			result := lhsValue ^ rhsValue

			// Check if result can be represented as an int
			if result >= math.MinInt32 && result <= math.MaxInt32 {
				return runtime.NewAtomValueInt(int(result))
			} else {
				return runtime.NewAtomValueNum(float64(result))
			}
		}
	default:
		panic("Invalid AST type")
	}
}
