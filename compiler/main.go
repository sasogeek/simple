package main

import (
	"fmt"
	"os"
	"path/filepath"
	"simple/lexer"
	"simple/parser"
)

func tokenize(l *lexer.Lexer) []lexer.Token {
	// Collect all tokens
	var tokens []lexer.Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		l.UpdateLastTwoTokens(tok)

		// Always append NEWLINE and handle indentation if necessary
		if tok.Type == lexer.TokenNewline {
			// If the previous two tokens were COLON followed by NEWLINE, expect an INDENT
			if l.LastTwoTokens[0].Type == lexer.TokenColon && l.LastTwoTokens[1].Type == lexer.TokenNewline {
				indentToken := l.HandleIndentAfterColonNewline()
				tokens = append(tokens, indentToken)
			} else {
				indentOrDedent := l.HandleNewline()
				if indentOrDedent.Type != lexer.TokenNewline {
					tokens = append(tokens, indentOrDedent)
				}
			}
		}

		// Stop if we reach EOF
		if tok.Type == lexer.TokenEOF {
			break
		}
	}
	return tokens
}

const version = "Simple v1.0.0"

func main() {
	// Check if the --version flag is passed
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		fmt.Println(version)
		return
	}

	if len(os.Args) < 2 {
		fmt.Println("Usage: ./simple <filename.simple>")
		return
	}

	filename := os.Args[1]
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	// Initialize the lexer from the lexer package
	l := lexer.NewLexer(string(content))
	var tokens = tokenize(l)

	// Print the tokens for debugging
	//fmt.Println("Tokens:")
	//for _, tok := range tokens {
	//	fmt.Printf("%+v\n", tok)
	//}

	// Initialize the parser with the tokens
	p := parser.NewParser(tokens)

	// Parse the entire program
	ast := p.ParseProgram()

	binaryName := filename[:len(filename)-7]
	cwd, _ := os.Getwd()
	outputDir := filepath.Join(cwd, binaryName)
	moduleName := binaryName

	if err := parser.CompileSimpleToGo(ast, outputDir, moduleName); err != nil {
		fmt.Printf("Error during compilation: %v\n", err)
		return
	}
}
