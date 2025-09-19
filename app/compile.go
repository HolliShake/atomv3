package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"regexp"
	"strconv"

	runtime "dev.runtime"
)

/*
 * Hide everything.
 */
type AtomCompile struct {
	state  *runtime.AtomState
	parser *AtomParser
}

func NewAtomCompile(parser *AtomParser, state *runtime.AtomState) *AtomCompile {
	return &AtomCompile{parser: parser, state: state}
}

func (c *AtomCompile) emit(atomFunc *runtime.AtomValue, opcode runtime.OpCode) {
	atomFunc.Value.(*runtime.AtomCode).Code =
		append(atomFunc.Value.(*runtime.AtomCode).Code, opcode)
}

func (c *AtomCompile) emitInt(atomFunc *runtime.AtomValue, opcode runtime.OpCode, intValue int) {
	// Convert int32 to 4 bytes using little-endian encoding
	bytes := []byte{0, 0, 0, 0}
	binary.LittleEndian.PutUint32(bytes, uint32(intValue))

	atomFunc.Value.(*runtime.AtomCode).Code =
		append(
			append(atomFunc.Value.(*runtime.AtomCode).Code, opcode),
			runtime.OpCode(bytes[0]),
			runtime.OpCode(bytes[1]),
			runtime.OpCode(bytes[2]),
			runtime.OpCode(bytes[3]),
		)
}

func (c *AtomCompile) emitNum(atomFunc *runtime.AtomValue, opcode runtime.OpCode, numValue float64) {
	bytes := []byte{0, 0, 0, 0, 0, 0, 0, 0}
	binary.LittleEndian.PutUint64(bytes, uint64(math.Float64bits(numValue)))

	atomFunc.Value.(*runtime.AtomCode).Code =
		append(
			append(atomFunc.Value.(*runtime.AtomCode).Code, opcode),
			runtime.OpCode(bytes[0]),
			runtime.OpCode(bytes[1]),
			runtime.OpCode(bytes[2]),
			runtime.OpCode(bytes[3]),
			runtime.OpCode(bytes[4]),
			runtime.OpCode(bytes[5]),
			runtime.OpCode(bytes[6]),
			runtime.OpCode(bytes[7]),
		)
}

func (c *AtomCompile) emitStr(atomFunc *runtime.AtomValue, opcode runtime.OpCode, strValue string) {
	bytes := []byte(strValue)

	opcodes := make([]runtime.OpCode, len(bytes)+1)
	for i, b := range bytes {
		opcodes[i] = runtime.OpCode(b)
	}
	opcodes[len(bytes)] = '\x00' // Null byte

	atomFunc.Value.(*runtime.AtomCode).Code =
		append(
			append(atomFunc.Value.(*runtime.AtomCode).Code, opcode),
			opcodes...,
		)
}

func (c *AtomCompile) emitWord(atomFunc *runtime.AtomValue, strValue string) {
	bytes := []byte(strValue)

	opcodes := make([]runtime.OpCode, len(bytes)+1)
	for i, b := range bytes {
		opcodes[i] = runtime.OpCode(b)
	}
	opcodes[len(bytes)] = '\x00' // Null byte

	atomFunc.Value.(*runtime.AtomCode).Code =
		append(
			atomFunc.Value.(*runtime.AtomCode).Code,
			opcodes...,
		)
}

func (c *AtomCompile) emitJump(atomFunc *runtime.AtomValue, opcode runtime.OpCode) int {
	c.emit(atomFunc, opcode)
	start := len(atomFunc.Value.(*runtime.AtomCode).Code)
	// Emit 4 placeholder bytes for the jump address
	for range 4 {
		c.emit(atomFunc, 0)
	}
	return start
}

func (c *AtomCompile) here(atomFunc *runtime.AtomValue) int {
	return len(atomFunc.Value.(*runtime.AtomCode).Code)
}

func (c *AtomCompile) label(atomFunc *runtime.AtomValue, jumpAddress int) {
	current := len(atomFunc.Value.(*runtime.AtomCode).Code)
	for i := range 4 {
		atomFunc.Value.(*runtime.AtomCode).Code[i+jumpAddress] =
			runtime.OpCode((current >> (8 * i)) & 0xFF)
	}
}

func (c *AtomCompile) labelContinue(atomFunc *runtime.AtomValue, continueStart int, to int) {
	for i := range 4 {
		atomFunc.Value.(*runtime.AtomCode).Code[i+continueStart] =
			runtime.OpCode((to >> (8 * i)) & 0xFF)
	}
}

func (c *AtomCompile) identifier(fn *runtime.AtomValue, ast *AtomAst, opcode runtime.OpCode) {
	c.emitStr(fn, opcode, ast.Str0)
}

func (c *AtomCompile) expression(scope *AtomScope, fn *runtime.AtomValue, ast *AtomAst) {
	switch ast.AstType {
	case AstTypeIdn:
		{
			c.identifier(fn, ast, runtime.OpLoadName)
		}

	case AstTypeInt:
		intValue, err := strconv.Atoi(ast.Str0)
		if err != nil {
			Error(
				c.parser.tokenizer.file,
				c.parser.tokenizer.data,
				"Invalid integer",
				ast.Position,
			)
		}
		c.emitInt(
			fn,
			runtime.OpLoadInt,
			intValue,
		)

	case AstTypeNum:
		numValue, err := strconv.ParseFloat(ast.Str0, 64)
		if err != nil {
			Error(
				c.parser.tokenizer.file,
				c.parser.tokenizer.data,
				"Invalid number",
				ast.Position,
			)
		}
		c.emitNum(fn, runtime.OpLoadNum, numValue)

	case AstTypeStr:
		c.emitStr(fn, runtime.OpLoadStr, ast.Str0)

	case AstTypeBool:
		var boolValue byte
		if ast.Str0 == "true" {
			boolValue = 1
		} else {
			boolValue = 0
		}
		c.emitInt(fn, runtime.OpLoadBool, int(boolValue))

	case AstTypeNull:
		c.emit(fn, runtime.OpLoadNull)

	case AstTypeArray:
		{
			for i := len(ast.Arr0) - 1; i >= 0; i-- {
				element := ast.Arr0[i]
				c.expression(scope, fn, element)
			}
			c.emitInt(fn, runtime.OpLoadArray, len(ast.Arr0))
		}

	case AstTypeObject:
		{
			for i := len(ast.Arr0) - 1; i >= 0; i-- {
				element := ast.Arr0[i]
				k := element.Ast0
				v := element.Ast1

				if k.AstType != AstTypeIdn {
					Error(
						c.parser.tokenizer.file,
						c.parser.tokenizer.data,
						"Expected identifier",
						k.Position,
					)
					return
				}

				c.expression(scope, fn, v)
				c.emitStr(fn, runtime.OpLoadStr, k.Str0)
			}
			c.emitInt(fn, runtime.OpLoadObject, len(ast.Arr0))
		}

	case AstTypeMember:
		{
			obj := ast.Ast0
			key := ast.Ast1
			if key.AstType != AstTypeIdn {
				Error(
					c.parser.tokenizer.file,
					c.parser.tokenizer.data,
					"Expected identifier",
					key.Position,
				)
				return
			}
			c.expression(scope, fn, obj)
			c.emitStr(fn, runtime.OpLoadStr, key.Str0)
			c.emit(fn, runtime.OpIndex)
		}

	case AstTypeIndex:
		{
			obj := ast.Ast0
			index := ast.Ast1
			c.expression(scope, fn, obj)
			c.expression(scope, fn, index)
			c.emit(fn, runtime.OpIndex)
		}

	case AstTypeCall:
		{
			funcAst := ast.Ast0
			args := ast.Arr0
			for i := len(args) - 1; i >= 0; i-- {
				c.expression(scope, fn, args[i])
			}
			c.expression(scope, fn, funcAst)
			c.emitInt(fn, runtime.OpCall, len(args))
		}

	case AstTypeAllocation:
		{
			ast0 := ast.Ast0
			if ast0.AstType != AstTypeCall {
				Error(
					c.parser.tokenizer.file,
					c.parser.tokenizer.data,
					"Expected call expression after new",
					ast0.Position,
				)
				return
			}

			constructorAst := ast0.Ast0
			args := ast0.Arr0
			for i := len(args) - 1; i >= 0; i-- {
				c.expression(scope, fn, args[i])
			}

			c.expression(scope, fn, constructorAst)
			c.emitInt(fn, runtime.OpCallConstructor, len(args))
		}

	case AstTypeUnaryNot:
		{
			c.expression(scope, fn, ast.Ast0)
			c.emit(fn, runtime.OpNot)
		}

	case AstTypeUnaryNeg:
		{
			c.expression(scope, fn, ast.Ast0)
			c.emit(fn, runtime.OpNeg)
		}

	case AstTypeUnaryPos:
		{
			c.expression(scope, fn, ast.Ast0)
			c.emit(fn, runtime.OpPos)
		}

	case AstTypeUnaryTypeof:
		{
			c.expression(scope, fn, ast.Ast0)
			c.emit(fn, runtime.OpTypeof)
		}

	case AstTypeUnaryAwait:
		{
			// Guard
			if !scope.InSide(AtomScopeTypeAsyncFunction, true) {
				Error(
					c.parser.tokenizer.file,
					c.parser.tokenizer.data,
					"Await must be in function scope",
					ast.Position,
				)
				return
			}
			callAst := ast.Ast0
			if callAst.AstType != AstTypeCall {
				Error(
					c.parser.tokenizer.file,
					c.parser.tokenizer.data,
					"Expected call expression after await",
					callAst.Position,
				)
				return
			}
			c.expression(scope, fn, callAst)
			c.emit(fn, runtime.OpAwait)
		}

	case AstTypeBinaryMul:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emit(fn, runtime.OpMul)
		}

	case AstTypeBinaryDiv:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emit(fn, runtime.OpDiv)
		}

	case AstTypeBinaryMod:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emit(fn, runtime.OpMod)
		}

	case AstTypeBinaryAdd:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emit(fn, runtime.OpAdd)
		}

	case AstTypeBinarySub:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emit(fn, runtime.OpSub)
		}

	case AstTypeBinaryShiftRight:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emit(fn, runtime.OpShr)
		}

	case AstTypeBinaryShiftLeft:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emit(fn, runtime.OpShl)
		}

	case AstTypeBinaryGreaterThan:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emit(fn, runtime.OpCmpGt)
		}

	case AstTypeBinaryGreaterThanEqual:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emit(fn, runtime.OpCmpGte)
		}

	case AstTypeBinaryLessThan:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emit(fn, runtime.OpCmpLt)
		}

	case AstTypeBinaryLessThanEqual:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emit(fn, runtime.OpCmpLte)
		}

	case AstTypeBinaryEqual:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emit(fn, runtime.OpCmpEq)
		}

	case AstTypeBinaryNotEqual:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emit(fn, runtime.OpCmpNe)
		}

	case AstTypeBinaryAnd:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emit(fn, runtime.OpAnd)
		}

	case AstTypeBinaryOr:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emit(fn, runtime.OpOr)
		}

	case AstTypeBinaryXor:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emit(fn, runtime.OpXor)
		}

	case AstTypeLogicalAnd:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			toEnd0 := c.emitJump(fn, runtime.OpJumpIfFalseOrPop)
			c.expression(scope, fn, rhs)
			c.label(fn, toEnd0)
		}

	case AstTypeLogicalOr:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			toEnd0 := c.emitJump(fn, runtime.OpJumpIfTrueOrPop)
			c.expression(scope, fn, rhs)
			c.label(fn, toEnd0)
		}

	case AstTypeAssign:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, rhs)
			c.emit(fn, runtime.OpDupTop)
			c.assign(scope, fn, lhs)
		}

	case AstTypeMulAssign:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emit(fn, runtime.OpMul)
			c.emit(fn, runtime.OpDupTop)
			c.assign(scope, fn, lhs)
		}

	case AstTypeDivAssign:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emit(fn, runtime.OpDiv)
			c.emit(fn, runtime.OpDupTop)
			c.assign(scope, fn, lhs)
		}

	case AstTypeModAssign:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emit(fn, runtime.OpMod)
			c.emit(fn, runtime.OpDupTop)
			c.assign(scope, fn, lhs)
		}

	case AstTypeAddAssign:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emit(fn, runtime.OpAdd)
			c.emit(fn, runtime.OpDupTop)
			c.assign(scope, fn, lhs)
		}

	case AstTypeSubAssign:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emit(fn, runtime.OpSub)
			c.emit(fn, runtime.OpDupTop)
			c.assign(scope, fn, lhs)
		}

	case AstTypeLeftShiftAssign:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emit(fn, runtime.OpShl)
			c.emit(fn, runtime.OpDupTop)
			c.assign(scope, fn, lhs)
		}

	case AstTypeRightShiftAssign:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emit(fn, runtime.OpShr)
			c.emit(fn, runtime.OpDupTop)
			c.assign(scope, fn, lhs)
		}

	case AstTypeBitwiseAndAssign:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emit(fn, runtime.OpAnd)
			c.emit(fn, runtime.OpDupTop)
			c.assign(scope, fn, lhs)
		}

	case AstTypeBitwiseOrAssign:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emit(fn, runtime.OpOr)
			c.emit(fn, runtime.OpDupTop)
			c.assign(scope, fn, lhs)
		}

	case AstTypeBitwiseXorAssign:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emit(fn, runtime.OpXor)
			c.emit(fn, runtime.OpDupTop)
			c.assign(scope, fn, lhs)
		}

	case AstTypeIfExpression:
		{
			condition := ast.Ast0
			thenValue := ast.Ast1
			elseValue := ast.Ast2

			c.expression(scope, fn, condition)
			toElse := c.emitJump(fn, runtime.OpPopJumpIfFalse)
			c.expression(scope, fn, thenValue)
			toEnd := c.emitJump(fn, runtime.OpJump)
			c.label(fn, toElse)
			c.expression(scope, fn, elseValue)
			c.label(fn, toEnd)

		}

	case AstTypeSwitchExpression:
		{
			condition := ast.Ast0
			defaultValue := ast.Ast1
			cases := ast.Arr0
			values := ast.Arr1

			c.expression(scope, fn, condition)

			toEndSwitch := []int{}

			for index, caseArray := range cases {
				cases := caseArray.Arr0
				value := values[index]
				storedJumps := []int{}
				for _, caseItem := range cases {
					c.expression(scope, fn, caseItem)
					jumpToValue := c.emitJump(fn, runtime.OpPeekJumpIfEqual)
					storedJumps = append(storedJumps, jumpToValue)
				}
				toNextCase := c.emitJump(fn, runtime.OpJump)

				// value
				for _, jump := range storedJumps {
					c.label(fn, jump)
				}
				// Pop condition if match
				c.emit(fn, runtime.OpPopTop)

				// value
				c.expression(scope, fn, value)
				jumpToEnd := c.emitJump(fn, runtime.OpJump)
				toEndSwitch = append(toEndSwitch, jumpToEnd)

				// Next?
				c.label(fn, toNextCase)
			}

			// Pop condition if default
			c.emit(fn, runtime.OpPopTop)

			// Default value
			c.expression(scope, fn, defaultValue)

			// End?
			for _, jump := range toEndSwitch {
				c.label(fn, jump)
			}
		}

	case AstTypeCatchExpression:
		{
			condition := ast.Ast0
			variable := ast.Ast1
			body := ast.Arr0

			//==========================
			atomFunc := runtime.NewAtomValueFunction(c.parser.tokenizer.file, "catch", false, 1)
			funScope := NewAtomScope(scope, AtomScopeTypeFunction)
			fnOffset := c.state.SaveFunction(atomFunc)

			c.expression(scope, fn, condition)
			toEndCatch := c.emitJump(fn, runtime.OpPopJumpIfNotError)

			// Variable as parameter
			c.emitStr(atomFunc, runtime.OpStoreFast, variable.Str0)

			// Body
			visibleReturn := false
			for _, stmt := range body {
				c.statement(funScope, atomFunc, stmt)
				if stmt.AstType == AstTypeReturnStatement {
					visibleReturn = true
					break
				}
			}
			if !visibleReturn {
				c.emit(atomFunc, runtime.OpLoadNull)
				c.emit(atomFunc, runtime.OpReturn)
			}

			// Load and call
			c.emitInt(fn, runtime.OpLoadFunction, fnOffset)
			c.emitInt(fn, runtime.OpCall, 1)

			// End Catch
			c.label(fn, toEndCatch)
		}

	default:
		Error(
			c.parser.tokenizer.file,
			c.parser.tokenizer.data,
			"Expected expression",
			ast.Position,
		)
	}
}

func (c *AtomCompile) assign(scope *AtomScope, fn *runtime.AtomValue, lhs *AtomAst) {
	switch lhs.AstType {
	case AstTypeIdn:
		{
			c.identifier(fn, lhs, runtime.OpStoreLocal)
		}

	case AstTypeMember:
		{
			c.expression(scope, fn, lhs.Ast0)
			c.emitStr(fn, runtime.OpLoadStr, lhs.Ast1.Str0)
			c.emit(fn, runtime.OpSetIndex)
		}

	case AstTypeIndex:
		{
			c.expression(scope, fn, lhs.Ast0)
			c.expression(scope, fn, lhs.Ast1)
			c.emit(fn, runtime.OpSetIndex)
		}

	default:
		{
			Error(
				c.parser.tokenizer.file,
				c.parser.tokenizer.data,
				"Expected identifier",
				lhs.Position,
			)
		}
	}
}

func (c *AtomCompile) statement(scope *AtomScope, fn *runtime.AtomValue, ast *AtomAst) {
	switch ast.AstType {
	case AstTypeBreakStatement:
		c.breakStatement(
			scope,
			fn,
			ast,
		)
	case AstTypeContinueStatement:
		c.continueStatement(
			scope,
			fn,
			ast,
		)
	case AstTypeReturnStatement:
		c.returnStatement(
			scope,
			fn,
			ast,
		)

	case AstTypeEmptyStatement:
		c.emptyStatement(
			scope,
			fn,
			ast,
		)

	case AstTypeExpressionStatement:
		c.expressionStatement(
			scope,
			fn,
			ast,
		)

	case AstTypeClass:
		c.classStatement(
			scope,
			fn,
			ast,
		)

	case AstTypeEnum:
		c.enumStatement(
			scope,
			fn,
			ast,
		)

	case AstTypeAsyncFunction,
		AstTypeFunction:
		c.function(
			scope,
			fn,
			ast,
			ast.AstType == AstTypeAsyncFunction,
		)

	case AstTypeBlock:
		c.block(
			scope,
			fn,
			ast,
		)

	case AstTypeVarStatement:
		c.varStatement(
			scope,
			fn,
			ast,
		)

	case AstTypeConstStatement:
		c.constStatement(
			scope,
			fn,
			ast,
		)

	case AstTypeLocalStatement:
		c.localStatement(
			scope,
			fn,
			ast,
		)

	case AstTypeImportStatement:
		c.importStatement(
			scope,
			fn,
			ast,
		)

	case AstTypeIfStatement:
		c.ifStatement(
			scope,
			fn,
			ast,
		)

	case AstTypeSwitchStatement:
		c.switchStatement(
			scope,
			fn,
			ast,
		)

	case AstTypeWhileStatement:
		c.whileStatement(
			scope,
			fn,
			ast,
		)

	case AstTypeDoWhileStatement:
		c.doWhileStatement(
			scope,
			fn,
			ast,
		)

	default:
		Error(
			c.parser.tokenizer.file,
			c.parser.tokenizer.data,
			"Expected statement",
			ast.Position,
		)
	}
}

func (c *AtomCompile) breakStatement(scope *AtomScope, fn *runtime.AtomValue, ast *AtomAst) {
	// Guard
	if !scope.InSide(AtomScopeTypeLoop, true) {
		Error(
			c.parser.tokenizer.file,
			c.parser.tokenizer.data,
			"Break statement must be inside a loop",
			ast.Position,
		)
		return
	}
	currentLoop := scope.GetCurrentLoop()
	currentLoop.AddBreak(
		c.emitJump(fn, runtime.OpJump),
	)
}

func (c *AtomCompile) continueStatement(scope *AtomScope, fn *runtime.AtomValue, ast *AtomAst) {
	// Guard
	if !scope.InSide(AtomScopeTypeLoop, true) {
		Error(
			c.parser.tokenizer.file,
			c.parser.tokenizer.data,
			"Continue statement must be inside a loop",
			ast.Position,
		)
		return
	}
	currentLoop := scope.GetCurrentLoop()
	currentLoop.AddContinue(
		c.emitJump(fn, runtime.OpAbsoluteJump),
	)
}

func (c *AtomCompile) returnStatement(scope *AtomScope, fn *runtime.AtomValue, ast *AtomAst) {
	// Guard
	if !(scope.InSide(AtomScopeTypeFunction, true) || scope.InSide(AtomScopeTypeAsyncFunction, true)) {
		Error(
			c.parser.tokenizer.file,
			c.parser.tokenizer.data,
			"Return statement must be inside a function",
			ast.Position,
		)
		return
	}
	if ast.Ast0 != nil {
		c.expression(scope, fn, ast.Ast0)
	} else {
		c.emit(fn, runtime.OpLoadNull)
	}
	c.emit(fn, runtime.OpReturn)
}

func (c *AtomCompile) emptyStatement(_ *AtomScope, fn *runtime.AtomValue, _ *AtomAst) {
	c.emit(fn, runtime.OpNoOp)
}

func (c *AtomCompile) expressionStatement(scope *AtomScope, fn *runtime.AtomValue, ast *AtomAst) {
	c.expression(scope, fn, ast.Ast0)
	c.emit(fn, runtime.OpPopTop)
}

func (c *AtomCompile) classStatement(scope *AtomScope, fn *runtime.AtomValue, ast *AtomAst) {
	// Guard
	if !scope.InSide(AtomScopeTypeGlobal, false) {
		Error(
			c.parser.tokenizer.file,
			c.parser.tokenizer.data,
			"Class statement must be in global scope",
			ast.Position,
		)
	}

	name := ast.Ast0
	base := ast.Ast1
	body := ast.Arr1

	if name.AstType != AstTypeIdn {
		Error(
			c.parser.tokenizer.file,
			c.parser.tokenizer.data,
			"Expected identifier",
			name.Position,
		)
	}

	if base != nil && base.AstType != AstTypeIdn {
		Error(
			c.parser.tokenizer.file,
			c.parser.tokenizer.data,
			"Expected identifier",
			base.Position,
		)
	}

	classScope := NewAtomScope(scope, AtomScopeTypeClass)

	//============================

	//============================

	items := 0

	// Body
	for _, stmt := range body {
		switch stmt.AstType {
		case AstTypeLocalStatement:
			items += len(stmt.Arr0)
			c.classVariable(classScope, fn, stmt)
		case AstTypeFunction:
			items += 1
			c.classFunction(classScope, fn, stmt)
		default:
			Error(
				c.parser.tokenizer.file,
				c.parser.tokenizer.data,
				"Expected function or variable declaration",
				stmt.Position,
			)
		}
	}

	//============================
	c.emitInt(fn, runtime.OpMakeClass, items)
	c.emitWord(fn, name.Str0)

	if base != nil {
		c.expression(scope, fn, base)
		c.emit(fn, runtime.OpExtendClass)
	}

	// Save
	c.emitStr(fn, runtime.OpInitVar, name.Str0)
	c.emit(fn, 1) // isGlobal is always true here
	c.emit(fn, 0) // Not constant
}

func (c *AtomCompile) classVariable(scope *AtomScope, fn *runtime.AtomValue, ast *AtomAst) {
	// Guard
	if !scope.InSide(AtomScopeTypeClass, false) {
		Error(
			c.parser.tokenizer.file,
			c.parser.tokenizer.data,
			"Variable must be defined in class scope",
			ast.Position,
		)
		return
	}

	for index, key := range ast.Arr0 {
		val := ast.Arr1[index]

		if key.AstType != AstTypeIdn {
			Error(
				c.parser.tokenizer.file,
				c.parser.tokenizer.data,
				"Expected identifier",
				key.Position,
			)
		}

		if val != nil {
			c.expression(scope, fn, val)
		} else {
			c.emit(fn, runtime.OpLoadNull)
		}

		c.emitStr(fn, runtime.OpLoadStr, key.Str0)
	}
}

func (c *AtomCompile) classFunction(scope *AtomScope, fn *runtime.AtomValue, ast *AtomAst) {
	// Guard
	if !scope.InSide(AtomScopeTypeClass, false) {
		Error(
			c.parser.tokenizer.file,
			c.parser.tokenizer.data,
			"Function must be defined in class scope",
			ast.Position,
		)
		return
	}

	funScope := NewAtomScope(scope, AtomScopeTypeFunction)
	atomFunc := runtime.NewAtomValueFunction(c.parser.tokenizer.file, ast.Ast0.Str0, false, len(ast.Arr0))

	params := ast.Arr0
	//============================
	fnOffset := c.state.SaveFunction(atomFunc)
	c.emitInt(fn, runtime.OpLoadFunction, fnOffset)
	c.emitStr(fn, runtime.OpLoadStr, ast.Ast0.Str0)
	//============================

	for _, param := range params {
		if param.AstType != AstTypeIdn {
			Error(
				c.parser.tokenizer.file,
				c.parser.tokenizer.data,
				"Expected identifier",
				param.Position,
			)
			return
		}

		// Save to symbol table
		c.emitStr(atomFunc, runtime.OpStoreFast, param.Str0)
	}
	body := ast.Arr1
	for _, stmt := range body {
		c.statement(funScope, atomFunc, stmt)
	}

	c.emit(atomFunc, runtime.OpLoadNull)
	c.emit(atomFunc, runtime.OpReturn)
}

func (c *AtomCompile) enumStatement(scope *AtomScope, fn *runtime.AtomValue, ast *AtomAst) {
	// Guard
	if !scope.InSide(AtomScopeTypeGlobal, false) {
		Error(
			c.parser.tokenizer.file,
			c.parser.tokenizer.data,
			"Enum statement must be in global scope",
			ast.Position,
		)
	}

	name := ast.Ast0
	names := ast.Arr0
	values := ast.Arr1

	if name.AstType != AstTypeIdn {
		Error(
			c.parser.tokenizer.file,
			c.parser.tokenizer.data,
			"Expected identifier",
			name.Position,
		)
		return
	}

	for index, name := range names {
		value := values[index]

		if name.AstType != AstTypeIdn {
			Error(
				c.parser.tokenizer.file,
				c.parser.tokenizer.data,
				"Expected identifier",
				name.Position,
			)
			return
		}

		if value == nil {
			c.emitInt(fn, runtime.OpLoadInt, index)
		} else {
			c.expression(scope, fn, value)
		}

		c.emitStr(fn, runtime.OpLoadStr, name.Str0)
	}

	c.emitInt(fn, runtime.OpMakeEnum, len(names))
	c.emitStr(fn, runtime.OpInitVar, name.Str0)
	c.emit(fn, 1) // isGlobal is always true here
	c.emit(fn, 0) // Not constant
}

func (c *AtomCompile) function(scope *AtomScope, fn *runtime.AtomValue, ast *AtomAst, async bool) {
	// Guard
	if !scope.InSide(AtomScopeTypeGlobal, false) {
		Error(
			c.parser.tokenizer.file,
			c.parser.tokenizer.data,
			"Function must be defined in global scope",
			ast.Position,
		)
		return
	}

	scopeType := AtomScopeTypeFunction
	if async {
		scopeType = AtomScopeTypeAsyncFunction
	}

	funScope := NewAtomScope(scope, scopeType)
	atomFunc := runtime.NewAtomValueFunction(c.parser.tokenizer.file, ast.Ast0.Str0, async, len(ast.Arr0))

	params := ast.Arr0
	//============================
	fnOffset := c.state.SaveFunction(atomFunc)

	// Save to symbol table first to allow captures to reference it
	c.emitInt(fn, runtime.OpLoadFunction, fnOffset)
	c.emitStr(fn, runtime.OpInitVar, ast.Ast0.Str0)
	c.emit(fn, 1) // isGlobal is always true here
	c.emit(fn, 0) // Not constant
	//============================

	for _, param := range params {
		if param.AstType != AstTypeIdn {
			Error(
				c.parser.tokenizer.file,
				c.parser.tokenizer.data,
				"Expected identifier",
				param.Position,
			)
			return
		}

		// Save to symbol table
		c.emitStr(atomFunc, runtime.OpStoreFast, param.Str0)
	}
	body := ast.Arr1
	visibleReturn := false
	for _, stmt := range body {
		c.statement(funScope, atomFunc, stmt)
		if stmt.AstType == AstTypeReturnStatement {
			visibleReturn = true
			break
		}
	}

	if !visibleReturn {
		c.emit(atomFunc, runtime.OpLoadNull)
		c.emit(atomFunc, runtime.OpReturn)
	}
}

func (c *AtomCompile) block(scope *AtomScope, fn *runtime.AtomValue, ast *AtomAst) {
	blockScope := NewAtomScope(scope, AtomScopeTypeBlock)
	c.emit(fn, runtime.OpEnterBlock)
	for _, stmt := range ast.Arr1 {
		c.statement(blockScope, fn, stmt)
	}
	c.emit(fn, runtime.OpExitBlock)
}

func (c *AtomCompile) varStatement(scope *AtomScope, fn *runtime.AtomValue, ast *AtomAst) {
	if !scope.InSide(AtomScopeTypeGlobal, false) {
		Error(
			c.parser.tokenizer.file,
			c.parser.tokenizer.data,
			"Var statement must be in global scope",
			ast.Position,
		)
		return
	}

	seenNames := map[string]bool{}

	for idx, key := range ast.Arr0 {
		val := ast.Arr1[idx]

		if key.AstType != AstTypeIdn {
			Error(
				c.parser.tokenizer.file,
				c.parser.tokenizer.data,
				"Expected identifier",
				key.Position,
			)
			return
		}

		if seenNames[key.Str0] {
			Error(
				c.parser.tokenizer.file,
				c.parser.tokenizer.data,
				fmt.Sprintf("Duplicate identifier: %s", key.Str0),
				key.Position,
			)
		}
		seenNames[key.Str0] = true

		if val == nil {
			c.emitInt(fn, runtime.OpLoadNull, 0)
		} else {
			c.expression(scope, fn, val)
		}

		c.emitStr(fn, runtime.OpInitVar, key.Str0)
		c.emit(fn, 1) // isGlobal is always true here
		c.emit(fn, 0) // Not constant
	}
}

func (c *AtomCompile) constStatement(scope *AtomScope, fn *runtime.AtomValue, ast *AtomAst) {
	isGlobal := scope.InSide(AtomScopeTypeGlobal, false)

	seenNames := map[string]bool{}

	for idx, key := range ast.Arr0 {
		val := ast.Arr1[idx]

		if key.AstType != AstTypeIdn {
			Error(
				c.parser.tokenizer.file,
				c.parser.tokenizer.data,
				"Expected identifier",
				key.Position,
			)
			return
		}

		if seenNames[key.Str0] {
			Error(
				c.parser.tokenizer.file,
				c.parser.tokenizer.data,
				fmt.Sprintf("Duplicate identifier: %s", key.Str0),
				key.Position,
			)
		}
		seenNames[key.Str0] = true

		if val == nil {
			c.emitInt(fn, runtime.OpLoadNull, 0)
		} else {
			c.expression(scope, fn, val)
		}

		c.emitStr(fn, runtime.OpInitVar, key.Str0)
		if isGlobal {
			c.emit(fn, 1)
		} else {
			c.emit(fn, 0)
		}
		c.emit(fn, 1) // Constant
	}
}

func (c *AtomCompile) localStatement(scope *AtomScope, fn *runtime.AtomValue, ast *AtomAst) {
	if !scope.InSide(AtomScopeTypeBlock, false) && !(scope.InSide(AtomScopeTypeFunction, false) || scope.InSide(AtomScopeTypeAsyncFunction, false)) {
		Error(
			c.parser.tokenizer.file,
			c.parser.tokenizer.data,
			"Local statement must be in block scope",
			ast.Position,
		)
		return
	}

	seenNames := map[string]bool{}

	for idx, key := range ast.Arr0 {
		val := ast.Arr1[idx]

		if key.AstType != AstTypeIdn {
			Error(
				c.parser.tokenizer.file,
				c.parser.tokenizer.data,
				"Expected identifier",
				key.Position,
			)
		}

		if seenNames[key.Str0] {
			Error(
				c.parser.tokenizer.file,
				c.parser.tokenizer.data,
				fmt.Sprintf("Duplicate identifier: %s", key.Str0),
				key.Position,
			)
		}
		seenNames[key.Str0] = true

		if val == nil {
			c.emitInt(fn, runtime.OpLoadNull, 0)
		} else {
			c.expression(scope, fn, val)
		}

		c.emitStr(fn, runtime.OpInitVar, key.Str0)
		c.emit(fn, 0) // isGlobal is always true here
		c.emit(fn, 0) // Not constant
	}
}

func (c *AtomCompile) importStatement(scope *AtomScope, fn *runtime.AtomValue, ast *AtomAst) {
	// Guard
	if !scope.InSide(AtomScopeTypeGlobal, false) {
		Error(
			c.parser.tokenizer.file,
			c.parser.tokenizer.data,
			"Import statement must be in global scope",
			ast.Position,
		)
		return
	}
	path := ast.Ast0
	names := ast.Arr0
	if path.AstType != AstTypeStr {
		Error(
			c.parser.tokenizer.file,
			c.parser.tokenizer.data,
			"Expected string",
			path.Position,
		)
		return
	}

	isBuiltin := func(name string) bool {
		// match if starts with 'atom:' and followed by module name using regex
		re := regexp.MustCompile(`^atom:([a-zA-Z_][a-zA-Z0-9_]*)$`)
		return re.MatchString(name)
	}

	validIdentifier := func(name string) bool {
		re := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
		return re.MatchString(name)
	}

	cleanNameWithoutExtension := func(name string) string {
		// Remove "atom:" prefix if it exists
		if isBuiltin(name) {
			re := regexp.MustCompile(`^atom:([a-zA-Z_][a-zA-Z0-9_]*)$`)
			name = re.ReplaceAllString(name, "$1")
		}

		re := regexp.MustCompile(`^([a-zA-Z_][a-zA-Z0-9_]*)\.atom$`)
		matches := re.FindStringSubmatch(name)
		if len(matches) > 1 {
			return matches[1]
		}
		return name
	}

	if !validIdentifier(cleanNameWithoutExtension(path.Str0)) {
		Error(
			c.parser.tokenizer.file,
			c.parser.tokenizer.data,
			"Invalid identifier",
			path.Position,
		)
		return
	}

	normalizedPath := cleanNameWithoutExtension(path.Str0)

	if isBuiltin(path.Str0) {
		c.emitStr(fn, runtime.OpLoadModule0, normalizedPath)
	} else {
		c.emitStr(fn, runtime.OpLoadModule1, path.Str0)
	}

	seenNames := make(map[string]bool)

	for _, name := range names {
		if name.AstType != AstTypeIdn {
			Error(
				c.parser.tokenizer.file,
				c.parser.tokenizer.data,
				"Expected identifier",
				name.Position,
			)
			return
		}

		// Check for duplicate names
		if seenNames[name.Str0] {
			Error(
				c.parser.tokenizer.file,
				c.parser.tokenizer.data,
				fmt.Sprintf("Duplicate identifier: %s", name.Str0),
				name.Position,
			)
			return
		}
		seenNames[name.Str0] = true

		// Save
		c.emitStr(fn, runtime.OpPluckAttribute, name.Str0)
		c.emitStr(fn, runtime.OpInitVar, name.Str0)
		c.emit(fn, 1) // isGlobal is always true here
		c.emit(fn, 0) // Not constant
	}

	// Save to table
	c.emitStr(fn, runtime.OpInitVar, normalizedPath)
	c.emit(fn, 1) // isGlobal is always true here
	c.emit(fn, 0) // Not constant
}

func (c *AtomCompile) ifStatement(scope *AtomScope, fn *runtime.AtomValue, ast *AtomAst) {
	isLogical := ast.Ast0.AstType == AstTypeLogicalAnd || ast.Ast0.AstType == AstTypeLogicalOr
	if !isLogical {
		c.expression(scope, fn, ast.Ast0)
		toElse := c.emitJump(fn, runtime.OpPopJumpIfFalse)
		c.statement(scope, fn, ast.Ast1)
		toEnd := c.emitJump(fn, runtime.OpJump)
		c.label(fn, toElse)
		if ast.Ast2 != nil {
			c.statement(scope, fn, ast.Ast2)
		}
		c.label(fn, toEnd)
	} else {
		isAnd := ast.Ast0.AstType == AstTypeLogicalAnd
		lhs := ast.Ast0.Ast0
		rhs := ast.Ast0.Ast1
		if isAnd {
			c.expression(scope, fn, lhs)
			toEnd0 := c.emitJump(fn, runtime.OpPopJumpIfFalse)
			c.expression(scope, fn, rhs)
			toEnd1 := c.emitJump(fn, runtime.OpPopJumpIfFalse)
			c.statement(scope, fn, ast.Ast1)
			toEnd2 := c.emitJump(fn, runtime.OpJump)
			c.label(fn, toEnd0)
			c.label(fn, toEnd1)
			if ast.Ast2 != nil {
				c.statement(scope, fn, ast.Ast2)
			}
			c.label(fn, toEnd2)
		} else {
			c.expression(scope, fn, lhs)
			toThen := c.emitJump(fn, runtime.OpPopJumpIfTrue)
			c.expression(scope, fn, rhs)
			toElse := c.emitJump(fn, runtime.OpPopJumpIfFalse)
			c.label(fn, toThen)
			c.statement(scope, fn, ast.Ast1)
			toEnd1 := c.emitJump(fn, runtime.OpJump)
			c.label(fn, toElse)
			if ast.Ast2 != nil {
				c.statement(scope, fn, ast.Ast2)
			}
			c.label(fn, toEnd1)
		}
	}
}

func (c *AtomCompile) switchStatement(scope *AtomScope, fn *runtime.AtomValue, ast *AtomAst) {
	{
		condition := ast.Ast0
		defaultValue := ast.Ast1
		cases := ast.Arr0
		values := ast.Arr1

		c.expression(scope, fn, condition)

		toEndSwitch := []int{}

		for index, caseArray := range cases {
			cases := caseArray.Arr0
			stmnt := values[index]
			storedJumps := []int{}
			for _, caseItem := range cases {
				c.expression(scope, fn, caseItem)
				jumpToValue := c.emitJump(fn, runtime.OpPeekJumpIfEqual)
				storedJumps = append(storedJumps, jumpToValue)
			}
			toNextCase := c.emitJump(fn, runtime.OpJump)

			// value
			for _, jump := range storedJumps {
				c.label(fn, jump)
			}
			// Pop condition if match
			c.emit(fn, runtime.OpPopTop)

			// statement
			c.statement(scope, fn, stmnt)
			jumpToEnd := c.emitJump(fn, runtime.OpJump)
			toEndSwitch = append(toEndSwitch, jumpToEnd)

			// Next?
			c.label(fn, toNextCase)
		}

		// Pop condition if default
		c.emit(fn, runtime.OpPopTop)

		// Default value
		c.statement(scope, fn, defaultValue)

		// End?
		for _, jump := range toEndSwitch {
			c.label(fn, jump)
		}
	}
}

func (c *AtomCompile) whileStatement(scope *AtomScope, fn *runtime.AtomValue, ast *AtomAst) {
	loopScope := NewAtomScope(scope, AtomScopeTypeLoop)
	isLogical := ast.Ast0.AstType == AstTypeLogicalAnd || ast.Ast0.AstType == AstTypeLogicalOr
	loopStart := c.here(fn)
	if !isLogical {
		c.expression(loopScope, fn, ast.Ast0)
		toEnd := c.emitJump(fn, runtime.OpPopJumpIfFalse)
		c.statement(loopScope, fn, ast.Ast1)
		c.emitInt(fn, runtime.OpAbsoluteJump, loopStart)
		c.label(fn, toEnd)
	} else {
		isAnd := ast.Ast0.AstType == AstTypeLogicalAnd
		lhs := ast.Ast0.Ast0
		rhs := ast.Ast0.Ast1
		if isAnd {
			c.expression(loopScope, fn, lhs)
			toEnd0 := c.emitJump(fn, runtime.OpPopJumpIfFalse)
			c.expression(loopScope, fn, rhs)
			toEnd1 := c.emitJump(fn, runtime.OpPopJumpIfFalse)
			c.statement(loopScope, fn, ast.Ast1)
			c.emitInt(fn, runtime.OpAbsoluteJump, loopStart)
			c.label(fn, toEnd0)
			c.label(fn, toEnd1)
		} else {
			c.expression(loopScope, fn, lhs)
			toThen := c.emitJump(fn, runtime.OpPopJumpIfTrue)
			c.expression(loopScope, fn, rhs)
			toEnd1 := c.emitJump(fn, runtime.OpPopJumpIfFalse)
			// Then?
			c.label(fn, toThen)
			c.statement(loopScope, fn, ast.Ast1)
			c.emitInt(fn, runtime.OpAbsoluteJump, loopStart)
			c.label(fn, toEnd1)
		}
	}

	for _, breakAddress := range loopScope.Breaks {
		c.label(fn, breakAddress)
	}
	for _, continueAddress := range loopScope.Continues {
		c.labelContinue(fn, continueAddress, loopStart)
	}
}

func (c *AtomCompile) doWhileStatement(scope *AtomScope, fn *runtime.AtomValue, ast *AtomAst) {
	loopScope := NewAtomScope(scope, AtomScopeTypeLoop)
	isLogical := ast.Ast0.AstType == AstTypeLogicalAnd || ast.Ast0.AstType == AstTypeLogicalOr
	loopStart := c.here(fn)
	if !isLogical {
		c.expression(loopScope, fn, ast.Ast0)
		toEnd := c.emitJump(fn, runtime.OpPopJumpIfFalse)
		c.statement(loopScope, fn, ast.Ast1)
		c.emitInt(fn, runtime.OpAbsoluteJump, loopStart)
		c.label(fn, toEnd)
	} else {
		isAnd := ast.Ast0.AstType == AstTypeLogicalAnd
		lhs := ast.Ast0.Ast0
		rhs := ast.Ast0.Ast1
		if isAnd {
			c.expression(loopScope, fn, lhs)
			toEnd0 := c.emitJump(fn, runtime.OpPopJumpIfFalse)
			c.expression(loopScope, fn, rhs)
			toEnd1 := c.emitJump(fn, runtime.OpPopJumpIfFalse)
			c.statement(loopScope, fn, ast.Ast1)
			c.emitInt(fn, runtime.OpAbsoluteJump, loopStart)
			c.label(fn, toEnd0)
			c.label(fn, toEnd1)
		} else {
			c.expression(loopScope, fn, lhs)
			toThen := c.emitJump(fn, runtime.OpPopJumpIfTrue)
			c.expression(loopScope, fn, rhs)
			toEnd1 := c.emitJump(fn, runtime.OpPopJumpIfFalse)
			// Then?
			c.label(fn, toThen)
			c.statement(loopScope, fn, ast.Ast1)
			c.emitInt(fn, runtime.OpAbsoluteJump, loopStart)
			c.label(fn, toEnd1)
		}
	}

	for _, breakAddress := range loopScope.Breaks {
		c.label(fn, breakAddress)
	}
	for _, continueAddress := range loopScope.Continues {
		c.labelContinue(fn, continueAddress, loopStart)
	}
}

func (c *AtomCompile) program(ast *AtomAst) *runtime.AtomValue {
	globalScope := NewAtomScope(nil, AtomScopeTypeGlobal)
	programFunc := runtime.NewAtomValueFunction(c.parser.tokenizer.file, "script", false, 0)
	body := ast.Arr1
	for _, stmt := range body {
		c.statement(globalScope, programFunc, stmt)
	}
	c.emit(programFunc, runtime.OpLoadNull)
	c.emit(programFunc, runtime.OpReturn)
	return programFunc
}

func (c *AtomCompile) Compile() *runtime.AtomValue {
	return c.program(c.parser.Parse())
}
