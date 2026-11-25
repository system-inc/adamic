package lower

import (
	"slices"
	"strconv"
	"strings"
	_ "unsafe"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

// Ruling 17, road (a), the dispatch cast (Adamic on #bbssdcy, family 19; this is #d1chbm7's proof, whose lowering
// is the next commit's): `const fn = (table as Record<K, F>)[x.kind]; ... fn === undefined ? ... : fn(x, ...)`, as
// tsc's forEachChild writes it, where the assertion erases each key's function to F and the call hands x to
// whatever the key found. The lookup is admitted where every condition below holds, each a function returning the
// reason it fails ("" where it holds), asked in this order so a declined candidate names the first that fails, and
// anything else keeps today's stop, refuseAnyRelation's at the assertion:
//   - the shape: the table's declared type is a mapped type with an as clause, [N in U as N["kind"]], over a union
//     whose members each hold one literal kind, so each key is one kind value and the element is read only at
//     x.kind (dispatchShape). A mapped type over an enum's keys, a Record annotation, an index signature and a
//     Partial of the ruled shape each keep the stop by name;
//   - Language's numeric-keys should: every key of the mapped type and every kind of x's declared type is a number
//     (dispatchNumericKeys), so a kind no entry names reads undefined, never a member Object gives every object
//     (her g03.ts, where "constructor" finds Object). Its support is Adamic's c13 answer: docs/rulings.md:89 keeps
//     an Object.assign that adds fields at NotYet, and a write to Object.prototype is refused, so no admitted
//     program puts a function where a lookup at a kind no entry names would find it;
//   - the proof, Adamic's never-written rule (her 19:12:02Z, condition 2 now): the table binds one name a const
//     declares and no export modifier exports (numericKeyBinding, dispatchBinding), and every reference to it by
//     the checker's symbol, in every module, is that name or a candidate lookup that passes every condition here
//     (dispatchProofReference). Passed, stored, re-exported, spread, read without the assertion, bound by a
//     separate assertion (`const view = table as Record<...>`, Language's r02) and any other reference keep the
//     stop, each by its kind, so nothing writes the table and each key still holds what tsc checked there;
//   - Adamic's function-type condition: each key's call signature is declared by a function type, which
//     strictFunctionTypes relates contravariantly, so tsc's check of the initializer proved each entry's declared
//     parameter holds its key's members. A method signature, which tsc relates both ways, keeps the stop (t04.ts,
//     `{ run(node: T): string }["run"]`) (dispatchFunctionType);
//   - Language's single-signature condition (21:16:05Z, joined at 21:33:33Z): each key's property is one function
//     type, never a union, with exactly one call signature. A union of function types, a distributive conditional's
//     among them (g05.ts), an alias or an overload set, lets an entry for one member stand at a key that holds
//     two (dispatchSingleSignature);
//   - Language's entries point (19:15:31Z) and Adamic's entries condition (21:47:33Z): the initializer is an object
//     literal whose every element is a property assignment, keyed by a numeric computed key, holding an arrow, a
//     function expression, or an identifier whose symbol's one declaration is a function declaration with a body.
//     So the lookup runs no program code, and the signature tsc related at the initializer is the one the body runs
//     under: an accessor, a method, a shorthand and a spread keep the stop by name, and so does an overloaded
//     declaration (c08.ts to c11.ts), whose overload tsc related instead of the body (dispatchEntries);
//   - no any (21:47:33Z): neither an entry's own type nor any of its declared parameters is any, whose relation to
//     the key proves nothing (c06.ts) (dispatchNoAny);
//   - the entries count (22:50:42Z): each entry declares no more parameters than its key's one call signature,
//     both counted from the checker, since tsc lets an entry with an extra parameter, optional or not, fill a slot
//     that never names it (c12-trailing.ts, c14-trailing-one-member.ts). A rest parameter, in an entry or a slot,
//     keeps the stop by its own reason (dispatchEntriesCount);
//   - Language's condition 1: x is a const or a parameter its function never assigns (kindUnionFixedBinding), read
//     by the same name for the lookup and every call, so nothing repoints it between them (r01.ts, t01.ts)
//     (dispatchSubjectFixed);
//   - Language's condition 3: kind is read once, into the local the lookup and its check share (dispatch_lower.go),
//     and x's kind is a plain data field on every member (plainDataLink) with no getter or setter the program writes
//     named kind (accessorName), judged by name over the program and failing closed, as the house judges a tag read
//     (enum_tag_flow.go's accessLocalTests). The program-wide install flag (propertyWritesOf's accessors) is for a
//     narrowing a later write can change (narrowed_nullish.go), not a tag read once; tsc's own site declined on it
//     alone, with no getter named kind anywhere (J-sgqs2np) (r03.ts) (dispatchPlainKind);
//   - the escapes, Language's and ruling 17's own pins: the lookup's value is the whole initializer of a const fn,
//     and every reference to fn is `fn === undefined` (fn on either side of ===) or the callee of a call, or an
//     optional call fn?.(...), whose first argument is x by the same name. fn.call, .apply, .bind, a spread
//     argument, storing, returning or passing fn, a read from a nested function and anything else keep the stop
//     (dispatchEscapes);
//   - the call condition (22:50:42Z) and the third clause (22:58:43Z): each call passes no more arguments than each
//     key's one call signature declares and at least its required count, and each argument after x is assignable
//     to that signature's own parameter at its position, the signature's own type parameters inferred from the
//     call as tsc infers them at a call (dispatchSlotAtCall). So the call is bounded both ways by the slot tsc
//     checked each entry against, and each entry's parameter holds what it's handed (dispatchCallCondition);
//   - the lowering's guard (#d1chbm7, commit 2b): a plain call of fn is the false arm of a conditional expression
//     whose condition is one of fn's presence tests, `fn === undefined ? A : fn(x, ...)`, parentheses aside, so a
//     kind no entry names never reaches it. An optional call answers undefined there by itself. A plain call
//     anywhere else keeps the stop (dispatchCallGuarded).
//
// Her 20:46:47Z rule is what the call hands x as: the union of every member the mapped type gives the key, the
// key's call signature's first parameter, read from the checker's mapped property for the key, never from the
// entry (dispatchKey's members). All the members under one key carry its one kind value, so the check stays one
// comparison, and for a one-member key it reads as the member.
//
// The scan runs once a lowering (dispatchCasts), the way functionConstructorRoads is the allocator road's, and
// refuseAnyRelation leaves an admitted lookup's assertion to it (dispatchCastOperand). An admitted lookup is lowered
// by dispatch_lower.go, at fn's declaration and at each of fn's uses.

// dispatchSite is one dispatch lookup the scan found: everything the conditions read, filled as they're asked, and
// whole once the lookup is admitted.
type dispatchSite struct {
	// lookup is the element access, (table as Record<K, F>)[x.kind]; assertion its as-expression; and subjectKind the
	// x.kind it's keyed by.
	lookup, assertion, subjectKind *ast.Node
	// table is the asserted name's symbol, and literal the object literal its declaration is initialized by.
	table   *ast.Symbol
	literal *ast.Node
	// subject is x's symbol.
	subject *ast.Symbol
	// callee is fn's symbol and calleeDeclaration its variable declaration, whose initializer is the lookup.
	callee            *ast.Symbol
	calleeDeclaration *ast.Node
	// presenceTests are fn's `fn === undefined` comparisons, by their binary expressions, and calls its calls.
	presenceTests []*ast.Node
	calls         []dispatchCall
	// keys are the table's entries, in the literal's order.
	keys []dispatchKey
	// kind is the local the lowering holds x's kind in, plus one, set where fn's declaration is lowered and 0 before
	// (dispatchDeclaration). Every use of fn reads it, never x's kind again.
	kind int
}

// dispatchCall is one call of fn: its call expression, and whether it's an optional call, fn?.(...).
type dispatchCall struct {
	node     *ast.Node
	optional bool
}

// dispatchKey is one entry of the table: its key's canonical number string (numericKey) and the number it spells,
// the property assignment, the function it holds (an arrow, a function expression or the function declaration an
// identifier binds), the key's one call signature as the checker's mapped property for the key gives it, and that
// signature's first parameter type, the union of every member the mapped type gives the key (nil for a signature
// with no parameter).
type dispatchKey struct {
	name      string
	value     float64
	entry     *ast.Node
	function  *ast.Node
	signature *checker.Signature
	members   *checker.Type
}

// dispatchCastRoads is the program's dispatch lookups, found once: each admitted one by its as-expression, and the
// reason each declined candidate gives, by its as-expression too. An admitted one is also found by fn's symbol
// (callees) and by each of fn's presence tests and calls (uses), where the lowering meets them.
type dispatchCastRoads struct {
	sites    map[*ast.Node]*dispatchSite
	declines map[*ast.Node]string
	callees  map[*ast.Symbol]*dispatchSite
	uses     map[*ast.Node]*dispatchSite
}

// The checker's own inference at a call, as chooseOverload runs it for a generic candidate
// (tsc/internal/checker/checker.go:9282 to :9294): an inference context over the signature's own type
// parameters, the type arguments inferred from the call's arguments and its contextual type, and the
// signature instantiated with them. These are references to the pinned checker, not a second inference engine.
//
//go:linkname dispatchInferenceContext github.com/microsoft/TypeScript/tsc/internal/checker.(*Checker).newInferenceContext
func dispatchInferenceContext(receiver *checker.Checker, typeParameters []*checker.Type, signature *checker.Signature, flags checker.InferenceFlags, compare checker.TypeComparer) *checker.InferenceContext

//go:linkname dispatchInferTypeArguments github.com/microsoft/TypeScript/tsc/internal/checker.(*Checker).inferTypeArguments
func dispatchInferTypeArguments(receiver *checker.Checker, node *ast.Node, signature *checker.Signature, arguments []*ast.Node, mode checker.CheckMode, context *checker.InferenceContext) []*checker.Type

//go:linkname dispatchSignatureInstantiation github.com/microsoft/TypeScript/tsc/internal/checker.(*Checker).getSignatureInstantiation
func dispatchSignatureInstantiation(receiver *checker.Checker, signature *checker.Signature, typeArguments []*checker.Type, javaScript bool, inferredTypeParameters []*checker.Type) *checker.Signature

// dispatchCasts walks every module once for dispatch lookups and judges each (the comment at the top). A
// candidate's own conditions are judged first, then the proof's references, which ask whether each other
// lookup of the same table passes its own: a lookup's own reference is its own to judge, by its own conditions.
func (l *lowering) dispatchCasts() *dispatchCastRoads {
	if l.dispatchCastScan != nil {
		return l.dispatchCastScan
	}
	roads := &dispatchCastRoads{sites: map[*ast.Node]*dispatchSite{}, declines: map[*ast.Node]string{}, callees: map[*ast.Symbol]*dispatchSite{}, uses: map[*ast.Node]*dispatchSite{}}
	l.dispatchCastScan = roads
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		// A program whose modules can't be ordered is one the scan can't read whole, so nothing is admitted.
		return roads
	}
	candidates := []*dispatchSite{}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if site := l.dispatchCandidate(node); site != nil {
			candidates = append(candidates, site)
		}
		node.ForEachChild(visit)
		return false
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	tableReasons, own := map[*ast.Node]string{}, map[*ast.Node]string{}
	for _, site := range candidates {
		tableReasons[site.assertion] = l.dispatchTableReason(site)
		own[site.assertion] = tableReasons[site.assertion]
		if own[site.assertion] == "" {
			own[site.assertion] = l.dispatchUseReason(site)
		}
	}
	references := map[*ast.Symbol][]*ast.Node{}
	for _, site := range candidates {
		reason := tableReasons[site.assertion]
		if reason == "" {
			if _, walked := references[site.table]; !walked {
				references[site.table] = l.dispatchTableReferences(site.table)
			}
			reason = l.dispatchProofReference(site, references[site.table], own)
		}
		if reason == "" {
			reason = own[site.assertion]
		}
		if reason != "" {
			roads.declines[site.assertion] = reason
			continue
		}
		roads.sites[site.assertion] = site
		roads.callees[site.callee] = site
		for _, test := range site.presenceTests {
			roads.uses[test] = site
		}
		for _, call := range site.calls {
			roads.uses[call.node] = site
		}
	}
	return roads
}

// dispatchCastOperand reports whether node, parentheses aside, is the operand of an admitted dispatch lookup's
// assertion: refuseAnyRelation leaves that relation to the dispatch's proof.
func (l *lowering) dispatchCastOperand(node *ast.Node) bool {
	operand := node
	for operand.Parent != nil && operand.Parent.Kind == ast.KindParenthesizedExpression {
		operand = operand.Parent
	}
	assertion := operand.Parent
	if assertion == nil || assertion.Kind != ast.KindAsExpression || assertion.AsAsExpression().Expression != operand {
		return false
	}
	_, admitted := l.dispatchCasts().sites[assertion]
	return admitted
}

// dispatchCandidate is node's dispatch site where node is a candidate, nil anywhere else: an element access, with
// no ?., whose object, parentheses aside, is an as-expression whose operand is a name, and whose key, parentheses
// aside, is x.kind, with no ?., on a name. Every other assertion is no candidate and keeps today's road.
func (l *lowering) dispatchCandidate(node *ast.Node) *dispatchSite {
	if node.Kind != ast.KindElementAccessExpression || node.AsElementAccessExpression().QuestionDotToken != nil {
		return nil
	}
	access := node.AsElementAccessExpression()
	assertion := ast.SkipParentheses(access.Expression)
	if assertion.Kind != ast.KindAsExpression || !ast.IsIdentifier(ast.SkipParentheses(assertion.AsAsExpression().Expression)) {
		return nil
	}
	key := ast.SkipParentheses(access.ArgumentExpression)
	if key.Kind != ast.KindPropertyAccessExpression || key.AsPropertyAccessExpression().QuestionDotToken != nil || key.Name().Text() != "kind" || !ast.IsIdentifier(ast.SkipParentheses(key.AsPropertyAccessExpression().Expression)) {
		return nil
	}
	return &dispatchSite{
		lookup:      node,
		assertion:   assertion,
		subjectKind: key,
		table:       l.symbol(ast.SkipParentheses(assertion.AsAsExpression().Expression)),
		subject:     l.symbol(ast.SkipParentheses(key.AsPropertyAccessExpression().Expression)),
	}
}

// dispatchTableReason is the first of the shape, the numeric keys and the proof's binding to fail, "" where all hold.
func (l *lowering) dispatchTableReason(site *dispatchSite) string {
	for _, condition := range []func(*dispatchSite) string{l.dispatchShape, l.dispatchNumericKeys, l.dispatchBinding} {
		if reason := condition(site); reason != "" {
			return reason
		}
	}
	return ""
}

// dispatchUseReason is the first of the conditions after the proof to fail, in their order, "" where all hold. Each
// may fill site as it holds: the entries its keys, the escapes its callee and uses.
func (l *lowering) dispatchUseReason(site *dispatchSite) string {
	for _, condition := range []func(*dispatchSite) string{l.dispatchFunctionType, l.dispatchSingleSignature, l.dispatchEntries, l.dispatchNoAny, l.dispatchEntriesCount, l.dispatchSubjectFixed, l.dispatchPlainKind, l.dispatchEscapes, l.dispatchCallCondition, l.dispatchCallGuarded} {
		if reason := condition(site); reason != "" {
			return reason
		}
	}
	return ""
}

// dispatchShape is the shape: the table's declared type is a mapped type whose as clause is its member's kind, over
// a union whose members each hold one literal kind. A library alias (Record, Partial), a type that isn't mapped and
// a mapped type with no as clause each name what they are.
func (l *lowering) dispatchShape(site *dispatchSite) string {
	if site.table == nil {
		return "shape: the table's name binds no symbol"
	}
	declared := l.checker.GetTypeOfSymbol(site.table)
	if alias := declared.Alias(); alias != nil && dispatchLibrarySymbol(alias.Symbol()) {
		switch alias.Symbol().Name {
		case "Record":
			return "shape: the table's declared type is a Record annotation"
		case "Partial":
			return "shape: the table's declared type is a Partial annotation"
		}
		return "shape: the table's declared type is the library's " + alias.Symbol().Name
	}
	if declared.Flags()&checker.TypeFlagsObject == 0 || declared.ObjectFlags()&checker.ObjectFlagsMapped == 0 {
		if len(l.checker.GetIndexInfosOfType(declared)) != 0 {
			return "shape: the table's declared type is an index signature"
		}
		return "shape: the table's declared type isn't a mapped type"
	}
	mapped := declared.AsMappedType()
	mapped.ResolveComponents(l.checker, declared)
	field := site.subjectKind.Name().Text()
	name := mapped.NameType()
	if name == nil {
		for _, member := range flattenedMembers(mapped.ConstraintType()) {
			if member.Flags()&checker.TypeFlagsEnumLiteral == 0 {
				return "shape: the table's declared type is a mapped type with no as clause"
			}
		}
		return "shape: the table's declared type is a mapped type over an enum's keys"
	}
	if name.Flags()&checker.TypeFlagsIndexedAccess == 0 || name.AsIndexedAccessType().ObjectType() != mapped.TypeParameter() || !dispatchStringLiteral(name.AsIndexedAccessType().IndexType(), field) {
		return "shape: the mapped type's as clause isn't its member's " + field
	}
	for _, member := range flattenedMembers(mapped.ConstraintType()) {
		if member.Flags()&checker.TypeFlagsObject == 0 {
			return "shape: a member of the mapped type's union isn't an object"
		}
		kind := l.checker.GetTypeOfPropertyOfType(member, field)
		if kind == nil || kind.Flags()&checker.TypeFlagsUnion != 0 || kind.Flags()&checker.TypeFlagsLiteral == 0 {
			return "shape: a member of the mapped type's union doesn't hold one literal " + field
		}
	}
	return ""
}

// dispatchNumericKeys is Language's numeric-keys should: each member of the mapped type's union, which names its
// key, holds a number kind, and each member of x's declared type holds only number kinds. It's asked once the shape
// holds.
func (l *lowering) dispatchNumericKeys(site *dispatchSite) string {
	declared := l.checker.GetTypeOfSymbol(site.table)
	if declared.Flags()&checker.TypeFlagsObject == 0 || declared.ObjectFlags()&checker.ObjectFlagsMapped == 0 {
		return "numeric keys: the table's declared type isn't a mapped type"
	}
	mapped := declared.AsMappedType()
	mapped.ResolveComponents(l.checker, declared)
	field := site.subjectKind.Name().Text()
	for _, member := range flattenedMembers(mapped.ConstraintType()) {
		if kind := l.checker.GetTypeOfPropertyOfType(member, field); kind == nil || kind.Flags()&checker.TypeFlagsNumberLiteral == 0 {
			return "numeric keys: a key of the table's type isn't a number"
		}
	}
	if site.subject == nil {
		return "numeric keys: x binds no symbol"
	}
	for _, member := range flattenedMembers(l.checker.GetTypeOfSymbol(site.subject)) {
		kind := l.checker.GetTypeOfPropertyOfType(member, field)
		if kind == nil {
			return "numeric keys: a member of x's type has no " + field
		}
		for _, literal := range flattenedMembers(kind) {
			if literal.Flags()&checker.TypeFlagsNumberLiteral == 0 {
				return "numeric keys: a " + field + " of x's type isn't a number"
			}
		}
	}
	return ""
}

// dispatchBinding is the proof's binding: the table is one name a const declares, initialized, and exported by no
// modifier (numericKeyBinding), so every way to the object is a reference to that name.
func (l *lowering) dispatchBinding(site *dispatchSite) string {
	declaration := site.table.ValueDeclaration
	if declaration == nil || declaration.Kind != ast.KindVariableDeclaration || declaration.Initializer() == nil {
		return "the proof: the table isn't a variable with an initializer"
	}
	if list := declaration.Parent; list == nil || list.Kind != ast.KindVariableDeclarationList || list.Flags&ast.NodeFlagsBlockScoped != ast.NodeFlagsConst {
		return "the proof: the table isn't a const"
	}
	if ast.GetCombinedModifierFlags(declaration)&ast.ModifierFlagsExport != 0 {
		return "the proof: the table is exported"
	}
	if numericKeyBinding(ast.SkipParentheses(declaration.Initializer())) != declaration {
		return "the proof: the table isn't one name a const binds"
	}
	return ""
}

// dispatchTableReferences is every reference to symbol, a table dispatchBinding holds, in every module, by the
// checker's symbol, in the order the modules and their text give them: read as a value, through a shorthand, or as
// an export specifier's local (numericKeyReference). Its declaration's name and a name in a type are none.
func (l *lowering) dispatchTableReferences(symbol *ast.Symbol) []*ast.Node {
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return nil
	}
	name := symbol.ValueDeclaration.Name()
	references := []*ast.Node{}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if ast.IsIdentifier(node) && node != name && !ast.IsPartOfTypeNode(node) && !inTypeQuery(node) && l.numericKeyReference(node, symbol) {
			references = append(references, node)
		}
		node.ForEachChild(visit)
		return false
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	return references
}

// dispatchProofReference is the proof's references for site: "" where every reference to the table is site's own
// operand or another candidate's whose own conditions hold (own), and otherwise the first other, by its kind.
func (l *lowering) dispatchProofReference(site *dispatchSite, references []*ast.Node, own map[*ast.Node]string) string {
	for _, reference := range references {
		assertion := dispatchAssertionOf(reference)
		if assertion == site.assertion {
			continue
		}
		if reason, candidate := own[assertion]; assertion != nil && candidate {
			if reason == "" {
				continue
			}
			return "the proof: the table is looked up by a dispatch that declines"
		}
		return "the proof: the table is " + dispatchUseKind(reference)
	}
	return ""
}

// dispatchAssertionOf is the as-expression reference, parentheses aside, is the operand of, or nil.
func dispatchAssertionOf(reference *ast.Node) *ast.Node {
	operand := reference
	for operand.Parent != nil && operand.Parent.Kind == ast.KindParenthesizedExpression {
		operand = operand.Parent
	}
	if parent := operand.Parent; parent != nil && parent.Kind == ast.KindAsExpression && parent.AsAsExpression().Expression == operand {
		return parent
	}
	return nil
}

// dispatchUseKind names what a reference that's no admitted use does with its value, for a reason: passed, stored,
// re-exported, spread, read without the assertion, bound by a separate assertion, and anything else by its
// parent's kind.
func dispatchUseKind(reference *ast.Node) string {
	switch reference.Parent.Kind {
	case ast.KindExportSpecifier, ast.KindExportAssignment:
		return "re-exported"
	case ast.KindShorthandPropertyAssignment:
		return "stored"
	}
	outer := reference
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	parent := outer.Parent
	switch parent.Kind {
	case ast.KindAsExpression, ast.KindTypeAssertionExpression, ast.KindSatisfiesExpression, ast.KindNonNullExpression:
		asserted := parent
		for asserted.Parent != nil && asserted.Parent.Kind == ast.KindParenthesizedExpression {
			asserted = asserted.Parent
		}
		if lookup := asserted.Parent; lookup != nil && lookup.Kind == ast.KindElementAccessExpression && lookup.Expression() == asserted {
			return "asserted for a lookup outside the dispatch's shape"
		}
		return "bound by a separate assertion"
	case ast.KindSpreadElement, ast.KindSpreadAssignment:
		return "spread"
	case ast.KindCallExpression, ast.KindNewExpression:
		if parent.Expression() == outer {
			return "called"
		}
		return "passed"
	case ast.KindElementAccessExpression, ast.KindPropertyAccessExpression:
		if parent.Expression() == outer {
			return "read without the assertion"
		}
	case ast.KindVariableDeclaration, ast.KindPropertyAssignment, ast.KindPropertyDeclaration, ast.KindArrayLiteralExpression:
		return "stored"
	case ast.KindBinaryExpression:
		if binary := parent.AsBinaryExpression(); binary.OperatorToken.Kind == ast.KindEqualsToken {
			if binary.Left == outer {
				return "written"
			}
			return "stored"
		}
	case ast.KindReturnStatement, ast.KindArrowFunction:
		return "returned"
	}
	return "used as " + strings.TrimPrefix(parent.Kind.String(), "Kind")
}

// dispatchFunctionType is Adamic's function-type condition: every call signature of every key's property, a union's
// members' included, is declared by a function type.
func (l *lowering) dispatchFunctionType(site *dispatchSite) string {
	for _, property := range l.checker.GetPropertiesOfType(l.checker.GetTypeOfSymbol(site.table)) {
		for _, member := range flattenedMembers(l.checker.GetTypeOfSymbol(property)) {
			for _, signature := range l.checker.GetSignaturesOfType(member, checker.SignatureKindCall) {
				declaration := signature.Declaration()
				if declaration != nil && declaration.Kind == ast.KindMethodSignature {
					return "function type: key " + property.Name + "'s call signature comes from a method signature"
				}
				if declaration == nil || declaration.Kind != ast.KindFunctionType {
					return "function type: key " + property.Name + "'s call signature isn't declared by a function type"
				}
			}
		}
	}
	return ""
}

// dispatchSingleSignature is Language's single-signature condition: every key's property is no union and has
// exactly one call signature.
func (l *lowering) dispatchSingleSignature(site *dispatchSite) string {
	for _, property := range l.checker.GetPropertiesOfType(l.checker.GetTypeOfSymbol(site.table)) {
		slot := l.checker.GetTypeOfSymbol(property)
		if slot.Flags()&checker.TypeFlagsUnion != 0 {
			return "single signature: key " + property.Name + "'s property is a union, not one function type"
		}
		if len(l.checker.GetSignaturesOfType(slot, checker.SignatureKindCall)) != 1 {
			return "single signature: key " + property.Name + "'s property doesn't have exactly one call signature"
		}
	}
	return ""
}

// dispatchEntries is Language's entries point with Adamic's entries condition, and fills site's literal and keys:
// every element of the initializer is a property assignment, keyed by a key numericKey names that the table's type
// has, holding an arrow, a function expression or an identifier dispatchIdentifierEntry admits.
func (l *lowering) dispatchEntries(site *dispatchSite) string {
	initializer := ast.SkipParentheses(site.table.ValueDeclaration.Initializer())
	if initializer.Kind != ast.KindObjectLiteralExpression {
		return "entries: the table's initializer isn't an object literal"
	}
	site.literal = initializer
	site.keys = nil
	declared := l.checker.GetTypeOfSymbol(site.table)
	for _, element := range initializer.AsObjectLiteralExpression().Properties.Nodes {
		switch element.Kind {
		case ast.KindPropertyAssignment:
		case ast.KindSpreadAssignment:
			return "entries: a spread element"
		case ast.KindShorthandPropertyAssignment:
			return "entries: a shorthand property"
		case ast.KindMethodDeclaration:
			return "entries: a method entry"
		case ast.KindGetAccessor, ast.KindSetAccessor:
			return "entries: an accessor entry"
		default:
			return "entries: an element that's a " + strings.TrimPrefix(element.Kind.String(), "Kind")
		}
		name, named := l.numericKey(element.Name())
		if !named {
			return "entries: an entry whose key isn't a numeric computed key"
		}
		property := l.checker.GetPropertyOfType(declared, name)
		if property == nil {
			return "entries: key " + name + " isn't a key of the table's type"
		}
		value := ast.SkipParentheses(element.Initializer())
		function := value
		switch {
		case ast.IsIdentifier(value):
			if reason := l.dispatchIdentifierEntry(value); reason != "" {
				return "entries: key " + name + "'s entry " + reason
			}
			function = l.checker.GetSymbolAtLocation(value).Declarations[0]
		case !ast.IsFunctionExpressionOrArrowFunction(value):
			return "entries: key " + name + "'s entry holds a " + strings.TrimPrefix(value.Kind.String(), "Kind")
		}
		// numericKey spells an integer from 0 to 2^53 - 1 by its digits, which parse back to it exactly.
		number, _ := strconv.ParseFloat(name, 64)
		key := dispatchKey{name: name, value: number, entry: element, function: function}
		if signatures := l.checker.GetSignaturesOfType(l.checker.GetTypeOfSymbol(property), checker.SignatureKindCall); len(signatures) == 1 {
			key.signature = signatures[0]
			if len(signatures[0].Parameters()) != 0 {
				key.members = l.checker.GetTypeAtPosition(signatures[0], 0)
			}
		}
		site.keys = append(site.keys, key)
	}
	return ""
}

// dispatchIdentifierEntry is Adamic's entries condition for an identifier entry: "" where the checker's symbol for
// it, unresolved, is no import alias and its one declaration is a function declaration with a body, so the
// signature tsc related at the initializer is the implementation's own; otherwise what it binds instead.
func (l *lowering) dispatchIdentifierEntry(value *ast.Node) string {
	symbol := l.checker.GetSymbolAtLocation(value)
	if symbol == nil {
		return "binds no symbol"
	}
	if symbol.Flags&ast.SymbolFlagsAlias != 0 {
		return "binds an import alias"
	}
	functions := 0
	for _, declaration := range symbol.Declarations {
		if declaration.Kind == ast.KindFunctionDeclaration {
			functions++
		}
	}
	if functions > 1 {
		return "binds an overloaded function declaration"
	}
	if len(symbol.Declarations) != 1 || symbol.Declarations[0].Kind != ast.KindFunctionDeclaration {
		return "binds something other than one function declaration"
	}
	if symbol.Declarations[0].Body() == nil {
		return "binds a function declaration with no body"
	}
	return ""
}

// dispatchNoAny is Adamic's no-any condition: no entry's own type is any, and no parameter its signature declares,
// its this among them, is typed any. It's asked once the entries hold.
func (l *lowering) dispatchNoAny(site *dispatchSite) string {
	for _, key := range site.keys {
		if l.checker.GetTypeAtLocation(ast.SkipParentheses(key.entry.Initializer())).Flags()&checker.TypeFlagsAny != 0 {
			return "no any: key " + key.name + "'s entry is typed any"
		}
		signature := l.checker.GetSignatureFromDeclaration(key.function)
		parameters := signature.Parameters()
		if this := signature.ThisParameter(); this != nil {
			parameters = append([]*ast.Symbol{this}, parameters...)
		}
		for _, parameter := range parameters {
			if l.checker.GetTypeOfSymbol(parameter).Flags()&checker.TypeFlagsAny != 0 {
				return "no any: key " + key.name + "'s entry has a parameter typed any"
			}
		}
	}
	return ""
}

// dispatchEntriesCount is Adamic's entries count: no rest parameter in a key's call signature or in its entry, and
// no entry declaring more parameters than its key's call signature, each counted from the checker's signature (an
// arrow's or a function expression's own, an identifier entry's declaration's). It's asked once the entries hold.
func (l *lowering) dispatchEntriesCount(site *dispatchSite) string {
	for _, key := range site.keys {
		if key.signature == nil {
			return "entries count: key " + key.name + " has no one call signature"
		}
		entry := l.checker.GetSignatureFromDeclaration(key.function)
		switch {
		case key.signature.HasRestParameter():
			return "entries count: key " + key.name + "'s call signature has a rest parameter"
		case entry.HasRestParameter():
			return "entries count: key " + key.name + "'s entry has a rest parameter"
		case len(entry.Parameters()) > len(key.signature.Parameters()):
			return "entries count: key " + key.name + "'s entry declares " + dispatchCount(len(entry.Parameters()), "parameter") + ", its call signature " + strconv.Itoa(len(key.signature.Parameters()))
		}
	}
	return ""
}

// dispatchSubjectFixed is Language's condition 1: x is a const or a parameter its function never assigns.
func (l *lowering) dispatchSubjectFixed(site *dispatchSite) string {
	if site.subject == nil || site.subject.ValueDeclaration == nil {
		return "condition 1: x binds no declaration"
	}
	if !l.kindUnionFixedBinding(site.subject) {
		return "condition 1: x is neither a const nor a never-assigned parameter"
	}
	return ""
}

// dispatchPlainKind is Language's condition 3: x's kind is a data field on every member its read can see, and no
// getter the program writes by that name may answer it.
func (l *lowering) dispatchPlainKind(site *dispatchSite) string {
	field := site.subjectKind.Name().Text()
	if !l.plainDataLink(site.subjectKind) {
		return "condition 3: x's " + field + " is an accessor or a method on some member"
	}
	if l.accessorName(field) {
		return "condition 3: the program writes a getter or setter that may be named " + field
	}
	return ""
}

// dispatchEscapes is the escapes, and fills site's callee and uses: the lookup is the whole initializer of a const
// no modifier exports, and every reference to it in its own function (or module, at the top), outside a nested
// function, is a presence test or a call whose first argument is x (dispatchCalleeUse).
func (l *lowering) dispatchEscapes(site *dispatchSite) string {
	outer := site.lookup
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	declaration := outer.Parent
	if declaration == nil || declaration.Kind != ast.KindVariableDeclaration || declaration.Initializer() != outer || !ast.IsIdentifier(declaration.Name()) {
		return "escapes: the lookup's value isn't the whole initializer of a name"
	}
	if list := declaration.Parent; list == nil || list.Kind != ast.KindVariableDeclarationList || list.Flags&ast.NodeFlagsBlockScoped != ast.NodeFlagsConst {
		return "escapes: fn isn't a const"
	}
	if ast.GetCombinedModifierFlags(declaration)&ast.ModifierFlagsExport != 0 {
		return "escapes: fn is exported"
	}
	site.callee, site.calleeDeclaration = l.symbol(declaration.Name()), declaration
	site.presenceTests, site.calls = nil, nil
	if site.callee == nil {
		return "escapes: fn binds no symbol"
	}
	container := declaration.Parent
	for container != nil && !ast.IsFunctionLike(container) && container.Kind != ast.KindSourceFile {
		container = container.Parent
	}
	if container == nil {
		return "escapes: fn is in no function or module"
	}
	reason := ""
	var visit func(node *ast.Node, nested bool) bool
	visit = func(node *ast.Node, nested bool) bool {
		if ast.IsIdentifier(node) && node != declaration.Name() && !ast.IsPartOfTypeNode(node) && !inTypeQuery(node) && l.numericKeyReference(node, site.callee) {
			reason = l.dispatchCalleeUse(site, node, nested)
			return reason != ""
		}
		inner := nested || ast.IsFunctionLike(node) || ast.IsClassLike(node)
		node.ForEachChild(func(child *ast.Node) bool { return visit(child, inner) })
		return reason != ""
	}
	container.ForEachChild(func(child *ast.Node) bool { return visit(child, false) })
	return reason
}

// dispatchCalleeUse judges one reference to fn, "" for an admitted use, which it records on site: one operand of
// === whose other operand is undefined, the global's own symbol, or the callee of a call or an optional call, with
// no spread, whose first argument is x by its symbol.
func (l *lowering) dispatchCalleeUse(site *dispatchSite, reference *ast.Node, nested bool) string {
	if nested {
		return "escapes: fn is read inside a nested function"
	}
	outer := reference
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	switch parent := outer.Parent; parent.Kind {
	case ast.KindBinaryExpression:
		comparison := parent.AsBinaryExpression()
		other := comparison.Right
		if comparison.Right == outer {
			other = comparison.Left
		}
		other = ast.SkipParentheses(other)
		if comparison.OperatorToken.Kind == ast.KindEqualsEqualsEqualsToken && ast.IsIdentifier(other) && l.symbol(other) == l.checker.GetUndefinedSymbol() {
			site.presenceTests = append(site.presenceTests, parent)
			return ""
		}
		if ast.IsEqualityOperator(comparison.OperatorToken.Kind) {
			return "escapes: fn is compared other than by === undefined"
		}
	case ast.KindCallExpression:
		call := parent.AsCallExpression()
		if call.Expression != outer {
			return "escapes: fn is passed"
		}
		for _, argument := range call.Arguments.Nodes {
			if argument.Kind == ast.KindSpreadElement {
				return "escapes: a call of fn spreads an argument"
			}
		}
		if len(call.Arguments.Nodes) == 0 {
			return "escapes: a call of fn passes no argument"
		}
		if first := ast.SkipParentheses(call.Arguments.Nodes[0]); !ast.IsIdentifier(first) || l.symbol(first) != site.subject {
			return "escapes: a call of fn whose first argument isn't x"
		}
		site.calls = append(site.calls, dispatchCall{node: parent, optional: call.QuestionDotToken != nil})
		return ""
	case ast.KindPropertyAccessExpression:
		if parent.Expression() == outer {
			switch member := parent.Name().Text(); member {
			case "call", "apply", "bind":
				return "escapes: fn." + member
			}
			return "escapes: a member of fn is read"
		}
	case ast.KindElementAccessExpression:
		if parent.Expression() == outer {
			return "escapes: a member of fn is read"
		}
	}
	return "escapes: fn is " + dispatchUseKind(reference)
}

// dispatchCallCondition is the call condition and the third clause: for each call and each key, the call passes no
// more arguments than the key's call signature declares and at least its required count, and each argument after x
// is assignable to the signature's parameter at its position once the signature's own type parameters are inferred
// from the call (dispatchSlotAtCall). It's asked once the entries count and the escapes hold.
func (l *lowering) dispatchCallCondition(site *dispatchSite) string {
	for _, call := range site.calls {
		arguments := call.node.AsCallExpression().Arguments.Nodes
		for _, key := range site.keys {
			if key.signature == nil {
				return "the call condition: key " + key.name + " has no one call signature"
			}
			if key.signature.HasRestParameter() {
				return "the call condition: key " + key.name + "'s call signature has a rest parameter"
			}
			if declared := len(key.signature.Parameters()); len(arguments) > declared {
				return "the call condition: a call passes " + dispatchCount(len(arguments), "argument") + ", key " + key.name + "'s call signature declares " + strconv.Itoa(declared)
			}
			if required := key.signature.MinArgumentCount(); len(arguments) < required {
				return "the third clause: a call passes " + dispatchCount(len(arguments), "argument") + ", key " + key.name + "'s call signature requires " + strconv.Itoa(required)
			}
			if len(arguments) < 2 {
				continue
			}
			slot := l.dispatchSlotAtCall(call.node, key.signature, arguments)
			for position := 1; position < len(arguments); position++ {
				if !l.checker.IsTypeAssignableTo(l.checker.GetTypeAtLocation(arguments[position]), l.checker.GetTypeAtPosition(slot, position)) {
					return "the call condition: a call's argument " + strconv.Itoa(position+1) + " isn't assignable to key " + key.name + "'s parameter there"
				}
			}
		}
	}
	return ""
}

// dispatchSlotAtCall is signature as this call instantiates it: unchanged where it has no type parameters of its
// own, and otherwise with them inferred from the call's arguments and contextual type by the checker's own
// inference, as chooseOverload infers a generic candidate's (an inference context, inferTypeArguments, then
// getSignatureInstantiation). Nothing here picks a type or erases one to any. The inferred type parameters
// getSignatureInstantiation takes last change only a returned function's signature, which no condition reads.
func (l *lowering) dispatchSlotAtCall(call *ast.Node, signature *checker.Signature, arguments []*ast.Node) *checker.Signature {
	if len(signature.TypeParameters()) == 0 {
		return signature
	}
	flags := checker.InferenceFlagsNone
	if ast.IsInJSFile(call) {
		flags = checker.InferenceFlagsAnyDefault
	}
	context := dispatchInferenceContext(l.checker, signature.TypeParameters(), signature, flags, nil)
	inferred := dispatchInferTypeArguments(l.checker, call, signature, arguments, checker.CheckModeNormal, context)
	return dispatchSignatureInstantiation(l.checker, signature, inferred, ast.IsInJSFile(signature.Declaration()), nil)
}

// dispatchCallGuarded is the lowering's guard: every plain call of fn, parentheses aside, is the false arm of a
// conditional expression whose condition, parentheses aside, is one of fn's presence tests, so the call runs only
// where some entry's key is x's kind, and the lowering's arms cover every kind that reaches it. An optional call
// lowers its own undefined for a kind no entry names, so it needs no guard. It's asked once the escapes hold.
func (l *lowering) dispatchCallGuarded(site *dispatchSite) string {
	for _, call := range site.calls {
		if call.optional {
			continue
		}
		outer := call.node
		for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
			outer = outer.Parent
		}
		conditional := outer.Parent
		if conditional == nil || conditional.Kind != ast.KindConditionalExpression || conditional.AsConditionalExpression().WhenFalse != outer || !slices.Contains(site.presenceTests, ast.SkipParentheses(conditional.AsConditionalExpression().Condition)) {
			return "escapes: a call of fn no presence test guards"
		}
	}
	return ""
}

// dispatchLibrarySymbol is whether symbol is declared in the bundled library or the prelude, as librarySignature
// reads a signature's.
func dispatchLibrarySymbol(symbol *ast.Symbol) bool {
	if symbol == nil || len(symbol.Declarations) == 0 {
		return false
	}
	file := ast.GetSourceFileOfNode(symbol.Declarations[0])
	return file != nil && (load.IsLibrary(file) || load.IsPrelude(file))
}

// dispatchStringLiteral is whether literal is the string literal type text spells.
func dispatchStringLiteral(literal *checker.Type, text string) bool {
	if literal == nil || literal.Flags()&checker.TypeFlagsStringLiteral == 0 {
		return false
	}
	value, spelled := literal.AsLiteralType().Value().(string)
	return spelled && value == text
}

// dispatchCount is count with noun, plural past one.
func dispatchCount(count int, noun string) string {
	if count == 1 {
		return "1 " + noun
	}
	return strconv.Itoa(count) + " " + noun + "s"
}
