package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

// Regular enum declarations emit a hoisted var in JavaScript. A call before that var's assignment
// can observe undefined, not a lexical TDZ. Until that failure participates in native exception
// cleanup, refuse such reads rather than emit a different error or dereference NULL.
func (l *lowering) enumInitialization(modules []*ast.SourceFile) error {
	pending := 0
	for _, module := range modules {
		for _, statement := range module.Statements.Nodes {
			if statement.Kind == ast.KindEnumDeclaration && !ast.HasSyntacticModifier(statement, ast.ModifierFlagsConst) {
				pending++
			}
		}
	}
	if pending == 0 {
		return nil
	}
	initialized := map[*ast.Node]bool{}
	var visit func(*ast.Node, map[*ast.Node]bool, bool) error
	visit = func(node *ast.Node, visiting map[*ast.Node]bool, execute bool) error {
		if node == nil || ast.IsTypeNode(node) || visiting[node] {
			return nil
		}
		visiting[node] = true
		defer delete(visiting, node)
		if ast.IsFunctionLike(node) && !execute {
			return nil
		}
		if declaration := l.enumObject(node); declaration != nil && !ast.HasSyntacticModifier(declaration, ast.ModifierFlagsConst) && !initialized[declaration] {
			return l.notYet(node, "reading an enum before its runtime initialization; move the call after the enum declaration")
		}
		// A callback argument is conservatively executed. Following named aliases and property
		// initializers covers calls whose syntax does not directly name a function declaration.
		callable := func(callee *ast.Node) *ast.Node {
			seen := map[*ast.Symbol]bool{}
			for callee != nil {
				callee = ast.SkipParentheses(callee)
				if ast.IsFunctionLike(callee) {
					return callee
				}
				symbol := l.symbol(callee)
				if symbol == nil || seen[symbol] {
					return nil
				}
				seen[symbol] = true
				declaration := symbol.ValueDeclaration
				if declaration == nil {
					return nil
				}
				if declaration.Kind == ast.KindVariableDeclaration && declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
					return nil
				}
				if declaration.Kind == ast.KindPropertyAssignment && !l.checker.IsReadonlySymbol(symbol) {
					return nil
				}
				if ast.IsFunctionLike(declaration) && declaration.Body() == nil && !load.IsLibrary(ast.GetSourceFileOfNode(declaration)) && !load.IsPrelude(ast.GetSourceFileOfNode(declaration)) {
					return nil
				}
				if declaration.Kind == ast.KindVariableDeclaration || declaration.Kind == ast.KindPropertyAssignment {
					callee = declaration.Initializer()
					continue
				}
				return declaration
			}
			return nil
		}
		if execute && !ast.IsFunctionLike(node) {
			if declaration := callable(node); declaration != nil && ast.IsFunctionLike(declaration) {
				if err := visit(declaration, visiting, true); err != nil {
					return err
				}
			} else if len(l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(node), checker.SignatureKindCall)) > 0 {
				return l.notYet(node, "a callback whose body is unknown before enum initialization; declare enums before executable module code")
			}
		}
		if node.Kind == ast.KindCallExpression || node.Kind == ast.KindNewExpression {
			declaration := callable(node.Expression())
			if declaration != nil && ast.IsFunctionLike(declaration) {
				if err := visit(declaration, visiting, true); err != nil {
					return err
				}
			} else if declaration == nil || !load.IsLibrary(ast.GetSourceFileOfNode(declaration)) && !load.IsPrelude(ast.GetSourceFileOfNode(declaration)) {
				return l.notYet(node, "an indirect call or class construction before enum initialization; declare enums before executable module code")
			}
			for _, argument := range node.Arguments() {
				if err := visit(argument, visiting, true); err != nil {
					return err
				}
			}
		}
		var found error
		node.ForEachChild(func(child *ast.Node) bool {
			if found == nil {
				found = visit(child, visiting, false)
			}
			return found != nil
		})
		return found
	}
	for _, module := range modules {
		for _, statement := range module.Statements.Nodes {
			if statement.Kind == ast.KindEnumDeclaration {
				initialized[statement] = true
				if !ast.HasSyntacticModifier(statement, ast.ModifierFlagsConst) {
					pending--
				}
				continue
			}
			if pending == 0 {
				continue
			}
			if statement.Kind == ast.KindFunctionDeclaration {
				continue
			}
			if statement.Kind == ast.KindClassDeclaration {
				for _, member := range statement.AsClassDeclaration().Members.Nodes {
					if ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic) || member.Kind == ast.KindClassStaticBlockDeclaration {
						if err := visit(member, map[*ast.Node]bool{}, false); err != nil {
							return err
						}
					}
				}
				continue
			}
			if err := visit(statement, map[*ast.Node]bool{}, false); err != nil {
				return err
			}
		}
	}
	return nil
}
