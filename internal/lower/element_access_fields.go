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
func (lowering *lowering) finiteElementRead(node *ast.Node, object ir.Expression) (ir.Expression, error) {
	access := node.AsElementAccessExpression()
	receiver := lowering.checker.GetNonNullableType(lowering.checker.GetTypeAtLocation(access.Expression))
	keyType := lowering.checker.GetTypeAtLocation(access.ArgumentExpression)
	members := []*checker.Type{keyType}
	if keyType.Flags()&checker.TypeFlagsUnion != 0 {
		members = keyType.Types()
	}
	names := []string{}
	tests := []ir.Expression{}
	fields := []*ast.Symbol{}
	for _, member := range members {
		var name string
		var test ir.Expression
		switch {
		case member.Flags()&checker.TypeFlagsStringLiteral != 0:
			name = member.AsLiteralType().Value().(string)
			test = ir.StringConstant{Index: lowering.constant(name)}
		case member.Flags()&checker.TypeFlagsNumberLiteral != 0:
			number := reflect.ValueOf(member.AsLiteralType().Value()).Float()
			if math.Trunc(number) != number || math.Abs(number) > 9007199254740991 {
				return nil, lowering.notYet(node, "a computed numeric key outside exact integer names")
			}
			if number == 0 {
				number = 0
			}
			name = strconv.FormatFloat(number, 'f', -1, 64)
			test = ir.NumberConstant{Value: number}
		default:
			return nil, lowering.notYet(node, describe(node))
		}
		field := lowering.checker.GetPropertyOfType(receiver, name)
		if field == nil ||
			accessorSymbol(field) ||
			(lowering.accessorNames[name] && field.Flags&ast.SymbolFlagsOptional != 0) ||
			name == "__proto__" {
			return nil, lowering.notYet(node, "a computed key without an own data field")
		}
		for _, declaration := range field.Declarations {
			if load.IsLibrary(ast.GetSourceFileOfNode(declaration)) || declaration.Kind == ast.KindMethodDeclaration {
				return nil, lowering.notYet(node, "a computed prototype or method read")
			}
		}
		stored, known := lowering.representation(lowering.checker.GetTypeOfSymbol(field))
		if !known ||
			censusFieldSlotless(stored) ||
			stored == ir.Weak ||
			lowering.includesNull(lowering.checker.GetTypeOfSymbol(field)) {
			return nil, lowering.notYet(node, "a computed field with an unsupported representation")
		}
		if receiver.Flags()&checker.TypeFlagsUnion != 0 {
			for _, alternative := range receiver.Types() {
				own := lowering.checker.GetPropertyOfType(alternative, name)
				if own == nil {
					return nil, lowering.notYet(node, "a computed field absent from a receiver alternative")
				}
				held, known := lowering.representation(lowering.checker.GetTypeOfSymbol(own))
				if !known || held != stored {
					return nil, lowering.notYet(
						node,
						"a computed field stored differently across receiver alternatives",
					)
				}
			}
		}
		names = append(names, name)
		tests = append(tests, test)
		fields = append(fields, field)
	}
	if len(names) == 0 {
		return nil, lowering.notYet(node, "a computed key without finite alternatives")
	}
	result, typeOfError := lowering.typeOf(node)
	if typeOfError != nil {
		return nil, typeOfError
	}
	key, expressionError := lowering.expression(access.ArgumentExpression)
	if expressionError != nil {
		return nil, expressionError
	}
	b := lowering.libraryArrayBuilder([]ir.Expression{object, key})
	heldObject, heldKey := b.read(b.parameters[0]), b.read(b.parameters[1])
	var read ir.Expression
	for index, name := range names {
		stored, _ := lowering.representation(lowering.checker.GetTypeOfSymbol(fields[index]))
		read = ir.Property{
			Object: heldObject,
			Name:   name,
			Of:     stored,
			Absent: fields[index].Flags&ast.SymbolFlagsOptional != 0,
		}
		read = fit(read, result)
		if read.Type() != result {
			return nil, lowering.notYet(node, "a computed field with differently represented alternatives")
		}
		b.body = append(b.body, ir.If{
			Condition: ir.Binary{Operator: ir.Equal, Left: heldKey, Right: tests[index]},
			Then:      []ir.Statement{ir.Return{Value: read}},
		})
	}
	b.body = append(
		b.body,
		ir.Panic{
			Message: ir.StringConstant{Index: lowering.constant("computed object key outside its proven finite union")},
		},
	)
	value := b.finish("element_access_fields", read)
	if result.IsReference() &&
		result != ir.Union &&
		!lowering.includesUndefined(lowering.checker.GetTypeAtLocation(node)) &&
		!lowering.acceptsUndefined(node) {
		for _, field := range fields {
			if lowering.includesUndefined(lowering.checker.GetTypeOfSymbol(field)) {
				return lowering.computedFieldPresence(node, value), nil
			}
		}
	}
	return lowering.defined(node, value), nil
}

// viewKeyedRead reads a readonly string-index view held as the object it views (readonlyObjectView)
// by a string key, the way tsc's getEnumMembers reads an enum's table, enumObject[name]. It's the
// run-time key read Reflect.get lowers to, ir.DynamicKeyedProperty: an own data slot answers with its
// value, a missing name with undefined and an inherited one as Node's, and the result is a union, so
// typeof value === "number" stays a real check where the view's type says string | number.
func (lowering *lowering) viewKeyedRead(node *ast.Node, object ir.Expression) (ir.Expression, bool, error) {
	access := node.AsElementAccessExpression()
	if !lowering.readonlyObjectView(lowering.checker.GetTypeAtLocation(access.Expression)) {
		return nil, false, nil
	}
	keyType := lowering.checker.GetTypeAtLocation(access.ArgumentExpression)
	if held, known := lowering.representation(
		keyType,
	); !known || held != ir.String || lowering.includesNull(keyType) || lowering.includesUndefined(keyType) {
		return nil, true, lowering.notYet(
			access.ArgumentExpression,
			"a readonly view read by a key that isn't a string (symbol keys and key coercion)",
		)
	}
	key, expressionError := lowering.expression(access.ArgumentExpression)
	if expressionError != nil {
		return nil, true, expressionError
	}
	if key.Type() != ir.String {
		return nil, true, lowering.notYet(
			access.ArgumentExpression,
			"a readonly view read by a key that isn't a string (symbol keys and key coercion)",
		)
	}
	// The read can call a class getter the key names, from inside the runtime, so it arms the same
	// program-wide getter judgment Reflect.get's read does (checkDynamicKeyedGetters).
	if checkDynamicKeyedGetterSyntaxError := lowering.checkDynamicKeyedGetterSyntax(
		node,
	); checkDynamicKeyedGetterSyntaxError != nil {
		return nil, true, checkDynamicKeyedGetterSyntaxError
	}
	if lowering.dynamicKeyedSite == nil {
		lowering.dynamicKeyedSite = node
	}
	lowering.graphDynamicSite = node
	return ir.DynamicKeyedProperty{Object: fit(object, ir.Union), Key: key}, true, nil
}

// Indexed narrowing needs the declared field types: a computed access does not
// necessarily have the symbol that the named-read presence helper uses.
func (lowering *lowering) computedFieldPresence(node *ast.Node, value ir.Expression) ir.Expression {
	message := "undefined where the checker narrowed it away: a call since the narrowing put it back"
	at := node
	for at.Parent != nil && at.Parent.Kind == ast.KindParenthesizedExpression {
		at = at.Parent
	}
	if parent := at.Parent; parent != nil &&
		parent.Kind == ast.KindPropertyAccessExpression &&
		parent.AsPropertyAccessExpression().Expression == at {
		if written := parent.Parent; written != nil &&
			written.Kind == ast.KindBinaryExpression &&
			written.AsBinaryExpression().Left == parent &&
			written.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken {
			return value
		}
		message = "TypeError: Cannot read properties of undefined (reading '" + parent.Name().Text() + "')"
	}
	if parent := at.Parent; parent != nil &&
		parent.Kind == ast.KindElementAccessExpression &&
		parent.AsElementAccessExpression().Expression == at {
		key := ast.SkipParentheses(parent.AsElementAccessExpression().ArgumentExpression)
		if key.Kind == ast.KindNumericLiteral || key.Kind == ast.KindStringLiteral {
			message = "TypeError: Cannot read properties of undefined (reading '" + key.Text() + "')"
		}
	}
	return ir.Defined{Value: value, Message: message}
}

// Until finite reads share the checked member boundary, refuse every matching
// object read in the program, including aliases and reads lowered before the cast.
func (lowering *lowering) refuseViewElementReads(fields map[string]bool) error {
	modules, moduleOrderError := lowering.moduleOrder(lowering.program.Files()[0])
	if moduleOrderError != nil {
		return moduleOrderError
	}
	var refused error
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if refused != nil {
			return true
		}
		if node.Kind == ast.KindElementAccessExpression && !ast.IsAssignmentTarget(node) {
			access := node.AsElementAccessExpression()
			key := lowering.checker.GetTypeAtLocation(access.ArgumentExpression)
			if key.Flags()&checker.TypeFlagsStringLiteral != 0 && fields[key.AsLiteralType().Value().(string)] {
				if of, _ := lowering.representation(
					lowering.checker.GetTypeAtLocation(access.Expression),
				); of == ir.Object {
					refused = &Refused{
						Where: lowering.program.Where(node),
						What:  "view member read by element access is not yet checked",
						Fix:   "use view." + key.AsLiteralType().Value().(string),
					}
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
