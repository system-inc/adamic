package structure

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/module"
	"github.com/system-inc/cohere/internal/lint/ecmascript/react"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// messageUseComponentPropertyTypeSuffixText is the rule's `useComponentPropertyTypeSuffix` message,
// whose wording lives in `policy/messages/react-component-require-properties-type-suffix.json`.
var messageUseComponentPropertyTypeSuffixText = policy.MessageOf("structure/react-component-require-properties-type-suffix", "useComponentPropertyTypeSuffix")

// messageUseComponentPropertyTypeSuffix is the finding, rendered when it is reported so the text
// comes from the current catalog.
func messageUseComponentPropertyTypeSuffix() rule.Message {
	return rule.Message{Id: messageUseComponentPropertyTypeSuffixText.Id, Description: messageUseComponentPropertyTypeSuffixText.Render(nil)}
}

// messageUseInterfaceSuffixText is the rule's `useInterfaceSuffix` message, whose wording lives in
// `policy/messages/react-component-require-properties-type-suffix.json`.
var messageUseInterfaceSuffixText = policy.MessageOf("structure/react-component-require-properties-type-suffix", "useInterfaceSuffix")

// messageUseInterfaceSuffix is the finding, rendered when it is reported so the text comes from the
// current catalog.
func messageUseInterfaceSuffix() rule.Message {
	return rule.Message{Id: messageUseInterfaceSuffixText.Id, Description: messageUseInterfaceSuffixText.Render(nil)}
}

// messageUseTypeAliasSuffixText is the rule's `useTypeAliasSuffix` message, whose wording lives in
// `policy/messages/react-component-require-properties-type-suffix.json`.
var messageUseTypeAliasSuffixText = policy.MessageOf("structure/react-component-require-properties-type-suffix", "useTypeAliasSuffix")

// messageUseTypeAliasSuffix is the finding, rendered when it is reported so the text comes from the
// current catalog.
func messageUseTypeAliasSuffix() rule.Message {
	return rule.Message{Id: messageUseTypeAliasSuffixText.Id, Description: messageUseTypeAliasSuffixText.Render(nil)}
}

// ReactComponentRequirePropertiesTypeSuffix flags a component property type not named `*Properties`.
//
//	valid:   function Button(properties: ButtonProperties) { ... }
//	invalid: function Button(properties: ButtonProps) { ... }
//	invalid: interface ButtonInterface {} function Button(properties: ButtonInterface) { ... }
//
// Three message ids, and they are three sites rather than three defects. One wrong name produces a
// finding at the parameter and another at the declaration, because a reader fixing this has to
// touch both and seeing only one would leave the rename half done.
//
// The declaration findings depend on a set the parameter visitor fills, so declaration order
// decides: an interface declared after the component that uses it is reported, one declared before
// is not seen yet when its own visitor runs. That is the original's behavior and it is reproduced
// rather than corrected, since the alternative walks the file twice for a finding whose use site is
// already reported.
//
// The rename is derived rather than appended blindly. `ButtonProps` becomes `ButtonProperties` and
// `ButtonInterface` becomes `ButtonProperties`, because appending would produce `ButtonPropsProperties`,
// which nobody wants and which the author would have to fix by hand anyway.
var ReactComponentRequirePropertiesTypeSuffix = rule.Rule{
	Name: "structure/react-component-require-properties-type-suffix",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if !FileContextFor(ctx.SourceFile.FileName().AsString()).IsReactFile {
			return nil
		}

		// Type names seen at a component's first parameter that lack the suffix. Filled by the
		// component visitors and read by the declaration ones.
		componentPropertyTypes := map[string]bool{}

		// Type names the file already declares. A rename onto one of these produces a duplicate
		// identifier, and where the taken name is the alias's own right-hand side it also produces a
		// type that references itself. Both parse, so the fix engine cannot refuse them.
		declaredTypeNames := map[string]bool{}

		checkParameterType := func(parameters []*ast.Node) {
			if len(parameters) == 0 {
				return
			}

			// A plain identifier parameter only, matching the original. A destructured parameter is
			// `react-component-no-destructuring`'s business, and its type is often a composed one
			// (`TypeA & { ... }`) rather than the single name this rule renames.
			//
			// Caught by the tree rather than by review: without this the rule reported
			// `MenuItem.tsx`, whose parameter destructures and is typed `MenuItemInterface`. That
			// is a genuine violation of the convention and the gate does not report it, so
			// reporting it here would have been a second only-cohere divergence, arrived at by
			// accident rather than by a ruling.
			parameterName := parameters[0].AsParameterDeclaration().Name()
			if parameterName == nil || parameterName.Kind != ast.KindIdentifier {
				return
			}

			typeName := typeReferenceName(parameterTypeNode(parameters[0]))
			if typeName == "" || strings.HasSuffix(typeName, "Properties") {
				return
			}

			componentPropertyTypes[typeName] = true

			// Reported on the type reference rather than the parameter, so the finding underlines
			// the name that has to change.
			typeNode := parameterTypeNode(parameters[0])
			if declaredTypeNames[propertiesTypeNameFor(typeName)] {
				// The target name is taken, so renaming onto it would break the file. Reported
				// without a fix rather than silenced: the name still violates the convention and
				// the author is the one who can decide which of the two survives.
				ctx.ReportNode(typeNode.AsTypeReferenceNode().TypeName,
					messageUseComponentPropertyTypeSuffix())
				return
			}
			ctx.ReportNodeWithFixes(typeNode.AsTypeReferenceNode().TypeName,
				messageUseComponentPropertyTypeSuffix(),
				ctx.ReplaceNode(typeNode.AsTypeReferenceNode().TypeName, propertiesTypeNameFor(typeName)))
		}

		reportDeclaration := func(name *ast.Node, message rule.Message) {
			if name == nil || name.Kind != ast.KindIdentifier {
				return
			}
			if !componentPropertyTypes[name.Text()] || strings.HasSuffix(name.Text(), "Properties") {
				return
			}
			if declaredTypeNames[propertiesTypeNameFor(name.Text())] {
				ctx.ReportNode(name, message)
				return
			}
			ctx.ReportNodeWithFixes(name, message, ctx.ReplaceNode(name, propertiesTypeNameFor(name.Text())))
		}

		return rule.Listeners{
			// The map has to be complete before any declaration is judged, and a single walk in
			// source order does not guarantee that: an interface declared above the component that
			// uses it is visited first, sees an empty map, and is skipped.
			//
			// That silence was the defect. The usage still got its fix, so the rule renamed
			// `CardProps` to `CardProperties` at the call site and left the declaration named
			// `CardProps`, producing a file that parses and does not compile. Declaration-first is
			// the conventional ordering, so it was the common case rather than an edge, and the
			// fixtures passed because on that ordering the rule emits one finding instead of two,
			// which `ExpectFindings` cannot distinguish from correct behavior.
			//
			// The source-file listener fires before its children, so filling the map here removes
			// the ordering assumption rather than reversing it. Same shape as `react/no-multi-comp`,
			// which collects in this listener for the same reason.
			ast.KindSourceFile: func(node *ast.Node) {
				// Every type name the file declares, so a rename can refuse a target that is
				// already taken. The four declaration kinds are the shelf's, which counts an enum
				// as well: an enum declares a type as much as a value, and renaming onto one
				// produces a duplicate identifier exactly as renaming onto an interface does.
				for name := range module.AllDeclaredTypeNames(ctx.SourceFile) {
					declaredTypeNames[name] = true
				}
				collectComponentPropertyTypeNames(node, componentPropertyTypes)
			},

			ast.KindFunctionDeclaration: func(node *ast.Node) {
				declaration := node.AsFunctionDeclaration()
				name := declaration.Name()
				if name == nil || !react.IsLikelyComponentName(name.Text()) {
					return
				}
				checkParameterType(parameterNodes(declaration.Parameters))
			},

			ast.KindVariableDeclaration: func(node *ast.Node) {
				declaration := node.AsVariableDeclaration()
				name := declaration.Name()
				if name == nil || name.Kind != ast.KindIdentifier || !react.IsLikelyComponentName(name.Text()) {
					return
				}
				// The nil check has to precede SkipParentheses, which dereferences its argument.
				// `let Button;` is ordinary code and reaches this listener.
				if declaration.Initializer == nil {
					return
				}
				initializer := ast.SkipParentheses(declaration.Initializer)
				if initializer == nil {
					return
				}
				switch initializer.Kind {
				case ast.KindArrowFunction:
					checkParameterType(parameterNodes(initializer.AsArrowFunction().Parameters))
				case ast.KindFunctionExpression:
					checkParameterType(parameterNodes(initializer.AsFunctionExpression().Parameters))
				}
			},

			ast.KindInterfaceDeclaration: func(node *ast.Node) {
				reportDeclaration(node.AsInterfaceDeclaration().Name(), messageUseInterfaceSuffix())
			},

			ast.KindTypeAliasDeclaration: func(node *ast.Node) {
				reportDeclaration(node.AsTypeAliasDeclaration().Name(), messageUseTypeAliasSuffix())
			},
		}
	},
}

// collectComponentPropertyTypeNames records every type a component in this file takes as properties.
//
// A pre-pass rather than a side effect of the component listener, because the declaration listener
// needs the complete set before it judges anything and walk order does not supply that. Reaching
// every component means walking the whole tree rather than the top level: a component can be
// declared inside a function, and the original's map was filled from a listener that saw those too.
func collectComponentPropertyTypeNames(sourceFile *ast.Node, into map[string]bool) {
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node == nil {
			return
		}

		var name *ast.Node
		var parameters []*ast.Node

		switch node.Kind {
		case ast.KindFunctionDeclaration:
			declaration := node.AsFunctionDeclaration()
			name = declaration.Name()
			parameters = parameterNodes(declaration.Parameters)

		case ast.KindVariableDeclaration:
			declaration := node.AsVariableDeclaration()
			name = declaration.Name()
			if name == nil || name.Kind != ast.KindIdentifier || declaration.Initializer == nil {
				break
			}
			initializer := ast.SkipParentheses(declaration.Initializer)
			if initializer == nil {
				break
			}
			switch initializer.Kind {
			case ast.KindArrowFunction:
				parameters = parameterNodes(initializer.AsArrowFunction().Parameters)
			case ast.KindFunctionExpression:
				parameters = parameterNodes(initializer.AsFunctionExpression().Parameters)
			}
		}

		if name != nil && name.Kind == ast.KindIdentifier &&
			react.IsLikelyComponentName(name.Text()) && len(parameters) == 1 {
			if typeName := typeReferenceName(parameterTypeNode(parameters[0])); typeName != "" &&
				!strings.HasSuffix(typeName, "Properties") {
				into[typeName] = true
			}
		}

		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}
	visit(sourceFile)
}

// propertiesTypeNameFor derives the corrected name rather than appending to the wrong one.
//
// `ButtonProps` and `ButtonInterface` both become `ButtonProperties`. Appending would produce
// `ButtonPropsProperties`, which nobody wants and which the author would have to fix by hand,
// making the fix worse than no fix.
func propertiesTypeNameFor(name string) string {
	for _, suffix := range []string{"Interface", "Props"} {
		if strings.HasSuffix(name, suffix) {
			return strings.TrimSuffix(name, suffix) + "Properties"
		}
	}
	return name + "Properties"
}
