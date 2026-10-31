package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
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
// branded supertype's: through a union's members or an array's elements, it can't be negated. visited
// ends a recursive alias, such as `type Nodes = Identifier | Nodes[]`, on its second visit.
func (l *lowering) brandedMembership(target *checker.Type, visited map[*checker.Type]bool) bool {
	target = l.concrete(target)
	if visited[target] {
		return false
	}
	visited[target] = true
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range target.Types() {
			if l.brandedMembership(member, visited) {
				return true
			}
		}
		return false
	}
	if l.checker.IsArrayType(target) {
		return l.brandedMembership(l.checker.GetElementTypeOfArrayType(target), visited)
	}
	return target.Flags()&checker.TypeFlagsObject != 0 && !isClassInstance(target) && l.fieldLiteral(target, "kind") == nil && l.brandedSupertype(target)
}

func (l *lowering) registerBrandedSupertypeViews(node *ast.Node, target *checker.Type) error {
	covered := []*checker.Type{target}
	for _, subtype := range l.brandedSubtypes(target) {
		covered = append(covered, subtype.declared)
	}
	if err := l.brandedSpreads(covered); err != nil {
		return err
	}
	if _, err := l.view(node, nil, target); err != nil {
		return err
	}
	for _, subtype := range l.brandedSubtypes(target) {
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

// brandedSubtypes is every type the program declares that's assignable to target, target aside, found once.
func (l *lowering) brandedSubtypes(target *checker.Type) []declaredType {
	proofs := l.proofs()
	if subtypes, found := proofs.brandedSubtypes[target]; found {
		return subtypes
	}
	subtypes := []declaredType{}
	for _, declaration := range l.declaredTypes() {
		if declaration.declared != target && declaration.declared.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) != 0 && l.checker.IsTypeAssignableTo(declaration.declared, target) {
			subtypes = append(subtypes, declaration)
		}
	}
	proofs.brandedSubtypes[target] = subtypes
	return subtypes
}

// brandedNarrowing keeps a direct call NotYet unless the checker's narrowing of its argument, on the true
// side, can only give S or a type its views cover (Compiler's musts on 9abcbdca and 8c7afca2). Every
// member of the argument's type must be a supertype of S, which narrows to S, or a declared subtype the
// views cover, which keeps itself. Any other member is NotYet, never modelled: the checker drops an
// unrelated member only beside one that's its subtype, and makes an intersection no view is registered
// for otherwise, so judging its drop by assignability would admit an X & S read raw. A branded supertype
// inside an array target is NotYet, since its elements narrow by the same rule with no call to judge at.
func (l *lowering) brandedNarrowing(node, argument *ast.Node, claim *checker.Type) error {
	claim = l.concrete(claim)
	targets := []*checker.Type{claim}
	if claim.Flags()&checker.TypeFlagsUnion != 0 {
		targets = claim.Types()
	}
	source := l.concrete(l.checker.GetTypeAtLocation(argument))
	members := []*checker.Type{source}
	if source.Flags()&checker.TypeFlagsUnion != 0 {
		members = source.Types()
	}
	for _, target := range targets {
		if !l.brandedMembership(target, map[*checker.Type]bool{}) {
			continue
		}
		if l.checker.IsArrayType(target) || target.Flags()&checker.TypeFlagsUnion != 0 {
			return l.notYet(node, "a branded supertype inside a checked predicate's array target")
		}
		covered := map[*checker.Type]bool{}
		for _, subtype := range l.brandedSubtypes(target) {
			covered[subtype.declared] = !subtype.generic && !isClassInstance(subtype.declared)
		}
		for _, member := range members {
			if !l.checker.IsTypeAssignableTo(target, member) && !(covered[member] && l.checker.IsTypeAssignableTo(member, target)) {
				return l.notYet(node, "a checked predicate target "+l.checker.TypeToString(target)+" that narrows "+l.checker.TypeToString(member)+", which is neither its supertype nor a subtype its views cover")
			}
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
		// A union's members are never unions, since the checker flattens them, so this goes one level down.
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
	for _, file := range l.programFiles() {
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

// declaredTypes is every interface, class and type alias the program's files declare, anywhere, a
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
	for _, file := range l.programFiles() {
		file.AsNode().ForEachChild(visit)
	}
	return proofs.declaredTypes
}

// programFiles is the program's files, its roots and every file they import or export from, type-only
// imports included, since a subtype declared in a file a root imports only for its types still gives
// values its type, found once. The library and the prelude declare no subtype a view must cover.
func (l *lowering) programFiles() []*ast.SourceFile {
	proofs := l.proofs()
	if proofs.files != nil {
		return proofs.files
	}
	proofs.files = []*ast.SourceFile{}
	seen := map[*ast.SourceFile]bool{}
	var visit func(file *ast.SourceFile)
	visit = func(file *ast.SourceFile) {
		if seen[file] || load.IsPrelude(file) || load.IsLibrary(file) {
			return
		}
		seen[file] = true
		proofs.files = append(proofs.files, file)
		for _, statement := range file.Statements.Nodes {
			var specifier *ast.Node
			switch statement.Kind {
			case ast.KindImportDeclaration, ast.KindExportDeclaration:
				specifier = statement.ModuleSpecifier()
			case ast.KindImportEqualsDeclaration:
				if reference := statement.AsImportEqualsDeclaration().ModuleReference; reference.Kind == ast.KindExternalModuleReference {
					specifier = reference.AsExternalModuleReference().Expression
				}
			}
			if specifier == nil {
				continue
			}
			if target := l.checker.GetSymbolAtLocation(specifier); target != nil && len(target.Declarations) > 0 && target.Declarations[0].Kind == ast.KindSourceFile {
				visit(target.Declarations[0].AsSourceFile())
			}
		}
	}
	for _, file := range l.program.Files() {
		visit(file)
	}
	return proofs.files
}
