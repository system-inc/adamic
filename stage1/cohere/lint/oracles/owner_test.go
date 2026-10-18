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
func TestProduct_LintGoOracle(t *testing.T) {
	t.Parallel()
	timed(t, func() { requireManifestOracle(t, lintGoOracleProduct(t)) })
}
func TestProduct_CompleteSuggestionGoOracle(t *testing.T) {
	t.Parallel()
	timed(t, func() {
		requireManifestOracle(t, filepath.Join(completeSuggestionOracleProduct(context.Background(), t, completeSuggestionFixture(t, false)), "oracle"))
	})
}
func TestProduct_jsx_parser(t *testing.T) {
	t.Parallel()
	timed(t, func() { requireOracleExecutable(t, jsxParserOracle(t), "--jsx-kinds") })
}
func TestProduct_jsx_membership(t *testing.T) {
	t.Parallel()
	timed(t, func() { requireMembershipOracle(t, jsxMembershipOracle(t)) })
}
func TestProduct_SuggestionAlongsideGoOracle(t *testing.T) {
	t.Parallel()
	timed(t, func() {
		directory, _ := suggestionAlongsideFixture(t)
		requireManifestOracle(t, suggestionAlongsideOracleProduct(t, directory))
	})
}
func TestProduct_RulesAgreeCapture(t *testing.T) {
	t.Parallel()
	timed(t, func() { requireOracleArtifacts(t, rulesAgreeCapture(t), "rows.json") })
}
func TestProduct_CompilerAgreementCheckerAllAnswer(t *testing.T) {
	t.Parallel()
	timed(t, func() { requireCheckerAnswer(t) })
}
func TestProduct_ProfileCompilationOracleOutputs(t *testing.T) {
	t.Parallel()
	timed(t, func() { requireProfileAnswers(t) })
}
func TestProduct_YepestaSetup(t *testing.T) {
	t.Parallel()
	timed(t, func() { requireYepestaBundle(t, yepestaPrepare(t)) })
}
func timed(t *testing.T, f func()) {
	start := time.Now()
	f()
	t.Logf("cold/fetch wall %.6fs", time.Since(start).Seconds())
}
