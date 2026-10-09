package core

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/property"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// buildNoAlertMessage renders the finding, which names the function being called.
//
// Built per report rather than held as a constant, because the name is interpolated and
// `rule.Message` has no rendering layer, so an id assertion cannot see anything this format string
// does.
func buildNoAlertMessage(name string) rule.Message {
	return rule.Message{
		Id: "unexpected",
		Description: fmt.Sprintf(
			"This calls `%s`, which freezes the whole page on a dialog the user cannot style, "+
				"move, or dismiss with anything but the one button the browser drew. It is almost "+
				"always debugging left in, and where it is deliberate it belongs in the "+
				"application's own interface rather than the browser's.", name),
	}
}

// noAlertProhibited are the three functions upstream watches.
//
// Upstream spells this as `/^(?:alert|confirm|prompt)$/u`, which is a set membership written as a
// regular expression. A set says the same thing and cannot be widened by an accidental missing
// anchor, which is the failure mode of the regex spelling.
var noAlertProhibited = map[string]bool{
	"alert":   true,
	"confirm": true,
	"prompt":  true,
}

// NoAlert flags a call to `alert`, `confirm` or `prompt`, bare or through the global object.
//
//	valid:   foo.alert(foo)
//	valid:   function alert() {} alert();
//	valid:   window[alert]();
//	valid:   function foo() { var window = bar; window.alert(); }
//	invalid: alert(foo)
//	invalid: window.alert(foo)
//	invalid: window['prompt'](foo)
//	invalid: globalThis.alert();
//
// # Two call shapes, and a shadow test on each
//
// A bare `alert(foo)` reports unless something in source declares the name. A member call reports
// only when the object is the global object, spelled `window` or `globalThis`, and only when THAT
// name is itself unshadowed. So `function foo() { var window = bar; window.alert(); }` is clean
// while the same call at file scope reports, and upstream ships both halves as one case.
//
// The shadow question is the same one `no-undef-init` asks and it is answered the same way: does the
// identifier resolve to a symbol something in a source file declares. `undefinedIsShadowed` is that
// predicate with the name baked in; this needs it for three different names, so the shared form
// lives in `identifierIsShadowed` and the older function now calls it.
//
// # `window[alert]()` is clean and `window['alert']()` reports
//
// The first is a computed access through a VARIABLE named `alert`, so the property read is whatever
// that variable holds. The second is a string subscript, which is the same property as the dotted
// form. `property.AccessedName` already draws exactly this line and is what `no-iterator` uses, so
// the distinction is inherited rather than re-derived.
//
// # `this.alert()` is decided ABOVE this rule, by source type
//
// Upstream reports `this.alert(foo)` at the top level of a SCRIPT, where `this` is the global
// object, and its `isGlobalThisReferenceOrGlobalWindow` has an arm for exactly that. Our harness and
// our tree are modules: `internal/rule_testing/program.go` pins `moduleDetection: "force"`, and in a
// module top-level `this` is `undefined` rather than the global.
//
// Measured rather than reasoned: driving the installed rule with `sourceType: "module"` returns
// CLEAN for both `this.alert(foo)` and `this['alert'](foo)`, and `unexpected` for both under
// `sourceType: "script"`. So the two `this` cases in upstream's corpus are clean HERE for the same
// reason they are clean in any module, and reproducing them as reporting would require a script-mode
// file this tree cannot produce. There is deliberately no `this` arm below, and the two cases sit in
// the silent list with this reasoning at the line.
var NoAlert = rule.Rule{
	Name:             "no-alert",
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindCallExpression: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				callee := ast.SkipParentheses(node.AsCallExpression().Expression)
				if callee == nil {
					return
				}

				// The bare form. A local binding of the name means the call is not the global one,
				// whatever it is spelled.
				if callee.Kind == ast.KindIdentifier {
					name := callee.Text()
					if noAlertProhibited[name] && !identifierIsShadowed(ctx, callee) {
						ctx.ReportNode(node, buildNoAlertMessage(name))
					}
					return
				}

				// The member form, which needs the receiver to BE the global object before the
				// property name matters at all.
				//
				// A callee that is neither an identifier nor a member access has no receiver to ask
				// about, and `memberAccessObject` answers nil for it. `super(...)` and `import(...)`
				// are the two that occur constantly: both are call expressions whose callee is a bare
				// keyword, and every subclass constructor in the tree carries a `super()`. Feeding
				// that nil to `ast.SkipParentheses` dereferences it, and because the walk recovers
				// per file rather than per rule, the panic took the WHOLE file away from every rule.
				// It cost 167 files, about five percent of the tree, silently unchecked while the run
				// still printed green.
				object := memberAccessObject(callee)
				if object == nil {
					return
				}
				object = ast.SkipParentheses(object)
				if !isGlobalObjectReference(ctx, object) {
					return
				}

				// `property.Static` answers for the dotted form and for a string subscript alike,
				// and answers nothing for a computed access through a variable, which is what makes
				// `window[alert]()` a clean case rather than a missed finding.
				name, _ := property.AccessedName(callee, property.Static)
				if noAlertProhibited[name] {
					ctx.ReportNode(node, buildNoAlertMessage(name))
				}
			},
		}
	},
}

// isGlobalObjectReference says whether an expression names the global object.
//
// Upstream accepts `window` unconditionally and `globalThis` only when the scope knows a binding by
// that name, which is its way of asking whether the environment is new enough to have it. Ours does
// not need that test: `globalThis` is in the standard library our checker reads, so an unshadowed
// `globalThis` resolves to the global in every file, and a shadowed one resolves to the shadow. The
// two spellings therefore answer through one path here where upstream needs two.
//
// The `this` arm upstream carries is deliberately absent; see the rule's doc comment for the
// measurement.
func isGlobalObjectReference(ctx rule.Context, node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindIdentifier {
		return false
	}
	name := node.Text()
	if name != "window" && name != "globalThis" {
		return false
	}
	return !identifierIsShadowed(ctx, node)
}
