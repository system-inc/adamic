package oracles

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/corpusfiles"
)

func TestCheckerAnswerMatchesLiveOracle(test *testing.T) {
	test.Parallel()
	source := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if source == "" {
		test.Skip("set ADAMIC_TYPESCRIPT_SOURCE to pinned v6.0.3")
	}
	corpusfiles.Upstream(test, source, compilerCommit, []string{"src/compiler"}, []string{"*.ts", "*.a"})
	path := filepath.Join(source, "src/compiler/checker.ts")
	live, err := os.ReadFile(path)
	if err != nil {
		test.Fatal(err)
	}
	fixture, err := os.ReadFile(
		filepath.Join(repository, "cohere/TypeScript/tsc/testdata/fixtures/compiler/checker.ts"),
	)
	if err != nil {
		test.Fatal(err)
	}
	if !bytes.Equal(live, fixture) {
		test.Fatal("checker oracle fixture differs from the pinned agreement source")
	}
	got := compilerAgreementCheckerAllAnswer(
		test,
		[]compilerAgreementCase{{path: path, key: compilerAgreementCheckerKey, rule: "all"}},
	)
	want := execute(test, "", lintGoOracleProduct(test), "--manifest", manifest(test, []string{path + "\tall"})).output
	if !bytes.Equal(got, want) {
		test.Fatal("snapshotted checker answer differs from the live Go oracle")
	}
	test.Logf("%d byte-identical live/snapshotted Go answer bytes", len(want))
}
