package structure

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const variablesDeclarations = "import { networkService } from './NetworkService.ts';\n" +
	"declare const Document: unknown;\n" +
	"type UserQueryVariablesType = { identifier: string; slug: string };\n" +
	"type SaveMutationVariablesType = { input: string };\n" +
	"type InferUseGraphQlQueryOptions<T> = { enabled?: boolean; document?: T };\n" +
	"type UserProfileType = { displayName: string };\n"

const variablesFile = "/repository/source/api/UserRequest.ts"

func TestNetworkRequireHookVariablesTypeFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{
			"a variables parameter typed inline",
			"export function useUserRequest(variables: { identifier: string }) {\n" +
				"    return networkService.useGraphQlQuery(Document, variables);\n}\n",
			[]string{"inlineVariablesType"},
		},
		{
			"a variables parameter typed with an unrelated named type",
			"export function useUserRequest(variables: UserProfileType) {\n" +
				"    return networkService.useGraphQlQuery(Document, variables);\n}\n",
			[]string{"inlineVariablesType"},
		},
		{
			"a parameter named input typed inline",
			"export function useSaveRequest(input: { input: string }) {\n" +
				"    return networkService.useGraphQlMutation(Document, input);\n}\n",
			[]string{"inlineVariablesType"},
		},
		// The second arm, and the case people actually write: a parameter named for the thing it
		// fetches rather than for being variables, whose inline type gives it away.
		{
			"a differently named parameter whose inline type holds identifier",
			"export function useUserRequest(user: { identifier: string }) {\n" +
				"    return networkService.useGraphQlQuery(Document, user);\n}\n",
			[]string{"inlineVariablesType"},
		},
		{
			"a differently named parameter whose inline type holds pagination",
			"export function useUserRequest(page: { pagination: string }) {\n" +
				"    return networkService.useGraphQlQuery(Document, page);\n}\n",
			[]string{"inlineVariablesType"},
		},
		{
			"the rebuild, property for property",
			"export function useUserRequest(variables: UserQueryVariablesType) {\n" +
				"    return networkService.useGraphQlQuery(Document, { identifier: variables.identifier, slug: variables.slug });\n}\n",
			[]string{"unnecessaryVariablesDestructuring"},
		},
		{
			"a single-property rebuild",
			"export function useUserRequest(variables: UserQueryVariablesType) {\n" +
				"    return networkService.useGraphQlQuery(Document, { identifier: variables.identifier });\n}\n",
			[]string{"unnecessaryVariablesDestructuring"},
		},
		// Both defects at once, which is the shape that actually appears: someone writes the type
		// by hand and rebuilds the object in the same breath.
		{
			"an inline type and a rebuild together",
			"export function useUserRequest(variables: { identifier: string }) {\n" +
				"    return networkService.useGraphQlQuery(Document, { identifier: variables.identifier });\n}\n",
			[]string{"inlineVariablesType", "unnecessaryVariablesDestructuring"},
		},
		{
			"a suspense query is judged like a query",
			"export function useUserRequest(variables: { identifier: string }) {\n" +
				"    return networkService.useSuspenseGraphQlQuery(Document, variables);\n}\n",
			[]string{"inlineVariablesType"},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, NetworkRequireHookVariablesType, variablesFile,
				variablesDeclarations+testCase.sourceText), testCase.wantIds...)
		})
	}
}

func TestNetworkRequireHookVariablesTypeStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// The shape the rule asks for, in both directions.
		{
			"a generated query variables type passed through",
			"export function useUserRequest(variables: UserQueryVariablesType) {\n" +
				"    return networkService.useGraphQlQuery(Document, variables);\n}\n",
		},
		{
			"a generated mutation variables type passed through",
			"export function useSaveRequest(variables: SaveMutationVariablesType) {\n" +
				"    return networkService.useGraphQlMutation(Document, variables);\n}\n",
		},
		// The options parameter belongs to another rule. Judging it here would report the generated
		// options type as a wrong variables type, so this fails if the exemption is dropped.
		{
			"an options parameter is not judged as variables",
			"export function useUserRequest(options?: InferUseGraphQlQueryOptions<typeof Document>) {\n" +
				"    return networkService.useGraphQlQuery(Document, undefined, options);\n}\n",
		},
		// The options exemption's own boundary, and it needs an inline type to reach. A parameter
		// named options carrying a named type is silent whether or not the exemption exists, since
		// the fallback arm only judges inline types, so the case above measures nothing about it.
		// Found by dropping the exemption and watching the suite stay green.
		//
		// An inline options type holding `id` is what a hand-written option shape actually looks
		// like, and it is exactly what the second arm would flag as a variables object.
		{
			"an options parameter typed inline is still not judged as variables",
			"export function useUserRequest(options?: { id?: string; enabled?: boolean }) {\n" +
				"    return networkService.useGraphQlQuery(Document, undefined, options);\n}\n",
		},
		{
			"a parameter named queryOptions typed inline is exempt too",
			"export function useUserRequest(queryOptions?: { pagination?: string }) {\n" +
				"    return networkService.useGraphQlQuery(Document, undefined, queryOptions);\n}\n",
		},
		// The second arm's boundary. Same shape as a firing case, differing only in whether the
		// inline type carries a telltale property, so this fails if the property list is widened
		// to "any inline type".
		{
			"a differently named parameter whose inline type has no GraphQL property",
			"export function useUserRequest(user: { displayName: string }) {\n" +
				"    return networkService.useGraphQlQuery(Document, user);\n}\n",
		},
		{
			"a differently named parameter with an unrelated named type",
			"export function useUserRequest(user: UserProfileType) {\n" +
				"    return networkService.useGraphQlQuery(Document, user);\n}\n",
		},
		{
			"an unannotated variables parameter",
			"export function useUserRequest(variables) {\n" +
				"    return networkService.useGraphQlQuery(Document, variables);\n}\n",
		},
		// A rename is not a rebuild. Collapsing this to the parameter would change what is sent, so
		// the rule must leave it alone.
		{
			"a renaming object literal is not a rebuild",
			"export function useUserRequest(variables: UserQueryVariablesType) {\n" +
				"    return networkService.useGraphQlQuery(Document, { identifier: variables.slug });\n}\n",
		},
		// A spread plus an addition is the parameter plus something, not a copy of it.
		// A bare spread is a pass-through written the long way, not a property-by-property rebuild,
		// and the rule must not claim it rebuilds anything. Found by making a spread count as a
		// rebuild property and watching the suite stay green: every other spread fixture here adds
		// something, so nothing measured the spread-only shape.
		{
			"a bare spread of the variables parameter is not a rebuild",
			"export function useUserRequest(variables: UserQueryVariablesType) {\n" +
				"    return networkService.useGraphQlQuery(Document, { ...variables });\n}\n",
		},
		{
			"a spread with an addition is not a rebuild",
			"export function useUserRequest(variables: UserQueryVariablesType) {\n" +
				"    return networkService.useGraphQlQuery(Document, { ...variables, slug: 'x' });\n}\n",
		},
		{
			"a literal that reads off something other than the variables parameter",
			"declare const other: { identifier: string };\n" +
				"export function useUserRequest(variables: UserQueryVariablesType) {\n" +
				"    return networkService.useGraphQlQuery(Document, { identifier: other.identifier });\n}\n",
		},
		// The rebuild check is queries only, because a mutation's second argument is not variables.
		{
			"a mutation is not checked for a rebuild",
			"export function useSaveRequest(variables: SaveMutationVariablesType) {\n" +
				"    return networkService.useGraphQlMutation(Document, { input: variables.input });\n}\n",
		},
		{
			"graphQlRequest imposes nothing",
			"export function useUserRequest(variables: { identifier: string }) {\n" +
				"    return networkService.graphQlRequest(Document, variables);\n}\n",
		},
		{
			"a hook that never calls NetworkService",
			"export function useUserRequest(variables: { identifier: string }) {\n    return variables;\n}\n",
		},
		{
			"a file with no hooks",
			"export const value = 1;\n",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NetworkRequireHookVariablesType, variablesFile,
				variablesDeclarations+testCase.sourceText))
		})
	}
}
