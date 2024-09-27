package parser

import "fmt"

// ########################################################
// ########################################################
// ##############      UPPER NODE        ###################
// ########################################################
// ########################################################

// UpperNode Represents the string.upper() method
type UpperNode struct {
	StringName string
}

func (n *UpperNode) String() string {
	return fmt.Sprintf("Upper(%s)", n.StringName)
}

func (n *UpperNode) GenerateGoCode() string {
	usedPackages["strings"] = true // Ensure the strings package is imported
	return fmt.Sprintf("%s = strings.ToUpper(%s)", n.StringName, n.StringName)
}

// ########################################################
// ########################################################
// ##############      LOWER NODE        ###################
// ########################################################
// ########################################################

// LowerNode Represents the string.lower() method
type LowerNode struct {
	StringName string
}

func (n *LowerNode) String() string {
	return fmt.Sprintf("Lower(%s)", n.StringName)
}

func (n *LowerNode) GenerateGoCode() string {
	usedPackages["strings"] = true // Ensure the strings package is imported
	return fmt.Sprintf("%s = strings.ToLower(%s)", n.StringName, n.StringName)
}

// ########################################################
// ########################################################
// ##############     REPLACE NODE       ###################
// ########################################################
// ########################################################

// ReplaceNode Represents the string.replace() method
type ReplaceNode struct {
	StringName   string
	OldSubstring ASTNode
	NewSubstring ASTNode
}

func (n *ReplaceNode) String() string {
	return fmt.Sprintf("Replace(%s, %v, %v)", n.StringName, n.OldSubstring, n.NewSubstring)
}

func (n *ReplaceNode) GenerateGoCode() string {
	usedPackages["strings"] = true // Ensure the strings package is imported
	return fmt.Sprintf("%s = strings.ReplaceAll(%s, %s, %s)", n.StringName, n.StringName, n.OldSubstring.GenerateGoCode(), n.NewSubstring.GenerateGoCode())
}

// ########################################################
// ########################################################
// ##############        SPLIT NODE         ###############
// ########################################################
// ########################################################

// SplitNode Represents the string.split(separator) method
type SplitNode struct {
	StringName string
	Separator  ASTNode
}

func (n *SplitNode) String() string {
	return fmt.Sprintf("Split(%s, %v)", n.StringName, n.Separator)
}

func (n *SplitNode) GenerateGoCode() string {
	usedPackages["strings"] = true // Ensure the strings package is imported
	return fmt.Sprintf("strings.Split(%s, %s)", n.StringName, n.Separator.GenerateGoCode())
}

// ########################################################
// ########################################################
// ##############        JOIN NODE         ################
// ########################################################
// ########################################################

// JoinNode Represents the string.join(list) method
type JoinNode struct {
	StringName string
	Elements   ASTNode
}

func (n *JoinNode) String() string {
	return fmt.Sprintf("Join(%s, %v)", n.StringName, n.Elements)
}

func (n *JoinNode) GenerateGoCode() string {
	usedPackages["strings"] = true // Ensure the strings package is imported
	return fmt.Sprintf("strings.Join(%s, %s)", n.Elements.GenerateGoCode(), n.StringName)
}

// ########################################################
// ########################################################
// ##############         FIND NODE         ################
// ########################################################
// ########################################################

// FindNode Represents the string.find(substring) method
type FindNode struct {
	StringName string
	Substring  ASTNode
}

func (n *FindNode) String() string {
	return fmt.Sprintf("Find(%s, %v)", n.StringName, n.Substring)
}

func (n *FindNode) GenerateGoCode() string {
	usedPackages["strings"] = true // Ensure the strings package is imported
	return fmt.Sprintf("strings.Index(%s, %s)", n.StringName, n.Substring.GenerateGoCode())
}

// ########################################################
// ########################################################
// ##############    STARTSWITH NODE       ################
// ########################################################
// ########################################################

// StartsWithNode Represents the string.startswith(prefix) method
type StartsWithNode struct {
	StringName string
	Prefix     ASTNode
}

func (n *StartsWithNode) String() string {
	return fmt.Sprintf("StartsWith(%s, %v)", n.StringName, n.Prefix)
}

func (n *StartsWithNode) GenerateGoCode() string {
	usedPackages["strings"] = true // Ensure the strings package is imported
	return fmt.Sprintf("strings.HasPrefix(%s, %s)", n.StringName, n.Prefix.GenerateGoCode())
}

// ########################################################
// ########################################################
// ##############     ENDSWITH NODE        ################
// ########################################################
// ########################################################

// EndsWithNode Represents the string.endswith(suffix) method
type EndsWithNode struct {
	StringName string
	Suffix     ASTNode
}

func (n *EndsWithNode) String() string {
	return fmt.Sprintf("EndsWith(%s, %v)", n.StringName, n.Suffix)
}

func (n *EndsWithNode) GenerateGoCode() string {
	usedPackages["strings"] = true // Ensure the strings package is imported
	return fmt.Sprintf("strings.HasSuffix(%s, %s)", n.StringName, n.Suffix.GenerateGoCode())
}

// ########################################################
// ########################################################
// ##############        STRIP NODE         ###############
// ########################################################
// ########################################################

// StripNode Represents the string.strip() method
type StripNode struct {
	StringName string
}

func (n *StripNode) String() string {
	return fmt.Sprintf("Strip(%s)", n.StringName)
}

func (n *StripNode) GenerateGoCode() string {
	usedPackages["strings"] = true // Ensure the strings package is imported
	return fmt.Sprintf("%s = strings.TrimSpace(%s)", n.StringName, n.StringName)
}

// ########################################################
// ########################################################
// ##############       LSTRIP NODE         ###############
// ########################################################
// ########################################################

// LStripNode Represents the string.lstrip() method
type LStripNode struct {
	StringName string
}

func (n *LStripNode) String() string {
	return fmt.Sprintf("LStrip(%s)", n.StringName)
}

func (n *LStripNode) GenerateGoCode() string {
	usedPackages["strings"] = true // Ensure the strings package is imported
	return fmt.Sprintf("%s = strings.TrimLeftFunc(%s, unicode.IsSpace)", n.StringName, n.StringName)
}

// ########################################################
// ########################################################
// ##############       RSTRIP NODE         ###############
// ########################################################
// ########################################################

// RStripNode Represents the string.rstrip() method
type RStripNode struct {
	StringName string
}

func (n *RStripNode) String() string {
	return fmt.Sprintf("RStrip(%s)", n.StringName)
}

func (n *RStripNode) GenerateGoCode() string {
	usedPackages["strings"] = true // Ensure the strings package is imported
	return fmt.Sprintf("%s = strings.TrimRightFunc(%s, unicode.IsSpace)", n.StringName, n.StringName)
}

// ########################################################
// ########################################################
// ##############     CAPITALIZE NODE       ################
// ########################################################
// ########################################################

// CapitalizeNode Represents the string.capitalize() method
type CapitalizeNode struct {
	StringName string
}

func (n *CapitalizeNode) String() string {
	return fmt.Sprintf("Capitalize(%s)", n.StringName)
}

func (n *CapitalizeNode) GenerateGoCode() string {
	usedPackages["strings"] = true // Ensure the strings package is imported
	return fmt.Sprintf("%s = strings.Title(%s)", n.StringName, n.StringName)
}

// ########################################################
// ########################################################
// ##############        COUNT NODE         ################
// ########################################################
// ########################################################

// CountNode Represents the string.count(substring) method
type CountNode struct {
	StringName string
	Substring  ASTNode
}

func (n *CountNode) String() string {
	return fmt.Sprintf("Count(%v)", n.Substring)
}

func (n *CountNode) GenerateGoCode() string {
	usedPackages["strings"] = true // Ensure the strings package is imported
	return fmt.Sprintf("strings.Count(%s, %s))", n.StringName, n.Substring.GenerateGoCode())
}
