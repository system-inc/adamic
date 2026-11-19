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
// it's viewed as there, refused the node rule 3300 judged and would have refused, hidden what that
// refusal names, the first absent optional field and the types it was found between, and fields every
// field the walk finds hidden there, hidden's first, each of which the view must check.
//
// The latent census never sees these views. Its unit lowerings never run refuse, so they record no entry,
// and a site this file views reads there as a clean conversion with no 3300 stop. That isn't proof the view
// was made: production runs refuse before every lowering (lowerWithElementKey), and only there does a
// relation cross as a checked view or stop in rule 3300's words.
type relationView struct {
	source, target *checker.Type
	refused        *ast.Node
	hidden         optionalWidening
	fields         []optionalWidening
}

// recordRelationView records found, the hiding rule's failure at node, as a relation view keyed by the node
// whose lowered value enters the slot, and says whether it did. Rule 3300 records it where it judges the
// relation, as recordSourceSlotWrite records Rule C's write where judgeViewAccess judges it, and since refuse
// runs before lowering, the entry is there before any lowering meets the value. It records only where every
// condition holds, and the rule refuses as before where one doesn't:
//
//   - the node is in a .ts file: a .a file refuses until the relation is proven (rulings.md:72);
//   - the relation is one value entering one slot (site), and a value judged against a second slot
//     (refuseInSlot, which runs while lowering) is no longer one: its entry would be viewed as one slot's
//     type alone;
//   - the source and the slot's type are open to no field the checker never names (relationViewOpen,
//     below), and the source isn't a registered Node host object (nodeHostObject);
//   - every field the walk finds hidden, found and each one after it (optionalWidenings), is one a view
//     can check (relationViewAdmits);
//   - the type viewed is one plain object type: the slot's type, or its one object member where the slot
//     adds undefined or null to it and the source holds neither (the T | undefined widening). A union of
//     objects would be legacyView's, which holds only scalar fields and no optional one, so it's refused.
//
// A failure found below the top of the relation (a field's own type) is admitted here and judged once the
// view is made: relationViewValue refuses it unless the view registered that field checked through the
// type it was found on, the nested contract viewSchema walks lazily at each depth (Language's condition 3).
func (l *lowering) recordRelationView(node *ast.Node, found *optionalWidening, site *relationSite) bool {
	if site == nil || !l.relationViewAdmits(found) {
		return false
	}
	if !strings.HasSuffix(l.program.FileName(ast.GetSourceFileOfNode(node)), ".ts") {
		return false
	}
	source := l.checker.GetTypeAtLocation(site.value)
	if l.relationViewOpen(source) || l.relationViewOpen(site.target) {
		return false
	}
	if _, host := l.nodeHostObject(source); host {
		return false
	}
	// found is the first field of this same walk, so the walk run whole names it first and every field
	// after it, each of which the view lets cross too.
	fields := l.optionalWidenings(source, site.target, site.skip)
	if len(fields) == 0 {
		return false
	}
	for index := range fields {
		if !l.relationViewAdmits(&fields[index]) {
			return false
		}
	}
	target := site.target
	if target.Flags()&checker.TypeFlagsUnion != 0 && !l.includesUndefined(source) && !l.includesNull(source) {
		target = l.checker.GetNonNullableType(target)
	}
	if target.Flags()&checker.TypeFlagsObject == 0 || l.isArrayType(target) || checker.IsTupleType(target) {
		return false
	}
	key := ast.SkipParentheses(site.value)
	if recorded, known := l.relationViews[key]; known && (recorded.source != source || recorded.target != target) {
		return false
	}
	if l.relationViews == nil {
		l.relationViews = map[*ast.Node]relationView{}
	}
	l.relationViews[key] = relationView{source: source, target: target, refused: node, hidden: *found, fields: fields}
	return true
}

// relationViewAdmits says whether a view can check field, a field the hiding rule's walk found hidden:
//
//   - it's an absent optional field, not a conflicting handle (handle) or a lib member the runtime's shape
//     doesn't hold (unstored);
//   - no type the walk passed to it has an index signature or is instantiable, a type parameter's
//     constraint included (constrained), where the value may hold the field under a key the checker never
//     names (Language's condition 1 on #0b9xqhh), and neither type it was found between may hold what
//     isn't a structural object the program declares (relationViewOpen: a class instance, object or {}, a
//     library's type);
//   - the type it was found on isn't a registered Node host object's, whose members are the host view
//     walk's to judge, as methodViewOwnedElsewhere defers to it;
//   - the field isn't a Weak<T>, and its type has no call or construct signature: a Weak field is kept as
//     a handle, not read as its target, and no checked read is built for one, nor for a callable field's
//     contract.
func (l *lowering) relationViewAdmits(field *optionalWidening) bool {
	if field.handle != nil || field.unstored != "" || field.constrained {
		return false
	}
	if l.relationViewOpen(field.source) || l.relationViewOpen(field.target) {
		return false
	}
	if _, host := l.nodeHostObject(field.source); host {
		return false
	}
	absent := l.checker.GetPropertyOfType(field.target, field.property)
	return absent != nil && !l.relationViewWeak(l.checker.GetTypeOfSymbol(absent)) && !l.callableViewContract(l.checker.GetTypeOfSymbol(absent))
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
// didn't register each absent field checked through the type it was found on, refuses in rule 3300's
// words.
//
// Each field is asserted, not only the first. viewSchema registers every field of every contract it
// reaches (interface_cast.go:332), but the hiding rule's walk reaches types no contract holds: a call
// signature's result and parameters (optional_widening.go's signature arm), and the fields below a
// contract the adapters left unsupported, which keeps its own fields only at the top (interface_cast.go:310).
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
	for _, field := range entry.fields {
		if !l.result.CheckedFields[ir.CheckedFieldKey(int(l.concrete(field.target).Id()), field.property)] {
			return nil, l.optionalWideningRefusal(entry.refused, &entry.hidden)
		}
	}
	if l.relationViewsTaken == nil {
		l.relationViewsTaken = map[*ast.Node]bool{}
	}
	l.relationViewsTaken[key] = true
	return viewed, nil
}

// checkRelationViewsTaken refuses, in rule 3300's words at the node it judged, every relation view no
// lowering took, once lowering is done: a value never lowered at all (a relation in a generic body no call
// instantiates), or lowered by a road that never passes value and doesn't refuse on its own, would
// otherwise cross with no view, its rule already skipped and its reads plain. The one such road the audit
// found (relation_views.mutants.md), a shorthand property's name, which shorthand reads as a local,
// refuses itself on every lowering, so one lowering's view can't mark a raw one taken; this pass holds
// any road found later. Before checkViewWidenedReads, which weighs every view the relations registered. Entries are
// judged in source order, so the stop is the same on every run.
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
