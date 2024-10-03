package parser

import (
	"fmt"
	"strconv"
	"strings"

	"simple/lexer"
)

// Type represents a type in the Simple language.
type Type interface {
	TypeName() string
	String() string
}

// BasicType represents basic built-in types like int, string, etc.
type BasicType struct {
	Name string
}

func (bt *BasicType) TypeName() string {
	return bt.Name
}

func (bt *BasicType) String() string {
	return bt.Name
}

// PointerType represents a pointer to another type.
type PointerType struct {
	ElementType Type
}

func (pt *PointerType) TypeName() string {
	return "*" + pt.ElementType.TypeName()
}

func (pt *PointerType) String() string {
	return "*" + pt.ElementType.String()
}

// NamedType represents a named type, possibly from an imported package.
type NamedType struct {
	Name    string // e.g., "ResponseWriter"
	Package string // e.g., "http"
}

func (nt *NamedType) TypeName() string {
	return nt.Name
}

func (nt *NamedType) String() string {
	if nt.Package != "" {
		return nt.Package + "." + nt.Name
	}
	return nt.Name
}

// ArrayType represents an array type with element type.
type ArrayType struct {
	ElementType Type
}

func (at *ArrayType) TypeName() string {
	return "[]" + at.ElementType.TypeName()
}

func (at *ArrayType) String() string {
	return fmt.Sprintf("[]%s", at.ElementType.String())
}

// ArrayLiteral represents an array/list literal in the code.
type ArrayLiteral struct {
	Token    lexer.Token // The '[' token
	Elements []Expression
	Type     Type // Inferred element type
}

func (al *ArrayLiteral) expressionNode()      {}
func (al *ArrayLiteral) TokenLiteral() string { return al.Token.Literal }
func (al *ArrayLiteral) String() string       { return al.Token.Literal }

// MapType represents a map/dictionary type with key and value types.
type MapType struct {
	KeyType   Type
	ValueType Type
}

func (mt *MapType) TypeName() string {
	return "map[" + mt.KeyType.TypeName() + "]" + mt.ValueType.TypeName()
}

func (mt *MapType) String() string {
	return fmt.Sprintf("map[%s]%s", mt.KeyType.String(), mt.ValueType.String())
}

// MapLiteral represents a map/dictionary literal in the code.
type MapLiteral struct {
	Token lexer.Token // The '{' token
	Pairs map[Expression]Expression
	Type  Type // Inferred key and value types
}

func (ml *MapLiteral) expressionNode()      {}
func (ml *MapLiteral) TokenLiteral() string { return ml.Token.Literal }
func (ml *MapLiteral) String() string       { return ml.Token.Literal }

// BuiltinType represents a built-in function type.
type BuiltinType struct {
	Name string
}

func (bt *BuiltinType) TypeName() string {
	return "builtin"
}

func (bt *BuiltinType) String() string {
	return bt.Name
}

// FunctionType represents function types.
type FunctionType struct {
	Parameters     []Identifier
	ParameterTypes []Type
	ReturnType     Type
}

func (ft *FunctionType) TypeName() string {
	return "function"
}

func (ft *FunctionType) String() string {
	params := []string{}
	for _, p := range ft.ParameterTypes {
		params = append(params, p.String())
	}
	if ft.ReturnType != nil {
		return fmt.Sprintf("func(%s) %s", strings.Join(params, ", "), ft.ReturnType.String())
	}
	return fmt.Sprintf("func(%s) %s", strings.Join(params, ", "), nil)

}

// In parser/ast.go

// TypeConversionExpression represents a type conversion.
type TypeConversionExpression struct {
	Token      lexer.Token
	Expression Expression
	TargetType Type
}

func (tce *TypeConversionExpression) expressionNode()      {}
func (tce *TypeConversionExpression) TokenLiteral() string { return tce.Token.Literal }
func (tce *TypeConversionExpression) String() string {
	return fmt.Sprintf("%s(%s)", tce.TargetType.String(), tce.Expression.String())
}

// StructType represents a struct type.
type StructType struct {
	Name string
}

func (st *StructType) TypeName() string {
	return "struct"
}

func (st *StructType) String() string {
	return st.Name
}

// InterfaceType represents an interface type.
type InterfaceType struct {
	Name string
}

func (it *InterfaceType) TypeName() string {
	return "interface"
}

func (it *InterfaceType) String() string {
	return it.Name
}

// Node represents a node in the AST.
type Node interface {
	TokenLiteral() string
	String() string
}

// Statement represents a statement in the AST.
type Statement interface {
	Node
	statementNode()
}

// Expression represents an expression in the AST.
type Expression interface {
	Node
	expressionNode()
}

// Program is the root node of the AST.
type Program struct {
	Statements []Statement
}

func (p *Program) statementNode()       {}
func (p *Program) TokenLiteral() string { return "" }
func (p *Program) String() string {
	var out strings.Builder
	for _, stmt := range p.Statements {
		out.WriteString(stmt.String())
		out.WriteString("\n")
	}
	return out.String()
}

// Identifier represents an identifier.
type Identifier struct {
	Token lexer.Token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

// IntegerLiteral represents an integer.
type IntegerLiteral struct {
	Token lexer.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

// BooleanLiteral represents a boolean value.
type BooleanLiteral struct {
	Token lexer.Token
	Value bool
}

func (bl *BooleanLiteral) expressionNode()      {}
func (bl *BooleanLiteral) TokenLiteral() string { return bl.Token.Literal }
func (bl *BooleanLiteral) String() string       { return bl.Token.Literal }

// StringLiteral represents a string literal.
type StringLiteral struct {
	Token lexer.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return "\"" + sl.Value + "\"" }

// FunctionLiteral represents a function definition.
type FunctionLiteral struct {
	Token      lexer.Token
	Name       *Identifier
	Parameters []*Identifier
	Body       *BlockStatement
}

func (fl *FunctionLiteral) expressionNode()      {}
func (fl *FunctionLiteral) statementNode()       {}
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FunctionLiteral) String() string {
	var out strings.Builder
	params := []string{}
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}
	out.WriteString("def ")
	out.WriteString(fl.Name.String())
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(")")
	out.WriteString(":\n")
	out.WriteString(fl.Body.String())
	return out.String()
}

// BlockStatement represents a block of statements.
type BlockStatement struct {
	Token      lexer.Token
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string {
	var out strings.Builder
	for _, stmt := range bs.Statements {
		out.WriteString(stmt.String())
		out.WriteString("\n")
	}
	return out.String()
}

// ExpressionStatement represents a statement consisting of a single expression.
type ExpressionStatement struct {
	Token      lexer.Token
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	return es.Expression.String()
}

// CallExpression represents a function call.
type CallExpression struct {
	Token     lexer.Token
	Function  Expression
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out strings.Builder
	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}
	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")
	return out.String()
}

// ReturnStatement represents a return statement.
type ReturnStatement struct {
	Token       lexer.Token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) String() string {
	var out strings.Builder
	out.WriteString("return ")
	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}
	return out.String()
}

// IfStatement represents an if statement.
type IfStatement struct {
	Token       lexer.Token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (is *IfStatement) statementNode()       {}
func (is *IfStatement) TokenLiteral() string { return is.Token.Literal }
func (is *IfStatement) String() string {
	var out strings.Builder
	out.WriteString("if ")
	out.WriteString(is.Condition.String())
	out.WriteString(":\n")
	out.WriteString(is.Consequence.String())
	if is.Alternative != nil {
		out.WriteString("else:\n")
		out.WriteString(is.Alternative.String())
	}
	return out.String()
}

// WhileStatement represents a while loop.
type WhileStatement struct {
	Token     lexer.Token
	Condition Expression
	Body      *BlockStatement
}

func (ws *WhileStatement) statementNode()       {}
func (ws *WhileStatement) TokenLiteral() string { return ws.Token.Literal }
func (ws *WhileStatement) String() string {
	var out strings.Builder
	out.WriteString("while ")
	out.WriteString(ws.Condition.String())
	out.WriteString(":\n")
	out.WriteString(ws.Body.String())
	return out.String()
}

// ForStatement represents a for loop.
type ForStatement struct {
	Token    lexer.Token
	Variable *Identifier
	Iterable Expression
	Body     *BlockStatement
}

func (fs *ForStatement) statementNode()       {}
func (fs *ForStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *ForStatement) String() string {
	var out strings.Builder
	out.WriteString("for ")
	out.WriteString(fs.Variable.String())
	out.WriteString(" in ")
	out.WriteString(fs.Iterable.String())
	out.WriteString(":\n")
	out.WriteString(fs.Body.String())
	return out.String()
}

// AssignmentStatement represents a variable assignment.
type AssignmentStatement struct {
	Token lexer.Token
	Name  *Identifier
	Value Expression
}

func (as *AssignmentStatement) statementNode()       {}
func (as *AssignmentStatement) TokenLiteral() string { return as.Token.Literal }
func (as *AssignmentStatement) String() string {
	var out strings.Builder
	out.WriteString(as.Name.String())
	out.WriteString(" = ")
	if as.Value != nil {
		out.WriteString(as.Value.String())
	}
	return out.String()
}

// ImportStatement represents an import statement.
type ImportStatement struct {
	Token          lexer.Token
	ImportedModule *StringLiteral
}

func (is *ImportStatement) statementNode()       {}
func (is *ImportStatement) TokenLiteral() string { return is.Token.Literal }
func (is *ImportStatement) String() string {
	var out strings.Builder
	out.WriteString("import ")
	out.WriteString(is.ImportedModule.String())
	return out.String()
}

// InfixExpression represents an infix expression.
type InfixExpression struct {
	Token    lexer.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	var out strings.Builder
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" ")
	out.WriteString(ie.Operator)
	out.WriteString(" ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")
	return out.String()
}

// PrefixExpression represents a prefix expression.
type PrefixExpression struct {
	Token    lexer.Token
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	var out strings.Builder
	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")
	return out.String()
}

// SelectorExpression represents an expression like "w.Write"
type SelectorExpression struct {
	Token    lexer.Token // The '.' token
	Left     Expression
	Selector *Identifier
}

func (se *SelectorExpression) expressionNode()      {}
func (se *SelectorExpression) TokenLiteral() string { return se.Token.Literal }
func (se *SelectorExpression) String() string {
	var out strings.Builder
	out.WriteString(se.Left.String())
	out.WriteString(".")
	out.WriteString(se.Selector.String())
	return out.String()
}

// Precedence levels.
const (
	_ int = iota
	LOWEST
	EQUALS      // == or !=
	LESSGREATER // >, <, >=, <=
	SUM         // + or -
	PRODUCT     // *, /, %
	PREFIX      // -X or !X
	CALL        // function calls
	SELECTOR    = iota + 1
)

// precedences maps token types to their precedence.
var precedences = map[lexer.TokenType]int{
	lexer.TokenEQ:        EQUALS,
	lexer.TokenNotEQ:     EQUALS,
	lexer.TokenLT:        LESSGREATER,
	lexer.TokenLTE:       LESSGREATER,
	lexer.TokenGT:        LESSGREATER,
	lexer.TokenGTE:       LESSGREATER,
	lexer.TokenPlus:      SUM,
	lexer.TokenMinus:     SUM,
	lexer.TokenAsterisk:  PRODUCT,
	lexer.TokenSlash:     PRODUCT,
	lexer.TokenModulo:    PRODUCT,
	lexer.TokenParenOpen: CALL, // For function calls
	lexer.TokenDot:       SELECTOR,
}

// Parser represents a parser.
type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  lexer.Token
	peekToken lexer.Token

	prefixParseFns map[lexer.TokenType]prefixParseFn
	infixParseFns  map[lexer.TokenType]infixParseFn
}

type (
	prefixParseFn func() Expression
	infixParseFn  func(Expression) Expression
)

// NewParser creates a new parser.
func NewParser(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:              l,
		errors:         []string{},
		prefixParseFns: make(map[lexer.TokenType]prefixParseFn),
		infixParseFns:  make(map[lexer.TokenType]infixParseFn),
	}

	// Register prefix parsers.
	p.registerPrefix(lexer.TokenIdentifier, p.parseIdentifier)
	p.registerPrefix(lexer.TokenNumber, p.parseIntegerLiteral)
	p.registerPrefix(lexer.TokenString, p.parseStringLiteral)
	p.registerPrefix(lexer.TokenBang, p.parsePrefixExpression)
	p.registerPrefix(lexer.TokenMinus, p.parsePrefixExpression)
	p.registerPrefix(lexer.TokenParenOpen, p.parseGroupedExpression)
	p.registerPrefix(lexer.TokenTrue, p.parseBooleanLiteral)
	p.registerPrefix(lexer.TokenFalse, p.parseBooleanLiteral)

	// Register infix parsers.
	p.registerInfix(lexer.TokenPlus, p.parseInfixExpression)
	p.registerInfix(lexer.TokenMinus, p.parseInfixExpression)
	p.registerInfix(lexer.TokenAsterisk, p.parseInfixExpression)
	p.registerInfix(lexer.TokenSlash, p.parseInfixExpression)
	p.registerInfix(lexer.TokenModulo, p.parseInfixExpression)
	p.registerInfix(lexer.TokenEQ, p.parseInfixExpression)
	p.registerInfix(lexer.TokenNotEQ, p.parseInfixExpression)
	p.registerInfix(lexer.TokenLT, p.parseInfixExpression)
	p.registerInfix(lexer.TokenLTE, p.parseInfixExpression)
	p.registerInfix(lexer.TokenGT, p.parseInfixExpression)
	p.registerInfix(lexer.TokenGTE, p.parseInfixExpression)
	p.registerInfix(lexer.TokenParenOpen, p.parseCallExpression)
	p.registerInfix(lexer.TokenDot, p.parseSelectorExpression)

	// Read two tokens to initialize curToken and peekToken.
	p.nextToken()
	p.nextToken()

	return p
}

// Errors returns parser errors.
func (p *Parser) Errors() []string {
	return p.errors
}

// registerPrefix registers a prefix parse function for a given token type.
func (p *Parser) registerPrefix(tokenType lexer.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

// registerInfix registers an infix parse function for a given token type.
func (p *Parser) registerInfix(tokenType lexer.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

// nextToken advances to the next token.
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// expectPeek checks if the next token is of the expected type and advances if it is.
func (p *Parser) expectPeek(t lexer.TokenType) bool {
	if p.peekToken.Type == t {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

// peekError records an error for unexpected peek token.
func (p *Parser) peekError(t lexer.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead (Line %d, Column %d)", t, p.peekToken.Type, p.peekToken.Line, p.peekToken.Column)
	p.errors = append(p.errors, msg)
}

// noPrefixParseFnError records an error for missing prefix parse function.
func (p *Parser) noPrefixParseFnError(t lexer.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found (Line %d, Column %d)", t, p.curToken.Line, p.curToken.Column)
	p.errors = append(p.errors, msg)
}

// peekPrecedence returns the precedence of the peek token.
func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

// curPrecedence returns the precedence of the current token.
func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

// ParseProgram parses the entire program and returns the AST.
func (p *Parser) ParseProgram() *Program {
	program := &Program{}
	program.Statements = []Statement{}

	for p.curToken.Type != lexer.TokenEOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

// parseStatement parses a single statement.
func (p *Parser) parseStatement() Statement {
	p.skipNewlines()

	// Check if we've reached EOF
	if p.curToken.Type == lexer.TokenEOF {
		return nil
	}

	switch p.curToken.Type {
	case lexer.TokenKeyword:
		switch p.curToken.Literal {
		case "def":
			return p.parseFunctionDefinition()
		case "return":
			return p.parseReturnStatement()
		case "if":
			return p.parseIfStatement()
		case "while":
			return p.parseWhileStatement()
		case "for":
			return p.parseForStatement()
		case "import":
			return p.parseImportStatement()
		default:
			return nil
		}
	case lexer.TokenIdentifier:
		if p.peekToken.Type == lexer.TokenAssign {
			return p.parseAssignmentStatement()
		} else {
			return p.parseExpressionStatement()
		}
	default:
		return p.parseExpressionStatement()
	}
}

// In parser/parser.go

func (p *Parser) parseSelectorExpression(left Expression) Expression {
	se := &SelectorExpression{
		Token: p.curToken, // The '.' token
		Left:  left,
	}

	if !p.expectPeek(lexer.TokenIdentifier) {
		return nil
	}

	se.Selector = &Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}

	return se
}

// parseAssignmentStatement parses a variable assignment.
func (p *Parser) parseAssignmentStatement() *AssignmentStatement {
	stmt := &AssignmentStatement{Token: p.curToken}

	// Parse the identifier on the left-hand side
	stmt.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Consume the '=' token
	if !p.expectPeek(lexer.TokenAssign) {
		return nil
	}

	// Consume the expression on the right-hand side
	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	// Optional: handle end of statement (e.g., semicolons)
	if p.peekToken.Type == lexer.TokenNewline {
		p.nextToken()
	}

	return stmt
}

// parseFunctionDefinition parses a function definition.
func (p *Parser) parseFunctionDefinition() Statement {
	fl := &FunctionLiteral{
		Token: p.curToken,
	}

	if !p.expectPeek(lexer.TokenIdentifier) {
		return nil
	}

	fl.Name = &Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}

	if !p.expectPeek(lexer.TokenParenOpen) {
		return nil
	}

	fl.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(lexer.TokenColon) {
		return nil
	}

	fl.Body = p.parseBlockStatement()

	return fl
}

// parseFunctionParameters parses function parameters.
func (p *Parser) parseFunctionParameters() []*Identifier {
	identifiers := []*Identifier{}

	if p.peekToken.Type == lexer.TokenParenClose {
		p.nextToken()
		return identifiers
	}

	p.nextToken()
	ident := &Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
	identifiers = append(identifiers, ident)

	for p.peekToken.Type == lexer.TokenComma {
		p.nextToken()
		p.nextToken()
		ident := &Identifier{
			Token: p.curToken,
			Value: p.curToken.Literal,
		}
		identifiers = append(identifiers, ident)
	}

	if !p.expectPeek(lexer.TokenParenClose) {
		return nil
	}

	return identifiers
}

// parseBlockStatement parses a block of statements.
func (p *Parser) parseBlockStatement() *BlockStatement {
	block := &BlockStatement{
		Token:      p.curToken,
		Statements: []Statement{},
	}

	if !p.expectPeek(lexer.TokenNewline) {
		return nil
	}

	p.skipNewlines()

	// Expect INDENT
	if p.peekToken.Type != lexer.TokenIndent {
		msg := fmt.Sprintf("expected INDENT, got %s instead (Line %d, Column %d)", p.peekToken.Type, p.peekToken.Line, p.peekToken.Column)
		p.errors = append(p.errors, msg)
		return nil
	}

	p.nextToken() // Move to INDENT
	p.nextToken() // Move to the first token inside the block

	for p.curToken.Type != lexer.TokenDedent && p.curToken.Type != lexer.TokenEOF {
		stmt := p.parseStatement()
		p.nextToken()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		} else {
			p.nextToken()
		}
	}

	// After the block, p.curToken should be DEDENT
	if p.curToken.Type != lexer.TokenDedent {
		msg := fmt.Sprintf("expected DEDENT, got %s instead (Line %d, Column %d)", p.curToken.Type, p.curToken.Line, p.curToken.Column)
		p.errors = append(p.errors, msg)
		return nil
	}

	return block
}

// parseReturnStatement parses a return statement.
func (p *Parser) parseReturnStatement() *ReturnStatement {
	rs := &ReturnStatement{
		Token: p.curToken,
	}

	p.nextToken()

	rs.ReturnValue = p.parseExpression(LOWEST)

	// Optional: handle end of statement
	if p.peekToken.Type == lexer.TokenNewline {
		p.nextToken()
	}

	return rs
}

// parseIfStatement parses an if statement.
func (p *Parser) parseIfStatement() *IfStatement {
	is := &IfStatement{
		Token: p.curToken,
	}

	p.nextToken()
	is.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.TokenColon) {
		return nil
	}

	is.Consequence = p.parseBlockStatement()

	if p.peekToken.Type == lexer.TokenKeyword && p.peekToken.Literal == "else" {
		p.nextToken() // Move to 'else'
		//p.nextToken() // Move to ':'

		if !p.expectPeek(lexer.TokenColon) {
			return nil
		}

		is.Alternative = p.parseBlockStatement()
	}

	return is
}

// parseWhileStatement parses a while loop.
func (p *Parser) parseWhileStatement() *WhileStatement {
	ws := &WhileStatement{
		Token: p.curToken,
	}

	p.nextToken()
	ws.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.TokenColon) {
		return nil
	}

	ws.Body = p.parseBlockStatement()

	return ws
}

// parseForStatement parses a for loop.
func (p *Parser) parseForStatement() *ForStatement {
	fs := &ForStatement{
		Token: p.curToken,
	}

	if !p.expectPeek(lexer.TokenIdentifier) {
		return nil
	}

	fs.Variable = &Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}

	if !p.expectPeek(lexer.TokenKeyword) || p.curToken.Literal != "in" {
		msg := fmt.Sprintf("expected 'in', got %s instead (Line %d, Column %d)", p.curToken.Literal, p.curToken.Line, p.curToken.Column)
		p.errors = append(p.errors, msg)
		return nil
	}

	p.nextToken()
	fs.Iterable = p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.TokenColon) {
		return nil
	}

	fs.Body = p.parseBlockStatement()

	return fs
}

// parseImportStatement parses an import statement.
func (p *Parser) parseImportStatement() *ImportStatement {
	is := &ImportStatement{
		Token: p.curToken,
	}

	if !p.expectPeek(lexer.TokenString) {
		return nil
	}

	is.ImportedModule = &StringLiteral{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}

	// Optional: handle end of statement
	if p.peekToken.Type == lexer.TokenNewline {
		p.nextToken()
	}

	return is
}

// parseExpressionStatement parses an expression statement.
func (p *Parser) parseExpressionStatement() *ExpressionStatement {
	if p.curToken.Type == lexer.TokenNewline {
		return nil
	}

	es := &ExpressionStatement{
		Token: p.curToken,
	}

	es.Expression = p.parseExpression(LOWEST)

	// Optional: handle end of statement
	if p.peekToken.Type == lexer.TokenNewline {
		p.nextToken()
	}

	return es
}

// parseExpression parses an expression with given precedence.
func (p *Parser) parseExpression(precedence int) Expression {
	//p.skipNewlines()

	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for p.peekToken.Type != lexer.TokenNewline && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()

		leftExp = infix(leftExp)
	}

	return leftExp
}

// parseIdentifier parses an identifier.
func (p *Parser) parseIdentifier() Expression {
	return &Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
}

// parseIntegerLiteral parses an integer literal.
func (p *Parser) parseIntegerLiteral() Expression {
	il := &IntegerLiteral{
		Token: p.curToken,
	}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer (Line %d, Column %d)", p.curToken.Literal, p.curToken.Line, p.curToken.Column)
		p.errors = append(p.errors, msg)
		return nil
	}
	il.Value = value
	return il
}

// parseStringLiteral parses a string literal.
func (p *Parser) parseStringLiteral() Expression {
	return &StringLiteral{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
}

// parseBooleanLiteral parses a boolean literal.
func (p *Parser) parseBooleanLiteral() Expression {
	value := false
	if p.curToken.Type == lexer.TokenTrue {
		value = true
	}
	return &BooleanLiteral{
		Token: p.curToken,
		Value: value,
	}
}

// parseGroupedExpression parses a grouped expression.
func (p *Parser) parseGroupedExpression() Expression {
	p.nextToken()
	exp := p.parseExpression(LOWEST)
	if !p.expectPeek(lexer.TokenParenClose) {
		return nil
	}
	return exp
}

// parsePrefixExpression parses a prefix expression.
func (p *Parser) parsePrefixExpression() Expression {
	pe := &PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	pe.Right = p.parseExpression(PREFIX)

	return pe
}

// parseInfixExpression parses an infix expression.
func (p *Parser) parseInfixExpression(left Expression) Expression {
	ie := &InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	ie.Right = p.parseExpression(precedence)

	return ie
}

// parseCallExpression parses a function call expression.
func (p *Parser) parseCallExpression(function Expression) Expression {
	ce := &CallExpression{
		Token:    p.curToken,
		Function: function,
	}

	ce.Arguments = p.parseExpressionList(lexer.TokenParenClose)

	return ce
}

// parseExpressionList parses a list of expressions separated by commas.
func (p *Parser) parseExpressionList(end lexer.TokenType) []Expression {
	list := []Expression{}

	if p.peekToken.Type == end {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))

	for p.peekToken.Type == lexer.TokenComma {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}

// skipNewlines skips over newline tokens.
func (p *Parser) skipNewlines() {
	for p.peekToken.Type == lexer.TokenNewline {
		p.nextToken()
	}
}

// NodeVisitor is a function that can be called for each node.
type NodeVisitor func(Node) bool

// Inspect traverses the AST and calls the visitor function for each node.
func Inspect(node Node, pre NodeVisitor) {
	if !pre(node) {
		return
	}

	switch n := node.(type) {
	case *Program:
		for _, stmt := range n.Statements {
			Inspect(stmt, pre)
		}
	case *ExpressionStatement:
		if n != nil {
			Inspect(n.Expression, pre)
		}

	case *CallExpression:
		Inspect(n.Function, pre)
		for _, arg := range n.Arguments {
			Inspect(arg, pre)
		}
	case *FunctionLiteral:
		for _, param := range n.Parameters {
			Inspect(param, pre)
		}
		Inspect(n.Body, pre)
	case *BlockStatement:
		for _, stmt := range n.Statements {
			Inspect(stmt, pre)
		}
	case *IfStatement:
		Inspect(n.Condition, pre)
		Inspect(n.Consequence, pre)
		if n.Alternative != nil {
			Inspect(n.Alternative, pre)
		}
	case *WhileStatement:
		Inspect(n.Condition, pre)
		Inspect(n.Body, pre)
	case *ForStatement:
		Inspect(n.Iterable, pre)
		Inspect(n.Body, pre)
	case *InfixExpression:
		Inspect(n.Left, pre)
		Inspect(n.Right, pre)
	case *PrefixExpression:
		Inspect(n.Right, pre)
	case *SelectorExpression:
		Inspect(n.Left, pre)
		Inspect(n.Selector, pre)
		// Add other node types as needed
	}
}
