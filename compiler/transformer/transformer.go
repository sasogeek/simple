package transformer

import (
	"fmt"
	"go/types"
	"simple/parser"
	"simple/semantic"
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

func (t *Transformer) Transform(node parser.Node) {
	switch n := node.(type) {
	case *parser.Program:
		for _, stmt := range n.Statements {
			t.Transform(stmt)
		}
	case *parser.ExpressionStatement:
		if n != nil {
			t.Transform(n.Expression)
		}
	case *parser.CallExpression:
		t.handleCallExpression(n)
	case *parser.FunctionLiteral:
		t.Transform(n.Body)
	case *parser.BlockStatement:
		for _, stmt := range n.Statements {
			t.Transform(stmt)
		}
		// Handle other node types as needed
	}
}

func (t *Transformer) handleCallExpression(ce *parser.CallExpression) {
	// Check if the function is a SelectorExpression (e.g., pkg.Func)
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
								ce.Arguments[paramId].(*parser.StringLiteral).Value = fmt.Sprintf("%s(\"%s\")", expectedType.Type().String(), ce.Arguments[paramId].(*parser.StringLiteral).Value)
							}
						}

					}

				}
			}
		}

		//pkgName := selExpr.Left.String()
		//funcName := selExpr.Selector.String()
		//fqFuncName := pkgName + "." + funcName

		// Retrieve the function type from the analyzer's ExternalFuncs
		if extFuncType, exists := t.analyzer.ExternalFuncs[fqFuncName]; exists {
			// Perform type conversion for arguments
			for i, arg := range ce.Arguments {
				// Infer the argument's type
				argType := t.analyzer.InferExpressionType(arg, true)
				expectedType := extFuncType.ParameterTypes[i]

				// Check if type conversion is needed
				if argType.String() != expectedType.String() {
					// Insert code or modify the AST to perform type conversion
					ce.Arguments[i] = t.wrapWithTypeConversion(arg, expectedType)
				}
			}

			// Store the expected return type for code generation
			t.analyzer.ExpectedReturnTypes[ce] = extFuncType.ReturnType
		}
	}

	// Recursively transform arguments
	for _, arg := range ce.Arguments {
		t.Transform(arg)
	}
}

func (t *Transformer) wrapWithTypeConversion(arg parser.Expression, targetType parser.Type) parser.Expression {
	argType := t.analyzer.InferExpressionType(arg, true)
	if t.analyzer.AreTypesCompatible(argType, targetType) {
		// No conversion needed
		return arg
	}

	// Use go/types to determine if a conversion is possible
	srcGoType := t.analyzer.GetGoTypeFromParserType(argType)
	destGoType := t.analyzer.GetGoTypeFromParserType(targetType)

	if types.ConvertibleTo(srcGoType, destGoType) {
		// Wrap the argument with a type conversion expression
		return &parser.TypeConversionExpression{
			Expression: arg,
			TargetType: targetType,
		}
	} else {
		// Report an error if conversion is not possible
		//t.analyzer.errors = append(t.analyzer.errors, fmt.Sprintf("Cannot convert type '%s' to '%s'", argType.String(), targetType.String()))
		return arg
	}
}
