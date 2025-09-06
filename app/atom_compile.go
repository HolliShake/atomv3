package main

import (
	"encoding/binary"
	"fmt"
	"math"
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
	// convert int32 to 4 bytes using little-endian encoding
	bytes := make([]byte, 4)
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
	bytes := make([]byte, 8)
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
	bytes := make([]byte, len(strValue))
	copy(bytes, []byte(strValue))

	opcodes := make([]runtime.OpCode, len(bytes))
	for i, b := range bytes {
		opcodes[i] = runtime.OpCode(b)
	}

	atomFunc.Value.(*runtime.AtomCode).Code =
		append(
			append(atomFunc.Value.(*runtime.AtomCode).Code, opcode),
			opcodes...,
		)

	atomFunc.Value.(*runtime.AtomCode).Code =
		append(
			append(atomFunc.Value.(*runtime.AtomCode).Code, opcode),
			0, // null byte
		)
}

func (c *AtomCompile) emitJump(atomFunc *runtime.AtomValue, opcode runtime.OpCode) int {
	c.emit(atomFunc, opcode)
	start := len(atomFunc.Value.(*runtime.AtomCode).Code)
	// Emit 4 placeholder bytes for the jump address
	for i := 0; i < 4; i++ {
		c.emit(atomFunc, 0)
	}
	return start
}

func (c *AtomCompile) label(atomFunc *runtime.AtomValue, jumpAddress int) {
	current := len(atomFunc.Value.(*runtime.AtomCode).Code)
	for i := range 4 {
		atomFunc.Value.(*runtime.AtomCode).Code[i+jumpAddress] =
			runtime.OpCode((current >> (8 * i)) & 0xFF)
	}
}

func (c *AtomCompile) expression(parentScope *AtomScope, parentFunc *runtime.AtomValue, ast *AtomAst) {
	switch ast.AstType {
	case AstTypeIdn:
		{
			if !parentScope.HasSymbol(ast.Str0) {
				Error(
					c.parser.tokenizer.file,
					c.parser.tokenizer.data,
					fmt.Sprintf("Symbol %s not found", ast.Str0),
					ast.Position,
				)
				return
			}
			symbol := parentScope.GetSymbol(ast.Str0)
			if parentScope.HasLocal(ast.Str0) {
				c.emitInt(parentFunc, runtime.OpLoadLocal, symbol.Offset)
				return
			}
			// Non-local symbol, save as capture
			functionScope := parentScope.GetCurrentFunction()
			captureOffset := parentFunc.Value.(*runtime.AtomCode).IncrementCapture()
			functionScope.AddCapture(NewAtomSymbol(
				ast.Str0,
				captureOffset,
				false,
			))
			c.emitInt(parentFunc, runtime.OpLoadCapture, captureOffset)
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

	default:
		Error(
			c.parser.tokenizer.file,
			c.parser.tokenizer.data,
			"Expected expression",
			ast.Position,
		)
	}
}

func (c *AtomCompile) statement(parentScope *AtomScope, parentFunc *runtime.AtomValue, ast *AtomAst) {
	switch ast.AstType {
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

	case AstTypeIfStatement:
		c.ifStatement(
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

func (c *AtomCompile) returnStatement(parentScope *AtomScope, parentFunc *runtime.AtomValue, ast *AtomAst) {
	if !parentScope.InSide(AtomScopeTypeFunction) {
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
	// Guard
	if parentScope.Type != AtomScopeTypeGlobal {
		Error(
			c.parser.tokenizer.file,
			c.parser.tokenizer.data,
			"Function must be defined in global scope",
			ast.Position,
		)
		return nil
	}

	funcScope := NewAtomScope(parentScope, AtomScopeTypeFunction)
	atomFunc := runtime.NewFunction(c.parser.tokenizer.file, ast.Ast0.Str0, len(ast.Arr0))
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
	offset := parentFunc.Value.(*runtime.AtomCode).IncrementLocal()
	parentScope.AddSymbol(NewAtomSymbol(
		ast.Ast0.Str0,
		offset,
		parentScope.Type == AtomScopeTypeGlobal,
	))
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
		if funcScope.HasSymbol(param.Str0) {
			Error(
				c.parser.tokenizer.file,
				c.parser.tokenizer.data,
				fmt.Sprintf("Symbol %s already defined", param.Str0),
				param.Position,
			)
			return nil
		}
		// Save to symbol table
		offset := atomFunc.Value.(*runtime.AtomCode).IncrementLocal()
		funcScope.AddSymbol(NewAtomSymbol(
			param.Str0,
			offset,
			false,
		))
		c.emitInt(atomFunc, runtime.OpStoreLocal, offset)
	}
	body := ast.Arr1
	for _, stmt := range body {
		c.statement(funcScope, atomFunc, stmt)
	}
	c.emit(atomFunc, runtime.OpLoadNull)
	c.emit(atomFunc, runtime.OpReturn)
	//============================
	c.emitInt(parentFunc, runtime.OpLoadFunction, fnOffset)
	c.emit(parentFunc, runtime.OpDupTop)
	c.emitInt(parentFunc, runtime.OpStoreLocal, offset)
	// Write captures
	for _, capture := range funcScope.Captures {
		offset := 0
		opcode := runtime.OpLoadLocal
		if parentScope.HasLocal(capture.Name) {
			opcode = runtime.OpLoadLocal
			offset = parentScope.GetSymbol(capture.Name).Offset
		} else if parentScope.HasCapture(capture.Name) {
			opcode = runtime.OpLoadCapture
			offset = parentScope.GetCapture(capture.Name).Offset
		} else {
			panic(fmt.Sprintf("Symbol %s not found", capture.Name))
		}
		c.emitInt(parentFunc, opcode, offset)
		c.emitInt(parentFunc, runtime.OpStoreCapture, capture.Offset)
	}
	//============================
	atomFunc.Value.(*runtime.AtomCode).AllocateLocals()
	atomFunc.Value.(*runtime.AtomCode).AllocateCaptures()
	return atomFunc
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

func (c *AtomCompile) program(ast *AtomAst) *runtime.AtomValue {
	globalScope := NewAtomScope(nil, AtomScopeTypeGlobal)
	programFunc := runtime.NewFunction(c.parser.tokenizer.file, "main", 0)
	body := ast.Arr1
	for _, stmt := range body {
		c.statement(globalScope, programFunc, stmt)
	}
	c.emit(programFunc, runtime.OpLoadNull)
	c.emit(programFunc, runtime.OpReturn)
	programFunc.Value.(*runtime.AtomCode).AllocateLocals()
	programFunc.Value.(*runtime.AtomCode).AllocateCaptures()
	return programFunc
}

func (c *AtomCompile) Compile() *runtime.AtomValue {
	return c.program(c.parser.Parse())
}
