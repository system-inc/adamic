package core

import (
	"encoding/json"
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// IdDenylistSettings is the decoded option surface: the names that may not be used.
//
// Upstream's schema is a bare array of unique strings, with no object form and no flags:
//
//	"id-denylist": ["error", "data", "err", "e", "cb", "callback"]
type IdDenylistSettings struct {
	// Denied holds every restricted name. A map because the only question ever asked of it is
	// membership, which is upstream's own `new Set(context.options)`.
	Denied map[string]bool
}

// DefaultIdDenylistSettings is an empty list, which is upstream's `defaultOptions: []` and means
// the rule denies nothing.
func DefaultIdDenylistSettings() IdDenylistSettings {
	return IdDenylistSettings{}
}

// DecodeIdDenylistOptions reads the denied names off the config.
//
// Hand rolled rather than routed through `rule.DecodeOptionsInto` because the wire shape is a JSON
// list of strings rather than an object, and the generic helper decodes into a struct.
//
// # Upstream's option surface is VARIADIC, and the list arrives whole
//
// Upstream reads `context.options`, which is every element of the config array after the severity:
//
//	"id-denylist": ["error", "data", "err", "cb"]
//
// The rule registers with `DecodeOptionList`, so this is handed `["data", "err", "cb"]`, upstream's
// list exactly. The config layer used to keep only the first element, and the first dry run against
// a seeded tree failed outright on the bare string it delivered:
//
//	rule configuration: rule id-denylist: decoding []string: json: cannot unmarshal string
//	into Go value of type []string
//
// The workaround was a nested spelling, `["error", ["data", "err", "cb"]]`, plus reading a bare
// string as a one-name list so the variadic spelling degraded to its first name. Both are gone: the
// first is an element that is not a string, which upstream's schema refuses, and the second was a
// silent drop of every name after the first.
//
// Empty input decodes to an empty list rather than failing. A rule configured as a bare "error" is
// handed nil options, and for this rule the correct answer is "deny nothing", which is the same
// answer upstream gives: its `denyList` is built from `context.options`, and an unconfigured rule
// has none.
func DecodeIdDenylistOptions(list []byte) (any, error) {
	settings := DefaultIdDenylistSettings()
	if len(list) == 0 {
		return settings, nil
	}

	var names []string
	if err := json.Unmarshal(list, &names); err != nil {
		return settings, fmt.Errorf(
			"id-denylist takes each name as its own string element, as in "+
				"[\"error\", \"data\", \"err\"]: %w", err)
	}
	if len(names) == 0 {
		return settings, nil
	}

	settings.Denied = make(map[string]bool, len(names))
	for _, name := range names {
		settings.Denied[name] = true
	}
	return settings, nil
}

// idDenylistSettingsFrom recovers the settings from whatever the config layer handed over.
func idDenylistSettingsFrom(options any) IdDenylistSettings {
	if settings, ok := rule.OptionsAs[IdDenylistSettings](options); ok {
		return settings
	}
	return DefaultIdDenylistSettings()
}

var messageIdDenylistRestricted = rule.Message{
	Id: "restricted",
	Description: "This name was denied by configuration. A denylist exists because the name says " +
		"nothing about what the value holds, so the next reader has to reconstruct it from the " +
		"surrounding code every time. Name it for what it is.",
}

var messageIdDenylistRestrictedPrivate = rule.Message{
	Id: "restrictedPrivate",
	Description: "This private field's name was denied by configuration. A denylist exists because " +
		"the name says nothing about what the value holds, so the next reader has to reconstruct " +
		"it from the surrounding code every time. Name it for what it is.",
}

// IdDenylist reports an identifier whose name appears in the configured denylist.
//
//	valid:   foo = "bar"                          (denying "bar": the string is not an identifier)
//	valid:   foo()                                (denying "foo": a callee is not a name we control)
//	valid:   foo.bar                              (denying "bar": reading a property we may not own)
//	valid:   import { foo as bar } from 'mod'     (denying "foo": the exporter chose that name)
//	valid:   const { foo: bar } = baz             (denying "foo": the source object's key)
//	valid:   Number.parseInt()                    (denying "Number": a global we did not name)
//	invalid: let foo = 1                          (denying "foo")
//	invalid: foo.bar = 1                          (denying "bar": a WRITE creates the property)
//	invalid: class C { #foo = 1; }                (denying "foo": reported as #foo)
//
// # The rule denies nothing until it is configured, and that is the whole reason it measured zero
//
// The audit beside this file records "violations in ahra: none" with a recommendation of Yes. That
// zero is not evidence the tree is clean, because the rule cannot report anything without an
// option naming what to deny, and nothing names one. It is the class of zero the standard calls
// out: an unconfigured options-driven rule is indistinguishable from a clean tree in a findings
// count. The rule is registered here and left unenabled; what it would catch depends entirely on
// the list Kirk chooses.
//
// # The audit's zero is an unconfigured rule, not an unreachable one, and that was measured
//
// A sibling port in this batch found `react/prefer-exact-props` carrying the same audit verdict as
// this rule -- "Yes, violations: none" -- while being structurally incapable of firing here: its
// only reporting path reads `settings.propWrapperFunctions`, a SHARED ESLint setting that
// `rule.Context` has no path to and that ahra configures nowhere. The ruling is at
// `internal/lint/rules/react/forbid_prop_types.go:158-183`. So the audit's zero cannot by itself
// tell a clean tree from a rule that cannot report at all, and both readings had to be separated
// here rather than assumed.
//
// They are different mechanisms and this rule is the ordinary one. It reads `context.options`,
// which the config layer hands the rule, rather than `context.settings`, which it does not. Checked
// three ways, because a grep for the absent thing is the weakest of them:
//
//	the rule file names no shared setting        0 hits for context.settings / sharedSettings
//	its ONLY import is ./utils/ast-utils         which itself has 0 such hits
//	  (the control, react/forbid-prop-types, reaches the setting INDIRECTLY through
//	   util/propWrapper.js:21, so a rule-file grep alone would have missed it -- which is
//	   exactly why the import was traced rather than trusted)
//	driven executably over all 145 corpus cases  0 change verdict when a settings block is present
//
// The last is the one that settles it, and it carries its own control: the same harness runs
// `react/prefer-exact-props` on a reporting case from its own corpus and measures 0 findings
// without settings against 1 with, so the probe demonstrably CAN detect a settings dependency and
// refuses to report if that control stays flat. This rule's zero is therefore "nothing is
// configured yet", which a denylist fixes, rather than "nothing can be".
// # What is checked is narrower than "every identifier", and each exclusion is a false-positive
// class upstream already paid for
//
// Upstream's `shouldCheck` is five tests, and the reasoning under all of them is the same: a
// denylist governs names WE choose, not names imposed on us from outside.
//
//	a callee or a constructor      foo() and new foo()      the function was named elsewhere
//	an import's exported name      import { foo as bar }    the exporting module chose it
//	a re-export's local name       export { foo as bar } from 'mod'
//	a destructuring source key     const { foo: bar } = x   the object's own key
//	a global reference             Number.parseInt()        the environment named it
//	an import attribute key        import x from 'm' with { type: 'json' }
//	new.target                     a meta property is not a name at all
//
// Member access splits rather than being excluded outright, and the split is the sharpest judgment
// in the rule: reading `foo.bar` is allowed because `foo` may be an object we do not own, while
// WRITING `foo.bar = 1` is denied because the write is what creates a property with that name.
// Upstream's `isAssignmentTarget` covers a plain assignment and every destructuring position, and
// all of them are reproduced below.
//
// # The global exemption is the one place this port cannot be faithful, and the reason is the
// configuration surface rather than the algorithm
//
// Upstream's `isReferenceToGlobalVariable` reads ESLint's global scope: a name is exempt when the
// scope holds a variable for it with ZERO definitions, which means the environment declared it and
// the source did not. Measured by mutation against the installed rule, that exemption decides
// **13 of the 145 corpus cases**, so it cannot be dropped:
//
//	node .../globalcount.cjs id-denylist id-denylist.json
//	  -> id-denylist: the global exemption changes 13/145 corpus verdicts
//
// The checker answers the same question through `idDenylistIsReferenceToGlobalVariable` below.
// Reaching for `identifierIsShadowed`, the helper already in this package, is the obvious move and
// it is WRONG here by exactly one row -- see that function's own comment for the measurement.
// Over all thirteen shapes, the four resolution outcomes are:
//
//	nil symbol                  not shadowed  -> the environment's  -> exempt
//	zero declarations           not shadowed  -> the environment's  -> exempt (undefined)
//	declared only in a .d.ts    not shadowed  -> the environment's  -> exempt (Number, Map)
//	declared in source          SHADOWED      -> ours               -> checked
//
// **Where it diverges, and why the divergence is not repairable here.** ESLint lets a config
// DECLARE globals that the code never mentions -- `languageOptions.globals: {myGlobal: "readonly"}`
// and the `/* global myGlobal */` comment directive. Eight corpus cases turn on that surface, five
// of which actually diverge -- they are pinned one by one, at MEASURED verdicts, in
// TestIdDenylistDivergesOnTheEslintGlobalsSurface. And
// cohere has no counterpart to it: `CohereSettings.json` carries no globals key, and the checker's
// answer comes from the lib and `@types` a file is actually compiled against. So for a name that is
// neither declared in source nor declared by any lib, upstream can be told it is a global and this
// port cannot:
//
//	/* global myGlobal: readonly */ myGlobal = 5;      upstream clean, here reported
//	foo = { [myGlobal]: 1 };  (globals: myGlobal)      upstream clean, here reported
//	/* globals Number: off */ Number.parseInt()        upstream reports, here exempt
//	var foo = undefined;      (globals: undefined off) upstream reports, here exempt
//
// That is the source-type hazard the standard names, one layer over: the two instruments are being
// asked different questions, and re-running the measurement reproduces the same difference. The
// checker's answer is the RIGHT one for a TypeScript tree, because an undeclared name here is a
// compile error rather than an ambient global, so the divergence is recorded rather than worked
// around. The fixtures below carry these cases at the verdict this port produces, with the
// upstream verdict named at the line.
//
// # Upstream deduplicates by source range and this port does not need to
//
// `reportedNodes` is a set of stringified ranges, and its comment says why: in ESTree a shorthand
// property `{ foo }` is TWO identifier nodes, a key and a value, sharing one range, so a rule
// visiting both would report twice. Measured here, `({ foo } = a)` is a single
// `KindShorthandPropertyAssignment` carrying exactly one name node, and `let { foo } = a` is a
// single `KindBindingElement` likewise. That is a parser difference rather than a judgment, so the
// dedup set is a workaround for a constraint this tree does not have and is not reproduced.
//
// # No fix
//
// Upstream offers none, and none is possible: the repair is to choose a different name, and nothing
// here knows which one the author meant.
var IdDenylist = rule.Rule{
	Name: "id-denylist",

	// Telling a global reference from a name the source declared is name resolution. See the
	// exemption section above for the four resolution outcomes and how each was measured.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings := idDenylistSettingsFrom(options)
		if len(settings.Denied) == 0 {
			// Upstream builds an empty Set and every identifier fails `isRestricted`, so the rule
			// reports nothing. Registering no listener is the same decision paid for once per file
			// rather than once per identifier, and it is what a rule configured as a bare severity
			// gets, since there is no sensible default denylist.
			//
			// This early exit is a MEASURED EQUIVALENCE rather than a discrimination, and it is
			// recorded so the next reader does not chase it. A mutation removing it survives every
			// fixture, correctly: with an empty map the per-identifier membership test below fails
			// for every name, so the rule reports nothing either way. The exit buys speed, not a
			// verdict, and no fixture can or should distinguish the two.
			return rule.Listeners{}
		}

		check := func(node *ast.Node) {
			if !settings.Denied[idDenylistNameOf(node)] {
				return
			}
			if !idDenylistShouldCheck(ctx, node) {
				return
			}
			if node.Kind == ast.KindPrivateIdentifier {
				ctx.ReportNode(node, rule.Message{
					Id: messageIdDenylistRestrictedPrivate.Id,
					Description: fmt.Sprintf("Identifier '#%s' is restricted. %s",
						idDenylistNameOf(node), messageIdDenylistRestrictedPrivate.Description),
				})
				return
			}
			ctx.ReportNode(node, rule.Message{
				Id: messageIdDenylistRestricted.Id,
				Description: fmt.Sprintf("Identifier '%s' is restricted. %s",
					idDenylistNameOf(node), messageIdDenylistRestricted.Description),
			})
		}

		return rule.Listeners{
			ast.KindIdentifier:        check,
			ast.KindPrivateIdentifier: check,
		}
	},
}

// idDenylistNameOf reads the name a denylist entry would have to match.
//
// A private identifier's `Text()` already carries the leading `#`, so it is trimmed: upstream
// matches `node.name`, which for a `PrivateIdentifier` is the name WITHOUT the hash, and renders
// the hash back into the message itself. A port matching the hashed spelling would need `#foo` in
// the config to deny `#foo`, which is not what the corpus configures.
func idDenylistNameOf(node *ast.Node) string {
	text := node.Text()
	if node.Kind == ast.KindPrivateIdentifier && len(text) > 0 && text[0] == '#' {
		return text[1:]
	}
	return text
}

// idDenylistShouldCheck is upstream's `shouldCheck`, arm for arm and in upstream's order.
func idDenylistShouldCheck(ctx rule.Context, node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}

	// Import attributes are named by the runtime that reads them, not by us.
	if idDenylistIsImportAttributeKey(node) {
		return false
	}

	// `new.target` and `import.meta` are single tokens spelled with a dot. The identifier after the
	// dot is not a name anybody chose.
	if parent.Kind == ast.KindMetaProperty {
		return false
	}

	// Member access has its own rule, and it is the sharpest judgment here. READING a property with
	// a denied name is allowed, because the object may be one we do not own. WRITING one is not,
	// because the write is what creates a property with that name.
	//
	// Upstream tests `parent.type === "MemberExpression" && parent.property === node &&
	// !parent.computed`. Here a non-computed member access is `KindPropertyAccessExpression` and a
	// computed one is `KindElementAccessExpression`, a separate kind, so the `!computed` half is
	// carried by the kind itself and the identifier inside `foo[bar]` never reaches this arm.
	if parent.Kind == ast.KindPropertyAccessExpression {
		if access := parent.AsPropertyAccessExpression(); access != nil && access.Name() == node {
			return idDenylistIsAssignmentTarget(parent)
		}
	}

	// A callee names a function somebody else declared.
	if parent.Kind == ast.KindCallExpression {
		if call := parent.AsCallExpression(); call != nil && call.Expression == node {
			return false
		}
	}
	if parent.Kind == ast.KindNewExpression {
		if construct := parent.AsNewExpression(); construct != nil && construct.Expression == node {
			return false
		}
	}

	if idDenylistIsRenamedImport(node) {
		return false
	}
	if idDenylistIsPropertyNameInDestructuring(node) {
		return false
	}
	// The environment named it, not us. See the exemption section on the rule for the four
	// resolution outcomes and where it diverges.
	if idDenylistIsReferenceToGlobalVariable(ctx, node) {
		return false
	}
	return true
}

// idDenylistIsReferenceToGlobalVariable is upstream's `isReferenceToGlobalVariable`.
//
// Upstream reads the global scope and requires THREE things at once: a variable exists there under
// this name, it has zero definitions, and this identifier is one of its references. All three
// matter, and the middle one is what makes the predicate narrow.
//
// The tempting shortcut is `!identifierIsShadowed`, the helper already in this package, and it is
// WRONG here by exactly one row. That helper answers "is any declaration of this symbol in a real
// source file", so it returns false both for a genuine ambient global and for a name that resolves
// to NOTHING AT ALL. Upstream separates those: an undeclared name is `scope.through` rather than a
// global-scope variable, so it is checked rather than exempted. Measured directly against the
// installed rule:
//
//	foo = 1;             denying "foo", nothing declares foo   REPORTS
//	foo = 1;             denying "foo", globals: {foo}         clean
//	Number.parseInt();   denying "Number"                      clean
//
// Using the shadow helper cost 19 of the 88 reporting fixtures on the first run -- every bare
// `foo = "bar"`, every `foo.bar()`, every `foo[bar]` -- because in a fixture nothing declares those
// names and they all came back with a nil symbol. The four rows this predicate draws instead:
//
//	nil symbol                  resolves to nothing  -> CHECKED  (upstream's scope.through)
//	zero declarations           an ambient global    -> exempt   (undefined)
//	declared only in a .d.ts    an environment global-> exempt   (Number, Map, Object, Array)
//	declared in source          ours                 -> CHECKED
func idDenylistIsReferenceToGlobalVariable(ctx rule.Context, node *ast.Node) bool {
	if ctx.TypeChecker == nil {
		return false
	}
	symbol := ctx.TypeChecker.GetSymbolAtLocation(node)
	if symbol == nil {
		// Upstream's `globalScope.set.get(name)` found nothing, so this is not a global reference.
		return false
	}
	if len(symbol.Declarations) == 0 {
		// A symbol the checker knows without a declaration node anywhere, which is what `undefined`
		// is. Upstream's `variable.defs.length === 0` says the same thing.
		return true
	}
	for _, declaration := range symbol.Declarations {
		if file := ast.GetSourceFileOfNode(declaration); file != nil && !file.IsDeclarationFile {
			// Declared in real source, so the name is one we chose. Upstream's non-empty `defs`.
			return false
		}
	}
	// Every declaration is in a `.d.ts`, so the environment declared it and we did not.
	return true
}

// idDenylistIsImportAttributeKey is ESLint's `astUtils.isImportAttributeKey`.
//
// An import attribute is read by the module loader, which decides what the keys mean, so a denylist
// governing names WE choose has no business there. Two spellings reach it and both are in the
// corpus, which is why the second arm is worth its machinery rather than being declined:
//
//	import foo from 'foo.json' with { type: 'json' }      static, its own node kind
//	import('foo.json', { with: { type: 'json' } })        dynamic, ordinary object literals
//
// The static form is trivial here because our parser gives it a dedicated `KindImportAttribute`
// whose name slot is exactly upstream's `parent.key`.
//
// The dynamic form is not, and the corpus pins precisely where the exemption STOPS. Eight cases
// exercise this, six clean and two reporting, and the two reporting ones are what make it a real
// judgment rather than "anything under an import call is exempt":
//
//	import('foo.json', { with: { [type]: 'json' } })   REPORTS -- a computed key is an expression
//	import('foo.json', { with: { type: json } })       REPORTS -- the VALUE is not a key
//	import('foo.json', { with: { type } })             clean   -- shorthand, key and value are one
//	import('foo.json', { 'with': { type: 'json' } })   clean   -- a quoted `with` still names it
//
// Upstream reaches the second arm by recursing: it exempts a non-computed key of an object literal
// whose parent object is either the `options` argument of an ImportExpression, or is itself the
// value of an exempt key. The recursion is what makes the nested `{ with: { type } }` work, and it
// is reproduced rather than flattened, because flattening it would exempt a key at any depth.
func idDenylistIsImportAttributeKey(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}

	// The static form: `import x from 'm' with { type: 'json' }`, and the same clause on an export.
	if parent.Kind == ast.KindImportAttribute {
		attribute := parent.AsImportAttribute()
		return attribute != nil && attribute.Name() == node
	}

	// The dynamic form. Only a non-computed key or a shorthand's single name qualifies; a computed
	// key's identifier has `KindComputedPropertyName` for a parent and never reaches here, and a
	// property's VALUE fails the name test below.
	var objectLiteral *ast.Node
	switch parent.Kind {
	case ast.KindPropertyAssignment:
		assignment := parent.AsPropertyAssignment()
		if assignment == nil || assignment.Name() != node {
			return false
		}
		objectLiteral = parent.Parent

	case ast.KindShorthandPropertyAssignment:
		// Upstream's `parent.value === node && parent.shorthand && !parent.method`. A shorthand is
		// one node here carrying one name, so being the name IS being the shorthand.
		assignment := parent.AsShorthandPropertyAssignment()
		if assignment == nil || assignment.Name() != node {
			return false
		}
		objectLiteral = parent.Parent

	default:
		return false
	}

	if objectLiteral == nil || objectLiteral.Kind != ast.KindObjectLiteralExpression {
		return false
	}
	return idDenylistIsImportOptionsObject(objectLiteral)
}

// idDenylistIsImportOptionsObject answers whether an object literal is, or sits inside, the options
// argument of a dynamic `import()`.
//
// This is the recursive half of upstream's `isImportAttributeKey`. The outer object is the call's
// second argument; an inner object qualifies when it is the value of a key that itself qualifies,
// which is what lets `{ with: { type } }` exempt `type` while `{ with: { type: json } }` still
// reports `json`.
func idDenylistIsImportOptionsObject(objectLiteral *ast.Node) bool {
	parent := objectLiteral.Parent
	if parent == nil {
		return false
	}

	// The outer object: the second argument of `import(specifier, options)`.
	if parent.Kind == ast.KindCallExpression && ast.IsImportCall(parent) {
		call := parent.AsCallExpression()
		if call == nil || call.Arguments == nil || len(call.Arguments.Nodes) < 2 {
			return false
		}
		return call.Arguments.Nodes[1] == objectLiteral
	}

	// An inner object: the value of a key that is itself an import attribute key. Upstream spells
	// this `isImportAttributeKey(objectExpressionParent.key)`, so the recursion runs through the
	// KEY rather than through this object, which is why a quoted `'with'` still qualifies -- the
	// name slot of the property assignment is what gets asked, whatever its spelling.
	if parent.Kind == ast.KindPropertyAssignment {
		assignment := parent.AsPropertyAssignment()
		if assignment == nil || assignment.Initializer != objectLiteral {
			return false
		}
		grandparent := parent.Parent
		if grandparent == nil || grandparent.Kind != ast.KindObjectLiteralExpression {
			return false
		}
		return idDenylistIsImportOptionsObject(grandparent)
	}
	return false
}

// idDenylistIsAssignmentTarget is upstream's `isAssignmentTarget`, which decides whether a member
// access is being written rather than read.
//
// Upstream's five arms are a plain assignment, an ArrayPattern element, a RestElement, an
// ObjectPattern property's value, and an AssignmentPattern's left side. Only the first and the
// destructuring-through-a-member forms can reach a MEMBER ACCESS at all, because a binding pattern
// declaring new names holds identifiers rather than member accesses. In this tree an assignment
// destructuring is written with literal syntax rather than binding syntax, so `[foo.bar] = a` is a
// `KindArrayLiteralExpression` and `({ x: foo.bar } = a)` a `KindPropertyAssignment` inside a
// `KindObjectLiteralExpression`, which is why those kinds appear here and the binding kinds do not.
func idDenylistIsAssignmentTarget(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}

	switch parent.Kind {
	// A plain assignment, which is upstream's `AssignmentExpression && parent.left === node`. Only
	// `=` counts: upstream tests the node type rather than the operator, and a compound assignment
	// such as `foo.bar += 1` is also an AssignmentExpression there, so it is also a write here.
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		return binary != nil && binary.Left == node && ast.IsAssignmentOperator(binary.OperatorToken.Kind)

	// `[foo.bar] = a`, upstream's ArrayPattern. An assignment destructuring parses as an array
	// LITERAL here and is reinterpreted as a pattern by the checker, so the element's parent is the
	// literal.
	//
	// The destructuring test is load-bearing and upstream needs no equivalent, because its parser
	// produces `ArrayPattern` ONLY for a genuine target and `ArrayExpression` for a value. Ours
	// produces one kind for both. Without the test, the bare statement `[foo.bar]` reads as a
	// pattern and reports, and upstream's corpus lists exactly that source as CLEAN.
	case ast.KindArrayLiteralExpression:
		return idDenylistIsDestructuringTarget(parent)

	// `[...foo.bar] = a` and `({ ...foo.bar } = a)`, upstream's RestElement. Same reasoning: a
	// spread in a value position is not a write.
	case ast.KindSpreadElement, ast.KindSpreadAssignment:
		return parent.Parent != nil && idDenylistIsDestructuringTarget(parent.Parent)

	// `({ x: foo.bar } = a)`, upstream's `Property && parent.value === node && ObjectPattern`. The
	// grandparent test is what separates a destructuring target from an ordinary object literal
	// value, which is a read.
	case ast.KindPropertyAssignment:
		assignment := parent.AsPropertyAssignment()
		if assignment == nil || assignment.Initializer != node {
			return false
		}
		return parent.Parent != nil && parent.Parent.Kind == ast.KindObjectLiteralExpression &&
			idDenylistIsDestructuringTarget(parent.Parent)

	// `[foo.bar = 1] = a`, upstream's `AssignmentPattern && parent.left === node`. A default inside
	// a destructuring assignment is an ordinary binary `=` here, so it is already covered by the
	// KindBinaryExpression arm above and this case exists only for the binding-pattern spelling,
	// where the target is a name rather than a member access and cannot reach this function.
	default:
		return false
	}
}

// idDenylistIsDestructuringTarget answers whether an object literal is standing in for a pattern.
//
// Upstream reads the node TYPE, because its parser produces an `ObjectPattern` for a destructuring
// target and an `ObjectExpression` for a value. Our parser produces `KindObjectLiteralExpression`
// for both and lets the checker reinterpret it, so the question has to be asked of the context: an
// object literal is a pattern exactly when it is the left side of an assignment, or is itself
// nested inside something that is.
func idDenylistIsDestructuringTarget(node *ast.Node) bool {
	current := node
	for current != nil {
		parent := current.Parent
		if parent == nil {
			return false
		}
		switch parent.Kind {
		case ast.KindBinaryExpression:
			binary := parent.AsBinaryExpression()
			return binary != nil && binary.Left == current &&
				binary.OperatorToken.Kind == ast.KindEqualsToken

		// Keep walking out through the shapes that can nest one pattern inside another.
		case ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment,
			ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression,
			ast.KindSpreadAssignment, ast.KindSpreadElement, ast.KindParenthesizedExpression:
			current = parent

		// A `for (const [a] of b)` head and a parameter both make the literal a pattern without an
		// `=` anywhere.
		case ast.KindForOfStatement, ast.KindForInStatement, ast.KindParameter:
			return true

		default:
			return false
		}
	}
	return false
}

// idDenylistIsRenamedImport is upstream's `isRenamedImport`: the name on the far side of an `as` in
// an import, or in a re-export, belongs to the module we are importing from.
//
// The two directions are not symmetric, and getting them the wrong way round is silent. For
// `import { a as b }` the EXPORTED name is `a` (our `PropertyName`) and the local binding is `b`.
// For `export { a as b } from 'mod'` the local half is `a` (also our `PropertyName`) and the
// exported name is `b`. So both arms exempt `PropertyName`, but they mean opposite things and only
// the export arm requires a `from` clause: `let foo; export { foo as bar };` re-exports a name this
// file DECLARED, so `foo` is ours and is checked, which is exactly what upstream's
// `parent.parent.source` test says.
func idDenylistIsRenamedImport(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}

	switch parent.Kind {
	case ast.KindImportSpecifier:
		specifier := parent.AsImportSpecifier()
		// `import { foo } from 'mod'` has no PropertyName at all, and upstream's
		// `parent.imported !== parent.local` is false for it, so the name is checked.
		return specifier != nil && specifier.PropertyName == node

	case ast.KindExportSpecifier:
		specifier := parent.AsExportSpecifier()
		if specifier == nil || specifier.PropertyName != node {
			return false
		}
		return idDenylistExportHasSource(parent)
	}
	return false
}

// idDenylistExportHasSource answers upstream's `parent.parent.source`: whether the export specifier
// belongs to a re-export rather than to an export of something this file declares.
func idDenylistExportHasSource(specifier *ast.Node) bool {
	// specifier -> KindNamedExports -> KindExportDeclaration
	namedExports := specifier.Parent
	if namedExports == nil || namedExports.Parent == nil {
		return false
	}
	declaration := namedExports.Parent.AsExportDeclaration()
	return declaration != nil && declaration.ModuleSpecifier != nil
}

// idDenylistIsPropertyNameInDestructuring is upstream's `isPropertyNameInDestructuring`: in
// `const { foo: bar } = baz`, the key `foo` is the SOURCE object's own property name and we did not
// choose it, so only the new binding `bar` is checked.
//
// Upstream reads `Property && parent.parent.type === "ObjectPattern" && parent.key === node`. This
// tree spells the same source two ways depending on whether it declares bindings, and both are
// reached here: a declaration destructuring is a `KindBindingElement` carrying a `PropertyName`,
// and an assignment destructuring is a `KindPropertyAssignment` inside an object literal being used
// as a pattern.
func idDenylistIsPropertyNameInDestructuring(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}

	switch parent.Kind {
	case ast.KindBindingElement:
		element := parent.AsBindingElement()
		if element == nil || element.PropertyName != node {
			return false
		}
		// Upstream's `!parent.computed`. A computed key is its own node here, so a key reached
		// through one is not the plain property name upstream exempts.
		return node.Kind != ast.KindComputedPropertyName

	case ast.KindPropertyAssignment:
		assignment := parent.AsPropertyAssignment()
		if assignment == nil || assignment.Name() != node {
			return false
		}
		return parent.Parent != nil && parent.Parent.Kind == ast.KindObjectLiteralExpression &&
			idDenylistIsDestructuringTarget(parent.Parent)
	}
	return false
}
