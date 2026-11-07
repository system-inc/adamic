package meter

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/toolchain"
)

// Keep these Python suites in the Go gate too. Their local fixtures include Go
// compiler witnesses; each suite has its own bound and file-backed output.
func TestReportPython(test *testing.T) {
	test.Parallel()
	runPython(test, "report_test")
}
func TestEntryPython(test *testing.T) {
	test.Parallel()
	runPython(test, "entry_test")
}
func TestCompilerPython(test *testing.T) {
	test.Parallel()
	runPython(test, "compiler_test")
}
func TestGeneratorPython(test *testing.T) {
	test.Parallel()
	runPython(test, "generator_test")
}
func TestNodeReadinessPython(test *testing.T) {
	test.Parallel()
	runPython(test, "node_readiness_test")
}

func runPython(test *testing.T, module string) {
	test.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()
	path := filepath.Join(test.TempDir(), "python.log")
	log, createError := os.Create(path)
	if createError != nil {
		test.Fatal(createError)
	}
	command := toolchain.CommandContextIn(ctx, ".", "python3", "-E", "-S", "-B", "-m", "unittest", module)
	command.Env = append(command.Env, "PYTHONDONTWRITEBYTECODE=1")
	command.Stdout, command.Stderr = log, log
	runError := command.Run()
	if closeError := log.Close(); closeError != nil {
		test.Fatal(closeError)
	}
	output, readError := os.ReadFile(path)
	if readError != nil {
		test.Fatal(readError)
	}
	if runError != nil {
		test.Fatalf("%s: %v\n%s", module, runError, output)
	}
	test.Logf("%s: %s", module, output)
}
