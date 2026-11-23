package cloud

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/toolchain"
)

// The CLI tests use only disposable local Git origins and synthetic Go events.
// Real three-way worker execution is recorded separately in the unit report. python3 is toolchain's, and it runs with
// git's environment, since the script drives only git: no home and no global configuration, so its fixtures set
// every identity they commit with.
func TestLintWaveCheck(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "lint-wave-check-tests.txt")
	output, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	// -B writes no bytecode, and -X pycache_prefix reads it only under an empty directory of the test's own, never a
	// __pycache__ in the tree, which no key holds; -s leaves the user site, and a usercustomize there, off sys.path.
	command := toolchain.Command("python3", "-B", "-s", "-X", "pycache_prefix="+t.TempDir(), "lint-wave-check-test.py")
	if command.Err == nil {
		command.Env = toolchain.GitEnv()
	}
	command.Stdout = output
	command.Stderr = output
	runError := command.Run()
	if err := output.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if runError != nil {
		t.Fatalf("worker check CLI tests: %v\n%s", runError, data)
	}
	t.Logf("worker check CLI tests passed:\n%s", data)
}
