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
	found := l.promiseIntoVoid(source, target, 0, map[[2]*checker.Type]bool{})
	if found == nil {
		return nil
	}
	if found.depthLimit {
		return l.notYet(node, "a structural relation deeper than 64 type pairs while checking promise-returning functions seen as void-returning")
	}
	what := "an unawaited async task: a function returning " + l.checker.TypeToString(found.returns) + " at the " + name + " site"
	if found.path != "" {
		what += " through " + found.path
	}
	what += " seen as " + l.checker.TypeToString(found.target) + ", whose caller drops its Promise"
	return &Refused{Where: l.program.Where(node), What: what,
		Fix: "await the task inside a function that returns void, or keep the Promise in the type it's seen as"}
}

func init() { refusalRules.register(3210, "refusePromiseIntoVoid", (*lowering).refusePromiseIntoVoid) }

// promiseDiscard is a position where a function whose result can hold a promise meets one that returns
// void: the source's return type there, the void-returning target, and the path to it.
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
	if returns := l.promiseReturning(source); returns != nil && l.voidReturning(target) {
		return &promiseDiscard{returns: returns, target: target}
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

// promiseReturning is the return type of the first of source's call signatures whose result can hold a
// promise, or nil.
func (l *lowering) promiseReturning(source *checker.Type) *checker.Type {
	for _, signature := range l.checker.GetSignaturesOfType(source, checker.SignatureKindCall) {
		if returns := l.checker.GetReturnTypeOfSignature(signature); l.mayHoldPromise(returns, 0) {
			return returns
		}
	}
	return nil
}

// voidReturning is whether any of target's call signatures returns exactly void, so a call through it
// drops whatever the function returned.
func (l *lowering) voidReturning(target *checker.Type) bool {
	for _, signature := range l.checker.GetSignaturesOfType(target, checker.SignatureKindCall) {
		if l.checker.GetReturnTypeOfSignature(signature).Flags() == checker.TypeFlagsVoid {
			return true
		}
	}
	return false
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
