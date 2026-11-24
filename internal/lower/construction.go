package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

// A fresh plain literal can reserve its factory's fixed scalar slots. Its asserted
// result remains a checked view; reserving storage proves no value initialized.
// Past that rule, a .ts literal reserves a target of any known field storage where
// castProof has no other road for it (constructionWidened).
func (l *lowering) constructionCast(node *ast.Node) bool {
	source := l.factoryConstructionSource(node)
	if source == nil {
		return false
	}
	target := l.concrete(l.checker.GetTypeAtLocation(node))
	if target.Flags()&checker.TypeFlagsObject == 0 || isClassInstance(target) || len(l.checker.GetSignaturesOfType(target, checker.SignatureKindCall)) != 0 || len(l.checker.GetSignaturesOfType(target, checker.SignatureKindConstruct)) != 0 {
		return false
	}
	fields := l.checker.GetPropertiesOfType(target)
	if len(fields) == 0 {
		return false
	}
	scalar := true
	for _, field := range fields {
		if strings.ContainsRune(field.Name, 0) || strings.HasPrefix(field.Name, "#") || field.Name == "__proto__" {
			return false
		}
		held, known := l.representation(l.checker.GetTypeOfSymbol(field))
		if !known {
			return false
		}
		if field.Flags&ast.SymbolFlagsOptional != 0 || held < ir.Number || held > ir.String || !interfaceScalar(l.checker.GetTypeOfSymbol(field)) {
			scalar = false
		}
	}
	if !scalar && !l.constructionWidened(node, source, target, fields) {
		return false
	}
	provided := map[string]bool{}
	for _, property := range source.AsObjectLiteralExpression().Properties.Nodes {
		if property.Kind != ast.KindPropertyAssignment && property.Kind != ast.KindShorthandPropertyAssignment {
			return false
		}
		name, known := l.methodName(property)
		if !known || name == "__proto__" {
			return false
		}
		provided[name] = true
		field := l.checker.GetPropertyOfType(target, name)
		if field == nil {
			return false
		}
		value := property.Name()
		if property.Kind == ast.KindPropertyAssignment {
			value = property.AsPropertyAssignment().Initializer
		}
		actual := l.checker.GetTypeAtLocation(value)
		wanted := l.checker.GetTypeOfSymbol(field)
		if scalar {
			if !interfaceScalar(actual) || !l.checker.IsTypeAssignableTo(actual, wanted) {
				return false
			}
			continue
		}
		// objectLiteral fits each written value to its declared slot, which would convert a value held
		// another way without a word: it's held as the slot holds it, all the way down, and seen as the
		// slot's type it widens nothing a write through the target could break, or the cast stays stopped.
		have, haveKnown := l.representation(actual)
		held, _ := l.representation(wanted)
		if !haveKnown || have != held || !l.checker.IsTypeAssignableTo(actual, wanted) || !l.sameKeeping(actual, wanted, map[[2]*checker.Type]bool{}) || !l.castAssignableAt(value, actual, wanted) {
			return false
		}
	}
	return len(provided) < len(fields)
}

// constructionWidened is whether a .ts literal may reserve a target whose fields aren't all required
// scalars (Platforms' ruling on #b5w3ycg, generic over any construction). It's admitted only where castProof
// would otherwise stop it as castSideways, so no cast that has a road today changes road: the literal's own
// type isn't assignable to the target, which castProof would take as an upcast first, and the target isn't
// assignable to it, which is a structural view. In .a the sideways cast keeps its refusal. A closure or Weak
// slot, a method and an accessor stay stopped: a call or a Weak read reaches its slot past the readiness
// check an unwritten required field's read stops at.
func (l *lowering) constructionWidened(node, source *ast.Node, target *checker.Type, fields []*ast.Symbol) bool {
	if !l.checkedAssertionSource(node) {
		return false
	}
	literal := l.concrete(l.checker.GetTypeAtLocation(source))
	if l.checker.IsTypeAssignableTo(literal, target) || l.checker.IsTypeAssignableTo(l.checker.GetWidenedType(literal), target) || l.objectCastCause(literal, target) != castSideways {
		return false
	}
	for _, field := range fields {
		if field.Flags&ast.SymbolFlagsMethod != 0 || accessorSymbol(field) {
			return false
		}
		switch held, _ := l.representation(l.checker.GetTypeOfSymbol(field)); held {
		case ir.Number, ir.Boolean, ir.String, ir.Object, ir.Array, ir.Union, ir.MaybeNumber, ir.MaybeBoolean:
		default:
			return false
		}
	}
	return true
}

// unwrittenField lays a target field the literal doesn't write, absent either way. A required one is
// Uninitialized too, so a read of it stops loudly (#dm8exc2); an optional one is really absent, and a read
// of it gives undefined, as Node's does.
func (l *lowering) unwrittenField(field *ast.Symbol) ir.Field {
	held, _ := l.representation(l.checker.GetTypeOfSymbol(field))
	zero := ir.Expression(zeroValue(held))
	if held.IsReference() {
		zero = ir.Undefined{Of: held}
	}
	return ir.Field{Name: field.Name, Value: zero, Absent: true, Uninitialized: field.Flags&ast.SymbolFlagsOptional == 0}
}

func (l *lowering) constructLiteral(node *ast.Node) (ir.Expression, error) {
	target := l.concrete(l.checker.GetTypeAtLocation(node))
	source := ast.SkipParentheses(node.AsAsExpression().Expression)
	if source.Kind == ast.KindCallExpression {
		return l.expression(source)
	}
	value, err := l.objectLiteral(source)
	if err != nil {
		return nil, err
	}
	literal := value.(ir.ObjectLiteral)
	present := map[string]bool{}
	for _, field := range literal.Fields {
		present[field.Name] = true
	}
	for _, field := range l.checker.GetPropertiesOfType(target) {
		if present[field.Name] {
			continue
		}
		literal.Fields = append(literal.Fields, l.unwrittenField(field))
	}
	return literal, nil
}

// A literal string key names the same fixed slot as a dot access. It is not a
// dynamic index-signature write and cannot add a slot the declared layout lacks.
func (l *lowering) constructionIndexWrite(target, valueNode *ast.Node) ([]ir.Statement, bool, error) {
	access := target.AsElementAccessExpression()
	key := ast.SkipParentheses(access.ArgumentExpression)
	if key.Kind != ast.KindStringLiteral || strings.ContainsRune(key.Text(), 0) || strings.HasPrefix(key.Text(), "#") || key.Text() == "__proto__" {
		return nil, false, nil
	}
	receiver := l.checker.GetTypeAtLocation(access.Expression)
	held, _ := l.representation(receiver)
	if held != ir.Object || isClassInstance(receiver) {
		return nil, false, nil
	}
	field := l.checker.GetPropertyOfType(receiver, key.Text())
	if field == nil || !interfaceScalar(l.checker.GetTypeOfSymbol(field)) {
		return nil, false, nil
	}
	object, err := l.expression(access.Expression)
	if err != nil {
		return nil, true, err
	}
	value, err := l.expression(valueNode)
	if err != nil {
		return nil, true, err
	}
	declared, known := l.representation(l.checker.GetTypeOfSymbol(field))
	if !known || value.Type() != declared {
		return nil, true, l.notYet(target, "a string-key field write requiring representation conversion")
	}
	return []ir.Statement{ir.SetProperty{Object: object, Name: key.Text(), Value: value, Site: l.writeSite(access.Expression)}}, true, nil
}

// A const initialized by one factory return has a closed physical layout. This
// proves JSON's descriptor, not the values of reserved or unset fields.
func (l *lowering) constructionJSON(node *ast.Node) (*ir.JSONSchema, bool) {
	if !ast.IsIdentifier(node) {
		return nil, false
	}
	symbol := l.symbol(node)
	if symbol == nil || len(symbol.Declarations) != 1 || symbol.Declarations[0].Kind != ast.KindVariableDeclaration {
		return nil, false
	}
	declaration := symbol.Declarations[0]
	if declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
		return nil, false
	}
	initializer := declaration.AsVariableDeclaration().Initializer
	if initializer == nil {
		return nil, false
	}
	initializer = ast.SkipParentheses(initializer)
	var assertion *ast.Node
	if initializer.Kind == ast.KindAsExpression {
		assertion = initializer
	}
	if initializer.Kind == ast.KindCallExpression {
		callee := l.symbol(ast.SkipParentheses(initializer.AsCallExpression().Expression))
		if callee == nil || len(callee.Declarations) != 1 || callee.Declarations[0].Kind != ast.KindFunctionDeclaration {
			return nil, false
		}
		body := callee.Declarations[0].Body()
		if body == nil || body.Kind != ast.KindBlock || len(body.AsBlock().Statements.Nodes) != 1 {
			return nil, false
		}
		statement := body.AsBlock().Statements.Nodes[0]
		if statement.Kind != ast.KindReturnStatement {
			return nil, false
		}
		assertion = statement.AsReturnStatement().Expression
	}
	if assertion == nil || assertion.Kind != ast.KindAsExpression || !l.constructionCast(assertion) {
		return nil, false
	}
	names := []string{}
	present := map[string]bool{}
	source := l.factoryConstructionSource(assertion)
	for _, property := range source.AsObjectLiteralExpression().Properties.Nodes {
		name, _ := l.methodName(property)
		names = append(names, name)
		present[name] = true
	}
	target := l.concrete(l.checker.GetTypeAtLocation(assertion))
	for _, field := range l.checker.GetPropertiesOfType(target) {
		if !present[field.Name] {
			names = append(names, field.Name)
		}
	}
	schema := &ir.JSONSchema{Kind: "object"}
	for index, name := range names {
		field := l.checker.GetPropertyOfType(target, name)
		held, _ := l.representation(l.checker.GetTypeOfSymbol(field))
		kind := map[ir.Type]string{ir.Number: "number", ir.Boolean: "boolean", ir.String: "string"}[held]
		if kind == "" {
			// A widened construction's other storage has no descriptor here, and an empty kind is none.
			return nil, false
		}
		schema.Fields = append(schema.Fields, ir.JSONField{Name: l.constant(name), Slot: index, Schema: &ir.JSONSchema{Kind: kind}})
	}
	return schema, true
}

// Only a single direct literal return fixes the allocation provenance. Branches,
// indirect calls, aliases, mutations and generic helper bodies remain checked views.
func (l *lowering) factoryConstructionSource(node *ast.Node) *ast.Node {
	source := ast.SkipParentheses(node.AsAsExpression().Expression)
	if source.Kind == ast.KindObjectLiteralExpression {
		return source
	}
	if source.Kind != ast.KindCallExpression {
		return nil
	}
	callee := l.symbol(ast.SkipParentheses(source.AsCallExpression().Expression))
	if callee == nil || len(callee.Declarations) != 1 || callee.Declarations[0].Kind != ast.KindFunctionDeclaration {
		return nil
	}
	body := callee.Declarations[0].Body()
	if body == nil || body.Kind != ast.KindBlock || len(body.AsBlock().Statements.Nodes) != 1 {
		return nil
	}
	statement := body.AsBlock().Statements.Nodes[0]
	if statement.Kind != ast.KindReturnStatement || statement.AsReturnStatement().Expression == nil {
		return nil
	}
	literal := ast.SkipParentheses(statement.AsReturnStatement().Expression)
	if literal.Kind != ast.KindObjectLiteralExpression {
		return nil
	}
	return literal
}

// Reserve each reviewed target at the original base factory allocation. These
// slots remain absent; changing a target never fabricates an initialized field.
func (l *lowering) reserveFactoryFields(node *ast.Node, literal ir.ObjectLiteral) (ir.ObjectLiteral, error) {
	if literal.Record || literal.Spread != nil {
		return literal, nil
	}
	present := map[string]ir.Type{}
	for _, field := range literal.Fields {
		present[field.Name] = field.Value.Type()
	}
	// reserved is where each slot laid here sits, so a target that requires a slot another target laid
	// optional makes it Uninitialized: a read through the one that requires it stops, as constructLiteral's.
	reserved := map[string]int{}
	var failure error
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		if failure != nil {
			return true
		}
		if n.Kind == ast.KindAsExpression && l.factoryConstructionSource(n) == node && l.constructionCast(n) {
			for _, field := range l.checker.GetPropertiesOfType(l.concrete(l.checker.GetTypeAtLocation(n))) {
				held, _ := l.representation(l.checker.GetTypeOfSymbol(field))
				laid := l.unwrittenField(field)
				if old, found := present[field.Name]; found {
					if old != held {
						failure = l.notYet(n, "factory target layouts with incompatible field storage")
					} else if index, ours := reserved[field.Name]; ours && laid.Uninitialized {
						literal.Fields[index].Uninitialized = true
					}
					continue
				}
				reserved[field.Name] = len(literal.Fields)
				literal.Fields = append(literal.Fields, laid)
				present[field.Name] = held
			}
		}
		return n.ForEachChild(visit)
	}
	for _, file := range l.program.Files() {
		file.AsNode().ForEachChild(visit)
	}
	return literal, failure
}
