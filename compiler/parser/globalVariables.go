package parser

// ########################################################
// ########################################################
// ##############     GLOBAL VARIABLES     ################
// ########################################################
// ########################################################

var usedPackages = map[string]bool{
	"fmt":  false,
	"os":   false,
	"math": false,
	// Add more standard packages if needed
}

// Declare a set to track declared variables
var declaredVariables = map[string]bool{}

var rangeUsed = false // Flag to track if range() is used

var symbolTable = make(map[string]string) // Global symbol table
