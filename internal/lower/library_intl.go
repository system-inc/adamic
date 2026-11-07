package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Intl.Collator lowers as far as tsc builds one (#zjndrx5 head 1, Adamic's ruling at 06:39Z): core.ts:2100's
//
//	new Intl.Collator(locale, { usage: "sort", sensitivity: "variant", numeric: true })
//
// its locale a string or undefined, and a call of the collator's compare. tsc makes its comparer
// lazily, for services and the harness alone, so no compile ever calls it, and compare isn't built
// yet: the collator is a host object holding a compare closure, bound to it by construction as
// ECMA-402 says, whose call stops on both backends in one fixed run-time NotYet
// (ir.IntlCollatorCompareStop). It never answers in ordinal order, which would be a silent miscompile
// (docs/rulings.md, localeCompare). A defined locale stops at construction (ir.IntlCollatorLocaleStop),
// since Node resolves a well-formed tag and throws a RangeError for a malformed one, and going on
// silently would answer a program that catches it differently (Runtime's must 1).
//
// Every other road to a collator is a compile-time NotYet that says what it is (intlCollatorRefusal):
// another option or value, options that aren't the literal written in place, another locale type,
// resolvedOptions, supportedLocalesOf, Intl.Collator called without new and a class extending it. So
// every collator a program makes was made here and holds the compare slot Lower's #9tw2b9z reads as a
// value (Adamic's note at 06:40Z). compare read as a value stays the method value refusal until then.
//
// Node's collator has no own property: compare is a getter on Intl.Collator.prototype, so a write to
// it throws a TypeError, and a copy or an enumeration of a collator holds nothing. TypeScript types
// compare as an assignable method, so those programs type-check, and native's shape holds compare in
// slot 0, where a write would land and go on silently. So a write to a collator's member, and
// Object.assign, entries, values, getOwnPropertyNames and hasOwn of one, are compile-time NotYets by
// name too (Runtime's must on head 1's C). A spread of one is the library spread rule's refusal
// (librarySpreadRefusal), delete of compare is tsc's own error (TS2790, compare isn't optional), and
// an in test and a run-time key read stop at run time on the host shape. A collator seen as any type
// but Intl.Collator is a NotYet at the view (intlCollatorViewRefusal), Object.keys' {} parameter among
// them, since a view could write compare.

// intlCollatorOptions is tsc's options literal as tsc writes it, in its order, each option with the
// one literal it's written with: usage "sort", sensitivity "variant" and numeric true.
var intlCollatorOptions = []struct {
	name string
	kind ast.Kind
	text string
}{
	{"usage", ast.KindStringLiteral, "sort"},
	{"sensitivity", ast.KindStringLiteral, "variant"},
	{"numeric", ast.KindTrueKeyword, ""},
}

// intlCollatorOnly is what every options NotYet says is built.
const intlCollatorOnly = `(only tsc's { usage: "sort", sensitivity: "variant", numeric: true } is built)`

// isIntlCollator reports whether node is the library's Intl.Collator, its constructor read off the
// Intl global.
func (l *lowering) isIntlCollator(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	return node.Kind == ast.KindPropertyAccessExpression && node.Name().Text() == "Collator" && l.isLibraryGlobal(node.AsPropertyAccessExpression().Expression, "Intl")
}

// isIntlCollatorType reports whether a type is the library's Intl.Collator, the interface a collator
// is, past undefined.
func (l *lowering) isIntlCollatorType(proven *checker.Type) bool {
	present := l.checker.GetNonNullableType(proven)
	if !l.isLibraryType(present, "Collator") {
		return false
	}
	parent := l.unbranded(present).Symbol().Parent
	return parent != nil && parent.Name == "Intl"
}

// intlCollatorRefusal is the refusal walk's stop for each road to a collator head 1 doesn't build,
// naming it, before anything of it lowers: new Intl.Collator other than tsc's, Intl.Collator called
// without new, a class extending it, supportedLocalesOf, a collator's resolvedOptions, a write to a
// collator's member, and an Object call that would read or copy a collator's own properties, which
// native's shape holds and Node's collator doesn't have.
func (l *lowering) intlCollatorRefusal(node *ast.Node) error {
	switch node.Kind {
	case ast.KindNewExpression:
		if l.isIntlCollator(node.AsNewExpression().Expression) {
			return l.intlCollatorArguments(node)
		}
	case ast.KindCallExpression:
		callee := ast.SkipParentheses(node.AsCallExpression().Expression)
		if l.isIntlCollator(callee) {
			return l.notYet(node, "Intl.Collator called without new")
		}
		if err := l.intlCollatorOwnProperties(node, callee); err != nil {
			return err
		}
		// Called, a member is judged at its call, which the walk meets before the member, so the call's
		// other rules can't speak for it first.
		if callee.Kind == ast.KindPropertyAccessExpression {
			return l.intlCollatorMember(callee)
		}
	case ast.KindClassDeclaration, ast.KindClassExpression:
		// At the class, which the walk meets before its heritage, so the class's other rules can't speak
		// for it first.
		for _, clause := range nodesOf(node.ClassLikeData().HeritageClauses) {
			if clause.AsHeritageClause().Token != ast.KindExtendsKeyword {
				continue
			}
			for _, base := range clause.AsHeritageClause().Types.Nodes {
				if l.isIntlCollator(base.AsExpressionWithTypeArguments().Expression) {
					return l.notYet(base, "a class extending Intl.Collator")
				}
			}
		}
	case ast.KindPropertyAccessExpression:
		if err := l.intlCollatorWrite(node, node.Name().Text()); err != nil {
			return err
		}
		return l.intlCollatorMember(node)
	case ast.KindElementAccessExpression:
		member := "a member named at run time"
		if argument := node.AsElementAccessExpression().ArgumentExpression; argument.Kind == ast.KindStringLiteral {
			member = argument.Text()
		}
		return l.intlCollatorWrite(node, member)
	}
	return nil
}

// intlCollatorOwnPropertyCalls are the Object calls that read or copy a value's own properties, which
// lowering would answer from native's shape: hasOwn reads the shape's names unguarded
// (library_object.c), and the others would copy or list its slots, compare among them, where Node's
// collator has no own property at all. Object.keys isn't one: its parameter is {}, a view of the
// collator, which intlCollatorViewRefusal stops (own_keys.a).
var intlCollatorOwnPropertyCalls = map[string]bool{"assign": true, "entries": true, "getOwnPropertyNames": true, "hasOwn": true, "values": true}

// intlCollatorOwnProperties is the stop for an Object call of intlCollatorOwnPropertyCalls handed a
// collator in any argument, Object.assign's target and sources alike, or nil for any other call.
func (l *lowering) intlCollatorOwnProperties(node, callee *ast.Node) error {
	if callee.Kind != ast.KindPropertyAccessExpression || !intlCollatorOwnPropertyCalls[callee.Name().Text()] || !l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Object") {
		return nil
	}
	for _, argument := range node.AsCallExpression().Arguments.Nodes {
		if l.isIntlCollatorType(l.checker.GetTypeAtLocation(argument)) {
			return l.notYet(node, "Object."+callee.Name().Text()+" on an Intl.Collator, which has no own property on Node")
		}
	}
	return nil
}

// intlCollatorWrite is the stop for a write to a collator's member, read off it by name or by key: on
// Node it has none of its own, and compare is a getter on Intl.Collator.prototype, so writing it throws
// a TypeError, where native would write the shape's slot and go on. It's nil for any other access.
func (l *lowering) intlCollatorWrite(node *ast.Node, member string) error {
	if !ast.IsAssignmentTarget(node) || !l.isIntlCollatorType(l.checker.GetTypeAtLocation(node.Expression())) {
		return nil
	}
	return l.notYet(node, "a write to an Intl.Collator's "+member+", which Node's collator holds only on Intl.Collator.prototype (compare there is a getter, and writing it throws a TypeError)")
}

// intlCollatorMember is the stop for Intl.Collator.supportedLocalesOf and a collator's
// resolvedOptions, read or called, or nil for any other member.
func (l *lowering) intlCollatorMember(node *ast.Node) error {
	access := node.AsPropertyAccessExpression()
	switch node.Name().Text() {
	case "supportedLocalesOf":
		if l.isIntlCollator(access.Expression) {
			return l.notYet(node, "Intl.Collator.supportedLocalesOf")
		}
	case "resolvedOptions":
		if l.isIntlCollatorType(l.checker.GetTypeAtLocation(access.Expression)) {
			return l.notYet(node, "Intl.Collator's resolvedOptions")
		}
	}
	return nil
}

func init() { refusalRules.register(150, "intlCollatorRefusal", (*lowering).intlCollatorRefusal) }

// intlCollatorView is what a collator seen as another type stops in.
const intlCollatorView = "an Intl.Collator seen as a type other than Intl.Collator (Node's collator has no own compare, and a view could write one)"

// intlCollatorViewRefusal is the stop for a collator going into a slot that sees it as another type
// (#zjndrx5 head 1, Language's must on its write stop). intlCollatorWrite judges a write by its
// receiver's type, so a collator held as { compare: F } took the write into native's slot 0 and answered
// from it, where Node throws a TypeError. A type says what a value looks like, never what reaches it, so
// the closed form is Library's: a value whose type holds an Intl.Collator goes only into a slot whose
// every member is Intl.Collator itself, unknown, any, null, undefined or a primitive, which can't hold a
// collator at all. Every other slot is a NotYet: a field view, readonly or not, since a readonly view is
// assignable to a writable one, an interface, {}, object and a type parameter. The slot is read at each
// site a value goes into one (intlCollatorSite) through slotType, so every relation site sees it,
// refuseInSlot's explicit slots too, and a value with no slot is judged against what tsc took it into,
// a conditional's reduced type, a literal's element type or a function's inferred result (impliedTarget).
func (l *lowering) intlCollatorViewRefusal(node *ast.Node) error {
	if node.Kind == ast.KindParenthesizedExpression || !l.isExpression(node) || !l.holdsIntlCollator(l.checker.GetTypeAtLocation(node)) {
		return nil
	}
	var slot *checker.Type
	if intlCollatorSite(node) {
		slot = l.slotType(node)
	}
	if slot == nil {
		slot = l.impliedTarget(node)
	}
	if slot == nil || l.intlCollatorSlot(slot) {
		return nil
	}
	return l.notYet(node, intlCollatorView)
}

func init() {
	refusalRules.register(160, "intlCollatorViewRefusal", (*lowering).intlCollatorViewRefusal)
}

// intlCollatorSite reports whether node is a value going into a slot tsc gives it: viewSite's sites, a
// shorthand property's name, an as or angle-bracket assertion's operand, whose slot is the asserted
// type, a yield's operand, the right of &&=, ||= and ??=, a binding element's default, and a
// conditional's branch or a logical operand, which take the slot their expression goes into.
// Parentheses around it are the same site.
func intlCollatorSite(node *ast.Node) bool {
	if viewSite(node) {
		return true
	}
	child := node
	for child.Parent != nil && child.Parent.Kind == ast.KindParenthesizedExpression {
		child = child.Parent
	}
	parent := child.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindShorthandPropertyAssignment:
		return parent.Name() == node
	case ast.KindAsExpression, ast.KindTypeAssertionExpression, ast.KindYieldExpression:
		return true
	case ast.KindBindingElement:
		return parent.AsBindingElement().Initializer == child
	case ast.KindConditionalExpression:
		conditional := parent.AsConditionalExpression()
		return conditional.WhenTrue == child || conditional.WhenFalse == child
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindAmpersandAmpersandEqualsToken, ast.KindBarBarEqualsToken, ast.KindQuestionQuestionEqualsToken, ast.KindAmpersandAmpersandToken, ast.KindCommaToken:
			return binary.Right == child
		case ast.KindBarBarToken, ast.KindQuestionQuestionToken:
			return true
		}
	}
	return false
}

// holdsIntlCollator reports whether a type is the library's Intl.Collator or a union holding it.
func (l *lowering) holdsIntlCollator(proven *checker.Type) bool {
	if proven.Flags()&checker.TypeFlagsUnion == 0 {
		return l.isIntlCollatorType(proven)
	}
	for _, member := range proven.Types() {
		if l.isIntlCollatorType(member) {
			return true
		}
	}
	return false
}

// intlCollatorSlot reports whether a slot shows a collator as nothing but itself: each member of it
// Intl.Collator, unknown, any, or a primitive (null and undefined among them).
func (l *lowering) intlCollatorSlot(slot *checker.Type) bool {
	members := []*checker.Type{slot}
	if slot.Flags()&checker.TypeFlagsUnion != 0 {
		members = slot.Types()
	}
	for _, member := range members {
		if member.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown|checker.TypeFlagsPrimitive) == 0 && !l.isIntlCollatorType(member) {
			return false
		}
	}
	return true
}

// intlCollatorBindingRefusal is the stop for an Intl.Collator bound to a constrained type parameter,
// written or inferred. f<T extends { compare: F }>(x: T) called with a collator binds T to
// Intl.Collator, so the argument's slot is Intl.Collator itself and intlCollatorViewRefusal lets it by,
// while the body sees x as the constraint and can write its compare. Any constraint but unknown or any
// is a NotYet, the closed form. An unconstrained parameter shows nothing to write, and a type parameter
// the library declares binds a body that's the runtime's own (constraintRelation's library exemption).
func (l *lowering) intlCollatorBindingRefusal(node *ast.Node) error {
	return l.typeArgumentBindings(node, func(binding typeArgumentBinding) error {
		if binding.argument == nil || !l.holdsIntlCollator(binding.argument) || binding.constraint.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 || l.libraryTypeParameter(binding.parameter) {
			return nil
		}
		where := binding.site
		if binding.written != nil {
			where = binding.written
		}
		return l.notYet(where, "an Intl.Collator bound to a type parameter constrained to "+l.checker.TypeToString(binding.constraint)+" (Node's collator has no own compare, and the generic's body could write one)")
	})
}

func init() {
	refusalRules.register(170, "intlCollatorBindingRefusal", (*lowering).intlCollatorBindingRefusal)
}

// intlCollatorArguments holds new Intl.Collator to tsc's two arguments: a locale typed string or
// undefined, and its options literal exactly as tsc writes it, in place, each option a plain field
// named in tsc's order with its one literal, so a literal held in a const, a spread, another order and
// a value cast to boolean are each a NotYet of their own (Language's roads at 06:41Z).
func (l *lowering) intlCollatorArguments(node *ast.Node) error {
	arguments := nodesOf(node.AsNewExpression().Arguments)
	for _, argument := range arguments {
		if argument.Kind == ast.KindSpreadElement {
			return l.notYet(argument, "new Intl.Collator with spread arguments")
		}
	}
	if len(arguments) != 2 {
		return l.notYet(node, "new Intl.Collator without exactly a locale and options "+intlCollatorOnly)
	}
	if !l.intlCollatorLocaleType(l.checker.GetTypeAtLocation(arguments[0])) {
		return l.notYet(arguments[0], "new Intl.Collator with a locale that isn't a string or undefined (a list of locales isn't built yet)")
	}
	if arguments[1].Kind != ast.KindObjectLiteralExpression {
		return l.notYet(arguments[1], "new Intl.Collator with options that aren't an object literal written in place "+intlCollatorOnly)
	}
	fields := arguments[1].AsObjectLiteralExpression().Properties.Nodes
	for index, field := range fields {
		if field.Kind != ast.KindPropertyAssignment || field.Name() == nil || !ast.IsIdentifier(field.Name()) {
			return l.notYet(field, "new Intl.Collator with an option that isn't a plain field "+intlCollatorOnly)
		}
		name := field.Name().Text()
		known := -1
		for position, option := range intlCollatorOptions {
			if option.name == name {
				known = position
			}
		}
		if known < 0 {
			return l.notYet(field, "new Intl.Collator with the option "+name+" "+intlCollatorOnly)
		}
		want, value := intlCollatorOptions[known], field.AsPropertyAssignment().Initializer
		if value.Kind != want.kind || want.kind == ast.KindStringLiteral && value.Text() != want.text {
			return l.notYet(field, "new Intl.Collator with the option "+name+": "+sourceExpression(value)+" "+intlCollatorOnly)
		}
		if known != index {
			return l.notYet(field, "new Intl.Collator with its options in another order than tsc's "+intlCollatorOnly)
		}
	}
	if len(fields) < len(intlCollatorOptions) {
		return l.notYet(arguments[1], "new Intl.Collator without the option "+intlCollatorOptions[len(fields)].name+" "+intlCollatorOnly)
	}
	return nil
}

// intlCollatorLocaleType reports whether a locale is typed as tsc's is, a string or undefined: each
// member of its type string-like or undefined.
func (l *lowering) intlCollatorLocaleType(proven *checker.Type) bool {
	members := []*checker.Type{proven}
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		members = proven.Types()
	}
	for _, member := range members {
		if member.Flags()&(checker.TypeFlagsStringLike|checker.TypeFlagsUndefined) == 0 {
			return false
		}
	}
	return true
}

// intlCollatorNew lowers tsc's new Intl.Collator, its arguments held to tsc's again, as the refusal
// walk held them: the locale, a string or undefined, makes the collator, and the options, constants
// whose every value is tsc's, lower to nothing.
func (l *lowering) intlCollatorNew(node *ast.Node) (ir.Expression, error) {
	if err := l.intlCollatorArguments(node); err != nil {
		return nil, err
	}
	locale, err := l.expression(node.AsNewExpression().Arguments.Nodes[0])
	if err != nil {
		return nil, err
	}
	locale = fit(locale, ir.String)
	if locale.Type() != ir.String {
		return nil, l.notYet(node.AsNewExpression().Arguments.Nodes[0], "new Intl.Collator with a locale not held as a string")
	}
	return ir.IntlCollatorNew{Locale: locale}, nil
}

// intlCollatorCompare lowers collator.compare(x, y), called on its collator: a call of the closure the
// collator holds in its compare slot, through the object as any method's is (ir.Property's Method),
// which stops (library_intl.c). It's false for any other call.
func (l *lowering) intlCollatorCompare(node *ast.Node) (ir.Expression, bool, error) {
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if callee.Kind != ast.KindPropertyAccessExpression || callee.Name().Text() != "compare" {
		return nil, false, nil
	}
	receiver := callee.AsPropertyAccessExpression().Expression
	if !l.isIntlCollatorType(l.checker.GetTypeAtLocation(receiver)) {
		return nil, false, nil
	}
	if node.Flags&ast.NodeFlagsOptionalChain != 0 {
		return nil, true, l.notYet(node, "Intl.Collator's compare through an optional chain; call it directly")
	}
	written := call.Arguments.Nodes
	for _, argument := range written {
		if argument.Kind == ast.KindSpreadElement {
			return nil, true, l.notYet(argument, "Intl.Collator's compare with spread arguments")
		}
	}
	if len(written) != 2 {
		return nil, true, l.notYet(node, "Intl.Collator's compare without exactly two strings")
	}
	collator, err := l.expression(receiver)
	if err != nil {
		return nil, true, err
	}
	if collator.Type() != ir.Object {
		return nil, true, l.notYet(receiver, "Intl.Collator's compare on a collator not held as an object")
	}
	arguments := []ir.Expression{}
	for _, argument := range written {
		value, err := l.expression(argument)
		if err != nil {
			return nil, true, err
		}
		if value.Type() != ir.String {
			return nil, true, l.notYet(argument, "Intl.Collator's compare of a value not held as a string")
		}
		arguments = append(arguments, value)
	}
	compare := ir.Property{Object: collator, Name: "compare", Of: ir.Closure, Method: true}
	return ir.CallClosure{Closure: compare, Arguments: arguments, FunctionType: int(l.concrete(l.checker.GetTypeAtLocation(callee)).Id()), Returns: ir.Number}, true, nil
}
