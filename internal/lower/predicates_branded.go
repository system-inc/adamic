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

// judgeBrandedCalls judges every call in scope by brandedCall, under the type mapper in force, which is
// the mapper scope's body is lowered under (Adamic's ruling on #c7w6j0y, Nov 1: every call in a function
// body is judged once, when that body is lowered, under that body's own mapper, by a walk over the body).
// A walk finds every CallExpression however it lowers, an assertion statement included, where a hook in
// one lowering path misses the others. It passes over a function declaration and a class, which are
// lowered, and judged, where they're first used or instantiated, so a unit never stops at a call it
// doesn't reach. It enters every other function written in scope, a closure, an object's method, or a
// generic arrow, method or function expression: those are lowered with scope, and a generic one's own
// type parameters have no instance door, so they stay unmapped and its claims over them are NotYet
// (Lower, Nov 1).
func (l *lowering) judgeBrandedCalls(scope *ast.Node) error {
	var found error
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		if found != nil {
			return true
		}
		if n != scope && (n.Kind == ast.KindFunctionDeclaration || n.Kind == ast.KindClassDeclaration || n.Kind == ast.KindClassExpression) {
			return false
		}
		if n.Kind == ast.KindCallExpression {
			if err := l.brandedCall(n); err != nil {
				found = err
				return true
			}
		}
		return n.ForEachChild(visit)
	}
	visit(scope)
	return found
}

// brandedRequest judges an instance request at its door, hit or miss, before the instance table is read:
// declaration's body under mapper, the instance's own, keyed by the instance's key. Two requests with one
// key are one instance, whose body is lowered once, so judging the request is judging that body, and a
// body shared by requests that differ, an unread parameter's, is judged with that parameter unmapped.
// TestBrandedRequestHoldsEveryInstanceDoor holds every write to the instance tables to a call of this.
func (l *lowering) brandedRequest(key string, declaration *ast.Node, mapper *typeMapper) error {
	proofs := l.proofs()
	if verdict, judged := proofs.brandedRequests[key]; judged {
		return predicateFailureCopy(verdict)
	}
	outer := l.typeMapper
	l.typeMapper = mapper
	verdict := l.judgeBrandedCalls(declaration)
	l.typeMapper = outer
	proofs.brandedRequests[key] = verdict
	return predicateFailureCopy(verdict)
}

// brandedCall judges a call whose resolved signature claims a type holding a branded supertype S of an
// argument: brandedNarrowing decides whether the checker's narrowing of the argument can only give S. The
// checker narrows a call's argument however the call reaches the predicate, directly, through the
// predicate held as a value or a function type, or through a method's claim (Language's must on
// 9e4f2184), and judgeBrandedCalls finds every one of those calls. A claim on this, or one whose argument
// is spread or missing, is NotYet, since no argument node says what the checker narrowed. So is a claim
// over a type parameter the body is lowered without knowing, an unread one or a generic arrow's own: the
// instance a call claims is then unknown, and every instance shares the one body (Adamic's ruling, Nov 1).
func (l *lowering) brandedCall(call *ast.Node) error {
	signature := l.checker.GetResolvedSignature(call)
	if signature == nil {
		return nil
	}
	claim := predicateOfSignature(l.checker, signature)
	if claim == nil || claim.Type() == nil {
		return nil
	}
	if parameter := l.unknownInstance(l.concrete(claim.Type())); parameter != nil {
		return l.notYet(call, "a checked predicate claim of "+l.checker.TypeToString(l.concrete(claim.Type()))+", over "+l.checker.TypeToString(parameter)+", which this body is lowered without knowing")
	}
	if !l.brandedMembership(claim.Type(), map[*checker.Type]bool{}) {
		return nil
	}
	target := l.checker.TypeToString(l.concrete(claim.Type()))
	if claim.Kind() != checker.TypePredicateKindIdentifier && claim.Kind() != checker.TypePredicateKindAssertsIdentifier {
		return l.notYet(call, "a checked predicate target "+target+" claimed of this")
	}
	index := int(claim.ParameterIndex())
	arguments := call.AsCallExpression().Arguments.Nodes
	if index < 0 || index >= len(arguments) {
		return l.notYet(call, "a checked predicate target "+target+" claimed of a missing argument")
	}
	for _, argument := range arguments[:index+1] {
		if argument.Kind == ast.KindSpreadElement {
			return l.notYet(call, "a checked predicate target "+target+" claimed of a spread argument")
		}
	}
	return l.brandedNarrowing(call, arguments[index], claim.Type())
}

// unknownInstance is a type parameter claim still holds once concreted, which the body it's judged in
// doesn't map, or nil. A claim that is a bare type parameter, or a union of them, whose base constraint
// holds only primitives can never be a branded supertype, an object, so it isn't one. A type parameter
// anywhere else, inside a reference or an alias, counts whatever its constraint, since Tagged<U> with
// U extends number is S when S is Tagged<number> (Compiler's must on 810ce7d9).
func (l *lowering) unknownInstance(claim *checker.Type) *checker.Type {
	if !l.mentionsTypeParameter(claim, map[*checker.Type]bool{}) {
		return nil
	}
	members := []*checker.Type{claim}
	if claim.Flags()&checker.TypeFlagsUnion != 0 {
		members = claim.Types()
	}
	for _, member := range members {
		if member.Flags()&checker.TypeFlagsTypeParameter == 0 {
			return claim
		}
		constraint := l.checker.GetBaseConstraintOfType(member)
		if constraint == nil || !predicatePrimitiveTarget(constraint) {
			return member
		}
	}
	return nil
}

// brandedNarrowedRead is the checker's flow type for argument at its first read where the call is known
// to have answered true: the true branch of an if or a conditional, the right of an &&, the statements
// after an assertion, through parentheses and a !. It's nil when that region isn't one of those, or holds
// no read of the argument, which brandedNarrowing takes as unknown: the narrowing can reach past the
// region, after an if whose else returns, or after the block an assertion ends, so a region with no read
// says nothing about the reads the narrowing gives.
func (l *lowering) brandedNarrowedRead(call, argument *ast.Node) *checker.Type {
	var regions []*ast.Node
	negated := false
	node := call
	for regions == nil {
		parent := node.Parent
		if parent == nil {
			return nil
		}
		switch {
		case parent.Kind == ast.KindParenthesizedExpression:
		case parent.Kind == ast.KindPrefixUnaryExpression && parent.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken:
			negated = !negated
		case parent.Kind == ast.KindBinaryExpression && parent.AsBinaryExpression().OperatorToken.Kind == ast.KindAmpersandAmpersandToken && !negated:
			if parent.AsBinaryExpression().Left == node {
				regions = []*ast.Node{parent.AsBinaryExpression().Right}
			}
		case parent.Kind == ast.KindConditionalExpression && parent.AsConditionalExpression().Condition == node:
			regions = []*ast.Node{parent.AsConditionalExpression().WhenTrue}
			if negated {
				regions = []*ast.Node{parent.AsConditionalExpression().WhenFalse}
			}
		case parent.Kind == ast.KindIfStatement && parent.AsIfStatement().Expression == node:
			region := parent.AsIfStatement().ThenStatement
			if negated {
				region = parent.AsIfStatement().ElseStatement
			}
			if region == nil {
				return nil
			}
			regions = []*ast.Node{region}
		case parent.Kind == ast.KindExpressionStatement && !negated && parent.Parent != nil && parent.Parent.Kind == ast.KindBlock:
			statements := parent.Parent.AsBlock().Statements.Nodes
			for position, statement := range statements {
				if statement == parent {
					regions = append([]*ast.Node{}, statements[position+1:]...)
				}
			}
			if regions == nil {
				return nil
			}
		default:
			return nil
		}
		node = parent
	}
	var found *checker.Type
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		if found != nil {
			return true
		}
		if (n.Kind == ast.KindIdentifier || n.Kind == ast.KindPropertyAccessExpression || n.Kind == ast.KindThisKeyword) && predicateMatchingReference(l.checker, argument, n) {
			found = l.checker.GetTypeAtLocation(n)
			return true
		}
		return n.ForEachChild(visit)
	}
	for _, region := range regions {
		if visit(region) {
			break
		}
	}
	return found
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
	targets := []*checker.Type{claim}
	if claim.Flags()&checker.TypeFlagsUnion != 0 {
		targets = claim.Types()
	}
	// The checker narrows the argument's own type, as written, so a type parameter T narrows to T & S,
	// whose instance is an intersection with an id of its own that no view covers (Language's must on
	// 7f4b7030), while tryCast's TIn narrows to its TOut, whose instance is S. Concreted, the two
	// arguments look alike, so an argument that mentions a type parameter, a generic object like Holder<K>
	// or Readonly<T> included (Language's must on 4a1fe092), is judged by what its reads see: the checker's
	// flow type at its first read on the true side, concreted under the body's mapper. The call stands only
	// when that's S itself, so every read goes through S's view, which checks each field it reads (Adamic's
	// ruling on 810ce7d9: the checker is asked, never modelled; Compiler's and Lower's reading, Nov 1, that
	// the concreted read is that question). An instance that's a union or an array narrows its members or
	// elements by rules this read doesn't see, and no read found says nothing, so both are NotYet.
	if declared := l.checker.GetTypeAtLocation(argument); l.concrete(declared) != declared || l.mentionsTypeParameter(declared, map[*checker.Type]bool{}) {
		instance := l.concrete(claim)
		if narrowed := l.brandedNarrowedRead(node, argument); narrowed != nil && l.concrete(narrowed).Id() == instance.Id() && instance.Flags()&checker.TypeFlagsUnion == 0 && !l.checker.IsArrayType(instance) {
			return nil
		}
		return l.notYet(node, "a checked predicate target "+l.checker.TypeToString(instance)+" over a generic argument "+l.checker.TypeToString(declared)+", whose reads on the true side aren't seen as "+l.checker.TypeToString(instance)+" alone")
	}
	claim = l.concrete(claim)
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
