package lexer

import (
	"unicode"
)

// Define token types
const (
	TokenEOF          = iota // End of file
	TokenIdentifier          // Variable and function names
	TokenKeyword             // def, class, if, else, etc.
	TokenOperator            // +, -, *, /, ==, !=, is, in, etc.
	TokenNumber              // 123, 3.14, etc.
	TokenString              // "hello", 'world'
	TokenNone                // Python-like None
	TokenTrue                // True
	TokenFalse               // False
	TokenIndent              // Indentation (4 spaces)
	TokenDedent              // Dedentation
	TokenNewline             // Newline
	TokenColon               // :
	TokenComma               // ,
	TokenParenOpen           // (
	TokenParenClose          // )
	TokenBracketOpen         // [
	TokenBracketClose        // ]
	TokenBraceOpen           // {
	TokenBraceClose          // }
	TokenAssert              // assert keyword
	TokenBreak               // break keyword
	TokenComment             // #
	TokenDot                 // .
	TokenIllegal             //
)

// Token structure
type Token struct {
	Type    int
	Literal string
	Line    int
	Column  int
}

// Lexer structure
type Lexer struct {
	input         string
	position      int
	readPosition  int
	currentChar   byte
	line          int
	column        int
	indentStack   []int    // Tracks indentation levels
	LastTwoTokens [2]Token // Tracks the last two tokens
}

// NewLexer Create a new lexer
func NewLexer(input string) *Lexer {
	l := &Lexer{input: input, indentStack: []int{0}}
	l.readChar()
	return l
}

// Read the next character in the input
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.currentChar = 0 // EOF
	} else {
		l.currentChar = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
	l.column++
}

// PeekChar Peek at the next character without advancing the position
func (l *Lexer) PeekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

// Skip whitespace (except for indentation)
func (l *Lexer) skipWhitespace() {
	for l.currentChar == ' ' || l.currentChar == '\t' {
		l.readChar()
	}
}

// HandleNewline Handle newlines and indentation after encountering a newline
func (l *Lexer) HandleNewline() Token {

	// Count spaces for indentation
	start := l.position
	for l.currentChar == ' ' || l.currentChar == '\t' {
		l.readChar()
	}
	indentLength := l.position - start

	// Compare indentation to the current level
	currentIndent := l.indentStack[len(l.indentStack)-1]
	if indentLength > currentIndent {
		l.indentStack = append(l.indentStack, indentLength)
		return Token{Type: TokenIndent, Literal: "INDENT", Line: l.line, Column: l.column}
	} else if indentLength < currentIndent {
		l.indentStack = l.indentStack[:len(l.indentStack)-1] // Pop the current indent level
		return Token{Type: TokenDedent, Literal: "DEDENT", Line: l.line, Column: l.column}
	}

	// No change in indentation level
	return Token{Type: TokenNewline, Literal: "\\n", Line: l.line, Column: l.column}
}

// HandleIndentAfterColonNewline Handle indentation after colon + newline
func (l *Lexer) HandleIndentAfterColonNewline() Token {
	// Count spaces for indentation
	start := l.position
	for l.currentChar == ' ' || l.currentChar == '\t' {
		l.readChar()
	}
	indentLength := l.position - start

	// Compare indentation to the current level
	currentIndent := l.indentStack[len(l.indentStack)-1]
	if indentLength > currentIndent { // New indent level
		l.indentStack = append(l.indentStack, indentLength)
		return Token{Type: TokenIndent, Literal: "INDENT", Line: l.line, Column: l.column}
	} else if indentLength < currentIndent { // Dedent level
		l.indentStack = l.indentStack[:len(l.indentStack)-1] // Pop the current indent level
		return Token{Type: TokenDedent, Literal: "DEDENT", Line: l.line, Column: l.column}
	}

	// No indentation change (stay at the same level)
	return Token{Type: TokenNewline, Literal: "\\n", Line: l.line, Column: l.column}
}

// UpdateLastTwoTokens Update last two tokens for colon + newline detection
func (l *Lexer) UpdateLastTwoTokens(tok Token) {
	l.LastTwoTokens[0] = l.LastTwoTokens[1]
	l.LastTwoTokens[1] = tok
}

// NextToken Get the next token from the input
func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	var tok Token
	tok.Line = l.line
	tok.Column = l.column

	switch l.currentChar {
	case '=':
		if l.PeekChar() == '=' {
			l.readChar()
			tok = Token{Type: TokenOperator, Literal: "==", Line: l.line, Column: l.column}
		} else {
			tok = Token{Type: TokenOperator, Literal: "=", Line: l.line, Column: l.column}
		}
	case '!':
		if l.PeekChar() == '=' {
			l.readChar()
			tok = Token{Type: TokenOperator, Literal: "!=", Line: l.line, Column: l.column}
		} else {
			tok = Token{Type: TokenIllegal, Literal: "!", Line: l.line, Column: l.column}
		}
	case '>':
		if l.PeekChar() == '=' {
			l.readChar()
			tok = Token{Type: TokenOperator, Literal: ">=", Line: l.line, Column: l.column}
		} else {
			tok = Token{Type: TokenOperator, Literal: ">", Line: l.line, Column: l.column}
		}
	case '<':
		if l.PeekChar() == '=' {
			l.readChar()
			tok = Token{Type: TokenOperator, Literal: "<=", Line: l.line, Column: l.column}
		} else {
			tok = Token{Type: TokenOperator, Literal: "<", Line: l.line, Column: l.column}
		}
	case ':':
		tok = Token{Type: TokenColon, Literal: ":", Line: l.line, Column: l.column}
		l.readChar() // Skip the colon
		//l.UpdateLastTwoTokens(tok)
		return tok
	case '#': // Comment
		l.skipComment()
		tok = l.NextToken() // Recurse to get the next non-comment token
	case '\n':
		// Emit a newline token first
		tok = Token{Type: TokenNewline, Literal: "\\n", Line: l.line, Column: l.column}
		l.readChar() // Consume the newline character
		//l.UpdateLastTwoTokens(tok)
		for l.PeekChar() == '\n' {
			l.readChar()
		}
		return tok
	case '+':
		tok = Token{Type: TokenOperator, Literal: "+", Line: l.line, Column: l.column}
	case '-':
		tok = Token{Type: TokenOperator, Literal: "-", Line: l.line, Column: l.column}
	case '*':
		tok = Token{Type: TokenOperator, Literal: "*", Line: l.line, Column: l.column}
	case '/':
		if l.PeekChar() == '/' {
			l.readChar()
			tok = Token{Type: TokenOperator, Literal: "//", Line: l.line, Column: l.column}
		} else {
			tok = Token{Type: TokenOperator, Literal: "/", Line: l.line, Column: l.column}
		}
	case '&':
		if l.PeekChar() == '&' {
			l.readChar()
			tok = Token{Type: TokenOperator, Literal: "&&", Line: l.line, Column: l.column}
		} else {
			tok = Token{Type: TokenIllegal, Literal: "&", Line: l.line, Column: l.column}
		}
	case '|':
		if l.PeekChar() == '|' {
			l.readChar()
			tok = Token{Type: TokenOperator, Literal: "||", Line: l.line, Column: l.column}
		} else {
			tok = Token{Type: TokenIllegal, Literal: "|", Line: l.line, Column: l.column}
		}
	case '"':
		tok = Token{Type: TokenString, Literal: l.readString('"'), Line: l.line, Column: l.column}
	case '\'':
		tok = Token{Type: TokenString, Literal: l.readString('\''), Line: l.line, Column: l.column}
	case '(':
		tok = Token{Type: TokenParenOpen, Literal: "(", Line: l.line, Column: l.column}
	case ')':
		tok = Token{Type: TokenParenClose, Literal: ")", Line: l.line, Column: l.column}
	case '{':
		tok = Token{Type: TokenBraceOpen, Literal: "{", Line: l.line, Column: l.column}
	case '}':
		tok = Token{Type: TokenBraceClose, Literal: "}", Line: l.line, Column: l.column}
	case '[':
		tok = Token{Type: TokenBracketOpen, Literal: "[", Line: l.line, Column: l.column}
	case ']':
		tok = Token{Type: TokenBracketClose, Literal: "]", Line: l.line, Column: l.column}
	case ',':
		tok = Token{Type: TokenComma, Literal: ",", Line: l.line, Column: l.column}
	case '.':
		tok = Token{Type: TokenDot, Literal: ".", Line: l.line, Column: l.column}
	default:
		// Handle identifiers, numbers, strings, etc.
		if isLetter(l.currentChar) {
			tok.Literal = l.readIdentifier()
			tok.Type = lookupKeyword(tok.Literal)
			return tok
		} else if isDigit(l.currentChar) {
			tok.Literal = l.readNumber()
			tok.Type = TokenNumber
			return tok
		} else {
			tok = Token{Type: TokenEOF, Literal: "", Line: l.line, Column: l.column}
		}
	}

	l.readChar() // Advance to the next character

	return tok
}

// Skip comments
func (l *Lexer) skipComment() {
	for l.currentChar != '\n' && l.currentChar != 0 {
		l.readChar()
	}
}

// Read identifiers (variable, function names)
func (l *Lexer) readIdentifier() string {
	start := l.position
	for isLetter(l.currentChar) || isDigit(l.currentChar) {
		l.readChar()
	}
	return l.input[start:l.position]
}

// Read numbers
func (l *Lexer) readNumber() string {
	start := l.position
	for isDigit(l.currentChar) || l.currentChar == '.' {
		l.readChar()
	}
	return l.input[start:l.position]
}

// Read strings (single or double quotes)
func (l *Lexer) readString(quote byte) string {
	start := l.position
	for {
		l.readChar()
		if l.currentChar == quote || l.currentChar == 0 {
			break
		}
	}
	return l.input[start : l.position+1]
}

// Check if a character is a letter
func isLetter(ch byte) bool {
	return unicode.IsLetter(rune(ch)) || ch == '_'
}

// Check if a character is a digit
func isDigit(ch byte) bool {
	return unicode.IsDigit(rune(ch))
}

// Lookup keywords and special values like None, True, False
func lookupKeyword(ident string) int {
	keywords := map[string]int{
		"def":     TokenKeyword,
		"class":   TokenKeyword,
		"if":      TokenKeyword,
		"else":    TokenKeyword,
		"elif":    TokenKeyword,
		"while":   TokenKeyword,
		"for":     TokenKeyword,
		"return":  TokenKeyword,
		"True":    TokenTrue,
		"False":   TokenFalse,
		"None":    TokenNone,
		"assert":  TokenAssert,
		"break":   TokenBreak,
		"in":      TokenOperator,
		"is":      TokenOperator,
		"print":   TokenKeyword,
		"and":     TokenOperator,
		"or":      TokenOperator,
		"not":     TokenOperator,
		"import":  TokenKeyword,
		"from":    TokenKeyword,
		"try":     TokenKeyword,
		"except":  TokenKeyword,
		"finally": TokenKeyword,
	}
	if tokType, ok := keywords[ident]; ok {
		return tokType
	}
	return TokenIdentifier
}
