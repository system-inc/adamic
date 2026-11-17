package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// Registration is the trust gate, not the interface's spelling (Adamic #bxaz0p1).
func (r *workerEnvironmentRules) kv(t *checker.Type) bool {
	if !load.WorkerKV(r.l.checker, t) {
		return false
	}
	_, registered := r.l.nodeHostObject(t)
	return registered
}
func (r *workerEnvironmentRules) kvRefused(n *ast.Node, form string) error {
	return &Refused{Where: r.l.program.Where(n), What: "workers:KvNamespace in " + form, Fix: "read a declared environment binding and keep it in an exact const local or parameter"}
}
func (r *workerEnvironmentRules) kvSlot(n *ast.Node, t *checker.Type) bool {
	if !r.kv(t) {
		return false
	}
	switch n.Kind {
	case ast.KindParameter:
		p := n.AsParameterDeclaration()
		return !ast.IsThisIdentifier(n.Name()) && p.QuestionToken == nil && p.DotDotDotToken == nil && !ast.HasSyntacticModifier(n, ast.ModifierFlagsParameterPropertyModifier)
	case ast.KindVariableDeclaration:
		return n.Parent.Flags&ast.NodeFlagsConst != 0 && workerEnvironmentFunction(n) != nil
	case ast.KindBindingElement:
		for p := n.Parent; p != nil && !ast.IsFunctionLike(p); p = p.Parent {
			if p.Kind == ast.KindParameter {
				return r.exact(r.l.checker.GetTypeAtLocation(p))
			}
			if p.Kind == ast.KindVariableDeclaration {
				return p.Name().Kind == ast.KindObjectBindingPattern && p.Initializer() != nil && r.environmentValue(r.l.checker.GetTypeAtLocation(p.Initializer())) && p.Parent.Flags&ast.NodeFlagsConst != 0 && workerEnvironmentFunction(p) != nil
			}
		}
	case ast.KindPropertySignature:
		if r.environment != nil {
			for _, f := range r.l.checker.GetPropertiesOfType(r.environment) {
				for _, d := range f.Declarations {
					if d == n {
						return true
					}
				}
			}
		}
	}
	return false
}
func (r *workerEnvironmentRules) kvOrigin(n *ast.Node) bool {
	n = ast.SkipParentheses(n)
	if (n.Kind == ast.KindPropertyAccessExpression || n.Kind == ast.KindElementAccessExpression) && r.environmentValue(r.l.checker.GetTypeAtLocation(n.Expression())) && r.declared(n) {
		return true
	}
	if n.Kind != ast.KindIdentifier {
		return false
	}
	symbol := r.l.checker.GetSymbolAtLocation(n)
	if symbol == nil {
		return false
	}
	for _, d := range symbol.Declarations {
		if r.kvSlot(d, r.l.checker.GetTypeOfSymbol(symbol)) {
			return true
		}
	}
	return false
}

// Lower #bxaz0p1 rule 3: a function with a KV parameter is only a direct callee.
func (r *workerEnvironmentRules) kvFunction(symbol *ast.Symbol) bool {
	if symbol == nil {
		return false
	}
	if symbol.Flags&ast.SymbolFlagsAlias != 0 {
		symbol = r.l.checker.GetAliasedSymbol(symbol)
	}
	for _, d := range symbol.Declarations {
		if !ast.IsFunctionLike(d) {
			continue
		}
		for _, p := range d.Parameters() {
			if t := r.l.checker.GetTypeAtLocation(p); r.kvSlot(p, t) && r.holds(t, map[*checker.Type]bool{}, 0, true) {
				return true
			}
		}
	}
	return false
}

// KV purity facts use existing function source positions, never new IR fields.
func (r *workerEnvironmentRules) noteCapability(n *ast.Node, t *checker.Type) {
	if r.capabilitySites == nil || t == nil || !r.holds(t, map[*checker.Type]bool{}, 0, true) {
		return
	}
	function := workerEnvironmentFunction(n)
	if ast.IsFunctionLike(n) {
		function = n
	}
	if function == nil {
		return
	}
	functions := []*ast.Node{function}
	if name := function.Name(); name != nil {
		if symbol := r.l.checker.GetSymbolAtLocation(name); symbol != nil {
			functions = append(functions, symbol.Declarations...)
		}
	}
	for _, declaration := range functions {
		if !ast.IsFunctionLike(declaration) {
			continue
		}
		name := declaration.Name()
		if name == nil {
			name = declaration
		}
		file := ast.GetSourceFileOfNode(name)
		line, column := scanner.GetLineAndCharacterOfPosition(file, scanner.GetTokenPosOfNode(name, file, false))
		r.capabilitySites[ir.SourcePosition{File: r.l.program.FileName(file), Line: line + 1, Column: column + 1}] = true
	}
}
func (r *workerEnvironmentRules) envOrigin(n *ast.Node) bool {
	n = ast.SkipParentheses(n)
	if n.Kind != ast.KindIdentifier {
		return false
	}
	symbol := r.l.checker.GetSymbolAtLocation(n)
	if symbol == nil || !r.exact(r.l.checker.GetTypeOfSymbol(symbol)) {
		return false
	}
	for _, d := range symbol.Declarations {
		if d.Kind == ast.KindParameter {
			return true
		}
		if d.Kind == ast.KindVariableDeclaration && d.Parent.Flags&ast.NodeFlagsConst != 0 && workerEnvironmentFunction(d) != nil {
			return true
		}
	}
	return false
}

func (r *workerEnvironmentRules) kvUse(n *ast.Node) error {
	// Rule 1 includes inferred generic bindings; judgeBinding's inventory supplies them.
	if err := r.l.typeArgumentBindings(n, func(b typeArgumentBinding) error {
		if r.holds(b.argument, map[*checker.Type]bool{}, 0, true) {
			site := b.written
			if site == nil {
				site = b.site
			}
			return r.kvRefused(site, "a generic binding")
		}
		return nil
	}); err != nil {
		return err
	}
	if n.Kind == ast.KindParenthesizedExpression || !r.l.isExpression(n) || ast.IsDeclarationName(n) || hasTypeAncestor(n) {
		return nil
	}
	if n.Parent != nil && n.Parent.Kind == ast.KindPropertyAccessExpression && n.Parent.Name() == n {
		return nil
	}
	// Rule 4 precedes the write's result and right-hand value.
	field := n
	if n.Kind == ast.KindBinaryExpression && ast.IsAssignmentTarget(n.AsBinaryExpression().Left) {
		field = ast.SkipParentheses(n.AsBinaryExpression().Left)
	}
	if (field.Kind == ast.KindPropertyAccessExpression || field.Kind == ast.KindElementAccessExpression) && ast.IsAssignmentTarget(field) && r.kv(r.l.checker.GetTypeAtLocation(field)) && r.environmentValue(r.l.checker.GetTypeAtLocation(field.Expression())) && r.declared(field) {
		return r.kvRefused(field, "a write to an environment field")
	}
	if n.Kind == ast.KindBinaryExpression && ast.IsAssignmentTarget(n.AsBinaryExpression().Left) && n.Parent != nil && n.Parent.Kind == ast.KindExpressionStatement {
		return nil
	}
	if ast.IsAssignmentTarget(n) && r.kv(r.l.checker.GetTypeAtLocation(n)) && r.parameterField(n) {
		return r.kvRefused(n, "a write to a destructured parameter binding")
	}
	if n.Kind == ast.KindIdentifier {
		symbol := r.l.checker.GetSymbolAtLocation(n)
		if r.kvFunction(symbol) {
			site := n
			for site.Parent != nil && site.Parent.Kind == ast.KindParenthesizedExpression {
				site = site.Parent
			}
			if site.Parent == nil || site.Parent.Kind != ast.KindCallExpression || site.Parent.AsCallExpression().Expression != site {
				return r.kvRefused(n, "a function used as a value instead of a direct callee")
			}
		}
		if symbol != nil {
			declared := r.l.checker.GetTypeOfSymbol(symbol)
			if r.kv(declared) || r.exact(declared) && r.holds(declared, map[*checker.Type]bool{}, 0, true) {
				for _, d := range symbol.Declarations {
					if owner := workerEnvironmentFunction(d); owner != nil && owner != workerEnvironmentFunction(n) {
						return r.kvRefused(n, "a closure capture")
					}
				}
			}
		}
	}
	name, source, target := r.l.relationSite(n)
	// Lower rule 2: every value into an admitted KV slot is an admitted read.
	// relationSite owns defaults, initializers, arguments, assignments and assertions.
	if r.exact(target) && r.holds(target, map[*checker.Type]bool{}, 0, true) && !r.envOrigin(n) {
		return r.kvRefused(n, "a program-created environment holding a binding")
	}
	if (n.Kind == ast.KindArrayLiteralExpression || n.Kind == ast.KindObjectLiteralExpression) && r.holds(r.l.checker.GetTypeAtLocation(n), map[*checker.Type]bool{}, 0, true) && !r.kv(r.l.checker.GetTypeAtLocation(n)) {
		return r.kvRefused(n, "a container value")
	}
	if r.kv(target) {
		if name == "assertion" || !r.kvOrigin(n) {
			return r.kvRefused(n, name+" from a program value")
		}
		return nil
	}
	if !r.kv(r.l.checker.GetTypeAtLocation(n)) {
		return nil
	}
	if target != nil {
		if reached := r.l.nodeHostViewAt(source, target); reached != "" {
			return r.l.nodeHostViewRefusal(n, source, target, reached)
		}
	}
	site := n
	for site.Parent != nil && site.Parent.Kind == ast.KindParenthesizedExpression {
		site = site.Parent
	}
	if parent := site.Parent; parent != nil && (parent.Kind == ast.KindPropertyAccessExpression || parent.Kind == ast.KindElementAccessExpression) && parent.Expression() == site && r.kvOrigin(n) {
		return nil
	}
	return r.kvRefused(n, "a value outside an admitted KV slot")
}
