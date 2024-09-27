package parser

import (
	"fmt"
	"simple/lexer"
	"strings"
)

// ########################################################
// ########################################################
// ##############     DICTIONARY NODE      ################
// ########################################################
// ########################################################

// DictionaryNode Node representing a dictionary (map)
type DictionaryNode struct {
	Pairs map[ASTNode]ASTNode
}

func (n *DictionaryNode) String() string {
	return fmt.Sprintf("Dictionary: %v", n.Pairs)
}

func (n *DictionaryNode) GenerateGoCode() string {
	var pairs []string
	for key, value := range n.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s", key.GenerateGoCode(), value.GenerateGoCode()))
	}
	// Return the Go code for a map with dynamic key/value types (map[interface{}]interface{})
	return fmt.Sprintf("map[interface{}]interface{}{%s}", strings.Join(pairs, ", "))
}

// ########################################################
// ########################################################
// ##############      UPDATE NODE        #################
// ########################################################
// ########################################################

// UpdateNode Represents the dictionary update operation (dict.update(another_dict))
type UpdateNode struct {
	DictName   string
	NewEntries ASTNode
}

func (n *UpdateNode) String() string {
	return fmt.Sprintf("Update(%s, %v)", n.DictName, n.NewEntries)
}

func (n *UpdateNode) GenerateGoCode() string {
	return fmt.Sprintf(`
for k, v := range %s {
    %s[k] = v
}`, n.NewEntries.GenerateGoCode(), n.DictName)
}

func (p *Parser) parseDictionary() ASTNode {
	p.nextToken() // Skip the '{'
	pairs := make(map[ASTNode]ASTNode)

	// Parse key-value pairs until we hit a closing '}'
	for p.currentToken().Type != lexer.TokenBraceClose {
		// Parse the key (can be any expression)
		key := p.parseExpression(0)

		// Expect a colon between key and value
		if p.currentToken().Type != lexer.TokenColon {
			panic(fmt.Sprintf("Expected ':' after dictionary key, got %v", p.currentToken()))
		}
		p.nextToken() // Skip ':'

		// Parse the value
		value := p.parseExpression(0)
		pairs[key] = value

		// If there's a comma, skip it
		if p.currentToken().Type == lexer.TokenComma {
			p.nextToken()
		} else {
			break
		}
	}

	// Expect and skip the closing '}'
	if p.currentToken().Type != lexer.TokenBraceClose {
		panic(fmt.Sprintf("Expected '}' but got %v", p.currentToken()))
	}
	p.nextToken() // Skip '}'
	return &DictionaryNode{Pairs: pairs}
}
