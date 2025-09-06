package main

import (
	"unicode"
)

type Tokenizer struct {
	file   string
	data   []rune
	pos    int
	line   int
	column int
}

func NewTokenizer(file string, data string) *Tokenizer {
	return &Tokenizer{
		file:   file,
		data:   []rune(data),
		pos:    0,
		line:   1,
		column: 1,
	}
}

// isKeyword checks if a string is a JavaScript keyword
func (t *Tokenizer) isKeyword(word string) bool {
	keywords := []string{
		KeyClass, KeyFunc, KeyVar, KeyConst, KeyLocal, KeyEnum,
		KeyImport, KeyContinue, KeyBreak, KeyReturn,
		KeyIf, KeyElse, KeySwitch, KeyCase, KeyDefault, KeyFor,
		KeyWhile, KeyDoWhile, KetTrue, KetFalse, KetNull, KeyNew,
	}

	for _, keyword := range keywords {
		if word == keyword {
			return true
		}
	}
	return false
}

// isLetter checks if a rune is a letter (including Unicode letters)
func (t *Tokenizer) isLetter(r rune) bool {
	return unicode.IsLetter(r) || r == '_' || r == '$'
}

// isDigit checks if a rune is a digit
func (t *Tokenizer) isDigit(r rune) bool {
	return unicode.IsDigit(r)
}

// isHexDigit checks if a rune is a hexadecimal digit
func (t *Tokenizer) isHexDigit(r rune) bool {
	return unicode.IsDigit(r) || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}

// isWhitespace checks if a rune is whitespace
func (t *Tokenizer) isWhitespace(r rune) bool {
	return unicode.IsSpace(r)
}

// containsDecimalOrScientific checks if a number string contains decimal point or scientific notation
func containsDecimalOrScientific(numStr string) bool {
	for _, r := range numStr {
		if r == '.' || r == 'e' || r == 'E' {
			return true
		}
	}
	return false
}

// current returns the current rune or 0 if at end
func (t *Tokenizer) current() rune {
	if t.pos >= len(t.data) {
		return 0
	}
	return t.data[t.pos]
}

// peek returns the next rune without advancing position
func (t *Tokenizer) peek() rune {
	if t.pos+1 >= len(t.data) {
		return 0
	}
	return t.data[t.pos+1]
}

// advance moves to the next character
func (t *Tokenizer) advance() {
	if t.pos < len(t.data) {
		if t.data[t.pos] == '\n' {
			t.line++
			t.column = 1
		} else {
			t.column++
		}
		t.pos++
	}
}

// skipWhitespace skips whitespace and comments
func (t *Tokenizer) skipWhitespace() {
	for t.pos < len(t.data) {
		r := t.current()
		if t.isWhitespace(r) {
			t.advance()
		} else if r == '/' && t.peek() == '/' {
			// Single line comment
			for t.pos < len(t.data) && t.current() != '\n' {
				t.advance()
			}
		} else if r == '/' && t.peek() == '*' {
			// Multi-line comment
			t.advance() // skip /
			t.advance() // skip *
			for t.pos < len(t.data)-1 {
				if t.current() == '*' && t.peek() == '/' {
					t.advance() // skip *
					t.advance() // skip /
					break
				}
				t.advance()
			}
		} else {
			break
		}
	}
}

// readString reads a string literal with Unicode support
func (t *Tokenizer) readString() (string, error) {
	quote := t.current()
	t.advance() // skip opening quote

	var result []rune
	for t.pos < len(t.data) {
		r := t.current()
		if r == quote {
			t.advance() // skip closing quote
			return string(result), nil
		} else if r == '\n' || r == '\r' {
			// Unescaped newline - break the string here
			// Don't advance past the newline, let the caller handle it
			return string(result), nil
		} else if r == '\\' {
			t.advance() // skip backslash
			if t.pos >= len(t.data) {
				break
			}
			escaped := t.current()
			switch escaped {
			case 'n':
				result = append(result, '\n')
			case 't':
				result = append(result, '\t')
			case 'r':
				result = append(result, '\r')
			case '\\':
				result = append(result, '\\')
			case '"', '\'':
				result = append(result, escaped)
			case 'u':
				// Unicode escape sequence \uXXXX
				t.advance() // skip 'u'
				if t.pos+4 <= len(t.data) {
					hex := string(t.data[t.pos : t.pos+4])
					// Parse hexadecimal Unicode codepoint
					var code rune
					for _, r := range hex {
						var digit int
						if r >= '0' && r <= '9' {
							digit = int(r - '0')
						} else if r >= 'a' && r <= 'f' {
							digit = int(r - 'a' + 10)
						} else if r >= 'A' && r <= 'F' {
							digit = int(r - 'A' + 10)
						} else {
							// Invalid hex digit
							result = append(result, '\\', 'u')
							break
						}
						code = code*16 + rune(digit)
					}
					if len(hex) == 4 {
						result = append(result, code)
						t.pos += 4
					}
				} else {
					result = append(result, '\\', 'u')
				}
			default:
				result = append(result, '\\', escaped)
			}
		} else {
			result = append(result, r)
		}
		t.advance()
	}
	return string(result), nil
}

// readNumber reads a numeric literal
func (t *Tokenizer) readNumber() string {
	var result []rune

	// Handle hexadecimal
	if t.current() == '0' && (t.peek() == 'x' || t.peek() == 'X') {
		result = append(result, t.current())
		t.advance()
		result = append(result, t.current())
		t.advance()
		for t.pos < len(t.data) && t.isHexDigit(t.current()) {
			result = append(result, t.current())
			t.advance()
		}
		return string(result)
	}

	// Handle decimal numbers
	for t.pos < len(t.data) && t.isDigit(t.current()) {
		result = append(result, t.current())
		t.advance()
	}

	// Handle decimal point
	if t.current() == '.' && t.pos+1 < len(t.data) && t.isDigit(t.peek()) {
		result = append(result, t.current())
		t.advance()
		for t.pos < len(t.data) && t.isDigit(t.current()) {
			result = append(result, t.current())
			t.advance()
		}
	}

	// Handle scientific notation
	if t.pos < len(t.data) && (t.current() == 'e' || t.current() == 'E') {
		result = append(result, t.current())
		t.advance()
		if t.pos < len(t.data) && (t.current() == '+' || t.current() == '-') {
			result = append(result, t.current())
			t.advance()
		}
		for t.pos < len(t.data) && t.isDigit(t.current()) {
			result = append(result, t.current())
			t.advance()
		}
	}

	return string(result)
}

// readIdentifier reads an identifier or keyword
func (t *Tokenizer) readIdentifier() string {
	var result []rune
	for t.pos < len(t.data) && (t.isLetter(t.current()) || t.isDigit(t.current())) {
		result = append(result, t.current())
		t.advance()
	}
	return string(result)
}

// NextToken returns the next token from the input
func (t *Tokenizer) NextToken() Token {
	t.skipWhitespace()

	if t.pos >= len(t.data) {
		return Token{
			ttype:    TokenTypeEof,
			value:    "",
			position: Position{LineStart: t.line, LineEnded: t.line, ColumnStart: t.column, ColumnEnded: t.column},
		}
	}

	startLine := t.line
	startColumn := t.column
	r := t.current()

	// String literals
	if r == '"' || r == '\'' {
		value, _ := t.readString()
		return Token{
			ttype:    TokenTypeStr,
			value:    value,
			position: Position{LineStart: startLine, LineEnded: t.line, ColumnStart: startColumn, ColumnEnded: t.column},
		}
	}

	// Numbers
	if t.isDigit(r) {
		value := t.readNumber()
		// Determine if it's a floating point number or integer
		tokenType := TokenTypeInt
		if containsDecimalOrScientific(value) {
			tokenType = TokenTypeNum
		}
		return Token{
			ttype:    tokenType,
			value:    value,
			position: Position{LineStart: startLine, LineEnded: t.line, ColumnStart: startColumn, ColumnEnded: t.column},
		}
	}

	// Identifiers and keywords
	if t.isLetter(r) {
		value := t.readIdentifier()
		tokenType := TokenTypeIdn
		if t.isKeyword(value) {
			tokenType = TokenTypeKey
		}
		return Token{
			ttype:    tokenType,
			value:    value,
			position: Position{LineStart: startLine, LineEnded: t.line, ColumnStart: startColumn, ColumnEnded: t.column},
		}
	}

	// Symbols and operators
	t.advance()
	symbol := string(r)

	// Handle multi-character operators
	if t.pos < len(t.data) {
		next := t.current()
		switch {
		case symbol == "=" && next == '=':
			t.advance()
			symbol = "=="
		case symbol == "!" && next == '=':
			t.advance()
			symbol = "!="
		case symbol == "=" && next == '=' && t.peek() == '=':
			t.advance()
			t.advance()
			symbol = "==="
		case symbol == "!" && next == '=' && t.peek() == '=':
			t.advance()
			t.advance()
			symbol = "!=="
		case symbol == "&" && next == '&':
			t.advance()
			symbol = "&&"
		case symbol == "|" && next == '|':
			t.advance()
			symbol = "||"
		case symbol == "+" && next == '+':
			t.advance()
			symbol = "++"
		case symbol == "-" && next == '-':
			t.advance()
			symbol = "--"
		case symbol == "+" && next == '=':
			t.advance()
			symbol = "+="
		case symbol == "-" && next == '=':
			t.advance()
			symbol = "-="
		case symbol == "*" && next == '=':
			t.advance()
			symbol = "*="
		case symbol == "/" && next == '=':
			t.advance()
			symbol = "/="
		case symbol == ">" && next == '=':
			t.advance()
			symbol = ">="
		case symbol == "<" && next == '=':
			t.advance()
			symbol = "<="
		case symbol == ">" && next == '>':
			t.advance()
			symbol = ">>"
		case symbol == "<" && next == '<':
			t.advance()
			symbol = "<<"
		}
	}

	return Token{
		ttype:    TokenTypeSym,
		value:    symbol,
		position: Position{LineStart: startLine, LineEnded: t.line, ColumnStart: startColumn, ColumnEnded: t.column},
	}
}
