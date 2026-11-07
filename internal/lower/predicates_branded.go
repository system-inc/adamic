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
func (lowering *lowering) brandedSupertype(target *checker.Type) bool {
	kind := lowering.checker.GetPropertyOfType(target, "kind")
	if kind == nil || lowering.fieldLiteral(target, "kind") != nil {
		return false
	}
	for _, property := range lowering.checker.GetPropertiesOfType(target) {
		if property.Flags&ast.SymbolFlagsOptional != 0 || checker.IsKnownSymbol(property) {
			continue
		}
		if phantomField(lowering.checker.GetTypeOfSymbol(property)) {
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
func (lowering *lowering) brandedSupertypeViews(node *ast.Node, source, target *checker.Type) error {
	proofs := lowering.proofs()
	key := brandedViewKey{source, target}
	if err, done := proofs.brandedViews[key]; done {
		return predicateFailureCopy(err)
	}
	registerBrandedSupertypeViewsError := lowering.registerBrandedSupertypeViews(node, source, target)
	proofs.brandedViews[key] = registerBrandedSupertypeViewsError
	return predicateFailureCopy(registerBrandedSupertypeViewsError)
}

// brandedMembership is whether target's membership check, as predicateMembership builds it, holds a
// branded supertype's: through a union's members or an array's elements, it can't be negated. visited
// ends a recursive alias, such as `type Nodes = Identifier | Nodes[]`, on its second visit.
func (lowering *lowering) brandedMembership(target *checker.Type, visited map[*checker.Type]bool) bool {
	target = lowering.concrete(target)
	if visited[target] {
		return false
	}
	visited[target] = true
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range target.Types() {
			if lowering.brandedMembership(member, visited) {
				return true
			}
		}
		return false
	}
	if lowering.isArrayType(target) {
		return lowering.brandedMembership(lowering.elementTypeOfArray(target), visited)
	}
	return target.Flags()&checker.TypeFlagsObject != 0 &&
		!isClassInstance(target) &&
		lowering.fieldLiteral(target, "kind") == nil &&
		lowering.brandedSupertype(target)
}

func (lowering *lowering) registerBrandedSupertypeViews(node *ast.Node, source, target *checker.Type) error {
	if classes := lowering.brandedClasses(target); len(classes) != 0 {
		return lowering.notYet(node, "a checked predicate target "+
			lowering.checker.TypeToString(target)+
			" that class "+
			lowering.checker.TypeToString(classes[0])+
			" is assignable to")
	}
	// The claim's source is what the view vouches for through a wider type, and the value enters the view
	// wherever a read is narrowed to S, as a tag-proven predicate's view does (view_widened.go).
	if _, viewError := lowering.view(node, nil, source, target); viewError != nil {
		return viewError
	}
	lowering.viewSources[len(lowering.viewSources)-1].narrows = true
	return nil
}

// brandedClasses is every class the program declares whose instances are assignable to target, found once.
func (lowering *lowering) brandedClasses(target *checker.Type) []*checker.Type {
	proofs := lowering.proofs()
	if classes, found := proofs.brandedClasses[target]; found {
		return classes
	}
	classes := []*checker.Type{}
	for _, class := range lowering.declaredClasses() {
		if lowering.checker.IsTypeAssignableTo(class, target) {
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
func (lowering *lowering) judgeBrandedCalls(scope *ast.Node) error {
	var found error
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if found != nil {
			return true
		}
		if node != scope && (node.Kind == ast.KindFunctionDeclaration ||
			node.Kind == ast.KindClassDeclaration ||
			node.Kind == ast.KindClassExpression) {
			return false
		}
		if node.Kind == ast.KindCallExpression {
			if brandedCallError := lowering.brandedCall(node); brandedCallError != nil {
				found = brandedCallError
				return true
			}
		}
		return node.ForEachChild(visit)
	}
	visit(scope)
	return found
}

// brandedRequest judges an instance request at its door, hit or miss, before the instance table is read:
// declaration's body under mapper, the instance's own, keyed by the instance's key. Two requests with one
// key are one instance, whose body is lowered once, so judging the request is judging that body, and a
// body shared by requests that differ, an unread parameter's, is judged with that parameter unmapped.
// TestBrandedRequestHoldsEveryInstanceDoor holds every write to the instance tables to a call of this.
func (lowering *lowering) brandedRequest(key string, declaration *ast.Node, mapper *typeMapper) error {
	proofs := lowering.proofs()
	if verdict, judged := proofs.brandedRequests[key]; judged {
		return predicateFailureCopy(verdict)
	}
	outer := lowering.typeMapper
	lowering.typeMapper = mapper
	verdict := lowering.judgeBrandedCalls(declaration)
	lowering.typeMapper = outer
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
func (lowering *lowering) brandedCall(call *ast.Node) error {
	signature := lowering.checker.GetResolvedSignature(call)
	if signature == nil {
		return nil
	}
	claim := predicateOfSignature(lowering.checker, signature)
	if claim == nil || claim.Type() == nil {
		return nil
	}
	if parameter := lowering.unknownInstance(lowering.concrete(claim.Type())); parameter != nil {
		return lowering.notYet(call, "a checked predicate claim of "+
			lowering.checker.TypeToString(lowering.concrete(claim.Type()))+
			", over "+
			lowering.checker.TypeToString(parameter)+
			", which this body is lowered without knowing")
	}
	if !lowering.brandedMembership(claim.Type(), map[*checker.Type]bool{}) {
		return nil
	}
	target := lowering.checker.TypeToString(lowering.concrete(claim.Type()))
	if claim.Kind() != checker.TypePredicateKindIdentifier &&
		claim.Kind() != checker.TypePredicateKindAssertsIdentifier {
		return lowering.notYet(call, "a checked predicate target "+target+" claimed of this")
	}
	index := int(claim.ParameterIndex())
	arguments := call.AsCallExpression().Arguments.Nodes
	if index < 0 || index >= len(arguments) {
		return lowering.notYet(call, "a checked predicate target "+target+" claimed of a missing argument")
	}
	for _, argument := range arguments[:index+1] {
		if argument.Kind == ast.KindSpreadElement {
			return lowering.notYet(call, "a checked predicate target "+target+" claimed of a spread argument")
		}
	}
	return lowering.brandedNarrowing(call, arguments[index], claim.Type())
}

// unknownInstance is a type parameter claim still holds once concreted, which the body it's judged in
// doesn't map, or nil. A claim that is a bare type parameter, or a union of them, whose base constraint
// holds only primitives can never be a branded supertype, an object, so it isn't one. A type parameter
// anywhere else, inside a reference or an alias, counts whatever its constraint, since Tagged<U> with
// U extends number is S when S is Tagged<number> (Compiler's must on 810ce7d9).
func (lowering *lowering) unknownInstance(claim *checker.Type) *checker.Type {
	if !lowering.mentionsTypeParameter(claim, map[*checker.Type]bool{}) {
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
		constraint := lowering.checker.GetBaseConstraintOfType(member)
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
func (lowering *lowering) brandedNarrowedRead(call, argument *ast.Node) *checker.Type {
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
		case parent.Kind == ast.KindPrefixUnaryExpression &&
			parent.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken:
			negated = !negated
		case parent.Kind == ast.KindBinaryExpression &&
			parent.AsBinaryExpression().OperatorToken.Kind == ast.KindAmpersandAmpersandToken &&
			!negated:
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
		case parent.Kind == ast.KindExpressionStatement &&
			!negated &&
			parent.Parent != nil &&
			parent.Parent.Kind == ast.KindBlock:
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
	visit = func(node *ast.Node) bool {
		if found != nil {
			return true
		}
		if (node.Kind == ast.KindIdentifier ||
			node.Kind == ast.KindPropertyAccessExpression ||
			node.Kind == ast.KindThisKeyword) && predicateMatchingReference(lowering.checker, argument, node) {
			found = lowering.checker.GetTypeAtLocation(node)
			return true
		}
		return node.ForEachChild(visit)
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
func (lowering *lowering) brandedNarrowing(node, argument *ast.Node, claim *checker.Type) error {
	// An argument already of the claim's own type narrows to itself, as tsc's `some(array)` claiming
	// `readonly T[]` of a `readonly T[]` does, whatever T is instantiated at: the result adds no member or
	// layout the argument didn't give, so there's nothing for a view to check, and it lowers as main does
	// (Adamic's ruling on 9e4f2184's census, core.ts:1203).
	if lowering.checker.GetTypeAtLocation(argument).Id() == claim.Id() {
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
	// the concreted read is that question). A read of S intersected only with S's supertypes, as tsc's
	// tryCast and cast read their value, stands too (brandedIntersection). An instance that's a union or an
	// array narrows its members or elements by rules this read doesn't see, and no read found says nothing,
	// so both are NotYet.
	if declared := lowering.checker.GetTypeAtLocation(
		argument,
	); lowering.concrete(declared) != declared || lowering.mentionsTypeParameter(declared, map[*checker.Type]bool{}) {
		instance := lowering.concrete(claim)
		narrowed := lowering.brandedNarrowedRead(node, argument)
		if narrowed == nil {
			return lowering.notYet(node, "a checked predicate target "+
				lowering.checker.TypeToString(instance)+
				" over a generic argument "+
				lowering.checker.TypeToString(declared)+
				", with no read of it on the true side")
		}
		read := lowering.concrete(narrowed)
		if (read.Id() == instance.Id() || lowering.brandedIntersection(read, instance)) &&
			instance.Flags()&checker.TypeFlagsUnion == 0 &&
			!lowering.isArrayType(instance) {
			return nil
		}
		return lowering.notYet(node, "a checked predicate target "+
			lowering.checker.TypeToString(instance)+
			" over a generic argument "+
			lowering.checker.TypeToString(declared)+
			", read on the true side as "+
			lowering.checker.TypeToString(read))
	}
	claim = lowering.concrete(claim)
	source := lowering.concrete(lowering.checker.GetTypeAtLocation(argument))
	members := []*checker.Type{source}
	if source.Flags()&checker.TypeFlagsUnion != 0 {
		members = source.Types()
	}
	for _, target := range targets {
		if !lowering.brandedMembership(target, map[*checker.Type]bool{}) {
			continue
		}
		if lowering.isArrayType(target) || target.Flags()&checker.TypeFlagsUnion != 0 {
			return lowering.notYet(node, "a branded supertype inside a checked predicate's array target")
		}
		for _, member := range members {
			if !lowering.checker.IsTypeAssignableTo(target, member) {
				return lowering.notYet(node, "a checked predicate target "+
					lowering.checker.TypeToString(target)+
					" that narrows "+
					lowering.checker.TypeToString(member)+
					", which isn't its supertype")
			}
		}
	}
	return nil
}

// brandedIntersection is whether read, a read's concreted type, is target intersected only with
// supertypes of target, as tsc's tryCast and cast read their value as TIn & TOut, which concretes to
// Node & TypeElement (Adamic's ruling, Nov 1). Each field of such an intersection is target's own field
// type met with a wider one, which is target's, so a read through it is a read through target's checked
// view, which checks the value is target where it reads. A member is a supertype when target is assignable
// to it by the checker's strict assignability, never comparability, which takes an optional field for a
// required one: target's view would check that field as maybe absent while the read types it as present.
// A member typed any or unknown adds nothing. A member that's still a type parameter, one the body is
// lowered without mapping, or anything else, is no supertype, so the read stays NotYet. And a member
// must hold only target's fields (brandedSupertypeMember), or the premise fails.
func (lowering *lowering) brandedIntersection(read, target *checker.Type) bool {
	if read.Flags()&checker.TypeFlagsIntersection == 0 {
		return false
	}
	found := false
	for _, member := range read.Types() {
		switch {
		case member.Id() == target.Id():
			found = true
		case member.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0:
		case lowering.mentionsTypeParameter(member, map[*checker.Type]bool{}) ||
			!lowering.brandedSupertypeMember(member, target):
			return false
		}
	}
	return found
}

// brandedSupertypeMember is whether member is a supertype of target the ruling's argument holds for:
// target strictly assignable to it, and every property it has, optional and inherited ones included, a
// property of target by name whose type in target is strictly assignable to member's. Assignability alone
// isn't that: a member's optional field target lacks, as Lettered's `a?: string` over TypeElement, keeps
// target assignable while its read, routed to target's id, is a field no view of target checks, so it was
// read raw, unchecked, where 021e43f3 lowered it (Lower's must, ruled (a) by Compiler, Lower and
// Language, Nov 2). A member with an index signature, or a call or construct signature, never counts:
// neither is a property, and either could route a read through target's id raw (Compiler, Nov 2).
func (lowering *lowering) brandedSupertypeMember(member, target *checker.Type) bool {
	if !lowering.checker.IsTypeAssignableTo(target, member) ||
		len(lowering.checker.GetIndexInfosOfType(member)) != 0 ||
		len(lowering.checker.GetSignaturesOfType(member, checker.SignatureKindCall)) != 0 ||
		len(lowering.checker.GetSignaturesOfType(member, checker.SignatureKindConstruct)) != 0 {
		return false
	}
	for _, property := range lowering.checker.GetPropertiesOfType(member) {
		own := lowering.checker.GetPropertyOfType(target, property.Name)
		if own == nil ||
			!lowering.checker.IsTypeAssignableTo(
				lowering.checker.GetTypeOfSymbol(own),
				lowering.checker.GetTypeOfSymbol(property),
			) {
			return false
		}
	}
	return true
}

// brandedIntersectionReceiver is the branded supertype S a receiver of S intersected only with supertypes
// of S is read through, or receiver as it is. A checked view's fields are keyed by its target's id
// (CheckedFieldKey), and the intersection has an id of its own, so without this a read through it is no
// read of S's view: g1's n.name at T = Node, admitted with no routing, was a widened read, NotYet as one
// that may hold a checked view's value (view_widened.go, probed on Golem, Nov 1). brandedIntersection is
// what makes it S's (Adamic's ruling, Nov 1). Only reads take it (viewReadReceiverTypeID); a write keeps
// the intersection's own id.
func (lowering *lowering) brandedIntersectionReceiver(receiver *checker.Type) *checker.Type {
	if receiver.Flags()&checker.TypeFlagsIntersection == 0 {
		return receiver
	}
	for _, member := range receiver.Types() {
		if member.Flags()&checker.TypeFlagsObject != 0 &&
			!isClassInstance(member) &&
			lowering.fieldLiteral(member, "kind") == nil &&
			lowering.brandedSupertype(member) &&
			lowering.brandedIntersection(receiver, member) {
			return member
		}
	}
	return receiver
}

// mentionsTypeParameter is whether `subjectType` is instantiable, or names an instantiable type through a union's or
// intersection's members, its alias's type arguments, or a reference's type arguments, as Readonly<T>
// and Holder<K> do. visited ends a recursive type on its second visit.
func (lowering *lowering) mentionsTypeParameter(subjectType *checker.Type, visited map[*checker.Type]bool) bool {
	if visited[subjectType] {
		return false
	}
	visited[subjectType] = true
	if subjectType.Flags()&checker.TypeFlagsInstantiable != 0 {
		return true
	}
	var parts []*checker.Type
	if subjectType.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		parts = append(parts, subjectType.Types()...)
	}
	if alias := subjectType.Alias(); alias != nil {
		parts = append(parts, alias.TypeArguments()...)
	}
	if subjectType.Flags()&checker.TypeFlagsObject != 0 && subjectType.ObjectFlags()&checker.ObjectFlagsReference != 0 {
		parts = append(parts, lowering.checker.GetTypeArguments(subjectType)...)
	}
	for _, part := range parts {
		if lowering.mentionsTypeParameter(part, visited) {
			return true
		}
	}
	return false
}

// declaredClasses is every class the program's files declare or write as an expression, anywhere, a
// function's local ones included, as its instance type, found once. A class expression has no pin: on
// main every one panics in the namespace walks (namespaces.go:271, namespaces_call_graph.go:84) before
// this runs, which is Lower's #nkxtfkq, so it's here to fail closed once that's fixed.
func (lowering *lowering) declaredClasses() []*checker.Type {
	proofs := lowering.proofs()
	if proofs.classes != nil {
		return proofs.classes
	}
	proofs.classes = []*checker.Type{}
	seen := map[*checker.Type]bool{}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindClassDeclaration || node.Kind == ast.KindClassExpression {
			if symbol := lowering.checker.GetSymbolOfDeclaration(node); symbol != nil {
				if declared := lowering.checker.GetDeclaredTypeOfSymbol(symbol); declared != nil && !seen[declared] {
					seen[declared] = true
					proofs.classes = append(proofs.classes, declared)
				}
			}
		}
		node.ForEachChild(visit)
		return false
	}
	for _, file := range lowering.programFiles() {
		file.AsNode().ForEachChild(visit)
	}
	return proofs.classes
}

// programFiles is the program's files, its roots and every file they import or export from, found once.
// It follows type-only imports too, which fails closed: a class declared only there makes no instance.
// The library and the prelude declare no class a branded supertype's view would read.
func (lowering *lowering) programFiles() []*ast.SourceFile {
	proofs := lowering.proofs()
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
			if target := lowering.checker.GetSymbolAtLocation(
				specifier,
			); target != nil && len(target.Declarations) > 0 && target.Declarations[0].Kind == ast.KindSourceFile {
				visit(target.Declarations[0].AsSourceFile())
			}
		}
	}
	for _, file := range lowering.program.Files() {
		visit(file)
	}
	return proofs.files
}
