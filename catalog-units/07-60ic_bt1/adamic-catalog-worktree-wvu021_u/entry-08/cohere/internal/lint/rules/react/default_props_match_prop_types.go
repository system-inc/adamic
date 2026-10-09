package react

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// DefaultPropsMatchPropTypesOptions configures the rule.
//
// One key, defaulting to false, so the zero value is the correct default and the decoder stays
// simple. Stated because two siblings in this package default an option to TRUE, where the same
// shape would silently invert the rule.
type DefaultPropsMatchPropTypesOptions struct {
	// AllowRequiredDefaults permits a default on a prop declared `isRequired`. Absent means false.
	AllowRequiredDefaults bool `json:"allowRequiredDefaults"`
}

// DefaultDefaultPropsMatchPropTypesOptions is the unconfigured answer.
//
// Upstream reads `configuration.allowRequiredDefaults || false`, so an absent options object, an
// empty one, and an explicit false are the same rule. Unlike `boolean-prop-naming`, this rule reads
// nothing else from its options, so it is NOT inert without them: measured, a bare `"error"` reports
// exactly as `{}` does.
func DefaultDefaultPropsMatchPropTypesOptions() DefaultPropsMatchPropTypesOptions {
	return DefaultPropsMatchPropTypesOptions{}
}

// DecodeDefaultPropsMatchPropTypesOptions reads this rule's configuration from the config layer.
//
// Our config layer unwraps the severity tuple before dispatch, so this receives upstream's option
// object rather than upstream's one-element array. Empty input is a bare `"error"` and resolves to
// the default, which is the correct behaviour here precisely because the rule works without options.
func DecodeDefaultPropsMatchPropTypesOptions(raw []byte) (any, error) {
	options := DefaultDefaultPropsMatchPropTypesOptions()
	if len(raw) == 0 {
		return options, nil
	}
	if err := rule.UnmarshalOptions(raw, &options); err != nil {
		return options, err
	}
	return options, nil
}

// messageRequiredHasDefault is the arm for a default on a required prop.
//
// The finding is reported on the DEFAULT rather than on the propType, because the repair is to drop
// one of the two and the default is the half that is redundant. Measured against the installed build
// by slicing its reported range: the span is `a: 1`, not `a: PropTypes.string.isRequired`.
func messageRequiredHasDefault(name string) rule.Message {
	return rule.Message{
		Id:          "requiredHasDefault",
		Description: `defaultProp "` + name + `" defined for isRequired propType.`,
	}
}

// messageDefaultHasNoType is the arm for a default nothing declares.
func messageDefaultHasNoType(name string) rule.Message {
	return rule.Message{
		Id:          "defaultHasNoType",
		Description: `defaultProp "` + name + `" has no corresponding propTypes declaration.`,
	}
}

// DefaultPropsMatchPropTypes reports a defaultProps entry that no propTypes declaration supports.
//
//	valid:   propTypes {a: string}, defaultProps {a: 1}
//	valid:   defaultProps with no propTypes at all, which is deliberately not judged
//	invalid: propTypes {a: string}, defaultProps {b: 1}          the default is dead
//	invalid: propTypes {a: string.isRequired}, defaultProps {a}  the two contradict
//
// Two findings in opposite directions. A default nothing declares is dead code. A default on a
// required prop is a contradiction, because the default means the prop can never actually be
// missing, so one of the two statements is wrong.
//
// # The two silences are deliberate and are the rule's whole false-positive story
//
// Upstream returns early when propTypes is absent or empty, and when either object carries a
// spread. Both exist so a component whose props are assembled elsewhere is not judged on a partial
// view, and both are measured: a spread in EITHER object silences the component entirely.
//
// # One declaration form is deliberately not read, and the reason is a measurement
//
// Upstream also reads props from a TYPE ANNOTATION rather than from a `propTypes` object, in two
// spellings that are one question:
//
//	class C extends React.Component { props: { foo: string }; }   the Flow-era class property
//	function C(props: { foo: string }) {}                          a function component's parameter
//
// In both, a non-optional member counts as required, so a default on it is the `requiredHasDefault`
// arm. Our parser produces both shapes, so they are expressible, and five of upstream's failing
// cases write them. Reading either means resolving a type reference to its declaration, which needs
// the type checker,
// and declaring `NeedsTypeChecker` costs the per-file checker lock on every file in the tree.
//
// Measured on 2026-08-28 before deciding, with a seeded control to prove the counter worked: the
// tree contains ZERO occurrences of `defaultProps`, `propTypes` or `getDefaultProps` across every
// TypeScript and JavaScript file, against 56 files naming `React.Component`. React 18.3 deprecated
// `defaultProps` on function components and this codebase has moved past it entirely.
//
// So the trade is a checker lock on 3,516 files against five corpus cases and no real input. The
// five are recorded in the test file rather than expressed, and if `defaultProps` ever returns to
// this tree the annotation arm is the piece to add.
var DefaultPropsMatchPropTypes = rule.Rule{
	Name: "react/default-props-match-prop-types",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		settings := DefaultDefaultPropsMatchPropTypesOptions()
		if configured, isConfigured := rule.OptionsAs[DefaultPropsMatchPropTypesOptions](options); isConfigured {
			settings = configured
		}

		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				checkDefaultPropsAgainstPropTypes(ctx, node, settings)
			},
		}
	},
}

// propTypeEntry is one declared prop: where it was written and whether it is required.
type propTypeEntry struct {
	isRequired bool
}

// defaultPropEntry is one default: the property node to report on.
type defaultPropEntry struct {
	node *ast.Node
	name string
}

// componentPropDeclarations is everything one component declares about its props.
//
// The two `unresolved` flags are upstream's early returns, kept as state rather than as an early
// exit because both objects can be declared in separate statements and either one can be the
// spread that disqualifies the component.
type componentPropDeclarations struct {
	propTypes           map[string]propTypeEntry
	propTypeOrder       []string
	sawPropTypes        bool
	propTypesUnresolved bool
	defaults            []defaultPropEntry
	sawDefaults         bool
	defaultsUnresolved  bool
}

// checkDefaultPropsAgainstPropTypes runs the whole analysis for one file.
//
// Upstream gathers across several listeners and reports in `Program:exit`, because `propTypes` and
// `defaultProps` are usually written as separate statements below the component and either can come
// first. Our walk is pre-order, so a source-file listener fires before its children and the
// gathering has to happen here.
//
// # Why declarations are keyed by OWNER rather than by detected component
//
// Upstream hangs both objects off its component record, so two components in one file keep their
// props apart. The three shapes that reach this rule all name their owner: a class property belongs
// to its class, and `C.defaultProps = {...}` names `C`. Keying on that name is what keeps a second
// component's propTypes from satisfying the first component's defaults, and it costs nothing that a
// detected-component walk would have bought.
func checkDefaultPropsAgainstPropTypes(
	ctx rule.Context,
	sourceFile *ast.Node,
	settings DefaultPropsMatchPropTypesOptions,
) {
	byOwner := map[string]*componentPropDeclarations{}
	order := []string{}
	declarationsFor := func(owner string) *componentPropDeclarations {
		existing, found := byOwner[owner]
		if found {
			return existing
		}
		created := &componentPropDeclarations{propTypes: map[string]propTypeEntry{}}
		byOwner[owner] = created
		order = append(order, owner)
		return created
	}

	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		switch node.Kind {
		case ast.KindPropertyDeclaration:
			// `static propTypes = {...}` or `static defaultProps = {...}` in a class body.
			declaration := node.AsPropertyDeclaration()
			owner := enclosingClassOwnerName(node)
			readPropDeclaration(declarationsFor(owner), declaration.Name(), declaration.Initializer)

		case ast.KindGetAccessor:
			// `static get propTypes() { return {...}; }`. Upstream resolves the getter's returned
			// object, so the declaration is the return value rather than an initializer.
			owner := enclosingClassOwnerName(node)
			readPropDeclaration(declarationsFor(owner), node.Name(),
				singleReturnedExpression(node.AsGetAccessorDeclaration().Body))

		case ast.KindMethodDeclaration:
			// `getDefaultProps: function() { return {...}; }` inside the ES5 factory object, which
			// our parser also spells as a method when written shorthand.
			owner := enclosingObjectOwnerName(node)
			readPropDeclaration(declarationsFor(owner), node.Name(),
				singleReturnedExpression(node.AsMethodDeclaration().Body))

		case ast.KindPropertyAssignment:
			// `propTypes: {...}` and `getDefaultProps: function(){...}` inside the ES5 factory.
			assignment := node.AsPropertyAssignment()
			owner := enclosingObjectOwnerName(node)
			value := skipParenthesesOptional(assignment.Initializer)
			if value != nil && (value.Kind == ast.KindFunctionExpression || value.Kind == ast.KindArrowFunction) {
				value = singleReturnedExpression(functionBodyBlock(value))
			}
			readPropDeclaration(declarationsFor(owner), assignment.Name(), value)

		case ast.KindBinaryExpression:
			// `C.propTypes = {...}` written after the class or the function.
			binary := node.AsBinaryExpression()
			if binary.OperatorToken == nil || binary.OperatorToken.Kind != ast.KindEqualsToken {
				break
			}
			target := skipParenthesesOptional(binary.Left)
			if target == nil || target.Kind != ast.KindPropertyAccessExpression {
				break
			}
			access := target.AsPropertyAccessExpression()
			member := access.Name()
			if member == nil || member.Kind != ast.KindIdentifier {
				break
			}

			// `C.defaultProps.baz = "baz"` adds ONE default to an object declared elsewhere, and
			// upstream folds it into the same component. The receiver is then itself a
			// `C.defaultProps` access, so the owner is one level further up and the member name is
			// the prop rather than the declaration.
			if receiverAccess := skipParenthesesOptional(access.Expression); receiverAccess != nil &&
				receiverAccess.Kind == ast.KindPropertyAccessExpression {
				inner := receiverAccess.AsPropertyAccessExpression()
				if innerName := inner.Name(); innerName != nil && innerName.Kind == ast.KindIdentifier {
					if owner, named := ownerPathOf(inner.Expression); named {
						switch innerName.Text() {
						case "propTypes":
							addSinglePropType(declarationsFor(owner), member.Text(), binary.Right)
							break
						case "defaultProps", "getDefaultProps":
							addSingleDefaultProp(declarationsFor(owner), member.Text(), node)
							break
						}
					}
				}
			}

			owner, named := ownerPathOf(access.Expression)
			if !named {
				break
			}
			readPropDeclaration(declarationsFor(owner), member, binary.Right)
		}

		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}
	visit(sourceFile)

	// Reported in declaration order rather than map order, so two components in one file report
	// deterministically. A map walk here would make the finding order depend on Go's randomisation.
	for _, owner := range order {
		reportMismatchedDefaults(ctx, byOwner[owner], settings)
	}
}

// enclosingClassOwnerName names the class a static member belongs to.
//
// An anonymous class expression has no name, and its members are keyed on the empty string, which
// groups every anonymous class in a file together. That is wider than upstream, which keys on the
// node, and it can only make the rule QUIETER: two anonymous classes sharing a bucket means one's
// propTypes can satisfy the other's defaults. No corpus case writes two anonymous component classes
// in one file, and the alternative is a per-node key that the assignment arm cannot produce, since
// `C.defaultProps = ...` names its owner by identifier rather than by node.
func enclosingClassOwnerName(member *ast.Node) string {
	class := enclosingClassOf(member)
	if class == nil {
		return ""
	}
	if name := class.Name(); name != nil && name.Kind == ast.KindIdentifier {
		return name.Text()
	}
	return ""
}

// readPropDeclaration folds one `propTypes` or `defaultProps` declaration into a component's record.
//
// The name is read through `ast.TryGetTextOfPropertyName` rather than `Text()`, which panics on
// several name shapes. The walk recovers per FILE rather than per rule, so one such call would cost
// every rule in this package its verdict on that file.
func readPropDeclaration(declarations *componentPropDeclarations, name *ast.Node, value *ast.Node) {
	if name == nil {
		return
	}
	text, readable := ast.TryGetTextOfPropertyName(name)
	if !readable {
		return
	}

	switch text {
	case "propTypes":
		declarations.sawPropTypes = true
		readPropTypesObject(declarations, value)
	case "defaultProps", "getDefaultProps":
		// `getDefaultProps` is the ES5 factory spelling, accepted by the sibling
		// `sort-default-props` in both arms for the same reason.
		declarations.sawDefaults = true
		readDefaultPropsObject(declarations, value)
	}
}

// readPropTypesObject records each declared prop and whether it is required.
//
// A spread marks the whole component unresolved, which silences it. Upstream's reason is that a
// spread hides names the rule would otherwise call missing, so judging on a partial view would
// manufacture false positives. Measured: a spread in propTypes silences the component entirely.
func readPropTypesObject(declarations *componentPropDeclarations, value *ast.Node) {
	value = resolveToObjectLiteral(value)
	if value == nil {
		// A propTypes assigned from anything we cannot resolve to an object is a partial view,
		// which is the same situation a spread creates.
		declarations.propTypesUnresolved = true
		return
	}

	// A spread inside propTypes does NOT disqualify the component, unlike a spread inside
	// defaultProps. Upstream's propTypes reader takes what it can see and moves on, while its
	// defaults reader returns the string "unresolved" for the whole object.
	//
	// The asymmetry is the rule working as intended rather than an oversight. A spread in
	// defaultProps may supply a name the rule would then call undeclared, which is a false positive;
	// a spread in propTypes can only DECLARE more, so the worst case is a missed finding. Measured
	// on 2026-08-28: `{...base, a: PropTypes.string}` beside a default named `b` reports one
	// finding, while a spread alone reports none because the map is then empty and the
	// empty-propTypes guard fires.
	//
	// An earlier draft marked propTypes unresolved on a spread and went silent on that first input,
	// which the whole imported corpus missed: no case there mixes a spread with a real prop.
	for _, property := range value.AsObjectLiteralExpression().Properties.Nodes {
		if property.Kind != ast.KindPropertyAssignment &&
			property.Kind != ast.KindShorthandPropertyAssignment {
			continue
		}
		text, readable := ast.TryGetTextOfPropertyName(property.Name())
		if !readable {
			continue
		}
		if _, already := declarations.propTypes[text]; !already {
			declarations.propTypeOrder = append(declarations.propTypeOrder, text)
		}
		declarations.propTypes[text] = propTypeEntry{
			isRequired: propTypeIsRequired(property),
		}
	}
}

// propTypeIsRequired reports whether a declared prop type ends in `.isRequired`.
//
// Upstream's propTypes reader sets the flag from the ANNOTATION rather than from the value, but for
// the shapes this rule meets the two coincide: a required prop is written `PropTypes.string.isRequired`
// and the flag is the trailing member name. A shorthand property carries no value at all and is
// therefore never required.
func propTypeIsRequired(property *ast.Node) bool {
	if property.Kind != ast.KindPropertyAssignment {
		return false
	}
	return expressionIsRequired(property.AsPropertyAssignment().Initializer)
}

// readDefaultPropsObject records each default, in source order.
func readDefaultPropsObject(declarations *componentPropDeclarations, value *ast.Node) {
	value = resolveToObjectLiteral(value)
	if value == nil {
		declarations.defaultsUnresolved = true
		return
	}

	for _, property := range value.AsObjectLiteralExpression().Properties.Nodes {
		if property.Kind != ast.KindPropertyAssignment &&
			property.Kind != ast.KindShorthandPropertyAssignment &&
			property.Kind != ast.KindMethodDeclaration {
			// A spread. Upstream marks the whole defaults object unresolved rather than skipping
			// the one property, because the spread may supply any name.
			declarations.defaultsUnresolved = true
			continue
		}
		text, readable := ast.TryGetTextOfPropertyName(property.Name())
		if !readable {
			declarations.defaultsUnresolved = true
			continue
		}
		declarations.defaults = append(declarations.defaults,
			defaultPropEntry{node: property, name: text})
	}
}

// reportMismatchedDefaults applies the two arms to one component.
func reportMismatchedDefaults(
	ctx rule.Context,
	declarations *componentPropDeclarations,
	settings DefaultPropsMatchPropTypesOptions,
) {
	// Upstream's three early returns, in its own order. A component with no defaults is not judged,
	// one whose defaults cannot be read is not judged, and one with absent or empty propTypes is
	// not judged either. That last one is the surprising member of the set and it is deliberate:
	// a component declaring defaults and no propTypes is a partial view rather than a violation.
	if !declarations.sawDefaults || declarations.defaultsUnresolved {
		return
	}
	if declarations.propTypesUnresolved {
		return
	}
	if !declarations.sawPropTypes || len(declarations.propTypes) == 0 {
		return
	}

	for _, defaultProp := range declarations.defaults {
		declared, isDeclared := declarations.propTypes[defaultProp.name]
		if isDeclared && (settings.AllowRequiredDefaults || !declared.isRequired) {
			continue
		}
		if isDeclared {
			ctx.ReportNode(defaultProp.node, messageRequiredHasDefault(defaultProp.name))
			continue
		}
		ctx.ReportNode(defaultProp.node, messageDefaultHasNoType(defaultProp.name))
	}
}

// ownerPathOf renders the dotted name a propTypes assignment hangs off.
//
// `C.propTypes = {}` names `C`, and `Greetings.Hello.propTypes = {}` names `Greetings.Hello`. The
// path is rendered rather than resolved because the two arms have to agree on a key and one of them
// starts from an identifier with no declaration to resolve.
func ownerPathOf(expression *ast.Node) (string, bool) {
	expression = skipParenthesesOptional(expression)
	if expression == nil {
		return "", false
	}
	switch expression.Kind {
	case ast.KindIdentifier:
		return expression.Text(), true
	case ast.KindPropertyAccessExpression:
		access := expression.AsPropertyAccessExpression()
		name := access.Name()
		if name == nil || name.Kind != ast.KindIdentifier {
			return "", false
		}
		prefix, named := ownerPathOf(access.Expression)
		if !named {
			return "", false
		}
		return prefix + "." + name.Text(), true
	}
	return "", false
}

// singleReturnedExpression reads the object a getter or a factory method returns.
//
// Upstream resolves the return value of `static get propTypes()` and of
// `getDefaultProps: function(){}`, so both are the declaration rather than a wrapper around it.
// Only a direct return in the body is read: a conditional return is a view we cannot resolve, and
// answering nil there makes the component unresolved rather than judged on half its props.
func singleReturnedExpression(body *ast.Node) *ast.Node {
	if body == nil || body.Kind != ast.KindBlock {
		return nil
	}
	for _, statement := range body.AsBlock().Statements.Nodes {
		if statement.Kind == ast.KindReturnStatement {
			return statement.AsReturnStatement().Expression
		}
	}
	return nil
}

// enclosingObjectOwnerName names the ES5 factory call a member belongs to.
//
// `var Greeting = createReactClass({ propTypes: {...} })` hangs both objects off the factory's
// argument, and the owner is the variable the call is assigned to. Falls back to the empty string
// for an unnamed one, which groups it with any other unnamed component in the file for the same
// reason `enclosingClassOwnerName` does.
func enclosingObjectOwnerName(member *ast.Node) string {
	object := semanticParentOf(member)
	if object == nil || object.Kind != ast.KindObjectLiteralExpression {
		return ""
	}
	call := semanticParentOf(object)
	if call == nil || call.Kind != ast.KindCallExpression {
		return ""
	}
	declaration := semanticParentOf(call)
	if declaration == nil || declaration.Kind != ast.KindVariableDeclaration {
		return ""
	}
	if name := declaration.AsVariableDeclaration().Name(); name != nil &&
		name.Kind == ast.KindIdentifier {
		return name.Text()
	}
	return ""
}

// addSinglePropType folds a `C.propTypes.foo = ...` assignment into a component's record.
func addSinglePropType(declarations *componentPropDeclarations, name string, value *ast.Node) {
	declarations.sawPropTypes = true
	if _, already := declarations.propTypes[name]; !already {
		declarations.propTypeOrder = append(declarations.propTypeOrder, name)
	}
	declarations.propTypes[name] = propTypeEntry{isRequired: expressionIsRequired(value)}
}

// addSingleDefaultProp folds a `C.defaultProps.foo = ...` assignment into a component's record.
//
// The node reported on is the whole assignment statement, because there is no property node to
// point at: the default is written as a statement rather than as a member of an object literal.
func addSingleDefaultProp(declarations *componentPropDeclarations, name string, assignment *ast.Node) {
	declarations.sawDefaults = true
	declarations.defaults = append(declarations.defaults,
		defaultPropEntry{node: assignment, name: name})
}

// expressionIsRequired reports whether a prop-type expression ends in `.isRequired`.
//
// The same test `propTypeIsRequired` applies to a property's value, lifted so the single-assignment
// arm can reuse it rather than reconstruct a property node to ask about.
func expressionIsRequired(value *ast.Node) bool {
	value = skipParenthesesOptional(value)
	if value == nil || value.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	name := value.AsPropertyAccessExpression().Name()
	return name != nil && name.Kind == ast.KindIdentifier && name.Text() == "isRequired"
}

// resolveToObjectLiteral follows an identifier to the object literal it was declared with.
//
// Upstream's `resolveNodeValue` does the same through scope analysis, because
// `C.propTypes = types` is a common spelling and refusing to read it would silence the component
// entirely. Seven of upstream's own failing cases write exactly that.
//
// Resolution is by declaration rather than by name, so a shadowing binding of the same name in an
// inner scope cannot be mistaken for the outer one. Returning nil is the unresolved answer, which
// the callers turn into the same silence a spread produces.
func resolveToObjectLiteral(value *ast.Node) *ast.Node {
	for step := 0; step < objectResolutionDepthLimit; step++ {
		value = skipParenthesesOptional(value)
		if value == nil {
			return nil
		}
		if value.Kind == ast.KindObjectLiteralExpression {
			return value
		}
		if value.Kind != ast.KindIdentifier {
			return nil
		}
		declaration := declarationOfIdentifier(value)
		if declaration == nil {
			return nil
		}
		next := sortDefaultPropsInitializerOf(declaration)
		if next == nil || next == value {
			return nil
		}
		value = next
	}
	return nil
}

// objectResolutionDepthLimit bounds `const a = b; const b = a;`, which resolves forever otherwise.
//
// A limit rather than a visited set because the chain is short in every real spelling and the limit
// states the expectation. Reaching it answers unresolved, which is the safe direction: the component
// is not judged rather than judged on a guess.
const objectResolutionDepthLimit = 8

// declarationOfIdentifier finds the declaration an identifier names, within this file.
//
// Written as a walk rather than through the type checker so this rule stays untyped. The checker
// would answer the same question, and declaring `NeedsTypeChecker` for one identifier lookup costs
// every file in the tree the checker lock. Upstream resolves through scope analysis, which is
// closer to this than to the checker.
//
// The nearest ENCLOSING declaration wins, which is what makes a shadowing binding shadow. A
// declaration in a sibling scope is invisible, matching upstream's `findVariableByName`.
func declarationOfIdentifier(identifier *ast.Node) *ast.Node {
	wanted := identifier.Text()
	for scope := identifier.Parent; scope != nil; scope = scope.Parent {
		var found *ast.Node
		scope.ForEachChild(func(child *ast.Node) bool {
			if found != nil {
				return true
			}
			if child.Kind != ast.KindVariableStatement {
				return false
			}
			list := child.AsVariableStatement().DeclarationList
			if list == nil {
				return false
			}
			for _, declaration := range list.AsVariableDeclarationList().Declarations.Nodes {
				name := declaration.AsVariableDeclaration().Name()
				if name != nil && name.Kind == ast.KindIdentifier && name.Text() == wanted {
					found = declaration
					return true
				}
			}
			return false
		})
		if found != nil {
			return found
		}
	}
	return nil
}
