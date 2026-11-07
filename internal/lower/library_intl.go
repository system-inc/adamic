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
// every collator a program makes was made here and holds the #compare slot. compare read as a value
// lowers only off the new expression itself (intlCollatorCompareValue, #9tw2b9z), and stays the method
// value refusal everywhere else.
//
// Node's collator has no own property: compare is a getter on Intl.Collator.prototype, so a write to
// it throws a TypeError, and a copy or an enumeration of a collator holds nothing. TypeScript types
// compare as an assignable method, so those programs type-check. Native's shape holds no field named
// compare either (Runtime's ruling at 16:11Z, the allocation form): compare is its private #compare
// slot, so no road by compare's name reaches it, and the JavaScript stand-in has compare as a getter on
// its prototype. A write by the collator's own type would still go nowhere Node's goes, so a write to
// a collator's member, and
// Object.assign, entries, values, getOwnPropertyNames and hasOwn of one, are compile-time NotYets by
// name too (Runtime's must on head 1's C). A spread of one is the library spread rule's refusal
// (librarySpreadRefusal), delete of compare is tsc's own error (TS2790, compare isn't optional), and
// Object.keys, an in test and a run-time key read stop at run time on the host shape.

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
// collator has no own property at all. Object.keys isn't one: it stops at run time on the host shape
// (#0sps03r's keyless form, own_keys.a).
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
// a TypeError, where native would go on without it. It's nil for any other access.
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

// intlCollatorCompare lowers collator.compare(x, y), called on its collator: a call through the object
// as any method's is (ir.Property's Method), whose callee is found at run time by the object itself,
// never by its declared type. The host collator's is the closure in its #compare slot, which stops
// (library_intl.c): native finds it by the collator's shape (object.c's adamic_object_callee), and the
// JavaScript stand-in by its prototype's getter. A program's own object typed Intl.Collator, a class
// implementing it or a literal, calls its own compare there. It's false for any other call.
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

// A collator's compare read as a value, not called, lowers to the closure in its compare slot (#9tw2b9z,
// F19 of the core.ts:2116 wall, on #zjndrx5's head 1). tsc's comparer is that read:
//
//	const comparer = new Intl.Collator(locale, { usage: "sort", sensitivity: "variant", numeric: true }).compare;
//
// The method value refusal stops a method read as a value because it loses its object, and this one
// loses nothing. Every collator a program holds was made by intlCollatorNew, since head 1 makes every
// other road a NotYet, so the slot is there. Its closure is bound by construction, capturing a cell and
// never the collator (library_intl.c, and the JavaScript backend's AdamicIntlCollator), so a read of it
// is the whole function with no receiver to lose. And it's one slot, so every read of one collator's
// compare is the one closure, === to itself as Node's compare is, whose getter caches its bound
// function, while two collators' compares differ, as Node's do.
//
// Only a read into a slot that sees it exactly as `(x: string, y: string) => number` is admitted, as
// Date.now's is (exactFunctionSlot), so a wider slot such as `(x: string, y: string) => unknown` or
// Function stays the method value refusal. tsc's const has no annotation, so its slot is the const
// itself, which the checker types as the read's own type: the whole initializer of a const named by an
// identifier with no annotation sees it exactly (intlCollatorCompareConst). That's this rule's alone,
// and Date.now's read with no slot stays refused. An === operand isn't a slot, so `c.compare ===
// c.compare` stays refused, as `f === Date.now` is, and identity is read through consts. Every other
// Intl member read as a value stays refused, and resolvedOptions stays intlCollatorMember's NotYet.

// intlCollatorCompareRead reports whether node reads a collator's compare as a value the slot read
// admits: compare, not called and not through an optional chain, on a receiver typed as the library's
// Intl.Collator, into a slot that sees it exactly as itself. methodValueRefusal lets it through
// (libraryMethodReadAllowed), and intlCollatorCompareValue lowers it, so the two never disagree.
func (l *lowering) intlCollatorCompareRead(node *ast.Node) bool {
	if node.Kind != ast.KindPropertyAccessExpression || node.Name().Text() != "compare" || called(node) || node.Flags&ast.NodeFlagsOptionalChain != 0 {
		return false
	}
	if !l.isIntlCollatorType(l.checker.GetTypeAtLocation(node.AsPropertyAccessExpression().Expression)) {
		return false
	}
	return l.exactFunctionSlot(node) || intlCollatorCompareConst(node)
}

// intlCollatorCompareConst reports whether node, through any parentheses, is the whole initializer of
// a const named by an identifier with no annotation, as tsc's comparer is: the checker types that const
// as node's own type, so the const sees the read exactly as itself. A pattern takes the read's parts,
// not the read, so it's no such slot.
func intlCollatorCompareConst(node *ast.Node) bool {
	if !constMethodInitializer(node) {
		return false
	}
	outer := node
	for outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	declaration := outer.Parent.AsVariableDeclaration()
	return declaration.Type == nil && ast.IsIdentifier(declaration.Name())
}

// intlCollatorCompareValue lowers a read intlCollatorCompareRead admits to the closure in the
// collator's compare slot, read as a field and not as a callee, so without ir.Property's Method:
// library_intl.c's collator holds it in its first slot (fields.go's {compare, #locale}), and the
// JavaScript backend's AdamicIntlCollator as its compare field. It's false for any other node.
func (l *lowering) intlCollatorCompareValue(node *ast.Node) (ir.Expression, bool, error) {
	if !l.intlCollatorCompareRead(node) {
		return nil, false, nil
	}
	receiver := node.AsPropertyAccessExpression().Expression
	collator, err := l.expression(receiver)
	if err != nil {
		return nil, true, err
	}
	if collator.Type() != ir.Object {
		return nil, true, l.notYet(receiver, "Intl.Collator's compare read as a value off a collator not held as an object")
	}
	return ir.Property{Object: collator, Name: "compare", Of: ir.Closure}, true, nil
}
