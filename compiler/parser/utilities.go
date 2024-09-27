package parser

import (
	"fmt"
	"simple/lexer"
	"strconv"
	"strings"
)

// ########################################################
// ########################################################
// ##############        UTILITIES         ################
// ########################################################
// ########################################################

// lookaheadToken allows looking ahead at the next token without consuming it
func (p *Parser) lookaheadToken(n int) lexer.Token {
	if p.current+n >= len(p.tokens) {
		return lexer.Token{Type: lexer.TokenEOF}
	}
	return p.tokens[p.current+n]
}

// Helper function to check the current token type
func (p *Parser) currentToken() lexer.Token {
	if p.current >= len(p.tokens) {
		return lexer.Token{Type: lexer.TokenEOF}
	}
	return p.tokens[p.current]
}

// Helper function to advance to the next token
func (p *Parser) nextToken() lexer.Token {
	p.current++
	return p.currentToken()
}

// Operator precedence map
var precedenceMap = map[string]int{
	"||": 1,                                            // Logical OR
	"&&": 2,                                            // Logical AND
	"==": 3, "!=": 3, "<": 3, ">": 3, "<=": 3, ">=": 3, // Comparison
	"+": 4, "-": 4, // Additive
	"*": 5, "/": 5, // Multiplicative
}

// Get operator precedence
func getPrecedence(op string) int {
	if prec, ok := precedenceMap[op]; ok {
		return prec
	}
	return -1 // Lowest precedence for unknown operators
}

// Helper function to determine the type of a literal (number, string, etc.)
func determineLiteralType(literal *LiteralNode) string {
	if _, err := strconv.Atoi(literal.Value); err == nil {
		return "int"
	}
	if _, err := strconv.ParseFloat(literal.Value, 64); err == nil {
		return "float"
	}
	if literal.Value[0] == '"' || literal.Value[0] == '\'' {
		return "string"
	}
	if variableType, ok := symbolTable[literal.Value]; ok {
		return variableType
	}
	return "unknown"
}

// ########################################################
// ########################################################
// ##############    NODE IMPLEMENTATION   ################
// ########################################################
// ########################################################

// ASTNode represents a node in the abstract syntax tree
type ASTNode interface {
	String() string         // For debugging, prints the node's structure
	GenerateGoCode() string // Generates equivalent Go code
}

// ########################################################
// ########################################################
// ##############       LITERAL NODE       ################
// ########################################################
// ########################################################

// LiteralNode Represents a literal value like a number or string
type LiteralNode struct {
	Value string
}

func (n *LiteralNode) String() string {
	return fmt.Sprintf("Literal: %s", n.Value)
}

func (n *LiteralNode) GenerateGoCode() string {
	// If the literal is a string, ensure it's formatted correctly for Go
	if len(n.Value) > 0 && (n.Value[0] == '\'' || n.Value[0] == '"') {
		// Convert single-quoted strings to double-quoted strings for Go compatibility
		if n.Value[0] == '\'' {
			n.Value = "\"" + n.Value[1:len(n.Value)-1] + "\""
		}
	}

	// Handle booleans: Convert 'True' -> 'true' and 'False' -> 'false'
	if n.Value == "True" {
		return "true"
	} else if n.Value == "False" {
		return "false"
	}

	// Handle None -> nil
	if n.Value == "nil" {
		return "nil"
	}

	// For numbers and identifiers, return them as-is
	return n.Value
}

// ########################################################
// ########################################################
// ##############   UNARY OPERATION NODE   ################
// ########################################################
// ########################################################

// UnaryOperationNode Represents a unary operation like -a or +b
type UnaryOperationNode struct {
	Operator string
	Operand  ASTNode
}

func (n *UnaryOperationNode) String() string {
	return fmt.Sprintf("UnaryOperation: %s%v", n.Operator, n.Operand)
}

func (n *UnaryOperationNode) GenerateGoCode() string {
	return fmt.Sprintf("%s%s", n.Operator, n.Operand.GenerateGoCode())
}

// ########################################################
// ########################################################
// ##############       BINARY NODE        ################
// ########################################################
// ########################################################

// BinaryOperationNode Represents a binary operation like `a == 5`
type BinaryOperationNode struct {
	Left     ASTNode
	Operator string
	Right    ASTNode
}

func (n *BinaryOperationNode) String() string {
	return fmt.Sprintf("BinaryOperation: %v %s %v", n.Left, n.Operator, n.Right)
}

func (n *BinaryOperationNode) GenerateGoCode() string {
	return fmt.Sprintf("%s %s %s", n.Left.GenerateGoCode(), n.Operator, n.Right.GenerateGoCode())
}

// ########################################################
// ########################################################
// ##############      NEWLINE NODE       ##################
// ########################################################
// ########################################################

// NewlineNode Represents a newline in the code (used for formatting in code generation)
type NewlineNode struct{}

func (n *NewlineNode) String() string {
	return "Newline"
}

func (n *NewlineNode) GenerateGoCode() string {
	return "\n" // Generate a newline in the Go code
}

// ########################################################
// ########################################################
// ##############     ASSIGNMENT NODE      ################
// ########################################################
// ########################################################

// AssignmentNode Represents an assignment like `a = 5`
type AssignmentNode struct {
	Variable string
	Value    ASTNode
}

func (n *AssignmentNode) String() string {
	return fmt.Sprintf("Assignment: %s = %v", n.Variable, n.Value)
}

// GenerateGoCode Generate Go code for assignments, handling redeclaration vs reassignment
func (n *AssignmentNode) GenerateGoCode() string {
	// Check if the variable has already been declared
	if declaredVariables[n.Variable] {
		// Use '=' for reassignment
		return fmt.Sprintf("%s = %s", n.Variable, n.Value.GenerateGoCode())
	} else {
		// Mark the variable as declared and use ':=' for the first declaration
		declaredVariables[n.Variable] = true
		return fmt.Sprintf("%s := %s", n.Variable, n.Value.GenerateGoCode())
	}
}

// ########################################################
// ########################################################
// ##############    FUNCTION CALL NODE    ################
// ########################################################
// ########################################################

// FunctionCallNode Represents a function call like foo(a, b)
type FunctionCallNode struct {
	FunctionName string
	Arguments    []ASTNode
}

func (n *FunctionCallNode) String() string {
	return fmt.Sprintf("FunctionCall: %s(%v)", n.FunctionName, n.Arguments)
}

func (n *FunctionCallNode) GenerateGoCode() string {
	var argCodes []string
	for _, arg := range n.Arguments {
		argCodes = append(argCodes, arg.GenerateGoCode())
	}
	return fmt.Sprintf("%s(%s)", n.FunctionName, strings.Join(argCodes, ", "))
}

// ########################################################
// ########################################################
// ##############        BLOCK NODE        ################
// ########################################################
// ########################################################

// BlockNode Represents a block of statements (e.g., a function body, if-else body, while loop body)
type BlockNode struct {
	Statements []ASTNode
}

func (n *BlockNode) String() string {
	result := "Block: [\n"
	for _, stmt := range n.Statements {
		result += fmt.Sprintf("  %v\n", stmt)
	}
	result += "]"
	return result
}

func (n *BlockNode) GenerateGoCode() string {
	var code []string
	for _, stmt := range n.Statements {
		code = append(code, stmt.GenerateGoCode())
	}
	return strings.Join(code, "\n")
}

// ########################################################
// ########################################################
// ##############        RANGE NODE        ################
// ########################################################
// ########################################################

// RangeNode Represents the range() function
type RangeNode struct {
	Start ASTNode
	Stop  ASTNode
	Step  ASTNode
}

// Handle the String() method for debugging
func (n *RangeNode) String() string {
	return fmt.Sprintf("Range(%v, %v, %v)", n.Start, n.Stop, n.Step)
}

// GenerateGoCode Generate Go code for range(start, stop, step) to create a slice of integers
func (n *RangeNode) GenerateGoCode() string {
	startCode := "0" // Default start is 0
	stepCode := "1"  // Default step is 1

	if n.Start != nil {
		startCode = n.Start.GenerateGoCode()
	}
	if n.Step != nil {
		stepCode = n.Step.GenerateGoCode()
	}

	return fmt.Sprintf("generateRange(%s, %s, %s)", startCode, n.Stop.GenerateGoCode(), stepCode)
}

// ########################################################
// ########################################################
// ##############          PARSERS           ##############
// ########################################################
// ########################################################

// Parse a single statement (either an assignment, method call, or a keyword statement)
func (p *Parser) parseStatement() ASTNode {
	tok := p.currentToken()

	// Handle keywords (if, while, for, etc.)
	if tok.Type == lexer.TokenKeyword {
		switch tok.Literal {
		case "if":
			return p.parseIf()
		case "while":
			return p.parseWhile()
		case "for":
			return p.parseFor()
		case "print":
			return p.parsePrint()
		}
	}

	// Handle identifiers (could be an assignment or method call like arr.append())
	if tok.Type == lexer.TokenIdentifier {
		// Look ahead to check if the next token is '=' (indicating an assignment)
		if p.lookaheadToken(1).Type == lexer.TokenOperator && p.lookaheadToken(1).Literal == "=" {
			stmt := p.parseAssignment() // Parse assignment
			//p.skipNewlines()            // Skip any newlines after the statement
			return stmt
		}
		// Look ahead to check if the next token is '.' (indicating a method call)
		if p.lookaheadToken(1).Type == lexer.TokenDot {
			stmt := p.parseMethodCall(tok.Literal) // Parse method call
			//p.skipNewlines()                       // Skip any newlines after the statement
			return stmt
		}

		// If neither '=' nor '.' follows the identifier, it's just a standalone identifier
		return &LiteralNode{Value: tok.Literal}
	}

	if tok.Type == lexer.TokenNewline {
		return &NewlineNode{}
	}

	panic(fmt.Sprintf("Unexpected token in statement: %v", tok))
}

// Parse a print statement like `print("Hello, Simple!")`
func (p *Parser) parsePrint() ASTNode {
	p.nextToken() // Skip 'print'

	if p.currentToken().Type != lexer.TokenParenOpen {
		panic(fmt.Sprintf("Expected '(' after 'print', got: %v", p.currentToken()))
	}
	p.nextToken() // Skip '('

	// Parse the string or argument to print
	arg := p.parseExpression(0)

	if p.currentToken().Type != lexer.TokenParenClose {
		panic(fmt.Sprintf("Expected ')' after print argument, got: %v", p.currentToken()))
	}
	p.nextToken() // Skip ')'

	return &PrintNode{Arg: arg}
}

// Parse an assignment like arr = [1, 2, 3]
func (p *Parser) parseAssignment() ASTNode {
	varName := p.currentToken().Literal
	p.nextToken() // Skip the identifier

	if p.currentToken().Type != lexer.TokenOperator || p.currentToken().Literal != "=" {
		panic("Expected '=' in assignment")
	}
	p.nextToken() // Skip the '='

	// Parse the right-hand side value (could be an array, literal, etc.)
	value := p.parseExpression(0)
	// Update the symbol table with the variable and its type
	switch value.(type) {
	case *ArrayNode:
		symbolTable[varName] = "array"
	case *DictionaryNode:
		symbolTable[varName] = "dictionary"
	case *LiteralNode:
		symbolTable[varName] = determineLiteralType(value.(*LiteralNode))
	default:
		symbolTable[varName] = "unknown"
	}

	// Ensure the parser advances to the next token after the assignment
	p.nextToken()

	return &AssignmentNode{Variable: varName, Value: value}
}

// Parse range(start, stop, step) expression
func (p *Parser) parseRange() ASTNode {
	p.nextToken() // Skip 'range'

	if p.currentToken().Type != lexer.TokenParenOpen {
		panic(fmt.Sprintf("Expected '(' after 'range', got: %v", p.currentToken()))
	}
	p.nextToken() // Skip '('

	var start, stop, step ASTNode
	if p.currentToken().Type != lexer.TokenParenClose {
		// Parse the first argument (stop or start)
		start = p.parseExpression(0)

		if p.currentToken().Type == lexer.TokenComma {
			p.nextToken() // Skip comma

			// Parse the second argument (stop)
			stop = p.parseExpression(0)

			if p.currentToken().Type == lexer.TokenComma {
				p.nextToken() // Skip comma

				// Parse the third argument (step)
				step = p.parseExpression(0)
			}
		} else {
			// If only one argument is provided, treat it as 'stop'
			stop = start
			start = nil
		}
	}

	if p.currentToken().Type != lexer.TokenParenClose {
		panic(fmt.Sprintf("Expected ')' after range arguments, got: %v", p.currentToken()))
	}
	p.nextToken() // Skip ')'

	return &RangeNode{Start: start, Stop: stop, Step: step}
}

// Parse a block of statements (indented code)
func (p *Parser) parseBlock() *BlockNode {
	var statements []ASTNode
	for p.currentToken().Type != lexer.TokenDedent && p.currentToken().Type != lexer.TokenEOF {
		stmt := p.parseStatement() // Parse each statement in the block
		statements = append(statements, stmt)

		// Skip over any newlines between statements
		if p.currentToken().Type == lexer.TokenNewline {
			p.nextToken()
		}
	}
	if p.currentToken().Type == lexer.TokenNewline {
		p.nextToken()
	}
	return &BlockNode{Statements: statements}
}

// Parse an expression with the given precedence
func (p *Parser) parseExpression(minPrec int) ASTNode {
	// If we're parsing a range expression
	if p.currentToken().Literal == "range" {
		return p.parseRange() // Handle the range expression directly
	}

	lhs := p.parseUnary()              // Start by parsing a unary expression
	return p.parseBinary(lhs, minPrec) // Parse the binary operations (if any)
}

// ParsePrimary handles primary expressions such as literals, identifiers, arrays, and method calls
func (p *Parser) parsePrimary() ASTNode {
	tok := p.currentToken()
	switch tok.Type {
	// Handle array literals
	case lexer.TokenBracketOpen:
		return p.parseArray()
	// Handle dictionary literals
	case lexer.TokenBraceOpen:
		return p.parseDictionary()
	// Handle literals like numbers, strings, booleans
	case lexer.TokenNumber, lexer.TokenString, lexer.TokenTrue, lexer.TokenFalse:
		p.nextToken() // Consume the literal
		return &LiteralNode{Value: tok.Literal}
	// Handle identifiers (variables or function calls)
	case lexer.TokenIdentifier:
		identifier := tok.Literal
		// Check if it's a method call (identifier followed by a dot)
		if p.lookaheadToken(1).Type == lexer.TokenDot {
			return p.parseMethodCall(identifier)
		}
		p.nextToken() // Consume the identifier

		// Check if it's a method call (identifier followed by a dot)
		//if p.currentToken().Type == lexer.TokenDot {
		//	return p.parseMethodCall(identifier) // Parse method call
		//}
		return &LiteralNode{Value: identifier} // Otherwise, treat it as a variable
	// Handle parenthesized expressions
	case lexer.TokenParenOpen:
		p.nextToken()                // Skip '('
		expr := p.parseExpression(0) // Parse the inner expression
		if p.currentToken().Type != lexer.TokenParenClose {
			panic(fmt.Sprintf("Expected ')' but got %v", p.currentToken()))
		}
		p.nextToken() // Skip ')'
		return expr
	// Handle illegal tokens
	case lexer.TokenIllegal:
		panic(fmt.Sprintf("Illegal token '%s' encountered at line %d, column %d\n", tok.Literal, tok.Line, tok.Column))
	// Handle unexpected tokens (like the issue you're encountering)
	default:
		panic(fmt.Sprintf("Unexpected token in expression: %v", tok))
	}
}

// Parse unary operators like -a or +a
func (p *Parser) parseUnary() ASTNode {
	if p.currentToken().Literal == "-" || p.currentToken().Literal == "+" {
		operator := p.currentToken().Literal
		p.nextToken()             // Skip operator
		operand := p.parseUnary() // Recursively parse the operand
		return &UnaryOperationNode{Operator: operator, Operand: operand}
	}
	return p.parsePrimary()
}

// Parse binary operations while respecting operator precedence
func (p *Parser) parseBinary(lhs ASTNode, minPrec int) ASTNode {
	for {
		tok := p.currentToken()
		prec := getPrecedence(tok.Literal)

		// If the precedence of the current operator is lower than minPrec, stop parsing
		if prec < minPrec {
			return lhs
		}

		// Otherwise, consume the operator and parse the right-hand side
		op := tok.Literal
		p.nextToken()         // Skip the operator
		rhs := p.parseUnary() // Parse the right-hand side (which could be a unary expression)

		// If the next operator has higher precedence, recursively parse the right-hand side
		for {
			nextPrec := getPrecedence(p.currentToken().Literal)
			if nextPrec > prec {
				rhs = p.parseBinary(rhs, nextPrec)
			} else {
				break
			}
		}

		// Build a BinaryOperationNode
		lhs = &BinaryOperationNode{
			Left:     lhs,
			Operator: op,
			Right:    rhs,
		}
	}
}

// Helper function to skip over newlines
func (p *Parser) skipNewlines() {
	for p.currentToken().Type == lexer.TokenNewline {
		p.nextToken()
	}
}

// Parse a method call like object.method(args)
func (p *Parser) parseMethodCall(objectName string) ASTNode {
	// The current token should be a dot ('.')
	//objectType, _ := symbolTable[objectName]
	//mName := p.lookaheadToken(2).Literal
	//inPlace, _ := ObjectMethodsProperties[objectType][mName]["in-place"]
	p.nextToken() // If the method alters the object, consume the next token before processing the method call

	if p.currentToken().Type != lexer.TokenDot {
		panic(fmt.Sprintf("Expected '.' after identifier, got %v", p.currentToken()))
	}

	p.nextToken() // Skip the '.'

	// The next token should be the method name (e.g., 'append', 'extend')
	if p.currentToken().Type != lexer.TokenIdentifier {
		panic(fmt.Sprintf("Expected method name after '.', got %v", p.currentToken()))
	}
	methodName := p.currentToken().Literal
	p.nextToken() // Skip the method name

	// The next token should be the opening parenthesis '('
	if p.currentToken().Type != lexer.TokenParenOpen {
		panic(fmt.Sprintf("Expected '(' after method name, got %v", p.currentToken()))
	}

	// Parse the arguments of the method call
	var arguments []ASTNode
	p.nextToken() // Skip the '('
	if p.currentToken().Type != lexer.TokenParenClose {
		for {
			// Parse each argument
			arg := p.parseExpression(0)
			arguments = append(arguments, arg)

			// If the next token is a comma, continue parsing arguments
			if p.currentToken().Type == lexer.TokenComma {
				p.nextToken() // Skip the comma
			} else {
				break
			}
		}
	}

	// Expect closing parenthesis
	if p.currentToken().Type != lexer.TokenParenClose {
		panic(fmt.Sprintf("Expected ')' after arguments, got %v", p.currentToken()))
	}
	p.nextToken() // Skip the ')'

	// Resolve the method call based on the object type
	return p.resolveMethodCall(objectName, methodName, arguments)
}

// Resolve the method call based on the object and method name
func (p *Parser) resolveMethodCall(objectName, methodName string, arguments []ASTNode) ASTNode {
	// Lookup the object's type in the symbol table
	objectType, exists := symbolTable[objectName]
	if !exists {
		panic(fmt.Sprintf("Unknown object '%s'", objectName))
	}

	// Dispatch method based on the object's type
	switch objectType {
	case "array":
		return p.resolveArrayMethodCall(objectName, methodName, arguments)
	case "dictionary":
		return p.resolveDictionaryMethodCall(objectName, methodName, arguments)
	case "string":
		return p.resolveStringMethodCall(objectName, methodName, arguments) // Handle string methods
	// Add more types as needed
	default:
		panic(fmt.Sprintf("Unknown method '%s' for object type '%s'", methodName, objectType))
	}
}

// Resolve string methods like 'upper', 'lower', 'replace', etc.
func (p *Parser) resolveStringMethodCall(objectName, methodName string, arguments []ASTNode) ASTNode {
	switch methodName {
	case "upper":
		return &UpperNode{StringName: objectName}
	case "lower":
		return &LowerNode{StringName: objectName}
	case "replace":
		if len(arguments) != 2 {
			panic("replace() expects exactly 2 arguments: old substring and new substring")
		}
		return &ReplaceNode{StringName: objectName, OldSubstring: arguments[0], NewSubstring: arguments[1]}
	case "split":
		if len(arguments) != 1 {
			panic("split() expects exactly 1 argument: the separator")
		}
		return &SplitNode{StringName: objectName, Separator: arguments[0]}
	case "join":
		if len(arguments) != 1 {
			panic("join() expects exactly 1 argument: the list to join")
		}
		return &JoinNode{StringName: objectName, Elements: arguments[0]}
	case "find":
		if len(arguments) != 1 {
			panic("find() expects exactly 1 argument: the substring to find")
		}
		return &FindNode{StringName: objectName, Substring: arguments[0]}
	case "startswith":
		if len(arguments) != 1 {
			panic("startswith() expects exactly 1 argument: the prefix")
		}
		return &StartsWithNode{StringName: objectName, Prefix: arguments[0]}
	case "endswith":
		if len(arguments) != 1 {
			panic("endswith() expects exactly 1 argument: the suffix")
		}
		return &EndsWithNode{StringName: objectName, Suffix: arguments[0]}
	case "strip":
		return &StripNode{StringName: objectName}
	case "lstrip":
		return &LStripNode{StringName: objectName}
	case "rstrip":
		return &RStripNode{StringName: objectName}
	case "capitalize":
		return &CapitalizeNode{StringName: objectName}
	case "count":
		if len(arguments) != 1 {
			panic("count() expects exactly 1 argument: the substring to count")
		}
		return &CountNode{StringName: objectName, Substring: arguments[0]}
	// Add more string methods here as needed
	default:
		panic(fmt.Sprintf("Unknown string method: %s", methodName))
	}
}

// Resolve array methods like 'append' and 'extend'
func (p *Parser) resolveArrayMethodCall(objectName, methodName string, arguments []ASTNode) ASTNode {
	if methodName == "append" {
		if len(arguments) != 1 {
			panic("append() expects exactly 1 argument")
		}
		return &AppendNode{ArrayName: objectName, Element: arguments[0]}
	} else if methodName == "extend" {
		if len(arguments) != 1 {
			panic("extend() expects exactly 1 argument")
		}
		return &ExtendNode{ArrayName: objectName, Elements: arguments[0]}
	}

	panic(fmt.Sprintf("Unknown array method: %s", methodName))
}

// Resolve dictionary methods (future expansion)
func (p *Parser) resolveDictionaryMethodCall(objectName, methodName string, arguments []ASTNode) ASTNode {
	// Example: Add methods like 'update' for dictionaries in the future
	if methodName == "update" {
		if len(arguments) != 1 {
			panic("update() expects exactly 1 argument")
		}
		return &UpdateNode{DictName: objectName, NewEntries: arguments[0]}
	}

	panic(fmt.Sprintf("Unknown dictionary method: %s", methodName))
}
