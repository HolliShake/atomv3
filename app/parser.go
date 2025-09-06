package main

import "fmt"

/*
 * Hide everything.
 */
type Parser struct {
	tokenizer *Tokenizer
	lookahead Token
}

func NewParser(tokenizer *Tokenizer) *Parser {
	return &Parser{tokenizer: tokenizer}
}

func (p *Parser) checkT(ttype TokenType) bool {
	return p.lookahead.Type == ttype
}

func (p *Parser) checkV(value string) bool {
	return p.lookahead.Value == value && (p.lookahead.Type == TokenTypeSym ||
		p.lookahead.Type == TokenTypeKey ||
		p.lookahead.Type == TokenTypeIdn)
}

func (p *Parser) acceptT(ttype TokenType) {
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

func (p *Parser) acceptV(value string) {
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

func (p *Parser) terminal() *Ast {
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

func (p *Parser) multiplicative() *Ast {
	ast := p.terminal()
	for p.checkT(TokenTypeSym) && (p.checkV("*") || p.checkV("/") || p.checkV("%")) {
		opt := p.lookahead
		p.acceptV(opt.Value)

		rhs := p.terminal()
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

func (p *Parser) additive() *Ast {
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

func (p *Parser) statement() *Ast {
	if p.checkT(TokenTypeKey) && p.checkV(KeyFunc) {
		return p.function()
	}
	return p.additive()
}

func (p *Parser) function() *Ast {
	start := p.lookahead.Position
	ended := start
	p.acceptV(KeyFunc)
	name := p.terminal()
	p.acceptV("(")
	// Parameters
	params := make([]*Ast, 0)
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
	body := make([]*Ast, 0)
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

func (p *Parser) program() *Ast {
	start := p.lookahead.Position
	ended := start
	body := make([]*Ast, 0)
	ast := p.statement()
	for ast != nil {
		body = append(body, ast)
		ast = p.statement()
	}
	ended = p.lookahead.Position
	p.acceptT(TokenTypeEof)
	return NewProgram(body, start.Merge(ended))
}

func (p *Parser) Parse() *Ast {
	p.lookahead = p.tokenizer.NextToken()
	return p.program()
}
