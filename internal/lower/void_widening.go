package lower

import (
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// A void call has already discarded its result. No runtime check can recover it
// when another view of the function promises a value, even through a union.
func (l *lowering) refuseVoidWidening(node *ast.Node) error {
	if node.Kind == ast.KindParenthesizedExpression || !l.isExpression(node) {
		return nil
	}
	site := node
	for site.Parent != nil && site.Parent.Kind == ast.KindParenthesizedExpression {
		site = site.Parent
	}
	parent := site.Parent
	if parent == nil {
		return nil
	}
	source := l.checker.GetTypeAtLocation(node)
	var target *checker.Type
	name := "relation"
	if viewSite(node) {
		target = l.checker.GetContextualType(node, checker.ContextFlagsNone)
	}
	switch parent.Kind {
	case ast.KindVariableDeclaration, ast.KindParameter, ast.KindPropertyDeclaration:
		name = "initializer"
	case ast.KindBinaryExpression:
		name = "assignment"
	case ast.KindReturnStatement, ast.KindArrowFunction:
		name = "return"
		for function := parent; function != nil; function = function.Parent {
			if !ast.IsFunctionLike(function) {
				continue
			}
			if ast.HasSyntacticModifier(function, ast.ModifierFlagsAsync) {
				returns := l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(function))
				target = checker.Checker_getAwaitedType(l.checker, returns)
			}
			break
		}
	case ast.KindPropertyAssignment:
		name = "object field"
	case ast.KindShorthandPropertyAssignment:
		name = "object field"
		if symbol := l.checker.GetShorthandAssignmentValueSymbol(parent); symbol != nil {
			source = l.checker.GetTypeOfSymbol(symbol)
		}
		if object := l.checker.GetContextualType(parent.Parent, checker.ContextFlagsNone); object != nil {
			target = l.checker.GetTypeOfPropertyOfType(object, parent.Name().Text())
		}
	case ast.KindArrayLiteralExpression:
		name = "array element"
	case ast.KindAsExpression:
		name = "assertion"
		target = l.checker.GetTypeAtLocation(parent)
	case ast.KindCallExpression, ast.KindNewExpression:
		name = "argument"
		// Explicit type arguments do not always contextualize the argument.
		// Read the instantiated parameter, not the generic declaration's binder.
		if signature := l.checker.GetResolvedSignature(parent); signature != nil {
			var arguments []*ast.Node
			if parent.Kind == ast.KindCallExpression {
				arguments = parent.AsCallExpression().Arguments.Nodes
			} else if parent.AsNewExpression().Arguments != nil {
				arguments = parent.AsNewExpression().Arguments.Nodes
			}
			parameters := signature.Parameters()
			for index, argument := range arguments {
				if argument != site || len(parameters) == 0 {
					continue
				}
				parameter := parameters[min(index, len(parameters)-1)]
				target = l.checker.GetTypeOfSymbol(parameter)
				if len(parameter.Declarations) > 0 && parameter.Declarations[0].Kind == ast.KindParameter && parameter.Declarations[0].AsParameterDeclaration().DotDotDotToken != nil {
					target = l.checker.GetElementTypeOfArrayType(target)
				}
			}
		}
	}
	if target == nil {
		target = l.impliedTarget(node)
	}
	if target == nil || source == nil {
		return nil
	}
	what := ""
	if name == "return" && source.Flags() == checker.TypeFlagsVoid && target.Flags() != checker.TypeFlagsVoid {
		what = "a void value at a return site whose function returns " + l.checker.TypeToString(target)
	} else if found := l.voidWidenedUnlessUndefined(node, source, target); found != nil {
		if found.depthLimit {
			return l.notYet(node, "a structural relation deeper than 64 type pairs while checking void function results")
		}
		what = "a void-returning function at the " + name + " site"
		if found.contained {
			what = "a void value at the " + name + " site"
		}
		if found.path != "" {
			what += " through " + found.path
		}
		what += " seen as " + l.checker.TypeToString(found.target)
	}
	if what == "" {
		return nil
	}
	return &Refused{Where: l.program.Where(node), What: what,
		Fix: "declare the function's real return type (adamic/no-void-widening)"}
}

func (l *lowering) onlyVoidCalls(source *checker.Type) bool {
	signatures := l.checker.GetSignaturesOfType(source, checker.SignatureKindCall)
	if len(signatures) == 0 {
		return false
	}
	for _, signature := range signatures {
		if l.checker.GetReturnTypeOfSignature(signature).Flags() != checker.TypeFlagsVoid {
			return false
		}
	}
	return true
}

func (l *lowering) hasWiderVoidCall(target *checker.Type) bool {
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range target.Types() {
			if l.hasWiderVoidCall(member) {
				return true
			}
		}
	}
	for _, signature := range l.checker.GetSignaturesOfType(target, checker.SignatureKindCall) {
		if l.checker.GetReturnTypeOfSignature(signature).Flags() != checker.TypeFlagsVoid {
			return true
		}
	}
	return false
}

// A view can hide the discarded result inside a container, including a phantom
// interface argument. Arrays use element slots, never their library methods.
type voidWidening struct {
	target     *checker.Type
	path       string
	depthLimit bool
	// contained is a void value at a position (#p7t8tmh), not a void-returning function.
	contained bool
}

func (l *lowering) voidWidened(source, target *checker.Type, depth int, seen map[[2]*checker.Type]bool) *voidWidening {
	if source == nil || target == nil || source == target {
		return nil
	}
	pair := [2]*checker.Type{source, target}
	if seen[pair] {
		return nil
	}
	seen[pair] = true
	if depth >= 64 {
		// Exhausting a proof must not silently admit a deeper unsafe callable.
		return &voidWidening{depthLimit: true}
	}
	if l.onlyVoidCalls(source) && l.hasWiderVoidCall(target) {
		return &voidWidening{target: target}
	}
	if source.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range source.Types() {
			if found := l.voidWidened(member, target, depth+1, seen); found != nil {
				return found
			}
		}
		return nil
	}
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range target.Types() {
			if found := l.voidWidened(source, member, depth+1, seen); found != nil {
				return found
			}
		}
		return nil
	}
	if source.Flags()&checker.TypeFlagsTypeParameter != 0 {
		return l.voidWidened(l.checker.GetBaseConstraintOfType(source), target, depth+1, seen)
	}
	if target.Flags()&checker.TypeFlagsTypeParameter != 0 {
		return l.voidWidened(source, l.checker.GetBaseConstraintOfType(target), depth+1, seen)
	}
	if source.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) == 0 || target.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) == 0 {
		return nil
	}
	if (l.checker.IsArrayType(source) || checker.IsTupleType(source)) && (l.checker.IsArrayType(target) || checker.IsTupleType(target)) {
		if checker.IsTupleType(source) && checker.IsTupleType(target) {
			sources, targets := l.checker.GetTypeArguments(source), l.checker.GetTypeArguments(target)
			for index, inside := range sources {
				if index < len(targets) {
					if found := l.voidContained(inside, targets[index], depth, seen); found != nil {
						found.path = "[" + strconv.Itoa(index) + "]" + found.path
						return found
					}
				}
			}
			return nil
		}
		if found := l.voidContained(l.checker.GetNumberIndexType(source), l.checker.GetNumberIndexType(target), depth, seen); found != nil {
			found.path = "[]" + found.path
			return found
		}
		return nil
	}
	for _, kind := range []checker.SignatureKind{checker.SignatureKindCall, checker.SignatureKindConstruct} {
		if found := l.voidWidenedSignatures(source, target, kind, depth, seen); found != nil {
			return found
		}
	}
	if source.ObjectFlags()&checker.ObjectFlagsReference != 0 && target.ObjectFlags()&checker.ObjectFlagsReference != 0 && source.Target() != nil && source.Target() == target.Target() {
		sources, targets := l.checker.GetTypeArguments(source), l.checker.GetTypeArguments(target)
		for index, inside := range sources {
			if index < len(targets) {
				if found := l.voidContained(inside, targets[index], depth, seen); found != nil {
					found.path = "<" + strconv.Itoa(index+1) + ">" + found.path
					return found
				}
			}
		}
	}
	for _, property := range l.checker.GetPropertiesOfType(target) {
		from := l.checker.GetTypeOfPropertyOfType(source, property.Name)
		to := l.checker.GetTypeOfPropertyOfType(target, property.Name)
		if found := l.voidContained(from, to, depth, seen); found != nil {
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

// A function can hand out a void function from its return, or receive one through
// a parameter the other view reads. Returns relate covariantly and parameters
// contravariantly, and every target signature is checked against every source one.
func (l *lowering) voidWidenedSignatures(source, target *checker.Type, kind checker.SignatureKind, depth int, seen map[[2]*checker.Type]bool) *voidWidening {
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
			// Each instantiation of a generic interface clones its methods' type
			// parameters, so Set's union<U> would build Set<T | U | U | ...> without
			// end. The canonical signature restores the original unconstrained ones.
			to, from := checker.Checker_getCanonicalSignature(l.checker, to), checker.Checker_getCanonicalSignature(l.checker, from)
			if found := l.voidContained(l.checker.GetReturnTypeOfSignature(from), l.checker.GetReturnTypeOfSignature(to), depth, seen); found != nil {
				found.path = call + "()" + found.path
				return found
			}
			for position := range voidRelatedPositions(from, to) {
				if found := l.voidContained(l.checker.GetTypeAtPosition(to, position), l.checker.GetTypeAtPosition(from, position), depth, seen); found != nil {
					found.path = call + "(" + strconv.Itoa(position) + ")" + found.path
					return found
				}
			}
		}
	}
	return nil
}

// voidContained relates one position inside a container: an element, a type argument, a property,
// a function's return, or (through the walk's swap) a parameter. A discard releases what a void call
// returns where Node keeps it, so void at a position seen wider there is a read waiting to happen:
// [1].map(() => make()) holds 42 on Node, and seen as readonly unknown[] an element would read it
// (#p7t8tmh, ruled Oct 27). A union that holds void isn't this leaf: H2 already keeps every void
// value out of an X | void slot, so only a position that is exactly void is.
func (l *lowering) voidContained(inside, wider *checker.Type, depth int, seen map[[2]*checker.Type]bool) *voidWidening {
	if inside != nil && wider != nil && inside.Flags() == checker.TypeFlagsVoid && wider.Flags() != checker.TypeFlagsVoid {
		return &voidWidening{target: wider, contained: true}
	}
	return l.voidWidened(inside, wider, depth+1, seen)
}

// voidWidenedUnlessUndefined is voidWidened, except that a function written in place whose every return
// is provably undefined isn't a lying void function (#9fj31n7, ruled Oct 28): its own result widens
// freely, then's and catch's slots included, since nothing it returns is a value Node would keep. Its
// parameters are still related. A concise body that's a call, a function value, a parameter or a field
// stays refused, since it could carry a value or a thenable that Node would follow.
func (l *lowering) voidWidenedUnlessUndefined(node *ast.Node, source, target *checker.Type) *voidWidening {
	found := l.voidWidened(source, target, 0, map[[2]*checker.Type]bool{})
	if found == nil || found.depthLimit || (found.path != "" && found.path != "()") || !l.returnsOnlyUndefined(node) {
		return found
	}
	return l.voidWidenedParameters(source, target, map[[2]*checker.Type]bool{})
}

// voidWidenedParameters relates only the parameter positions of source's call signatures to target's.
func (l *lowering) voidWidenedParameters(source, target *checker.Type, seen map[[2]*checker.Type]bool) *voidWidening {
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range target.Types() {
			if found := l.voidWidenedParameters(source, member, seen); found != nil {
				return found
			}
		}
		return nil
	}
	for _, to := range l.checker.GetSignaturesOfType(target, checker.SignatureKindCall) {
		for _, from := range l.checker.GetSignaturesOfType(source, checker.SignatureKindCall) {
			to, from := checker.Checker_getCanonicalSignature(l.checker, to), checker.Checker_getCanonicalSignature(l.checker, from)
			for position := range voidRelatedPositions(from, to) {
				if found := l.voidContained(l.checker.GetTypeAtPosition(to, position), l.checker.GetTypeAtPosition(from, position), 0, seen); found != nil {
					found.path = "(" + strconv.Itoa(position) + ")" + found.path
					return found
				}
			}
		}
	}
	return nil
}

// returnsOnlyUndefined is whether node is an arrow or a function expression, neither async nor a
// generator, whose every return is provably undefined: a block with no return of a value, or a concise
// body that's undefined or void's result. Nested functions' returns are their own.
func (l *lowering) returnsOnlyUndefined(node *ast.Node) bool {
	if node.Kind != ast.KindArrowFunction && node.Kind != ast.KindFunctionExpression || ast.HasSyntacticModifier(node, ast.ModifierFlagsAsync) {
		return false
	}
	if node.Kind == ast.KindFunctionExpression && node.AsFunctionExpression().AsteriskToken != nil {
		return false
	}
	body := node.Body()
	if body == nil {
		return false
	}
	if body.Kind != ast.KindBlock {
		return l.provablyUndefined(body)
	}
	undefinedOnly := true
	var visit ast.Visitor
	visit = func(child *ast.Node) bool {
		if !undefinedOnly || ast.IsFunctionLike(child) || ast.IsClassLike(child) {
			return false
		}
		if child.Kind == ast.KindReturnStatement {
			if returned := child.AsReturnStatement().Expression; returned != nil && !l.provablyUndefined(returned) {
				undefinedOnly = false
				return true
			}
		}
		return child.ForEachChild(visit)
	}
	body.ForEachChild(visit)
	return undefinedOnly
}

// provablyUndefined is whether expression is void's result or the value undefined.
func (l *lowering) provablyUndefined(expression *ast.Node) bool {
	expression = ast.SkipParentheses(expression)
	return expression.Kind == ast.KindVoidExpression || ast.IsIdentifier(expression) && expression.Text() == "undefined" && l.checker.GetTypeAtLocation(expression).Flags() == checker.TypeFlagsUndefined
}

// voidRelatedPositions is how many parameter positions of to flow into from. A function never sees an
// argument past its own parameters, so `() => { ... }` reads nothing forEach or then hands it; past a
// rest parameter, every position reads its element type.
func voidRelatedPositions(from, to *checker.Signature) int {
	if from.HasRestParameter() {
		return max(len(from.Parameters()), len(to.Parameters()))
	}
	return len(from.Parameters())
}
