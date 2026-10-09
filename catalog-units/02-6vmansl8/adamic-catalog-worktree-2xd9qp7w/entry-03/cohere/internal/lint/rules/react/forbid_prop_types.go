package react

import (
	"encoding/json"
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/imports"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// ForbidPropTypesOptions configures which prop types are refused and where they are looked for.
type ForbidPropTypesOptions struct {
	// Forbid is the list of prop type names to refuse. Absent means upstream's default of
	// `any`, `array` and `object`; an explicitly empty list forbids nothing, which is a
	// different thing and is why the wire field is a pointer.
	Forbid *[]string `json:"forbid"`

	// CheckContextTypes extends the rule to `contextTypes` declarations.
	CheckContextTypes bool `json:"checkContextTypes"`

	// CheckChildContextTypes extends the rule to `childContextTypes` declarations.
	CheckChildContextTypes bool `json:"checkChildContextTypes"`
}

// forbidPropTypesDefaultForbid is upstream's `DEFAULTS`.
var forbidPropTypesDefaultForbid = []string{"any", "array", "object"}

// DefaultForbidPropTypesOptions is the unconfigured answer.
func DefaultForbidPropTypesOptions() ForbidPropTypesOptions {
	return ForbidPropTypesOptions{}
}

// DecodeForbidPropTypesOptions reads this rule's configuration from the config layer.
//
// Hand-rolled rather than `rule.DecodeOptionsInto` for one key. `forbid` has to be distinguishable
// as ABSENT, because absent means upstream's three defaults while an explicitly empty list forbids
// nothing. The generic helper would decode both to a nil slice and silently turn `forbid: []` into
// the default list, which is the opposite of what it asks for. Measured on the installed build:
// `{forbid: []}` over `PropTypes.any` is clean.
func DecodeForbidPropTypesOptions(raw []byte) (any, error) {
	options := DefaultForbidPropTypesOptions()
	if len(raw) == 0 {
		return options, nil
	}
	// Lenient on purpose: eslint-plugin-react 7.37.5's schema sets `additionalProperties: true` on
	// this object, so ESLint loads an extra key and so must this.
	if err := json.Unmarshal(raw, &options); err != nil {
		return options, err
	}
	return options, nil
}

// forbiddenTypes returns the effective forbid list.
func (o ForbidPropTypesOptions) forbiddenTypes() []string {
	if o.Forbid == nil {
		return forbidPropTypesDefaultForbid
	}
	return *o.Forbid
}

// forbidPropTypesMessage builds the finding, which names the offending type.
func forbidPropTypesMessage(target string) rule.Message {
	return rule.Message{
		Id: "forbiddenPropType",
		Description: fmt.Sprintf(
			"Prop type %q is forbidden here. A prop typed this loosely documents nothing and "+
				"checks nothing: every reader still has to find a call site to learn what shape "+
				"is really expected, and a wrong shape reaches the component unreported. Declare "+
				"the shape it actually takes instead.",
			target,
		),
	}
}

// ForbidPropTypes flags prop type declarations naming a type the configuration refuses.
//
//	valid:   Component.propTypes = { a: PropTypes.string }
//	valid:   Component.propTypes = { a: PropTypes.shape({ b: PropTypes.string }) }
//	invalid: Component.propTypes = { a: PropTypes.any }
//	invalid: Component.propTypes = { a: PropTypes.arrayOf(PropTypes.object) }
//	invalid: class C { static propTypes = { a: PropTypes.array } }
//
// Ported from `react/forbid-prop-types` in `eslint-plugin-react`, read from the clone at
// `lib/rules/forbid-prop-types.js`. Three options, one message, no fixer. Everything below was
// measured by driving the installed build (7.37.5) rather than reasoned from the source.
//
// # There is no file-suffix gate
//
// Measured under `.tsx`, `.jsx` and `.js`: all three report.
//
// # It does not ask whether anything is a component
//
// This reads as a React rule and is a purely syntactic one. A `static propTypes` on a plain class
// with no heritage reports, and so does `o.propTypes = {...}` on a bare object literal. Measured
// both ways. That is upstream being loose rather than upstream being right, and a port that added a
// component gate would go silent on inputs upstream reports while every imported fixture stayed
// green, because the corpus writes real components throughout.
//
// # What is looked for, and where
//
// Six shapes carry a prop types object, and all six are reachable:
//
//	class C { static propTypes = {...} }              a class property
//	C.propTypes = {...}                               an assignment to a member
//	createReactClass({ propTypes: {...} })            a property of an object literal
//	class C { static get propTypes() { return {...} } } a getter's return value
//	C.propTypes = someIdentifier                      an identifier resolved to its object
//	{ a: PropTypes.shape({...}) }                     nested inside a shape call
//
// `contextTypes` and `childContextTypes` are checked only when the matching option is on, which is
// upstream's default of off for both. Measured.
//
// # What counts as a forbidden type
//
// The value is unwrapped in three ways before its name is read, and each was measured:
//
//	PropTypes.any                reports `any`
//	PropTypes.any.isRequired     reports `any`; the `isRequired` member is peeled off first
//	PropTypes.arrayOf(X)         reports on the ARGUMENT's name rather than on `arrayOf`, so
//	                             `arrayOf(PropTypes.object)` reports `object`
//	a: any                       reports `any`; a bare identifier is read as the type name
//	PropTypes.oneOfType([...])   SILENT, because the argument is an array literal rather than a
//	                             member expression or an identifier
//
// The finding always points at the PROPERTY, not at the value, so `a: PropTypes.any` underlines the
// whole `a: PropTypes.any`. For a nested shape it underlines the inner property. Measured by
// slicing the reported range.
//
// # The foreign-package guard, and it is module-level state
//
// Upstream tracks imports across the file: importing anything named `PropTypes` from a module that
// is neither `react` nor `prop-types` sets a flag that makes every subsequent check demand the
// local name match. Measured: `import { PropTypes } from 'other-lib'` makes an otherwise reporting
// file clean, while the same code importing from `prop-types` reports, and an aliased
// `import PT from 'prop-types'` reports on `PT.any`.
//
// The state is per file and the import declaration may sit after the usage in source order, so the
// imports are collected in one pass before anything is judged. Upstream gets the same effect from
// ESLint visiting the whole program, and reproducing it needs an explicit pre-pass here.
//
// # One arm is not reproduced, and the reason is a missing substrate rather than a decision
//
// Upstream also unwraps a "prop wrapper" call, so `C.propTypes = forbidExtraProps({...})` is looked
// inside. Which functions count comes from `settings.propWrapperFunctions`, a SHARED setting read
// by four rules in this plugin. Our `rule.Context` exposes `SourceFile`, `Program`, `TypeChecker`,
// `Report` and `FileCache` and no path to configuration settings, and the setting is not in this
// rule's own schema so it cannot be carried as a rule option without inventing a surface the other
// three rules would not share.
//
// Measured cost: 14 of upstream's 111 corpus cases carry that setting, all of them reporting, and
// they assert 26 of the 82 findings. Every one is the same shape, a props object wrapped in a call.
// The other 97 cases, including all 48 passing ones, are reproduced here.
//
// This is a stated divergence rather than a silent one, and it costs findings rather than inventing
// them: a wrapped props object is simply not looked inside. If a settings path is added to
// `rule.Context`, the arm is `checkNode`'s call-expression branch and the 14 cases are in the
// corpus waiting.
//
// # The same missing setting makes a SIBLING RULE unportable outright, which is worth knowing here
//
// `react/prefer-exact-props` is on the audit as an unported line item rated Yes with no violations,
// and it is not work outstanding. The same `settings.propWrapperFunctions` this arm needs is the
// only input its reporting branch has: `getExactPropWrapperFunctions(context)` at its line 40 is
// where `exactWrappers` comes from, and its `propTypes` message requires that set to be non-empty.
//
// Where this rule loses 14 of 111 cases to the gap, that one loses everything. Measured against the
// installed 7.37.5 build on identical input:
//
//	no settings                    0 findings
//	settings.propWrapperFunctions  1 finding, messageId propTypes
//
// Its corpus is 31 cases, 17 of which carry a settings block. Replaying all 12 of its REPORTING
// cases with no settings: all twelve go silent, none survive. Its other message id is `flow` and
// needs Flow annotations a TypeScript tree does not have. Ahra configures `propWrapperFunctions`
// nowhere.
//
// So a port of it would register, pass a clean fixture set, and be structurally incapable of firing
// here, which is the shape the audit's zero cannot distinguish from a clean tree. It also ships no
// fixer, despite reading as though it does: its `meta` block is docs, messages and schema, and the
// string `fixer` does not appear in the file.
//
// Recorded here rather than in a rule file of its own, because this is the file that already
// explains the missing surface and the next reader chasing that audit line will arrive at the
// setting before they arrive at the rule.
var ForbidPropTypes = rule.Rule{
	Name: "react/forbid-prop-types",
	// Declared for one arm only: `C.propTypes = someIdentifier` resolves that identifier to the
	// object literal it was declared with, which upstream answers from its own variable index.
	// Every other arm is syntactic, so a nil checker costs exactly that arm rather than the rule,
	// and `resolveToObjectLiteral` guards for it. The fixture below pins that the typed harness is
	// required, so a later revert fails loudly instead of silently losing one shape.
	NeedsTypeChecker: true,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, ok := rule.OptionsAs[ForbidPropTypesOptions](options)
		if !ok {
			// A rule configured as bare `"error"` arrives with nil options, which type-asserts to
			// the zero value. That happens to be the right default here, since a nil Forbid means
			// upstream's three defaults, but it is written out so the reasoning is at the line.
			settings = DefaultForbidPropTypesOptions()
		}

		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				walker := &forbidPropTypesWalker{ctx: ctx, settings: settings}
				walker.collectImports(node)
				walker.walk(node)
			},
		}
	},
}

// forbidPropTypesWalker carries the per-file import state the rule's judgment depends on.
type forbidPropTypesWalker struct {
	ctx      rule.Context
	settings ForbidPropTypesOptions

	// propTypesPackageName is the local name bound to the prop-types package, when one is imported.
	propTypesPackageName string

	// reactPackageName is the local name bound to react, when it is imported.
	reactPackageName string

	// isForeignPropTypesPackage records that something named `PropTypes` was imported from a module
	// that is neither `react` nor `prop-types`. Upstream uses this to stop treating every
	// `PropTypes.x` in the file as the real one.
	isForeignPropTypesPackage bool
}

// collectImports runs before any judgment, because an import may follow the usage it governs.
//
// ESLint visits the whole program and the import listener happens to fire first for a top-level
// import, but nothing guarantees that ordering for the shapes this rule meets, and our walk is a
// single pre-order pass. Collecting first makes the dependence explicit rather than incidental.
func (w *forbidPropTypesWalker) collectImports(sourceFile *ast.Node) {
	sourceFile.ForEachChild(func(child *ast.Node) bool {
		if child.Kind != ast.KindImportDeclaration {
			return false
		}
		declaration := child.AsImportDeclaration()
		moduleSpecifier := declaration.ModuleSpecifier
		if moduleSpecifier == nil || moduleSpecifier.Kind != ast.KindStringLiteral {
			return false
		}

		switch moduleSpecifier.Text() {
		case "prop-types":
			if name := forbidPropTypesFirstLocalName(child); name != "" {
				w.propTypesPackageName = name
			}

		case "react":
			if name := forbidPropTypesFirstLocalName(child); name != "" {
				w.reactPackageName = name
			}
			// `import { PropTypes } from 'react'` binds the prop types package under whatever
			// local name that specifier uses.
			if local := forbidPropTypesLocalNameForImported(child, "PropTypes"); local != "" {
				w.propTypesPackageName = local
			}

		default:
			// Anything named `PropTypes` coming from a third module means the bare name in this
			// file is somebody else's. Measured: it makes an otherwise reporting file clean.
			if forbidPropTypesBindsLocalName(child, "PropTypes") {
				w.isForeignPropTypesPackage = true
			}
		}
		return false
	})
}

// forbidPropTypesFirstLocalName returns the first local binding an import declaration introduces.
//
// Upstream reads `node.specifiers[0].local.name`, which is the default import when there is one and
// otherwise the first named specifier, so the search order below is upstream's rather than a
// preference. A namespace import sits between the two in the same slot ESTree would list first
// after a default, which is why it is consulted before the named list.
//
// `imports.BindingsOf` answers which of the three shapes the statement carries, with the default,
// the namespace and the named specifiers already separated, so the kind switch this used to hold is
// the utility's rather than each rule's.
func forbidPropTypesFirstLocalName(node *ast.Node) string {
	bindings := imports.BindingsOf(node)

	if bindings.Default != nil && bindings.Default.Kind == ast.KindIdentifier {
		return bindings.Default.Text()
	}
	if bindings.Namespace != nil {
		if name := bindings.Namespace.Name(); name != nil {
			return name.Text()
		}
	}
	for _, specifier := range bindings.Named {
		if name := specifier.Name(); name != nil {
			return name.Text()
		}
	}
	return ""
}

// forbidPropTypesLocalNameForImported finds the local name a named import binds for one imported
// name, or "".
//
// `import { PropTypes as PT }` binds `PT` locally for the imported `PropTypes`, and upstream reads
// exactly that pair. Only the named list can carry a rename, which is why this reads that field
// alone rather than the whole Bindings.
func forbidPropTypesLocalNameForImported(node *ast.Node, imported string) string {
	for _, element := range imports.BindingsOf(node).Named {
		specifier := element.AsImportSpecifier()
		local := specifier.Name()
		if local == nil {
			continue
		}
		// `PropertyName` is the imported name when the specifier renames, and nil when it does not,
		// in which case the local name is also the imported one.
		importedName := specifier.PropertyName
		if importedName == nil {
			if local.Text() == imported {
				return local.Text()
			}
			continue
		}
		if importedName.Kind == ast.KindIdentifier && importedName.Text() == imported {
			return local.Text()
		}
	}
	return ""
}

// forbidPropTypesBindsLocalName reports whether an import declaration binds a given LOCAL name.
//
// Upstream tests `node.specifiers.some((x) => x.local.name === 'PropTypes')`, which is the local
// side rather than the imported one, so `import { Foo as PropTypes } from 'other'` trips the guard.
// All three binding shapes are searched because `specifiers` holds all three together upstream.
func forbidPropTypesBindsLocalName(node *ast.Node, local string) bool {
	bindings := imports.BindingsOf(node)

	if bindings.Default != nil && bindings.Default.Kind == ast.KindIdentifier &&
		bindings.Default.Text() == local {
		return true
	}
	if bindings.Namespace != nil {
		if name := bindings.Namespace.Name(); name != nil && name.Text() == local {
			return true
		}
	}
	for _, specifier := range bindings.Named {
		if name := specifier.Name(); name != nil && name.Text() == local {
			return true
		}
	}
	return false
}

// walk visits every node, dispatching the shapes that can carry a prop types object.
func (w *forbidPropTypesWalker) walk(node *ast.Node) {
	if node == nil {
		return
	}

	switch node.Kind {
	case ast.KindPropertyDeclaration:
		w.checkClassProperty(node)

	case ast.KindBinaryExpression:
		w.checkMemberAssignment(node)

	case ast.KindObjectLiteralExpression:
		w.checkObjectLiteral(node)

	case ast.KindGetAccessor:
		w.checkGetter(node)

	case ast.KindCallExpression:
		w.checkShapeCall(node)
	}

	node.ForEachChild(func(child *ast.Node) bool {
		w.walk(child)
		return false
	})
}

// declarationNameKind says which of the three declaration names a node carries, or "".
func (w *forbidPropTypesWalker) declarationNameKind(name string) bool {
	switch name {
	case "propTypes":
		return true
	case "contextTypes":
		return w.settings.CheckContextTypes
	case "childContextTypes":
		return w.settings.CheckChildContextTypes
	}
	return false
}

// checkClassProperty handles `class C { static propTypes = {...} }`.
//
// Upstream does not require the property to be static and neither does this: `propsUtil` tests the
// key's name alone.
func (w *forbidPropTypesWalker) checkClassProperty(node *ast.Node) {
	property := node.AsPropertyDeclaration()
	name := property.Name()
	if name == nil || name.Kind != ast.KindIdentifier ||
		!w.declarationNameKind(name.Text()) {
		return
	}
	w.checkNode(property.Initializer)
}

// checkMemberAssignment handles `C.propTypes = {...}`.
func (w *forbidPropTypesWalker) checkMemberAssignment(node *ast.Node) {
	assignment := node.AsBinaryExpression()
	if !ast.IsAssignmentOperator(assignment.OperatorToken.Kind) {
		// EVERY assignment operator qualifies, not only `=`, and this line was wrong before it was
		// measured. Upstream anchors on the member expression and reads `node.parent.right`, which
		// any assignment carries, so `C.propTypes ||= {...}` and `+=` both report. Verified on the
		// installed build across four operators, all reporting.
		//
		// The first draft narrowed this to `KindEqualsToken` on the reasoning that a compound
		// operator has no meaning for a prop types object. That reasoning is about what the code
		// ought to mean rather than what upstream decides, and it is exactly the "documenting a
		// limit you did not measure" failure: it survived every one of the 97 imported cases,
		// because the corpus writes only plain assignment.
		return
	}
	target := assignment.Left
	if target == nil || target.Kind != ast.KindPropertyAccessExpression {
		return
	}
	name := target.AsPropertyAccessExpression().Name()
	if name == nil || name.Kind != ast.KindIdentifier ||
		!w.declarationNameKind(name.Text()) {
		return
	}
	w.checkNode(assignment.Right)
}

// checkObjectLiteral handles `createReactClass({ propTypes: {...} })`.
//
// Upstream visits every object expression and looks for a property with one of the three names, so
// this is not restricted to a factory call and an ordinary object literal carrying a `propTypes`
// key is checked too. Measured.
func (w *forbidPropTypesWalker) checkObjectLiteral(node *ast.Node) {
	properties := node.AsObjectLiteralExpression().Properties
	if properties == nil {
		return
	}
	for _, property := range properties.Nodes {
		if property.Kind != ast.KindPropertyAssignment {
			continue
		}
		assignment := property.AsPropertyAssignment()
		name := assignment.Name()
		if name == nil || name.Kind != ast.KindIdentifier ||
			!w.declarationNameKind(name.Text()) {
			continue
		}
		// Upstream checks only for an object expression here, not the full `checkNode` surface, so
		// a factory whose `propTypes` is an IDENTIFIER is silent while the same identifier through
		// a member assignment reports. That asymmetry is real and measured, and
		// TestForbidPropTypesObjectLiteralArmDoesNotFollowIdentifiers pins it.
		//
		// The kind test on this line is nonetheless EQUIVALENT rather than load-bearing, and a
		// mutation widening it survives correctly. `checkProperties` opens with the same test and
		// returns on anything that is not an object literal, so no input can distinguish the two
		// versions. What produces the asymmetry is this arm calling `checkProperties` directly
		// rather than calling `checkNode`, which is a different line.
		//
		// Kept because it states the arm's own contract at the point of the call rather than
		// leaving a reader to follow the callee, and because widening it would read as an
		// invitation to pass anything. All four callers of `checkProperties` were enumerated when
		// this verdict was taken; a caller that skipped that entry guard would void it.
		value := assignment.Initializer
		if value != nil && value.Kind == ast.KindObjectLiteralExpression {
			w.checkProperties(value)
		}
	}
}

// checkGetter handles `static get propTypes() { return {...} }`.
func (w *forbidPropTypesWalker) checkGetter(node *ast.Node) {
	accessor := node.AsGetAccessorDeclaration()
	name := accessor.Name()
	if name == nil || name.Kind != ast.KindIdentifier ||
		!w.declarationNameKind(name.Text()) {
		return
	}
	body := accessor.Body
	if body == nil {
		return
	}
	// Upstream's `findReturnStatement` takes the LAST return statement in the body's statement
	// list, at the top level only.
	statements := body.AsBlock().Statements
	if statements == nil {
		return
	}
	for index := len(statements.Nodes) - 1; index >= 0; index-- {
		statement := statements.Nodes[index]
		if statement.Kind != ast.KindReturnStatement {
			continue
		}
		w.checkNode(statement.AsReturnStatement().Expression)
		return
	}
}

// checkShapeCall handles a `shape({...})` call anywhere, which upstream visits independently of the
// declaration it sits in.
//
// That independence is real and measured: a `shape` call reached through a nested prop type is
// checked by this arm rather than by the recursion, which is why a shape's own properties report
// even though `checkProperties` does not recurse into them.
func (w *forbidPropTypesWalker) checkShapeCall(node *ast.Node) {
	call := node.AsCallExpression()
	callee := call.Expression
	if callee == nil {
		return
	}

	// The RECEIVER has to look like the prop types package, and this guard is the whole reason
	// `Yup.object().shape({...})` is clean. Upstream declines a member callee whose object fails
	// `isPropTypesPackage` and is not itself a prop types declaration.
	//
	// `isPropTypesPackage` accepts an identifier or a member expression and nothing else, so a CALL
	// receiver such as `Yup.object()` falls through to false whatever is imported. That is what
	// separates a validation schema from a prop type, and it is not the foreign-import flag:
	// measured with and without the `yup` import, both clean, while `X.propTypes = { a:
	// PropTypes.shape({...}) }` reports either way.
	//
	// Three of upstream's passing cases are exactly this shape and all three reported before this
	// guard was written. They are the reason step 3 exists: nothing about the rule's description
	// suggests a validation library is the false positive to worry about.
	if callee.Kind == ast.KindPropertyAccessExpression {
		receiver := callee.AsPropertyAccessExpression().Expression
		if !w.isPropTypesPackage(receiver) && !w.isPropTypesDeclarationAccess(callee) {
			return
		}
	}

	calleeName := ""
	switch callee.Kind {
	case ast.KindIdentifier:
		calleeName = callee.Text()
	case ast.KindPropertyAccessExpression:
		if name := callee.AsPropertyAccessExpression().Name(); name != nil &&
			name.Kind == ast.KindIdentifier {
			calleeName = name.Text()
		}
	}
	if calleeName != "shape" {
		return
	}

	if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
		return
	}
	argument := call.Arguments.Nodes[0]
	if argument != nil && argument.Kind == ast.KindObjectLiteralExpression {
		w.checkProperties(argument)
	}
}

// checkNode resolves the several things a prop types declaration's value can be.
//
// Upstream's `checkNode` has a fourth branch for a prop wrapper call, which is not reproduced. See
// the rule doc: which functions count comes from a shared setting `rule.Context` does not expose.
func (w *forbidPropTypesWalker) checkNode(node *ast.Node) {
	if node == nil {
		return
	}
	switch node.Kind {
	case ast.KindObjectLiteralExpression:
		w.checkProperties(node)

	case ast.KindIdentifier:
		w.checkProperties(w.resolveToObjectLiteral(node))
	}
}

// resolveToObjectLiteral follows an identifier to the object literal it was initialized with.
//
// Upstream asks its own variable index; the checker answers the same question. Only a variable
// declaration with an object literal initializer qualifies, which is upstream's
// `propTypesObject && propTypesObject.properties` test.
func (w *forbidPropTypesWalker) resolveToObjectLiteral(identifier *ast.Node) *ast.Node {
	if w.ctx.TypeChecker == nil {
		return nil
	}
	symbol := w.ctx.TypeChecker.GetSymbolAtLocation(identifier)
	if symbol == nil {
		return nil
	}
	for _, declaration := range symbol.Declarations {
		if declaration == nil || declaration.Kind != ast.KindVariableDeclaration {
			continue
		}
		initializer := declaration.AsVariableDeclaration().Initializer
		if initializer != nil && initializer.Kind == ast.KindObjectLiteralExpression {
			return initializer
		}
	}
	return nil
}

// checkProperties reads each property of a prop types object and reports the forbidden ones.
//
// It deliberately does NOT recurse into a nested object literal. A nested `shape({...})` is reached
// by the shape arm instead, which visits every such call in the file independently.
func (w *forbidPropTypesWalker) checkProperties(object *ast.Node) {
	if object == nil || object.Kind != ast.KindObjectLiteralExpression {
		return
	}
	properties := object.AsObjectLiteralExpression().Properties
	if properties == nil {
		return
	}

	for _, property := range properties.Nodes {
		if property.Kind != ast.KindPropertyAssignment {
			// A spread carries no single type name, which is upstream declining anything that is
			// not a `Property`.
			continue
		}
		w.checkPropertyValue(property, property.AsPropertyAssignment().Initializer)
	}
}

// checkPropertyValue unwraps one property's value and reports if the resulting name is forbidden.
//
// The finding is anchored on the PROPERTY rather than on the value, which is upstream passing the
// declaration node to `report`.
func (w *forbidPropTypesWalker) checkPropertyValue(property *ast.Node, value *ast.Node) {
	if value == nil {
		return
	}

	// `PropTypes.any.isRequired` peels to `PropTypes.any`.
	if value.Kind == ast.KindPropertyAccessExpression {
		access := value.AsPropertyAccessExpression()
		if name := access.Name(); name != nil && name.Kind == ast.KindIdentifier &&
			name.Text() == "isRequired" {
			value = access.Expression
			if value == nil {
				return
			}
		}
	}

	// `PropTypes.arrayOf(PropTypes.object)` reports on the ARGUMENT's name, then continues with the
	// callee so the wrapper's own name is considered too.
	if value.Kind == ast.KindCallExpression {
		call := value.AsCallExpression()
		if !w.isPropTypesPackage(call.Expression) {
			return
		}
		if call.Arguments != nil {
			for _, argument := range call.Arguments.Nodes {
				w.reportIfForbidden(property, forbidPropTypesTypeName(argument))
			}
		}
		value = call.Expression
		if value == nil {
			return
		}
	}

	if !w.isPropTypesPackage(value) {
		return
	}
	w.reportIfForbidden(property, forbidPropTypesTypeName(value))
}

// forbidPropTypesTypeName reads the type name off a value, which upstream does two ways.
//
// A member expression contributes its property name, so `PropTypes.any` gives `any`; a bare
// identifier contributes its own name, so `a: any` gives `any`. Anything else, notably an array
// literal, contributes nothing, which is why `oneOfType([...])` is silent.
func forbidPropTypesTypeName(node *ast.Node) string {
	if node == nil {
		return ""
	}
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		if name := node.AsPropertyAccessExpression().Name(); name != nil &&
			name.Kind == ast.KindIdentifier {
			return name.Text()
		}
	case ast.KindIdentifier:
		return node.Text()
	}
	return ""
}

// isPropTypesPackage reports whether a value is reached through something this file treats as the
// prop types package.
//
// Upstream's test is written as a disjunction whose last term dominates: unless a foreign
// `PropTypes` was imported, EVERYTHING passes. So in an ordinary file with no such import this is
// always true, and the whole predicate exists to narrow the rule once the name has been claimed by
// another module. Measured: with a foreign import present, only the recorded local name qualifies.
func (w *forbidPropTypesWalker) isPropTypesPackage(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindIdentifier:
		return node.Text() == w.propTypesPackageName || !w.isForeignPropTypesPackage

	case ast.KindPropertyAccessExpression:
		receiver := node.AsPropertyAccessExpression().Expression
		if receiver == nil {
			return !w.isForeignPropTypesPackage
		}
		if receiver.Kind == ast.KindIdentifier {
			// The member arm compares against `reactPackageName` ONLY, never against
			// `propTypesPackageName`. That reads as an oversight upstream, since a
			// `import PT from 'prop-types'` alias is exactly what you would expect to qualify
			// here, and a first draft of this port helpfully added it.
			//
			// It is a divergence. Measured on the installed build with both a prop-types alias and
			// a foreign PropTypes import present: `PT.any` is SILENT upstream, in either import
			// order, while the same file without the foreign import reports. Adding the second
			// comparison made this port report where upstream does not, and the whole imported
			// corpus stayed green because no case writes both imports at once.
			//
			// The identifier arm above does consult `propTypesPackageName`, which is why the two
			// arms are not symmetric here.
			return receiver.Text() == w.reactPackageName || !w.isForeignPropTypesPackage
		}
		return !w.isForeignPropTypesPackage
	}
	return false
}

// reportIfForbidden reports when a type name is on the forbid list.
func (w *forbidPropTypesWalker) reportIfForbidden(property *ast.Node, typeName string) {
	if typeName == "" {
		return
	}
	for _, forbidden := range w.settings.forbiddenTypes() {
		if forbidden == typeName {
			w.ctx.ReportNode(property, forbidPropTypesMessage(typeName))
			return
		}
	}
}

// isPropTypesDeclarationAccess reports whether a member access names one of the three declarations.
//
// Upstream's shape guard has a second escape hatch beside the package test: a callee that IS a prop
// types declaration passes even when its receiver does not look like the package. Reproduced for
// fidelity; no corpus case reaches it, and the shape it would need is a `shape` call written
// directly as `something.propTypes(...)`.
func (w *forbidPropTypesWalker) isPropTypesDeclarationAccess(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	name := node.AsPropertyAccessExpression().Name()
	return name != nil && name.Kind == ast.KindIdentifier && w.declarationNameKind(name.Text())
}
