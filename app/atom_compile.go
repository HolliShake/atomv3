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
	atomFunc.Value.(*runtime.AtomCode).OpCodes =
		append(atomFunc.Value.(*runtime.AtomCode).OpCodes, opcode)
}

func (c *AtomCompile) emitInt(atomFunc *runtime.AtomValue, opcode runtime.OpCode, intValue int) {
	// convert int32 to 4 bytes using little-endian encoding
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, uint32(intValue))

	atomFunc.Value.(*runtime.AtomCode).OpCodes =
		append(
			append(atomFunc.Value.(*runtime.AtomCode).OpCodes, opcode),
			runtime.OpCode(bytes[0]),
			runtime.OpCode(bytes[1]),
			runtime.OpCode(bytes[2]),
			runtime.OpCode(bytes[3]),
		)
}

func (c *AtomCompile) emitNum(atomFunc *runtime.AtomValue, opcode runtime.OpCode, numValue float64) {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, uint64(math.Float64bits(numValue)))

	atomFunc.Value.(*runtime.AtomCode).OpCodes =
		append(
			append(atomFunc.Value.(*runtime.AtomCode).OpCodes, opcode),
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

	atomFunc.Value.(*runtime.AtomCode).OpCodes =
		append(
			append(atomFunc.Value.(*runtime.AtomCode).OpCodes, opcode),
			opcodes...,
		)

	atomFunc.Value.(*runtime.AtomCode).OpCodes =
		append(
			append(atomFunc.Value.(*runtime.AtomCode).OpCodes, opcode),
			0, // null byte
		)
}

func (c *AtomCompile) emitJump(atomFunc *runtime.AtomValue, opcode runtime.OpCode) int {
	c.emit(atomFunc, opcode)
	start := len(atomFunc.Value.(*runtime.AtomCode).OpCodes)
	// Emit 4 placeholder bytes for the jump address
	for i := 0; i < 4; i++ {
		c.emit(atomFunc, 0)
	}
	return start
}

func (c *AtomCompile) label(atomFunc *runtime.AtomValue, jumpAddress int) {
	current := len(atomFunc.Value.(*runtime.AtomCode).OpCodes)
	for i := range 4 {
		atomFunc.Value.(*runtime.AtomCode).OpCodes[i+jumpAddress] =
			runtime.OpCode((current >> (8 * i)) & 0xFF)
	}
}

func (c *AtomCompile) expression(parent *AtomScope, atomFunc *runtime.AtomValue, ast *AtomAst) {
	switch ast.AstType {
	case AstTypeIdn:
		if !parent.HasSymbol(ast.Str0) {
			Error(
				c.parser.tokenizer.file,
				c.parser.tokenizer.data,
				fmt.Sprintf("Symbol %s not found", ast.Str0),
				ast.Position,
			)
			return
		}
		symbol := parent.GetSymbol(ast.Str0)
		c.emitInt(atomFunc, runtime.OpLoadLocal, symbol.Offset)

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
			atomFunc,
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
		c.emitNum(atomFunc, runtime.OpLoadNum, numValue)

	case AstTypeStr:
		c.emitStr(atomFunc, runtime.OpLoadStr, ast.Str0)

	case AstTypeCall:
		{
			funcAst := ast.Ast0
			args := ast.Arr0
			for i := len(args) - 1; i >= 0; i-- {
				c.expression(parent, atomFunc, args[i])
			}
			c.expression(parent, atomFunc, funcAst)
			c.emitInt(atomFunc, runtime.OpCall, len(args))
		}

	case AstTypeBinaryMul:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(parent, atomFunc, lhs)
			c.expression(parent, atomFunc, rhs)
			c.emit(atomFunc, runtime.OpMul)
		}

	case AstTypeBinaryDiv:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(parent, atomFunc, lhs)
			c.expression(parent, atomFunc, rhs)
			c.emit(atomFunc, runtime.OpDiv)
		}

	case AstTypeBinaryMod:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(parent, atomFunc, lhs)
			c.expression(parent, atomFunc, rhs)
			c.emit(atomFunc, runtime.OpMod)
		}

	case AstTypeBinaryAdd:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(parent, atomFunc, lhs)
			c.expression(parent, atomFunc, rhs)
			c.emit(atomFunc, runtime.OpAdd)
		}

	case AstTypeBinarySub:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(parent, atomFunc, lhs)
			c.expression(parent, atomFunc, rhs)
			c.emit(atomFunc, runtime.OpSub)
		}

	case AstTypeBinaryShiftRight:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(parent, atomFunc, lhs)
			c.expression(parent, atomFunc, rhs)
			c.emit(atomFunc, runtime.OpShr)
		}

	case AstTypeBinaryShiftLeft:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(parent, atomFunc, lhs)
			c.expression(parent, atomFunc, rhs)
			c.emit(atomFunc, runtime.OpShl)
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

func (c *AtomCompile) statement(parent *AtomScope, atomFunc *runtime.AtomValue, ast *AtomAst) {
	switch ast.AstType {
	case AstTypeReturnStatement:
		c.returnStatement(
			parent,
			atomFunc,
			ast,
		)

	case AstTypeEmptyStatement:
		c.emptyStatement(
			parent,
			atomFunc,
			ast,
		)

	case AstTypeExpressionStatement:
		c.expressionStatement(
			parent,
			atomFunc,
			ast,
		)

	case AstTypeFunction:
		c.function(
			parent,
			atomFunc,
			ast,
		)

	case AstTypeIfStatement:
		c.ifStatement(
			parent,
			atomFunc,
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
	funcScope := NewAtomScope(parentScope, AtomScopeTypeFunction)
	atomFunc := runtime.NewFunction(c.parser.tokenizer.file, ast.Str0, len(ast.Arr0))
	params := ast.Arr0
	if parentScope.Type != AtomScopeTypeGlobal {
		Error(
			c.parser.tokenizer.file,
			c.parser.tokenizer.data,
			"Function must be defined in global scope",
			ast.Position,
		)
		return nil
	}
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
	atomFunc.Value.(*runtime.AtomCode).AllocateLocals()

	// save function to function table
	fnOffset := c.state.SaveFunction(atomFunc)
	c.emitInt(parentFunc, runtime.OpLoadFunction, fnOffset)

	if parentScope.HasLocal(ast.Ast0.Str0) {
		Error(
			c.parser.tokenizer.file,
			c.parser.tokenizer.data,
			fmt.Sprintf("Symbol %s already defined", ast.Ast0.Str0),
			ast.Ast0.Position,
		)
	}

	offset := parentFunc.Value.(*runtime.AtomCode).IncrementLocal()
	parentScope.AddSymbol(NewAtomSymbol(
		ast.Ast0.Str0,
		offset,
		parentScope.Type == AtomScopeTypeGlobal,
	))
	c.emitInt(parentFunc, runtime.OpStoreLocal, offset)
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
	return programFunc
}

func (c *AtomCompile) Compile() *runtime.AtomValue {
	return c.program(c.parser.Parse())
}
