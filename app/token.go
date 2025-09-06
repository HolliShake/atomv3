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

/*
 * Export everything for Compiler
 */
type Token struct {
	Type     TokenType
	Value    string
	Position Position
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
	typeStr := t.Type.String()
	return fmt.Sprintf("Token { %s %s %d:%d-%d:%d }", typeStr, t.Value, t.Position.LineStart, t.Position.ColmStart, t.Position.LineEnded, t.Position.ColmEnded)
}
