package oracles

import (
	"context"
	"testing"
)

func LintGoOracle(test *testing.T) string { ensureEnvironment(); return lintGoOracleProduct(test) }
func CompleteSuggestionOracle(ctx context.Context, test *testing.T, directory string) string {
	ensureEnvironment()
	return completeSuggestionOracleProduct(ctx, test, directory)
}
func JSXParserOracle(test *testing.T) string { ensureEnvironment(); return jsxParserOracle(test) }
func JSXMembershipOracle(test *testing.T) string {
	ensureEnvironment()
	return jsxMembershipOracle(test)
}
func RulesAgreeCapture(test *testing.T) string { ensureEnvironment(); return rulesAgreeCapture(test) }
func ShardsAgreeCapture(test *testing.T) []string {
	ensureEnvironment()
	return testShardsAgreeUpstream(test)
}
func ProfileOracleOutputs(test *testing.T, oracle, manifestPath string) ([]byte, []byte) {
	ensureEnvironment()
	return compilationOracleOutputs(test, oracle, manifestPath)
}
func YepestaSetup(test *testing.T) string { ensureEnvironment(); return yepestaPrepare(test) }

type CheckerRow struct{ Path, Key, Rule string }

func CheckerAllAnswer(test *testing.T, rows []CheckerRow) []byte {
	ensureEnvironment()
	owned := make([]compilerAgreementCase, len(rows))
	for i, c := range rows {
		owned[i] = compilerAgreementCase{path: c.Path, key: c.Key, rule: c.Rule}
	}
	return compilerAgreementCheckerAllAnswer(test, owned)
}
func SuggestionAlongsideOracle(test *testing.T, directory string) string {
	ensureEnvironment()
	return suggestionAlongsideOracleProduct(test, directory)
}
func FindProfileOracleOutputs(test *testing.T) {
	ensureEnvironment()
	prepareProfileOracleOutputs(test)
}
func FindSuggestionAlongsideOracle(test *testing.T) {
	ensureEnvironment()
	directory, _ := suggestionAlongsideFixture(test)
	suggestionAlongsideOracleProduct(test, directory)
}
