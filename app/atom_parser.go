package main

import "fmt"

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
	return nil
}

func (p *AtomParser) keyValue() *AtomAst {
	key := p.terminal()
	if key == nil {
		return nil
	}
	p.acceptV(":")
	val := p.primary()
	return NewKeyValue(key, val, key.Position.Merge(val.Position))
}

func (p *AtomParser) group() *AtomAst {
	start := p.lookahead.Position
	ended := start
	if p.checkT(TokenTypeSym) && p.checkV("(") {
		p.acceptV("(")
		ast := p.mandatory()
		p.acceptV(")")
		return ast
	} else if p.checkT(TokenTypeSym) && p.checkV("[") {
		p.acceptV("[")
		elements := make([]*AtomAst, 0)
		n := p.primary()
		if n != nil {
			elements = append(elements, n)
			for p.checkT(TokenTypeSym) && p.checkV(",") {
				p.acceptV(",")
				n = p.primary()
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
		elements := make([]*AtomAst, 0)
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
	}
	return p.terminal()
}

func (p *AtomParser) memberOrCall() *AtomAst {
	ast := p.group()
	for p.checkT(TokenTypeSym) && (p.checkV("(")) {
		if p.checkV("(") {
			args := make([]*AtomAst, 0)
			p.acceptV("(")
			// arguments
			arg := p.primary()
			if arg != nil {
				args = append(args, arg)
				for p.checkT(TokenTypeSym) && p.checkV(",") {
					p.acceptV(",")
					arg = p.primary()
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

func (p *AtomParser) multiplicative() *AtomAst {
	ast := p.memberOrCall()
	for p.checkT(TokenTypeSym) && (p.checkV("*") || p.checkV("/") || p.checkV("%")) {
		opt := p.lookahead
		p.acceptV(opt.Value)

		rhs := p.memberOrCall()
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

func (p *AtomParser) primary() *AtomAst {
	return p.logical()
}

func (p *AtomParser) mandatory() *AtomAst {
	ast := p.primary()
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
	if p.checkT(TokenTypeKey) && p.checkV(KeyFunc) {
		return p.function()
	} else if p.checkT(TokenTypeSym) && p.checkV("{") {
		return p.block()
	} else if p.checkT(TokenTypeKey) && p.checkV(KeyVar) {
		return p.varStatement()
	} else if p.checkT(TokenTypeKey) && p.checkV(KeyConst) {
		return p.constStatement()
	} else if p.checkT(TokenTypeKey) && p.checkV(KeyLocal) {
		return p.localStatement()
	} else if p.checkT(TokenTypeKey) && p.checkV(KeyIf) {
		return p.ifStatement()
	} else if p.checkT(TokenTypeKey) && p.checkV(KeyReturn) {
		return p.returnStatement()
	}
	return p.expressionStatement()
}

func (p *AtomParser) function() *AtomAst {
	start := p.lookahead.Position
	ended := start
	p.acceptV(KeyFunc)
	name := p.terminal()
	p.acceptV("(")
	// Parameters
	params := make([]*AtomAst, 0)
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
	body := make([]*AtomAst, 0)
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
	body := make([]*AtomAst, 0)
	stmt := p.statement()
	for stmt != nil {
		body = append(body, stmt)
		stmt = p.statement()
	}
	ended = p.lookahead.Position
	p.acceptV("}")
	return NewBlock(body, start.Merge(ended))
}

func (p *AtomParser) varStatement() *AtomAst {
	start := p.lookahead.Position
	ended := start
	p.acceptV(KeyVar)

	keys := make([]*AtomAst, 0)
	vals := make([]*AtomAst, 0)

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

	keys := make([]*AtomAst, 0)
	vals := make([]*AtomAst, 0)

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

	keys := make([]*AtomAst, 0)
	vals := make([]*AtomAst, 0)

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
	condition := p.primary()
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

func (p *AtomParser) returnStatement() *AtomAst {
	start := p.lookahead.Position
	ended := start
	p.acceptV(KeyReturn)
	expr := p.primary()
	p.acceptV(";")
	return NewReturnStatement(
		expr,
		start.Merge(ended),
	)
}

func (p *AtomParser) expressionStatement() *AtomAst {
	expr := p.primary()
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
	body := make([]*AtomAst, 0)
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
