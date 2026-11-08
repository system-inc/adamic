package lower

import (
	"sort"
	"strconv"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

func (lowering *lowering) jsonCall(node *ast.Node, name string) (ir.Expression, bool, error) {
	if name == "parse" {
		return nil, true, &Refused{
			Where: lowering.program.Where(node),
			What:  "JSON.parse: its result's type can't be proven from the text",
			Fix:   "a checked parse against a declared type is a later design; construct typed values explicitly for now",
		}
	}
	if name != "stringify" {
		return nil, true, lowering.notYet(node, "JSON."+name)
	}
	args := node.AsCallExpression().Arguments.Nodes
	if len(args) > 3 || hasSpread(node) {
		return nil, true, lowering.notYet(node, "JSON.stringify with spread or extra arguments")
	}
	result := ir.JSONStringify{Value: ir.Undefined{}, Schema: &ir.JSONSchema{Kind: "undefined"}}
	var jsonInputError error
	if len(args) > 0 {
		result.Value, result.Schema, jsonInputError = lowering.jsonInput(args[0])
		if jsonInputError != nil {
			return nil, true, jsonInputError
		}
	}
	if len(args) > 1 {
		if subjectType := lowering.checker.GetTypeAtLocation(
			args[1],
		); len(lowering.checker.GetSignaturesOfType(subjectType, checker.SignatureKindCall)) != 0 {
			return nil, true, lowering.notYet(
				args[1],
				"JSON.stringify replacer functions (the callback must have a proven type for every visited value and its holder)",
			)
		}
		result.Replacer, result.ReplacerSchema, jsonInputError = lowering.jsonInput(args[1])
		if jsonInputError != nil {
			return nil, true, jsonInputError
		}
		k := result.ReplacerSchema.Kind
		if k != "array" && k != "tuple" && k != "null" && k != "undefined" {
			return nil, true, lowering.notYet(
				args[1],
				"JSON.stringify replacer other than a key array, null or undefined",
			)
		}
		// Only scalar keys: objects could run coercion code, which this slice does not invoke.
		if k == "tuple" {
			for _, field := range result.ReplacerSchema.Fields {
				if !jsonScalar(field.Schema) {
					return nil, true, lowering.notYet(args[1], "JSON.stringify replacer keys with object coercion")
				}
			}
		}
		if k == "array" && !jsonScalar(result.ReplacerSchema.Element) {
			return nil, true, lowering.notYet(args[1], "JSON.stringify replacer keys with object coercion")
		}
	}
	if len(args) > 2 {
		result.Space, result.SpaceSchema, jsonInputError = lowering.jsonInput(args[2])
		if jsonInputError != nil {
			return nil, true, jsonInputError
		}
		if !jsonScalar(result.SpaceSchema) {
			return nil, true, lowering.notYet(args[2], "JSON.stringify space requiring object coercion")
		}
	}
	return result, true, nil
}
func jsonScalar(schema *ir.JSONSchema) bool {
	return schema.Kind == "number" ||
		schema.Kind == "boolean" ||
		schema.Kind == "string" ||
		schema.Kind == "undefined" ||
		schema.Kind == "null" ||
		schema.Kind == "union" ||
		schema.Kind == "maybe_number" ||
		schema.Kind == "maybe_boolean"
}

// Direct literals establish their complete shape, and preserve evaluation order. Anything read
// through an object type is refused: that type can hide additional fields or a toJSON method.
func (lowering *lowering) jsonInput(node *ast.Node) (ir.Expression, *ir.JSONSchema, error) {
	n := ast.SkipParentheses(node)
	if lowering.uninitializedInitializer(n) && lowering.placeholderDeclaration(n) {
		// This temporary field has a known nullish arm and never escapes the serializer.
		if _, null := lowering.placeholderInitialValue(n, ir.Union).(ir.Box); null {
			return ir.JSONNull{}, &ir.JSONSchema{Kind: "null"}, nil
		}
		return ir.Undefined{}, &ir.JSONSchema{Kind: "undefined"}, nil
	}
	if n.Kind == ast.KindNullKeyword {
		return ir.JSONNull{}, &ir.JSONSchema{Kind: "null"}, nil
	}
	if n.Kind == ast.KindObjectLiteralExpression && lowering.recordLiteralElement(n) != nil {
		value, expressionError := lowering.expression(n)
		if expressionError != nil {
			return nil, nil, expressionError
		}
		proven := lowering.slotType(n)
		if lowering.recordElement(proven) == nil {
			proven = lowering.checker.GetTypeAtLocation(n)
		}
		schema, jsonTypeError := lowering.jsonType(node, proven, 0)
		return value, schema, jsonTypeError
	}
	if n.Kind == ast.KindArrayLiteralExpression || n.Kind == ast.KindObjectLiteralExpression {
		literal := ir.ObjectLiteral{Tuple: n.Kind == ast.KindArrayLiteralExpression}
		schema := &ir.JSONSchema{Kind: "object"}
		if literal.Tuple {
			schema.Kind = "tuple"
		}
		add := func(name string, node *ast.Node) error {
			if name == "toJSON" || name == "__proto__" {
				return lowering.notYet(node, "JSON.stringify a literal with "+name+" semantics")
			}
			for _, field := range literal.Fields {
				if field.Name == name {
					return lowering.notYet(node, "JSON.stringify duplicate literal keys")
				}
			}
			value, child, jsonLiteralFieldError := lowering.jsonLiteralField(node)
			if jsonLiteralFieldError != nil {
				return jsonLiteralFieldError
			}
			if value.Type() == ir.MaybeBoolean {
				return lowering.notYet(node, "JSON.stringify a literal field with a two-word representation")
			}
			schema.Fields = append(
				schema.Fields,
				ir.JSONField{Name: lowering.constant(name), Slot: len(literal.Fields), Schema: child},
			)
			literal.Fields = append(literal.Fields, ir.Field{
				Name:          name,
				Value:         value,
				Uninitialized: lowering.uninitializedInitializer(node),
				Unset:         lowering.uninitializedInitializer(node),
			})
			return nil
		}
		if literal.Tuple {
			for index, v := range n.AsArrayLiteralExpression().Elements.Nodes {
				if v.Kind == ast.KindSpreadElement || v.Kind == ast.KindOmittedExpression {
					return nil, nil, lowering.notYet(v, "JSON.stringify a spread or hole in a literal")
				}
				if addError := add(strconv.Itoa(index), v); addError != nil {
					return nil, nil, addError
				}
			}
		} else {
			for _, f := range n.AsObjectLiteralExpression().Properties.Nodes {
				if f.Kind != ast.KindPropertyAssignment && f.Kind != ast.KindShorthandPropertyAssignment {
					return nil, nil, lowering.notYet(f, "JSON.stringify a literal with spread or methods")
				}
				key := f.Name()
				if !ast.IsIdentifier(key) && key.Kind != ast.KindStringLiteral && key.Kind != ast.KindNumericLiteral {
					return nil, nil, lowering.notYet(key, "JSON.stringify computed keys")
				}
				name := key.Text()
				if key.Kind == ast.KindNumericLiteral {
					number, parseError := strconv.ParseFloat(name, 64)
					if parseError != nil {
						return nil, nil, lowering.notYet(key, "JSON.stringify this numeric key")
					}
					name = strconv.FormatFloat(number, 'f', -1, 64)
					if _, ok := jsonIndex(name); !ok {
						return nil, nil, lowering.notYet(
							key,
							"JSON.stringify a numeric key outside the array-index range (spell it as a string)",
						)
					}
				}
				value := f
				if f.Kind == ast.KindPropertyAssignment {
					value = f.AsPropertyAssignment().Initializer
				}
				if addError := add(name, value); addError != nil {
					return nil, nil, addError
				}
			}
			// Integer indexes precede other strings, whose order is their insertion order.
			sort.SliceStable(schema.Fields, func(left, right int) bool {
				a, ai := jsonIndex(lowering.result.Strings[schema.Fields[left].Name])
				b, bi := jsonIndex(lowering.result.Strings[schema.Fields[right].Name])
				if ai != bi {
					return ai
				}
				return ai && a < b
			})
		}
		return literal, schema, nil
	}
	if schema, known := lowering.constructionJson(n); known {
		value, expressionError := lowering.expression(node)
		return value, schema, expressionError
	}
	schema, jsonTypeError := lowering.jsonType(node, lowering.checker.GetTypeAtLocation(node), 0)
	if jsonTypeError != nil {
		return nil, nil, jsonTypeError
	}
	value, expressionError := lowering.expression(node)
	return value, schema, expressionError
}
func jsonIndex(name string) (uint64, bool) {
	number, parseError := strconv.ParseUint(name, 10, 32)
	return number, parseError == nil && number < 4294967295 && strconv.FormatUint(number, 10) == name
}
func (lowering *lowering) jsonType(node *ast.Node, subjectType *checker.Type, depth int) (*ir.JSONSchema, error) {
	if depth == 0 && lowering.censusCallReturnsUndefined(node) {
		return &ir.JSONSchema{Kind: "undefined"}, nil
	}
	if depth > 64 {
		return nil, lowering.notYet(node, "JSON.stringify recursive array types")
	}
	if subjectType.Flags()&checker.TypeFlagsUndefined != 0 {
		return &ir.JSONSchema{Kind: "undefined"}, nil
	}
	if subjectType.Flags()&checker.TypeFlagsNever != 0 {
		return &ir.JSONSchema{Kind: "undefined"}, nil
	}
	of, known := lowering.representation(subjectType)
	if !known {
		return nil, lowering.notYet(node, "JSON.stringify a value of type "+lowering.checker.TypeToString(subjectType))
	}
	kinds := map[ir.Type]string{
		ir.Number:       "number",
		ir.Boolean:      "boolean",
		ir.String:       "string",
		ir.Map:          "map",
		ir.Closure:      "function",
		ir.MaybeNumber:  "maybe_number",
		ir.MaybeBoolean: "maybe_boolean",
		ir.Union:        "union",
	}
	if of == ir.Record {
		child, jsonTypeError := lowering.jsonType(
			node,
			lowering.recordElement(lowering.checker.GetNonNullableType(subjectType)),
			depth+1,
		)
		if jsonTypeError != nil {
			return nil, jsonTypeError
		}
		if child.Kind == "function" || child.Kind == "union" {
			return nil, lowering.notYet(node, "JSON.stringify record values that may provide a callable toJSON")
		}
		return &ir.JSONSchema{Kind: "record", Element: child}, nil
	}
	if of == ir.Array {
		element := lowering.elementTypeOfArray(lowering.checker.GetNonNullableType(subjectType))
		if _, extraElement, known := lowering.nodeArrayLayoutOf(
			lowering.checker.GetNonNullableType(subjectType),
		); known {
			element = extraElement
		}
		if element == nil {
			return nil, lowering.notYet(node, "JSON.stringify an array without a proven element type")
		}
		child, jsonTypeError := lowering.jsonType(node, element, depth+1)
		return &ir.JSONSchema{Kind: "array", Element: child}, jsonTypeError
	}
	if of == ir.Object || of == ir.Weak {
		return nil, lowering.notYet(
			node,
			"JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata)",
		)
	}
	if subjectType.Flags()&checker.TypeFlagsUnion != 0 {
		for _, memberType := range subjectType.Types() {
			child, jsonTypeError := lowering.jsonType(node, memberType, depth+1)
			if jsonTypeError != nil {
				return nil, jsonTypeError
			}
			if of == ir.Union &&
				(child.Kind == "array" || child.Kind == "object" || child.Kind == "tuple" || child.Kind == "record") {
				return nil, lowering.notYet(
					node,
					"JSON.stringify a union containing containers without runtime element metadata",
				)
			}
		}
	}
	kind, ok := kinds[of]
	if !ok {
		return nil, lowering.notYet(node, "JSON.stringify this representation")
	}
	return &ir.JSONSchema{Kind: kind}, nil
}

// A shorthand name is a property symbol, not the binding it reads. Resolve
// the value symbol just as ordinary object-literal lowering does. Its declared
// type describes the stored representation even when the occurrence is narrowed.
func (lowering *lowering) jsonLiteralField(node *ast.Node) (ir.Expression, *ir.JSONSchema, error) {
	if node.Kind != ast.KindShorthandPropertyAssignment {
		return lowering.jsonInput(node)
	}
	value, shorthandError := lowering.shorthand(node)
	if shorthandError != nil {
		return nil, nil, shorthandError
	}
	symbol := lowering.checker.GetShorthandAssignmentValueSymbol(node)
	if symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
		symbol = lowering.checker.GetAliasedSymbol(symbol)
	}
	if symbol == nil {
		return nil, nil, lowering.notYet(node, "JSON.stringify shorthand without a value binding")
	}
	schema, jsonTypeError := lowering.jsonType(node.Name(), lowering.checker.GetTypeOfSymbol(symbol), 0)
	return value, schema, jsonTypeError
}
