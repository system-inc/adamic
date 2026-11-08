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
	if l.isArrayType(target) {
		return l.brandedMembership(l.elementTypeOfArray(target), visited)
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
// side, can't give a lie a type it's read raw through (Compiler's musts on 9abcbdca and 8c7afca2). The
// danger is a narrowing that drops a member of the argument's type while keeping a type no view checks: a
// value of the dropped kind, said to be S by a lying predicate, is then read as the kept type. Nothing is
// modelled: the checker is asked what the narrowing gives, and the call stands on one of these roads.
//
//   - Every member of the argument's type is a supertype of each branded arm, which narrows to that arm.
//     A subtype of S keeps its own type, whose fields S's view doesn't check, and the checker drops an
//     unrelated member only beside one that's its subtype, so this road judges the written type alone,
//     never assignability, and only for an argument that mentions no type parameter.
//   - Rule 1 (brandedChainCall): no true-side read exists, and none can.
//   - Rule 2, read off the checker's flow type at the first true-side read (brandedNarrowedRead): (b) the
//     read is the argument's own type, by the checker's identity, so the narrowing dropped nothing and a
//     lie changes no static type; or, member by member, the read concreted under the body's mapper is
//     (a) a branded arm S, or S intersected only with S's supertypes (brandedIntersection), read through
//     S's view, or (c) an unbranded arm with its own tag, read through that arm's view
//     (brandedTaggedArm). Adamic's ruling on #fx9waqy, Nov 7, with Compiler's and Language's conditions.
//
// No read found says nothing, since the narrowing can reach past the region, so it's NotYet outside rule
// 1. A branded supertype inside an array target is NotYet, since its elements narrow by the same rule with
// no call to judge at. Whatever no road admits keeps the words it had before this rule: the first member
// that isn't a supertype for a written argument, and the generic argument's read for one that mentions a
// type parameter.
func (l *lowering) brandedNarrowing(node, argument *ast.Node, claim *checker.Type) error {
	declared := l.checker.GetTypeAtLocation(argument)
	// An argument already of the claim's own type narrows to itself, as tsc's `some(array)` claiming
	// `readonly T[]` of a `readonly T[]` does, whatever T is instantiated at: the result adds no member or
	// layout the argument didn't give, so there's nothing for a view to check, and it lowers as main does
	// (Adamic's ruling on 9e4f2184's census, core.ts:1203).
	if declared.Id() == claim.Id() {
		return nil
	}
	instance := l.concrete(claim)
	arms := []*checker.Type{instance}
	if instance.Flags()&checker.TypeFlagsUnion != 0 {
		arms = instance.Types()
	}
	for _, arm := range arms {
		if l.isArrayType(arm) && l.brandedMembership(arm, map[*checker.Type]bool{}) {
			return l.notYet(node, "a branded supertype inside a checked predicate's array target")
		}
	}
	// The checker narrows the argument's own type, as written, so a type parameter T narrows to T & S,
	// whose instance is an intersection with an id of its own that no view covers (Language's must on
	// 7f4b7030), while tryCast's TIn narrows to its TOut, whose instance is S. Concreted, the two
	// arguments look alike, so an argument that mentions a type parameter, a generic object like Holder<K>
	// or Readonly<T> included (Language's must on 4a1fe092), is judged only by what its reads see (Adamic's
	// ruling on 810ce7d9: the checker is asked, never modelled; Compiler's and Lower's reading, Nov 1, that
	// the concreted read is that question).
	generic := l.concrete(declared) != declared || l.mentionsTypeParameter(declared, map[*checker.Type]bool{})
	var stop error
	if !generic {
		source := l.concrete(declared)
		members := []*checker.Type{source}
		if source.Flags()&checker.TypeFlagsUnion != 0 {
			members = source.Types()
		}
		for _, arm := range arms {
			if !l.brandedMembership(arm, map[*checker.Type]bool{}) {
				continue
			}
			for _, member := range members {
				if stop == nil && !l.checker.IsTypeAssignableTo(arm, member) {
					stop = l.notYet(node, "a checked predicate target "+l.checker.TypeToString(arm)+" that narrows "+l.checker.TypeToString(member)+", which isn't its supertype")
				}
			}
		}
		if stop == nil {
			return nil
		}
	}
	narrowed := l.brandedNarrowedRead(node, argument)
	if narrowed == nil {
		if l.brandedChainCall(node, argument) {
			return nil
		}
		if stop != nil {
			return stop
		}
		return l.notYet(node, "a checked predicate target "+l.checker.TypeToString(instance)+" over a generic argument "+l.checker.TypeToString(declared)+", with no read of it on the true side")
	}
	// (b) is the checker's own identity of two types, never a comparison of their members: a union written
	// in another order, or reduced another way, is equal by structure and can still have dropped a member
	// the checker sees, so only the same type says nothing was dropped (Adamic's and Language's condition).
	if narrowed.Id() == declared.Id() {
		return nil
	}
	read := l.concrete(narrowed)
	members := []*checker.Type{read}
	if read.Flags()&checker.TypeFlagsUnion != 0 {
		members = read.Types()
	}
	for _, member := range members {
		if l.brandedArmRead(member, arms) || l.brandedTaggedArm(member, arms) {
			continue
		}
		if stop != nil {
			return stop
		}
		return l.notYet(node, "a checked predicate target "+l.checker.TypeToString(instance)+" over a generic argument "+l.checker.TypeToString(declared)+", read on the true side as "+l.checker.TypeToString(read))
	}
	return nil
}

// brandedArmRead is rule 2(a): member, one member of a true-side read concreted, is a branded arm S of
// the claim itself, or S intersected only with S's supertypes, as tsc's tryCast and cast read their value
// (brandedIntersection). Either is read through S's view, which checks each field it reads.
func (l *lowering) brandedArmRead(member *checker.Type, arms []*checker.Type) bool {
	for _, arm := range arms {
		if !l.brandedMembership(arm, map[*checker.Type]bool{}) {
			continue
		}
		if member.Id() == arm.Id() || l.brandedIntersection(member, arm) {
			return true
		}
	}
	return false
}

// brandedTaggedArm is rule 2(c): member, one member of a true-side read concreted, is an arm of the claim
// with its own literal kind, the tag the call's check tests for that arm (predicateMembership), which
// registers that arm's view, so every read narrowed to it checks its layout. The membership check is a
// union of its arms' tests, so a value can pass by a branded arm's, which checks only that it's an object,
// and still be read as this arm: the view is what stops that lie at its first read, never the tag. The tag
// must still tell the arm apart by value, as a cast's must (rulings.md:13): from every other arm's tag,
// and from every other member of its own enum (tag_values.go, the one rule #fvmqee2's verifier shares),
// since `enum K { First = 1, Alias = 1 }` makes two types with one value.
func (l *lowering) brandedTaggedArm(member *checker.Type, arms []*checker.Type) bool {
	if member.Flags()&checker.TypeFlagsObject == 0 || isClassInstance(member) {
		return false
	}
	arm := false
	for _, candidate := range arms {
		arm = arm || candidate.Id() == member.Id()
	}
	tag := l.fieldLiteral(member, "kind")
	if !arm || tag == nil {
		return false
	}
	if _, _, known := l.literalConstant(tag); !known || !l.enumValueUnique(tag) {
		return false
	}
	for _, other := range arms {
		if other.Id() == member.Id() || other.Flags()&checker.TypeFlagsObject == 0 || isClassInstance(other) {
			continue
		}
		if otherTag := l.fieldLiteral(other, "kind"); otherTag != nil && !tagLiteralsDistinct(tag, otherTag) {
			return false
		}
	}
	return true
}

// brandedChainCall is rule 1 (Adamic's ruling on #fx9waqy, Nov 7, in Compiler's form): a call that
// stands with no true-side read because none can exist. core.ts's orPredicate2 is its case, `return p =>
// f1(p) || f2(p)`: neither call's narrowing reaches a read, and the chain's value is the arrow's result,
// whose claim is checked at every call through it (brandedResultChecked). It holds when all of these do.
//
//   - The call is an operand of a chain of `||` and `&&`, through parentheses, whose whole value is what a
//     function returns, its arrow body or a return statement's expression.
//   - The argument is a name bound by one of that function's parameters.
//   - The function's own result claim is a checked claim.
//   - Every other occurrence of the name in the function, walked into nested functions and classes, since
//     a closure made after the narrowing reads the parameter narrowed (TypeScript 5.4 and later), is an
//     argument of another call in the same chain whose checked or vouched claim names that argument, at a
//     parameter of exactly the binding's own type. So it's passed on, never read through a narrower type,
//     stored, captured or returned (brandedChainUse).
//
// The `&&` form, f1(p) && f2(p), never reaches here for f1: f2's argument is a read on f1's true side,
// judged by rule 2(a) as an upcast to f2's own parameter. Pins: Compiler's `p => f1(p) || (p as
// R2).x !== undefined` and Language's `p => f1(p) && (() => p.x)()`, both NotYet. Mutant: the use walk
// dropped, and both lower.
func (l *lowering) brandedChainCall(call, argument *ast.Node) bool {
	if argument.Kind != ast.KindIdentifier {
		return false
	}
	binding := l.symbol(argument)
	if binding == nil || binding.ValueDeclaration == nil || binding.ValueDeclaration.Kind != ast.KindParameter {
		return false
	}
	top := brandedChainTop(call)
	if top == call {
		return false
	}
	function := brandedChainReturner(top)
	if function == nil || binding.ValueDeclaration.Parent != function || !l.brandedResultChecked(function) {
		return false
	}
	name := binding.ValueDeclaration.Name()
	declared := l.checker.GetTypeOfSymbol(binding)
	sound := true
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		if !sound {
			return true
		}
		// Every identifier spelled as the binding counts, whatever it resolves to, so a shorthand `{ p }`,
		// whose name resolves to the property, can't carry the value out unseen. A same-named property or
		// a shadowing name fails closed. A type position reads nothing at run time, a claim's own `p is R`
		// among them, so it doesn't count.
		if n.Kind == ast.KindIdentifier && n != name && n.Text() == name.Text() && !ast.IsPartOfTypeNode(n) {
			sound = l.symbol(n) == binding && l.brandedChainUse(n, top, declared)
		}
		return n.ForEachChild(visit)
	}
	visit(function)
	return sound
}

// brandedChainUse is whether use, a reference to a rule-1 binding, is an argument of a call in the chain
// whose top is top, at the parameter its claim names, typed exactly as the binding is, with a claim that's
// checked at that call or vouched for, so the call passes the value on and reads nothing through a
// narrower type.
func (l *lowering) brandedChainUse(use, top *ast.Node, declared *checker.Type) bool {
	call := use.Parent
	if call == nil || call.Kind != ast.KindCallExpression || call.AsCallExpression().Expression == use || brandedChainTop(call) != top {
		return false
	}
	index := -1
	for position, argument := range call.AsCallExpression().Arguments.Nodes {
		if argument.Kind == ast.KindSpreadElement {
			return false
		}
		if argument == use {
			index = position
			break
		}
	}
	signature := l.checker.GetResolvedSignature(call)
	if index < 0 || signature == nil || signature.Declaration() == nil || index >= len(signature.Parameters()) {
		return false
	}
	claim := predicateOfSignature(l.checker, signature)
	if claim == nil || claim.Kind() != checker.TypePredicateKindIdentifier || int(claim.ParameterIndex()) != index {
		return false
	}
	if l.checker.GetTypeOfSymbol(signature.Parameters()[index]).Id() != declared.Id() {
		return false
	}
	return l.checkedFunctionTypeClaim(signature.Declaration().Type()) || l.predicateCallVouched(signature.Declaration(), claim)
}

// brandedChainTop is the outermost expression a chain of `||` and `&&` makes around node, through
// parentheses, or node itself when it's no operand of one.
func brandedChainTop(node *ast.Node) *ast.Node {
	for node.Parent != nil {
		parent := node.Parent
		if parent.Kind == ast.KindBinaryExpression {
			operator := parent.AsBinaryExpression().OperatorToken.Kind
			if operator != ast.KindBarBarToken && operator != ast.KindAmpersandAmpersandToken {
				break
			}
		} else if parent.Kind != ast.KindParenthesizedExpression {
			break
		}
		node = parent
	}
	return node
}

// brandedChainReturner is the function whose result value is, when value is an arrow's expression body or
// a return statement's expression, or nil.
func brandedChainReturner(value *ast.Node) *ast.Node {
	parent := value.Parent
	switch {
	case parent == nil:
		return nil
	case parent.Kind == ast.KindArrowFunction && parent.Body() == value:
		return parent
	case parent.Kind == ast.KindReturnStatement && parent.AsReturnStatement().Expression == value:
		return ast.GetContainingFunction(parent)
	}
	return nil
}

// brandedResultChecked is whether function's result claim is checked at every call that narrows by it:
// a checked closure checks its own claim at each return (checkedPredicateClosure), and a function its
// producer returns under a written checked claim, as orPredicate2's `(p: P) => p is R1 | R2`, leaves only
// through that type (predicateReturnedUnderWrittenClaim).
func (l *lowering) brandedResultChecked(function *ast.Node) bool {
	return l.checkedPredicateClosure(function) || l.predicateReturnedUnderWrittenClaim(function)
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
func (l *lowering) brandedIntersection(read, target *checker.Type) bool {
	if read.Flags()&checker.TypeFlagsIntersection == 0 {
		return false
	}
	found := false
	for _, member := range read.Types() {
		switch {
		case member.Id() == target.Id():
			found = true
		case member.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0:
		case l.mentionsTypeParameter(member, map[*checker.Type]bool{}) || !l.brandedSupertypeMember(member, target):
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
func (l *lowering) brandedSupertypeMember(member, target *checker.Type) bool {
	if !l.checker.IsTypeAssignableTo(target, member) || len(l.checker.GetIndexInfosOfType(member)) != 0 || len(l.checker.GetSignaturesOfType(member, checker.SignatureKindCall)) != 0 || len(l.checker.GetSignaturesOfType(member, checker.SignatureKindConstruct)) != 0 {
		return false
	}
	for _, property := range l.checker.GetPropertiesOfType(member) {
		own := l.checker.GetPropertyOfType(target, property.Name)
		if own == nil || !l.checker.IsTypeAssignableTo(l.checker.GetTypeOfSymbol(own), l.checker.GetTypeOfSymbol(property)) {
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
func (l *lowering) brandedIntersectionReceiver(receiver *checker.Type) *checker.Type {
	if receiver.Flags()&checker.TypeFlagsIntersection == 0 {
		return receiver
	}
	for _, member := range receiver.Types() {
		if member.Flags()&checker.TypeFlagsObject != 0 && !isClassInstance(member) && l.fieldLiteral(member, "kind") == nil && l.brandedSupertype(member) && l.brandedIntersection(receiver, member) {
			return member
		}
	}
	return receiver
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
