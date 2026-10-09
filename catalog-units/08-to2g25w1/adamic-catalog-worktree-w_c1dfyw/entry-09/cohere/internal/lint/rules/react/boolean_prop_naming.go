package react

import (
	"errors"
	"regexp"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// BooleanPropNamingOptions configures the rule.
//
// Upstream's schema declares a `default` for `rule` and `validateNested`, and ESLint applies a
// schema default only when the options OBJECT is present. With no options at all,
// `context.options[0]` is undefined, `config.rule` is null, and every listener early-returns: the
// rule is completely inert. Measured on the installed build, one input three ways: no options
// reports zero, `{}` reports one, and an explicit rule reports one.
//
// That is why this rule declares `RequiresOptions` at registration. Enabling it as a bare `"error"`
// would reproduce upstream's inertness exactly, which is faithful and useless: it would sit in the
// config looking enforced while enforcing nothing, which is the shape this project has already been
// bitten by twice.
type BooleanPropNamingOptions struct {
	// Rule is the pattern a boolean prop's name must match.
	Rule string `json:"rule"`

	// PropTypeNames are the prop-type member names treated as boolean. Absent means `bool`.
	PropTypeNames []string `json:"propTypeNames"`

	// ValidateNested descends into `PropTypes.shape({...})` arguments. Absent means false.
	ValidateNested bool `json:"validateNested"`

	// Message replaces the default finding text entirely when set.
	Message string `json:"message"`
}

// DefaultBooleanPropNamingRulePattern is upstream's schema default.
//
// Reproduced as a constant rather than left to the decoder's zero value, because the zero value is
// the empty string and an empty pattern matches everything, which would silence the rule instead of
// applying the documented default.
const DefaultBooleanPropNamingRulePattern = "^(is|has)[A-Z]([A-Za-z0-9]?)+"

// defaultBooleanPropTypeName is the single prop-type member upstream treats as boolean by default.
const defaultBooleanPropTypeName = "bool"

// DefaultBooleanPropNamingOptions is the configuration an options object with no keys resolves to.
//
// This matches ESLint handing `{}` to the rule, which is the only unconfigured shape that does
// anything at all. See the type's doc comment for why the no-options shape is refused instead.
func DefaultBooleanPropNamingOptions() BooleanPropNamingOptions {
	return BooleanPropNamingOptions{
		Rule:           DefaultBooleanPropNamingRulePattern,
		PropTypeNames:  []string{defaultBooleanPropTypeName},
		ValidateNested: false,
	}
}

// DecodeBooleanPropNamingOptions reads this rule's configuration from the config layer.
//
// Hand-rolled rather than `rule.DecodeOptionsInto` because two keys carry non-zero defaults that
// the generic helper would leave empty: an absent `rule` has to become the documented pattern
// rather than the empty string, and an absent `propTypeNames` has to become `["bool"]` rather than
// an empty list. Both zero values silence the rule rather than merely narrowing it.
func DecodeBooleanPropNamingOptions(raw []byte) (any, error) {
	options := DefaultBooleanPropNamingOptions()
	if len(raw) == 0 {
		// Refusing empty input is what makes `RequiresOptions` mean anything. That field only
		// changes behaviour when the DECODER errors: `OptionsRegistry.Decode` swallows a decode
		// error on empty input for a rule that is not required, and propagates it for one that is.
		// A decoder returning defaults here would make the flag inert, which an earlier draft did,
		// and the rule then ran tree-wide on a bare `"error"` exactly as upstream does not.
		return options, errBooleanPropNamingNeedsOptions
	}

	// Decoded into a wire struct with pointer-free fields, then defaulted, so an explicitly empty
	// list is distinguishable from an absent key only where that distinction matters. Upstream
	// reads `config.propTypeNames || ['bool']`, so an explicitly empty array ALSO falls back, which
	// this reproduces by testing length rather than nil.
	var wire BooleanPropNamingOptions
	if err := rule.UnmarshalOptions(raw, &wire); err != nil {
		return options, err
	}
	if wire.Rule != "" {
		options.Rule = wire.Rule
	}
	if len(wire.PropTypeNames) > 0 {
		options.PropTypeNames = wire.PropTypeNames
	}
	options.ValidateNested = wire.ValidateNested
	options.Message = wire.Message

	if _, err := regexp.Compile(options.Rule); err != nil {
		return options, err
	}
	return options, nil
}

// messageBooleanPropNaming builds the finding for one mismatched prop.
//
// The apostrophe in "doesn't" is upstream's own U+2019, written as the escape \u2019 rather than as
// a literal so it is visible in review and cannot be mistaken for the editor damage this project has
// twice had to clean up. A straight apostrophe here would diverge from every other tool's output for
// the same finding.
//
// An earlier draft carried this same comment above a LITERAL curly apostrophe, which is the exact
// shape of claiming something in a comment and citing the comment back as evidence. The byte scan
// is what caught it.
func messageBooleanPropNaming(propertyName string, pattern string, override string) rule.Message {
	if override != "" {
		// Upstream reports the configured message with no id at all when `message` is set. Our
		// report surface always carries an id, so the id stays and only the text is replaced, which
		// is the closest expressible thing and is recorded in the test file.
		//
		// The placeholders are ESLint's own `{{ name }}` interpolation, which the rule feeds three
		// values: `propName`, `pattern`, and `component`, the last of which upstream sets to the
		// prop name too rather than to the component's name. Both the spaced and unspaced spellings
		// are accepted, measured against the installed build on all three names.
		return rule.Message{Id: "patternMismatch", Description: interpolateMessage(override, map[string]string{
			"propName":  propertyName,
			"pattern":   pattern,
			"component": propertyName,
		})}
	}
	return rule.Message{
		Id:          "patternMismatch",
		Description: "Prop name `" + propertyName + "` doesn\u2019t match rule `" + pattern + "`",
	}
}

// BooleanPropNaming reports a boolean prop whose name does not match a configured pattern.
//
//	valid:   C.propTypes = { isEnabled: PropTypes.bool }
//	valid:   C.propTypes = { count: PropTypes.number }      not boolean, not judged
//	invalid: C.propTypes = { enabled: PropTypes.bool }
//	invalid: interface Props { enabled: boolean }
//
// The regex is the easy half. The work is knowing which props are BOOLEAN, and upstream reads that
// from three independent sources: a `propTypes` object, a TypeScript props type, and the prop-type
// member names in `propTypeNames`.
//
// # Requires options, deliberately
//
// Registered with `RequiresOptions` because upstream is inert without an options object; see
// BooleanPropNamingOptions. A bare `"error"` here would be a rule that lints nothing.
var BooleanPropNaming = rule.Rule{
	Name: "react/boolean-prop-naming",

	// The TypeScript half resolves a props type reference to its declaration, which is what lets
	// `function C(p: Props)` reach `interface Props`. Upstream keeps its own map keyed by name;
	// resolution answers the same question without being fooled by a shadowing declaration.
	NeedsTypeChecker: true,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}
		// Declared once per file rather than once per listener. `NeedsTypeChecker` governs the
		// registration path and says nothing about the harness path, where a Context is built by
		// hand and the checker can be nil.
		if ctx.TypeChecker == nil {
			return nil
		}

		settings, hasSettings := rule.OptionsAs[BooleanPropNamingOptions](options)
		if !hasSettings {
			// RequiresOptions means the config layer declines the file before this, so reaching
			// here without options is a harness path. Declining matches upstream's own behaviour
			// with no options object rather than inventing a default nobody configured.
			//
			// This decline is not observable in a findings count and a mutation neutralising it
			// survived the whole suite. The reason, measured by A/B on 2026-08-28: without it the
			// zero-value Rule is the empty string, `regexp.Compile("")` succeeds, and an empty
			// pattern matches every name, so the rule reports nothing either way.
			//
			// Kept because the two are equal only in output. Without it the rule registers a
			// listener on every file and walks each one to build findings it will never emit, and
			// a later edit making the empty pattern mean something would turn a silent decline
			// into silent over-reporting. Declining says what is meant.
			return nil
		}
		pattern, err := regexp.Compile(settings.Rule)
		if err != nil {
			return nil
		}

		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				checkBooleanPropNaming(ctx, node, settings, pattern)
			},
		}
	},
}

// checkBooleanPropNaming runs the whole analysis for one file.
//
// Upstream gathers across several listeners and reports in `Program:exit`, because a `propTypes`
// assignment can sit far below the component and a props interface can be declared anywhere in the
// file. Our walk is pre-order, so a source-file listener fires before its children and the
// gathering has to happen here.
func checkBooleanPropNaming(
	ctx rule.Context,
	sourceFile *ast.Node,
	settings BooleanPropNamingOptions,
	pattern *regexp.Regexp,
) {
	// Findings are collected before reporting so a prop reachable through two routes, a propTypes
	// object and a props type, reports once rather than twice.
	reported := map[*ast.Node]bool{}
	report := func(property *ast.Node, name string) {
		if reported[property] {
			return
		}
		reported[property] = true
		ctx.ReportNode(property, messageBooleanPropNaming(name, settings.Rule, settings.Message))
	}

	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		switch node.Kind {
		case ast.KindPropertyDeclaration:
			// `static propTypes = {...}` inside a class body.
			declaration := node.AsPropertyDeclaration()
			if !isPropTypesName(declaration.Name()) {
				break
			}
			checkPropTypesValue(ctx, declaration.Initializer, settings, pattern, report)

		case ast.KindBinaryExpression:
			// `C.propTypes = {...}` written after the class or function.
			binary := node.AsBinaryExpression()
			if binary.OperatorToken == nil || binary.OperatorToken.Kind != ast.KindEqualsToken {
				break
			}
			target := skipParenthesesOptional(binary.Left)
			if target == nil || target.Kind != ast.KindPropertyAccessExpression {
				break
			}
			if !isPropTypesName(target.AsPropertyAccessExpression().Name()) {
				break
			}
			checkPropTypesValue(ctx, binary.Right, settings, pattern, report)

		case ast.KindPropertyAssignment:
			// `{ propTypes: {...} }` inside an object passed to the ES5 factory.
			if !isPropTypesName(node.AsPropertyAssignment().Name()) {
				break
			}
			checkPropTypesValue(ctx, node.AsPropertyAssignment().Initializer, settings, pattern, report)
		}

		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}
	visit(sourceFile)

	// The TypeScript half. Every detected component's first parameter type is resolved and its
	// members judged, which is what reaches `function C(p: Props)` through an interface declared
	// elsewhere in the file.
	for _, component := range collectDetectedComponents(ctx, sourceFile) {
		propertiesType := componentPropsTypeNode(component.node)
		if propertiesType == nil {
			continue
		}
		for _, member := range collectTypeMembers(ctx, propertiesType, map[*ast.Node]bool{}) {
			checkTypeMember(member, pattern, report)
		}
	}
}

// isPropTypesName reports whether a member name is `propTypes`.
//
// Upstream's `isPropTypesDeclaration` also accepts a string-literal key, so `'propTypes'` counts.
// The kind is checked before reading the text because `Text()` panics on several name shapes, and
// a computed key is a binding pattern's cousin in that respect.
func isPropTypesName(name *ast.Node) bool {
	if name == nil {
		return false
	}
	switch name.Kind {
	case ast.KindIdentifier, ast.KindStringLiteral:
		return name.Text() == "propTypes"
	}
	return false
}

// checkPropTypesValue judges the object literal a `propTypes` declaration holds.
//
// # A wrapper call is deliberately NOT unwrapped
//
// Upstream also accepts `propTypes = forbidExtraProps({...})`, but only when the callee is named in
// `settings.react.propWrapperFunctions`. That list is a settings surface we do not have, and with it
// unconfigured upstream unwraps NOTHING: measured on the installed build with no settings,
// `wrap({...})`, `forbidExtraProps({...})` and `merge({}, x, {...})` all report zero while the same
// props in a bare object report one. Configuring `forbidExtraProps` flips only that one to one.
//
// So declining every call form is not a narrowing of upstream, it IS upstream in the only
// configuration we can express. An earlier draft unwrapped calls unconditionally on the reasoning
// that a missing gate should fail open; that reported on two of upstream's own passing cases, both
// of which write a wrapper the corpus deliberately leaves unconfigured.
func checkPropTypesValue(
	ctx rule.Context,
	value *ast.Node,
	settings BooleanPropNamingOptions,
	pattern *regexp.Regexp,
	report func(*ast.Node, string),
) {
	value = skipParenthesesOptional(value)
	if value == nil || value.Kind != ast.KindObjectLiteralExpression {
		return
	}
	runBooleanPropCheck(value.AsObjectLiteralExpression().Properties, settings, pattern, report)
}

// runBooleanPropCheck walks one propTypes object's own properties.
//
// Upstream recurses only when `validateNested` is set AND the property's value is a call, which is
// how `PropTypes.shape({...})` is reached. Both halves are load bearing and measured: the same
// shape input reports zero under the default and one with the option on.
func runBooleanPropCheck(
	properties *ast.NodeList,
	settings BooleanPropNamingOptions,
	pattern *regexp.Regexp,
	report func(*ast.Node, string),
) {
	if properties == nil {
		return
	}
	for _, property := range properties.Nodes {
		if property.Kind != ast.KindPropertyAssignment {
			// A spread carries no key. Upstream returns null from `getPropKey` for it and the
			// property is skipped rather than reported.
			continue
		}
		assignment := property.AsPropertyAssignment()

		if settings.ValidateNested && isNestedPropTypesCall(assignment.Initializer) {
			nested := skipParenthesesOptional(assignment.Initializer)
			arguments := nested.AsCallExpression().Arguments
			if arguments != nil && len(arguments.Nodes) > 0 {
				first := skipParenthesesOptional(arguments.Nodes[0])
				if first != nil && first.Kind == ast.KindObjectLiteralExpression {
					runBooleanPropCheck(first.AsObjectLiteralExpression().Properties,
						settings, pattern, report)
				}
			}
			continue
		}

		if !isBooleanPropTypeValue(assignment.Initializer, settings.PropTypeNames) {
			continue
		}
		name, hasName := propTypesKeyName(assignment.Name())
		if pattern.MatchString(name) && hasName {
			continue
		}
		report(property, name)
	}
}

// isNestedPropTypesCall reports whether a prop's declared type is a call, which is upstream's
// `nestedPropTypes` test and is what `PropTypes.shape({...})` matches.
//
// Upstream tests only that the value IS a call, not that the callee is `shape`, so
// `PropTypes.oneOfType([...])` and any other call reach the recursion too. Reproduced rather than
// narrowed: the corpus writes `shape` and `arrayOf` both.
func isNestedPropTypesCall(value *ast.Node) bool {
	value = skipParenthesesOptional(value)
	return value != nil && value.Kind == ast.KindCallExpression
}

// isBooleanPropTypeValue reports whether a declared prop type names one of the boolean prop-type
// members.
//
// Upstream's `getPropKey` accepts four spellings and all four are measured reporting:
//
//	PropTypes.bool             a property access, the member name is read
//	React.PropTypes.bool       a deeper access, still the last member name
//	bool                       a bare identifier
//	PropTypes.bool.isRequired  the `isRequired` suffix is peeled and the name beneath is read
//
// The `isRequired` peel is one level only, matching upstream: it reads `node.value.object.property`
// and stops, so `PropTypes.bool.isRequired.isRequired` yields nothing and is silent.
func isBooleanPropTypeValue(value *ast.Node, propTypeNames []string) bool {
	value = skipParenthesesOptional(value)
	if value == nil {
		return false
	}

	name := ""
	switch value.Kind {
	case ast.KindIdentifier:
		name = value.Text()

	case ast.KindPropertyAccessExpression:
		access := value.AsPropertyAccessExpression()
		property := access.Name()
		if property == nil || property.Kind != ast.KindIdentifier {
			return false
		}
		if property.Text() == "isRequired" {
			object := skipParenthesesOptional(access.Expression)
			if object == nil || object.Kind != ast.KindPropertyAccessExpression {
				return false
			}
			inner := object.AsPropertyAccessExpression().Name()
			if inner == nil || inner.Kind != ast.KindIdentifier {
				return false
			}
			name = inner.Text()
		} else {
			name = property.Text()
		}

	default:
		return false
	}

	for _, wanted := range propTypeNames {
		if name == wanted {
			return true
		}
	}
	return false
}

// propTypesKeyName reads the name upstream would read from a propTypes property key.
//
// # An upstream defect, reproduced deliberately
//
// Upstream reads `node.key.name`, which exists only on an identifier key. A QUOTED key such as
// `'isEnabled'` and a numeric key both yield undefined, the regex is tested against the string
// "undefined", and the prop reports no matter what it is called. Measured against the installed
// build on 2026-08-28: `'enabled'`, `'isEnabled'` and `1` all report, and the message for each
// renders the literal word undefined.
//
// So a quoted key can never satisfy this rule upstream. That is reproduced here, because which
// inputs report is the rule's contract, and silently accepting `'isEnabled'` would make this port
// disagree with every other tool on the same file. The second return says whether a real name was
// read, so the caller can report even when the pattern happens to match the placeholder text.
//
// A computed key is different and is NOT a defect: upstream reads `node.key.name` and a computed
// `[k]` has an identifier `k` underneath, so it renders `k` and is judged against it. Measured.
func propTypesKeyName(key *ast.Node) (string, bool) {
	if key == nil {
		return "undefined", false
	}
	switch key.Kind {
	case ast.KindIdentifier:
		return key.Text(), true

	case ast.KindComputedPropertyName:
		inner := skipParenthesesOptional(key.AsComputedPropertyName().Expression)
		if inner != nil && inner.Kind == ast.KindIdentifier {
			return inner.Text(), true
		}
		return "undefined", false

	default:
		// A string or numeric literal key. Upstream renders undefined and always reports.
		return "undefined", false
	}
}

// componentPropsTypeNode returns the type node of a component's first parameter, or nil.
//
// Upstream also handles a props type reached through a generic annotation on the BINDING, as in
// `const C: React.FC<Props> = ...`, by reading the type argument list. That form is handled here
// too, because it is the common modern spelling and the corpus writes it.
func componentPropsTypeNode(component *ast.Node) *ast.Node {
	if parameters := componentParameters(component); parameters != nil && len(parameters.Nodes) > 0 {
		if declared := parameters.Nodes[0].AsParameterDeclaration().Type; declared != nil {
			return declared
		}
	}

	// `const C: React.FC<Props> = (p) => ...`. The annotation sits on the variable, and the props
	// type is the first type argument that is a reference or a literal.
	parent := semanticParentOf(component)
	if parent == nil || parent.Kind != ast.KindVariableDeclaration {
		return nil
	}
	annotation := parent.AsVariableDeclaration().Type
	if annotation == nil || annotation.Kind != ast.KindTypeReference {
		return nil
	}
	arguments := annotation.AsTypeReferenceNode().TypeArguments
	if arguments == nil {
		return nil
	}
	for _, argument := range arguments.Nodes {
		switch argument.Kind {
		case ast.KindTypeReference, ast.KindTypeLiteral:
			return argument
		}
	}
	return nil
}

// collectTypeMembers flattens a props type into the property signatures it declares.
//
// # What our parser gives, measured rather than ported
//
// Upstream unwraps five shapes and two of them have no input here at all:
//
//	TSInterfaceBody       ABSENT. An interface's members hang directly off the declaration, so
//	                      there is no body node to step through.
//	ObjectTypeAnnotation  ABSENT. That is Flow, and this parser produces no Flow nodes.
//
// One that reads as vestigial is LIVE and would have been dropped by a port that only followed
// upstream's estree shapes:
//
//	TSParenthesizedType   PRESENT, at depth two. `type Props = ({ a: boolean } & { b: boolean })`
//	                      wraps the intersection in a real KindParenthesizedType node, so without
//	                      the unwrap the members are never reached and the rule goes silent on
//	                      exactly the shape somebody wrote parentheses around for clarity.
//
// All three measured with a probe over six declaration shapes on 2026-08-28, with the walk
// disabled as a control to prove the probe could report nothing.
//
// The visited set breaks reference cycles: `interface A { self: A }` resolves back to itself and
// would otherwise recurse forever.
func collectTypeMembers(ctx rule.Context, typeNode *ast.Node, visited map[*ast.Node]bool) []*ast.Node {
	if typeNode == nil || visited[typeNode] {
		return nil
	}
	visited[typeNode] = true

	switch typeNode.Kind {
	case ast.KindParenthesizedType:
		return collectTypeMembers(ctx, typeNode.AsParenthesizedTypeNode().Type, visited)

	case ast.KindTypeLiteral:
		return typeNode.AsTypeLiteralNode().Members.Nodes

	case ast.KindInterfaceDeclaration:
		// Members hang directly off the declaration. Upstream steps through a `TSInterfaceBody`
		// here; that node does not exist in this parser, measured.
		return typeNode.AsInterfaceDeclaration().Members.Nodes

	case ast.KindIntersectionType:
		var members []*ast.Node
		for _, constituent := range typeNode.AsIntersectionTypeNode().Types.Nodes {
			members = append(members, collectTypeMembers(ctx, constituent, visited)...)
		}
		return members

	case ast.KindUnionType:
		// Upstream walks union members exactly as it walks intersection members, in
		// `findAllTypeAnnotations`. Reproduced rather than narrowed, even though a union of prop
		// shapes is an unusual thing to write.
		var members []*ast.Node
		for _, constituent := range typeNode.AsUnionTypeNode().Types.Nodes {
			members = append(members, collectTypeMembers(ctx, constituent, visited)...)
		}
		return members

	case ast.KindTypeReference:
		// A named props type. Upstream keeps its own map from name to declaration; resolution
		// answers the same question and is not fooled by a shadowing declaration of the same name.
		return collectTypeMembers(ctx, resolvedTypeDeclarationBody(ctx, typeNode), visited)
	}
	return nil
}

// resolvedTypeDeclarationBody resolves a type reference to the members-bearing node it names.
//
// An interface resolves to the declaration itself, because our parser hangs members directly off it
// with no body wrapper. A type alias resolves to whatever it aliases, which is then walked by the
// caller and may itself be a parenthesized intersection.
func resolvedTypeDeclarationBody(ctx rule.Context, reference *ast.Node) *ast.Node {
	typeName := reference.AsTypeReferenceNode().TypeName
	if typeName == nil || typeName.Kind != ast.KindIdentifier {
		return nil
	}
	symbol := ctx.TypeChecker.GetSymbolAtLocation(typeName)
	if symbol == nil {
		return nil
	}
	// Every declaration is offered rather than the first, because a symbol can carry more than one
	// through declaration merging and the question is "which declarations contribute members",
	// not "what single thing is this name".
	for _, declaration := range symbol.Declarations {
		switch declaration.Kind {
		case ast.KindInterfaceDeclaration:
			return declaration
		case ast.KindTypeAliasDeclaration:
			return declaration.AsTypeAliasDeclaration().Type
		}
	}
	return nil
}

// checkTypeMember judges one member of a props type.
//
// Upstream's `tsCheck` accepts a member only when it is a property signature whose annotation is
// exactly the boolean keyword, so an optional `enabled?: boolean` counts and a
// `enabled: boolean | undefined` does not. Both are reproduced by testing the annotation kind
// directly rather than asking the checker whether the type is assignable to boolean, which would be
// a wider question than upstream asks.
func checkTypeMember(member *ast.Node, pattern *regexp.Regexp, report func(*ast.Node, string)) {
	if member == nil || member.Kind != ast.KindPropertySignature {
		return
	}
	signature := member.AsPropertySignatureDeclaration()
	if signature.Type == nil || signature.Type.Kind != ast.KindBooleanKeyword {
		return
	}

	name, hasName := propTypesKeyName(signature.Name())
	if pattern.MatchString(name) && hasName {
		return
	}
	report(member, name)
}

// messagePlaceholderPattern matches ESLint's `{{ name }}` interpolation, with optional surrounding
// whitespace, which is the spelling its own `interpolate` accepts.
var messagePlaceholderPattern = regexp.MustCompile(`\{\{\s*([A-Za-z_][A-Za-z0-9_]*)\s*\}\}`)

// interpolateMessage substitutes ESLint's message placeholders.
//
// A placeholder naming something the rule does not provide is left untouched, which is ESLint's own
// behaviour rather than an omission: its `interpolate` returns the raw text when the key is absent.
func interpolateMessage(template string, values map[string]string) string {
	return messagePlaceholderPattern.ReplaceAllStringFunc(template, func(match string) string {
		name := messagePlaceholderPattern.FindStringSubmatch(match)[1]
		if replacement, hasValue := values[name]; hasValue {
			return replacement
		}
		return match
	})
}

// errBooleanPropNamingNeedsOptions is returned for an absent options object.
//
// The text names the pattern because that is the key somebody has to write, and a configuration
// error a reader cannot act on is barely better than silence.
var errBooleanPropNamingNeedsOptions = errors.New(
	"needs an options object, because upstream reads its pattern from the options and is inert " +
		"without one; write {\"rule\": \"" + DefaultBooleanPropNamingRulePattern + "\"} to get " +
		"upstream's documented default")
