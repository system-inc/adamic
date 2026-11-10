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
		return nil, true, &Refused{Where: l.program.Where(node), What: "JSON.parse: its result's type can't be proven from the text", Fix: "a checked parse against a declared type is a later design; construct typed values explicitly for now"}
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
			return nil, true, l.notYet(args[1], "JSON.stringify replacer functions (the callback must have a proven type for every visited value and its holder)")
		}
		result.Replacer, result.ReplacerSchema, err = l.jsonInput(args[1])
		if err != nil {
			return nil, true, err
		}
		k := result.ReplacerSchema.Kind
		if k != "array" && k != "tuple" && k != "null" && k != "undefined" {
			return nil, true, l.notYet(args[1], "JSON.stringify replacer other than a key array, null or undefined")
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
	l.recordWholeViewArrayUse(node, result.Value)
	return result, true, nil
}
func jsonScalar(s *ir.JSONSchema) bool {
	return s.Kind == "number" || s.Kind == "boolean" || s.Kind == "string" || s.Kind == "undefined" || s.Kind == "null" || s.Kind == "union" || s.Kind == "maybe_number" || s.Kind == "maybe_boolean"
}

// Direct literals establish their complete shape, and preserve evaluation order. Anything read
// through an object type is printed from its allocation at run time (jsonReference): that type can
// hide additional fields, and a cast can make it lie about the ones it shows.
func (l *lowering) jsonInput(node *ast.Node) (ir.Expression, *ir.JSONSchema, error) {
	n := ast.SkipParentheses(node)
	if l.uninitializedInitializer(n) && l.placeholderDeclaration(n) {
		// This temporary field has a known nullish arm and never escapes the serializer.
		if _, null := l.placeholderInitialValue(n, ir.Union).(ir.Box); null {
			return ir.JSONNull{}, &ir.JSONSchema{Kind: "null"}, nil
		}
		return ir.Undefined{}, &ir.JSONSchema{Kind: "undefined"}, nil
	}
	if n.Kind == ast.KindNullKeyword {
		return ir.JSONNull{}, &ir.JSONSchema{Kind: "null"}, nil
	}
	if n.Kind == ast.KindObjectLiteralExpression && l.recordLiteralElement(n) != nil {
		value, err := l.expression(n)
		if err != nil {
			return nil, nil, err
		}
		proven := l.slotType(n)
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
			value, child, err := l.jsonLiteralField(v)
			if err != nil {
				return err
			}
			if value.Type() == ir.MaybeBoolean {
				return l.notYet(v, "JSON.stringify a literal field with a two-word representation")
			}
			schema.Fields = append(schema.Fields, ir.JSONField{Name: l.constant(name), Slot: len(literal.Fields), Schema: child})
			literal.Fields = append(literal.Fields, ir.Field{Name: name, Value: value, Uninitialized: l.uninitializedInitializer(v), Unset: l.uninitializedInitializer(v)})
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
				if f.Kind != ast.KindPropertyAssignment && f.Kind != ast.KindShorthandPropertyAssignment {
					return nil, nil, l.notYet(f, "JSON.stringify a literal with spread or methods")
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
				value := f
				if f.Kind == ast.KindPropertyAssignment {
					value = f.AsPropertyAssignment().Initializer
				}
				if err := add(name, value); err != nil {
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
	if schema, known := l.constructionJSON(n); known {
		value, err := l.expression(node)
		return value, schema, err
	}
	proven := l.checker.GetTypeAtLocation(node)
	schema, err := l.jsonType(node, proven, 0)
	if err != nil {
		return nil, nil, err
	}
	value, err := l.expression(node)
	if err != nil {
		return nil, nil, err
	}
	return l.jsonReferenceValue(value, proven, schema), schema, nil
}
func jsonIndex(name string) (uint64, bool) {
	n, err := strconv.ParseUint(name, 10, 32)
	return n, err == nil && n < 4294967295 && strconv.FormatUint(n, 10) == name
}
func (l *lowering) jsonType(node *ast.Node, t *checker.Type, depth int) (*ir.JSONSchema, error) {
	if depth == 0 && l.censusCallReturnsUndefined(node) {
		return &ir.JSONSchema{Kind: "undefined"}, nil
	}
	if depth > 64 {
		return nil, l.notYet(node, "JSON.stringify recursive array types")
	}
	if t.Flags()&checker.TypeFlagsUndefined != 0 {
		return &ir.JSONSchema{Kind: "undefined"}, nil
	}
	if t.Flags()&checker.TypeFlagsNever != 0 {
		return &ir.JSONSchema{Kind: "undefined"}, nil
	}
	if t.Flags()&checker.TypeFlagsNull != 0 {
		return &ir.JSONSchema{Kind: "null"}, nil
	}
	of, known := l.representation(t)
	if !known {
		return nil, l.notYet(node, "JSON.stringify a value of type "+l.checker.TypeToString(t))
	}
	kinds := map[ir.Type]string{ir.Number: "number", ir.Boolean: "boolean", ir.String: "string", ir.Map: "map", ir.Closure: "function", ir.MaybeNumber: "maybe_number", ir.MaybeBoolean: "maybe_boolean", ir.Union: "union"}
	if of == ir.Record {
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
		element := l.elementTypeOfArray(l.checker.GetNonNullableType(t))
		if _, extraElement, known := l.nodeArrayLayoutOf(l.checker.GetNonNullableType(t)); known {
			element = extraElement
		}
		if element == nil {
			return nil, l.notYet(node, "JSON.stringify an array without a proven element type")
		}
		child, err := l.jsonType(node, element, depth+1)
		return &ir.JSONSchema{Kind: "array", Element: child}, err
	}
	if of == ir.Object || of == ir.Weak {
		return l.jsonReference(node, t, of)
	}
	reference := false
	if t.Flags()&checker.TypeFlagsUnion != 0 {
		for _, m := range t.Types() {
			child, err := l.jsonType(node, m, depth+1)
			if err != nil {
				return nil, err
			}
			if of == ir.Union && (child.Kind == "array" || child.Kind == "object" || child.Kind == "tuple" || child.Kind == "record") {
				return nil, l.notYet(node, "JSON.stringify a union containing containers without runtime element metadata")
			}
			reference = reference || child.Kind == "reference"
		}
	}
	if of == ir.Union && reference {
		// A union holding an object reference is read by its heap header like any union, and the
		// object it holds is described from its allocation.
		return &ir.JSONSchema{Kind: "reference"}, nil
	}
	kind, ok := kinds[of]
	if !ok {
		return nil, l.notYet(node, "JSON.stringify this representation")
	}
	return &ir.JSONSchema{Kind: kind}, nil
}

// A shorthand name is a property symbol, not the binding it reads. Resolve
// the value symbol just as ordinary object-literal lowering does. Its declared
// type describes the stored representation even when the occurrence is narrowed.
func (l *lowering) jsonLiteralField(node *ast.Node) (ir.Expression, *ir.JSONSchema, error) {
	if node.Kind != ast.KindShorthandPropertyAssignment {
		return l.jsonInput(node)
	}
	value, err := l.shorthand(node)
	if err != nil {
		return nil, nil, err
	}
	symbol := l.checker.GetShorthandAssignmentValueSymbol(node)
	if symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
		symbol = l.checker.GetAliasedSymbol(symbol)
	}
	if symbol == nil {
		return nil, nil, l.notYet(node, "JSON.stringify shorthand without a value binding")
	}
	proven := l.checker.GetTypeOfSymbol(symbol)
	schema, err := l.jsonType(node.Name(), proven, 0)
	if err != nil {
		return nil, nil, err
	}
	return l.jsonReferenceValue(value, proven, schema), schema, nil
}

// jsonReferenceRead is an object reference JSON.stringify prints, and the type it was proven at,
// judged by jsonReferenceReach once every shape the program makes is known.
type jsonReferenceRead struct {
	node   *ast.Node
	proven *checker.Type
}

// jsonReference admits an object reference, which json_objects.c prints from its allocation at run
// time: its actual names in Node's order, each slot by the representation byte its last write left,
// never by this type, which a cast can make lie. What the type can reach is judged after lowering
// (jsonReferenceReach), when the shapes it may really be are all known.
func (l *lowering) jsonReference(node *ast.Node, t *checker.Type, of ir.Type) (*ir.JSONSchema, error) {
	if of == ir.Weak {
		return nil, l.notYet(node, "JSON.stringify a Weak reference (Node follows it, so what it reaches could close a cycle)")
	}
	present := l.checker.GetNonNullableType(t)
	members := []*checker.Type{present}
	if present.Flags()&checker.TypeFlagsUnion != 0 {
		members = present.Types()
	}
	for _, member := range members {
		if member.Flags()&checker.TypeFlagsObject == 0 || checker.IsTupleType(member) {
			return nil, l.notYet(node, "JSON.stringify a value of type "+l.checker.TypeToString(t)+" held as an object reference")
		}
	}
	l.jsonReferenceReads = append(l.jsonReferenceReads, jsonReferenceRead{node: node, proven: t})
	return &ir.JSONSchema{Kind: "reference"}, nil
}

// jsonReferenceValue holds an object reference where the encoder reads a union. An object held as
// T | null keeps null as a null pointer (representation), which a union reads as undefined, so that
// one becomes the union's null; T and T | undefined are the union as they are.
func (l *lowering) jsonReferenceValue(value ir.Expression, proven *checker.Type, schema *ir.JSONSchema) ir.Expression {
	if schema.Kind != "reference" || value.Type() != ir.Object {
		return value
	}
	if l.includesNull(proven) {
		return ir.Logical{Left: value, Right: ir.Box{Value: ir.Null{Of: ir.Object}}, Of: ir.Union, KeepTruthy: true}
	}
	return fit(value, ir.Union)
}

// jsonReferenceReach refuses what an object reference JSON.stringify prints can reach and the
// runtime can't print as Node does, judged as the cycle finder judges reach: through every field,
// element and value, and every object type the program makes that can be seen as a type reached,
// with fields that type doesn't show. A function is left, as Node omits it. It runs after
// graphTypes, so a graph type is known; json_objects.c stops at run time on what a type can't show.
func (f *cycleFinder) jsonReferenceReach() error {
	for _, read := range f.l.jsonReferenceReads {
		seen := map[*checker.Type]bool{}
		queue := []*checker.Type{read.proven}
		for len(queue) != 0 {
			proven := queue[len(queue)-1]
			queue = queue[:len(queue)-1]
			if proven == nil || seen[proven] {
				continue
			}
			seen[proven] = true
			reason, next := f.jsonReach(proven)
			if reason != "" {
				return f.l.notYet(read.node, "JSON.stringify "+reason)
			}
			queue = append(queue, next...)
		}
	}
	return nil
}

// jsonReach is why JSON.stringify can't print what holds a value of this type, or else the types
// it holds.
func (f *cycleFinder) jsonReach(proven *checker.Type) (string, []*checker.Type) {
	l := f.l
	name := l.checker.TypeToString(proven)
	if f.weak(proven) {
		return "a value that can reach a Weak reference, " + name + " (Node follows it, so what it reaches could close a cycle)", nil
	}
	if l.result.GraphTypes[int(proven.Id())] {
		return "a value that can reach " + name + ", a graph type whose values may form a cycle", nil
	}
	if proven.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		return "", proven.Types()
	}
	if proven.Flags()&checker.TypeFlagsObject == 0 {
		return "", nil
	}
	if len(l.checker.GetSignaturesOfType(proven, checker.SignatureKindConstruct)) != 0 {
		return "a value that can reach " + name + ", a class itself", nil
	}
	if f.isFunction(proven) {
		return "", nil
	}
	if element := l.recordElement(proven); element != nil {
		return "", []*checker.Type{element}
	}
	if l.isArrayType(proven) || checker.IsTupleType(proven) {
		return "", l.checker.GetTypeArguments(proven)
	}
	if l.jsonLibrary(proven, map[*checker.Type]bool{}) {
		return "a value that can reach " + name + ", a library value whose own fields aren't what Node prints", nil
	}
	// The checker's properties of a class include what its bases declare: a toJSON a base declares is
	// on the prototype chain, and Node calls it as it calls the leaf's own.
	if l.checker.GetPropertyOfType(proven, "toJSON") != nil {
		return "a value that can reach " + name + ", which has a toJSON member", nil
	}
	next := []*checker.Type{}
	for _, property := range l.checker.GetPropertiesOfType(proven) {
		if accessorSymbol(property) && literalAccessor(property) {
			return "a value that can reach " + name + ", whose own getter Node calls", nil
		}
	}
	for _, field := range f.fields(proven) {
		if declared := l.checker.GetNonMissingTypeOfSymbol(field); jsonNullablePointer(l, declared) {
			return "a value that can reach " + name + "." + field.Name + ", a field of type " + l.checker.TypeToString(declared) + " held as a pointer, whose null and undefined are both a null pointer when a variable wrote it (lifted by #ckxcnr6)", nil
		}
		next = append(next, l.checker.GetTypeOfSymbol(field))
	}
	for _, shape := range f.shapeCandidates(proven) {
		if shape != proven && l.checker.IsTypeAssignableTo(shape, proven) {
			next = append(next, shape)
		}
	}
	return "", next
}

// jsonNullablePointer reports whether a field's type is held as a bare pointer (an object, an array, a
// map, a function, a promise, a typed array) and admits null or undefined. Both are then a null pointer,
// and the slot's byte tells them apart only when a literal null or undefined wrote it
// (emit_objects.go's fieldRepresentation), so json_objects.c couldn't say which Node prints. A string, a
// union, a weak link and a record are boxed where null is possible, so their null pointer is undefined
// alone, and an optional field's absence is its presence, never a pointer: GetNonMissingTypeOfSymbol
// leaves the missing type out.
func jsonNullablePointer(l *lowering, declared *checker.Type) bool {
	if !l.includesNull(declared) && !l.includesUndefined(declared) {
		return false
	}
	of, known := l.representation(declared)
	if !known {
		return false
	}
	switch of {
	case ir.Object, ir.Array, ir.Map, ir.Closure, ir.Promise, ir.Uint8Array, ir.Int32Array, ir.Float64Array, ir.Uint16Array, ir.Generator:
		return true
	}
	return false
}

// jsonLibrary reports whether a type is the library's, the prelude's or Node's, or a class extending one:
// such a value's storage is its own (a Date's, an Error's, a Map's), never the own fields Node prints. An
// interface extending one declares no storage; the allocation behind it is judged as its own shape.
func (l *lowering) jsonLibrary(proven *checker.Type, seen map[*checker.Type]bool) bool {
	if seen[proven] {
		return false
	}
	seen[proven] = true
	if symbol := l.unbranded(proven).Symbol(); symbol != nil && libraryDeclared(symbol) {
		return true
	}
	if !isClassInstance(proven) {
		return false
	}
	for _, base := range l.classBases(proven) {
		if l.jsonLibrary(base, seen) {
			return true
		}
	}
	return false
}

// literalAccessor reports whether an accessor is an object literal's own, which Node calls when it
// prints the object; a class's is on its prototype, which Node leaves.
func literalAccessor(property *ast.Symbol) bool {
	for _, declaration := range property.Declarations {
		if declaration.Parent != nil && declaration.Parent.Kind == ast.KindObjectLiteralExpression {
			return true
		}
	}
	return false
}
