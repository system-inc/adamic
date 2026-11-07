package lower

import (
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

// A member typed any holds whatever was stored there, and a relation that sees it as a typed member
// reads it as that type with no check: `const h: HasText = loose` turns `text: any` into a string, which
// native then reads as one (#tmdhz1d, Language's probe on main 31fdd16d, an AddressSanitizer death where
// Node prints 5). #bw3xg7c rules any refused; until a declared any member is, a relation that carries one
// into a member that isn't any or unknown is NotYet at the site, through every position voidWidened walks:
// properties, array and tuple elements, a reference's type arguments, and signatures' returns
// (covariantly) and parameters (contravariantly). A value of type any at the site itself is the
// expression's own NotYet, so the walk starts one position in.
//
// Ownership is the source expression's (Adamic's ruling on #tmdhz1d, Lower's line): a program local,
// parameter, field or literal counts, whatever narrowed it, so v2, q4 (`const loose: any[]`) and
// Array.isArray's any[] on a program parameter all count, and a library result stored in a program local
// counts too. Exempt until Language's axiom audit lands, and then replaced by its modelled list: a library
// call's direct result (Object.entries over a record, a method value's call) and a program callback
// passed as a library signature's argument, whose parameters the library fills (replace's callback).
func (lowering *lowering) refuseAnyRelation(node *ast.Node) error {
	name, source, target := lowering.relationSite(node)
	if target == nil || source == nil || source.Flags()&checker.TypeFlagsAny != 0 || lowering.libraryCallResult(node) {
		return nil
	}
	found := lowering.anyRelated(
		source,
		target,
		name == "argument" && lowering.libraryCallArgument(node),
		0,
		map[[2]*checker.Type]bool{},
	)
	if found == nil {
		return nil
	}
	// The lowest-priority stop: another rule's refusal anywhere in the module is the stronger reason, as
	// Array.prototype.fill's method value and Object.entries over a graph record already are, so this one
	// is kept for refuse to return only when nothing else refused.
	if lowering.refusal.deferred == nil {
		if found.depthLimit {
			lowering.refusal.deferred = lowering.notYet(
				node,
				"a structural relation deeper than 64 type pairs while checking members typed any",
			)
		} else {
			lowering.refusal.deferred = lowering.notYet(node, "a member typed any ("+
				found.path+
				") related to "+
				lowering.checker.TypeToString(found.target)+
				" at the "+
				name+
				" site")
		}
	}
	return nil
}

func init() { refusalRules.register(3270, "refuseAnyRelation", (*lowering).refuseAnyRelation) }

// anyRelation is a position where a member typed any meets a member that isn't any or unknown.
type anyRelation struct {
	target     *checker.Type
	path       string
	depthLimit bool
}

// anyRelated walks source against target as voidWidened does, with a seen set of pairs and a bound of
// 64, past which it answers depthLimit so the caller fails closed. Unions relate member by member. Every
// value-carrying position counts its any whoever wrote it, the library included (Language's position rule
// on 99634edf): an element, a property, a type argument, a return. Array.isArray's `any[]` seen as a
// `string[]` holds whatever the value held. Only a parameter position is exempt, and only between two
// signatures the library declares (anyRelatedSignatures).
func (lowering *lowering) anyRelated(
	source, target *checker.Type,
	callback bool,
	depth int,
	seen map[[2]*checker.Type]bool,
) *anyRelation {
	if source == nil || target == nil || source == target {
		return nil
	}
	pair := [2]*checker.Type{source, target}
	if seen[pair] {
		return nil
	}
	seen[pair] = true
	if depth >= 64 {
		return &anyRelation{depthLimit: true}
	}
	if source.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range source.Types() {
			if found := lowering.anyRelated(member, target, callback, depth+1, seen); found != nil {
				return found
			}
		}
		return nil
	}
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range target.Types() {
			if found := lowering.anyRelated(source, member, callback, depth+1, seen); found != nil {
				return found
			}
		}
		return nil
	}
	if source.Flags()&checker.TypeFlagsTypeParameter != 0 {
		return lowering.anyRelated(lowering.checker.GetBaseConstraintOfType(source), target, callback, depth+1, seen)
	}
	if target.Flags()&checker.TypeFlagsTypeParameter != 0 {
		return lowering.anyRelated(source, lowering.checker.GetBaseConstraintOfType(target), callback, depth+1, seen)
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
					if found := lowering.anyContained(inside, targets[index], depth, seen); found != nil {
						found.path = "[" + strconv.Itoa(index) + "]" + found.path
						return found
					}
				}
			}
			return nil
		}
		if found := lowering.anyContained(
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
		if found := lowering.anyRelatedSignatures(source, target, kind, callback, depth, seen); found != nil {
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
				if found := lowering.anyContained(inside, targets[index], depth, seen); found != nil {
					found.path = "<" + strconv.Itoa(index+1) + ">" + found.path
					return found
				}
			}
		}
	}
	for _, property := range lowering.checker.GetPropertiesOfType(target) {
		from := lowering.checker.GetTypeOfPropertyOfType(source, property.Name)
		to := lowering.checker.GetTypeOfPropertyOfType(target, property.Name)
		if found := lowering.anyContained(from, to, depth, seen); found != nil {
			step := property.Name
			if strings.HasPrefix(property.Name, ast.InternalSymbolNamePrefix) {
				// A well-known symbol's key is internal; name it as the source spells it.
				step = lowering.checker.SymbolToString(property)
			}
			if found.path != "" &&
				!strings.HasPrefix(found.path, "[") &&
				!strings.HasPrefix(found.path, "<") &&
				!strings.HasPrefix(found.path, "(") {
				step += "."
			}
			found.path = step + found.path
			return found
		}
	}
	return nil
}

// anyRelatedSignatures relates returns covariantly and parameters contravariantly, every target
// signature against every source one, through canonical signatures as voidWidenedSignatures does. A
// return always counts. Parameters are skipped when both signatures are the library's: a library
// callback's any parameter related to another library signature (a typed array's reduce, Set's union) is
// the library's own modelling of what flows in, and no value the program holds passes through it. So are
// a callback's own parameters when it's passed as a library call's argument (callback): the library fills
// them, until the axiom audit says how.
func (lowering *lowering) anyRelatedSignatures(
	source, target *checker.Type,
	kind checker.SignatureKind,
	callback bool,
	depth int,
	seen map[[2]*checker.Type]bool,
) *anyRelation {
	sources := lowering.checker.GetSignaturesOfType(source, kind)
	targets := lowering.checker.GetSignaturesOfType(target, kind)
	if len(sources) == 0 || len(targets) == 0 {
		return nil
	}
	for _, to := range targets {
		for _, from := range sources {
			to, from := checker.Checker_getCanonicalSignature(
				lowering.checker,
				to,
			), checker.Checker_getCanonicalSignature(lowering.checker, from)
			if found := lowering.anyContained(
				lowering.checker.GetReturnTypeOfSignature(from),
				lowering.checker.GetReturnTypeOfSignature(to),
				depth,
				seen,
			); found != nil {
				found.path = "()" + found.path
				return found
			}
			if callback || librarySignature(from) && librarySignature(to) {
				continue
			}
			for position := range anyRelatedPositions(from, to) {
				if found := lowering.anyContained(
					lowering.checker.GetTypeAtPosition(to, position),
					lowering.checker.GetTypeAtPosition(from, position),
					depth,
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

// anyRelatedPositions is how many parameter positions carry a value from a caller through to into from:
// those to declares, or every one of from's past a rest parameter of to's. A position to doesn't declare
// gets no argument, and the checker's any for it there is no value anyone passes, as with
// `const f: (path: string) => string = fs.realpathSync.native` and its options.
func anyRelatedPositions(from, to *checker.Signature) int {
	if to.HasRestParameter() {
		return voidRelatedPositions(from, to)
	}
	return min(voidRelatedPositions(from, to), len(to.Parameters()))
}

// libraryCallResult is whether node, past parentheses, is a call or new expression whose resolved signature
// the library declares: its result is the library's, until the axiom audit lands.
func (lowering *lowering) libraryCallResult(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindCallExpression && node.Kind != ast.KindNewExpression {
		return false
	}
	signature := lowering.checker.GetResolvedSignature(node)
	return signature != nil && librarySignature(signature)
}

// libraryCallArgument is whether node is an argument of a call or new expression whose resolved signature
// the library declares. A value given a slot by a caller (slotSet) is written into that slot, not handed
// to the library, so it isn't one.
func (lowering *lowering) libraryCallArgument(node *ast.Node) bool {
	if lowering.slotSet(node) {
		return false
	}
	site := node
	for site.Parent != nil && site.Parent.Kind == ast.KindParenthesizedExpression {
		site = site.Parent
	}
	if site.Parent == nil || site.Parent.Kind != ast.KindCallExpression && site.Parent.Kind != ast.KindNewExpression {
		return false
	}
	signature := lowering.checker.GetResolvedSignature(site.Parent)
	return signature != nil && librarySignature(signature)
}

// librarySignature is whether a signature is declared in the bundled library or the prelude. A declaration
// file is not the library: Node's own declarations stay counted until their host audit lands, and a .d.ts
// the program ships is the program's own (Compiler's must on dab38125).
func librarySignature(signature *checker.Signature) bool {
	declaration := signature.Declaration()
	if declaration == nil {
		return false
	}
	file := ast.GetSourceFileOfNode(declaration)
	return file != nil && (load.IsLibrary(file) || load.IsPrelude(file))
}

// anyContained relates one position inside a container: a member typed any seen as anything but any
// or unknown is the hazard, and anything else is walked in turn.
func (lowering *lowering) anyContained(
	inside, wider *checker.Type,
	depth int,
	seen map[[2]*checker.Type]bool,
) *anyRelation {
	if inside != nil &&
		wider != nil &&
		inside.Flags()&checker.TypeFlagsAny != 0 &&
		wider.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) == 0 {
		return &anyRelation{target: wider}
	}
	return lowering.anyRelated(inside, wider, false, depth+1, seen)
}
