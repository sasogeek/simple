package parser

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"simple/lexer"
	"strings"
)

// ########################################################
// ########################################################
// ##############         PARSER           ################
// ########################################################
// ########################################################

// Parser structure
type Parser struct {
	tokens  []lexer.Token
	current int
}

// NewParser initializes the parser with the tokens from the lexer
func NewParser(tokens []lexer.Token) *Parser {
	return &Parser{tokens: tokens, current: 0}
}

// ParseProgram starts the parsing process
func (p *Parser) ParseProgram() *BlockNode {
	block := &BlockNode{Statements: []ASTNode{}}

	for p.currentToken().Type != lexer.TokenEOF {
		stmt := p.parseStatement()
		block.Statements = append(block.Statements, stmt)

		// Skip over any newlines after a statement
		if p.currentToken().Type == lexer.TokenNewline {
			p.nextToken()
		}
	}

	return block
}

// ########################################################
// ########################################################
// ##############     CONVERTER/COMPILER      #############
// ########################################################
// ########################################################

// Dynamically generate the import statements based on used packages
func writeImports(file *os.File) {
	file.WriteString("import (\n")
	for pkg, used := range usedPackages {
		if used {
			file.WriteString(fmt.Sprintf("\t\"%s\"\n", pkg))
		}
	}
	file.WriteString(")\n\n")
}

// GenerateGoProgram generates Go code from the AST and writes it to a Go file
func GenerateGoProgram(ast ASTNode, outputDir string) error {
	// Create the Go source file
	goCode := ast.GenerateGoCode()
	outputFile := filepath.Join(outputDir, "main.go")
	file, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	// Write the package declaration and imports
	_, err = file.WriteString(fmt.Sprintf("package main\n\n"))
	if err != nil {
		return err
	}

	// Write the import section dynamically based on used packages
	writeImports(file)

	// If range() is used, add the generateRange() function
	if rangeUsed {
		_, err := file.WriteString(generateRangeFunction())
		if err != nil {
			return err
		}
	}

	// Write the main function and the rest of the AST
	_, err = file.WriteString("func main() {\n")
	if err != nil {
		return err
	}
	_, err = file.WriteString(goCode)
	if err != nil {
		return err
	}
	_, err = file.WriteString("\n}\n")
	if err != nil {
		return err
	}

	return nil
}

// generateRangeFunction returns the Go code for the generateRange() function
func generateRangeFunction() string {
	return `
func generateRange(start, stop, step int) []int {
	if step == 0 {
		panic("Step cannot be 0")
	}
	var result []int
	if step > 0 {
		for i := start; i < stop; i += step {
			result = append(result, i)
		}
	} else {
		for i := start; i > stop; i += step {
			result = append(result, i)
		}
	}
	return result
}
`
}

// Initialize a Go module in the output directory (if it's not already initialized)
func initializeGoModule(outputDir string, moduleName string) error {
	// Check if go.mod already exists, skip if it does
	goModFile := filepath.Join(outputDir, "go.mod")
	if _, err := os.Stat(goModFile); err == nil {
		return nil // go.mod already exists, no need to initialize
	}

	// Run 'go mod init' in the output directory
	//cmd := exec.Command("go", "version")
	cmd := exec.Command("go", "mod", "init", moduleName)
	cmd.Dir = outputDir
	//cmd.Stdin = os.Stdin
	//cmd.Stdout = os.Stdout
	//cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Failed to create go.mod; %s\n", err)
	}
	return err
}

// Run 'go mod tidy' to clean up imports
func runGoModTidy(outputDir string) error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = outputDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to run go mod tidy: %s\n", string(output))
		return err
	}
	return nil
}

func moveFile(src, dst string) error {
	// Open the source file
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Create the destination file
	destinationFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	// Copy the contents from source to destination
	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}

	// Remove the original file
	err = os.Remove(src)
	if err != nil {
		return err
	}

	return nil
}

// Build and run the generated Go program
func buildAndRunGoProgram(outputDir string) error {
	cwd, _ := os.Getwd()
	binaryName := fmt.Sprintf(strings.Split(outputDir, "/")[len(strings.Split(outputDir, "/"))-1])
	cmd := exec.Command("go", "build", "-o", binaryName)
	cmd.Dir = outputDir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	os.Rename(outputDir, filepath.Join(cwd, binaryName+"_"))
	os.Rename(filepath.Join(filepath.Join(cwd, binaryName+"_"), binaryName), filepath.Join(cwd, binaryName))
	os.RemoveAll(filepath.Join(cwd, binaryName+"_"))

	cmd = exec.Command(fmt.Sprintf("./%s", binaryName))
	cmd.Dir = cwd
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	return err
}

func CompileSimpleToGo(ast ASTNode, outputDir string, moduleName string) error {
	cwd, _ := os.Getwd()
	os.RemoveAll(filepath.Join(cwd, moduleName))
	os.MkdirAll(outputDir, 0755)
	// Step 1: Generate the Go program (without manually handling imports)
	if err := GenerateGoProgram(ast, outputDir); err != nil {
		return err
	}
	// Step 2: Initialize a Go module in the output directory (if needed)
	if err := initializeGoModule(outputDir, moduleName); err != nil {
		fmt.Println(err)
		return err
	}
	// Step 3: Run 'go mod tidy' to manage the imports automatically
	runGoModTidy(outputDir)

	// Step 4: Build and run the generated Go program
	return buildAndRunGoProgram(outputDir)
}
