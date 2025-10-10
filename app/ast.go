package main

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
	AstTypeBase
	AstTypeArray
	AstTypeObject
	AstTypeKeyValue
	AstTypeAsyncFunctionExpression
	AstTypeFunctionExpression
	AstTypeCall
	AstTypeIndex
	AstTypeMember
	AstTypeAllocation
	AstTypePostfixInc
	AstTypePostfixDec
	AstTypeUnaryBitNot
	AstTypeUnaryNot
	AstTypeUnaryNeg
	AstTypeUnaryPos
	AstTypeUnaryInc
	AstTypeUnaryDec
	AstTypeUnaryTypeof
	AstTypeUnaryAwait
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
	AstTypeAssign
	AstTypeMulAssign
	AstTypeDivAssign
	AstTypeModAssign
	AstTypeAddAssign
	AstTypeSubAssign
	AstTypeLeftShiftAssign
	AstTypeRightShiftAssign
	AstTypeBitwiseAndAssign
	AstTypeBitwiseOrAssign
	AstTypeBitwiseXorAssign
	AstTypeIfExpression
	AstTypeSwitchExpression
	AstTypeCatchExpression
	AstTypeBreakStatement
	AstTypeContinueStatement
	AstTypeReturnStatement
	AstTypeEmptyStatement
	AstTypeExpressionStatement
	AstTypeClass
	AstTypeEnum
	AstTypeAsyncFunction
	AstTypeFunction
	AstTypeBlock
	AstTypeVarStatement
	AstTypeConstStatement
	AstTypeLocalStatement
	AstTypeImportStatement
	AstTypeIfStatement
	AstTypeSwitchStatement
	AstTypeWhileStatement
	AstTypeDoWhileStatement
	AstTypeForStatement
	AstTypeProgram
	AstInvalid
)

func NewAtomAst(astType AtomAstType, position AtomPosition) *AtomAst {
	return &AtomAst{
		AstType:  astType,
		Str0:     "",
		Ast0:     nil,
		Ast1:     nil,
		Ast2:     nil,
		Ast3:     nil,
		Arr0:     nil,
		Arr1:     nil,
		Arr2:     nil,
		Position: position,
	}
}

func getPostfixAstType(op AtomToken) AtomAstType {
	switch op.Value {
	case "++":
		return AstTypePostfixInc
	case "--":
		return AstTypePostfixDec
	default:
		return AstInvalid
	}
}

func getUnaryAstType(op AtomToken) AtomAstType {
	switch op.Value {
	case "~":
		return AstTypeUnaryBitNot
	case "!":
		return AstTypeUnaryNot
	case "-":
		return AstTypeUnaryNeg
	case "+":
		return AstTypeUnaryPos
	case "++":
		return AstTypeUnaryInc
	case "--":
		return AstTypeUnaryDec
	case "typeof":
		return AstTypeUnaryTypeof
	case "await":
		return AstTypeUnaryAwait
	default:
		return AstInvalid
	}
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
	case "=":
		return AstTypeAssign
	case "*=":
		return AstTypeMulAssign
	case "/=":
		return AstTypeDivAssign
	case "%=":
		return AstTypeModAssign
	case "+=":
		return AstTypeAddAssign
	case "-=":
		return AstTypeSubAssign
	case ">>=":
		return AstTypeRightShiftAssign
	case "<<=":
		return AstTypeLeftShiftAssign
	case "&=":
		return AstTypeBitwiseAndAssign
	case "|=":
		return AstTypeBitwiseOrAssign
	case "^=":
		return AstTypeBitwiseXorAssign
	default:
		return AstInvalid
	}
}

func NewTerminal(astType AtomAstType, value string, position AtomPosition) *AtomAst {
	ast := NewAtomAst(astType, position)
	ast.Str0 = value
	return ast
}

func NewArray(elements []*AtomAst, position AtomPosition) *AtomAst {
	ast := NewAtomAst(AstTypeArray, position)
	ast.Arr0 = elements
	return ast
}

func NewObject(elements []*AtomAst, position AtomPosition) *AtomAst {
	ast := NewAtomAst(AstTypeObject, position)
	ast.Arr0 = elements
	return ast
}

func NewKeyValue(key *AtomAst, val *AtomAst, position AtomPosition) *AtomAst {
	ast := NewAtomAst(AstTypeKeyValue, position)
	ast.Ast0 = key
	ast.Ast1 = val
	return ast
}

func NewFunctionExpression(astType AtomAstType, params []*AtomAst, body []*AtomAst, position AtomPosition) *AtomAst {
	ast := NewAtomAst(astType, position)
	ast.Arr0 = params
	ast.Arr1 = body
	return ast
}

func NewMember(obj *AtomAst, key *AtomAst, position AtomPosition) *AtomAst {
	ast := NewAtomAst(AstTypeMember, position)
	ast.Ast0 = obj
	ast.Ast1 = key
	return ast
}

func NewIndex(obj *AtomAst, index *AtomAst, position AtomPosition) *AtomAst {
	ast := NewAtomAst(AstTypeIndex, position)
	ast.Ast0 = obj
	ast.Ast1 = index
	return ast
}

func NewCall(ast0 *AtomAst, args []*AtomAst, position AtomPosition) *AtomAst {
	ast := NewAtomAst(AstTypeCall, position)
	ast.Ast0 = ast0
	ast.Arr0 = args
	return ast
}

func NewAllocation(ast0 *AtomAst, position AtomPosition) *AtomAst {
	ast := NewAtomAst(AstTypeAllocation, position)
	ast.Ast0 = ast0
	return ast
}

func NewPostfix(op AtomToken, ast0 *AtomAst, position AtomPosition) *AtomAst {
	ast := NewAtomAst(getPostfixAstType(op), position)
	ast.Ast0 = ast0
	return ast
}

func NewUnary(op AtomToken, ast0 *AtomAst, position AtomPosition) *AtomAst {
	ast := NewAtomAst(getUnaryAstType(op), position)
	ast.Ast0 = ast0
	return ast
}

func NewBinary(ast0 *AtomAst, op AtomToken, ast1 *AtomAst, position AtomPosition) *AtomAst {
	ast := NewAtomAst(getBinaryAstType(op), position)
	ast.Ast0 = ast0
	ast.Ast1 = ast1
	return ast
}

func NewIfExpression(condition *AtomAst, thenValue *AtomAst, elseValue *AtomAst, position AtomPosition) *AtomAst {
	ast := NewAtomAst(AstTypeIfExpression, position)
	ast.Ast0 = condition
	ast.Ast1 = thenValue
	ast.Ast2 = elseValue
	return ast
}

func NewSwitchExpression(condition *AtomAst, cases []*AtomAst, values []*AtomAst, value *AtomAst, position AtomPosition) *AtomAst {
	ast := NewAtomAst(AstTypeSwitchExpression, position)
	ast.Ast0 = condition
	ast.Ast1 = value
	ast.Arr0 = cases
	ast.Arr1 = values
	return ast
}

func NewCatchExpression(condition *AtomAst, variable *AtomAst, body []*AtomAst, position AtomPosition) *AtomAst {
	ast := NewAtomAst(AstTypeCatchExpression, position)
	ast.Ast0 = condition
	ast.Ast1 = variable
	ast.Arr0 = body
	return ast
}

func NewImportStatement(path *AtomAst, names []*AtomAst, position AtomPosition) *AtomAst {
	ast := NewAtomAst(AstTypeImportStatement, position)
	ast.Ast0 = path
	ast.Arr0 = names
	return ast
}

func NewBreakStatement(position AtomPosition) *AtomAst {
	ast := NewAtomAst(AstTypeBreakStatement, position)
	return ast
}

func NewContinueStatement(position AtomPosition) *AtomAst {
	ast := NewAtomAst(AstTypeContinueStatement, position)
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

func NewClassStatement(name *AtomAst, base *AtomAst, body []*AtomAst, position AtomPosition) *AtomAst {
	ast := NewAtomAst(AstTypeClass, position)
	ast.Ast0 = name
	ast.Ast1 = base
	ast.Arr1 = body
	return ast
}

func NewEnumStatement(name *AtomAst, names []*AtomAst, values []*AtomAst, position AtomPosition) *AtomAst {
	ast := NewAtomAst(AstTypeEnum, position)
	ast.Ast0 = name
	ast.Arr0 = names
	ast.Arr1 = values
	return ast
}

func NewFunction(astType AtomAstType, name *AtomAst, params []*AtomAst, body []*AtomAst, position AtomPosition) *AtomAst {
	ast := NewAtomAst(astType, position)
	ast.Ast0 = name
	ast.Arr0 = params
	ast.Arr1 = body
	return ast
}

func NewBlock(body []*AtomAst, position AtomPosition) *AtomAst {
	ast := NewAtomAst(AstTypeBlock, position)
	ast.Arr0 = body
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

func NewSwitchStatement(condition *AtomAst, cases []*AtomAst, values []*AtomAst, value *AtomAst, position AtomPosition) *AtomAst {
	ast := NewAtomAst(AstTypeSwitchStatement, position)
	ast.Ast0 = condition
	ast.Ast1 = value
	ast.Arr0 = cases
	ast.Arr1 = values
	return ast
}

func NewWhileStatement(condition *AtomAst, body *AtomAst, position AtomPosition) *AtomAst {
	ast := NewAtomAst(AstTypeWhileStatement, position)
	ast.Ast0 = condition
	ast.Ast1 = body
	return ast
}

func NewDoWhileStatement(body *AtomAst, condition *AtomAst, position AtomPosition) *AtomAst {
	ast := NewAtomAst(AstTypeDoWhileStatement, position)
	ast.Ast0 = condition
	ast.Ast1 = body
	return ast
}

func NewForStatement(initializer *AtomAst, condition *AtomAst, updater *AtomAst, body *AtomAst, position AtomPosition) *AtomAst {
	ast := NewAtomAst(AstTypeForStatement, position)
	ast.Ast0 = initializer
	ast.Ast1 = condition
	ast.Ast2 = updater
	ast.Ast3 = body
	return ast
}

func NewProgram(body []*AtomAst, position AtomPosition) *AtomAst {
	ast := NewAtomAst(AstTypeProgram, position)
	ast.Arr1 = body
	return ast
}
