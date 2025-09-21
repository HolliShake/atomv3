package main

import (
	"slices"
	"unicode"
)

/*
 * Hide everything.
 */
type AtomTokenizer struct {
	file   string
	data   []rune
	pos    int
	line   int
	column int
}

func NewAtomTokenizer(file string, data string) *AtomTokenizer {
	return &AtomTokenizer{
		file:   file,
		data:   []rune(data),
		pos:    0,
		line:   1,
		column: 1,
	}
}

// isKeyword checks if a string is a JavaScript keyword
func (t *AtomTokenizer) isKeyword(word string) bool {
	keywords := []string{
		KeyClass, KeyExtends, KeyAsync, KeyFunc, KeyVar, KeyConst, KeyLocal, KeyEnum,
		KeyImport, KeyFrom, KeyContinue, KeyBreak, KeyReturn,
		KeyIf, KeyElse, KeySwitch, KeyCase, KeyDefault, KeyCatch, KeyFor,
		KeyWhile, KeyDo, KetTrue, KetFalse, KetNull, KeyNew, KeyTypeof, KeyAwait,
	}

	return slices.Contains(keywords, word)
}

// isLetter checks if a rune is a letter (including Unicode letters)
func (t *AtomTokenizer) isLetter(r rune) bool {
	return unicode.IsLetter(r) || r == '_' || r == '$'
}

// isDigit checks if a rune is a digit
func (t *AtomTokenizer) isDigit(r rune) bool {
	return unicode.IsDigit(r)
}

// isHexDigit checks if a rune is a hexadecimal digit
func (t *AtomTokenizer) isHexDigit(r rune) bool {
	return unicode.IsDigit(r) || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}

// isBinaryDigit checks if a rune is a binary digit
func (t *AtomTokenizer) isBinaryDigit(r rune) bool {
	return r == '0' || r == '1'
}

// isOctalDigit checks if a rune is an octal digit
func (t *AtomTokenizer) isOctalDigit(r rune) bool {
	return r >= '0' && r <= '7'
}

// isWhitespace checks if a rune is whitespace
func (t *AtomTokenizer) isWhitespace(r rune) bool {
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
func (t *AtomTokenizer) current() rune {
	if t.pos >= len(t.data) {
		return 0
	}
	return t.data[t.pos]
}

// peek returns the next rune without advancing position
func (t *AtomTokenizer) peek() rune {
	if t.pos+1 >= len(t.data) {
		return 0
	}
	return t.data[t.pos+1]
}

// advance moves to the next character
func (t *AtomTokenizer) advance() {
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
func (t *AtomTokenizer) skipWhitespace() {
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
func (t *AtomTokenizer) readString() (string, error) {
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
func (t *AtomTokenizer) readNumber() string {
	var result []rune

	// Handle special number formats starting with 0
	if t.current() == '0' && t.pos+1 < len(t.data) {
		next := t.peek()

		// Handle hexadecimal (0x or 0X)
		if next == 'x' || next == 'X' {
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

		// Handle binary (0b or 0B)
		if next == 'b' || next == 'B' {
			result = append(result, t.current())
			t.advance()
			result = append(result, t.current())
			t.advance()
			for t.pos < len(t.data) && t.isBinaryDigit(t.current()) {
				result = append(result, t.current())
				t.advance()
			}
			return string(result)
		}

		// Handle octal (0o or 0O)
		if next == 'o' || next == 'O' {
			result = append(result, t.current())
			t.advance()
			result = append(result, t.current())
			t.advance()
			for t.pos < len(t.data) && t.isOctalDigit(t.current()) {
				result = append(result, t.current())
				t.advance()
			}
			return string(result)
		}
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
func (t *AtomTokenizer) readIdentifier() string {
	var result []rune
	for t.pos < len(t.data) && (t.isLetter(t.current()) || t.isDigit(t.current())) {
		result = append(result, t.current())
		t.advance()
	}
	return string(result)
}

// NextToken returns the next token from the input
func (t *AtomTokenizer) NextToken() AtomToken {
	t.skipWhitespace()

	if t.pos >= len(t.data) {
		return AtomToken{
			Type:     TokenTypeEof,
			Value:    "",
			Position: AtomPosition{LineStart: t.line, LineEnded: t.line, ColmStart: t.column, ColmEnded: t.column},
		}
	}

	startLine := t.line
	startColumn := t.column
	r := t.current()

	// String literals
	if r == '"' || r == '\'' {
		value, _ := t.readString()
		return AtomToken{
			Type:     TokenTypeStr,
			Value:    value,
			Position: AtomPosition{LineStart: startLine, LineEnded: t.line, ColmStart: startColumn, ColmEnded: t.column},
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
		return AtomToken{
			Type:     tokenType,
			Value:    value,
			Position: AtomPosition{LineStart: startLine, LineEnded: t.line, ColmStart: startColumn, ColmEnded: t.column},
		}
	}

	// Identifiers and keywords
	if t.isLetter(r) {
		value := t.readIdentifier()
		tokenType := TokenTypeIdn
		if t.isKeyword(value) {
			tokenType = TokenTypeKey
		}
		return AtomToken{
			Type:     tokenType,
			Value:    value,
			Position: AtomPosition{LineStart: startLine, LineEnded: t.line, ColmStart: startColumn, ColmEnded: t.column},
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
		case symbol == "=" && next == '>':
			t.advance()
			symbol = "=>"
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
		case symbol == "%" && next == '=':
			t.advance()
			symbol = "%="
		case symbol == ">" && next == '=':
			t.advance()
			symbol = ">="
		case symbol == "<" && next == '=':
			t.advance()
			symbol = "<="
		case symbol == ">" && next == '>':
			t.advance()
			symbol = ">>"
			if t.pos < len(t.data) && t.current() == '=' {
				t.advance()
				symbol = ">>="
			}
		case symbol == "<" && next == '<':
			t.advance()
			symbol = "<<"
			if t.pos < len(t.data) && t.current() == '=' {
				t.advance()
				symbol = "<<="
			}
		case symbol == "&" && next == '=':
			t.advance()
			symbol = "&="
		case symbol == "|" && next == '=':
			t.advance()
			symbol = "|="
		case symbol == "^" && next == '=':
			t.advance()
			symbol = "^="
		}
	}

	return AtomToken{
		Type:     TokenTypeSym,
		Value:    symbol,
		Position: AtomPosition{LineStart: startLine, LineEnded: t.line, ColmStart: startColumn, ColmEnded: t.column},
	}
}
