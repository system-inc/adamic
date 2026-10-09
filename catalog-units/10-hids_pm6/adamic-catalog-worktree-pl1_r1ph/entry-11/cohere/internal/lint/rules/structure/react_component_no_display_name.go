package structure

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/react"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// reactComponentNoDisplayNameText is the rule's message, whose wording lives in
// `policy/messages/react-component-no-display-name.json`.
var reactComponentNoDisplayNameText = policy.MessageOf("structure/react-component-no-display-name", "noDisplayNameAssignment")

// messageNoDisplayNameAssignment is the finding, rendered when it is reported so the text comes from the current catalog.
func messageNoDisplayNameAssignment() rule.Message {
	return rule.Message{Id: reactComponentNoDisplayNameText.Id, Description: reactComponentNoDisplayNameText.Render(nil)}
}

// anonymousComponentWrapperNames are the wrappers whose result carries no name of its own.
//
// A component built by one of these genuinely has nothing for DevTools to read, so the assignment
// is the only way to label it and the rule has to allow it. Anything else assigning displayName is
// copying a name that already exists.
var anonymousComponentWrapperNames = map[string]bool{
	"memo":       true,
	"forwardRef": true,
}

// ReactComponentNoDisplayName flags a hand-written displayName assignment on a component.
//
//	valid:   function Field() { ... }                      // the name is already there
//	valid:   const Field = memo(function () { ... }); Field.displayName = 'Field'
//	valid:   event.displayName = organization.displayName  // a domain field, not a component
//	invalid: function Field() { ... } Field.displayName = 'Field'
//
// # Only a component's displayName
//
// `displayName` is also an ordinary field name: an organization, a user, a calendar event all have
// one. Matching the property name alone reported every one of those, and with the rule on in
// api-phi-health all 17 findings were domain fields in plain `.ts` services (#qpms5xz, #gf3a2m8).
// So the target has to be a component this file declares at its top level, recognized by the
// detector the other `react-component` rules share: a capitalized function `IsLikelyReactComponent`
// accepts, a capitalized class or binding extending a React component base, or a `memo` or `forwardRef` result.
//
// The set is gathered from the whole file the first time an assignment needs it, not during the
// walk, because a function declaration hoists: `Field.displayName = 'Field'` above
// `function Field()` runs, and names a component all the same. A file that never assigns displayName
// never pays for the scan.
//
// The silent cases are the right polarity for this rule. A component it cannot see (imported, a
// parameter, a member of an object, a binding from some other factory) is a missed finding, and a
// missed finding here costs a redundant line; a false one asks someone to delete a field their data
// needs.
//
// The exemption is why this rule needs state rather than a single listener. Whether an assignment
// is allowed depends on how the target was declared, which is a different node seen earlier, so the
// rule collects wrapper-initialized bindings during the walk and consults that set at the
// assignment.
//
// Declaration order is therefore load-bearing, and it matches the original: a binding declared
// after the assignment does not exempt it. That is a real limit rather than an oversight, and it
// is the right one, because `Field.displayName = 'x'` above `const Field = memo(...)` is a
// temporal dead zone error at runtime anyway. A rule that exempted it would be excusing code that
// cannot run.
var ReactComponentNoDisplayName = rule.Rule{
	Name: "structure/react-component-no-display-name",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// Scoped to this file's walk, which is the whole lifetime that matters: the original keys
		// the same set on one `create` call per file.
		wrappedAsAnonymous := map[string]bool{}
		// Nil until the first displayName assignment asks for it.
		var componentNames map[string]bool

		return rule.Listeners{
			ast.KindVariableDeclaration: func(node *ast.Node) {
				declaration := node.AsVariableDeclaration()
				name := declaration.Name()
				if name == nil || name.Kind != ast.KindIdentifier {
					return
				}
				if isAnonymousWrapperCall(declaration.Initializer) {
					wrappedAsAnonymous[name.Text()] = true
				}
			},

			ast.KindBinaryExpression: func(node *ast.Node) {
				binary := node.AsBinaryExpression()

				// Only a plain assignment. A compound one (`+=`) cannot be establishing a display
				// name, and reporting it would be a claim about arithmetic on a string.
				if binary.OperatorToken == nil || binary.OperatorToken.Kind != ast.KindEqualsToken {
					return
				}

				left := ast.SkipParentheses(binary.Left)
				if left == nil || left.Kind != ast.KindPropertyAccessExpression {
					// An element access (`Field['displayName']`) is deliberately not matched,
					// following the original. It is not how anyone writes this, and matching it
					// would mean deciding what a computed key resolves to.
					return
				}

				access := left.AsPropertyAccessExpression()
				property := access.Name()
				if property == nil || property.Kind != ast.KindIdentifier || property.Text() != "displayName" {
					return
				}

				target := ast.SkipParentheses(access.Expression)
				if target == nil || target.Kind != ast.KindIdentifier || wrappedAsAnonymous[target.Text()] {
					return
				}
				if componentNames == nil {
					componentNames = componentsDeclaredAtTopLevel(ctx.SourceFile.AsNode())
				}
				if !componentNames[target.Text()] {
					return
				}

				ctx.ReportNode(node, messageNoDisplayNameAssignment())
			},
		}
	},
}

// componentsDeclaredAtTopLevel returns the names of the components a file declares at its top level.
//
// Top level only, as `componentsDeclaredInFile` beside this reads it: a component is declared there,
// and so is the assignment that names it. The returned map is never nil, so the caller computes it
// once even for a file with no components.
func componentsDeclaredAtTopLevel(sourceFile *ast.Node) map[string]bool {
	names := map[string]bool{}
	record := func(name *ast.Node) {
		if name != nil && name.Kind == ast.KindIdentifier && react.IsLikelyComponentName(name.Text()) {
			names[name.Text()] = true
		}
	}
	sourceFile.ForEachChild(func(statement *ast.Node) bool {
		switch statement.Kind {
		case ast.KindFunctionDeclaration:
			if IsLikelyReactComponent(statement) {
				record(statement.Name())
			}
		case ast.KindClassDeclaration:
			if react.IsEs6ComponentClass(statement) {
				record(statement.Name())
			}
		case ast.KindVariableStatement:
			declarationList := statement.AsVariableStatement().DeclarationList
			if declarationList == nil {
				return false
			}
			for _, declarationNode := range declarationList.AsVariableDeclarationList().Declarations.Nodes {
				declaration := declarationNode.AsVariableDeclaration()
				if declaration.Initializer == nil {
					continue
				}
				initializer := ast.SkipParentheses(declaration.Initializer)
				// A memo or forwardRef result is a component too. Declared above the assignment it is exempt,
				// by the walk; below it, the assignment is a temporal dead zone error and reports.
				if IsLikelyReactComponent(initializer) || react.IsEs6ComponentClass(initializer) || isAnonymousWrapperCall(initializer) {
					record(declaration.Name())
				}
			}
		}
		return false
	})
	return names
}

// isAnonymousWrapperCall reports `memo(...)`, `forwardRef(...)`, or their React.-qualified forms.
//
// The nil check comes before SkipParentheses rather than after it. SkipParentheses dereferences its
// argument, so passing it a nil initializer panics, and a declaration without one is ordinary:
// `declare const memo` and `let Field` both reach here. Found by a fixture crashing the walk, which
// is the failure mode worth keeping in mind for the rest of this family, since almost every rule
// here reaches for an initializer that does not have to exist.
func isAnonymousWrapperCall(initializer *ast.Node) bool {
	if initializer == nil {
		return false
	}
	initializer = ast.SkipParentheses(initializer)
	if initializer == nil || initializer.Kind != ast.KindCallExpression {
		return false
	}

	callee := ast.SkipParentheses(initializer.AsCallExpression().Expression)
	if callee == nil {
		return false
	}

	switch callee.Kind {
	case ast.KindIdentifier:
		return anonymousComponentWrapperNames[callee.Text()]
	case ast.KindPropertyAccessExpression:
		return react.IsNamespacedMember(callee, func(name string) bool {
			return anonymousComponentWrapperNames[name]
		})
	}
	return false
}
