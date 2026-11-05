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
func TestReportPython(t *testing.T) {
	t.Parallel()
	runPython(t, "report_test")
}
func TestEntryPython(t *testing.T) {
	t.Parallel()
	runPython(t, "entry_test")
}
func TestCompilerPython(t *testing.T) {
	t.Parallel()
	runPython(t, "compiler_test")
}
func TestGeneratorPython(t *testing.T) {
	t.Parallel()
	runPython(t, "generator_test")
}
func TestNodeReadinessPython(t *testing.T) {
	t.Parallel()
	runPython(t, "node_readiness_test")
}

func runPython(t *testing.T, module string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()
	path := filepath.Join(t.TempDir(), "python.log")
	log, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	command := toolchain.CommandContextIn(ctx, ".", "python3", "-E", "-S", "-B", "-m", "unittest", module)
	command.Env = append(command.Env, "PYTHONDONTWRITEBYTECODE=1")
	command.Stdout, command.Stderr = log, log
	runErr := command.Run()
	if err := log.Close(); err != nil {
		t.Fatal(err)
	}
	output, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if runErr != nil {
		t.Fatalf("%s: %v\n%s", module, runErr, output)
	}
	t.Logf("%s: %s", module, output)
}
