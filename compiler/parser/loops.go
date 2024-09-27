package parser

import (
	"fmt"
	"simple/lexer"
)

// ########################################################
// ########################################################
// ##############        WHILE NODE        ################
// ########################################################
// ########################################################

// WhileNode Represents a while loop
type WhileNode struct {
	Condition ASTNode
	Body      *BlockNode
}

func (n *WhileNode) String() string {
	return fmt.Sprintf("While: %v do %v", n.Condition, n.Body)
}

func (n *WhileNode) GenerateGoCode() string {
	return fmt.Sprintf("for %s {\n%s\n}", n.Condition.GenerateGoCode(), n.Body.GenerateGoCode())
}

// ########################################################
// ########################################################
// ##############         FOR NODE         ################
// ########################################################
// ########################################################

// ForNode Represents a for loop
type ForNode struct {
	LoopVar  string
	Iterable ASTNode
	Body     *BlockNode
}

func (n *ForNode) String() string {
	return fmt.Sprintf("For: %v in %v do %v", n.LoopVar, n.Iterable, n.Body)
}

func (n *ForNode) GenerateGoCode() string {
	// If the iterable is a RangeNode, generate a Go 'for' loop with 'range'
	if _, ok := n.Iterable.(*RangeNode); ok {
		return fmt.Sprintf("for _, %s := range %s {\n%s\n}", n.LoopVar, n.Iterable.GenerateGoCode(), n.Body.GenerateGoCode())
	}
	// For other iterables (e.g., lists, arrays), handle them differently
	return fmt.Sprintf("for _, %s := range %s {\n%s\n}", n.LoopVar, n.Iterable.GenerateGoCode(), n.Body.GenerateGoCode())
}

// Parse a while loop
func (p *Parser) parseWhile() ASTNode {
	p.nextToken() // Skip 'while'

	// Parse the loop condition using parseExpression
	condition := p.parseExpression(0)

	if p.currentToken().Type != lexer.TokenColon {
		panic(fmt.Sprintf("Expected ':' after while condition, got: %v", p.currentToken()))
	}
	p.nextToken() // Skip ':'

	if p.currentToken().Type != lexer.TokenNewline {
		panic(fmt.Sprintf("Expected newline after ':', got: %v", p.currentToken()))
	}
	p.nextToken() // Skip newline

	if p.currentToken().Type != lexer.TokenIndent {
		panic(fmt.Sprintf("Expected INDENT after 'while', got: %v", p.currentToken()))
	}
	p.nextToken() // Skip indent

	body := p.parseBlock()

	if p.currentToken().Type != lexer.TokenDedent {
		panic(fmt.Sprintf("Expected DEDENT after while block, got: %v", p.currentToken()))
	}
	p.nextToken() // Skip dedent

	return &WhileNode{Condition: condition, Body: body}
}

// Parse a for loop (Python style: for item in iterable)
func (p *Parser) parseFor() ASTNode {
	p.nextToken() // Skip 'for'
	loopVar := p.currentToken().Literal
	p.nextToken() // Move past the loop variable

	if p.currentToken().Literal != "in" {
		panic(fmt.Sprintf("Expected 'in' after loop variable, got: %v", p.currentToken()))
	}
	p.nextToken() // Skip 'in'

	// Check if the iterable is 'range'
	var iterable ASTNode
	if p.currentToken().Literal == "range" {
		rangeUsed = true // Track that range() is used
		iterable = p.parseRange()
	} else {
		iterable = p.parseExpression(0)
	}

	if p.currentToken().Type != lexer.TokenColon {
		panic(fmt.Sprintf("Expected ':' after 'for' loop, got: %v", p.currentToken()))
	}
	p.nextToken() // Skip ':'

	if p.currentToken().Type != lexer.TokenNewline {
		panic(fmt.Sprintf("Expected newline after ':', got: %v", p.currentToken()))
	}
	p.nextToken() // Skip newline

	if p.currentToken().Type != lexer.TokenIndent {
		panic(fmt.Sprintf("Expected INDENT after 'for', got: %v", p.currentToken()))
	}
	p.nextToken() // Skip indent

	body := p.parseBlock()

	if p.currentToken().Type != lexer.TokenDedent {
		panic(fmt.Sprintf("Expected DEDENT after 'for' block, got: %v", p.currentToken()))
	}
	p.nextToken() // Skip dedent

	return &ForNode{LoopVar: loopVar, Iterable: iterable, Body: body}
}
