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

func (c *AtomCompile) identifier(parentScope *AtomScope, parentFunc *runtime.AtomValue, ast *AtomAst, opcode runtime.OpCode) {
	if !parentScope.HasSymbol(ast.Str0) {
		parentScope.Dump()
		Error(
			c.parser.tokenizer.file,
			c.parser.tokenizer.data,
			fmt.Sprintf("Symbol %s not found", ast.Str0),
			ast.Position,
		)
		return
	}
	symbol := parentScope.GetSymbol(ast.Str0)
	if parentScope.HasCapture(ast.Str0) {
		captureSymbol := parentScope.GetCapture(ast.Str0)
		c.emitInt(parentFunc, opcode, captureSymbol.Offset)
		return
	}
	if parentScope.HasLocal(ast.Str0) {
		c.emitInt(parentFunc, opcode, symbol.Offset)
		return
	}
	// Non-local symbol, save as capture
	functionScope := parentScope.GetCurrentFunction()

	if functionScope == nil {
		// Global?
		c.emitInt(parentFunc, opcode, symbol.Offset)
		return
	}

	captureOffset := parentFunc.Value.(*runtime.AtomCode).IncrementLocal()
	functionScope.AddCapture(NewCaptureAtomSymbol(
		ast.Str0,
		captureOffset,
		symbol.Global,
		symbol.Const,
		true,
	))
	c.emitInt(parentFunc, opcode, captureOffset)
}

func (c *AtomCompile) expression(parentScope *AtomScope, parentFunc *runtime.AtomValue, ast *AtomAst) {
	switch ast.AstType {
	case AstTypeIdn:
		{
			c.identifier(parentScope, parentFunc, ast, runtime.OpLoadLocal)
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
			parentFunc,
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
		c.emitNum(parentFunc, runtime.OpLoadNum, numValue)

	case AstTypeStr:
		c.emitStr(parentFunc, runtime.OpLoadStr, ast.Str0)

	case AstTypeBool:
		var boolValue byte
		if ast.Str0 == "true" {
			boolValue = 1
		} else {
			boolValue = 0
		}
		c.emitInt(parentFunc, runtime.OpLoadBool, int(boolValue))

	case AstTypeNull:
		c.emit(parentFunc, runtime.OpLoadNull)

	case AstTypeArray:
		{
			for i := len(ast.Arr0) - 1; i >= 0; i-- {
				element := ast.Arr0[i]
				c.expression(parentScope, parentFunc, element)
			}
			c.emitInt(parentFunc, runtime.OpLoadArray, len(ast.Arr0))
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

				c.expression(parentScope, parentFunc, v)
				c.emitStr(parentFunc, runtime.OpLoadStr, k.Str0)
			}
			c.emitInt(parentFunc, runtime.OpLoadObject, len(ast.Arr0))
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
			c.expression(parentScope, parentFunc, obj)
			c.emitStr(parentFunc, runtime.OpLoadStr, key.Str0)
			c.emit(parentFunc, runtime.OpIndex)
		}

	case AstTypeIndex:
		{
			obj := ast.Ast0
			index := ast.Ast1
			c.expression(parentScope, parentFunc, obj)
			c.expression(parentScope, parentFunc, index)
			c.emit(parentFunc, runtime.OpIndex)
		}

	case AstTypeCall:
		{
			funcAst := ast.Ast0
			args := ast.Arr0
			for i := len(args) - 1; i >= 0; i-- {
				c.expression(parentScope, parentFunc, args[i])
			}
			c.expression(parentScope, parentFunc, funcAst)
			c.emitInt(parentFunc, runtime.OpCall, len(args))
		}

	case AstTypeUnaryNot:
		{
			c.expression(parentScope, parentFunc, ast.Ast0)
			c.emit(parentFunc, runtime.OpNot)
		}

	case AstTypeUnaryNeg:
		{
			c.expression(parentScope, parentFunc, ast.Ast0)
			c.emit(parentFunc, runtime.OpNeg)
		}

	case AstTypeUnaryPos:
		{
			c.expression(parentScope, parentFunc, ast.Ast0)
			c.emit(parentFunc, runtime.OpPos)
		}

	case AstTypeBinaryMul:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(parentScope, parentFunc, lhs)
			c.expression(parentScope, parentFunc, rhs)
			c.emit(parentFunc, runtime.OpMul)
		}

	case AstTypeBinaryDiv:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(parentScope, parentFunc, lhs)
			c.expression(parentScope, parentFunc, rhs)
			c.emit(parentFunc, runtime.OpDiv)
		}

	case AstTypeBinaryMod:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(parentScope, parentFunc, lhs)
			c.expression(parentScope, parentFunc, rhs)
			c.emit(parentFunc, runtime.OpMod)
		}

	case AstTypeBinaryAdd:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(parentScope, parentFunc, lhs)
			c.expression(parentScope, parentFunc, rhs)
			c.emit(parentFunc, runtime.OpAdd)
		}

	case AstTypeBinarySub:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(parentScope, parentFunc, lhs)
			c.expression(parentScope, parentFunc, rhs)
			c.emit(parentFunc, runtime.OpSub)
		}

	case AstTypeBinaryShiftRight:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(parentScope, parentFunc, lhs)
			c.expression(parentScope, parentFunc, rhs)
			c.emit(parentFunc, runtime.OpShr)
		}

	case AstTypeBinaryShiftLeft:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(parentScope, parentFunc, lhs)
			c.expression(parentScope, parentFunc, rhs)
			c.emit(parentFunc, runtime.OpShl)
		}

	case AstTypeBinaryGreaterThan:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(parentScope, parentFunc, lhs)
			c.expression(parentScope, parentFunc, rhs)
			c.emit(parentFunc, runtime.OpCmpGt)
		}

	case AstTypeBinaryGreaterThanEqual:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(parentScope, parentFunc, lhs)
			c.expression(parentScope, parentFunc, rhs)
			c.emit(parentFunc, runtime.OpCmpGte)
		}

	case AstTypeBinaryLessThan:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(parentScope, parentFunc, lhs)
			c.expression(parentScope, parentFunc, rhs)
			c.emit(parentFunc, runtime.OpCmpLt)
		}

	case AstTypeBinaryLessThanEqual:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(parentScope, parentFunc, lhs)
			c.expression(parentScope, parentFunc, rhs)
			c.emit(parentFunc, runtime.OpCmpLte)
		}

	case AstTypeBinaryEqual:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(parentScope, parentFunc, lhs)
			c.expression(parentScope, parentFunc, rhs)
			c.emit(parentFunc, runtime.OpCmpEq)
		}

	case AstTypeBinaryNotEqual:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(parentScope, parentFunc, lhs)
			c.expression(parentScope, parentFunc, rhs)
			c.emit(parentFunc, runtime.OpCmpNe)
		}

	case AstTypeBinaryAnd:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(parentScope, parentFunc, lhs)
			c.expression(parentScope, parentFunc, rhs)
			c.emit(parentFunc, runtime.OpAnd)
		}

	case AstTypeBinaryOr:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(parentScope, parentFunc, lhs)
			c.expression(parentScope, parentFunc, rhs)
			c.emit(parentFunc, runtime.OpOr)
		}

	case AstTypeBinaryXor:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(parentScope, parentFunc, lhs)
			c.expression(parentScope, parentFunc, rhs)
			c.emit(parentFunc, runtime.OpXor)
		}

	case AstTypeLogicalAnd:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(parentScope, parentFunc, lhs)
			toEnd0 := c.emitJump(parentFunc, runtime.OpJumpIfFalseOrPop)
			c.expression(parentScope, parentFunc, rhs)
			c.label(parentFunc, toEnd0)
		}

	case AstTypeLogicalOr:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(parentScope, parentFunc, lhs)
			toEnd0 := c.emitJump(parentFunc, runtime.OpJumpIfTrueOrPop)
			c.expression(parentScope, parentFunc, rhs)
			c.label(parentFunc, toEnd0)
		}

	case AstTypeAssign:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(parentScope, parentFunc, rhs)
			c.emit(parentFunc, runtime.OpDupTop)
			c.assign(parentScope, parentFunc, lhs)
		}

	case AstTypeMulAssign:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(parentScope, parentFunc, lhs)
			c.expression(parentScope, parentFunc, rhs)
			c.emit(parentFunc, runtime.OpMul)
			c.emit(parentFunc, runtime.OpDupTop)
			c.assign(parentScope, parentFunc, lhs)
		}

	case AstTypeDivAssign:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(parentScope, parentFunc, lhs)
			c.expression(parentScope, parentFunc, rhs)
			c.emit(parentFunc, runtime.OpDiv)
			c.emit(parentFunc, runtime.OpDupTop)
			c.assign(parentScope, parentFunc, lhs)
		}

	case AstTypeModAssign:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(parentScope, parentFunc, lhs)
			c.expression(parentScope, parentFunc, rhs)
			c.emit(parentFunc, runtime.OpMod)
			c.emit(parentFunc, runtime.OpDupTop)
			c.assign(parentScope, parentFunc, lhs)
		}

	case AstTypeAddAssign:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(parentScope, parentFunc, lhs)
			c.expression(parentScope, parentFunc, rhs)
			c.emit(parentFunc, runtime.OpAdd)
			c.emit(parentFunc, runtime.OpDupTop)
			c.assign(parentScope, parentFunc, lhs)
		}

	case AstTypeSubAssign:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(parentScope, parentFunc, lhs)
			c.expression(parentScope, parentFunc, rhs)
			c.emit(parentFunc, runtime.OpSub)
			c.emit(parentFunc, runtime.OpDupTop)
			c.assign(parentScope, parentFunc, lhs)
		}

	case AstTypeLeftShiftAssign:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(parentScope, parentFunc, lhs)
			c.expression(parentScope, parentFunc, rhs)
			c.emit(parentFunc, runtime.OpShl)
			c.emit(parentFunc, runtime.OpDupTop)
			c.assign(parentScope, parentFunc, lhs)
		}

	case AstTypeRightShiftAssign:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(parentScope, parentFunc, lhs)
			c.expression(parentScope, parentFunc, rhs)
			c.emit(parentFunc, runtime.OpShr)
			c.emit(parentFunc, runtime.OpDupTop)
			c.assign(parentScope, parentFunc, lhs)
		}

	case AstTypeBitwiseAndAssign:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(parentScope, parentFunc, lhs)
			c.expression(parentScope, parentFunc, rhs)
			c.emit(parentFunc, runtime.OpAnd)
			c.emit(parentFunc, runtime.OpDupTop)
			c.assign(parentScope, parentFunc, lhs)
		}

	case AstTypeBitwiseOrAssign:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(parentScope, parentFunc, lhs)
			c.expression(parentScope, parentFunc, rhs)
			c.emit(parentFunc, runtime.OpOr)
			c.emit(parentFunc, runtime.OpDupTop)
			c.assign(parentScope, parentFunc, lhs)
		}

	case AstTypeBitwiseXorAssign:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(parentScope, parentFunc, lhs)
			c.expression(parentScope, parentFunc, rhs)
			c.emit(parentFunc, runtime.OpXor)
			c.emit(parentFunc, runtime.OpDupTop)
			c.assign(parentScope, parentFunc, lhs)
		}

	case AstTypeIfExpression:
		{
			condition := ast.Ast0
			thenValue := ast.Ast1
			elseValue := ast.Ast2

			c.expression(parentScope, parentFunc, condition)
			toElse := c.emitJump(parentFunc, runtime.OpPopJumpIfFalse)
			c.expression(parentScope, parentFunc, thenValue)
			toEnd := c.emitJump(parentFunc, runtime.OpJump)
			c.label(parentFunc, toElse)
			c.expression(parentScope, parentFunc, elseValue)
			c.label(parentFunc, toEnd)

		}

	case AstTypeSwitchExpression:
		{
			condition := ast.Ast0
			defaultValue := ast.Ast1
			cases := ast.Arr0
			values := ast.Arr1

			c.expression(parentScope, parentFunc, condition)

			toEndSwitch := []int{}

			for index, caseArray := range cases {
				cases := caseArray.Arr0
				value := values[index]
				storedJumps := []int{}
				for _, caseItem := range cases {
					c.expression(parentScope, parentFunc, caseItem)
					jumpToValue := c.emitJump(parentFunc, runtime.OpPeekJumpIfEqual)
					storedJumps = append(storedJumps, jumpToValue)
				}
				toNextCase := c.emitJump(parentFunc, runtime.OpJump)

				// value
				for _, jump := range storedJumps {
					c.label(parentFunc, jump)
				}
				// Pop condition if match
				c.emit(parentFunc, runtime.OpPopTop)

				// value
				c.expression(parentScope, parentFunc, value)
				jumpToEnd := c.emitJump(parentFunc, runtime.OpJump)
				toEndSwitch = append(toEndSwitch, jumpToEnd)

				// Next?
				c.label(parentFunc, toNextCase)
			}

			// Pop condition if default
			c.emit(parentFunc, runtime.OpPopTop)

			// Default value
			c.expression(parentScope, parentFunc, defaultValue)

			// End?
			for _, jump := range toEndSwitch {
				c.label(parentFunc, jump)
			}
		}

	case AstTypeCatchExpression:
		{
			condition := ast.Ast0
			variable := ast.Ast1
			body := ast.Arr0

			parentCode := parentFunc.Value.(*runtime.AtomCode)
			//==========================
			atomFunc := runtime.NewAtomValueFunction(c.parser.tokenizer.file, "catch", 1)
			atomCode := atomFunc.Value.(*runtime.AtomCode)
			funScope := NewAtomScope(parentScope, AtomScopeTypeFunction)
			fnOffset := c.state.SaveFunction(atomFunc)

			c.expression(parentScope, parentFunc, condition)
			toEndCatch := c.emitJump(parentFunc, runtime.OpPopJumpIfNotError)

			// Variable as parameter
			offset := atomCode.IncrementLocal()
			funScope.AddSymbol(NewAtomSymbol(
				variable.Str0,
				offset,
				false,
			))
			c.emitInt(atomFunc, runtime.OpStoreLocal, offset)

			// Body
			for _, stmt := range body {
				c.statement(funScope, atomFunc, stmt)
			}
			c.emit(atomFunc, runtime.OpLoadNull)
			c.emit(atomFunc, runtime.OpReturn)

			// Write captures
			for _, capture := range funScope.Captures() {
				offset := 0
				if parentScope.HasLocal(capture.Name) {
					offset = parentScope.GetSymbol(capture.Name).Offset
				} else {
					// Possible, not handled properly
					panic(fmt.Sprintf("Capture %s not found", capture.Name))
				}
				atomCode.CopyCellFrom(parentCode, offset, capture.Offset)
			}

			// Load and call
			c.emitInt(parentFunc, runtime.OpLoadFunction, fnOffset)
			c.emitInt(parentFunc, runtime.OpCall, 1)

			// End Catch
			c.label(parentFunc, toEndCatch)
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

func (c *AtomCompile) assign(parentScope *AtomScope, parentFunc *runtime.AtomValue, lhs *AtomAst) {
	switch lhs.AstType {
	case AstTypeIdn:
		{
			c.identifier(parentScope, parentFunc, lhs, runtime.OpStoreLocal)
		}

	case AstTypeMember:
		{
			c.expression(parentScope, parentFunc, lhs.Ast0)
			c.emitStr(parentFunc, runtime.OpLoadStr, lhs.Ast1.Str0)
			c.emit(parentFunc, runtime.OpSetIndex)
		}

	case AstTypeIndex:
		{
			c.expression(parentScope, parentFunc, lhs.Ast0)
			c.expression(parentScope, parentFunc, lhs.Ast1)
			c.emit(parentFunc, runtime.OpSetIndex)
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

func (c *AtomCompile) assign0(parentScope *AtomScope, parentFunc *runtime.AtomValue, lhs *AtomAst) {
	switch lhs.AstType {
	case AstTypeIdn:
		{
			c.expression(parentScope, parentFunc, lhs)
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

func (c *AtomCompile) statement(parentScope *AtomScope, parentFunc *runtime.AtomValue, ast *AtomAst) {
	switch ast.AstType {
	case AstTypeBreakStatement:
		c.breakStatement(
			parentScope,
			parentFunc,
			ast,
		)
	case AstTypeContinueStatement:
		c.continueStatement(
			parentScope,
			parentFunc,
			ast,
		)
	case AstTypeReturnStatement:
		c.returnStatement(
			parentScope,
			parentFunc,
			ast,
		)

	case AstTypeEmptyStatement:
		c.emptyStatement(
			parentScope,
			parentFunc,
			ast,
		)

	case AstTypeExpressionStatement:
		c.expressionStatement(
			parentScope,
			parentFunc,
			ast,
		)

	case AstTypeFunction:
		c.function(
			parentScope,
			parentFunc,
			ast,
		)

	case AstTypeBlock:
		c.block(
			parentScope,
			parentFunc,
			ast,
		)

	case AstTypeVarStatement:
		c.varStatement(
			parentScope,
			parentFunc,
			ast,
		)

	case AstTypeConstStatement:
		c.constStatement(
			parentScope,
			parentFunc,
			ast,
		)

	case AstTypeLocalStatement:
		c.localStatement(
			parentScope,
			parentFunc,
			ast,
		)

	case AstTypeImportStatement:
		c.importStatement(
			parentScope,
			parentFunc,
			ast,
		)

	case AstTypeIfStatement:
		c.ifStatement(
			parentScope,
			parentFunc,
			ast,
		)

	case AstTypeSwitchStatement:
		c.switchStatement(
			parentScope,
			parentFunc,
			ast,
		)

	case AstTypeWhileStatement:
		c.whileStatement(
			parentScope,
			parentFunc,
			ast,
		)

	case AstTypeDoWhileStatement:
		c.doWhileStatement(
			parentScope,
			parentFunc,
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

func (c *AtomCompile) breakStatement(parentScope *AtomScope, parentFunc *runtime.AtomValue, ast *AtomAst) {
	// Guard
	if !parentScope.InSide(AtomScopeTypeLoop, true) {
		Error(
			c.parser.tokenizer.file,
			c.parser.tokenizer.data,
			"Break statement must be inside a loop",
			ast.Position,
		)
		return
	}
	currentLoop := parentScope.GetCurrentLoop()
	currentLoop.AddBreak(
		c.emitJump(parentFunc, runtime.OpJump),
	)
}

func (c *AtomCompile) continueStatement(parentScope *AtomScope, parentFunc *runtime.AtomValue, ast *AtomAst) {
	// Guard
	if !parentScope.InSide(AtomScopeTypeLoop, true) {
		Error(
			c.parser.tokenizer.file,
			c.parser.tokenizer.data,
			"Continue statement must be inside a loop",
			ast.Position,
		)
		return
	}
	currentLoop := parentScope.GetCurrentLoop()
	currentLoop.AddContinue(
		c.emitJump(parentFunc, runtime.OpAbsoluteJump),
	)
}

func (c *AtomCompile) returnStatement(parentScope *AtomScope, parentFunc *runtime.AtomValue, ast *AtomAst) {
	// Guard
	if !parentScope.InSide(AtomScopeTypeFunction, true) {
		Error(
			c.parser.tokenizer.file,
			c.parser.tokenizer.data,
			"Return statement must be inside a function",
			ast.Position,
		)
		return
	}
	c.expression(parentScope, parentFunc, ast.Ast0)
	c.emit(parentFunc, runtime.OpReturn)
}

func (c *AtomCompile) emptyStatement(_ *AtomScope, parentFunc *runtime.AtomValue, _ *AtomAst) {
	c.emit(parentFunc, runtime.OpNoOp)
}

func (c *AtomCompile) expressionStatement(parentScope *AtomScope, parentFunc *runtime.AtomValue, ast *AtomAst) {
	c.expression(parentScope, parentFunc, ast.Ast0)
	c.emit(parentFunc, runtime.OpPopTop)
}

func (c *AtomCompile) function(parentScope *AtomScope, parentFunc *runtime.AtomValue, ast *AtomAst) *runtime.AtomValue {
	parentCode := parentFunc.Value.(*runtime.AtomCode)
	// Guard
	if !parentScope.InSide(AtomScopeTypeGlobal, false) {
		Error(
			c.parser.tokenizer.file,
			c.parser.tokenizer.data,
			"Function must be defined in global scope",
			ast.Position,
		)
		return nil
	}
	funScope := NewAtomScope(parentScope, AtomScopeTypeFunction)
	atomFunc := runtime.NewAtomValueFunction(c.parser.tokenizer.file, ast.Ast0.Str0, len(ast.Arr0))
	atomCode := atomFunc.Value.(*runtime.AtomCode)

	params := ast.Arr0
	//============================
	fnOffset := c.state.SaveFunction(atomFunc)
	if parentScope.HasLocal(ast.Ast0.Str0) {
		Error(
			c.parser.tokenizer.file,
			c.parser.tokenizer.data,
			fmt.Sprintf("Symbol %s already defined", ast.Ast0.Str0),
			ast.Ast0.Position,
		)
	}
	// Save to symbol table first to allow captures to reference it
	offset := parentCode.IncrementLocal()
	parentScope.AddSymbol(NewAtomSymbol(
		ast.Ast0.Str0,
		offset,
		parentScope.InSide(AtomScopeTypeGlobal, false),
	))
	c.emitInt(parentFunc, runtime.OpLoadFunction, fnOffset)
	c.emitInt(parentFunc, runtime.OpStoreLocal, offset)
	//============================

	for _, param := range params {
		if param.AstType != AstTypeIdn {
			Error(
				c.parser.tokenizer.file,
				c.parser.tokenizer.data,
				"Expected identifier",
				param.Position,
			)
			return nil
		}
		if funScope.HasLocal(param.Str0) {
			Error(
				c.parser.tokenizer.file,
				c.parser.tokenizer.data,
				fmt.Sprintf("Symbol %s already defined", param.Str0),
				param.Position,
			)
			return nil
		}
		// Save to symbol table
		offset := atomCode.IncrementLocal()
		funScope.AddSymbol(NewAtomSymbol(
			param.Str0,
			offset,
			false,
		))
		c.emitInt(atomFunc, runtime.OpStoreLocal, offset)
	}
	body := ast.Arr1
	for _, stmt := range body {
		c.statement(funScope, atomFunc, stmt)
	}
	c.emit(atomFunc, runtime.OpLoadNull)
	c.emit(atomFunc, runtime.OpReturn)

	// Write captures
	for _, capture := range funScope.Captures() {
		offset := 0
		if parentScope.HasLocal(capture.Name) {
			offset = parentScope.GetSymbol(capture.Name).Offset
		} else {
			// Possible, not handled properly
			panic(fmt.Sprintf("Capture %s not found", capture.Name))
		}
		atomCode.CopyCellFrom(parentCode, offset, capture.Offset)
	}

	return atomFunc
}

func (c *AtomCompile) block(parentScope *AtomScope, parentFunc *runtime.AtomValue, ast *AtomAst) {
	blockScope := NewAtomScope(parentScope, AtomScopeTypeBlock)
	for _, stmt := range ast.Arr1 {
		c.statement(blockScope, parentFunc, stmt)
	}
}

func (c *AtomCompile) varStatement(parentScope *AtomScope, parentFunc *runtime.AtomValue, ast *AtomAst) {
	if !parentScope.InSide(AtomScopeTypeGlobal, false) {
		Error(
			c.parser.tokenizer.file,
			c.parser.tokenizer.data,
			"Var statement must be in global scope",
			ast.Position,
		)
		return
	}
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
		if val == nil {
			c.emitInt(parentFunc, runtime.OpLoadNull, 0)
		} else {
			c.expression(parentScope, parentFunc, val)
		}
		if parentScope.HasLocal(key.Str0) {
			Error(
				c.parser.tokenizer.file,
				c.parser.tokenizer.data,
				fmt.Sprintf("Symbol %s already defined", key.Str0),
				key.Position,
			)
			return
		}
		offset := parentFunc.Value.(*runtime.AtomCode).IncrementLocal()
		parentScope.AddSymbol(NewAtomSymbol(
			key.Str0,
			offset,
			parentScope.InSide(AtomScopeTypeGlobal, false),
		))
		c.emitInt(parentFunc, runtime.OpStoreGlobal, offset)
	}
}

func (c *AtomCompile) constStatement(parentScope *AtomScope, parentFunc *runtime.AtomValue, ast *AtomAst) {
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
		if val == nil {
			c.emitInt(parentFunc, runtime.OpLoadNull, 0)
		} else {
			c.expression(parentScope, parentFunc, val)
		}
		if parentScope.HasLocal(key.Str0) {
			Error(
				c.parser.tokenizer.file,
				c.parser.tokenizer.data,
				fmt.Sprintf("Symbol %s already defined", key.Str0),
				key.Position,
			)
			return
		}
		offset := parentFunc.Value.(*runtime.AtomCode).IncrementLocal()
		parentScope.AddSymbol(NewConstAtomSymbol(
			key.Str0,
			offset,
			parentScope.InSide(AtomScopeTypeGlobal, false),
		))
		c.emitInt(parentFunc, runtime.OpStoreGlobal, offset)
	}
}

func (c *AtomCompile) localStatement(parentScope *AtomScope, parentFunc *runtime.AtomValue, ast *AtomAst) {
	if !parentScope.InSide(AtomScopeTypeBlock, false) && !parentScope.InSide(AtomScopeTypeFunction, false) {
		Error(
			c.parser.tokenizer.file,
			c.parser.tokenizer.data,
			"Local statement must be in block scope",
			ast.Position,
		)
		return
	}
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
		if val == nil {
			c.emitInt(parentFunc, runtime.OpLoadNull, 0)
		} else {
			c.expression(parentScope, parentFunc, val)
		}
		if parentScope.HasLocal(key.Str0) {
			Error(
				c.parser.tokenizer.file,
				c.parser.tokenizer.data,
				fmt.Sprintf("Symbol %s already defined", key.Str0),
				key.Position,
			)
			return
		}
		offset := parentFunc.Value.(*runtime.AtomCode).IncrementLocal()
		parentScope.AddSymbol(NewAtomSymbol(
			key.Str0,
			offset,
			parentScope.InSide(AtomScopeTypeGlobal, false),
		))
		c.emitInt(parentFunc, runtime.OpStoreLocal, offset)
	}
}

func (c *AtomCompile) importStatement(parentScope *AtomScope, parentFunc *runtime.AtomValue, ast *AtomAst) {
	// Guard
	if !parentScope.InSide(AtomScopeTypeGlobal, false) {
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
		c.emitStr(parentFunc, runtime.OpLoadModule0, normalizedPath)
	} else {
		c.emitStr(parentFunc, runtime.OpLoadModule1, path.Str0)
	}

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

		if parentScope.HasLocal(name.Str0) {
			Error(
				c.parser.tokenizer.file,
				c.parser.tokenizer.data,
				fmt.Sprintf("Symbol %s already defined", name.Str0),
				name.Position,
			)
			return
		}

		// Save
		offset := parentFunc.Value.(*runtime.AtomCode).IncrementLocal()
		parentScope.AddSymbol(NewAtomSymbol(
			name.Str0,
			offset,
			parentScope.InSide(AtomScopeTypeGlobal, false),
		))
		c.emitStr(parentFunc, runtime.OpPluckAttribute, name.Str0)
		c.emitInt(parentFunc, runtime.OpStoreGlobal, offset)
	}

	if parentScope.HasLocal(normalizedPath) {
		Error(
			c.parser.tokenizer.file,
			c.parser.tokenizer.data,
			fmt.Sprintf("Symbol %s already defined", normalizedPath),
			path.Position,
		)
		return
	}

	// Save to table
	offset := parentFunc.Value.(*runtime.AtomCode).IncrementLocal()
	parentScope.AddSymbol(NewAtomSymbol(
		normalizedPath,
		offset,
		parentScope.InSide(AtomScopeTypeGlobal, false),
	))
	c.emitInt(parentFunc, runtime.OpStoreGlobal, offset)
}

func (c *AtomCompile) ifStatement(parentScope *AtomScope, parentFunc *runtime.AtomValue, ast *AtomAst) {
	isLogical := ast.Ast0.AstType == AstTypeLogicalAnd || ast.Ast0.AstType == AstTypeLogicalOr
	if !isLogical {
		c.expression(parentScope, parentFunc, ast.Ast0)
		toElse := c.emitJump(parentFunc, runtime.OpPopJumpIfFalse)
		c.statement(parentScope, parentFunc, ast.Ast1)
		toEnd := c.emitJump(parentFunc, runtime.OpJump)
		c.label(parentFunc, toElse)
		if ast.Ast2 != nil {
			c.statement(parentScope, parentFunc, ast.Ast2)
		}
		c.label(parentFunc, toEnd)
	} else {
		isAnd := ast.Ast0.AstType == AstTypeLogicalAnd
		lhs := ast.Ast0.Ast0
		rhs := ast.Ast0.Ast1
		if isAnd {
			c.expression(parentScope, parentFunc, lhs)
			toEnd0 := c.emitJump(parentFunc, runtime.OpPopJumpIfFalse)
			c.expression(parentScope, parentFunc, rhs)
			toEnd1 := c.emitJump(parentFunc, runtime.OpPopJumpIfFalse)
			c.statement(parentScope, parentFunc, ast.Ast1)
			toEnd2 := c.emitJump(parentFunc, runtime.OpJump)
			c.label(parentFunc, toEnd0)
			c.label(parentFunc, toEnd1)
			if ast.Ast2 != nil {
				c.statement(parentScope, parentFunc, ast.Ast2)
			}
			c.label(parentFunc, toEnd2)
		} else {
			c.expression(parentScope, parentFunc, lhs)
			toThen := c.emitJump(parentFunc, runtime.OpPopJumpIfTrue)
			c.expression(parentScope, parentFunc, rhs)
			toElse := c.emitJump(parentFunc, runtime.OpPopJumpIfFalse)
			c.label(parentFunc, toThen)
			c.statement(parentScope, parentFunc, ast.Ast1)
			toEnd1 := c.emitJump(parentFunc, runtime.OpJump)
			c.label(parentFunc, toElse)
			if ast.Ast2 != nil {
				c.statement(parentScope, parentFunc, ast.Ast2)
			}
			c.label(parentFunc, toEnd1)
		}
	}
}

func (c *AtomCompile) switchStatement(parentScope *AtomScope, parentFunc *runtime.AtomValue, ast *AtomAst) {
	{
		condition := ast.Ast0
		defaultValue := ast.Ast1
		cases := ast.Arr0
		values := ast.Arr1

		c.expression(parentScope, parentFunc, condition)

		toEndSwitch := []int{}

		for index, caseArray := range cases {
			cases := caseArray.Arr0
			stmnt := values[index]
			storedJumps := []int{}
			for _, caseItem := range cases {
				c.expression(parentScope, parentFunc, caseItem)
				jumpToValue := c.emitJump(parentFunc, runtime.OpPeekJumpIfEqual)
				storedJumps = append(storedJumps, jumpToValue)
			}
			toNextCase := c.emitJump(parentFunc, runtime.OpJump)

			// value
			for _, jump := range storedJumps {
				c.label(parentFunc, jump)
			}
			// Pop condition if match
			c.emit(parentFunc, runtime.OpPopTop)

			// statement
			c.statement(parentScope, parentFunc, stmnt)
			jumpToEnd := c.emitJump(parentFunc, runtime.OpJump)
			toEndSwitch = append(toEndSwitch, jumpToEnd)

			// Next?
			c.label(parentFunc, toNextCase)
		}

		// Pop condition if default
		c.emit(parentFunc, runtime.OpPopTop)

		// Default value
		c.statement(parentScope, parentFunc, defaultValue)

		// End?
		for _, jump := range toEndSwitch {
			c.label(parentFunc, jump)
		}
	}
}

func (c *AtomCompile) whileStatement(parentScope *AtomScope, parentFunc *runtime.AtomValue, ast *AtomAst) {
	loopScope := NewAtomScope(parentScope, AtomScopeTypeLoop)
	isLogical := ast.Ast0.AstType == AstTypeLogicalAnd || ast.Ast0.AstType == AstTypeLogicalOr
	loopStart := c.here(parentFunc)
	if !isLogical {
		c.expression(loopScope, parentFunc, ast.Ast0)
		toEnd := c.emitJump(parentFunc, runtime.OpPopJumpIfFalse)
		c.statement(loopScope, parentFunc, ast.Ast1)
		c.emitInt(parentFunc, runtime.OpAbsoluteJump, loopStart)
		c.label(parentFunc, toEnd)
	} else {
		isAnd := ast.Ast0.AstType == AstTypeLogicalAnd
		lhs := ast.Ast0.Ast0
		rhs := ast.Ast0.Ast1
		if isAnd {
			c.expression(loopScope, parentFunc, lhs)
			toEnd0 := c.emitJump(parentFunc, runtime.OpPopJumpIfFalse)
			c.expression(loopScope, parentFunc, rhs)
			toEnd1 := c.emitJump(parentFunc, runtime.OpPopJumpIfFalse)
			c.statement(loopScope, parentFunc, ast.Ast1)
			c.emitInt(parentFunc, runtime.OpAbsoluteJump, loopStart)
			c.label(parentFunc, toEnd0)
			c.label(parentFunc, toEnd1)
		} else {
			c.expression(loopScope, parentFunc, lhs)
			toThen := c.emitJump(parentFunc, runtime.OpPopJumpIfTrue)
			c.expression(loopScope, parentFunc, rhs)
			toEnd1 := c.emitJump(parentFunc, runtime.OpPopJumpIfFalse)
			// Then?
			c.label(parentFunc, toThen)
			c.statement(loopScope, parentFunc, ast.Ast1)
			c.emitInt(parentFunc, runtime.OpAbsoluteJump, loopStart)
			c.label(parentFunc, toEnd1)
		}
	}

	for _, breakAddress := range loopScope.Breaks {
		c.label(parentFunc, breakAddress)
	}
	for _, continueAddress := range loopScope.Continues {
		c.labelContinue(parentFunc, continueAddress, loopStart)
	}
}

func (c *AtomCompile) doWhileStatement(parentScope *AtomScope, parentFunc *runtime.AtomValue, ast *AtomAst) {
	loopScope := NewAtomScope(parentScope, AtomScopeTypeLoop)
	isLogical := ast.Ast0.AstType == AstTypeLogicalAnd || ast.Ast0.AstType == AstTypeLogicalOr
	loopStart := c.here(parentFunc)
	if !isLogical {
		c.expression(loopScope, parentFunc, ast.Ast0)
		toEnd := c.emitJump(parentFunc, runtime.OpPopJumpIfFalse)
		c.statement(loopScope, parentFunc, ast.Ast1)
		c.emitInt(parentFunc, runtime.OpAbsoluteJump, loopStart)
		c.label(parentFunc, toEnd)
	} else {
		isAnd := ast.Ast0.AstType == AstTypeLogicalAnd
		lhs := ast.Ast0.Ast0
		rhs := ast.Ast0.Ast1
		if isAnd {
			c.expression(loopScope, parentFunc, lhs)
			toEnd0 := c.emitJump(parentFunc, runtime.OpPopJumpIfFalse)
			c.expression(loopScope, parentFunc, rhs)
			toEnd1 := c.emitJump(parentFunc, runtime.OpPopJumpIfFalse)
			c.statement(loopScope, parentFunc, ast.Ast1)
			c.emitInt(parentFunc, runtime.OpAbsoluteJump, loopStart)
			c.label(parentFunc, toEnd0)
			c.label(parentFunc, toEnd1)
		} else {
			c.expression(loopScope, parentFunc, lhs)
			toThen := c.emitJump(parentFunc, runtime.OpPopJumpIfTrue)
			c.expression(loopScope, parentFunc, rhs)
			toEnd1 := c.emitJump(parentFunc, runtime.OpPopJumpIfFalse)
			// Then?
			c.label(parentFunc, toThen)
			c.statement(loopScope, parentFunc, ast.Ast1)
			c.emitInt(parentFunc, runtime.OpAbsoluteJump, loopStart)
			c.label(parentFunc, toEnd1)
		}
	}

	for _, breakAddress := range loopScope.Breaks {
		c.label(parentFunc, breakAddress)
	}
	for _, continueAddress := range loopScope.Continues {
		c.labelContinue(parentFunc, continueAddress, loopStart)
	}
}

func (c *AtomCompile) program(ast *AtomAst) *runtime.AtomValue {
	globalScope := NewAtomScope(nil, AtomScopeTypeGlobal)
	programFunc := runtime.NewAtomValueFunction(c.parser.tokenizer.file, "main", 0)
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
