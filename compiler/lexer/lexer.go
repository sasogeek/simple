package lexer

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// TokenType represents the type of a token.
type TokenType string

// Token types.
const (
	TokenEOF          TokenType = "EOF"
	TokenIdentifier   TokenType = "IDENTIFIER"
	TokenNumber       TokenType = "NUMBER"
	TokenString       TokenType = "STRING"
	TokenOperator     TokenType = "OPERATOR"
	TokenKeyword      TokenType = "KEYWORD"
	TokenNewline      TokenType = "NEWLINE"
	TokenIndent       TokenType = "INDENT"
	TokenDedent       TokenType = "DEDENT"
	TokenIllegal      TokenType = "ILLEGAL"
	TokenColon        TokenType = ":"
	TokenSemicolon    TokenType = ";"
	TokenComma        TokenType = ","
	TokenParenOpen    TokenType = "("
	TokenParenClose   TokenType = ")"
	TokenBracketOpen  TokenType = "["
	TokenBracketClose TokenType = "]"
	TokenBraceOpen    TokenType = "{"
	TokenBraceClose   TokenType = "}"
	TokenDot          TokenType = "DOT"

	// Comparison Operators
	TokenEQ    TokenType = "=="
	TokenNotEQ TokenType = "!="
	TokenGT    TokenType = ">"
	TokenGTE   TokenType = ">="
	TokenLT    TokenType = "<"
	TokenLTE   TokenType = "<="

	// Boolean Literals
	TokenTrue  TokenType = "TRUE"
	TokenFalse TokenType = "FALSE"

	// Arithmetic Operators
	TokenPlus     TokenType = "+"
	TokenMinus    TokenType = "-"
	TokenAsterisk TokenType = "*"
	TokenSlash    TokenType = "/"
	TokenModulo   TokenType = "%"
	TokenBang     TokenType = "!"

	// Assignment Operator
	TokenAssign TokenType = "="

	TokenDefer TokenType = "defer"
	TokenGo    TokenType = "go"
)

// Token represents a lexical token.
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

// keywords maps keyword strings to their token types.
var keywords = map[string]TokenType{
	"def":    TokenKeyword, // Function definition
	"return": TokenKeyword,
	"if":     TokenKeyword,
	"else":   TokenKeyword,
	"elif":   TokenKeyword,
	"while":  TokenKeyword,
	"for":    TokenKeyword,
	"in":     TokenKeyword,
	"import": TokenKeyword,
	"defer":  TokenDefer,
	"go":     TokenGo,
	"print":  TokenIdentifier,
	"True":   TokenTrue,
	"False":  TokenFalse,
	"None":   TokenKeyword,
}

// LookupIdent checks if an identifier is a keyword and returns the appropriate token type.
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return TokenIdentifier
}

// Lexer represents a lexical scanner.
type Lexer struct {
	input         string
	position      int     // Current position in input (points to current char)
	readPosition  int     // Current reading position in input (after current char)
	ch            rune    // Current character under examination
	line          int     // Current line number
	column        int     // Current column number
	indentStack   []int   // Stack to keep track of indentation levels
	pendingTokens []Token // Queue for INDENT/DEDENT tokens
	AtNewLine     bool    // Indicates if the lexer is at the start of a new line
}

// NewLexer initializes a new Lexer.
func NewLexer(input string) *Lexer {
	l := &Lexer{
		input:         input,
		line:          1,
		column:        0,
		indentStack:   []int{0}, // Start with indentation level 0
		pendingTokens: []Token{},
		AtNewLine:     true, // Start at a new line
	}
	l.readChar()
	return l
}

// readChar reads the next character.
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		r, width := utf8.DecodeRuneInString(l.input[l.readPosition:])
		l.ch = r
		l.readPosition += width
	}

	if l.ch == '\n' {
		l.line++
		l.column = 0
		//l.AtNewLine = true
	} else {
		l.column++
	}

	l.position = l.readPosition
}

// peekChar peeks at the next character without advancing positions.
func (l *Lexer) peekChar() rune {
	if l.readPosition >= len(l.input) {
		return 0 // ASCII code for NUL
	}
	r, _ := utf8.DecodeRuneInString(l.input[l.readPosition:])
	return r
}

// NextToken scans the next token and returns it.
func (l *Lexer) NextToken() Token {
	// If there are pending tokens (INDENT/DEDENT), emit them first
	if len(l.pendingTokens) > 0 {
		tok := l.pendingTokens[0]
		l.pendingTokens = l.pendingTokens[1:]
		return tok
	}

	// Handle indentation at the start of a new line
	if l.AtNewLine {
		tok := l.handleIndentation()
		l.AtNewLine = false // Reset the flag after handling indentation
		if tok.Type != TokenNewline && tok.Type != TokenEOF {
			return tok
		}
	}

	l.skipWhitespace()

	var tok Token

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = Token{Type: TokenEQ, Literal: literal, Line: l.line, Column: l.column - 1}
		} else {
			tok = Token{Type: TokenAssign, Literal: string(l.ch), Line: l.line, Column: l.column}
		}
	case '+':
		tok = Token{Type: TokenPlus, Literal: string(l.ch), Line: l.line, Column: l.column}
	case '-':
		tok = Token{Type: TokenMinus, Literal: string(l.ch), Line: l.line, Column: l.column}
	case '*':
		tok = Token{Type: TokenAsterisk, Literal: string(l.ch), Line: l.line, Column: l.column}
	case '/':
		tok = Token{Type: TokenSlash, Literal: string(l.ch), Line: l.line, Column: l.column}
	case '%':
		tok = Token{Type: TokenModulo, Literal: string(l.ch), Line: l.line, Column: l.column}
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = Token{Type: TokenNotEQ, Literal: literal, Line: l.line, Column: l.column - 1}
		} else {
			tok = Token{Type: TokenBang, Literal: string(l.ch), Line: l.line, Column: l.column}
		}
	case '>':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = Token{Type: TokenGTE, Literal: literal, Line: l.line, Column: l.column - 1}
		} else {
			tok = Token{Type: TokenGT, Literal: string(l.ch), Line: l.line, Column: l.column}
		}
	case '<':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = Token{Type: TokenLTE, Literal: literal, Line: l.line, Column: l.column - 1}
		} else {
			tok = Token{Type: TokenLT, Literal: string(l.ch), Line: l.line, Column: l.column}
		}
	case '(':
		tok = Token{Type: TokenParenOpen, Literal: string(l.ch), Line: l.line, Column: l.column}
	case ')':
		tok = Token{Type: TokenParenClose, Literal: string(l.ch), Line: l.line, Column: l.column}
	case '[':
		tok = Token{Type: TokenBracketOpen, Literal: string(l.ch), Line: l.line, Column: l.column}
	case ']':
		tok = Token{Type: TokenBracketClose, Literal: string(l.ch), Line: l.line, Column: l.column}
	case '{':
		tok = Token{Type: TokenBraceOpen, Literal: string(l.ch), Line: l.line, Column: l.column}
	case '}':
		tok = Token{Type: TokenBraceClose, Literal: string(l.ch), Line: l.line, Column: l.column}
	case ',':
		tok = Token{Type: TokenComma, Literal: string(l.ch), Line: l.line, Column: l.column}
	case ';':
		tok = Token{Type: TokenSemicolon, Literal: string(l.ch), Line: l.line, Column: l.column}
	case ':':
		tok = Token{Type: TokenColon, Literal: string(l.ch), Line: l.line, Column: l.column}
	case '"', '\'', '`':
		//quoteChar := l.ch
		literal := l.readString(l.ch)
		//if quoteChar == '`' {
		//	literal = strings.Trim(literal, `\"`)
		//}
		tok = Token{Type: TokenString, Literal: literal, Line: l.line, Column: l.column - len(literal) - 1}
		return tok
	case '\n':
		tok = Token{Type: TokenNewline, Literal: "\\n", Line: l.line, Column: l.column}
		l.readChar()
		l.AtNewLine = true
		return tok
	case 0:
		// At EOF, emit DEDENT tokens for any remaining indentation levels
		if len(l.indentStack) > 1 {
			l.indentStack = l.indentStack[:len(l.indentStack)-1]
			return Token{Type: TokenDedent, Literal: "DEDENT", Line: l.line, Column: l.column}
		}
		tok = Token{Type: TokenEOF, Literal: "", Line: l.line, Column: l.column}
	case '.':
		tok = Token{Type: TokenDot, Literal: string(l.ch), Line: l.line, Column: l.column}
	default:
		if l.ch == '&' || isLetter(l.ch) {
			literal := l.readIdentifier()
			tokenType := LookupIdent(literal)
			tok = Token{Type: tokenType, Literal: literal, Line: l.line, Column: l.column - len(literal) + 1}
			return tok
		} else if isDigit(l.ch) {
			literal := l.readNumber()
			tok = Token{Type: TokenNumber, Literal: literal, Line: l.line, Column: l.column - len(literal) + 1}
			return tok
		} else {
			tok = Token{Type: TokenIllegal, Literal: string(l.ch), Line: l.line, Column: l.column}
		}
	}

	l.readChar()

	return tok
}

// PeekAhead looks ahead by n tokens and returns the token at that position without advancing the lexer's state.
func (l *Lexer) PeekAhead(n int) Token {
	// Save the lexer's state
	savedPosition := l.position
	savedReadPosition := l.readPosition
	savedCh := l.ch
	savedLine := l.line
	savedColumn := l.column
	savedIndentStack := make([]int, len(l.indentStack))
	copy(savedIndentStack, l.indentStack)
	savedPendingTokens := make([]Token, len(l.pendingTokens))
	copy(savedPendingTokens, l.pendingTokens)
	savedAtNewLine := l.AtNewLine

	var tok Token
	for i := 0; i <= n; i++ {
		tok = l.NextToken()
		if tok.Type == TokenEOF {
			break
		}
	}

	// Restore the lexer's state
	l.position = savedPosition
	l.readPosition = savedReadPosition
	l.ch = savedCh
	l.line = savedLine
	l.column = savedColumn
	l.indentStack = savedIndentStack
	l.pendingTokens = savedPendingTokens
	l.AtNewLine = savedAtNewLine

	return tok
}

// skipWhitespace skips over spaces and tabs.
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
		l.readChar()
	}
}

// readIdentifier reads an identifier and advances the lexer's positions.
func (l *Lexer) readIdentifier() string {
	position := l.readPosition - 1
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '&' || l.ch == '{' || l.ch == '}' {
		l.readChar()
	}
	return l.input[position : l.readPosition-1]
}

// readNumber reads a number (integer or float) and advances the lexer's positions.
func (l *Lexer) readNumber() string {
	position := l.readPosition - 1
	hasDot := false

	for {
		if isDigit(l.ch) {
			l.readChar()
		} else if l.ch == '.' && !hasDot {
			hasDot = true
			l.readChar()
		} else {
			break
		}
	}

	return l.input[position : l.readPosition-1]
}

// readString reads a string literal, handling escape sequences and multi-line strings.
func (l *Lexer) readString(quoteChar rune) string {
	var sb strings.Builder
	l.readChar() // Skip the opening quote

	isTripleQuoted := false

	// Helper function to peek ahead n runes without advancing the lexer's position
	peekAhead := func(n int) rune {
		position := l.readPosition
		var ch rune
		for i := 0; i < n; i++ {
			if position >= len(l.input) {
				return 0 // NUL
			}
			r, width := utf8.DecodeRuneInString(l.input[position:])
			ch = r
			position += width
		}
		return ch
	}

	// Check if it's a triple-quoted string
	if l.ch == quoteChar && l.peekChar() == quoteChar {
		isTripleQuoted = true
		// Consume the next two quoteChars
		l.readChar() // Skip second quoteChar
		l.readChar() // Skip third quoteChar
	}

	if quoteChar == '`' {
		// Backtick strings are raw strings; read until the closing backtick
		for {
			if l.ch == '`' {
				//fmt.Println(string(l.peekChar()))
				l.readChar() // Consume closing backtick
				break
			}
			if l.ch == 0 {
				// Reached EOF without closing backtick
				break
			}
			sb.WriteRune(l.ch)
			l.readChar()
		}
	} else if isTripleQuoted {
		// Triple-quoted string
		for {
			if l.ch == quoteChar && l.peekChar() == quoteChar && peekAhead(1) == quoteChar {
				// Consume the three quoteChars
				l.readChar() // Skip first quoteChar
				l.readChar() // Skip second quoteChar
				l.readChar() // Skip third quoteChar
				break
			}
			if l.ch == 0 {
				// Reached EOF without closing triple quote
				break
			}
			if l.ch == '\\' {
				l.readChar()
				switch l.ch {
				case 'n':
					sb.WriteRune('\n')
				case 't':
					sb.WriteRune('\t')
				case 'r':
					sb.WriteRune('\r')
				case '\\':
					sb.WriteRune('\\')
				case '\'':
					sb.WriteRune('\'')
				case '"':
					sb.WriteRune('"')
				default:
					// Unknown escape sequence, include the backslash and the character
					sb.WriteRune('\\')
					sb.WriteRune(l.ch)
				}
				l.readChar()
			} else {
				sb.WriteRune(l.ch)
				l.readChar()
			}
		}
	} else {
		// Single-quoted or double-quoted string
		for {
			if l.ch == quoteChar {
				l.readChar() // Consume closing quote
				break
			}
			if l.ch == '\\' {
				l.readChar()
				switch l.ch {
				case 'n':
					sb.WriteRune('\n')
				case 't':
					sb.WriteRune('\t')
				case 'r':
					sb.WriteRune('\r')
				case '\\':
					sb.WriteRune('\\')
				case '\'':
					sb.WriteRune('\'')
				case '"':
					sb.WriteRune('"')
				default:
					// Unknown escape sequence, include the backslash and the character
					sb.WriteRune('\\')
					sb.WriteRune(l.ch)
				}
				l.readChar()
			} else if l.ch == 0 || l.ch == '\n' {
				// Reached EOF or newline without closing quote
				break
			} else {
				sb.WriteRune(l.ch)
				l.readChar()
			}
		}
	}

	return sb.String()
}

// handleIndentation handles indentation at the start of a new line.
func (l *Lexer) handleIndentation() Token {
	spaces := 0
	for {
		if l.ch == ' ' {
			spaces++
			l.readChar()
		} else if l.ch == '\t' {
			spaces += 4 // convert tab to 4 spaces
			l.readChar()
		} else {
			break
		}
	}

	currentIndent := l.indentStack[len(l.indentStack)-1]

	if spaces > currentIndent {
		l.indentStack = append(l.indentStack, spaces)
		return Token{Type: TokenIndent, Literal: "INDENT", Line: l.line, Column: l.column}
	} else if spaces < currentIndent && l.ch != '\n' {
		var dedentTokens []Token
		for len(l.indentStack) > 1 && spaces < l.indentStack[len(l.indentStack)-1] {
			l.indentStack = l.indentStack[:len(l.indentStack)-1]
			dedentTokens = append(dedentTokens, Token{Type: TokenDedent, Literal: "DEDENT", Line: l.line, Column: l.column})
		}
		if len(l.indentStack) == 0 || spaces != l.indentStack[len(l.indentStack)-1] {
			return Token{Type: TokenIllegal, Literal: fmt.Sprintf("Invalid dedent at line %d", l.line), Line: l.line, Column: l.column}
		}
		l.pendingTokens = append(dedentTokens, l.pendingTokens...)
		tok := l.pendingTokens[0]
		l.pendingTokens = l.pendingTokens[1:]
		return tok
	}

	// Indentation level is the same
	return Token{Type: TokenNewline, Literal: "\\n", Line: l.line, Column: l.column}
}

// isLetter checks if the rune is a letter or underscore.
func isLetter(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_'
}

// isDigit checks if the rune is a digit.
func isDigit(ch rune) bool {
	return unicode.IsDigit(ch)
}
