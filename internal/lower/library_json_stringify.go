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
		value, err := l.jsonParse(node, false)
		return value, true, err
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
	if result.Schema.ContainsParsed() && len(args) > 1 {
		return nil, true, l.notYet(node, "JSON.stringify parsed values with replacer or space without full runtime metadata")
	}
	if len(args) > 1 {
		t := l.checker.GetTypeAtLocation(args[1])
		signatures := l.checker.GetSignaturesOfType(t, checker.SignatureKindCall)
		if len(signatures) != 0 {
			if len(signatures) != 1 {
				return nil, true, l.notYet(args[1], "JSON.stringify overloaded replacer")
			}
			parameters := signatures[0].Parameters()
			if len(parameters) > 2 {
				return nil, true, l.notYet(args[1], "JSON.stringify replacer holder or context without proven metadata")
			}
			if len(parameters) > 0 {
				of, known := l.representation(l.checker.GetTypeOfSymbol(parameters[0]))
				if !known || of != ir.String || !l.jsonScalarArgument(ir.String, l.checker.GetTypeOfSymbol(parameters[0])) {
					return nil, true, l.notYet(args[1], "JSON.stringify replacer key without proven string type")
				}
			}
			if len(parameters) > 1 {
				of, known := l.representation(l.checker.GetTypeOfSymbol(parameters[1]))
				kinds := map[ir.Type]string{ir.Number: "number", ir.String: "string", ir.Boolean: "boolean"}
				if !known || kinds[of] == "" || result.Schema.Kind != kinds[of] || !l.jsonScalarArgument(of, l.checker.GetTypeOfSymbol(parameters[1])) {
					return nil, true, l.notYet(args[1], "JSON.stringify replacer value without a proven scalar input")
				}
			}
			returned := l.checker.GetReturnTypeOfSignature(signatures[0])
			if returned.Flags()&checker.TypeFlagsVoid != 0 && !l.jsonCallbackReturnsUndefined(args[1]) {
				return nil, true, l.notYet(args[1], "JSON.stringify void replacer without proof that its actual return is undefined")
			}
			child := &ir.JSONSchema{Kind: "undefined"}
			if returned.Flags()&(checker.TypeFlagsVoid|checker.TypeFlagsUndefined|checker.TypeFlagsNever) == 0 {
				of, known := l.representation(returned)
				kinds := map[ir.Type]string{ir.Number: "number", ir.String: "string", ir.Boolean: "boolean"}
				if !known || kinds[of] == "" {
					return nil, true, l.notYet(args[1], "JSON.stringify replacer return without a proven scalar representation")
				}
				child.Kind = kinds[of]
				if l.includesUndefined(returned) && of != ir.String {
					return nil, true, l.notYet(args[1], "JSON.stringify optional replacer result without proven metadata")
				}
			}
			// Every return is primitive, so SerializeJSONProperty visits only the root.
			result.Replacer, err = l.expression(args[1])
			result.ReplacerSchema = &ir.JSONSchema{Kind: "root_replacer", Element: child}
		} else {
			result.Replacer, result.ReplacerSchema, err = l.jsonInput(args[1])
		}
		if err != nil {
			return nil, true, err
		}
		k := result.ReplacerSchema.Kind
		if k == "parsed" || k == "function" || (k == "union" && !l.writable(t)) {
			return nil, true, l.notYet(args[1], "JSON.stringify replacer without proven array identity or callability")
		}
		// Only scalar keys: objects could run coercion code, which this slice does not invoke.
		if k == "tuple" {
			for _, f := range result.ReplacerSchema.Fields {
				if !jsonScalar(f.Schema) {
					return nil, true, l.notYet(args[1], "JSON.stringify replacer keys with object coercion")
				}
			}
		}
		if k == "array" && !jsonScalar(result.ReplacerSchema.Element) {
			return nil, true, l.notYet(args[1], "JSON.stringify replacer keys with object coercion")
		}
	}
	if len(args) > 2 {
		result.Space, result.SpaceSchema, err = l.jsonInput(args[2])
		if err != nil {
			return nil, true, err
		}
		if !jsonScalar(result.SpaceSchema) {
			return nil, true, l.notYet(args[2], "JSON.stringify space requiring object coercion")
		}
	}
	return result, true, nil
}
func jsonScalar(s *ir.JSONSchema) bool {
	return s.Kind == "number" || s.Kind == "boolean" || s.Kind == "string" || s.Kind == "undefined" || s.Kind == "null" || s.Kind == "union" || s.Kind == "maybe_number" || s.Kind == "maybe_boolean"
}

// Literal origins establish complete shapes and preserve evaluation order. Structural views
// alone remain insufficient: they can hide additional fields or a toJSON method.
func (l *lowering) jsonInput(node *ast.Node) (ir.Expression, *ir.JSONSchema, error) {
	n := ast.SkipParentheses(node)
	if n.Kind == ast.KindNewExpression && l.isLibraryGlobal(n.AsNewExpression().Expression, "Number") {
		// Keep this JSON-only intrinsic slot independent of the Number method slice.
		arguments := n.AsNewExpression().Arguments
		var value ir.Expression = ir.NumberConstant{}
		var err error
		if arguments != nil && len(arguments.Nodes) > 0 {
			if len(arguments.Nodes) != 1 || arguments.Nodes[0].Kind == ast.KindSpreadElement {
				return nil, nil, l.notYet(n, "JSON.stringify boxed primitive with spread or extra arguments")
			}
			value, err = l.libraryNumber(arguments.Nodes[0])
		}
		return value, &ir.JSONSchema{Kind: "number"}, err
	}
	if n.Kind == ast.KindNewExpression && (l.isLibraryGlobal(n.AsNewExpression().Expression, "String") || l.isLibraryGlobal(n.AsNewExpression().Expression, "Boolean")) {
		boolean := l.isLibraryGlobal(n.AsNewExpression().Expression, "Boolean")
		arguments := n.AsNewExpression().Arguments
		kind := "string"
		var value ir.Expression = ir.StringConstant{Index: l.constant("")}
		if boolean {
			kind = "boolean"
			value = ir.BooleanConstant{}
		}
		if arguments != nil && len(arguments.Nodes) > 0 {
			if len(arguments.Nodes) != 1 || arguments.Nodes[0].Kind == ast.KindSpreadElement {
				return nil, nil, l.notYet(n, "JSON.stringify boxed primitive with spread or extra arguments")
			}
			var err error
			if boolean {
				value, err = l.expression(arguments.Nodes[0])
				if err == nil && value.Type() != ir.Boolean {
					return nil, nil, l.notYet(n, "JSON.stringify boxed Boolean without a proven boolean argument")
				}
			} else {
				value, err = l.stringConversion(arguments.Nodes[0])
			}
			if err != nil {
				return nil, nil, err
			}
		}
		return value, &ir.JSONSchema{Kind: kind}, nil
	}
	if n.Kind == ast.KindCallExpression {
		callee := ast.SkipParentheses(n.AsCallExpression().Expression)
		if callee.Kind == ast.KindPropertyAccessExpression && callee.Name().Text() == "parse" && l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "JSON") {
			value, err := l.jsonParse(n, true)
			return value, &ir.JSONSchema{Kind: "parsed"}, err
		}
	}
	if n.Kind == ast.KindRegularExpressionLiteral {
		// A fresh intrinsic RegExp has no enumerable own fields. References stay refused:
		// structural views and expando properties require complete runtime shape evidence.
		value, err := l.expression(n)
		return value, &ir.JSONSchema{Kind: "regexp"}, err
	}
	if n.Kind == ast.KindNullKeyword {
		return ir.JSONNull{}, &ir.JSONSchema{Kind: "null"}, nil
	}
	if n.Kind == ast.KindArrayLiteralExpression || n.Kind == ast.KindObjectLiteralExpression {
		literal := ir.ObjectLiteral{Tuple: n.Kind == ast.KindArrayLiteralExpression}
		schema := &ir.JSONSchema{Kind: "object"}
		if literal.Tuple {
			schema.Kind = "tuple"
		}
		add := func(name string, v *ast.Node) error {
			if name == "__proto__" {
				return l.notYet(v, "JSON.stringify a literal with "+name+" semantics")
			}
			if name == "toJSON" && len(l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(v), checker.SignatureKindCall)) == 0 && l.jsonMayCall(l.checker.GetTypeAtLocation(v)) {
				return l.notYet(v, "JSON.stringify toJSON without proven callability")
			}
			if name == "toJSON" && len(l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(v), checker.SignatureKindCall)) != 0 {
				signatures := l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(v), checker.SignatureKindCall)
				if len(signatures) != 1 || len(signatures[0].Parameters()) != 0 {
					return l.notYet(v, "JSON.stringify toJSON without a proven zero-argument callback")
				}
				returned := l.checker.GetReturnTypeOfSignature(signatures[0])
				child := &ir.JSONSchema{Kind: "undefined"}
				if returned.Flags()&checker.TypeFlagsVoid != 0 && !l.jsonCallbackReturnsUndefined(v) {
					return l.notYet(v, "JSON.stringify void toJSON without proof of its actual return")
				}
				if returned.Flags()&(checker.TypeFlagsVoid|checker.TypeFlagsUndefined|checker.TypeFlagsNever) == 0 {
					var err error
					child, err = l.jsonType(v, returned, 0)
					if err != nil {
						return err
					}
				}
				if child.Kind != "number" && child.Kind != "string" && child.Kind != "boolean" && child.Kind != "undefined" && child.Kind != "array" {
					return l.notYet(v, "JSON.stringify toJSON return without proven primitive or array metadata")
				}
				if child.Kind == "array" && !l.jsonArrayCallbackReturn(v, child) {
					return l.notYet(v, "JSON.stringify toJSON array return without proven literal element metadata")
				}
				schema.Kind, schema.Element = "toJSON", child
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
				if v.Kind == ast.KindOmittedExpression {
					schema.Fields = append(schema.Fields, ir.JSONField{Name: l.constant(strconv.Itoa(i)), Slot: len(literal.Fields), Schema: &ir.JSONSchema{Kind: "undefined"}})
					literal.Fields = append(literal.Fields, ir.Field{Name: strconv.Itoa(i), Value: ir.Undefined{}})
					continue
				}
				if v.Kind == ast.KindSpreadElement {
					return nil, nil, l.notYet(v, "JSON.stringify a spread or hole in a literal")
				}
				if err := add(strconv.Itoa(i), v); err != nil {
					return nil, nil, err
				}
			}
		} else {
			for _, f := range n.AsObjectLiteralExpression().Properties.Nodes {
				if f.Kind != ast.KindPropertyAssignment {
					return nil, nil, l.notYet(f, "JSON.stringify a literal with spread, shorthand or methods")
				}
				key := f.Name()
				if !ast.IsIdentifier(key) && key.Kind != ast.KindStringLiteral && key.Kind != ast.KindNumericLiteral {
					return nil, nil, l.notYet(key, "JSON.stringify computed keys")
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
	if origin := l.jsonObjectOrigin(n, 0); origin != nil && ast.IsIdentifier(n) {
		if l.jsonLocalOnly(n, origin) {
			schema, err := l.jsonNativeLiteralSchema(origin, 0)
			if err != nil {
				return nil, nil, err
			}
			value, err := l.expression(n)
			return value, schema, err
		}
		schema := &ir.JSONSchema{Kind: "object"}
		for slot, field := range origin.AsObjectLiteralExpression().Properties.Nodes {
			name := field.Name().Text()
			if name == "toJSON" {
				return nil, nil, l.notYet(node, "JSON.stringify a literal with toJSON semantics")
			}
			property := l.checker.GetPropertyOfType(l.checker.GetTypeAtLocation(n), name)
			child, err := l.jsonType(node, l.checker.GetTypeOfSymbol(property), 0)
			if err != nil {
				return nil, nil, err
			}
			if !jsonScalar(child) || child.Kind == "maybe_boolean" {
				return nil, nil, l.notYet(node, "JSON.stringify an object's non-scalar fields without complete value metadata")
			}
			schema.Fields = append(schema.Fields, ir.JSONField{Name: l.constant(name), Slot: slot, Schema: child})
		}
		sort.SliceStable(schema.Fields, func(i, j int) bool {
			a, ai := jsonIndex(l.result.Strings[schema.Fields[i].Name])
			b, bi := jsonIndex(l.result.Strings[schema.Fields[j].Name])
			if ai != bi {
				return ai
			}
			return ai && a < b
		})
		value, err := l.expression(n)
		return value, schema, err
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
	generic := t.Flags()&checker.TypeFlagsTypeParameter != 0
	t = l.concrete(t)
	// Generic instantiations with the same Union representation share a body. Even a scalar
	// first call could therefore supply its descriptor to a later container union (probe B).
	// Refuse generic boxed unions before that body is cached. Nullable references have
	// a complete concrete schema and must retain their proven empty-case bit.
	if held, known := l.representation(t); generic && known && held == ir.Union {
		return nil, l.notYet(node, "JSON.stringify a generic union without per-instantiation container metadata")
	}
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
	if l.isLibraryType(l.checker.GetNonNullableType(t), "Date") {
		if l.includesUndefined(t) || l.includesNull(t) {
			return nil, l.notYet(node, "JSON.stringify an optional Date")
		}
		return &ir.JSONSchema{Kind: "date"}, nil
	}
	of, known := l.representation(t)
	if !known {
		return nil, l.notYet(node, "JSON.stringify a value of type "+l.checker.TypeToString(t))
	}
	kinds := map[ir.Type]string{ir.Number: "number", ir.Boolean: "boolean", ir.String: "string", ir.Map: "map", ir.Closure: "function", ir.MaybeNumber: "maybe_number", ir.MaybeBoolean: "maybe_boolean", ir.Union: "union"}
	if of == ir.Array {
		element := l.checker.GetElementTypeOfArrayType(l.checker.GetNonNullableType(t))
		if element == nil {
			return nil, l.notYet(node, "JSON.stringify an array without a proven element type")
		}
		child, err := l.jsonType(node, element, depth+1)
		return &ir.JSONSchema{Kind: "array", Element: child, Null: l.includesNull(t)}, err
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
			if of == ir.Union && (child.Kind == "array" || child.Kind == "object" || child.Kind == "tuple") {
				return nil, l.notYet(node, "JSON.stringify a union containing containers without runtime element metadata")
			}
		}
	}
	kind, ok := kinds[of]
	if !ok {
		return nil, l.notYet(node, "JSON.stringify this representation")
	}
	return &ir.JSONSchema{Kind: kind, Null: l.includesNull(t)}, nil
}

// A JSON-local complete-origin proof, without adding a language reflection operation. Reject
// annotations and every rebinding; scalar field stores preserve the fixed native shape.
func (l *lowering) jsonObjectOrigin(node *ast.Node, depth int) *ast.Node {
	if depth > 16 {
		return nil
	}
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindObjectLiteralExpression {
		if l.exactObject(node, 0) {
			return node
		}
		return nil
	}
	if !ast.IsIdentifier(node) {
		return nil
	}
	symbol := l.symbol(node)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return nil
	}
	declaration := symbol.Declarations[0]
	if declaration.Kind != ast.KindVariableDeclaration {
		return nil
	}
	variable := declaration.AsVariableDeclaration()
	if variable.Type != nil || variable.Initializer == nil {
		return nil
	}
	safe := true
	var contains ast.Visitor
	contains = func(part *ast.Node) bool {
		if ast.IsIdentifier(part) && l.symbol(part) == symbol {
			safe = false
			return true
		}
		return part.ForEachChild(contains)
	}
	var visit ast.Visitor
	visit = func(part *ast.Node) bool {
		if part.Kind == ast.KindBinaryExpression {
			binary := part.AsBinaryExpression()
			target := ast.SkipParentheses(binary.Left)
			if ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
				if ast.IsIdentifier(target) && l.symbol(target) == symbol {
					safe = false
				}
				if target.Kind != ast.KindPropertyAccessExpression && target.Kind != ast.KindElementAccessExpression {
					contains(target)
				}
			}
		}
		if part.Kind == ast.KindForOfStatement || part.Kind == ast.KindForInStatement {
			contains(part.AsForInOrOfStatement().Initializer)
		}
		if !safe {
			return true
		}
		return part.ForEachChild(visit)
	}
	ast.GetSourceFileOfNode(declaration).AsNode().ForEachChild(visit)
	if !safe {
		return nil
	}
	return l.jsonObjectOrigin(variable.Initializer, depth+1)
}

func (l *lowering) jsonMayCall(t *checker.Type) bool {
	t = l.concrete(t)
	if len(l.checker.GetSignaturesOfType(t, checker.SignatureKindCall)) != 0 {
		return true
	}
	if t.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range t.Types() {
			if l.jsonMayCall(member) {
				return true
			}
		}
	}
	return false
}

// Nested schemas need a stronger proof than scalar fields: every value is made in the literal,
// and the binding is used only as JSON.stringify's first argument. Reject aliases, property
// reads/stores, exports and shorthand escapes rather than infer hidden descendant shapes.
func (l *lowering) jsonLocalOnly(node, origin *ast.Node) bool {
	symbol := l.symbol(node)
	declaration := symbol.Declarations[0]
	if ast.SkipParentheses(declaration.AsVariableDeclaration().Initializer) != origin {
		return false
	}
	safe := true
	var visit ast.Visitor
	visit = func(part *ast.Node) bool {
		if part.Kind == ast.KindExportDeclaration || ast.HasSyntacticModifier(part, ast.ModifierFlagsExport) {
			safe = false
			return true
		}
		if part.Kind == ast.KindShorthandPropertyAssignment {
			value := l.checker.GetShorthandAssignmentValueSymbol(part)
			if value != nil && l.checker.GetExportSymbolOfSymbol(value) == symbol {
				safe = false
				return true
			}
		}
		if ast.IsIdentifier(part) && l.symbol(part) == symbol && part != declaration.Name() {
			use := part
			for use.Parent != nil && use.Parent.Kind == ast.KindParenthesizedExpression {
				use = use.Parent
			}
			parent := use.Parent
			if parent == nil || parent.Kind != ast.KindCallExpression {
				safe = false
				return true
			}
			call := parent.AsCallExpression()
			callee := ast.SkipParentheses(call.Expression)
			if len(call.Arguments.Nodes) == 0 || call.Arguments.Nodes[0] != use || callee.Kind != ast.KindPropertyAccessExpression || callee.Name().Text() != "stringify" || !l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "JSON") {
				safe = false
				return true
			}
		}
		return part.ForEachChild(visit)
	}
	ast.GetSourceFileOfNode(declaration).AsNode().ForEachChild(visit)
	return safe
}
func (l *lowering) jsonNativeLiteralSchema(node *ast.Node, depth int) (*ir.JSONSchema, error) {
	node = ast.SkipParentheses(node)
	if depth > 64 || !l.exactObject(node, 0) {
		return nil, l.notYet(node, "JSON.stringify nested origin without complete bounded literal metadata")
	}
	schema := &ir.JSONSchema{Kind: "object"}
	for slot, field := range node.AsObjectLiteralExpression().Properties.Nodes {
		if field.Kind != ast.KindPropertyAssignment || field.Name().Text() == "toJSON" {
			return nil, l.notYet(field, "JSON.stringify nested origin with shorthand or toJSON metadata")
		}
		value := ast.SkipParentheses(field.AsPropertyAssignment().Initializer)
		var child *ir.JSONSchema
		var err error
		if value.Kind == ast.KindObjectLiteralExpression {
			child, err = l.jsonNativeLiteralSchema(value, depth+1)
		} else {
			child, err = l.jsonType(value, l.checker.GetTypeAtLocation(value), depth+1)
			if err == nil && ((!jsonScalar(child) && (child.Kind != "array" || value.Kind != ast.KindArrayLiteralExpression)) || child.Kind == "maybe_boolean") {
				return nil, l.notYet(value, "JSON.stringify nested origin with escaping or non-scalar descendants")
			}
		}
		if err != nil {
			return nil, err
		}
		if child.Kind == "array" && !jsonNativeArrayLiteral(value, child) {
			return nil, l.notYet(value, "JSON.stringify nested arrays without a complete literal origin")
		}
		schema.Fields = append(schema.Fields, ir.JSONField{Name: l.constant(field.Name().Text()), Slot: slot, Schema: child})
	}
	sort.SliceStable(schema.Fields, func(i, j int) bool {
		a, ai := jsonIndex(l.result.Strings[schema.Fields[i].Name])
		b, bi := jsonIndex(l.result.Strings[schema.Fields[j].Name])
		if ai != bi {
			return ai
		}
		return ai && a < b
	})
	return schema, nil
}

func jsonNativeArrayLiteral(node *ast.Node, schema *ir.JSONSchema) bool {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindArrayLiteralExpression {
		return false
	}
	for _, element := range node.AsArrayLiteralExpression().Elements.Nodes {
		if element.Kind == ast.KindSpreadElement || element.Kind == ast.KindOmittedExpression {
			return false
		}
		if schema.Element.Kind == "array" && !jsonNativeArrayLiteral(element, schema.Element) {
			return false
		}
	}
	return true
}

// A function's readonly return view can widen array element slots. Only a directly returned
// constant literal supplies complete metadata until closure results carry their actual schema.
func (l *lowering) jsonArrayCallbackReturn(node *ast.Node, schema *ir.JSONSchema) bool {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindArrowFunction && node.Kind != ast.KindFunctionExpression {
		return false
	}
	body := node.Body()
	if body != nil && body.Kind == ast.KindBlock {
		statements := body.AsBlock().Statements.Nodes
		if len(statements) != 1 || statements[0].Kind != ast.KindReturnStatement {
			return false
		}
		body = statements[0].AsReturnStatement().Expression
	}
	if body == nil || !jsonConstantArray(body) {
		return false
	}
	actual, err := l.jsonType(body, l.checker.GetTypeAtLocation(body), 0)
	return err == nil && jsonSameSchema(actual, schema)
}
func jsonConstantArray(node *ast.Node) bool {
	if node.Kind != ast.KindArrayLiteralExpression {
		return false
	}
	for _, child := range node.AsArrayLiteralExpression().Elements.Nodes {
		child = ast.SkipParentheses(child)
		switch child.Kind {
		case ast.KindArrayLiteralExpression:
			if !jsonConstantArray(child) {
				return false
			}
		case ast.KindNumericLiteral, ast.KindStringLiteral, ast.KindTrueKeyword, ast.KindFalseKeyword:
		default:
			return false
		}
	}
	return true
}
func jsonSameSchema(a, b *ir.JSONSchema) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.Kind == b.Kind && len(a.Fields) == 0 && len(b.Fields) == 0 && jsonSameSchema(a.Element, b.Element)
}
