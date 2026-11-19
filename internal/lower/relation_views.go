package lower

import (
	"sort"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// relationView is a relation rulings.md:72 lets cross as an untagged checked view of its slot's type
// instead of refusing it (#0b9xqhh): source is the type of the value entering the slot, target the type
// it's viewed as there, refused the node rule 3300 judged and would have refused, and hidden what that
// refusal names, the absent optional field and the types it was found between.
//
// The latent census never sees these views. Its unit lowerings never run refuse, so they record no entry,
// and a site this file views reads there as a clean conversion with no 3300 stop. That isn't proof the view
// was made: production runs refuse before every lowering (lowerWithElementKey), and only there does a
// relation cross as a checked view or stop in rule 3300's words.
type relationView struct {
	source, target *checker.Type
	refused        *ast.Node
	hidden         optionalWidening
}

// recordRelationView records found, the hiding rule's failure at node, as a relation view keyed by the node
// whose lowered value enters the slot, and says whether it did. Rule 3300 records it where it judges the
// relation, as recordSourceSlotWrite records Rule C's write where judgeViewAccess judges it, and since refuse
// runs before lowering, the entry is there before any lowering meets the value. It records only where every
// condition holds, and the rule refuses as before where one doesn't:
//
//   - the node is in a .ts file: a .a file refuses until the relation is proven (rulings.md:72);
//   - the failure is an absent optional field, not a conflicting handle (found.handle) or a lib member
//     the runtime's shape doesn't hold (found.unstored), and the relation is one value entering one slot
//     (site);
//   - no type the walk passed has an index signature or is instantiable, a type parameter's constraint
//     included, where the value may hold the field under a key the checker never names (Language's
//     condition 1 on #0b9xqhh), and none may hold what isn't a structural object the program declares
//     (relationViewOpen: a class instance, object or {}, a library's type);
//   - the source isn't a registered Node host object (nodeHostObject), whose members are the host view
//     walk's to judge, as methodViewOwnedElsewhere defers to it;
//   - the absent field isn't a Weak<T>, and its type has no call or construct signature: a view can't
//     check a callable contract today;
//   - the type viewed is one plain object type: the slot's type, or its one object member where the slot
//     adds undefined or null to it and the source holds neither (the T | undefined widening). A union of
//     objects would be legacyView's, which holds only scalar fields and no optional one, so it's refused.
//
// A failure found below the top of the relation (a field's own type) is admitted here and judged once the
// view is made: relationViewValue refuses it unless the view registered that field checked through the
// type it was found on, the nested contract viewSchema walks lazily at each depth (Language's condition 3).
func (l *lowering) recordRelationView(node *ast.Node, found *optionalWidening, site *relationSite) bool {
	if site == nil || found.handle != nil || found.unstored != "" || found.constrained {
		return false
	}
	if !strings.HasSuffix(l.program.FileName(ast.GetSourceFileOfNode(node)), ".ts") {
		return false
	}
	source := l.checker.GetTypeAtLocation(site.value)
	if l.relationViewOpen(source) || l.relationViewOpen(found.source) || l.relationViewOpen(found.target) || l.relationViewOpen(site.target) {
		return false
	}
	if _, host := l.nodeHostObject(source); host {
		return false
	}
	if _, host := l.nodeHostObject(found.source); host {
		return false
	}
	// A Weak field is kept as a handle, not read as its target, and no checked read is built for one, nor
	// for a callable field's contract.
	if absent := l.checker.GetPropertyOfType(found.target, found.property); absent == nil || l.relationViewWeak(l.checker.GetTypeOfSymbol(absent)) || l.callableViewContract(l.checker.GetTypeOfSymbol(absent)) {
		return false
	}
	target := site.target
	if target.Flags()&checker.TypeFlagsUnion != 0 && !l.includesUndefined(source) && !l.includesNull(source) {
		target = l.checker.GetNonNullableType(target)
	}
	if target.Flags()&checker.TypeFlagsObject == 0 || l.isArrayType(target) || checker.IsTupleType(target) {
		return false
	}
	if l.relationViews == nil {
		l.relationViews = map[*ast.Node]relationView{}
	}
	l.relationViews[ast.SkipParentheses(site.value)] = relationView{source: source, target: target, refused: node, hidden: *found}
	return true
}

// relationViewOpen says whether a type may hold a field under a key the checker never names, or may hold a
// value that isn't a structural object the program declares: it, or a member of its union or intersection,
// has an index signature, is instantiable, is a class instance, is object or {} (which hold any object, a
// record, a host value or a primitive), or is a type only a library declares (a host value's, read by the
// runtime's own shape, #sn788gx).
func (l *lowering) relationViewOpen(proven *checker.Type) bool {
	if proven == nil {
		return true
	}
	if proven.Flags()&(checker.TypeFlagsInstantiable|checker.TypeFlagsNonPrimitive) != 0 || isClassInstance(proven) || len(l.checker.GetIndexInfosOfType(proven)) != 0 {
		return true
	}
	if proven.Flags()&checker.TypeFlagsObject != 0 && len(l.checker.GetPropertiesOfType(proven)) == 0 && len(l.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)) == 0 && len(l.checker.GetSignaturesOfType(proven, checker.SignatureKindConstruct)) == 0 {
		return true
	}
	if symbol := proven.Symbol(); symbol != nil && len(symbol.Declarations) != 0 && !programDeclared(symbol) {
		return true
	}
	if proven.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		for _, member := range proven.Types() {
			if l.relationViewOpen(member) {
				return true
			}
		}
	}
	return false
}

// relationViewWeak says whether a field's type is a Weak<T>, or a union holding one.
func (l *lowering) relationViewWeak(declared *checker.Type) bool {
	if l.weakTarget(declared) != nil {
		return true
	}
	if declared.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range declared.Types() {
			if l.weakTarget(member) != nil {
				return true
			}
		}
	}
	return false
}

// relationViewValue views value, node's lowered value, as its relation view's type when rule 3300 recorded
// one for node, and gives back value otherwise. value calls it each time it lowers node, which may be more
// than once (a generic body for each instantiation, a program region's discovery lowering and the real
// one, a speculative copy), so every lowering views it: the entry is never deleted, and relationViewsTaken
// marks that some lowering did. The view is view's, as structuralViewCast's is, without the cast's
// requirement that the target be assignable to the source; it sits where a cast's view sits, inside
// value and before expression's graphAllocation. A value held other than as an object, or a view that
// didn't register the absent field checked through the type it was found on, refuses in rule 3300's words.
func (l *lowering) relationViewValue(node *ast.Node, value ir.Expression) (ir.Expression, error) {
	key := ast.SkipParentheses(node)
	entry, recorded := l.relationViews[key]
	if !recorded {
		return value, nil
	}
	if value.Type() != ir.Object {
		return nil, l.optionalWideningRefusal(entry.refused, &entry.hidden)
	}
	viewed, err := l.view(node, value, l.concrete(entry.source), l.concrete(entry.target))
	if err != nil {
		return nil, err
	}
	if !l.result.CheckedFields[ir.CheckedFieldKey(int(l.concrete(entry.hidden.target).Id()), entry.hidden.property)] {
		return nil, l.optionalWideningRefusal(entry.refused, &entry.hidden)
	}
	if l.relationViewsTaken == nil {
		l.relationViewsTaken = map[*ast.Node]bool{}
	}
	l.relationViewsTaken[key] = true
	return viewed, nil
}

// checkRelationViewsTaken refuses, in rule 3300's words at the node it judged, every relation view no
// lowering took, once lowering is done: a value lowered by a path that never passes value (a shorthand
// property's name, which shorthand reads as a local) or never lowered at all would otherwise cross with no
// view, its rule already skipped and its reads plain. Before checkViewWidenedReads, which weighs every
// view the relations registered. Entries are judged in source order, so the stop is the same on every run.
func (l *lowering) checkRelationViewsTaken() error {
	var untaken []*ast.Node
	for key := range l.relationViews {
		if !l.relationViewsTaken[key] {
			untaken = append(untaken, key)
		}
	}
	if len(untaken) == 0 {
		return nil
	}
	sort.Slice(untaken, func(i, j int) bool {
		left, right := l.relationViews[untaken[i]].refused, l.relationViews[untaken[j]].refused
		leftFile, rightFile := l.program.FileName(ast.GetSourceFileOfNode(left)), l.program.FileName(ast.GetSourceFileOfNode(right))
		if leftFile != rightFile {
			return leftFile < rightFile
		}
		return left.Pos() < right.Pos()
	})
	entry := l.relationViews[untaken[0]]
	return l.optionalWideningRefusal(entry.refused, &entry.hidden)
}
