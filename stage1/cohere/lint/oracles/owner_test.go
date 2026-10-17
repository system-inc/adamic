package oracles

import (
	"context"
	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	if err := os.Chdir(".."); err != nil {
		panic(err)
	}
	if _, err := registry.Stale("."); err != nil {
		panic(err)
	}
	ensureEnvironment()
	var err error
	sharedDirectory, err = os.MkdirTemp("", "lint-oracles-")
	if err != nil {
		panic(err)
	}
	oracleProducer = true
	code := m.Run()
	os.RemoveAll(sharedDirectory)
	os.Exit(code)
}
func TestProduct_LintGoOracle(t *testing.T) {
	t.Parallel()
	timed(t, func() { lintGoOracleProduct(t) })
}
func TestProduct_CompleteSuggestionGoOracle(t *testing.T) {
	t.Parallel()
	timed(t, func() { completeSuggestionOracleProduct(context.Background(), t, completeSuggestionFixture(t, false)) })
}
func TestProduct_jsx_parser(t *testing.T) { t.Parallel(); timed(t, func() { jsxParserOracle(t) }) }
func TestProduct_jsx_membership(t *testing.T) {
	t.Parallel()
	timed(t, func() { jsxMembershipOracle(t) })
}
func TestProduct_SuggestionAlongsideGoOracle(t *testing.T) {
	t.Parallel()
	timed(t, func() { directory, _ := suggestionAlongsideFixture(t); suggestionAlongsideOracleProduct(t, directory) })
}
func TestProduct_RulesAgreeCapture(t *testing.T) {
	t.Parallel()
	timed(t, func() { rulesAgreeCapture(t) })
}
func TestProduct_ShardsAgreeCapture(t *testing.T) {
	t.Parallel()
	timed(t, func() { testShardsAgreeUpstream(t) })
}
func TestProduct_CompilerAgreementCheckerAllAnswer(t *testing.T) {
	t.Parallel()
	timed(t, func() { prepareCheckerAnswer(t) })
}
func TestProduct_ProfileCompilationOracleOutputs(t *testing.T) {
	t.Parallel()
	timed(t, func() { prepareProfileOracleOutputs(t) })
}
func TestProduct_YepestaSetup(t *testing.T) { t.Parallel(); timed(t, func() { yepestaPrepare(t) }) }
func timed(t *testing.T, f func()) {
	start := time.Now()
	f()
	t.Logf("cold/fetch wall %.6fs", time.Since(start).Seconds())
}
