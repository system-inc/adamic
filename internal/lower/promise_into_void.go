package lower

import (
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// A function whose result can hold a promise, seen as one that returns void, hands its promise to a
// call site that drops it: `const run: () => void = fails; run();`. The floating-task refusal (:355)
// judges a call's own type, so it never sees that discard, and a rejection there exits 0 natively where
// Node reports it and exits 1 (#8j8ecw0, Runtime Performance's probe on #995spev). Adamic ruled it :355
// applied to every road a promise can reach a discard by: Refused at the relation, in .a and .ts, through
// every position the void-widening walk reads. A thenable counts as a promise: Node's await and its
// unhandled-rejection rule follow any callable then. The run-time rule that lifts this stop is Library's
// (#995spev, point 3): a rejection no reaction took by the end of the drain exits 1, on every road.
func (l *lowering) refusePromiseIntoVoid(node *ast.Node) error {
	name, source, target := l.relationSite(node)
	if target == nil || source == nil {
		return nil
	}
	if name == "return" && l.inAsyncFunction(node) {
		// An async function's return adopts a promise it's handed, so its value is read awaited, as
		// relationSite reads the target (`return fails();` keeps the task).
		if awaited := checker.Checker_getAwaitedType(l.checker, source); awaited != nil {
			source = awaited
		}
	}
	found := l.promiseIntoVoid(source, target, 0, map[[2]*checker.Type]bool{})
	if found == nil {
		return nil
	}
	if found.depthLimit {
		return l.notYet(node, "a structural relation deeper than 64 type pairs while checking promise-returning functions seen as void-returning")
	}
	what := "an unawaited async task: a " + l.checker.TypeToString(found.returns) + " at the " + name + " site"
	if found.path != "" {
		what += " through " + found.path
	}
	what += " seen as " + l.checker.TypeToString(found.target) + ", a type that names no Promise, so whoever holds it drops the task"
	return &Refused{Where: l.program.Where(node), What: what,
		Fix: "await the task inside a function that returns void, or keep the Promise in the type it's seen as"}
}

func init() { refusalRules.register(3210, "refusePromiseIntoVoid", (*lowering).refusePromiseIntoVoid) }

// promiseDiscard is a position where a value that can hold a promise meets a type that names none: the
// value's type there (returns), the target, and the path to it.
type promiseDiscard struct {
	returns    *checker.Type
	target     *checker.Type
	path       string
	depthLimit bool
}

// promiseIntoVoid walks source against target through the positions voidWidened walks: union members on
// either side, a type parameter's constraint, array and tuple elements, a reference's type arguments,
// signatures' returns covariantly and parameters contravariantly (so a callee typed to take a void
// callback, seen as taking a promise-returning one, is found too), and properties. Its bound is 64 pairs,
// past which it answers depthLimit so the caller fails closed. This is the third walk of that shape,
// beside voidWidened and anyRelated; folding them into one is owed.
func (l *lowering) promiseIntoVoid(source, target *checker.Type, depth int, seen map[[2]*checker.Type]bool) *promiseDiscard {
	if source == nil || target == nil || source == target {
		return nil
	}
	pair := [2]*checker.Type{source, target}
	if seen[pair] {
		return nil
	}
	seen[pair] = true
	if depth >= 64 {
		return &promiseDiscard{depthLimit: true}
	}
	// A value that can hold a promise seen as a type that names none: void, unknown, {}, object, any, or
	// a union with no promise member (Language's (a), ruled on #8j8ecw0). A function's result reaches
	// here through its signature's return, so a promise-returning function seen as a void-returning one
	// is this leaf one position in. A widening that still says Promise relates on.
	if l.mayHoldPromise(source, 0) && !l.mayHoldPromise(target, 0) && target.Flags()&checker.TypeFlagsNever == 0 {
		return &promiseDiscard{returns: source, target: target}
	}
	if source.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range source.Types() {
			if found := l.promiseIntoVoid(member, target, depth+1, seen); found != nil {
				return found
			}
		}
		return nil
	}
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range target.Types() {
			if found := l.promiseIntoVoid(source, member, depth+1, seen); found != nil {
				return found
			}
		}
		return nil
	}
	if source.Flags()&checker.TypeFlagsTypeParameter != 0 {
		return l.promiseIntoVoid(l.checker.GetBaseConstraintOfType(source), target, depth+1, seen)
	}
	if target.Flags()&checker.TypeFlagsTypeParameter != 0 {
		return l.promiseIntoVoid(source, l.checker.GetBaseConstraintOfType(target), depth+1, seen)
	}
	if source.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) == 0 || target.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) == 0 {
		return nil
	}
	if (l.isArrayType(source) || checker.IsTupleType(source)) && (l.isArrayType(target) || checker.IsTupleType(target)) {
		if checker.IsTupleType(source) && checker.IsTupleType(target) {
			sources, targets := l.checker.GetTypeArguments(source), l.checker.GetTypeArguments(target)
			for index, inside := range sources {
				if index < len(targets) {
					if found := l.promiseIntoVoid(inside, targets[index], depth+1, seen); found != nil {
						found.path = "[" + strconv.Itoa(index) + "]" + found.path
						return found
					}
				}
			}
			return nil
		}
		if found := l.promiseIntoVoid(l.checker.GetNumberIndexType(source), l.checker.GetNumberIndexType(target), depth+1, seen); found != nil {
			found.path = "[]" + found.path
			return found
		}
		return nil
	}
	for _, kind := range []checker.SignatureKind{checker.SignatureKindCall, checker.SignatureKindConstruct} {
		if found := l.promiseIntoVoidSignatures(source, target, kind, depth, seen); found != nil {
			return found
		}
	}
	if source.ObjectFlags()&checker.ObjectFlagsReference != 0 && target.ObjectFlags()&checker.ObjectFlagsReference != 0 && source.Target() != nil && source.Target() == target.Target() {
		sources, targets := l.checker.GetTypeArguments(source), l.checker.GetTypeArguments(target)
		for index, inside := range sources {
			if index < len(targets) {
				if found := l.promiseIntoVoid(inside, targets[index], depth+1, seen); found != nil {
					found.path = "<" + l.typeArgumentName(source, index) + ">" + found.path
					return found
				}
			}
		}
	}
	for _, property := range l.checker.GetPropertiesOfType(target) {
		from := l.checker.GetTypeOfPropertyOfType(source, property.Name)
		to := l.checker.GetTypeOfPropertyOfType(target, property.Name)
		if found := l.promiseIntoVoid(from, to, depth+1, seen); found != nil {
			step := "." + property.Name
			if strings.HasPrefix(property.Name, ast.InternalSymbolNamePrefix) {
				// A well-known symbol's key is internal; name it as the source spells it.
				step = l.checker.SymbolToString(property)
			}
			found.path = step + found.path
			return found
		}
	}
	return nil
}

// promiseIntoVoidSignatures relates returns covariantly and parameters contravariantly, every target
// signature against every source one, through canonical signatures as voidWidenedSignatures does.
func (l *lowering) promiseIntoVoidSignatures(source, target *checker.Type, kind checker.SignatureKind, depth int, seen map[[2]*checker.Type]bool) *promiseDiscard {
	sources := l.checker.GetSignaturesOfType(source, kind)
	targets := l.checker.GetSignaturesOfType(target, kind)
	if len(sources) == 0 || len(targets) == 0 {
		return nil
	}
	call := ""
	if kind == checker.SignatureKindConstruct {
		call = "new"
	}
	for _, to := range targets {
		for _, from := range sources {
			to, from := checker.Checker_getCanonicalSignature(l.checker, to), checker.Checker_getCanonicalSignature(l.checker, from)
			if found := l.promiseIntoVoid(l.checker.GetReturnTypeOfSignature(from), l.checker.GetReturnTypeOfSignature(to), depth+1, seen); found != nil {
				found.path = call + "()" + found.path
				return found
			}
			for position := range voidRelatedPositions(from, to) {
				if found := l.promiseIntoVoid(l.checker.GetTypeAtPosition(to, position), l.checker.GetTypeAtPosition(from, position), depth+1, seen); found != nil {
					found.path = call + "(" + strconv.Itoa(position) + ")" + found.path
					return found
				}
			}
		}
	}
	return nil
}

// mayHoldPromise is whether a value of t can be a promise: a member of a union or an intersection that
// can, a type parameter whose constraint can, or an object whose then is callable, which is what Node's
// await and its rejection tracking follow. Past a depth of 8 it answers true, so a type it can't see the
// end of fails closed.
func (l *lowering) mayHoldPromise(t *checker.Type, depth int) bool {
	if t == nil {
		return false
	}
	if depth >= 8 {
		return true
	}
	if t.Flags()&checker.TypeFlagsUnionOrIntersection != 0 {
		for _, member := range t.Types() {
			if l.mayHoldPromise(member, depth+1) {
				return true
			}
		}
		return false
	}
	if t.Flags()&checker.TypeFlagsTypeParameter != 0 {
		constraint := l.checker.GetBaseConstraintOfType(t)
		return constraint != nil && constraint != t && l.mayHoldPromise(constraint, depth+1)
	}
	if t.Flags()&checker.TypeFlagsObject == 0 {
		return false
	}
	then := l.checker.GetPropertyOfType(t, "then")
	return then != nil && len(l.checker.GetSignaturesOfType(l.checker.GetTypeOfSymbol(then), checker.SignatureKindCall)) > 0
}

// inAsyncFunction is whether the function node's return belongs to is async, found as relationSite finds
// it: the first function-like ancestor.
func (l *lowering) inAsyncFunction(node *ast.Node) bool {
	for function := node.Parent; function != nil; function = function.Parent {
		if ast.IsFunctionLike(function) {
			return ast.HasSyntacticModifier(function, ast.ModifierFlagsAsync)
		}
	}
	return false
}

// refusePromiseDiscard refuses an expression thrown away (discarded: a statement, void's operand, a for
// statement's initializer or incrementor, a comma's left) whose type can hold a promise, itself or inside
// an array, a tuple or a reference's type arguments: `ok && fails();` and `[0].map(fails);` drop their
// tasks where Node reports the rejection and exits 1 (Language's (b), r06, ruled on #8j8ecw0). A call
// thrown away whole, as a statement or void's operand, is :355's own (statements.go and expression.go
// refuse it by its lowered type), so this one leaves it there and judges the rest.
func (l *lowering) refusePromiseDiscard(node *ast.Node) error {
	if node.Kind == ast.KindParenthesizedExpression || !l.isExpression(node) {
		return nil
	}
	site := node
	for site.Parent != nil && site.Parent.Kind == ast.KindParenthesizedExpression {
		site = site.Parent
	}
	if !discarded(site) {
		return nil
	}
	held := l.checker.GetTypeAtLocation(node)
	whole := l.mayHoldPromise(held, 0)
	if whole && node.Kind == ast.KindCallExpression && (site.Parent.Kind == ast.KindExpressionStatement || site.Parent.Kind == ast.KindVoidExpression) {
		return nil
	}
	if !whole && !l.holdsPromiseInside(held, 0, map[*checker.Type]bool{}) {
		return nil
	}
	return &Refused{Where: l.program.Where(node), What: "an unawaited async task: a " + l.checker.TypeToString(held) + " thrown away, whose tasks nothing awaits",
		Fix: "await the tasks (Promise.all for many), or keep them and return them"}
}

func init() { refusalRules.register(3220, "refusePromiseDiscard", (*lowering).refusePromiseDiscard) }

// holdsPromiseInside is whether a value of t holds a promise in an element of an array or a tuple, or
// in a reference's type argument (a Map's values, a Set's members, a promise of a promise), at any depth.
// A property of an object type isn't followed: an instance thrown away whose field holds a task is the
// run-time rule's to judge (#r7tcjjq). Past a depth of 8 it answers true, so a type it can't see the end
// of fails closed.
func (l *lowering) holdsPromiseInside(t *checker.Type, depth int, seen map[*checker.Type]bool) bool {
	if t == nil || seen[t] {
		return false
	}
	seen[t] = true
	if depth >= 8 {
		return true
	}
	if t.Flags()&checker.TypeFlagsUnionOrIntersection != 0 {
		for _, member := range t.Types() {
			if l.holdsPromiseInside(member, depth+1, seen) {
				return true
			}
		}
		return false
	}
	if t.Flags()&checker.TypeFlagsObject == 0 || t.ObjectFlags()&checker.ObjectFlagsReference == 0 {
		return false
	}
	for _, argument := range l.checker.GetTypeArguments(t) {
		if l.mayHoldPromise(argument, 0) || l.holdsPromiseInside(argument, depth+1, seen) {
			return true
		}
	}
	return false
}
