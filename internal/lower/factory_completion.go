package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// A fresh staged allocation reserves the complete declared shape. Its absent
// slots are unreadable until assigned, rather than fictional initialized zeros.
// This first surface is a local factory initializer, not an assertion over an
// existing object or a generic allocator whose construction is still unknown.
func (l *lowering) factoryStage(node *ast.Node, target *checker.Type) (bool, error) {
	expression := ast.SkipParentheses(node.AsAsExpression().Expression)
	if expression.Kind != ast.KindObjectLiteralExpression || target.Flags()&checker.TypeFlagsObject == 0 || isClassInstance(target) {
		return false, nil
	}
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	if parent == nil || parent.Kind != ast.KindVariableDeclaration || !ast.IsIdentifier(parent.Name()) {
		return false, nil
	}
	factory := parent.Parent
	for factory != nil && !ast.IsFunctionLike(factory) {
		factory = factory.Parent
	}
	if factory == nil {
		return false, nil
	}
	if len(l.checker.GetSignaturesOfType(target, checker.SignatureKindCall)) != 0 || l.checker.IsArrayType(target) {
		return false, nil
	}
	fields, fixed := interfaceLiteralFields(expression)
	if !fixed {
		return false, nil
	}
	missing := false
	for _, property := range l.checker.GetPropertiesOfType(target) {
		missing = missing || fields[property.Name] == nil
	}
	if !missing {
		return false, nil
	}
	for _, property := range l.checker.GetPropertiesOfType(target) {
		declared := l.concrete(l.checker.GetTypeOfSymbol(property))
		of, known := l.representation(declared)
		if property.Flags&ast.SymbolFlagsOptional != 0 {
			return false, l.notYet(node, "an optional staged factory field "+property.Name+"; requires step 17 presence representation (#z00sxvc)")
		}
		if !known || slotless(of) || l.callableViewContract(declared) || accessorSymbol(property) {
			return false, l.notYet(node, "a staged factory field "+property.Name+" without a checked storage representation")
		}
		if initial := fields[property.Name]; initial != nil && !l.censusRelated(l.checker.GetTypeAtLocation(initial), declared) {
			return false, &Refused{Where: l.program.Where(initial), What: "a staged factory initializer incompatible with field " + property.Name, Fix: "initialize the field with a value admitted by its declared type"}
		}
	}
	// Reserved absent fields do not yet have step 17's observable presence bits.
	// Refuse observers rather than turn absence into a present undefined key.
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return false, err
	}
	var found error
	var visit ast.Visitor
	visit = func(part *ast.Node) bool {
		computedObserver := false
		if part.Kind == ast.KindElementAccessExpression {
			key := ast.SkipParentheses(part.AsElementAccessExpression().ArgumentExpression)
			computedObserver = key.Kind != ast.KindStringLiteral || key.Text() == "hasOwnProperty" || key.Text() == "propertyIsEnumerable"
		}
		if computedObserver || part.Kind == ast.KindForInStatement || part.Kind == ast.KindSpreadAssignment || part.Kind == ast.KindInKeyword || (part.Kind == ast.KindPropertyAccessExpression && (part.Name().Text() == "hasOwnProperty" || part.Name().Text() == "propertyIsEnumerable")) || part.Kind == ast.KindIdentifier && (part.Text() == "Object" || part.Text() == "Reflect" || part.Text() == "JSON") {
			found = l.notYet(part, "observing staged factory field presence before step 17 (#z00sxvc)")
			return true
		}
		if found == nil {
			part.ForEachChild(visit)
		}
		return found != nil
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
		if found != nil {
			return false, found
		}
	}
	return true, nil
}

func (l *lowering) factoryAllocation(node *ast.Node, value ir.Expression, target *checker.Type) (ir.Expression, error) {
	literal, fresh := value.(ir.ObjectLiteral)
	if !fresh || literal.Spread != nil {
		return nil, l.notYet(node, "a staged factory allocation without a fresh fixed literal")
	}
	made := map[string]bool{}
	for _, field := range literal.Fields {
		made[field.Name] = true
	}
	for _, property := range l.checker.GetPropertiesOfType(target) {
		if made[property.Name] {
			continue
		}
		of, known := l.representation(l.concrete(l.checker.GetTypeOfSymbol(property)))
		if !known {
			return nil, l.notYet(node, "a staged factory field without a representation")
		}
		empty := ir.Expression(zeroValue(of))
		if of.IsReference() {
			empty = ir.Undefined{Of: of}
		}
		literal.Fields = append(literal.Fields, ir.Field{Name: property.Name, Value: empty, Uninitialized: true})
	}
	return literal, nil
}
