package transformer

import (
	"fmt"
	"go/types"
	"simple/lexer"
	"simple/parser"
	"simple/semantic"
	"slices"
	"strings"
)

type Transformer struct {
	analyzer *semantic.Analyzer
}

func NewTransformer(analyzer *semantic.Analyzer) *Transformer {
	return &Transformer{
		analyzer: analyzer,
	}
}

func (t *Transformer) Transform(node parser.Node, rNode parser.Node) {
	switch n := node.(type) {
	case *parser.Program:
		for _, stmt := range n.Statements {
			t.Transform(stmt, rNode)
		}
	case *parser.ExpressionStatement:
		if n != nil {
			t.Transform(n.Expression, rNode)
		}
	case *parser.CallExpression:
		t.handleCallExpression(n, rNode)
	case *parser.FunctionLiteral:
		prevTable := t.analyzer.CurrentTable
		t.analyzer.CurrentTable = t.analyzer.SymbolTables.Tables[n.Name.Value]
		t.Transform(n.Body, rNode)
		t.analyzer.CurrentTable = prevTable
	case *parser.BlockStatement:
		for _, stmt := range n.Statements {
			t.Transform(stmt, rNode)
		}
	case *parser.AssignmentStatement:
		t.handleAssignmentStatement(n, rNode)
		// Handle other node types as needed
	case *parser.ReturnStatement:
		t.handleReturnStatement(n, rNode)
	}
}

func (t *Transformer) handleReturnStatement(rs *parser.ReturnStatement, rNode parser.Node) {
	// Transform the return value
	t.Transform(rs.ReturnValue, rNode)

	// Infer the type of the return value
	returnTypes := t.analyzer.InferExpressionTypes(rs.ReturnValue, true)

	// Update the enclosing function's return types
	enclosingFunc := t.analyzer.CurrentTable.Name
	funcSymbol, exists := t.analyzer.CurrentTable.Resolve(enclosingFunc)
	if exists {
		if funcType, ok := funcSymbol.Type.(*parser.FunctionType); ok {
			funcType.ReturnTypes = returnTypes
		}
	}

}

func (t *Transformer) handleAssignmentStatement(as *parser.AssignmentStatement, rNode parser.Node) {
	// Transform the RHS expression
	t.Transform(as.Value, rNode)

	// Infer the type(s) of the RHS expression(s)
	varTypes := t.analyzer.InferExpressionTypes(as.Value, true) // Returns []parser.Type
	if len(as.Value.String()) > 2 {
		if as.Value.String()[len(as.Value.String())-1:] == "}" {
			varTypes = []parser.Type{&parser.BasicType{Name: as.Value.String()[:strings.Index(as.Value.String(), "{")]}}
		}
	}
	// Update the symbol table and variable types
	for i, leftExpr := range as.Left {
		var currentVarType parser.Type
		if i < len(varTypes) {
			currentVarType = varTypes[i]
		} else {
			// Handle the case where there are fewer types than variables
			currentVarType = &parser.BasicType{Name: "any"}
		}

		switch expr := leftExpr.(type) {
		case *parser.Identifier:
			name := expr.Value
			symbol, exists := t.analyzer.CurrentTable.Resolve(name)
			if !exists {
				// Define the new variable in the symbol table
				t.analyzer.CurrentTable.Define(name, &semantic.Symbol{
					Name:  name,
					Type:  currentVarType,
					Scope: t.analyzer.CurrentTable.Name,
				})
				t.analyzer.Assignments[name] = map[string][]string{"types": []string{}}
				t.analyzer.Assignments[name]["types"] = append(t.analyzer.Assignments[name]["types"], currentVarType.String())
			} else {
				switch symbol.Type.(type) {
				case *parser.BasicType:
					switch currentVarType.(type) {
					case *parser.BasicType:
						if currentVarType.(*parser.BasicType).Name != symbol.Type.(*parser.BasicType).Name {
							symbol.Type = currentVarType
							t.updateFunctionParameterTypes(name, symbol.Type, rNode)
							if t.analyzer.Assignments[name]["types"] == nil {
								t.analyzer.Assignments[name] = map[string][]string{"types": []string{}}
								t.analyzer.Assignments[name]["types"] = append(t.analyzer.Assignments[name]["types"], currentVarType.String())
							} else {
								if !slices.Contains(t.analyzer.Assignments[name]["types"], currentVarType.String()) {
									t.analyzer.Assignments[name]["types"] = append(t.analyzer.Assignments[name]["types"], currentVarType.String())
								}
							}
							return
						}
					case *parser.ArrayType:
						if currentVarType.(*parser.ArrayType).ElementType.(*parser.BasicType).Name != symbol.Type.(*parser.BasicType).Name {
							symbol.Type = currentVarType
							t.updateFunctionParameterTypes(name, symbol.Type, rNode)
							if t.analyzer.Assignments[name]["types"] == nil {
								t.analyzer.Assignments[name] = map[string][]string{"types": []string{}}
								t.analyzer.Assignments[name]["types"] = append(t.analyzer.Assignments[name]["types"], currentVarType.String())
							} else {
								if !slices.Contains(t.analyzer.Assignments[name]["types"], currentVarType.String()) {
									t.analyzer.Assignments[name]["types"] = append(t.analyzer.Assignments[name]["types"], currentVarType.String())
								}
							}
							return
						}
					}
				case *parser.ArrayType:
					switch currentVarType.(type) {
					case *parser.BasicType:
						if currentVarType.(*parser.BasicType).Name != symbol.Type.(*parser.BasicType).Name {
							symbol.Type = currentVarType
							t.updateFunctionParameterTypes(name, symbol.Type, rNode)
							if t.analyzer.Assignments[name]["types"] == nil {
								t.analyzer.Assignments[name] = map[string][]string{"types": []string{}}
								t.analyzer.Assignments[name]["types"] = append(t.analyzer.Assignments[name]["types"], currentVarType.String())
							} else {
								if !slices.Contains(t.analyzer.Assignments[name]["types"], currentVarType.String()) {
									t.analyzer.Assignments[name]["types"] = append(t.analyzer.Assignments[name]["types"], currentVarType.String())
								}
							}
							return
						}
					case *parser.ArrayType:
						if currentVarType.(*parser.ArrayType).ElementType.(*parser.BasicType).Name != symbol.Type.(*parser.ArrayType).ElementType.(*parser.BasicType).Name {
							symbol.Type = currentVarType
							t.updateFunctionParameterTypes(name, symbol.Type, rNode)
							if t.analyzer.Assignments[name]["types"] == nil {
								t.analyzer.Assignments[name] = map[string][]string{"types": []string{}}
								t.analyzer.Assignments[name]["types"] = append(t.analyzer.Assignments[name]["types"], currentVarType.String())
							} else {
								if !slices.Contains(t.analyzer.Assignments[name]["types"], currentVarType.String()) {
									t.analyzer.Assignments[name]["types"] = append(t.analyzer.Assignments[name]["types"], currentVarType.String())
								}
							}
							return
						}
					}
				}
				// Update the symbol's type
				//symbol.Type = currentVarType
			}
			t.updateFunctionParameterTypes(name, currentVarType, rNode)
			// Handle other types of LHS expressions if needed
		}

		// Check if the LHS variable is used in function calls

	}
}

func (t *Transformer) updateFunctionParameterTypes(varName string, varType parser.Type, rNode parser.Node) {
	// Iterate over all function calls in the program
	parser.Inspect(rNode, func(node parser.Node) bool {
		callExpr, ok := node.(*parser.CallExpression)
		if !ok {
			return true
		}

		// Check if the variable is used as an argument
		for i, arg := range callExpr.Arguments {
			if ident, ok := arg.(*parser.Identifier); ok && ident.Value == varName {
				// Retrieve the function symbol
				funcName := callExpr.Function.String()
				funcSymbol, exists := t.analyzer.CurrentTable.Resolve(funcName)
				if !exists {
					continue
				}

				// Update the parameter type
				if funcType, ok := funcSymbol.Type.(*parser.FunctionType); ok {
					if i < len(funcType.ParameterTypes) {
						funcType.ParameterTypes[i] = varType
					}
				}
			}
		}
		return true
	})
}

func (t *Transformer) handleCallExpression(ce *parser.CallExpression, rNode parser.Node) {
	// Check if the function is a SelectorExpression (e.g., pkg.Func)
	switch ce.Function.(type) {
	case *parser.SelectorExpression:
		if selExpr, ok := ce.Function.(*parser.SelectorExpression); ok {

			pkgName := selExpr.Left.String()
			funcName := selExpr.Selector.String()
			fqFuncName := pkgName + "." + funcName
			if symbol, exists := t.analyzer.CurrentTable.Resolve(selExpr.Left.String()); exists {
				switch symbol.Type.(type) {
				case *parser.NamedType:
					pkgName = fmt.Sprintf("%s", strings.Split(symbol.Type.(*parser.NamedType).Package, "/")[len(strings.Split(symbol.Type.(*parser.NamedType).Package, "/"))-1])
					pkgFuncName := symbol.Type.(*parser.NamedType).Name
					funcSymbol, exsts := t.analyzer.CurrentTable.Resolve(pkgFuncName)
					if exsts {
						//var methods []interface{}
						var methodName string
						//var expectedType interface{}
						for i := range funcSymbol.GoType.(*types.Interface).NumMethods() {
							methodName = funcSymbol.GoType.(*types.Interface).Method(i).Name()
							methodSig := funcSymbol.GoType.(*types.Interface).Method(i).Signature()
							if methodName == funcName {
								for paramId := range ce.Arguments {
									expectedType := methodSig.Params().At(paramId)
									switch ce.Arguments[paramId].(type) {
									case *parser.StringLiteral:
										ce.Arguments[paramId].(*parser.StringLiteral).Value = fmt.Sprintf("%s(\"%s\")", expectedType.Type().String(), ce.Arguments[paramId].(*parser.StringLiteral).String())
									case *parser.Identifier:
										ce.Arguments[paramId].(*parser.Identifier).Value = fmt.Sprintf("%s(%s)", expectedType.Type().String(), ce.Arguments[paramId].(*parser.Identifier).String())
									case *parser.InfixExpression:
										stringValue := ""
										stringLiteral := &parser.StringLiteral{
											Token: lexer.Token{Type: lexer.TokenString},
											Value: stringValue,
										}
										stringLiteral.Value = t.expressionToString(ce.Arguments[paramId].(*parser.InfixExpression))
										stringLiteral.Value = fmt.Sprintf("%s(%s)", expectedType.Type().String(), stringLiteral.Value)
										ce.Arguments[paramId] = stringLiteral
									}
								}
							}

						}

					}
				case *parser.PointerType:
					pkgName = fmt.Sprintf("%s", strings.Split(symbol.Type.(*parser.PointerType).ElementType.(*parser.NamedType).Package, "/")[len(strings.Split(symbol.Type.(*parser.PointerType).ElementType.(*parser.NamedType).Package, "/"))-1])
					pkgFuncName := symbol.Type.(*parser.PointerType).ElementType.(*parser.NamedType).Name
					funcSymbol, exsts := t.analyzer.CurrentTable.Resolve(pkgFuncName)
					if exsts {
						//var methods []interface{}
						var methodName string
						//var expectedType interface{}
						switch funcSymbol.GoType.(type) {
						case *types.Named:
							f := false
							for i := range funcSymbol.GoType.(*types.Named).NumMethods() {
								methodName = funcSymbol.GoType.(*types.Named).Method(i).Name()
								methodSig := funcSymbol.GoType.(*types.Named).Method(i).Signature()
								if methodName == funcName {
									f = true
									for paramId := range ce.Arguments {
										expectedType := methodSig.Params().At(paramId)
										switch ce.Arguments[paramId].(type) {
										case *parser.StringLiteral:
											//ce.Arguments[paramId].(*parser.StringLiteral).Value = fmt.Sprintf("%s(\"%s\")", expectedType.Type().String(), ce.Arguments[paramId].(*parser.StringLiteral).String())
										case *parser.Identifier:
											ce.Arguments[paramId].(*parser.Identifier).Value = fmt.Sprintf("%s(%s)", expectedType.Type().String(), ce.Arguments[paramId].(*parser.Identifier).String())
										case *parser.InfixExpression:
											stringValue := ""
											stringLiteral := &parser.StringLiteral{
												Token: lexer.Token{Type: lexer.TokenString},
												Value: stringValue,
											}
											stringLiteral.Value = t.expressionToString(ce.Arguments[paramId].(*parser.InfixExpression))
											stringLiteral.Value = fmt.Sprintf("%s(%s)", expectedType.Type().String(), stringLiteral.Value)
											ce.Arguments[paramId] = stringLiteral
										}
									}
								}
							}

							if !f {
								switch structType := funcSymbol.GoType.(*types.Named).Underlying().(type) {
								case *types.Struct:
									for i := 0; i < structType.NumFields(); i++ {
										field := structType.Field(i)
										if field.Embedded() {
											// If it's an embedded field, recursively print its fields and methods
											embeddedType := field.Type()

											if namedEmbedded, ok := embeddedType.(*types.Named); ok {

												for i := range namedEmbedded.NumMethods() {
													methodName = namedEmbedded.Method(i).Name()
													methodSig := namedEmbedded.Method(i).Signature()
													if methodName == funcName {
														f = true
														for paramId := range ce.Arguments {
															expectedType := methodSig.Params().At(paramId)
															switch ce.Arguments[paramId].(type) {
															case *parser.StringLiteral:
																//ce.Arguments[paramId].(*parser.StringLiteral).Value = ce.Arguments[paramId].(*parser.StringLiteral).String()
															case *parser.Identifier:
																switch et := expectedType.Type().(type) {
																case *types.Slice:
																	switch et.Elem().Underlying().(type) {
																	case *types.Signature:
																		if fAsArgSymbol, found := t.analyzer.CurrentTable.Resolve(ce.Arguments[paramId].String()); found {
																			for argId, _ := range fAsArgSymbol.Type.(*parser.FunctionType).ParameterTypes {
																				switch et.Elem().Underlying().(*types.Signature).Params().At(argId).Type().(type) {
																				case *types.Pointer:
																					etType := et.Elem().Underlying().(*types.Signature).Params().At(0).Type().String()
																					etTypeString := strings.Split(etType, "/")[len(strings.Split(etType, "/"))-1]
																					fAsArgSymbol.Type.(*parser.FunctionType).ParameterTypes[argId] = &parser.BasicType{Name: fmt.Sprintf("*%s", etTypeString)}
																				}
																			}
																		}
																	}
																}

																//ce.Arguments[paramId].(*parser.Identifier).Value = fmt.Sprintf("%s(%s)", strings.Split(expectedType.Type().String(), "/")[len(strings.Split(expectedType.Type().String(), "/"))-1], ce.Arguments[paramId].(*parser.Identifier).String())
															case *parser.InfixExpression:
																stringValue := ""
																stringLiteral := &parser.StringLiteral{
																	Token: lexer.Token{Type: lexer.TokenString},
																	Value: stringValue,
																}
																stringLiteral.Value = t.expressionToString(ce.Arguments[paramId].(*parser.InfixExpression))
																stringLiteral.Value = fmt.Sprintf("%s(%s)", expectedType.Type().String(), stringLiteral.Value)
																ce.Arguments[paramId] = stringLiteral
															}
														}
													}
												}

											}
										}
									}
								}

							}

						}

					}
				}
			}

			//pkgName := selExpr.Left.String()
			//funcName := selExpr.Selector.String()
			//fqFuncName := pkgName + "." + funcName

			if extFuncType, exists := t.analyzer.ExternalFuncs[fqFuncName]; exists {
				// Perform type conversion for arguments
				for i, arg := range ce.Arguments {
					if i < len(extFuncType.ParameterTypes) {
						argType := t.analyzer.InferExpressionTypes(arg, true)[0]
						expectedType := extFuncType.ParameterTypes[i]

						// Check if type conversion is needed
						if argType.String() != expectedType.String() {
							// Insert code or modify the AST to perform type conversion
							ce.Arguments[i] = t.wrapWithTypeConversion(arg, expectedType)
						}
					}

				}

				// Store the expected return type for code generation
				t.analyzer.ExpectedReturnTypes[ce] = extFuncType.ReturnTypes
			}
		}
		//case *parser.Identifier:
		//	funcSymbol, _ := t.analyzer.CurrentTable.Resolve(ce.Function.String())
		//	for paramId := range ce.Arguments {
		//		expectedType := funcSymbol.GoType.(*types.Signature).Params().At(paramId)
		//		switch ce.Arguments[paramId].(type) {
		//		case *parser.StringLiteral:
		//			ce.Arguments[paramId].(*parser.StringLiteral).Value = fmt.Sprintf("%s(\"%s\")", expectedType.Type().String(), ce.Arguments[paramId].(*parser.StringLiteral).String())
		//		case *parser.Identifier:
		//			ce.Arguments[paramId].(*parser.Identifier).Value = fmt.Sprintf("%s(%s)", expectedType.Type().String(), ce.Arguments[paramId].(*parser.Identifier).String())
		//		case *parser.InfixExpression:
		//			stringValue := ""
		//			stringLiteral := &parser.StringLiteral{
		//				Token: lexer.Token{Type: lexer.TokenString},
		//				Value: stringValue,
		//			}
		//			stringLiteral.Value = t.expressionToString(ce.Arguments[paramId].(*parser.InfixExpression))
		//			stringLiteral.Value = fmt.Sprintf("%s(%s)", expectedType.Type().String(), stringLiteral.Value)
		//			ce.Arguments[paramId] = stringLiteral
		//		}
		//	}
	}

	// Recursively transform arguments
	for _, arg := range ce.Arguments {
		t.Transform(arg, rNode)
	}
}

func (t *Transformer) expressionToString(expr parser.Expression) string {
	switch e := expr.(type) {
	case *parser.InfixExpression:
		ls := t.expressionToString(e.Left)
		rs := t.expressionToString(e.Right)
		switch e.Operator {
		case "+":
			return fmt.Sprintf("fmt.Sprintf(\"%%v %%v\", %s, %s)", ls, rs)
		default:
			return fmt.Sprintf("fmt.Sprintf(\"%%v %%v\", %s, %s)", ls, rs)
		}
	case *parser.Identifier:
		return fmt.Sprintf("fmt.Sprintf(\"%%v\", %s)", e.Value)
	case *parser.StringLiteral:
		return fmt.Sprintf("%q", e.Value)
	default:
		return e.String()
	}
}

func (t *Transformer) wrapWithTypeConversion(arg parser.Expression, targetType parser.Type) parser.Expression {
	argType := t.analyzer.InferExpressionTypes(arg, true)[0]
	if t.analyzer.AreTypesCompatible(argType, targetType) {
		return arg
	}

	srcGoType := t.analyzer.GetGoTypeFromParserType(argType)
	destGoType := t.analyzer.GetGoTypeFromParserType(targetType)

	if types.ConvertibleTo(srcGoType, destGoType) {
		return &parser.TypeConversionExpression{
			Expression: arg,
			TargetType: targetType,
		}
	} else {
		//t.analyzer.errors = append(t.analyzer.errors, fmt.Sprintf("Cannot convert type '%s' to '%s'", argType.String(), targetType.String()))
		return arg
	}
}
