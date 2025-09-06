package main

import "fmt"

type AstType int

/*
 * Export everything for Compiler
 */
type Ast struct {
	AstType  AstType
	Str0     string
	Ast0     *Ast
	Ast1     *Ast
	Ast2     *Ast
	Ast3     *Ast
	Arr0     []*Ast
	Arr1     []*Ast
	Arr2     []*Ast
	Position Position
}

const (
	AstTypeIdn AstType = iota
	AstTypeInt
	AstTypeNum
	AstTypeStr
	AstTypeBool
	AstTypeNull
	AstTypeArray
	AstTypeObject
	AstTypeBinaryMul
	AstTypeBinaryDiv
	AstTypeBinaryMod
	AstTypeBinaryAdd
	AstTypeBinarySub
	AstTypeFunction
	AstTypeProgram
	AstInvalid
)

func NewAst(astType AstType, position Position) *Ast {
	ast := new(Ast)
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

func getBinaryAstType(op Token) AstType {
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
	default:
		return AstInvalid
	}
}

func NewTerminal(astType AstType, value string, position Position) *Ast {
	ast := NewAst(astType, position)
	ast.Str0 = value
	return ast
}

func NewBinary(ast0 *Ast, op Token, ast1 *Ast, position Position) *Ast {
	ast := NewAst(getBinaryAstType(op), position)
	ast.Ast0 = ast0
	ast.Ast1 = ast1
	return ast
}

func NewFunction(name *Ast, params []*Ast, body []*Ast, position Position) *Ast {
	ast := NewAst(AstTypeFunction, position)
	ast.Ast0 = name
	ast.Arr0 = params
	ast.Arr1 = body
	return ast
}

func NewProgram(body []*Ast, position Position) *Ast {
	ast := NewAst(AstTypeProgram, position)
	ast.Arr1 = body
	return ast
}

func (a *Ast) String() string {
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
	default:
		return a.String()
	}
}
