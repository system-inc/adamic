package structure

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// messageUsePropertiesNotPropsText is the rule's `usePropertiesNotProps` message, whose wording
// lives in `policy/messages/react-component-require-properties-parameter.json`.
var messageUsePropertiesNotPropsText = policy.MessageOf("structure/react-component-require-properties-parameter", "usePropertiesNotProps")

// messageUsePropertiesNotProps is the finding, rendered when it is reported so the text comes from
// the current catalog.
func messageUsePropertiesNotProps() rule.Message {
	return rule.Message{Id: messageUsePropertiesNotPropsText.Id, Description: messageUsePropertiesNotPropsText.Render(nil)}
}

// messageRenameToPropertiesText is the rule's `renameToProperties` message, whose wording lives in
// `policy/messages/react-component-require-properties-parameter.json`.
var messageRenameToPropertiesText = policy.MessageOf("structure/react-component-require-properties-parameter", "renameToProperties")

// messageRenameToProperties is the suggestion, rendered when it is reported so the text comes from
// the current catalog.
func messageRenameToProperties() rule.Message {
	return rule.Message{Id: messageRenameToPropertiesText.Id, Description: messageRenameToPropertiesText.Render(nil)}
}

// ReactComponentRequirePropertiesParameter flags a component whose first parameter is named `props`.
//
//	valid:   function Button(properties: ButtonProperties) { ... }
//	valid:   function Button({ label, ...buttonProperties }) { ... }
//	invalid: function Button(props: ButtonProperties) { ... }
//
// Only the name, and only when it is exactly `props`. A destructured parameter has no single name
// to judge and is another rule's business; a parameter named anything else is the author's choice.
//
// A suggestion rather than a fix, and the distinction is the whole of it.
//
// The rename rewrites only the declaration, not the uses, so `props.label` inside a function whose
// parameter is now `properties` does not compile. That is deliberate: the alternative renames every
// reference and has to decide which `props` in the body are the parameter and which are somebody
// else's, and a rename that stops at the declaration fails immediately and visibly where one that
// guesses at references fails quietly and in the wrong place.
//
// **But a deliberate visible break only works if somebody sees it.** This shipped as
// `ReportNodeWithFixes`, which the engine applies unattended, so the entire argument above rested on
// a human reading a diff that nothing showed them. The reasoning was right and the verb was wrong:
// a fix preserves what the code means, and this changes it by design.
//
// So the break stays and the automation goes. The author is shown the rename, and is the one who
// then updates the body, which is the work only they can do correctly anyway.
var ReactComponentRequirePropertiesParameter = rule.Rule{
	Name: "structure/react-component-require-properties-parameter",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if !FileContextFor(ctx.SourceFile.FileName().AsString()).IsReactFile {
			return nil
		}

		check := func(functionNode *ast.Node, parameters []*ast.Node) {
			if len(parameters) == 0 {
				return
			}
			if !HasJsxOrReactHookCalls(functionNode) {
				// A function that renders nothing is not a component, and its parameter names are
				// nobody's business but its author's.
				return
			}

			name := parameters[0].AsParameterDeclaration().Name()
			if name == nil || name.Kind != ast.KindIdentifier || name.Text() != "props" {
				return
			}
			ctx.ReportNodeWithSuggestions(name, messageUsePropertiesNotProps(), rule.Suggestion{
				Message: messageRenameToProperties(),
				Fixes:   []rule.Fix{ctx.ReplaceNode(name, "properties")},
			})
		}

		return rule.Listeners{
			ast.KindFunctionDeclaration: func(node *ast.Node) {
				check(node, parameterNodes(node.AsFunctionDeclaration().Parameters))
			},
			ast.KindFunctionExpression: func(node *ast.Node) {
				check(node, parameterNodes(node.AsFunctionExpression().Parameters))
			},
			ast.KindArrowFunction: func(node *ast.Node) {
				check(node, parameterNodes(node.AsArrowFunction().Parameters))
			},
		}
	},
}
