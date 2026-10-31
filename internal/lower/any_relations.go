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
func (l *lowering) refuseAnyRelation(node *ast.Node) error {
	name, source, target := l.relationSite(node)
	if target == nil || source == nil || source.Flags()&checker.TypeFlagsAny != 0 {
		return nil
	}
	found := l.anyRelated(source, target, 0, map[[2]*checker.Type]bool{})
	if found == nil {
		return nil
	}
	// The lowest-priority stop: another rule's refusal anywhere in the module is the stronger reason, as
	// Array.prototype.fill's method value and Object.entries over a graph record already are, so this one
	// is kept for refuse to return only when nothing else refused.
	if l.refusal.deferred == nil {
		if found.depthLimit {
			l.refusal.deferred = l.notYet(node, "a structural relation deeper than 64 type pairs while checking members typed any")
		} else {
			l.refusal.deferred = l.notYet(node, "a member typed any ("+found.path+") related to "+l.checker.TypeToString(found.target)+" at the "+name+" site")
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
func (l *lowering) anyRelated(source, target *checker.Type, depth int, seen map[[2]*checker.Type]bool) *anyRelation {
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
			if found := l.anyRelated(member, target, depth+1, seen); found != nil {
				return found
			}
		}
		return nil
	}
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range target.Types() {
			if found := l.anyRelated(source, member, depth+1, seen); found != nil {
				return found
			}
		}
		return nil
	}
	if source.Flags()&checker.TypeFlagsTypeParameter != 0 {
		return l.anyRelated(l.checker.GetBaseConstraintOfType(source), target, depth+1, seen)
	}
	if target.Flags()&checker.TypeFlagsTypeParameter != 0 {
		return l.anyRelated(source, l.checker.GetBaseConstraintOfType(target), depth+1, seen)
	}
	if source.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) == 0 || target.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) == 0 {
		return nil
	}
	if (l.checker.IsArrayType(source) || checker.IsTupleType(source)) && (l.checker.IsArrayType(target) || checker.IsTupleType(target)) {
		if checker.IsTupleType(source) && checker.IsTupleType(target) {
			sources, targets := l.checker.GetTypeArguments(source), l.checker.GetTypeArguments(target)
			for index, inside := range sources {
				if index < len(targets) {
					if found := l.anyContained(inside, targets[index], depth, seen); found != nil {
						found.path = "[" + strconv.Itoa(index) + "]" + found.path
						return found
					}
				}
			}
			return nil
		}
		if found := l.anyContained(l.checker.GetNumberIndexType(source), l.checker.GetNumberIndexType(target), depth, seen); found != nil {
			found.path = "[]" + found.path
			return found
		}
		return nil
	}
	for _, kind := range []checker.SignatureKind{checker.SignatureKindCall, checker.SignatureKindConstruct} {
		if found := l.anyRelatedSignatures(source, target, kind, depth, seen); found != nil {
			return found
		}
	}
	if source.ObjectFlags()&checker.ObjectFlagsReference != 0 && target.ObjectFlags()&checker.ObjectFlagsReference != 0 && source.Target() != nil && source.Target() == target.Target() {
		sources, targets := l.checker.GetTypeArguments(source), l.checker.GetTypeArguments(target)
		for index, inside := range sources {
			if index < len(targets) {
				if found := l.anyContained(inside, targets[index], depth, seen); found != nil {
					found.path = "<" + strconv.Itoa(index+1) + ">" + found.path
					return found
				}
			}
		}
	}
	for _, property := range l.checker.GetPropertiesOfType(target) {
		from := l.checker.GetTypeOfPropertyOfType(source, property.Name)
		to := l.checker.GetTypeOfPropertyOfType(target, property.Name)
		if found := l.anyContained(from, to, depth, seen); found != nil {
			step := property.Name
			if strings.HasPrefix(property.Name, ast.InternalSymbolNamePrefix) {
				// A well-known symbol's key is internal; name it as the source spells it.
				step = l.checker.SymbolToString(property)
			}
			if found.path != "" && !strings.HasPrefix(found.path, "[") && !strings.HasPrefix(found.path, "<") && !strings.HasPrefix(found.path, "(") {
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
// the library's own modelling of what flows in, and no value the program holds passes through it.
func (l *lowering) anyRelatedSignatures(source, target *checker.Type, kind checker.SignatureKind, depth int, seen map[[2]*checker.Type]bool) *anyRelation {
	sources := l.checker.GetSignaturesOfType(source, kind)
	targets := l.checker.GetSignaturesOfType(target, kind)
	if len(sources) == 0 || len(targets) == 0 {
		return nil
	}
	for _, to := range targets {
		for _, from := range sources {
			to, from := checker.Checker_getCanonicalSignature(l.checker, to), checker.Checker_getCanonicalSignature(l.checker, from)
			if found := l.anyContained(l.checker.GetReturnTypeOfSignature(from), l.checker.GetReturnTypeOfSignature(to), depth, seen); found != nil {
				found.path = "()" + found.path
				return found
			}
			if librarySignature(from) && librarySignature(to) {
				continue
			}
			for position := range anyRelatedPositions(from, to) {
				if found := l.anyContained(l.checker.GetTypeAtPosition(to, position), l.checker.GetTypeAtPosition(from, position), depth, seen); found != nil {
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

// librarySignature is whether a signature is declared in the library, the prelude or a declaration file.
func librarySignature(signature *checker.Signature) bool {
	declaration := signature.Declaration()
	if declaration == nil {
		return false
	}
	file := ast.GetSourceFileOfNode(declaration)
	return file != nil && (file.IsDeclarationFile || load.IsLibrary(file) || load.IsPrelude(file))
}

// anyContained relates one position inside a container: a member typed any seen as anything but any
// or unknown is the hazard, and anything else is walked in turn.
func (l *lowering) anyContained(inside, wider *checker.Type, depth int, seen map[[2]*checker.Type]bool) *anyRelation {
	if inside != nil && wider != nil && inside.Flags()&checker.TypeFlagsAny != 0 && wider.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) == 0 {
		return &anyRelation{target: wider}
	}
	return l.anyRelated(inside, wider, depth+1, seen)
}
