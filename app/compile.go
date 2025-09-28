package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	runtime "dev.runtime"
)

type AtomPendingVariable struct {
	ast      *AtomAst
	atomFunc *runtime.AtomValue
	index    int
}

/*
 * Hide everything.
 */
type AtomCompile struct {
	state            *runtime.AtomState
	parser           *AtomParser
	pendingVariables []AtomPendingVariable
}

func NewAtomCompile(parser *AtomParser, state *runtime.AtomState) *AtomCompile {
	return &AtomCompile{
		parser:           parser,
		state:            state,
		pendingVariables: []AtomPendingVariable{},
	}
}

func isConstant(ast *AtomAst) bool {
	switch ast.AstType {
	case AstTypeNum,
		AstTypeStr,
		AstTypeBool,
		AstTypeNull:
		return true
	case AstTypeInt:
		return !strings.HasSuffix(ast.Str0, "n") && !strings.HasSuffix(ast.Str0, "N")
	case AstTypeBinaryMul,
		AstTypeBinaryDiv,
		AstTypeBinaryMod,
		AstTypeBinaryAdd,
		AstTypeBinarySub,
		AstTypeBinaryShiftRight,
		AstTypeBinaryShiftLeft,
		AstTypeBinaryGreaterThan,
		AstTypeBinaryGreaterThanEqual,
		AstTypeBinaryLessThan,
		AstTypeBinaryLessThanEqual,
		AstTypeBinaryEqual,
		AstTypeBinaryNotEqual,
		AstTypeBinaryAnd,
		AstTypeBinaryOr,
		AstTypeBinaryXor:
		// Careful for recursive calls
		return isConstant(ast.Ast0) && isConstant(ast.Ast1)
	default:
		return false
	}
}

func hasDeclairation(block []*AtomAst) bool {
	for _, stmt := range block {
		if stmt.AstType == AstTypeVarStatement || stmt.AstType == AstTypeConstStatement || stmt.AstType == AstTypeLocalStatement {
			return true
		}
	}
	return false
}

func arrayReverse(path []string) []string {
	reverse := []string{}
	for i := len(path) - 1; i >= 0; i-- {
		reverse = append(reverse, path[i])
	}
	return reverse
}

func currentLoop(scope *AtomScope) *AtomScope {
	for current := scope; current != nil; current = current.Parent {
		if current.Type == AtomScopeTypeLoop {
			return current
		}
	}
	return nil
}

func currentBlock(scope *AtomScope) *AtomScope {
	for current := scope; current != nil; current = current.Parent {
		if current.Type == AtomScopeTypeBlock {
			return current
		}
	}
	return nil
}

func sendBreak(scope *AtomScope, jumpAddress int) {
	loop := currentLoop(scope)
	loop.Breaks = append(loop.Breaks, jumpAddress)
}

func sendContinue(scope *AtomScope, jumpAddress int) {
	loop := currentLoop(scope)
	loop.Continues = append(loop.Continues, jumpAddress)
}

func (c *AtomCompile) emitByRuntimeValue(fn, obj *runtime.AtomValue) {
	switch obj.Type {
	case runtime.AtomTypeInt:
		c.emitInt(fn, runtime.OpLoadInt, int(obj.Value.(int32)))
	case runtime.AtomTypeNum:
		c.emitNum(fn, runtime.OpLoadNum, obj.Value.(float64))
	case runtime.AtomTypeStr:
		c.emitStr(fn, runtime.OpLoadStr, obj.Value.(string))
	case runtime.AtomTypeBool:
		if runtime.CoerceToBool(obj) {
			c.emitInt(fn, runtime.OpLoadBool, 1)
		} else {
			c.emitInt(fn, runtime.OpLoadBool, 0)
		}
	case runtime.AtomTypeNull:
		c.emit(fn, runtime.OpLoadNull)
	default:
		panic(fmt.Sprintf("invalid type: %d", obj.Type))
	}
}

func (c *AtomCompile) emit(atomFunc *runtime.AtomValue, opcode runtime.OpCode) {
	atomFunc.Value.(*runtime.AtomCode).Code =
		append(atomFunc.Value.(*runtime.AtomCode).Code, opcode)
}

func (c *AtomCompile) emitInt(atomFunc *runtime.AtomValue, opcode runtime.OpCode, intValue int) {
	// Convert int32 to 4 bytes using little-endian encoding
	bytes := []byte{0, 0, 0, 0}
	binary.LittleEndian.PutUint32(bytes, uint32(intValue))

	atomFunc.Value.(*runtime.AtomCode).Code = append(
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

	atomFunc.Value.(*runtime.AtomCode).Code = append(
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

func (c *AtomCompile) emitLine(atomFunc *runtime.AtomValue, pos AtomPosition) {
	address := len(atomFunc.Value.(*runtime.AtomCode).Code)
	atomFunc.Value.(*runtime.AtomCode).Line = append(atomFunc.Value.(*runtime.AtomCode).Line, runtime.AtomDebugLine{
		Line:    pos.LineStart,
		Address: address,
	})
}

func (c *AtomCompile) emitVar(atomFunc *runtime.AtomValue, scope *AtomScope, ast *AtomAst, global, constant bool) {
	if _, exists := scope.Names[ast.Str0]; exists {
		Error(
			c.parser.tokenizer.file,
			c.parser.tokenizer.data,
			fmt.Sprintf("Variable %s already exists", ast.Str0),
			ast.Position,
		)
	}

	aliases := []string{}
	current := scope
	for current != nil {
		if current.Alias != "" {
			aliases = append(aliases, current.Alias)
		}
		current = current.Parent
	}

	// Namespaced??
	path := append(arrayReverse(aliases), ast.Str0)
	name := ast.Str0
	if len(path) > 1 && global {
		name = strings.Join(path, "::")
		// convert to global
		for scope.Parent != nil {
			scope = scope.Parent
		}
	}

	// Save to symbol table
	scope.Names[name] = NewAtomSymbol(
		name,
		global,
		constant,
	)

	c.emitLine(atomFunc, ast.Position)
	c.emitStr(atomFunc, runtime.OpInitLocal, name)
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

func (c *AtomCompile) lookup(scope *AtomScope, symbol string) *AtomSymbol {
	for current := scope; current != nil; current = current.Parent {
		if _, exists := current.Names[symbol]; exists {
			return current.Names[symbol]
		}
	}
	panic(fmt.Sprintf("symbol '%s' not found in scope", symbol))
}

func (c *AtomCompile) isDefined(scope *AtomScope, symbol string) bool {
	for current := scope; current != nil; current = current.Parent {
		if _, exists := current.Names[symbol]; exists {
			return true
		}
	}
	return false
}

func (c *AtomCompile) identifier(fn *runtime.AtomValue, scope *AtomScope, ast *AtomAst, opcode runtime.OpCode) {
	// Preprocessor like
	if opcode == runtime.OpLoadName && ast.Str0 == "__name__" {
		c.emitStr(fn, runtime.OpLoadStr, "script")
		return
	} else if opcode == runtime.OpLoadName && ast.Str0 == "__file__" {
		c.emitStr(fn, runtime.OpLoadStr, c.parser.tokenizer.file)
		return
	} else if opcode == runtime.OpLoadName && ast.Str0 == "__dir__" {
		absPath, err := filepath.Abs(c.parser.tokenizer.file)
		if err != nil {
			Error(
				c.parser.tokenizer.file,
				c.parser.tokenizer.data,
				"Failed to get absolute path",
				ast.Position,
			)
		}
		c.emitStr(fn, runtime.OpLoadStr, filepath.Dir(absPath))
		return
	} else if opcode == runtime.OpLoadName && ast.Str0 == "__line__" {
		c.emitInt(fn, runtime.OpLoadInt, ast.Position.LineStart)
		return
	}

	if !c.isDefined(scope, ast.Str0) {
		// Resolve to global
		c.emitStr(fn, runtime.OpLoadName, ast.Str0)
		c.pendingVariables = append(c.pendingVariables, AtomPendingVariable{
			ast:      ast,
			atomFunc: fn,
			index:    0,
		})
		return
	}
	symbol := c.lookup(scope, ast.Str0)
	if opcode == runtime.OpStoreLocal && symbol.constant {
		Error(
			c.parser.tokenizer.file,
			c.parser.tokenizer.data,
			"Cannot store to constant variable",
			ast.Position,
		)
	}
	c.emitStr(fn, opcode, ast.Str0)
}

func (c *AtomCompile) expression(scope *AtomScope, fn *runtime.AtomValue, ast *AtomAst) {
	switch ast.AstType {
	case AstTypeIdn:
		{
			c.emitLine(fn, ast.Position)
			c.identifier(fn, scope, ast, runtime.OpLoadName)
		}

	case AstTypeInt:
		c.emitLine(fn, ast.Position)

		if strings.HasSuffix(ast.Str0, "n") || strings.HasSuffix(ast.Str0, "N") {
			ast.Str0 = strings.TrimSuffix(ast.Str0, "n")
			ast.Str0 = strings.TrimSuffix(ast.Str0, "N")
			c.emitStr(fn, runtime.OpLoadBigInt, ast.Str0)
			return
		}

		intValue, err := strconv.Atoi(ast.Str0)
		var overflowed bool
		if after, ok := strings.CutPrefix(ast.Str0, "0x"); ok {
			_intValue, _err := strconv.ParseInt(after, 16, 64)
			if _intValue > math.MaxInt32 || _intValue < math.MinInt32 {
				overflowed = true
			} else {
				intValue = int(_intValue)
			}
			err = _err
		} else if after, ok := strings.CutPrefix(ast.Str0, "0o"); ok {
			_intValue, _err := strconv.ParseInt(after, 8, 64)
			if _intValue > math.MaxInt32 || _intValue < math.MinInt32 {
				overflowed = true
			} else {
				intValue = int(_intValue)
			}
			err = _err
		} else if after, ok := strings.CutPrefix(ast.Str0, "0b"); ok {
			_intValue, _err := strconv.ParseInt(after, 2, 64)
			if _intValue > math.MaxInt32 || _intValue < math.MinInt32 {
				overflowed = true
			} else {
				intValue = int(_intValue)
			}
			err = _err
		} else {
			// Check for overflow in decimal case
			_intValue, _err := strconv.ParseInt(ast.Str0, 10, 64)
			if _err == nil && (_intValue > math.MaxInt32 || _intValue < math.MinInt32) {
				overflowed = true
			}
		}

		// If overflow detected, promote to float
		if overflowed {
			numValue, numErr := strconv.ParseFloat(ast.Str0, 64)
			if numErr != nil {
				Error(
					c.parser.tokenizer.file,
					c.parser.tokenizer.data,
					"Invalid number",
					ast.Position,
				)
			}
			c.emitNum(fn, runtime.OpLoadNum, numValue)
			return
		}

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
		c.emitLine(fn, ast.Position)
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
		c.emitLine(fn, ast.Position)
		c.emitStr(fn, runtime.OpLoadStr, ast.Str0)

	case AstTypeBool:
		c.emitLine(fn, ast.Position)
		var boolValue byte
		if ast.Str0 == "true" {
			boolValue = 1
		} else {
			boolValue = 0
		}
		c.emitInt(fn, runtime.OpLoadBool, int(boolValue))

	case AstTypeNull:
		c.emitLine(fn, ast.Position)
		c.emit(fn, runtime.OpLoadNull)

	case AstTypeBase:
		// Guard
		if !scope.InSide(AtomScopeTypeClass, true) && !scope.InSide(AtomScopeTypeFunction, true) && !scope.InSide(AtomScopeTypeAsyncFunction, true) {
			Error(
				c.parser.tokenizer.file,
				c.parser.tokenizer.data,
				"Base must be in class, function, or async function scope",
				ast.Position,
			)
		}

		c.emitLine(fn, ast.Position)
		c.emit(fn, runtime.OpLoadBase)

	case AstTypeArray:
		{
			for i := 0; i < len(ast.Arr0); i++ {
				element := ast.Arr0[i]
				c.expression(scope, fn, element)
			}
			c.emitLine(fn, ast.Position)
			c.emitInt(fn, runtime.OpLoadArray, len(ast.Arr0))
		}

	case AstTypeObject:
		{
			for i := len(ast.Arr0) - 1; i >= 0; i-- {
				element := ast.Arr0[i]
				k := element.Ast0
				v := element.Ast1

				if k.AstType != AstTypeIdn && k.AstType != AstTypeStr {
					Error(
						c.parser.tokenizer.file,
						c.parser.tokenizer.data,
						"Expected identifier or string",
						k.Position,
					)
					return
				}

				c.expression(scope, fn, v)
				c.emitLine(fn, ast.Position)
				c.emitStr(fn, runtime.OpLoadStr, k.Str0)
			}
			c.emitLine(fn, ast.Position)
			c.emitInt(fn, runtime.OpLoadObject, len(ast.Arr0))
		}

	case AstTypeAsyncFunctionExpression,
		AstTypeFunctionExpression:
		{
			async := ast.AstType == AstTypeAsyncFunctionExpression
			scopeType := AtomScopeTypeFunction
			if async {
				scopeType = AtomScopeTypeAsyncFunction
			}

			funScope := NewAtomScope(scope, scopeType)
			atomFunc := runtime.NewAtomGenericValue(
				runtime.AtomTypeFunc,
				runtime.NewAtomCode(c.parser.tokenizer.file, "anonymous", async, len(ast.Arr0)),
			)

			params := ast.Arr0
			//============================
			fnOffset := c.state.SaveFunction(atomFunc)

			// Save to symbol table first to allow captures to reference it
			c.emitLine(fn, ast.Position)
			c.emitInt(fn, runtime.OpLoadFunction, fnOffset)
			//============================

			for i := len(params) - 1; i >= 0; i-- {
				param := params[i]
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
				c.emitVar(atomFunc, funScope, param, false, false)
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
				c.emitLine(atomFunc, ast.Position)
				c.emit(atomFunc, runtime.OpLoadNull)
				c.emitLine(atomFunc, ast.Position)
				c.emit(atomFunc, runtime.OpReturn)
			}
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
			c.emitLine(fn, ast.Position)
			c.emitStr(fn, runtime.OpLoadStr, key.Str0)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpIndex)
		}

	case AstTypeIndex:
		{
			obj := ast.Ast0
			key := ast.Ast1
			c.expression(scope, fn, obj)
			c.expression(scope, fn, key)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpIndex)
		}

	case AstTypeCall:
		{
			funcAst := ast.Ast0
			args := ast.Arr0
			for i := 0; i < len(args); i++ {
				c.expression(scope, fn, args[i])
			}
			c.expression(scope, fn, funcAst)
			c.emitLine(fn, ast.Position)
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

			for i := range len(args) {
				c.expression(scope, fn, args[i])
			}

			c.expression(scope, fn, constructorAst)
			c.emitLine(fn, ast.Position)
			c.emitInt(fn, runtime.OpCallConstructor, len(args))
		}

	case AstTypePostfixInc:
		{
			c.assignOp0(scope, fn, ast.Ast0, true)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpInc)
			c.assignOp1(scope, fn, ast.Ast0, true)
		}

	case AstTypePostfixDec:
		{
			c.assignOp0(scope, fn, ast.Ast0, true)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpDec)
			c.assignOp1(scope, fn, ast.Ast0, true)
		}

	case AstTypeUnaryNot:
		{
			c.expression(scope, fn, ast.Ast0)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpNot)
		}

	case AstTypeUnaryNeg:
		{
			c.expression(scope, fn, ast.Ast0)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpNeg)
		}

	case AstTypeUnaryPos:
		{
			c.expression(scope, fn, ast.Ast0)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpPos)
		}

	case AstTypeUnaryInc:
		{
			c.assignOp0(scope, fn, ast.Ast0, false)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpInc)
			c.assignOp1(scope, fn, ast.Ast0, false)
		}

	case AstTypeUnaryDec:
		{
			c.assignOp0(scope, fn, ast.Ast0, false)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpDec)
			c.assignOp1(scope, fn, ast.Ast0, false)
		}

	case AstTypeUnaryTypeof:
		{
			c.expression(scope, fn, ast.Ast0)
			c.emitLine(fn, ast.Position)
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
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpAwait)
		}

	case AstTypeBinaryMul:
		{
			if isConstant(ast) {
				result := Eval(c, ast)
				c.emitLine(fn, ast.Position)
				c.emitByRuntimeValue(fn, result)
				return
			}
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpMul)
		}

	case AstTypeBinaryDiv:
		{
			if isConstant(ast) {
				result := Eval(c, ast)
				c.emitLine(fn, ast.Position)
				c.emitByRuntimeValue(fn, result)
				return
			}
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpDiv)
		}

	case AstTypeBinaryMod:
		{
			if isConstant(ast) {
				result := Eval(c, ast)
				c.emitLine(fn, ast.Position)
				c.emitByRuntimeValue(fn, result)
				return
			}
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpMod)
		}

	case AstTypeBinaryAdd:
		{
			if isConstant(ast) {
				result := Eval(c, ast)
				c.emitLine(fn, ast.Position)
				c.emitByRuntimeValue(fn, result)
				return
			}
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpAdd)
		}

	case AstTypeBinarySub:
		{
			if isConstant(ast) {
				result := Eval(c, ast)
				c.emitLine(fn, ast.Position)
				c.emitByRuntimeValue(fn, result)
				return
			}
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpSub)
		}

	case AstTypeBinaryShiftRight:
		{
			if isConstant(ast) {
				result := Eval(c, ast)
				c.emitLine(fn, ast.Position)
				c.emitByRuntimeValue(fn, result)
				return
			}
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpShr)
		}

	case AstTypeBinaryShiftLeft:
		{
			if isConstant(ast) {
				result := Eval(c, ast)
				c.emitLine(fn, ast.Position)
				c.emitByRuntimeValue(fn, result)
				return
			}
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpShl)
		}

	case AstTypeBinaryGreaterThan:
		{
			if isConstant(ast) {
				result := Eval(c, ast)
				c.emitLine(fn, ast.Position)
				c.emitByRuntimeValue(fn, result)
				return
			}
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpCmpGt)
		}

	case AstTypeBinaryGreaterThanEqual:
		{
			if isConstant(ast) {
				result := Eval(c, ast)
				c.emitLine(fn, ast.Position)
				c.emitByRuntimeValue(fn, result)
				return
			}
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpCmpGte)
		}

	case AstTypeBinaryLessThan:
		{
			if isConstant(ast) {
				result := Eval(c, ast)
				c.emitLine(fn, ast.Position)
				c.emitByRuntimeValue(fn, result)
				return
			}
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpCmpLt)
		}

	case AstTypeBinaryLessThanEqual:
		{
			if isConstant(ast) {
				result := Eval(c, ast)
				c.emitLine(fn, ast.Position)
				c.emitByRuntimeValue(fn, result)
				return
			}
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpCmpLte)
		}

	case AstTypeBinaryEqual:
		{
			if isConstant(ast) {
				result := Eval(c, ast)
				c.emitLine(fn, ast.Position)
				c.emitByRuntimeValue(fn, result)
				return
			}
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpCmpEq)
		}

	case AstTypeBinaryNotEqual:
		{
			if isConstant(ast) {
				result := Eval(c, ast)
				c.emitLine(fn, ast.Position)
				c.emitByRuntimeValue(fn, result)
				return
			}
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpCmpNe)
		}

	case AstTypeBinaryAnd:
		{
			if isConstant(ast) {
				result := Eval(c, ast)
				c.emitLine(fn, ast.Position)
				c.emitByRuntimeValue(fn, result)
				return
			}
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpAnd)
		}

	case AstTypeBinaryOr:
		{
			if isConstant(ast) {
				result := Eval(c, ast)
				c.emitLine(fn, ast.Position)
				c.emitByRuntimeValue(fn, result)
				return
			}
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpOr)
		}

	case AstTypeBinaryXor:
		{
			if isConstant(ast) {
				result := Eval(c, ast)
				c.emitLine(fn, ast.Position)
				c.emitByRuntimeValue(fn, result)
				return
			}
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpXor)
		}

	case AstTypeLogicalAnd:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.emitLine(fn, ast.Position)
			toEnd0 := c.emitJump(fn, runtime.OpJumpIfFalseOrPop)
			c.expression(scope, fn, rhs)
			c.label(fn, toEnd0)
		}

	case AstTypeLogicalOr:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.emitLine(fn, ast.Position)
			toEnd0 := c.emitJump(fn, runtime.OpJumpIfTrueOrPop)
			c.expression(scope, fn, rhs)
			c.label(fn, toEnd0)
		}

	case AstTypeAssign:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, rhs)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpDupTop)
			c.assign(scope, fn, lhs)
		}

	case AstTypeMulAssign:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpMul)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpDupTop)
			c.assign(scope, fn, lhs)
		}

	case AstTypeDivAssign:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpDiv)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpDupTop)
			c.assign(scope, fn, lhs)
		}

	case AstTypeModAssign:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpMod)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpDupTop)
			c.assign(scope, fn, lhs)
		}

	case AstTypeAddAssign:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpAdd)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpDupTop)
			c.assign(scope, fn, lhs)
		}

	case AstTypeSubAssign:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpSub)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpDupTop)
			c.assign(scope, fn, lhs)
		}

	case AstTypeLeftShiftAssign:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpShl)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpDupTop)
			c.assign(scope, fn, lhs)
		}

	case AstTypeRightShiftAssign:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpShr)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpDupTop)
			c.assign(scope, fn, lhs)
		}

	case AstTypeBitwiseAndAssign:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpAnd)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpDupTop)
			c.assign(scope, fn, lhs)
		}

	case AstTypeBitwiseOrAssign:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpOr)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpDupTop)
			c.assign(scope, fn, lhs)
		}

	case AstTypeBitwiseXorAssign:
		{
			lhs := ast.Ast0
			rhs := ast.Ast1
			c.expression(scope, fn, lhs)
			c.expression(scope, fn, rhs)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpXor)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpDupTop)
			c.assign(scope, fn, lhs)
		}

	case AstTypeIfExpression:
		{
			condition := ast.Ast0
			thenValue := ast.Ast1
			elseValue := ast.Ast2

			c.expression(scope, fn, condition)
			c.emitLine(fn, ast.Position)
			toElse := c.emitJump(fn, runtime.OpPopJumpIfFalse)
			c.expression(scope, fn, thenValue)
			c.emitLine(fn, ast.Position)
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
					c.emitLine(fn, ast.Position)
					jumpToValue := c.emitJump(fn, runtime.OpPeekJumpIfEqual)
					storedJumps = append(storedJumps, jumpToValue)
				}
				c.emitLine(fn, ast.Position)
				toNextCase := c.emitJump(fn, runtime.OpJump)

				// value
				for _, jump := range storedJumps {
					c.label(fn, jump)
				}
				// Pop condition if match
				c.emitLine(fn, ast.Position)
				c.emit(fn, runtime.OpPopTop)

				// value
				c.expression(scope, fn, value)
				c.emitLine(fn, ast.Position)
				jumpToEnd := c.emitJump(fn, runtime.OpJump)
				toEndSwitch = append(toEndSwitch, jumpToEnd)

				// Next?
				c.label(fn, toNextCase)
			}

			// Pop condition if default
			c.emitLine(fn, ast.Position)
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
			atomFunc := runtime.NewAtomGenericValue(
				runtime.AtomTypeFunc,
				runtime.NewAtomCode(c.parser.tokenizer.file, "catch", false, 1),
			)
			funScope := NewAtomScope(scope, AtomScopeTypeFunction)
			fnOffset := c.state.SaveFunction(atomFunc)

			c.expression(scope, fn, condition)
			c.emitLine(fn, ast.Position)
			toEndCatch := c.emitJump(fn, runtime.OpPopJumpIfNotError)

			// Variable as parameter
			c.emitVar(atomFunc, funScope, variable, false, false)

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
				c.emitLine(atomFunc, ast.Position)
				c.emit(atomFunc, runtime.OpLoadNull)
				c.emitLine(atomFunc, ast.Position)
				c.emit(atomFunc, runtime.OpReturn)
			}

			// Load and call
			c.emitLine(fn, ast.Position)
			c.emitInt(fn, runtime.OpLoadFunction, fnOffset)
			c.emitLine(fn, ast.Position)
			c.emitInt(fn, runtime.OpCall, 1)

			// End Catch
			c.label(fn, toEndCatch)
		}

	default:
		Error(
			c.parser.tokenizer.file,
			c.parser.tokenizer.data,
			fmt.Sprintf("Expected expression, got %d", ast.AstType),
			ast.Position,
		)
	}
}

func (c *AtomCompile) assignOp0(scope *AtomScope, fn *runtime.AtomValue, ast *AtomAst, postfix bool) {
	switch ast.AstType {
	case AstTypeIdn:
		{
			c.emitLine(fn, ast.Position)
			c.identifier(fn, scope, ast, runtime.OpLoadName)

			if postfix {
				c.emit(fn, runtime.OpDupTop)
			}
		}
	case AstTypeMember:
		{
			obj := ast.Ast0
			key := ast.Ast1
			c.expression(scope, fn, obj)
			c.emitLine(fn, ast.Position)
			c.emitStr(fn, runtime.OpLoadStr, key.Str0)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpDupTop2)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpIndex)

			if postfix {
				c.emit(fn, runtime.OpDupTop)
			}
		}
	case AstTypeIndex:
		{
			obj := ast.Ast0
			key := ast.Ast1
			c.expression(scope, fn, obj)
			c.expression(scope, fn, key)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpDupTop2)
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpIndex)

			if postfix {
				c.emit(fn, runtime.OpDupTop)
			}
		}
	default:
		Error(
			c.parser.tokenizer.file,
			c.parser.tokenizer.data,
			"Expected identifier or member expression",
			ast.Position,
		)
	}
}

func (c *AtomCompile) assignOp1(scope *AtomScope, fn *runtime.AtomValue, ast *AtomAst, postfix bool) {
	switch ast.AstType {
	case AstTypeIdn:
		{
			c.emit(fn, runtime.OpDupTop)

			if postfix {
				c.emit(fn, runtime.OpRot2)
				c.emit(fn, runtime.OpPopTop)
			}

			c.emitLine(fn, ast.Position)
			c.identifier(fn, scope, ast, runtime.OpStoreLocal)
		}
	case AstTypeMember,
		AstTypeIndex:
		{
			/*
				unary order
				var A = { B: 2 };
				++A.B;
				==>
				ROT4    :    => [B, A, 3, 3]
				DUP_TOP :    => [A, B, 3, 3]
				INC     :	 => [A, B, 3]
				INDEX   :	 => [A, B, 2]
				DUP_TOP2 :   => [A, B, A, B]
				LOAD_STR: B  => [A, A, B]
				DUP_TOP : 	 => [A, A]
				LOAD_OBJ: A  => [A, ]
			*/

			c.emit(fn, runtime.OpDupTop)

			if postfix {
				c.emit(fn, runtime.OpRot2)
				c.emit(fn, runtime.OpPopTop)
			}

			// [A, B, 3, 3] -> [3, A, B, 3]
			c.emit(fn, runtime.OpRot4)
			// [3, A, B, 3] -> [3, 3, A, B]
			c.emit(fn, runtime.OpRot4)

			// SET_INDEX
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpSetIndex)
		}
	default:
		Error(
			c.parser.tokenizer.file,
			c.parser.tokenizer.data,
			"Expected identifier or member expression",
			ast.Position,
		)
	}
}

func (c *AtomCompile) assign(scope *AtomScope, fn *runtime.AtomValue, lhs *AtomAst) {
	switch lhs.AstType {
	case AstTypeIdn:
		{
			c.emitLine(fn, lhs.Position)
			c.identifier(fn, scope, lhs, runtime.OpStoreLocal)
		}

	case AstTypeMember:
		{
			c.expression(scope, fn, lhs.Ast0)
			c.emitLine(fn, lhs.Position)
			c.emitStr(fn, runtime.OpLoadStr, lhs.Ast1.Str0)
			c.emitLine(fn, lhs.Position)
			c.emit(fn, runtime.OpSetIndex)
		}

	case AstTypeIndex:
		{
			c.expression(scope, fn, lhs.Ast0)
			c.expression(scope, fn, lhs.Ast1)
			c.emitLine(fn, lhs.Position)
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

	case AstTypeForStatement:
		c.forStatement(
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

	depth := 0
	for block := currentBlock(scope); block != nil && block.Type == AtomScopeTypeBlock; block = block.Parent {
		depth++
	}
	if depth != 0 {
		c.emitLine(fn, ast.Position)
		c.emitInt(fn, runtime.OpExitBlock, depth)
	}

	c.emitLine(fn, ast.Position)
	sendBreak(scope, c.emitJump(fn, runtime.OpJump))
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

	depth := 0
	for block := currentBlock(scope); block != nil && block.Type == AtomScopeTypeBlock; block = block.Parent {
		depth++
	}
	if depth != 0 {
		c.emitLine(fn, ast.Position)
		c.emitInt(fn, runtime.OpExitBlock, depth)
	}

	c.emitLine(fn, ast.Position)
	sendContinue(scope, c.emitJump(fn, runtime.OpAbsoluteJump))
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
		c.emitLine(fn, ast.Position)
		c.emit(fn, runtime.OpLoadNull)
	}

	depth := 0
	for block := currentBlock(scope); block != nil && block.Type == AtomScopeTypeBlock; block = block.Parent {
		depth++
	}
	if depth != 0 {
		c.emitLine(fn, ast.Position)
		c.emitInt(fn, runtime.OpExitBlock, depth)
	}

	c.emitLine(fn, ast.Position)
	c.emit(fn, runtime.OpReturn)
}

func (c *AtomCompile) emptyStatement(_ *AtomScope, fn *runtime.AtomValue, ast *AtomAst) {
	c.emitLine(fn, ast.Position)
	c.emit(fn, runtime.OpNoOp)
}

func (c *AtomCompile) expressionStatement(scope *AtomScope, fn *runtime.AtomValue, ast *AtomAst) {
	c.expression(scope, fn, ast.Ast0)
	c.emitLine(fn, ast.Position)
	c.emit(fn, runtime.OpPopTop)
}

func (c *AtomCompile) classStatement(scope *AtomScope, fn *runtime.AtomValue, ast *AtomAst) {
	// Guard
	// Allowed only in global or namespace scope
	if !scope.InSide(AtomScopeTypeGlobal, false) && !scope.InSide(AtomScopeTypeNamespace, false) {
		Error(
			c.parser.tokenizer.file,
			c.parser.tokenizer.data,
			"Class statement must be in global or namespace scope",
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

	// initilize class as null
	c.emitLine(fn, ast.Position)
	c.emit(fn, runtime.OpLoadNull)

	// Save
	c.emitVar(fn, scope, name, true, false)

	items := 0

	// Body
	for _, stmt := range body {
		switch stmt.AstType {
		case AstTypeLocalStatement:
			items += len(stmt.Arr0)
			c.classVariable(classScope, fn, stmt)
		case AstTypeFunction,
			AstTypeAsyncFunction:
			items += 1
			c.classFunction(classScope, fn, stmt, stmt.AstType == AstTypeAsyncFunction)
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
	c.emitLine(fn, ast.Position)
	c.emitInt(fn, runtime.OpMakeClass, items)
	c.emitWord(fn, name.Str0)

	if base != nil {
		c.expression(scope, fn, base)
		c.emitLine(fn, ast.Position)
		c.emit(fn, runtime.OpExtendClass)
	}

	// Save
	c.identifier(fn, scope, name, runtime.OpStoreLocal)
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
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpLoadNull)
		}
		c.emitLine(fn, ast.Position)
		c.emitStr(fn, runtime.OpLoadStr, key.Str0)
	}
}

func (c *AtomCompile) classFunction(scope *AtomScope, fn *runtime.AtomValue, ast *AtomAst, async bool) {
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

	scopeType := AtomScopeTypeFunction
	if async {
		scopeType = AtomScopeTypeAsyncFunction
	}

	funScope := NewAtomScope(scope, scopeType)
	atomFunc := runtime.NewAtomGenericValue(
		runtime.AtomTypeFunc,
		runtime.NewAtomCode(c.parser.tokenizer.file, ast.Ast0.Str0, async, len(ast.Arr0)),
	)

	params := ast.Arr0
	//============================
	fnOffset := c.state.SaveFunction(atomFunc)
	c.emitLine(fn, ast.Position)
	c.emitInt(fn, runtime.OpLoadFunction, fnOffset)
	c.emitLine(fn, ast.Position)
	c.emitStr(fn, runtime.OpLoadStr, ast.Ast0.Str0)
	//============================

	for i := len(params) - 1; i >= 0; i-- {
		param := params[i]
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
		c.emitVar(atomFunc, funScope, param, false, false)
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
		c.emitLine(atomFunc, ast.Position)
		c.emit(atomFunc, runtime.OpLoadNull)
		c.emitLine(atomFunc, ast.Position)
		c.emit(atomFunc, runtime.OpReturn)
	}
}

func (c *AtomCompile) enumStatement(scope *AtomScope, fn *runtime.AtomValue, ast *AtomAst) {
	// Guard
	// Allowed only in global
	if !scope.InSide(AtomScopeTypeGlobal, false) && !scope.InSide(AtomScopeTypeNamespace, false) {
		Error(
			c.parser.tokenizer.file,
			c.parser.tokenizer.data,
			"Enum statement must be in global or namespace scope",
			ast.Position,
		)
		return
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
			c.emitLine(fn, ast.Position)
			c.emitInt(fn, runtime.OpLoadInt, index)
		} else {
			c.expression(scope, fn, value)
		}
		c.emitLine(fn, ast.Position)
		c.emitStr(fn, runtime.OpLoadStr, name.Str0)
	}

	c.emitLine(fn, ast.Position)
	c.emitInt(fn, runtime.OpMakeEnum, len(names))
	c.emitVar(fn, scope, name, true, false)
}

func (c *AtomCompile) function(scope *AtomScope, fn *runtime.AtomValue, ast *AtomAst, async bool) {
	// Guard
	// Allowed only in global or namespace scope
	if !scope.InSide(AtomScopeTypeGlobal, false) && !scope.InSide(AtomScopeTypeNamespace, false) {
		Error(
			c.parser.tokenizer.file,
			c.parser.tokenizer.data,
			"Function must be defined in global or namespace scope",
			ast.Position,
		)
		return
	}

	scopeType := AtomScopeTypeFunction
	if async {
		scopeType = AtomScopeTypeAsyncFunction
	}

	funScope := NewAtomScope(scope, scopeType)
	atomFunc := runtime.NewAtomGenericValue(
		runtime.AtomTypeFunc,
		runtime.NewAtomCode(c.parser.tokenizer.file, ast.Ast0.Str0, async, len(ast.Arr0)),
	)

	params := ast.Arr0
	//============================
	fnOffset := c.state.SaveFunction(atomFunc)

	// Save to symbol table first to allow captures to reference it
	c.emitLine(fn, ast.Position)
	c.emitInt(fn, runtime.OpLoadFunction, fnOffset)
	c.emitVar(fn, scope, ast.Ast0, true, false)
	//============================

	for i := len(params) - 1; i >= 0; i-- {
		param := params[i]
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
		c.emitVar(atomFunc, funScope, param, false, false)
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
		c.emitLine(atomFunc, ast.Position)
		c.emit(atomFunc, runtime.OpLoadNull)
		c.emitLine(atomFunc, ast.Position)
		c.emit(atomFunc, runtime.OpReturn)
	}
}

func (c *AtomCompile) block(scope *AtomScope, fn *runtime.AtomValue, ast *AtomAst) {
	var blockScope *AtomScope
	var dec = hasDeclairation(ast.Arr0)
	if dec {
		blockScope = NewAtomScope(scope, AtomScopeTypeBlock)
		c.emitInt(fn, runtime.OpEnterBlock, 1)
	} else {
		blockScope = NewAtomScope(scope, AtomScopeTypeBlockNoEnv)
	}

	for _, stmt := range ast.Arr0 {
		c.statement(blockScope, fn, stmt)
	}

	if dec {
		c.emitInt(fn, runtime.OpExitBlock, 1)
	}
}

func (c *AtomCompile) varStatement(scope *AtomScope, fn *runtime.AtomValue, ast *AtomAst) {
	// Guard
	// Allowed only in global or namespace scope
	if !scope.InSide(AtomScopeTypeGlobal, false) && !scope.InSide(AtomScopeTypeNamespace, false) {
		Error(
			c.parser.tokenizer.file,
			c.parser.tokenizer.data,
			"Var statement must be in global or namespace scope",
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
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpLoadNull)
		} else {
			c.expression(scope, fn, val)
		}

		c.emitVar(
			fn,
			scope,
			key,
			true,
			false,
		)
	}
}

func (c *AtomCompile) constStatement(scope *AtomScope, fn *runtime.AtomValue, ast *AtomAst) {
	// Guard
	// Allowed on any scope except single scope
	if scope.InSide(AtomScopeTypeSingle, false) {
		// error, not allowed in single scope
		// if (true)
		// 	const x = 2;
		Error(
			c.parser.tokenizer.file,
			c.parser.tokenizer.data,
			"Const statement must be in block scope",
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
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpLoadNull)
		} else {
			c.expression(scope, fn, val)
		}

		c.emitVar(
			fn,
			scope,
			key,
			scope.InSide(AtomScopeTypeGlobal, false) || scope.InSide(AtomScopeTypeNamespace, false),
			true,
		)
	}
}

func (c *AtomCompile) localStatement(scope *AtomScope, fn *runtime.AtomValue, ast *AtomAst) {
	// Guard
	// Allowed only in block, function, async function, and loop scope
	if !scope.InSide(AtomScopeTypeBlock, false) &&
		!scope.InSide(AtomScopeTypeBlockNoEnv, false) &&
		!scope.InSide(AtomScopeTypeFunction, false) &&
		!scope.InSide(AtomScopeTypeAsyncFunction, false) &&
		!scope.InSide(AtomScopeTypeLoop, false) {
		// error, not allowed in single scope
		// if (true)
		// 	local x = 2;
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
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpLoadNull)
		} else {
			c.expression(scope, fn, val)
		}

		c.emitVar(
			fn,
			scope,
			key,
			false,
			false,
		)
	}
}

func (c *AtomCompile) importStatement(scope *AtomScope, fn *runtime.AtomValue, ast *AtomAst) {
	// Guard
	// Allowed only in global scope
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

	isRelative := func(name string) bool {
		return strings.HasPrefix(name, "./") || strings.HasPrefix(name, "../")
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
		} else if isRelative(name) {
			// Remove relative path prefixes
			for strings.HasPrefix(name, "./") || strings.HasPrefix(name, "../") {
				if strings.HasPrefix(name, "../") {
					name = strings.TrimPrefix(name, "../")
				} else {
					name = strings.TrimPrefix(name, "./")
				}
			}
			segments := strings.Split(name, "/")
			name = segments[len(segments)-1]
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
		c.emitLine(fn, ast.Position)
		c.emitStr(fn, runtime.OpLoadModule, normalizedPath)

	} else if !isRelative(path.Str0) {
		// Absolute path
		absPath := filepath.Join(c.state.Path, "lib", normalizedPath)
		// Check if is dir
		if stat, error := os.Stat(absPath); error == nil && stat.IsDir() {
			absPath = filepath.Join(absPath, "index.atom")
			// Check if exists
			if _, error := os.Stat(absPath); os.IsNotExist(error) {
				Error(
					c.parser.tokenizer.file,
					c.parser.tokenizer.data,
					fmt.Sprintf("Module %s not found", absPath),
					ast.Position,
				)
			}
		} else {
			// Check if exists
			absPath += ".atom"
		}

		// Check if exists
		if _, error := os.Stat(absPath); os.IsNotExist(error) {
			Error(
				c.parser.tokenizer.file,
				c.parser.tokenizer.data,
				fmt.Sprintf("Module %s not found", absPath),
				ast.Position,
			)
		}

		// Readable?
		if _, error := os.Stat(absPath); error != nil {
			Error(
				c.parser.tokenizer.file,
				c.parser.tokenizer.data,
				fmt.Sprintf("Module %s is not readable", absPath),
				ast.Position,
			)
		}

		if exists := c.state.SaveModule(normalizedPath); !exists {
			// Not exists, compile and export
			t := NewAtomTokenizer(absPath, readFile(absPath))
			p := NewAtomParser(t)
			c := NewAtomCompile(p, c.state)
			i := c.Export()

			c.emitLine(fn, ast.Position)
			c.emitInt(fn, runtime.OpLoadFunction, i)
			c.emitLine(fn, ast.Position)
			c.emitInt(fn, runtime.OpCall, 0)

			// Save to table
			c.emitLine(fn, ast.Position)
			c.emitStr(fn, runtime.OpStoreModule, normalizedPath)
		}

		c.emitLine(fn, ast.Position)
		c.emitStr(fn, runtime.OpLoadModule, normalizedPath)

	} else {
		// Relative path or path with step?

		currentPath := filepath.Dir(c.parser.tokenizer.file)
		absPath := ""

		if strings.HasPrefix(path.Str0, "./") {
			absPath = filepath.Join(currentPath, path.Str0)
			newPath, err := filepath.Abs(absPath)
			if err != nil {
				Error(
					c.parser.tokenizer.file,
					c.parser.tokenizer.data,
					"Failed to get absolute path",
					ast.Position,
				)
			}
			absPath = newPath
		} else if strings.HasPrefix(path.Str0, "../") {
			// subtract currentPath for 1 dir
			absPath = filepath.Join(currentPath, path.Str0)
			newPath, err := filepath.Abs(absPath)
			if err != nil {
				Error(
					c.parser.tokenizer.file,
					c.parser.tokenizer.data,
					"Failed to get absolute path",
					ast.Position,
				)
			}
			absPath = newPath
		} else {
			absPath = filepath.Join(currentPath, path.Str0)
			newPath, err := filepath.Abs(absPath)
			if err != nil {
				Error(
					c.parser.tokenizer.file,
					c.parser.tokenizer.data,
					"Failed to get absolute path",
					ast.Position,
				)
			}
			absPath = newPath
		}

		// Check if exists
		if _, error := os.Stat(absPath); os.IsNotExist(error) {
			Error(
				c.parser.tokenizer.file,
				c.parser.tokenizer.data,
				fmt.Sprintf("Module %s not found", absPath),
				ast.Position,
			)
		}

		// Readable?
		if _, error := os.Stat(absPath); error != nil {
			Error(
				c.parser.tokenizer.file,
				c.parser.tokenizer.data,
				fmt.Sprintf("Module %s is not readable", absPath),
				ast.Position,
			)
		}

		if exists := c.state.SaveModule(normalizedPath); !exists {
			// Not exists, compile and export
			t := NewAtomTokenizer(absPath, readFile(absPath))
			p := NewAtomParser(t)
			c := NewAtomCompile(p, c.state)
			i := c.Export()

			c.emitLine(fn, ast.Position)
			c.emitInt(fn, runtime.OpLoadFunction, i)
			c.emitLine(fn, ast.Position)
			c.emitInt(fn, runtime.OpCall, 0)

			// Save to table
			c.emitLine(fn, ast.Position)
			c.emitStr(fn, runtime.OpStoreModule, normalizedPath)
		}

		c.emitLine(fn, ast.Position)
		c.emitStr(fn, runtime.OpLoadModule, normalizedPath)
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
		c.emitLine(fn, ast.Position)
		c.emitStr(fn, runtime.OpPluckAttribute, name.Str0)

		c.emitVar(
			fn,
			scope,
			name,
			true,
			false,
		)
	}

	// Save to table
	c.emitVar(
		fn,
		scope,
		NewTerminal(
			AstTypeIdn,
			normalizedPath,
			path.Position,
		),
		true,
		false,
	)
}

func (c *AtomCompile) ifStatement(scope *AtomScope, fn *runtime.AtomValue, ast *AtomAst) {
	c.expression(scope, fn, ast.Ast0)
	c.emitLine(fn, ast.Position)
	toElse := c.emitJump(fn, runtime.OpPopJumpIfFalse)
	single := NewAtomScope(scope, AtomScopeTypeSingle)
	c.statement(single, fn, ast.Ast1)
	c.emitLine(fn, ast.Position)
	toEnd := -1
	if ast.Ast2 != nil {
		toEnd = c.emitJump(fn, runtime.OpJump)
		c.label(fn, toElse)
		single := NewAtomScope(scope, AtomScopeTypeSingle)
		c.statement(single, fn, ast.Ast2)
	} else {
		c.label(fn, toElse)
	}
	if toEnd != -1 {
		c.label(fn, toEnd)
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
				c.emitLine(fn, ast.Position)
				jumpToValue := c.emitJump(fn, runtime.OpPeekJumpIfEqual)
				storedJumps = append(storedJumps, jumpToValue)
			}
			c.emitLine(fn, ast.Position)
			toNextCase := c.emitJump(fn, runtime.OpJump)

			// value
			for _, jump := range storedJumps {
				c.label(fn, jump)
			}
			// Pop condition if match
			c.emitLine(fn, ast.Position)
			c.emit(fn, runtime.OpPopTop)

			// statement
			single := NewAtomScope(scope, AtomScopeTypeSingle)
			c.statement(single, fn, stmnt)
			c.emitLine(fn, ast.Position)
			jumpToEnd := c.emitJump(fn, runtime.OpJump)
			toEndSwitch = append(toEndSwitch, jumpToEnd)

			// Next?
			c.label(fn, toNextCase)
		}

		// Pop condition if default
		c.emitLine(fn, ast.Position)
		c.emit(fn, runtime.OpPopTop)

		// Default value
		single := NewAtomScope(scope, AtomScopeTypeSingle)
		c.statement(single, fn, defaultValue)

		// End?
		for _, jump := range toEndSwitch {
			c.label(fn, jump)
		}
	}
}

func (c *AtomCompile) whileStatement(scope *AtomScope, fn *runtime.AtomValue, ast *AtomAst) {
	loopScope := NewAtomScope(scope, AtomScopeTypeLoop)
	loopStart := c.here(fn)
	if ast.Ast1.AstType == AstTypeBlock && hasDeclairation(ast.Ast1.Arr0) {
		c.emitInt(fn, runtime.OpEnterBlock, 1)
		c.expression(loopScope, fn, ast.Ast0)
		c.emitLine(fn, ast.Position)
		toEnd := c.emitJump(fn, runtime.OpPopJumpIfFalse)
		blockScope := NewAtomScope(loopScope, AtomScopeTypeBlock)
		for _, stmt := range ast.Ast1.Arr0 {
			c.statement(blockScope, fn, stmt)
		}
		c.emitLine(fn, ast.Position)
		c.emitInt(fn, runtime.OpAbsoluteJump, loopStart)
		c.label(fn, toEnd)
		c.emitInt(fn, runtime.OpExitBlock, 1)
	} else if ast.Ast1.AstType == AstTypeBlock {
		c.expression(loopScope, fn, ast.Ast0)
		c.emitLine(fn, ast.Position)
		toEnd := c.emitJump(fn, runtime.OpPopJumpIfFalse)
		blockScope := NewAtomScope(loopScope, AtomScopeTypeBlockNoEnv)
		for _, stmt := range ast.Ast1.Arr0 {
			c.statement(blockScope, fn, stmt)
		}
		c.emitLine(fn, ast.Position)
		c.emitInt(fn, runtime.OpAbsoluteJump, loopStart)
		c.label(fn, toEnd)
	} else {
		c.expression(loopScope, fn, ast.Ast0)
		c.emitLine(fn, ast.Position)
		toEnd := c.emitJump(fn, runtime.OpPopJumpIfFalse)
		single := NewAtomScope(loopScope, AtomScopeTypeSingle)
		c.statement(single, fn, ast.Ast1)
		c.emitLine(fn, ast.Position)
		c.emitInt(fn, runtime.OpAbsoluteJump, loopStart)
		c.label(fn, toEnd)
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
	loopStart := c.here(fn)
	if ast.Ast1.AstType == AstTypeBlock && hasDeclairation(ast.Ast1.Arr0) {
		blockScope := NewAtomScope(loopScope, AtomScopeTypeBlock)
		c.emitInt(fn, runtime.OpEnterBlock, 1)
		for _, stmt := range ast.Ast1.Arr0 {
			c.statement(blockScope, fn, stmt)
		}

		// Modify opcodes for continue
		for _, continueAddress := range loopScope.Continues {
			fn.Value.(*runtime.AtomCode).Code[continueAddress-1] = runtime.OpJump
			c.label(fn, continueAddress)
		}

		// check condition
		c.expression(loopScope, fn, ast.Ast0)
		c.emitLine(fn, ast.Position)
		toEnd := c.emitJump(fn, runtime.OpPopJumpIfFalse)

		// jump to start
		c.emitLine(fn, ast.Position)
		c.emitInt(fn, runtime.OpAbsoluteJump, loopStart)
		c.label(fn, toEnd)
		c.emitInt(fn, runtime.OpExitBlock, 1)
	} else if ast.Ast1.AstType == AstTypeBlock {
		blockScope := NewAtomScope(loopScope, AtomScopeTypeBlockNoEnv)
		for _, stmt := range ast.Ast1.Arr0 {
			c.statement(blockScope, fn, stmt)
		}

		// Modify opcodes for continue
		for _, continueAddress := range loopScope.Continues {
			fn.Value.(*runtime.AtomCode).Code[continueAddress-1] = runtime.OpJump
			c.label(fn, continueAddress)
		}

		// check condition
		c.expression(loopScope, fn, ast.Ast0)
		c.emitLine(fn, ast.Position)
		toEnd := c.emitJump(fn, runtime.OpPopJumpIfFalse)

		// jump to start
		c.emitLine(fn, ast.Position)
		c.emitInt(fn, runtime.OpAbsoluteJump, loopStart)
		c.label(fn, toEnd)
	} else {
		single := NewAtomScope(loopScope, AtomScopeTypeSingle)
		c.statement(single, fn, ast.Ast1)

		// Modify opcodes for continue
		for _, continueAddress := range loopScope.Continues {
			fn.Value.(*runtime.AtomCode).Code[continueAddress-1] = runtime.OpJump
			c.label(fn, continueAddress)
		}

		// check condition
		c.expression(loopScope, fn, ast.Ast0)
		c.emitLine(fn, ast.Position)
		toEnd := c.emitJump(fn, runtime.OpPopJumpIfFalse)

		// jump to start
		c.emitLine(fn, ast.Position)
		c.emitInt(fn, runtime.OpAbsoluteJump, loopStart)
		c.label(fn, toEnd)
	}
	for _, breakAddress := range loopScope.Breaks {
		c.label(fn, breakAddress)
	}
}

func (c *AtomCompile) forStatement(scope *AtomScope, fn *runtime.AtomValue, ast *AtomAst) {
	loopScope := NewAtomScope(scope, AtomScopeTypeLoop)

	initializer := ast.Ast0
	condition := ast.Ast1
	updater := ast.Ast2
	body := ast.Ast3

	if initializer != nil {
		c.statement(loopScope, fn, initializer)
	}

	loopStart := c.here(fn)
	jump0 := 0

	if condition != nil {
		c.expression(loopScope, fn, condition)
		c.emitLine(fn, ast.Position)
		jump0 = c.emitJump(fn, runtime.OpPopJumpIfFalse)
	}

	// body
	single := NewAtomScope(loopScope, AtomScopeTypeSingle)
	if body.AstType == AstTypeBlock && hasDeclairation(body.Arr0) {
		blockScope := NewAtomScope(loopScope, AtomScopeTypeBlock)
		c.emitInt(fn, runtime.OpEnterBlock, 1)
		for _, stmt := range body.Arr0 {
			c.statement(blockScope, fn, stmt)
		}

		// Modify opcodes for continue
		for _, continueAddress := range loopScope.Continues {
			fn.Value.(*runtime.AtomCode).Code[continueAddress-1] = runtime.OpJump
			c.label(fn, continueAddress)
		}

		// Updater
		if updater != nil {
			c.expression(loopScope, fn, updater)
			c.emitLine(fn, updater.Position)
			c.emit(fn, runtime.OpPopTop)
		}

		// Loop
		c.emitLine(fn, ast.Position)
		c.emitInt(fn, runtime.OpAbsoluteJump, loopStart)

		// End loop
		if condition != nil {
			c.label(fn, jump0)
		}

		c.emitInt(fn, runtime.OpExitBlock, 1)
	} else if body.AstType == AstTypeBlock {
		blockScope := NewAtomScope(loopScope, AtomScopeTypeBlockNoEnv)
		for _, stmt := range body.Arr0 {
			c.statement(blockScope, fn, stmt)
		}

		// Modify opcodes for continue
		for _, continueAddress := range loopScope.Continues {
			fn.Value.(*runtime.AtomCode).Code[continueAddress-1] = runtime.OpJump
			c.label(fn, continueAddress)
		}

		// Updater
		if updater != nil {
			c.expression(loopScope, fn, updater)
			c.emitLine(fn, updater.Position)
			c.emit(fn, runtime.OpPopTop)
		}

		// Loop
		c.emitLine(fn, ast.Position)
		c.emitInt(fn, runtime.OpAbsoluteJump, loopStart)

		// End loop
		if condition != nil {
			c.label(fn, jump0)
		}
	} else {
		c.statement(single, fn, body)

		// Modify opcodes for continue
		for _, continueAddress := range loopScope.Continues {
			fn.Value.(*runtime.AtomCode).Code[continueAddress-1] = runtime.OpJump
			c.label(fn, continueAddress)
		}

		// Updater
		if updater != nil {
			c.expression(loopScope, fn, updater)
			c.emitLine(fn, updater.Position)
			c.emit(fn, runtime.OpPopTop)
		}

		// Loop
		c.emitLine(fn, ast.Position)
		c.emitInt(fn, runtime.OpAbsoluteJump, loopStart)

		// End loop
		if condition != nil {
			c.label(fn, jump0)
		}
	}

	for _, breakAddress := range loopScope.Breaks {
		c.label(fn, breakAddress)
	}
}

func (c *AtomCompile) program(ast *AtomAst) *runtime.AtomValue {
	globalScope := NewAtomScope(nil, AtomScopeTypeGlobal)
	programFunc := runtime.NewAtomGenericValue(
		runtime.AtomTypeFunc,
		runtime.NewAtomCode(c.parser.tokenizer.file, "script", false, 0),
	)
	body := ast.Arr1
	for _, stmt := range body {
		c.statement(globalScope, programFunc, stmt)
	}
	c.emitLine(programFunc, ast.Position)
	c.emit(programFunc, runtime.OpLoadNull)
	c.emitLine(programFunc, ast.Position)
	c.emit(programFunc, runtime.OpReturn)

	// Resolve
	for _, pendingVariable := range c.pendingVariables {
		/*
		 * Variables that have been referenced but do not exist yet,
		 * we mark them as global captured variables
		 */
		if !c.isDefined(globalScope, pendingVariable.ast.Str0) {
			Error(
				c.parser.tokenizer.file,
				c.parser.tokenizer.data,
				fmt.Sprintf("Variable %s is not defined", pendingVariable.ast.Str0),
				pendingVariable.ast.Position,
			)
		}
	}

	return programFunc
}

func (c *AtomCompile) Export() int {
	ast := c.parser.Parse()
	if exists := c.state.SaveModule(c.parser.tokenizer.file); exists {
		panic("Already exists (not handled properly)!")
	}
	globalScope := NewAtomScope(nil, AtomScopeTypeGlobal)
	programFunc := runtime.NewAtomGenericValue(
		runtime.AtomTypeFunc,
		runtime.NewAtomCode(c.parser.tokenizer.file, "script", false, 0),
	)
	body := ast.Arr1
	for _, stmt := range body {
		c.statement(globalScope, programFunc, stmt)
	}

	// Get global vars
	count := 0
	for _, name := range globalScope.Names {
		if !name.global {
			continue
		}
		count++
		c.emitLine(programFunc, ast.Position)
		c.emitStr(programFunc, runtime.OpLoadName, name.name)
		c.emitLine(programFunc, ast.Position)
		c.emitStr(programFunc, runtime.OpLoadStr, name.name)
	}
	c.emitLine(programFunc, ast.Position)
	c.emitInt(programFunc, runtime.OpMakeModule, count)

	c.emitLine(programFunc, ast.Position)
	c.emit(programFunc, runtime.OpReturn)

	// Resolve
	for _, pendingVariable := range c.pendingVariables {
		/*
		 * Variables that have been referenced but do not exist yet,
		 * we mark them as global captured variables
		 */
		if !c.isDefined(globalScope, pendingVariable.ast.Str0) {
			Error(
				c.parser.tokenizer.file,
				c.parser.tokenizer.data,
				fmt.Sprintf("Variable %s is not defined", pendingVariable.ast.Str0),
				pendingVariable.ast.Position,
			)
		}
	}

	return c.state.SaveFunction(programFunc)
}

func (c *AtomCompile) Compile() *runtime.AtomValue {
	return c.program(c.parser.Parse())
}
