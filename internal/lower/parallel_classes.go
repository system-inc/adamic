package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// Admit only a source method of the receiver's concrete, heritage-free class.
// Other dispatch shapes remain refused. The ordinary effect proof still visits
// every argument, method body, capture and transitive call.
func (p *parallelProof) readonlyMethod(callee *ast.Node, chain []string) (bool, error) {
	access := callee.AsPropertyAccessExpression()
	symbol := p.l.symbol(callee)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return false, nil
	}
	method := symbol.Declarations[0]
	if method.Kind != ast.KindMethodDeclaration || method.Body() == nil || ast.HasSyntacticModifier(method, ast.ModifierFlagsStatic) {
		return false, nil
	}
	class := method.Parent
	if class == nil || class.Kind != ast.KindClassDeclaration {
		return false, nil
	}
	receiver := p.l.concrete(p.l.checker.GetTypeAtLocation(access.Expression))
	if receiver.Symbol() == nil || receiver.Symbol() != p.l.symbol(class.Name()) {
		return false, nil
	}
	// Until all-target method resolution is available at this source proof stage,
	// refuse a program containing any inheritance, even an unrelated hierarchy.
	modules, err := p.l.moduleOrder(p.l.program.Files()[0])
	if err != nil {
		return true, err
	}
	unstable := false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindHeritageClause && ast.IsClassLike(node.Parent) {
			unstable = true
		}
		// A source declaration is not a dispatch proof if its method can be replaced.
		// Name matching deliberately includes writes through structural aliases.
		if target := parallelWrite(node); target != nil {
			target = ast.SkipParentheses(target)
			if target.Kind == ast.KindPropertyAccessExpression && target.Name().Text() == method.Name().Text() {
				unstable = true
			}
			if target.Kind == ast.KindElementAccessExpression {
				unstable = true
			}
		}
		return node.ForEachChild(visit)
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	if unstable {
		return false, nil
	}
	if why := p.l.shareableWith(receiver, "receiver", access.Expression, p.shareableFunctions); why != "" {
		return true, p.effect(callee, chain, "has a receiver that is not Shareable: "+why)
	}
	if p.readonlyMethods == nil {
		p.readonlyMethods = map[*ast.Node]bool{}
	}
	p.readonlyMethods[method] = true
	return true, p.function(method, chain)
}

// This is only a direct readonly data-field read in the admitted method itself.
// No alias, nested closure, accessor, method reference or this escape is blessed.
func (p *parallelProof) readonlyThis(node, owner *ast.Node) bool {
	if !p.readonlyMethods[owner] {
		return false
	}
	parent := node.Parent
	if parent == nil || parent.Kind != ast.KindPropertyAccessExpression || parent.AsPropertyAccessExpression().Expression != node {
		return false
	}
	symbol := p.l.symbol(parent)
	if symbol == nil || !p.l.checker.IsReadonlySymbol(symbol) || len(symbol.Declarations) != 1 {
		return false
	}
	field := symbol.Declarations[0]
	return field.Kind == ast.KindPropertyDeclaration && field.Parent == owner.Parent && !ast.HasSyntacticModifier(field, ast.ModifierFlagsStatic)
}
