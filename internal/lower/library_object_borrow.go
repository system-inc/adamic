package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// The standard library types Object.prototype as any. Recognize this one
// intrinsic by its global declaration, never by a structural function signature.
func (l *lowering) borrowedOwnIntrinsic(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindPropertyAccessExpression || node.Name().Text() != "hasOwnProperty" || node.AsPropertyAccessExpression().QuestionDotToken != nil {
		return false
	}
	prototype := ast.SkipParentheses(node.AsPropertyAccessExpression().Expression)
	return prototype.Kind == ast.KindPropertyAccessExpression && prototype.Name().Text() == "prototype" && prototype.AsPropertyAccessExpression().QuestionDotToken == nil && l.isLibraryGlobal(prototype.AsPropertyAccessExpression().Expression, "Object")
}

// The cached value has an honest explicit signature. Inferred any aliases and
// escaping intrinsic identity remain outside this first shape.
func (l *lowering) borrowedOwnAlias(node *ast.Node) *ast.Node {
	node = ast.SkipParentheses(node)
	if !ast.IsIdentifier(node) {
		return nil
	}
	symbol := l.symbol(node)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return nil
	}
	declaration := symbol.Declarations[0]
	if declaration.Kind != ast.KindVariableDeclaration || declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
		return nil
	}
	variable := declaration.AsVariableDeclaration()
	if variable.Type == nil || variable.Initializer == nil || !l.borrowedOwnIntrinsic(variable.Initializer) {
		return nil
	}
	signatures := l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(declaration.Name()), checker.SignatureKindCall)
	if len(signatures) != 1 {
		return nil
	}
	signature := signatures[0]
	if len(signature.Parameters()) != 1 || signature.ThisParameter() == nil || l.checker.TypeToString(l.checker.GetTypeOfSymbol(signature.ThisParameter())) != "object" || l.checker.TypeToString(l.checker.GetTypeOfSymbol(signature.Parameters()[0])) != "string" || l.checker.TypeToString(l.checker.GetReturnTypeOfSignature(signature)) != "boolean" {
		return nil
	}
	return declaration
}

func (l *lowering) borrowedOwnUses(declaration *ast.Node) error {
	symbol := l.symbol(declaration.Name())
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return err
	}
	var gap error
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if gap != nil {
			return true
		}
		if ast.IsIdentifier(node) && l.symbol(node) == symbol && node != declaration.Name() {
			parent := node.Parent
			if parent == nil || parent.Kind != ast.KindPropertyAccessExpression || parent.AsPropertyAccessExpression().Expression != node || parent.Name().Text() != "call" || parent.Parent == nil || parent.Parent.Kind != ast.KindCallExpression || parent.Parent.AsCallExpression().Expression != parent {
				gap = l.notYet(node, "Object.prototype.hasOwnProperty alias escaping its proven .call use")
				return true
			}
		}
		return node.ForEachChild(visit)
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	if statement := declaration.Parent.Parent; statement != nil && ast.HasSyntacticModifier(statement, ast.ModifierFlagsExport) {
		return l.notYet(declaration, "exported Object.prototype.hasOwnProperty alias identity")
	}
	return gap
}

func (l *lowering) borrowedOwnClosure() ir.Expression {
	function := len(l.result.Functions)
	object, key := len(l.result.Locals), len(l.result.Locals)+1
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "object", Type: ir.Object, Function: function}, ir.Local{Name: "key", Type: ir.String, Function: function})
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "library_borrowed_has_own", Parameters: []int{object, key}, Returns: ir.Boolean, Closure: true, Body: []ir.Statement{ir.Return{Value: ir.HasOwn{Object: ir.Read{Local: object, Of: ir.Object}, Key: ir.Read{Local: key, Of: ir.String}}}}})
	return ir.MakeClosure{Function: function}
}

// Retain the cached binding's real initialization and reads. Erasing it would
// incorrectly allow a call through a binding still in its temporal dead zone.
func (l *lowering) borrowedOwnValue(node *ast.Node) (ir.Expression, bool, error) {
	if !l.borrowedOwnIntrinsic(node) {
		return nil, false, nil
	}
	outer := node
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	if outer.Parent == nil || outer.Parent.Kind != ast.KindVariableDeclaration || l.borrowedOwnAlias(outer.Parent.Name()) == nil {
		return nil, true, l.notYet(node, "Object.prototype.hasOwnProperty value without a const (this: object, key: string) => boolean signature")
	}
	if err := l.borrowedOwnUses(outer.Parent); err != nil {
		return nil, true, err
	}
	return l.borrowedOwnClosure(), true, nil
}

func (l *lowering) borrowedOwnCall(node *ast.Node) (ir.Expression, bool, error) {
	if node.Kind != ast.KindCallExpression {
		return nil, false, nil
	}
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if callee.Kind != ast.KindPropertyAccessExpression || callee.Name().Text() != "call" {
		return nil, false, nil
	}
	receiver := ast.SkipParentheses(callee.AsPropertyAccessExpression().Expression)
	direct := l.borrowedOwnIntrinsic(receiver)
	alias := l.borrowedOwnAlias(receiver)
	if direct {
		// The merged intrinsic handler also proves primitive and indexed receivers.
		// Keep that wider supported family; this path specializes cached aliases.
		return nil, false, nil
	}
	if !direct && alias == nil {
		return nil, false, nil
	}
	if call.QuestionDotToken != nil || callee.AsPropertyAccessExpression().QuestionDotToken != nil || len(call.Arguments.Nodes) != 2 || hasSpread(node) {
		return nil, true, l.notYet(node, "Object.prototype.hasOwnProperty.call with other than one present receiver and string key")
	}
	if !l.exactObject(call.Arguments.Nodes[0], 0) {
		return nil, true, l.notYet(node, "Object.prototype.hasOwnProperty.call without a proven complete plain literal shape")
	}
	if key, known := l.representation(l.checker.GetTypeAtLocation(call.Arguments.Nodes[1])); !known || key != ir.String || l.includesUndefined(l.checker.GetTypeAtLocation(call.Arguments.Nodes[1])) || l.includesNull(l.checker.GetTypeAtLocation(call.Arguments.Nodes[1])) {
		return nil, true, l.notYet(node, "Object.prototype.hasOwnProperty.call with a key requiring ToPropertyKey")
	}
	var closure ir.Expression
	var err error
	if direct {
		closure = l.borrowedOwnClosure()
	} else {
		if err := l.borrowedOwnUses(alias); err != nil {
			return nil, true, err
		}
		closure, err = l.expression(receiver)
		if err != nil {
			return nil, true, err
		}
	}
	arguments := []ir.Expression{}
	for _, argument := range call.Arguments.Nodes {
		value, err := l.expression(argument)
		if err != nil {
			return nil, true, err
		}
		arguments = append(arguments, value)
	}
	return ir.CallClosure{Closure: closure, Arguments: arguments, Returns: ir.Boolean}, true, nil
}

// This detached intrinsic always receives an explicit receiver at the admitted
// call sites. It therefore preserves this rather than losing it.
func (l *lowering) borrowedOwnReadAllowed(node *ast.Node) bool {
	if !l.borrowedOwnIntrinsic(node) {
		return false
	}
	outer := node
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	parent := outer.Parent
	if parent == nil {
		return false
	}
	if parent.Kind == ast.KindVariableDeclaration {
		return l.borrowedOwnAlias(parent.Name()) != nil
	}
	return parent.Kind == ast.KindPropertyAccessExpression && parent.Name().Text() == "call" && parent.AsPropertyAccessExpression().Expression == outer && called(parent)
}
