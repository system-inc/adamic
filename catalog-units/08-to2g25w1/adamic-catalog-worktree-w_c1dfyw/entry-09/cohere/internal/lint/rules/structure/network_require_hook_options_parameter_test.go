package structure

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const optionsDeclarations = "import { networkService } from './NetworkService.ts';\n" +
	"declare const Document: unknown;\n" +
	"type InferUseGraphQlQueryOptions<T> = { enabled?: boolean; document?: T };\n" +
	"type InferUseGraphQlMutationOptions<T> = { onSuccess?: () => void; document?: T };\n" +
	"type OtherOptionsType = { enabled?: boolean };\n" +
	"type UserQueryVariablesType = { identifier: string };\n"

const optionsFile = "/repository/source/api/UserRequest.ts"

func TestNetworkRequireHookOptionsParameterFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{
			"a query hook with no parameters at all",
			"export function useUserRequest() {\n    return networkService.useGraphQlQuery(Document);\n}\n",
			[]string{"missingOptionsParameter"},
		},
		{
			"a query hook whose only parameter is variables",
			"export function useUserRequest(variables: UserQueryVariablesType) {\n" +
				"    return networkService.useGraphQlQuery(Document, variables);\n}\n",
			[]string{"missingOptionsParameter"},
		},
		{
			"a mutation hook with no options parameter",
			"export function useSaveRequest() {\n    return networkService.useGraphQlMutation(Document);\n}\n",
			[]string{"missingOptionsParameter"},
		},
		{
			"an arrow hook with no options parameter",
			"export const useUserRequest = () => {\n    return networkService.useGraphQlQuery(Document);\n};\n",
			[]string{"missingOptionsParameter"},
		},
		// The signature promises control the call silently drops, which is the worse of the two
		// failures and why it has its own id.
		{
			"a query hook that accepts options and does not forward them",
			"export function useUserRequest(options?: InferUseGraphQlQueryOptions<typeof Document>) {\n" +
				"    return networkService.useGraphQlQuery(Document);\n}\n",
			[]string{"optionsParameterNotPassedThrough"},
		},
		{
			"a mutation hook that accepts options and does not forward them",
			"export function useSaveRequest(options?: InferUseGraphQlMutationOptions<typeof Document>) {\n" +
				"    return networkService.useGraphQlMutation(Document);\n}\n",
			[]string{"optionsParameterNotPassedThrough"},
		},
		// Forwarded at the wrong position. The option reaches the call and lands in the variables
		// slot, so it is dropped exactly as if it were never passed.
		{
			"options forwarded in the variables position of a three-argument query",
			"export function useUserRequest(variables: UserQueryVariablesType, options?: InferUseGraphQlQueryOptions<typeof Document>) {\n" +
				"    return networkService.useGraphQlQuery(Document, options, variables);\n}\n",
			[]string{"optionsParameterNotPassedThrough"},
		},
		{
			"an options parameter typed with something else",
			"export function useUserRequest(options?: OtherOptionsType) {\n" +
				"    return networkService.useGraphQlQuery(Document, undefined, options);\n}\n",
			[]string{"incorrectOptionsParameterType"},
		},
		{
			"an untyped options parameter",
			"export function useUserRequest(options?: unknown) {\n" +
				"    return networkService.useGraphQlQuery(Document, undefined, options);\n}\n",
			[]string{"incorrectOptionsParameterType"},
		},
		// A qualified type name is not accepted. Matching the last segment would let an import from
		// the wrong place satisfy a rule that exists to pin one specific generated type.
		{
			"an options parameter typed with a qualified name",
			"declare namespace Wrong { type InferUseGraphQlQueryOptions<T> = { document?: T } }\n" +
				"export function useUserRequest(options?: Wrong.InferUseGraphQlQueryOptions<typeof Document>) {\n" +
				"    return networkService.useGraphQlQuery(Document, undefined, options);\n}\n",
			[]string{"incorrectOptionsParameterType"},
		},
		// Both defects at once: wrong type and never forwarded.
		{
			"a wrongly typed options parameter that is also not forwarded",
			"export function useUserRequest(options?: OtherOptionsType) {\n" +
				"    return networkService.useGraphQlQuery(Document);\n}\n",
			[]string{"incorrectOptionsParameterType", "optionsParameterNotPassedThrough"},
		},
		// Judged per call, not per hook. Two calls each missing options is two findings, and this
		// case fails if the rule collapses a hook to one verdict.
		{
			"a hook making two calls, neither given options",
			"export function useUserRequest() {\n" +
				"    networkService.useGraphQlQuery(Document);\n" +
				"    return networkService.useGraphQlMutation(Document);\n}\n",
			[]string{"missingOptionsParameter", "missingOptionsParameter"},
		},
		// A destructured last parameter has no single name, so it is not the options parameter and
		// the hook reads as having none.
		{
			"a destructured last parameter",
			"export function useUserRequest({ options }: { options?: InferUseGraphQlQueryOptions<typeof Document> }) {\n" +
				"    return networkService.useGraphQlQuery(Document, undefined, options);\n}\n",
			[]string{"missingOptionsParameter"},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, NetworkRequireHookOptionsParameter, optionsFile,
				optionsDeclarations+testCase.sourceText), testCase.wantIds...)
		})
	}
}

func TestNetworkRequireHookOptionsParameterStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// The shape the rule asks for, in the three-argument query form.
		{
			"a query forwarding options third with variables second",
			"export function useUserRequest(variables: UserQueryVariablesType, options?: InferUseGraphQlQueryOptions<typeof Document>) {\n" +
				"    return networkService.useGraphQlQuery(Document, variables, options);\n}\n",
		},
		{
			"a query forwarding options third with undefined second",
			"export function useUserRequest(options?: InferUseGraphQlQueryOptions<typeof Document>) {\n" +
				"    return networkService.useGraphQlQuery(Document, undefined, options);\n}\n",
		},
		// The position derivation's other branch: a query whose second argument already looks like
		// options is satisfied at position two. This is the case a fixed "options are third" rule
		// would report wrongly, so it is the boundary the derivation exists for.
		{
			"a query forwarding options second, with no variables",
			"export function useUserRequest(options?: InferUseGraphQlQueryOptions<typeof Document>) {\n" +
				"    return networkService.useGraphQlQuery(Document, options);\n}\n",
		},
		{
			"a mutation forwarding options second",
			"export function useSaveRequest(options?: InferUseGraphQlMutationOptions<typeof Document>) {\n" +
				"    return networkService.useGraphQlMutation(Document, options);\n}\n",
		},
		// A hook adding its own defaults forwards everything the caller supplied. A rule accepting
		// only a bare identifier would report this correct code.
		{
			"options forwarded by spread into an object literal",
			"export function useUserRequest(options?: InferUseGraphQlQueryOptions<typeof Document>) {\n" +
				"    return networkService.useGraphQlQuery(Document, undefined, { ...options, enabled: true });\n}\n",
		},
		{
			"a differently named options parameter",
			"export function useUserRequest(queryOptions?: InferUseGraphQlQueryOptions<typeof Document>) {\n" +
				"    return networkService.useGraphQlQuery(Document, undefined, queryOptions);\n}\n",
		},
		{
			"an arrow hook forwarding options correctly",
			"export const useUserRequest = (options?: InferUseGraphQlQueryOptions<typeof Document>) => {\n" +
				"    return networkService.useGraphQlQuery(Document, undefined, options);\n};\n",
		},
		// graphQlRequest is not a hook call and takes no options object, so it imposes nothing. A
		// rule that judged all four methods would report every use of it.
		{
			"a hook calling graphQlRequest with no options",
			"export function useUserRequest() {\n    return networkService.graphQlRequest(Document);\n}\n",
		},
		{
			"a hook that never calls NetworkService",
			"export function useUserRequest() {\n    return 1;\n}\n",
		},
		{
			"a file with no hooks",
			"export const value = 1;\n",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NetworkRequireHookOptionsParameter, optionsFile,
				optionsDeclarations+testCase.sourceText))
		})
	}
}
