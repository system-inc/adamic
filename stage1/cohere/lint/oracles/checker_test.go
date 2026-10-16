package oracles

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/corpusfiles"
)

func TestCheckerAnswerMatchesLiveOracle(t *testing.T) {
	t.Parallel()
	source := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if source == "" {
		t.Skip("set ADAMIC_TYPESCRIPT_SOURCE to pinned v6.0.3")
	}
	corpusfiles.Upstream(t, source, compilerCommit, []string{"src/compiler"}, []string{"*.ts", "*.a"})
	path := filepath.Join(source, "src/compiler/checker.ts")
	live, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	fixture, err := os.ReadFile(filepath.Join(repository, "cohere/TypeScript/tsc/testdata/fixtures/compiler/checker.ts"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(live, fixture) {
		t.Fatal("checker oracle fixture differs from the pinned agreement source")
	}
	got := compilerAgreementCheckerAllAnswer(t, []compilerAgreementCase{{path: path, key: compilerAgreementCheckerKey, rule: "all"}})
	want := execute(t, "", lintGoOracleProduct(t), "--manifest", manifest(t, []string{path + "\tall"})).output
	if !bytes.Equal(got, want) {
		t.Fatal("snapshotted checker answer differs from the live Go oracle")
	}
	t.Logf("%d byte-identical live/snapshotted Go answer bytes", len(want))
}
