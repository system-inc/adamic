package structure

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/react"
	"github.com/system-inc/cohere/internal/lint/ecmascript/scope"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// messageNoDestructuringText is the rule's `noDestructuring` message, whose wording lives in
// `policy/messages/react-component-no-destructuring.json`.
var messageNoDestructuringText = policy.MessageOf("structure/react-component-no-destructuring", "noDestructuring")

// messageNoDestructuring is the finding, rendered when it is reported so the text comes from the
// current catalog.
func messageNoDestructuring() rule.Message {
	return rule.Message{Id: messageNoDestructuringText.Id, Description: messageNoDestructuringText.Render(nil)}
}

// messageNoDestructuringFromPropsSourceText is the rule's `noDestructuringFromPropsSource` message,
// whose wording lives in `policy/messages/react-component-no-destructuring.json`.
var messageNoDestructuringFromPropsSourceText = policy.MessageOf("structure/react-component-no-destructuring", "noDestructuringFromPropsSource")

// messageNoDestructuringFromPropsSource is the finding, rendered when it is reported so the text
// comes from the current catalog.
func messageNoDestructuringFromPropsSource() rule.Message {
	return rule.Message{Id: messageNoDestructuringFromPropsSourceText.Id, Description: messageNoDestructuringFromPropsSourceText.Render(nil)}
}

// messageRequirePropertiesSuffixText is the rule's `requirePropertiesSuffix` message, whose wording
// lives in `policy/messages/react-component-no-destructuring.json`.
var messageRequirePropertiesSuffixText = policy.MessageOf("structure/react-component-no-destructuring", "requirePropertiesSuffix")

// messageRequirePropertiesSuffix is the finding, rendered when it is reported so the text comes
// from the current catalog.
func messageRequirePropertiesSuffix() rule.Message {
	return rule.Message{Id: messageRequirePropertiesSuffixText.Id, Description: messageRequirePropertiesSuffixText.Render(nil)}
}

// messageSemanticSpreadNameText is the rule's `semanticSpreadName` message, whose wording lives in
// `policy/messages/react-component-no-destructuring.json`.
var messageSemanticSpreadNameText = policy.MessageOf("structure/react-component-no-destructuring", "semanticSpreadName")

// messageSemanticSpreadName is the finding, rendered when it is reported so the text comes from the
// current catalog.
func messageSemanticSpreadName() rule.Message {
	return rule.Message{Id: messageSemanticSpreadNameText.Id, Description: messageSemanticSpreadNameText.Render(nil)}
}

// messageSpreadMustBeUsedText is the rule's `spreadMustBeUsed` message, whose wording lives in
// `policy/messages/react-component-no-destructuring.json`.
var messageSpreadMustBeUsedText = policy.MessageOf("structure/react-component-no-destructuring", "spreadMustBeUsed")

// messageSpreadMustBeUsed is the finding, rendered when it is reported so the text comes from the
// current catalog.
func messageSpreadMustBeUsed() rule.Message {
	return rule.Message{Id: messageSpreadMustBeUsedText.Id, Description: messageSpreadMustBeUsedText.Render(nil)}
}

// genericSpreadNames are rest names that describe being a rest rather than what they hold.
var genericSpreadNames = map[string]bool{
	"restProperties":      true,
	"otherProperties":     true,
	"remainingProperties": true,
	"extraProperties":     true,
}

// propertiesSourceNames are the names a component's properties object goes by.
var propertiesSourceNames = map[string]bool{
	"properties": true,
	"props":      true,
}

// ReactComponentNoDestructuring flags destructuring a component's properties without a used rest.
//
//	valid:   function Button(properties: ButtonProperties) { return <button>{properties.label}</button> }
//	valid:   function Button({ label, ...buttonProperties }) { return <button {...buttonProperties}>{label}</button> }
//	invalid: function Button({ label }) { ... }                       no rest
//	invalid: function Button({ label, ...rest }) { ... }              rest not named for an element
//	invalid: function Button({ label, ...restProperties }) { ... }    named for being a rest
//	invalid: function Button({ label, ...buttonProperties }) { }      rest never used
//	invalid: function Button(properties) { const { label } = properties }
//
// Five message ids for five distinct repairs, which is the whole reason this rule is long. A
// missing rest, a badly named one, a generically named one, and an unused one are four different
// edits, and one message covering all of them would say only that something is wrong.
//
// The `useImperativeHandle` exemption is the subtle part. A component destructuring `ref` and
// exposing an imperative handle is not the leaf-DOM passthrough this rule's rest-spread convention
// was designed for, and React 19 makes `ref` an ordinary property that has to be taken out at the
// boundary. Requiring a rest there would be asking for a shape that does not fit.
var ReactComponentNoDestructuring = rule.Rule{
	Name: "structure/react-component-no-destructuring",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if !FileContextFor(ctx.SourceFile.FileName().AsString()).IsReactFile {
			return nil
		}

		checkParameters := func(node *ast.Node, parameters []*ast.Node) {
			name := scope.NameOf(node)
			if name == "" || !react.IsLikelyComponentName(name) {
				return
			}
			if !HasJsxOrReactHookCalls(node) {
				// A capitalized function that renders nothing is not a component, and holding it
				// to a component's convention would flag ordinary factories and builders.
				return
			}

			for _, parameter := range parameters {
				pattern := parameter.AsParameterDeclaration().Name()
				if pattern == nil || pattern.Kind != ast.KindObjectBindingPattern {
					continue
				}
				if message, ok := destructuringProblem(pattern, node); ok {
					ctx.ReportNode(pattern, message)
				}
			}
		}

		return rule.Listeners{
			ast.KindFunctionDeclaration: func(node *ast.Node) {
				checkParameters(node, parameterNodes(node.AsFunctionDeclaration().Parameters))
			},
			ast.KindFunctionExpression: func(node *ast.Node) {
				checkParameters(node, parameterNodes(node.AsFunctionExpression().Parameters))
			},
			ast.KindArrowFunction: func(node *ast.Node) {
				checkParameters(node, parameterNodes(node.AsArrowFunction().Parameters))
			},

			// `const { label } = properties` inside a component body, which is the same mistake
			// written one line later.
			ast.KindVariableDeclaration: func(node *ast.Node) {
				declaration := node.AsVariableDeclaration()

				pattern := declaration.Name()
				if pattern == nil || pattern.Kind != ast.KindObjectBindingPattern {
					return
				}
				// SkipParentheses dereferences its argument, so the nil check precedes it. A
				// binding pattern with no initializer is a syntax error but reaches the walk
				// while someone is mid-edit, and a linter that panics on a half-written file is
				// worse than one that says nothing about it.
				if declaration.Initializer == nil {
					return
				}
				initializer := ast.SkipParentheses(declaration.Initializer)
				if initializer == nil || initializer.Kind != ast.KindIdentifier {
					return
				}
				if !propertiesSourceNames[initializer.Text()] {
					return
				}

				containing := scope.EnclosingFunctionLike(node)
				if containing == nil || !react.IsLikelyComponentName(scope.NameOf(containing)) {
					return
				}

				// No rest at all is its own message here, because the repair is different: the
				// whole statement goes, rather than a rest being added to it.
				if restBindingName(pattern) == nil {
					ctx.ReportNode(pattern, messageNoDestructuringFromPropsSource())
					return
				}
				if message, ok := destructuringProblem(pattern, containing); ok {
					ctx.ReportNode(pattern, message)
				}
			},
		}
	},
}

// destructuringProblem judges one object pattern, returning the message when it is wrong.
func destructuringProblem(pattern *ast.Node, functionNode *ast.Node) (rule.Message, bool) {
	// A component taking `ref` out and exposing an imperative handle is not a passthrough, so the
	// rest-spread convention does not apply to it.
	if bindsReferenceShorthand(pattern) && callsNamedHook(scope.BodyOf(functionNode), "useImperativeHandle") {
		return rule.Message{}, false
	}

	restName := restBindingName(pattern)
	if restName == nil {
		return messageNoDestructuring(), true
	}
	name := restName.Text()

	if !strings.HasSuffix(name, "Properties") {
		return messageRequirePropertiesSuffix(), true
	}
	if genericSpreadNames[name] {
		return messageSemanticSpreadName(), true
	}
	if !bodyReferencesName(scope.BodyOf(functionNode), name) {
		return messageSpreadMustBeUsed(), true
	}

	return rule.Message{}, false
}

// restBindingName returns the identifier a rest element binds, or nil when there is no simple one.
//
// A rest that destructures further (`...{ a }`) is not valid JavaScript, so the only shapes here
// are an identifier or nothing.
func restBindingName(pattern *ast.Node) *ast.Node {
	for _, element := range pattern.AsBindingPattern().Elements.Nodes {
		binding := element.AsBindingElement()
		if binding.DotDotDotToken == nil {
			continue
		}
		name := binding.Name()
		if name != nil && name.Kind == ast.KindIdentifier {
			return name
		}
	}
	return nil
}

// bindsReferenceShorthand reports whether the pattern takes a property named ref.
func bindsReferenceShorthand(pattern *ast.Node) bool {
	for _, element := range pattern.AsBindingPattern().Elements.Nodes {
		binding := element.AsBindingElement()
		if binding.DotDotDotToken != nil {
			continue
		}

		// The bound name for a shorthand, the source property for a renamed one. The rule is about
		// which property is taken, not what it is called locally.
		source := binding.PropertyName
		if source == nil {
			source = binding.Name()
		}
		if source != nil && source.Kind == ast.KindIdentifier && source.Text() == "ref" {
			return true
		}
	}
	return false
}

// spreadUsageDepthLimit bounds the body walks, matching the original's limit of 40.
//
// Deliberately not the JSX walk's limit of 20: these two searches have different originals with
// different bounds, and the bound is behavior. A rest variable used deeper than this reads as
// unused to the gate, and an unbounded walk would find it and silence a finding the gate has.
const spreadUsageDepthLimit = 40

// bodyReferencesName reports whether an identifier of this name appears anywhere in a body.
//
// A plain name search rather than a resolved one, matching the original. It over-approximates:
// a shadowed binding of the same name would count as a use. That direction is chosen, since the
// alternative reports a rest that is used and tells the author to delete working code.
func bodyReferencesName(body *ast.Node, name string) bool {
	if body == nil {
		return false
	}

	found := false
	var visit func(*ast.Node, int)
	visit = func(current *ast.Node, depth int) {
		if current == nil || found || depth > spreadUsageDepthLimit {
			return
		}
		if current.Kind == ast.KindIdentifier && current.Text() == name {
			found = true
			return
		}
		current.ForEachChild(func(child *ast.Node) bool {
			visit(child, depth+1)
			return found
		})
	}
	visit(body, 0)
	return found
}

// callsNamedHook reports a call to a specifically named hook anywhere in a body.
func callsNamedHook(body *ast.Node, hookName string) bool {
	if body == nil {
		return false
	}

	found := false
	var visit func(*ast.Node, int)
	visit = func(current *ast.Node, depth int) {
		if current == nil || found || depth > spreadUsageDepthLimit {
			return
		}
		if current.Kind == ast.KindCallExpression {
			callee := ast.SkipParentheses(current.AsCallExpression().Expression)
			switch {
			case callee == nil:
			case callee.Kind == ast.KindIdentifier && callee.Text() == hookName:
				found = true
				return
			case callee.Kind == ast.KindPropertyAccessExpression:
				method := callee.AsPropertyAccessExpression().Name()
				if method != nil && method.Text() == hookName {
					found = true
					return
				}
			}
		}
		current.ForEachChild(func(child *ast.Node) bool {
			visit(child, depth+1)
			return found
		})
	}
	visit(body, 0)
	return found
}
