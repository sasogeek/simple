package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"simple/codegen"
	"simple/lexer"
	"simple/parser"
	"simple/semantic"
	"simple/transformer"
	"strings"
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
	cmd := exec.Command(binaryName)
	cmd.Dir = filepath.Dir(binaryName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to run binary: %w", err)
	}

	return nil
}

func stdlib() ([]string, error) {
	var files []string
	usr, err := user.Current()
	homeDir := usr.HomeDir
	dir := filepath.Join(homeDir, "simple/stdlib")
	entries, err := os.ReadDir(dir)
	//fmt.Println("entries: ", entries)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}

	return files, nil
}

func compile(content string, outputDir string, isMain bool) {
	// Initialize Lexer
	l := lexer.NewLexer(content)

	// Initialize Parser
	p := parser.NewParser(l)

	// Parse the program
	ast := p.ParseProgram()

	// Initialize Semantic Analyzer
	analyzer := semantic.NewAnalyzer()

	// Perform Semantic Analysis
	analyzer.Analyze(ast, []parser.Statement{})

	// Initialize Transformer
	transformer := transformer.NewTransformer(analyzer)

	// Perform Transformation
	transformer.Transform(ast, ast)

	// Initialize Code Generator
	cg := codegen.NewCodeGenerator(outputDir, analyzer, isMain)

	// Generate Go Code
	err := cg.GenerateCode(ast)
	if err != nil {
		fmt.Println("Error:", err)
		//return
	}
}

const version = "Simple 0.0.4"

func main() {
	// Check if the --version flag is passed
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		fmt.Println(version)
		return
	}

	//filename := "examples/test.simple"
	filename := os.Args[1]
	mainContent, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	// Code Generation
	binaryName := filename[:len(filename)-7]
	cwd, _ := os.Getwd()
	outputDir := filepath.Join(cwd, binaryName)
	os.MkdirAll(outputDir, os.ModePerm)
	//fmt.Println("output directory: ", outputDir)

	stdlibFiles, err := stdlib()
	for _, file := range stdlibFiles {
		content, err := os.ReadFile(file)
		if err == nil {
			destDir := filepath.Join(outputDir, "lib/"+strings.Split(filepath.Base(file), ".")[0])
			//fmt.Println("stdlib dest: ", destDir)
			os.MkdirAll(destDir, os.ModePerm)
			compile(string(content), destDir, false)
		}
	}

	compile(string(mainContent), outputDir, true)

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

	fmt.Printf("%s/%s\n", outputDir, binaryName)

	// Step 3: Run the binary
	err = runBinary(filepath.Join(outputDir, filepath.Base(binaryName)))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
}
