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
// it's viewed as there (relationViewPair: for a container both are arrays, and where both sides are
// nullable both are the object arms), refused the node rule 3300 judged and would have refused, hidden what that
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
//   - a call's argument is judged against a slot that holds no type parameter at any depth
//     (relationSlotGeneric, below);
//   - the relation is one of the shapes relationViewPair admits, each of whose pairs is open to no field
//     the checker never names (relationViewOpen, below) and has no registered Node host object for its
//     source (nodeHostObject);
//   - every field the walk finds hidden, found and each one after it (optionalWidenings), is one a view
//     can check (relationViewAdmits).
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
	// A call's argument is judged against the slot slotType gives it, the checker's contextual type, which
	// for an argument is the call's resolved signature's instantiated parameter at that position
	// (getContextualTypeForArgumentAtIndex reads getResolvedSignature, then getTypeAtPosition), so inference
	// from every argument is in it and the callee's declaration isn't. The view stands only where that slot
	// holds no type parameter: in a generic body the resolved signature still names the body's own type
	// parameters, and a relation there is refused, never viewed as one instantiation's type under the type
	// mapper (Lower's hook 2 on #0b9xqhh).
	if relationArgument(site.value) && l.relationSlotGeneric(site.target, map[*checker.Type]bool{}) {
		return false
	}
	source := l.checker.GetTypeAtLocation(site.value)
	viewedSource, target, admitted := l.relationViewPair(source, site.target)
	if !admitted {
		return false
	}
	// found is the first field of this same walk, so the walk run whole names it first and every field
	// after it, each of which the view lets cross too. For an array the walk goes through the element pair.
	fields := l.optionalWidenings(source, site.target, site.skip)
	if len(fields) == 0 {
		return false
	}
	for index := range fields {
		if !l.relationViewAdmits(&fields[index]) {
			return false
		}
	}
	key := ast.SkipParentheses(site.value)
	if recorded, known := l.relationViews[key]; known && (recorded.source != viewedSource || recorded.target != target) {
		return false
	}
	if l.relationViews == nil {
		l.relationViews = map[*ast.Node]relationView{}
	}
	l.relationViews[key] = relationView{source: viewedSource, target: target, refused: node, hidden: *found, fields: fields}
	return true
}

// relationViewPair is the pair a value of type source entering a slot of type slot is viewed as, and says
// whether the relation is one of the shapes rulings.md:72 admits:
//
//   - one value into one plain object type, the slot's type, or its one object member where the slot adds
//     undefined or null to it and the source holds neither (the T | undefined widening). A union of
//     objects would be legacyView's, which holds only scalar fields and no optional one, so it's refused;
//   - nullable on both sides: each side is one plain object type with undefined or null beside it, and
//     both are held as one nullable pointer (l.representation says Object for each), so a value crosses
//     with no conversion. The view is the two object arms', and the nullish arm crosses unchanged. A side
//     boxed as a union (T | null | undefined) is another representation, and refused;
//   - a container (relationViewContainer, below).
//
// A pair is open to no field the checker never names and to no value that isn't a structural object the
// program declares (relationViewOpen), and its source isn't a registered Node host object.
func (l *lowering) relationViewPair(source, slot *checker.Type) (*checker.Type, *checker.Type, bool) {
	if l.isArrayType(source) || l.isArrayType(slot) || checker.IsTupleType(source) || checker.IsTupleType(slot) {
		return l.relationViewContainer(source, slot)
	}
	if l.relationViewOpen(source) || l.relationViewOpen(slot) {
		return nil, nil, false
	}
	if _, host := l.nodeHostObject(source); host {
		return nil, nil, false
	}
	target := slot
	if slot.Flags()&checker.TypeFlagsUnion != 0 {
		if l.includesUndefined(source) || l.includesNull(source) {
			if !l.relationViewNullablePointer(source) || !l.relationViewNullablePointer(slot) {
				return nil, nil, false
			}
			source = l.checker.GetNonNullableType(source)
		}
		target = l.checker.GetNonNullableType(slot)
	}
	if target.Flags()&checker.TypeFlagsObject == 0 || l.isArrayType(target) || checker.IsTupleType(target) {
		return nil, nil, false
	}
	return source, target, true
}

// relationViewNullablePointer says whether a type is one plain object type with undefined or null beside
// it, held as one nullable pointer.
func (l *lowering) relationViewNullablePointer(nullable *checker.Type) bool {
	present := l.checker.GetNonNullableType(nullable)
	if present == nullable || present.Flags()&checker.TypeFlagsObject == 0 || l.isArrayType(present) || checker.IsTupleType(present) {
		return false
	}
	of, known := l.representation(nullable)
	return known && of == ir.Object
}

// relationViewContainer is the pair an array entering a readonly array slot is viewed as: the arrays
// themselves, whose view (view's) builds the array's contract with its element's inside, so the element
// pair registers its checked fields at the element type's id as any viewed pair does, with no entry of
// its own for the container (Lower's condition 1 on #0b9xqhh). It admits only:
//
//   - an array or readonly array source, never a tuple, going into readonly T[] or ReadonlyArray<T>. A
//     mutable target array keeps the refusal: a T written into storage typed S[] is the invariant-mutable
//     hazard, whatever tsc accepts (condition 3);
//   - one representation on both sides, element storage included (condition 4): both arrays held as
//     arrays, each with the program's plain layout (no NodeArray layout and no phantom brand, whose storage
//     this doesn't read), and both elements held as objects;
//   - an element target that's one plain object type, and an element pair open to nothing
//     (relationViewOpen) whose source isn't a registered Node host object.
func (l *lowering) relationViewContainer(source, slot *checker.Type) (*checker.Type, *checker.Type, bool) {
	if !l.isArrayType(source) || !l.isArrayType(slot) || checker.IsTupleType(source) || checker.IsTupleType(slot) || !l.isLibraryType(slot, "ReadonlyArray") {
		return nil, nil, false
	}
	if l.unbranded(source) != source || l.unbranded(slot) != slot {
		return nil, nil, false
	}
	if _, _, layout := l.nodeArrayLayoutOf(source); layout {
		return nil, nil, false
	}
	if _, _, layout := l.nodeArrayLayoutOf(slot); layout {
		return nil, nil, false
	}
	held, seen := l.elementTypeOfArray(source), l.elementTypeOfArray(slot)
	if held == nil || seen == nil || seen.Flags()&checker.TypeFlagsObject == 0 || l.isArrayType(seen) || checker.IsTupleType(seen) {
		return nil, nil, false
	}
	if l.relationViewOpen(held) || l.relationViewOpen(seen) {
		return nil, nil, false
	}
	if _, host := l.nodeHostObject(held); host {
		return nil, nil, false
	}
	array, arrayKnown := l.representation(source)
	slotArray, slotKnown := l.representation(slot)
	element, elementKnown := l.representation(held)
	slotElement, slotElementKnown := l.representation(seen)
	if !arrayKnown || !slotKnown || !elementKnown || !slotElementKnown || array != ir.Array || slotArray != ir.Array || element != ir.Object || slotElement != ir.Object {
		return nil, nil, false
	}
	return source, slot, true
}

// relationArgument says whether value, its parentheses aside, is an argument of a call or a new
// expression.
func relationArgument(value *ast.Node) bool {
	child := value
	for child.Parent != nil && child.Parent.Kind == ast.KindParenthesizedExpression {
		child = child.Parent
	}
	switch parent := child.Parent; {
	case parent == nil:
		return false
	case parent.Kind == ast.KindCallExpression:
		return parent.AsCallExpression().Expression != child
	case parent.Kind == ast.KindNewExpression:
		return parent.AsNewExpression().Expression != child
	}
	return false
}

// relationSlotGeneric says whether a slot's type holds a free type parameter at any depth: itself, a union
// or intersection member, an alias's or a reference's type argument, or, for an anonymous object type (a
// type literal or a function type, which a generic body instantiates with its own parameters), a
// property's type, an index signature's value, or a call or construct signature's parameter or result. A
// signature's own type parameters are bound inside it, so they aren't free there. An interface's or a
// class's members name only its own type parameters, bound by its type arguments, and an outer one a
// local declaration captures is one of those arguments too, so a reference is judged by its arguments.
// Any other instantiable type (an indexed access, a conditional, keyof), and a mapped type, holds one,
// failing closed.
func (l *lowering) relationSlotGeneric(slot *checker.Type, visited map[*checker.Type]bool) bool {
	return l.relationSlotHolds(slot, map[*checker.Type]bool{}, visited)
}

// relationSlotHolds is relationSlotGeneric's walk, with bound the type parameters the signatures it's
// inside declare. visited is shared across bound sets, so a type met first inside a signature and again
// outside it answers from the first meeting. That can't change an answer: a type that holds one of the
// signature's own parameters exists only inside that signature (another instantiation is another type),
// and a type that holds none answers the same under either bound set (Lower's should on #0b9xqhh).
func (l *lowering) relationSlotHolds(slot *checker.Type, bound, visited map[*checker.Type]bool) bool {
	if slot == nil {
		return false
	}
	if slot.Flags()&checker.TypeFlagsTypeParameter != 0 {
		return !bound[slot]
	}
	if slot.Flags()&checker.TypeFlagsInstantiable != 0 || slot.ObjectFlags()&checker.ObjectFlagsMapped != 0 {
		return true
	}
	if visited[slot] {
		return false
	}
	visited[slot] = true
	var parts []*checker.Type
	if slot.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		parts = append(parts, slot.Types()...)
	}
	if alias := slot.Alias(); alias != nil {
		parts = append(parts, alias.TypeArguments()...)
	}
	if slot.Flags()&checker.TypeFlagsObject != 0 {
		if slot.ObjectFlags()&checker.ObjectFlagsReference != 0 {
			parts = append(parts, l.checker.GetTypeArguments(slot)...)
		}
		if slot.ObjectFlags()&checker.ObjectFlagsAnonymous != 0 {
			for _, property := range l.checker.GetPropertiesOfType(slot) {
				parts = append(parts, l.checker.GetTypeOfSymbol(property))
			}
			for _, info := range l.checker.GetIndexInfosOfType(slot) {
				parts = append(parts, info.ValueType())
			}
			for _, kind := range []checker.SignatureKind{checker.SignatureKindCall, checker.SignatureKindConstruct} {
				for _, signature := range l.checker.GetSignaturesOfType(slot, kind) {
					inside := bound
					if declared := signature.TypeParameters(); len(declared) != 0 {
						inside = map[*checker.Type]bool{}
						for parameter := range bound {
							inside[parameter] = true
						}
						for _, parameter := range declared {
							inside[parameter] = true
						}
					}
					for _, parameter := range signature.Parameters() {
						if l.relationSlotHolds(l.checker.GetTypeOfSymbol(parameter), inside, visited) {
							return true
						}
					}
					if l.relationSlotHolds(l.checker.GetReturnTypeOfSignature(signature), inside, visited) {
						return true
					}
				}
			}
		}
	}
	for _, part := range parts {
		if l.relationSlotHolds(part, bound, visited) {
			return true
		}
	}
	return false
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
// value and before expression's graphAllocation. A value held other than as an object (an array, for a
// container's view), or a view that didn't register each absent field checked through the type it was
// found on, refuses in rule 3300's words. A container's view is kept among relationContainerOrigins, the
// values whose elements a whole-object reader stops on (checkViewArrays).
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
	container := l.isArrayType(entry.target)
	if held := value.Type(); held != ir.Object && !(container && held == ir.Array) {
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
	if container {
		l.relationContainerOrigins = append(l.relationContainerOrigins, value)
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
