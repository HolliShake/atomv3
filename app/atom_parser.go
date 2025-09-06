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

func (p *AtomParser) memberOrCall() *AtomAst {
	ast := p.terminal()
	for p.checkT(TokenTypeSym) && (p.checkV("(")) {
		if p.checkV("(") {
			args := make([]*AtomAst, 0)
			p.acceptV("(")
			// arguments
			arg := p.terminal()
			if arg != nil {
				args = append(args, arg)
				for p.checkT(TokenTypeSym) && p.checkV(",") {
					p.acceptV(",")
					arg = p.terminal()
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
		ast = NewBinary(ast, opt, rhs, ast.Position.Merge(rhs.Position))
	}
	return ast
}

func (p *AtomParser) primary() *AtomAst {
	return p.additive()
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
	return NewFunction(name, params, body, start.Merge(ended))
}

func (p *AtomParser) returnStatement() *AtomAst {
	start := p.lookahead.Position
	ended := start
	p.acceptV(KeyReturn)
	expr := p.primary()
	p.acceptV(";")
	return NewReturnStatement(expr, start.Merge(ended))
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
			return NewEmptyStatement(start.Merge(ended))
		}
		return nil
	}
	ended := p.lookahead.Position
	p.acceptV(";")
	return NewExpressionStatement(expr, expr.Position.Merge(ended))
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
	return NewProgram(body, start.Merge(ended))
}

func (p *AtomParser) Parse() *AtomAst {
	p.lookahead = p.tokenizer.NextToken()
	return p.program()
}
