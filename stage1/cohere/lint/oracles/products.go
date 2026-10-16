package oracles

import (
	"context"
	"testing"
)

func LintGoOracle(t *testing.T) string { ensureEnvironment(); return lintGoOracleProduct(t) }
func CompleteSuggestionOracle(ctx context.Context, t *testing.T, directory string) string {
	ensureEnvironment()
	return completeSuggestionOracleProduct(ctx, t, directory)
}
func JSXParserOracle(t *testing.T) string     { ensureEnvironment(); return jsxParserOracle(t) }
func JSXMembershipOracle(t *testing.T) string { ensureEnvironment(); return jsxMembershipOracle(t) }
func RulesAgreeCapture(t *testing.T) string   { ensureEnvironment(); return rulesAgreeCapture(t) }
func ShardsAgreeCapture(t *testing.T) []string {
	ensureEnvironment()
	return testShardsAgreeUpstream(t)
}
func ProfileOracleOutputs(t *testing.T, oracle, manifestPath string) ([]byte, []byte) {
	ensureEnvironment()
	return compilationOracleOutputs(t, oracle, manifestPath)
}
func YepestaSetup(t *testing.T) string { ensureEnvironment(); return yepestaPrepare(t) }

type CheckerRow struct{ Path, Key, Rule string }

func CheckerAllAnswer(t *testing.T, rows []CheckerRow) []byte {
	ensureEnvironment()
	owned := make([]compilerAgreementCase, len(rows))
	for i, c := range rows {
		owned[i] = compilerAgreementCase{path: c.Path, key: c.Key, rule: c.Rule}
	}
	return compilerAgreementCheckerAllAnswer(t, owned)
}
func SuggestionAlongsideOracle(t *testing.T, directory string) string {
	ensureEnvironment()
	return suggestionAlongsideOracleProduct(t, directory)
}
func FindProfileOracleOutputs(t *testing.T) { ensureEnvironment(); prepareProfileOracleOutputs(t) }
func FindSuggestionAlongsideOracle(t *testing.T) {
	ensureEnvironment()
	directory, _ := suggestionAlongsideFixture(t)
	suggestionAlongsideOracleProduct(t, directory)
}
