package main

import (
	"fmt"
	"os"
	"path/filepath"
	"simple/codegen"
	"simple/lexer"
	"simple/parser"
	"simple/semantic"
	"simple/transformer"
)

func main() {
	//if len(os.Args) < 2 {
	//	fmt.Println("Usage: ./simple <filename.simple>")
	//	return
	//}

	filename := "test.simple"
	//filename := os.Args[1]
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	// Initialize Lexer
	l := lexer.NewLexer(string(content))

	//for {
	//	tok := l.NextToken()
	//	fmt.Printf("Type: %-10s Literal: %-10q Line: %d Column: %d\n", tok.Type, tok.Literal, tok.Line, tok.Column)
	//	if tok.Type == lexer.TokenEOF {
	//		break
	//	}
	//}

	// Initialize Parser
	p := parser.NewParser(l)

	// Parse the program
	ast := p.ParseProgram()

	if len(p.Errors()) > 0 {
		fmt.Println("Parser Errors:")
		for _, msg := range p.Errors() {
			fmt.Println("\t" + msg)
		}
		os.Exit(1)
	}

	// Initialize Semantic Analyzer
	analyzer := semantic.NewAnalyzer()

	// Perform Semantic Analysis
	analyzer.Analyze(ast)

	if len(analyzer.Errors()) > 0 {
		fmt.Println("Semantic Errors:")
		for _, msg := range analyzer.Errors() {
			fmt.Println("\t" + msg)
		}
		os.Exit(1)
	}

	// Initialize Transformer
	transformer := transformer.NewTransformer(analyzer)

	// Perform Transformation
	transformer.Transform(ast)

	// Code Generation
	binaryName := filename[:len(filename)-7]
	cwd, _ := os.Getwd()
	outputDir := filepath.Join(cwd, binaryName)
	os.MkdirAll(outputDir, os.ModePerm)

	// Initialize Code Generator
	cg := codegen.NewCodeGenerator(outputDir, analyzer)

	// Generate Go Code
	err = cg.GenerateCode(ast)
	if err != nil {
		fmt.Printf("Code Generation Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Compilation successful! Generated Go code is in the '", outputDir, "' directory.")
}
