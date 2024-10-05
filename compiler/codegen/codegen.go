package codegen

import (
	"fmt"
	"os"
	"path/filepath"
	"simple/parser"
	"simple/semantic"
	"strings"
)

// CodeGenerator generates Go code from the AST.
type CodeGenerator struct {
	outputDir   string
	imports     map[string]bool
	indentLevel int
	analyzer    *semantic.Analyzer
	Returns     map[string]map[string]bool
}

func NewCodeGenerator(outputDir string, analyzer *semantic.Analyzer) *CodeGenerator {
	return &CodeGenerator{
		outputDir:   outputDir,
		imports:     make(map[string]bool),
		indentLevel: 0,
		analyzer:    analyzer,
		Returns:     make(map[string]map[string]bool),
	}
}

// GenerateCode generates Go code from the program.
func (cg *CodeGenerator) GenerateCode(program *parser.Program) error {
	mainFilePath := filepath.Join(cg.outputDir, "main.go")
	mainFile, err := os.Create(mainFilePath)
	if err != nil {
		return err
	}
	defer mainFile.Close()

	fmt.Fprintln(mainFile, "package main\n")

	// Collect imports
	cg.collectImports(program)

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
			cg.generateFunction(mainFile, stmt.(*parser.FunctionLiteral))
		}
	}

	// Generate main function
	fmt.Fprintln(mainFile, "func main() {")
	cg.indentLevel++
	for _, stmt := range program.Statements {
		if _, ok := stmt.(*parser.FunctionLiteral); !ok {
			cg.generateStatement(mainFile, stmt)
		}
	}
	cg.indentLevel--
	fmt.Fprintln(mainFile, "}")

	return nil
}

// collectImports collects imports from the program.
func (cg *CodeGenerator) collectImports(program *parser.Program) {
	for _, stmt := range program.Statements {
		if imp, ok := stmt.(*parser.ImportStatement); ok {
			module := strings.Trim(imp.ImportedModule.Value, "\"")
			cg.imports[module] = true
		}
	}

	// If 'print' is used, add 'fmt' to imports
	if cg.isBuiltinUsed("print", program) {
		cg.imports["fmt"] = true
	}
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
func (cg *CodeGenerator) generateFunction(file *os.File, fn *parser.FunctionLiteral) {
	// Get the function symbol from the symbol table
	symbol, ok := cg.analyzer.GlobalTable.Resolve(fn.Name.Value)
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
			switch functionType.ParameterTypes[i].(type) {
			case *parser.NamedType:
				paramType = strings.Split(functionType.ParameterTypes[i].String(), "/")[len(strings.Split(functionType.ParameterTypes[i].String(), "/"))-1]
			case *parser.PointerType:
				paramType = "*" + strings.Split(functionType.ParameterTypes[i].String(), "/")[len(strings.Split(functionType.ParameterTypes[i].String(), "/"))-1]
			case *parser.BasicType:
				paramType = functionType.ParameterTypes[i].String()
			}

			// Check if paramType includes a package prefix
			//if strings.Contains(paramType, ".") {
			//	parts := strings.Split(paramType, ".")
			//	if len(parts) == 2 {
			//		pkgName, typeName := parts[0], parts[1]
			//		//cg.imports[pkgName] = true
			//		paramType = fmt.Sprintf("%s.%s", pkgName, typeName)
			//	}
			//}
		}
		params = append(params, fmt.Sprintf("%s %s", p.Value, paramType))
	}

	// Determine return type
	returnType := ""
	cg.Returns["currentFunc"] = map[string]bool{"expects": false, "done": false}

	if functionType.ReturnType.TypeName() != "void" {
		returnType = functionType.ReturnType.String()
		cg.Returns["currentFunc"]["expects"] = true
	}

	if returnType != "" {
		fmt.Fprintf(file, "func %s(%s) %s {\n", fn.Name.Value, strings.Join(params, ", "), returnType)
	} else {
		fmt.Fprintf(file, "func %s(%s) {\n", fn.Name.Value, strings.Join(params, ", "))
	}
	cg.indentLevel++
	cg.generateBlockStatement(file, fn.Body)
	cg.indentLevel--
	if returnType != "" {
		if cg.Returns["currentFunc"]["expects"] && cg.Returns["currentFunc"]["done"] {
			fmt.Fprintf(file, "}\n")
		} else {
			fmt.Fprintf(file, "return interface{}(0)}\n", returnType)
		}

	} else {
		fmt.Fprintln(file, "}\n")
	}
	fmt.Fprintln(file) // Add an empty line for readability
	cg.Returns["currentFunc"]["expects"] = false
	cg.Returns["currentFunc"]["done"] = false
}

func (cg *CodeGenerator) generateAssignmentStatement(file *os.File, as *parser.AssignmentStatement) {
	cg.writeIndent(file)
	// Get the variable's type from the symbol table
	symbol, found := cg.analyzer.CurrentTable.Resolve(as.Name.Value)
	if !found {
		symbol, found = cg.analyzer.GlobalTable.Resolve(as.Name.Value)
		if !found {
			fmt.Fprintf(os.Stderr, "Undefined variable: %s\n", as.Name.Value)
			return
		}

	}
	if symbol.Metadata == nil {
		//varType := symbol.Type.String()
		fmt.Fprintf(file, "%s := ", as.Name.Value)
		cg.generateExpression(file, as.Value)
		fmt.Fprintln(file)
		symbol.Metadata = map[string]any{"set": true}
	} else {
		fmt.Fprintf(file, "%s = ", as.Name.Value)
		cg.generateExpression(file, as.Value)
		fmt.Fprintln(file)
		symbol.Metadata = map[string]any{"set": true}
	}

}

// generateStatement generates Go code for a statement.
func (cg *CodeGenerator) generateStatement(file *os.File, stmt parser.Statement) {
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
		cg.generateIfStatement(file, s)
	case *parser.WhileStatement:
		cg.generateWhileStatement(file, s)
	case *parser.ForStatement:
		cg.generateForStatement(file, s)
	default:
		// Handle other statements as needed
	}
}

// generateExpression generates Go code for an expression.
func (cg *CodeGenerator) generateExpression(file *os.File, expr parser.Expression) {
	switch e := expr.(type) {
	case *parser.Identifier:
		fmt.Fprint(file, e.Value)
	case *parser.IntegerLiteral:
		fmt.Fprint(file, e.Value)
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
		fmt.Fprint(file, e.Left.String())
		fmt.Fprint(file, "[")
		fmt.Fprint(file, e.Index.String())
		fmt.Fprint(file, "]")
	default:
		// Handle other expressions as needed
	}
}

func (cg *CodeGenerator) generateArrayLiteral(file *os.File, arr *parser.ArrayLiteral) {
	fmt.Fprint(file, "[]any{")
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

// In codegen/codegen.go

func (cg *CodeGenerator) generateSelectorExpression(file *os.File, se *parser.SelectorExpression) {
	// Generate code for the left expression
	cg.generateExpression(file, se.Left)

	// Generate the dot
	fmt.Fprint(file, ".")

	// Generate the selector (method or field name)
	cg.generateExpression(file, se.Selector)
}

func (cg *CodeGenerator) isNumeric(str string, slice []string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}

// generateInfixExpression generates Go code for an infix expression.
func (cg *CodeGenerator) generateInfixExpression(file *os.File, ie *parser.InfixExpression) {
	switch ie.Operator {
	case "+", "-", "*", "/", "%", "<", "<=", ">", ">=", "==":
		leftType := cg.getExpressionType(ie.Left)
		rightType := cg.getExpressionType(ie.Right)

		numeric := cg.isNumeric(leftType.String(), []string{"int", "float"})
		numeric = cg.isNumeric(rightType.String(), []string{"int", "float"})

		if !numeric {
			// If either side is not an int, convert both sides to strings
			fmt.Fprint(file, "(")

			// Convert left side to string
			switch left := ie.Left.(type) {
			case *parser.Identifier:
				fmt.Fprintf(file, "fmt.Sprintf(\"%%v\", %s)", left.Value)
			case *parser.InfixExpression:
				cg.generateInfixExpression(file, left)
			default:
				fmt.Fprint(file, "fmt.Sprintf(\"%v\", ")
				cg.generateExpression(file, ie.Left)
				fmt.Fprint(file, ")")
			}

			fmt.Fprint(file, " + ")

			// Convert right side to string
			switch right := ie.Right.(type) {
			case *parser.Identifier:
				fmt.Fprintf(file, "fmt.Sprintf(\"%%v\", %s)", right.Value)
			case *parser.InfixExpression:
				cg.generateInfixExpression(file, right)
			default:
				fmt.Fprint(file, "fmt.Sprintf(\"%v\", ")
				cg.generateExpression(file, ie.Right)
				fmt.Fprint(file, ")")
			}

			fmt.Fprint(file, ")")
			return
		}

		if leftType.String() == "float" || rightType.String() == "float" {
			// Handle other type combinations as needed
			fmt.Fprintf(file, "float64(")
			cg.generateExpression(file, ie.Left)
			fmt.Fprintf(file, ") %s ", ie.Operator)

			//switch ie.Left.(type) {
			//case *parser.IndexExpression:
			//	fmt.Fprintf(file, ") %s ", leftType.String(), ie.Operator)
			//case *parser.Identifier:
			//	fmt.Fprintf(file, ")  %s ", ie.Operator)
			//default:
			//	fmt.Fprintf(file, ") %s ", ie.Operator)
			//}

			fmt.Fprintf(file, "float64(")
			cg.generateExpression(file, ie.Right)
			fmt.Fprintf(file, ")")
			//switch ie.Right.(type) {
			//case *parser.IndexExpression:
			//	fmt.Fprintf(file, ".(float64)")
			//case *parser.Identifier:
			//	fmt.Fprintf(file, ".(float64)")
			//default:
			//	fmt.Fprintf(file, ".(float64)")
			//}
		} else {
			// Handle other type combinations as needed
			fmt.Fprintf(file, "int(")
			cg.generateExpression(file, ie.Left)
			fmt.Fprintf(file, ") %s ", ie.Operator)
			//switch ie.Left.(type) {
			//case *parser.IndexExpression:
			//	fmt.Fprintf(file, ".(int64) %s ", ie.Operator)
			//case *parser.Identifier:
			//	fmt.Fprintf(file, ".(int64) %s ", ie.Operator)
			//default:
			//	fmt.Fprintf(file, ".(int64) %s ", ie.Operator)
			//}

			fmt.Fprintf(file, "int(")
			cg.generateExpression(file, ie.Right)
			fmt.Fprintf(file, ")")
			//switch ie.Right.(type) {
			//case *parser.IndexExpression:
			//	fmt.Fprintf(file, ".(int64)")
			//case *parser.Identifier:
			//	fmt.Fprintf(file, ".(int64)")
			//default:
			//	fmt.Fprintf(file, ".(int64)")
			//}
			return
		}

	default:
		// Handle other operators
		cg.generateExpression(file, ie.Left)
		fmt.Fprintf(file, " %s ", ie.Operator)
		cg.generateExpression(file, ie.Right)
	}
}

// getExpressionType retrieves the type of an expression from the symbol table.
func (cg *CodeGenerator) getExpressionType(expr parser.Expression) parser.Type {
	switch e := expr.(type) {
	case *parser.Identifier:
		symbol, found := cg.analyzer.CurrentTable.Resolve(e.Value)
		if !found {
			symbol, found = cg.analyzer.GlobalTable.Resolve(e.Value)
			if !found {
				return &parser.BasicType{Name: "interface{}"}
			}
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
					return ft.ReturnType
				}
			}
		}
		return &parser.BasicType{Name: "interface{}"}
	case *parser.IndexExpression:
		return &parser.BasicType{Name: "int"}
	case *parser.InfixExpression:
		return cg.getExpressionType(e.Left)
	default:
		return &parser.BasicType{Name: "interface{}"}
	}
}

// generatePrefixExpression generates Go code for a prefix expression.
func (cg *CodeGenerator) generatePrefixExpression(file *os.File, pe *parser.PrefixExpression) {
	fmt.Fprintf(file, "(%s", pe.Operator)
	cg.generateExpression(file, pe.Right)
	fmt.Fprint(file, ")")
}

// generateCallExpression generates Go code for a function call.
func (cg *CodeGenerator) generateCallExpression(file *os.File, ce *parser.CallExpression) {
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
	funcType := cg.analyzer.InferExpressionType(ce.Function, true)
	var paramTypes []parser.Type
	if ft, ok := funcType.(*parser.FunctionType); ok {
		paramTypes = ft.ParameterTypes
	}

	cg.generateExpression(file, ce.Function)
	fmt.Fprint(file, "(")
	for i, arg := range ce.Arguments {
		argType := cg.analyzer.InferExpressionType(arg, true)
		var expectedType parser.Type
		if len(paramTypes) > 0 {
			expectedType = paramTypes[i]
		} else {
			expectedType = &parser.BasicType{Name: "interface{}"}
		}

		needsConversion, conversionFunc := cg.needsTypeConversion(argType, expectedType)
		if needsConversion {
			fmt.Fprint(file, conversionFunc+"(")
			cg.generateExpression(file, arg)
			fmt.Fprint(file, ")")
		} else {
			if arg == nil {
				fmt.Fprint(file, "nil")
			} else {
				cg.generateExpression(file, arg)
			}
		}

		if i < len(ce.Arguments)-1 {
			fmt.Fprint(file, ", ")
		}
	}
	fmt.Fprint(file, ")")
}

func (cg *CodeGenerator) needsTypeConversion(argType, expectedType parser.Type) (bool, string) {
	argTypeName := argType.String()
	expectedTypeName := expectedType.String()

	// Handle common conversions
	if argTypeName == "string" && expectedTypeName == "[]byte" {
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
func (cg *CodeGenerator) generateBlockStatement(file *os.File, block *parser.BlockStatement) {
	if block != nil {
		for _, stmt := range block.Statements {
			cg.generateStatement(file, stmt)
		}
	}
}

// generateIfStatement generates Go code for an if statement.
func (cg *CodeGenerator) generateIfStatement(file *os.File, is *parser.IfStatement) {
	cg.writeIndent(file)
	fmt.Fprint(file, "if ")
	cg.generateExpression(file, is.Condition)
	fmt.Fprintln(file, " {")
	cg.indentLevel++
	cg.generateBlockStatement(file, is.Consequence)
	cg.indentLevel--
	cg.writeIndent(file)
	if is.Alternative != nil {
		fmt.Fprintln(file, "} else {")
		cg.indentLevel++
		cg.generateBlockStatement(file, is.Alternative)
		cg.indentLevel--
		cg.writeIndent(file)
		fmt.Fprintln(file, "}")
	} else {
		fmt.Fprintln(file, "}")
	}
}

// generateWhileStatement generates Go code for a while loop.
func (cg *CodeGenerator) generateWhileStatement(file *os.File, ws *parser.WhileStatement) {
	cg.writeIndent(file)
	fmt.Fprint(file, "for ")
	cg.generateExpression(file, ws.Condition)
	fmt.Fprintln(file, " {")
	cg.indentLevel++
	cg.generateBlockStatement(file, ws.Body)
	cg.indentLevel--
	cg.writeIndent(file)
	fmt.Fprintln(file, "}")
}

// generateForStatement generates Go code for a for loop.
func (cg *CodeGenerator) generateForStatement(file *os.File, fs *parser.ForStatement) {
	cg.writeIndent(file)
	switch fs.Iterable.(type) {
	case *parser.IntegerLiteral:
		fmt.Fprintf(file, "for %s := range ", fs.Variable.Value)
	case *parser.ArrayLiteral:
		fmt.Fprintf(file, "for _, %s := range ", fs.Variable.Value)
	case *parser.Identifier:
		symbol, _ := cg.analyzer.CurrentTable.Resolve(fs.Iterable.(*parser.Identifier).Value)
		switch symbol.Type.(type) {
		case *parser.BasicType:
			switch symbol.Type.(*parser.BasicType).Name {
			case "int":
				fmt.Fprintf(file, "for _, %s := range ", fs.Variable.Value)
			case "slice", "map":
				fmt.Fprintf(file, "for %s, _ := range ", fs.Variable.Value)
			}
		}

	default:
		fmt.Fprintf(file, "for %s, _ := range ", fs.Variable.Value)
	}

	cg.generateExpression(file, fs.Iterable)
	fmt.Fprintln(file, " {")
	cg.indentLevel++
	cg.generateBlockStatement(file, fs.Body)
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
