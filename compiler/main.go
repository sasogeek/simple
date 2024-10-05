package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"simple/codegen"
	"simple/lexer"
	"simple/parser"
	"simple/semantic"
	"simple/transformer"
)

// Function to navigate to a directory and create go.mod with a given Go version
func createGoMod(dir, goVersion string) error {
	err := os.Chdir(dir)
	if err != nil {
		return fmt.Errorf("failed to navigate to directory: %w", err)
	}

	if _, err = os.Stat("go.mod"); err == nil {

	} else if os.IsNotExist(err) {
		cmd := exec.Command("go", "mod", "init", filepath.Base(dir))
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("failed to create go.mod file: %w", err)
		}
		fmt.Println("go.mod file created successfully.")
	} else {
		return fmt.Errorf("failed to check if go.mod exists: %w", err)
	}

	// Update the Go version in the go.mod file
	cmd := exec.Command("go", "mod", "edit", "-go", goVersion)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	cmd = exec.Command("go", "mod", "tidy")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to set Go version: %w", err)
	}

	return nil
}

// Function to run go build and return the binary's name
func buildGoProject(dir string) (string, error) {
	binaryName := filepath.Base(dir)

	// Run go build
	cmd := exec.Command("go", "build", "-o", binaryName)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to build the project: %w", err)
	}

	return binaryName, nil
}

// Function to run the binary
func runBinary(binaryName string) error {
	if _, err := os.Stat(binaryName); os.IsNotExist(err) {
		return fmt.Errorf("binary does not exist: %w", err)
	}

	// Execute the binary
	cmd := exec.Command("./" + binaryName)
	cmd.Dir = filepath.Dir(binaryName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to run binary: %w", err)
	}

	return nil
}

const version = "Simple 3.2024.10"

func main() {
	// Check if the --version flag is passed
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		fmt.Println(version)
		return
	}

	//filename := "test.simple"
	filename := os.Args[1]
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	// Initialize Lexer
	l := lexer.NewLexer(string(content))

	// Initialize Parser
	p := parser.NewParser(l)

	// Parse the program
	ast := p.ParseProgram()

	// Initialize Semantic Analyzer
	analyzer := semantic.NewAnalyzer()

	// Perform Semantic Analysis
	analyzer.Analyze(ast)

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
	//if err != nil {
	//	fmt.Println("Error:", err)
	//	return
	//}

	goVersion := "1.23.1"

	// Step 1: Create go.mod file
	err = createGoMod(outputDir, goVersion)
	//if err != nil {
	//	fmt.Println("Error:", err)
	//	return
	//}

	// Step 2: Build the project
	_, err = buildGoProject(outputDir)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Build successful! Binary: %s\n", binaryName)

	// Step 3: Run the binary
	err = runBinary(binaryName)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Compilation successful! Generated Go code is in the '", outputDir, "' directory.")
}
