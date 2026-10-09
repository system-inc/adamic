package structure

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/react"
	"github.com/system-inc/cohere/internal/lint/ecmascript/scope"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// reactHookNoPropertiesInDependenciesText is the rule's message, whose wording lives in
// `policy/messages/react-hook-no-properties-in-dependencies.json`.
var reactHookNoPropertiesInDependenciesText = policy.MessageOf("structure/react-hook-no-properties-in-dependencies", "extractPropertiesFirst")

// messageExtractPropertiesFirst is the finding, rendered when it is reported so the text comes from
// the current catalog.
func messageExtractPropertiesFirst() rule.Message {
	return rule.Message{Id: reactHookNoPropertiesInDependenciesText.Id, Description: reactHookNoPropertiesInDependenciesText.Render(nil)}
}

// hooksWithDependencies are the React hooks that take a dependency array.
//
// useImperativeHandle is the one whose array is not the second argument, and that difference is
// carried at the call site rather than here.
var hooksWithDependencies = map[string]bool{
	"useCallback":         true,
	"useMemo":             true,
	"useEffect":           true,
	"useLayoutEffect":     true,
	"useImperativeHandle": true,
}

// ReactHookNoPropertiesInDependencies flags a whole `properties` object in a hook dependency array.
//
//	valid:   React.useEffect(() => { run(properties.id); }, [propertiesId])
//	valid:   React.useEffect(() => { run(properties.id); }, [properties.id])
//	invalid: React.useEffect(() => { run(properties.id); }, [properties])
//
// Ported from `structure/react-hook-no-properties-in-dependencies`.
//
// # Only `React.useEffect`, never a bare `useEffect`
//
// The original matches its callee through `isReactMethodCall`, which requires the callee to be a
// member expression whose object is the identifier `React`. A bare `useEffect(...)` imported
// directly is a plain identifier callee and never matches, so it is invisible to the rule.
//
// That is a real narrowing rather than an oversight to correct, and it costs nothing here: measured
// on the tree, 435 `React.<hook>` call sites and 0 bare ones, because this codebase imports React as
// a namespace. Widening it would report call sites the gate does not, which is the failure mode this
// port is judged against.
//
// # The dependency array is the third argument for one hook
//
// `useImperativeHandle(ref, factory, dependencies)` puts its array third; every other hook here puts
// it second. Reading the second argument for all five would check a factory function for
// `useImperativeHandle` and find nothing, so the rule would be quietly dead on that hook alone.
//
// # Where the report lands
//
// On the offending element inside the array rather than on the call, matching the original. An array
// listing `properties` twice reports twice, which is what the original does, and anchoring on the
// call would collapse them into one finding at a line the author has to search.
var ReactHookNoPropertiesInDependencies = rule.Rule{
	Name: "structure/react-hook-no-properties-in-dependencies",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil || !FileContextFor(ctx.SourceFile.FileName().AsString()).IsReactFile {
			return nil
		}

		return rule.Listeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				if call == nil || call.Expression == nil {
					return
				}

				hookName := reactHookNameOf(call.Expression)
				if hookName == "" || !hooksWithDependencies[hookName] {
					return
				}

				dependencies := dependencyArrayArgument(call, hookName)
				if dependencies == nil {
					return
				}

				// The containment test is deliberately made once per call rather than once per
				// element: it walks the ancestor chain, and the answer cannot differ between two
				// elements of the same array.
				if !isInsideComponent(node) {
					return
				}

				for _, element := range dependencies.AsArrayLiteralExpression().Elements.Nodes {
					// A bare identifier only. `properties.id` is the shape the rule is asking for
					// and must stay silent, and a spread or a call is not the whole object either.
					if element == nil || element.Kind != ast.KindIdentifier {
						continue
					}
					if element.Text() != "properties" {
						continue
					}
					ctx.ReportNode(element, messageExtractPropertiesFirst())
				}
			},
		}
	},
}

// reactHookNameOf returns the hook name in a `React.useThing` callee, or "" for anything else.
func reactHookNameOf(callee *ast.Node) string {
	// `react.IsNamespacedMember` rather than a hand-rolled receiver test, which additionally skips
	// parentheses on the receiver: `(React).useEffect(...)` is the same call and the hand-rolled
	// version declined it while looking correct. The predicate captures the name rather than
	// answering about it, since this function returns the name it found.
	found := ""
	react.IsNamespacedMember(callee, func(name string) bool {
		found = name
		return true
	})
	return found
}

// dependencyArrayArgument returns the hook's dependency array, or nil when it has none or when the
// argument in that position is something other than an array literal.
//
// A hook called with a variable rather than a literal (`React.useEffect(run, dependencies)`) has
// nothing to inspect, and guessing what the variable holds is not this rule's business.
func dependencyArrayArgument(call *ast.CallExpression, hookName string) *ast.Node {
	index := 1
	if hookName == "useImperativeHandle" {
		index = 2
	}
	if call.Arguments == nil || len(call.Arguments.Nodes) <= index {
		return nil
	}
	argument := call.Arguments.Nodes[index]
	if argument == nil || argument.Kind != ast.KindArrayLiteralExpression {
		return nil
	}
	return argument
}

// isInsideComponent reports whether any enclosing function is named like a component.
//
// Ported from the original's `isInsideComponent`: a name starting with a capital that is not also a
// hook name.
//
// # The hook exclusion cannot fire, and it is kept anyway
//
// `react.IsLikelyComponentName` tests the first character for a capital and `react.IsHookName` tests for a
// leading `use` followed by a capital, so the two are mutually exclusive: `useThing` starts with a
// lowercase `u` and is rejected before the exclusion is consulted. Confirmed by evaluating both
// predicates over the real names rather than by reading them, and the original has the same
// structure with `/^[A-Z]/` and `/^use[A-Z]/`.
//
// So no fixture can kill a mutant that deletes the exclusion, and mutation testing reported it as a
// survivor. It stays because it is what the original says, and because the pair only stays disjoint
// while both predicates keep their current definitions: widening either one silently makes this
// clause load-bearing at a line the change does not touch. Recorded here rather than left looking
// like an untested guard.
//
// Every enclosing function is tested rather than only the nearest, matching the original's loop. The
// nearest is usually the anonymous callback passed to the hook itself, which has no name at all, so
// stopping there would silence the rule everywhere.
func isInsideComponent(node *ast.Node) bool {
	for current := node.Parent; current != nil; current = current.Parent {
		name := scope.NameOf(current)
		if name != "" && react.IsLikelyComponentName(name) && !react.IsHookName(name) {
			return true
		}
	}
	return false
}
