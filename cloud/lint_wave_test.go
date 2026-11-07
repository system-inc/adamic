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
func TestLintWaveCheck(test *testing.T) {
	test.Parallel()
	path := filepath.Join(test.TempDir(), "lint-wave-check-tests.txt")
	output, createError := os.Create(path)
	if createError != nil {
		test.Fatal(createError)
	}
	command := toolchain.Command("python3", "-B", "lint-wave-check-test.py")
	if command.Err == nil {
		command.Env = toolchain.GitEnv()
	}
	command.Stdout = output
	command.Stderr = output
	runError := command.Run()
	if closeError := output.Close(); closeError != nil {
		test.Fatal(closeError)
	}
	data, readError := os.ReadFile(path)
	if readError != nil {
		test.Fatal(readError)
	}
	if runError != nil {
		test.Fatalf("worker check CLI tests: %v\n%s", runError, data)
	}
	test.Logf("worker check CLI tests passed:\n%s", data)
}
