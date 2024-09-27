package parser

import (
	"fmt"
	"simple/lexer"
)

// ########################################################
// ########################################################
// ##############         IF NODE          ################
// ########################################################
// ########################################################

// IfNode / Represents an if statement
type IfNode struct {
	Condition ASTNode
	Body      *BlockNode
	ElseBody  *BlockNode
}

func (n *IfNode) String() string {
	if n.ElseBody != nil {
		return fmt.Sprintf("If: %v then %v else %v", n.Condition, n.Body, n.ElseBody)
	}
	return fmt.Sprintf("If: %v then %v", n.Condition, n.Body)
}

func (n *IfNode) GenerateGoCode() string {
	ifCode := fmt.Sprintf("if %s {\n%s\n}", n.Condition.GenerateGoCode(), n.Body.GenerateGoCode())
	if n.ElseBody != nil {
		ifCode += fmt.Sprintf(" else {\n%s\n}", n.ElseBody.GenerateGoCode())
	}
	return ifCode
}

// Parse an if statement
func (p *Parser) parseIf() ASTNode {
	p.nextToken() // Skip 'if'

	// Parse the condition using the new parseExpression
	condition := p.parseExpression(0)

	// Expect a colon (':') after the condition
	if p.currentToken().Type != lexer.TokenColon {
		panic(fmt.Sprintf("Expected ':' after if condition, got: %v", p.currentToken()))
	}
	p.nextToken() // Skip ':'

	// Expect a newline after the colon
	if p.currentToken().Type != lexer.TokenNewline {
		panic(fmt.Sprintf("Expected newline after ':', got: %v", p.currentToken()))
	}
	p.nextToken() // Skip newline

	// Parse the indented block for the if body
	if p.currentToken().Type != lexer.TokenIndent {
		panic(fmt.Sprintf("Expected INDENT after 'if', got: %v", p.currentToken()))
	}
	p.nextToken() // Skip indent

	body := p.parseBlock()

	// Expect dedent after the if block
	if p.currentToken().Type != lexer.TokenDedent {
		panic(fmt.Sprintf("Expected DEDENT after if block, got: %v", p.currentToken()))
	}
	p.nextToken() // Skip dedent

	// Optionally parse the else block
	var elseBody *BlockNode
	if p.currentToken().Type == lexer.TokenKeyword && p.currentToken().Literal == "else" {
		p.nextToken() // Skip 'else'
		if p.currentToken().Type != lexer.TokenColon {
			panic(fmt.Sprintf("Expected ':' after 'else', got: %v", p.currentToken()))
		}
		p.nextToken() // Skip ':'

		if p.currentToken().Type != lexer.TokenNewline {
			panic(fmt.Sprintf("Expected newline after ':', got: %v", p.currentToken()))
		}
		p.nextToken() // Skip newline

		if p.currentToken().Type != lexer.TokenIndent {
			panic(fmt.Sprintf("Expected INDENT after 'else', got: %v", p.currentToken()))
		}
		p.nextToken() // Skip indent

		elseBody = p.parseBlock()

		if p.currentToken().Type != lexer.TokenDedent {
			panic(fmt.Sprintf("Expected DEDENT after else block, got: %v", p.currentToken()))
		}
		p.nextToken() // Skip dedent
	}

	return &IfNode{Condition: condition, Body: body, ElseBody: elseBody}
}
