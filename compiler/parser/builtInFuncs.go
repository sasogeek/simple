package parser

import "fmt"

// ########################################################
// ########################################################
// ##############        PRINT NODE        ################
// ########################################################
// ########################################################

// PrintNode Represents a print statement
type PrintNode struct {
	Arg ASTNode
}

func (n *PrintNode) String() string {
	return fmt.Sprintf("Print: %v", n.Arg)
}

func (n *PrintNode) GenerateGoCode() string {
	usedPackages["fmt"] = true // Mark fmt as used
	return fmt.Sprintf("fmt.Println(%s)", n.Arg.GenerateGoCode())
}
