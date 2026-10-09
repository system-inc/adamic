package base

import (
	"regexp"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// consistencyRequirePaginationArgumentNameText is the rule's message, whose wording lives in
// `policy/messages/consistency-require-pagination-argument-name.json`.
var consistencyRequirePaginationArgumentNameText = policy.MessageOf("base/consistency-require-pagination-argument-name", "invalidName")

func messagePaginationDecoratorInvalidName() rule.Message {
	return rule.Message{
		Id:          consistencyRequirePaginationArgumentNameText.Id,
		Description: consistencyRequirePaginationArgumentNameText.Render(nil),
	}
}

// paginationInputTypeName matches the base type and every `@PaginationInputFor` subclass.
//
// The suffix is the whole test, which is the original's decision rather than a shortcut: a subclass
// is generated per paginated type and there is no list of them to check against, so the name is what
// says a type is a paginator.
var paginationInputTypeName = regexp.MustCompile(`PaginationInput$`)

// paginationArgumentName matches the disambiguating form, `prefixPagination`.
var paginationArgumentName = regexp.MustCompile(`Pagination$`)

// ConsistencyRequirePaginationArgumentName requires a GraphQlArgument of PaginationInput type to be named for what it is.
//
//	valid:   m(@GraphQlArgument(() => UserPaginationInput) pagination: UserPaginationInput)
//	valid:   m(@GraphQlArgument('postsPagination') p: PaginationInput)
//	valid:   m(@GraphQlArgument('items') items: string)
//	invalid: m(@GraphQlArgument('bad') p: PaginationInput)
//	invalid: m(@GraphQlArgument(() => PaginationInput) items: PaginationInput)
//
// # A base rule rather than an upstream port, so the source is the spec
//
// This comes from `api-phi-health/libraries/base/code-quality/lint/rules/ConsistencyRequirePaginationArgumentNameRule.ts`
// and there is no upstream corpus to import. That inverts the usual fixture advice: the cases below
// are written rather than imported, so they carry the same beliefs as this file and cannot catch a
// wrong belief on their own. The mitigation is the source repository, which is a real oracle: this
// rule was run over `api-phi-health` and compared against what the existing rule reports there. What
// that comparison found is recorded at the bottom of this comment.
//
// # A PaginationInput argument is recognized two ways
//
// An explicit type thunk on the decorator, `@GraphQlArgument(() => XPaginationInput)`, or the
// parameter's own type annotation, which is what the inferred-type and options-only forms rely on.
// Either is enough, matching the original's `||`. Both match by the `PaginationInput$` suffix.
//
// The thunk's body must be a bare identifier. `() => Foo.Bar` is a property access and `() => {...}`
// is a block, and neither is matched. That is the original's shape test reproduced, and it is also
// what keeps this rule away from a node whose `Text()` panics: measured, an arrow body here can be
// `KindPropertyAccessExpression`, and reading text off one takes down every rule in the package for
// that whole file. The kind is checked before any text is read.
//
// # The name is the decorator's first string literal, or the parameter's own
//
// `@GraphQlArgument('postsPagination') p: PaginationInput` is named by the literal, and
// `@GraphQlArgument() pagination: PaginationInput` by the parameter. The finding points at whichever
// one supplied the name, which is the token the author has to edit.
//
// # Where this port is simpler than the original, and why that is not a shortcut
//
// The original walks from the decorator to its parameter through `resolveDecoratedParameterNode`,
// thirty lines handling three ESTree shapes -- a plain identifier, an `AssignmentPattern` for a
// defaulted parameter, and a `TSParameterProperty` for a constructor parameter property -- plus a
// membership test against the enclosing function's `params` array so an identifier inside the
// decorator's own arguments is never mistaken for the parameter.
//
// Our parser needs none of it. Measured across all three shapes plus a parameter property: the
// decorator's parent is `KindParameter` every time, and the defaulted and parameter-property forms
// differ only in fields hanging off that same node. So the walk is `node.Parent` with a kind guard,
// and the membership test has nothing to protect against, since a decorator's own arguments are not
// its parent. This is fidelity to the decision rather than to the workaround; the original's walk
// exists for a parser difference we do not have.
//
// The kind guard IS load bearing, and it is the one thing the simplification must keep. A decorator
// can sit on a class, a method, a property or an accessor, and all four were measured reaching this
// listener with a parent of that kind rather than `KindParameter`.
//
// # No fix, matching the original
//
// The correct name is a naming judgment. `pagination` is right when there is one paginator and wrong
// when there are several, and the rule cannot tell which query it is looking at.
var ConsistencyRequirePaginationArgumentName = rule.Rule{
	Name: "base/consistency-require-pagination-argument-name",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindDecorator: func(node *ast.Node) {
				checkPaginationDecorator(ctx, node)
			},
		}
	},
}

// checkPaginationDecorator judges one decorator.
func checkPaginationDecorator(ctx rule.Context, node *ast.Node) {
	decorator := node.AsDecorator()
	if decorator == nil || decorator.Expression == nil ||
		decorator.Expression.Kind != ast.KindCallExpression {
		return
	}
	call := decorator.Expression.AsCallExpression()
	// A bare identifier callee only. `@Namespace.GraphQlArgument(...)` is a property access and the
	// original does not match it either, so the kind is checked before the text is read.
	if call.Expression == nil || call.Expression.Kind != ast.KindIdentifier ||
		call.Expression.Text() != "GraphQlArgument" {
		return
	}

	// The decorated parameter. See the doc above: our parser puts it directly at the decorator's
	// parent for every shape the original walks for, and the kind guard is what makes that safe,
	// since a decorator on a class or a method reaches here too.
	parent := node.Parent
	if parent == nil || parent.Kind != ast.KindParameter {
		return
	}
	parameter := parent.AsParameterDeclaration()
	if parameter == nil {
		return
	}
	parameterName := parameter.Name()
	if parameterName == nil || parameterName.Kind != ast.KindIdentifier {
		return
	}

	arguments := []*ast.Node{}
	if call.Arguments != nil {
		arguments = call.Arguments.Nodes
	}

	if !isAPaginationArgument(arguments, parameter) {
		return
	}

	// The name is the first string literal when the decorator gives one, otherwise the parameter's.
	// The node reported is whichever supplied it, so the finding lands on the token to edit.
	name := parameterName.Text()
	reported := parameterName
	if literal := firstStringLiteralArgument(arguments); literal != nil {
		name = literal.Text()
		reported = literal
	}

	if name == "pagination" || paginationArgumentName.MatchString(name) {
		return
	}
	ctx.ReportNode(reported, messagePaginationDecoratorInvalidName())
}

// isAPaginationArgument says whether a decorated parameter is a paginator.
//
// Either the decorator carries an explicit type thunk naming a PaginationInput type, or the
// parameter's own annotation is one. The original tests both with an `||` and so does this.
func isAPaginationArgument(arguments []*ast.Node, parameter *ast.ParameterDeclaration) bool {
	for _, argument := range arguments {
		if argument.Kind != ast.KindArrowFunction {
			continue
		}
		body := argument.AsArrowFunction().Body
		// A bare identifier body only. A property access or a block body is not matched by the
		// original, and reading `Text()` off a property access panics, which costs every rule in
		// this package every finding in the file. The kind is checked first for both reasons.
		if body != nil && body.Kind == ast.KindIdentifier &&
			paginationInputTypeName.MatchString(body.Text()) {
			return true
		}
	}
	return paginationInputTypeName.MatchString(typeReferenceName(parameter.Type))
}

// typeReferenceName reads the name of a bare type reference, or the empty string.
//
// The original returns null for anything that is not a `TSTypeReference` to an `Identifier`, so a
// keyword type, a generic instantiation's arguments and a qualified name all fall through. A
// qualified name is the case worth naming: `A.B` arrives as a `KindQualifiedName` rather than an
// identifier, and reading text off it is the panic this file avoids everywhere.
func typeReferenceName(typeNode *ast.Node) string {
	if typeNode == nil || typeNode.Kind != ast.KindTypeReference {
		return ""
	}
	name := typeNode.AsTypeReferenceNode().TypeName
	if name == nil || name.Kind != ast.KindIdentifier {
		return ""
	}
	return name.Text()
}

// firstStringLiteralArgument gives the decorator's first string argument, or nil.
//
// The original reads `decoratorArguments[0]` and checks it is a string literal, so a decorator whose
// first argument is a thunk or an options object supplies no name and the parameter's own is used.
// That ordering matters: `@GraphQlArgument(() => X, 'name')` takes the parameter's name upstream,
// not `'name'`, and this reproduces it by looking only at position zero.
func firstStringLiteralArgument(arguments []*ast.Node) *ast.Node {
	if len(arguments) == 0 {
		return nil
	}
	first := arguments[0]
	if first.Kind == ast.KindStringLiteral || first.Kind == ast.KindNoSubstitutionTemplateLiteral {
		return first
	}
	return nil
}
