package lower

import (
	"sort"
	"strconv"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

func (l *lowering) jsonCall(node *ast.Node, name string) (ir.Expression, bool, error) {
	if name == "parse" {
		return nil, true, &Refused{Where: l.program.Where(node), What: "JSON.parse: its result's type can't be proven from the text", Fix: "a checked parse against a declared type is a later design; construct typed values explicitly for now (adamic/checked-json-parse)"}
	}
	if name != "stringify" {
		return nil, true, l.notYet(node, "JSON."+name)
	}
	args := node.AsCallExpression().Arguments.Nodes
	if len(args) > 3 || hasSpread(node) {
		return nil, true, l.notYet(node, "JSON.stringify with spread or extra arguments")
	}
	result := ir.JSONStringify{Value: ir.Undefined{}, Schema: &ir.JSONSchema{Kind: "undefined"}}
	var err error
	if len(args) > 0 {
		result.Value, result.Schema, err = l.jsonInput(args[0])
		if err != nil {
			return nil, true, err
		}
	}
	if len(args) > 1 {
		if t := l.checker.GetTypeAtLocation(args[1]); len(l.checker.GetSignaturesOfType(t, checker.SignatureKindCall)) != 0 {
			return nil, true, l.notYet(args[1], "JSON.stringify replacer functions (the callback must have a proven type for every visited value and its holder) (construct the serializable values explicitly before stringify, or use a key-array replacer to select fields)")
		}
		result.Replacer, result.ReplacerSchema, err = l.jsonInput(args[1])
		if err != nil {
			return nil, true, err
		}
		k := result.ReplacerSchema.Kind
		if k != "array" && k != "tuple" && k != "null" && k != "undefined" {
			return nil, true, l.notYet(args[1], "JSON.stringify replacer other than a key array, null or undefined (pass an array of string keys, null or undefined as the replacer)")
		}
		// Only scalar keys: objects could run coercion code, which this slice does not invoke.
		if k == "tuple" {
			for _, f := range result.ReplacerSchema.Fields {
				if !jsonScalar(f.Schema) {
					return nil, true, l.notYet(args[1], "JSON.stringify replacer keys with object coercion (convert replacer keys to strings explicitly before the call)")
				}
			}
		}
		if k == "array" && !jsonScalar(result.ReplacerSchema.Element) {
			return nil, true, l.notYet(args[1], "JSON.stringify replacer keys with object coercion (convert replacer keys to strings explicitly before the call)")
		}
	}
	if len(args) > 2 {
		result.Space, result.SpaceSchema, err = l.jsonInput(args[2])
		if err != nil {
			return nil, true, err
		}
		if !jsonScalar(result.SpaceSchema) {
			return nil, true, l.notYet(args[2], "JSON.stringify space requiring object coercion (pass a number or string as the space argument)")
		}
	}
	return result, true, nil
}
func jsonScalar(s *ir.JSONSchema) bool {
	return s.Kind == "number" || s.Kind == "boolean" || s.Kind == "string" || s.Kind == "undefined" || s.Kind == "null" || s.Kind == "union" || s.Kind == "maybe_number" || s.Kind == "maybe_boolean"
}

// Direct literals establish their complete shape, and preserve evaluation order. Anything read
// through an object type is refused: that type can hide additional fields or a toJSON method.
func (l *lowering) jsonInput(node *ast.Node) (ir.Expression, *ir.JSONSchema, error) {
	n := ast.SkipParentheses(node)
	if l.numericTypedArray(l.checker.GetTypeAtLocation(node)) {
		return nil, nil, l.notYet(node, "JSON.stringify typed arrays requires their own numeric keys")
	}
	if l.program.ContainsSparseArrays() && (n.Kind == ast.KindArrayLiteralExpression || l.checker.IsArrayType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(node)))) {
		return nil, nil, l.notYet(node, "JSON.stringify arrays in a program with sparse arrays")
	}
	if n.Kind == ast.KindNullKeyword {
		return ir.JSONNull{}, &ir.JSONSchema{Kind: "null"}, nil
	}
	if n.Kind == ast.KindObjectLiteralExpression && l.recordLiteralElement(n) != nil {
		value, err := l.expression(n)
		if err != nil {
			return nil, nil, err
		}
		proven := l.checker.GetContextualType(n, checker.ContextFlagsNone)
		if l.recordElement(proven) == nil {
			proven = l.checker.GetTypeAtLocation(n)
		}
		schema, err := l.jsonType(node, proven, 0)
		return value, schema, err
	}
	if n.Kind == ast.KindArrayLiteralExpression || n.Kind == ast.KindObjectLiteralExpression {
		literal := ir.ObjectLiteral{Tuple: n.Kind == ast.KindArrayLiteralExpression}
		schema := &ir.JSONSchema{Kind: "object"}
		if literal.Tuple {
			schema.Kind = "tuple"
		}
		add := func(name string, v *ast.Node) error {
			if name == "toJSON" || name == "__proto__" {
				return l.notYet(v, "JSON.stringify a literal with "+name+" semantics")
			}
			for _, f := range literal.Fields {
				if f.Name == name {
					return l.notYet(v, "JSON.stringify duplicate literal keys")
				}
			}
			value, child, err := l.jsonInput(v)
			if err != nil {
				return err
			}
			if value.Type() == ir.MaybeBoolean {
				return l.notYet(v, "JSON.stringify a literal field with a two-word representation")
			}
			schema.Fields = append(schema.Fields, ir.JSONField{Name: l.constant(name), Slot: len(literal.Fields), Schema: child})
			literal.Fields = append(literal.Fields, ir.Field{Name: name, Value: value})
			return nil
		}
		if literal.Tuple {
			for i, v := range n.AsArrayLiteralExpression().Elements.Nodes {
				if v.Kind == ast.KindSpreadElement || v.Kind == ast.KindOmittedExpression {
					return nil, nil, l.notYet(v, "JSON.stringify a spread or hole in a literal")
				}
				if err := add(strconv.Itoa(i), v); err != nil {
					return nil, nil, err
				}
			}
		} else {
			for _, f := range n.AsObjectLiteralExpression().Properties.Nodes {
				if f.Kind != ast.KindPropertyAssignment {
					return nil, nil, l.notYet(f, "JSON.stringify a literal with spread, shorthand or methods (write explicit field: value entries in a plain literal)")
				}
				key := f.Name()
				if !ast.IsIdentifier(key) && key.Kind != ast.KindStringLiteral && key.Kind != ast.KindNumericLiteral {
					return nil, nil, l.notYet(key, "JSON.stringify computed keys (spell each key as an identifier or string literal)")
				}
				name := key.Text()
				if key.Kind == ast.KindNumericLiteral {
					v, err := strconv.ParseFloat(name, 64)
					if err != nil {
						return nil, nil, l.notYet(key, "JSON.stringify this numeric key")
					}
					name = strconv.FormatFloat(v, 'f', -1, 64)
					if _, ok := jsonIndex(name); !ok {
						return nil, nil, l.notYet(key, "JSON.stringify a numeric key outside the array-index range (spell it as a string)")
					}
				}
				if err := add(name, f.AsPropertyAssignment().Initializer); err != nil {
					return nil, nil, err
				}
			}
			// Integer indexes precede other strings, whose order is their insertion order.
			sort.SliceStable(schema.Fields, func(i, j int) bool {
				a, ai := jsonIndex(l.result.Strings[schema.Fields[i].Name])
				b, bi := jsonIndex(l.result.Strings[schema.Fields[j].Name])
				if ai != bi {
					return ai
				}
				return ai && a < b
			})
		}
		return literal, schema, nil
	}
	schema, err := l.jsonType(node, l.checker.GetTypeAtLocation(node), 0)
	if err != nil {
		return nil, nil, err
	}
	value, err := l.expression(node)
	return value, schema, err
}
func jsonIndex(name string) (uint64, bool) {
	n, err := strconv.ParseUint(name, 10, 32)
	return n, err == nil && n < 4294967295 && strconv.FormatUint(n, 10) == name
}
func (l *lowering) jsonType(node *ast.Node, t *checker.Type, depth int) (*ir.JSONSchema, error) {
	if depth > 64 {
		return nil, l.notYet(node, "JSON.stringify recursive array types")
	}
	if t.Flags()&checker.TypeFlagsNull != 0 {
		return &ir.JSONSchema{Kind: "null"}, nil
	}
	if t.Flags()&checker.TypeFlagsUndefined != 0 {
		return &ir.JSONSchema{Kind: "undefined"}, nil
	}
	if t.Flags()&checker.TypeFlagsNever != 0 {
		return &ir.JSONSchema{Kind: "undefined"}, nil
	}
	of, known := l.representation(t)
	if !known {
		return nil, l.notYet(node, "JSON.stringify a value of type "+l.checker.TypeToString(t))
	}
	kinds := map[ir.Type]string{ir.Number: "number", ir.Boolean: "boolean", ir.String: "string", ir.Map: "map", ir.Closure: "function", ir.MaybeNumber: "maybe_number", ir.MaybeBoolean: "maybe_boolean", ir.Union: "union"}
	if of == ir.Record {
		if l.includesNull(t) {
			return nil, l.notYet(node, "JSON.stringify a value of type "+l.checker.TypeToString(t))
		}
		child, err := l.jsonType(node, l.recordElement(l.checker.GetNonNullableType(t)), depth+1)
		if err != nil {
			return nil, err
		}
		if child.Kind == "function" || child.Kind == "union" {
			return nil, l.notYet(node, "JSON.stringify record values that may provide a callable toJSON")
		}
		return &ir.JSONSchema{Kind: "record", Element: child}, nil
	}
	if of == ir.Array {
		element := l.checker.GetElementTypeOfArrayType(l.checker.GetNonNullableType(t))
		if element == nil {
			return nil, l.notYet(node, "JSON.stringify an array without a proven element type")
		}
		child, err := l.jsonType(node, element, depth+1)
		return &ir.JSONSchema{Kind: "array", Element: child}, err
	}
	if of == ir.Object || of == ir.Weak {
		return nil, l.notYet(node, "JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata)")
	}
	if t.Flags()&checker.TypeFlagsUnion != 0 {
		for _, m := range t.Types() {
			child, err := l.jsonType(node, m, depth+1)
			if err != nil {
				return nil, err
			}
			if of == ir.Union && (child.Kind == "array" || child.Kind == "object" || child.Kind == "tuple" || child.Kind == "record") {
				return nil, l.notYet(node, "JSON.stringify a union containing containers without runtime element metadata")
			}
		}
	}
	kind, ok := kinds[of]
	if !ok {
		return nil, l.notYet(node, "JSON.stringify this representation")
	}
	return &ir.JSONSchema{Kind: kind}, nil
}
