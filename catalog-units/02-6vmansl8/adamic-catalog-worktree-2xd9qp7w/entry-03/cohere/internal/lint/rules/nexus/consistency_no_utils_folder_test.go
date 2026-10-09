package nexus

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

func TestConsistencyNoUtilsFolderFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		fileName string
		wantIds  []string
	}{
		{"/project/source/utils/Thing.ts", []string{"noUtils"}},
		{"/project/source/_utils/Thing.ts", []string{"noUnderscoreUtils"}},
		{"/project/utils/nested/deep/Thing.ts", []string{"noUtils"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.fileName, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoUtilsFolder, testCase.fileName, "export const Value = 1;\n")
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

func TestConsistencyNoUtilsFolderStaysSilent(t *testing.T) {
	t.Parallel()

	// The last two are the ones that matter. A substring match would flag both, and a rule that
	// fires on the exact spelling it is asking for is worse than no rule at all.
	cases := []string{
		"/project/source/utilities/Thing.ts",
		"/project/source/_utilities/Thing.ts",
		"/project/source/components/Thing.ts",
		"/project/source/utilsomething/Thing.ts",
		"/project/source/myutils/Thing.ts",
	}
	for _, fileName := range cases {
		t.Run(fileName, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoUtilsFolder, fileName, "export const Value = 1;\n")
			rule_testing.ExpectClean(t, result)
		})
	}
}
