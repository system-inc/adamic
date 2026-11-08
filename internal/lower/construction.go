package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

// A fresh plain literal can reserve its factory's fixed scalar slots. Its asserted
// result remains a checked view; reserving storage proves no value initialized.
func (lowering *lowering) constructionCast(node *ast.Node) bool {
	source := lowering.factoryConstructionSource(node)
	if source == nil {
		return false
	}
	target := lowering.concrete(lowering.checker.GetTypeAtLocation(node))
	if target.Flags()&checker.TypeFlagsObject == 0 ||
		isClassInstance(target) ||
		len(lowering.checker.GetSignaturesOfType(target, checker.SignatureKindCall)) != 0 ||
		len(lowering.checker.GetSignaturesOfType(target, checker.SignatureKindConstruct)) != 0 {
		return false
	}
	fields := lowering.checker.GetPropertiesOfType(target)
	if len(fields) == 0 {
		return false
	}
	for _, field := range fields {
		if strings.ContainsRune(field.Name, 0) || strings.HasPrefix(field.Name, "#") || field.Name == "__proto__" {
			return false
		}
		held, known := lowering.representation(lowering.checker.GetTypeOfSymbol(field))
		if field.Flags&ast.SymbolFlagsOptional != 0 ||
			!known ||
			held < ir.Number ||
			held > ir.String ||
			!interfaceScalar(lowering.checker.GetTypeOfSymbol(field)) {
			return false
		}
	}
	provided := map[string]bool{}
	for _, property := range source.AsObjectLiteralExpression().Properties.Nodes {
		if property.Kind != ast.KindPropertyAssignment && property.Kind != ast.KindShorthandPropertyAssignment {
			return false
		}
		name, known := lowering.methodName(property)
		if !known || name == "__proto__" {
			return false
		}
		provided[name] = true
		field := lowering.checker.GetPropertyOfType(target, name)
		if field == nil {
			return false
		}
		value := property.Name()
		if property.Kind == ast.KindPropertyAssignment {
			value = property.AsPropertyAssignment().Initializer
		}
		actual := lowering.checker.GetTypeAtLocation(value)
		wanted := lowering.checker.GetTypeOfSymbol(field)
		if !interfaceScalar(actual) || !lowering.checker.IsTypeAssignableTo(actual, wanted) {
			return false
		}
	}
	return len(provided) < len(fields)
}

func (lowering *lowering) constructLiteral(node *ast.Node) (ir.Expression, error) {
	target := lowering.concrete(lowering.checker.GetTypeAtLocation(node))
	source := ast.SkipParentheses(node.AsAsExpression().Expression)
	if source.Kind == ast.KindCallExpression {
		return lowering.expression(source)
	}
	value, objectLiteralError := lowering.objectLiteral(source)
	if objectLiteralError != nil {
		return nil, objectLiteralError
	}
	literal := value.(ir.ObjectLiteral)
	present := map[string]bool{}
	for _, field := range literal.Fields {
		present[field.Name] = true
	}
	for _, field := range lowering.checker.GetPropertiesOfType(target) {
		if present[field.Name] {
			continue
		}
		held, _ := lowering.representation(lowering.checker.GetTypeOfSymbol(field))
		zero := ir.Expression(zeroValue(held))
		if held.IsReference() {
			zero = ir.Undefined{Of: held}
		}
		literal.Fields = append(
			literal.Fields,
			ir.Field{Name: field.Name, Value: zero, Absent: true, Uninitialized: true},
		)
	}
	return literal, nil
}

// A literal string key names the same fixed slot as a dot access. It is not a
// dynamic index-signature write and cannot add a slot the declared layout lacks.
func (lowering *lowering) constructionIndexWrite(target, valueNode *ast.Node) ([]ir.Statement, bool, error) {
	access := target.AsElementAccessExpression()
	key := ast.SkipParentheses(access.ArgumentExpression)
	if key.Kind != ast.KindStringLiteral ||
		strings.ContainsRune(key.Text(), 0) ||
		strings.HasPrefix(key.Text(), "#") ||
		key.Text() == "__proto__" {
		return nil, false, nil
	}
	receiver := lowering.checker.GetTypeAtLocation(access.Expression)
	held, _ := lowering.representation(receiver)
	if held != ir.Object || isClassInstance(receiver) {
		return nil, false, nil
	}
	field := lowering.checker.GetPropertyOfType(receiver, key.Text())
	if field == nil || !interfaceScalar(lowering.checker.GetTypeOfSymbol(field)) {
		return nil, false, nil
	}
	object, expressionError := lowering.expression(access.Expression)
	if expressionError != nil {
		return nil, true, expressionError
	}
	value, expressionError := lowering.expression(valueNode)
	if expressionError != nil {
		return nil, true, expressionError
	}
	declared, known := lowering.representation(lowering.checker.GetTypeOfSymbol(field))
	if !known || value.Type() != declared {
		return nil, true, lowering.notYet(target, "a string-key field write requiring representation conversion")
	}
	return []ir.Statement{
		ir.SetProperty{Object: object, Name: key.Text(), Value: value, Site: lowering.writeSite(access.Expression)},
	}, true, nil
}

// A const initialized by one factory return has a closed physical layout. This
// proves JSON's descriptor, not the values of reserved or unset fields.
func (lowering *lowering) constructionJson(node *ast.Node) (*ir.JSONSchema, bool) {
	if !ast.IsIdentifier(node) {
		return nil, false
	}
	symbol := lowering.symbol(node)
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
		callee := lowering.symbol(ast.SkipParentheses(initializer.AsCallExpression().Expression))
		if callee == nil ||
			len(callee.Declarations) != 1 ||
			callee.Declarations[0].Kind != ast.KindFunctionDeclaration {
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
	if assertion == nil || assertion.Kind != ast.KindAsExpression || !lowering.constructionCast(assertion) {
		return nil, false
	}
	names := []string{}
	present := map[string]bool{}
	source := lowering.factoryConstructionSource(assertion)
	for _, property := range source.AsObjectLiteralExpression().Properties.Nodes {
		name, _ := lowering.methodName(property)
		names = append(names, name)
		present[name] = true
	}
	target := lowering.concrete(lowering.checker.GetTypeAtLocation(assertion))
	for _, field := range lowering.checker.GetPropertiesOfType(target) {
		if !present[field.Name] {
			names = append(names, field.Name)
		}
	}
	schema := &ir.JSONSchema{Kind: "object"}
	for index, name := range names {
		field := lowering.checker.GetPropertyOfType(target, name)
		held, _ := lowering.representation(lowering.checker.GetTypeOfSymbol(field))
		kind := map[ir.Type]string{ir.Number: "number", ir.Boolean: "boolean", ir.String: "string"}[held]
		schema.Fields = append(
			schema.Fields,
			ir.JSONField{Name: lowering.constant(name), Slot: index, Schema: &ir.JSONSchema{Kind: kind}},
		)
	}
	return schema, true
}

// Only a single direct literal return fixes the allocation provenance. Branches,
// indirect calls, aliases, mutations and generic helper bodies remain checked views.
func (lowering *lowering) factoryConstructionSource(node *ast.Node) *ast.Node {
	source := ast.SkipParentheses(node.AsAsExpression().Expression)
	if source.Kind == ast.KindObjectLiteralExpression {
		return source
	}
	if source.Kind != ast.KindCallExpression {
		return nil
	}
	callee := lowering.symbol(ast.SkipParentheses(source.AsCallExpression().Expression))
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
func (lowering *lowering) reserveFactoryFields(node *ast.Node, literal ir.ObjectLiteral) (ir.ObjectLiteral, error) {
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
		if n.Kind == ast.KindAsExpression &&
			lowering.factoryConstructionSource(n) == node &&
			lowering.constructionCast(n) {
			for _, field := range lowering.checker.GetPropertiesOfType(
				lowering.concrete(lowering.checker.GetTypeAtLocation(n)),
			) {
				held, _ := lowering.representation(lowering.checker.GetTypeOfSymbol(field))
				if old, found := present[field.Name]; found {
					if old != held {
						failure = lowering.notYet(n, "factory target layouts with incompatible field storage")
					}
					continue
				}
				zero := ir.Expression(zeroValue(held))
				if held.IsReference() {
					zero = ir.Undefined{Of: held}
				}
				literal.Fields = append(
					literal.Fields,
					ir.Field{Name: field.Name, Value: zero, Absent: true, Uninitialized: true},
				)
				present[field.Name] = held
			}
		}
		return n.ForEachChild(visit)
	}
	for _, file := range lowering.program.Files() {
		file.AsNode().ForEachChild(visit)
	}
	return literal, failure
}
