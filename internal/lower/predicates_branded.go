package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// A branded supertype, such as tsc's TypeElement or LiteralExpression, is an interface over many node
// kinds: its kind is wide, and a brand field typed any or void, which no value carries, keeps other nodes
// from being assignable to it. Which kinds belong is a naming convention no run-time check can recover:
// a kind set built from the types assignable to it leaves out JsxText, which isLiteralExpression takes,
// and every late-bound declaration isLateBound takes (#c7w6j0y, measured on tsc Oct 31). So a checked
// predicate claiming one checks at the call only that the value is an object, and checks its shape where
// it's relied on: S's view checks every field read through S, and so does the view of every type the
// program declares assignable to S, since narrowing, a cast or a subtype's parameter gives the value one
// of those types and a lie carrying a real subtype's kind must still stop at its first bad read. A
// subtype no view can take, a class or a generic, leaves S NotYet.

// brandedSupertype is whether target is the shape that rule is keyed on, judged by type, never by name:
// a kind property that isn't one literal, and at least one property typed exactly any or void.
func (l *lowering) brandedSupertype(target *checker.Type) bool {
	kind := l.checker.GetPropertyOfType(target, "kind")
	if kind == nil || l.fieldLiteral(target, "kind") != nil {
		return false
	}
	for _, property := range l.checker.GetPropertiesOfType(target) {
		if property.Flags&ast.SymbolFlagsOptional == 0 && l.checker.GetTypeOfSymbol(property).Flags()&(checker.TypeFlagsAny|checker.TypeFlagsVoid) != 0 {
			return true
		}
	}
	return false
}

// brandedSupertypeViews registers S's view and its subtypes' once, keeping the error that stopped them.
func (l *lowering) brandedSupertypeViews(node *ast.Node, target *checker.Type) error {
	proofs := l.proofs()
	if err, done := proofs.brandedViews[target]; done {
		return predicateFailureCopy(err)
	}
	err := l.registerBrandedSupertypeViews(node, target)
	proofs.brandedViews[target] = err
	return predicateFailureCopy(err)
}

// brandedMembership is whether target's membership check, as predicateMembership builds it, holds a
// branded supertype's: through a union's members or an array's elements, it can't be negated.
func (l *lowering) brandedMembership(target *checker.Type) bool {
	target = l.concrete(target)
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range target.Types() {
			if l.brandedMembership(member) {
				return true
			}
		}
		return false
	}
	if l.checker.IsArrayType(target) {
		return l.brandedMembership(l.checker.GetElementTypeOfArrayType(target))
	}
	return target.Flags()&checker.TypeFlagsObject != 0 && !isClassInstance(target) && l.fieldLiteral(target, "kind") == nil && l.brandedSupertype(target)
}

func (l *lowering) registerBrandedSupertypeViews(node *ast.Node, target *checker.Type) error {
	covered := []*checker.Type{target}
	var subtypes []declaredType
	for _, declaration := range l.declaredTypes() {
		if declaration.declared != target && declaration.declared.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) != 0 && l.checker.IsTypeAssignableTo(declaration.declared, target) {
			covered = append(covered, declaration.declared)
			subtypes = append(subtypes, declaration)
		}
	}
	if err := l.brandedSpreads(covered); err != nil {
		return err
	}
	if _, err := l.view(node, nil, target); err != nil {
		return err
	}
	for _, subtype := range subtypes {
		if isClassInstance(subtype.declared) {
			return l.notYet(node, "a checked predicate target "+l.checker.TypeToString(target)+" that class "+l.checker.TypeToString(subtype.declared)+" is assignable to")
		}
		if subtype.generic {
			return l.notYet(node, "a checked predicate target "+l.checker.TypeToString(target)+" that generic "+l.checker.TypeToString(subtype.declared)+" is assignable to")
		}
		if _, err := l.view(node, nil, subtype.declared); err != nil {
			return err
		}
	}
	return nil
}

// brandedSpreads keeps a branded supertype NotYet while the program copies a value of a type its views cover whole,
// by an object spread, a rest binding or Object.assign: the copy moves each field as the checker types
// it, with no view's check, so a lying layout would reach a raw read through the copy (#yq6fkp5).
func (l *lowering) brandedSpreads(covered []*checker.Type) error {
	ids := map[int]bool{}
	for _, t := range covered {
		ids[int(t.Id())] = true
	}
	var copies func(t *checker.Type) bool
	copies = func(t *checker.Type) bool {
		t = l.checker.GetNonNullableType(l.concrete(t))
		if t.Flags()&checker.TypeFlagsUnion != 0 {
			for _, member := range t.Types() {
				if copies(member) {
					return true
				}
			}
			return false
		}
		return ids[l.viewReceiverTypeID(t)]
	}
	var site *ast.Node
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		if site != nil {
			return true
		}
		switch n.Kind {
		case ast.KindSpreadAssignment:
			if copies(l.checker.GetTypeAtLocation(n.AsSpreadAssignment().Expression)) {
				site = n
			}
		case ast.KindBindingElement:
			if n.AsBindingElement().DotDotDotToken != nil && n.Parent != nil && n.Parent.Kind == ast.KindObjectBindingPattern && copies(l.checker.GetTypeAtLocation(n.Parent)) {
				site = n
			}
		case ast.KindCallExpression:
			callee := n.AsCallExpression().Expression
			if callee.Kind == ast.KindPropertyAccessExpression && callee.Name().Text() == "assign" && l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Object") {
				for _, argument := range n.AsCallExpression().Arguments.Nodes {
					if copies(l.checker.GetTypeAtLocation(argument)) {
						site = n
					}
				}
			}
		}
		if site == nil {
			n.ForEachChild(visit)
		}
		return site != nil
	}
	for _, file := range l.program.Files() {
		file.AsNode().ForEachChild(visit)
	}
	if site != nil {
		// Named at the copy, as view's own refusals are, so the reason stays one string for the meter.
		return l.notYet(site, "a whole-object copy of a branded supertype's value, which no view checks")
	}
	return nil
}

// declaredType is one interface, class or type alias the program declares, as its declared type, and
// whether it has type parameters, whose instances are other types no view registered here would cover.
type declaredType struct {
	declared *checker.Type
	generic  bool
}

// declaredTypes is every interface, class and type alias the program's own files declare, anywhere, a
// function's local ones included, found once.
func (l *lowering) declaredTypes() []declaredType {
	proofs := l.proofs()
	if proofs.declaredTypes != nil {
		return proofs.declaredTypes
	}
	proofs.declaredTypes = []declaredType{}
	seen := map[*checker.Type]bool{}
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		switch n.Kind {
		case ast.KindInterfaceDeclaration, ast.KindClassDeclaration, ast.KindTypeAliasDeclaration:
			if symbol := l.checker.GetSymbolOfDeclaration(n); symbol != nil {
				if declared := l.checker.GetDeclaredTypeOfSymbol(symbol); declared != nil && !seen[declared] {
					seen[declared] = true
					proofs.declaredTypes = append(proofs.declaredTypes, declaredType{declared: declared, generic: len(n.TypeParameters()) != 0})
				}
			}
		}
		n.ForEachChild(visit)
		return false
	}
	for _, file := range l.program.Files() {
		file.AsNode().ForEachChild(visit)
	}
	return proofs.declaredTypes
}
