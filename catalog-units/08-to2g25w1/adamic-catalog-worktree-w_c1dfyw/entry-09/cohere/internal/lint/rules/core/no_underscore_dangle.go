package core

import (
	"fmt"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoUnderscoreDangleOptions carries upstream's nine options.
//
// # THREE of the nine default to TRUE, and the Go zero value is wrong for all three
//
// A first draft of this rule assumed every default was the zero value, wrote that down as a reasoned
// observation, and was wrong about `allowFunctionParams`, `allowInArrayDestructuring` and
// `allowInObjectDestructuring`. Upstream's `defaultOptions` block settles it:
//
//	allow: []                      allowAfterSuper: false        allowAfterThis: false
//	allowAfterThisConstructor: false
//	allowFunctionParams: TRUE      allowInArrayDestructuring: TRUE
//	allowInObjectDestructuring: TRUE
//	enforceInClassFields: false    enforceInMethodNames: false
//
// So out of the box `function foo(_bar) {}`, `const [_foo] = arr` and `const { _foo } = obj` are all
// CLEAN, and twelve of upstream's valid cases assert exactly that. The corpus caught this; nothing
// in the source's shape suggests it.
//
// The three are pointers rather than bools for that reason. A `*bool` distinguishes "absent, so use
// upstream's default" from "explicitly false", which a plain bool cannot: `{"allowFunctionParams":
// false}` and an absent key are different configurations and upstream's corpus asserts both.
type NoUnderscoreDangleOptions struct {
	// Allow lists identifiers that are permitted whole, e.g. `["__proto__"]`.
	Allow []string `json:"allow"`
	// AllowAfterThis permits `this._foo`.
	AllowAfterThis bool `json:"allowAfterThis"`
	// AllowAfterSuper permits `super._foo`.
	AllowAfterSuper bool `json:"allowAfterSuper"`
	// AllowAfterThisConstructor permits `this.constructor._foo`.
	AllowAfterThisConstructor bool `json:"allowAfterThisConstructor"`
	// EnforceInMethodNames turns the method-name arm ON. Off by default.
	EnforceInMethodNames bool `json:"enforceInMethodNames"`
	// AllowFunctionParams permits an underscored parameter name. Defaults TRUE.
	AllowFunctionParams *bool `json:"allowFunctionParams"`
	// EnforceInClassFields turns the class-field arm ON. Off by default.
	EnforceInClassFields bool `json:"enforceInClassFields"`
	// AllowInArrayDestructuring permits a name coined by an array pattern. Defaults TRUE.
	AllowInArrayDestructuring *bool `json:"allowInArrayDestructuring"`
	// AllowInObjectDestructuring permits a name coined by an object pattern. Defaults TRUE.
	AllowInObjectDestructuring *bool `json:"allowInObjectDestructuring"`
}

// DecodeNoUnderscoreDangleOptions reads the option object off the config.
//
// # The config layer hands this a different shape than a fixture does
//
// `internal/lint/configuration/configuration.go` keeps `setting.Options = tuple[1]`, a single JSON
// value after the severity, while ESLint's `context.options` is every element after it. This rule's
// `meta.schema` declares ONE element, so the two coincide and the object arrives whole. Written out
// rather than assumed: a decoder wrong for the config shape passes every fixture, because a fixture
// hands it bytes the test built rather than bytes the config layer sliced.
//
// `additionalProperties: false` upstream means an unknown key is refused at config load. That
// refusal is reproduced here, and it matters more for this rule than for most: there are nine keys,
// four of them start with `allow` and two with `enforce`, and `allowInMethodNames` is a plausible
// thing to type for `enforceInMethodNames`. Silently ignoring it would leave a project believing it
// had turned an arm off when it had never been on.
func DecodeNoUnderscoreDangleOptions(raw []byte) (any, error) {
	var options NoUnderscoreDangleOptions
	if len(raw) == 0 {
		return options, nil
	}
	if err := rule.UnmarshalOptions(raw, &options); err != nil {
		return options, fmt.Errorf("no-underscore-dangle: %w", err)
	}
	return options, nil
}

// noUnderscoreDangleFlag reads a tri-state option, applying upstream's default when it is absent.
func noUnderscoreDangleFlag(configured *bool, defaultValue bool) bool {
	if configured == nil {
		return defaultValue
	}
	return *configured
}

var messageNoUnderscoreDangle = rule.Message{
	Id: "unexpectedUnderscore",
	Description: "A leading or trailing underscore is a convention for privacy that the language " +
		"does not enforce, so it tells a reader something is off limits while doing nothing to " +
		"keep them out. Where privacy is real, `#private` or a closure says so and is checked.",
}

// NoUnderscoreDangle flags an identifier that begins or ends with an underscore.
//
//	valid:   var foo = 1
//	valid:   var _ = require('underscore')
//	valid:   foo.bar.__proto__
//	valid:   class foo { _method() {} }              enforceInMethodNames defaults OFF
//	invalid: var _foo = 1
//	invalid: function foo_() {}
//	invalid: this._prop
//
// # Six arms, and two of them are off by default
//
// Upstream registers eight listeners over five judgments: a function's own name, a function's
// parameters, a variable declarator's bindings, a member access, a method name, and a class field.
// The last two are gated behind `enforceInMethodNames` and `enforceInClassFields`, both defaulting
// FALSE, so out of the box this rule says nothing about `class A { _method() {} }`. That is worth
// stating because the audit's violation count was taken with the defaults, and turning either flag
// on is a different rule.
//
// # Three names are special-cased and they are not the same three
//
//	`_` alone            never reported anywhere. `identifier !== "_"` sits inside the dangling
//	                     test itself, so it protects the underscore library's `var _ = require(...)`
//	                     in every arm at once.
//	`__proto__`          exempt in a MEMBER ACCESS only. `var __proto__ = 1` reports, and upstream's
//	                     corpus asserts both directions.
//	`#name`              a private identifier renders with its hash in the message, in the method and
//	                     class-field arms. Measured: `#_bar` reports as "'#_bar'".
//
// # The destructuring arms ask which pattern COINED the name, not which pattern contains it
//
// Upstream walks up from the identifier to the first `VariableDeclarator`, `ArrayPattern` or
// `ObjectPattern` and tests THAT node's type. So in
// `const { foo: [_bar, { a: _a }] } = ...` the name `_bar` is coined by an array pattern and `_a` by
// an object one, in the same declarator, and the two flags exempt them independently. Upstream's
// corpus asserts exactly that with `allowInArrayDestructuring: true, allowInObjectDestructuring:
// false`, which reports `_a` and not `_bar`.
//
// A port testing "is this declarator's binding a pattern at all" collapses the two flags into one
// and passes most of the corpus. The nested case is what separates them.
//
// # The span is the DECLARATOR, not the identifier, and several findings can share it
//
// Upstream reports `node` from the `VariableDeclarator` listener, so every underscored name a
// declarator coins reports at the same span covering the whole declaration. Measured:
// `const { foo: [_bar, { a: _a, b } ] } = ...` produces TWO findings, both at columns 7 to 72, whose
// only difference is the identifier in the message. A port asserting one finding per span, or
// anchoring on the identifier, is wrong on both counts and a message-id fixture cannot see it.
//
// The other arms differ and each was measured: a function declaration reports the whole function, a
// parameter reports the parameter, a member access reports the whole access, and a method or field
// reports the whole member.
//
// # No fixer
//
// `meta.fixable` is unset and the rule is `frozen`. Renaming an identifier is a refactor across
// every reference, which a span-local repair cannot do.
var NoUnderscoreDangle = rule.Rule{
	Name: "no-underscore-dangle",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, _ := rule.OptionsAs[NoUnderscoreDangleOptions](options)

		reportNamed := func(node *ast.Node, identifier string) {
			ctx.ReportNode(node, rule.Message{
				Id: messageNoUnderscoreDangle.Id,
				Description: fmt.Sprintf("Unexpected dangling '_' in '%s'. %s",
					identifier, messageNoUnderscoreDangle.Description),
			})
		}

		checkFunction := func(node *ast.Node) {
			// Upstream checks the NAME only for a FunctionDeclaration. A named function expression
			// or an arrow assigned to a variable is not checked here; the variable arm catches the
			// binding instead.
			if node.Kind == ast.KindFunctionDeclaration {
				if name := node.Name(); name != nil && name.Kind == ast.KindIdentifier {
					if noUnderscoreDangleReportable(name.Text(), settings) {
						reportNamed(node, name.Text())
					}
				}
			}
			if noUnderscoreDangleFlag(settings.AllowFunctionParams, true) {
				return
			}
			functionLike := node.FunctionLikeData()
			if functionLike == nil || functionLike.Parameters == nil {
				return
			}
			for _, parameter := range functionLike.Parameters.Nodes {
				// Upstream unwraps a RestElement to its argument and an AssignmentPattern to its
				// left, then requires an Identifier. Our parser puts the rest token and the
				// initializer on the parameter itself, so the same three shapes are one node whose
				// Name() is the binding, and a destructured parameter's name is a binding pattern
				// rather than an identifier and is declined by the kind guard.
				name := parameter.Name()
				if name == nil || name.Kind != ast.KindIdentifier {
					continue
				}
				if noUnderscoreDangleReportable(name.Text(), settings) {
					// The PARAMETER is reported, not its name. Measured at columns 14 to 22 for
					// `function foo(_bar = 0) {}`, which covers the default too.
					reportNamed(parameter, name.Text())
				}
			}
		}

		checkMember := func(node *ast.Node, expression *ast.Node, name *ast.Node) {
			if name == nil || name.Kind != ast.KindIdentifier {
				return
			}
			identifier := name.Text()
			if !noUnderscoreDangleHasDanglingUnderscore(identifier) {
				return
			}
			if expression != nil {
				if expression.Kind == ast.KindThisKeyword && settings.AllowAfterThis {
					return
				}
				if expression.Kind == ast.KindSuperKeyword && settings.AllowAfterSuper {
					return
				}
				if settings.AllowAfterThisConstructor &&
					noUnderscoreDangleIsThisConstructor(expression) {
					return
				}
			}
			// `__proto__` is exempt in a member access and nowhere else.
			if identifier == "__proto__" {
				return
			}
			if slices.Contains(settings.Allow, identifier) {
				return
			}
			reportNamed(node, identifier)
		}

		checkMemberName := func(node *ast.Node, gated bool) {
			if !gated {
				return
			}
			name := node.Name()
			if name == nil {
				return
			}
			// A private identifier's Text() already carries the leading hash here, which is what
			// upstream builds by hand with a template literal. The dangling test runs on the name
			// WITHOUT the hash, so `#_bar` is dangling and `#bar` is not.
			identifier := name.Text()
			bare := identifier
			if name.Kind == ast.KindPrivateIdentifier {
				bare = identifier[1:]
			} else if name.Kind != ast.KindIdentifier {
				// A computed or literal key is not an identifier upstream either: it reads
				// `node.key.name`, which is undefined for both, and the `typeof !== "undefined"`
				// guard declines.
				return
			}
			if !noUnderscoreDangleHasDanglingUnderscore(bare) {
				return
			}
			if slices.Contains(settings.Allow, bare) {
				return
			}
			reportNamed(node, identifier)
		}

		return rule.Listeners{
			ast.KindFunctionDeclaration: checkFunction,
			ast.KindFunctionExpression:  checkFunction,
			ast.KindArrowFunction:       checkFunction,

			ast.KindVariableDeclaration: func(node *ast.Node) {
				declaration := node.AsVariableDeclaration()
				noUnderscoreDangleWalkBindings(declaration.Name(), declaration.Name(),
					func(name string, coinedBy ast.Kind) {
						if !noUnderscoreDangleReportable(name, settings) {
							return
						}
						if coinedBy == ast.KindArrayBindingPattern &&
							noUnderscoreDangleFlag(settings.AllowInArrayDestructuring, true) {
							return
						}
						if coinedBy == ast.KindObjectBindingPattern &&
							noUnderscoreDangleFlag(settings.AllowInObjectDestructuring, true) {
							return
						}
						// The DECLARATOR is reported, so several findings share one span.
						reportNamed(node, name)
					})
			},

			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				access := node.AsPropertyAccessExpression()
				checkMember(node, access.Expression, access.Name())
			},

			// Upstream's MethodDefinition and its `Property` with `method: true`. Our parser gives
			// an object literal method its own kind, so both arrive here.
			ast.KindMethodDeclaration: func(node *ast.Node) {
				checkMemberName(node, settings.EnforceInMethodNames)
				// A method's PARAMETERS are checked whatever `enforceInMethodNames` says, because
				// upstream reaches them through its FunctionExpression listener: in ESTree a
				// method's value IS a FunctionExpression. Our parser has no such wrapper, so
				// without this line `const o = { onClick(_bar) {} }` is silently unchecked. Three
				// of upstream's reporting cases are exactly that shape.
				checkFunction(node)
			},
			ast.KindPropertyDeclaration: func(node *ast.Node) {
				checkMemberName(node, settings.EnforceInClassFields)
			},
		}
	},
}

// noUnderscoreDangleReportable is the whole-name test the non-member arms share.
func noUnderscoreDangleReportable(identifier string, settings NoUnderscoreDangleOptions) bool {
	return noUnderscoreDangleHasDanglingUnderscore(identifier) &&
		!slices.Contains(settings.Allow, identifier)
}

// noUnderscoreDangleHasDanglingUnderscore is upstream's `hasDanglingUnderscore`.
//
// The `identifier !== "_"` clause lives INSIDE this test rather than beside one arm, which is what
// makes `var _ = require('underscore')` clean everywhere at once. A port hoisting that exemption
// into the variable arm alone passes the corpus, whose only `_` case is a variable.
func noUnderscoreDangleHasDanglingUnderscore(identifier string) bool {
	if identifier == "" || identifier == "_" {
		return false
	}
	return identifier[0] == '_' || identifier[len(identifier)-1] == '_'
}

// noUnderscoreDangleIsThisConstructor is upstream's `isThisConstructorReference`.
//
// The receiver of the access must itself be `this.constructor`, so this reads two levels down.
func noUnderscoreDangleIsThisConstructor(expression *ast.Node) bool {
	if expression.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	inner := expression.AsPropertyAccessExpression()
	if inner.Name() == nil || inner.Name().Text() != "constructor" {
		return false
	}
	return inner.Expression != nil && inner.Expression.Kind == ast.KindThisKeyword
}

// noUnderscoreDangleWalkBindings visits every name a declarator coins, with the pattern that coined
// it.
//
// `coinedBy` is upstream's walk up to the nearest `VariableDeclarator`, `ArrayPattern` or
// `ObjectPattern`, expressed as a walk DOWN instead: the enclosing pattern is known on the way in,
// so no parent chain is needed. A plain `const _foo = 1` has no pattern at all and reports under
// both flags, which upstream gets from landing on the VariableDeclarator.
//
// The recursion is what the nested case needs. `const { foo: [_bar, { a: _a }] }` coins `_bar` under
// an ARRAY pattern and `_a` under an OBJECT one, and the two flags exempt them independently.
func noUnderscoreDangleWalkBindings(node *ast.Node, root *ast.Node, visit func(string, ast.Kind)) {
	if node == nil {
		return
	}
	switch node.Kind {
	case ast.KindIdentifier:
		coinedBy := ast.KindUnknown
		if node.Parent != nil && node != root {
			if element := node.Parent; element.Kind == ast.KindBindingElement && element.Parent != nil {
				coinedBy = element.Parent.Kind
			}
		}
		visit(node.Text(), coinedBy)
	case ast.KindObjectBindingPattern:
		for _, element := range node.AsBindingPattern().Elements.Nodes {
			noUnderscoreDangleWalkBindings(element.AsBindingElement().Name(), root, visit)
		}
	case ast.KindArrayBindingPattern:
		for _, element := range node.AsBindingPattern().Elements.Nodes {
			// An elision in `const [, a] = x` is an OmittedExpression with no name.
			if element.Kind != ast.KindBindingElement {
				continue
			}
			noUnderscoreDangleWalkBindings(element.AsBindingElement().Name(), root, visit)
		}
	}
}
