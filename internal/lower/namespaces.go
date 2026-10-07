package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

func namespaceStatements(declaration *ast.Node) []*ast.Node {
	body := declaration.AsModuleDeclaration().Body
	if body == nil {
		return nil
	}
	if body.Kind == ast.KindModuleDeclaration {
		return []*ast.Node{body}
	}
	return body.AsModuleBlock().Statements.Nodes
}

func namespaceRuntime(declaration *ast.Node) bool {
	for _, member := range namespaceStatements(declaration) {
		switch member.Kind {
		case ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration, ast.KindEmptyStatement:
		case ast.KindModuleDeclaration:
			if namespaceRuntime(member) {
				return true
			}
		default:
			return true
		}
	}
	return false
}

// Flatten scoped declarations, never names: the checker's symbols retain every namespace and
// private binding's identity. The qualified-only subset cannot observe a namespace object.
func namespaceDeclarations(statements []*ast.Node) []*ast.Node {
	result := []*ast.Node{}
	for _, statement := range statements {
		if statement.Kind == ast.KindModuleDeclaration {
			result = append(result, namespaceDeclarations(namespaceStatements(statement))...)
		} else {
			result = append(result, statement)
		}
	}
	return result
}

func (l *lowering) namespaceDeclaration(node *ast.Node) *ast.Node {
	symbol := l.symbol(node)
	if symbol == nil {
		return nil
	}
	for _, declaration := range symbol.Declarations {
		if declaration.Kind == ast.KindModuleDeclaration && ast.IsIdentifier(declaration.Name()) {
			return declaration
		}
	}
	return nil
}

func (l *lowering) namespaceMember(node *ast.Node) bool {
	module := false
	if node.Kind == ast.KindPropertyAccessExpression || node.Kind == ast.KindElementAccessExpression {
		module = l.moduleNamespace(node.Expression())
		if !module && l.namespaceDeclaration(node.Expression()) == nil {
			return false
		}
	}
	symbol := l.symbol(node)
	if symbol == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		parent := declaration.Parent
		if declaration.Kind == ast.KindVariableDeclaration {
			parent = parent.Parent.Parent
		}
		if parent != nil && (parent.Kind == ast.KindModuleBlock || module && parent.Kind == ast.KindSourceFile) {
			return true
		}
	}
	return false
}

func (l *lowering) namespaceExpression(node *ast.Node) (ir.Expression, bool, error) {
	if node.Kind != ast.KindPropertyAccessExpression && node.Kind != ast.KindElementAccessExpression || !l.namespaceMember(node) {
		return nil, false, nil
	}
	if node.Kind == ast.KindElementAccessExpression {
		return nil, true, l.notYet(node, "a computed namespace member; use a qualified name")
	}
	if node.AsPropertyAccessExpression().QuestionDotToken != nil || node.Flags&ast.NodeFlagsOptionalChain != 0 {
		return nil, true, l.notYet(node, "an optional namespace read; use a qualified name after initialization")
	}
	if function, found := l.functions[l.symbol(node)]; found {
		value, err := l.functionValue(node, function)
		return value, true, err
	}
	if local, found := l.local(node); found {
		read := ir.Expression(ir.Read{Local: local, Of: l.result.Locals[local].Type, Checked: l.checkedModuleRead(node, local)})
		if read.Type() == ir.Union {
			parent := node.Parent
			for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
				parent = parent.Parent
			}
			observing := comparedWithUndefined(node) || parent != nil && parent.Kind == ast.KindTypeOfExpression
			if narrowed, known := l.representation(l.checker.GetTypeAtLocation(node)); known && narrowed != ir.Union && !observing {
				return nil, true, l.notYet(node, "a narrowed namespace union member; read into a local before narrowing")
			}
		}
		if declared := read.Type(); declared.IsMaybe() {
			if narrowed, _ := l.representation(l.checker.GetTypeAtLocation(node)); narrowed == declared.Present() && !l.acceptsUndefined(node) {
				read = ir.Unwrap{Value: read}
			}
		}
		return l.defined(node, read), true, nil
	}
	return nil, true, l.notYet(node, "a namespace member without a lowered binding")
}

func (l *lowering) namespaceValueNode(node *ast.Node) bool {
	if parent := node.Parent; parent != nil && parent.Kind == ast.KindPropertyAccessExpression && parent.Name() == node {
		return false
	}
	if !l.isExpression(node) || ast.IsDeclarationName(node) {
		return false
	}
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		if ast.IsTypeNode(parent) || parent.Kind == ast.KindImportDeclaration || parent.Kind == ast.KindExportDeclaration {
			return false
		}
		if ast.IsStatement(parent) {
			break
		}
	}
	return true
}

func (l *lowering) namespaceRefusal(node *ast.Node) error {
	if err := l.moduleNamespaceRefusal(node); err != nil {
		return err
	}
	if node.Kind == ast.KindModuleDeclaration {
		if !ast.IsIdentifier(node.Name()) || ast.HasSyntacticModifier(node, ast.ModifierFlagsAmbient) {
			return l.notYet(node, "an ambient namespace or external module; use named file imports")
		}
		if node.Parent.Kind != ast.KindSourceFile && node.Parent.Kind != ast.KindModuleBlock && node.Parent.Kind != ast.KindModuleDeclaration {
			return l.notYet(node, "a namespace outside module scope")
		}
		symbol := l.symbol(node.Name())
		for _, declaration := range symbol.Declarations {
			if declaration == node || declaration.Kind == ast.KindInterfaceDeclaration || declaration.Kind == ast.KindTypeAliasDeclaration {
				continue
			}
			if declaration.Kind == ast.KindFunctionDeclaration && declaration.Parent == node.Parent {
				if namespaceOwnThis(declaration) {
					return &Refused{Where: l.program.Where(declaration), What: "this in a callable namespace function; qualified and detached calls have different receivers", Fix: "pass the state explicitly"}
				}
				for _, parameter := range declaration.Parameters() {
					if parameter.Name().Text() == "this" {
						return l.notYet(parameter, "an explicit callable-namespace this parameter; pass state explicitly")
					}
				}
				continue
			}
			return l.notYet(node, "a reopened namespace or namespace merged with a runtime value; put the declarations in one namespace or use a module")
		}
		for _, member := range namespaceStatements(node) {
			if l.namespaceMergedFunction(node.Name()) && ast.HasSyntacticModifier(member, ast.ModifierFlagsExport) {
				if member.Kind == ast.KindVariableStatement {
					for _, variable := range member.AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes {
						if variable.Name() != nil && callableNamespaceIntrinsic(variable.Name().Text()) {
							return l.notYet(variable, "a callable namespace export colliding with a function intrinsic")
						}
					}
				} else if member.Name() != nil && callableNamespaceIntrinsic(member.Name().Text()) {
					return l.notYet(member, "a callable namespace export colliding with a function intrinsic")
				}
			}
			switch member.Kind {
			case ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration, ast.KindEmptyStatement, ast.KindModuleDeclaration, ast.KindClassDeclaration, ast.KindEnumDeclaration:
			case ast.KindExpressionStatement, ast.KindBlock, ast.KindIfStatement, ast.KindForStatement, ast.KindWhileStatement, ast.KindDoStatement, ast.KindForOfStatement, ast.KindSwitchStatement:
				// Statement lowering keeps module evaluation order. Calls still pass preflight.
			case ast.KindFunctionDeclaration:
				if namespaceOwnThis(member) {
					return &Refused{Where: l.program.Where(member), What: "this in a namespace function; a qualified call and a detached call have different receivers", Fix: "pass the state as an explicit parameter instead of using this"}
				}
				for _, parameter := range member.Parameters() {
					if parameter.Name().Kind == ast.KindThisKeyword || parameter.Name().Text() == "this" {
						return l.notYet(parameter, "an explicit namespace-function this parameter; pass state explicitly")
					}
				}
			case ast.KindVariableStatement:
				list := member.AsVariableStatement().DeclarationList
				for _, variable := range list.AsVariableDeclarationList().Declarations.Nodes {
					if !ast.IsIdentifier(variable.Name()) && (list.Flags&ast.NodeFlagsBlockScoped != 0 || !namespaceObjectBindings(variable.Name())) {
						return l.notYet(variable, "destructuring inside a namespace; use plain singleton bindings")
					}
				}
			default:
				return l.notYet(member, "an unsupported namespace statement or declaration")
			}
		}
	}
	if node.Kind == ast.KindVariableDeclarationList && node.Flags&ast.NodeFlagsBlockScoped == 0 && !namespaceVariable(node) {
		for parent := node.Parent; parent != nil; parent = parent.Parent {
			if ast.IsFunctionLike(parent) {
				break
			}
			if parent.Kind == ast.KindModuleBlock {
				return l.notYet(node, "var inside namespace control flow; direct singleton var is hoisted, block var is not yet modeled")
			}
		}
	}
	if l.namespaceValueNode(node) && l.namespaceDeclaration(node) != nil {
		parent := node.Parent
		if l.namespaceMergedFunction(node) {
			if l.callableNamespaceUse(node) {
				return nil
			}
			return l.notYet(node, "observing a callable namespace object; only direct calls, fixed qualified members, typeof and canonical function identity are represented")
		}
		if parent == nil || parent.Kind != ast.KindPropertyAccessExpression || parent.AsPropertyAccessExpression().Expression != node {
			return l.notYet(node, "a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports")
		}
	}
	if l.isExpression(node) && l.namespaceMember(node) {
		parent := node.Parent
		if parent != nil && parent.Kind == ast.KindBinaryExpression && parent.AsBinaryExpression().Left == node && ast.IsAssignmentOperator(parent.AsBinaryExpression().OperatorToken.Kind) {
			if node.Kind == ast.KindPropertyAccessExpression || node.Kind == ast.KindElementAccessExpression {
				symbol := l.symbol(node)
				mutable := false
				for _, declaration := range symbol.Declarations {
					if declaration.Kind == ast.KindVariableDeclaration && declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
						mutable = true
					}
				}
				if !mutable {
					return l.notYet(node, "replacing a namespace export; exported functions and constants have fixed identity")
				}
			}
		}
	}
	return nil
}

// Namespace vars are hoisted and their export assignments are staged in JavaScript. Until we
// model those partial objects, arbitrary calls wait until every runtime namespace is initialized.
// Bodies of functions are deferred; private bindings keep their ordinary checked lexical storage.
func (l *lowering) namespaceInitialization(modules []*ast.SourceFile) error {
	if err := l.enumInitialization(modules); err != nil {
		return err
	}
	pending := 0
	for _, module := range modules {
		var count ast.Visitor
		count = func(node *ast.Node) bool {
			if node.Kind == ast.KindModuleDeclaration && namespaceRuntime(node) {
				pending++
			}
			return node.ForEachChild(count)
		}
		module.AsNode().ForEachChild(count)
	}
	if pending == 0 {
		return nil
	}
	initialized := map[*ast.Node]bool{}
	var visit func(*ast.Node) error
	visit = func(node *ast.Node) error {
		if node == nil || ast.IsTypeNode(node) || ast.IsFunctionLike(node) {
			return nil
		}
		if node.Kind == ast.KindModuleDeclaration {
			for _, statement := range namespaceStatements(node) {
				if err := visit(statement); err != nil {
					return err
				}
			}
			initialized[node] = true
			if namespaceRuntime(node) {
				pending--
			}
			return nil
		}
		if l.namespaceValueNode(node) {
			if declaration := l.namespaceDeclaration(node); declaration != nil && namespaceRuntime(declaration) && !initialized[declaration] {
				return l.notYet(node, "a namespace read before runtime initialization; put namespaces before executable module code")
			}
		}
		if pending > 0 && (node.Kind == ast.KindCallExpression && !l.isConsole(node.Expression()) || node.Kind == ast.KindNewExpression) {
			return l.notYet(node, "a call before all runtime namespaces are initialized; put namespaces before executable module code")
		}
		var found error
		node.ForEachChild(func(child *ast.Node) bool { found = visit(child); return found != nil })
		return found
	}
	for _, module := range modules {
		for _, statement := range module.Statements.Nodes {
			if err := visit(statement); err != nil {
				return err
			}
		}
	}
	return nil
}

// namespaceVariable identifies direct singleton storage, excluding function-local var.
func namespaceVariable(list *ast.Node) bool {
	return list.Parent != nil && list.Parent.Parent != nil && list.Parent.Parent.Kind == ast.KindModuleBlock
}

func (l *lowering) namespaceBody(node *ast.Node) ([]ir.Statement, error) {
	hoisted := []ir.Statement{}
	for _, statement := range namespaceStatements(node) {
		if statement.Kind != ast.KindVariableStatement {
			continue
		}
		list := statement.AsVariableStatement().DeclarationList
		if list.Flags&ast.NodeFlagsBlockScoped != 0 {
			continue
		}
		for _, declaration := range list.AsVariableDeclarationList().Declarations.Nodes {
			names := []*ast.Node{declaration.Name()}
			if !ast.IsIdentifier(declaration.Name()) {
				if !namespaceObjectBindings(declaration.Name()) {
					return nil, l.notYet(declaration.Name(), "destructuring inside a namespace; only flat object var bindings are hoisted")
				}
				names = nil
				for _, binding := range declaration.Name().AsBindingPattern().Elements.Nodes {
					names = append(names, binding.Name())
				}
			}
			for _, name := range names {
				local, err := l.declareLocal(name)
				if err != nil {
					return nil, err
				}
				var value ir.Expression
				if l.includesUndefined(l.checker.GetTypeAtLocation(name)) {
					value = fit(ir.Undefined{Of: l.result.Locals[local].Type}, l.result.Locals[local].Type)
				}
				hoisted = append(hoisted, ir.Declare{Local: local, Value: value})
			}
		}
	}
	body, err := l.statements(namespaceStatements(node))
	return append(hoisted, body...), err
}

// The parser's factory destructuring declares symbols on the renamed identifiers,
// not on the ObjectBindingPattern. Other binding forms retain their own NotYet.
func namespaceObjectBindings(pattern *ast.Node) bool {
	if pattern.Kind != ast.KindObjectBindingPattern {
		return false
	}
	for _, binding := range pattern.AsBindingPattern().Elements.Nodes {
		declared := binding.AsBindingElement()
		if !ast.IsIdentifier(binding.Name()) || declared.Initializer != nil || declared.DotDotDotToken != nil {
			return false
		}
		if declared.PropertyName != nil && !ast.IsIdentifier(declared.PropertyName) && declared.PropertyName.Kind != ast.KindStringLiteral {
			return false
		}
	}
	return true
}
