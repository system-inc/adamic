package structure

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/module"
	"github.com/system-inc/cohere/internal/lint/ecmascript/react"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// reactComponentNoConstAssignmentText is the rule's message, whose wording lives in
// `policy/messages/react-component-no-const-assignment.json`.
var reactComponentNoConstAssignmentText = policy.MessageOf("structure/react-component-no-const-assignment", "noConstAssignment")

// messageNoConstAssignment names the component and what to write instead. The declaration to write is
// one value, code the rule builds, since whether it is exported changes the code and not the wording.
func messageNoConstAssignment(identifierName string, isExported bool) rule.Message {
	prefix := ""
	if isExported {
		prefix = "export "
	}
	return rule.Message{
		Id:          reactComponentNoConstAssignmentText.Id,
		Description: reactComponentNoConstAssignmentText.Render(map[string]string{"declaration": prefix + "function " + identifierName + "()"}),
	}
}

// ReactComponentNoConstAssignment reports a component declared as a const holding a function.
//
//	valid:   export function Thing() { return <div />; }
//	valid:   export const Thing = React.forwardRef(function (properties, reference) { ... })
//	valid:   export const Thing = { render() {} }
//	invalid: export const Thing = () => <div />
//	invalid: const Thing = function () { return <div />; }
//
// Ported from `structure/react-component-no-const-assignment`.
//
// # This rule reports and does not fix, which is a deliberate departure
//
// The original is `fixable: 'code'` and rewrites the whole declaration into a function statement.
// That fix is not ported, and the reason is three specific defects rather than general caution.
// Each was read out of the original rather than inferred:
//
//   - **A type annotation is dropped.** `const Thing: React.FC<Properties> = () => {}` becomes
//     `function Thing() {}`. The fix builds its replacement from the identifier name and the
//     initializer's parameters and body; `declaration.id.typeAnnotation` is never read.
//   - **`async` is lost.** The arrow branch emits `function Name(params) body` with no modifier,
//     so `const Thing = async () => {}` becomes a non-async function. That is a behavior change
//     the type checker catches only if the result is awaited somewhere typed.
//   - **Two declarators corrupt the statement.** The rule reports per declarator and the fix
//     replaces the whole `VariableDeclaration`, so `const A = () => {}, B = () => {}` produces two
//     fixes over the same range with different text. One is applied and the other discarded, which
//     silently deletes a component.
//
// The tree has no instances of this construct at all, so none of these can be caught by running the
// rule over it, and fixtures only prove what their author already believed. Reporting without a fix
// is the subset that can be shown correct, and it loses nothing the gate has today: the gate has
// been holding this at zero for months, so there is nothing waiting to be fixed.
//
// # forwardRef is exempt, and the exemption cannot fire
//
// A `React.forwardRef(...)` component has to be a const, because the value is what the call returns.
// The original checks the callee explicitly, before its initializer-kind test.
//
// That check is unreachable in both implementations, and mutation testing is what showed it:
// deleting the exemption here kills no fixture. A `forwardRef(...)` initializer is a call
// expression, and the kind test two lines below already rejects everything that is not a function
// expression or an arrow, so the exemption never decides anything.
//
// Kept because it is what the original says, and because it is dead only by a relationship between
// two checks rather than by its own logic: widening the kind test to accept calls, which a later
// rule about component factories might want, silently makes this exemption load-bearing at a line
// that change does not touch. A reader who found it deleted would have no way to learn that.
var ReactComponentNoConstAssignment = rule.Rule{
	Name: "structure/react-component-no-const-assignment",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil || !FileContextFor(ctx.SourceFile.FileName().AsString()).IsReactFile {
			return nil
		}

		return rule.Listeners{
			ast.KindVariableStatement: func(node *ast.Node) {
				statement := node.AsVariableStatement()
				if statement.DeclarationList == nil {
					return
				}
				// `export` is a modifier on the statement here rather than a wrapping node, which is
				// the whole difference from ESTree's two visitors. The original registers
				// ExportNamedDeclaration and VariableDeclaration separately and guards the second
				// against double-reporting the first; one listener covers both cases here, and the
				// export status is read from the modifier list.
				isExported := module.IsExportedByName(statement.Modifiers())

				for _, declarationNode := range statement.DeclarationList.AsVariableDeclarationList().Declarations.Nodes {
					declaration := declarationNode.AsVariableDeclaration()
					name := declaration.Name()
					if name == nil || name.Kind != ast.KindIdentifier {
						continue
					}
					if !react.IsLikelyComponentName(name.Text()) {
						continue
					}
					initializer := declaration.Initializer
					if initializer == nil {
						continue
					}
					if isForwardRefCall(initializer) {
						continue
					}
					// A function expression or an arrow, and nothing else. A capitalized const
					// holding an object, a string or a call is not a component declaration.
					if initializer.Kind != ast.KindFunctionExpression && initializer.Kind != ast.KindArrowFunction {
						continue
					}
					ctx.ReportNode(name, messageNoConstAssignment(name.Text(), isExported))
				}
			},
		}
	},
}
