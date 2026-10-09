package core

import (
	"encoding/json"
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// messageMaxNestedCallbacksId is the id. The message itself is built per finding, because the
// rendered text carries both the depth reached and the limit, and neither is a constant a reader
// could compare a diagnostic against by identity.
const messageMaxNestedCallbacksId = "exceed"

// maxNestedCallbacksMessage renders the finding.
//
// Upstream's template is `Too many nested callbacks ({{num}}). Maximum allowed is {{max}}.` Both
// values move and both are integers, so passing them in the wrong order renders a sentence that
// reads perfectly and says the opposite thing. That is why the tests assert the rendered string
// rather than only the id.
func maxNestedCallbacksMessage(depth int, maximum int) rule.Message {
	return rule.Message{
		Id: messageMaxNestedCallbacksId,
		Description: fmt.Sprintf("Too many nested callbacks (%d). Maximum allowed is %d. "+
			"Each layer of callback moves the code that runs later further from the code that "+
			"decided to run it, so by this depth the order of execution can no longer be read off "+
			"the page. Name the inner functions and call them by name, or move to promises.",
			depth, maximum),
	}
}

// MaxNestedCallbacksOptions carries upstream's one positional option, which has two shapes.
//
// Upstream's schema is a `oneOf` over an integer and an object, so `[3]` and `[{max: 3}]` are the
// same configuration. Maximum is a pointer because 0 is both a legal limit here and the zero value
// of the type, and those two must stay distinguishable: `[0]` is a real configuration meaning no
// callback may be a call argument at all, and upstream's corpus tests it.
type MaxNestedCallbacksOptions struct {
	// Maximum is how deep a callback may be nested. Upstream defaults it to 10.
	Maximum *int

	// CheckConstructorCallCallbacks counts a function passed to `new Foo(...)` as a callback.
	// Off by default, so `new Promise(() => {})` is clean at any limit unless it is turned on.
	CheckConstructorCallCallbacks *bool
}

// DefaultMaxNestedCallbacksSettings is upstream's `defaultOptions: [10]`.
func DefaultMaxNestedCallbacksSettings() MaxNestedCallbacksOptions {
	maximum := 10
	checkConstructorCallCallbacks := false
	return MaxNestedCallbacksOptions{
		Maximum:                       &maximum,
		CheckConstructorCallCallbacks: &checkConstructorCallCallbacks,
	}
}

// maxNestedCallbacksObjectShape is the object arm of upstream's `oneOf`.
//
// Both spellings of the limit are carried because upstream accepts both and prefers `maximum`, and
// the preference is expressed in a way that has an observable defect. See the decoder.
type maxNestedCallbacksObjectShape struct {
	Maximum                       *int  `json:"maximum"`
	Max                           *int  `json:"max"`
	CheckConstructorCallCallbacks *bool `json:"checkConstructorCallCallbacks"`
}

// DecodeMaxNestedCallbacksOptions turns the configured value into options.
//
// Hand-rolled rather than `rule.DecodeOptionsInto` because the wire value is polymorphic, the
// default is not the zero value, and the two spellings of the limit interact in a way no struct tag
// can express.
//
// The value may be a bare INTEGER: cohere's config layer strips the severity tuple, so a config
// writing `["error", 3]` hands this the JSON `3`, which no struct can unmarshal.
//
// An ABSENT value must mean 10 rather than 0. A rule configured as a bare `"error"` reaches the
// decoder with empty input, and a plain int field would read that as a limit of zero, reporting
// every callback in the tree. That is the live config's own shape, so it is the ordinary path.
//
// An object with NEITHER key must also mean 10. Upstream reaches the object arm only when
// `Object.hasOwn(option, "maximum") || Object.hasOwn(option, "max")`, so `{}` and
// `{checkConstructorCallCallbacks: true}` both leave the limit at ten. Measured against the
// installed build at 10.8.1: `[{}]` is clean on a two-deep nest that `[{max: 1}]` reports.
//
// # An upstream defect this reproduces
//
// Upstream resolves the two spellings with `option.maximum || option.max`, which is a truthiness
// test rather than a presence test, so a `maximum` of ZERO falls through to `max`. Measured:
//
//	[{maximum: 0, max: 5}]   clean on a one-deep callback, so the limit in force is 5
//	[{maximum: 5, max: 1}]   clean on a two-deep nest, so the limit in force is 5
//	[{maximum: 1, max: 5}]   reports on a two-deep nest, so the limit in force is 1
//
// A user who writes `{maximum: 0, max: 5}` asking for zero silently gets five. Reproduced rather
// than corrected, because a port that quietly disagrees with the tool it is compared against turns
// a shared defect into a divergence nobody can see. The intuitive reading is that an explicit 0
// wins, and it is wrong.
func DecodeMaxNestedCallbacksOptions(raw []byte) (any, error) {
	if len(raw) == 0 {
		return DefaultMaxNestedCallbacksSettings(), nil
	}

	settings := DefaultMaxNestedCallbacksSettings()

	// The integer arm first, because it is the narrower shape: a JSON number cannot also decode as
	// an object, so trying it first cannot swallow a case belonging to the other arm.
	var configuredMaximum int
	if err := json.Unmarshal(raw, &configuredMaximum); err == nil {
		if configuredMaximum < 0 {
			return DefaultMaxNestedCallbacksSettings(), fmt.Errorf(
				"max-nested-callbacks takes a maximum of at least 0, got %d", configuredMaximum)
		}
		settings.Maximum = &configuredMaximum
		return settings, nil
	}

	var object maxNestedCallbacksObjectShape
	if err := rule.UnmarshalOptions(raw, &object); err != nil {
		return DefaultMaxNestedCallbacksSettings(), err
	}
	// `option.maximum || option.max`, truthiness and all. A zero `maximum` is falsy in JavaScript
	// and falls through to `max`; only when both are absent or both are zero does the default stand.
	var resolved *int
	if object.Maximum != nil && *object.Maximum != 0 {
		resolved = object.Maximum
	} else if object.Max != nil && *object.Max != 0 {
		resolved = object.Max
	} else if object.Maximum != nil {
		resolved = object.Maximum
	} else if object.Max != nil {
		resolved = object.Max
	}
	if resolved != nil {
		if *resolved < 0 {
			return DefaultMaxNestedCallbacksSettings(), fmt.Errorf(
				"max-nested-callbacks takes a maximum of at least 0, got %d", *resolved)
		}
		settings.Maximum = resolved
	}
	if object.CheckConstructorCallCallbacks != nil {
		settings.CheckConstructorCallCallbacks = object.CheckConstructorCallCallbacks
	}
	return settings, nil
}

// MaxNestedCallbacks flags a function passed as a call argument nested deeper than the limit.
//
//	valid:   foo(function() { bar(thing, function(data) {}); });        under a maximum of 3
//	valid:   fn(function(){}, function(){}, function(){});              siblings do not accumulate
//	valid:   (function() {})();                                         a callee is not a callback
//	valid:   new Promise(() => {});                                     unless the option is on
//	invalid: foo(function() { bar(thing, function(data) { baz(function() {}); }); });
//
// # What counts, and what the depth is counted over
//
// Only a function that is an ARGUMENT to a call counts. A function that is the callee does not, so
// an immediately invoked expression is clean at a limit of zero, and neither does a function stored
// in a variable, held in an object property, or used as a method body. Those all appear in the
// corpus and each one is a class of false positive a looser reading would ship.
//
// The count is the depth of the callback CHAIN rather than the number of callbacks in the file.
// Upstream pushes on entering such a function and pops on leaving it, so two callbacks side by side
// at the same level each reach depth one. Three of upstream's invalid cases exist only to pin this:
// each writes a function declaration or an unrelated callback beside the nest and asserts a single
// finding rather than several.
//
// # Why this walks the tree itself
//
// Upstream keeps its stack with paired enter and exit listeners. This tree has no exit hook and the
// walk is pre-order, so a listener on the function kinds could push and never pop, and the depth
// would grow monotonically over the file: the second sibling callback would report at a depth the
// first one earned. So the whole judgment happens under one KindSourceFile listener, which fires
// before its children, and the recursion carries the depth down and unwinds it on return. That is
// the shipped shape for a rule that has to gather before it can judge.
//
// # Where a finding points
//
// At the function HEAD rather than the whole function, which upstream gets from
// `astUtils.getFunctionHeadLoc`. That helper has three branches and only two are reachable here:
//
//	an arrow                     the `=>` token alone, so `async () => {}` reports on the arrow
//	                             and not on the `async`
//	anything else                the function's own start through to the opening paren of its
//	                             parameters, so `async function*` and `function named` are both
//	                             covered whole while the parameter list is not
//	a Property or method parent   UNREACHABLE from this rule
//
// The third branch is the one worth writing down. A function whose parent is an object property or
// a class method is never a call argument, so `checkFunction` returns before it can report, and no
// input can reach that branch through this rule. Measured against the installed build rather than
// argued: `foo(function(){ bar({m: function(){}}) })` and the shorthand `bar({m(){}})` are both
// clean at a limit of 1 while the same shape with a bare function argument reports. The callers
// enumerated for that verdict are the two listeners upstream registers and nothing else; a future
// change that counted method bodies would void it.
var MaxNestedCallbacks = rule.Rule{
	Name: "max-nested-callbacks",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// A rule configured as a bare severity is handed nil, and the zero value of the struct
		// carries a nil Maximum, so the default has to be restored somewhere.
		//
		// It is restored in BOTH places on purpose, and the two are mutually redundant: mutating
		// either one alone leaves every fixture green because the other still supplies the ten,
		// and mutating both at once fails four. That pair is invisible to a sweep that changes one
		// site at a time, so it is recorded here rather than left looking like two guards where one
		// would do. The redundancy is cheap and the failure it prevents is a limit of zero, which
		// reports every callback in the tree.
		settings, isSettings := rule.OptionsAs[MaxNestedCallbacksOptions](options)
		if !isSettings {
			settings = DefaultMaxNestedCallbacksSettings()
		}
		maximum := 10
		if settings.Maximum != nil {
			maximum = *settings.Maximum
		}
		checkConstructorCallCallbacks := settings.CheckConstructorCallCallbacks != nil &&
			*settings.CheckConstructorCallCallbacks

		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				var walk func(current *ast.Node, depth int)
				walk = func(current *ast.Node, depth int) {
					next := depth
					if maxNestedCallbacksIsCountedCallback(current, checkConstructorCallCallbacks) {
						next = depth + 1
						if next > maximum {
							ctx.ReportRange(maxNestedCallbacksHeadRange(ctx, current),
								maxNestedCallbacksMessage(next, maximum))
						}
					}
					current.ForEachChild(func(child *ast.Node) bool {
						walk(child, next)
						return false
					})
				}
				walk(node, 0)
			},
		}
	},
}

// maxNestedCallbacksIsCountedCallback answers whether this node is a function being passed as an
// argument to a call, which is the only thing the depth counts.
//
// The callee test is upstream's `parent.callee === node` and it is not redundant with the argument
// test. An immediately invoked function is the CHILD of a call expression, so a check that only
// asked about the parent's kind would count `(function() {})()` and report it at a limit of zero,
// which upstream's corpus writes as a passing case twice.
func maxNestedCallbacksIsCountedCallback(node *ast.Node, checkConstructorCallCallbacks bool) bool {
	if node.Kind != ast.KindFunctionExpression && node.Kind != ast.KindArrowFunction {
		return false
	}
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindCallExpression:
		return parent.AsCallExpression().Expression != node
	case ast.KindNewExpression:
		if !checkConstructorCallCallbacks {
			return false
		}
		return parent.AsNewExpression().Expression != node
	}
	return false
}

// maxNestedCallbacksHeadRange is the reachable half of upstream's `getFunctionHeadLoc`.
//
// For an arrow the head is the `=>` token by itself. For anything else it runs from the function's
// own first token through to the opening paren of its parameters, which places `async` and a
// generator star and a name inside the span and leaves the parameters outside it. The paren is
// found at one before the parameter list's start rather than by scanning, because the list's
// position is inside the paren by construction. Verified across a named function, an async one, a
// generator, an async generator, one carrying type parameters, one with a comment before its
// parens, and one broken across three lines.
func maxNestedCallbacksHeadRange(ctx rule.Context, node *ast.Node) core.TextRange {
	start := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, node.Pos()).Pos()

	if node.Kind == ast.KindArrowFunction {
		arrow := node.AsArrowFunction().EqualsGreaterThanToken
		if arrow == nil {
			return core.NewTextRange(start, node.End())
		}
		return core.NewTextRange(
			scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, arrow.Pos()).Pos(), arrow.End())
	}

	parameters := node.AsFunctionExpression().Parameters
	if parameters == nil || parameters.Pos() <= start {
		return core.NewTextRange(start, node.End())
	}
	return core.NewTextRange(start, parameters.Pos()-1)
}
