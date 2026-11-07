package oracles

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

func requireOracleArtifacts(test *testing.T, directory string, names ...string) {
	test.Helper()
	var artifacts []buildcache.Artifact
	for _, name := range names {
		artifacts = append(artifacts, buildcache.Artifact{Name: name})
	}
	buildcache.RequireArtifacts(test, directory, artifacts...)
}

func requireOracleExecutable(test *testing.T, binary string, args ...string) {
	test.Helper()
	buildcache.RequireArtifacts(
		test,
		filepath.Dir(binary),
		buildcache.Artifact{Name: filepath.Base(binary), Executable: true, Arguments: args},
	)
}

func requireManifestOracle(test *testing.T, binary string) {
	test.Helper()
	path := filepath.Join(test.TempDir(), "empty-manifest")
	if writeError := os.WriteFile(path, nil, 0600); writeError != nil {
		test.Fatal(writeError)
	}
	requireOracleExecutable(test, binary, "--manifest", path)
}

// The membership helper takes a positional manifest and does not skip blank rows.
func requireMembershipOracle(test *testing.T, binary string) {
	test.Helper()
	directory := test.TempDir()
	source, manifest := filepath.Join(directory, "probe.tsx"), filepath.Join(directory, "manifest")
	if writeError := os.WriteFile(source, []byte("const value = <div />;\n"), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	if writeError := os.WriteFile(manifest, []byte(source+"\n"), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	requireOracleExecutable(test, binary, manifest)
}

func requireCheckerAnswer(test *testing.T) {
	test.Helper()
	prepareCheckerAnswer(test)
	path := filepath.Join(repository, "cohere/TypeScript/tsc/testdata/fixtures/compiler/checker.ts")
	oracle, readError := os.ReadFile(lintGoOracleProduct(test))
	if readError != nil {
		test.Fatal(readError)
	}
	source, readError := os.ReadFile(path)
	if readError != nil {
		test.Fatal(readError)
	}
	inputs := compilerAgreementCheckerAllInputs(
		oracle,
		[]compilerAgreementCase{{path: path, key: compilerAgreementCheckerKey, rule: "all"}},
		[][]byte{source},
	)
	requireOracleArtifacts(test, oracleProductDirectory(test, inputs, nil), "answer")
}

func requireProfileAnswers(test *testing.T) {
	test.Helper()
	compilationCase(test)
	directory := test.TempDir()
	copyPort(test, directory, "", "")
	prepareRegistry(test, directory)
	sourcePath := ownedWitnesses(test, directory, "no-var")[0]
	path := manifest(test, []string{sourcePath + "\tno-var"})
	oraclePath := lintGoOracleProduct(test)
	compilationOracleOutputs(test, oraclePath, path)
	binary, readError := os.ReadFile(oraclePath)
	if readError != nil {
		test.Fatal(readError)
	}
	source, readError := os.ReadFile(sourcePath)
	if readError != nil {
		test.Fatal(readError)
	}
	inputs := compilationOracleInputs(test, binary, source)
	requireOracleArtifacts(test, oracleProductDirectory(test, inputs, nil), "wire", "count")
}

func requireYepestaBundle(test *testing.T, directory string) {
	test.Helper()
	requireOracleArtifacts(test, directory, "bundle.json")
	requireManifestOracle(test, filepath.Join(directory, "oracle"))
}
