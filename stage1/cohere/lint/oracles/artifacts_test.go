package oracles

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

func requireOracleArtifacts(t *testing.T, directory string, names ...string) {
	t.Helper()
	var artifacts []buildcache.Artifact
	for _, name := range names {
		artifacts = append(artifacts, buildcache.Artifact{Name: name})
	}
	buildcache.RequireArtifacts(t, directory, artifacts...)
}

func requireOracleExecutable(t *testing.T, binary string, args ...string) {
	t.Helper()
	buildcache.RequireArtifacts(t, filepath.Dir(binary), buildcache.Artifact{Name: filepath.Base(binary), Executable: true, Arguments: args})
}

func requireManifestOracle(t *testing.T, binary string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "empty-manifest")
	if err := os.WriteFile(path, nil, 0600); err != nil {
		t.Fatal(err)
	}
	requireOracleExecutable(t, binary, "--manifest", path)
}

// The membership helper takes a positional manifest and does not skip blank rows.
func requireMembershipOracle(t *testing.T, binary string) {
	t.Helper()
	directory := t.TempDir()
	source, manifest := filepath.Join(directory, "probe.tsx"), filepath.Join(directory, "manifest")
	if err := os.WriteFile(source, []byte("const value = <div />;\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifest, []byte(source+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	requireOracleExecutable(t, binary, manifest)
}

func requireCheckerAnswer(t *testing.T) {
	t.Helper()
	prepareCheckerAnswer(t)
	path := filepath.Join(repository, "cohere/TypeScript/tsc/testdata/fixtures/compiler/checker.ts")
	oracle, err := os.ReadFile(lintGoOracleProduct(t))
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	inputs := compilerAgreementCheckerAllInputs(oracle, []compilerAgreementCase{{path: path, key: compilerAgreementCheckerKey, rule: "all"}}, [][]byte{source})
	requireOracleArtifacts(t, oracleProductDirectory(t, inputs, nil), "answer")
}

func requireProfileAnswers(t *testing.T) {
	t.Helper()
	compilationCase(t)
	directory := t.TempDir()
	copyPort(t, directory, "", "")
	prepareRegistry(t, directory)
	sourcePath := ownedWitnesses(t, directory, "no-var")[0]
	path := manifest(t, []string{sourcePath + "\tno-var"})
	oraclePath := lintGoOracleProduct(t)
	compilationOracleOutputs(t, oraclePath, path)
	binary, err := os.ReadFile(oraclePath)
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	inputs := compilationOracleInputs(t, binary, source)
	requireOracleArtifacts(t, oracleProductDirectory(t, inputs, nil), "wire", "count")
}

func requireShardsCapture(t *testing.T) {
	t.Helper()
	rows := testShardsAgreeUpstream(t)
	if len(rows) == 0 {
		t.Fatal("capture has no cases")
	}
	directory := oracleProductDirectory(t, testShardsAgreeInputs(t, "capture"), nil)
	requireOracleArtifacts(t, directory, "rows")
}

func requireYepestaBundle(t *testing.T, directory string) {
	t.Helper()
	requireOracleArtifacts(t, directory, "bundle.json")
	requireManifestOracle(t, filepath.Join(directory, "oracle"))
}
