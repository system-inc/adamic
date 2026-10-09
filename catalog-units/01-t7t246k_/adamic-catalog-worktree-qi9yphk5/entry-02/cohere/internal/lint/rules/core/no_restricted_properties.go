package core

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoRestrictedPropertiesRestriction is one entry in the rule's list.
//
// Upstream's schema requires at least one of `object` and `property`, and forbids pairing
// `allowObjects` with `object` or `allowProperties` with `property`, because each pair is
// self-contradictory: an entry naming one object has nothing to allow other objects for.
type NoRestrictedPropertiesRestriction struct {
	// Object restricts accesses through a receiver spelled this way. Compared by NAME, not by
	// binding: a local `foo` shadowing an import still matches, because this is a spelling rule.
	Object string `json:"object"`

	// Property restricts accesses to a property spelled this way.
	Property string `json:"property"`

	// AllowObjects exempts these receivers from a property-only restriction.
	AllowObjects []string `json:"allowObjects"`

	// AllowProperties exempts these properties from an object-only restriction.
	AllowProperties []string `json:"allowProperties"`

	// Message is appended to the finding, and is where a project says what to use instead.
	Message string `json:"message"`
}

// NoRestrictedPropertiesOptions is the rule's whole configuration.
//
// Upstream's options are variadic: each restriction object is its own option element,
// `["error", {"object": "a", "property": "b"}, {"property": "c"}]`. The rule registers with
// `DecodeOptionList`, so the decoder reads that list, one restriction per element, with upstream's
// own field names.
type NoRestrictedPropertiesOptions struct {
	// Restrictions is the list. Empty means the rule has nothing to enforce and registers nothing,
	// which is upstream's `if (restrictedCalls.length === 0) return {}`.
	Restrictions []NoRestrictedPropertiesRestriction
}

// DecodeNoRestrictedPropertiesOptions reads this rule's configuration from the config layer.
//
// Hand-rolled so an entry naming neither an object nor a property fails loudly rather than matching
// everything or nothing silently. Upstream expresses that as a schema `anyOf`, which we have no
// equivalent for, so the check lives here.
//
// Every element is one restriction, read strictly: a key upstream's schema does not declare is
// refused, which is its `additionalProperties: false`, because a misspelled `message` or
// `allowObjects` would otherwise be accepted and silently have no effect. The `{"restrictions": [...]}`
// wrapper this decoder used to read, when the config layer could deliver only one element, is gone
// with it: it was never upstream's spelling.
func DecodeNoRestrictedPropertiesOptions(list []byte) (any, error) {
	var options NoRestrictedPropertiesOptions
	if len(list) == 0 {
		return options, nil
	}
	var elements []json.RawMessage
	if err := json.Unmarshal(list, &elements); err != nil {
		return options, fmt.Errorf(
			"no-restricted-properties: the option list is not a JSON array: %w", err)
	}
	for index, element := range elements {
		var restriction NoRestrictedPropertiesRestriction
		if err := rule.UnmarshalOptions(element, &restriction); err != nil {
			return NoRestrictedPropertiesOptions{}, fmt.Errorf(
				"no-restricted-properties element %d: %w", index+1, err)
		}
		options.Restrictions = append(options.Restrictions, restriction)
	}
	for index, restriction := range options.Restrictions {
		if restriction.Object == "" && restriction.Property == "" {
			return options, fmt.Errorf(
				"no-restricted-properties: restriction %d names neither an object nor a "+
					"property, so it can never match anything", index)
		}
		// Upstream's schema forbids each of these pairs as self-contradictory: an entry naming one
		// object has nothing to allow other objects for.
		if restriction.Object != "" && len(restriction.AllowObjects) > 0 {
			return options, fmt.Errorf(
				"no-restricted-properties: restriction %d pairs `object` with `allowObjects`, "+
					"which restricts one receiver and then exempts receivers", index)
		}
		if restriction.Property != "" && len(restriction.AllowProperties) > 0 {
			return options, fmt.Errorf(
				"no-restricted-properties: restriction %d pairs `property` with "+
					"`allowProperties`, which restricts one property and then exempts properties",
				index)
		}
	}
	return options, nil
}

// NoRestrictedProperties reports an access to a property a project has banned.
//
//	valid:   foo.bar          // with no restriction naming it
//	invalid: foo.bar          // with { object: "foo", property: "bar" }
//	invalid: anything.bar     // with { property: "bar" }
//	invalid: foo.anything     // with { object: "foo" }
//	invalid: var { bar } = foo;
//
// # This rule enforces nothing until somebody configures it
//
// With an empty list upstream returns no listeners at all, and this reproduces that. That is worth
// saying plainly because it makes the rule's violation count structurally zero rather than
// evidence of a clean tree: an audit measuring this rule unconfigured reports no violations on a
// codebase full of whatever it would have banned. Verified against the installed build --
// `foo.bar; baz.qux;` with no options reports nothing.
//
// So it is registered here and NOT enabled. Enabling it means choosing restrictions, which is a
// decision about this codebase rather than a porting question.
//
// # Three kinds of restriction, checked in one order
//
//	object + property   the specific access `foo.bar`
//	object only         every property on `foo`, minus `allowProperties`
//	property only       that property on every receiver, minus `allowObjects`
//
// An object-specific match takes priority over a property-only one, so the two never both report on
// one access. That ordering is upstream's and it decides which MESSAGE a reader sees.
//
// # Names, not bindings
//
// The receiver is compared by its spelling, so a local `foo` shadowing an import still matches. That
// is right for a rule a project configures by name: the restriction is written as a string.
//
// A consequence worth stating because it looks like a bug: `foo.bar.baz` with `object: "bar"` is
// CLEAN, because the receiver of `.baz` is a member expression rather than an identifier and has no
// name to compare. Measured against the installed build.
var NoRestrictedProperties = rule.Rule{
	Name: "no-restricted-properties",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, configured := rule.OptionsAs[NoRestrictedPropertiesOptions](options)
		if !configured || len(settings.Restrictions) == 0 {
			// Upstream returns an empty visitor object here rather than a no-op listener, so the
			// rule costs nothing at all when it has nothing to enforce.
			return rule.Listeners{}
		}
		index := buildNoRestrictedPropertiesIndex(settings.Restrictions)

		return rule.Listeners{
			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				checkNoRestrictedPropertiesAccess(ctx, node, index)
			},
			ast.KindElementAccessExpression: func(node *ast.Node) {
				checkNoRestrictedPropertiesAccess(ctx, node, index)
			},
			ast.KindObjectBindingPattern: func(node *ast.Node) {
				checkNoRestrictedPropertiesPattern(ctx, node, index)
			},
			// Destructuring in an ASSIGNMENT position, which our parser gives a different node
			// kind than destructuring in a declaration. See the doc on the checker below.
			ast.KindObjectLiteralExpression: func(node *ast.Node) {
				checkNoRestrictedPropertiesAssignmentPattern(ctx, node, index)
			},
		}
	},
}

// noRestrictedPropertiesIndex is the configuration turned into the three lookups the rule makes.
//
// Built once per file rather than scanned per access, which matters because a project restricting
// twenty properties would otherwise do twenty comparisons on every member expression in the tree.
type noRestrictedPropertiesIndex struct {
	// byObjectAndProperty holds the fully-specific restrictions, keyed object then property.
	byObjectAndProperty map[string]map[string]NoRestrictedPropertiesRestriction
	// byObject holds the object-only restrictions.
	byObject map[string]NoRestrictedPropertiesRestriction
	// byProperty holds the property-only restrictions.
	byProperty map[string]NoRestrictedPropertiesRestriction
}

// buildNoRestrictedPropertiesIndex sorts the configured restrictions into their three lookups.
func buildNoRestrictedPropertiesIndex(
	restrictions []NoRestrictedPropertiesRestriction,
) noRestrictedPropertiesIndex {
	index := noRestrictedPropertiesIndex{
		byObjectAndProperty: map[string]map[string]NoRestrictedPropertiesRestriction{},
		byObject:            map[string]NoRestrictedPropertiesRestriction{},
		byProperty:          map[string]NoRestrictedPropertiesRestriction{},
	}
	for _, restriction := range restrictions {
		switch {
		case restriction.Object == "":
			index.byProperty[restriction.Property] = restriction
		case restriction.Property == "":
			index.byObject[restriction.Object] = restriction
		default:
			if index.byObjectAndProperty[restriction.Object] == nil {
				index.byObjectAndProperty[restriction.Object] =
					map[string]NoRestrictedPropertiesRestriction{}
			}
			index.byObjectAndProperty[restriction.Object][restriction.Property] = restriction
		}
	}
	return index
}

// checkNoRestrictedPropertiesAccess judges one member access.
func checkNoRestrictedPropertiesAccess(
	ctx rule.Context,
	node *ast.Node,
	index noRestrictedPropertiesIndex,
) {
	propertyName, named := noRestrictedPropertiesAccessedName(node)
	if !named {
		return
	}
	reportNoRestrictedProperties(ctx, node, index,
		noRestrictedPropertiesReceiverName(node), propertyName)
}

// checkNoRestrictedPropertiesPattern judges every property a destructuring pattern reads.
//
// `var { bar } = foo` reads `foo.bar` without writing a member expression, so the pattern is a
// second entry point into the same judgment. The receiver is the initializer when it is a bare
// identifier, and nothing otherwise -- `var { bar } = foo.baz` is clean, measured.
//
// Every finding points at the PATTERN rather than at the property, which is upstream's `node` and
// is why one pattern restricting two properties reports twice on the same span.
func checkNoRestrictedPropertiesPattern(
	ctx rule.Context,
	node *ast.Node,
	index noRestrictedPropertiesIndex,
) {
	objectName := noRestrictedPropertiesPatternSource(node)
	node.ForEachChild(func(element *ast.Node) bool {
		if element.Kind != ast.KindBindingElement {
			return false
		}
		binding := element.AsBindingElement()
		// A shorthand `{ bar }` has no property name node, so the binding name IS the property
		// read. A rest element `{ ...rest }` reads no single property and is skipped, which is
		// upstream's behaviour rather than an omission: it has no key to compare.
		key := binding.PropertyName
		if key == nil {
			if binding.DotDotDotToken != nil {
				return false
			}
			key = binding.Name()
		}
		propertyName, named := noRestrictedPropertiesKeyName(key)
		if !named {
			return false
		}
		reportNoRestrictedProperties(ctx, node, index, objectName, propertyName)
		return false
	})
}

// checkNoRestrictedPropertiesAssignmentPattern judges destructuring written as an assignment.
//
// # One ESTree node, two of ours
//
// ESTree calls both `var {bar} = foo` and `({bar} = foo)` an `ObjectPattern`, so upstream needs one
// listener. Our parser only knows a declaration is a pattern; in an assignment it has already
// committed to parsing `{bar}` as an object LITERAL, whose members are property assignments rather
// than binding elements. So the same judgment needs a second entry point over a different shape.
//
// This is invisible from upstream's source and invisible from its corpus verdicts. It surfaced as
// six imported cases reporting nothing, all of them assignment-position destructuring, while every
// declaration case passed. A port stopping at the binding pattern would have shipped with a whole
// syntactic form silently exempt.
//
// Not every object literal is a pattern, and the ones that are not must stay clean: `foo({bar: 1})`
// is an ordinary literal. The test is positional, walking up through the nesting a pattern can
// have -- a nested property, an array element -- to an assignment's left side.
func checkNoRestrictedPropertiesAssignmentPattern(
	ctx rule.Context,
	node *ast.Node,
	index noRestrictedPropertiesIndex,
) {
	if !noRestrictedPropertiesIsAssignmentTarget(node) {
		return
	}
	objectName := noRestrictedPropertiesAssignmentSource(node)
	node.ForEachChild(func(member *ast.Node) bool {
		var key *ast.Node
		switch member.Kind {
		case ast.KindPropertyAssignment:
			key = member.AsPropertyAssignment().Name()
		case ast.KindShorthandPropertyAssignment:
			key = member.AsShorthandPropertyAssignment().Name()
		default:
			// A spread reads no single property, matching the rest element on the binding side.
			return false
		}
		propertyName, named := noRestrictedPropertiesKeyName(key)
		if !named {
			return false
		}
		reportNoRestrictedProperties(ctx, node, index, objectName, propertyName)
		return false
	})
}

// noRestrictedPropertiesIsAssignmentTarget answers whether an object literal is being destructured.
//
// Walks outward through the shapes a pattern can nest inside -- a property's value, an array
// element -- and answers true only on reaching the left side of an `=`. An object literal anywhere
// else is an ordinary literal and this rule has nothing to say about it.
func noRestrictedPropertiesIsAssignmentTarget(node *ast.Node) bool {
	child := node
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		switch parent.Kind {
		case ast.KindPropertyAssignment, ast.KindArrayLiteralExpression,
			ast.KindShorthandPropertyAssignment, ast.KindSpreadAssignment,
			ast.KindSpreadElement, ast.KindParenthesizedExpression,
			// An enclosing object literal, which is what a NESTED pattern sits inside:
			// `({ bar: { bad } } = foo)` walks inner literal, property, outer literal, assignment.
			// Omitting this arm left exactly that shape silent while every flat pattern passed.
			ast.KindObjectLiteralExpression:
			// Still inside the pattern; keep walking.
		case ast.KindBinaryExpression:
			binary := parent.AsBinaryExpression()
			return binary.OperatorToken != nil &&
				binary.OperatorToken.Kind == ast.KindEqualsToken && binary.Left == child
		default:
			return false
		}
		child = parent
	}
	return false
}

// noRestrictedPropertiesAssignmentSource returns the name of the object an assignment pattern reads.
//
// Only the OUTERMOST pattern has one: in `({ bar: { bad } } = foo)` the inner pattern reads
// `foo.bar`, whose name is not an identifier, so it answers "" and matches only a property-only
// restriction. Measured -- upstream reports `bad` there with the property-only message.
func noRestrictedPropertiesAssignmentSource(node *ast.Node) string {
	parent := node.Parent
	if parent == nil || parent.Kind != ast.KindBinaryExpression {
		return ""
	}
	binary := parent.AsBinaryExpression()
	if binary.OperatorToken == nil || binary.OperatorToken.Kind != ast.KindEqualsToken ||
		binary.Left != node || binary.Right == nil {
		return ""
	}
	if right := ast.SkipParentheses(binary.Right); right.Kind == ast.KindIdentifier {
		return right.Text()
	}
	return ""
}

// noRestrictedPropertiesPatternSource returns the name of the object a pattern destructures.
//
// Three parents can supply one, and all three require a bare identifier: a declarator's initializer,
// an assignment's right side, and a parameter default. Anything else answers "", which makes the
// access match only a property-only restriction.
func noRestrictedPropertiesPatternSource(node *ast.Node) string {
	parent := node.Parent
	if parent == nil {
		return ""
	}
	var source *ast.Node
	switch parent.Kind {
	case ast.KindVariableDeclaration:
		source = parent.AsVariableDeclaration().Initializer
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		if binary.OperatorToken != nil && binary.OperatorToken.Kind == ast.KindEqualsToken {
			source = binary.Right
		}
	case ast.KindParameter:
		source = parent.AsParameterDeclaration().Initializer
	case ast.KindBindingElement:
		source = parent.AsBindingElement().Initializer
	}
	if source == nil || source.Kind != ast.KindIdentifier {
		return ""
	}
	return source.Text()
}

// reportNoRestrictedProperties applies the three lookups in upstream's order and reports at most
// once.
//
// The object-specific arms are consulted first and, crucially, an object-specific MATCH suppresses
// the property-only arm even when the specific one is exempted by `allowProperties`. That is
// upstream's `else if` and it is a real decision rather than a shortcut: a project that restricts
// an object while allowing one property has said that property is fine on that object, and a
// property-only rule should not then override it.
func reportNoRestrictedProperties(
	ctx rule.Context,
	node *ast.Node,
	index noRestrictedPropertiesIndex,
	objectName string,
	propertyName string,
) {
	// A fully-specific restriction wins; failing that, an object-only one.
	matched, matchedObject := NoRestrictedPropertiesRestriction{}, false
	if byProperty, known := index.byObjectAndProperty[objectName]; known {
		matched, matchedObject = byProperty[propertyName]
	}
	if !matchedObject {
		matched, matchedObject = index.byObject[objectName]
	}

	if matchedObject {
		if noRestrictedPropertiesListed(propertyName, matched.AllowProperties) {
			// Exempted, and the property-only arm is NOT consulted. See the doc above.
			return
		}
		allowance := ""
		if matched.AllowProperties != nil {
			allowance = fmt.Sprintf(" Only these properties are allowed: %s.",
				strings.Join(matched.AllowProperties, ", "))
		}
		ctx.ReportNode(node, rule.Message{
			Id: "restrictedObjectProperty",
			Description: fmt.Sprintf("`%s.%s` is restricted from being used.%s%s",
				objectName, propertyName, allowance,
				noRestrictedPropertiesSuffix(matched.Message)),
		})
		return
	}

	globalMatch, matchedProperty := index.byProperty[propertyName]
	if !matchedProperty ||
		noRestrictedPropertiesListed(objectName, globalMatch.AllowObjects) {
		return
	}
	allowance := ""
	if globalMatch.AllowObjects != nil {
		allowance = fmt.Sprintf(" Property `%s` is only allowed on these objects: %s.",
			propertyName, strings.Join(globalMatch.AllowObjects, ", "))
	}
	ctx.ReportNode(node, rule.Message{
		Id: "restrictedProperty",
		Description: fmt.Sprintf("`%s` is restricted from being used.%s%s",
			propertyName, allowance, noRestrictedPropertiesSuffix(globalMatch.Message)),
	})
}

// noRestrictedPropertiesSuffix renders a configured message, space-separated, or nothing.
func noRestrictedPropertiesSuffix(message string) string {
	if message == "" {
		return ""
	}
	return " " + message
}

// noRestrictedPropertiesListed answers whether a name is on an allowance list.
//
// A nil list is not an empty allowance: it means no allowance was configured, and upstream's
// `if (!allowedList) return false` says the same thing. The distinction is visible in the rendered
// message, which names the allowed set only when one exists.
func noRestrictedPropertiesListed(name string, allowed []string) bool {
	for _, candidate := range allowed {
		if candidate == name {
			return true
		}
	}
	return false
}

// noRestrictedPropertiesReceiverName returns the spelling of an access's receiver, or "".
//
// Upstream reads `node.object && node.object.name`, which is undefined for anything but an
// identifier. So `foo.bar.baz` has no receiver name for its outer access, and `(foo).bar` DOES have
// one, because espree gives parentheses no node. The parenthesis skip here is what reproduces that
// second case rather than widening the rule.
func noRestrictedPropertiesReceiverName(node *ast.Node) string {
	var receiver *ast.Node
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		receiver = node.AsPropertyAccessExpression().Expression
	case ast.KindElementAccessExpression:
		receiver = node.AsElementAccessExpression().Expression
	}
	if receiver == nil {
		return ""
	}
	if unwrapped := ast.SkipParentheses(receiver); unwrapped.Kind == ast.KindIdentifier {
		return unwrapped.Text()
	}
	return ""
}

// noRestrictedPropertiesAccessedName returns the property an access reads, when the syntax settles
// it.
//
// # Why this is not `property.AccessedName`
//
// The shelf helper is close and disagrees with upstream on four of the eight spellings this rule can
// meet, which was measured with a probe rather than read off its doc comment:
//
//	spelling            upstream            property.Static
//	`this.#foo`         null, so CLEAN      "#foo", so it would report
//	`foo[/re/]`         "/re/", reports     nothing, so it would go silent
//	`foo[null]`         "null", reports     nothing, so it would go silent
//	`foo[1n]`           "1", reports        nothing, so it would go silent
//
// Two of those cost findings and one adds a false one, and the private-name row is in upstream's own
// corpus as a REPORTING case for `foo['#foo']` beside a clean `this.#foo`. Widening the shelf helper
// would change four other rules, so this rule carries its own predicate and the shelf keeps its.
func noRestrictedPropertiesAccessedName(node *ast.Node) (string, bool) {
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		return noRestrictedPropertiesKeyName(node.AsPropertyAccessExpression().Name())
	case ast.KindElementAccessExpression:
		argument := node.AsElementAccessExpression().ArgumentExpression
		if argument == nil {
			return "", false
		}
		return noRestrictedPropertiesLiteralName(ast.SkipParentheses(argument))
	}
	return "", false
}

// noRestrictedPropertiesKeyName returns the property a key node names.
//
// A bare identifier key is its own name; everything else goes through the literal reader, which is
// upstream's `getStaticPropertyName` splitting on `computed`.
func noRestrictedPropertiesKeyName(key *ast.Node) (string, bool) {
	if key == nil {
		return "", false
	}
	if key.Kind == ast.KindIdentifier {
		return key.Text(), true
	}
	if key.Kind == ast.KindComputedPropertyName {
		inner := key.AsComputedPropertyName().Expression
		if inner == nil {
			return "", false
		}
		return noRestrictedPropertiesLiteralName(ast.SkipParentheses(inner))
	}
	return noRestrictedPropertiesLiteralName(key)
}

// noRestrictedPropertiesLiteralName renders a literal the way upstream's `getStaticStringValue`
// does.
//
// Upstream stringifies the literal's VALUE, which is why a regular expression answers its own source
// including the slashes, `null` answers "null", and a bigint answers its digits without the `n`.
// Each of those is a spelling somebody could plausibly restrict, and each was measured against the
// installed build rather than inferred.
//
// A private identifier answers nothing, which is what keeps `this.#foo` clean while `foo['#foo']`
// reports. Upstream reaches that by falling off the end of a switch rather than by a deliberate
// arm, so it reads like an oversight; it is reproduced because the corpus asserts both halves.
func noRestrictedPropertiesLiteralName(node *ast.Node) (string, bool) {
	if node == nil {
		return "", false
	}
	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		return node.Text(), true
	case ast.KindNumericLiteral:
		return node.Text(), true
	case ast.KindBigIntLiteral:
		// `String(1n)` is "1": the suffix is syntax rather than part of the value.
		return strings.TrimSuffix(node.Text(), "n"), true
	case ast.KindTrueKeyword:
		return "true", true
	case ast.KindFalseKeyword:
		return "false", true
	case ast.KindNullKeyword:
		return "null", true
	case ast.KindRegularExpressionLiteral:
		// The literal's own source, slashes and flags included, which is what `String(/a/)` gives.
		return node.Text(), true
	}
	return "", false
}
