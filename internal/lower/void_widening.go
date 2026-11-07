package lower

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// A void call has already discarded its result. No runtime check can recover it
// when another view of the function promises a value, even through a union.
func (lowering *lowering) refuseVoidWidening(node *ast.Node) error {
	if lowering.generatorStepped(node) {
		return nil // A spread or Array.from never reads a generator's last value (generators.go).
	}
	name, source, target := lowering.relationSite(node)
	if target == nil || source == nil {
		return nil
	}
	what := ""
	if name == "return" && source.Flags() == checker.TypeFlagsVoid && target.Flags() != checker.TypeFlagsVoid {
		what = "a void value at a return site whose function returns " + lowering.checker.TypeToString(target)
	} else if found := lowering.voidWidenedUnlessUndefined(node, source, target); found != nil {
		if found.depthLimit {
			return lowering.notYet(
				node,
				"a structural relation deeper than 64 type pairs while checking void function results",
			)
		}
		what = "a void-returning function at the " + name + " site"
		if found.contained {
			what = "a void value at the " + name + " site"
		}
		if found.path != "" {
			what += " through " + found.path
		}
		what += " seen as " + lowering.checker.TypeToString(found.target)
	}
	if what == "" {
		return nil
	}
	return &Refused{Where: lowering.program.Where(node), What: what,
		Fix: "declare the function's real return type (adamic/no-void-widening)"}
}

func init() { refusalRules.register(3200, "refuseVoidWidening", (*lowering).refuseVoidWidening) }

// relationSite is the relation an expression takes part in where it stands: the site's name, the
// expression's type and the type it's seen as there, or nil types when it's no relation's source. It
// covers initializers, assignments, returns (awaited in an async function), arguments (a rest
// parameter's element), object fields and shorthand fields, array elements, assertions and the implied
// target. refuseVoidWidening and refuseAnyRelation share it.
func (lowering *lowering) relationSite(node *ast.Node) (string, *checker.Type, *checker.Type) {
	if node.Kind == ast.KindParenthesizedExpression || !lowering.isExpression(node) {
		return "", nil, nil
	}
	site := node
	for site.Parent != nil && site.Parent.Kind == ast.KindParenthesizedExpression {
		site = site.Parent
	}
	parent := site.Parent
	if parent == nil {
		return "", nil, nil
	}
	source := lowering.checker.GetTypeAtLocation(node)
	var target *checker.Type
	name := "relation"
	if viewSite(node) {
		target = lowering.slotType(node)
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
				returns := lowering.checker.GetReturnTypeOfSignature(
					lowering.checker.GetSignatureFromDeclaration(function),
				)
				target = checker.Checker_getAwaitedType(lowering.checker, returns)
			}
			break
		}
	case ast.KindPropertyAssignment:
		name = "object field"
	case ast.KindShorthandPropertyAssignment:
		name = "object field"
		if symbol := lowering.checker.GetShorthandAssignmentValueSymbol(parent); symbol != nil {
			source = lowering.checker.GetTypeOfSymbol(symbol)
		}
		if object := lowering.slotType(parent.Parent); object != nil {
			target = lowering.checker.GetTypeOfPropertyOfType(object, parent.Name().Text())
		}
	case ast.KindArrayLiteralExpression:
		name = "array element"
	case ast.KindAsExpression:
		name = "assertion"
		target = lowering.checker.GetTypeAtLocation(parent)
	case ast.KindCallExpression, ast.KindNewExpression:
		name = "argument"
		// Explicit type arguments do not always contextualize the argument.
		// Read the instantiated parameter, not the generic declaration's binder.
		if signature := lowering.checker.GetResolvedSignature(parent); signature != nil {
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
				target = lowering.checker.GetTypeOfSymbol(parameter)
				if len(parameter.Declarations) > 0 &&
					parameter.Declarations[0].Kind == ast.KindParameter &&
					parameter.Declarations[0].AsParameterDeclaration().DotDotDotToken != nil {
					target = lowering.elementTypeOfArray(target)
				}
			}
		}
	}
	if lowering.slotSet(node) {
		// The slot a caller set is what the value is written into, not the resolved signature's parameter.
		target = lowering.slotType(node)
	}
	if target == nil {
		target = lowering.impliedTarget(node)
	}
	return name, source, target
}

func (lowering *lowering) onlyVoidCalls(source *checker.Type) bool {
	signatures := lowering.checker.GetSignaturesOfType(source, checker.SignatureKindCall)
	if len(signatures) == 0 {
		return false
	}
	for _, signature := range signatures {
		if lowering.checker.GetReturnTypeOfSignature(signature).Flags() != checker.TypeFlagsVoid {
			return false
		}
	}
	return true
}

func (lowering *lowering) hasWiderVoidCall(target *checker.Type) bool {
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range target.Types() {
			if lowering.hasWiderVoidCall(member) {
				return true
			}
		}
	}
	for _, signature := range lowering.checker.GetSignaturesOfType(target, checker.SignatureKindCall) {
		if lowering.checker.GetReturnTypeOfSignature(signature).Flags() != checker.TypeFlagsVoid {
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

func (lowering *lowering) voidWidened(
	source, target *checker.Type,
	depth int,
	seen map[[2]*checker.Type]bool,
) *voidWidening {
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
	if lowering.onlyVoidCalls(source) && lowering.hasWiderVoidCall(target) {
		return &voidWidening{target: target}
	}
	if source.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range source.Types() {
			if found := lowering.voidWidened(member, target, depth+1, seen); found != nil {
				return found
			}
		}
		return nil
	}
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range lowering.reachableArms(source, target.Types()) {
			if found := lowering.voidWidened(source, member, depth+1, seen); found != nil {
				return found
			}
		}
		return nil
	}
	if source.Flags()&checker.TypeFlagsTypeParameter != 0 {
		return lowering.voidWidened(lowering.checker.GetBaseConstraintOfType(source), target, depth+1, seen)
	}
	if target.Flags()&checker.TypeFlagsTypeParameter != 0 {
		return lowering.voidWidened(source, lowering.checker.GetBaseConstraintOfType(target), depth+1, seen)
	}
	if source.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) == 0 ||
		target.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) == 0 {
		return nil
	}
	if (lowering.isArrayType(source) || checker.IsTupleType(source)) &&
		(lowering.isArrayType(target) || checker.IsTupleType(target)) {
		if checker.IsTupleType(source) && checker.IsTupleType(target) {
			sources, targets := lowering.checker.GetTypeArguments(source), lowering.checker.GetTypeArguments(target)
			for index, inside := range sources {
				if index < len(targets) {
					if found := lowering.voidContained(inside, targets[index], depth, seen); found != nil {
						found.path = "[" + strconv.Itoa(index) + "]" + found.path
						return found
					}
				}
			}
			return nil
		}
		if found := lowering.voidContained(
			lowering.checker.GetNumberIndexType(source),
			lowering.checker.GetNumberIndexType(target),
			depth,
			seen,
		); found != nil {
			found.path = "[]" + found.path
			return found
		}
		return nil
	}
	for _, kind := range []checker.SignatureKind{checker.SignatureKindCall, checker.SignatureKindConstruct} {
		if found := lowering.voidWidenedSignatures(source, target, kind, depth, seen); found != nil {
			return found
		}
	}
	if source.ObjectFlags()&checker.ObjectFlagsReference != 0 &&
		target.ObjectFlags()&checker.ObjectFlagsReference != 0 &&
		source.Target() != nil &&
		source.Target() == target.Target() {
		sources, targets := lowering.checker.GetTypeArguments(source), lowering.checker.GetTypeArguments(target)
		for index, inside := range sources {
			if index < len(targets) {
				if found := lowering.voidContained(inside, targets[index], depth, seen); found != nil {
					found.path = "<" + lowering.typeArgumentName(source, index) + ">" + found.path
					return found
				}
			}
		}
	}
	for _, property := range lowering.checker.GetPropertiesOfType(target) {
		from := lowering.checker.GetTypeOfPropertyOfType(source, property.Name)
		to := lowering.checker.GetTypeOfPropertyOfType(target, property.Name)
		if found := lowering.voidContained(from, to, depth, seen); found != nil {
			step := "." + property.Name
			if strings.HasPrefix(property.Name, ast.InternalSymbolNamePrefix) {
				// A well-known symbol's key is internal; name it as the source spells it.
				step = lowering.checker.SymbolToString(property)
			}
			found.path = step + found.path
			return found
		}
	}
	return nil
}

// typeArgumentName names a reference's type argument at index by its target's type parameter, as the
// declaration spells it, so a path reads next()<TReturn> rather than next()<1>, a position the reader
// would have to count to in the lib (#v97ejyb). A target with no parameter there keeps the 1-based
// position, so a step is never empty.
func (lowering *lowering) typeArgumentName(reference *checker.Type, index int) string {
	if declared := reference.Target().AsInterfaceType(); declared != nil {
		if parameters := declared.TypeParameters(); index < len(parameters) {
			if name := lowering.checker.TypeToString(parameters[index]); name != "" {
				return name
			}
		}
	}
	return strconv.Itoa(index + 1)
}

// reachableArms is the arms of a target union a value of source can be read through: every arm except
// one source is disjoint from by a discriminant, a property both declare whose types are unit literals
// with no value in common. A generator's IteratorReturnResult<void> (done: true) is never read through
// IteratorYieldResult<number> (done?: false), so relating the two refused `Iterable<number, void>`
// taking a function*'s call (#wcva20f, #54c1r7k). Disjointness is the only proof (Compiler's rule,
// ruled by Adamic, Nov 5): a source that isn't assignable to an arm can still be read through it, since
// a program selects an arm by a user predicate, an in test or a typeof on another field, so tsc's
// assignability would drop a pairing that's really read. A source disjoint from every arm, which a
// checked relation shouldn't produce, keeps them all.
func (lowering *lowering) reachableArms(source *checker.Type, arms []*checker.Type) []*checker.Type {
	reachable := make([]*checker.Type, 0, len(arms))
	for _, arm := range arms {
		if !lowering.discriminatedApart(source, arm) {
			reachable = append(reachable, arm)
		}
	}
	if len(reachable) == 0 {
		return arms
	}
	return reachable
}

// discriminatedApart is whether source and arm share a property whose types on each side are unit
// literal types, or unions of them, with no value in common: done: true against done?: false, or
// kind: 'a' against kind: 'b'. A property that isn't all unit literals on either side proves nothing.
func (lowering *lowering) discriminatedApart(source, arm *checker.Type) bool {
	if source.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) == 0 ||
		arm.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) == 0 {
		return false
	}
	for _, property := range lowering.checker.GetPropertiesOfType(arm) {
		from := lowering.checker.GetTypeOfPropertyOfType(source, property.Name)
		to := lowering.checker.GetTypeOfPropertyOfType(arm, property.Name)
		if from == nil || to == nil {
			continue
		}
		fromValues, fromUnits := lowering.unitValues(from)
		toValues, toUnits := lowering.unitValues(to)
		if !fromUnits || !toUnits {
			continue
		}
		// An optional property can be absent, which reads undefined, whether or not its type says so: under
		// exactOptionalPropertyTypes it doesn't, and two absent kinds would read as apart (Compiler, #wcva20f).
		if inside := lowering.checker.GetPropertyOfType(
			source,
			property.Name,
		); inside != nil && inside.Flags&ast.SymbolFlagsOptional != 0 {
			fromValues["undefined"] = true
		}
		if property.Flags&ast.SymbolFlagsOptional != 0 {
			toValues["undefined"] = true
		}
		shared := false
		for value := range fromValues {
			shared = shared || toValues[value]
		}
		if !shared {
			return true
		}
	}
	return false
}

// unitValues is the values a type admits when it's a unit literal type or a union of them: true,
// false, undefined, null, and string, number and bigint literals. A numeric enum's whole type is open,
// so a type holding it has none (holdsOpenNumericEnum). A type that admits any other value has none too.
func (lowering *lowering) unitValues(t *checker.Type) (map[string]bool, bool) {
	members := []*checker.Type{t}
	if t.Flags()&checker.TypeFlagsUnion != 0 {
		members = t.Types()
	}
	if lowering.holdsOpenNumericEnum(members) {
		return nil, false
	}
	values := map[string]bool{}
	for _, member := range members {
		switch flags := member.Flags(); {
		case flags&checker.TypeFlagsBooleanLiteral != 0:
			values["boolean "+lowering.checker.TypeToString(member)] = true
		case flags&checker.TypeFlagsUndefined != 0:
			values["undefined"] = true
		case flags&checker.TypeFlagsNull != 0:
			values["null"] = true
		case flags&checker.TypeFlagsStringLiteral != 0:
			values["string "+fmt.Sprint(member.AsLiteralType().Value())] = true
		case flags&checker.TypeFlagsNumberLiteral != 0:
			values["number "+fmt.Sprint(member.AsLiteralType().Value())] = true
		case flags&checker.TypeFlagsBigIntLiteral != 0:
			values["bigint "+fmt.Sprint(member.AsLiteralType().Value())] = true
		default:
			return nil, false
		}
	}
	return values, true
}

// holdsOpenNumericEnum is whether members, a union's flattened members, hold a numeric enum's whole
// type. A numeric enum is open (rulings.md:216): any number can be stored in it, so it isn't a closed
// set of values (Compiler's must on #wcva20f). A member literal, or a union of some of the members, is
// closed, since every road from a number into an enum-literal tag slot is refused (rulings.md:220,
// #w7w2asn), exactly as far as the open-enum tag walk trusts it (Language, Nov 5). tsc flattens a union
// naming every member into the enum's own members, so an enum all of whose members are here is the open
// type, also inside a wider union (Kind | "x"). An enum-literal number of an enum that isn't purely
// numeric fails closed. wholeNumericEnums decides which enums are whole, as it does for
// openNumericEnumType (#4cfvp4q).
func (lowering *lowering) holdsOpenNumericEnum(members []*checker.Type) bool {
	whole, impure := lowering.wholeNumericEnums(members)
	return impure || len(whole) > 0
}

// A function can hand out a void function from its return, or receive one through
// a parameter the other view reads. Returns relate covariantly and parameters
// contravariantly, and every target signature is checked against every source one.
func (lowering *lowering) voidWidenedSignatures(
	source, target *checker.Type,
	kind checker.SignatureKind,
	depth int,
	seen map[[2]*checker.Type]bool,
) *voidWidening {
	sources := lowering.checker.GetSignaturesOfType(source, kind)
	targets := lowering.checker.GetSignaturesOfType(target, kind)
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
			to, from := checker.Checker_getCanonicalSignature(
				lowering.checker,
				to,
			), checker.Checker_getCanonicalSignature(lowering.checker, from)
			if found := lowering.voidContained(
				lowering.checker.GetReturnTypeOfSignature(from),
				lowering.checker.GetReturnTypeOfSignature(to),
				depth,
				seen,
			); found != nil {
				found.path = call + "()" + found.path
				return found
			}
			for position := range voidRelatedPositions(from, to) {
				if found := lowering.voidContained(
					lowering.checker.GetTypeAtPosition(to, position),
					lowering.checker.GetTypeAtPosition(from, position),
					depth,
					seen,
				); found != nil {
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
func (lowering *lowering) voidContained(
	inside, wider *checker.Type,
	depth int,
	seen map[[2]*checker.Type]bool,
) *voidWidening {
	if inside != nil &&
		wider != nil &&
		inside.Flags() == checker.TypeFlagsVoid &&
		wider.Flags() != checker.TypeFlagsVoid {
		return &voidWidening{target: wider, contained: true}
	}
	return lowering.voidWidened(inside, wider, depth+1, seen)
}

// voidWidenedUnlessUndefined is voidWidened, except that a function written in place whose every return
// is provably undefined isn't a lying void function (#9fj31n7, ruled Oct 28): its own result widens
// freely, then's and catch's slots included, since nothing it returns is a value Node would keep. Its
// parameters are still related. A concise body that's a call, a function value, a parameter or a field
// stays refused, since it could carry a value or a thenable that Node would follow.
func (lowering *lowering) voidWidenedUnlessUndefined(node *ast.Node, source, target *checker.Type) *voidWidening {
	found := lowering.voidWidened(source, target, 0, map[[2]*checker.Type]bool{})
	if found == nil ||
		found.depthLimit ||
		(found.path != "" && found.path != "()") ||
		!lowering.returnsOnlyUndefined(node) {
		return found
	}
	return lowering.voidWidenedParameters(source, target, map[[2]*checker.Type]bool{})
}

// voidWidenedParameters relates only the parameter positions of source's call signatures to target's.
func (lowering *lowering) voidWidenedParameters(
	source, target *checker.Type,
	seen map[[2]*checker.Type]bool,
) *voidWidening {
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range target.Types() {
			if found := lowering.voidWidenedParameters(source, member, seen); found != nil {
				return found
			}
		}
		return nil
	}
	for _, to := range lowering.checker.GetSignaturesOfType(target, checker.SignatureKindCall) {
		for _, from := range lowering.checker.GetSignaturesOfType(source, checker.SignatureKindCall) {
			to, from := checker.Checker_getCanonicalSignature(
				lowering.checker,
				to,
			), checker.Checker_getCanonicalSignature(lowering.checker, from)
			for position := range voidRelatedPositions(from, to) {
				if found := lowering.voidContained(
					lowering.checker.GetTypeAtPosition(to, position),
					lowering.checker.GetTypeAtPosition(from, position),
					0,
					seen,
				); found != nil {
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
// body that's undefined or void's result. Nested functions' returns are their own. The return-site
// refusal is the guard for a body returning a void value; this predicate keeps the admission exactly
// the ruled shape, which is why a mutant loosening either branch here leaves `() => make()` refused.
func (lowering *lowering) returnsOnlyUndefined(node *ast.Node) bool {
	if node.Kind != ast.KindArrowFunction && node.Kind != ast.KindFunctionExpression ||
		ast.HasSyntacticModifier(node, ast.ModifierFlagsAsync) {
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
		return lowering.provablyUndefined(body)
	}
	undefinedOnly := true
	var visit ast.Visitor
	visit = func(child *ast.Node) bool {
		if !undefinedOnly || ast.IsFunctionLike(child) || ast.IsClassLike(child) {
			return false
		}
		if child.Kind == ast.KindReturnStatement {
			if returned := child.AsReturnStatement().Expression; returned != nil &&
				!lowering.provablyUndefined(returned) {
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
func (lowering *lowering) provablyUndefined(expression *ast.Node) bool {
	expression = ast.SkipParentheses(expression)
	return expression.Kind == ast.KindVoidExpression || ast.IsIdentifier(expression) &&
		expression.Text() == "undefined" &&
		lowering.checker.GetTypeAtLocation(expression).Flags() == checker.TypeFlagsUndefined
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
