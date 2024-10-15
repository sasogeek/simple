package codegen

import (
	"fmt"
	"github.com/sasogeek/simple/compiler/lexer"
	"github.com/sasogeek/simple/compiler/parser"
	"github.com/sasogeek/simple/compiler/semantic"
	"github.com/sasogeek/simple/compiler/transformer"
	"go/types"
	"os"
	"path/filepath"
	"strings"
)

// CodeGenerator generates Go code from the AST.
type CodeGenerator struct {
	outputDir   string
	imports     map[string]bool
	indentLevel int
	analyzer    *semantic.Analyzer
	Returns     map[string]map[string]bool
	isMain      bool
	stdLib      map[string]bool
}

func NewCodeGenerator(outputDir string, analyzer *semantic.Analyzer, isMain bool) *CodeGenerator {
	stdLib := map[string]bool{
		"json": true,
	}
	return &CodeGenerator{
		outputDir:   outputDir,
		imports:     make(map[string]bool),
		indentLevel: 0,
		analyzer:    analyzer,
		Returns:     make(map[string]map[string]bool),
		isMain:      isMain,
		stdLib:      stdLib,
	}
}

// GenerateCode generates Go code from the program.
func (cg *CodeGenerator) GenerateCode(program *parser.Program) error {

	if cg.isMain {
		mainFilePath := filepath.Join(cg.outputDir, "main.go")
		mainFile, err := os.Create(mainFilePath)
		if err != nil {
			return err
		}
		defer mainFile.Close()

		fmt.Fprintln(mainFile, "package main\n")

		// Collect imports
		err = cg.collectImports(program)
		if err != nil {
			return err
		}

		// Write imports
		if len(cg.imports) > 0 {
			fmt.Fprintln(mainFile, "import (")
			for imp := range cg.imports {
				fmt.Fprintf(mainFile, "\t%q\n", imp)
			}
			fmt.Fprintln(mainFile, ")\n")
		}

		// Generate code for global statements (functions)
		for _, stmt := range program.Statements {
			if _, ok := stmt.(*parser.FunctionLiteral); ok {
				cg.generateFunction(mainFile, stmt.(*parser.FunctionLiteral), cg.analyzer.CurrentTable, false)
			}
		}

		// Generate main function
		fmt.Fprintln(mainFile, "func main() {")
		cg.indentLevel++
		for _, stmt := range program.Statements {
			if _, ok := stmt.(*parser.FunctionLiteral); !ok {
				cg.generateStatement(mainFile, stmt, cg.analyzer.CurrentTable)
			}
		}
		cg.indentLevel--
		fmt.Fprintln(mainFile, "}")

		return nil

	} else {
		mainFilePath := filepath.Join(cg.outputDir, fmt.Sprintf("%s.go", filepath.Base(cg.outputDir)))
		mainFile, err := os.Create(mainFilePath)
		if err != nil {
			return err
		}
		defer mainFile.Close()
		fmt.Fprintln(mainFile, fmt.Sprintf("package %s\n", filepath.Base(cg.outputDir)))

		// Collect imports
		err = cg.collectImports(program)
		if err != nil {
			return err
		}

		// Write imports
		if len(cg.imports) > 0 {
			fmt.Fprintln(mainFile, "import (")
			for imp := range cg.imports {
				fmt.Fprintf(mainFile, "\t%q\n", imp)
			}
			fmt.Fprintln(mainFile, ")\n")
		}

		// Generate code for global statements (functions)
		for _, stmt := range program.Statements {
			if _, ok := stmt.(*parser.FunctionLiteral); ok {
				cg.generateFunction(mainFile, stmt.(*parser.FunctionLiteral), cg.analyzer.CurrentTable, true)
			}
		}

		return nil

	}

}

// collectImports collects imports from the program.
func (cg *CodeGenerator) collectImports(program *parser.Program) error {
	for _, stmt := range program.Statements {
		if imp, ok := stmt.(*parser.ImportStatement); ok {
			if imp.IsSimpleImport {
				// Handle simple import
				packageName := imp.ImportedModule.Value
				if err := cg.processSimpleImport(packageName); err != nil {
					return fmt.Errorf("failed to process simple import '%s': %v", packageName, err)
				}
			} else {
				// Handle Go import
				module := strings.Trim(imp.ImportedModule.Value, "\"")
				cg.imports[module] = true
			}
		}
	}

	cg.imports["fmt"] = true

	return nil
}

// processSimpleImport processes a simple import by generating a separate Go package.
func (cg *CodeGenerator) processSimpleImport(packageName string) error {
	// Prevent processing the same package multiple times
	if cg.stdLib[packageName] {
		cg.imports[fmt.Sprintf("%s/lib/%s", filepath.Base(cg.outputDir), packageName)] = true
		return nil
	}
	if _, alreadyProcessed := cg.imports[packageName]; alreadyProcessed {
		return nil
	}

	// Assume the simple file has a .simple extension
	dir := filepath.Dir(cg.outputDir)
	simpleFilePath := filepath.Join(dir, packageName+".simple")
	data, err := os.ReadFile(simpleFilePath)
	if err != nil {

		return fmt.Errorf("could not read simple file '%s': %v", simpleFilePath, err)
	}

	l := lexer.NewLexer(string(data))

	// Parse the simple file
	p := parser.NewParser(l)
	ast := p.ParseProgram()

	// Perform semantic analysis
	analyzer := semantic.NewAnalyzer()
	analyzer.Analyze(ast, []parser.Statement{})

	// Initialize Transformer
	transformer := transformer.NewTransformer(analyzer)

	// Perform Transformation
	transformer.Transform(ast, ast)

	// Create a directory for the package
	packageDir := filepath.Join(cg.outputDir, packageName)
	if err := os.MkdirAll(packageDir, os.ModePerm); err != nil {
		return fmt.Errorf("could not create package directory '%s': %v", packageDir, err)
	}

	// Initialize a new CodeGenerator for the package
	packageGenerator := NewCodeGenerator(packageDir, analyzer, false)

	// Generate Go code for the imported package
	packageGenerator.GenerateCode(ast)

	// Add the package to imports (use relative path or module path as needed)
	// Here, we assume the package can be imported using its directory name
	cg.imports[fmt.Sprintf("%s/%s", filepath.Base(cg.outputDir), packageName)] = true

	return nil
}

func capitalize(name string) string {
	if name == "" {
		return ""
	}
	return strings.ToUpper(name[:1]) + name[1:]
}

// Helper function to check if a built-in function is used
func (cg *CodeGenerator) isBuiltinUsed(name string, program *parser.Program) bool {
	found := false
	parser.Inspect(program, func(n parser.Node) bool {
		if ce, ok := n.(*parser.CallExpression); ok {
			if ident, ok := ce.Function.(*parser.Identifier); ok {
				if ident.Value == name {
					found = true
					return false // Stop traversal
				}
			}
		}
		return true
	})
	if !found {
		fmt.Printf("Builtin function '%s' NOT found in AST\n", name)
	}
	return found
}

// generateFunction generates Go code for a function definition.
func (cg *CodeGenerator) generateFunction(file *os.File, fn *parser.FunctionLiteral, prevSymbolTable *semantic.SymbolTable, exported bool) {
	funcName := fn.Name.Value
	if exported {
		funcName = capitalize(funcName)
	}

	// Get the function symbol from the symbol table
	symbol, ok := cg.analyzer.CurrentTable.Resolve(fn.Name.Value)
	if !ok {
		fmt.Fprintf(os.Stderr, "Undefined function: %s\n", fn.Name.Value)
		return
	}

	functionType, ok := symbol.Type.(*parser.FunctionType)
	if !ok {
		fmt.Fprintf(os.Stderr, "Symbol '%s' is not a function\n", fn.Name.Value)
		return
	}

	params := []string{}
	for i, p := range fn.Parameters {
		paramType := "interface{}" // Default type
		if i < len(functionType.ParameterTypes) {
			switch pt := functionType.ParameterTypes[i].(type) {
			case *parser.NamedType:
				paramType = strings.Split(pt.String(), "/")[len(strings.Split(pt.String(), "/"))-1]
			case *parser.PointerType:
				elemType := strings.Split(pt.ElementType.String(), "/")
				paramType = "*" + elemType[len(elemType)-1]
			case *parser.BasicType:
				paramType = pt.String()
			case *parser.ArrayType:
				paramType = pt.String()
			case *parser.MapType:
				paramType = pt.String()
			}
		}
		params = append(params, fmt.Sprintf("%s %s", p.Value, paramType))
		paramSymbol, _ := cg.analyzer.SymbolTables.Tables[fn.Name.Value].Resolve(p.Value)
		paramSymbol.Metadata = map[string]any{"set": true}
		paramSymbol.Name = p.Value
		paramSymbol.GoType = cg.analyzer.GetGoTypeFromParserType(functionType.ParameterTypes[i])
		paramSymbol.Type = functionType.ParameterTypes[i]
	}

	// Determine return type
	returnType := ""
	cg.Returns["currentFunc"] = map[string]bool{"expects": false, "done": false}

	if len(functionType.ReturnTypes) > 0 {
		// Build the return type string
		returnTypeNames := []string{}
		for _, rt := range functionType.ReturnTypes {
			returnTypeNames = append(returnTypeNames, rt.String())
		}
		if len(returnTypeNames) == 1 {
			if returnTypeNames[0] == "void" {
				returnType = ""
			} else {
				returnType = returnTypeNames[0]
			}
		} else {
			returnType = fmt.Sprintf("(%s)", strings.Join(returnTypeNames, ", "))
		}
		cg.Returns["currentFunc"]["expects"] = true
	}

	cg.writeIndent(file)
	functionSymbol, ok := prevSymbolTable.Resolve(cg.analyzer.CurrentTable.Name)
	if ok {
		switch functionSymbol.Type.(type) {
		case *parser.FunctionType:
			if returnType != "" {
				if functionSymbol.Metadata == nil {
					fmt.Fprintf(file, "%s := func(%s) %s {\n", funcName, strings.Join(params, ", "), returnType)
					functionSymbol.Metadata = map[string]any{"set": true}
				} else {
					fmt.Fprintf(file, "%s = func(%s) %s {\n", funcName, strings.Join(params, ", "), returnType)
					functionSymbol.Metadata["set"] = true
				}
			} else {
				if functionSymbol.Metadata == nil {
					fmt.Fprintf(file, "%s := func(%s) {\n", funcName, strings.Join(params, ", "))
					functionSymbol.Metadata = map[string]any{"set": true}
				} else {
					fmt.Fprintf(file, "%s = func(%s) {\n", funcName, strings.Join(params, ", "))
					functionSymbol.Metadata["set"] = true
				}
			}
		default:
			if returnType != "" {
				fmt.Fprintf(file, "func %s(%s) %s {\n", funcName, strings.Join(params, ", "), returnType)
			} else {
				fmt.Fprintf(file, "func %s(%s) {\n", funcName, strings.Join(params, ", "))
			}
		}
	} else {
		if returnType != "" {
			fmt.Fprintf(file, "func %s(%s) %s {\n", funcName, strings.Join(params, ", "), returnType)
		} else {
			fmt.Fprintf(file, "func %s(%s) {\n", funcName, strings.Join(params, ", "))
		}
	}

	cg.indentLevel++
	prevTable := cg.analyzer.CurrentTable
	cg.analyzer.CurrentTable = cg.analyzer.SymbolTables.Tables[fn.Name.Value]
	cg.generateBlockStatement(file, fn.Body, prevTable)
	cg.indentLevel--
	cg.writeIndent(file)
	if returnType != "" {
		if cg.Returns["currentFunc"]["expects"] && cg.Returns["currentFunc"]["done"] {
			fmt.Fprintf(file, "}\n")
		} else if cg.Returns["currentFunc"]["expects"] && !cg.Returns["currentFunc"]["done"] {
			// Generate default return values
			defaultReturnValues := []string{}
			for _, rt := range functionType.ReturnTypes {
				switch rt.String() {
				case "int":
					defaultReturnValues = append(defaultReturnValues, "0")
				case "string":
					defaultReturnValues = append(defaultReturnValues, "\"\"")
				case "bool":
					defaultReturnValues = append(defaultReturnValues, "false")
				default:
					defaultReturnValues = append(defaultReturnValues, "nil")
				}
			}
			fmt.Fprintf(file, "return %s\n", strings.Join(defaultReturnValues, ", "))
			fmt.Fprintf(file, "}\n")
		} else {
			fmt.Fprintf(file, "}\n")
		}
	} else {
		fmt.Fprintln(file, "}\n")
	}
	fmt.Fprintln(file) // Add an empty line for readability
	cg.analyzer.CurrentTable = prevTable
	cg.Returns["currentFunc"]["expects"] = false
	cg.Returns["currentFunc"]["done"] = false
}

func (cg *CodeGenerator) generateAssignmentStatement(file *os.File, as *parser.AssignmentStatement) {
	cg.writeIndent(file)

	// Collect left-hand side expressions and check if any variables are undeclared
	lhsExpressions := []string{}
	useShortDeclaration := false

	for _, expr := range as.Left {
		exprStr := expr.String()

		// Check if the left-hand side is a simple identifier
		if ident, ok := expr.(*parser.Identifier); ok {
			symbol, found := cg.analyzer.CurrentTable.Resolve(ident.Value)
			if found && symbol.Metadata == nil {
				useShortDeclaration = true
			}
		}

		lhsExpressions = append(lhsExpressions, exprStr)
	}

	// Build the assignment operator
	assignmentOperator := "="
	if useShortDeclaration {
		assignmentOperator = ":="
	}

	// Generate the assignment statement
	for ex := range lhsExpressions {
		if _, found := cg.analyzer.CurrentTable.Resolve(lhsExpressions[ex]); found {
			if len(cg.analyzer.Assignments[lhsExpressions[ex]]["types"]) > 1 {
				switch as.Value.(type) {
				case *parser.CallExpression:
					continue
				case *parser.InfixExpression:
					continue
				}
				if assignmentOperator == ":=" {
					fmt.Fprintf(file, "var %s any\n", lhsExpressions[ex])
					cg.writeIndent(file)
					assignmentOperator = "="
				}
			}
		}
	}
	fmt.Fprintf(file, "%s %s ", strings.Join(lhsExpressions, ", "), assignmentOperator)
	cg.generateExpression(file, as.Value)
	fmt.Fprintln(file)

	// Update the symbol table
	for _, expr := range as.Left {
		// Only identifiers need to be added or updated in the symbol table
		if ident, ok := expr.(*parser.Identifier); ok {
			symbol, found := cg.analyzer.CurrentTable.Resolve(ident.Value)
			if !found {
				// Define the new variable in the symbol table
				// You may need to determine the actual type based on the expression
				cg.analyzer.CurrentTable.Define(ident.Value, &semantic.Symbol{
					Type:     &parser.BasicType{Name: "interface{}"}, // Adjust the type as needed
					Metadata: map[string]any{"set": true},
				})
			} else {
				// Update existing symbol metadata
				if symbol.Metadata == nil {
					symbol.Metadata = map[string]any{"set": true}
				} else {
					symbol.Metadata["set"] = true
				}
			}
		}
	}
}

// generateStatement generates Go code for a statement.
func (cg *CodeGenerator) generateStatement(file *os.File, stmt parser.Statement, prevSymbolTable *semantic.SymbolTable) {
	switch s := stmt.(type) {
	case *parser.ExpressionStatement:
		if s != nil {
			cg.writeIndent(file)
			cg.generateExpression(file, s.Expression)
			fmt.Fprintln(file)

		}
	case *parser.AssignmentStatement:
		cg.generateAssignmentStatement(file, s)
	case *parser.ReturnStatement:

		cg.writeIndent(file)
		fmt.Fprint(file, "return ")
		if s.ReturnValue != nil {
			cg.generateExpression(file, s.ReturnValue)
			cg.Returns["currentFunc"]["done"] = true
		}
		fmt.Fprintln(file)
	case *parser.IfStatement:
		cg.generateIfStatement(file, s, prevSymbolTable)
	case *parser.WhileStatement:
		cg.generateWhileStatement(file, s, prevSymbolTable)
	case *parser.ForStatement:
		cg.generateForStatement(file, s, prevSymbolTable)
	case *parser.FunctionLiteral:
		if cg.isMain {
			cg.generateFunction(file, s, prevSymbolTable, false)
		} else {
			cg.generateFunction(file, s, prevSymbolTable, true)
		}
	case *parser.DeferStatement:
		cg.generateExpression(file, s.Expression)
	case *parser.GoStatement:
		cg.generateExpression(file, s.Expression)
	default:
		// Handle other statements as needed
	}
}

// generateExpression generates Go code for an expression.
func (cg *CodeGenerator) generateExpression(file *os.File, expr parser.Expression) {
	switch e := expr.(type) {
	case *parser.Identifier:
		//if symbol, found := cg.analyzer.CurrentTable.Resolve(e.Value); found {
		//	switch symbol.Type.(type) {
		//	case *parser.BasicType:
		//		fmt.Fprintf(file, "%s(%s)", symbol.Type.String(), e.Value)
		//	default:
		//		fmt.Fprint(file, e.Value)
		//	}
		//} else {
		//	fmt.Fprintf(file, "%s", e.Value)
		//}
		fmt.Fprint(file, e.Value)
	case *parser.IntegerLiteral:
		fmt.Fprint(file, e.TokenLiteral())
	case *parser.StringLiteral:
		switch strings.Contains(e.Value, "[]byte") {
		case true:
			s := e.Value[7 : len(e.Value)-1]
			s = strings.Trim(s, "\"")
			fmt.Fprintf(file, "[]byte(%s)", s)
		default:
			fmt.Fprintf(file, "%q", e.Value)
		}

	case *parser.BooleanLiteral:
		if e.Value {
			fmt.Fprint(file, "true")
		} else {
			fmt.Fprint(file, "false")
		}
	case *parser.CallExpression:
		cg.generateCallExpression(file, e)
	case *parser.DeferLiteral:
		cg.writeIndent(file)
		fmt.Fprintf(file, "%s", e.Value)
	case *parser.GoLiteral:
		cg.writeIndent(file)
		fmt.Fprintf(file, "%s", e.Value)
	case *parser.InfixExpression:
		cg.generateInfixExpression(file, e)
	case *parser.PrefixExpression:
		cg.generatePrefixExpression(file, e)
	case *parser.SelectorExpression:
		cg.generateSelectorExpression(file, e)
	case *parser.TypeConversionExpression:
		cg.generateTypeConversionExpression(file, e)
	case *parser.ArrayLiteral:
		cg.generateArrayLiteral(file, e)
	case *parser.MapLiteral:
		cg.generateMapLiteral(file, e)
	case *parser.IndexExpression:
		fmt.Fprint(file, e.String())
	default:

	}
}

// isImportedPackage checks if a given identifier is an imported package.
func (cg *CodeGenerator) isImportedPackage(ident string) bool {
	_, exists := cg.imports[ident]
	return exists
}

func (cg *CodeGenerator) generateArrayLiteral(file *os.File, arr *parser.ArrayLiteral) {
	fmt.Fprintf(file, "[]%s{", arr.Type.String())
	for _, el := range arr.Elements {
		fmt.Fprint(file, el)
		fmt.Fprint(file, ", ")
	}
	fmt.Fprint(file, "}")
}

func (cg *CodeGenerator) generateMapLiteral(file *os.File, m *parser.MapLiteral) {
	// Determine the map type
	keyType := "any"
	valueType := "any"

	if m.Type != nil {
		if mt, ok := m.Type.(*parser.MapType); ok {
			keyType = mt.KeyType.String()
			valueType = mt.ValueType.String()
		}
	}

	// Write the map type
	fmt.Fprintf(file, "map[%s]%s{", keyType, valueType)

	// Iterate over key-value pairs
	first := true
	for key, value := range m.Pairs {
		if !first {
			fmt.Fprint(file, ", ")
		}
		first = false

		// Generate key expression
		cg.generateExpression(file, key)
		fmt.Fprint(file, ": ")

		// Generate value expression
		cg.generateExpression(file, value)
	}

	fmt.Fprint(file, "}")
}

func (cg *CodeGenerator) generateTypeConversionExpression(file *os.File, expr *parser.TypeConversionExpression) string {
	//exprCode := cg.generateExpression(file, expr.Expression)
	targetTypeCode := cg.typeToGoString(expr.TargetType)

	// Generate Go code for the type conversion
	return fmt.Sprintf("%s(%s)", targetTypeCode, expr.Expression.String())
}

// Helper method to convert parser.Type to Go type string
func (cg *CodeGenerator) typeToGoString(t parser.Type) string {
	switch typ := t.(type) {
	case *parser.BasicType:
		return typ.Name
	case *parser.PointerType:
		return "*" + cg.typeToGoString(typ.ElementType)
	case *parser.NamedType:
		if typ.Package != "" {
			return fmt.Sprintf("%s.%s", typ.Package, typ.Name)
		}
		return typ.Name
	// Handle other type kinds...
	default:
		return "interface{}"
	}
}

func (cg *CodeGenerator) generateSelectorExpression(file *os.File, se *parser.SelectorExpression) {
	// Generate code for the left expression
	cg.generateExpression(file, se.Left)

	// Generate the dot
	fmt.Fprint(file, ".")

	// Generate the selector (method or field name)
	cg.generateExpression(file, se.Selector)
}

// generateInfixExpression generates Go code for an infix expression.
func (cg *CodeGenerator) generateInfixExpression(file *os.File, ie *parser.InfixExpression) {
	switch ie.Operator {
	case "+", "-", "*", "/", "%", "<", "<=", ">", ">=", "==":
		leftType := cg.getExpressionType(ie.Left)
		rightType := cg.getExpressionType(ie.Right)

		leftNumeric := cg.isNumericType(cg.analyzer.GetGoTypeFromParserType(leftType))
		rightNumeric := cg.isNumericType(cg.analyzer.GetGoTypeFromParserType(rightType))
		numeric := leftNumeric && rightNumeric

		isLeftString := leftType.String() == "string"
		isRightString := rightType.String() == "string"

		if isLeftString || isRightString {
			// If either side is a string, convert both sides to strings
			//fmt.Fprint(file, "(")
			cg.generateStringExpression(file, ie.Left)
			fmt.Fprintf(file, " %s ", ie.Operator)
			cg.generateStringExpression(file, ie.Right)
			//fmt.Fprint(file, ")")
			return
		} else if numeric {
			// Both sides are numeric, check if type casting is necessary
			castType := cg.getNumericCastType(cg.analyzer.GetGoTypeFromParserType(leftType), cg.analyzer.GetGoTypeFromParserType(rightType))
			//fmt.Fprint(file, "(")
			cg.generateNumericExpression(file, ie.Left, castType)
			fmt.Fprintf(file, " %s ", ie.Operator)
			cg.generateNumericExpression(file, ie.Right, castType)
			//fmt.Fprint(file, ")")
			return
		} else if leftNumeric || rightNumeric {
			// at least one side numeric, check if type casting is necessary
			castType := cg.getNumericCastType(cg.analyzer.GetGoTypeFromParserType(leftType), cg.analyzer.GetGoTypeFromParserType(rightType))
			//fmt.Fprint(file, "(")
			cg.generateNumericExpression(file, ie.Left, castType)
			fmt.Fprintf(file, " %s ", ie.Operator)
			cg.generateNumericExpression(file, ie.Right, castType)
			//fmt.Fprint(file, ")")
			return
		} else {
			// Handle other types without casting
			//fmt.Fprint(file, "(")
			cg.generateExpression(file, ie.Left)
			fmt.Fprintf(file, " %s ", ie.Operator)
			cg.generateExpression(file, ie.Right)
			//fmt.Fprint(file, ")")
			return
		}

	default:
		// Handle other operators
		cg.generateExpression(file, ie.Left)
		fmt.Fprintf(file, " %s ", ie.Operator)
		cg.generateExpression(file, ie.Right)
	}
}

func (cg *CodeGenerator) isNumericType(typ types.Type) bool {
	switch typ := typ.(type) {
	case *types.Basic:
		return typ.Info()&types.IsNumeric != 0
	default:
		return false
	}
}

func (cg *CodeGenerator) getNumericCastType(leftType, rightType types.Type) string {
	if types.Identical(leftType, rightType) {
		return leftType.String()
	}
	// Prioritize float over int
	if strings.Contains(leftType.String(), "float") || strings.Contains(rightType.String(), "float") {
		return "float64"
	}
	return "int"
}

func (cg *CodeGenerator) generateNumericExpression(file *os.File, expr parser.Expression, castType string) {
	exprType := cg.getExpressionType(expr)
	if exprType.String() != castType {
		cg.generateExpression(file, expr)
		switch expr.(type) {
		case *parser.Identifier:
			fmt.Fprintf(file, ".(%s)", castType)
		}
	} else {
		cg.generateExpression(file, expr)
	}
}

func (cg *CodeGenerator) generateStringExpression(file *os.File, expr parser.Expression) {
	fmt.Fprint(file, "fmt.Sprintf(\"%v\", ")
	cg.generateExpression(file, expr)
	fmt.Fprint(file, ")")
}

// getExpressionType retrieves the type of an expression from the symbol table.
func (cg *CodeGenerator) getExpressionType(expr parser.Expression) parser.Type {
	switch e := expr.(type) {
	case *parser.Identifier:
		symbol, found := cg.analyzer.CurrentTable.Resolve(e.Value)
		if !found {
			return &parser.BasicType{Name: "interface{}"}
		}
		return symbol.Type
	case *parser.IntegerLiteral:
		return &parser.BasicType{Name: "int"}
	case *parser.StringLiteral:
		return &parser.BasicType{Name: "string"}
	case *parser.BooleanLiteral:
		return &parser.BasicType{Name: "bool"}
	case *parser.CallExpression:
		// Handle call expressions accordingly
		if ident, ok := e.Function.(*parser.Identifier); ok {
			symbol, found := cg.analyzer.GlobalTable.Resolve(ident.Value)
			if found {
				if ft, ok := symbol.Type.(*parser.FunctionType); ok {
					if len(ft.ReturnTypes) > 0 {
						// Return the first return type for simplicity
						return ft.ReturnTypes[0]
					}
					// If there are no return types, return 'void'
					return &parser.BasicType{Name: "void"}
				}
			}
		}
		return &parser.BasicType{Name: "interface{}"}
	case *parser.IndexExpression:
		return &parser.BasicType{Name: "int"}
	case *parser.InfixExpression:
		return cg.getExpressionType(e.Left)
	case *parser.SelectorExpression:
		// Handle qualified identifiers (e.g., "math.Pi")
		if ident, ok := e.Left.(*parser.Identifier); ok {
			fqName := fmt.Sprintf("%s.%s", ident.Value, e.Selector.Value)
			symbol, ok := cg.analyzer.CurrentTable.Resolve(fqName)
			if ok {
				return symbol.Type
			}
			return &parser.BasicType{Name: "interface{}"}
		}
		return &parser.BasicType{Name: "interface{}"}
	default:
		return &parser.BasicType{Name: "interface{}"}
	}
}

// generatePrefixExpression generates Go code for a prefix expression.
func (cg *CodeGenerator) generatePrefixExpression(file *os.File, pe *parser.PrefixExpression) {
	fmt.Fprintf(file, "%s ", pe.Operator)
	cg.generateExpression(file, pe.Right)
	//fmt.Fprint(file, ")")
}

// generateCallExpression generates Go code for a function call.
func (cg *CodeGenerator) generateCallExpression(file *os.File, ce *parser.CallExpression) {
	switch ce.Function.(type) {
	case *parser.SelectorExpression:
		switch ce.Function.(*parser.SelectorExpression).Left.(type) {
		case *parser.Identifier:
			if cg.isImportedPackage(fmt.Sprintf("%s/%s", filepath.Base(cg.outputDir), ce.Function.(*parser.SelectorExpression).Left.(*parser.Identifier).Value)) {
				ce.Function.(*parser.SelectorExpression).Selector.Value = capitalize(ce.Function.(*parser.SelectorExpression).Selector.Value)
			}
		}
	}

	// Check if this CallExpression needs any wrappers
	wrappers, ok := cg.analyzer.WrapFunctionCalls[ce]
	if ok && len(wrappers) > 0 {
		// Create a copy of arguments to modify
		modifiedArgs := make([]parser.Expression, len(ce.Arguments))
		copy(modifiedArgs, ce.Arguments)

		// Apply wrappers to the specified arguments
		for _, wrapperInfo := range wrappers {
			if wrapperInfo.ArgIndex < len(modifiedArgs) {
				// Replace the argument with the wrapped version
				wrappedArg := &parser.Identifier{
					Value: wrapperInfo.Wrapper + "(" + modifiedArgs[wrapperInfo.ArgIndex].String() + ")",
				}
				modifiedArgs[wrapperInfo.ArgIndex] = wrappedArg
			}
		}

		// Generate the function call with wrapped arguments
		cg.generateExpression(file, ce.Function)
		fmt.Fprint(file, "(")
		for i, arg := range modifiedArgs {
			cg.generateExpression(file, arg)
			if i < len(modifiedArgs)-1 {
				fmt.Fprint(file, ", ")
			}
		}
		fmt.Fprint(file, ")")
		return
	}

	// Existing special cases (e.g., print, len)
	if ident, ok := ce.Function.(*parser.Identifier); ok {
		switch ident.Value {
		case "print":
			// Handle 'print' as a special case
			fmt.Fprint(file, "fmt.Println(")
			for i, arg := range ce.Arguments {
				cg.generateExpression(file, arg)
				if i < len(ce.Arguments)-1 {
					fmt.Fprint(file, ", ")
				}
			}
			fmt.Fprint(file, ")")
			return
		case "len":
			// Handle 'len' as a special case
			fmt.Fprint(file, "len(")
			for i, arg := range ce.Arguments {
				cg.generateExpression(file, arg)
				if i < len(ce.Arguments)-1 {
					fmt.Fprint(file, ", ")
				}
			}
			fmt.Fprint(file, ")")
			return
		case "make":
			switch ce.Arguments[0].(*parser.Identifier).Value {
			case "chan":
				fmt.Fprint(file, "make(chan any, ")
				for i, arg := range ce.Arguments[1:] {
					cg.generateExpression(file, arg)
					if i < len(ce.Arguments[1:])-1 {
						fmt.Fprint(file, ", ")
					}
				}
				fmt.Fprint(file, ")")
				return
			}
		}
	}

	//if se, ok := ce.Function.(*parser.SelectorExpression); ok {
	//	switch se.Left.(type) {
	//	case *parser.CallExpression:
	//		cg.generateCallExpression(file, se.Left.(*parser.CallExpression))
	//		return
	//	}
	//}

	// Handle generic function calls
	// Infer the function type
	//funcType := cg.analyzer.InferExpressionTypes(ce.Function, true)[0]
	//var paramTypes []parser.Type
	//if ft, ok := funcType.(*parser.FunctionType); ok {
	//	paramTypes = ft.ParameterTypes
	//}

	cg.generateExpression(file, ce.Function)
	fmt.Fprint(file, "(")
	for i, arg := range ce.Arguments {
		//argType := cg.analyzer.InferExpressionTypes(arg, true)[0]
		//var expectedType parser.Type
		//if i < len(paramTypes) {
		//	expectedType = paramTypes[i]
		//} else {
		//	expectedType = &parser.BasicType{Name: "interface{}"}
		//}

		//needsConversion, conversionFunc := cg.needsTypeConversion(argType, expectedType)
		//if needsConversion {
		//	fmt.Fprint(file, conversionFunc+"(")
		//	cg.generateExpression(file, arg)
		//	fmt.Fprint(file, ")")
		//} else {
		if arg == nil {
			fmt.Fprint(file, "nil")
		} else {
			cg.generateExpression(file, arg)
			switch arg.(type) {
			case *parser.Identifier:
				isInterface := false
				var castType string
				symbol, exsts := cg.analyzer.CurrentTable.Resolve(arg.String())
				if exsts {
					argType := symbol.Type.String()
					count := 0
					for _, obj := range cg.analyzer.Objects {
						if mapObj, exists := obj[arg.String()]; exists {
							if mapObj["function"] == cg.analyzer.CurrentTable.Name {
								if (argType == "interface{}" || argType == "any") && (mapObj["type"] != "interface{}" && mapObj["type"] != "any") {
									castType = mapObj["type"]
									count++
								}
								if count > 1 {
									isInterface = true
									break
									// report warning, potentially calling a go method with an argument that is not the correct type.
								}
							}
						}
					}
					if !isInterface && castType != "" {
						fmt.Fprintf(file, ".(%s)", castType)
					}
				}

			}
		}
		//}

		if i < len(ce.Arguments)-1 {
			fmt.Fprint(file, ", ")
		}
	}
	fmt.Fprint(file, ")")
}

func (cg *CodeGenerator) needsTypeConversion(argType, expectedType parser.Type) (bool, string) {
	var argTypeName string
	if argType == nil {
		argTypeName = "nil"
	} else {
		argTypeName = argType.String()
	}
	expectedTypeName := expectedType.String()

	// Handle common conversions
	if (argTypeName == "string" || argTypeName == "interface{}") && expectedTypeName == "[]byte" {
		return true, "[]byte"
	}
	if argTypeName == "int" && expectedTypeName == "float64" {
		return true, "float64"
	}
	// Add more conversion rules as needed

	// No conversion needed
	return false, ""
}

// generateBlockStatement generates Go code for a block of statements.
func (cg *CodeGenerator) generateBlockStatement(file *os.File, block *parser.BlockStatement, prevSymbolTable *semantic.SymbolTable) {
	if block != nil {
		for _, stmt := range block.Statements {
			cg.generateStatement(file, stmt, prevSymbolTable)
		}
	}
}

// generateIfStatement generates Go code for an if statement.
func (cg *CodeGenerator) generateIfStatement(file *os.File, is *parser.IfStatement, prevSymbolTable *semantic.SymbolTable) {
	cg.writeIndent(file)
	fmt.Fprint(file, "if ")
	cg.generateExpression(file, is.Condition)
	fmt.Fprintln(file, " {")
	cg.indentLevel++
	cg.generateBlockStatement(file, is.Consequence, prevSymbolTable)
	cg.indentLevel--
	cg.writeIndent(file)
	if is.Alternative != nil {
		fmt.Fprintln(file, "} else {")
		cg.indentLevel++
		cg.generateBlockStatement(file, is.Alternative, prevSymbolTable)
		cg.indentLevel--
		cg.writeIndent(file)
		fmt.Fprintln(file, "}")
	} else {
		fmt.Fprintln(file, "}")
	}
}

// generateWhileStatement generates Go code for a while loop.
func (cg *CodeGenerator) generateWhileStatement(file *os.File, ws *parser.WhileStatement, prevSymbolTable *semantic.SymbolTable) {
	cg.writeIndent(file)
	fmt.Fprint(file, "for ")
	cg.generateExpression(file, ws.Condition)
	switch ws.Condition.(type) {
	case *parser.InfixExpression:
		switch ws.Condition.(*parser.InfixExpression).Left.(type) {
		case *parser.Identifier:
			symbol, _ := cg.analyzer.CurrentTable.Resolve(ws.Condition.(*parser.InfixExpression).Left.(*parser.Identifier).Value)
			symbol.Metadata = map[string]any{"set": true}
		}
		switch ws.Condition.(*parser.InfixExpression).Right.(type) {
		case *parser.Identifier:
			symbol, _ := cg.analyzer.CurrentTable.Resolve(ws.Condition.(*parser.InfixExpression).Right.(*parser.Identifier).Value)
			symbol.Metadata = map[string]any{"set": true}
		}
	}
	fmt.Fprintln(file, " {")
	cg.indentLevel++
	cg.generateBlockStatement(file, ws.Body, prevSymbolTable)
	cg.indentLevel--
	cg.writeIndent(file)
	fmt.Fprintln(file, "}")
}

// generateForStatement generates Go code for a for loop.
func (cg *CodeGenerator) generateForStatement(file *os.File, fs *parser.ForStatement, prevSymbolTable *semantic.SymbolTable) {
	cg.writeIndent(file)
	switch fs.Iterable.(type) {
	case *parser.IntegerLiteral:
		fmt.Fprintf(file, "for %s := range ", fs.Variable.Value)
	case *parser.ArrayLiteral:
		fmt.Fprintf(file, "for _, %s := range ", fs.Variable.Value)
	case *parser.Identifier:
		symbol, _ := cg.analyzer.CurrentTable.Resolve(fs.Iterable.(*parser.Identifier).Value)
		switch st := symbol.Type.(type) {
		case *parser.BasicType:
			switch symbol.Type.(*parser.BasicType).Name {
			case "int":
				if fs.Variable.Value == "_" {
					fmt.Fprint(file, "for _ = range ")
				} else {
					fmt.Fprintf(file, "for %s := range ", fs.Variable.Value)
				}
			case "[]any":
				fmt.Fprintf(file, "for _, %s := range ", fs.Variable.Value)
			case "chan any":
				fmt.Fprintf(file, "for %s := range ", fs.Variable.Value)
			default:
				if strings.Contains(st.Name, "map") {
					fmt.Fprintf(file, "for %s, _ := range ", fs.Variable.Value)
				} else {
					fmt.Fprintf(file, "for _, %s := range ", fs.Variable.Value)
				}
			}
		}
		symbol.Metadata = map[string]any{"set": true}

	default:
		fmt.Fprintf(file, "for %s, _ := range ", fs.Variable.Value)
	}

	cg.generateExpression(file, fs.Iterable)
	fmt.Fprintln(file, " {")
	cg.indentLevel++
	cg.generateBlockStatement(file, fs.Body, prevSymbolTable)
	cg.indentLevel--
	cg.writeIndent(file)
	fmt.Fprintln(file, "}")
}

// writeIndent writes indentation.
func (cg *CodeGenerator) writeIndent(file *os.File) {
	for i := 0; i < cg.indentLevel; i++ {
		fmt.Fprint(file, "\t")
	}
}
