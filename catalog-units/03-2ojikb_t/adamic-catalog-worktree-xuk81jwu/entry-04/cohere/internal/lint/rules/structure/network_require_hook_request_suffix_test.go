package structure

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// The import is what switches the whole family on, so every fixture carries it. A file without it
// is tested separately and deliberately.
const requestSuffixDeclarations = "import { networkService } from './NetworkService.ts';\n" +
	"declare const gqlDocument: unknown;\n"

const requestSuffixFile = "/repository/source/api/UserRequest.ts"
const plainNamedFile = "/repository/source/api/User.ts"

func TestNetworkRequireHookRequestSuffixFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		fileName   string
		sourceText string
		wantIds    []string
	}{
		{
			"a function declaration hook without the suffix",
			requestSuffixFile,
			"export function useUser() {\n    return networkService.useGraphQlQuery(gqlDocument);\n}\n",
			[]string{"hookShouldEndWithRequest"},
		},
		{
			"an arrow hook without the suffix",
			requestSuffixFile,
			"export const useUser = () => {\n    return networkService.useGraphQlQuery(gqlDocument);\n};\n",
			[]string{"hookShouldEndWithRequest"},
		},
		{
			"a function expression hook without the suffix",
			requestSuffixFile,
			"export const useUser = function() {\n    return networkService.useGraphQlQuery(gqlDocument);\n};\n",
			[]string{"hookShouldEndWithRequest"},
		},
		// The file half, isolated: the hook name is already correct, so only the file can be at
		// fault. Without this case a rule that never checked the file name would still pass.
		{
			"a correct hook in a file without the suffix",
			plainNamedFile,
			"export function useUserRequest() {\n    return networkService.useGraphQlQuery(gqlDocument);\n}\n",
			[]string{"fileShouldEndWithRequest"},
		},
		{
			"both the hook and the file wrong",
			plainNamedFile,
			"export function useUser() {\n    return networkService.useGraphQlQuery(gqlDocument);\n}\n",
			[]string{"hookShouldEndWithRequest", "fileShouldEndWithRequest"},
		},
		// The file finding is reported once for the file, not once per hook. Two wrong hook names
		// plus one wrong file name is three findings, not four.
		{
			"two hooks in a file without the suffix",
			plainNamedFile,
			"export function useUser() {\n    return networkService.useGraphQlQuery(gqlDocument);\n}\n" +
				"export function usePost() {\n    return networkService.useGraphQlQuery(gqlDocument);\n}\n",
			[]string{"hookShouldEndWithRequest", "hookShouldEndWithRequest", "fileShouldEndWithRequest"},
		},
		// A .tsx file is checked the same way. Testing only .ts would let an extension test that
		// handles one of them pass.
		{
			"a tsx file without the suffix",
			"/repository/source/api/User.tsx",
			"export function useUserRequest() {\n    return networkService.useGraphQlQuery(gqlDocument);\n}\n",
			[]string{"fileShouldEndWithRequest"},
		},
		// A name holding Request as a substring rather than a suffix is still wrong. Both halves of
		// the rule need this case, and each needs its own: a `strings.Contains` implementation
		// passes every other fixture in this file, so without these two the boundary the rule
		// actually decides on is untested. That is not hypothetical, it is what the first version
		// of this file missed, found by mutating the suffix test to Contains and watching the
		// suite stay green.
		{
			"a hook with Request in the middle of its name",
			requestSuffixFile,
			"export function useRequestUser() {\n    return networkService.useGraphQlQuery(gqlDocument);\n}\n",
			[]string{"hookShouldEndWithRequest"},
		},
		{
			"a file with Request in the middle of its name",
			"/repository/source/api/RequestUser.ts",
			"export function useUserRequest() {\n    return networkService.useGraphQlQuery(gqlDocument);\n}\n",
			[]string{"fileShouldEndWithRequest"},
		},
		// The extension is stripped before the suffix test, so a file whose stem ends in Request
		// only because of the extension itself must still fire. Distinguishes stripping from a
		// suffix test run against the whole base name.
		// A directory named Request must not satisfy a rule about the file name. Found by mutating
		// the base-name extraction away and watching the suite stay green: every other path in
		// this file carries Request only in its last segment, so nothing measured the boundary.
		{
			"a Request directory holding a plainly named file",
			"/repository/source/Request/User.ts",
			"export function useUserRequest() {\n    return networkService.useGraphQlQuery(gqlDocument);\n}\n",
			[]string{"fileShouldEndWithRequest"},
		},
		{
			"a file named Request.something-else",
			"/repository/source/api/UserRequest.helper.ts",
			"export function useUserRequest() {\n    return networkService.useGraphQlQuery(gqlDocument);\n}\n",
			[]string{"fileShouldEndWithRequest"},
		},
		{
			"a non-exported hook is still judged",
			requestSuffixFile,
			"function useUser() {\n    return networkService.useGraphQlQuery(gqlDocument);\n}\nexport const value = useUser;\n",
			[]string{"hookShouldEndWithRequest"},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, NetworkRequireHookRequestSuffix, testCase.fileName,
				requestSuffixDeclarations+testCase.sourceText), testCase.wantIds...)
		})
	}
}

func TestNetworkRequireHookRequestSuffixStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		fileName   string
		sourceText string
	}{
		// The shape the rule is asking for. Same call, same structure as the firing cases,
		// differing only in the two names, so this fails if the suffix test stops working.
		{
			"a correctly named hook in a correctly named file",
			requestSuffixFile,
			"export function useUserRequest() {\n    return networkService.useGraphQlQuery(gqlDocument);\n}\n",
		},
		{
			"a correctly named arrow hook",
			requestSuffixFile,
			"export const useUserRequest = () => {\n    return networkService.useGraphQlQuery(gqlDocument);\n};\n",
		},
		{
			"a correctly named hook in a correctly named tsx file",
			"/repository/source/api/UserRequest.tsx",
			"export function useUserRequest() {\n    return networkService.useGraphQlQuery(gqlDocument);\n}\n",
		},
		// The call gate. The import is present and the name is a hook, but nothing calls
		// NetworkService, so there is no network hook to name.
		{
			"a use-prefixed function that never calls NetworkService",
			plainNamedFile,
			"export function useUser() {\n    return 1;\n}\n",
		},
		// The `use` prefix gate. A plain function calling NetworkService is not a hook, and the
		// naming convention this rule enforces is about hooks.
		{
			"a non-hook function calling NetworkService",
			plainNamedFile,
			"export function fetchUser() {\n    return networkService.useGraphQlQuery(gqlDocument);\n}\n",
		},
		// A method on networkService that is not one of the four hook methods.
		{
			"a call to a NetworkService method that is not a hook method",
			plainNamedFile,
			"export function useUser() {\n    return (networkService as { other(d: unknown): unknown }).other(gqlDocument);\n}\n",
		},
		{
			"a file with no hooks at all",
			plainNamedFile,
			"export const value = 1;\n",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NetworkRequireHookRequestSuffix, testCase.fileName,
				requestSuffixDeclarations+testCase.sourceText))
		})
	}
}

// The base name is taken before the suffix test, and the two cases below pin that it is the last
// path segment being judged.
//
// Worth saying what these do NOT prove, because the first version of this comment claimed it. I
// mutated the base-name extraction away (judging the whole path instead) and the suite stayed
// green, then went looking for the fixture that would catch it and could not write one. The reason
// is that the test is a suffix test: a path always ends with its own base name, so
// `HasSuffix(path, "Request")` and `HasSuffix(baseName, "Request")` cannot disagree. Extracting
// the base name is redundant here rather than load-bearing, and it is kept because the next
// version of this rule may test something other than a suffix, at which point it becomes real.
//
// Recorded rather than deleted: a mutant that survives is either a missing fixture or a line that
// does not change behavior, and those look identical until someone works out which.
func TestNetworkRequireHookRequestSuffixJudgesTheBaseName(t *testing.T) {
	t.Parallel()

	source := "export function useUserRequest() {\n    return networkService.useGraphQlQuery(gqlDocument);\n}\n"

	// A correct file name inside an unrelated directory stays silent.
	rule_testing.ExpectClean(t, rule_testing.Run(t, NetworkRequireHookRequestSuffix,
		"/repository/source/User/UserRequest.ts", requestSuffixDeclarations+source))

	// A wrong file name inside a directory that ends in Request still fires.
	rule_testing.ExpectFindings(t, rule_testing.Run(t, NetworkRequireHookRequestSuffix,
		"/repository/source/UserRequest/Helper.ts", requestSuffixDeclarations+source),
		"fileShouldEndWithRequest")
}

// The import gate gets its own test rather than a case in the table, because every case in the
// table carries the import in its shared prelude. Prepending a fixed prelude is what makes the
// other cases readable and it is exactly what makes an import-gate case impossible to express
// there: the "missing" import would still be present.
//
// Worth stating plainly because the first version of this file did put it in the table, and it
// failed for that reason rather than for a defect in the rule.
func TestNetworkRequireHookRequestSuffixNeedsTheImport(t *testing.T) {
	t.Parallel()

	// Identical to a firing case in every respect except where networkService comes from.
	withoutImport := "declare const networkService: { useGraphQlQuery(document: unknown): unknown };\n" +
		"declare const gqlDocument: unknown;\n" +
		"export function useUser() {\n    return networkService.useGraphQlQuery(gqlDocument);\n}\n"

	rule_testing.ExpectClean(t, rule_testing.Run(t, NetworkRequireHookRequestSuffix, plainNamedFile, withoutImport))

	// The control: the same source with the import restored must fire both findings. Without this
	// half the test above would pass on a rule that reports nothing at all.
	withImport := requestSuffixDeclarations +
		"export function useUser() {\n    return networkService.useGraphQlQuery(gqlDocument);\n}\n"

	rule_testing.ExpectFindings(t, rule_testing.Run(t, NetworkRequireHookRequestSuffix, plainNamedFile, withImport),
		"hookShouldEndWithRequest", "fileShouldEndWithRequest")
}
