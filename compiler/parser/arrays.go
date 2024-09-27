package parser

import (
	"fmt"
	"simple/lexer"
	"strings"
)

// ########################################################
// ########################################################
// ##############       ARRAY NODE         ################
// ########################################################
// ########################################################

// ArrayNode Node representing an array (list)
type ArrayNode struct {
	Elements []ASTNode
}

func (n *ArrayNode) String() string {
	return fmt.Sprintf("Array: %v", n.Elements)
}

func (n *ArrayNode) GenerateGoCode() string {
	var elementCodes []string
	for _, elem := range n.Elements {
		elementCodes = append(elementCodes, elem.GenerateGoCode())
	}
	// Return the Go code for an array of mixed types (as []interface{})
	return fmt.Sprintf("[]interface{}{%s}", strings.Join(elementCodes, ", "))
}

// ########################################################
// ########################################################
// ##############      APPEND NODE        #################
// ########################################################
// ########################################################

// AppendNode Represents the array append operation (arr.append(element))
type AppendNode struct {
	ArrayName string
	Element   ASTNode
}

func (n *AppendNode) String() string {
	return fmt.Sprintf("Append(%s, %v)", n.ArrayName, n.Element)
}

func (n *AppendNode) GenerateGoCode() string {
	return fmt.Sprintf("%s = append(%s, %s)", n.ArrayName, n.ArrayName, n.Element.GenerateGoCode())
}

// ########################################################
// ########################################################
// ##############      EXTEND NODE        #################
// ########################################################
// ########################################################

// ExtendNode Represents the array extend operation (arr.extend(another_array))
type ExtendNode struct {
	ArrayName string
	Elements  ASTNode
}

func (n *ExtendNode) String() string {
	return fmt.Sprintf("Extend(%s, %v)", n.ArrayName, n.Elements)
}

func (n *ExtendNode) GenerateGoCode() string {
	return fmt.Sprintf("%s = append(%s, %s...)", n.ArrayName, n.ArrayName, n.Elements.GenerateGoCode())
}

func (p *Parser) parseArray() ASTNode {
	p.nextToken() // Skip the '['
	var elements []ASTNode

	// Parse each element in the array until we hit a closing ']'
	for p.currentToken().Type != lexer.TokenBracketClose {
		element := p.parseExpression(0) // Parse the array elements
		elements = append(elements, element)

		// If there's a comma, skip it
		if p.currentToken().Type == lexer.TokenComma {
			p.nextToken() // Skip comma
		} else {
			break
		}
	}

	// Expect and skip the closing ']'
	if p.currentToken().Type != lexer.TokenBracketClose {
		panic(fmt.Sprintf("Expected ']' but got %v", p.currentToken()))
	}
	p.nextToken() // Skip ']'
	return &ArrayNode{Elements: elements}
}
