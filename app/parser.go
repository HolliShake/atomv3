package main

import (
	"fmt"
)

/*
 * Hide everything.
 */
type AtomParser struct {
	tokenizer *AtomTokenizer
	lookahead AtomToken
}

func NewAtomParser(tokenizer *AtomTokenizer) *AtomParser {
	return &AtomParser{tokenizer: tokenizer}
}

func (p *AtomParser) checkT(ttype AtomTokenType) bool {
	return p.lookahead.Type == ttype
}

func (p *AtomParser) checkV(value string) bool {
	return p.lookahead.Value == value && (p.lookahead.Type == TokenTypeSym ||
		p.lookahead.Type == TokenTypeKey ||
		p.lookahead.Type == TokenTypeIdn)
}

func (p *AtomParser) acceptT(ttype AtomTokenType) {
	if p.checkT(ttype) {
		p.lookahead = p.tokenizer.NextToken()
		return
	}
	expected := ttype.String()
	Error(
		p.tokenizer.file,
		p.tokenizer.data,
		fmt.Sprintf("Expected %s, got %s", expected, p.lookahead.Type.String()),
		p.lookahead.Position,
	)
}

func (p *AtomParser) acceptV(value string) {
	if p.checkV(value) {
		p.lookahead = p.tokenizer.NextToken()
		return
	}
	expected := value
	Error(
		p.tokenizer.file,
		p.tokenizer.data,
		fmt.Sprintf("Expected %s, got %s", expected, p.lookahead.Value),
		p.lookahead.Position,
	)
}

func (p *AtomParser) terminal() *AtomAst {
	if p.checkT(TokenTypeIdn) {
		ast := NewTerminal(
			AstTypeIdn,
			p.lookahead.Value,
			p.lookahead.Position,
		)
		p.acceptT(TokenTypeIdn)
		return ast
	}
	if p.checkT(TokenTypeInt) {
		ast := NewTerminal(
			AstTypeInt,
			p.lookahead.Value,
			p.lookahead.Position,
		)
		p.acceptT(TokenTypeInt)
		return ast
	}
	if p.checkT(TokenTypeNum) {
		ast := NewTerminal(
			AstTypeNum,
			p.lookahead.Value,
			p.lookahead.Position,
		)
		p.acceptT(TokenTypeNum)
		return ast
	}
	if p.checkT(TokenTypeStr) {
		ast := NewTerminal(
			AstTypeStr,
			p.lookahead.Value,
			p.lookahead.Position,
		)
		p.acceptT(TokenTypeStr)
		return ast
	}
	if p.checkT(TokenTypeKey) && (p.checkV(KetTrue) || p.checkV(KetFalse)) {
		ast := NewTerminal(
			AstTypeBool,
			p.lookahead.Value,
			p.lookahead.Position,
		)
		p.acceptT(TokenTypeKey)
		return ast
	}
	if p.checkT(TokenTypeKey) && p.checkV(KetNull) {
		ast := NewTerminal(
			AstTypeNull,
			p.lookahead.Value,
			p.lookahead.Position,
		)
		p.acceptT(TokenTypeKey)
		return ast
	}
	return nil
}

func (p *AtomParser) keyValue() *AtomAst {
	key := p.terminal()
	if key == nil {
		return nil
	}
	p.acceptV(":")
	val := p.expression()
	return NewKeyValue(key, val, key.Position.Merge(val.Position))
}

func (p *AtomParser) primary() *AtomAst {
	start := p.lookahead.Position
	ended := start
	if p.checkT(TokenTypeSym) && p.checkV("(") {
		p.acceptV("(")
		ast := p.mandatory()
		p.acceptV(")")
		return ast
	} else if p.checkT(TokenTypeSym) && p.checkV("[") {
		p.acceptV("[")
		elements := []*AtomAst{}
		n := p.expression()
		if n != nil {
			elements = append(elements, n)
			for p.checkT(TokenTypeSym) && p.checkV(",") {
				p.acceptV(",")
				n = p.expression()
				if n == nil {
					Error(
						p.tokenizer.file,
						p.tokenizer.data,
						"Expected expression after comma",
						p.lookahead.Position,
					)
					return nil
				}
				elements = append(elements, n)
			}
		}
		ended = p.lookahead.Position
		p.acceptV("]")
		return NewArray(elements, start.Merge(ended))
	} else if p.checkT(TokenTypeSym) && p.checkV("{") {
		p.acceptV("{")
		elements := []*AtomAst{}
		n := p.keyValue()
		if n != nil {
			elements = append(elements, n)
			for p.checkT(TokenTypeSym) && p.checkV(",") {
				p.acceptV(",")
				n = p.keyValue()
				if n == nil {
					Error(
						p.tokenizer.file,
						p.tokenizer.data,
						"Expected key-value pair after comma",
						p.lookahead.Position,
					)
					return nil
				}
				elements = append(elements, n)
			}
		}
		ended = p.lookahead.Position
		p.acceptV("}")
		return NewObject(elements, start.Merge(ended))
	} else if p.checkT(TokenTypeKey) && (p.checkV(KeyAsync) || p.checkV(KeyFunc)) {
		async := p.checkV(KeyAsync)
		start := p.lookahead.Position
		ended := start

		astType := AstTypeFunctionExpression
		if async {
			astType = AstTypeAsyncFunctionExpression
			p.acceptV(KeyAsync)
		}
		p.acceptV(KeyFunc)
		p.acceptV("(")
		// Parameters
		params := []*AtomAst{}
		param := p.terminal()
		if param != nil {
			params = append(params, param)
			for p.checkT(TokenTypeSym) && p.checkV(",") {
				p.acceptV(",")
				param = p.terminal()
				params = append(params, param)
			}
		}
		p.acceptV(")")
		p.acceptV("{")
		// Body
		body := []*AtomAst{}
		stmt := p.statement()
		if stmt != nil {
			for stmt != nil {
				body = append(body, stmt)
				stmt = p.statement()
			}
		}
		ended = p.lookahead.Position
		p.acceptV("}")

		return NewFunctionExpression(
			astType,
			params,
			body,
			start.Merge(ended),
		)
	}
	return p.terminal()
}

func (p *AtomParser) memberOrCall() *AtomAst {
	ast := p.primary()
	for p.checkT(TokenTypeSym) && (p.checkV(".") || p.checkV("[") || p.checkV("(")) {

		if p.checkV(".") {
			p.acceptV(".")
			key := p.terminal()
			if key == nil {
				Error(
					p.tokenizer.file,
					p.tokenizer.data,
					"Expected identifier",
					p.lookahead.Position,
				)
				return nil
			}
			ast = NewMember(
				ast,
				key,
				ast.Position.Merge(key.Position),
			)
		} else if p.checkV("[") {
			p.acceptV("[")
			index := p.expression()
			if index == nil {
				Error(
					p.tokenizer.file,
					p.tokenizer.data,
					"Expected expression",
					p.lookahead.Position,
				)
				return nil
			}
			p.acceptV("]")
			ast = NewIndex(
				ast,
				index,
				ast.Position.Merge(index.Position),
			)
		} else if p.checkV("(") {
			args := []*AtomAst{}
			p.acceptV("(")
			// arguments
			arg := p.expression()
			if arg != nil {
				args = append(args, arg)
				for p.checkT(TokenTypeSym) && p.checkV(",") {
					p.acceptV(",")
					arg = p.expression()
					if arg == nil {
						Error(
							p.tokenizer.file,
							p.tokenizer.data,
							"Expected expression after comma",
							p.lookahead.Position,
						)
						return nil
					}
					args = append(args, arg)
				}
			}
			p.acceptV(")")
			ast = NewCall(ast, args, ast.Position.Merge(p.lookahead.Position))
		}
	}
	return ast
}

func (p *AtomParser) allocation() *AtomAst {
	if !(p.checkT(TokenTypeKey) && p.checkV(KeyNew)) {
		return p.memberOrCall()
	}

	start := p.lookahead.Position
	ended := start

	p.acceptV(KeyNew)

	call := p.allocation()
	if call == nil {
		Error(
			p.tokenizer.file,
			p.tokenizer.data,
			"Expected expression after new",
			p.lookahead.Position,
		)
		return nil
	}

	ended = call.Position

	return NewAllocation(
		call,
		start.Merge(ended),
	)
}

func (p *AtomParser) unary() *AtomAst {
	if p.checkT(TokenTypeSym) && (p.checkV("!") || p.checkV("+") || p.checkV("-")) {
		opt := p.lookahead
		p.acceptV(opt.Value)
		rhs := p.allocation()
		if rhs == nil {
			Error(
				p.tokenizer.file,
				p.tokenizer.data,
				fmt.Sprintf("Expected expression after %s, got %s", opt.Value, p.lookahead.Type.String()),
				p.lookahead.Position,
			)
			return nil
		}
		return NewUnary(
			opt,
			rhs,
			rhs.Position.Merge(p.lookahead.Position),
		)
	} else if p.checkT(TokenTypeKey) && p.checkV(KeyTypeof) {
		opt := p.lookahead
		p.acceptV(opt.Value)
		rhs := p.allocation()
		if rhs == nil {
			Error(
				p.tokenizer.file,
				p.tokenizer.data,
				fmt.Sprintf("Expected expression after %s, got %s", opt.Value, p.lookahead.Type.String()),
				p.lookahead.Position,
			)
			return nil
		}
		return NewUnary(
			opt,
			rhs,
			rhs.Position.Merge(p.lookahead.Position),
		)
	} else if p.checkT(TokenTypeKey) && p.checkV(KeyAwait) {
		opt := p.lookahead
		p.acceptV(opt.Value)
		rhs := p.allocation()
		if rhs == nil {
			Error(
				p.tokenizer.file,
				p.tokenizer.data,
				fmt.Sprintf("Expected expression after %s, got %s", opt.Value, p.lookahead.Type.String()),
				p.lookahead.Position,
			)
			return nil
		}
		return NewUnary(
			opt,
			rhs,
			rhs.Position.Merge(p.lookahead.Position),
		)
	}
	return p.allocation()
}

func (p *AtomParser) multiplicative() *AtomAst {
	ast := p.unary()
	for p.checkT(TokenTypeSym) && (p.checkV("*") || p.checkV("/") || p.checkV("%")) {
		opt := p.lookahead
		p.acceptV(opt.Value)

		rhs := p.unary()
		if rhs == nil {
			Error(
				p.tokenizer.file,
				p.tokenizer.data,
				fmt.Sprintf("Expected expression after %s, got %s", opt.Value, p.lookahead.Type.String()),
				p.lookahead.Position,
			)
			return nil
		}
		ast = NewBinary(
			ast,
			opt,
			rhs,
			ast.Position.Merge(rhs.Position),
		)
	}
	return ast
}

func (p *AtomParser) additive() *AtomAst {
	ast := p.multiplicative()
	for p.checkT(TokenTypeSym) && (p.checkV("+") || p.checkV("-")) {
		opt := p.lookahead
		p.acceptV(opt.Value)

		rhs := p.multiplicative()
		if rhs == nil {
			Error(
				p.tokenizer.file,
				p.tokenizer.data,
				fmt.Sprintf("Expected expression after %s, got %s", opt.Value, p.lookahead.Type.String()),
				p.lookahead.Position,
			)
			return nil
		}
		ast = NewBinary(
			ast,
			opt,
			rhs,
			ast.Position.Merge(rhs.Position),
		)
	}
	return ast
}

func (p *AtomParser) shift() *AtomAst {
	ast := p.additive()
	for p.checkT(TokenTypeSym) && (p.checkV(">>") || p.checkV("<<")) {
		opt := p.lookahead
		p.acceptV(opt.Value)

		rhs := p.additive()
		if rhs == nil {
			Error(
				p.tokenizer.file,
				p.tokenizer.data,
				fmt.Sprintf("Expected expression after %s, got %s", opt.Value, p.lookahead.Type.String()),
				p.lookahead.Position,
			)
			return nil
		}
		ast = NewBinary(
			ast,
			opt,
			rhs,
			ast.Position.Merge(rhs.Position),
		)
	}
	return ast
}

func (p *AtomParser) relational() *AtomAst {
	ast := p.shift()
	for p.checkT(TokenTypeSym) && (p.checkV("<") || p.checkV("<=") || p.checkV(">") || p.checkV(">=")) {
		opt := p.lookahead
		p.acceptV(opt.Value)

		rhs := p.shift()
		if rhs == nil {
			Error(
				p.tokenizer.file,
				p.tokenizer.data,
				fmt.Sprintf("Expected expression after %s, got %s", opt.Value, p.lookahead.Type.String()),
				p.lookahead.Position,
			)
			return nil
		}
		ast = NewBinary(
			ast,
			opt,
			rhs,
			ast.Position.Merge(rhs.Position),
		)
	}
	return ast
}

func (p *AtomParser) equality() *AtomAst {
	ast := p.relational()
	for p.checkT(TokenTypeSym) && (p.checkV("==") || p.checkV("!=") || p.checkV("===") || p.checkV("!==")) {
		opt := p.lookahead
		p.acceptV(opt.Value)

		rhs := p.relational()
		if rhs == nil {
			Error(
				p.tokenizer.file,
				p.tokenizer.data,
				fmt.Sprintf("Expected expression after %s, got %s", opt.Value, p.lookahead.Type.String()),
				p.lookahead.Position,
			)
			return nil
		}
		ast = NewBinary(
			ast,
			opt,
			rhs,
			ast.Position.Merge(rhs.Position),
		)
	}
	return ast
}

func (p *AtomParser) bitwise() *AtomAst {
	ast := p.equality()
	for p.checkT(TokenTypeSym) && (p.checkV("&") || p.checkV("|") || p.checkV("^")) {
		opt := p.lookahead
		p.acceptV(opt.Value)

		rhs := p.equality()
		if rhs == nil {
			Error(
				p.tokenizer.file,
				p.tokenizer.data,
				fmt.Sprintf("Expected expression after %s, got %s", opt.Value, p.lookahead.Type.String()),
				p.lookahead.Position,
			)
			return nil
		}
		ast = NewBinary(
			ast,
			opt,
			rhs,
			ast.Position.Merge(rhs.Position),
		)
	}
	return ast
}

func (p *AtomParser) logical() *AtomAst {
	ast := p.bitwise()
	for p.checkT(TokenTypeSym) && (p.checkV("&&") || p.checkV("||")) {
		opt := p.lookahead
		p.acceptV(opt.Value)

		rhs := p.bitwise()
		if rhs == nil {
			Error(
				p.tokenizer.file,
				p.tokenizer.data,
				fmt.Sprintf("Expected expression after %s, got %s", opt.Value, p.lookahead.Type.String()),
				p.lookahead.Position,
			)
			return nil
		}
		ast = NewBinary(
			ast,
			opt,
			rhs,
			ast.Position.Merge(rhs.Position),
		)
	}
	return ast
}

func (p *AtomParser) assign() *AtomAst {
	ast := p.logical()
	for p.checkT(TokenTypeSym) && p.checkV("=") {
		opt := p.lookahead
		p.acceptV(opt.Value)

		rhs := p.logical()
		if rhs == nil {
			Error(
				p.tokenizer.file,
				p.tokenizer.data,
				fmt.Sprintf("Expected expression after %s, got %s", opt.Value, p.lookahead.Type.String()),
				p.lookahead.Position,
			)
			return nil
		}
		ast = NewBinary(
			ast,
			opt,
			rhs,
			ast.Position.Merge(rhs.Position),
		)
	}
	return ast
}

func (p *AtomParser) multiplicativeAssign() *AtomAst {
	ast := p.assign()
	for p.checkT(TokenTypeSym) && (p.checkV("*=") || p.checkV("/=") || p.checkV("%=")) {
		opt := p.lookahead
		p.acceptV(opt.Value)

		rhs := p.assign()
		if rhs == nil {
			Error(
				p.tokenizer.file,
				p.tokenizer.data,
				fmt.Sprintf("Expected expression after %s, got %s", opt.Value, p.lookahead.Type.String()),
				p.lookahead.Position,
			)
			return nil
		}
		ast = NewBinary(
			ast,
			opt,
			rhs,
			ast.Position.Merge(rhs.Position),
		)
	}
	return ast
}

func (p *AtomParser) additiveAssign() *AtomAst {
	ast := p.multiplicativeAssign()
	for p.checkT(TokenTypeSym) && (p.checkV("+=") || p.checkV("-=")) {
		opt := p.lookahead
		p.acceptV(opt.Value)

		rhs := p.multiplicativeAssign()
		if rhs == nil {
			Error(
				p.tokenizer.file,
				p.tokenizer.data,
				fmt.Sprintf("Expected expression after %s, got %s", opt.Value, p.lookahead.Type.String()),
				p.lookahead.Position,
			)
			return nil
		}
		ast = NewBinary(
			ast,
			opt,
			rhs,
			ast.Position.Merge(rhs.Position),
		)
	}
	return ast
}

func (p *AtomParser) shiftAssign() *AtomAst {
	ast := p.additiveAssign()
	for p.checkT(TokenTypeSym) && (p.checkV(">>=") || p.checkV("<<=")) {
		opt := p.lookahead
		p.acceptV(opt.Value)

		rhs := p.additiveAssign()
		if rhs == nil {
			Error(
				p.tokenizer.file,
				p.tokenizer.data,
				fmt.Sprintf("Expected expression after %s, got %s", opt.Value, p.lookahead.Type.String()),
				p.lookahead.Position,
			)
			return nil
		}
		ast = NewBinary(
			ast,
			opt,
			rhs,
			ast.Position.Merge(rhs.Position),
		)
	}
	return ast
}

func (p *AtomParser) bitwiseAssign() *AtomAst {
	ast := p.shiftAssign()
	for p.checkT(TokenTypeSym) && (p.checkV("&=") || p.checkV("|=") || p.checkV("^=")) {
		opt := p.lookahead
		p.acceptV(opt.Value)

		rhs := p.shiftAssign()
		if rhs == nil {
			Error(
				p.tokenizer.file,
				p.tokenizer.data,
				fmt.Sprintf("Expected expression after %s, got %s", opt.Value, p.lookahead.Type.String()),
				p.lookahead.Position,
			)
			return nil
		}
		ast = NewBinary(
			ast,
			opt,
			rhs,
			ast.Position.Merge(rhs.Position),
		)
	}
	return ast
}

func (p *AtomParser) ifExpression() *AtomAst {
	start := p.lookahead.Position
	ended := start
	if !(p.checkT(TokenTypeKey) && p.checkV(KeyIf)) {
		return p.bitwiseAssign()
	}
	p.acceptV(KeyIf)
	p.acceptV("(")
	condition := p.ifExpression()
	if condition == nil {
		Error(
			p.tokenizer.file,
			p.tokenizer.data,
			"Expected expression",
			p.lookahead.Position,
		)
		return nil
	}
	p.acceptV(")")
	thenValue := p.ifExpression()

	p.acceptV(KeyElse)

	ended = p.lookahead.Position
	elseValue := p.ifExpression()

	if elseValue == nil {
		Error(
			p.tokenizer.file,
			p.tokenizer.data,
			"Expected expression",
			p.lookahead.Position,
		)
		return nil
	}

	return NewIfExpression(
		condition,
		thenValue,
		elseValue,
		start.Merge(ended),
	)
}

func (p *AtomParser) switchExpression() *AtomAst {
	start := p.lookahead.Position
	ended := start
	ast := p.ifExpression()
	if !(p.checkT(TokenTypeKey) && p.checkV(KeySwitch)) {
		return ast
	}
	p.acceptV(KeySwitch)
	p.acceptV("{")
	cases := []*AtomAst{}
	values := []*AtomAst{}

	for p.checkT(TokenTypeKey) && p.checkV(KeyCase) {
		p.acceptV(KeyCase)
		p.acceptV("(")
		patterns := []*AtomAst{}

		start = p.lookahead.Position
		ended = start

		pattern := p.expression()
		if pattern == nil {
			Error(
				p.tokenizer.file,
				p.tokenizer.data,
				"Expected pattern",
				p.lookahead.Position,
			)
			return nil
		}
		patterns = append(patterns, pattern)
		for p.checkT(TokenTypeSym) && p.checkV(",") {
			p.acceptV(",")
			pattern = p.expression()
			if pattern == nil {
				Error(
					p.tokenizer.file,
					p.tokenizer.data,
					"Expected pattern",
					p.lookahead.Position,
				)
				return nil
			}
			ended = p.lookahead.Position
			patterns = append(patterns, pattern)
		}
		p.acceptV(")")
		p.acceptV("=>")

		value := p.switchExpression()

		if value == nil {
			Error(
				p.tokenizer.file,
				p.tokenizer.data,
				"Expected expression",
				p.lookahead.Position,
			)
			return nil
		}
		cases = append(cases, NewArray(
			patterns,
			start.Merge(ended),
		))

		values = append(values, value)
	}

	p.acceptV(KeyDefault)
	p.acceptV("=>")
	value := p.switchExpression()

	if value == nil {
		Error(
			p.tokenizer.file,
			p.tokenizer.data,
			"Expected expression",
			p.lookahead.Position,
		)
		return nil
	}

	ended = p.lookahead.Position
	p.acceptV("}")

	return NewSwitchExpression(
		ast,
		cases,
		values,
		value,
		start.Merge(ended),
	)
}

func (p *AtomParser) catchExpression() *AtomAst {
	start := p.lookahead.Position
	ended := start

	ast := p.switchExpression()

	if !(p.checkT(TokenTypeKey) && p.checkV(KeyCatch)) {
		return ast
	}

	p.acceptV(KeyCatch)
	p.acceptV("(")
	variable := p.terminal()
	if variable == nil {
		Error(
			p.tokenizer.file,
			p.tokenizer.data,
			"Expected identifier",
			p.lookahead.Position,
		)
		return nil
	}
	p.acceptV(")")

	p.acceptV("{")

	body := []*AtomAst{}

	stmt := p.statement()
	for stmt != nil {
		body = append(body, stmt)
		stmt = p.statement()
	}

	ended = p.lookahead.Position
	p.acceptV("}")

	return NewCatchExpression(
		ast,
		variable,
		body,
		start.Merge(ended),
	)
}

func (p *AtomParser) expression() *AtomAst {
	return p.catchExpression()
}

func (p *AtomParser) mandatory() *AtomAst {
	ast := p.expression()
	if ast == nil {
		Error(
			p.tokenizer.file,
			p.tokenizer.data,
			"Expected expression",
			p.lookahead.Position,
		)
		return nil
	}
	return ast
}

func (p *AtomParser) statement() *AtomAst {
	if p.checkT(TokenTypeKey) && p.checkV(KeyClass) {
		return p.classStatement()
	} else if p.checkT(TokenTypeKey) && p.checkV(KeyEnum) {
		return p.enumStatement()
	} else if p.checkT(TokenTypeKey) && p.checkV(KeyAsync) {
		return p.function(true)
	} else if p.checkT(TokenTypeKey) && p.checkV(KeyFunc) {
		return p.function(false)
	} else if p.checkT(TokenTypeSym) && p.checkV("{") {
		return p.block()
	} else if p.checkT(TokenTypeKey) && p.checkV(KeyImport) {
		return p.importStatement()
	} else if p.checkT(TokenTypeKey) && p.checkV(KeyVar) {
		return p.varStatement()
	} else if p.checkT(TokenTypeKey) && p.checkV(KeyConst) {
		return p.constStatement()
	} else if p.checkT(TokenTypeKey) && p.checkV(KeyLocal) {
		return p.localStatement()
	} else if p.checkT(TokenTypeKey) && p.checkV(KeyIf) {
		return p.ifStatement()
	} else if p.checkT(TokenTypeKey) && p.checkV(KeySwitch) {
		return p.switchStatement()
	} else if p.checkT(TokenTypeKey) && p.checkV(KeyWhile) {
		return p.whileStatement()
	} else if p.checkT(TokenTypeKey) && p.checkV(KeyDo) {
		return p.doWhileStatement()
	} else if p.checkT(TokenTypeKey) && p.checkV(KeyBreak) {
		return p.breakStatement()
	} else if p.checkT(TokenTypeKey) && p.checkV(KeyContinue) {
		return p.continueStatement()
	} else if p.checkT(TokenTypeKey) && p.checkV(KeyReturn) {
		return p.returnStatement()
	}
	return p.expressionStatement()
}

func (p *AtomParser) classStatement() *AtomAst {
	start := p.lookahead.Position
	ended := start

	p.acceptV(KeyClass)

	name := p.terminal()
	if name == nil {
		Error(
			p.tokenizer.file,
			p.tokenizer.data,
			"Expected identifier",
			p.lookahead.Position,
		)
		return nil
	}

	var base *AtomAst = nil

	if p.checkT(TokenTypeKey) && p.checkV(KeyExtends) {
		p.acceptV(KeyExtends)
		base = p.expression()
		if base == nil {
			Error(
				p.tokenizer.file,
				p.tokenizer.data,
				"Expected identifier",
				p.lookahead.Position,
			)
			return nil
		}
	}

	p.acceptV("{")

	body := []*AtomAst{}
	stmt := p.statement()
	for stmt != nil {
		body = append(body, stmt)
		stmt = p.statement()
	}

	ended = p.lookahead.Position
	p.acceptV("}")

	return NewClassStatement(
		name,
		base,
		body,
		start.Merge(ended),
	)
}

func (p *AtomParser) enumStatement() *AtomAst {
	start := p.lookahead.Position
	ended := start

	p.acceptV(KeyEnum)

	name := p.terminal()
	if name == nil {
		Error(
			p.tokenizer.file,
			p.tokenizer.data,
			"Expected identifier",
			p.lookahead.Position,
		)
		return nil
	}

	p.acceptV("{")

	names := []*AtomAst{}
	values := []*AtomAst{}

	var nameN *AtomAst = p.terminal()
	var valueN *AtomAst = nil

	if nameN == nil {
		Error(
			p.tokenizer.file,
			p.tokenizer.data,
			"Expected identifier",
			p.lookahead.Position,
		)
		return nil
	}

	if p.checkT(TokenTypeSym) && p.checkV("=") {
		p.acceptV("=")
		valueN = p.mandatory()
	}

	names = append(names, nameN)
	values = append(values, valueN)

	for p.checkT(TokenTypeSym) && p.checkV(",") {
		p.acceptV(",")
		nameN = p.terminal()
		if nameN == nil {
			Error(
				p.tokenizer.file,
				p.tokenizer.data,
				"Expected identifier",
				p.lookahead.Position,
			)
			return nil
		}

		valueN = nil
		if p.checkT(TokenTypeSym) && p.checkV("=") {
			p.acceptV("=")
			valueN = p.mandatory()
		}

		names = append(names, nameN)
		values = append(values, valueN)
	}

	ended = p.lookahead.Position
	p.acceptV("}")

	return NewEnumStatement(
		name,
		names,
		values,
		start.Merge(ended),
	)
}

func (p *AtomParser) function(async bool) *AtomAst {
	start := p.lookahead.Position
	ended := start
	astType := AstTypeFunction
	if async {
		astType = AstTypeAsyncFunction
		p.acceptV(KeyAsync)
	}
	p.acceptV(KeyFunc)
	name := p.terminal()
	p.acceptV("(")
	// Parameters
	params := []*AtomAst{}
	param := p.terminal()
	if param != nil {
		params = append(params, param)
		for p.checkT(TokenTypeSym) && p.checkV(",") {
			p.acceptV(",")
			param = p.terminal()
			params = append(params, param)
		}
	}
	p.acceptV(")")
	p.acceptV("{")
	// Body
	body := []*AtomAst{}
	stmt := p.statement()
	if stmt != nil {
		for stmt != nil {
			body = append(body, stmt)
			stmt = p.statement()
		}
	}
	ended = p.lookahead.Position
	p.acceptV("}")
	return NewFunction(
		astType,
		name,
		params,
		body,
		start.Merge(ended),
	)
}

func (p *AtomParser) block() *AtomAst {
	start := p.lookahead.Position
	ended := start
	p.acceptV("{")
	body := []*AtomAst{}
	stmt := p.statement()
	for stmt != nil {
		body = append(body, stmt)
		stmt = p.statement()
	}
	ended = p.lookahead.Position
	p.acceptV("}")
	return NewBlock(body, start.Merge(ended))
}

func (p *AtomParser) importStatement() *AtomAst {
	start := p.lookahead.Position
	ended := start

	p.acceptV(KeyImport)

	names := []*AtomAst{}
	var path *AtomAst = nil

	if p.checkT(TokenTypeSym) && p.checkV("[") {
		p.acceptV("[")
		nameN := p.terminal()
		if nameN == nil {
			Error(
				p.tokenizer.file,
				p.tokenizer.data,
				"Expected identifier",
				p.lookahead.Position,
			)
			return nil
		}
		names = append(names, nameN)
		for p.checkT(TokenTypeSym) && p.checkV(",") {
			p.acceptV(",")
			nameN = p.terminal()
			if nameN == nil {
				Error(
					p.tokenizer.file,
					p.tokenizer.data,
					"Expected identifier",
					p.lookahead.Position,
				)
				return nil
			}
			names = append(names, nameN)
		}
		p.acceptV("]")

		p.acceptV(KeyFrom)

	}

	path = p.terminal()

	ended = p.lookahead.Position
	p.acceptV(";")

	return NewImportStatement(
		path,
		names,
		start.Merge(ended),
	)
}

func (p *AtomParser) varStatement() *AtomAst {
	start := p.lookahead.Position
	ended := start
	p.acceptV(KeyVar)

	keys := []*AtomAst{}
	vals := []*AtomAst{}

	var key *AtomAst = p.terminal()
	var val *AtomAst = nil
	if key == nil {
		Error(
			p.tokenizer.file,
			p.tokenizer.data,
			"Expected identifier",
			p.lookahead.Position,
		)
		return nil
	}
	if p.checkT(TokenTypeSym) && p.checkV("=") {
		p.acceptV("=")
		val = p.mandatory()
	}
	keys = append(keys, key)
	vals = append(vals, val)
	for p.checkT(TokenTypeSym) && p.checkV(",") {
		p.acceptV(",")
		key = p.terminal()
		if key == nil {
			Error(
				p.tokenizer.file,
				p.tokenizer.data,
				"Expected identifier",
				p.lookahead.Position,
			)
			return nil
		}
		val = nil
		if p.checkT(TokenTypeSym) && p.checkV("=") {
			p.acceptV("=")
			val = p.mandatory()
		}

		keys = append(keys, key)
		vals = append(vals, val)
	}
	ended = p.lookahead.Position
	p.acceptV(";")
	return NewVarStatement(
		keys,
		vals,
		start.Merge(ended),
	)
}

func (p *AtomParser) constStatement() *AtomAst {
	start := p.lookahead.Position
	ended := start
	p.acceptV(KeyConst)

	keys := []*AtomAst{}
	vals := []*AtomAst{}

	var key *AtomAst = p.terminal()
	var val *AtomAst = nil
	if key == nil {
		Error(
			p.tokenizer.file,
			p.tokenizer.data,
			"Expected identifier",
			p.lookahead.Position,
		)
		return nil
	}
	if p.checkT(TokenTypeSym) && p.checkV("=") {
		p.acceptV("=")
		val = p.mandatory()
	}
	keys = append(keys, key)
	vals = append(vals, val)
	for p.checkT(TokenTypeSym) && p.checkV(",") {
		p.acceptV(",")
		key = p.terminal()
		if key == nil {
			Error(
				p.tokenizer.file,
				p.tokenizer.data,
				"Expected identifier",
				p.lookahead.Position,
			)
			return nil
		}
		val = nil
		if p.checkT(TokenTypeSym) && p.checkV("=") {
			p.acceptV("=")
			val = p.mandatory()
		}

		keys = append(keys, key)
		vals = append(vals, val)
	}
	ended = p.lookahead.Position
	p.acceptV(";")
	return NewConstStatement(
		keys,
		vals,
		start.Merge(ended),
	)
}

func (p *AtomParser) localStatement() *AtomAst {
	start := p.lookahead.Position
	ended := start
	p.acceptV(KeyLocal)

	keys := []*AtomAst{}
	vals := []*AtomAst{}

	var key *AtomAst = p.terminal()
	var val *AtomAst = nil
	if key == nil {
		Error(
			p.tokenizer.file,
			p.tokenizer.data,
			"Expected identifier",
			p.lookahead.Position,
		)
		return nil
	}
	if p.checkT(TokenTypeSym) && p.checkV("=") {
		p.acceptV("=")
		val = p.mandatory()
	}
	keys = append(keys, key)
	vals = append(vals, val)
	for p.checkT(TokenTypeSym) && p.checkV(",") {
		p.acceptV(",")
		key = p.terminal()
		if key == nil {
			Error(
				p.tokenizer.file,
				p.tokenizer.data,
				"Expected identifier",
				p.lookahead.Position,
			)
			return nil
		}
		val = nil
		if p.checkT(TokenTypeSym) && p.checkV("=") {
			p.acceptV("=")
			val = p.mandatory()
		}

		keys = append(keys, key)
		vals = append(vals, val)
	}
	ended = p.lookahead.Position
	p.acceptV(";")
	return NewLocalStatement(
		keys,
		vals,
		start.Merge(ended),
	)
}

func (p *AtomParser) ifStatement() *AtomAst {
	start := p.lookahead.Position
	ended := start
	p.acceptV(KeyIf)
	p.acceptV("(")
	condition := p.expression()
	if condition == nil {
		Error(
			p.tokenizer.file,
			p.tokenizer.data,
			"Expected expression",
			p.lookahead.Position,
		)
		return nil
	}
	p.acceptV(")")
	thenValue := p.statement()
	if thenValue == nil {
		Error(
			p.tokenizer.file,
			p.tokenizer.data,
			"Expected statement",
			p.lookahead.Position,
		)
		return nil
	}

	var elseValue *AtomAst = nil
	ended = thenValue.Position

	if p.checkT(TokenTypeKey) && p.checkV(KeyElse) {
		p.acceptV(KeyElse)
		elseValue = p.statement()
		if elseValue == nil {
			Error(
				p.tokenizer.file,
				p.tokenizer.data,
				"Expected statement",
				p.lookahead.Position,
			)
			return nil
		}
		ended = elseValue.Position
	}
	return NewIfStatement(
		condition,
		thenValue,
		elseValue,
		start.Merge(ended),
	)
}

func (p *AtomParser) switchStatement() *AtomAst {
	start := p.lookahead.Position
	ended := start
	p.acceptV(KeySwitch)
	p.acceptV("(")
	condition := p.expression()
	if condition == nil {
		Error(
			p.tokenizer.file,
			p.tokenizer.data,
			"Expected expression",
			p.lookahead.Position,
		)
		return nil
	}
	p.acceptV(")")
	p.acceptV("{")
	cases := []*AtomAst{}
	values := []*AtomAst{}
	for p.checkT(TokenTypeKey) && p.checkV(KeyCase) {
		p.acceptV(KeyCase)
		p.acceptV("(")
		patterns := []*AtomAst{}
		pattern := p.expression()
		if pattern == nil {
			Error(
				p.tokenizer.file,
				p.tokenizer.data,
				"Expected pattern",
				p.lookahead.Position,
			)
			return nil
		}
		patterns = append(patterns, pattern)
		for p.checkT(TokenTypeSym) && p.checkV(",") {
			p.acceptV(",")
			pattern = p.expression()
			if pattern == nil {
				Error(
					p.tokenizer.file,
					p.tokenizer.data,
					"Expected pattern",
					p.lookahead.Position,
				)
				return nil
			}
			patterns = append(patterns, pattern)
		}
		p.acceptV(")")

		p.acceptV(":")

		value := p.statement()
		if value == nil {
			Error(
				p.tokenizer.file,
				p.tokenizer.data,
				"Expected statement",
				p.lookahead.Position,
			)
			return nil
		}
		cases = append(cases, NewArray(
			patterns,
			start.Merge(ended),
		))
		values = append(values, value)
	}
	p.acceptV(KeyDefault)
	p.acceptV(":")
	value := p.statement()
	if value == nil {
		Error(
			p.tokenizer.file,
			p.tokenizer.data,
			"Expected statement",
			p.lookahead.Position,
		)
		return nil
	}
	ended = p.lookahead.Position
	p.acceptV("}")
	return NewSwitchStatement(
		condition,
		cases,
		values,
		value,
		start.Merge(ended),
	)
}

func (p *AtomParser) whileStatement() *AtomAst {
	start := p.lookahead.Position
	ended := start
	p.acceptV(KeyWhile)
	p.acceptV("(")
	condition := p.expression()
	if condition == nil {
		Error(
			p.tokenizer.file,
			p.tokenizer.data,
			"Expected expression",
			p.lookahead.Position,
		)
		return nil
	}
	p.acceptV(")")

	body := p.statement()
	if body == nil {
		Error(
			p.tokenizer.file,
			p.tokenizer.data,
			"Expected statement",
			p.lookahead.Position,
		)
		return nil
	}

	ended = body.Position

	return NewWhileStatement(
		condition,
		body,
		start.Merge(ended),
	)
}

func (p *AtomParser) doWhileStatement() *AtomAst {
	start := p.lookahead.Position
	ended := start
	p.acceptV(KeyDo)
	body := p.statement()
	if body == nil {
		Error(
			p.tokenizer.file,
			p.tokenizer.data,
			"Expected statement",
			p.lookahead.Position,
		)
		return nil
	}
	p.acceptV(KeyWhile)
	p.acceptV("(")
	condition := p.expression()
	if condition == nil {
		Error(
			p.tokenizer.file,
			p.tokenizer.data,
			"Expected expression",
			p.lookahead.Position,
		)
		return nil
	}
	ended = p.lookahead.Position
	p.acceptV(")")
	return NewDoWhileStatement(
		body,
		condition,
		start.Merge(ended),
	)
}

func (p *AtomParser) breakStatement() *AtomAst {
	start := p.lookahead.Position
	ended := start
	p.acceptV(KeyBreak)
	p.acceptV(";")
	return NewBreakStatement(
		start.Merge(ended),
	)
}

func (p *AtomParser) continueStatement() *AtomAst {
	start := p.lookahead.Position
	ended := start
	p.acceptV(KeyContinue)
	p.acceptV(";")
	return NewContinueStatement(
		start.Merge(ended),
	)
}

func (p *AtomParser) returnStatement() *AtomAst {
	start := p.lookahead.Position
	ended := start
	p.acceptV(KeyReturn)
	expr := p.expression()
	p.acceptV(";")
	return NewReturnStatement(
		expr,
		start.Merge(ended),
	)
}

func (p *AtomParser) expressionStatement() *AtomAst {
	expr := p.expression()
	if expr == nil {
		if p.checkV(";") {
			start := p.lookahead.Position
			ended := start
			for p.checkV(";") {
				ended = p.lookahead.Position
				p.acceptV(";")
			}
			return NewEmptyStatement(
				start.Merge(ended),
			)
		}
		return nil
	}
	ended := p.lookahead.Position
	p.acceptV(";")
	return NewExpressionStatement(
		expr,
		expr.Position.Merge(ended),
	)
}

func (p *AtomParser) program() *AtomAst {
	start := p.lookahead.Position
	ended := start
	body := []*AtomAst{}
	ast := p.statement()
	for ast != nil {
		body = append(body, ast)
		ast = p.statement()
	}
	ended = p.lookahead.Position
	p.acceptT(TokenTypeEof)
	return NewProgram(
		body,
		start.Merge(ended),
	)
}

func (p *AtomParser) Parse() *AtomAst {
	p.lookahead = p.tokenizer.NextToken()
	return p.program()
}
