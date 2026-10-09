package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// A class method is statically called; own literal methods and arrow fields are closures with
// different argument layouts. A structural view must not hide a different layout.
func memberConvention(member *ast.Symbol) *ast.Node {
	if member == nil || len(member.Declarations) == 0 {
		return nil
	}
	declaration := member.Declarations[0]
	if declaration.Kind == ast.KindMethodDeclaration {
		if declaration.Parent.Kind == ast.KindClassDeclaration {
			return declaration.Parent
		}
		return objectMethodConvention
	}
	return arrowMethodConvention
}

var objectMethodConvention = &ast.Node{}
var arrowMethodConvention = &ast.Node{}

// Fresh literal assignability performs excess-property checking. Runtime values do not lose
// their extra fields when viewed structurally, so compare required members individually instead.
func (l *lowering) iterationShapeFits(shape, view *checker.Type) bool {
	for _, expected := range l.checker.GetPropertiesOfType(view) {
		actual := l.checker.GetPropertyOfType(shape, expected.Name)
		if actual == nil {
			if expected.Flags&ast.SymbolFlagsOptional != 0 {
				continue
			}
			return false
		}
		if actual.Flags&ast.SymbolFlagsOptional != 0 && expected.Flags&ast.SymbolFlagsOptional == 0 {
			return false
		}
		if !l.checker.IsTypeAssignableTo(l.checker.GetTypeOfSymbol(actual), l.checker.GetTypeOfSymbol(expected)) {
			return false
		}
	}
	return true
}

// A property typed as an ordinary callback can hide a literal method. Its closure needs a
// receiver argument. Do not read it using the callback convention until that origin is retained.
func (l *lowering) erasedLiteralMethod(node *ast.Node) error {
	access := node.AsPropertyAccessExpression()
	of, known := l.representation(l.checker.GetTypeAtLocation(node))
	if !known || of != ir.Closure {
		return nil
	}
	member := l.checker.GetSymbolAtLocation(node)
	if member != nil && literalMethod(member) {
		return nil
	}
	return l.erasedMethodField(node, l.checker.GetTypeAtLocation(access.Expression), access.Name().Text())
}

func (l *lowering) erasedMethodField(where *ast.Node, view *checker.Type, name string) error {
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return err
	}
	var found error
	var visit ast.Visitor
	visit = func(candidate *ast.Node) bool {
		if found != nil {
			return true
		}
		if candidate.Kind == ast.KindObjectLiteralExpression || candidate.Kind == ast.KindNewExpression {
			shape := l.checker.GetTypeAtLocation(candidate)
			field := l.checker.GetPropertyOfType(shape, name)
			if field != nil && l.iterationShapeFits(shape, view) {
				if !isCallee(where) {
					for _, root := range l.checker.GetRootSymbols(field) {
						for _, declaration := range root.Declarations {
							var initializer *ast.Node
							if declaration.Kind == ast.KindPropertyAssignment {
								initializer = declaration.AsPropertyAssignment().Initializer
							}
							if declaration.Kind == ast.KindPropertyDeclaration {
								initializer = declaration.AsPropertyDeclaration().Initializer
							}
							if initializer != nil && dynamicReceiverFunction(ast.SkipParentheses(initializer)) {
								found = &Refused{Where: l.program.Where(where), What: "a receiver-dependent function field read as a value (unbound-method)", Fix: "call it in an arrow that keeps its object"}
								return true
							}
						}
					}
				}

				if literalMethod(field) {
					if isCallee(where) { // Calls carry the runtime closure's receiver convention.
						return candidate.ForEachChild(visit)
					}
					found = l.notYet(where, "a literal method through a view that erases its receiver")
					return true
				}
				for _, root := range l.checker.GetRootSymbols(field) {
					for _, declaration := range root.Declarations {
						if declaration.Kind == ast.KindMethodDeclaration && declaration.Parent.Kind == ast.KindClassDeclaration {
							// Method calls preserve the existing class prototype dispatch. Arrow-field
							// views and destructuring still erase the represented method origin.
							member := l.checker.GetPropertyOfType(l.checker.GetNonNullableType(view), name)
							if isCallee(where) && (isClassInstance(l.checker.GetNonNullableType(view)) || (member != nil && member.Flags&ast.SymbolFlagsMethod != 0) || optionalInvocationCallee(where)) {
								continue
							}
							found = l.notYet(where, "a class method through a view that erases its prototype origin")
							return true
						}
					}
				}
			}
		}
		return candidate.ForEachChild(visit)
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	return found
}

// Known factories pin the actual method origins. Comparing every structurally compatible shape
// would otherwise refuse ordinary Range/Iterator class pairs with overlapping public fields.
func (l *lowering) iterationOrigin(node *ast.Node, members []*ast.Symbol, closePresence *bool, depth int) bool {
	if node == nil || depth > 16 {
		return false
	}
	node = ast.SkipParentheses(node)
	switch node.Kind {
	case ast.KindObjectLiteralExpression, ast.KindNewExpression, ast.KindThisKeyword:
		if node.Kind == ast.KindObjectLiteralExpression {
			for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
				if property.Kind == ast.KindSpreadAssignment {
					return false
				}
			}
		}
		if node.Kind == ast.KindNewExpression {
			class := l.classes[l.symbol(node.AsNewExpression().Expression)]
			if class == nil {
				return false
			}
			for _, member := range class.Members() {
				if member.Kind == ast.KindConstructor && hasConstructorReturn(member.Body()) {
					return false
				}
			}
		}
		shape := l.checker.GetTypeAtLocation(node)
		for _, expected := range members {
			actual := l.checker.GetPropertyOfType(shape, expected.Name)
			if actual == nil || !l.sameMemberOrigin(actual, expected) {
				return false
			}
		}
		if closePresence != nil && (l.checker.GetPropertyOfType(shape, "return") != nil) != *closePresence {
			return false
		}
		return true
	case ast.KindIdentifier:
		symbol := l.symbol(node)
		if symbol == nil || len(symbol.Declarations) != 1 {
			return false
		}
		declaration := symbol.Declarations[0]
		if declaration.Kind != ast.KindVariableDeclaration || declaration.Parent == nil || declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
			return false
		}
		return l.iterationOrigin(declaration.AsVariableDeclaration().Initializer, members, closePresence, depth+1)
	case ast.KindCallExpression:
		callee := ast.SkipParentheses(node.AsCallExpression().Expression)
		if !ast.IsIdentifier(callee) {
			return false
		}
		symbol := l.symbol(callee)
		if symbol == nil || len(symbol.Declarations) != 1 {
			return false
		}
		body := l.iterationCallableBody(symbol, depth+1)
		return l.iterationReturns(body, members, closePresence, depth+1)
	case ast.KindConditionalExpression:
		expression := node.AsConditionalExpression()
		return l.iterationOrigin(expression.WhenTrue, members, closePresence, depth+1) && l.iterationOrigin(expression.WhenFalse, members, closePresence, depth+1)
	}
	return false
}

func (l *lowering) sameMemberOrigin(actual, expected *ast.Symbol) bool {
	wanted := map[*ast.Node]bool{}
	for _, root := range l.checker.GetRootSymbols(expected) {
		for _, declaration := range root.Declarations {
			wanted[declaration] = true
		}
	}
	count := 0
	for _, root := range l.checker.GetRootSymbols(actual) {
		for _, declaration := range root.Declarations {
			if !wanted[declaration] {
				return false
			}
			count++
		}
	}
	return count > 0
}

func (l *lowering) iterationReturns(body *ast.Node, members []*ast.Symbol, closePresence *bool, depth int) bool {
	if body == nil {
		return false
	}
	if body.Kind != ast.KindBlock {
		return l.iterationOrigin(body, members, closePresence, depth+1)
	}
	found, sound := false, true
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if ast.IsFunctionLike(node) {
			return false
		}
		if node.Kind == ast.KindReturnStatement {
			found = true
			sound = sound && l.iterationOrigin(node.AsReturnStatement().Expression, members, closePresence, depth+1)
			return false
		}
		return node.ForEachChild(visit)
	}
	body.ForEachChild(visit)
	return found && sound
}

func (l *lowering) iterationFactoryKnown(member *ast.Symbol, next, close *ast.Symbol) bool {
	roots := l.checker.GetRootSymbols(member)
	if len(roots) != 1 || len(roots[0].Declarations) != 1 {
		return false
	}
	declaration := roots[0].Declarations[0]
	var body *ast.Node
	switch declaration.Kind {
	case ast.KindMethodDeclaration:
		body = declaration.Body()
	case ast.KindPropertyAssignment:
		initializer := ast.SkipParentheses(declaration.AsPropertyAssignment().Initializer)
		if initializer.Kind == ast.KindArrowFunction {
			body = initializer.Body()
		}
	}
	members := []*ast.Symbol{next}
	if close != nil {
		members = append(members, close)
	}
	presence := close != nil
	return l.iterationReturns(body, members, &presence, 0)
}

func hasConstructorReturn(body *ast.Node) bool {
	if body == nil {
		return false
	}
	found := false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if ast.IsFunctionLike(node) {
			return false
		}
		if node.Kind == ast.KindReturnStatement {
			found = true
			return true
		}
		return node.ForEachChild(visit)
	}
	body.ForEachChild(visit)
	return found
}

func (l *lowering) iterationCallableBody(symbol *ast.Symbol, depth int) *ast.Node {
	if symbol == nil || len(symbol.Declarations) != 1 || depth > 16 {
		return nil
	}
	declaration := symbol.Declarations[0]
	if declaration.Kind == ast.KindFunctionDeclaration {
		return declaration.Body()
	}
	if declaration.Kind != ast.KindVariableDeclaration || declaration.Parent == nil || declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
		return nil
	}
	initializer := declaration.AsVariableDeclaration().Initializer
	if initializer == nil {
		return nil
	}
	initializer = ast.SkipParentheses(initializer)
	if initializer.Kind == ast.KindArrowFunction {
		return initializer.Body()
	}
	if ast.IsIdentifier(initializer) {
		return l.iterationCallableBody(l.symbol(initializer), depth+1)
	}
	return nil
}

// Generic classes are monomorphized. A readonly structural view can change their arguments
// without changing method declarations, so declaration identity alone is not a sufficient proof.
func (l *lowering) sameIterationClass(actual, expected *checker.Type) bool {
	actual, expected = l.concrete(actual), l.concrete(expected)
	if !isClassInstance(actual) || !isClassInstance(expected) || actual.Symbol() != expected.Symbol() {
		return false
	}
	left, right := l.checker.GetTypeArguments(actual), l.checker.GetTypeArguments(expected)
	if len(left) != len(right) {
		return false
	}
	for index, arg := range left {
		other := right[index]
		if !l.checker.IsTypeAssignableTo(arg, other) || !l.checker.IsTypeAssignableTo(other, arg) {
			return false
		}
		if (isClassInstance(arg) || isClassInstance(other)) && arg.Symbol() != other.Symbol() {
			return false
		}
	}
	return true
}

func (l *lowering) knownIterationClass(node *ast.Node, expected *checker.Type, depth int) bool {
	if node == nil || depth > 16 {
		return false
	}
	node = ast.SkipParentheses(node)
	switch node.Kind {
	case ast.KindNewExpression:
		return l.sameIterationClass(l.checker.GetTypeAtLocation(node), expected)
	case ast.KindIdentifier:
		symbol := l.symbol(node)
		if symbol == nil || len(symbol.Declarations) != 1 {
			return false
		}
		declaration := symbol.Declarations[0]
		if declaration.Kind != ast.KindVariableDeclaration || declaration.Parent == nil || declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
			return false
		}
		return l.knownIterationClass(declaration.AsVariableDeclaration().Initializer, expected, depth+1)
	case ast.KindCallExpression:
		callee := ast.SkipParentheses(node.AsCallExpression().Expression)
		if !ast.IsIdentifier(callee) {
			return false
		}
		return l.iterationReturnsClass(l.iterationCallableBody(l.symbol(callee), depth+1), expected, nil, depth+1)
	case ast.KindConditionalExpression:
		expression := node.AsConditionalExpression()
		return l.knownIterationClass(expression.WhenTrue, expected, depth+1) && l.knownIterationClass(expression.WhenFalse, expected, depth+1)
	}
	return false
}

func (l *lowering) iterationReturnsClass(body *ast.Node, expected, receiver *checker.Type, depth int) bool {
	if body == nil {
		return false
	}
	matches := func(node *ast.Node) bool {
		if node != nil && ast.SkipParentheses(node).Kind == ast.KindThisKeyword {
			return receiver != nil && l.sameIterationClass(receiver, expected)
		}
		return l.knownIterationClass(node, expected, depth+1)
	}
	if body.Kind != ast.KindBlock {
		return matches(body)
	}
	found, sound := false, true
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if ast.IsFunctionLike(node) {
			return false
		}
		if node.Kind == ast.KindReturnStatement {
			found = true
			sound = sound && matches(node.AsReturnStatement().Expression)
			return false
		}
		return node.ForEachChild(visit)
	}
	body.ForEachChild(visit)
	return found && sound
}

func (l *lowering) genericIteratorFactoryKnown(member *ast.Symbol, source, iterator *checker.Type) bool {
	roots := l.checker.GetRootSymbols(member)
	if len(roots) != 1 || len(roots[0].Declarations) != 1 {
		return false
	}
	declaration := roots[0].Declarations[0]
	var body *ast.Node
	if declaration.Kind == ast.KindMethodDeclaration {
		body = declaration.Body()
	}
	return l.iterationReturnsClass(body, iterator, source, 0)
}

// A returned receiver keeps its dynamic subclass, even when the result annotation names
// the base. Follow local aliases too; they do not make this an exact base instance.
func (l *lowering) iterationFactoryReturnsThis(member *ast.Symbol) bool {
	var receiver func(*ast.Node, int) bool
	receiver = func(node *ast.Node, depth int) bool {
		if node == nil || depth > 16 {
			return false
		}
		node = ast.SkipParentheses(node)
		if node.Kind == ast.KindThisKeyword {
			return true
		}
		if ast.IsIdentifier(node) {
			symbol := l.symbol(node)
			if symbol != nil && len(symbol.Declarations) == 1 && symbol.Declarations[0].Kind == ast.KindVariableDeclaration {
				return receiver(symbol.Declarations[0].AsVariableDeclaration().Initializer, depth+1)
			}
		}
		return false
	}
	for _, root := range l.checker.GetRootSymbols(member) {
		for _, declaration := range root.Declarations {
			if declaration.Kind != ast.KindMethodDeclaration || declaration.Body() == nil {
				continue
			}
			found := false
			var visit ast.Visitor
			visit = func(node *ast.Node) bool {
				if ast.IsFunctionLike(node) {
					return false
				}
				if node.Kind == ast.KindReturnStatement && receiver(node.AsReturnStatement().Expression, 0) {
					found = true
					return true
				}
				return node.ForEachChild(visit)
			}
			declaration.Body().ForEachChild(visit)
			if found {
				return true
			}
		}
	}
	return false
}

// Unknown receiver origins matter only when a declared descendant can change the
// protocol. A factory returning a base receiver remains sound in a closed class tree.
func (l *lowering) iteratorReceiverOverrides(modules []*ast.SourceFile, iterator *checker.Type, next, close *ast.Symbol) bool {
	class := l.classNodeFor(iterator)
	if class == nil {
		return false
	}
	changed := false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindClassDeclaration && node.Name() != nil && node != class {
			shape := l.checker.GetTypeAtLocation(node.Name())
			if l.classView(shape, class) != nil {
				actualNext := l.checker.GetPropertyOfType(shape, "next")
				actualClose := l.checker.GetPropertyOfType(shape, "return")
				if actualNext == nil || !l.sameMemberOrigin(actualNext, next) || (actualClose != nil) != (close != nil) || (close != nil && !l.sameMemberOrigin(actualClose, close)) {
					changed = true
					return true
				}
			}
		}
		return node.ForEachChild(visit)
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	return changed
}

// Arrows retain lexical this; only an ordinary function needs the call-site
// receiver. Do not let a structural property type erase that dependency.
func dynamicReceiverFunction(node *ast.Node) bool {
	if node.Kind != ast.KindFunctionExpression {
		return false
	}
	if ast.GetThisParameter(node) != nil {
		return true
	}
	found := false
	var visit ast.Visitor
	visit = func(inner *ast.Node) bool {
		if ast.IsFunctionLike(inner) && inner.Kind != ast.KindArrowFunction {
			return false
		}
		if inner.Kind == ast.KindThisKeyword && !ast.IsPartOfTypeNode(inner) {
			found = true
		}
		return inner.ForEachChild(visit)
	}
	if node.Body() != nil {
		node.Body().ForEachChild(visit)
	}
	return found
}
