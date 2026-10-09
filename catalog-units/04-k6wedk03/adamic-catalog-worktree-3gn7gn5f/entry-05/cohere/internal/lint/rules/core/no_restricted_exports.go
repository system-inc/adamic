package core

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	esregexp "github.com/system-inc/cohere/internal/lint/ecmascript/regexp"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// messageNoRestrictedExportsRestrictedNamedId is the id for a name the project has banned from
// being exported.
const messageNoRestrictedExportsRestrictedNamedId = "restrictedNamed"

// messageNoRestrictedExportsRestrictedDefault is the id for a default export the project has banned.
//
// A constant rather than a builder, because unlike the named message nothing moves in it: upstream's
// template is `Exporting 'default' is restricted.` with no interpolation at all.
var messageNoRestrictedExportsRestrictedDefault = rule.Message{
	Id: "restrictedDefault",
	Description: "Exporting `default` is restricted here. A default export has no name at the " +
		"point of import, so every importer gets to pick its own, and the same module ends up " +
		"referred to by three spellings across the codebase. Export it by name instead.",
}

// noRestrictedExportsRestrictedNamedMessage renders the finding for a banned exported name.
//
// The name is interpolated because a reader needs to know WHICH of the configured names this
// export matched, and with a pattern configured the matching name is not in the config at all.
func noRestrictedExportsRestrictedNamedMessage(name string) rule.Message {
	return rule.Message{
		Id: messageNoRestrictedExportsRestrictedNamedId,
		Description: fmt.Sprintf("`%s` is restricted from being used as an exported name. The "+
			"project has banned this name at its module boundary, usually because it collides "+
			"with something ambient or reads as a different thing than it is. Rename the export, "+
			"or export it under an alias that says what a caller is getting.", name),
	}
}

// NoRestrictedExportsRestrictDefaultExports carries upstream's five booleans, each naming one
// syntactic way a module can export a default.
//
// They are separate options rather than one switch because the five are not equally objectionable
// and a project usually wants some of them. Each is a pointer so an absent key stays distinguishable
// from an explicit `false`, and every one of them defaults to off.
type NoRestrictedExportsRestrictDefaultExports struct {
	// Direct bans `export default foo;`, `export default 42;` and `export default function foo() {}`.
	Direct *bool `json:"direct"`

	// Named bans `export { foo as default };` with no source module.
	Named *bool `json:"named"`

	// DefaultFrom bans `export { default } from "mod";` and `export { default as default } from "mod";`
	DefaultFrom *bool `json:"defaultFrom"`

	// NamedFrom bans `export { foo as default } from "mod";`
	NamedFrom *bool `json:"namedFrom"`

	// NamespaceFrom bans `export * as default from "mod";`
	NamespaceFrom *bool `json:"namespaceFrom"`
}

// NoRestrictedExportsOptions is the rule's whole configuration.
//
// Upstream's schema is an `anyOf` over two object shapes whose only difference is that the second
// forbids the literal string `default` inside `restrictedNamedExports` while permitting
// `restrictDefaultExports` beside it. That constraint is checked in the decoder rather than
// expressed in the type, because Go has no shape for "this array may not contain this string".
type NoRestrictedExportsOptions struct {
	// RestrictedNamedExports is the list of exported names to ban outright.
	RestrictedNamedExports []string `json:"restrictedNamedExports"`

	// RestrictedNamedExportsPattern bans every exported name matching this JavaScript regular
	// expression. Applied under the `u` flag, which is upstream's.
	RestrictedNamedExportsPattern string `json:"restrictedNamedExportsPattern"`

	// RestrictDefaultExports names which of the five default-export forms are banned.
	RestrictDefaultExports *NoRestrictedExportsRestrictDefaultExports `json:"restrictDefaultExports"`
}

// DecodeNoRestrictedExportsOptions reads this rule's configuration from the config layer.
//
// Hand-rolled for two reasons the generic decoder cannot serve. Empty input is legal and means an
// unconfigured rule that enforces nothing, where `rule.DecodeOptionsInto` errors. And upstream's
// schema constraint -- `default` may appear in `restrictedNamedExports` only when
// `restrictDefaultExports` is absent -- is a cross-field rule with no struct-tag spelling.
//
// The pattern is compiled here rather than per file so a project that misspells it learns at
// configuration time. Upstream compiles per report and would throw at lint time instead; failing
// earlier is strictly better and changes no verdict.
func DecodeNoRestrictedExportsOptions(raw []byte) (any, error) {
	var options NoRestrictedExportsOptions
	if len(raw) == 0 {
		return options, nil
	}
	if err := rule.UnmarshalOptions(raw, &options); err != nil {
		return options, err
	}

	if options.RestrictedNamedExportsPattern != "" {
		if _, err := esregexp.Compile(options.RestrictedNamedExportsPattern, "u"); err != nil {
			return options, fmt.Errorf(
				"no-restricted-exports: restrictedNamedExportsPattern %q is not a valid regular "+
					"expression: %w", options.RestrictedNamedExportsPattern, err)
		}
	}

	// Upstream's second schema arm carries `pattern: "^(?!default$)"` on the array's items, which
	// is how it says the two cannot be combined. Listing `default` under restrictedNamedExports
	// already bans every named form, so pairing it with the flags that ban the individual forms is
	// a configuration that contradicts itself about which message a reader should see.
	if options.RestrictDefaultExports != nil {
		for _, name := range options.RestrictedNamedExports {
			if name == "default" {
				return options, fmt.Errorf(
					"no-restricted-exports: `default` cannot be listed in restrictedNamedExports " +
						"alongside restrictDefaultExports, because the list already bans every " +
						"named default form the flags select between")
			}
		}
	}
	return options, nil
}

// NoRestrictedExports reports an export whose name a project has banned.
//
//	valid:   export var a;                        with nothing configured
//	valid:   export { b as a } from 'foo';        with restrictedNamedExports: ['b']
//	valid:   export default foo;                  unless restrictDefaultExports.direct
//	invalid: export var a;                        with restrictedNamedExports: ['a']
//	invalid: export let { b: { c: a } } = {};     with restrictedNamedExports: ['a']
//	invalid: export * as a from 'a';              with restrictedNamedExports: ['a']
//	invalid: export { foo as default } from 'm';  with restrictDefaultExports.namedFrom
//
// # This rule enforces nothing until somebody configures it
//
// With no options at all upstream reads `context.options[0]` as `{}` and every check falls through,
// so the rule is silent on every file. That is worth saying plainly because it makes an audit's
// violation count structurally zero rather than evidence of a clean tree: measured through the
// installed ESLint 10.8.1, the 95 cases upstream annotates valid include `export var a;` under no
// options at all.
//
// So it is registered here and NOT enabled. Enabling it means choosing which names a project bans,
// which is a decision about this codebase rather than a porting question.
//
// # The exported name is the one the importer types, never the local one
//
// Every check in this rule reads the name a consumer of the module would write, and that is the
// distinction the corpus spends most of its cases on. `export { b as a }` exports `a` and binds `b`,
// so `restrictedNamedExports: ['b']` is CLEAN and `['a']` reports; `export { a as b }` is the
// mirror. Measured across sixteen corpus cases that differ only in which side of `as` carries the
// restricted name.
//
// This tree spells that pair the opposite way round from the field names in ESTree, which is worth
// stating because reading either one as "the name" gives a rule that is exactly backwards on every
// aliased export. On an `ExportSpecifier` here, `PropertyName` is the LOCAL name and `Name()` is the
// EXPORTED one, and `PropertyName` is nil for the unaliased `export { a }`. Measured with an
// ancestry probe over `export { b as 'a' } from 'foo'`.
//
// # A string literal is a legal export name and carries the cooked value
//
// `export { b as 'a' }` and `export { b as 'a' }` both export the name `a`, and the corpus
// tests both. `Node.Text()` answers with the cooked value for a string literal as well as for an
// identifier, so the two shapes need no separate arm. What they DO need is that the empty string and
// a string with spaces are real names: `export { ” } from 'foo'` under `restrictedNamedExports:
// [”]` reports, and under `[' ']` is clean. A port comparing against a zero value rather than
// against the configured list gets both of those wrong.
//
// # A destructured export declares every name the pattern binds
//
// `export let { b: { c: a = d } = e } = {};` exports exactly one name, `a`, and neither `b`, `c`,
// `d` nor `e`. Upstream reaches that through `sourceCode.getDeclaredVariables(declaration)`, which
// is the scope analyser's answer to "what did this statement bind". There is no scope analyser in
// this walk, so the equivalent is a walk of the binding pattern collecting the names in binding
// POSITION, and the two agree on all fourteen destructuring cases in the corpus.
//
// The shapes that separate a correct walk from a plausible one are all in the corpus, and each of
// them is a place a name appears without being bound:
//
//	export var { a: b } = {};        binds `b`; `a` is a property key
//	export let { b = a } = {};       binds `b`; `a` is a default VALUE
//	export const [b] = [a];          binds `b`; `a` is on the initializer side
//	export var [a = a] = [];         binds `a` once, not twice
//
// The last is the sharpest: the same identifier appears twice, once as a binding and once as the
// default it falls back to, and upstream reports once.
//
// # `export default` reaches this rule as two unrelated node kinds
//
// `export default foo;` is a `KindExportAssignment`, while `export default function foo() {}` is an
// ordinary `KindFunctionDeclaration` carrying both an export and a default modifier. There is no
// single node for "a default export" to listen on, so both arms are needed and a port listening only
// on the first is silent on half of upstream's `direct` cases. Measured on both shapes.
//
// `export = foo` is TypeScript's own construct and shares `KindExportAssignment` with the default
// export, separated by `IsExportEquals`. It is not a default export and upstream, which has no such
// syntax, has nothing to say about it.
var NoRestrictedExports = rule.Rule{
	Name: "no-restricted-exports",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, configured := rule.OptionsAs[NoRestrictedExportsOptions](options)
		if !configured {
			return rule.Listeners{}
		}

		restricted := map[string]bool{}
		for _, name := range settings.RestrictedNamedExports {
			restricted[name] = true
		}
		var pattern *esregexp.RegExp
		if settings.RestrictedNamedExportsPattern != "" {
			// The decoder already refused an uncompilable pattern, so an error here would mean the
			// rule was handed options that never went through it. Falling back to no pattern is the
			// quiet answer and the wrong one; there is nothing to report against, so the rule simply
			// does not apply the pattern rather than pretending it matched nothing.
			compiled, err := esregexp.Compile(settings.RestrictedNamedExportsPattern, "u")
			if err == nil {
				pattern = compiled
			}
		}

		// Upstream returns its listeners regardless of whether anything is configured, and every
		// check inside them falls through. Declining here instead is the same verdict for less
		// work, and it matches what the sibling `no-restricted-properties` already does.
		if len(restricted) == 0 && pattern == nil && settings.RestrictDefaultExports == nil {
			return rule.Listeners{}
		}

		checker := &noRestrictedExportsChecker{
			ctx:        ctx,
			restricted: restricted,
			pattern:    pattern,
			defaults:   settings.RestrictDefaultExports,
		}

		return rule.Listeners{
			// `export var a;`, `export function a() {}`, `export class A {}` -- a declaration
			// carrying an export modifier, which is upstream's ExportNamedDeclaration WITH a
			// declaration.
			ast.KindVariableStatement:    checker.checkDeclarationStatement,
			ast.KindFunctionDeclaration:  checker.checkDeclarationStatement,
			ast.KindClassDeclaration:     checker.checkDeclarationStatement,
			ast.KindInterfaceDeclaration: checker.checkDeclarationStatement,
			ast.KindTypeAliasDeclaration: checker.checkDeclarationStatement,
			ast.KindEnumDeclaration:      checker.checkDeclarationStatement,
			ast.KindModuleDeclaration:    checker.checkDeclarationStatement,

			// `export { a }`, `export { a } from 'm'`, `export * as a from 'm'`, `export * from 'm'`
			// -- upstream's ExportNamedDeclaration without a declaration, plus ExportAllDeclaration.
			ast.KindExportDeclaration: checker.checkExportDeclaration,

			// `export default foo;`
			ast.KindExportAssignment: checker.checkExportAssignment,
		}
	},
}

// noRestrictedExportsChecker carries the resolved configuration across the listeners.
type noRestrictedExportsChecker struct {
	ctx        rule.Context
	restricted map[string]bool
	pattern    *esregexp.RegExp
	defaults   *NoRestrictedExportsRestrictDefaultExports
}

// checkDeclarationStatement handles a declaration that carries its own export modifier.
//
// This is upstream's `ExportNamedDeclaration` arm for the case where `node.declaration` is set, plus
// its `ExportDefaultDeclaration` arm for the declaration forms, because this tree does not separate
// the two: `export default function foo() {}` is a function declaration with two modifiers on it,
// not a distinct node.
func (c *noRestrictedExportsChecker) checkDeclarationStatement(node *ast.Node) {
	if !noRestrictedExportsHasModifier(node, ast.KindExportKeyword) {
		return
	}

	if noRestrictedExportsHasModifier(node, ast.KindDefaultKeyword) {
		// `export default function foo() {}`. Upstream's ExportDefaultDeclaration arm reports the
		// whole statement and never consults the name, so `restrictedNamedExports: ['foo']` is
		// clean on it while `restrictDefaultExports.direct` reports. Both measured.
		c.reportDefaultIf(node, c.defaults != nil && noRestrictedExportsIsOn(c.defaults.Direct))
		return
	}

	if node.Kind == ast.KindVariableStatement {
		// Upstream maps `getDeclaredVariables` over the declaration and checks each bound name.
		// Every declarator in the list contributes, so `export const b = 1, a = 2;` checks both.
		list := node.AsVariableStatement().DeclarationList
		if list == nil {
			return
		}
		declarations := list.AsVariableDeclarationList().Declarations
		if declarations == nil {
			return
		}
		for _, declaration := range declarations.Nodes {
			name := declaration.Name()
			if name == nil {
				continue
			}
			noRestrictedExportsForEachBoundName(name, func(bound *ast.Node) {
				c.checkExportedName(bound)
			})
		}
		return
	}

	// A function, class, interface, type alias, enum or namespace exports exactly its own name.
	if name := node.Name(); name != nil {
		c.checkExportedName(name)
	}
}

// checkExportDeclaration handles every `export ... from`-shaped statement and the bare
// `export { ... }`.
//
// This is upstream's ExportNamedDeclaration-without-declaration arm and its ExportAllDeclaration arm
// at once, because this tree gives both the same node kind and separates them by what hangs off it:
// a `KindNamedExports` child for the specifier list, a `KindNamespaceExport` child for
// `export * as a`, and neither for the bare `export * from 'm'`.
func (c *noRestrictedExportsChecker) checkExportDeclaration(node *ast.Node) {
	declaration := node.AsExportDeclaration()
	if declaration == nil || declaration.ExportClause == nil {
		// `export * from 'foo';` exports no name of its own. Upstream's ExportAllDeclaration arm
		// guards on `node.exported`, which is null here, so it checks nothing. The corpus tests
		// `export * from 'a'` under `restrictedNamedExports: ['a']` and it is clean: `a` there is
		// the module, not an exported name.
		return
	}

	switch declaration.ExportClause.Kind {
	case ast.KindNamespaceExport:
		// `export * as a from 'm'`. Upstream's ExportAllDeclaration arm with `node.exported` set.
		if name := declaration.ExportClause.AsNamespaceExport().Name(); name != nil {
			c.checkExportedName(name)
		}

	case ast.KindNamedExports:
		elements := declaration.ExportClause.AsNamedExports().Elements
		if elements == nil {
			return
		}
		for _, specifier := range elements.Nodes {
			if name := specifier.Name(); name != nil {
				c.checkExportedName(name)
			}
		}
	}
}

// checkExportAssignment handles `export default <expression>;`.
//
// `export = foo` shares this node kind and is not a default export. Upstream has no such syntax and
// therefore no opinion, so it is skipped rather than reported.
func (c *noRestrictedExportsChecker) checkExportAssignment(node *ast.Node) {
	assignment := node.AsExportAssignment()
	if assignment == nil || assignment.IsExportEquals {
		return
	}
	c.reportDefaultIf(node, c.defaults != nil && noRestrictedExportsIsOn(c.defaults.Direct))
}

// checkExportedName is upstream's `checkExportedName`, and the order of its arms decides which
// message a reader sees.
//
// A name matching the restricted list or the pattern reports `restrictedNamed` and RETURNS, so the
// default arms below never run for it. That ordering is upstream's and it is observable: with both
// `restrictedNamedExports: ['default']` and a `restrictDefaultExports` flag configured, the named
// message wins. Upstream's schema forbids exactly that combination, which is why the decoder here
// refuses it too.
func (c *noRestrictedExportsChecker) checkExportedName(node *ast.Node) {
	name := noRestrictedExportsNameText(node)

	// The pattern deliberately does not apply to `default`. Upstream guards with
	// `name !== "default"` before testing, so `restrictedNamedExportsPattern: 'default'` is clean on
	// every one of the five default forms -- five corpus cases turn on this, and a port that tested
	// the pattern unconditionally would report all five under the wrong message.
	matchesPattern := false
	if c.pattern != nil && name != "default" {
		matchesPattern = c.pattern.Test(name)
	}

	if matchesPattern || c.restricted[name] {
		c.ctx.ReportNode(node, noRestrictedExportsRestrictedNamedMessage(name))
		return
	}

	if name != "default" || c.defaults == nil {
		return
	}

	parent := node.Parent
	if parent == nil {
		return
	}

	if parent.Kind == ast.KindNamespaceExport {
		// `export * as default from 'mod';`
		c.reportDefaultIf(node, noRestrictedExportsIsOn(c.defaults.NamespaceFrom))
		return
	}

	if parent.Kind != ast.KindExportSpecifier {
		return
	}

	// Which of the three specifier flags applies turns on two questions: whether the statement has
	// a source module, and what the LOCAL side of the specifier is. Upstream reads the local name
	// through `node.parent.local`, which for an unaliased `export { default }` is the same node as
	// the exported name.
	hasSource := noRestrictedExportsSpecifierHasSource(parent)
	localName := name
	if propertyName := parent.AsExportSpecifier().PropertyName; propertyName != nil {
		localName = noRestrictedExportsNameText(propertyName)
	}

	if !hasSource {
		// `export { foo as default };` -- a local binding re-exported as the default.
		c.reportDefaultIf(node, noRestrictedExportsIsOn(c.defaults.Named))
		return
	}

	if localName == "default" {
		// `export { default } from 'mod';` and `export { default as default } from 'mod';`
		c.reportDefaultIf(node, noRestrictedExportsIsOn(c.defaults.DefaultFrom))
		return
	}
	// `export { foo as default } from 'mod';`
	c.reportDefaultIf(node, noRestrictedExportsIsOn(c.defaults.NamedFrom))
}

// reportDefaultIf reports the restricted-default message when the flag that selects this form is on.
func (c *noRestrictedExportsChecker) reportDefaultIf(node *ast.Node, restricted bool) {
	if restricted {
		c.ctx.ReportNode(node, messageNoRestrictedExportsRestrictedDefault)
	}
}

// noRestrictedExportsIsOn reads an optional boolean, where absent means off.
//
// Every one of upstream's five flags defaults to off, so a nil pointer and an explicit `false` mean
// the same thing here. They are still stored as pointers, because a decoder that collapsed them
// would be unable to tell a configuration that names a flag from one that does not, and the next
// option added to this rule may not default to off.
func noRestrictedExportsIsOn(flag *bool) bool {
	return flag != nil && *flag
}

// noRestrictedExportsNameText reads the name an identifier or a string literal spells.
//
// A string literal is a legal module export name -- `export { b as 'a' }` -- and `Text()` answers
// with the COOKED value for one, so `'a'` reads as `a` exactly as upstream's
// `getModuleExportName` does. The empty string is a legal name too, so an unreadable node and an
// empty name are not distinguished here; the caller compares against a configured list where `""`
// may legitimately appear.
func noRestrictedExportsNameText(node *ast.Node) string {
	if node == nil {
		return ""
	}
	switch node.Kind {
	case ast.KindIdentifier, ast.KindStringLiteral:
		return node.Text()
	}
	return ""
}

// noRestrictedExportsSpecifierHasSource answers whether the export statement holding this specifier
// names a module to re-export from.
//
// Upstream reads `node.parent.parent.source`. Here the specifier's grandparent is the export
// declaration, since `KindNamedExports` sits between them.
func noRestrictedExportsSpecifierHasSource(specifier *ast.Node) bool {
	namedExports := specifier.Parent
	if namedExports == nil {
		return false
	}
	declaration := namedExports.Parent
	if declaration == nil || declaration.Kind != ast.KindExportDeclaration {
		return false
	}
	return declaration.AsExportDeclaration().ModuleSpecifier != nil
}

// noRestrictedExportsHasModifier answers whether a declaration carries a given modifier keyword.
func noRestrictedExportsHasModifier(node *ast.Node, kind ast.Kind) bool {
	modifiers := node.Modifiers()
	if modifiers == nil {
		return false
	}
	for _, modifier := range modifiers.Nodes {
		if modifier.Kind == kind {
			return true
		}
	}
	return false
}

// noRestrictedExportsForEachBoundName visits every identifier a binding name puts into scope.
//
// This stands in for upstream's `sourceCode.getDeclaredVariables(declaration)`, which asks the scope
// analyser what a statement bound. There is no scope analyser in this walk, so the question is
// answered structurally instead, and the two agree on all fourteen destructuring cases in upstream's
// corpus.
//
// What makes that agreement possible is that a binding pattern separates its positions into distinct
// fields rather than distinct kinds, so the walk never has to guess. On a `KindBindingElement`:
//
//	PropertyName   the key being read from the object, `a` in `{ a: b }`     NOT bound
//	Name()         the binding, `b` in `{ a: b }` or `a` in `{ a }`          BOUND
//	Initializer    the default value, `a` in `{ b = a }`                     NOT bound
//
// So only `Name()` is followed, and it is followed RECURSIVELY, because a nested pattern is itself
// a name: `export var [{ a }] = [];` binds `a` two levels down. Reading only the top level reports
// nothing on four corpus cases, and reading every identifier in the subtree reports `a` twice on
// `export var [a = a] = [];`, where upstream reports once.
//
// A rest element (`{ ...rest }`) reaches this as a binding element whose `DotDotDotToken` is set and
// whose `Name()` is the bound identifier, so it needs no arm of its own.
func noRestrictedExportsForEachBoundName(name *ast.Node, visit func(bound *ast.Node)) {
	if name == nil {
		return
	}
	switch name.Kind {
	case ast.KindIdentifier:
		visit(name)

	case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
		elements := name.AsBindingPattern().Elements
		if elements == nil {
			return
		}
		for _, element := range elements.Nodes {
			if element.Kind == ast.KindOmittedExpression {
				// A hole in an array pattern, `export var [, a] = [];`. It binds nothing.
				continue
			}
			noRestrictedExportsForEachBoundName(element.Name(), visit)
		}
	}
}
