package structure

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/module"
	"github.com/system-inc/cohere/internal/lint/ecmascript/react"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// reactComponentRequireNamedExportText is the rule's message, whose wording lives in
// `policy/messages/react-component-require-named-export.json`.
var reactComponentRequireNamedExportText = policy.MessageOf("structure/react-component-require-named-export", "componentRequiresNamedExport")

// messageComponentRequiresNamedExport is the finding, rendered when it is reported so the text
// comes from the current catalog.
func messageComponentRequiresNamedExport() rule.Message {
	return rule.Message{Id: reactComponentRequireNamedExportText.Id, Description: reactComponentRequireNamedExportText.Render(nil)}
}

// ReactComponentRequireNamedExport flags a component file with no named export.
//
//	valid:   export function Button() { return <div />; }
//	valid:   function Button() { return <div />; } export { Button };
//	valid:   export default function Page() { ... }        in a Next.js page, which is exempt by path
//	invalid: function Button() { return <div />; } export default Button;
//
// # Where the file has to be for this to run
//
// Only `.tsx` and `.jsx`, and never a Next.js special file. A page, layout, error, not-found,
// loading or route file has a default export because the framework reads it by name from the
// filesystem, so requiring a named export there would be requiring the file to break.
//
// # The exposure, and why that changes what the fixtures are for
//
// This is the opposite situation from a rule whose construct is absent. Measured on the tree before
// this was written: 1,192 non-special `.tsx` files, 1,116 of them declaring a component with a named
// export, and 0 that would report. So the rule runs on more than a thousand files and has to stay
// silent on every one.
//
// That inverts where the risk sits. A rule that is too narrow reports nothing and looks correct; a
// rule that is too eager reports hundreds of times and is obvious immediately. The dangerous
// direction here is the second one being subtly true of a shape that is rare, so the clean fixtures
// carry the weight and each one names a shape that already exists in the tree.
//
// # One thing in the original that cannot fire, reproduced as a comment rather than as code
//
// The original consults an `isNested` flag and returns early when the file's first component is
// nested inside another function. That branch is unreachable. `isNested` is a parameter threaded
// through its collectors and every one of the four call sites passes `false`; nothing recurses into
// a function body, so no collected component is ever marked nested.
//
// Verified by reading all four call sites rather than inferred from the shape. Porting the check
// would mean writing nesting detection the original does not have, which would make this rule
// quieter than the gate on files the gate reports. So the flag is not ported, and the reason is
// recorded here because a later reader comparing the two files will otherwise see a missing guard.
var ReactComponentRequireNamedExport = rule.Rule{
	Name:       "structure/react-component-require-named-export",
	NoListener: rule.NoListenerAnswersInRun,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		fileContext := FileContextFor(ctx.SourceFile.FileName().AsString())
		if !fileContext.IsReactFile || fileContext.IsSpecialNextJsFile {
			return nil
		}

		// The decision is about the file rather than about any node in it, so it is made once here
		// rather than in a listener. Returning nil listeners means the walk never enters this file
		// on this rule's behalf, which is the cheapest a rule can be.
		components := componentsDeclaredInFile(ctx.SourceFile.AsNode())
		if len(components) == 0 {
			return nil
		}
		for _, component := range components {
			if component.hasNamedExport {
				return nil
			}
		}

		// The first component declared, matching the original. A file with several components and no
		// named export reports once, naming the first, because the fix is to export the one the file
		// is about and the rest are its helpers.
		ctx.ReportNode(components[0].nameNode, messageComponentRequiresNamedExport())
		return nil
	},
}

// declaredComponent is one component this file declares, and whether it is exported by name.
type declaredComponent struct {
	name           string
	nameNode       *ast.Node
	hasNamedExport bool
}

// componentsDeclaredInFile collects the components a file declares, in source order.
//
// Order is load-bearing: the finding names the first one, so a map would make the message depend on
// iteration order and the same file would report different names on different runs.
//
// Only top-level statements are examined, matching the original, which walks `program.body` and
// never descends into a function. A component declared inside another function is invisible to both.
func componentsDeclaredInFile(sourceFile *ast.Node) []declaredComponent {
	var components []declaredComponent

	// record adds a component, or upgrades an existing entry's export status. The original
	// deduplicates by name and function node for the same reason: a function declaration reached
	// both as a plain statement and as an export declaration must not count twice.
	record := func(name *ast.Node, exportedByName bool) {
		if name == nil || name.Kind != ast.KindIdentifier {
			return
		}
		text := name.Text()
		for index := range components {
			if components[index].name == text {
				if exportedByName {
					components[index].hasNamedExport = true
				}
				return
			}
		}
		components = append(components, declaredComponent{
			name: text, nameNode: name, hasNamedExport: exportedByName,
		})
	}

	sourceFile.ForEachChild(func(statement *ast.Node) bool {
		switch statement.Kind {
		case ast.KindFunctionDeclaration:
			declaration := statement.AsFunctionDeclaration()
			name := declaration.Name()
			if name == nil || !react.IsLikelyComponentName(name.Text()) {
				return false
			}
			// IsLikelyReactComponent rather than HasJsxOrReactHookCalls, and the difference is not
			// cosmetic: the original's analysis calls this predicate, and it accepts a function
			// whose first parameter is named `properties` before looking for any JSX. Using the
			// JSX search instead reported ReportBlockRenderer.tsx on the live tree, a file that
			// does export its component by name, because that component dispatches from a switch
			// and a bounded property walk cannot enter a switch arm.
			if !IsLikelyReactComponent(statement) {
				return false
			}
			record(name, module.IsExportedByName(declaration.Modifiers()))

		case ast.KindVariableStatement:
			variableStatement := statement.AsVariableStatement()
			declarationList := variableStatement.DeclarationList
			if declarationList == nil {
				return false
			}
			exportedByName := module.IsExportedByName(variableStatement.Modifiers())
			for _, declarationNode := range declarationList.AsVariableDeclarationList().Declarations.Nodes {
				declaration := declarationNode.AsVariableDeclaration()
				name := declaration.Name()
				if name == nil || name.Kind != ast.KindIdentifier || !react.IsLikelyComponentName(name.Text()) {
					continue
				}
				if declaration.Initializer == nil {
					continue
				}
				// `const Thing = React.forwardRef(...)` counts as a component regardless of whether
				// the search finds JSX inside it, matching the original, which special-cases the
				// call before applying its component test. A forwardRef wrapping a render function
				// that only spreads props has no JSX of its own and is still a component.
				if isForwardRefCall(declaration.Initializer) || IsLikelyReactComponent(declaration.Initializer) {
					record(name, exportedByName)
				}
			}

		case ast.KindExportDeclaration:
			// `export { Thing }` naming something declared above. A re-export carries a module
			// specifier and names nothing this file declares, so it cannot satisfy the requirement
			// and is skipped.
			declaration := statement.AsExportDeclaration()
			if declaration.ModuleSpecifier != nil || declaration.ExportClause == nil {
				return false
			}
			if declaration.ExportClause.Kind != ast.KindNamedExports {
				return false
			}
			for _, element := range declaration.ExportClause.AsNamedExports().Elements.Nodes {
				specifier := element.AsExportSpecifier()
				if specifier == nil {
					continue
				}
				// Two names, and confusing them is the defect this shape invites.
				//
				//	export { Button }                Name() = Button,  PropertyName = nil
				//	export { Button as Other }       Name() = Other,   PropertyName = Button
				//	export { Button as default }     Name() = default, PropertyName = Button
				//
				// Name() is the name the outside world sees; PropertyName is the local declaration
				// being exported. The component to mark is the local one, and the name that decides
				// whether this counts as a named export is the exported one.
				//
				// Measured against the parser rather than assumed. Matching the component against
				// Name() alone works for the plain form and silently fails for both aliased forms:
				// `export { Button as Other }` would leave Button unmarked and report a file that
				// does export it, and the `default` guard would be unreachable because "default"
				// never equals a component's name.
				exported := specifier.Name()
				if exported == nil || exported.Kind != ast.KindIdentifier {
					continue
				}
				local := exported
				if specifier.PropertyName != nil && specifier.PropertyName.Kind == ast.KindIdentifier {
					local = specifier.PropertyName
				}
				// `export { Thing as default }` is a default export wearing named-export syntax, so
				// it does not satisfy a rule that exists to require a name consumers must use.
				if exported.Text() == "default" {
					continue
				}
				for index := range components {
					if components[index].name == local.Text() {
						components[index].hasNamedExport = true
					}
				}
			}
		}
		return false
	})

	return components
}

// isForwardRefCall reports a `React.forwardRef(...)` or bare `forwardRef(...)` initializer.
func isForwardRefCall(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindCallExpression {
		return false
	}
	callee := node.AsCallExpression().Expression
	if callee == nil {
		return false
	}
	switch callee.Kind {
	case ast.KindIdentifier:
		return callee.Text() == "forwardRef"
	case ast.KindPropertyAccessExpression:
		name := callee.AsPropertyAccessExpression().Name()
		return name != nil && name.Kind == ast.KindIdentifier && name.Text() == "forwardRef"
	}
	return false
}
