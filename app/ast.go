package main

type AstType int

type Ast struct {
	astType  AstType
	str0     string
	ast0     *Ast
	ast1     *Ast
	ast2     *Ast
	ast3     *Ast
	arr0     []*Ast
	arr1     []*Ast
	arr2     []*Ast
	position Position
}

const (
	AstTypeInt AstType = iota
	AstTypeNum
	AstTypeStr
	AstTypeBool
	AstTypeNull
	AstTypeClass
	AstTypeEnum
	AstTypeImport
	AstTypeFunc
	AstTypeVar
	AstTypeConst
	AstTypeLocal
	AstTypeBreak
	AstTypeContinue
	AstTypeReturn
	AstTypeIf
	AstTypeElse
	AstTypeSwitch
	AstTypeCase
	AstTypeDefault
	AstTypeFor
	AstTypeWhile
	AstTypeDoWhile
	AstTypeNew
	AstTypeThis
	AstInvalid
)

func NewAst() *Ast {
	ast := new(Ast)
	ast.astType = AstInvalid
	ast.str0 = ""
	ast.ast0 = nil
	ast.ast1 = nil
	ast.ast2 = nil
	ast.ast3 = nil
	ast.arr0 = nil
	ast.arr1 = nil
	ast.arr2 = nil
	ast.position = Position{}
	return ast
}

func NewTerminal(str0 string) *Ast {
	ast := NewAst()
	ast.str0 = str0
	return ast
}
