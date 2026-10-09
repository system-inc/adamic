package react

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/comments"
	"github.com/system-inc/cohere/internal/lint/ecmascript/imports"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoDeprecated flags a React API that React itself has deprecated.
//
//	valid:   React.createElement('p', {}, null)
//	valid:   ReactDOM.findDOMNode(instance)
//	valid:   class Foo { componentWillMount() {} }              (not a component)
//	valid:   var Foo = { componentWillMount: function() {} }     (not a component)
//	valid:   class Foo extends React.Component { UNSAFE_componentWillMount() {} }
//	invalid: React.renderComponent()
//	invalid: React.createClass({})
//	invalid: this.transferPropsTo()
//	invalid: import { createClass } from 'react'
//	invalid: var { createClass } = require('react')
//	invalid: class Foo extends React.Component { componentWillMount() {} }
//
// Ported from `react/no-deprecated` in `eslint-plugin-react`, which is the implementation that
// defined this rule. Every behavior recorded below was measured by driving the installed build,
// version 7.37.5, through the ESLint Linter API on inputs written for the question. All 55 imported
// corpus cases were run through that build first and it agreed with every one of them, so the
// corpus and the oracle are one authority here rather than two.
//
// # The version gate collapses, and that is the whole reason this rule is small
//
// Upstream guards every entry with `testReactVersion(context, '>= ' + since)`. When no React
// version is configured, `getReactVersionFromContext` falls back to `ULTIMATE_LATEST_SEMVER`, the
// literal `999.999.999` at `util/version.js:16`, so every comparison is true and every entry is in
// force. Auto-detection from `node_modules/react` runs only when the configuration says
// `version: "detect"` explicitly; absence does not trigger it.
//
// cohere reads no React settings (the loader refuses `settings.react`), so no React version can
// reach a rule here by any route. The faithful reading is to
// reproduce the answer upstream gives under this repository's configuration, and that answer is
// that all 29 entries fire. Measured directly: one file writing all 29 deprecated spellings under
// no settings produces 29 findings on the installed build.
//
// So the `since` version survives here only as the number the message prints. It decides nothing.
//
// # The member arm matches SOURCE TEXT, not a resolved path, and that is load-bearing
//
// Upstream's `MemberExpression` listener calls `checkDeprecation(node, getText(context, node))`,
// which is the member expression's own source slice. A map lookup on that string is not the same
// question as "does this expression denote React.renderComponent", and the difference is visible
// in four directions, all measured on the installed build and all silent there:
//
//	React . renderComponent()       spaces inside the member          silent
//	React\n  .renderComponent()     a newline inside the member       silent
//	React/*x*/.renderComponent()    a comment inside the member       silent
//	React['renderComponent']()      the computed spelling             silent
//	React?.renderComponent()        the optional-chain spelling       silent
//
// Every one of those denotes the deprecated API and upstream reports none of them. Reproduced
// rather than improved on: a port that resolved the path instead would start reporting five shapes
// upstream is silent about, and no imported fixture could see it because the corpus writes none of
// them. The source slice is compared verbatim, so our answer moves with upstream's for any
// whitespace or trivia a future file happens to contain.
//
// The nesting consequence is the same mechanism read forward. `React.PropTypes.component` reports
// TWICE, because the listener fires on the outer member and again on the inner `React.PropTypes`,
// and both slices are keys in the map. That is upstream's behavior, measured, and it is why the
// walk here visits every property access rather than only the outermost.
//
// # The pragma, which renames React for the whole file
//
// `pragmaUtil.getFromContext` scans EVERY comment in the file for `@jsx <name>`, takes the part
// before the first dot, and uses it in place of `React`. Four things about it were measured because
// the source leaves each ambiguous:
//
//	/** @jsx Foo */ Foo.renderComponent()      reports as Foo.renderComponent
//	/** @jsx Foo.bar */ Foo.renderComponent()  reports; only the part before the dot is taken
//	// @jsx Foo (a LINE comment)               reports; the scan is not restricted to block comments
//	Foo.renderComponent()\n/** @jsx Foo */     reports; position in the file does not matter
//	/** @jsx Foo */ React.renderComponent()    SILENT; the pragma REPLACES React rather than adding
//
// The last is the one a port is most likely to get wrong in the safe-looking direction, and it is
// the reason the table below is built per file rather than being a package-level constant.
//
// The settings-based pragma (`settings.react.pragma`) is unreachable here for the same reason the
// version is, so the comment scan is the only way the pragma moves.
//
// # The lifecycle arm, and the three ways it declines
//
// `checkLifeCycleMethods` runs only on a node `isES5Component` or `isES6Component` accepts, and
// then reads `getPropertyName` on each member. That name is `nameNode.name`, which exists on an
// identifier and does not exist on anything else, so two shapes are silent upstream:
//
//	class Foo extends React.Component { ['componentWillMount']() {} }   silent, computed key
//	class Foo extends React.Component { 'componentWillMount'() {} }     silent, string key
//
// Both measured. Neither is in the corpus. A port reading the key's text through a helper that
// cooks a string literal would report both and look more correct while diverging.
//
// What does NOT decline is anything about the member's kind. A getter, a static, and a class field
// holding a non-function all report, measured, so this arm tests the NAME and nothing else.
//
// And the ES5 factory name here is `createReactClass` alone, from `getCreateClassFromContext`'s
// default. `React.createClass({componentWillMount(){}})` therefore reports ONE finding, for the
// member expression `React.createClass`, and not the lifecycle method inside it. Measured; it is
// the case where the two arms most look like they should agree and do not.
//
// # What is deliberately not ported: the JSDoc component
//
// `isES6Component` opens with `isExplicitComponent`, which reads a JSDoc `@extends React.Component`
// or `@augments React.Component` tag through `sourceCode.getJSDocComment`. Four inputs written for
// it are all silent on the installed build:
//
//	/** @extends React.Component */ class F { componentWillMount() {} }   silent
//	/** @augments React.Component */ class F { componentWillMount() {} }  silent
//
// So the branch is not reproduced, and this paragraph records the measurement rather than an
// argument: the command was the same Linter drive every other line here cites, and it reported
// nothing for either tag. If a later reader finds an arrangement where it does fire, this is the
// note that says the absence here was measured on two shapes and not proved over all of them.
var NoDeprecated = rule.Rule{
	// No namespace prefix. The config writes `react/no-deprecated` and the parity guard strips the
	// namespace on a `/` boundary, so `react-no-deprecated` would match no inventory entry and lint
	// no files while passing every fixture in this package.
	Name: "react/no-deprecated",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// The pragma is read once per file rather than per node, because it is a property of the
		// file's comments and cannot change between two nodes in it. It also has to be read before
		// the first node is judged, which is why the table is built here and not lazily.
		pragma := reactPragmaFor(ctx)
		deprecations := defaultPragmaDeprecations
		if pragma != defaultPragma {
			deprecations = deprecationsForPragma(pragma)
		}

		report := func(node *ast.Node, name string) {
			entry, deprecated := deprecations[name]
			if !deprecated {
				return
			}
			ctx.Report(rule.Diagnostic{
				Range:      rule.TokenRange(ctx.SourceFile, node),
				Message:    rule.Message{Id: noDeprecatedMessageId, Description: entry.describe(name)},
				SourceFile: ctx.SourceFile,
			})
		}

		return rule.Listeners{
			// Every property access, not only the outermost. `React.PropTypes.component` reports
			// twice upstream and the inner finding is only reachable by visiting the inner node.
			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				report(node, sourceSliceOf(ctx, node))
			},

			ast.KindImportDeclaration: func(node *ast.Node) {
				moduleName, named := deprecatedModuleAlias(node)
				if !named {
					return
				}
				forEachNamedImportSpecifier(node, func(specifier *ast.Node, importedName string) {
					report(specifier, moduleName+"."+importedName)
				})
			},

			// A destructuring binding, from either `require('react')` or a bare React-ish
			// identifier. Upstream anchors this on `VariableDeclarator`; ours is the same node
			// under a different name.
			ast.KindVariableDeclaration: func(node *ast.Node) {
				moduleName, named := destructuredReactModuleName(node, pragma)
				if !named {
					return
				}
				forEachDestructuredProperty(node, func(property *ast.Node, key string) {
					report(property, moduleName+"."+key)
				})
			},

			ast.KindClassDeclaration:        reportLifeCycleMethods(report),
			ast.KindClassExpression:         reportLifeCycleMethods(report),
			ast.KindObjectLiteralExpression: reportLifeCycleMethods(report),
		}
	},
}

// noDeprecatedMessageId is the single id this rule reports under.
//
// Upstream has one `messages` entry and interpolates four slots into it. The id is therefore not
// enough to tell two findings apart, which is why the fixtures for this rule assert the rendered
// Description rather than only the id: a mutation moving a slot leaves the id and the count fixed.
const noDeprecatedMessageId = "deprecated"

// deprecation is one row of upstream's table.
//
// `since` is printed and decides nothing, for the reason in the rule's doc comment. `replacement`
// and `reference` are empty when upstream's array is shorter than three, and each empty one drops
// its whole clause from the message rather than printing an empty one.
type deprecation struct {
	since       string
	replacement string
	reference   string
}

// describe renders the finding exactly as upstream's message template does.
//
// The template is `{{oldMethod}} is deprecated since React {{version}}{{newMethod}}{{refs}}`, where
// `newMethod` is `, use X instead` or the empty string and `refs` is `, see Y` or the empty string.
// Those two conditionals live in the report call upstream rather than in the template, so a port
// interpolating four raw slots would print a stray comma on the two entries that carry no
// replacement. `React.isValidClass` and `React.addons.LinkedStateMixin` are those two, and both are
// pinned by a fixture asserting this whole string.
func (entry deprecation) describe(oldMethod string) string {
	description := fmt.Sprintf("%s is deprecated since React %s", oldMethod, entry.since)
	if entry.replacement != "" {
		description += ", use " + entry.replacement + " instead"
	}
	if entry.reference != "" {
		description += ", see " + entry.reference
	}
	return description
}

// The three reference URLs upstream repeats, named once each.
//
// Upstream builds the lifecycle ones by concatenating a documentation anchor with a codemod
// sentence, and the trailing space inside the first half is part of the rendered message rather
// than a typo in this file. Copied byte for byte from `no-deprecated.js:73` and pinned by a fixture
// asserting the whole rendered string.
const (
	unsafeLifecycleCodemodNote = ". Use https://github.com/reactjs/react-codemod#rename-unsafe-lifecycles " +
		"to automatically update your components."
	createRootReference = "https://reactjs.org/link/switch-to-createroot"
)

// deprecationsForPragma builds upstream's table with the file's pragma substituted for React.
//
// Built per file rather than once at package scope because the pragma comes from the file's own
// comments, and a package-level map keyed on `React.` would go silent on every file that renames
// it and report on every file that keeps `React` while the pragma says otherwise. Both directions
// were measured; see the rule's doc comment.
//
// The rows and their order are upstream's, from `getDeprecated` at `no-deprecated.js:34`. The three
// entries whose key does not begin with the pragma (`this.transferPropsTo`, and the `ReactPerf` and
// `Perf` pairs) are upstream's too: those names are fixed and the pragma does not reach them.
// defaultPragma is the pragma a file has when no comment names another, which is nearly every file.
const defaultPragma = "React"

// defaultPragmaDeprecations is the table for the default pragma, built once rather than per file.
// Rebuilding it was most of this rule's time: 230ms of 296ms over ahra's 3,787 files. It is only
// read, so every file's walk can share it.
var defaultPragmaDeprecations = deprecationsForPragma(defaultPragma)

func deprecationsForPragma(pragma string) map[string]deprecation {
	return map[string]deprecation{
		// 0.12.0
		pragma + ".renderComponent":               {since: "0.12.0", replacement: pragma + ".render"},
		pragma + ".renderComponentToString":       {since: "0.12.0", replacement: pragma + ".renderToString"},
		pragma + ".renderComponentToStaticMarkup": {since: "0.12.0", replacement: pragma + ".renderToStaticMarkup"},
		pragma + ".isValidComponent":              {since: "0.12.0", replacement: pragma + ".isValidElement"},
		pragma + ".PropTypes.component":           {since: "0.12.0", replacement: pragma + ".PropTypes.element"},
		pragma + ".PropTypes.renderable":          {since: "0.12.0", replacement: pragma + ".PropTypes.node"},
		pragma + ".isValidClass":                  {since: "0.12.0"},
		"this.transferPropsTo":                    {since: "0.12.0", replacement: "spread operator ({...})"},

		// 0.13.0
		pragma + ".addons.classSet":       {since: "0.13.0", replacement: "the npm module classnames"},
		pragma + ".addons.cloneWithProps": {since: "0.13.0", replacement: pragma + ".cloneElement"},

		// 0.14.0
		pragma + ".render":                 {since: "0.14.0", replacement: "ReactDOM.render"},
		pragma + ".unmountComponentAtNode": {since: "0.14.0", replacement: "ReactDOM.unmountComponentAtNode"},
		pragma + ".findDOMNode":            {since: "0.14.0", replacement: "ReactDOM.findDOMNode"},
		pragma + ".renderToString":         {since: "0.14.0", replacement: "ReactDOMServer.renderToString"},
		pragma + ".renderToStaticMarkup":   {since: "0.14.0", replacement: "ReactDOMServer.renderToStaticMarkup"},

		// 15.0.0
		pragma + ".addons.LinkedStateMixin":   {since: "15.0.0"},
		"ReactPerf.printDOM":                  {since: "15.0.0", replacement: "ReactPerf.printOperations"},
		"Perf.printDOM":                       {since: "15.0.0", replacement: "Perf.printOperations"},
		"ReactPerf.getMeasurementsSummaryMap": {since: "15.0.0", replacement: "ReactPerf.getWasted"},
		"Perf.getMeasurementsSummaryMap":      {since: "15.0.0", replacement: "Perf.getWasted"},

		// 15.5.0
		pragma + ".createClass":      {since: "15.5.0", replacement: "the npm module create-react-class"},
		pragma + ".addons.TestUtils": {since: "15.5.0", replacement: "ReactDOM.TestUtils"},
		pragma + ".PropTypes":        {since: "15.5.0", replacement: "the npm module prop-types"},

		// 15.6.0
		pragma + ".DOM": {since: "15.6.0", replacement: "the npm module react-dom-factories"},

		// 16.9.0. Upstream's own comment records that these three are legacy rather than removed,
		// and reports them anyway.
		"componentWillMount": {
			since:       "16.9.0",
			replacement: "UNSAFE_componentWillMount",
			reference: "https://reactjs.org/docs/react-component.html#unsafe_componentwillmount" +
				unsafeLifecycleCodemodNote,
		},
		"componentWillReceiveProps": {
			since:       "16.9.0",
			replacement: "UNSAFE_componentWillReceiveProps",
			reference: "https://reactjs.org/docs/react-component.html#unsafe_componentwillreceiveprops" +
				unsafeLifecycleCodemodNote,
		},
		"componentWillUpdate": {
			since:       "16.9.0",
			replacement: "UNSAFE_componentWillUpdate",
			reference: "https://reactjs.org/docs/react-component.html#unsafe_componentwillupdate" +
				unsafeLifecycleCodemodNote,
		},

		// 18.0.0
		"ReactDOM.render":                   {since: "18.0.0", replacement: "createRoot", reference: createRootReference},
		"ReactDOM.hydrate":                  {since: "18.0.0", replacement: "hydrateRoot", reference: createRootReference},
		"ReactDOM.unmountComponentAtNode":   {since: "18.0.0", replacement: "root.unmount", reference: createRootReference},
		"ReactDOMServer.renderToNodeStream": {since: "18.0.0", replacement: "renderToPipeableStream", reference: "https://reactjs.org/docs/react-dom-server.html#rendertonodestream"},
	}
}

// deprecatedModules maps an import specifier to the object name its exports are deprecated under.
//
// Upstream's `MODULES` maps each module to a LIST of names and then uses only `[0]`, everywhere it
// is read. `react-addons-perf` is the one entry whose list has two, and its second name (`Perf`) is
// never reached by any code path: the import arm takes `[0]`, and `getReactModuleName`'s
// `arguments[0].value` branch also takes `[0]`. Verified by reading both call sites and by driving
// `import { printDOM } from 'react-addons-perf'`, which reports as `ReactPerf.printDOM` and never as
// `Perf.printDOM`. So one name per module here rather than a list, with the discarded name recorded
// in `perfIdentifierNames` below, where it IS reachable.
var deprecatedModules = map[string]string{
	"react":             "React",
	"react-addons-perf": "ReactPerf",
	"react-dom":         "ReactDOM",
	"react-dom/server":  "ReactDOMServer",
}

// perfIdentifierNames is where `react-addons-perf`'s second name becomes reachable.
//
// `getReactModuleName`'s other branch matches `node.init.name` against every name in the list, so
// `var {printDOM} = Perf` resolves to the module and reports as `Perf.printDOM`. That is the only
// route by which the second entry is read, which is why it is here and not in the map above.
// Measured: `var {printDOM} = Perf;` reports `Perf.printDOM is deprecated since React 15.0.0`.
var moduleNamesByIdentifier = map[string]string{
	"React":          "React",
	"ReactPerf":      "ReactPerf",
	"Perf":           "Perf",
	"ReactDOM":       "ReactDOM",
	"ReactDOMServer": "ReactDOMServer",
}

// reactPragmaFor reads the `@jsx` annotation out of the file's comments, or answers React.
//
// Upstream's regular expression is `/@jsx\s+([^\s]+)/` applied to the comment's VALUE, which is its
// text with the delimiters stripped, and it takes the first comment in the file that matches at all
// rather than the first well-formed one. Both details are reproduced: the scan is over every
// comment in source order, and the captured token is cut at its first dot.
//
// `comments.ForFile` is the shelf's cached per-file scan. Reaching for the scanner directly would
// duplicate a derivation three other rules already share, which is the cost this shelf exists to
// avoid.
func reactPragmaFor(ctx rule.Context) string {
	for _, comment := range comments.ForFile(ctx) {
		candidate, found := jsxAnnotationIn(commentValueOf(comment))
		if !found {
			continue
		}
		// Upstream falls back to React on a name that is not an identifier, after warning. The
		// warning has no counterpart here; the fallback does.
		if !isJavaScriptIdentifier(candidate) {
			return defaultPragma
		}
		return candidate
	}
	return defaultPragma
}

// commentValueOf strips a comment's delimiters, which is what ESLint hands a rule as `.value`.
//
// The shelf keeps the full source text including delimiters, deliberately, because what counts as
// content differs by rule. Upstream matches against the delimiter-free value, and the difference is
// reachable: a comment written `/*@jsx*/` has `@jsx` adjacent to the delimiter, and leaving the
// delimiters in would let `*/` be captured as part of a name.
func commentValueOf(comment comments.Comment) string {
	text := comment.Text
	if comment.IsBlock {
		text = strings.TrimPrefix(text, "/*")
		return strings.TrimSuffix(text, "*/")
	}
	// The `//` strip is EQUIVALENT under the current `jsxAnnotationIn` and is kept anyway.
	//
	// Measured: a mutant replacing this line with a bare `return text` survives the whole fixture
	// set, and its inverse (returning the empty string) is caught by five lines, so the arm is
	// reached and the fixtures can see it. The reason no input distinguishes the two is mechanical:
	// `jsxAnnotationIn` searches for `@jsx` at any offset and reads the name from what FOLLOWS it,
	// so a `//` sitting before the marker can never land inside a captured name. The `*/` on the
	// block arm can, which is why that one has a fixture and this one has this paragraph.
	//
	// Kept because it is what ESLint hands a rule, and because the equivalence is a property of the
	// current scanner rather than of the rule: a future `jsxAnnotationIn` anchored at the start of
	// the text would make this line load-bearing again with nothing to announce the change.
	return strings.TrimPrefix(text, "//")
}

// jsxAnnotationIn finds `@jsx` followed by whitespace and a run of non-whitespace.
//
// Hand-rolled rather than compiled from a pattern because the shape is fixed and this runs once per
// file. The dot cut is upstream's `matches[1].split('.')[0]`, so `@jsx Foo.bar` yields `Foo`.
func jsxAnnotationIn(text string) (string, bool) {
	const marker = "@jsx"
	for offset := 0; ; {
		index := strings.Index(text[offset:], marker)
		if index < 0 {
			return "", false
		}
		rest := text[offset+index+len(marker):]
		trimmed := strings.TrimLeft(rest, " \t\r\n\f\v")
		// Upstream's `\s+` requires at least one space, so `@jsxFoo` does not match. Comparing
		// lengths is how that requirement survives the trim.
		if len(trimmed) < len(rest) && trimmed != "" {
			name := trimmed
			if end := strings.IndexAny(name, " \t\r\n\f\v"); end >= 0 {
				name = name[:end]
			}
			if dot := strings.IndexByte(name, '.'); dot >= 0 {
				name = name[:dot]
			}
			if name != "" {
				return name, true
			}
		}
		offset += index + len(marker)
	}
}

// isJavaScriptIdentifier reproduces upstream's `/^[_$a-zA-Z][_$a-zA-Z0-9]*$/`.
//
// Upstream's own comment beside that pattern says it "does not check for reserved keywords or
// unicode characters", so it is deliberately ASCII-only and a port widening it to unicode would
// accept pragma names upstream rejects.
func isJavaScriptIdentifier(name string) bool {
	if name == "" {
		return false
	}
	for index := 0; index < len(name); index++ {
		character := name[index]
		switch {
		case character == '_' || character == '$':
		case character >= 'a' && character <= 'z':
		case character >= 'A' && character <= 'Z':
		case index > 0 && character >= '0' && character <= '9':
		default:
			return false
		}
	}
	return true
}

// sourceSliceOf returns a node's own source text, which is what upstream compares.
//
// `TokenRange` is what `ctx.ReportNode` uses, so the slice starts where the finding would be
// anchored rather than at the node's leading trivia. That matters for a member expression whose
// receiver carries a comment: the trivia belongs to the receiver rather than to the member, and
// including it would make the key unmatchable for a reason upstream does not have.
func sourceSliceOf(ctx rule.Context, node *ast.Node) string {
	textRange := rule.TokenRange(ctx.SourceFile, node)
	text := ctx.SourceFile.Text()
	start, end := textRange.Pos(), textRange.End()
	if start < 0 || end > len(text) || start > end {
		return ""
	}
	return text[start:end]
}

// deprecatedModuleAlias answers the object name an import's exports are deprecated under.
func deprecatedModuleAlias(node *ast.Node) (string, bool) {
	declaration := node.AsImportDeclaration()
	if declaration == nil || declaration.ModuleSpecifier == nil {
		return "", false
	}
	if !ast.IsStringLiteralLike(declaration.ModuleSpecifier) {
		return "", false
	}
	moduleName, known := deprecatedModules[declaration.ModuleSpecifier.Text()]
	return moduleName, known
}

// forEachNamedImportSpecifier visits each `{ name }` or `{ name as alias }` binding.
//
// Upstream filters on `'imported' in s && s.imported`, which is true only for a named import: a
// default import and a namespace import both lack it. Both are therefore silent, measured, and the
// kind check below is how that filter is spelled here.
//
// The name visited is the IMPORTED one rather than the local alias, which is upstream's
// `specifier.imported.name`. `import { createClass as cc }` reports as `React.createClass`, and the
// span is the whole specifier including the alias. Measured; the corpus writes no alias.
func forEachNamedImportSpecifier(node *ast.Node, visit func(specifier *ast.Node, importedName string)) {
	// Named is empty for a default-only or namespace-only import, which is how upstream's
	// `'imported' in s` filter reads here: both shapes are silent.
	for _, element := range imports.BindingsOf(node).Named {
		if element.Kind != ast.KindImportSpecifier {
			continue
		}
		specifier := element.AsImportSpecifier()
		// `PropertyName` is set only when an alias is written; without one the local name IS the
		// imported name. Reading only `Name()` would report the alias, which upstream does not do.
		nameNode := specifier.PropertyName
		if nameNode == nil {
			nameNode = specifier.Name()
		}
		if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
			continue
		}
		visit(element, nameNode.Text())
	}
}

// destructuredReactModuleName answers the object name a destructuring binding's keys belong to.
//
// Upstream's `VariableDeclarator` arm accepts two initializers, and its guard is a pair of
// conjunctions that reduce to: the binding must be an object pattern, AND either the initializer is
// a bare identifier naming a known module, or it is a `require(...)` of a known module specifier.
//
// The `pragma` parameter reproduces the fallback in upstream's report call,
// `reactModuleName || pragma`. It is reachable only when `getReactModuleName` answered false while
// the require branch answered true, which cannot happen for any module in the table, since every
// specifier that satisfies the require branch also has a name. Kept because dropping it would be a
// silent narrowing of a line upstream wrote deliberately, and it costs one parameter.
func destructuredReactModuleName(node *ast.Node, pragma string) (string, bool) {
	declaration := node.AsVariableDeclaration()
	if declaration == nil || declaration.Initializer == nil {
		return "", false
	}
	// The object pattern is what makes this arm apply at all. `var [createClass] = require('react')`
	// is an array pattern and is silent upstream, measured.
	if declaration.Name() == nil || declaration.Name().Kind != ast.KindObjectBindingPattern {
		return "", false
	}

	initializer := declaration.Initializer
	if initializer.Kind == ast.KindIdentifier {
		moduleName, known := moduleNamesByIdentifier[initializer.Text()]
		return moduleName, known
	}

	if initializer.Kind != ast.KindCallExpression {
		return "", false
	}
	call := initializer.AsCallExpression()
	callee := call.Expression
	if callee == nil || callee.Kind != ast.KindIdentifier || callee.Text() != "require" {
		return "", false
	}
	if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
		return "", false
	}
	specifier := call.Arguments.Nodes[0]
	if !ast.IsStringLiteralLike(specifier) {
		return "", false
	}
	moduleName, known := deprecatedModules[specifier.Text()]
	if !known {
		return "", false
	}
	if moduleName == "" {
		return pragma, true
	}
	return moduleName, true
}

// forEachDestructuredProperty visits each `{ key }` or `{ key: local }` element of an object pattern.
//
// Upstream filters `p.type !== 'RestElement' && p.key`, so a rest element is skipped and so is
// anything with no key. The name read is `property.key.name`, an IDENTIFIER key, so a computed key
// is silent: `var {['createClass']: c} = require('react')` reports nothing upstream, measured. The
// span is the whole property, which is why the element is what gets visited rather than its key.
func forEachDestructuredProperty(node *ast.Node, visit func(property *ast.Node, key string)) {
	declaration := node.AsVariableDeclaration()
	if declaration == nil || declaration.Name() == nil {
		return
	}
	pattern := declaration.Name().AsBindingPattern()
	if pattern == nil || pattern.Elements == nil {
		return
	}
	for _, element := range pattern.Elements.Nodes {
		if element.Kind != ast.KindBindingElement {
			continue
		}
		binding := element.AsBindingElement()
		// A rest element carries a dots token and no property name, which is upstream's
		// `RestElement` filter.
		if binding.DotDotDotToken != nil {
			continue
		}
		// `PropertyName` is set only when the binding renames; without it the binding's own name
		// is the key. A computed key is a `ComputedPropertyName` rather than an identifier and is
		// declined here, matching upstream's `property.key.name` being undefined for it.
		keyNode := binding.PropertyName
		if keyNode == nil {
			keyNode = binding.Name()
		}
		if keyNode == nil || keyNode.Kind != ast.KindIdentifier {
			continue
		}
		visit(element, keyNode.Text())
	}
}

// reportLifeCycleMethods builds the listener shared by the three component-bearing kinds.
//
// One closure rather than three, because upstream registers the same function under
// `ClassDeclaration`, `ClassExpression` and `ObjectExpression` and the arms must not drift.
func reportLifeCycleMethods(report func(node *ast.Node, name string)) func(node *ast.Node) {
	return func(node *ast.Node) {
		if !isDeprecatedEs5Component(node) && !isDeprecatedEs6Component(node) {
			return
		}
		forEachComponentMemberName(node, func(nameNode *ast.Node, name string) {
			report(nameNode, name)
		})
	}
}

// isDeprecatedEs5Component reports whether an object literal is the argument to createReactClass.
//
// Upstream's `isES5Component` looks at `node.parent.callee`, so the object literal is a component
// only when it sits directly in a call. The factory name is `getCreateClassFromContext`'s default,
// `createReactClass`, and the namespaced form requires the pragma as its object.
//
// The shelf's `react.IsEs5ComponentCall` is NOT used here and the difference is measured: it
// accepts `createClass` as well as `createReactClass`, and upstream accepts only the latter for
// this question. `var F = createClass({componentWillMount(){}})` is silent on the installed build
// and would report through the shelf helper. Its own doc comment records that widening as
// deliberate for the rules that motivated it, which is exactly the case the brief warns about:
// a helper correct for its own caller and wrong for a neighbour that looks the same.
func isDeprecatedEs5Component(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindObjectLiteralExpression {
		return false
	}
	parent := node.Parent
	if parent == nil || parent.Kind != ast.KindCallExpression {
		return false
	}
	// Upstream reads `node.parent.callee` with no paren skipping, but espree produces no
	// parenthesized node at all, so `(createReactClass)({...})` reaches its check as a bare
	// identifier. Our parser keeps the node, so skipping is what reproduces upstream's answer
	// rather than widening past it. Measured: that spelling reports on the installed build.
	callee := ast.SkipParentheses(parent.AsCallExpression().Expression)
	if callee == nil {
		return false
	}
	switch callee.Kind {
	case ast.KindIdentifier:
		return callee.Text() == createReactClassFactoryName
	case ast.KindPropertyAccessExpression:
		access := callee.AsPropertyAccessExpression()
		object := ast.SkipParentheses(access.Expression)
		if object == nil || object.Kind != ast.KindIdentifier || object.Text() != "React" {
			return false
		}
		name := access.Name()
		return name != nil && name.Kind == ast.KindIdentifier &&
			name.Text() == createReactClassFactoryName
	}
	return false
}

// createReactClassFactoryName is `getCreateClassFromContext`'s default and the only reachable value.
//
// `settings.react.createClass` would override it and cohere reads no React settings, so this is
// the answer for every file. Named rather than inlined because it appears in the doc comment above
// as a measured divergence from the shelf helper, and a reader checking that claim should find one
// spelling.
const createReactClassFactoryName = "createReactClass"

// isDeprecatedEs6Component reports whether a class extends a React component base.
//
// This reproduces upstream's `isES6Component` minus its JSDoc branch, which is recorded as
// unmeasured-reachable in the rule's doc comment.
//
// **`react.IsEs6ComponentClass` is not used, and the reason is a paren difference in the opposite
// direction from the one its doc comment describes.** That helper skips parentheses on the
// RECEIVER (`(React).Component`) and not on the heritage expression as a whole, so
// `class F extends (React.Component)` answers false there. Upstream reports it, because espree
// produces no parenthesized node and `node.superClass` is the bare member expression however the
// parens are written. All four spellings were measured on the installed build and all four report:
//
//	class F extends (React.Component) { componentWillMount() {} }    reports
//	class F extends (React).Component { componentWillMount() {} }    reports
//	class F extends ((React.Component)) { componentWillMount() {} }  reports
//	class F extends (Component) { componentWillMount() {} }          reports
//
// Skipping parentheses at both positions is what reproduces that, and it is fidelity rather than a
// widening: it makes our parser answer the question espree's parser was never asked.
func isDeprecatedEs6Component(node *ast.Node) bool {
	if node == nil {
		return false
	}
	var heritageClauses *ast.NodeList
	switch node.Kind {
	case ast.KindClassDeclaration:
		heritageClauses = node.AsClassDeclaration().HeritageClauses
	case ast.KindClassExpression:
		heritageClauses = node.AsClassExpression().HeritageClauses
	default:
		return false
	}
	if heritageClauses == nil {
		return false
	}

	for _, clause := range heritageClauses.Nodes {
		if clause.Kind != ast.KindHeritageClause {
			continue
		}
		// An `implements` clause is a `KindTypeReference` and an `extends` clause is a
		// `KindExpressionWithTypeArguments`, so the kind check below already declines implements
		// without a token comparison. Upstream reads `node.superClass`, which only exists for
		// extends.
		types := clause.AsHeritageClause().Types
		if types == nil {
			continue
		}
		for _, typeNode := range types.Nodes {
			if typeNode.Kind != ast.KindExpressionWithTypeArguments {
				continue
			}
			if isDeprecatedComponentBase(typeNode.AsExpressionWithTypeArguments().Expression) {
				return true
			}
		}
	}
	return false
}

// isDeprecatedComponentBase matches upstream's `/^(Pure)?Component$/` against a base expression.
//
// Parentheses are skipped at both positions, for the reason recorded on the caller.
func isDeprecatedComponentBase(expression *ast.Node) bool {
	expression = ast.SkipParentheses(expression)
	if expression == nil {
		return false
	}
	switch expression.Kind {
	case ast.KindIdentifier:
		return isDeprecatedComponentBaseName(expression.Text())
	case ast.KindPropertyAccessExpression:
		access := expression.AsPropertyAccessExpression()
		object := ast.SkipParentheses(access.Expression)
		if object == nil || object.Kind != ast.KindIdentifier || object.Text() != "React" {
			return false
		}
		name := access.Name()
		return name != nil && name.Kind == ast.KindIdentifier &&
			isDeprecatedComponentBaseName(name.Text())
	}
	return false
}

// isDeprecatedComponentBaseName accepts the two base classes upstream's pattern matches.
func isDeprecatedComponentBaseName(name string) bool {
	return name == "Component" || name == "PureComponent"
}

// forEachComponentMemberName visits every member of a component whose name is a plain identifier.
//
// Upstream's `getComponentProperties` returns `node.body.body` for a class and `node.properties`
// for an object, and then `getPropertyName` reads `nameNode.name`. That property exists on an
// identifier and on nothing else, which is why a computed key and a string-literal key are both
// silent upstream. Both were measured and neither is in the corpus.
//
// Nothing here filters on what the member IS. A getter, a static method, and a class field holding
// a number all report on the installed build, so this arm tests the name and only the name.
func forEachComponentMemberName(node *ast.Node, visit func(nameNode *ast.Node, name string)) {
	var members *ast.NodeList
	switch node.Kind {
	case ast.KindClassDeclaration:
		members = node.AsClassDeclaration().Members
	case ast.KindClassExpression:
		members = node.AsClassExpression().Members
	case ast.KindObjectLiteralExpression:
		members = node.AsObjectLiteralExpression().Properties
	}
	if members == nil {
		return
	}
	for _, member := range members.Nodes {
		nameNode := member.Name()
		if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
			continue
		}
		visit(nameNode, nameNode.Text())
	}
}
