package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// noteNamespaceObjects selects live containers before signatures and function bodies are lowered.
func (l *lowering) noteNamespaceObjects(modules []*ast.SourceFile) {
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if l.namespaceValueNode(node) {
			if d := l.namespaceDeclaration(node); d != nil && !l.namespaceMergedFunction(node) {
				parent := node.Parent
				if parent == nil || parent.Kind != ast.KindPropertyAccessExpression || parent.Expression() != node {
					l.namespaceObjectLocal(d)
				}
			}
		}
		if node.Kind == ast.KindModuleDeclaration && ast.IsIdentifier(node.Name()) {
			count := 0
			for _, d := range l.symbol(node.Name()).Declarations {
				if d.Kind == ast.KindModuleDeclaration {
					count++
				}
			}
			if count > 1 {
				l.namespaceObjectLocal(node)
			}
		}
		return node.ForEachChild(visit)
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
}

func (l *lowering) namespaceObjectDeclaration(d *ast.Node) bool {
	if d == nil || d.Kind != ast.KindModuleDeclaration {
		return false
	}
	local, ok := l.locals[l.symbol(d.Name())]
	return ok && l.result.Locals[local].NamespaceObject
}

func (l *lowering) namespaceObjectLocal(d *ast.Node) int {
	if l.locals == nil {
		l.locals = map[*ast.Symbol]int{}
	}
	symbol := l.symbol(d.Name())
	if local, ok := l.locals[symbol]; ok && l.result.Locals[local].NamespaceObject {
		return local
	}
	local := len(l.result.Locals)
	ready := local + 1
	l.locals[symbol] = local
	l.result.Locals = append(l.result.Locals, ir.Local{Name: d.Name().Text(), Type: ir.Object, Global: true, Function: -1, NamespaceObject: true, NamespaceReady: ready + 1}, ir.Local{Name: d.Name().Text() + "_created", Type: ir.Boolean, Global: true, Function: -1})
	l.forwarderValues = append(l.forwarderValues, ir.Declare{Local: local, Value: ir.Undefined{Of: ir.Object}}, ir.Declare{Local: ready, Value: ir.BooleanConstant{Value: false}})
	return local
}

func (l *lowering) namespaceExport(node *ast.Node) (*ast.Node, string) {
	symbol := l.symbol(node)
	if symbol == nil {
		return nil, ""
	}
	for _, d := range symbol.Declarations {
		statement := d
		if d.Kind == ast.KindVariableDeclaration {
			statement = d.Parent.Parent
		}
		if statement.Parent != nil && statement.Parent.Kind == ast.KindModuleBlock && ast.HasSyntacticModifier(statement, ast.ModifierFlagsExport) {
			owner := statement.Parent.Parent
			if l.namespaceObjectDeclaration(owner) {
				return owner, symbol.Name
			}
		}
	}
	return nil, ""
}

func (l *lowering) namespaceExportRead(node *ast.Node) (ir.Expression, bool, error) {
	owner, name := l.namespaceExport(node)
	if owner == nil {
		return nil, false, nil
	}
	of, err := l.typeOf(node)
	if err != nil {
		return nil, true, err
	}
	local := l.namespaceObjectLocal(owner)
	object := ir.Expression(ir.Read{Local: local, Of: ir.Object})
	if node.Kind == ast.KindPropertyAccessExpression || node.Kind == ast.KindElementAccessExpression {
		object, err = l.expression(node.Expression())
		if err != nil {
			return nil, true, err
		}
	}
	if comparedWithUndefined(node) || node.Parent != nil && (node.Parent.Kind == ast.KindTypeOfExpression || node.Parent.Kind == ast.KindCallExpression && l.isLibraryGlobal(node.Parent.Expression(), "String") || node.Parent.Kind == ast.KindTemplateSpan) {
		of = ir.Union
	}
	return ir.Property{Object: object, Name: name, Of: of, Namespace: true, View: sourceExpression(node)}, true, nil
}

func (l *lowering) namespaceObjectExpression(node *ast.Node) (ir.Expression, bool, error) {
	if d := l.namespaceDeclaration(node); l.namespaceObjectDeclaration(d) {
		return ir.Read{Local: l.namespaceObjectLocal(d), Of: ir.Object}, true, nil
	}
	if node.Kind == ast.KindPropertyAccessExpression || node.Kind == ast.KindElementAccessExpression {
		return l.namespaceExportRead(node)
	}
	return nil, false, nil
}

func (l *lowering) namespaceObjectVariable(d *ast.Node) ([]ir.Statement, bool, error) {
	owner, name := l.namespaceExport(d.Name())
	if owner == nil {
		return nil, false, nil
	}
	if l.includesNull(l.checker.GetTypeOfSymbol(l.symbol(d.Name()))) {
		return nil, true, l.notYet(d, "a nullable namespace export; its staged absence and null require distinct tags")
	}
	if d.Initializer() == nil {
		return nil, true, nil
	} // tsc emits no assignment for an uninitialized export.
	value, err := l.expression(d.Initializer())
	if err != nil {
		return nil, true, err
	}
	readonly := d.Parent.Flags&ast.NodeFlagsConst != 0
	return []ir.Statement{l.namespaceInstall(owner, name, value, readonly)}, true, nil
}

func (l *lowering) namespaceInstall(owner *ast.Node, name string, value ir.Expression, readonly bool) ir.SetProperty {
	return ir.SetProperty{Object: ir.Read{Local: l.namespaceObjectLocal(owner), Of: ir.Object}, Name: name, Value: fit(value, ir.Union), Record: true, NamespaceInstall: true, NamespaceReadonly: readonly}
}

func (l *lowering) namespaceObjectBody(node *ast.Node) ([]ir.Statement, error) {
	local := l.namespaceObjectLocal(node)
	hoisted, err := l.namespaceHoisted(node)
	if err != nil {
		return nil, err
	}
	body := []ir.Statement{ir.If{Condition: ir.IsUndefined{Value: ir.Read{Local: local, Of: ir.Object}}, Then: []ir.Statement{ir.Assign{Local: local, Value: ir.ObjectLiteral{Record: true, Namespace: true}}, ir.Assign{Local: l.namespaceReadyLocal(node), Value: ir.BooleanConstant{Value: true}}}}}
	body = append(body, hoisted...)
	for _, statement := range namespaceStatements(node) {
		if ast.HasSyntacticModifier(statement, ast.ModifierFlagsExport) {
			if statement.Name() != nil && statement.Name().Text() == "__proto__" {
				return nil, l.notYet(statement, "an escaped namespace export named __proto__; its assignment invokes a prototype setter")
			}
			if statement.Kind == ast.KindVariableStatement {
				for _, variable := range statement.AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes {
					if variable.Name() != nil && variable.Name().Text() == "__proto__" {
						return nil, l.notYet(variable, "an escaped namespace export named __proto__; its assignment invokes a prototype setter")
					}
				}
			}
		}
		lowered, err := l.statement(statement)
		if err != nil {
			return nil, err
		}
		body = append(body, lowered...)
		if !ast.HasSyntacticModifier(statement, ast.ModifierFlagsExport) {
			continue
		}
		switch statement.Kind {
		case ast.KindFunctionDeclaration:
			function, ok := l.functions[l.symbol(statement.Name())]
			if !ok {
				return nil, l.notYet(statement, "a generic function installed on an escaped namespace")
			}
			value, err := l.functionValue(statement.Name(), function)
			if err != nil {
				return nil, err
			}
			body = append(body, l.namespaceInstall(node, statement.Name().Text(), value, true))
		case ast.KindEnumDeclaration:
			if !ast.HasSyntacticModifier(statement, ast.ModifierFlagsConst) {
				return nil, l.notYet(statement, "a runtime enum installed on an escaped namespace")
			}
		case ast.KindClassDeclaration, ast.KindModuleDeclaration:
			return nil, l.notYet(statement, "a class or nested runtime namespace installed on an escaped namespace")
		}
	}
	return body, nil
}

func (l *lowering) namespaceObjectAssignment(node, target *ast.Node, operator ast.Kind, compound bool) ([]ir.Statement, bool, error) {
	owner, name := l.namespaceExport(target)
	if owner == nil {
		return nil, false, nil
	}
	object, prefix, err := l.namespaceSavedReceiver(target, owner)
	if err != nil {
		return nil, true, err
	}
	value, err := l.expression(node.AsBinaryExpression().Right)
	if err != nil {
		return nil, true, err
	}
	if compound {
		current, _, err := l.namespaceExportRead(target)
		if err != nil {
			return nil, true, err
		}
		property := current.(ir.Property)
		property.Object = object
		value, err = l.combine(node, operator, property, value)
		if err != nil {
			return nil, true, err
		}
	}
	return append(prefix, ir.SetProperty{Object: object, Name: name, Value: fit(value, ir.Union), Record: true}), true, nil
}

func (l *lowering) namespaceObjectIncrement(node, operand *ast.Node, operator ast.Kind) ([]ir.Statement, bool, error) {
	owner, name := l.namespaceExport(operand)
	if owner == nil {
		return nil, false, nil
	}
	object, prefix, err := l.namespaceSavedReceiver(operand, owner)
	if err != nil {
		return nil, true, err
	}
	current, _, err := l.namespaceExportRead(operand)
	if err != nil {
		return nil, true, err
	}
	property := current.(ir.Property)
	property.Object = object
	step := ir.Add
	if operator == ast.KindMinusMinusToken {
		step = ir.Subtract
	}
	value := ir.Binary{Operator: step, Left: property, Right: ir.NumberConstant{Value: 1}}
	return append(prefix, ir.SetProperty{Object: object, Name: name, Value: fit(value, ir.Union), Record: true}), true, nil
}

func (l *lowering) namespaceSavedReceiver(node, owner *ast.Node) (ir.Expression, []ir.Statement, error) {
	object, err := l.namespaceExportObject(node, owner)
	if err != nil {
		return nil, nil, err
	}
	if ast.IsIdentifier(node) {
		return object, nil, nil
	}
	local := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "namespace_receiver", Type: ir.Object, Function: l.functionIndex})
	return ir.Read{Local: local, Of: ir.Object}, []ir.Statement{ir.Declare{Local: local, Value: object}}, nil
}

func (l *lowering) namespaceType(proven *checker.Type) bool {
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		for _, part := range proven.Types() {
			if l.namespaceType(part) {
				return true
			}
		}
		return false
	}
	symbol := proven.Symbol()
	if symbol == nil {
		return false
	}
	for _, d := range symbol.Declarations {
		if d.Kind == ast.KindModuleDeclaration && l.namespaceObjectDeclaration(d) {
			return true
		}
	}
	return false
}

func (l *lowering) hasNamespaceObjects() bool {
	for _, local := range l.result.Locals {
		if local.NamespaceObject {
			return true
		}
	}
	return false
}

func (l *lowering) namespaceExportObject(node, owner *ast.Node) (ir.Expression, error) {
	if node.Kind == ast.KindPropertyAccessExpression || node.Kind == ast.KindElementAccessExpression {
		return l.expression(node.Expression())
	}
	return ir.Read{Local: l.namespaceObjectLocal(owner), Of: ir.Object}, nil
}

// TypeScript erases blocks containing only types and const enums.
func namespaceObjectRuntime(node *ast.Node) bool {
	if !namespaceRuntime(node) {
		return false
	}
	for _, member := range namespaceStatements(node) {
		switch member.Kind {
		case ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration, ast.KindEmptyStatement:
		case ast.KindEnumDeclaration:
			if !ast.HasSyntacticModifier(member, ast.ModifierFlagsConst) {
				return true
			}
		case ast.KindModuleDeclaration:
			if namespaceObjectRuntime(member) {
				return true
			}
		default:
			return true
		}
	}
	return false
}
