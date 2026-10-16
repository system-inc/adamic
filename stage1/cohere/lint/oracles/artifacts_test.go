package oracles

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

// These two getters return answer bytes rather than a directory. Resolve the already-held
// product by its existing identity through buildcache's portable path API; do not rebuild it.
func heldOracleDirectory(t *testing.T, inputs buildcache.Inputs) string {
	t.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	inputs.Name += " oracle-owner-v1"
	return buildcache.Absolute("<build cache>/" + buildcache.NameKey(root, inputs))
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
	requireOracleArtifacts(t, heldOracleDirectory(t, inputs), "answer")
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
	witness := strings.SplitN(compilationCase(t), "\t", 2)[0]
	inputs := buildcache.Inputs{Name: "lint-profile-oracle-outputs", Files: []string{witness, "stage1/cohere/lint/oracles"}, Flags: []string{"mode=no-var", fmt.Sprintf("oracle-sha256=%x", sha256.Sum256(binary)), fmt.Sprintf("case-sha256=%x", sha256.Sum256(source))}, Toolchain: []string{runtime.GOOS, runtime.GOARCH}}
	requireOracleArtifacts(t, heldOracleDirectory(t, inputs), "wire", "count")
}

func requireShardsCapture(t *testing.T) {
	t.Helper()
	rows := testShardsAgreeUpstream(t)
	if len(rows) == 0 {
		t.Fatal("capture has no cases")
	}
	directory := heldOracleDirectory(t, testShardsAgreeInputs(t, "capture"))
	requireOracleArtifacts(t, directory, "rows")
}

func requireYepestaBundle(t *testing.T, directory string) {
	t.Helper()
	requireOracleArtifacts(t, directory, "bundle.json")
	requireManifestOracle(t, filepath.Join(directory, "oracle"))
}
