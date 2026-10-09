package core

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// IdMatchSettings is the decoded option surface.
//
// Upstream's schema is a tuple: a pattern string, then an object of four booleans.
//
//	"id-match": ["error", "^[a-z]+([A-Z][a-z]+)*$", { "properties": true }]
type IdMatchSettings struct {
	// Pattern is the compiled regular expression every checked name must match. Nil means the rule
	// does nothing, which is what an unconfigured rule gets.
	Pattern *regexp.Regexp
	// PatternText is the pattern as written, because the message quotes it back at the reader.
	PatternText string
	// CheckProperties extends the rule to object properties and member accesses.
	CheckProperties bool
	// CheckClassFields extends the rule to class field names, public and private.
	CheckClassFields bool
	// OnlyDeclarations narrows the rule to function declarations and variable declarators.
	OnlyDeclarations bool
	// IgnoreDestructuring exempts a name a destructuring pattern merely re-uses, while still
	// checking one it invents.
	IgnoreDestructuring bool
}

// DefaultIdMatchSettings is upstream's `defaultOptions`: match everything, check nothing extra.
//
// Upstream's default pattern is `^.+$`, which matches every non-empty name, so an unconfigured rule
// reports nothing. That is spelled here as a nil pattern rather than as a compiled `^.+$`, because
// the two are behaviourally identical and nil lets the rule decline the file without walking it.
func DefaultIdMatchSettings() IdMatchSettings {
	return IdMatchSettings{}
}

// idMatchRawOptions is the wire shape of the second tuple element.
//
// Every field is a plain bool rather than a pointer because every default is FALSE, so a zero value
// and an absent key mean the same thing. That is the opposite of `no-invalid-this`, whose one
// option defaults true and therefore needs pointers to tell absent from false.
type idMatchRawOptions struct {
	Properties          bool `json:"properties"`
	ClassFields         bool `json:"classFields"`
	OnlyDeclarations    bool `json:"onlyDeclarations"`
	IgnoreDestructuring bool `json:"ignoreDestructuring"`
}

// DecodeIdMatchOptions reads the pattern and the flag object off the config.
//
// Hand rolled rather than routed through `rule.DecodeOptionsInto` because the wire shape is a
// heterogeneous list: element one is a string and element two is an object, which no single Go
// struct decodes.
//
// # Upstream's option surface is two elements, and both arrive
//
// Upstream reads `context.options` as a pattern followed by a flag object:
//
//	"id-match": ["error", "^[a-z]+$", { "properties": true }]
//
// The rule registers with `DecodeOptionList` and is handed `["^[a-z]+$", {"properties": true}]`.
// The config layer used to keep only the first element, so that spelling arrived as the bare
// pattern with every flag silently off, quieter than id-denylist's outright error on the same cause.
// The nested `["error", ["^[a-z]+$", {...}]]` workaround is refused now: its first element is not a
// string.
//
// Empty input decodes to the default settings rather than failing, which is the answer for a rule
// configured as a bare "error": upstream's default pattern matches every name, so such a rule
// reports nothing rather than erroring.
//
// A pattern that does not compile is returned as an error rather than silently ignored. Upstream
// would throw at `new RegExp(pattern, "u")`, and a rule that quietly matched everything instead
// would be inert while looking configured, which is the worst of the two failures. So is a flag key
// upstream's schema does not declare, which `additionalProperties: false` refuses there.
func DecodeIdMatchOptions(list []byte) (any, error) {
	settings := DefaultIdMatchSettings()
	elements, err := rule.OptionElements(list, 2)
	if err != nil || len(elements) == 0 {
		return settings, err
	}

	var patternText string
	if err := json.Unmarshal(elements[0], &patternText); err != nil {
		return settings, fmt.Errorf("id-match element 1 takes a pattern string, got %s", elements[0])
	}

	if len(elements) > 1 {
		var flags idMatchRawOptions
		if err := rule.UnmarshalOptions(elements[1], &flags); err != nil {
			return settings, fmt.Errorf("id-match element 2: %w", err)
		}
		settings.CheckProperties = flags.Properties
		settings.CheckClassFields = flags.ClassFields
		settings.OnlyDeclarations = flags.OnlyDeclarations
		settings.IgnoreDestructuring = flags.IgnoreDestructuring
	}

	if patternText == "" {
		return settings, nil
	}

	// Upstream compiles with the `u` flag, which in JavaScript makes the pattern operate on code
	// points rather than UTF-16 units. Go's regexp is already code-point oriented, so there is no
	// flag to add: the behaviour the `u` flag buys is the only behaviour Go offers.
	compiled, err := regexp.Compile(patternText)
	if err != nil {
		return settings, fmt.Errorf("id-match pattern %q does not compile: %w", patternText, err)
	}
	settings.Pattern = compiled
	settings.PatternText = patternText
	return settings, nil
}

// idMatchSettingsFrom recovers the settings from whatever the config layer handed over.
func idMatchSettingsFrom(options any) IdMatchSettings {
	if settings, ok := rule.OptionsAs[IdMatchSettings](options); ok {
		return settings
	}
	return DefaultIdMatchSettings()
}

var messageIdMatchNotMatch = rule.Message{
	Id: "notMatch",
	Description: "A naming convention is enforced here so that a reader can tell what kind of thing " +
		"a name refers to without looking it up. Rename it to fit the pattern.",
}

var messageIdMatchNotMatchPrivate = rule.Message{
	Id: "notMatchPrivate",
	Description: "A naming convention is enforced here so that a reader can tell what kind of thing " +
		"a name refers to without looking it up. Rename this private field to fit the pattern.",
}

// IdMatch reports an identifier whose name does not match the configured regular expression.
//
//	valid:   let fooBar = 1                       (pattern ^[a-z]+([A-Z][a-z]+)*$)
//	valid:   foo.no_under                         (properties off, the default)
//	valid:   no_camelcased()                      (a callee names something declared elsewhere)
//	valid:   Object.keys(bar)                     (a global we did not name)
//	valid:   let { category_id } = query          (ignoreDestructuring on)
//	invalid: let no_camelcased = 1
//	invalid: obj.no_under = 2                     (properties on)
//	invalid: class C { no_under = 1; }            (classFields on)
//	invalid: let { category_id: category_alias }  (a NEW name, so ignoreDestructuring does not help)
//
// # The rule matches nothing until it is configured, and that is the whole reason it measured zero
//
// The audit beside this file records "violations in ahra: none" with a recommendation of Yes. That
// zero is not evidence the tree is clean: upstream's default pattern is `^.+$`, which every
// non-empty name matches, so an unconfigured rule cannot report. It is the class of zero the
// standard calls out. The rule is registered here and left unenabled; what it catches depends
// entirely on the pattern Kirk chooses.
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
//	driven executably over all 100 corpus cases  0 change verdict when a settings block is present
//
// The last is the one that settles it, and it carries its own control: the same harness runs
// `react/prefer-exact-props` on a reporting case from its own corpus and measures 0 findings
// without settings against 1 with, so the probe demonstrably CAN detect a settings dependency and
// refuses to report if that control stays flat. This rule's zero is therefore "nothing is
// configured yet", which a pattern fixes, rather than "nothing can be".
// # Four options, and each one gates a different arm rather than narrowing one filter
//
//	properties          object literal keys, member accesses, and shorthand values
//	classFields         class field names, public and private
//	onlyDeclarations    narrows everything to function declarations and variable declarators
//	ignoreDestructuring exempts a name a pattern RE-USES, never one it invents
//
// The last is the subtlest and the corpus pins it precisely. Measured against the installed rule
// with `ignoreDestructuring: true`:
//
//	let { category_id } = query                    clean   -- the name came from the object
//	let { category_id: category_id } = query        clean   -- same name, still re-used
//	let { category_id = 1 } = query                 clean   -- a default does not rename
//	let { category_id: category_alias } = query    REPORTS -- category_alias is a NEW name
//	let { a: b, ...other_props } = query           REPORTS -- a rest element always invents
//
// So the option asks "did this pattern coin this name", not "is this inside a pattern". A port
// reading it as the latter loses the last two rows and no count would say so.
//
// # Where ESTree's ObjectPattern goes here, and why the destructuring arm is two arms
//
// Upstream's whole destructuring branch keys on `parent.type === "Property"` with
// `parent.parent.type === "ObjectPattern"`, reading `parent.key`, `parent.value` and
// `parent.shorthand`. This tree spells the same source two different ways depending on whether it
// DECLARES bindings, and both had to be handled:
//
//	let { a: b } = x        KindBindingElement inside KindObjectBindingPattern
//	({ a: b } = x)          KindPropertyAssignment inside KindObjectLiteralExpression
//
// The binding form is the friendlier of the two, and it collapses upstream's three fields into
// two slots on one node: `PropertyName` is upstream's key when a rename happened and is ABSENT for
// a shorthand, and `Name()` is always the binding that comes into scope. So upstream's
// `parent.key.name === parent.value.name` test, which asks whether a rename happened, is here just
// "is PropertyName nil", and the shorthand flag needs no separate field at all.
//
// # The member-access arm is upstream's three-way split, and it is not a single condition
//
// Upstream's MemberExpression branch has three arms and they are not alternatives to one test:
//
//	the OBJECT of the access                      foo.bar        checks foo
//	the property of an assignment's LEFT side     foo.bar = 1    checks bar
//	an assignment whose right side is not itself
//	  a member access                             foo.bar = baz  checks bar
//
// The third is why `foo.no_under = boom.bam_pow` reports only ONCE, on `no_under`, while
// `foo.no_under = 2` also reports once: the right-hand `bam_pow` is the property of a member access
// being READ, and reading a property we may not own is not our naming decision. That distinction
// is reproduced rather than simplified, because collapsing it reports every property read in the
// tree the moment `properties` is on.
//
// # The global exemption, and where it diverges
//
// Upstream's `isReferenceToGlobalVariable` reads ESLint's global scope. Measured by mutation
// against the installed rule, it decides **2 of the 100 corpus cases**, both of them `Object` and
// `Array`:
//
//	node .../globalcount.cjs id-match id-match.json
//	  -> id-match: the global exemption changes 2/100 corpus verdicts
//
// Unlike `id-denylist`, whose corpus leans on ESLint's `globals` CONFIG for nine cases, id-match's
// corpus carries no `globals` languageOption and no `/* global */` directive at all, so both cases
// are real lib globals the checker resolves by itself. This port therefore reproduces the whole
// exemption with no divergence, through `idDenylistIsReferenceToGlobalVariable`, whose four
// resolution rows are documented on that function.
//
// # No fix
//
// Upstream offers none, and none is possible: the repair is to rename, and nothing here knows what
// the author meant. Renaming would also have to reach every reference, which is a refactor rather
// than a lint fix.
var IdMatch = rule.Rule{
	Name: "id-match",

	// Telling a global reference from a name the source declared is name resolution.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings := idMatchSettingsFrom(options)
		if settings.Pattern == nil {
			// Upstream's default pattern `^.+$` matches every non-empty name, so an unconfigured
			// rule reports nothing. Registering no listener is the same decision paid once per file.
			return rule.Listeners{}
		}

		return rule.Listeners{
			ast.KindIdentifier: func(node *ast.Node) {
				checkIdMatchIdentifier(ctx, node, settings)
			},
			ast.KindPrivateIdentifier: func(node *ast.Node) {
				checkIdMatchPrivateIdentifier(ctx, node, settings)
			},
		}
	},
}

// idMatchIsInvalid answers upstream's `isInvalid`: the name fails the pattern.
func idMatchIsInvalid(settings IdMatchSettings, name string) bool {
	return !settings.Pattern.MatchString(name)
}

// reportIdMatch emits the finding, rendering the name and the pattern the way upstream does.
func reportIdMatch(ctx rule.Context, node *ast.Node, settings IdMatchSettings) {
	name := node.Text()
	if node.Kind == ast.KindPrivateIdentifier {
		// `Text()` already carries the `#`, and upstream renders it too, from a name that does not.
		ctx.ReportNode(node, rule.Message{
			Id: messageIdMatchNotMatchPrivate.Id,
			Description: fmt.Sprintf("Identifier '%s' does not match the pattern '%s'. %s",
				name, settings.PatternText, messageIdMatchNotMatchPrivate.Description),
		})
		return
	}
	ctx.ReportNode(node, rule.Message{
		Id: messageIdMatchNotMatch.Id,
		Description: fmt.Sprintf("Identifier '%s' does not match the pattern '%s'. %s",
			name, settings.PatternText, messageIdMatchNotMatch.Description),
	})
}

// checkIdMatchPrivateIdentifier is upstream's `PrivateIdentifier` listener.
//
// A private name is only ever a class member, so the only question is whether class fields are
// being checked at all. Upstream gates on `PropertyDefinition` specifically, which means a private
// METHOD is checked regardless of the `classFields` option, and a private FIELD only when it is on.
// Measured: `class C { #no_under() {} }` reports with classFields off and
// `class C { #no_under = 1; }` does not.
func checkIdMatchPrivateIdentifier(ctx rule.Context, node *ast.Node, settings IdMatchSettings) {
	if node.Parent != nil && node.Parent.Kind == ast.KindPropertyDeclaration && !settings.CheckClassFields {
		return
	}
	// The pattern is matched against the name WITHOUT the hash, which is upstream's `node.name`.
	name := node.Text()
	if len(name) > 0 && name[0] == '#' {
		name = name[1:]
	}
	if idMatchIsInvalid(settings, name) {
		reportIdMatch(ctx, node, settings)
	}
}

// checkIdMatchIdentifier is upstream's `Identifier` listener, arm for arm and in upstream's order.
//
// The order matters: the arms are a chain of `else if`, so the first one whose shape matches wins
// and the rest never run. Reordering them changes verdicts.
func checkIdMatchIdentifier(ctx rule.Context, node *ast.Node, settings IdMatchSettings) {
	parent := node.Parent
	if parent == nil {
		return
	}
	name := node.Text()

	// Upstream's three unconditional exits, in its order.
	if idDenylistIsReferenceToGlobalVariable(ctx, node) {
		return
	}
	if idDenylistIsImportAttributeKey(node) {
		return
	}
	if parent.Kind == ast.KindMetaProperty {
		return
	}

	switch {
	// The member-access arm. A computed access is `KindElementAccessExpression`, a separate kind
	// here, and upstream's MemberExpression covers both spellings; but every one of its three arms
	// tests a NON-computed shape (`parent.object`, `effectiveParent.left.property.name`), so only
	// the property-access kind can satisfy them and the element-access kind falls through to the
	// default arm exactly as a computed member does upstream.
	case parent.Kind == ast.KindPropertyAccessExpression:
		checkIdMatchMemberAccess(ctx, node, parent, name, settings)

	// An object literal's own key, which upstream separated out for eslint/eslint#15123 so that a
	// key is governed by `properties` alone rather than by the destructuring logic below.
	case parent.Kind == ast.KindPropertyAssignment && parent.Parent != nil &&
		parent.Parent.Kind == ast.KindObjectLiteralExpression &&
		parent.AsPropertyAssignment() != nil && parent.AsPropertyAssignment().Name() == node &&
		!idDenylistIsDestructuringTarget(parent.Parent):
		if settings.CheckProperties && idMatchIsInvalid(settings, name) {
			reportIdMatch(ctx, node, settings)
		}

	// The destructuring arms, in the two spellings this tree has for them.
	case parent.Kind == ast.KindBindingElement:
		checkIdMatchBindingElement(ctx, node, parent, name, settings)

	// The assignment-destructuring spelling. The guard is the destructuring test rather than the
	// node kind alone, because an ordinary object literal uses the same kinds: without it, a
	// property VALUE such as the `no_under` in `var obj = {key: no_under}` reaches this arm, finds
	// no pattern, and returns silently instead of falling through to the default arm that should
	// judge it. That cost two upstream cases on the first fixture run and neither was a count
	// anybody could have read as a shape problem.
	case (parent.Kind == ast.KindShorthandPropertyAssignment ||
		parent.Kind == ast.KindPropertyAssignment) &&
		parent.Parent != nil && parent.Parent.Kind == ast.KindObjectLiteralExpression &&
		idDenylistIsDestructuringTarget(parent.Parent):
		checkIdMatchAssignmentPattern(ctx, node, parent, name, settings)

	// An import binding. Only the LOCAL name is ours; the exported name on the far side of an `as`
	// was chosen by the module we import from.
	case parent.Kind == ast.KindImportSpecifier || parent.Kind == ast.KindImportClause ||
		parent.Kind == ast.KindNamespaceImport:
		if idMatchIsLocalImportBinding(node, parent) && idMatchIsInvalid(settings, name) {
			reportIdMatch(ctx, node, settings)
		}

	// A class field's name, public. The private spelling is a different node kind and has its own
	// listener above.
	case parent.Kind == ast.KindPropertyDeclaration:
		if settings.CheckClassFields && idMatchIsInvalid(settings, name) {
			reportIdMatch(ctx, node, settings)
		}

	// Everything else, gated by `onlyDeclarations` and by the callee exclusion.
	default:
		if idMatchShouldReport(idMatchEffectiveParentOf(node), name, settings) {
			reportIdMatch(ctx, node, settings)
		}
	}
}

// idMatchEffectiveParentOf is upstream's `effectiveParent`: for an identifier inside a member
// access, the node ABOVE the access, and otherwise the immediate parent.
//
// It exists so that `foo.bar = 1` can be judged by the assignment rather than by the access.
func idMatchEffectiveParentOf(node *ast.Node) *ast.Node {
	parent := node.Parent
	if parent != nil && parent.Kind == ast.KindPropertyAccessExpression {
		return parent.Parent
	}
	return parent
}

// idMatchShouldReport is upstream's `shouldReport`.
//
// Two gates and the pattern test. `onlyDeclarations` narrows to the two shapes that introduce a
// name into a scope, and the callee exclusion drops a name that was declared somewhere else.
func idMatchShouldReport(effectiveParent *ast.Node, name string, settings IdMatchSettings) bool {
	if effectiveParent == nil {
		return false
	}
	if settings.OnlyDeclarations && !idMatchIsDeclaration(effectiveParent) {
		return false
	}
	if idMatchIsAllowedParent(effectiveParent) {
		return false
	}
	return idMatchIsInvalid(settings, name)
}

// idMatchIsDeclaration is upstream's DECLARATION_TYPES: a function declaration or a variable
// declarator.
func idMatchIsDeclaration(node *ast.Node) bool {
	return node.Kind == ast.KindFunctionDeclaration || node.Kind == ast.KindVariableDeclaration
}

// idMatchIsAllowedParent is upstream's ALLOWED_PARENT_TYPES: a call or a construction, whose callee
// names something declared elsewhere.
func idMatchIsAllowedParent(node *ast.Node) bool {
	return node.Kind == ast.KindCallExpression || node.Kind == ast.KindNewExpression
}

// checkIdMatchMemberAccess is upstream's MemberExpression branch, all three arms.
func checkIdMatchMemberAccess(ctx rule.Context, node *ast.Node, access *ast.Node, name string,
	settings IdMatchSettings) {

	if !settings.CheckProperties {
		return
	}

	expression := access.AsPropertyAccessExpression()
	if expression == nil {
		return
	}

	// Arm one: the OBJECT of the access is always checked, because it is a plain reference to a
	// name and nothing about the access makes it a property. Upstream spells this
	// `parent.object.type === "Identifier" && parent.object.name === name`, which is a name
	// comparison rather than a node identity one, so `foo.foo` satisfies it for BOTH identifiers.
	// That looks like a bug and is upstream's behaviour, so it is reproduced: measured on
	// `no_under.no_under` with properties on, the installed rule reports twice.
	if expression.Expression != nil && expression.Expression.Kind == ast.KindIdentifier &&
		expression.Expression.Text() == name {
		if idMatchIsInvalid(settings, name) {
			reportIdMatch(ctx, node, settings)
		}
		return
	}

	effectiveParent := idMatchEffectiveParentOf(node)
	if effectiveParent == nil || effectiveParent.Kind != ast.KindBinaryExpression {
		return
	}
	assignment := effectiveParent.AsBinaryExpression()
	if assignment == nil || assignment.OperatorToken.Kind != ast.KindEqualsToken {
		return
	}

	// Arm two: the property on the LEFT of an assignment, which is a write and therefore names
	// something we are creating.
	if assignment.Left != nil && assignment.Left.Kind == ast.KindPropertyAccessExpression {
		if left := assignment.Left.AsPropertyAccessExpression(); left != nil &&
			left.Name() != nil && left.Name().Text() == name {
			if idMatchIsInvalid(settings, name) {
				reportIdMatch(ctx, node, settings)
			}
			return
		}
	}

	// Arm three: an assignment whose RIGHT side is not itself a member access. This is what stops
	// `foo.no_under = boom.bam_pow` from reporting `bam_pow`: reading a property off an object we
	// may not own is not our naming decision.
	if assignment.Right != nil && assignment.Right.Kind != ast.KindPropertyAccessExpression {
		if idMatchIsInvalid(settings, name) {
			reportIdMatch(ctx, node, settings)
		}
	}
}

// checkIdMatchBindingElement judges a name inside a DECLARATION destructuring, which is upstream's
// `Property` inside an `ObjectPattern` and also its `AssignmentPattern`.
//
// One `KindBindingElement` carries what ESTree splits over three fields:
//
//	let { a }        PropertyName nil,  Name() = a     upstream's shorthand
//	let { a: b }     PropertyName = a,  Name() = b     upstream's key and value
//	let { a = 1 }    PropertyName nil,  Name() = a     shorthand with a default
//	let { ...rest }  PropertyName nil,  Name() = rest, DotDotDotToken set
//
// So "did this pattern coin a new name", which is what `ignoreDestructuring` turns on, is exactly
// "does this element rename or spread".
func checkIdMatchBindingElement(ctx rule.Context, node *ast.Node, element *ast.Node, name string,
	settings IdMatchSettings) {

	binding := element.AsBindingElement()
	if binding == nil {
		return
	}

	// Only the BINDING is ours. In `let { a: b } = x` the key `a` is the source object's own
	// property name and we did not choose it, so it is never checked even when it fails the
	// pattern -- measured against the installed rule, `var { bad_key: goodName } = q` is clean.
	//
	// One test rather than two. An earlier draft also returned early on `PropertyName == node`,
	// which reads as the explicit statement of that exclusion and is entirely redundant with this
	// line: a key is by definition not the name, so it exits here anyway. The mutation sweep is
	// what established the redundancy -- deleting the PropertyName guard alone survived every
	// fixture, while deleting both together is caught by seven.
	if binding.Name() != node {
		return
	}

	// An array pattern's elements are positional, so nothing about them comes from a property name
	// and upstream reaches them through its default arm rather than its Property branch.
	if element.Parent != nil && element.Parent.Kind == ast.KindArrayBindingPattern {
		if idMatchShouldReport(idMatchEffectiveParentOf(node), name, settings) {
			reportIdMatch(ctx, node, settings)
		}
		return
	}

	if settings.IgnoreDestructuring {
		// A rename or a rest element invents a name, and inventing is what the option does NOT
		// excuse. A plain shorthand merely re-uses the object's own key, which it does.
		//
		// "Renamed" is a comparison of the two NAMES rather than of the two slots, which is
		// upstream's `parent.key.name === parent.value.name`. The distinguishing case is in the
		// corpus: `let { category_id: category_id } = query` writes the key out longhand while
		// coining nothing, and it is CLEAN under this option. A port testing only whether
		// PropertyName is present reports it, and the fixture that catches that is the only one in
		// a hundred that can.
		coinsANewName := binding.DotDotDotToken != nil ||
			(binding.PropertyName != nil && binding.PropertyName.Kind == ast.KindIdentifier &&
				binding.PropertyName.Text() != name)
		if !coinsANewName {
			return
		}
	}

	// `properties` gates a destructured name the same way it gates an object key: upstream's
	// `!checkProperties && !parent.computed` returns before the report. A rest element is exempt
	// from that gate, because a rest name is never a property name.
	if !settings.CheckProperties && binding.DotDotDotToken == nil {
		return
	}

	if settings.OnlyDeclarations && !idMatchIsDeclaration(idMatchEnclosingDeclarationOf(element)) {
		return
	}
	if idMatchIsInvalid(settings, name) {
		reportIdMatch(ctx, node, settings)
	}
}

// idMatchEnclosingDeclarationOf walks out of a binding pattern to whatever declares it.
//
// Upstream reads `effectiveParent`, which for a destructured name is the enclosing
// VariableDeclarator, because in ESTree the pattern hangs directly off it. Here the pattern is
// several nodes deep, so the walk is explicit.
func idMatchEnclosingDeclarationOf(element *ast.Node) *ast.Node {
	for current := element; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindVariableDeclaration, ast.KindFunctionDeclaration, ast.KindParameter:
			return current
		}
	}
	return nil
}

// checkIdMatchAssignmentPattern judges a name inside an ASSIGNMENT destructuring, which our parser
// spells with object-literal nodes rather than binding nodes.
//
//	({ a } = x)       KindShorthandPropertyAssignment
//	({ a: b } = x)    KindPropertyAssignment, name a, initializer b
func checkIdMatchAssignmentPattern(ctx rule.Context, node *ast.Node, property *ast.Node, name string,
	settings IdMatchSettings) {

	objectLiteral := property.Parent
	if objectLiteral == nil || !idDenylistIsDestructuringTarget(objectLiteral) {
		return
	}

	renamed := false
	switch property.Kind {
	case ast.KindPropertyAssignment:
		assignment := property.AsPropertyAssignment()
		if assignment == nil {
			return
		}
		// The source object's own key. Upstream's `!assignmentKeyEqualsValue && parent.key === node`
		// returns before any report.
		if assignment.Name() == node {
			return
		}
		if assignment.Initializer != node {
			return
		}
		// Upstream's `assignmentKeyEqualsValue` compares the two NAMES, so writing the key out
		// longhand with the same spelling coins nothing. See the binding-element arm for the
		// corpus case that pins this.
		renamed = assignment.Name() == nil || assignment.Name().Kind != ast.KindIdentifier ||
			assignment.Name().Text() != name

	case ast.KindShorthandPropertyAssignment:
		if assignment := property.AsShorthandPropertyAssignment(); assignment == nil ||
			assignment.Name() != node {
			return
		}
	}

	if settings.IgnoreDestructuring && !renamed {
		return
	}

	// No `properties` gate and no `onlyDeclarations` gate, and both absences were MEASURED rather
	// than reasoned. An earlier draft carried both, on the strength of the binding-declaration arm
	// carrying them, and it was wrong on four shapes at once. Against the installed rule:
	//
	//	({ no_under } = bar)                             1 finding, no options at all
	//	({ no_under } = bar)   properties: false          1 finding
	//	({ no_under } = bar)   onlyDeclarations: true     1 finding
	//	({ no_under } = bar)   ignoreDestructuring: true  clean
	//
	// The reason is in upstream's control flow rather than in its intent. Its `checkProperties`
	// and `onlyDeclarations` gates sit AFTER the ObjectPattern block, and that block reports and
	// falls through rather than returning, so a name reported there has already been reported by
	// the time either gate is consulted. Reproducing the gates here reproduced a structure upstream
	// does not have.
	if idMatchIsInvalid(settings, name) {
		reportIdMatch(ctx, node, settings)
	}
}

// idMatchIsLocalImportBinding answers upstream's `parent.local && parent.local.name === node.name`:
// the name this file binds, rather than the one the exporting module chose.
func idMatchIsLocalImportBinding(node *ast.Node, parent *ast.Node) bool {
	switch parent.Kind {
	case ast.KindImportSpecifier:
		// For `import { a as b }`, PropertyName is `a` (the exported name) and Name() is `b`.
		specifier := parent.AsImportSpecifier()
		return specifier != nil && specifier.Name() == node

	case ast.KindImportClause:
		// `import foo from 'mod'`, upstream's ImportDefaultSpecifier.
		clause := parent.AsImportClause()
		return clause != nil && clause.Name() == node

	case ast.KindNamespaceImport:
		// `import * as foo from 'mod'`, upstream's ImportNamespaceSpecifier.
		namespace := parent.AsNamespaceImport()
		return namespace != nil && namespace.Name() == node
	}
	return false
}
