package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

// A fresh plain literal can reserve its factory's fixed scalar slots. Its asserted
// result remains a checked view; reserving storage proves no value initialized.
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
	for _, field := range fields {
		if strings.ContainsRune(field.Name, 0) || strings.HasPrefix(field.Name, "#") || field.Name == "__proto__" {
			return false
		}
		held, known := l.representation(l.checker.GetTypeOfSymbol(field))
		if field.Flags&ast.SymbolFlagsOptional != 0 || !known || held < ir.Number || held > ir.String || !interfaceScalar(l.checker.GetTypeOfSymbol(field)) {
			return false
		}
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
		if !interfaceScalar(actual) || !l.checker.IsTypeAssignableTo(actual, wanted) {
			return false
		}
	}
	return len(provided) < len(fields)
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
		held, _ := l.representation(l.checker.GetTypeOfSymbol(field))
		zero := ir.Expression(zeroValue(held))
		if held.IsReference() {
			zero = ir.Undefined{Of: held}
		}
		literal.Fields = append(literal.Fields, ir.Field{Name: field.Name, Value: zero, Absent: true, Uninitialized: true})
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
	var failure error
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		if failure != nil {
			return true
		}
		if n.Kind == ast.KindAsExpression && l.factoryConstructionSource(n) == node && l.constructionCast(n) {
			for _, field := range l.checker.GetPropertiesOfType(l.concrete(l.checker.GetTypeAtLocation(n))) {
				held, _ := l.representation(l.checker.GetTypeOfSymbol(field))
				if old, found := present[field.Name]; found {
					if old != held {
						failure = l.notYet(n, "factory target layouts with incompatible field storage")
					}
					continue
				}
				zero := ir.Expression(zeroValue(held))
				if held.IsReference() {
					zero = ir.Undefined{Of: held}
				}
				literal.Fields = append(literal.Fields, ir.Field{Name: field.Name, Value: zero, Absent: true, Uninitialized: true})
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
