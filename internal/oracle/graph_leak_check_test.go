package oracle

import (
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
)

// These adapters keep graph review probes on the graph worker's shared leak
// check while preserving main's existing oracle leak path.
func leakChecked(t *testing.T, code, sanitized string) string {
	t.Helper()
	report, err := leakcheck.Check(leakcheck.Program{
		C:         code,
		Sanitized: sanitized,
		Counted:   filepath.Join(t.TempDir(), "counted"),
		Execute: func(environment []string, name string, arguments ...string) leakcheck.Run {
			return leakRun(executeWith(t, environment, name, arguments...))
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return report
}

func leakRun(result run) leakcheck.Run {
	return leakcheck.Run{Stdout: result.stdout, Stderr: result.stderr, ExitCode: result.exitCode}
}
