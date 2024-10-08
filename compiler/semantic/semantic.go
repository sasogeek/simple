package semantic

import (
	"fmt"
	"github.com/google/uuid"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/packages"
	"simple/parser"
	"strings"
)

// Symbol represents a symbol in the symbol table.
type Symbol struct {
	Name     string
	Type     parser.Type
	Scope    string // "global", "local", "builtin", "imported"
	GoType   types.Type
	Metadata map[string]any
}

// SymbolTable represents a symbol table with scope chaining.
type SymbolTable struct {
	Symbols map[string]*Symbol
	Outer   *SymbolTable
	Name    string
}

type SymbolTables struct {
	Tables map[string]*SymbolTable
}

// NewSymbolTable creates a new symbol table.
func NewSymbolTable(outer *SymbolTable, name string) *SymbolTable {
	return &SymbolTable{
		Symbols: make(map[string]*Symbol),
		Outer:   outer,
		Name:    name,
	}
}

// Define adds a symbol to the symbol table.
func (st *SymbolTable) Define(name string, symbol *Symbol) {
	st.Symbols[name] = symbol
}

// Resolve looks up a symbol in the symbol table chain.
func (st *SymbolTable) Resolve(name string) (*Symbol, bool) {
	symbol, ok := st.Symbols[name]
	if !ok && st.Outer != nil {
		return st.Outer.Resolve(name)
	}
	return symbol, ok
}

// ExternalInterface represents an interface from an external Go package
type ExternalInterface struct {
	Package     string
	Name        string
	MethodNames []string
	Methods     []*parser.FunctionType
}

// Analyzer performs semantic analysis on the AST.
type Analyzer struct {
	GlobalTable         *SymbolTable
	CurrentTable        *SymbolTable
	SymbolTables        *SymbolTables
	errors              []string
	importedPackages    map[string]*packages.Package
	WrapFunctionCalls   map[*parser.CallExpression][]WrapperInfo
	ExternalFuncs       map[string]*parser.FunctionType // key: "package.Func"
	ExternalInterfaces  map[string]*ExternalInterface
	ExpectedReturnTypes map[*parser.CallExpression]parser.Type
}

// NewAnalyzer creates a new semantic analyzer.
func NewAnalyzer() *Analyzer {
	global := NewSymbolTable(nil, "global")
	analyzer := &Analyzer{
		GlobalTable:         global,
		CurrentTable:        global,
		SymbolTables:        &SymbolTables{Tables: map[string]*SymbolTable{"global": global}},
		errors:              []string{},
		importedPackages:    make(map[string]*packages.Package),
		WrapFunctionCalls:   make(map[*parser.CallExpression][]WrapperInfo),
		ExternalFuncs:       make(map[string]*parser.FunctionType),
		ExternalInterfaces:  make(map[string]*ExternalInterface),
		ExpectedReturnTypes: make(map[*parser.CallExpression]parser.Type),
	}

	// Initialize built-in functions
	analyzer.initBuiltins()

	return analyzer
}

type WrapperInfo struct {
	ArgIndex int
	Wrapper  string
}

func (a *Analyzer) GetGoTypeFromParserType(pt parser.Type) types.Type {
	switch t := pt.(type) {
	case *parser.BasicType:
		switch t.Name {
		case "int":
			return types.Typ[types.Int]
		case "string":
			return types.Typ[types.String]
		case "bool":
			return types.Typ[types.Bool]
		case "void":
			return types.Typ[types.UnsafePointer] // Represents 'void' as an unsafe pointer
		default:
			// Attempt to resolve the type from the symbol table (e.g., imported types)
			symbol, ok := a.GlobalTable.Resolve(t.Name)
			if ok {
				return symbol.GoType
			}
			// Default to interface{} for unknown types
			return types.NewInterface(nil, nil)
		}
	case *parser.FunctionType:
		// Create a types.Signature for the function
		return a.createGoSignatureFromFunctionType(t)
	case *parser.InterfaceType:
		// Resolve the interface type
		symbol, ok := a.GlobalTable.Resolve(t.Name)
		if ok {
			return symbol.GoType
		}
		// Default to an empty interface if not found
		return types.NewInterface(nil, nil)
	case *parser.StructType:
		// Resolve the struct type
		symbol, ok := a.GlobalTable.Resolve(t.Name)
		if ok {
			return symbol.GoType
		}
		// Default to struct{} if not found
		return types.NewStruct(nil, nil)
	default:
		// Default to interface{} for unknown type kinds
		return types.NewInterface(nil, nil)
	}
}

func (a *Analyzer) createGoSignatureFromFunctionType(ft *parser.FunctionType) *types.Signature {
	// Create parameter list
	var params *types.Tuple
	if len(ft.ParameterTypes) > 0 {
		var paramVars []*types.Var
		for i, pt := range ft.ParameterTypes {
			paramType := a.GetGoTypeFromParserType(pt)
			if paramType == nil {
				a.errors = append(a.errors, fmt.Sprintf("Unknown parameter type: %s", pt.String()))
				return nil
			}
			paramName := fmt.Sprintf("param%d", i+1) // Assign generic names
			paramVars = append(paramVars, types.NewVar(token.NoPos, nil, paramName, paramType))
		}
		params = types.NewTuple(paramVars...)
	} else {
		params = types.NewTuple()
	}

	// Create result list
	var results *types.Tuple
	switch ft.ReturnType != nil {
	case true:
		if ft.ReturnType.TypeName() != "void" {
			returnType := a.GetGoTypeFromParserType(ft.ReturnType)
			if returnType == nil {
				a.errors = append(a.errors, fmt.Sprintf("Unknown return type: %s", ft.ReturnType.String()))
				return nil
			}
			results = types.NewTuple(types.NewVar(token.NoPos, nil, "ret", returnType))
		} else {
			results = types.NewTuple()
		}
	case false:
		results = types.NewTuple()
	}

	// Create the signature
	sig := types.NewSignatureType(nil, nil, nil, params, results, false)
	return sig
}

// Errors returns the list of semantic errors.
func (a *Analyzer) Errors() []string {
	return a.errors
}

// initBuiltins adds built-in functions to the global symbol table.
func (a *Analyzer) initBuiltins() {
	// Define the 'print' built-in function
	printFunctionType := &parser.FunctionType{
		ParameterTypes: []parser.Type{&parser.BasicType{Name: "interface{}"}},
		ReturnType:     &parser.BasicType{Name: "void"},
	}
	symbol := &Symbol{
		Name:   "print",
		Type:   printFunctionType,
		Scope:  "builtin",
		GoType: a.createGoSignatureFromFunctionType(printFunctionType),
	}
	a.GlobalTable.Define("print", symbol)

	// Add other built-in functions if needed
}

// Analyze performs semantic analysis on the AST node.
func (a *Analyzer) Analyze(node parser.Node, remainingStatements []parser.Statement) {
	switch n := node.(type) {
	case *parser.Program:
		if n != nil {
			for i, stmt := range n.Statements {
				a.Analyze(stmt, n.Statements[i+1:])
			}
		}
	case *parser.FunctionLiteral:
		if n != nil {
			a.handleFunctionLiteral(n)
		}
	case *parser.ExpressionStatement:
		if n != nil {
			a.Analyze(n.Expression, remainingStatements)
		}

	case *parser.CallExpression:
		if n != nil {
			a.handleCallExpression(n)
		}
	case *parser.AssignmentStatement:
		if n != nil {
			a.handleAssignmentStatement(n, remainingStatements)
		}
	case *parser.Identifier:
		if n != nil {
			a.handleIdentifier(n, false)
		}
	case *parser.IfStatement:
		if n != nil {
			a.Analyze(n.Condition, remainingStatements)
			a.Analyze(n.Consequence, remainingStatements)
			a.Analyze(n.Alternative, remainingStatements)
		}
	case *parser.WhileStatement:
		if n != nil {
			a.Analyze(n.Condition, remainingStatements)
			a.Analyze(n.Body, remainingStatements)
		}
	case *parser.ForStatement:
		if n != nil {
			a.Analyze(n.Iterable, remainingStatements)
			switch n.Iterable.(type) {
			case *parser.Identifier:
				symbol, found := a.CurrentTable.Resolve(n.Iterable.(*parser.Identifier).Value)
				if found {
					switch symbol.Type.(type) {
					case *parser.BasicType:
						a.CurrentTable.Define(n.Variable.Value, &Symbol{
							Name:  n.Variable.Value,
							Type:  &parser.BasicType{Name: "int"}, // Initial type
							Scope: "local",
						})
					default:
						a.CurrentTable.Define(n.Variable.Value, &Symbol{
							Name:  n.Variable.Value,
							Type:  &parser.BasicType{Name: "interface{}"}, // Initial type
							Scope: "local",
						})
					}
				} else {
					a.CurrentTable.Define(n.Variable.Value, &Symbol{
						Name:  n.Variable.Value,
						Type:  &parser.BasicType{Name: "interface{}"}, // Initial type
						Scope: "local",
					})
				}
			}
			a.Analyze(n.Body, remainingStatements)
		}
	case *parser.ReturnStatement:
		if n != nil {
			a.Analyze(n.ReturnValue, remainingStatements)
		}
	case *parser.BlockStatement:
		if n != nil {
			for i, stmt := range n.Statements {
				a.Analyze(stmt, n.Statements[i+1:])
			}
		}
	case *parser.ImportStatement:
		if n != nil {
			a.handleImportStatement(n)
		}
	default:
		// Handle other node types as needed
	}
}

// handleFunctionLiteral processes function definitions.
func (a *Analyzer) handleFunctionLiteral(fl *parser.FunctionLiteral) {
	// Initialize function type with parameter types and 'void' return type
	// and define the function in the global table
	paramTypes := make([]parser.Type, len(fl.Parameters))
	params := make([]parser.Identifier, len(fl.Parameters))
	for i := range fl.Parameters {
		paramTypes[i] = &parser.BasicType{Name: "interface{}"} // Initial type
		params[i] = *fl.Parameters[i]
		paramSymbol := &Symbol{
			Name:  fl.Parameters[i].Value,
			Type:  paramTypes[i],
			Scope: fl.Name.Value,
		}
		a.CurrentTable.Define(paramSymbol.Name, paramSymbol)
	}

	functionType := &parser.FunctionType{
		Parameters:     params,
		ParameterTypes: paramTypes,
		ReturnType:     &parser.BasicType{Name: "void"},
	}

	// Define the function symbol in the global table
	symbol := &Symbol{
		Name:  fl.Name.Value,
		Type:  functionType,
		Scope: a.CurrentTable.Name,
	}

	// Create a types.Signature for the function
	//sig := a.createGoSignatureFromFunctionType(functionType)
	//if sig == nil {
	//	a.errors = append(a.errors, fmt.Sprintf("Failed to create Go signature for function '%s'", fl.Name.Value))
	//	return
	//}
	//symbol.GoType = sig

	//a.GlobalTable.Define(fl.Name.Value, symbol)
	a.CurrentTable.Define(fl.Name.Value, symbol)

	// Create a new symbol table for the function scope
	prevTable := a.CurrentTable
	funcTable := NewSymbolTable(a.CurrentTable, fl.Name.Value)
	a.SymbolTables.Tables[fl.Name.Value] = funcTable
	a.CurrentTable = funcTable

	// Define function parameters in the new scope
	for i, param := range fl.Parameters {
		paramType := a.GetGoTypeFromParserType(paramTypes[i])
		if paramType == nil {
			a.errors = append(a.errors, fmt.Sprintf("Unknown parameter type for '%s'", param.Value))
			continue
		}
		a.CurrentTable.Define(param.Value, &Symbol{
			Name:   param.Value,
			Type:   paramTypes[i],
			Scope:  fl.Name.Value,
			GoType: paramType,
		})
	}

	// Analyze the function body
	a.Analyze(fl.Body, []parser.Statement{fl.Body})

	// Infer parameter types based on usage
	a.InferFunctionParameterTypes(fl, funcTable)

	// Infer return type based on return statements
	functionType.ReturnType = a.InferFunctionReturnType(fl.Body, funcTable)

	// Update the function's GoType based on inferred return type
	functionTypeInferred := a.createGoSignatureFromFunctionType(functionType)
	if functionTypeInferred == nil {
		a.errors = append(a.errors, fmt.Sprintf("Failed to infer Go signature for function '%s'", fl.Name.Value))
		return
	}
	symbol.GoType = functionTypeInferred

	// Restore the previous symbol table
	a.CurrentTable = prevTable
}

// InferFunctionParameterTypes Infers and updates parameter types based on their usage.
func (a *Analyzer) InferFunctionParameterTypes(fl *parser.FunctionLiteral, funcTable *SymbolTable) {
	prevTable := a.CurrentTable
	a.CurrentTable = funcTable

	// Traverse the function body to find how parameters are used
	parser.Inspect(fl.Body, func(n parser.Node) bool {
		switch expr := n.(type) {
		case *parser.InfixExpression:
			if expr.Operator == "+" {
				leftType := a.InferExpressionType(expr.Left, true)
				rightType := a.InferExpressionType(expr.Right, true)

				// Check if parameters are involved and update their types
				a.updateParameterType(fl, expr.Left, leftType)
				a.updateParameterType(fl, expr.Right, rightType)
			}
		}
		return true
	})

	a.CurrentTable = prevTable

	// Update the function's parameter types in the symbol table
	functionSymbol, _ := a.CurrentTable.Resolve(fl.Name.Value)
	functionType, ok := functionSymbol.Type.(*parser.FunctionType)
	if !ok {
		return
	}

	for i, param := range fl.Parameters {
		symbol, found := funcTable.Resolve(param.Value)
		if found {
			functionType.ParameterTypes[i] = symbol.Type
		}
	}
}

// updateParameterType updates the type of a parameter if it is an identifier.
func (a *Analyzer) updateParameterType(fl *parser.FunctionLiteral, expr parser.Expression, newType parser.Type) {
	if ident, ok := expr.(*parser.Identifier); ok {
		// Check if the identifier is a parameter
		for _, param := range fl.Parameters {
			if param.Value == ident.Value {
				symbol, found := a.CurrentTable.Resolve(ident.Value)
				if found {
					symbol.Type = newType
				}
				break
			}
		}
	}
}

// InferFunctionReturnType Infers the return type of a function based on its return statements.
func (a *Analyzer) InferFunctionReturnType(body *parser.BlockStatement, funcTable *SymbolTable) parser.Type {
	var returnTypes []parser.Type
	prevTable := a.CurrentTable
	a.CurrentTable = funcTable

	parser.Inspect(body, func(n parser.Node) bool {
		if retStmt, ok := n.(*parser.ReturnStatement); ok {
			if retStmt.ReturnValue != nil {
				retType := a.InferExpressionType(retStmt.ReturnValue, false)
				returnTypes = append(returnTypes, retType)
			} else {
				returnTypes = append(returnTypes, &parser.BasicType{Name: "void"})
			}
		}
		return true
	})

	a.CurrentTable = prevTable

	if len(returnTypes) == 0 {
		return &parser.BasicType{Name: "void"}
	}

	firstType := returnTypes[0]
	for _, t := range returnTypes[1:] {
		if t.String() != firstType.String() {
			return &parser.BasicType{Name: "interface{}"}
		}
	}

	return firstType
}

// handleAssignmentStatement processes variable assignments.
func (a *Analyzer) handleAssignmentStatement(as *parser.AssignmentStatement, remainingStatements []parser.Statement) {
	a.Analyze(as.Value, remainingStatements)
	// Infer the type of the value
	varType := a.InferExpressionType(as.Value, true)

	// Determine the scope based on the current symbol table
	scope := a.CurrentTable.Name
	//if a.CurrentTable == a.GlobalTable {
	//	scope = "global"
	//}
	// Assign the Inferred type to the variable

	// attempt to resolve it, if it doesn't exist, its the first time we're defining it.
	// if it does exist and the type is different, rename it.
	symbol, exists := a.CurrentTable.Resolve(as.Name.Value)
	if !exists {
		a.CurrentTable.Define(as.Name.Value, &Symbol{
			Name:  as.Name.Value,
			Type:  varType,
			Scope: scope,
		})
	} else {
		prevName := symbol.Name
		if symbol.Type.TypeName() != varType.TypeName() {
			vid := strings.Replace(uuid.NewString(), "-", "", -1)[:5]
			as.Name.Value = as.Name.Value + vid
			a.CurrentTable.Define(as.Name.Value, &Symbol{
				Name:  as.Name.Value,
				Type:  varType,
				Scope: scope,
			})
			// Update any references to the variable in the remaining statements
			for _, stmt := range remainingStatements {
				a.updateVariableReferences(stmt, prevName, as.Name.Value)
			}
		}

	}

	// Assign the Inferred type to the variable
	//a.GlobalTable.Define(as.Name.Value, &Symbol{
	//	Name:  as.Name.Value,
	//	Type:  varType,
	//	Scope: scope,
	//})
}

// updateVariableReferences updates references to a variable in a given statement to the new name.
func (a *Analyzer) updateVariableReferences(stmt parser.Statement, oldName, newName string) {
	switch n := stmt.(type) {
	case *parser.AssignmentStatement:
		if n != nil {
			if n.Name.Value == oldName {
				n.Name.Value = newName
			}
			a.updateVariableReferencesInExpression(n.Value, oldName, newName)
		}
	case *parser.ExpressionStatement:
		if n != nil {
			a.updateVariableReferencesInExpression(n.Expression, oldName, newName)
		}
	case *parser.IfStatement:
		if n != nil {
			a.updateVariableReferencesInExpression(n.Condition, oldName, newName)
			a.updateVariableReferences(n.Consequence, oldName, newName)
			if n.Alternative != nil {
				a.updateVariableReferences(n.Alternative, oldName, newName)
			}
		}
	case *parser.WhileStatement:
		if n != nil {
			a.updateVariableReferencesInExpression(n.Condition, oldName, newName)
			a.updateVariableReferences(n.Body, oldName, newName)
		}
	case *parser.ForStatement:
		if n != nil {
			a.updateVariableReferencesInExpression(n.Iterable, oldName, newName)
			a.updateVariableReferences(n.Body, oldName, newName)
			if n.Variable != nil && n.Variable.Value == oldName {
				n.Variable.Value = newName
			}
		}
	case *parser.ReturnStatement:
		if n != nil {
			if n.ReturnValue != nil {
				a.updateVariableReferencesInExpression(n.ReturnValue, oldName, newName)
			}
		}
	case *parser.BlockStatement:
		if n != nil {
			for _, s := range n.Statements {
				a.updateVariableReferences(s, oldName, newName)
			}
		}
	case *parser.FunctionLiteral:
		if n != nil {
			for _, param := range n.Parameters {
				if param.Value == oldName {
					param.Value = newName
				}
			}
			a.updateVariableReferences(n.Body, oldName, newName)
		}
	}
}

// updateVariableReferencesInExpression updates references to a variable in a given expression to the new name.
func (a *Analyzer) updateVariableReferencesInExpression(expr parser.Expression, oldName, newName string) {
	switch e := expr.(type) {
	case *parser.Identifier:
		if e.Value == oldName {
			e.Value = newName
		}
	case *parser.InfixExpression:
		a.updateVariableReferencesInExpression(e.Left, oldName, newName)
		a.updateVariableReferencesInExpression(e.Right, oldName, newName)
	case *parser.PrefixExpression:
		a.updateVariableReferencesInExpression(e.Right, oldName, newName)
	case *parser.CallExpression:
		a.updateVariableReferencesInExpression(e.Function, oldName, newName)
		for _, arg := range e.Arguments {
			a.updateVariableReferencesInExpression(arg, oldName, newName)
		}
	case *parser.IndexExpression:
		a.updateVariableReferencesInExpression(e.Left, oldName, newName)
		a.updateVariableReferencesInExpression(e.Index, oldName, newName)
	case *parser.SelectorExpression:
		a.updateVariableReferencesInExpression(e.Left, oldName, newName)
	case *parser.ArrayLiteral:
		for _, elem := range e.Elements {
			a.updateVariableReferencesInExpression(elem, oldName, newName)
		}
	case *parser.MapLiteral:
		for key, value := range e.Pairs {
			a.updateVariableReferencesInExpression(key, oldName, newName)
			a.updateVariableReferencesInExpression(value, oldName, newName)
		}
	case *parser.FunctionLiteral:
		for _, param := range e.Parameters {
			if param.Value == oldName {
				param.Value = newName
			}
		}
		a.updateVariableReferences(e.Body, oldName, newName)
	}
}

// handleCallExpression processes function calls.
func (a *Analyzer) handleCallExpression(ce *parser.CallExpression) {
	// Analyze the function being called
	funcType := a.InferExpressionType(ce.Function, true)
	switch funcType.(type) {
	case *parser.FunctionType:
		ft := funcType.(*parser.FunctionType)
		for i, arg := range ce.Arguments {
			a.Analyze(arg, []parser.Statement{})
			argType := a.InferExpressionType(arg, true)
			var prevType parser.Type
			if i < len(ft.ParameterTypes) {
				paramType := ft.ParameterTypes[i]
				if paramType.String() != "interface{}" {
					if argType.String() != paramType.String() {
						// Additional check: if paramType is an interface, check if argType implements it
						if a.doesTypeImplement(paramType, argType) {
							// If the argument type implements the parameter type interface,
							// and the argument is a function, we might need to wrap it
							if _, ok := argType.(*parser.FunctionType); ok {
								// Find the adapter needed
								wrapper := a.findAdapterForInterface(paramType)
								if wrapper != "" {
									// Record that this call expression's argument needs wrapping
									a.WrapFunctionCalls[ce] = append(a.WrapFunctionCalls[ce], WrapperInfo{
										ArgIndex: i,
										Wrapper:  wrapper,
									})
								}
							}
						} else {
							prevType = argType
							argType = paramType
							switch arg.(type) {
							case *parser.Identifier:
								//fmt.Println(prevType)
								switch prevType.(type) {
								case *parser.FunctionType:
									if symbol, ok := a.CurrentTable.Resolve(arg.String()); ok {
										//symbol.Type = argType
										for x, _ := range symbol.Type.(*parser.FunctionType).ParameterTypes {
											symbol.Type.(*parser.FunctionType).ParameterTypes[x] = a.ExternalFuncs[fmt.Sprintf("%s.%s", ce.Function.(*parser.SelectorExpression).Left, ce.Function.(*parser.SelectorExpression).Selector)].ParameterTypes[i].(*parser.FunctionType).ParameterTypes[x]
											a.CurrentTable.Define(arg.String(), symbol)
											//a.CurrentTable.Define(arg.String(), symbol)
											param := symbol.Type.(*parser.FunctionType).Parameters[x]
											//if _, exsts := a.CurrentTable.Resolve(param.Value); exsts {
											paramSymbol := &Symbol{
												Name: param.Value,
												Type: a.ExternalFuncs[fmt.Sprintf("%s.%s", ce.Function.(*parser.SelectorExpression).Left, ce.Function.(*parser.SelectorExpression).Selector)].ParameterTypes[i].(*parser.FunctionType).ParameterTypes[x],
											}
											a.CurrentTable.Define(param.Value, paramSymbol)
											//a.CurrentTable.Define(param.Value, paramSymbol)
											//}
										}
									}

								}

							}

						}
					}

				} else {
					// Adopt the argument type
					ft.ParameterTypes[i] = argType
				}
			}
		}
		switch len(ft.ParameterTypes) > len(ce.Arguments) {
		case true:
			for len(ft.ParameterTypes) > len(ce.Arguments) {
				ce.Arguments = append(ce.Arguments, nil)
			}

		}

	}

}

// doesTypeImplement checks if argType implements paramType interface
func (a *Analyzer) doesTypeImplement(paramType parser.Type, argType parser.Type) bool {
	// Retrieve the underlying go/types.Type for paramType
	var paramGoType types.Type
	switch pt := paramType.(type) {
	case *parser.InterfaceType:
		// Interface name may include package alias, e.g., "http.Handler"
		parts := strings.Split(pt.Name, ".")
		if len(parts) == 2 {
			interfaceName := parts[1]
			symbol, ok := a.GlobalTable.Resolve(interfaceName)
			if !ok {
				a.errors = append(a.errors, fmt.Sprintf("Interface '%s' not found", interfaceName))
				return false
			}
			paramGoType = symbol.GoType
		} else {
			// Interface name without package alias
			symbol, ok := a.GlobalTable.Resolve(pt.Name)
			if !ok {
				a.errors = append(a.errors, fmt.Sprintf("Interface '%s' not found", pt.Name))
				return false
			}
			paramGoType = symbol.GoType
		}
	default:
		// Currently, only handling interface types
		return false
	}

	// Retrieve the underlying go/types.Type for argType
	var argGoType types.Type
	switch at := argType.(type) {
	case *parser.FunctionType:
		// Create a types.Signature for the function
		sig := a.createGoSignatureFromFunctionType(at)
		if sig == nil {
			a.errors = append(a.errors, "Failed to create Go signature from FunctionType")
			return false
		}
		argGoType = sig
	case *parser.BasicType:
		// Map basic types
		switch strings.ToLower(at.Name) {
		case "int":
			argGoType = types.Typ[types.Int]
		case "string":
			argGoType = types.Typ[types.String]
		case "bool":
			argGoType = types.Typ[types.Bool]
		default:
			// Attempt to resolve the type from the symbol table (e.g., imported types)
			symbol, ok := a.GlobalTable.Resolve(at.Name)
			if ok {
				argGoType = symbol.GoType
			} else {
				// Default to interface{} for unknown types
				argGoType = types.Universe.Lookup("any").Type()
			}
		}
	default:
		// Handle other types as needed
		return false
	}

	// Ensure paramGoType is an interface
	paramInterface, ok := paramGoType.Underlying().(*types.Interface)
	if !ok {
		// paramType is not an interface
		a.errors = append(a.errors, fmt.Sprintf("Type '%s' is not an interface", paramType.String()))
		return false
	}

	// Use types.Implements
	if types.Implements(argGoType, paramInterface) {
		return true
	}

	// Also check if the pointer to argGoType implements the interface
	ptrArgGoType := types.NewPointer(argGoType)
	if types.Implements(ptrArgGoType, paramInterface) {
		return true
	}

	return false
}

// handleIdentifier processes identifier usage.
func (a *Analyzer) handleIdentifier(id *parser.Identifier, reportErrors bool) {
	// Resolve the identifier in the current and outer scopes
	_, found := a.CurrentTable.Resolve(id.Value)
	if !found {
		_, found = a.GlobalTable.Resolve(id.Value)
		if !found {
			if reportErrors {
				//a.errors = append(a.errors, fmt.Sprintf("Undefined identifier: %s", id.Value))
			}
			return
		}
	}
	// Optionally, use the symbol's type for further analysis
}

// InferExpressionType Infers the type of an expression.
func (a *Analyzer) InferExpressionType(expr parser.Expression, reportErrors bool) parser.Type {
	switch e := expr.(type) {
	case *parser.IntegerLiteral:
		switch strings.Contains(e.Token.Literal, ".") {
		case true:
			return &parser.BasicType{Name: "float"}
		case false:
			return &parser.BasicType{Name: "int"}
		}
		return &parser.BasicType{Name: "int"}
	case *parser.StringLiteral:
		return &parser.BasicType{Name: "string"}
	case *parser.BooleanLiteral:
		return &parser.BasicType{Name: "bool"}
	case *parser.ArrayLiteral:
		return &parser.BasicType{Name: "slice"}
	case *parser.MapLiteral:
		return &parser.BasicType{Name: "map"}
	case *parser.Identifier:
		symbol, found := a.CurrentTable.Resolve(e.Value)
		if !found {
			//symbol, found = a.GlobalTable.Resolve(e.Value)
			//if !found {
			//	if reportErrors {
			//a.errors = append(a.errors, fmt.Sprintf("Undefined identifier: %s", e.Value))
			//}
			return &parser.BasicType{Name: "interface{}"}
			//}
		}
		return symbol.Type
	case *parser.CallExpression:
		funcType := a.InferExpressionType(e.Function, reportErrors)
		if ft, ok := funcType.(*parser.FunctionType); ok {
			// Analyze arguments if needed
			for i, arg := range e.Arguments {
				argType := a.InferExpressionType(arg, reportErrors)
				expectedType := ft.ParameterTypes[i]
				if !a.AreTypesCompatible(argType, expectedType) {
					// Perform type conversion or report error
				}
			}
			return ft.ReturnType
		}
		if reportErrors {
			a.errors = append(a.errors, fmt.Sprintf("Cannot determine return type of function '%s'", e.Function.String()))
		}
		return &parser.BasicType{Name: "interface{}"}
	case *parser.InfixExpression:
		leftType := a.InferExpressionType(e.Left, reportErrors)
		rightType := a.InferExpressionType(e.Right, reportErrors)
		switch e.Operator {
		case "+", "-", "*", "/", "%", "<", "<=", ">", ">=", "==":
			switch leftType.String() {
			case "string":
				return &parser.BasicType{Name: "string"}
			case "int":
				switch rightType.String() {
				case "int":
					return &parser.BasicType{Name: "int"}
				case "float":
					return &parser.BasicType{Name: "float"}
				}
			}
			switch rightType.String() {
			case "string":
				return &parser.BasicType{Name: "string"}
			case "int":
				switch leftType.String() {
				case "int":
					return &parser.BasicType{Name: "int"}
				case "float":
					return &parser.BasicType{Name: "float"}
				}
			}
		default:
			return &parser.BasicType{Name: "interface{}"}
		}
		return &parser.BasicType{Name: "interface{}"}
	case *parser.PrefixExpression:
		rightType := a.InferExpressionType(e.Right, reportErrors)
		switch e.Operator {
		case "!":
			return &parser.BasicType{Name: "bool"}
		case "-":
			return rightType
		}
		return &parser.BasicType{Name: "interface{}"}
	case *parser.SelectorExpression:
		if pkgMethod, exists := a.GlobalTable.Symbols[fmt.Sprintf("%s.%s", expr.(*parser.SelectorExpression).Left, expr.(*parser.SelectorExpression).Selector)]; exists {
			functionType := pkgMethod.Type
			return functionType
		}
		leftType := a.InferExpressionType(e.Left, reportErrors)
		if leftType == nil {
			if reportErrors {
				a.errors = append(a.errors, fmt.Sprintf("Unknown type for expression: %s", e.Left.String()))
			}
			return nil
		}

		// Retrieve the Go type from leftType
		leftGoType := a.GetGoTypeFromParserType(leftType)
		if leftGoType == nil {
			if reportErrors {
				a.errors = append(a.errors, fmt.Sprintf("Cannot resolve type for %s", leftType.String()))
			}
			return nil
		}

		// Look up the method or field
		sel := e.Selector.Value
		obj, _, _ := types.LookupFieldOrMethod(leftGoType, true, a.packageScope(), sel)

		switch obj := obj.(type) {
		case *types.Func:
			// Method found
			sig := obj.Type().(*types.Signature)
			functionType := a.functionTypeFromSignature(sig)
			return functionType
		case *types.Var:
			// Field found
			fieldType := &parser.BasicType{Name: obj.Type().String()}
			return fieldType
		default:
			if reportErrors {
				//a.errors = append(a.errors, fmt.Sprintf("Unsupported selector type for '%s.%s'", leftType.String(), sel))
			}
			return nil
		}
	default:
		return &parser.BasicType{Name: "interface{}"}
	}
}

func (a *Analyzer) AreTypesCompatible(srcType, destType parser.Type) bool {
	srcGoType := a.GetGoTypeFromParserType(srcType)
	destGoType := a.GetGoTypeFromParserType(destType)

	if srcGoType == nil || destGoType == nil {
		return false
	}

	// Use types.AssignableTo to check compatibility
	return types.AssignableTo(srcGoType, destGoType)
}

func (a *Analyzer) packageScope() *types.Package {
	// Return the current package scope
	// You may need to adjust this based on your Analyzer's implementation
	return nil // Replace with actual package scope if available
}

func (a *Analyzer) functionTypeFromSignature(sig *types.Signature) *parser.FunctionType {
	paramTypes := []parser.Type{}
	for i := 0; i < sig.Params().Len(); i++ {
		param := sig.Params().At(i)
		paramTypes = append(paramTypes, &parser.BasicType{Name: param.Type().String()})
	}
	var returnType parser.Type = &parser.BasicType{Name: "void"}
	if sig.Results().Len() > 0 {
		result := sig.Results().At(0)
		returnType = &parser.BasicType{Name: result.Type().String()}
	}
	return &parser.FunctionType{
		ParameterTypes: paramTypes,
		ReturnType:     returnType,
	}
}

// InferFunctionType Infers the type of an anonymous function (if supported).
func (a *Analyzer) InferFunctionType(fl *parser.FunctionLiteral) parser.Type {
	// Similar to handleFunctionLiteral but for function expressions
	// For simplicity, return 'interface{}' here
	return &parser.BasicType{Name: "interface{}"}
}

// handleImportStatement processes import statements.
func (a *Analyzer) handleImportStatement(is *parser.ImportStatement) {
	modulePath := strings.Trim(is.ImportedModule.Value, "\"")
	if _, exists := a.importedPackages[modulePath]; exists {
		// Package already imported
		return
	}

	// Load the package using golang.org/x/tools/go/packages
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedDeps,
	}
	pkgs, err := packages.Load(cfg, modulePath)
	if err != nil || len(pkgs) == 0 {
		a.errors = append(a.errors, fmt.Sprintf("Failed to load package: %s", modulePath))
		return
	}

	pkg := pkgs[0]
	a.importedPackages[modulePath] = pkg

	a.extractExternalFunctions(pkg)
	a.extractExternalInterfaces(pkg)

	// Add exported functions and types to the symbol table
	scope := pkg.Types.Scope()
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		if obj == nil || !obj.Exported() {
			continue
		}

		switch obj := obj.(type) {
		case *types.Func:
			// Handle functions
			sig := obj.Type().(*types.Signature)
			paramTypes := []parser.Type{}
			for i := 0; i < sig.Params().Len(); i++ {
				param := sig.Params().At(i)
				paramTypes = append(paramTypes, &parser.BasicType{Name: param.Type().String()})
			}
			//var returnType parser.Type = &parser.BasicType{Name: "void"}
			//if sig.Results().Len() > 0 {
			//	// For simplicity, handle only the first return value
			//	result := sig.Results().At(0)
			//	returnType = &parser.BasicType{Name: result.Type().String()}
			//}
			functionType := a.functionTypeFromSignature(sig)
			symbol := &Symbol{
				Name:   pkg.Name + "." + name,
				Type:   functionType,
				Scope:  "imported",
				GoType: sig,
			}
			a.GlobalTable.Define(pkg.Name+"."+name, symbol)
		case *types.TypeName:
			// Handle structs and interfaces
			named, ok := obj.Type().(*types.Named)
			if !ok {
				continue
			}
			switch named.Underlying().(type) {
			case *types.Struct:
				structType := &parser.StructType{Name: named.Obj().Name()}
				symbol := &Symbol{
					Name:   named.Obj().Name(),
					Type:   structType,
					Scope:  "imported",
					GoType: named,
				}
				a.GlobalTable.Define(named.Obj().Name(), symbol)
			case *types.Interface:
				interfaceType := &parser.InterfaceType{Name: named.Obj().Name()}
				symbol := &Symbol{
					Name:   named.Obj().Name(),
					Type:   interfaceType,
					Scope:  "imported",
					GoType: named.Underlying(),
				}
				a.GlobalTable.Define(named.Obj().Name(), symbol)
			}
		}
	}
}

func (a *Analyzer) findAdapterForInterface(paramType parser.Type) string {
	switch pt := paramType.(type) {
	case *parser.InterfaceType:
		// Interface name may include package alias, e.g., "http.Handler"
		parts := strings.Split(pt.Name, ".")
		if len(parts) == 2 {
			pkgAlias, interfaceName := parts[0], parts[1]
			adapterFuncName := interfaceName + "Func"
			return fmt.Sprintf("%s.%s", pkgAlias, adapterFuncName)
		} else {
			// Interface name without package alias
			adapterFuncName := pt.Name + "Func"
			return adapterFuncName
		}
	default:
		return ""
	}
}

// extractExternalInterfaces extracts exported interfaces from a loaded package.
func (a *Analyzer) extractExternalInterfaces(pkg *packages.Package) {
	scope := pkg.Types.Scope()
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		if obj == nil {
			continue
		}

		// We are interested in interfaces only
		iface, ok := obj.Type().Underlying().(*types.Interface)
		if !ok {
			continue
		}

		// Ensure the interface is exported
		if !obj.Exported() {
			continue
		}

		// Extract method signatures
		methodNames := []string{}
		methods := []*parser.FunctionType{}
		for i := 0; i < iface.NumMethods(); i++ {
			method := iface.Method(i)
			methodName := method.Name()
			methodNames = append(methodNames, methodName)

			sig, ok := method.Type().(*types.Signature)
			if !ok {
				continue
			}

			// Extract parameter types
			paramTypes := []parser.Type{}
			params := sig.Params()
			for j := 0; j < params.Len(); j++ {
				param := params.At(j)
				paramTypes = append(paramTypes, a.convertGoType(param.Type()))
			}

			// Extract return types
			returnTypes := []parser.Type{}
			results := sig.Results()
			for j := 0; j < results.Len(); j++ {
				returnTypes = append(returnTypes, a.convertGoType(results.At(j).Type()))
			}

			methods = append(methods, &parser.FunctionType{
				ParameterTypes: paramTypes,
				ReturnType:     a.combineReturnTypes(returnTypes),
				//ReturnTypes: returnTypes,
			})
		}

		// Fully qualified interface name (e.g., "http.ResponseWriter")
		fqIfaceName := fmt.Sprintf("%s.%s", pkg.Name, name)

		// Populate the ExternalInterfaces map
		a.ExternalInterfaces[fqIfaceName] = &ExternalInterface{
			Package:     pkg.Name,
			Name:        name,
			MethodNames: methodNames,
			Methods:     methods,
		}
	}
}

// combineReturnTypes simplifies handling single vs multiple return types.
func (a *Analyzer) combineReturnTypes(returnTypes []parser.Type) parser.Type {
	if len(returnTypes) == 0 {
		return &parser.BasicType{Name: ""}
	} else if len(returnTypes) == 1 {
		return returnTypes[0]
	}
	// For multiple return types, you might define a TupleType or similar.
	// For simplicity, return interface{}.
	return &parser.BasicType{Name: "interface{}"}
}

// extractExternalFunctions extracts exported functions from a loaded package.
func (a *Analyzer) extractExternalFunctions(pkg *packages.Package) {
	scope := pkg.Types.Scope()
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		if obj == nil {
			continue
		}

		// We are interested in functions only
		funcObj, ok := obj.(*types.Func)
		if !ok {
			continue
		}

		// Ensure the function is exported
		if !funcObj.Exported() {
			continue
		}

		// Get the function signature
		sig, ok := funcObj.Type().(*types.Signature)
		if !ok {
			continue
		}

		// Extract parameter types
		params := sig.Params()
		paramTypes := []parser.Type{}
		for i := 0; i < params.Len(); i++ {
			param := params.At(i)
			paramTypes = append(paramTypes, a.convertGoType(param.Type()))
		}

		// Extract return types
		results := sig.Results()
		var returnType parser.Type
		if results.Len() > 0 {
			// For simplicity, handle single return type
			returnType = a.convertGoType(results.At(0).Type())
		}

		// Fully qualified function name (e.g., "http.ListenAndServe")
		fqFuncName := fmt.Sprintf("%s.%s", pkg.Name, funcObj.Name())

		// Populate the ExternalFuncs map
		a.ExternalFuncs[fqFuncName] = &parser.FunctionType{
			ParameterTypes: paramTypes,
			ReturnType:     returnType,
		}
	}
}

// convertGoType converts Go's types.Type to Simple's parser.Type.
func (a *Analyzer) convertGoType(goType types.Type) parser.Type {
	switch t := goType.(type) {
	case *types.Basic:
		return &parser.BasicType{Name: t.Name()}
	case *types.Pointer:
		elemType := a.convertGoType(t.Elem())
		return &parser.PointerType{ElementType: elemType}
	case *types.Named:
		obj := t.Obj()
		pkgPath := ""
		if obj.Pkg() != nil {
			pkgPath = obj.Pkg().Path()
		}
		return &parser.NamedType{
			Name:    obj.Name(),
			Package: pkgPath,
		}
	case *types.Signature:
		// Handle function types
		paramTypes := []parser.Type{}
		params := t.Params()
		for i := 0; i < params.Len(); i++ {
			param := params.At(i)
			paramTypes = append(paramTypes, a.convertGoType(param.Type()))
		}

		var returnType parser.Type
		results := t.Results()
		if results.Len() > 0 {
			// For simplicity, handle single return type
			returnType = a.convertGoType(results.At(0).Type())
		}

		return &parser.FunctionType{
			ParameterTypes: paramTypes,
			ReturnType:     returnType,
		}
	case *types.Slice:
		elemType := a.convertGoType(t.Elem())
		return &parser.ArrayType{ElementType: elemType}
	case *types.Map:
		keyType := a.convertGoType(t.Key())
		valueType := a.convertGoType(t.Elem())
		return &parser.MapType{KeyType: keyType, ValueType: valueType}
	default:
		// Handle other types as needed
		return &parser.BasicType{Name: "interface{}"}
	}
}
