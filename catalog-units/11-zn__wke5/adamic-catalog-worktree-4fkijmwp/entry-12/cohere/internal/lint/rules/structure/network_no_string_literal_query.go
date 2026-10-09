package structure

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// graphQlQueryMethods are the methods whose first argument must be a gql-tagged template.
//
// Matched on the property name alone, with no test of the receiver. That is the original's choice
// and it is deliberate: the sibling rules in this family test for the `networkService` binding
// because they are about hooks that file offers, while this one is about how a GraphQL document is
// written anywhere it is written. A wrapper, a re-export, or a differently-named service instance
// still takes a document, and a string literal is just as wrong there.
var graphQlQueryMethods = map[string]bool{
	"useGraphQlQuery":         true,
	"useGraphQlMutation":      true,
	"useSuspenseGraphQlQuery": true,
	"graphQlRequest":          true,
}

// networkNoStringLiteralQueryText is the rule's message, whose wording lives in
// `policy/messages/network-no-string-literal-query.json`.
var networkNoStringLiteralQueryText = policy.MessageOf("structure/network-no-string-literal-query", "noStringLiteralGraphQlQuery")

// messageNoStringLiteralGraphQlQuery is the finding, rendered when it is reported so the text comes from the current catalog.
func messageNoStringLiteralGraphQlQuery() rule.Message {
	return rule.Message{
		Id:          networkNoStringLiteralQueryText.Id,
		Description: networkNoStringLiteralQueryText.Render(nil),
	}
}

// NetworkNoStringLiteralQuery flags a GraphQL document passed as a string rather than a gql tag.
//
//	valid:   networkService.useGraphQlQuery(gql(`query { user { id } }`))
//	invalid: networkService.useGraphQlQuery('query { user { id } }')
//	invalid: networkService.useGraphQlQuery(`query { user { id } }`)
//	invalid: const Document = 'query { ... }'; networkService.useGraphQlQuery(Document)
//
// The third shape is why this rule resolves identifiers rather than testing the argument node.
// Hoisting a document into a `const` is the natural thing to do when it is reused, and it is
// precisely the case a syntactic check misses: the call site then reads as an identifier, which
// looks the same whether it holds a gql result or a string.
//
// The original resolves through ESLint's scope chain. Reproduced here by walking enclosing scopes
// for a variable declaration of that name, which is the same search rather than a type query: this
// rule has to work on a file that does not type-check, since a wrong document is exactly the kind of
// mistake that arrives with other errors.
var NetworkNoStringLiteralQuery = rule.Rule{
	Name: "structure/network-no-string-literal-query",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				if !isGraphQlMethodCall(call) {
					return
				}
				if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
					return
				}

				firstArgument := call.Arguments.Nodes[0]
				if resolvesToStringLiteral(firstArgument, map[*ast.Node]bool{}) {
					ctx.ReportNode(firstArgument, messageNoStringLiteralGraphQlQuery())
				}
			},
		}
	},
}

// isGraphQlMethodCall reports `<anything>.useGraphQlQuery(...)` and its three siblings.
func isGraphQlMethodCall(call *ast.CallExpression) bool {
	callee := ast.SkipParentheses(call.Expression)
	if callee == nil || callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	method := callee.AsPropertyAccessExpression().Name()
	return method != nil && method.Kind == ast.KindIdentifier && graphQlQueryMethods[method.Text()]
}

// resolvesToStringLiteral reports whether a node is, or resolves to, a string or template literal.
//
// The visited set breaks the cycle in `const a = b; const b = a`, which is not valid code but is
// reachable while someone is mid-edit, and a linter that hangs on a half-written file is worse than
// one that misses a finding on it.
func resolvesToStringLiteral(node *ast.Node, visited map[*ast.Node]bool) bool {
	node = ast.SkipParentheses(node)
	if node == nil || visited[node] {
		return false
	}
	visited[node] = true

	switch node.Kind {
	case ast.KindStringLiteral, ast.KindTemplateExpression, ast.KindNoSubstitutionTemplateLiteral:
		return true

	case ast.KindAsExpression:
		return resolvesToStringLiteral(node.AsAsExpression().Expression, visited)

	case ast.KindTypeAssertionExpression:
		return resolvesToStringLiteral(node.AsTypeAssertion().Expression, visited)

	case ast.KindIdentifier:
		initializer := variableInitializerInScope(node, node.Text())
		if initializer == nil {
			return false
		}
		return resolvesToStringLiteral(initializer, visited)
	}

	return false
}

// variableInitializerInScope finds the initializer of the nearest `const name = ...` binding.
//
// Walks outward from the use rather than downward from the file, so an inner binding shadows an
// outer one the way the language does. Only variable declarations are considered, matching the
// original: a parameter or an import has no initializer to read, and a rule that guessed at one
// would report on a document it never saw.
func variableInitializerInScope(reference *ast.Node, name string) *ast.Node {
	for scope := reference.Parent; scope != nil; scope = scope.Parent {
		var found *ast.Node

		scope.ForEachChild(func(statement *ast.Node) bool {
			if statement.Kind != ast.KindVariableStatement {
				return false
			}
			declarationList := statement.AsVariableStatement().DeclarationList
			if declarationList == nil {
				return false
			}
			for _, declarationNode := range declarationList.AsVariableDeclarationList().Declarations.Nodes {
				declaration := declarationNode.AsVariableDeclaration()
				declaredName := declaration.Name()
				if declaredName != nil && declaredName.Kind == ast.KindIdentifier &&
					declaredName.Text() == name && declaration.Initializer != nil {
					found = declaration.Initializer
					return true
				}
			}
			return false
		})

		if found != nil {
			return found
		}
	}
	return nil
}
