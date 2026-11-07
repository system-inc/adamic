package oracles

import (
	"context"
	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
	"os"
	"path/filepath"
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
func TestProduct_LintGoOracle(test *testing.T) {
	test.Parallel()
	timed(test, func() { requireManifestOracle(test, lintGoOracleProduct(test)) })
}
func TestProduct_CompleteSuggestionGoOracle(test *testing.T) {
	test.Parallel()
	timed(test, func() {
		requireManifestOracle(test, filepath.Join(
			completeSuggestionOracleProduct(context.Background(), test, completeSuggestionFixture(test, false)),
			"oracle",
		))
	})
}
func TestProduct_jsx_parser(test *testing.T) {
	test.Parallel()
	timed(test, func() { requireOracleExecutable(test, jsxParserOracle(test), "--jsx-kinds") })
}
func TestProduct_jsx_membership(test *testing.T) {
	test.Parallel()
	timed(test, func() { requireMembershipOracle(test, jsxMembershipOracle(test)) })
}
func TestProduct_SuggestionAlongsideGoOracle(test *testing.T) {
	test.Parallel()
	timed(test, func() {
		directory, _ := suggestionAlongsideFixture(test)
		requireManifestOracle(test, suggestionAlongsideOracleProduct(test, directory))
	})
}
func TestProduct_RulesAgreeCapture(test *testing.T) {
	test.Parallel()
	timed(test, func() { requireOracleArtifacts(test, rulesAgreeCapture(test), "rows.json") })
}
func TestProduct_CompilerAgreementCheckerAllAnswer(test *testing.T) {
	test.Parallel()
	timed(test, func() { requireCheckerAnswer(test) })
}
func TestProduct_ProfileCompilationOracleOutputs(test *testing.T) {
	test.Parallel()
	timed(test, func() { requireProfileAnswers(test) })
}
func TestProduct_YepestaSetup(test *testing.T) {
	test.Parallel()
	timed(test, func() { requireYepestaBundle(test, yepestaPrepare(test)) })
}
func timed(test *testing.T, f func()) {
	start := time.Now()
	f()
	test.Logf("cold/fetch wall %.6fs", time.Since(start).Seconds())
}
