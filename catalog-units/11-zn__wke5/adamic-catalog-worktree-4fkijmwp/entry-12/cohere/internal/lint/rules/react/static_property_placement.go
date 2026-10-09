package react

import (
	"encoding/json"
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// StaticPropertyPlacementPosition names one of the three places a React static may be written.
type StaticPropertyPlacementPosition string

const (
	// StaticPublicField is `static propTypes = {}` inside the class body.
	StaticPublicField StaticPropertyPlacementPosition = "StaticPublicField"

	// StaticGetter is `static get propTypes() { return {}; }` inside the class body.
	StaticGetter StaticPropertyPlacementPosition = "StaticGetter"

	// PropertyAssignment is `MyComponent.propTypes = {}` after the class.
	PropertyAssignment StaticPropertyPlacementPosition = "PropertyAssignment"
)

// StaticPropertyPlacementOptions is the expected position for each of the six properties.
//
// Upstream takes two positional options: a default position, then an object overriding it per
// property. Both are flattened into this map at decode time, so the rule reads one answer per
// property and never has to re-apply the default.
type StaticPropertyPlacementOptions struct {
	Positions map[string]StaticPropertyPlacementPosition
}

// staticPropertyPlacementProperties is the fixed set of names this rule judges, in upstream's own
// declaration order.
//
// The order is load-bearing rather than cosmetic. Upstream finds the FIRST matching name and
// reports under it, and `props` with a type annotation matches `propTypes` while `context` with one
// matches `contextTypes`, so a node could in principle answer two predicates. Keeping the order
// keeps the reported name the same as upstream's.
var staticPropertyPlacementProperties = []string{
	"propTypes",
	"defaultProps",
	"childContextTypes",
	"contextTypes",
	"contextType",
	"displayName",
}

// DefaultStaticPropertyPlacementOptions is the unconfigured answer: every property is expected as a
// static public field.
func DefaultStaticPropertyPlacementOptions() StaticPropertyPlacementOptions {
	positions := make(map[string]StaticPropertyPlacementPosition, len(staticPropertyPlacementProperties))
	for _, property := range staticPropertyPlacementProperties {
		positions[property] = StaticPublicField
	}
	return StaticPropertyPlacementOptions{Positions: positions}
}

// DecodeStaticPropertyPlacementOptions reads upstream's option list off the config.
//
// Upstream's option surface is POSITIONAL, `[default, overrides]`, and the rule registers with
// `DecodeOptionList` so both slots arrive:
//
//	nothing                                                    a static public field everywhere
//	["static getter"]                                          that position for every property
//	["property assignment", {"displayName": "static getter"}]  default plus per-property overrides
//
// Upstream's schema refuses a position outside its three-value enum and an override key outside its
// six properties, before the rule runs. This decoder used to let each fall back to the default
// instead, which is a configured value read and then ignored. Each is refused now, as is a third
// element. The bare-string spelling this decoder accepted while the config layer delivered one
// element is refused too: it was never upstream's.
func DecodeStaticPropertyPlacementOptions(list []byte) (any, error) {
	options := DefaultStaticPropertyPlacementOptions()
	slots, err := rule.OptionElements(list, 2)
	if err != nil || len(slots) == 0 {
		return options, err
	}

	var defaultSpelling string
	if err := json.Unmarshal(slots[0], &defaultSpelling); err != nil {
		return options, fmt.Errorf("static-property-placement element 1 takes a position string, got %s", slots[0])
	}
	position, recognised := staticPropertyPlacementPositionFor(defaultSpelling)
	if !recognised {
		return options, fmt.Errorf("static-property-placement: unknown position %q, wanted one of "+
			"\"static public field\", \"static getter\", \"property assignment\"", defaultSpelling)
	}
	for _, property := range staticPropertyPlacementProperties {
		options.Positions[property] = position
	}
	if len(slots) < 2 {
		return options, nil
	}

	var overrides map[string]string
	if err := json.Unmarshal(slots[1], &overrides); err != nil {
		return options, fmt.Errorf("static-property-placement element 2: %w", err)
	}
	for property, spelling := range overrides {
		if _, isKnown := options.Positions[property]; !isKnown {
			return options, fmt.Errorf("static-property-placement: %q is not a property this rule "+
				"judges, so its override would never be read", property)
		}
		overridePosition, recognised := staticPropertyPlacementPositionFor(spelling)
		if !recognised {
			return options, fmt.Errorf("static-property-placement: unknown position %q for %s",
				spelling, property)
		}
		options.Positions[property] = overridePosition
	}
	return options, nil
}

// staticPropertyPlacementPositionFor maps upstream's option spelling onto ours.
//
// Upstream's spellings carry spaces and are what a configuration file writes, so they are the wire
// format. The Go values are PascalCase per this codebase's convention for a string-literal union,
// which is why the two spellings differ and why this function exists rather than a direct cast.
func staticPropertyPlacementPositionFor(spelling string) (StaticPropertyPlacementPosition, bool) {
	switch spelling {
	case "static public field":
		return StaticPublicField, true
	case "static getter":
		return StaticGetter, true
	case "property assignment":
		return PropertyAssignment, true
	}
	return "", false
}

// messageStaticPropertyPlacement builds the finding for one misplaced property.
//
// The property name is interpolated, so this needs a rendered-text assertion: a message-id
// assertion cannot see anything the interpolation does, and all three arms share the same shape.
func messageStaticPropertyPlacement(
	position StaticPropertyPlacementPosition,
	propertyName string,
) rule.Message {
	switch position {
	case StaticGetter:
		return rule.Message{
			Id: "notGetterClassFunc",
			Description: "'" + propertyName + "' should be declared as a static getter class " +
				"function. This file's configuration puts React statics in the class body as " +
				"getters, and a reader looking for them will look there.",
		}
	case PropertyAssignment:
		return rule.Message{
			Id: "declareOutsideClass",
			Description: "'" + propertyName + "' should be declared outside the class body. " +
				"This file's configuration puts React statics after the class as assignments, " +
				"and a reader looking for them will look there.",
		}
	}
	return rule.Message{
		Id: "notStaticClassProp",
		Description: "'" + propertyName + "' should be declared as a static class property. " +
			"This file's configuration puts React statics in the class body as static fields, " +
			"and a reader looking for them will look there.",
	}
}

// StaticPropertyPlacement enforces where a React component's static properties are written.
//
//	valid:   class C extends React.Component { static propTypes = {}; }         under the default
//	valid:   C.propTypes = {}                                                   under "property assignment"
//	valid:   class C extends React.Component { static other = {}; }             not a React static
//	invalid: class C extends React.Component { propTypes = {}; }                not static
//	invalid: class C extends React.Component { static get propTypes() {} }      under the default
//	invalid: C.propTypes = {}                                                   under the default
//
// # The six properties, and why the order they are listed in matters
//
// `propTypes`, `defaultProps`, `childContextTypes`, `contextTypes`, `contextType`, `displayName`.
// Upstream finds the FIRST name whose predicate matches and reports under that name, and two of the
// predicates accept a second spelling: a class field named `props` carrying a type annotation counts
// as `propTypes`, and one named `context` carrying one counts as `contextTypes`. Keeping upstream's
// order keeps the reported name identical.
//
// # Three arms, and the asymmetry between them
//
// A class field is judged against `StaticPublicField`, a static getter against `StaticGetter`, and
// an assignment after the class against `PropertyAssignment`. Each arm asks "is the configured
// position the one I am", and additionally, for the two in-class positions, whether the member is
// actually `static`.
//
// The getter arm is narrower than it looks: upstream requires BOTH `static` and `kind === 'get'`
// before checking anything, so a NON-static getter is silent under every configuration. Measured
// against the installed build. That is a gap rather than a decision, and it is reproduced.
//
// # Which assignments count
//
// `C.propTypes = {}` counts when `C` resolves to a class extending a React base. Upstream walks the
// member path, resolves the head through scope, and then follows references and definitions to find
// the component; this asks the checker for the symbol and looks at its declarations, which answers
// the same question. Measured agreement on the shapes that separate them:
//
//	class C extends React.Component {}; C.propTypes = {}     reports
//	const C = class extends React.Component {}; C.propTypes = {}  reports
//	const ns = {C: class extends React.Component {}}; ns.C.propTypes = {}  reports
//	class C extends Other {}; C.propTypes = {}               SILENT, not a React class
//	function C(){}; C.propTypes = {}                         SILENT, not an ES6 class
//
// An assignment written INSIDE a class body is skipped entirely, which is upstream's
// `isContextInClass` walking the scope chain for an enclosing class. Measured silent.
//
// # No fix
//
// `meta.fixable` is null upstream. Moving a static between the three positions changes more than
// spelling: a class field and a getter differ in when the value is evaluated, and an assignment
// after the class is a different statement entirely. None of the three rewrites is behaviour
// preserving in general, so none is offered here either.
var StaticPropertyPlacement = rule.Rule{
	Name:             "react/static-property-placement",
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		settings := DefaultStaticPropertyPlacementOptions()
		if configured, isConfigured := rule.OptionsAs[StaticPropertyPlacementOptions](options); isConfigured &&
			configured.Positions != nil {
			settings = configured
		}

		// judge is upstream's `reportNodeIncorrectlyPositioned`. The node reports when the
		// configured position differs from the one this arm represents, OR when the configured
		// position is one of the two in-class ones and the member is not actually static.
		judge := func(node *ast.Node, propertyName string, arm StaticPropertyPlacementPosition, isStatic bool) {
			expected, isJudged := settings.Positions[propertyName]
			if !isJudged {
				return
			}
			mismatched := expected != arm
			notStaticEnough := !isStatic &&
				(expected == StaticPublicField || expected == StaticGetter)
			if !mismatched && !notStaticEnough {
				return
			}
			ctx.ReportNode(node, messageStaticPropertyPlacement(expected, propertyName))
		}

		return rule.Listeners{
			ast.KindPropertyDeclaration: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				if !staticPropertyPlacementInReactClass(node) {
					return
				}
				propertyName, named := staticPropertyPlacementNameOf(node)
				if !named {
					return
				}
				judge(node, propertyName, StaticPublicField, staticPropertyPlacementIsStatic(node))
			},

			ast.KindGetAccessor: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				// Upstream requires static AND a getter before it looks at the name at all, so a
				// non-static getter is silent under every configuration. Measured; reproduced.
				if !staticPropertyPlacementIsStatic(node) {
					return
				}
				if !staticPropertyPlacementInReactClass(node) {
					return
				}
				propertyName, named := staticPropertyPlacementNameOf(node)
				if !named {
					return
				}
				judge(node, propertyName, StaticGetter, true)
			},

			ast.KindBinaryExpression: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				expression := node.AsBinaryExpression()
				if expression.OperatorToken == nil ||
					expression.OperatorToken.Kind != ast.KindEqualsToken {
					return
				}
				target := expression.Left
				if target == nil || target.Kind != ast.KindPropertyAccessExpression {
					return
				}
				// An assignment written inside a class body is skipped, which is upstream's
				// `isContextInClass` walking the scope chain for an enclosing class declaration.
				if staticPropertyPlacementInsideAClass(node) {
					return
				}
				propertyName, named := staticPropertyPlacementNameOf(target)
				if !named {
					return
				}
				if !staticPropertyPlacementReceiverIsReactClass(ctx, target.AsPropertyAccessExpression().Expression) {
					return
				}
				// An assignment is never `static`, and upstream passes `node.static` as undefined
				// here, which is falsy. So a configuration expecting either in-class position
				// reports through the second half of the test rather than the first.
				judge(target, propertyName, PropertyAssignment, false)
			},
		}
	},
}

// staticPropertyPlacementNameOf answers which of the six names a node declares, if any.
//
// The two Flow-shaped spellings are checked first for a class field, matching upstream: a field
// named `props` carrying a type annotation is a `propTypes` declaration, and one named `context`
// carrying one is a `contextTypes` declaration. Upstream's predicates run in list order and each
// tests its Flow spelling before falling through to the plain name comparison, so a field named
// `props` answers `propTypes` and never reaches the later names.
func staticPropertyPlacementNameOf(node *ast.Node) (string, bool) {
	name := node.Name()
	if name == nil {
		return "", false
	}
	text, readable := ast.TryGetTextOfPropertyName(name)
	if !readable {
		return "", false
	}

	if node.Kind == ast.KindPropertyDeclaration && node.AsPropertyDeclaration().Type != nil {
		switch text {
		case "props":
			return "propTypes", true
		case "context":
			return "contextTypes", true
		}
	}

	for _, property := range staticPropertyPlacementProperties {
		if text == property {
			return property, true
		}
	}
	return "", false
}

// staticPropertyPlacementIsStatic reports whether a class member carries the static modifier.
func staticPropertyPlacementIsStatic(node *ast.Node) bool {
	modifiers := node.Modifiers()
	if modifiers == nil {
		return false
	}
	for _, modifier := range modifiers.Nodes {
		if modifier.Kind == ast.KindStaticKeyword {
			return true
		}
	}
	return false
}

// staticPropertyPlacementInReactClass reports whether a class member belongs to a React component
// class, which is upstream's `getParentES6Component`.
func staticPropertyPlacementInReactClass(member *ast.Node) bool {
	body := member.Parent
	if body == nil {
		return false
	}
	return isComponentClass(body)
}

// staticPropertyPlacementInsideAClass reports whether a node sits anywhere inside a class body.
//
// Upstream's `isContextInClass` walks the SCOPE chain looking for a block that is a class
// declaration, which reaches through nested functions, so an assignment inside a method of a class
// is skipped. Walking parents reaches the same answer for the same inputs and does not need a scope
// table.
func staticPropertyPlacementInsideAClass(node *ast.Node) bool {
	for current := node.Parent; current != nil; current = current.Parent {
		if current.Kind == ast.KindClassDeclaration || current.Kind == ast.KindClassExpression {
			return true
		}
	}
	return false
}

// staticPropertyPlacementReceiverIsReactClass answers whether the left side of an assignment names
// a React component class, which is upstream's `getRelatedComponent` plus `isES6Component`.
//
// Upstream walks the member path, resolves the head identifier through scope, then hunts through
// that variable's references and definitions for the component node. We ask the checker for the
// symbol at whichever node the path ends on and look at what it was declared as, which answers the
// same question for every shape measured:
//
//	class C extends React.Component {}                     a class declaration
//	const C = class extends React.Component {}             a variable whose initializer is a class
//	const ns = {C: class extends React.Component {}}       a property whose value is a class
//
// The loop over declarations rather than an index is the standing hazard in this tree: a merged
// declaration can put the class at a position other than zero.
func staticPropertyPlacementReceiverIsReactClass(ctx rule.Context, receiver *ast.Node) bool {
	if receiver == nil {
		return false
	}
	symbol := ctx.TypeChecker.GetSymbolAtLocation(receiver)
	if symbol == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if staticPropertyPlacementDeclaresReactClass(declaration) {
			return true
		}
	}
	return false
}

// staticPropertyPlacementDeclaresReactClass answers whether one declaration is, or holds, a React
// component class.
func staticPropertyPlacementDeclaresReactClass(declaration *ast.Node) bool {
	if declaration == nil {
		return false
	}
	if isComponentClass(declaration) {
		return true
	}
	switch declaration.Kind {
	case ast.KindVariableDeclaration:
		return isComponentClass(declaration.AsVariableDeclaration().Initializer)
	case ast.KindPropertyAssignment:
		return isComponentClass(declaration.AsPropertyAssignment().Initializer)
	}
	return false
}
