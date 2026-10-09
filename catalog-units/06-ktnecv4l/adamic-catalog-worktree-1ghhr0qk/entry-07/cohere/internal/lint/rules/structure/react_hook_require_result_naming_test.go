package structure

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const resultNamingFile = "/repository/source/components/Panel.tsx"

const resultNamingDeclarations = "declare function useAccountQuery(): unknown;\n" +
	"declare function useWatch(): unknown;\n" +
	"declare function useSprings(): unknown;\n"

func TestReactHookRequireResultNamingFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantId     string
	}{
		// Generic names, which say only that something was fetched.
		{
			"a result named data",
			"export function Panel() {\n    const data = useAccountQuery();\n    return data;\n}\n",
			"hookResultNamingGeneric",
		},
		{
			"a result named result",
			"export function Panel() {\n    const result = useAccountQuery();\n    return result;\n}\n",
			"hookResultNamingGeneric",
		},
		{
			"a two-character name",
			"export function Panel() {\n    const ab = useAccountQuery();\n    return ab;\n}\n",
			"hookResultNamingGeneric",
		},
		// Suffixes that describe the variable rather than its contents.
		{
			"a result ending in Result",
			"export function Panel() {\n    const somethingResult = useAccountQuery();\n    return somethingResult;\n}\n",
			"hookResultNamingBadSuffix",
		},
		{
			"a result ending in State",
			"export function Panel() {\n    const somethingState = useAccountQuery();\n    return somethingState;\n}\n",
			"hookResultNamingBadSuffix",
		},
		// A name unrelated to the hook, and not generic or badly suffixed.
		{
			"an unrelated name",
			"export function Panel() {\n    const widget = useAccountQuery();\n    return widget;\n}\n",
			"hookResultNamingMismatch",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, ReactHookRequireResultNaming, resultNamingFile,
				resultNamingDeclarations+testCase.sourceText), testCase.wantId)
		})
	}
}

func TestReactHookRequireResultNamingStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		fileName   string
		sourceText string
	}{
		// The shape the rule asks for.
		{
			"the expected name",
			resultNamingFile,
			"export function Panel() {\n    const accountQuery = useAccountQuery();\n    return accountQuery;\n}\n",
		},
		// The acceptance test is loose on purpose: both of these are better names than the bare
		// expected one, and a narrow test would report improvements.
		{
			"a name starting with the expected one",
			resultNamingFile,
			"export function Panel() {\n    const watchEmail = useWatch();\n    return watchEmail;\n}\n",
		},
		{
			"a name ending with the expected one",
			resultNamingFile,
			"export function Panel() {\n    const emailWatch = useWatch();\n    return emailWatch;\n}\n",
		},
		// A plural hook accepts a name built on its singular. `lineSprings` does not measure that
		// arm, because it also ends with `springs` and the suffix check accepts it first. The
		// singular has to appear without the plural for the arm to be reached.
		{
			"a singular name for a plural hook",
			resultNamingFile,
			"export function Panel() {\n    const lineSprings = useSprings();\n    return lineSprings;\n}\n",
		},
		{
			"a singular name that is not also the plural",
			resultNamingFile,
			"export function Panel() {\n    const springLine = useSprings();\n    return springLine;\n}\n",
		},
		// The single-invocation gate, which is what keeps the rule from being noise. Two calls have
		// to be told apart by their names, so a prefix is exactly right there.
		{
			"two calls to the same hook",
			resultNamingFile,
			"export function Panel() {\n    const first = useAccountQuery();\n" +
				"    const second = useAccountQuery();\n    return [first, second];\n}\n",
		},
		// A destructured result has no single name to judge.
		{
			"a destructured result",
			resultNamingFile,
			"export function Panel() {\n    const { data } = useAccountQuery() as { data: unknown };\n    return data;\n}\n",
		},
		// A call that is not a hook. Two of them, so a rule counting every call would see the
		// second as a repeat invocation and silence the first, which is what the count gate does
		// for real hooks.
		{
			"two non-hook calls do not affect the count",
			resultNamingFile,
			"declare function fetchAccount(): unknown;\n" +
				"export function Panel() {\n    fetchAccount();\n" +
				"    const accountQuery = useAccountQuery();\n    return accountQuery;\n}\n",
		},
		{
			"a non-hook call",
			resultNamingFile,
			"declare function fetchAccount(): unknown;\n" +
				"export function Panel() {\n    const data = fetchAccount();\n    return data;\n}\n",
		},
		// Outside a React file the rule declines before looking.
		{
			"a hook result in a ts file",
			"/repository/source/api/Thing.ts",
			"export function run() {\n    const data = useAccountQuery();\n    return data;\n}\n",
		},
		{
			"a file with no hooks",
			resultNamingFile,
			"export const value = 1;\n",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, ReactHookRequireResultNaming, testCase.fileName,
				resultNamingDeclarations+testCase.sourceText))
		})
	}
}

// The message is chosen by an ordered test, not by combining complaints.
//
// A name can be both generic and badly suffixed. The suffix check runs first because it names a
// concrete edit, where the generic message only says the name is thin. Pinned so the order is a
// decision rather than an accident of how the conditions were written.
func TestReactHookRequireResultNamingPicksOneComplaint(t *testing.T) {
	t.Parallel()

	// `dataResult` is generic-adjacent and badly suffixed. The suffix wins.
	rule_testing.ExpectFindings(t, rule_testing.Run(t, ReactHookRequireResultNaming, resultNamingFile,
		resultNamingDeclarations+
			"export function Panel() {\n    const dataResult = useAccountQuery();\n    return dataResult;\n}\n"),
		"hookResultNamingBadSuffix")
}
