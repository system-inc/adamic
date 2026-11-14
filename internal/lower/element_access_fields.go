package lower

import (
	"math"
	"reflect"
	"strconv"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// A finite key reads the same own data slots as named access. Ordinary IR calls
// hold the receiver before evaluating the key, including a key that replaces it.
func (l *lowering) finiteElementRead(node *ast.Node, object ir.Expression) (ir.Expression, error) {
	access := node.AsElementAccessExpression()
	names, tests, fields, err := l.finiteKeyFields(node)
	if err != nil {
		return nil, err
	}
	result, err := l.typeOf(node)
	if err != nil {
		return nil, err
	}
	key, err := l.expression(access.ArgumentExpression)
	if err != nil {
		return nil, err
	}
	b := l.libraryArrayBuilder([]ir.Expression{object, key})
	heldObject, heldKey := b.read(b.parameters[0]), b.read(b.parameters[1])
	var read ir.Expression
	for index, name := range names {
		stored, _ := l.representation(l.checker.GetTypeOfSymbol(fields[index]))
		read = ir.Property{Object: heldObject, Name: name, Of: stored, Absent: fields[index].Flags&ast.SymbolFlagsOptional != 0}
		read = fit(read, result)
		if read.Type() != result {
			return nil, l.notYet(node, "a computed field with differently represented alternatives")
		}
		b.body = append(b.body, ir.If{Condition: ir.Binary{Operator: ir.Equal, Left: heldKey, Right: tests[index]}, Then: []ir.Statement{ir.Return{Value: read}}})
	}
	b.body = append(b.body, ir.Panic{Message: ir.StringConstant{Index: l.constant("computed object key outside its proven finite union")}})
	value := b.finish("element_access_fields", read)
	if result.IsReference() && result != ir.Union && !l.includesUndefined(l.checker.GetTypeAtLocation(node)) && !l.acceptsUndefined(node) {
		for _, field := range fields {
			if l.includesUndefined(l.checker.GetTypeOfSymbol(field)) {
				return l.computedFieldPresence(node, value), nil
			}
		}
	}
	return l.defined(node, value), nil
}

// finiteKeyFields is the one proof that a computed key names only own data fields of its receiver, which
// the read (finiteElementRead) and the write (finiteElementWrite) both stand on, so neither can admit a
// key the other refuses. Each member of the key's type must be a literal that names a field (literalKeyName)
// held as one data slot, of one representation in every receiver alternative. It gives each field's name,
// the test that matches the run-time key to it, and its symbol, or the NotYet for the first member it
// can't prove.
func (l *lowering) finiteKeyFields(node *ast.Node) ([]string, []ir.Expression, []*ast.Symbol, error) {
	access := node.AsElementAccessExpression()
	receiver := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(access.Expression))
	keyType := l.checker.GetTypeAtLocation(access.ArgumentExpression)
	members := []*checker.Type{keyType}
	if keyType.Flags()&checker.TypeFlagsUnion != 0 {
		members = keyType.Types()
	}
	names := []string{}
	tests := []ir.Expression{}
	fields := []*ast.Symbol{}
	for _, member := range members {
		name, named := literalKeyName(member)
		if !named && member.Flags()&checker.TypeFlagsNumberLiteral != 0 {
			return nil, nil, nil, l.notYet(node, "a computed numeric key outside exact integer names")
		}
		if !named {
			return nil, nil, nil, l.notYet(node, describe(node))
		}
		var test ir.Expression
		if member.Flags()&checker.TypeFlagsNumberLiteral != 0 {
			// The name is an exact integer's spelling, so it parses back to the number the key holds.
			number, _ := strconv.ParseFloat(name, 64)
			test = ir.NumberConstant{Value: number}
		} else {
			test = ir.StringConstant{Index: l.constant(name)}
		}
		field := l.checker.GetPropertyOfType(receiver, name)
		if field == nil || accessorSymbol(field) || (l.accessorNames[name] && field.Flags&ast.SymbolFlagsOptional != 0) || name == "__proto__" {
			return nil, nil, nil, l.notYet(node, "a computed key without an own data field")
		}
		for _, declaration := range field.Declarations {
			if load.IsLibrary(ast.GetSourceFileOfNode(declaration)) || declaration.Kind == ast.KindMethodDeclaration {
				return nil, nil, nil, l.notYet(node, "a computed prototype or method read")
			}
		}
		stored, known := l.representation(l.checker.GetTypeOfSymbol(field))
		if !known || censusFieldSlotless(stored) || stored == ir.Weak || l.includesNull(l.checker.GetTypeOfSymbol(field)) {
			return nil, nil, nil, l.notYet(node, "a computed field with an unsupported representation")
		}
		if receiver.Flags()&checker.TypeFlagsUnion != 0 {
			for _, alternative := range receiver.Types() {
				own := l.checker.GetPropertyOfType(alternative, name)
				if own == nil {
					return nil, nil, nil, l.notYet(node, "a computed field absent from a receiver alternative")
				}
				held, known := l.representation(l.checker.GetTypeOfSymbol(own))
				if !known || held != stored {
					return nil, nil, nil, l.notYet(node, "a computed field stored differently across receiver alternatives")
				}
			}
		}
		names = append(names, name)
		tests = append(tests, test)
		fields = append(fields, field)
	}
	if len(names) == 0 {
		return nil, nil, nil, l.notYet(node, "a computed key without finite alternatives")
	}
	return names, tests, fields, nil
}

// literalKeyName is the own field name a literal key type names, spelled as the key reads at run time: a
// string literal as written, and a number literal that's an exact integer as String writes it, -0 as 0. A
// number with a fraction or past 2^53, whose spelling this doesn't make, and any other type name none.
func literalKeyName(member *checker.Type) (string, bool) {
	switch {
	case member.Flags()&checker.TypeFlagsStringLiteral != 0:
		return member.AsLiteralType().Value().(string), true
	case member.Flags()&checker.TypeFlagsNumberLiteral != 0:
		number := reflect.ValueOf(member.AsLiteralType().Value()).Float()
		if math.Trunc(number) != number || math.Abs(number) > 9007199254740991 {
			return "", false
		}
		if number == 0 {
			number = 0
		}
		return strconv.FormatFloat(number, 'f', -1, 64), true
	}
	return "", false
}

// finiteElementWrite lowers object[key] = value where object is a fixed object and the key's type is a finite
// union of its own data fields' names, proven as the read proves it (finiteKeyFields): tsc's
// `(customLevels || (customLevels = {}))[level] = Number(customLevel)` (sys.ts:160). JavaScript evaluates
// the receiver, then the key, then the value, and only then stores, so each is held in a local of its own in
// that order, and a value that rebinds the receiver or the key still writes where they were. Each name's arm
// is the store o.name = value makes (fieldStore), so checked views, the cycle finder and Nullish bytes see
// this write as they see that one, and a key outside the union panics loudly, as the read does.
//
// What the read admits and a write can't: an index signature anywhere in the receiver's type is record
// storage's wall, not a fixed object's; a readonly field, which tsc lets a constructor write through this;
// a method a program declares, which setProperty only replaces under its host-member judgment; an optional
// field some producer left out, as absentOptionalWrite judges o.name = value; and a value not held as every
// named field is.
func (l *lowering) finiteElementWrite(target *ast.Node, object ir.Expression, valueNode *ast.Node) ([]ir.Statement, error) {
	access := target.AsElementAccessExpression()
	receiver := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(access.Expression))
	alternatives := []*checker.Type{receiver}
	if receiver.Flags()&checker.TypeFlagsUnion != 0 {
		alternatives = append(alternatives, receiver.Types()...)
	}
	for _, alternative := range alternatives {
		if len(l.checker.GetIndexInfosOfType(alternative)) != 0 {
			return nil, l.notYet(target, "a computed field write to an object with an index signature")
		}
	}
	names, tests, fields, err := l.finiteKeyFields(target)
	if err != nil {
		return nil, err
	}
	stored := make([]ir.Type, len(fields))
	for index, field := range fields {
		if l.checker.IsReadonlySymbol(field) {
			return nil, l.notYet(target, "a computed write to readonly field "+names[index])
		}
		if field.Flags&ast.SymbolFlagsMethod != 0 {
			return nil, l.notYet(target, "replacing a represented method at runtime")
		}
		// A computed access spells no path an if could have tested, so no presence test vouches for it.
		if err := l.absentOptionalFieldWrite(target, access.Expression, field, names[index], false); err != nil {
			return nil, err
		}
		// The registry judges o.name = value against the name's own declared type, where the value's slot
		// here is what tsc writes through a union key, the named fields' types intersected. So each named
		// field is judged as its own slot, the whole registry (refuseInSlot) and lowering's own slot checks
		// (valueSlotStops), as namedWriteStop judges a record's declared property (#b6m00y7). A value whose
		// representation isn't known is judged as a reference, so this never leans on a later stop.
		slot := l.checker.GetNonMissingTypeOfSymbol(field)
		if err := l.refuseInSlot(valueNode, slot); err != nil {
			return nil, err
		}
		held, known := l.representation(l.checker.GetTypeAtLocation(valueNode))
		if err := l.valueSlotStops(valueNode, slot, !known || held.IsReference()); err != nil {
			return nil, err
		}
		if stored[index], err = l.typeOfSymbol(target, field); err != nil {
			return nil, err
		}
	}
	key, err := l.expression(access.ArgumentExpression)
	if err != nil {
		return nil, err
	}
	value, err := l.expression(valueNode)
	if err != nil {
		return nil, err
	}
	// Where every named field is held one way, the value takes that form before it's held, so a literal
	// undefined becomes the field's own missing value, as it does for o.name = undefined.
	shared := true
	for _, of := range stored {
		shared = shared && of == stored[0]
	}
	if shared {
		value = fit(value, stored[0])
	}
	objectLocal := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "object", Type: object.Type(), Function: l.functionIndex})
	keyLocal := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "key", Type: key.Type(), Function: l.functionIndex})
	valueLocal := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "value", Type: value.Type(), Function: l.functionIndex})
	heldKey := ir.Read{Local: keyLocal, Of: key.Type()}
	stores := make([]ir.Statement, len(names))
	for index, field := range fields {
		written, of, declared, err := l.fieldWriteValue(target, field, ir.Read{Local: valueLocal, Of: value.Type()})
		if err != nil {
			return nil, err
		}
		if written.Type() != of {
			return nil, l.notYet(target, "a computed field write whose value isn't held as every named field is")
		}
		stores[index] = l.fieldStore(target, access.Expression, valueNode, field, declared, names[index], ir.Read{Local: objectLocal, Of: object.Type()}, written)
	}
	chain := []ir.Statement{ir.Panic{Message: ir.StringConstant{Index: l.constant("computed object key outside its proven finite union")}}}
	for index := len(stores) - 1; index >= 0; index-- {
		chain = []ir.Statement{ir.If{Condition: ir.Binary{Operator: ir.Equal, Left: heldKey, Right: tests[index]}, Then: []ir.Statement{stores[index]}, Else: chain}}
	}
	body := []ir.Statement{
		ir.Declare{Local: objectLocal, Value: object},
		ir.Declare{Local: keyLocal, Value: key},
		ir.Declare{Local: valueLocal, Value: value},
	}
	return []ir.Statement{ir.Block{Body: append(body, chain...)}}, nil
}

// viewKeyedRead reads a readonly string-index view held as the object it views (readonlyObjectView)
// by a string key, the way tsc's getEnumMembers reads an enum's table, enumObject[name]. It's the
// run-time key read Reflect.get lowers to, ir.DynamicKeyedProperty: an own data slot answers with its
// value, a missing name with undefined and an inherited one as Node's, and the result is a union, so
// typeof value === "number" stays a real check where the view's type says string | number.
func (l *lowering) viewKeyedRead(node *ast.Node, object ir.Expression) (ir.Expression, bool, error) {
	access := node.AsElementAccessExpression()
	if !l.readonlyObjectView(l.checker.GetTypeAtLocation(access.Expression)) {
		return nil, false, nil
	}
	keyType := l.checker.GetTypeAtLocation(access.ArgumentExpression)
	if held, known := l.representation(keyType); !known || held != ir.String || l.includesNull(keyType) || l.includesUndefined(keyType) {
		return nil, true, l.notYet(access.ArgumentExpression, "a readonly view read by a key that isn't a string (symbol keys and key coercion)")
	}
	key, err := l.expression(access.ArgumentExpression)
	if err != nil {
		return nil, true, err
	}
	if key.Type() != ir.String {
		return nil, true, l.notYet(access.ArgumentExpression, "a readonly view read by a key that isn't a string (symbol keys and key coercion)")
	}
	// The read can call a class getter the key names, from inside the runtime, so it arms the same
	// program-wide getter judgment Reflect.get's read does (checkDynamicKeyedGetters).
	if err := l.checkDynamicKeyedGetterSyntax(node); err != nil {
		return nil, true, err
	}
	if l.dynamicKeyedSite == nil {
		l.dynamicKeyedSite = node
	}
	l.graphDynamicSite = node
	return ir.DynamicKeyedProperty{Object: fit(object, ir.Union), Key: key}, true, nil
}

// Indexed narrowing needs the declared field types: a computed access does not
// necessarily have the symbol that the named-read presence helper uses.
func (l *lowering) computedFieldPresence(node *ast.Node, value ir.Expression) ir.Expression {
	message := "undefined where the checker narrowed it away: a call since the narrowing put it back"
	at := node
	for at.Parent != nil && at.Parent.Kind == ast.KindParenthesizedExpression {
		at = at.Parent
	}
	if parent := at.Parent; parent != nil && parent.Kind == ast.KindPropertyAccessExpression && parent.AsPropertyAccessExpression().Expression == at {
		if written := parent.Parent; written != nil && written.Kind == ast.KindBinaryExpression && written.AsBinaryExpression().Left == parent && written.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken {
			return value
		}
		message = "TypeError: Cannot read properties of undefined (reading '" + parent.Name().Text() + "')"
	}
	if parent := at.Parent; parent != nil && parent.Kind == ast.KindElementAccessExpression && parent.AsElementAccessExpression().Expression == at {
		key := ast.SkipParentheses(parent.AsElementAccessExpression().ArgumentExpression)
		if key.Kind == ast.KindNumericLiteral || key.Kind == ast.KindStringLiteral {
			message = "TypeError: Cannot read properties of undefined (reading '" + key.Text() + "')"
		}
	}
	return ir.Defined{Value: value, Message: message}
}

// Until finite reads share the checked member boundary, refuse every matching
// object read in the program, including aliases and reads lowered before the cast.
func (l *lowering) refuseViewElementReads(fields map[string]bool) error {
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return err
	}
	var refused error
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if refused != nil {
			return true
		}
		if node.Kind == ast.KindElementAccessExpression && !ast.IsAssignmentTarget(node) {
			access := node.AsElementAccessExpression()
			key := l.checker.GetTypeAtLocation(access.ArgumentExpression)
			if key.Flags()&checker.TypeFlagsStringLiteral != 0 && fields[key.AsLiteralType().Value().(string)] {
				if of, _ := l.representation(l.checker.GetTypeAtLocation(access.Expression)); of == ir.Object {
					refused = &Refused{Where: l.program.Where(node), What: "view member read by element access is not yet checked", Fix: "use view." + key.AsLiteralType().Value().(string)}
					return true
				}
			}
		}
		node.ForEachChild(visit)
		return refused != nil
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	return refused
}
