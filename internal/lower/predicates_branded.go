package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

// A branded supertype, such as tsc's TypeElement or LiteralExpression, is an interface over many node
// kinds: its kind is wide, and a brand field typed undefined, which no value carries, keeps other nodes
// from being assignable to it. Which kinds belong is a naming convention no run-time check can recover:
// a kind set built from the types assignable to it leaves out JsxText, which isLiteralExpression takes,
// and every late-bound declaration isLateBound takes (#c7w6j0y, measured on tsc Oct 31). So a checked
// predicate claiming one checks at the call only that the value is an object, and checks its shape where
// it's relied on: S's view checks every field read through S. S's view is the only one (road 2, ruled by
// Lower, Compiler and Adamic on Nov 1). A value it admits can still be seen as another type with no cast,
// a field-less restatement (`interface Restated extends TypeElement {}`), an alias or an inline type of
// S's shape, or any wider type, and a read through that type is #80cs7rz's: checkViewWidenedReads makes
// it NotYet wherever the value may come from S's entry, so a lie never reaches it raw (view_widened.go).
// Narrowing never gives a lie a subtype's type: brandedNarrowing fails closed.

// brandedSupertype is whether target is the shape that rule is keyed on, judged by type, never by name:
// a kind property that isn't one literal, and at least one brand member, a required member that isn't
// symbol-keyed, phantom as phantomField judges a primitive brand's (void or undefined), so the two rules
// can't drift when a brand is retyped: tsc's 36 are undefined since #ptd2v1r. A member typed any is never
// a brand, since #bw3xg7c refuses any (Adamic's correction, #tmdhz1d).
func (l *lowering) brandedSupertype(target *checker.Type) bool {
	kind := l.checker.GetPropertyOfType(target, "kind")
	if kind == nil || l.fieldLiteral(target, "kind") != nil {
		return false
	}
	for _, property := range l.checker.GetPropertiesOfType(target) {
		if property.Flags&ast.SymbolFlagsOptional != 0 || checker.IsKnownSymbol(property) {
			continue
		}
		if phantomField(l.checker.GetTypeOfSymbol(property)) {
			return true
		}
	}
	return false
}

// brandedViewKey is a claim's source and its branded supertype, as brandedViews keys them. It's a named
// struct, not a [2]*checker.Type, so the latent census's state copy (stage3/census/latent/statecopy)
// copies it field by field, as it copies nodeHostViewKey.
type brandedViewKey struct{ source, target *checker.Type }

// brandedSupertypeViews registers S's view once for each source the claim is made from, keeping the
// error that stopped it: what a view vouches for through a wider type is what its source already held
// (view_widened.go), so a source is part of the registration.
func (l *lowering) brandedSupertypeViews(node *ast.Node, source, target *checker.Type) error {
	proofs := l.proofs()
	key := brandedViewKey{source, target}
	if err, done := proofs.brandedViews[key]; done {
		return predicateFailureCopy(err)
	}
	err := l.registerBrandedSupertypeViews(node, source, target)
	proofs.brandedViews[key] = err
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

func (l *lowering) registerBrandedSupertypeViews(node *ast.Node, source, target *checker.Type) error {
	if classes := l.brandedClasses(target); len(classes) != 0 {
		return l.notYet(node, "a checked predicate target "+l.checker.TypeToString(target)+" that class "+l.checker.TypeToString(classes[0])+" is assignable to")
	}
	if err := l.brandedCalls(target); err != nil {
		return err
	}
	// The claim's source is what the view vouches for through a wider type, and the value enters the view
	// wherever a read is narrowed to S, as a tag-proven predicate's view does (view_widened.go).
	if _, err := l.view(node, nil, source, target); err != nil {
		return err
	}
	l.viewSources[len(l.viewSources)-1].narrows = true
	return nil
}

// brandedClasses is every class the program declares whose instances are assignable to target, found once.
func (l *lowering) brandedClasses(target *checker.Type) []*checker.Type {
	proofs := l.proofs()
	if classes, found := proofs.brandedClasses[target]; found {
		return classes
	}
	classes := []*checker.Type{}
	for _, class := range l.declaredClasses() {
		if l.checker.IsTypeAssignableTo(class, target) {
			classes = append(classes, class)
		}
	}
	proofs.brandedClasses[target] = classes
	return classes
}

// brandedCalls judges, once per branded supertype S, every call in the program whose resolved signature
// claims S of an argument, by brandedNarrowing. The checker narrows a call's argument by the claim however
// the call reaches the predicate: directly, through the predicate held as a value or a function type, or
// through a method's claim (Language's must on 9e4f2184). Those paths check the claim in different places,
// and the forwarder's and a closure's check holds no call site at all, so the calls are found here, where
// S's view is registered, which every path reaches through predicateMembership. A claim on this, or one
// whose argument is spread or missing, is NotYet, since no argument node says what the checker narrowed.
func (l *lowering) brandedCalls(target *checker.Type) error {
	proofs := l.proofs()
	if err, done := proofs.brandedCalls[target]; done {
		return predicateFailureCopy(err)
	}
	// The walk is done once per S, so it reads every call as written, with no instance's mapper: which
	// instance happens to be lowering when S's view registers says nothing about the calls in other
	// bodies. A claim naming a type parameter is judged for every instance at once (brandedCall).
	mapper := l.typeMapper
	l.typeMapper = nil
	defer func() { l.typeMapper = mapper }()
	var err error
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		if err != nil {
			return true
		}
		if n.Kind == ast.KindCallExpression {
			err = l.brandedCall(n, target)
		}
		if err == nil {
			n.ForEachChild(visit)
		}
		return err != nil
	}
	for _, file := range l.programFiles() {
		if file.AsNode().ForEachChild(visit) {
			break
		}
	}
	proofs.brandedCalls[target] = err
	return predicateFailureCopy(err)
}

// brandedCall judges one call by brandedNarrowing when its resolved signature claims a type holding target.
func (l *lowering) brandedCall(call *ast.Node, target *checker.Type) error {
	signature := l.checker.GetResolvedSignature(call)
	if signature == nil {
		return nil
	}
	claim := predicateOfSignature(l.checker, signature)
	if claim == nil || claim.Type() == nil {
		return nil
	}
	generic := l.mentionsTypeParameter(claim.Type(), map[*checker.Type]bool{})
	if generic && !l.brandedInstantiable(claim.Type(), target) || !generic && !l.brandedHolds(claim.Type(), target, map[*checker.Type]bool{}) {
		return nil
	}
	if claim.Kind() != checker.TypePredicateKindIdentifier && claim.Kind() != checker.TypePredicateKindAssertsIdentifier {
		return l.notYet(call, "a checked predicate target "+l.checker.TypeToString(target)+" claimed of this")
	}
	index := int(claim.ParameterIndex())
	arguments := call.AsCallExpression().Arguments.Nodes
	if index < 0 || index >= len(arguments) {
		return l.notYet(call, "a checked predicate target "+l.checker.TypeToString(target)+" claimed of a missing argument")
	}
	for _, argument := range arguments[:index+1] {
		if argument.Kind == ast.KindSpreadElement {
			return l.notYet(call, "a checked predicate target "+l.checker.TypeToString(target)+" claimed of a spread argument")
		}
	}
	if generic {
		// An argument of the claim's own type narrows to itself in every instance (brandedNarrowing). Any
		// other narrowing by a claim S may instantiate depends on the instance, which this walk can't see.
		if l.checker.GetTypeAtLocation(arguments[index]).Id() == claim.Type().Id() {
			return nil
		}
		return l.notYet(call, "a checked predicate target "+l.checker.TypeToString(target)+" that the generic claim "+l.checker.TypeToString(claim.Type())+" may be instantiated at")
	}
	return l.brandedNarrowing(call, arguments[index], claim.Type())
}

// brandedInstantiable is whether claim could be instantiated at a type holding target. A bare type
// parameter, or a union of them, can be target only when its base constraint admits it, or it has none.
// A type parameter anywhere else, a reference's or an alias's type argument or an intersection's member,
// counts as instantiable whatever its constraint: `Tagged<U>` with `U extends number` is the generic
// branded supertype Tagged<number> at U = number, which S's own constraint test would never see
// (Compiler's must on 810ce7d9).
func (l *lowering) brandedInstantiable(claim, target *checker.Type) bool {
	if claim.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range claim.Types() {
			if l.brandedInstantiable(member, target) {
				return true
			}
		}
		return false
	}
	if claim.Flags()&checker.TypeFlagsTypeParameter != 0 {
		constraint := l.checker.GetBaseConstraintOfType(claim)
		return constraint == nil || l.checker.IsTypeAssignableTo(target, constraint)
	}
	return l.mentionsTypeParameter(claim, map[*checker.Type]bool{})
}

// brandedHolds is whether claim is target, or holds it through a union's members or an array's elements,
// as brandedMembership walks them, read as written. visited ends a recursive alias on its second visit.
func (l *lowering) brandedHolds(claim, target *checker.Type, visited map[*checker.Type]bool) bool {
	if claim.Id() == target.Id() {
		return true
	}
	if visited[claim] {
		return false
	}
	visited[claim] = true
	if claim.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range claim.Types() {
			if l.brandedHolds(member, target, visited) {
				return true
			}
		}
		return false
	}
	return l.checker.IsArrayType(claim) && l.brandedHolds(l.checker.GetElementTypeOfArrayType(claim), target, visited)
}

// brandedNarrowing keeps a call NotYet unless the checker's narrowing of its argument, on the true
// side, can only give S (Compiler's musts on 9abcbdca and 8c7afca2). Every member of the argument's type
// must be a supertype of S, which narrows to S. Any other member is NotYet, never modelled: a subtype of
// S keeps its own type, whose fields S's view doesn't check and whose value no entry of S's names, and
// the checker drops an unrelated member only beside one that's its subtype, and makes an intersection no
// view is registered for otherwise, so judging either by assignability would admit a lie read raw. A
// branded supertype inside an array target is NotYet, since its elements narrow by the same rule with no
// call to judge at.
func (l *lowering) brandedNarrowing(node, argument *ast.Node, claim *checker.Type) error {
	// An argument already of the claim's own type narrows to itself, as tsc's `some(array)` claiming
	// `readonly T[]` of a `readonly T[]` does, whatever T is instantiated at: the result adds no member or
	// layout the argument didn't give, so there's nothing for a view to check, and it lowers as main does
	// (Adamic's ruling on 9e4f2184's census, core.ts:1203).
	if l.checker.GetTypeAtLocation(argument).Id() == claim.Id() {
		return nil
	}
	claim = l.concrete(claim)
	targets := []*checker.Type{claim}
	if claim.Flags()&checker.TypeFlagsUnion != 0 {
		targets = claim.Types()
	}
	// The checker narrows the argument's own type, so a type parameter T narrows to T & S, whose
	// instance is an intersection with an id of its own that no view covers (Language's must on
	// 7f4b7030). Judging concrete(T) would see a supertype and admit it. Any type the instance changes
	// mentions a type parameter, a generic object like Holder<K> or Readonly<T> included (Language's
	// must on 4a1fe092), so that's the test, not the type's own flags. A body lowered with no mapper
	// leaves the type as it is, so mentionsTypeParameter names the parameter there too.
	if declared := l.checker.GetTypeAtLocation(argument); l.concrete(declared) != declared || l.mentionsTypeParameter(declared, map[*checker.Type]bool{}) {
		return l.notYet(node, "a checked predicate target "+l.checker.TypeToString(claim)+" over a generic argument "+l.checker.TypeToString(declared)+", which narrows to an intersection no view covers")
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
		for _, member := range members {
			if !l.checker.IsTypeAssignableTo(target, member) {
				return l.notYet(node, "a checked predicate target "+l.checker.TypeToString(target)+" that narrows "+l.checker.TypeToString(member)+", which isn't its supertype")
			}
		}
	}
	return nil
}

// mentionsTypeParameter is whether t is instantiable, or names an instantiable type through a union's or
// intersection's members, its alias's type arguments, or a reference's type arguments, as Readonly<T>
// and Holder<K> do. visited ends a recursive type on its second visit.
func (l *lowering) mentionsTypeParameter(t *checker.Type, visited map[*checker.Type]bool) bool {
	if visited[t] {
		return false
	}
	visited[t] = true
	if t.Flags()&checker.TypeFlagsInstantiable != 0 {
		return true
	}
	var parts []*checker.Type
	if t.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		parts = append(parts, t.Types()...)
	}
	if alias := t.Alias(); alias != nil {
		parts = append(parts, alias.TypeArguments()...)
	}
	if t.Flags()&checker.TypeFlagsObject != 0 && t.ObjectFlags()&checker.ObjectFlagsReference != 0 {
		parts = append(parts, l.checker.GetTypeArguments(t)...)
	}
	for _, part := range parts {
		if l.mentionsTypeParameter(part, visited) {
			return true
		}
	}
	return false
}

// declaredClasses is every class the program's files declare or write as an expression, anywhere, a
// function's local ones included, as its instance type, found once. A class expression has no pin: on
// main every one panics in the namespace walks (namespaces.go:271, namespaces_call_graph.go:84) before
// this runs, which is Lower's #nkxtfkq, so it's here to fail closed once that's fixed.
func (l *lowering) declaredClasses() []*checker.Type {
	proofs := l.proofs()
	if proofs.classes != nil {
		return proofs.classes
	}
	proofs.classes = []*checker.Type{}
	seen := map[*checker.Type]bool{}
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		if n.Kind == ast.KindClassDeclaration || n.Kind == ast.KindClassExpression {
			if symbol := l.checker.GetSymbolOfDeclaration(n); symbol != nil {
				if declared := l.checker.GetDeclaredTypeOfSymbol(symbol); declared != nil && !seen[declared] {
					seen[declared] = true
					proofs.classes = append(proofs.classes, declared)
				}
			}
		}
		n.ForEachChild(visit)
		return false
	}
	for _, file := range l.programFiles() {
		file.AsNode().ForEachChild(visit)
	}
	return proofs.classes
}

// programFiles is the program's files, its roots and every file they import or export from, found once.
// It follows type-only imports too, which fails closed: a class declared only there makes no instance.
// The library and the prelude declare no class a branded supertype's view would read.
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
