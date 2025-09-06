package main

import "fmt"

type AtomAstType int

/*
 * Export everything for Compiler
 */
type AtomAst struct {
	AstType  AtomAstType
	Str0     string
	Ast0     *AtomAst
	Ast1     *AtomAst
	Ast2     *AtomAst
	Ast3     *AtomAst
	Arr0     []*AtomAst
	Arr1     []*AtomAst
	Arr2     []*AtomAst
	Position AtomPosition
}

const (
	AstTypeIdn AtomAstType = iota
	AstTypeInt
	AstTypeNum
	AstTypeStr
	AstTypeBool
	AstTypeNull
	AstTypeArray
	AstTypeObject
	AstTypeCall
	AstTypeBinaryMul
	AstTypeBinaryDiv
	AstTypeBinaryMod
	AstTypeBinaryAdd
	AstTypeBinarySub
	AstTypeBinaryShiftRight
	AstTypeBinaryShiftLeft
	AstTypeBinaryGreaterThan
	AstTypeBinaryGreaterThanEqual
	AstTypeBinaryLessThan
	AstTypeBinaryLessThanEqual
	AstTypeBinaryEqual
	AstTypeBinaryNotEqual
	AstTypeBinaryAnd
	AstTypeBinaryOr
	AstTypeBinaryXor
	AstTypeLogicalAnd
	AstTypeLogicalOr
	AstTypeReturnStatement
	AstTypeEmptyStatement
	AstTypeExpressionStatement
	AstTypeFunction
	AstTypeVarStatement
	AstTypeConstStatement
	AstTypeLocalStatement
	AstTypeIfStatement
	AstTypeProgram
	AstInvalid
)

func NewAtomAst(astType AtomAstType, position AtomPosition) *AtomAst {
	ast := new(AtomAst)
	ast.AstType = astType
	ast.Str0 = ""
	ast.Ast0 = nil
	ast.Ast1 = nil
	ast.Ast2 = nil
	ast.Ast3 = nil
	ast.Arr0 = nil
	ast.Arr1 = nil
	ast.Arr2 = nil
	ast.Position = position
	return ast
}

func getBinaryAstType(op AtomToken) AtomAstType {
	switch op.Value {
	case "*":
		return AstTypeBinaryMul
	case "/":
		return AstTypeBinaryDiv
	case "%":
		return AstTypeBinaryMod
	case "+":
		return AstTypeBinaryAdd
	case "-":
		return AstTypeBinarySub
	case ">>":
		return AstTypeBinaryShiftRight
	case "<<":
		return AstTypeBinaryShiftLeft
	case ">":
		return AstTypeBinaryGreaterThan
	case ">=":
		return AstTypeBinaryGreaterThanEqual
	case "<":
		return AstTypeBinaryLessThan
	case "<=":
		return AstTypeBinaryLessThanEqual
	case "==":
		return AstTypeBinaryEqual
	case "!=":
		return AstTypeBinaryNotEqual
	case "&":
		return AstTypeBinaryAnd
	case "|":
		return AstTypeBinaryOr
	case "^":
		return AstTypeBinaryXor
	case "&&":
		return AstTypeLogicalAnd
	case "||":
		return AstTypeLogicalOr
	default:
		return AstInvalid
	}
}

func NewTerminal(astType AtomAstType, value string, position AtomPosition) *AtomAst {
	ast := NewAtomAst(astType, position)
	ast.Str0 = value
	return ast
}

func NewCall(ast0 *AtomAst, args []*AtomAst, position AtomPosition) *AtomAst {
	ast := NewAtomAst(AstTypeCall, position)
	ast.Ast0 = ast0
	ast.Arr0 = args
	return ast
}

func NewBinary(ast0 *AtomAst, op AtomToken, ast1 *AtomAst, position AtomPosition) *AtomAst {
	ast := NewAtomAst(getBinaryAstType(op), position)
	ast.Ast0 = ast0
	ast.Ast1 = ast1
	return ast
}

func NewReturnStatement(expr *AtomAst, position AtomPosition) *AtomAst {
	ast := NewAtomAst(AstTypeReturnStatement, position)
	ast.Ast0 = expr
	return ast
}

func NewEmptyStatement(position AtomPosition) *AtomAst {
	ast := NewAtomAst(AstTypeEmptyStatement, position)
	return ast
}

func NewExpressionStatement(expr *AtomAst, position AtomPosition) *AtomAst {
	ast := NewAtomAst(AstTypeExpressionStatement, position)
	ast.Ast0 = expr
	return ast
}

func NewFunction(name *AtomAst, params []*AtomAst, body []*AtomAst, position AtomPosition) *AtomAst {
	ast := NewAtomAst(AstTypeFunction, position)
	ast.Ast0 = name
	ast.Arr0 = params
	ast.Arr1 = body
	return ast
}

func NewVarStatement(keys []*AtomAst, vals []*AtomAst, position AtomPosition) *AtomAst {
	ast := NewAtomAst(AstTypeVarStatement, position)
	ast.Arr0 = keys
	ast.Arr1 = vals
	return ast
}

func NewConstStatement(keys []*AtomAst, vals []*AtomAst, position AtomPosition) *AtomAst {
	ast := NewAtomAst(AstTypeConstStatement, position)
	ast.Arr0 = keys
	ast.Arr1 = vals
	return ast
}

func NewLocalStatement(keys []*AtomAst, vals []*AtomAst, position AtomPosition) *AtomAst {
	ast := NewAtomAst(AstTypeLocalStatement, position)
	ast.Arr0 = keys
	ast.Arr1 = vals
	return ast
}

func NewIfStatement(condition *AtomAst, thenValue *AtomAst, elseValue *AtomAst, position AtomPosition) *AtomAst {
	ast := NewAtomAst(AstTypeIfStatement, position)
	ast.Ast0 = condition
	ast.Ast1 = thenValue
	ast.Ast2 = elseValue
	return ast
}

func NewProgram(body []*AtomAst, position AtomPosition) *AtomAst {
	ast := NewAtomAst(AstTypeProgram, position)
	ast.Arr1 = body
	return ast
}

func (a *AtomAst) String() string {
	switch a.AstType {
	case AstTypeIdn,
		AstTypeInt,
		AstTypeNum,
		AstTypeStr,
		AstTypeBool,
		AstTypeNull:
		return a.Str0
	case AstTypeArray:
		return "[]"
	case AstTypeObject:
		return "{}"
	case AstTypeBinaryMul:
		return fmt.Sprintf("(%s * %s)", a.Ast0.String(), a.Ast1.String())
	case AstTypeBinaryDiv:
		return fmt.Sprintf("(%s / %s)", a.Ast0.String(), a.Ast1.String())
	case AstTypeBinaryMod:
		return fmt.Sprintf("(%s %% %s)", a.Ast0.String(), a.Ast1.String())
	case AstTypeBinaryAdd:
		return fmt.Sprintf("(%s + %s)", a.Ast0.String(), a.Ast1.String())
	case AstTypeBinarySub:
		return fmt.Sprintf("(%s - %s)", a.Ast0.String(), a.Ast1.String())
	case AstTypeFunction:
		{
			params := ""
			for idx, param := range a.Arr0 {
				params += param.String()
				if idx < len(a.Arr0)-1 {
					params += ", "
				}
			}
			body := ""
			for idx, stmt := range a.Arr1 {
				body += "\t" + stmt.String()
				if idx < len(a.Arr1)-1 {
					body += "\n"
				}
			}
			return fmt.Sprintf("func %s(%s) {\n%s\n}", a.Ast0.String(), params, body)
		}
	case AstTypeVarStatement:
		vals := ""
		for idx, key := range a.Arr0 {
			val := a.Arr1[idx]
			vals += key.String() + " = " + val.String()
			if idx < len(a.Arr0)-1 {
				vals += ", "
			}
		}
		return fmt.Sprintf("var %s;", vals)
	case AstTypeConstStatement:
		vals := ""
		for idx, key := range a.Arr0 {
			val := a.Arr1[idx]
			vals += key.String() + " = " + val.String()
			if idx < len(a.Arr0)-1 {
				vals += ", "
			}
		}
		return fmt.Sprintf("const %s;", vals)
	case AstTypeLocalStatement:
		vals := ""
		for idx, key := range a.Arr0 {
			val := a.Arr1[idx]
			vals += key.String() + " = " + val.String()
			if idx < len(a.Arr0)-1 {
				vals += ", "
			}
		}
		return fmt.Sprintf("local %s;", vals)
	case AstTypeIfStatement:
		return fmt.Sprintf("if (%s) {\n%s\n} else {\n%s\n}", a.Ast0.String(), a.Ast1.String(), a.Ast2.String())
	default:
		return a.String()
	}
}
