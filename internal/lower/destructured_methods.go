package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/load"
)

// Unlike erasedMethodField's compatible-allocation scan, this proof starts at
// the pattern's source. An unrelated new C() can own the same method that an
// object literal only stores as c.method.bind(c).
type destructuredMethodProof struct {
	*lowering
	modules []*ast.SourceFile
	active  map[*ast.Node]bool
}

func (l *lowering) detachableDestructuredMethod(where *ast.Node, field string) bool {
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil || where.Parent == nil || where.Parent.Parent == nil {
		return false
	}
	proof := destructuredMethodProof{l, modules, make(map[*ast.Node]bool)}
	owner := where.Parent.Parent
	if proof.mayReplaceMember(field) && !proof.freshParameterSource(owner) {
		return false
	}
	switch owner.Kind {
	case ast.KindVariableDeclaration, ast.KindParameter:
		return proof.patternSource(owner, field)
	case ast.KindBinaryExpression:
		binary := owner.AsBinaryExpression()
		return binary.Left == where.Parent && binary.OperatorToken.Kind == ast.KindEqualsToken && proof.value(binary.Right, field)
	}
	return false
}

// A direct write through any view naming the field must be considered. An
// unbounded element write or a library bulk write keeps the original stop.
// We do not infer that an alias is unmodified from its declared type alone.
func (p *destructuredMethodProof) mayReplaceMember(name string) bool {
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		write := ast.IsAssignmentTarget(node) || node.Parent != nil && node.Parent.Kind == ast.KindDeleteExpression
		switch node.Kind {
		case ast.KindPropertyAccessExpression:
			if write && node.Name().Text() == name {
				return true
			}
		case ast.KindElementAccessExpression:
			if write {
				key := node.AsElementAccessExpression().ArgumentExpression
				if key == nil || key.Kind != ast.KindStringLiteral || key.Text() == name {
					return true
				}
			}
		case ast.KindCallExpression:
			call := node.AsCallExpression()
			callee := ast.SkipParentheses(call.Expression)
			if callee.Kind == ast.KindPropertyAccessExpression && p.libraryMember(callee) {
				object := callee.AsPropertyAccessExpression().Expression
				switch callee.Name().Text() {
				case "assign", "set", "defineProperty", "defineProperties", "setPrototypeOf":
					if p.isLibraryGlobal(object, "Object") || p.isLibraryGlobal(object, "Reflect") {
						return true
					}
				}
			}
		}
		return node.ForEachChild(visit)
	}
	for _, module := range p.modules {
		if module.AsNode().ForEachChild(visit) {
			return true
		}
	}
	return false
}

// A nonempty field selects a member of the value. An empty field asks whether
// the value itself is detachable. Cycles and unrecognized value roads fail closed.
func (p *destructuredMethodProof) value(node *ast.Node, field string) bool {
	if node == nil {
		return false
	}
	node = ast.SkipParentheses(node)
	if p.active[node] {
		return false
	}
	p.active[node] = true
	defer delete(p.active, node)

	switch node.Kind {
	case ast.KindArrowFunction:
		return field == ""
	case ast.KindCallExpression:
		callee := ast.SkipParentheses(node.AsCallExpression().Expression)
		if field == "" && callee.Kind == ast.KindPropertyAccessExpression && callee.Name().Text() == "bind" && p.libraryMember(callee) {
			return true // Ordinary lowering still judges the bind and its arguments.
		}
		fn := p.directFunction(callee)
		return fn != nil && p.results(fn.Body(), field)
	case ast.KindObjectLiteralExpression:
		return field != "" && p.literalMember(node, field)
	case ast.KindNewExpression:
		return field != "" && p.instanceMember(node, field)
	case ast.KindIdentifier:
		symbol := p.valueSymbol(node)
		if symbol == nil || len(symbol.Declarations) != 1 || p.assignedSymbol(symbol) {
			return false
		}
		declaration := symbol.Declarations[0]
		switch declaration.Kind {
		case ast.KindVariableDeclaration:
			return p.value(declaration.Initializer(), field)
		case ast.KindParameter:
			return p.patternSource(declaration, field)
		case ast.KindBindingElement:
			if field != "" || declaration.Initializer() != nil {
				return false
			}
			binding := declaration.AsBindingElement()
			if binding.DotDotDotToken != nil || declaration.Parent.Kind != ast.KindObjectBindingPattern {
				return false
			}
			name := binding.PropertyName
			if name == nil {
				name = declaration.Name()
			}
			if !ast.IsIdentifier(name) && name.Kind != ast.KindStringLiteral {
				return false
			}
			return p.patternSource(declaration.Parent.Parent, name.Text())
		}
	case ast.KindConditionalExpression:
		choice := node.AsConditionalExpression()
		return p.value(choice.WhenTrue, field) && p.value(choice.WhenFalse, field)
	case ast.KindAsExpression:
		return p.value(node.AsAsExpression().Expression, field)
	case ast.KindNonNullExpression:
		return p.value(node.AsNonNullExpression().Expression, field)
	}
	return false
}

func (p *destructuredMethodProof) literalMember(node *ast.Node, field string) bool {
	var provider *ast.Node
	for _, member := range node.AsObjectLiteralExpression().Properties.Nodes {
		name, named := memberKeyText(member)
		if !named {
			// A spread or a computed key can replace a preceding instance.
			return false
		}
		if name == field {
			if provider != nil {
				return false
			}
			provider = member
		}
	}
	return p.member(provider)
}

func (p *destructuredMethodProof) instanceMember(node *ast.Node, field string) bool {
	ctor := node.AsNewExpression().Expression
	if !ast.IsIdentifier(ctor) {
		return false
	}
	symbol := p.symbol(ctor)
	if symbol == nil || len(symbol.Declarations) != 1 || p.assignedSymbol(symbol) {
		return false
	}
	class := symbol.Declarations[0]
	if class.Kind != ast.KindClassDeclaration || ast.GetClassExtendsHeritageElement(class) != nil {
		return false
	}
	for _, member := range class.Members() {
		if member.Kind == ast.KindConstructor && hasConstructorReturn(member.Body()) {
			return false
		}
	}
	slots := p.checker.GetPropertyOfType(p.checker.GetTypeAtLocation(node), field)
	if slots == nil {
		return false
	}
	roots := p.checker.GetRootSymbols(slots)
	return len(roots) == 1 && len(roots[0].Declarations) == 1 && p.member(roots[0].Declarations[0])
}

func (p *destructuredMethodProof) member(node *ast.Node) bool {
	if node == nil || load.IsLibrary(ast.GetSourceFileOfNode(node)) {
		return false
	}
	switch node.Kind {
	case ast.KindPropertyAssignment, ast.KindPropertyDeclaration:
		return p.value(node.Initializer(), "")
	case ast.KindShorthandPropertyAssignment:
		return p.value(node.Name(), "")
	case ast.KindGetAccessor:
		return p.results(node.Body(), "")
	}
	return false
}

func (p *destructuredMethodProof) directFunction(callee *ast.Node) *ast.Node {
	if !ast.IsIdentifier(callee) {
		return nil
	}
	symbol := p.symbol(callee)
	if symbol == nil || len(symbol.Declarations) != 1 || p.assignedSymbol(symbol) {
		return nil
	}
	declaration := symbol.Declarations[0]
	if declaration.Kind != ast.KindFunctionDeclaration || declaration.Body() == nil {
		return nil
	}
	if generatorDeclaration(declaration) || ast.HasSyntacticModifier(declaration, ast.ModifierFlagsAsync) {
		return nil
	}
	return declaration
}

func (p *destructuredMethodProof) results(body *ast.Node, field string) bool {
	if body == nil {
		return false
	}
	if body.Kind != ast.KindBlock {
		return p.value(body, field)
	}
	statements := body.AsBlock().Statements.Nodes
	if len(statements) == 0 || statements[len(statements)-1].Kind != ast.KindReturnStatement {
		return false
	}
	var invalid ast.Visitor
	invalid = func(node *ast.Node) bool {
		if ast.IsFunctionLike(node) {
			return false
		}
		if node.Kind == ast.KindReturnStatement {
			return !p.value(node.AsReturnStatement().Expression, field)
		}
		return node.ForEachChild(invalid)
	}
	return !body.ForEachChild(invalid)
}

// An inline data literal given straight to an object parameter pattern has
// no alias before its reads. This is the only exemption from the global write
// check. Accessors and spreads could expose or replace the original slots.
func (p *destructuredMethodProof) freshParameterSource(owner *ast.Node) bool {
	if owner.Kind != ast.KindParameter || owner.Name().Kind != ast.KindObjectBindingPattern {
		return false
	}
	return p.parameterArguments(owner, func(argument *ast.Node) bool {
		argument = ast.SkipParentheses(argument)
		if argument.Kind != ast.KindObjectLiteralExpression {
			return false
		}
		for _, member := range argument.AsObjectLiteralExpression().Properties.Nodes {
			switch member.Kind {
			case ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment, ast.KindMethodDeclaration:
			default:
				return false
			}
		}
		return true
	})
}

func (p *destructuredMethodProof) patternSource(owner *ast.Node, field string) bool {
	if owner.Kind == ast.KindVariableDeclaration {
		return !p.mayReplaceMember(field) && p.value(owner.Initializer(), field)
	}
	if owner.Kind != ast.KindParameter {
		return false
	}
	// This also covers a field selected from a forwarded destructured binding.
	if p.mayReplaceMember(field) && !p.freshParameterSource(owner) {
		return false
	}
	return p.parameterArguments(owner, func(argument *ast.Node) bool {
		return p.value(argument, field)
	})
}

func (p *destructuredMethodProof) parameterArguments(owner *ast.Node, prove func(*ast.Node) bool) bool {
	if owner.Initializer() != nil || owner.AsParameterDeclaration().DotDotDotToken != nil {
		return false
	}
	if name := owner.Name(); ast.IsIdentifier(name) && p.assignedSymbol(p.symbol(name)) {
		return false
	}
	fn := owner.Parent
	if fn.Kind != ast.KindFunctionDeclaration || fn.Name() == nil {
		return false
	}
	symbol := p.valueSymbol(fn.Name())
	if symbol == nil || len(symbol.Declarations) != 1 || p.assignedSymbol(symbol) {
		return false
	}
	entry := p.program.Files()[0].AsNode().Symbol()
	if entry != nil {
		for _, exported := range p.checker.GetExportsOfModule(entry) {
			if p.resolvedExport(exported) == symbol {
				return false
			}
		}
	}
	parameters := fn.Parameters()
	// An explicit this parameter is first and consumes no call argument.
	if len(parameters) > 0 && ast.IsIdentifier(parameters[0].Name()) && parameters[0].Name().Text() == "this" {
		return false
	}
	index := -1
	for i, param := range parameters {
		if param == owner {
			index = i
			break
		}
	}
	if index < 0 {
		return false
	}
	found := false
	var invalid ast.Visitor
	invalid = func(node *ast.Node) bool {
		if node.Kind == ast.KindCallExpression {
			call := node.AsCallExpression()
			callee := ast.SkipParentheses(call.Expression)
			if ast.IsIdentifier(callee) && p.valueSymbol(callee) == symbol {
				found = true
				if len(call.Arguments.Nodes) <= index || hasSpread(node) || !prove(call.Arguments.Nodes[index]) {
					return true
				}
				for _, argument := range call.Arguments.Nodes {
					if invalid(argument) {
						return true
					}
				}
				return false
			}
		}
		if ast.IsIdentifier(node) && node != fn.Name() && !ast.IsPartOfTypeNode(node) && p.valueSymbol(node) == symbol {
			return true // Aliasing or an opaque call can add another provider.
		}
		return node.ForEachChild(invalid)
	}
	for _, module := range p.modules {
		if module.AsNode().ForEachChild(invalid) {
			return false
		}
	}
	return found
}
