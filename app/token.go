package main

import "fmt"

type TokenType int

const (
	TokenTypeKey TokenType = iota
	TokenTypeIdn
	TokenTypeInt
	TokenTypeNum
	TokenTypeStr
	TokenTypeSym
	TokenTypeEof
)

type Token struct {
	ttype    TokenType
	value    string
	position Position
}

func (t TokenType) String() string {
	switch t {
	case TokenTypeKey:
		return "KEYWORD"
	case TokenTypeIdn:
		return "IDENTIFIER"
	case TokenTypeInt:
		return "INTEGER"
	case TokenTypeNum:
		return "NUMBER"
	case TokenTypeStr:
		return "STRING"
	case TokenTypeSym:
		return "SYMBOL"
	case TokenTypeEof:
		return "EOF"
	}
	return "UNKNOWN"
}

func (t *Token) String() string {
	typeStr := t.ttype.String()
	return fmt.Sprintf("Token { %s %s %d:%d-%d:%d }", typeStr, t.value, t.position.LineStart, t.position.ColumnStart, t.position.LineEnded, t.position.ColumnEnded)
}
