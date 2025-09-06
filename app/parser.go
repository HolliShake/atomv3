package main

import "fmt"

type Parser struct {
	tokenizer *Tokenizer
	lookahead Token
}

func NewParser(tokenizer *Tokenizer) *Parser {
	return &Parser{tokenizer: tokenizer}
}

func (p *Parser) checkT(ttype TokenType) bool {
	return p.lookahead.ttype == ttype
}

func (p *Parser) checkV(value string) bool {
	return p.lookahead.value == value && (p.lookahead.ttype == TokenTypeSym ||
		p.lookahead.ttype == TokenTypeKey ||
		p.lookahead.ttype == TokenTypeIdn)
}

func (p *Parser) acceptT(ttype TokenType) {
	if p.checkT(ttype) {
		p.lookahead = p.tokenizer.NextToken()
	}
	expected := ttype.String()
	Error(p.tokenizer.file, p.tokenizer.data, fmt.Sprintf("Expected %s, got %s", expected, p.lookahead.ttype.String()), p.lookahead.position)
}

func (p *Parser) acceptV(value string) bool {
	if p.checkV(value) {
		p.lookahead = p.tokenizer.NextToken()
		return true
	}
	expected := value
	Error(p.tokenizer.file, p.tokenizer.data, fmt.Sprintf("Expected %s, got %s", expected, p.lookahead.value), p.lookahead.position)
	return false
}

func (p *Parser) terminal() *Ast {

}
