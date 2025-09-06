package main

import (
	"encoding/binary"
	"math"
	"strconv"

	runtime "dev.runtime"
)

/*
 * Hide everything.
 */
type Compile struct {
	state  *runtime.AtomState
	parser *Parser
}

func NewCompile(parser *Parser, state *runtime.AtomState) *Compile {
	return &Compile{parser: parser, state: state}
}

func (c *Compile) emit(atomFunc *runtime.AtomValue, opcode runtime.OpCode) {
	atomFunc.Value.(*runtime.Code).OpCodes =
		append(atomFunc.Value.(*runtime.Code).OpCodes, opcode)
}

func (c *Compile) emitInt(atomFunc *runtime.AtomValue, opcode runtime.OpCode, intValue int) {
	// convert int32 to 4 bytes using little-endian encoding
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, uint32(intValue))

	atomFunc.Value.(*runtime.Code).OpCodes =
		append(
			append(atomFunc.Value.(*runtime.Code).OpCodes, opcode),
			runtime.OpCode(bytes[0]),
			runtime.OpCode(bytes[1]),
			runtime.OpCode(bytes[2]),
			runtime.OpCode(bytes[3]),
		)
}

func (c *Compile) emitNum(atomFunc *runtime.AtomValue, opcode runtime.OpCode, numValue float64) {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, uint64(math.Float64bits(numValue)))

	atomFunc.Value.(*runtime.Code).OpCodes =
		append(
			append(atomFunc.Value.(*runtime.Code).OpCodes, opcode),
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

func (c *Compile) emitStr(atomFunc *runtime.AtomValue, opcode runtime.OpCode, strValue string) {
	bytes := make([]byte, len(strValue))
	copy(bytes, []byte(strValue))

	opcodes := make([]runtime.OpCode, len(bytes))
	for i, b := range bytes {
		opcodes[i] = runtime.OpCode(b)
	}

	atomFunc.Value.(*runtime.Code).OpCodes =
		append(
			append(atomFunc.Value.(*runtime.Code).OpCodes, opcode),
			opcodes...,
		)

	atomFunc.Value.(*runtime.Code).OpCodes =
		append(
			append(atomFunc.Value.(*runtime.Code).OpCodes, opcode),
			0, // null byte
		)
}

func (c *Compile) emitJump(atomFunc *runtime.AtomValue, opcode runtime.OpCode) int {
	c.emit(atomFunc, opcode)
	start := len(atomFunc.Value.(*runtime.Code).OpCodes)
	c.emit(atomFunc, opcode)
	return start
}

func (c *Compile) label(atomFunc *runtime.AtomValue, jumpAddress int) {
	current := len(atomFunc.Value.(*runtime.Code).OpCodes)
	for i := range 4 {
		atomFunc.Value.(*runtime.Code).OpCodes[i+jumpAddress] =
			runtime.OpCode((current >> (8 * i)) & 0xFF)
	}
}

func (c *Compile) expression(atomFunc *runtime.AtomValue, ast *Ast) {
	switch ast.AstType {
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

	case AstTypeBinaryMul:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(atomFunc, lhs)
			c.expression(atomFunc, rhs)
			c.emit(atomFunc, runtime.OpMul)
		}

	case AstTypeBinaryDiv:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(atomFunc, lhs)
			c.expression(atomFunc, rhs)
			c.emit(atomFunc, runtime.OpDiv)
		}

	case AstTypeBinaryMod:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(atomFunc, lhs)
			c.expression(atomFunc, rhs)
			c.emit(atomFunc, runtime.OpMod)
		}

	case AstTypeBinaryAdd:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(atomFunc, lhs)
			c.expression(atomFunc, rhs)
			c.emit(atomFunc, runtime.OpAdd)
		}

	case AstTypeBinarySub:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(atomFunc, lhs)
			c.expression(atomFunc, rhs)
			c.emit(atomFunc, runtime.OpSub)
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

func (c *Compile) statement(atomFunc *runtime.AtomValue, ast *Ast) {
	switch ast.AstType {
	case AstTypeFunction:
		c.function(ast)
	default:
		c.expression(atomFunc, ast)
	}
}

func (c *Compile) function(ast *Ast) *runtime.AtomValue {
	atomFunc := runtime.NewAtomValue(runtime.AtomTypeFunc)
	atomFunc.Value = runtime.NewCode(c.parser.tokenizer.file, ast.Str0)

	// compile parameters
	params := ast.Arr0
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

		// increment local count and add to locals
		// offset := atomFunc.Value.(*runtime.Code).IncrementLocal()

		// emit opcode
		// c.emitInt(atomFunc, runtime.OpStoreLocal, offset)
	}

	// compile body
	body := ast.Arr1
	for _, stmt := range body {
		c.statement(atomFunc, stmt)
	}

	// emit return opcode
	c.emit(atomFunc, runtime.OpReturn)

	// allocate locals
	atomFunc.Value.(*runtime.Code).AllocateLocals()

	return atomFunc
}

func (c *Compile) Compile() *runtime.AtomValue {
	return c.function(c.parser.Parse())
}
