package main

import "fmt"

type AtomTokenType int

const (
	TokenTypeKey AtomTokenType = iota
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
type AtomToken struct {
	Type     AtomTokenType
	Value    string
	Position AtomPosition
}

func (t AtomTokenType) String() string {
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

func (t *AtomToken) String() string {
	typeStr := t.Type.String()
	return fmt.Sprintf("Token { %s %s %d:%d-%d:%d }", typeStr, t.Value, t.Position.LineStart, t.Position.ColmStart, t.Position.LineEnded, t.Position.ColmEnded)
}
