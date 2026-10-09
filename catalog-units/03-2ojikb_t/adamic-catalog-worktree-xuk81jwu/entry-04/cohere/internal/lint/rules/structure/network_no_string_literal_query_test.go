package structure

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const queryFile = "/repository/source/api/UserRequest.ts"

const queryDeclarations = "declare const networkService: {\n" +
	"    useGraphQlQuery(document: unknown, variables?: unknown): unknown;\n" +
	"    useGraphQlMutation(document: unknown): unknown;\n" +
	"    useSuspenseGraphQlQuery(document: unknown): unknown;\n" +
	"    graphQlRequest(document: unknown): unknown;\n" +
	"    other(document: unknown): unknown;\n" +
	"};\n" +
	"declare function gql(document: unknown): unknown;\n" +
	"declare const identifier: string;\n"

func TestNetworkNoStringLiteralQueryFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{
			"a plain string document",
			"export function run() {\n    networkService.useGraphQlQuery('query { user { id } }');\n}\n",
		},
		// A template with no substitutions is a string wearing backticks and reaches a different
		// node kind, so it needs its own case or the rule is silent on the shape people actually
		// write a multi-line document in.
		{
			"a template with no substitutions",
			"export function run() {\n    networkService.useGraphQlQuery(`query { user { id } }`);\n}\n",
		},
		{
			"a template with interpolation",
			"export function run() {\n    networkService.useGraphQlQuery(`query { user(id: \"${identifier}\") { id } }`);\n}\n",
		},
		// The reason the rule resolves identifiers at all. Hoisting a reused document into a const
		// is the natural refactor, and it is exactly what turns the call site into an identifier
		// that a syntactic check reads as fine.
		{
			"an identifier holding a string",
			"const UserDocument = 'query { user { id } }';\nexport function run() {\n    networkService.useGraphQlQuery(UserDocument);\n}\n",
		},
		{
			"an identifier holding a template",
			"const UserDocument = `query { user { id } }`;\nexport function run() {\n    networkService.useGraphQlQuery(UserDocument);\n}\n",
		},
		{
			"an identifier resolving through a second identifier",
			"const Inner = 'query { user { id } }';\nconst UserDocument = Inner;\nexport function run() {\n    networkService.useGraphQlQuery(UserDocument);\n}\n",
		},
		{
			"a string behind an as-expression",
			"export function run() {\n    networkService.useGraphQlQuery('query { user { id } }' as string);\n}\n",
		},
		// All four methods take a document, so all four are checked. Testing one and trusting the
		// map is how a rule ends up enforcing a quarter of what its name claims.
		{
			"useGraphQlMutation",
			"export function run() {\n    networkService.useGraphQlMutation('mutation { save { id } }');\n}\n",
		},
		{
			"useSuspenseGraphQlQuery",
			"export function run() {\n    networkService.useSuspenseGraphQlQuery('query { user { id } }');\n}\n",
		},
		{
			"graphQlRequest",
			"export function run() {\n    networkService.graphQlRequest('query { user { id } }');\n}\n",
		},
		// The receiver is deliberately not judged, so a wrapper or a differently-named instance is
		// still checked. This is the boundary against its sibling rules, which do test the name.
		{
			"a receiver that is not networkService",
			"declare const client: { useGraphQlQuery(document: unknown): unknown };\nexport function run() {\n    client.useGraphQlQuery('query { user { id } }');\n}\n",
		},
		{
			"an inner binding shadowing an outer gql document",
			"const UserDocument = gql(`query { user { id } }`);\nexport function run() {\n    const UserDocument = 'query { user { id } }';\n    networkService.useGraphQlQuery(UserDocument);\n}\n",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, NetworkNoStringLiteralQuery, queryFile,
				queryDeclarations+testCase.sourceText), "noStringLiteralGraphQlQuery")
		})
	}
}

func TestNetworkNoStringLiteralQueryStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// The point of the rule: this is the shape it is asking for, and it must not fire on it.
		{
			"a gql tagged document",
			"export function run() {\n    networkService.useGraphQlQuery(gql(`query { user { id } }`));\n}\n",
		},
		// The identifier arm's own boundary. Same call shape as the firing case, differing only in
		// what the const holds, so this case fails if the arm stops resolving and starts guessing.
		{
			"an identifier holding a gql result",
			"const UserDocument = gql(`query { user { id } }`);\nexport function run() {\n    networkService.useGraphQlQuery(UserDocument);\n}\n",
		},
		{
			"an inner binding shadowing an outer string",
			"const UserDocument = 'query { user { id } }';\nexport function run() {\n    const UserDocument = gql(`query { user { id } }`);\n    networkService.useGraphQlQuery(UserDocument);\n}\n",
		},
		{
			"an identifier with no binding in the file",
			"declare const ImportedDocument: unknown;\nexport function run() {\n    networkService.useGraphQlQuery(ImportedDocument);\n}\n",
		},
		{
			"a parameter, which has no initializer to read",
			"export function run(document: unknown) {\n    networkService.useGraphQlQuery(document);\n}\n",
		},
		// Only the first argument is the document. Variables are objects and may legitimately hold
		// strings, so a rule that checked every argument would fire on correct code constantly.
		{
			"a string in the variables argument",
			"export function run() {\n    networkService.useGraphQlQuery(gql(`query { user { id } }`), { identifier: 'abc' });\n}\n",
		},
		{
			"a method that does not take a document",
			"export function run() {\n    networkService.other('query { user { id } }');\n}\n",
		},
		{
			"a bare call that is not a member access",
			"declare function useGraphQlQuery(document: unknown): unknown;\nexport function run() {\n    useGraphQlQuery('query { user { id } }');\n}\n",
		},
		{
			"a call with no arguments",
			"export function run() {\n    networkService.graphQlRequest();\n}\n",
		},
		{
			"a cycle between two bindings does not hang",
			"declare const seed: unknown;\nexport function run() {\n    const first: unknown = seed;\n    networkService.useGraphQlQuery(first);\n}\n",
		},
		{"no calls at all", "export const Value = 1;\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NetworkNoStringLiteralQuery, queryFile,
				queryDeclarations+testCase.sourceText))
		})
	}
}
